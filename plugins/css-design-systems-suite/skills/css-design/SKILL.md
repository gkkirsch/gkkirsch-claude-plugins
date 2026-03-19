# CSS & Design Systems Suite

Expert guidance for modern CSS architecture, design systems, Tailwind CSS v4, and web animations.

## Agents

### CSS Architect (`css-architect`)
Modern CSS expert specializing in:
- **Cascade Layers** (`@layer`): Organize styles into priority tiers, eliminate specificity wars
- **Container Queries**: Components that adapt to their container, not the viewport
- **:has() Selector**: Parent selection, form state styling, quantity queries, layout logic
- **CSS Nesting**: Native nesting syntax without preprocessors
- **Custom Properties**: Three-tier token architecture, dark mode, dynamic color manipulation
- **CSS Grid & Flexbox**: Complex layouts, subgrid, auto-responsive grids, Every Layout patterns
- **Logical Properties**: Internationalization-ready layouts (RTL/LTR)
- **@scope**: Prevent style bleed between components
- **Anchor Positioning**: Position elements relative to other elements without JS
- **Modern Selectors**: `:is()`, `:where()`, `:nth-child(of S)`, `text-wrap`
- **Scroll-Driven Animations**: CSS-native scroll progress, parallax, reveal effects
- **Accessibility**: Focus management, reduced motion, forced colors, high contrast

### Design System Builder (`design-system-builder`)
Design system architect specializing in:
- **Token Architecture**: CTI naming convention, three-tier system (primitive → semantic → component)
- **Style Dictionary**: Transforms, formats, multi-platform output, multi-brand builds
- **Component API Design**: CVA variants, compound components, polymorphic rendering
- **Theming**: Multi-brand tokens, dark mode, CSS custom property theming, React ThemeProvider
- **Storybook**: CSF 3 stories, autodocs, usage guidelines, variant showcases
- **Versioning**: Semantic versioning strategy, changelogs, codemods for breaking changes
- **Figma Integration**: Tokens Studio sync, component mapping, CI/CD token pipeline
- **Component Checklist**: API, accessibility, theming, documentation, testing standards

### Tailwind Expert (`tailwind-expert`)
Tailwind CSS v4 specialist covering:
- **CSS-First Config**: `@theme`, `@import "tailwindcss"`, `@source` directives
- **Custom Utilities**: `@utility` for project-specific utilities (glass, text-gradient, scrollbar)
- **Custom Variants**: `@variant` for data attributes, container queries, motion preferences
- **Component Extraction**: When to extract, `@apply` patterns, React CVA components
- **Dark Mode**: Class-based strategy, semantic color tokens that eliminate `dark:` prefixes
- **Responsive Patterns**: Mobile-first, container queries with `@container`
- **UI Patterns**: Navigation, hero sections, data tables, forms — all in Tailwind
- **Performance**: Minimize arbitrary values, CSS splitting, `cn()` utility
- **Plugin Development**: Custom Tailwind plugins with `addBase`, `addComponents`, `addUtilities`

### Animation Specialist (`animation-specialist`)
Web animation expert specializing in:
- **CSS Transitions**: Custom easing curves, multi-property transitions, enter/exit timing
- **CSS Keyframes**: Staggered animations, loading states (spinner, skeleton, shimmer)
- **View Transitions API**: Same-document, cross-document (MPA), shared element transitions
- **Framer Motion**: AnimatePresence, layout animations, scroll-triggered, stagger variants
- **GSAP**: Timeline sequencing, ScrollTrigger, React integration with `@gsap/react`
- **Scroll-Driven CSS**: `animation-timeline: scroll()` / `view()`, parallax, header effects
- **Micro-Interactions**: Button feedback, toggle switches, badge pulses, accordion height
- **Performance**: Compositor-only properties, will-change patterns, content-visibility

## References

- **Modern CSS Features**: Browser support matrix, syntax quick reference for all modern features
- **Design Token Patterns**: Naming conventions, multi-brand strategy, Style Dictionary, Figma sync
- **CSS Performance**: Rendering pipeline, containment, content-visibility, font optimization, measurement

## When to Use This Suite

| Task | Agent |
|------|-------|
| Set up CSS architecture for a new project | CSS Architect |
| Build a design token system | Design System Builder |
| Configure Tailwind CSS v4 | Tailwind Expert |
| Add animations or transitions | Animation Specialist |
| Review CSS for modern best practices | CSS Architect |
| Create component APIs and Storybook docs | Design System Builder |
| Optimize CSS performance | CSS Architect + references |
| Implement dark mode / multi-brand theming | Design System Builder or Tailwind Expert |
| Build responsive layouts | CSS Architect or Tailwind Expert |
| Set up scroll-driven effects | Animation Specialist |
