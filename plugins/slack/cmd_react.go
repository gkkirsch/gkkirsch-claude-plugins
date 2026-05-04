package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

func cmdReact(args []string) {
	fs := flag.NewFlagSet("react", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(stderr(), "usage: director-slack react <CHANNEL_ID/TS> <emoji>")
	}
	_ = fs.Parse(args)
	if fs.NArg() < 2 {
		fs.Usage()
		os.Exit(2)
	}
	channelID, ts, err := splitMsgRef(fs.Arg(0))
	if err != nil {
		dieJSON(err.Error())
	}
	emoji := fs.Arg(1)

	tok, err := ensureToken()
	if err != nil {
		dieJSON(err.Error())
	}
	c := newClient(tok)
	ctx := context.Background()

	if err := c.addReaction(ctx, channelID, ts, emoji); err != nil {
		dieJSON(err.Error())
	}
	emitJSON(map[string]any{
		"success":    true,
		"action":     "reacted",
		"channel_id": channelID,
		"ts":         ts,
		"emoji":      emoji,
	})
}
