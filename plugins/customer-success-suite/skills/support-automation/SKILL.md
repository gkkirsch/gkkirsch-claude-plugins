---
name: support-automation
description: >
  Customer support automation and self-service patterns.
  Use when building help centers, implementing chatbots, creating
  auto-response systems, or designing escalation workflows.
  Triggers: "support automation", "help center", "chatbot", "auto-reply",
  "ticket routing", "escalation", "self-service", "knowledge base", "FAQ".
  NOT for: onboarding flows (see onboarding-retention), marketing emails, sales automation.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Support Automation

## Ticket Routing & Classification

```typescript
// lib/ticket-router.ts
interface SupportTicket {
  id: string;
  subject: string;
  body: string;
  from: string;
  priority: 'low' | 'medium' | 'high' | 'urgent';
  category: string | null;
  assignedTo: string | null;
  status: 'new' | 'open' | 'pending' | 'resolved' | 'closed';
  createdAt: Date;
  metadata: {
    userPlan: string;
    accountAge: number; // days
    previousTickets: number;
    healthScore: number;
  };
}

interface RoutingRule {
  name: string;
  condition: (ticket: SupportTicket) => boolean;
  action: {
    assignTo?: string;       // team or agent
    priority?: SupportTicket['priority'];
    autoReply?: string;      // template ID
    escalate?: boolean;
    tags?: string[];
  };
}

const routingRules: RoutingRule[] = [
  {
    name: 'VIP escalation',
    condition: (t) => t.metadata.userPlan === 'enterprise' || t.metadata.healthScore < 20,
    action: { assignTo: 'senior-support', priority: 'high', escalate: true },
  },
  {
    name: 'Billing auto-route',
    condition: (t) => /billing|invoice|charge|refund|cancel/i.test(t.subject + t.body),
    action: { assignTo: 'billing-team', tags: ['billing'] },
  },
  {
    name: 'Bug report',
    condition: (t) => /bug|error|crash|broken|not working|500/i.test(t.subject + t.body),
    action: { assignTo: 'engineering-support', priority: 'high', tags: ['bug'] },
  },
  {
    name: 'How-to questions',
    condition: (t) => /how (do|can|to)|where (is|do)|what (is|does)/i.test(t.subject),
    action: { autoReply: 'knowledge-base-search', tags: ['how-to'] },
  },
  {
    name: 'New user grace',
    condition: (t) => t.metadata.accountAge < 7 && t.metadata.previousTickets === 0,
    action: { priority: 'high', tags: ['new-user'], assignTo: 'onboarding-support' },
  },
];

function routeTicket(ticket: SupportTicket): SupportTicket {
  for (const rule of routingRules) {
    if (rule.condition(ticket)) {
      return {
        ...ticket,
        assignedTo: rule.action.assignTo ?? ticket.assignedTo,
        priority: rule.action.priority ?? ticket.priority,
        category: rule.name,
      };
    }
  }
  // Default: round-robin to general support
  return { ...ticket, assignedTo: 'general-support' };
}
```

## Auto-Response Templates

```typescript
// templates/auto-responses.ts
interface ResponseTemplate {
  id: string;
  trigger: string;
  subject: string;
  body: string;
  requiresFollowUp: boolean;
}

const templates: ResponseTemplate[] = [
  {
    id: 'password-reset',
    trigger: 'password|reset|cant login|locked out',
    subject: 'Password Reset Instructions',
    body: `Hi {{name}},

You can reset your password here: {{resetLink}}

This link expires in 24 hours. If you didn't request this, you can safely ignore this email.

If you're still having trouble, reply to this email and we'll help you get back in.

Best,
{{agentName}}`,
    requiresFollowUp: false,
  },
  {
    id: 'knowledge-base-search',
    trigger: 'how to|where is|how do',
    subject: 'Re: {{originalSubject}}',
    body: `Hi {{name}},

Thanks for reaching out! Based on your question, these articles might help:

{{#each matchedArticles}}
- [{{this.title}}]({{this.url}})
{{/each}}

If none of these address your question, reply to this email and a team member will follow up within {{slaHours}} hours.

Best,
{{productName}} Support`,
    requiresFollowUp: true,
  },
  {
    id: 'bug-acknowledgment',
    trigger: 'bug|error|broken|not working',
    subject: 'Re: {{originalSubject}} — We\'re on it',
    body: `Hi {{name}},

Thanks for reporting this. We've logged it and our engineering team is investigating.

**What we know so far:**
- Issue received: {{timestamp}}
- Ticket #: {{ticketId}}
- Priority: {{priority}}

We'll update you within {{slaHours}} hours with next steps.

In the meantime, {{workaroundSuggestion}}

Best,
{{agentName}}`,
    requiresFollowUp: true,
  },
];
```

## Knowledge Base Structure

```markdown
# Knowledge Base Architecture

## Category Structure (recommended)
```
Getting Started/
├── Quick Start Guide
├── Account Setup
├── First Project Tutorial
└── Team Invitations

Features/
├── [Feature A]/
│   ├── Overview
│   ├── How to [common task]
│   └── Advanced: [power user feature]
├── [Feature B]/
│   └── ...

Integrations/
├── Slack Integration
├── GitHub Integration
└── API Documentation

Billing & Account/
├── Plans & Pricing
├── Upgrade/Downgrade
├── Invoice & Payment
└── Cancel Account

Troubleshooting/
├── Common Errors
├── Performance Issues
├── Browser Compatibility
└── Known Issues
```

## Article Template
```markdown
# [Action-oriented title: "How to..." or "Setting up..."]

**Last updated**: [date]
**Applies to**: [Free / Pro / Enterprise] plans

## Overview
[1-2 sentences explaining what this article covers and when you'd need it]

## Prerequisites
- [Requirement 1]
- [Requirement 2]

## Steps

### Step 1: [Action verb + object]
[Instructions with screenshot]

### Step 2: [Action verb + object]
[Instructions with screenshot]

## Troubleshooting

### [Common issue 1]
**Symptom**: [What the user sees]
**Fix**: [Steps to resolve]

### [Common issue 2]
**Symptom**: [What the user sees]
**Fix**: [Steps to resolve]

## Related Articles
- [Link 1]
- [Link 2]
```

## SLA & Escalation Rules
```
| Priority | First Response | Resolution Target | Escalation |
|----------|---------------|-------------------|------------|
| Urgent   | 1 hour        | 4 hours           | → VP Eng after 2h |
| High     | 4 hours       | 24 hours          | → Team Lead after 8h |
| Medium   | 8 hours       | 48 hours          | → Manager after 24h |
| Low      | 24 hours      | 5 business days   | → Auto-close after 14d |
```

## Satisfaction Measurement

```typescript
// lib/csat.ts
interface SurveyResponse {
  ticketId: string;
  rating: 1 | 2 | 3 | 4 | 5;
  comment?: string;
  respondedAt: Date;
}

function calculateCSAT(responses: SurveyResponse[]): {
  score: number;        // Percentage of 4-5 ratings
  average: number;      // Mean rating
  nps: number;          // Net Promoter Score approximation
  responseRate: number; // Surveys sent vs responded
  breakdown: Record<number, number>;
} {
  if (responses.length === 0) {
    return { score: 0, average: 0, nps: 0, responseRate: 0, breakdown: {} };
  }

  const breakdown: Record<number, number> = { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 };
  let sum = 0;

  for (const r of responses) {
    breakdown[r.rating]++;
    sum += r.rating;
  }

  const satisfied = breakdown[4] + breakdown[5];
  const detractors = breakdown[1] + breakdown[2];
  const total = responses.length;

  return {
    score: Math.round((satisfied / total) * 100),
    average: Math.round((sum / total) * 10) / 10,
    nps: Math.round(((breakdown[5] - detractors) / total) * 100),
    responseRate: 0, // Set externally based on surveys sent
    breakdown,
  };
}
```

## Gotchas

1. **Auto-closing tickets without resolution** -- Auto-closing "no response after 7 days" feels efficient but frustrates users who were busy. Send 2 follow-ups before closing, and make it trivial to re-open. A closed-without-resolution ticket = a silent churn signal.

2. **Bot deflection rate as a vanity metric** -- "Our chatbot handles 60% of tickets" means nothing if users are frustrated and churning. Measure resolution rate (was the issue actually solved?), not deflection rate. Track how many bot-handled users open follow-up tickets.

3. **Knowledge base articles nobody reads** -- If articles don't appear in search results or auto-responses, they're invisible. Connect your KB to ticket routing (auto-suggest relevant articles), in-app help (contextual tooltips), and chatbot responses. An unlinked article is a dead article.

4. **SLA based on business hours only** -- Your SLA says "4 hours" but that's 4 business hours. A Friday 5pm urgent ticket gets a Monday 1pm response. If you serve global customers, define SLAs in real hours for urgent/high priority, or staff weekend coverage.

5. **No feedback loop to product** -- Support tickets are the richest source of product feedback. If ticket patterns (repeated feature requests, common confusion points) don't reach the product team weekly, you're fixing symptoms not causes. Tag tickets with product areas and report trends.

6. **Over-automating human problems** -- Billing disputes, angry enterprise customers, and data loss incidents need a human. Never auto-respond to cancellation requests or angry messages with a bot. Route emotional/high-stakes tickets to senior human agents immediately.
