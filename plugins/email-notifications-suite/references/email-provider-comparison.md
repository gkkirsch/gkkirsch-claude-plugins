# Email Provider Comparison

## Provider Matrix

| Feature | Resend | SendGrid | AWS SES | Postmark | Mailgun |
|---------|--------|----------|---------|----------|---------|
| **Free tier** | 3,000/mo | 100/day | 62,000/mo (12 mo) | 100/mo | 5,000/mo (3 mo) |
| **Paid starting** | $20/mo (50K) | $19.95/mo (50K) | $0.10/1K | $15/mo (10K) | $35/mo (50K) |
| **Best for** | Developers, React Email | High volume, templates | Cost at scale | Transactional only | Flexibility |
| **SDK quality** | Excellent (TS-first) | Good | AWS-style | Good | Good |
| **Template system** | React Email | Dynamic templates | Basic | MJML-based | Handlebars |
| **Deliverability** | High | High | Depends on setup | Highest | High |
| **Analytics** | Basic | Comprehensive | CloudWatch | Detailed | Good |
| **Webhooks** | Yes | Yes | SNS | Yes | Yes |
| **Dedicated IP** | $40/mo | Included (Pro) | $24.95/mo | Included (some) | $59/mo |
| **SMTP relay** | Yes | Yes | Yes | Yes | Yes |
| **API style** | REST (clean) | REST | AWS SDK | REST | REST |

## Decision Guide

### Choose Resend if:
- You're a developer who wants the best DX
- You use React (React Email integration is native)
- You want TypeScript-first SDK
- Your volume is under 100K/month
- You want simple, clean APIs

### Choose SendGrid if:
- You need a visual template editor (non-developers editing templates)
- You need comprehensive analytics and reporting
- You're sending marketing AND transactional email
- You want dynamic templates with Handlebars
- Your volume is 100K-1M+/month

### Choose AWS SES if:
- Cost is the primary concern (cheapest at scale by far)
- You're already on AWS
- You have email expertise to handle deliverability yourself
- You're sending 1M+/month
- You don't need a template system (build your own)

### Choose Postmark if:
- Deliverability is your #1 priority
- You ONLY send transactional email (they reject marketing)
- You want the fastest delivery times
- You need detailed delivery analytics
- You're willing to pay more for reliability

### Choose Mailgun if:
- You need email validation/verification
- You want flexible routing rules
- You need comprehensive logs and analytics
- You want both inbound and outbound email
- Your European users need EU data residency

## DNS Records Quick Reference

### SPF (Sender Policy Framework)

```
# TXT record on your domain
v=spf1 include:_spf.google.com include:sendgrid.net ~all

# Provider-specific includes:
# Resend:   include:amazonses.com
# SendGrid: include:sendgrid.net
# AWS SES:  include:amazonses.com
# Postmark: include:spf.mtasv.net
# Mailgun:  include:mailgun.org
```

### DKIM (DomainKeys Identified Mail)

Each provider gives you CNAME or TXT records to add. Usually 2-3 records:

```
# Example (provider-specific, they give you the values):
s1._domainkey.yourdomain.com → CNAME → s1.domainkey.u1234.wl.sendgrid.net
s2._domainkey.yourdomain.com → CNAME → s2.domainkey.u1234.wl.sendgrid.net
```

### DMARC (Domain-based Message Authentication)

```
# TXT record: _dmarc.yourdomain.com
# Start with monitoring (p=none), then move to quarantine, then reject

# Phase 1: Monitor only (2-4 weeks)
v=DMARC1; p=none; rua=mailto:dmarc@yourdomain.com; pct=100

# Phase 2: Quarantine suspicious (2-4 weeks)
v=DMARC1; p=quarantine; rua=mailto:dmarc@yourdomain.com; pct=25

# Phase 3: Reject failures
v=DMARC1; p=reject; rua=mailto:dmarc@yourdomain.com; pct=100
```

## Cost Comparison at Scale

| Monthly Volume | Resend | SendGrid | AWS SES | Postmark |
|---------------|--------|----------|---------|----------|
| 10,000 | Free | Free | $1 | $15 |
| 50,000 | $20 | $19.95 | $5 | $50 |
| 100,000 | $40 | $19.95 | $10 | $100 |
| 500,000 | $170 | $89.95 | $50 | $375 |
| 1,000,000 | $340 | $89.95 | $100 | $700 |
| 5,000,000 | Custom | $449 | $500 | Custom |

*Prices approximate, check current pricing. AWS SES is consistently the cheapest at volume.*

## Bounce & Complaint Handling

### Webhook Events to Handle

| Event | Action | Priority |
|-------|--------|----------|
| `bounce.hard` | Remove address permanently | Critical |
| `bounce.soft` | Retry 3x, then remove | High |
| `complaint` | Remove address immediately | Critical |
| `unsubscribe` | Update preferences | High |
| `delivered` | Update delivery status | Low |
| `opened` | Track engagement | Low |
| `clicked` | Track engagement | Low |

### Bounce Categories

| Type | Meaning | Action |
|------|---------|--------|
| Hard bounce | Address doesn't exist | Remove immediately |
| Soft bounce | Mailbox full, server down | Retry, remove after 3 failures |
| Block | ISP rejected | Check reputation, fix content |
| Spam complaint | User marked as spam | Remove immediately, investigate |
| Unsubscribe | User clicked unsubscribe | Update preferences |

## Rate Limiting Best Practices

| Provider | Default Rate Limit | Recommendation |
|----------|-------------------|----------------|
| Resend | 10/sec (free), higher on paid | Queue with 8/sec max |
| SendGrid | Varies by plan | Queue with plan-appropriate limits |
| AWS SES | 14/sec (default, request increase) | Queue with 10/sec, increase as needed |
| Postmark | 50/sec | Queue with 40/sec max |
| Mailgun | Varies | Queue with plan-appropriate limits |

Always use a queue (BullMQ, SQS, etc.) with rate limiting. Never send directly from request handlers at scale.
