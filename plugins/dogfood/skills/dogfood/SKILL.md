---
name: dogfood
description: "Use after building any feature, fixing any bug, or making any non-trivial code change. Drives the outermost layer the user touches (browser for UI, real HTTP call for endpoints, real message send for integrations, real agent run for orchestrator behavior), captures concrete proof (screenshots, response bodies, db rows, log lines, traces), and — most importantly — builds a tight feedback loop you can iterate inside before you say done. Never claim the feature works without exercising it end-to-end and producing an artifact. Keywords: e2e, end-to-end, test, verify, validate, prove, smoke test, regression, browser test, integration test, dogfood, ship, done, feedback loop, iterate, debug."
---

# Dogfood: prove it works at the outermost layer, with a loop you can iterate inside

## The principle

A feature is not done until you've used it *as the user would* and seen it succeed.

- Code that compiles is not code that works.
- Tests that pass are not features that work.
- A code review approving the diff is not the diff producing the right output.

Only exercising the outermost layer — the one a real user touches — and watching the system respond correctly counts as evidence.

## Match the rigor to the risk (not everything needs the full loop)

This skill describes the *full* e2e + feedback-loop discipline. That is the right tool when the work is non-trivial, user-facing, or has real downside if it's wrong. It is *overkill* for:

- Doc and comment edits
- Trivial renames / type fixes
- One-shot scripts that run once and never again
- Code you're throwing away after a single experiment
- Pure internal refactors with no behavior change (where existing tests are the contract)

For those: a quick smoke test (does it still compile / does the page still load) is the right amount of evidence. Don't manufacture a full loop where there's nothing to iterate on.

For everything else — anything a real user will exercise, anything that changes behavior, anything where "it broke" is a regression — run the loop. The cost is small; the cost of shipping a broken outer layer is large.

If you're not sure which bucket a change is in, default to the loop. The overhead of testing something that didn't strictly need it is much smaller than the overhead of debugging something you didn't.

## The deeper principle: build a tight feedback loop *first*

The most important thing you can do for yourself is **design a fast, accurate, observable loop**. If your loop is slow, opaque, or unreliable, every iteration is a guess and you'll burn through tries without converging on the real problem.

A good feedback loop has three properties:

1. **Fast** — you can run it in seconds, not minutes. If a single iteration takes a minute, you'll do five and quit; if it takes ten seconds, you'll do thirty and find it.
2. **Specific** — when something fails, you see *why*, not just *that*. Logs, stack traces, structured errors, network responses, db state — not "it broke."
3. **Reachable** — you have the *levers* to act on what you see: reload, retry, mutate state, restart the server, query the db, kill the agent, change the input.

Before you start testing, ask: *do I have a loop I can iterate inside?* If not, build one before you build anything else.

## Step 1 — find your levers

You can't drive what you can't observe, and you can't iterate without controls. Inventory both before you start:

**Read levers** (observe what's happening)
- Server logs (`tail -f path/to/log`)
- Browser console (`agent-browser console` or the read_console_messages tool)
- Browser network panel (read_network_requests)
- Database queries (`sqlite3 db.sqlite3 "select …"`, `psql -c "…"`)
- HTTP responses (`curl -i …`)
- Agent traces (`roster trace <agent-id> --tail 30`)
- Process state (`pgrep -lf …`, `lsof -iTCP:<port>`)
- File system (`ls`, `cat`, `find -newer`)

**Write levers** (act on what you see)
- Reload / restart the dev server
- Click / type via agent-browser
- Hit an endpoint with curl
- Mutate db rows (test data, reset state)
- Send a real message via the slack/messaging plugin
- Spawn or notify an agent
- Edit a file and let the watcher reload

If you don't know what levers your specific app has, **find out before testing**. Read the README, scan `package.json` / `Makefile` / `Procfile`, ask the user. Flying blind is the slow path.

## Step 2 — identify the outermost layer

Different features surface at different layers. Test at the outermost — that exercises every layer beneath as a side effect.

| You built… | Outermost layer | How to exercise |
|---|---|---|
| A page, button, form, panel | The rendered UI in a browser | agent-browser navigate + click + assert DOM |
| An HTTP endpoint | The request | curl, capture response status + body |
| A CLI command / binary | The shell | run with sample args, check stdout + exit code |
| A background job / queue worker | The trigger + side effect | enqueue, then check the side effect (db, log, file) |
| An agent / orchestrator behavior | The agent's run | spawn or notify, read trace, validate output |
| An integration (Slack/Telegram/iMessage/email) | The recipient app | send a real message, verify it appears |
| A file / artifact (PDF, image, doc) | The rendered file | open it, screenshot, inspect content |
| A database migration / schema change | The data model | introspect schema + sample row counts before/after |

If a feature spans layers (UI → API → DB), test at the outermost (UI). Don't shortcut to "I'll just call the API directly" — you'd be skipping exactly the layer the user lives in.

## Step 3 — exercise it (per-category playbooks)

### UI changes

Goal: render → interact → see the response.

```bash
# Make sure dev server is running
lsof -iTCP:5173 -sTCP:LISTEN

# Open the page
agent-browser open "http://localhost:5173/the-route"
agent-browser snapshot -i        # find your new element by ref
agent-browser click @e3          # or fill, select, etc.
agent-browser snapshot -i        # re-snapshot to see the post-click DOM
agent-browser screenshot /tmp/proof.png
```

If your change has a side effect (writes a row, fires an API call), verify *that* too — UI saying "saved" doesn't mean it actually saved.

### Backend / endpoint changes

Goal: real request → expected response → side effect verified.

```bash
curl -sS -i -X POST http://localhost:8080/api/widgets \
  -H 'Content-Type: application/json' \
  -d '{"name":"test"}'

# Now verify the side effect
sqlite3 ./data.db "SELECT * FROM widgets ORDER BY id DESC LIMIT 1;"
```

A 200 response is necessary but not sufficient. Always check whatever the endpoint was *supposed to cause*.

### CLI / binary

```bash
./your-binary --flag value sample-input
echo "exit=$?"
# Inspect outputs / files written / state changed
```

### Background jobs

Trigger the job, then look for its side effect. Examples:
- Job writes to a queue → `redis-cli LRANGE queue 0 -1`
- Job inserts a row → query the table
- Job sends an email → check the email recipient (or the smtp log)
- Job hits an external API → check the external system, or the outbound HTTP log

### Agent / orchestrator behavior

The trace IS your proof. After triggering the agent:

```bash
roster notify <agent-id> "<test prompt>" --from dogfood
# wait a beat for it to actually do work
roster trace <agent-id> --tail 30
roster describe <agent-id>
```

Look for: did it use the tools you expected, did it produce the expected output shape, did it errors flag in the trace, did it complete (status: ready) or stall (status: streaming forever).

### Integrations (Slack, Telegram, iMessage, email)

The outermost layer is the recipient seeing the message. Two paths:

**Path A — read it back via API.** If you have a tool that talks to the integration, send + read.

```bash
# Slack (using the slack plugin):
slack send "#test-channel" "dogfood probe $(date +%s)"
slack read "#test-channel" -n 1   # confirm it landed
```

**Path B — open the recipient's web client.** If no API tool, open the web app in agent-browser and look at the result.

```bash
agent-browser open "https://app.slack.com/client/.../test-channel"
agent-browser snapshot -i
# look for the message text in the snapshot
```

For iMessage on macOS: send a real text from another device or Apple ID, then watch the orch process the inbound. The Director Tasks panel + the dispatcher's chat surface both show what arrived.

### File / artifact generation

Open the file, see it rendered. For images: screenshot via agent-browser. For PDFs: render to image first (`pdftoppm`) or open in Preview. For markdown: inspect the rendered HTML or the structured content.

### Database migrations / schema changes

```bash
# Before
sqlite3 db "SELECT sql FROM sqlite_master WHERE type='table' AND name='widgets';"

# Run migration
./migrate up

# After
sqlite3 db "SELECT sql FROM sqlite_master WHERE type='table' AND name='widgets';"
sqlite3 db "SELECT count(*) FROM widgets;"   # row preservation check
```

## Step 4 — capture concrete proof

Proof is an *artifact* the user can examine, not a sentence. Pick at least one:

- **Screenshot file path** — for visual changes
- **HTTP response body** — pasted in the report
- **SQL query result** — pasted in the report
- **Log line excerpt** — grep'd to the relevant lines
- **`roster trace` excerpt** — for agent runs
- **"Recipient saw it"** — explicit confirmation from the recipient app

If you can't produce an artifact, you haven't tested. "It works" without an artifact is faith, not evidence.

## Step 5 — iterate with intent

If your first run fails, *don't* immediately patch and retry. Use the loop:

1. **Read the failure signal.** Logs, console errors, response body, trace. What does it actually say?
2. **Form a hypothesis.** "The bug is X because Y." If you can't articulate a hypothesis, you don't have enough information yet — go find more (better logs, db state, inspector, etc.).
3. **Make the smallest change that tests the hypothesis.** One variable at a time.
4. **Run the loop again.** Same exact steps. Did the signal change?
5. **If iterations feel slow or vague, stop and improve the loop itself.** Add a log line. Pretty-print a payload. Add a `console.log`. Improving the loop pays back every subsequent iteration.

If you're three iterations in and not converging, your loop is broken. Fix the loop, not the bug.

## Step 6 — report

Tell the user, in this shape:

> **What I tested:** the [outermost layer] — e.g. "the new task creation flow in the Tasks panel"
> **How I exercised it:** [commands / steps]
> **Proof:** [paths / excerpts / references]
> **Anything I couldn't test:** [be honest — no test theatre]

If you can't honestly say "I drove the outer layer and it worked," don't say "it works."

## Anti-patterns (do not do these)

- ❌ "Tests pass." → tests verify modules, not the system end-to-end.
- ❌ "TypeScript compiles." → compilation says nothing about runtime behavior.
- ❌ "I read the code, it looks right." → reading is not testing.
- ❌ "The API returned 200." → 200 doesn't mean the data landed correctly.
- ❌ "I mocked the dependency and the test passed." → mocks verify your test setup, not your code.
- ❌ "It worked before, the change is small, it's fine." → optimism is not evidence.
- ❌ Skipping the side-effect check. The visible response is one signal; what the system actually did is another. Verify both.

## When you genuinely can't reach the outer layer

Sometimes the outermost layer isn't reachable in your sandbox: a paid API, a hardware device, a real customer interaction, a production-only resource. When that's true:

1. **Say so explicitly.** "I can't exercise X end-to-end because Y."
2. **Test the closest reachable layer instead.** Document what that layer is and what it does/doesn't prove.
3. **Mark the feature as "implemented, not E2E-verified"** so the user knows what they're shipping.

Honest "I tested layer N-1 because layer N requires Z" beats a confident-sounding "it works" that turns into a customer-facing bug.

## Keep the loop honest

The point of this skill is not ceremony. It's that an agent which won't ship without proof will produce dramatically fewer regressions than one which won't ship without optimism. The cost of a tight loop is small. The cost of shipping broken code is large.

Build the loop. Use it. Show the proof.
