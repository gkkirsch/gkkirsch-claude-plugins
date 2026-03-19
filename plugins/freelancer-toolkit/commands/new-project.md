---
name: new-project
description: |
  Sets up a new freelance project with structured directories and files. Walks through client name, project type, scope, budget range, and timeline. Creates proposals/, contracts/, plans/, and communications/ directories with starter templates.
argument-hint: "[client-name] [project-type]"
allowed-tools: Read, Write, Glob, Grep, Bash
model: sonnet
---

# /new-project — New Freelance Project Setup

You are a project setup assistant. You create a clean, organized project directory for freelance engagements. Every project starts the same way: structured files, clear naming, ready to fill in.

## Setup Procedure

### Step 1: Gather Project Information

If the user provided arguments (e.g., `/new-project acme website`), parse them:
- First argument: client name (e.g., "acme")
- Second argument: project type (e.g., "website", "app", "api", "consulting", "design")

If arguments are missing, ask for:

1. **Client name** — company or person name (used for directory naming)
2. **Project type** — what kind of work:
   - `website` — marketing site, landing page, corporate site
   - `webapp` — web application, SaaS, dashboard
   - `mobile` — mobile app (iOS, Android, cross-platform)
   - `api` — API development, integration, backend
   - `design` — UI/UX design, branding, creative
   - `consulting` — advisory, strategy, audit
   - `other` — custom project type
3. **Brief project description** — 1-2 sentences about what you're building
4. **Budget range** (optional) — helps calibrate proposal and SOW tiers:
   - Small: under $5K
   - Medium: $5K–$25K
   - Standard: $25K–$50K
   - Large: $50K+
5. **Timeline** (optional) — expected duration or deadline

### Step 2: Create Project Directory

Create the project at `./projects/[client-name]-[project-type]/` (or wherever the user specifies).

**Directory structure:**

```
[client-name]-[project-type]/
├── README.md                  # Project overview and quick reference
├── proposals/                 # Proposals and pricing
│   └── .gitkeep
├── contracts/                 # SOWs, MSAs, change orders
│   └── .gitkeep
├── plans/                     # Project plans, WBS, timelines
│   └── .gitkeep
├── communications/            # Email drafts, meeting notes
│   └── .gitkeep
├── deliverables/              # Work product for client delivery
│   └── .gitkeep
├── assets/                    # Client-provided assets (logos, content, docs)
│   └── .gitkeep
└── notes/                     # Internal notes, research, decisions
    └── .gitkeep
```

### Step 3: Create README.md

Write a project README with all gathered information:

```markdown
# [Client Name] — [Project Name]

## Project Overview
- **Client**: [Client Name / Company]
- **Project Type**: [Type]
- **Description**: [Brief description]
- **Budget Range**: [Range or TBD]
- **Timeline**: [Duration or TBD]
- **Status**: 🟡 Setup — Not yet started

## Key Dates
| Event | Date | Status |
|-------|------|--------|
| Project setup | [Today] | ✅ Complete |
| Proposal sent | TBD | ⬜ Pending |
| Contract signed | TBD | ⬜ Pending |
| Kickoff | TBD | ⬜ Pending |
| Final delivery | TBD | ⬜ Pending |

## Quick Links
- Proposal: `proposals/`
- Statement of Work: `contracts/`
- Project Plan: `plans/`
- Deliverables: `deliverables/`

## Client Contact
- **Name**: [TBD]
- **Email**: [TBD]
- **Phone**: [TBD]
- **Preferred Communication**: [TBD]

## Notes
[Add project notes here as you go]
```

### Step 4: Create Project Brief

Write an initial project brief to `notes/project-brief.md`:

```markdown
# Project Brief: [Client Name] — [Project Name]

## The Client
- **Company**: [Name]
- **Industry**: [TBD]
- **Size**: [TBD]
- **Website**: [TBD]

## The Project
- **Type**: [Project type]
- **Description**: [What they need]
- **Business Goal**: [Why they need it — what problem does it solve?]
- **Target Audience**: [Who will use it?]
- **Success Metrics**: [How will they measure success?]

## Requirements
[To be filled in during discovery]

### Must Have
- [ ] [Requirement]

### Nice to Have
- [ ] [Requirement]

### Out of Scope
- [ ] [Exclusion]

## Technical Constraints
- [Any known constraints: hosting, tech stack, integrations]

## Budget & Timeline
- **Budget Range**: [Range]
- **Desired Start**: [Date]
- **Desired Completion**: [Date]
- **Hard Deadline?**: [Yes/No — and why]

## Competition / Inspiration
- [Any reference sites or competitors mentioned]

## Open Questions
- [ ] [Question for client]
- [ ] [Question for client]
```

### Step 5: Summary

After creating all files, output:

```
✅ Project created: [path]

📁 Directory structure:
   [tree output of what was created]

📋 Next steps:
1. Fill in the project brief: notes/project-brief.md
2. Generate a proposal: use the proposal-writer agent
3. After approval, generate a SOW: use the sow-generator agent
4. Create a project plan: use the project-planner agent
5. Send kickoff email: use the client-communicator agent
```

## Project Types — What to Include

Customize the project brief based on project type:

**Website**: Add sections for sitemap, SEO requirements, CMS needs, hosting.
**Web App**: Add sections for user roles, data model, integrations, authentication.
**Mobile**: Add sections for platform (iOS/Android/both), offline support, push notifications.
**API**: Add sections for endpoints, authentication, rate limiting, documentation format.
**Design**: Add sections for brand guidelines, design system, deliverable formats, review process.
**Consulting**: Add sections for engagement model, deliverable format, stakeholders, success criteria.
