package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type targetType int

const (
	tgtUnknown targetType = iota
	tgtChannelName
	tgtChannelID
	tgtUserDM
	tgtUserHandle
	tgtUserEmail
	tgtGroupDM
)

type parsedTarget struct {
	kind    targetType
	raw     string
	channel string   // for channel name (no #) or channel ID
	user    string   // for DM (name/handle/email/ID)
	users   []string // for group DM
}

var (
	reChannelID = regexp.MustCompile(`^[CG][A-Z0-9]{8,}$`)
	reUserID    = regexp.MustCompile(`^U[A-Z0-9]{8,}$`)
	reEmail     = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

func parseTarget(s string) parsedTarget {
	s = strings.TrimSpace(s)
	pt := parsedTarget{raw: s}
	if s == "" {
		return pt
	}
	if reChannelID.MatchString(s) {
		pt.kind = tgtChannelID
		pt.channel = s
		return pt
	}
	if reUserID.MatchString(s) {
		pt.kind = tgtUserDM
		pt.user = s
		return pt
	}
	if strings.HasPrefix(s, "#") {
		pt.kind = tgtChannelName
		pt.channel = strings.ToLower(strings.TrimPrefix(s, "#"))
		return pt
	}
	if reEmail.MatchString(s) {
		pt.kind = tgtUserEmail
		pt.user = strings.ToLower(s)
		return pt
	}
	if strings.HasPrefix(s, "@") {
		pt.kind = tgtUserHandle
		pt.user = strings.ToLower(strings.TrimPrefix(s, "@"))
		return pt
	}
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		var names []string
		ok := true
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" || reChannelID.MatchString(p) || reUserID.MatchString(p) {
				ok = false
				break
			}
			names = append(names, p)
		}
		if ok && len(names) >= 2 {
			pt.kind = tgtGroupDM
			pt.users = names
			return pt
		}
	}
	pt.kind = tgtUserDM
	pt.user = s
	return pt
}

func (p parsedTarget) display() string {
	switch p.kind {
	case tgtChannelName:
		return "channel #" + p.channel
	case tgtChannelID:
		return "channel " + p.channel
	case tgtUserDM:
		return "DM to " + p.user
	case tgtUserHandle:
		return "DM to @" + p.user
	case tgtUserEmail:
		return "DM to " + p.user
	case tgtGroupDM:
		return "group DM with " + strings.Join(p.users, ", ")
	default:
		return "unknown target: " + p.raw
	}
}

// resolution represents a successful target → channel ID resolution.
type resolution struct {
	channelID   string
	description string
	source      string
}

// resolveTarget routes a parsed target to the right Slack lookup.
func resolveTarget(ctx context.Context, c *apiClient, raw string) (resolution, error) {
	pt := parseTarget(raw)
	switch pt.kind {
	case tgtChannelID:
		return resolution{channelID: pt.channel, description: pt.display(), source: "direct"}, nil
	case tgtChannelName:
		id, err := c.findChannelByName(ctx, pt.channel)
		if err != nil {
			return resolution{}, err
		}
		return resolution{channelID: id, description: pt.display(), source: "channel_lookup"}, nil
	case tgtUserDM:
		uid, source, err := resolveUserName(ctx, c, pt.user)
		if err != nil {
			return resolution{}, err
		}
		ch, err := c.openDM(ctx, uid)
		if err != nil {
			return resolution{}, err
		}
		return resolution{channelID: ch, description: pt.display(), source: source}, nil
	case tgtUserHandle:
		uid, source, err := resolveUserHandle(ctx, c, pt.user)
		if err != nil {
			return resolution{}, err
		}
		ch, err := c.openDM(ctx, uid)
		if err != nil {
			return resolution{}, err
		}
		return resolution{channelID: ch, description: pt.display(), source: source}, nil
	case tgtUserEmail:
		u, err := c.lookupByEmail(ctx, pt.user)
		if err != nil {
			return resolution{}, err
		}
		ch, err := c.openDM(ctx, u.ID)
		if err != nil {
			return resolution{}, err
		}
		return resolution{channelID: ch, description: "DM to " + u.bestName(), source: "email"}, nil
	case tgtGroupDM:
		var ids []string
		for _, name := range pt.users {
			uid, _, err := resolveUserName(ctx, c, name)
			if err != nil {
				return resolution{}, fmt.Errorf("resolve %q: %w", name, err)
			}
			ids = append(ids, uid)
		}
		ch, err := c.openGroupDM(ctx, ids)
		if err != nil {
			return resolution{}, err
		}
		return resolution{channelID: ch, description: pt.display(), source: "group_dm"}, nil
	default:
		return resolution{}, fmt.Errorf("could not parse target: %s", raw)
	}
}

// resolveUserName turns a free-form name like "Alice" into a Slack user ID.
// Strategy: recent contacts cache → search.messages from:<name> → error.
func resolveUserName(ctx context.Context, c *apiClient, name string) (uid, source string, err error) {
	if reUserID.MatchString(name) {
		return name, "direct_id", nil
	}
	if uid := recentContactsLookup(name); uid != "" {
		return uid, "recent_contacts", nil
	}
	uid, err = searchUserByName(ctx, c, name)
	if err != nil {
		return "", "", err
	}
	return uid, "search", nil
}

func resolveUserHandle(ctx context.Context, c *apiClient, handle string) (uid, source string, err error) {
	if uid := recentContactsLookupHandle(handle); uid != "" {
		return uid, "recent_contacts_by_handle", nil
	}
	uid, err = searchUserByName(ctx, c, handle)
	if err != nil {
		return "", "", fmt.Errorf("no user found with handle @%s", handle)
	}
	return uid, "search_by_handle", nil
}

// searchUserByName uses search.messages with from:<name> to find a user ID.
// Falls through to a users.list scan as a last resort would be too slow on
// large workspaces, so we accept that totally-new users may not resolve until
// they appear in search.
func searchUserByName(ctx context.Context, c *apiClient, name string) (string, error) {
	res, err := c.searchMessages(ctx, "from:"+name, 5, "timestamp")
	if err != nil {
		return "", err
	}
	if res != nil && len(res.Matches) > 0 {
		for _, m := range res.Matches {
			if m.User != "" {
				recentContactsAdd(name, m.User, m.Username)
				return m.User, nil
			}
		}
	}
	return "", fmt.Errorf("no user found matching %q", name)
}
