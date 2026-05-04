package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

// --- resolve ---

func cmdResolve(args []string) {
	fs := flag.NewFlagSet("resolve", flag.ExitOnError)
	debug := fs.Bool("debug", false, "parse only, don't hit Slack")
	fs.Usage = func() {
		fmt.Fprintln(stderr(), "usage: director-slack resolve <target> [--debug]")
	}
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(2)
	}
	target := fs.Arg(0)
	pt := parseTarget(target)
	out := map[string]any{
		"raw_input":   target,
		"parsed_type": kindName(pt.kind),
		"display":     pt.display(),
		"parsed_values": map[string]any{
			"channel": pt.channel,
			"user":    pt.user,
			"users":   pt.users,
		},
	}
	if !*debug {
		tok, err := ensureToken()
		if err != nil {
			out["resolution"] = map[string]any{"success": false, "error": err.Error()}
			emitJSON(out)
			return
		}
		c := newClient(tok)
		ctx := context.Background()
		res, err := resolveTarget(ctx, c, target)
		if err != nil {
			out["resolution"] = map[string]any{"success": false, "error": err.Error()}
		} else {
			out["resolution"] = map[string]any{
				"success":    true,
				"channel_id": res.channelID,
				"source":     res.source,
			}
		}
	}
	emitJSON(out)
}

func kindName(k targetType) string {
	switch k {
	case tgtChannelName:
		return "CHANNEL_NAME"
	case tgtChannelID:
		return "CHANNEL_ID"
	case tgtUserDM:
		return "USER_DM"
	case tgtUserHandle:
		return "USER_HANDLE"
	case tgtUserEmail:
		return "USER_EMAIL"
	case tgtGroupDM:
		return "GROUP_DM"
	default:
		return "UNKNOWN"
	}
}

// --- cache ---

func cmdCache(args []string) {
	fs := flag.NewFlagSet("cache", flag.ExitOnError)
	clear := fs.Bool("clear", false, "clear all caches")
	rateStat := fs.Bool("rate-status", false, "show rate-limit usage")
	_ = fs.Parse(args)

	if *rateStat {
		emitJSON(map[string]any{"success": true, "rate_status": rateStatus()})
		return
	}
	if *clear {
		_ = channelCacheClear()
		_ = recentContactsClear()
		_ = rateClear()
		_ = os.Remove(cachePath("users.json"))
		emitJSON(map[string]any{"success": true, "action": "cleared"})
		return
	}
	emitJSON(map[string]any{
		"success":      true,
		"cache_dir":    cacheDir(),
		"rate_status":  rateStatus(),
	})
}

// --- auth ---

func cmdAuth(args []string) {
	fs := flag.NewFlagSet("auth", flag.ExitOnError)
	reauth := fs.Bool("reauth", false, "force re-authentication")
	clear := fs.Bool("clear", false, "delete the stored token")
	status := fs.Bool("status", false, "report whether a token is stored")
	_ = fs.Parse(args)

	if *clear {
		if err := keychainDelete(); err != nil {
			dieJSON(err.Error())
		}
		emitJSON(map[string]any{"success": true, "action": "cleared"})
		return
	}
	if *status {
		tok, err := keychainGet()
		if err != nil {
			dieJSON(err.Error())
		}
		emitJSON(map[string]any{"success": true, "authenticated": tok != ""})
		return
	}
	if _, err := runOAuth(*reauth); err != nil {
		dieJSON(err.Error())
	}
	emitJSON(map[string]any{"success": true, "action": "authorized"})
}
