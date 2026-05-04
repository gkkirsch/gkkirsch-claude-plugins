package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

func cmdActivity(args []string) {
	fs := flag.NewFlagSet("activity", flag.ExitOnError)
	days := fs.IntP("days", "d", 7, "look back N days")
	since := fs.StringP("since", "s", "", "since YYYY-MM-DD (overrides --days)")
	limit := fs.IntP("limit", "n", 100, "max messages per query")
	asJSON := fs.Bool("json", false, "raw JSON output")
	fs.Usage = func() {
		fmt.Fprintln(stderr(), "usage: director-slack activity [--days N | --since YYYY-MM-DD] [-n N] [--json]")
	}
	_ = fs.Parse(args)

	cutoff := time.Now().AddDate(0, 0, -*days)
	if *since != "" {
		t, err := time.Parse("2006-01-02", *since)
		if err != nil {
			dieJSON("invalid --since: " + err.Error())
		}
		cutoff = t
	}
	after := cutoff.Format("2006-01-02")

	tok, err := ensureToken()
	if err != nil {
		dieJSON(err.Error())
	}
	c := newClient(tok)
	ctx := context.Background()

	from, err := c.searchMessages(ctx, "from:me after:"+after, *limit, "timestamp")
	if err != nil {
		dieJSON(err.Error())
	}
	to, err := c.searchMessages(ctx, "to:me after:"+after, *limit, "timestamp")
	if err != nil {
		dieJSON(err.Error())
	}

	type bucket struct {
		Channel string
		Name    string
		Sent    int
		Recv    int
		Last    time.Time
	}
	channels := map[string]*bucket{}
	track := func(hits []searchHit, sent bool) {
		for _, h := range hits {
			b, ok := channels[h.Channel.ID]
			if !ok {
				b = &bucket{Channel: h.Channel.ID, Name: h.Channel.Name}
				channels[h.Channel.ID] = b
			}
			if sent {
				b.Sent++
			} else {
				b.Recv++
			}
			t := tsToTime(h.Ts)
			if t.After(b.Last) {
				b.Last = t
			}
		}
	}
	track(from.Matches, true)
	track(to.Matches, false)

	if *asJSON {
		enc := json.NewEncoder(stdout())
		enc.SetIndent("", "  ")
		_ = enc.Encode(channels)
		return
	}

	var rows []*bucket
	for _, b := range channels {
		rows = append(rows, b)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Last.After(rows[j].Last) })

	fmt.Fprintf(stdout(), "=== activity since %s ===\n\n", after)
	for _, b := range rows {
		name := b.Name
		if name == "" {
			name = b.Channel
		} else if !strings.HasPrefix(name, "#") {
			name = "#" + name
		}
		fmt.Fprintf(stdout(), "%s\n", name)
		fmt.Fprintf(stdout(), "  sent: %d | received: %d | last: %s\n",
			b.Sent, b.Recv, b.Last.Format("2006-01-02 15:04"))
	}
	if len(rows) == 0 {
		fmt.Fprintln(stdout(), "(no activity in window)")
	}
	_ = os.Stdout.Sync()
}
