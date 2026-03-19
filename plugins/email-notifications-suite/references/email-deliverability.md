# Email Deliverability Guide

## DNS Records

### SPF (Sender Policy Framework)

Tells receiving servers which IPs can send email from your domain.

```dns
# TXT record on yourdomain.com
v=spf1 include:_spf.google.com include:sendgrid.net include:amazonses.com ~all
```

| Qualifier | Meaning |
|-----------|---------|
| `+` (default) | Pass — authorized sender |
| `~` | SoftFail — probably not authorized (use this for `all`) |
| `-` | HardFail — definitely not authorized |
| `?` | Neutral |

**Common includes:**

| Provider | SPF Include |
|----------|-------------|
| Google Workspace | `include:_spf.google.com` |
| SendGrid | `include:sendgrid.net` |
| AWS SES | `include:amazonses.com` |
| Resend | `include:_spf.resend.com` |
| Mailgun | `include:mailgun.org` |
| Postmark | `include:spf.mtasv.net` |

### DKIM (DomainKeys Identified Mail)

Cryptographic signature proving the email hasn't been tampered with.

```dns
# CNAME or TXT record (provider-specific)
# Example for SendGrid:
s1._domainkey.yourdomain.com  CNAME  s1.domainkey.u12345.wl.sendgrid.net
s2._domainkey.yourdomain.com  CNAME  s2.domainkey.u12345.wl.sendgrid.net

# Example for Resend:
resend._domainkey.yourdomain.com  TXT  "v=DKIM1; k=rsa; p=MIGf..."
```

Each provider gives you specific DNS records to add. Follow their setup guides.

### DMARC (Domain-based Message Authentication)

Policy that tells receivers what to do when SPF/DKIM fail.

```dns
# Start with monitoring (p=none)
_dmarc.yourdomain.com  TXT  "v=DMARC1; p=none; rua=mailto:dmarc-reports@yourdomain.com"

# After monitoring, enforce (p=quarantine or p=reject)
_dmarc.yourdomain.com  TXT  "v=DMARC1; p=quarantine; rua=mailto:dmarc-reports@yourdomain.com; pct=100"
```

| Tag | Meaning |
|-----|---------|
| `p=none` | Monitor only, don't reject anything |
| `p=quarantine` | Send failing messages to spam |
| `p=reject` | Reject failing messages entirely |
| `rua=` | Where to send aggregate reports |
| `ruf=` | Where to send forensic (failure) reports |
| `pct=` | Percentage of messages to apply policy to |

**Recommended rollout:**
1. Start with `p=none` for 2-4 weeks
2. Monitor reports (use dmarcian.com or similar)
3. Move to `p=quarantine; pct=10` (10% of failing)
4. Gradually increase pct to 100
5. Move to `p=reject` when confident

### Return-Path / Envelope From

Use a subdomain for bounce handling:

```dns
bounce.yourdomain.com  MX  10 feedback-smtp.us-east-1.amazonses.com
```

## Deliverability Best Practices

### Content

- **Subject lines:** Avoid ALL CAPS, excessive punctuation (!!!), spam trigger words ("free", "act now", "limited time")
- **From name:** Use a recognizable name, not just an email address
- **Unsubscribe:** One-click unsubscribe header (List-Unsubscribe) + visible link in body
- **HTML + text:** Always include both HTML and plain text versions
- **Image-to-text ratio:** Don't send image-only emails. Keep text-to-image ratio above 60:40
- **Links:** Use your own domain for links (not URL shorteners like bit.ly — they're often blacklisted)

### List Hygiene

| Action | When |
|--------|------|
| Remove hard bounces | Immediately (after first hard bounce) |
| Remove soft bounces | After 3 consecutive soft bounces |
| Remove unsubscribes | Immediately (legally required) |
| Remove inactive users | After 6-12 months of no engagement |
| Verify new signups | Double opt-in (confirmation email) |

### Sending Practices

- **Warm up new domains/IPs:** Start with 50-100 emails/day, increase 50% daily
- **Consistent volume:** Don't go from 100/day to 10,000/day overnight
- **Segment sends:** Don't blast your entire list at once
- **Monitor bounce rates:** > 2% bounce rate is a red flag
- **Monitor spam complaints:** > 0.1% complaint rate triggers ISP scrutiny
- **Send from subdomains:** `mail.yourdomain.com` for transactional, keep `yourdomain.com` clean

### IP Reputation

| Factor | Impact |
|--------|--------|
| Bounce rate | High bounces = poor list quality = lower reputation |
| Spam complaints | Most damaging factor — even 0.1% triggers issues |
| Engagement | Opens and clicks improve reputation |
| Spam trap hits | Hitting a spam trap can blacklist your IP |
| Volume consistency | Sudden spikes look like spam |
| Authentication | SPF/DKIM/DMARC alignment improves trust |

## Testing Deliverability

### Free Tools

| Tool | URL | What It Checks |
|------|-----|----------------|
| **Mail-Tester** | mail-tester.com | SPF, DKIM, DMARC, content, blacklists |
| **MXToolbox** | mxtoolbox.com | DNS records, blacklist check, SMTP diagnostics |
| **Google Postmaster** | postmaster.google.com | Gmail-specific reputation and delivery data |
| **DMARC Analyzer** | dmarcanalyzer.com | DMARC report analysis |

### Check DNS Records

```bash
# SPF
dig TXT yourdomain.com | grep spf

# DKIM
dig TXT selector._domainkey.yourdomain.com

# DMARC
dig TXT _dmarc.yourdomain.com

# MX
dig MX yourdomain.com
```

### Blacklist Check

```bash
# Check if your IP is blacklisted
# Use mxtoolbox.com/blacklists.aspx
# Or programmatically:
dig +short your.ip.reversed.bl.spamcop.net
# Returns 127.0.0.2 if listed
```

## Bounce Handling

### Types

| Type | Meaning | Action |
|------|---------|--------|
| **Hard bounce** | Invalid address, domain doesn't exist | Remove immediately |
| **Soft bounce** | Mailbox full, server down, message too large | Retry 3 times, then remove |
| **Block** | Rejected by recipient server (policy, reputation) | Investigate, fix issue |
| **Complaint** | User clicked "report spam" | Remove immediately, investigate |

### Implementation

```typescript
// Webhook handler for bounce notifications
app.post('/webhooks/email-events', (req, res) => {
  const events = req.body;

  for (const event of events) {
    switch (event.type) {
      case 'bounce':
        if (event.bounceType === 'permanent') {
          // Hard bounce — remove address
          await db.user.update({
            where: { email: event.email },
            data: { emailValid: false, bounceCount: { increment: 1 } },
          });
        } else {
          // Soft bounce — increment counter
          await db.user.update({
            where: { email: event.email },
            data: { bounceCount: { increment: 1 } },
          });
          // Remove after 3 soft bounces
        }
        break;

      case 'complaint':
        // Spam complaint — stop all email
        await db.user.update({
          where: { email: event.email },
          data: { emailOptOut: true, emailValid: false },
        });
        break;

      case 'unsubscribe':
        await db.user.update({
          where: { email: event.email },
          data: { emailOptOut: true },
        });
        break;
    }
  }

  res.json({ received: true });
});
```

## Gmail-Specific Requirements (Feb 2024+)

Google now requires for bulk senders (5,000+ emails/day to Gmail):

1. **SPF and DKIM** authentication on sending domain
2. **DMARC** policy (at minimum `p=none`)
3. **One-click unsubscribe** via `List-Unsubscribe` header
4. **Spam complaint rate** below 0.10% (warning at 0.30%)
5. **Valid forward and reverse DNS** (PTR records)
6. **TLS** for SMTP connections

```typescript
// Add List-Unsubscribe header
await resend.emails.send({
  // ...
  headers: {
    'List-Unsubscribe': '<https://yourdomain.com/unsubscribe?token=xxx>',
    'List-Unsubscribe-Post': 'List-Unsubscribe=One-Click',
  },
});
```

## Monitoring Dashboard Metrics

Track these daily:

| Metric | Healthy Range | Alert Threshold |
|--------|--------------|-----------------|
| Delivery rate | > 97% | < 95% |
| Open rate | 15-25% (varies by industry) | < 10% |
| Click rate | 2-5% | < 1% |
| Bounce rate | < 2% | > 3% |
| Spam complaint rate | < 0.05% | > 0.1% |
| Unsubscribe rate | < 0.5% | > 1% |
