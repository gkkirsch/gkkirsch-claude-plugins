package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

func cmdSend(args []string) {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "resolve target but don't send")
	noCache := fs.Bool("no-cache", false, "bypass channel cache")
	noReact := fs.Bool("no-react", false, "skip the auto :robot_face: reaction")
	fs.Usage = func() {
		fmt.Fprintln(stderr(), "usage: director-slack send <target> <message> [--dry-run] [--no-cache] [--no-react]")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)
	rest := fs.Args()
	if len(rest) < 2 {
		fs.Usage()
		os.Exit(2)
	}
	target := rest[0]
	message := rest[1]

	if *noCache {
		_ = channelCacheClear()
	}

	tok, err := ensureToken()
	if err != nil {
		dieJSON(err.Error())
	}
	c := newClient(tok)
	ctx := context.Background()

	res, err := resolveTarget(ctx, c, target)
	if err != nil {
		dieJSON(err.Error())
	}
	if *dryRun {
		emitJSON(map[string]any{
			"success":     true,
			"action":      "dry_run",
			"target":      target,
			"description": res.description,
			"channel_id":  res.channelID,
			"source":      res.source,
		})
		return
	}

	ts, err := c.postMessage(ctx, res.channelID, message, "", false)
	if err != nil {
		dieJSON(err.Error())
	}

	if !*noReact {
		_ = c.addReaction(ctx, res.channelID, ts, "robot_face")
	}

	emitJSON(map[string]any{
		"success":     true,
		"action":      "sent",
		"target":      target,
		"description": res.description,
		"channel_id":  res.channelID,
		"ts":          ts,
		"source":      res.source,
	})
}
