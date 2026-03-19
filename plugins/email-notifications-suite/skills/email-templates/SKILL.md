---
name: email-templates
description: >
  Build responsive, cross-client email templates using React Email, MJML,
  or hand-coded HTML. Covers dark mode support, mobile responsiveness,
  component patterns, and email client compatibility.
  Triggers: "email template", "email design", "react email", "mjml",
  "html email", "responsive email", "email component".
  NOT for: sending emails (use transactional-email) or notification systems.
version: 1.0.0
argument-hint: "[react-email|mjml|html]"
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# Email Templates

Build responsive, cross-client email templates.

## Approach Selection

| Approach | Best For | Pros | Cons |
|----------|----------|------|------|
| **React Email** | TypeScript projects, component reuse | Type-safe, composable, great DX | Newer ecosystem |
| **MJML** | Marketing emails, designers | Easy responsive, good docs | Extra build step |
| **Hand-coded HTML** | Simple transactional, full control | No dependencies | Verbose, hard to maintain |

## React Email

### Setup

```bash
npm install @react-email/components react-email
npx create-email  # optional — scaffolds project
```

### Base Components

```tsx
// emails/components/Layout.tsx
import {
  Body,
  Container,
  Head,
  Html,
  Preview,
  Section,
  Text,
  Link,
  Img,
  Hr,
} from '@react-email/components';

interface LayoutProps {
  preview: string;
  children: React.ReactNode;
}

export function Layout({ preview, children }: LayoutProps) {
  return (
    <Html>
      <Head />
      <Preview>{preview}</Preview>
      <Body style={body}>
        <Container style={container}>
          {/* Logo */}
          <Section style={logoSection}>
            <Img
              src={`${process.env.APP_URL}/logo.png`}
              width={120}
              height={36}
              alt="App Name"
            />
          </Section>

          {/* Content */}
          <Section style={content}>
            {children}
          </Section>

          {/* Footer */}
          <Section style={footer}>
            <Hr style={hr} />
            <Text style={footerText}>
              &copy; {new Date().getFullYear()} App Name. All rights reserved.
            </Text>
            <Text style={footerText}>
              <Link href={`${process.env.APP_URL}/unsubscribe`} style={footerLink}>
                Unsubscribe
              </Link>
              {' | '}
              <Link href={`${process.env.APP_URL}/preferences`} style={footerLink}>
                Email preferences
              </Link>
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  );
}

// Styles (inline for email compatibility)
const body = {
  backgroundColor: '#f6f9fc',
  fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
  margin: '0',
  padding: '0',
};

const container = {
  margin: '0 auto',
  padding: '20px 0 48px',
  maxWidth: '600px',
};

const logoSection = {
  padding: '24px 32px 0',
};

const content = {
  backgroundColor: '#ffffff',
  borderRadius: '8px',
  padding: '32px',
  margin: '24px 0',
};

const footer = {
  padding: '0 32px',
};

const hr = {
  borderColor: '#e6ebf1',
  margin: '20px 0',
};

const footerText = {
  color: '#8898aa',
  fontSize: '12px',
  lineHeight: '16px',
  textAlign: 'center' as const,
};

const footerLink = {
  color: '#8898aa',
  textDecoration: 'underline',
};
```

### Button Component

```tsx
// emails/components/Button.tsx
import { Button as EmailButton } from '@react-email/components';

interface ButtonProps {
  href: string;
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'danger';
}

export function Button({ href, children, variant = 'primary' }: ButtonProps) {
  const colors = {
    primary: { bg: '#2563eb', text: '#ffffff' },
    secondary: { bg: '#f1f5f9', text: '#1e293b' },
    danger: { bg: '#dc2626', text: '#ffffff' },
  };

  const { bg, text } = colors[variant];

  return (
    <EmailButton
      href={href}
      style={{
        backgroundColor: bg,
        borderRadius: '6px',
        color: text,
        fontSize: '16px',
        fontWeight: '600',
        textDecoration: 'none',
        textAlign: 'center' as const,
        display: 'inline-block',
        padding: '12px 24px',
      }}
    >
      {children}
    </EmailButton>
  );
}
```

### Welcome Email Template

```tsx
// emails/Welcome.tsx
import { Text, Section } from '@react-email/components';
import { Layout } from './components/Layout';
import { Button } from './components/Button';

interface WelcomeEmailProps {
  name: string;
  loginUrl: string;
}

export default function WelcomeEmail({ name, loginUrl }: WelcomeEmailProps) {
  return (
    <Layout preview={`Welcome to App Name, ${name}!`}>
      <Text style={heading}>Welcome aboard, {name}!</Text>
      <Text style={paragraph}>
        Thanks for signing up. We're excited to have you. Here's everything you
        need to get started:
      </Text>

      <Section style={featureList}>
        <Text style={feature}>1. Set up your profile</Text>
        <Text style={feature}>2. Create your first project</Text>
        <Text style={feature}>3. Invite your team</Text>
      </Section>

      <Section style={{ textAlign: 'center', margin: '32px 0' }}>
        <Button href={loginUrl}>Get Started</Button>
      </Section>

      <Text style={paragraph}>
        Have questions? Just reply to this email — we read every message.
      </Text>
    </Layout>
  );
}

const heading = { fontSize: '24px', fontWeight: '700', color: '#1a1a1a', margin: '0 0 16px' };
const paragraph = { fontSize: '16px', lineHeight: '24px', color: '#4a4a4a', margin: '0 0 16px' };
const featureList = { margin: '24px 0', padding: '16px 24px', backgroundColor: '#f8fafc', borderRadius: '6px' };
const feature = { fontSize: '15px', lineHeight: '28px', color: '#374151', margin: '0' };
```

### Password Reset Template

```tsx
// emails/PasswordReset.tsx
import { Text, Section, Code } from '@react-email/components';
import { Layout } from './components/Layout';
import { Button } from './components/Button';

interface PasswordResetProps {
  name: string;
  resetUrl: string;
  expiresIn: string;
  ipAddress?: string;
}

export default function PasswordReset({ name, resetUrl, expiresIn, ipAddress }: PasswordResetProps) {
  return (
    <Layout preview="Reset your password">
      <Text style={heading}>Reset your password</Text>
      <Text style={paragraph}>Hi {name},</Text>
      <Text style={paragraph}>
        We received a request to reset your password. Click the button below to
        choose a new one:
      </Text>

      <Section style={{ textAlign: 'center', margin: '32px 0' }}>
        <Button href={resetUrl}>Reset Password</Button>
      </Section>

      <Text style={muted}>
        This link expires in {expiresIn}. If you didn't request a password reset,
        you can safely ignore this email.
      </Text>

      {ipAddress && (
        <Text style={muted}>
          This request was made from IP address: <Code>{ipAddress}</Code>
        </Text>
      )}
    </Layout>
  );
}

const heading = { fontSize: '24px', fontWeight: '700', color: '#1a1a1a', margin: '0 0 16px' };
const paragraph = { fontSize: '16px', lineHeight: '24px', color: '#4a4a4a', margin: '0 0 16px' };
const muted = { fontSize: '13px', lineHeight: '20px', color: '#8898aa', margin: '16px 0 0' };
```

### Order Confirmation Template

```tsx
// emails/OrderConfirmation.tsx
import { Text, Section, Row, Column, Img } from '@react-email/components';
import { Layout } from './components/Layout';
import { Button } from './components/Button';

interface OrderItem {
  name: string;
  quantity: number;
  price: string;
  total: string;
  image?: string;
}

interface OrderConfirmationProps {
  orderId: string;
  items: OrderItem[];
  subtotal: string;
  shipping: string;
  tax: string;
  total: string;
  shippingAddress: string;
  orderUrl: string;
}

export default function OrderConfirmation({
  orderId, items, subtotal, shipping, tax, total, shippingAddress, orderUrl,
}: OrderConfirmationProps) {
  return (
    <Layout preview={`Order confirmed — #${orderId}`}>
      <Text style={heading}>Order Confirmed</Text>
      <Text style={paragraph}>
        Your order <strong>#{orderId}</strong> has been placed successfully.
      </Text>

      {/* Items */}
      <Section style={table}>
        {items.map((item, i) => (
          <Row key={i} style={tableRow}>
            <Column style={{ width: '48px' }}>
              {item.image && (
                <Img src={item.image} width={48} height={48} alt={item.name}
                  style={{ borderRadius: '4px' }} />
              )}
            </Column>
            <Column style={{ paddingLeft: '12px' }}>
              <Text style={itemName}>{item.name}</Text>
              <Text style={itemMeta}>Qty: {item.quantity} x {item.price}</Text>
            </Column>
            <Column style={{ textAlign: 'right' }}>
              <Text style={itemTotal}>{item.total}</Text>
            </Column>
          </Row>
        ))}
      </Section>

      {/* Totals */}
      <Section style={totalsSection}>
        <Row><Column><Text style={totalLabel}>Subtotal</Text></Column><Column style={rightCol}><Text style={totalValue}>{subtotal}</Text></Column></Row>
        <Row><Column><Text style={totalLabel}>Shipping</Text></Column><Column style={rightCol}><Text style={totalValue}>{shipping}</Text></Column></Row>
        <Row><Column><Text style={totalLabel}>Tax</Text></Column><Column style={rightCol}><Text style={totalValue}>{tax}</Text></Column></Row>
        <Row style={totalRow}><Column><Text style={totalLabelBold}>Total</Text></Column><Column style={rightCol}><Text style={totalValueBold}>{total}</Text></Column></Row>
      </Section>

      {/* Shipping */}
      <Section style={addressSection}>
        <Text style={addressLabel}>Shipping to:</Text>
        <Text style={addressText}>{shippingAddress}</Text>
      </Section>

      <Section style={{ textAlign: 'center', margin: '24px 0' }}>
        <Button href={orderUrl}>View Order</Button>
      </Section>
    </Layout>
  );
}

const heading = { fontSize: '24px', fontWeight: '700', color: '#1a1a1a', margin: '0 0 16px' };
const paragraph = { fontSize: '16px', lineHeight: '24px', color: '#4a4a4a', margin: '0 0 24px' };
const table = { margin: '0 0 24px' };
const tableRow = { borderBottom: '1px solid #eee', padding: '12px 0' };
const itemName = { fontSize: '14px', fontWeight: '600', color: '#1a1a1a', margin: '0' };
const itemMeta = { fontSize: '13px', color: '#666', margin: '4px 0 0' };
const itemTotal = { fontSize: '14px', fontWeight: '600', color: '#1a1a1a', margin: '0' };
const totalsSection = { borderTop: '2px solid #eee', paddingTop: '16px' };
const totalLabel = { fontSize: '14px', color: '#666', margin: '4px 0' };
const totalValue = { fontSize: '14px', color: '#333', margin: '4px 0', textAlign: 'right' as const };
const totalRow = { borderTop: '1px solid #eee', paddingTop: '8px', marginTop: '8px' };
const totalLabelBold = { fontSize: '16px', fontWeight: '700', color: '#1a1a1a', margin: '8px 0' };
const totalValueBold = { fontSize: '16px', fontWeight: '700', color: '#1a1a1a', margin: '8px 0', textAlign: 'right' as const };
const rightCol = { textAlign: 'right' as const };
const addressSection = { backgroundColor: '#f8fafc', borderRadius: '6px', padding: '16px', margin: '16px 0' };
const addressLabel = { fontSize: '12px', fontWeight: '600', textTransform: 'uppercase' as const, color: '#666', margin: '0 0 8px' };
const addressText = { fontSize: '14px', color: '#333', margin: '0', whiteSpace: 'pre-line' as const };
```

### Rendering Emails

```typescript
// lib/email/render.ts
import { render } from '@react-email/render';
import WelcomeEmail from '../emails/Welcome';
import PasswordReset from '../emails/PasswordReset';
import OrderConfirmation from '../emails/OrderConfirmation';

const templates = {
  welcome: WelcomeEmail,
  'password-reset': PasswordReset,
  'order-confirmation': OrderConfirmation,
} as const;

export async function renderTemplate(
  name: keyof typeof templates,
  props: Record<string, any>
): Promise<{ subject: string; html: string; text: string }> {
  const Template = templates[name];
  if (!Template) throw new Error(`Unknown template: ${name}`);

  const component = Template(props as any);
  const html = await render(component);
  const text = await render(component, { plainText: true });

  // Subject extracted from Preview text or hardcoded map
  const subjects: Record<string, string> = {
    welcome: `Welcome to App Name, ${props.name}!`,
    'password-reset': 'Reset your password',
    'order-confirmation': `Order confirmed — #${props.orderId}`,
  };

  return { subject: subjects[name] || name, html, text };
}
```

### Preview Server

```bash
# Start the React Email dev server to preview templates
npx react-email dev --dir emails --port 3001
```

## MJML (Alternative)

### Setup

```bash
npm install mjml
```

### MJML Template

```xml
<!-- templates/welcome.mjml -->
<mjml>
  <mj-head>
    <mj-attributes>
      <mj-all font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" />
      <mj-text font-size="16px" line-height="24px" color="#4a4a4a" />
      <mj-button background-color="#2563eb" border-radius="6px" font-size="16px" font-weight="600" />
    </mj-attributes>
    <mj-preview>Welcome to App Name, {{name}}!</mj-preview>
  </mj-head>

  <mj-body background-color="#f6f9fc">
    <!-- Logo -->
    <mj-section padding="24px 0 0">
      <mj-column>
        <mj-image src="{{logoUrl}}" width="120px" alt="App Name" />
      </mj-column>
    </mj-section>

    <!-- Content -->
    <mj-section background-color="#ffffff" border-radius="8px" padding="32px">
      <mj-column>
        <mj-text font-size="24px" font-weight="700" color="#1a1a1a">
          Welcome aboard, {{name}}!
        </mj-text>
        <mj-text>
          Thanks for signing up. We're excited to have you.
        </mj-text>
        <mj-button href="{{loginUrl}}">
          Get Started
        </mj-button>
      </mj-column>
    </mj-section>

    <!-- Footer -->
    <mj-section padding="20px 0">
      <mj-column>
        <mj-text font-size="12px" color="#8898aa" align="center">
          &copy; 2025 App Name. All rights reserved.
        </mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
```

### Compile MJML

```typescript
import mjml2html from 'mjml';
import { readFileSync } from 'fs';

function renderMjml(templateName: string, data: Record<string, any>): string {
  let mjml = readFileSync(`templates/${templateName}.mjml`, 'utf-8');

  // Replace variables
  for (const [key, value] of Object.entries(data)) {
    mjml = mjml.replaceAll(`{{${key}}}`, String(value));
  }

  const { html, errors } = mjml2html(mjml, { validationLevel: 'strict' });
  if (errors.length) console.warn('MJML warnings:', errors);

  return html;
}
```

## Dark Mode Support

```css
/* Add to email <head> for dark mode support */
@media (prefers-color-scheme: dark) {
  .email-body { background-color: #1a1a2e !important; }
  .email-card { background-color: #16213e !important; }
  .email-text { color: #e0e0e0 !important; }
  .email-heading { color: #ffffff !important; }
  .email-muted { color: #a0a0a0 !important; }
  .email-link { color: #60a5fa !important; }
}

/* Meta tag for Apple Mail */
<meta name="color-scheme" content="light dark">
<meta name="supported-color-schemes" content="light dark">
```

## Email Client Compatibility

| Feature | Gmail | Outlook | Apple Mail | Yahoo |
|---------|-------|---------|------------|-------|
| `<style>` in `<head>` | Yes | Partial | Yes | Yes |
| Media queries | No | No | Yes | No |
| Flexbox | No | No | Yes | No |
| CSS Grid | No | No | Partial | No |
| `background-image` | Yes | No | Yes | Yes |
| `border-radius` | Yes | No | Yes | Yes |
| Dark mode CSS | No | Partial | Yes | No |
| `max-width` | Yes | No | Yes | Yes |
| Web fonts | Yes | No | Yes | No |

### Safe Patterns

- Use tables for layout (not div/flexbox/grid)
- Inline all critical styles
- Use `align` and `valign` attributes on `<td>`
- Set explicit widths on images
- Use `<!--[if mso]>` conditionals for Outlook
- Always include `alt` text on images
- Test with Litmus or Email on Acid

## Testing

```bash
# React Email preview server
npx react-email dev

# Send test email (Resend)
npx ts-node -e "
  import { render } from '@react-email/render';
  import { Resend } from 'resend';
  import WelcomeEmail from './emails/Welcome';

  const resend = new Resend(process.env.RESEND_API_KEY);
  const html = await render(WelcomeEmail({ name: 'Test', loginUrl: 'https://example.com' }));
  await resend.emails.send({ from: 'test@yourdomain.com', to: 'you@email.com', subject: 'Test', html });
"
```
