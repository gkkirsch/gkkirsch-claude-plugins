---
name: onboarding-retention
description: >
  Customer onboarding flows and retention patterns for SaaS products.
  Use when designing user onboarding, building activation flows,
  implementing retention campaigns, or reducing churn.
  Triggers: "onboarding", "user activation", "retention", "churn reduction",
  "welcome email", "product tour", "activation metrics", "user lifecycle".
  NOT for: customer support ticketing, billing/payments, marketing acquisition.
version: 1.0.0
allowed-tools: Read, Grep, Glob, Bash
---

# Onboarding & Retention

## Onboarding Flow Architecture

```typescript
// types/onboarding.ts
interface OnboardingStep {
  id: string;
  title: string;
  description: string;
  isCompleted: boolean;
  isRequired: boolean;
  action: () => void | Promise<void>;
  completionCriteria: string; // Human-readable for docs
}

interface OnboardingConfig {
  steps: OnboardingStep[];
  activationMilestone: string; // The "aha moment" step
  maxDaysToActivate: number;
  skipAllowed: boolean;
}

// Example: project management SaaS onboarding
const onboardingConfig: OnboardingConfig = {
  activationMilestone: 'create-first-project',
  maxDaysToActivate: 7,
  skipAllowed: true,
  steps: [
    {
      id: 'complete-profile',
      title: 'Complete your profile',
      description: 'Add your name and team role',
      isCompleted: false,
      isRequired: true,
      action: () => navigateTo('/settings/profile'),
      completionCriteria: 'User has set name and role fields',
    },
    {
      id: 'invite-team',
      title: 'Invite your team',
      description: 'Collaboration works best with teammates',
      isCompleted: false,
      isRequired: false,
      action: () => openInviteModal(),
      completionCriteria: 'At least 1 team invite sent',
    },
    {
      id: 'create-first-project',
      title: 'Create your first project',  // ← Activation milestone
      description: 'Start organizing your work',
      isCompleted: false,
      isRequired: true,
      action: () => navigateTo('/projects/new'),
      completionCriteria: 'User has created 1+ project',
    },
    {
      id: 'add-first-task',
      title: 'Add a task',
      description: 'Break work into actionable items',
      isCompleted: false,
      isRequired: false,
      action: () => openTaskCreator(),
      completionCriteria: 'User has created 1+ task in a project',
    },
  ],
};
```

## Activation Tracking

```typescript
// lib/activation.ts — track user progress toward activation
interface ActivationEvent {
  userId: string;
  event: string;
  timestamp: Date;
  properties?: Record<string, unknown>;
}

interface ActivationMetrics {
  signupDate: Date;
  activatedAt: Date | null;
  daysToActivate: number | null;
  completedSteps: string[];
  currentStep: string | null;
  isActivated: boolean;
  isAtRisk: boolean; // Past half of maxDaysToActivate without activation
}

function calculateActivationMetrics(
  events: ActivationEvent[],
  config: OnboardingConfig,
): ActivationMetrics {
  const signup = events.find(e => e.event === 'user.signed_up');
  if (!signup) throw new Error('No signup event found');

  const completedSteps = config.steps
    .filter(step => events.some(e => e.event === `onboarding.${step.id}.completed`))
    .map(s => s.id);

  const activationEvent = events.find(
    e => e.event === `onboarding.${config.activationMilestone}.completed`
  );

  const daysSinceSignup = Math.floor(
    (Date.now() - signup.timestamp.getTime()) / (1000 * 60 * 60 * 24)
  );

  const nextIncomplete = config.steps.find(s => !completedSteps.includes(s.id));

  return {
    signupDate: signup.timestamp,
    activatedAt: activationEvent?.timestamp ?? null,
    daysToActivate: activationEvent
      ? Math.floor((activationEvent.timestamp.getTime() - signup.timestamp.getTime()) / (1000 * 60 * 60 * 24))
      : null,
    completedSteps,
    currentStep: nextIncomplete?.id ?? null,
    isActivated: !!activationEvent,
    isAtRisk: !activationEvent && daysSinceSignup > config.maxDaysToActivate / 2,
  };
}
```

## Welcome Email Sequence

```markdown
# Welcome Email Sequence (7-day activation drip)

## Email 1: Welcome (Immediately after signup)
**Subject**: Welcome to [Product] — here's your quick start guide
**Goal**: Get user to complete profile
**CTA**: "Complete Your Profile" → link to /settings/profile
**Content**: Brief value prop recap, 3-step quick start, link to help docs

## Email 2: Social proof (Day 1)
**Subject**: How [Company Name] saved 10 hours/week with [Product]
**Goal**: Show real-world value
**CTA**: "See How They Did It" → link to case study
**Trigger**: Only send if user has NOT completed profile yet

## Email 3: Activation nudge (Day 3)
**Subject**: Your first project is waiting
**Goal**: Drive activation milestone
**CTA**: "Create Your First Project" → deep link to /projects/new
**Trigger**: Only if NOT yet activated
**Content**: 60-second video showing project creation

## Email 4: Team invite (Day 5)
**Subject**: [Product] is better with your team
**Goal**: Drive team invites (increases retention 3x)
**CTA**: "Invite Your Team" → link to /settings/team
**Trigger**: Only if activated but no team invites sent

## Email 5: Feature highlight (Day 7)
**Subject**: 3 features you haven't tried yet
**Goal**: Deepen engagement
**CTA**: Feature-specific deep links
**Trigger**: Only if activated

## Email 6: At-risk re-engagement (Day 7)
**Subject**: We noticed you haven't started yet
**Goal**: Save at-risk users
**CTA**: "Book a 15-min Setup Call" → calendly link
**Trigger**: Only if NOT activated by day 7
**Content**: Offer personal onboarding help
```

## Retention Health Score

```typescript
// lib/health-score.ts
interface UserActivity {
  loginCount30d: number;
  actionsPerSession: number;
  featuresBreadth: number; // unique features used / total features
  teamSize: number;
  lastActiveAt: Date;
  supportTickets30d: number;
  npsScore: number | null;
}

interface HealthScore {
  score: number; // 0-100
  grade: 'healthy' | 'neutral' | 'at-risk' | 'churning';
  factors: { name: string; score: number; weight: number }[];
  recommendations: string[];
}

function calculateHealthScore(activity: UserActivity): HealthScore {
  const factors = [
    {
      name: 'Login frequency',
      score: Math.min(activity.loginCount30d / 20, 1) * 100,
      weight: 0.25,
    },
    {
      name: 'Feature adoption',
      score: activity.featuresBreadth * 100,
      weight: 0.20,
    },
    {
      name: 'Team engagement',
      score: Math.min(activity.teamSize / 5, 1) * 100,
      weight: 0.20,
    },
    {
      name: 'Recency',
      score: getRecencyScore(activity.lastActiveAt),
      weight: 0.15,
    },
    {
      name: 'Depth per session',
      score: Math.min(activity.actionsPerSession / 10, 1) * 100,
      weight: 0.10,
    },
    {
      name: 'Satisfaction',
      score: activity.npsScore !== null ? (activity.npsScore / 10) * 100 : 50,
      weight: 0.10,
    },
  ];

  const score = Math.round(
    factors.reduce((sum, f) => sum + f.score * f.weight, 0)
  );

  const grade: HealthScore['grade'] =
    score >= 70 ? 'healthy' :
    score >= 45 ? 'neutral' :
    score >= 20 ? 'at-risk' : 'churning';

  const recommendations: string[] = [];
  const sorted = [...factors].sort((a, b) => a.score - b.score);

  if (sorted[0].name === 'Team engagement') {
    recommendations.push('Send team invite reminder — multi-user accounts retain 3x better');
  }
  if (sorted[0].name === 'Feature adoption') {
    recommendations.push('Trigger in-app feature tour for unused high-value features');
  }
  if (sorted[0].name === 'Login frequency') {
    recommendations.push('Send re-engagement email with personalized content');
  }

  return { score, grade, factors, recommendations };
}

function getRecencyScore(lastActive: Date): number {
  const daysSince = (Date.now() - lastActive.getTime()) / (1000 * 60 * 60 * 24);
  if (daysSince <= 1) return 100;
  if (daysSince <= 3) return 80;
  if (daysSince <= 7) return 60;
  if (daysSince <= 14) return 30;
  if (daysSince <= 30) return 10;
  return 0;
}
```

## Gotchas

1. **Too many onboarding steps** -- More than 5 steps in initial onboarding causes drop-off. Show 3-4 core steps, reveal advanced steps after activation. Users who see a 10-step checklist feel overwhelmed and abandon.

2. **Activation milestone too late** -- If your "aha moment" requires 10 minutes of setup, users churn before reaching it. The activation milestone should be achievable in under 3 minutes. Pre-fill data, skip optional steps, reduce friction to zero.

3. **Sending emails to activated users** -- Activation drip emails should stop once the user activates. Sending "you haven't started yet" to someone who's been using the product daily destroys trust. Always check activation status before sending.

4. **Health score without action** -- A health score dashboard that nobody acts on is wasted effort. Each grade (at-risk, churning) must trigger automated actions: email sequences, CSM alerts, in-app messages. Score without intervention is just a vanity metric.

5. **One-size-fits-all onboarding** -- A solo founder and a 50-person team have different needs. Segment onboarding by use case, team size, or role. Ask 1-2 qualifying questions at signup and branch the onboarding flow accordingly.

6. **Ignoring negative signals** -- Support ticket spikes, repeated failed actions, and settings page visits without changes are churn signals. Track these negative events alongside positive ones. A user who visits the cancellation page but doesn't cancel is screaming for help.
