# A/B Testing Guide

A comprehensive, practitioner-level guide to designing, running, and learning from A/B tests across marketing channels. This is not theory — it is a working playbook with formulas, sample sizes, real test ideas, and decision frameworks.

---

## 1. Hypothesis Formation Framework

Every test starts with a hypothesis. A weak hypothesis leads to inconclusive results regardless of sample size.

**Hypothesis template:**
```
If we [specific change], then [specific metric] will [increase/decrease] by [estimated amount]
because [reason grounded in data, research, or user insight].
```

**Good hypothesis examples:**
- "If we change the CTA button from 'Learn More' to 'Start Free Trial', then the landing page CVR will increase by 15% because the current CTA is vague and does not communicate the free trial offer."
- "If we add customer logos above the fold on the pricing page, then demo request rate will increase by 10% because social proof reduces purchase anxiety for enterprise buyers."
- "If we reduce the email signup form from 5 fields to 2 fields (name + email), then form completion rate will increase by 25% because friction analysis shows 60% of users abandon at field 3."

**Bad hypothesis examples (and why):**
- "If we make the page better, conversions will go up." (Not specific — what changes? How much?)
- "Let's test a new headline." (No predicted outcome or reasoning)
- "Red buttons convert better than green." (No reasoning — this is a guess, not a hypothesis)

**Hypothesis quality checklist:**
- [ ] Identifies one specific, isolated change
- [ ] Names the exact metric being measured
- [ ] Predicts a direction AND magnitude of change
- [ ] Provides a reason grounded in evidence (analytics data, user research, best practices)
- [ ] Is falsifiable — you can clearly determine if it was wrong

---

## 2. Sample Size Calculations

### The Core Formula

For a two-proportion z-test (comparing conversion rates between two variations):

```
n = (Z_alpha/2 + Z_beta)^2 * (p1*(1-p1) + p2*(1-p2)) / (p1 - p2)^2
```

Where:
- `n` = required sample size per variation
- `Z_alpha/2` = z-score for significance level (1.96 for 95% confidence)
- `Z_beta` = z-score for statistical power (0.84 for 80% power)
- `p1` = baseline conversion rate (control)
- `p2` = expected conversion rate (variation) = p1 * (1 + MDE)
- `MDE` = minimum detectable effect (relative change you want to detect)

### Sample Size Lookup Table (95% Confidence, 80% Power)

| Baseline CVR | 5% MDE | 10% MDE | 15% MDE | 20% MDE | 25% MDE | 30% MDE |
|-------------|--------|---------|---------|---------|---------|---------|
| 1% | 636,612 | 159,590 | 71,160 | 40,218 | 25,894 | 18,100 |
| 2% | 312,390 | 78,534 | 35,074 | 19,846 | 12,795 | 8,958 |
| 3% | 204,337 | 51,434 | 23,002 | 13,032 | 8,410 | 5,892 |
| 5% | 118,768 | 29,939 | 13,404 | 7,604 | 4,912 | 3,446 |
| 8% | 71,190 | 17,981 | 8,065 | 4,582 | 2,964 | 2,082 |
| 10% | 55,810 | 14,114 | 6,334 | 3,602 | 2,330 | 1,638 |
| 15% | 35,066 | 8,886 | 3,994 | 2,274 | 1,474 | 1,038 |
| 20% | 24,558 | 6,236 | 2,808 | 1,602 | 1,040 | 734 |
| 30% | 14,392 | 3,670 | 1,658 | 950 | 618 | 438 |

**How to use this table:**
1. Find your current conversion rate in the leftmost column
2. Decide the minimum improvement you want to detect (MDE)
3. The cell value is the number of visitors needed PER VARIATION
4. Multiply by 2 for total traffic needed (control + variation)
5. Divide total traffic needed by your daily traffic to get test duration in days

**Example:**
- Current landing page CVR: 5%
- You want to detect a 20% relative improvement (5% -> 6%)
- Required sample: 7,604 per variation = 15,208 total
- Your daily traffic: 500 visitors
- Test duration: 15,208 / 500 = ~31 days

### For Revenue/Continuous Metrics

When testing average order value, revenue per visitor, or other continuous metrics:

```
n = 2 * (Z_alpha/2 + Z_beta)^2 * sigma^2 / delta^2
```

Where:
- `sigma` = standard deviation of the metric
- `delta` = minimum detectable difference in absolute terms

Tip: Revenue metrics typically have high variance. You will need 2-5x more sample than conversion rate tests. Consider using revenue per visitor (RPV) rather than AOV to reduce variance.

---

## 3. Test Duration Rules

### Minimum Duration

**Hard minimum: 14 days (2 full weeks)**

Rationale:
- Captures weekday vs. weekend behavior differences
- Reduces impact of daily traffic fluctuations
- Avoids novelty effects (new visitors behave differently than returning)
- Accounts for varying user intent by day of week

### Duration Guidelines

| Traffic Level | Simple CVR Test (20% MDE) | Medium CVR Test (10% MDE) | Precise CVR Test (5% MDE) |
|--------------|--------------------------|--------------------------|--------------------------|
| 100/day | 4-8 weeks | 8-16 weeks | Not feasible |
| 500/day | 2-4 weeks | 4-8 weeks | 8-16 weeks |
| 2,000/day | 2 weeks (minimum) | 2-3 weeks | 4-8 weeks |
| 10,000/day | 2 weeks (minimum) | 2 weeks (minimum) | 2-4 weeks |
| 50,000/day | 2 weeks (minimum) | 2 weeks (minimum) | 2 weeks (minimum) |

### Maximum Duration

**Hard maximum: 6 weeks**

After 6 weeks:
- External factors (seasonality, market changes, competitor actions) contaminate results
- Cookie churn means your test populations are no longer clean
- Opportunity cost of not implementing the winner becomes significant
- If you have not reached significance in 6 weeks, your effect size is likely too small to matter

### When to Stop a Test Early

**Stop if:**
- There is a critical bug or UX issue causing user harm
- One variation is performing catastrophically worse (>50% drop in primary metric)
- The test is incorrectly implemented (tracking errors discovered mid-test)

**Never stop early because:**
- One variation looks like it is winning after a few days
- You "have enough data" before reaching your pre-calculated sample size
- A stakeholder is impatient
- Results look "obvious" — this is peeking bias and will inflate your false positive rate to 20-30%

---

## 4. Common Testing Mistakes

### Mistake 1: Peeking at Results

**Problem:** Checking results daily and stopping when you see significance inflates false positive rates from 5% to 20-30%.

**Solution:** Calculate your required sample size before the test. Do not look at statistical significance until you reach that sample size. If you must monitor, use a sequential testing framework (see Section 7).

### Mistake 2: Testing Too Many Things at Once

**Problem:** Changing headline, image, CTA, and layout simultaneously means you cannot attribute the result to any single change.

**Solution:** Change one element at a time. If you must test multiple elements, use a proper multivariate testing design with sufficient traffic (see Section 6).

### Mistake 3: Ignoring Segment Effects

**Problem:** A test shows no overall winner, but variation B is dramatically better for mobile users and worse for desktop users. These effects cancel out in aggregate.

**Solution:** Pre-register 2-3 segments of interest before the test (device type, new vs. returning, traffic source). Analyze these segments post-test, but only treat segment-level results as directional unless they reach significance independently.

### Mistake 4: Wrong Success Metric

**Problem:** You optimize for click-through rate but what you actually care about is revenue. A higher CTR with lower purchase intent produces worse business outcomes.

**Solution:** Always use the metric closest to revenue as your primary metric. Use upstream metrics (CTR, engagement) as secondary/diagnostic metrics only.

### Mistake 5: Survivorship Bias in Test Selection

**Problem:** Only testing "safe" changes means you only find small improvements. Your testing program stagnates.

**Solution:** Use the 70/20/10 rule for your testing portfolio:
- 70% incremental tests (button colors, copy tweaks, layout adjustments) — low risk, low reward
- 20% moderate tests (new page sections, different value props, pricing changes) — medium risk, medium reward
- 10% radical tests (completely new landing pages, different offers, new funnels) — high risk, high potential reward

### Mistake 6: Not Accounting for Multiple Comparisons

**Problem:** Testing 5 variations against a control. With alpha=0.05 and 5 comparisons, the probability of at least one false positive is 1 - (1-0.05)^5 = 22.6%.

**Solution:** Apply Bonferroni correction: divide alpha by the number of comparisons. For 5 variations: use alpha = 0.05/5 = 0.01 per comparison. Better yet, limit tests to A/B (2 variations) whenever possible.

### Mistake 7: Ignoring Practical Significance

**Problem:** A test reaches statistical significance with a 0.3% improvement in conversion rate. This is statistically real but practically meaningless.

**Solution:** Define your minimum detectable effect before the test based on business impact. If a 0.3% improvement does not justify the implementation cost, set your MDE higher and require a larger effect to declare a winner.

---

## 5. What to Test (By Channel)

### Website / Landing Pages
- Headline (value proposition framing)
- Sub-headline (supporting statement)
- Hero image (product shot vs. lifestyle vs. illustration)
- CTA button text ("Get Started" vs. "Start Free Trial" vs. "See Pricing")
- CTA button color and size
- Form length (fewer fields vs. more qualified leads)
- Social proof placement (above fold vs. below)
- Pricing page layout (comparison table vs. cards)
- Navigation presence (with nav vs. no nav on landing pages)
- Page length (short-form vs. long-form)

### Email
- Subject line (question vs. statement vs. number)
- Preview text
- Send time (morning vs. afternoon, weekday vs. weekend)
- Sender name (person vs. company)
- Email length (short vs. long)
- Number of CTAs (single CTA vs. multiple)
- Personalization depth (name only vs. behavioral)
- Plain text vs. HTML design
- Image placement (above vs. below the fold)
- PS line (with vs. without)

### Paid Ads
- Ad headline variations
- Ad description text
- Display URL paths
- Ad extensions combinations
- Bidding strategy (manual CPC vs. target CPA vs. maximize conversions)
- Audience targeting (broad vs. narrow)
- Creative format (single image vs. carousel vs. video)
- Landing page destination (homepage vs. dedicated LP)
- Ad placement (feed vs. stories vs. reels)
- Offer type (discount vs. free trial vs. free shipping)

---

## 6. Multivariate Testing

### When to Use Multivariate Testing

Use MVT when:
- You have very high traffic (>50,000 visitors per week to the test page)
- You need to understand interaction effects between elements
- You are optimizing a stable, mature page where incremental gains matter

Do NOT use MVT when:
- Traffic is moderate or low — you will never reach significance
- You are in early optimization stages — A/B tests will find bigger wins faster
- You have more than 3 elements to test — the number of combinations grows exponentially

### Factorial Design Example

Testing 2 headlines x 2 images x 2 CTAs = 8 combinations

| Combination | Headline | Image | CTA |
|------------|----------|-------|-----|
| 1 | A | A | A |
| 2 | A | A | B |
| 3 | A | B | A |
| 4 | A | B | B |
| 5 | B | A | A |
| 6 | B | A | B |
| 7 | B | B | A |
| 8 | B | B | B |

Traffic needed: 8x the per-variation sample size from the lookup table.

For a 5% baseline CVR and 20% MDE: 7,604 x 8 = 60,832 total visitors minimum.

### Fractional Factorial Design

If full factorial requires too much traffic, use a fractional factorial design that tests a subset of combinations. This can identify main effects but may miss interaction effects. Consult a statistician for designs with >3 factors.

---

## 7. Sequential Testing vs. Simultaneous

### Simultaneous Testing (Standard A/B)

- All variations run at the same time
- Random assignment of users to variations
- Eliminates time-based confounds (day of week, seasonality)
- This is the default and recommended approach

### Sequential Testing (Always Valid Inference)

When you need to monitor results continuously without inflating false positives:

- Use a spending function (O'Brien-Fleming or Pocock bounds) that adjusts the significance threshold at each peek
- O'Brien-Fleming: Very conservative at early looks, approaches standard alpha at final look
- Pocock: Equal spending at each look (more power early, less at the end)

**O'Brien-Fleming approximate bounds (5 equally spaced looks, overall alpha=0.05):**

| Look | Information Fraction | Significance Threshold |
|------|---------------------|----------------------|
| 1 | 20% | p < 0.00005 |
| 2 | 40% | p < 0.004 |
| 3 | 60% | p < 0.014 |
| 4 | 80% | p < 0.029 |
| 5 | 100% | p < 0.041 |

This approach lets you stop early for very large effects while maintaining overall error rates.

---

## 8. Bayesian vs. Frequentist Approaches

### Frequentist (Traditional A/B Testing)

- **Question answered:** "If there is no real difference, what is the probability of seeing results this extreme?"
- **Output:** p-value and confidence interval
- **Decision rule:** If p < 0.05, reject the null hypothesis
- **Pros:** Well-understood, widely accepted, easy to calculate sample sizes
- **Cons:** Requires fixed sample size, does not give "probability B is better than A"

### Bayesian A/B Testing

- **Question answered:** "What is the probability that B is better than A, given the data we have observed?"
- **Output:** Probability of each variation being the best, expected loss
- **Decision rule:** If P(B > A) > 95% AND expected loss < threshold, declare B the winner
- **Pros:** Natural interpretation ("there is a 97% chance B is better"), flexible stopping rules, works well with small samples
- **Cons:** Requires choosing priors, less standardized, can be harder to explain to stakeholders

### When to Use Which

| Scenario | Recommended Approach |
|----------|---------------------|
| Standard website test, good traffic | Frequentist |
| Low traffic, need faster decisions | Bayesian |
| Regulatory or academic context | Frequentist |
| Continuous experimentation platform | Bayesian |
| Stakeholders want "probability of winning" | Bayesian |
| You need to peek at results regularly | Bayesian or Sequential Frequentist |

---

## 9. Twenty Real A/B Test Ideas with Expected Impact

| # | Test | Element | Expected Impact | Effort |
|---|------|---------|----------------|--------|
| 1 | Replace stock photo hero with product screenshot | Landing page image | +10-30% CVR | Low |
| 2 | Add specific number to headline ("Join 10,000+ teams") | Headline | +5-15% CVR | Low |
| 3 | Reduce form fields from 5 to 3 | Lead gen form | +15-30% form completion | Low |
| 4 | Add live chat widget to pricing page | Pricing page | +5-15% demo requests | Medium |
| 5 | Test annual vs. monthly default pricing toggle | Pricing page | +10-20% annual plan selection | Low |
| 6 | Add customer testimonial video above fold | Landing page | +10-25% CVR | Medium |
| 7 | Change CTA from "Submit" to action-specific verb | Any form | +5-15% form submission | Low |
| 8 | Send emails from a person name vs. company name | Email | +10-25% open rate | Low |
| 9 | Test short subject line (<5 words) vs. long (>10 words) | Email | +5-20% open rate | Low |
| 10 | Add exit-intent popup with discount offer | E-commerce | +3-8% overall CVR | Medium |
| 11 | Test urgency element ("Only 3 left") vs. no urgency | Product page | +5-15% add-to-cart | Low |
| 12 | Simplify navigation to 4 items vs. 8 items | Site-wide | +3-10% goal completion | Medium |
| 13 | Add free shipping threshold banner ("Free shipping over $50") | E-commerce | +5-15% AOV | Low |
| 14 | Test video ad vs. static image ad | Facebook Ads | +10-40% CTR | High |
| 15 | Move social proof (star ratings) into ad creative | Facebook/Google Ads | +5-15% CTR | Medium |
| 16 | Test long-form sales page vs. short-form | Landing page | +/-20% CVR (varies) | High |
| 17 | Add money-back guarantee badge near CTA | Checkout/pricing | +5-10% CVR | Low |
| 18 | Test Tuesday 10am vs. Thursday 2pm email send | Email | +5-15% open rate | Low |
| 19 | Replace bullet points with benefit-oriented icons | Landing page | +3-8% engagement | Medium |
| 20 | Test in-line form vs. popup form for lead capture | Blog/content | +10-30% conversion | Medium |

---

## 10. Test Documentation Template

After every test, document the following:

```
## Test Name: [Descriptive Name]
## Test ID: [YYYY-MM-###]
## Date: [Start Date] - [End Date]

### Hypothesis
If we [change], then [metric] will [direction] by [amount] because [reason].

### Test Design
- Type: A/B / A/B/n / Multivariate
- Primary metric: [metric name]
- Secondary metrics: [metric 1], [metric 2]
- Traffic allocation: [50/50, 80/20, etc.]
- Targeting: [All visitors / Segment]
- Minimum sample size: [X per variation]
- Expected duration: [X days]

### Variations
- Control (A): [Description]
- Variation (B): [Description]
- [Additional variations if applicable]

### Results
- Sample size: Control [X], Variation [X]
- Primary metric: Control [X%], Variation [X%]
- Relative lift: [+/- X%]
- Statistical significance: [p-value] / [confidence level]
- Power: [X%]

### Decision
- Winner: [Control / Variation] / Inconclusive
- Action: [Implement / Iterate / Abandon]
- Implementation date: [Date]

### Learnings
- [Key insight 1]
- [Key insight 2]
- [What to test next based on this result]
```

---

## 11. ICE Prioritization Framework

Score every test idea on three dimensions from 1-10:

### Impact (1-10)
How much will this move your primary business metric if the test wins?
- 1-3: Minor improvements (<5% lift on a secondary metric)
- 4-6: Moderate improvements (5-15% lift on primary metric)
- 7-8: Significant improvements (15-30% lift or affects high-traffic pages)
- 9-10: Transformative (>30% lift or fundamentally changes conversion flow)

### Confidence (1-10)
How confident are you that this test will produce a positive result?
- 1-3: Pure speculation, no supporting data
- 4-6: Some directional evidence (competitor analysis, best practices)
- 7-8: Strong evidence (user research, analytics data, past test results)
- 9-10: Near-certain (proven in similar contexts, fixing a known bug)

### Ease (1-10)
How easy is it to implement and run this test?
- 1-3: Requires engineering, multiple teams, weeks of work
- 4-6: Requires design + development, can be done in 1-2 weeks
- 7-8: Can be done with a testing tool (no code), 1-3 days
- 9-10: Copy change, can be live in hours

### Scoring and Prioritization

```
ICE Score = (Impact + Confidence + Ease) / 3
```

| Score Range | Priority | Action |
|-------------|----------|--------|
| 8.0-10.0 | Run immediately | This is a no-brainer, launch ASAP |
| 6.0-7.9 | High priority | Schedule for next testing slot |
| 4.0-5.9 | Medium priority | Add to backlog, run when capacity allows |
| 2.0-3.9 | Low priority | Consider deprioritizing or refining the hypothesis |
| 1.0-1.9 | Skip | Not worth the effort right now |

**Example prioritization:**

| Test Idea | Impact | Confidence | Ease | ICE Score | Priority |
|-----------|--------|-----------|------|-----------|----------|
| Shorten signup form | 7 | 8 | 9 | 8.0 | Run immediately |
| New hero image | 6 | 5 | 8 | 6.3 | High priority |
| Redesign pricing page | 9 | 6 | 3 | 6.0 | Medium (high effort) |
| Change footer links | 2 | 3 | 9 | 4.7 | Low priority |
| New checkout flow | 9 | 7 | 2 | 6.0 | Medium (high effort) |

Maintain a testing backlog sorted by ICE score. Review and re-score quarterly as priorities shift.
