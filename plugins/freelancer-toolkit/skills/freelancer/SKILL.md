---
name: freelancer
description: >
  Freelancer Business Toolkit — automate proposals, statements of work, client emails, project plans, and invoices.
  Triggers: "write proposal", "create SOW", "draft client email", "project plan", "new project", "scope change", "payment reminder", "freelancer".
  Dispatches the appropriate specialist agent based on your request.
  NOT for: coding, debugging, deployment, or tasks unrelated to freelance business operations.
version: 1.0.0
argument-hint: "<proposal|sow|email|plan|project> [details]"
user-invocable: true
allowed-tools: Read, Write, Glob, Grep, Bash
model: sonnet
---

# Freelancer Business Toolkit

Stop spending hours on admin. This toolkit automates the business side of freelancing — proposals that win, contracts that protect, emails that get responses, and plans that keep projects on track.

## Available Agents

### Proposal Writer (`proposal-writer`)
Generates polished, persuasive client proposals from project requirements. Supports fixed-price, hourly, retainer, and value-based pricing. Produces client-ready markdown with executive summary, scope, timeline, deliverables, pricing, and terms.

**Invoke**: Dispatch via Task tool with `subagent_type: "proposal-writer"`.

**Example prompts**:
- "Write a proposal for Acme Corp — they need a React web app for inventory management, $30K budget, 8-week timeline."
- "Create a value-based proposal for an e-commerce redesign that should increase conversion by 15%."
- "Draft a retainer proposal for ongoing development work, 20 hours/month."

### SOW Generator (`sow-generator`)
Creates legally-structured Statements of Work with proper scope, deliverables, timelines, payment terms, IP clauses, and termination provisions. Three tiers: Simple (under $5K), Standard ($5K–$50K), and Enterprise ($50K+).

**Invoke**: Dispatch via Task tool with `subagent_type: "sow-generator"`.

**Example prompts**:
- "Generate a SOW for the Acme Corp web app project based on the approved proposal."
- "Create an enterprise SOW for a $75K platform migration with 3 phases."
- "Write a simple SOW for a $3K landing page project."

### Client Communicator (`client-communicator`)
Drafts professional emails for every stage of a freelance engagement. Templates for kickoff, status updates, milestone delivery, scope changes, payment reminders, completion, and follow-ups. Reads your project files to personalize every message.

**Invoke**: Dispatch via Task tool with `subagent_type: "client-communicator"`.

**Example prompts**:
- "Draft a kickoff email for the Acme Corp project starting next Monday."
- "Write a polite payment reminder — Invoice #2024-015 is 10 days overdue."
- "Create a scope change email — client wants to add user authentication, wasn't in the original scope."
- "Draft a project completion email and ask for a testimonial."

### Project Planner (`project-planner`)
Creates comprehensive project plans with Work Breakdown Structure, Gantt-style timeline, risk assessment, resource allocation, and milestone definitions with acceptance criteria.

**Invoke**: Dispatch via Task tool with `subagent_type: "project-planner"`.

**Example prompts**:
- "Create a project plan for a 6-week React web app build with design, development, and QA phases."
- "Plan a website redesign project — I have a designer for 10 hrs/week and I'm doing the development."
- "Break down an API integration project into phases with risk assessment."

## Slash Commands

### `/new-project [client-name] [project-type]`
Sets up a new freelance project with organized directories and starter templates. Creates proposals/, contracts/, plans/, communications/, and deliverables/ folders with a README and project brief.

**Example**: `/new-project acme webapp`

## Typical Workflow

A complete freelance project flows through these stages:

### 1. Project Setup
```
/new-project acme webapp
```
Creates your project structure and initial brief.

### 2. Write Proposal
```
Task tool:
  subagent_type: "proposal-writer"
  description: "Write proposal for Acme Corp"
  prompt: "Write a proposal for Acme Corp — React web app for inventory management. Budget: $25K-35K. Timeline: 8 weeks. Key features: real-time inventory tracking, barcode scanning, reporting dashboard."
  mode: "bypassPermissions"
```

### 3. Generate SOW
After the proposal is approved:
```
Task tool:
  subagent_type: "sow-generator"
  description: "Generate SOW for Acme Corp"
  prompt: "Create a standard SOW based on the approved proposal at proposals/acme-corp-proposal.md. Client: Acme Corp. Payment: 30% upfront, 30% at midpoint, 40% on delivery."
  mode: "bypassPermissions"
```

### 4. Create Project Plan
```
Task tool:
  subagent_type: "project-planner"
  description: "Plan Acme Corp project"
  prompt: "Create a detailed project plan for the Acme inventory app. I'm the solo developer. 30 hrs/week available. Key phases: Design (1 week), Core Development (4 weeks), Testing & Polish (2 weeks), Deployment (1 week)."
  mode: "bypassPermissions"
```

### 5. Communicate with Client
Throughout the project:
```
Task tool:
  subagent_type: "client-communicator"
  description: "Draft kickoff email"
  prompt: "Draft a project kickoff email for Acme Corp. Project starts Monday. I need their brand assets and API documentation by Wednesday. Weekly status updates on Fridays."
  mode: "bypassPermissions"
```

## Reference Materials

This toolkit includes expert reference documents:

- **`references/pricing-strategies.md`** — Value-based pricing, rate calculation formulas, project type pricing guides, retainer vs project pricing decisions
- **`references/proposal-templates.md`** — Complete proposal templates (simple, standard, enterprise), winning structures, objection handling, terms and conditions
- **`references/client-management.md`** — Communication best practices, scope creep prevention, difficult conversations, red flag indicators

Agents automatically consult these references when generating documents.

## Tips for Best Results

1. **Be specific with project details.** The more context you give (client name, budget, timeline, features), the better the output.
2. **Provide existing files.** If you have a proposal, the SOW generator will use it. If you have a SOW, the planner will reference it.
3. **Customize after generation.** Agents produce excellent first drafts — review and personalize before sending to clients.
4. **Use the right agent.** Don't ask the proposal writer to draft emails. Each agent is specialized for its task.
5. **Start with `/new-project`.** A structured project directory makes everything else work better.
