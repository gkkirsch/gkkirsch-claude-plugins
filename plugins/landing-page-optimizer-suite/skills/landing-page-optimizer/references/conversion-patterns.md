# Conversion Patterns Library

50+ proven conversion patterns organized by landing page section. Each pattern includes a description,
when to use it, conversion psychology, and implementation code examples in React/Tailwind.

---

## Hero Section Patterns

### Pattern 1: Direct Promise Hero

**"Get [Result] in [Timeframe]"**

The most straightforward and often highest-converting hero pattern. Works when you have a clear,
measurable outcome to promise.

**When to use**: When you can make a specific, believable promise about results.
**Conversion psychology**: Specificity builds credibility. A timeframe creates urgency and sets expectations.

```tsx
function DirectPromiseHero() {
  return (
    <section className="relative overflow-hidden bg-white px-6 py-24 sm:py-32 lg:px-8">
      <div className="mx-auto max-w-2xl text-center">
        <h1 className="text-4xl font-bold tracking-tight text-gray-900 sm:text-6xl">
          Cut Your Reporting Time by 80% in the First Week
        </h1>
        <p className="mt-6 text-lg leading-8 text-gray-600">
          Automated analytics that turn raw data into executive-ready reports.
          No spreadsheets. No manual work. Just insights that drive decisions.
        </p>
        <div className="mt-10 flex items-center justify-center gap-x-6">
          <a href="#" className="rounded-lg bg-indigo-600 px-6 py-3.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 transition-colors">
            Start my free trial
          </a>
        </div>
        <p className="mt-4 text-sm text-gray-500">No credit card required. Free for 14 days.</p>
        <p className="mt-6 text-sm font-medium text-gray-700">
          Trusted by 10,000+ marketing teams worldwide
        </p>
      </div>
    </section>
  );
}
```

### Pattern 2: Social Proof Hero

**"Join [X]+ [audience] who [achieve result]"**

Leverages the bandwagon effect. Effective for products with significant traction.

**When to use**: When you have impressive user numbers or recognizable customer logos.
**Conversion psychology**: Social proof reduces anxiety. "If 50,000 others trust this, it must be good."

```tsx
function SocialProofHero() {
  return (
    <section className="bg-white px-6 py-24 sm:py-32 lg:px-8">
      <div className="mx-auto max-w-2xl text-center">
        <div className="mb-8 inline-flex items-center gap-2 rounded-full bg-green-50 px-4 py-2 text-sm text-green-700">
          <span className="h-2 w-2 rounded-full bg-green-500" />
          50,847 teams already using DataFlow
        </div>
        <h1 className="text-4xl font-bold tracking-tight text-gray-900 sm:text-6xl">
          Join 50,000+ Marketing Teams Who Report in Minutes, Not Hours
        </h1>
        <p className="mt-6 text-lg leading-8 text-gray-600">
          The analytics platform that Fortune 500 companies and startups
          both love. See why teams switch from spreadsheets and never look back.
        </p>
        <div className="mt-10 flex items-center justify-center gap-x-6">
          <a href="#" className="rounded-lg bg-indigo-600 px-6 py-3.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500">
            See it in action
          </a>
          <a href="#" className="text-sm font-semibold leading-6 text-gray-900">
            Read customer stories <span aria-hidden="true">&rarr;</span>
          </a>
        </div>
      </div>
      {/* Logo bar */}
      <div className="mx-auto mt-16 max-w-5xl">
        <p className="text-center text-sm font-medium text-gray-500 mb-8">
          Trusted by teams at
        </p>
        <div className="flex items-center justify-center gap-x-12 grayscale opacity-60">
          {/* Company logos here */}
        </div>
      </div>
    </section>
  );
}
```

### Pattern 3: Video Hero

**Auto-play background video with overlay CTA**

Captures attention with motion. Best for products that are visual or experiential.

**When to use**: When your product has a compelling visual demo or the outcome is visual.
**Conversion psychology**: Video increases time-on-page and engagement. Seeing the product in action
reduces uncertainty about what it does.

Implementation note: Use a short (15-30s) looping video. Mute by default. Ensure a fallback image
for slow connections. Overlay text must be readable against all frames.

### Pattern 4: Interactive Hero

**Calculator, quiz, or embedded tool**

Turns the hero into an engagement device. The visitor starts using your product before signing up.

**When to use**: When you can give value before asking for anything. ROI calculators, savings
estimators, quizzes, or mini-versions of your tool.
**Conversion psychology**: The endowment effect — once someone invests time and sees personalized results,
they feel ownership and are more likely to convert.

```tsx
function InteractiveHero() {
  return (
    <section className="bg-gradient-to-b from-gray-50 to-white px-6 py-24 lg:px-8">
      <div className="mx-auto max-w-4xl">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl">
            How Much Time Are You Wasting on Reports?
          </h1>
          <p className="mt-4 text-lg text-gray-600">
            Find out in 30 seconds. Enter your numbers below.
          </p>
        </div>
        <div className="rounded-2xl border border-gray-200 bg-white p-8 shadow-lg">
          {/* Interactive calculator component */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Hours spent on reporting per week
              </label>
              <input
                type="number"
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="e.g., 10"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Team members involved
              </label>
              <input
                type="number"
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="e.g., 5"
              />
            </div>
          </div>
          <div className="mt-8 rounded-lg bg-indigo-50 p-6 text-center">
            <p className="text-sm text-indigo-600 font-medium">Your estimated savings</p>
            <p className="text-4xl font-bold text-indigo-900 mt-2">40 hours/month</p>
            <p className="text-sm text-indigo-700 mt-1">That is $8,000/month in recovered productivity</p>
          </div>
          <div className="mt-6 text-center">
            <a href="#" className="rounded-lg bg-indigo-600 px-8 py-3 text-sm font-semibold text-white hover:bg-indigo-500">
              Start saving time today
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
```

### Pattern 5: Before/After Hero

**Visual comparison showing the transformation**

Powerful for products with a visible outcome. Slider comparisons, side-by-side layouts, or
state transitions.

**When to use**: When the before/after difference is dramatic and visual.
**Conversion psychology**: The contrast principle — placing the old way next to the new way makes
the improvement feel more significant.

### Pattern 6: Testimonial Hero

**Customer quote as the headline**

Lets a real customer make your case. The most credible form of headline.

**When to use**: When you have a customer testimonial that is more compelling than anything you
could write yourself.
**Conversion psychology**: Third-party validation is more trusted than first-party claims. A customer
saying "This saved us $200K" is more believable than the company saying "Save $200K."

### Pattern 7: Data Hero

**"$2.3M saved by our customers" or a live metric counter**

Leads with an aggregate result that is too large to ignore.

**When to use**: When you have impressive aggregate numbers to share.
**Conversion psychology**: Large numbers create an anchoring effect and signal widespread adoption.

### Pattern 8: Problem Statement Hero

**"Tired of [pain]?" or "Still [doing painful thing]?"**

Starts with the problem to create immediate resonance with visitors experiencing that pain.

**When to use**: When the audience is problem-aware but not yet solution-aware.
**Conversion psychology**: Starting with the problem triggers recognition ("That is me!") and
primes the visitor to want a solution.

### Pattern 9: Question Hero

**"What if you could [dream state]?"**

Opens with a question that plants the seed of possibility.

**When to use**: When the desired outcome feels aspirational or transformative.
**Conversion psychology**: Questions engage the brain actively. The reader cannot help but
mentally answer, which creates buy-in before any claim is made.

### Pattern 10: Minimalist Hero

**Product screenshot + single CTA, minimal text**

For products where the interface speaks for itself. Clean, confident, no fluff.

**When to use**: When your product has a beautiful interface and your audience is tech-savvy.
**Conversion psychology**: Confidence in simplicity. If a product can sell itself with one screenshot,
it signals quality and ease of use.

```tsx
function MinimalistHero() {
  return (
    <section className="bg-white px-6 py-16 lg:px-8">
      <div className="mx-auto max-w-4xl text-center">
        <h1 className="text-3xl font-semibold tracking-tight text-gray-900 sm:text-4xl">
          Analytics that make sense.
        </h1>
        <div className="mt-8">
          <a href="#" className="rounded-lg bg-gray-900 px-6 py-3 text-sm font-semibold text-white hover:bg-gray-800">
            Try it free
          </a>
        </div>
        <div className="mt-12 rounded-xl border border-gray-200 shadow-2xl overflow-hidden">
          <img
            src="/dashboard-screenshot.png"
            alt="DataFlow dashboard showing real-time analytics"
            className="w-full"
          />
        </div>
      </div>
    </section>
  );
}
```

---

## Social Proof Patterns

### Pattern 11: Logo Bar

A horizontal row of recognizable company logos. Grayscale with hover color.

**When to use**: When you have well-known customers. Even 4-5 recognizable logos create instant credibility.
**Placement**: Immediately below the hero section, or within the hero section itself.

```tsx
function LogoBar() {
  const logos = [
    { name: "Stripe", src: "/logos/stripe.svg" },
    { name: "Shopify", src: "/logos/shopify.svg" },
    { name: "Notion", src: "/logos/notion.svg" },
    { name: "Vercel", src: "/logos/vercel.svg" },
    { name: "Linear", src: "/logos/linear.svg" },
  ];

  return (
    <div className="bg-gray-50 py-12">
      <div className="mx-auto max-w-5xl px-6 lg:px-8">
        <p className="text-center text-sm font-medium text-gray-500 mb-8">
          Trusted by industry leaders
        </p>
        <div className="flex flex-wrap items-center justify-center gap-x-12 gap-y-6">
          {logos.map((logo) => (
            <img
              key={logo.name}
              src={logo.src}
              alt={logo.name}
              className="h-8 grayscale opacity-60 hover:grayscale-0 hover:opacity-100 transition-all"
            />
          ))}
        </div>
      </div>
    </div>
  );
}
```

### Pattern 12: Testimonial Cards

Individual customer testimonials with photo, name, title, and company.

**When to use**: When you have 3+ specific testimonials with measurable results.
**Key elements**: Real photo, full name, job title, company name, and a specific result ("increased conversions by 47%").

```tsx
function TestimonialCards() {
  const testimonials = [
    {
      quote: "We went from spending 20 hours a week on reporting to 2 hours. The ROI was immediate.",
      name: "Sarah Chen",
      title: "VP of Marketing",
      company: "TechCorp",
      avatar: "/avatars/sarah.jpg",
      metric: "10x faster reporting",
    },
    {
      quote: "DataFlow gave us visibility we never had before. We caught a $200K budget leak in the first month.",
      name: "Marcus Johnson",
      title: "CMO",
      company: "GrowthCo",
      avatar: "/avatars/marcus.jpg",
      metric: "$200K saved",
    },
    {
      quote: "Best analytics tool we have ever used. Setup took 15 minutes and we had insights the same day.",
      name: "Emily Rodriguez",
      title: "Director of Growth",
      company: "ScaleUp Inc",
      avatar: "/avatars/emily.jpg",
      metric: "15-minute setup",
    },
  ];

  return (
    <section className="bg-white py-24 px-6 lg:px-8">
      <div className="mx-auto max-w-6xl">
        <h2 className="text-center text-3xl font-bold text-gray-900 mb-16">
          Hear from teams like yours
        </h2>
        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          {testimonials.map((t) => (
            <div key={t.name} className="rounded-2xl border border-gray-100 bg-gray-50 p-8">
              <div className="mb-4 inline-block rounded-full bg-indigo-100 px-3 py-1 text-sm font-medium text-indigo-700">
                {t.metric}
              </div>
              <p className="text-gray-700 leading-relaxed">&ldquo;{t.quote}&rdquo;</p>
              <div className="mt-6 flex items-center gap-3">
                <img src={t.avatar} alt={t.name} className="h-10 w-10 rounded-full" />
                <div>
                  <p className="text-sm font-semibold text-gray-900">{t.name}</p>
                  <p className="text-sm text-gray-500">{t.title}, {t.company}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
```

### Pattern 13: Metric Counters

Large, animated numbers that display aggregate social proof.

**When to use**: When your numbers are impressive at scale (users, revenue generated, time saved).
**Key elements**: Use specific numbers (not rounded). Animate counting up on scroll into view.

### Pattern 14: Case Study Snippets

Short summaries (3-5 sentences) of customer success stories with a "Read full story" link.

**When to use**: For B2B and high-consideration purchases where prospects need detailed proof.
**Key elements**: Company name, challenge, result with metrics, time to value.

### Pattern 15: Star Ratings and Reviews

Display aggregate ratings from third-party platforms (G2, Capterra, Product Hunt, Trustpilot).

**When to use**: When you have strong ratings (4.5+ stars) on recognized platforms.
**Key elements**: Star visualization, numeric rating, total number of reviews, platform badge.

### Pattern 16: Media Mentions

"As seen in" bar with publication logos (TechCrunch, Forbes, Wired, etc.).

**When to use**: When you have been featured in recognizable publications.
**Key elements**: Publication logos, optional quote snippets.

### Pattern 17: Live User Count

Real-time or near-real-time display of users currently online or recently active.

**When to use**: For products with high concurrent usage. Creates FOMO and signals popularity.
**Key elements**: Animated dot, count that updates, "people using this right now" language.

### Pattern 18: Community Size

Showcase the size and activity of your user community (Slack, Discord, GitHub, forum).

**When to use**: For developer tools and products with active communities.
**Key elements**: Community platform logo, member count, recent activity metrics.

---

## CTA Patterns

### Pattern 19: Benefit-Driven CTA

Button copy that states the benefit of clicking, not the action.

**Examples**: "Get more leads", "Start saving time", "See my results"
**Why it works**: Focuses on what the visitor gets, not what they have to do.

### Pattern 20: Urgency CTA

CTA with time-limited or quantity-limited element.

**Examples**: "Claim your spot (only 12 left)", "Start free trial — ends Friday"
**Why it works**: Loss aversion — people are more motivated by fear of losing than desire to gain.
**Caution**: Must be genuine. Fake urgency destroys trust.

### Pattern 21: Risk-Reversal CTA

CTA paired with a guarantee that eliminates risk.

```tsx
function RiskReversalCTA() {
  return (
    <div className="text-center">
      <a href="#" className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-8 py-4 text-base font-semibold text-white hover:bg-indigo-500 transition-colors">
        Start my free trial
      </a>
      <div className="mt-4 flex items-center justify-center gap-4 text-sm text-gray-500">
        <span className="flex items-center gap-1">
          <svg className="h-4 w-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
          </svg>
          No credit card required
        </span>
        <span className="flex items-center gap-1">
          <svg className="h-4 w-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
          </svg>
          Cancel anytime
        </span>
        <span className="flex items-center gap-1">
          <svg className="h-4 w-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
          </svg>
          30-day money-back guarantee
        </span>
      </div>
    </div>
  );
}
```

### Pattern 22: Micro-Commitment CTA

A small, non-threatening first step that leads to the full conversion.

**Examples**: "See a demo" (instead of "Buy now"), "Take the quiz" (instead of "Sign up")
**Why it works**: Lower perceived commitment reduces friction. Once someone takes a small step,
consistency bias makes them more likely to continue.

### Pattern 23: Sticky Header/Footer CTA

A CTA that remains visible as the visitor scrolls.

**When to use**: On long-form landing pages where the primary CTA scrolls out of view.
**Implementation**: Fixed position bar at the top or bottom of the viewport, appearing after
the user scrolls past the hero CTA.

```tsx
function StickyFooterCTA() {
  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 border-t border-gray-200 bg-white/95 backdrop-blur-sm px-6 py-3">
      <div className="mx-auto flex max-w-5xl items-center justify-between">
        <div>
          <p className="text-sm font-semibold text-gray-900">Ready to get started?</p>
          <p className="text-xs text-gray-500">Free 14-day trial. No credit card required.</p>
        </div>
        <a href="#" className="rounded-lg bg-indigo-600 px-6 py-2.5 text-sm font-semibold text-white hover:bg-indigo-500">
          Start free trial
        </a>
      </div>
    </div>
  );
}
```

### Pattern 24: Exit Intent CTA

A modal that appears when the visitor moves their cursor toward the browser close/back button.

**When to use**: As a last-chance conversion opportunity. Do not use on mobile (no cursor intent).
**Content options**: Discount offer, lead magnet, testimonial, or a survey question.

### Pattern 25: Progressive Disclosure CTA

Multi-step CTA where the first click reveals more options or a short form.

**When to use**: When the conversion requires information but you want a low-friction first click.
**Example**: Button says "Get my custom quote" -> clicks -> reveals 3-field form inline.

### Pattern 26: Social CTA

CTA that shows how many others have taken the same action.

**Examples**: "Join 12,847 marketers — Start free trial", "432 people signed up today"
**Why it works**: Real-time social proof at the moment of decision reduces anxiety.

---

## Trust Patterns

### Pattern 27: Money-Back Guarantee Badge

Visual badge displaying the guarantee period and policy.

**When to use**: For any paid product. The longer the guarantee, the lower the anxiety.
**Key elements**: Shield icon, guarantee duration, "no questions asked" language.

### Pattern 28: Security Certification Badges

SSL, SOC 2, GDPR, HIPAA, or PCI compliance badges near forms and payment sections.

**When to use**: Whenever collecting sensitive data (payment info, personal data, health data).
**Key elements**: Recognized certification logos, not custom-made badges.

### Pattern 29: Payment Method Badges

Visa, Mastercard, PayPal, Apple Pay logos near pricing and checkout.

**When to use**: On any page with payment. Familiar payment logos reduce transaction anxiety.

### Pattern 30: Industry Certifications

ISO, industry-specific certifications, awards, and recognitions.

**When to use**: In industries where certification matters (healthcare, finance, education).

### Pattern 31: Privacy Commitment

Explicit statement about data handling, linked to full privacy policy.

**When to use**: Near every form. "We never share your data" or "Your email is safe with us."

### Pattern 32: Refund Policy Preview

Clear, visible refund policy summary near the CTA.

**When to use**: For direct sales. A clear refund policy actually increases sales.

### Pattern 33: Customer Support Access

Display of support channels and response times.

**When to use**: For products where post-purchase support is a concern.
**Examples**: "Live chat support", "Average response time: 2 hours", "Dedicated account manager"

### Pattern 34: Free Trial / Freemium

Offering a risk-free way to experience the product.

**When to use**: For software products. Free trials convert 15-25% to paid on average.
**Key elements**: Duration, what is included, no credit card required (if applicable).

---

## Form Patterns

### Pattern 35: Single Field Form

The most frictionless form — just an email address.

**When to use**: For newsletter signups, early access, and top-of-funnel lead generation.
**Conversion rate**: Typically 3-5x higher than multi-field forms.

```tsx
function SingleFieldForm() {
  return (
    <form className="flex gap-2 max-w-md mx-auto">
      <input
        type="email"
        placeholder="Enter your work email"
        className="flex-1 rounded-lg border border-gray-300 px-4 py-3 text-sm focus:border-indigo-500 focus:ring-indigo-500"
      />
      <button type="submit" className="rounded-lg bg-indigo-600 px-6 py-3 text-sm font-semibold text-white hover:bg-indigo-500 whitespace-nowrap">
        Get started free
      </button>
    </form>
  );
}
```

### Pattern 36: Multi-Step Form

Breaking a long form into 2-4 steps with a progress indicator.

**When to use**: When you need 5+ fields. Multi-step forms convert 86% better than single-step forms with the same number of fields.
**Key elements**: Progress bar, step labels, ability to go back, save progress.

### Pattern 37: Inline Form

Form fields embedded within the page content rather than in a separate section.

**When to use**: When the form is short (1-3 fields) and contextually relevant to surrounding content.

### Pattern 38: Conversational Form

Form that feels like a chat or interview, asking one question at a time.

**When to use**: For complex intake forms (insurance quotes, custom product configuration).
**Conversion rate**: Up to 40% higher engagement than traditional forms.

### Pattern 39: Quiz/Assessment Form

Interactive quiz that provides personalized results in exchange for contact info.

**When to use**: When you can segment prospects and provide tailored value.
**Key elements**: Engaging questions, progress bar, valuable results, email gate on results.

### Pattern 40: ROI Calculator Form

Calculator that demonstrates value and captures lead information.

**When to use**: For B2B products where ROI is a key buying criterion.
**Key elements**: Relevant input fields, visual results, comparison to current state.

### Pattern 41: Configurator Form

Product configuration tool that lets prospects build their custom solution.

**When to use**: For products with variable pricing or features.
**Key elements**: Visual configuration, real-time price updates, clear specifications.

### Pattern 42: Lead Magnet Form

Form offering a valuable resource (ebook, template, checklist) in exchange for contact info.

**When to use**: For top-of-funnel lead generation when the visitor is not ready to buy.
**Key elements**: Resource preview image, clear description of value, instant delivery.

---

## Pricing Patterns

### Pattern 43: Comparison Table

Side-by-side feature comparison of pricing tiers.

**When to use**: When you have 2-4 pricing tiers with clear feature differentiation.
**Key elements**: Highlighted "most popular" tier, annual/monthly toggle, feature checkmarks.

```tsx
function PricingComparison() {
  return (
    <section className="bg-white py-24 px-6 lg:px-8">
      <div className="mx-auto max-w-5xl">
        <h2 className="text-center text-3xl font-bold text-gray-900">
          Simple, transparent pricing
        </h2>
        <p className="mt-4 text-center text-lg text-gray-600">
          No hidden fees. No surprises. Cancel anytime.
        </p>
        <div className="mt-16 grid grid-cols-1 gap-8 md:grid-cols-3">
          {/* Starter */}
          <div className="rounded-2xl border border-gray-200 p-8">
            <h3 className="text-lg font-semibold text-gray-900">Starter</h3>
            <p className="mt-2 text-sm text-gray-500">For individuals and small teams</p>
            <p className="mt-6"><span className="text-4xl font-bold text-gray-900">$29</span><span className="text-gray-500">/month</span></p>
            <a href="#" className="mt-8 block rounded-lg border border-indigo-600 px-4 py-2.5 text-center text-sm font-semibold text-indigo-600 hover:bg-indigo-50">
              Start free trial
            </a>
          </div>
          {/* Pro - Highlighted */}
          <div className="relative rounded-2xl border-2 border-indigo-600 p-8">
            <div className="absolute -top-4 left-1/2 -translate-x-1/2 rounded-full bg-indigo-600 px-4 py-1 text-xs font-semibold text-white">
              Most popular
            </div>
            <h3 className="text-lg font-semibold text-gray-900">Pro</h3>
            <p className="mt-2 text-sm text-gray-500">For growing teams</p>
            <p className="mt-6"><span className="text-4xl font-bold text-gray-900">$79</span><span className="text-gray-500">/month</span></p>
            <a href="#" className="mt-8 block rounded-lg bg-indigo-600 px-4 py-2.5 text-center text-sm font-semibold text-white hover:bg-indigo-500">
              Start free trial
            </a>
          </div>
          {/* Enterprise */}
          <div className="rounded-2xl border border-gray-200 p-8">
            <h3 className="text-lg font-semibold text-gray-900">Enterprise</h3>
            <p className="mt-2 text-sm text-gray-500">For large organizations</p>
            <p className="mt-6"><span className="text-4xl font-bold text-gray-900">Custom</span></p>
            <a href="#" className="mt-8 block rounded-lg border border-indigo-600 px-4 py-2.5 text-center text-sm font-semibold text-indigo-600 hover:bg-indigo-50">
              Contact sales
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
```

### Pattern 44: Highlighted Recommended Plan

Visual emphasis on the plan you want most visitors to choose.

**When to use**: Always, when showing multiple plans. Most people choose the middle option.
**Key elements**: Larger card, different border/background color, "Recommended" or "Most Popular" badge.

### Pattern 45: Annual Savings Toggle

Monthly/annual toggle showing the percentage saved on annual billing.

**When to use**: When annual pricing offers a meaningful discount (15-30%).
**Key elements**: Toggle switch, savings badge ("Save 20%"), annual price shown per month.

### Pattern 46: Feature Gating

Clear display of which features unlock at each pricing tier.

**When to use**: When features are the primary differentiator between tiers.
**Key elements**: Checkmark/X grid, feature descriptions on hover, grouped by category.

### Pattern 47: Free Tier

A permanently free plan with limited features.

**When to use**: For PLG (product-led growth) strategies. Reduces trial friction to zero.
**Key elements**: Clear limits, easy upgrade path, no credit card required.

### Pattern 48: Money-Back Guarantee at Pricing

Guarantee badge placed directly adjacent to the pricing/CTA area.

**When to use**: For all paid products. Placing the guarantee next to the price reduces
purchase anxiety at the critical decision point.

### Pattern 49: ROI Calculator at Pricing

An interactive calculator that shows the return on investment at each price point.

**When to use**: For B2B products where the price needs justification.
**Key elements**: Industry benchmarks, customizable inputs, clear ROI visualization.

### Pattern 50: Custom Quote CTA

"Get a custom quote" or "Talk to sales" for enterprise and complex pricing.

**When to use**: When pricing varies significantly by customer size, usage, or requirements.
**Key elements**: Quick form (company size, use case), expected response time, no-pressure language.

---

## Bonus Patterns

### Pattern 51: Sticky Table of Contents

A sidebar or sticky nav that shows page sections and highlights the current one.

**When to use**: On long-form landing pages (2,000+ words). Helps visitors navigate and signals
comprehensive content.

### Pattern 52: Comparison vs. Competitors

A direct comparison table showing your product versus named competitors.

**When to use**: When visitors are comparing solutions and your product wins on key criteria.
**Caution**: Be fair and accurate. Inaccurate comparisons damage trust.

### Pattern 53: Interactive Product Tour

A guided walkthrough of the product embedded in the landing page.

**When to use**: For products with complex UIs where a static screenshot does not do justice.
**Key elements**: Step-by-step with annotations, clickable hotspots, progress indicator.

### Pattern 54: Countdown Timer

A real-time countdown to the end of a promotion or limited offer.

**When to use**: Only when the deadline is real. Fake countdowns destroy credibility.
**Key elements**: Days, hours, minutes, seconds. Explanation of what happens when time expires.

### Pattern 55: Personalized CTA

CTA copy that changes based on visitor segment, referral source, or behavior.

**When to use**: When you have visitor segmentation data (UTM parameters, cookies, geolocation).
**Examples**: "Start your agency plan" for visitors from agency-related ads.

---

## Pattern Selection Guide

Choose patterns based on your conversion goal:

| Goal | Primary Patterns | Supporting Patterns |
|------|-----------------|-------------------|
| SaaS free trial | Direct Promise Hero, Single Field Form, Risk-Reversal CTA | Logo Bar, Testimonial Cards, Feature Gating |
| B2B lead gen | Interactive Hero, Multi-Step Form, Case Study Snippets | Comparison Table, ROI Calculator, Metric Counters |
| Direct sale | Before/After Hero, Comparison Table, Money-Back Guarantee | Urgency CTA, Star Ratings, Countdown Timer |
| Newsletter signup | Problem Statement Hero, Single Field Form, Lead Magnet | Social Proof CTA, Community Size |
| Demo booking | Social Proof Hero, Micro-Commitment CTA, Testimonial Cards | Logo Bar, Media Mentions, Customer Support Access |
