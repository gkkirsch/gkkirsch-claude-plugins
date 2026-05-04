package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func stdout() io.Writer { return os.Stdout }
func stderr() io.Writer { return os.Stderr }

func emitJSON(v any) {
	enc := json.NewEncoder(stdout())
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func emitJSONErr(v any) {
	enc := json.NewEncoder(stderr())
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func getEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

// renderMessages prints messages in the agent-friendly format used by the
// Python skill: timestamp header, [Display|id:UID] text, msg ref, replies, reactions.
func renderMessages(ctx context.Context, c *apiClient, header string, msgs []slackMessage, channelID string) {
	fmt.Fprintf(stdout(), "=== %s ===\n", header)
	fmt.Fprintf(stdout(), "Last %d messages\n\n", len(msgs))
	for _, m := range msgs {
		ts := tsToTime(m.Ts)
		fmt.Fprintf(stdout(), "--- %s ---\n", ts.Format("2006-01-02 15:04:05"))
		name := displayName(ctx, c, m.User, m.Username, m.BotID)
		fmt.Fprintf(stdout(), "[%s|id:%s] %s\n", name, m.User, strings.ReplaceAll(m.Text, "\n", "\n  "))
		fmt.Fprintf(stdout(), "  msg: %s/%s\n", channelID, m.Ts)
		extras := []string{}
		if m.ReplyCount > 0 {
			extras = append(extras, fmt.Sprintf("+%d replies", m.ReplyCount))
		}
		for _, r := range m.Reactions {
			extras = append(extras, fmt.Sprintf(":%s:(%d)", r.Name, r.Count))
		}
		if len(extras) > 0 {
			fmt.Fprintf(stdout(), "  %s\n", strings.Join(extras, " | "))
		}
		for _, f := range m.Files {
			fmt.Fprintf(stdout(), "  file: %s (%s)\n", f.Name, f.URL)
		}
		fmt.Fprintln(stdout())
	}
}

func displayName(ctx context.Context, c *apiClient, userID, username, botID string) string {
	if userID == "" && username != "" {
		return username
	}
	if userID == "" && botID != "" {
		return "bot:" + botID
	}
	if name := userNameGet(userID); name != "" {
		return name
	}
	if c == nil {
		return userID
	}
	u, err := c.usersInfo(ctx, userID)
	if err != nil || u == nil {
		return userID
	}
	name := u.bestName()
	userNameSet(userID, name)
	return name
}

// tsToTime parses a Slack ts ("1736934225.000100") to a time.Time.
func tsToTime(ts string) time.Time {
	s := ts
	if i := strings.IndexByte(ts, '.'); i >= 0 {
		s = ts[:i]
	}
	var sec int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return time.Time{}
		}
		sec = sec*10 + int64(c-'0')
	}
	return time.Unix(sec, 0)
}
