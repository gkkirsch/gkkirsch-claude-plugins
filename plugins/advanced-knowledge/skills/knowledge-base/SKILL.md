---
name: knowledge-base
description: Build and maintain a personal knowledge base using LLM-compiled markdown wikis. Inspired by Andrej Karpathy's approach to knowledge management.
when-to-use: When the user asks to compile knowledge, query the wiki, ingest sources, or lint the knowledge base. Also use proactively after research tasks to file findings.
user-invocable: true
---

# Knowledge Base

A Karpathy-style knowledge management system. Raw sources go into the
`library/` root, compiled wiki articles come out in `library/wiki/`.
The LLM is the compiler — it reads, summarizes, extracts concepts,
and builds cross-references.

## Directory Structure

`library/` is the root the user sees in Director's Library panel. Drop
sources directly there; processed outputs get their own subfolders
alongside.

```
library/
├── <source files>      # Drop articles, papers, notes, research, data
│                       # straight here. Anything not under a
│                       # processed subfolder is treated as raw.
└── wiki/               # LLM-compiled articles (the knowledge base)
    ├── concepts/       # One .md per concept
    ├── summaries/      # Source document summaries
    ├── connections/    # Cross-references, relationship maps
    ├── queries/        # Filed-back Q&A outputs
    └── index.md        # Auto-maintained master index
```

> **Note:** Daily logs, learnings, and reflections live in `memory/`
> (a sibling of `library/`, not under it) — use the `/memory` skill
> for those.

When scanning `library/` for raw sources, ignore the `wiki/` subdir
and any other plugin-owned subdirectory the orch carves out
(`audits/`, `research/`, `notes/`, etc).

## Commands

### `/knowledge-base compile`

Scan `library/` for new or updated source files (skipping the
processed subdirs above) and compile into `library/wiki/`.

**Steps:**
1. Glob `library/*.md`, `library/*.txt`, `library/*.pdf`, etc — any
   readable content. Exclude `wiki/` (its files are processed output).
   Exclude `wiki/` and any sibling directory of structured output.
2. For each source file:
   a. Check if a summary already exists in `library/wiki/summaries/`
      (match by filename slug).
   b. If missing or source is newer: read the source, write a summary
      to `library/wiki/summaries/<slug>.md`.
   c. Extract key concepts from the source.
   d. For each concept: create or update
      `library/wiki/concepts/<concept-slug>.md`.
   e. Add cross-references: when a concept article mentions another
      concept, link them.
3. Update `library/wiki/connections/map.md` — a relationship map
   showing how concepts connect.
4. Rebuild `library/wiki/index.md` — master index with links to all
   summaries, concepts, and connections.

**Summary format** (`library/wiki/summaries/<slug>.md`):
```markdown
---
source: <filename>
compiled: 2026-04-05
---

# Summary: <Title>

## Key Points
- ...

## Concepts Extracted
- [[concept-name]] — brief note on relevance

## Source Details
- Type: article/paper/note/research
- Length: ~X words
```

**Concept format** (`library/wiki/concepts/<slug>.md`):
```markdown
---
concept: <Name>
related: [concept-a, concept-b]
sources: [file1.md, file2.md]
last-updated: 2026-04-05
---

# <Concept Name>

<Definition and explanation>

## Key Details
- ...

## Connections
- Related to [[other-concept]] because...

## Sources
- Summarized from [[summary-slug]]
```

**Index format** (`library/wiki/index.md`):
```markdown
# Knowledge Base Index

Last compiled: 2026-04-05

## Concepts
- [Concept A](concepts/concept-a.md) — one-line description
- [Concept B](concepts/concept-b.md) — one-line description

## Summaries
- [Source Title](summaries/source-slug.md) — source type, date compiled

## Connections
- [Relationship Map](connections/map.md)
```

### `/knowledge-base query "<question>"`

Research the wiki to answer a question.

**Steps:**
1. Read `library/wiki/index.md` to understand what's available.
2. Identify relevant wiki articles (concepts, summaries, connections).
3. Read those articles.
4. Synthesize an answer using wiki knowledge.
5. If the wiki doesn't cover the topic, say so and suggest what raw
   sources might help.
6. Save the answer to `library/wiki/queries/<date>-<slug>.md`:

```markdown
---
question: "<original question>"
date: 2026-04-05
sources-consulted: [concepts/x.md, summaries/y.md]
---

# Q: <question>

<synthesized answer>

## Sources Used
- ...
```

### `/knowledge-base lint`

Health-check the wiki for consistency and completeness.

**Steps:**
1. List source files at the root of `library/` — check each has a
   corresponding summary in `library/wiki/summaries/`.
2. List all concept articles — check cross-references point to
   existing files.
3. Check `library/wiki/index.md` lists all concepts and summaries.
4. Look for duplicate or overlapping concept articles.
5. Check for concepts mentioned in summaries but without their own
   article.
6. Report findings in a structured format:

```
## Lint Report — <date>

### Missing Summaries (raw sources without wiki summaries)
- file1.md — no summary found

### Broken References
- concepts/x.md references concepts/y.md — file not found

### Missing Concepts (mentioned but no article)
- "machine learning" mentioned in 3 summaries, no concept article

### Index Sync Issues
- concepts/new-thing.md exists but not in index.md

### Auto-fixes Applied
- Added 2 missing entries to index.md
- Fixed 1 broken cross-reference
```

Auto-fix what's safe (index updates, broken links to renamed files).
Flag everything else.

### `/knowledge-base ingest <url-or-path>`

Add a new source to the knowledge base.

**Steps:**
1. If argument looks like a URL (starts with http):
   a. Fetch the content using WebFetch.
   b. Convert to clean markdown (strip nav, ads, boilerplate).
   c. Save to `library/<slugified-title>.md` with frontmatter noting
      the source URL.
2. If argument is a file path:
   a. Read the file.
   b. Copy content to `library/<filename>`.
3. Run an incremental compile for just the new source (don't
   recompile everything).
4. Report what was added and what concepts were extracted.

**Ingested file frontmatter:**
```markdown
---
source-url: https://example.com/article  # if from URL
ingested: 2026-04-05
type: article
---
```

## Integration with Daily Work

- After any research task: file findings into `library/` (root) and
  run compile.
- When starting a new topic: query the wiki first to see what you
  already know.
- For learnings, reflections, and session logs: use the `/memory`
  skill instead.

## Important

- ALL writes go to `library/` — never to `.claude/`.
- Sources live at the root of `library/`. The wiki and its queries
  live under `library/wiki/`. Memory is separate at `memory/`
  (sibling of library, not nested).
- Use wiki-style links `[[concept-name]]` for cross-references within
  articles.
- Keep concept articles focused — one concept per file.
- Summaries should be standalone — readable without the original
  source.
- The index is the entry point — keep it accurate and complete.
