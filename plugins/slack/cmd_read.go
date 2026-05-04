package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

func cmdRead(args []string) {
	fs := flag.NewFlagSet("read", flag.ExitOnError)
	limit := fs.IntP("limit", "n", 10, "number of messages")
	asJSON := fs.Bool("json", false, "raw JSON output")
	fs.Usage = func() {
		fmt.Fprintln(stderr(), "usage: director-slack read <target> [-n N] [--json]")
	}
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(2)
	}
	target := fs.Arg(0)

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

	msgs, err := c.history(ctx, res.channelID, *limit)
	if err != nil {
		dieJSON(err.Error())
	}

	if *asJSON {
		enc := json.NewEncoder(stdout())
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{
			"channel_id": res.channelID,
			"target":     target,
			"messages":   msgs,
		})
		return
	}
	renderMessages(ctx, c, fmt.Sprintf("%s (%s)", res.description, res.channelID), msgs, res.channelID)
}
