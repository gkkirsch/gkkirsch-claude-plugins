package main

import (
	"context"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

func cmdUpload(args []string) {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	message := fs.StringP("message", "m", "", "optional message")
	title := fs.StringP("title", "t", "", "file title (defaults to filename)")
	dryRun := fs.Bool("dry-run", false, "resolve target but don't upload")
	fs.Usage = func() {
		fmt.Fprintln(stderr(), "usage: director-slack upload <target> <path> [-m MESSAGE] [-t TITLE] [--dry-run]")
	}
	_ = fs.Parse(args)
	if fs.NArg() < 2 {
		fs.Usage()
		os.Exit(2)
	}
	target := fs.Arg(0)
	path := fs.Arg(1)

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
			"path":        path,
		})
		return
	}

	if err := c.uploadFile(ctx, res.channelID, path, *message, *title); err != nil {
		dieJSON(err.Error())
	}
	emitJSON(map[string]any{
		"success":     true,
		"action":      "uploaded",
		"target":      target,
		"description": res.description,
		"channel_id":  res.channelID,
		"path":        path,
	})
}
