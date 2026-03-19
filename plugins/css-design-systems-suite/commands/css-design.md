# /css-design Command

Activate the CSS & Design Systems Suite for expert guidance on modern CSS, design systems, Tailwind CSS, and animations.

## Usage

```
/css-design [subcommand] [description]
```

## Subcommands

### `/css-design architect`
Activate the **CSS Architect** agent for modern CSS architecture.

Best for:
- Setting up cascade layers (`@layer`) for a project
- Building responsive layouts with CSS Grid, Flexbox, and container queries
- Implementing `:has()` patterns for parent-aware styling
- CSS custom properties architecture (3-tier token system)
- Logical properties for internationalization
- CSS nesting, `@scope`, anchor positioning
- Modern selectors (`:is()`, `:where()`, `:nth-child(of S)`)

Example: `/css-design architect Set up a cascade layer system for my component library`

### `/css-design system`
Activate the **Design System Builder** agent for design system construction.

Best for:
- Design token architecture (naming, tiers, Style Dictionary)
- Component API design (variants, compound components, slots)
- Multi-brand theming and dark mode
- Storybook documentation and stories
- Versioning strategy and changelog management
- Figma-to-code token sync
- Accessibility compliance across components

Example: `/css-design system Create a 3-tier token architecture with Style Dictionary`

### `/css-design tailwind`
Activate the **Tailwind Expert** agent for Tailwind CSS v4.

Best for:
- Tailwind v4 CSS-first configuration (`@theme`, `@variant`, `@utility`)
- Custom utilities and variants
- Component extraction (when and how)
- Dark mode strategies (class-based, semantic tokens)
- Responsive and container query patterns
- Tailwind plugin development
- Performance optimization and CSS splitting

Example: `/css-design tailwind Set up Tailwind v4 with a custom design token theme`

### `/css-design animate`
Activate the **Animation Specialist** agent for web animations.

Best for:
- CSS transitions and keyframe animations
- View Transitions API (same-document and cross-document)
- Framer Motion (React declarative animation)
- GSAP timelines and ScrollTrigger
- Scroll-driven animations (CSS `animation-timeline`)
- Micro-interactions (buttons, toggles, accordions)
- Animation performance optimization

Example: `/css-design animate Add scroll-driven reveal animations to my sections`

### `/css-design review`
Review existing CSS for modern best practices.

Checks for:
- Usage of cascade layers vs specificity hacks
- Container queries vs viewport media queries for components
- Logical properties vs physical properties
- Modern color functions (oklch) vs legacy (hex/rgb/hsl)
- Performance issues (layout thrashing, excessive repaints)
- Accessibility (reduced motion, forced colors, contrast)
- Custom properties architecture

Example: `/css-design review Review my component styles for modern CSS improvements`

## Quick Reference

| Task | Agent |
|------|-------|
| Layout & architecture | `architect` |
| Tokens & components | `system` |
| Tailwind configuration | `tailwind` |
| Motion & transitions | `animate` |
| Code review | `review` |
