package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

func cmdSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	limit := fs.IntP("limit", "n", 20, "number of results")
	sort := fs.String("sort", "timestamp", "sort: timestamp|score")
	asJSON := fs.Bool("json", false, "raw JSON output")
	fs.Usage = func() {
		fmt.Fprintln(stderr(), `usage: director-slack search <query> [-n N] [--sort timestamp|score] [--json]`)
	}
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(2)
	}
	query := strings.Join(fs.Args(), " ")

	tok, err := ensureToken()
	if err != nil {
		dieJSON(err.Error())
	}
	c := newClient(tok)
	ctx := context.Background()

	r, err := c.searchMessages(ctx, query, *limit, *sort)
	if err != nil {
		dieJSON(err.Error())
	}
	if *asJSON {
		enc := json.NewEncoder(stdout())
		enc.SetIndent("", "  ")
		_ = enc.Encode(r)
		return
	}
	fmt.Fprintf(stdout(), "=== search: %q (%d total) ===\n\n", query, r.Total)
	for _, m := range r.Matches {
		fmt.Fprintf(stdout(), "--- %s ---\n", tsToTime(m.Ts).Format("2006-01-02 15:04:05"))
		channel := m.Channel.Name
		if channel != "" {
			channel = "#" + channel
		} else {
			channel = m.Channel.ID
		}
		fmt.Fprintf(stdout(), "[%s in %s] %s\n", m.Username, channel, m.Text)
		fmt.Fprintf(stdout(), "  msg: %s/%s\n", m.Channel.ID, m.Ts)
		if m.Permalink != "" {
			fmt.Fprintf(stdout(), "  link: %s\n", m.Permalink)
		}
		fmt.Fprintln(stdout())
	}
}
