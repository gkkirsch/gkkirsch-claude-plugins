// director-slack — Slack CLI for Claude Code skills.
// Single static binary, zero runtime deps. Token in macOS Keychain,
// caches in ~/.cache/director-slack/.
package main

import (
	"fmt"
	"os"
)

const usage = `director-slack — Slack CLI

Usage:
  director-slack <command> [args...]

Commands:
  send <target> <message>            Send a message
  read <target> [-n N]               Read channel history
  search <query>                     Search messages
  thread read <ref>                  Read a thread
  thread reply <ref> <message>       Reply to a thread
  react <ref> <emoji>                Add a reaction
  upload <target> <path> [-m MSG]    Upload a file
  activity [--days N]                Recent activity (last 7d default)
  resolve <target> [--debug]         Debug target resolution
  cache [--clear|--rate-status|--warm-users]
  auth [--reauth|--clear|--status]   OAuth + keychain management

Targets:
  #channel-name       Channel by name
  C01ABC123           Channel by ID
  Alice               DM by name
  Alice Smith         DM by full name
  @alice.smith        DM by handle
  alice@example.com   DM by email
  Alice, Bob          Group DM (comma-separated)

Output is JSON for sends/cache/auth; plain text for reads/search/activity.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, rest := os.Args[1], os.Args[2:]
	switch cmd {
	case "send":
		cmdSend(rest)
	case "read":
		cmdRead(rest)
	case "search":
		cmdSearch(rest)
	case "thread":
		cmdThread(rest)
	case "react":
		cmdReact(rest)
	case "upload":
		cmdUpload(rest)
	case "activity":
		cmdActivity(rest)
	case "resolve":
		cmdResolve(rest)
	case "cache":
		cmdCache(rest)
	case "auth":
		cmdAuth(rest)
	case "-h", "--help", "help":
		fmt.Print(usage)
	case "version", "-v", "--version":
		fmt.Println("director-slack 1.0.0")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", cmd, usage)
		os.Exit(2)
	}
}

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+msg+"\n", args...)
	os.Exit(1)
}

func dieJSON(err string) {
	emitJSON(map[string]any{"success": false, "error": err})
	os.Exit(1)
}
