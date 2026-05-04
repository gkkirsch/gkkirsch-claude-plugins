package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

func cmdThread(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(stderr(), "usage: director-slack thread <read|reply> ...")
		os.Exit(2)
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "read":
		threadRead(rest)
	case "reply":
		threadReply(rest)
	default:
		fmt.Fprintf(stderr(), "unknown thread subcommand: %s\n", sub)
		os.Exit(2)
	}
}

func threadRead(args []string) {
	fs := flag.NewFlagSet("thread read", flag.ExitOnError)
	limit := fs.IntP("limit", "n", 50, "max replies")
	fs.Usage = func() {
		fmt.Fprintln(stderr(), "usage: director-slack thread read <CHANNEL_ID/TS> [-n N]")
	}
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(2)
	}
	channelID, ts, err := splitMsgRef(fs.Arg(0))
	if err != nil {
		dieJSON(err.Error())
	}

	tok, err := ensureToken()
	if err != nil {
		dieJSON(err.Error())
	}
	c := newClient(tok)
	ctx := context.Background()

	msgs, err := c.replies(ctx, channelID, ts, *limit)
	if err != nil {
		dieJSON(err.Error())
	}
	renderMessages(ctx, c, fmt.Sprintf("thread %s/%s", channelID, ts), msgs, channelID)
}

func threadReply(args []string) {
	fs := flag.NewFlagSet("thread reply", flag.ExitOnError)
	broadcast := fs.Bool("broadcast", false, "also post to channel")
	noReact := fs.Bool("no-react", false, "skip auto :robot_face: reaction")
	fs.Usage = func() {
		fmt.Fprintln(stderr(), "usage: director-slack thread reply <CHANNEL_ID/TS> <message> [--broadcast] [--no-react]")
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
	message := strings.Join(fs.Args()[1:], " ")

	tok, err := ensureToken()
	if err != nil {
		dieJSON(err.Error())
	}
	c := newClient(tok)
	ctx := context.Background()

	newTs, err := c.postMessage(ctx, channelID, message, ts, *broadcast)
	if err != nil {
		dieJSON(err.Error())
	}
	if !*noReact {
		_ = c.addReaction(ctx, channelID, newTs, "robot_face")
	}
	emitJSON(map[string]any{
		"success":    true,
		"action":     "replied",
		"channel_id": channelID,
		"thread_ts":  ts,
		"ts":         newTs,
		"broadcast":  *broadcast,
	})
}

func splitMsgRef(ref string) (channelID, ts string, err error) {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid msg ref %q (expected CHANNEL_ID/TS)", ref)
	}
	return parts[0], parts[1], nil
}
