# Email Deliverability Guide

## The Deliverability Stack

```
Sender → Authentication (SPF/DKIM/DMARC) → Reputation → Content → Inbox
```

Every layer must pass for your email to reach the inbox. Failure at any layer = spam folder or rejection.

## Authentication: The Big Three

### 1. SPF (Sender Policy Framework)

**What it does**: Declares which servers are allowed to send email from your domain.

**How to set up**:
```
# DNS TXT record on yourdomain.com
v=spf1 include:_spf.resend.com include:_spf.google.com ~all
```

**Rules**:
- Only ONE SPF record per domain (combine with `include:`)
- `~all` = soft fail (recommended for setup), `-all` = hard fail (strict)
- Max 10 DNS lookups (each `include:` counts as 1)
- If you exceed 10 lookups, flatten your SPF record

**Testing**:
```bash
# Check SPF record
dig TXT yourdomain.com | grep spf

# Online tools
# mxtoolbox.com/spf.aspx
# dmarcian.com/spf-survey/
```

### 2. DKIM (DomainKeys Identified Mail)

**What it does**: Cryptographically signs emails so recipients can verify they weren't tampered with.

**How to set up**: Your email provider generates DKIM keys. You add CNAME or TXT records to DNS.

```
# Example CNAME records (provider gives you these):
s1._domainkey.yourdomain.com → CNAME → s1.domainkey.provider.com
s2._domainkey.yourdomain.com → CNAME → s2.domainkey.provider.com
```

**Testing**:
```bash
# Check DKIM record
dig TXT s1._domainkey.yourdomain.com

# Send a test email and check headers for:
# DKIM-Signature: v=1; a=rsa-sha256; d=yourdomain.com; ...
# Authentication-Results: dkim=pass
```

### 3. DMARC (Domain-based Message Authentication, Reporting, and Conformance)

**What it does**: Tells receiving servers what to do when SPF or DKIM fails, and sends you reports.

**Rollout plan** (don't skip phases):

```
# Phase 1: Monitor (2-4 weeks) — collect data, don't block anything
_dmarc.yourdomain.com  TXT  "v=DMARC1; p=none; rua=mailto:dmarc-reports@yourdomain.com; pct=100"

# Phase 2: Quarantine 10% (2 weeks) — start catching failures
_dmarc.yourdomain.com  TXT  "v=DMARC1; p=quarantine; pct=10; rua=mailto:dmarc-reports@yourdomain.com"

# Phase 3: Quarantine 50% (2 weeks)
_dmarc.yourdomain.com  TXT  "v=DMARC1; p=quarantine; pct=50; rua=mailto:dmarc-reports@yourdomain.com"

# Phase 4: Reject (final) — block all unauthenticated email
_dmarc.yourdomain.com  TXT  "v=DMARC1; p=reject; pct=100; rua=mailto:dmarc-reports@yourdomain.com"
```

**DMARC report analyzers** (free):
- dmarcian.com
- postmarkapp.com/dmarc
- valimail.com/dmarc-monitor

## Sender Reputation

### What Affects Reputation

| Factor | Impact | How to Monitor |
|--------|--------|---------------|
| Bounce rate | High | Keep under 2% |
| Spam complaints | Very high | Keep under 0.1% |
| Spam trap hits | Critical | Clean lists regularly |
| Send volume consistency | Medium | Don't spike suddenly |
| Engagement (opens/clicks) | Medium | Track per campaign |
| Unsubscribe rate | Low-Medium | Keep under 1% |
| Blacklist presence | Critical | Check regularly |

### Reputation Monitoring Tools

| Tool | What It Checks | Free? |
|------|---------------|-------|
| Google Postmaster Tools | Gmail-specific reputation | Yes |
| Microsoft SNDS | Outlook/Hotmail reputation | Yes |
| MXToolbox | Blacklists, DNS, SMTP | Free tier |
| Sender Score (Validity) | Overall reputation score | Free |
| Talos Intelligence | IP reputation | Free |
| BarracudaCentral | Barracuda blacklist | Free |
| mail-tester.com | Send a test, get a score | Free (limited) |

### IP Warming Schedule

When you get a new dedicated IP, you must warm it gradually:

| Day | Daily Volume | Notes |
|-----|-------------|-------|
| 1-2 | 50 | Send to most engaged users |
| 3-4 | 100 | |
| 5-6 | 500 | |
| 7-8 | 1,000 | |
| 9-10 | 5,000 | |
| 11-14 | 10,000 | |
| 15-21 | 25,000 | Monitor bounce/complaint rates |
| 22-28 | 50,000 | |
| 29+ | Target volume | Maintain consistent sending |

**Rules during warming**:
- Only send to confirmed, engaged recipients
- Monitor bounces and complaints daily
- If bounce rate > 5% or complaints > 0.1%, slow down
- Maintain consistent daily volume (don't skip days)

## Content Best Practices

### Subject Lines

**Do**:
- Keep under 50 characters (mobile preview)
- Use the recipient's name when natural
- Be specific: "Your order #1234 shipped" not "Order update"
- Create urgency naturally: "Your trial ends in 3 days"

**Don't**:
- ALL CAPS
- Excessive punctuation!!!
- Spam trigger words: "FREE", "ACT NOW", "LIMITED TIME"
- Misleading subjects (CAN-SPAM violation)
- Re: or Fwd: when it's not a reply/forward

### Email Body

**Do**:
- Keep HTML under 100KB (ideally under 50KB)
- Include both HTML and plain text versions
- Use a reasonable text-to-image ratio (60% text, 40% images)
- Include a visible, working unsubscribe link
- Include your physical mailing address (CAN-SPAM requirement)
- Use `alt` text on all images

**Don't**:
- Use shortened URLs (bit.ly, etc.) — looks spammy
- Include attachments in transactional email
- Use too many links (keeps it under 10)
- Hide text in images (text should be selectable)
- Use JavaScript (stripped by all email clients)
- Use forms in email (limited support, security concerns)

### List Hygiene

```typescript
// Automated list cleaning
async function cleanEmailList() {
  // 1. Remove hard bounces immediately
  await db.user.updateMany({
    where: { emailBounceType: 'hard' },
    data: { emailOptOut: true },
  });

  // 2. Remove users who complained
  await db.user.updateMany({
    where: { emailComplaint: true },
    data: { emailOptOut: true },
  });

  // 3. Flag inactive users (no opens in 90 days)
  const ninetyDaysAgo = new Date(Date.now() - 90 * 86400000);
  const inactive = await db.user.findMany({
    where: {
      lastEmailOpen: { lt: ninetyDaysAgo },
      emailOptOut: false,
    },
  });

  // Send re-engagement campaign, then remove if no response in 30 days

  // 4. Remove role-based addresses
  const roleAddresses = ['admin@', 'info@', 'support@', 'sales@', 'noreply@'];
  // Flag for review, don't auto-remove
}
```

## Troubleshooting Deliverability

### Email Landing in Spam

1. **Check authentication**: SPF, DKIM, DMARC all passing?
   ```bash
   # Send test email, check raw headers for:
   # spf=pass
   # dkim=pass
   # dmarc=pass
   ```

2. **Check blacklists**: Is your IP/domain listed?
   ```
   mxtoolbox.com/blacklists.aspx
   ```

3. **Check content**: Run through spam filters
   ```
   mail-tester.com — send a test email, get a 1-10 score
   ```

4. **Check reputation**: What do mailbox providers think of you?
   ```
   Google Postmaster Tools → gmail.com/postmaster
   ```

5. **Check engagement**: Are people opening and clicking?
   - Low engagement → providers assume spam
   - Solution: re-engage or remove inactive subscribers

### Common Issues

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| All emails bouncing | DNS misconfigured | Check SPF/DKIM records |
| Gmail spam folder | Low engagement, no DMARC | Improve engagement, add DMARC |
| Outlook blocking | IP reputation | Check SNDS, consider dedicated IP |
| Delayed delivery | Rate limiting | Reduce send rate, check queue |
| High bounce rate | Stale list | Clean list, verify addresses |
| "Via" label in Gmail | Missing DKIM or SPF alignment | Align DKIM domain with From domain |

## Legal Requirements

### CAN-SPAM (United States)

- Don't use false or misleading header information
- Don't use deceptive subject lines
- Include your physical postal address
- Tell recipients how to opt out
- Honor opt-out requests within 10 business days
- Monitor what others are doing on your behalf

### GDPR (European Union)

- Explicit consent required for marketing email
- Transactional email doesn't require consent (legitimate interest)
- Must provide easy opt-out mechanism
- Must respond to data access/deletion requests
- Keep records of consent (when, how, what they agreed to)
- Privacy policy must describe email practices

### CASL (Canada)

- Express consent required (implied consent expires after 2 years)
- Must identify sender clearly
- Must include unsubscribe mechanism
- Unsubscribe must be processed within 10 business days

## Metrics to Track

| Metric | Target | Warning | Critical |
|--------|--------|---------|----------|
| Delivery rate | >98% | <95% | <90% |
| Open rate | >20% | <15% | <10% |
| Click rate | >2% | <1% | <0.5% |
| Bounce rate | <2% | >3% | >5% |
| Spam complaint rate | <0.05% | >0.08% | >0.1% |
| Unsubscribe rate | <0.5% | >1% | >2% |
| List growth rate | >2%/mo | 0% | Negative |

## Quick Diagnostic Checklist

```
[ ] SPF record exists and passes
[ ] DKIM record exists and passes
[ ] DMARC record exists (at least p=none)
[ ] Not on any major blacklists
[ ] Bounce rate under 2%
[ ] Complaint rate under 0.1%
[ ] Unsubscribe link visible and working
[ ] Physical address in footer
[ ] Both HTML and plain text versions
[ ] Images have alt text
[ ] Subject line under 50 characters
[ ] No shortened URLs (bit.ly etc.)
[ ] List cleaned in last 90 days
[ ] Sending from verified domain (not gmail/yahoo)
[ ] Consistent sending volume (no spikes)
```
