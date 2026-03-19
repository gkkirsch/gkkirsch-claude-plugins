---
name: keyboard-navigation
description: >
  Implement and fix keyboard navigation patterns. Ensures all interactive elements
  are reachable and operable by keyboard. Covers focus management, skip navigation,
  focus traps, roving tabindex, and custom keyboard shortcuts.
  Triggers: "keyboard navigation", "keyboard accessible", "tab order",
  "focus management", "skip link", "keyboard trap".
  NOT for: keyboard shortcuts in desktop apps.
version: 1.0.0
argument-hint: "[component or page to fix]"
allowed-tools: Read, Grep, Glob, Write, Edit
---

# Keyboard Navigation

Make your application fully navigable by keyboard alone.

## Step 1: Audit Keyboard Access

### Find interactive elements missing keyboard support

```
# Elements with click handlers that aren't natively focusable
Grep: "onClick=\|@click=\|on:click=" — then check if element is <button>, <a>, <input>, <select>, <textarea>

# Divs and spans acting as buttons
Grep: "<div.*onClick\|<span.*onClick\|<div.*@click\|<span.*@click"

# Mouse-only interactions
Grep: "onMouseDown\|onMouseUp\|onMouseOver\|onMouseEnter\|onDragStart"

# Focus visibility removed
Grep: "outline:\s*none\|outline:\s*0\|\*:focus.*outline"
```

### Check for keyboard handlers
```
Grep: "onKeyDown\|onKeyUp\|onKeyPress\|@keydown\|@keyup"
```

## Step 2: Fix Common Issues

### Issue 1: Clickable div/span → Use native elements

```tsx
// BAD
<div className="btn" onClick={handleClick}>
  Submit
</div>

// GOOD
<button type="button" onClick={handleClick} className="btn">
  Submit
</button>
```

If you MUST use a non-native element (rare):
```tsx
<div
  role="button"
  tabIndex={0}
  onClick={handleClick}
  onKeyDown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick(e);
    }
  }}
>
  Submit
</div>
```

### Issue 2: Missing focus indicator

```css
/* BAD — removes focus for everyone */
*:focus {
  outline: none;
}

/* GOOD — custom focus indicator that's visible */
*:focus-visible {
  outline: 2px solid #005fcc;
  outline-offset: 2px;
}

/* GOOD — remove default only on mouse click, keep for keyboard */
button:focus:not(:focus-visible) {
  outline: none;
}
button:focus-visible {
  outline: 2px solid #005fcc;
  outline-offset: 2px;
  border-radius: 3px;
}
```

**Tailwind CSS:**
```html
<button class="focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 focus:outline-none">
  Submit
</button>
```

### Issue 3: Skip Navigation Link

Add as the FIRST focusable element on the page:

```html
<body>
  <a href="#main-content" class="skip-link">
    Skip to main content
  </a>
  <nav>...</nav>
  <main id="main-content" tabindex="-1">
    ...
  </main>
</body>

<style>
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  background: #000;
  color: #fff;
  padding: 8px 16px;
  z-index: 100;
  transition: top 0.2s;
}
.skip-link:focus {
  top: 0;
}
</style>
```

**React/Next.js:**
```tsx
// components/SkipLink.tsx
export function SkipLink() {
  return (
    <a
      href="#main-content"
      className="sr-only focus:not-sr-only focus:absolute focus:top-0 focus:left-0 focus:z-50 focus:bg-black focus:text-white focus:p-2"
    >
      Skip to main content
    </a>
  );
}

// In layout:
<SkipLink />
<nav>...</nav>
<main id="main-content" tabIndex={-1}>...</main>
```

### Issue 4: Focus Trapping in Modals

When a modal is open, Tab/Shift+Tab must cycle through modal elements only — never escape to the page behind.

```tsx
function useFocusTrap(isActive: boolean) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isActive || !containerRef.current) return;

    const container = containerRef.current;
    const getFocusableElements = () =>
      container.querySelectorAll<HTMLElement>(
        'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
      );

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      const focusable = getFocusableElements();
      if (focusable.length === 0) return;

      const first = focusable[0];
      const last = focusable[focusable.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isActive]);

  return containerRef;
}
```

### Issue 5: Focus Restoration

When a modal/dropdown/popover closes, focus must return to the element that opened it.

```tsx
function useRestoreFocus() {
  const triggerRef = useRef<HTMLElement | null>(null);

  const saveTrigger = () => {
    triggerRef.current = document.activeElement as HTMLElement;
  };

  const restoreFocus = () => {
    triggerRef.current?.focus();
    triggerRef.current = null;
  };

  return { saveTrigger, restoreFocus };
}
```

### Issue 6: Roving Tabindex for Widget Groups

For tab lists, menu bars, radio groups, toolbars — only ONE item in the group should be in the tab order.

```tsx
function useRovingTabIndex(itemCount: number) {
  const [activeIndex, setActiveIndex] = useState(0);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    let newIndex = activeIndex;
    switch (e.key) {
      case 'ArrowRight':
      case 'ArrowDown':
        newIndex = (activeIndex + 1) % itemCount;
        break;
      case 'ArrowLeft':
      case 'ArrowUp':
        newIndex = (activeIndex - 1 + itemCount) % itemCount;
        break;
      case 'Home':
        newIndex = 0;
        break;
      case 'End':
        newIndex = itemCount - 1;
        break;
      default:
        return;
    }
    e.preventDefault();
    setActiveIndex(newIndex);
  };

  const getTabIndex = (index: number) => (index === activeIndex ? 0 : -1);

  return { activeIndex, setActiveIndex, handleKeyDown, getTabIndex };
}
```

### Issue 7: Route Change Announcements (SPA)

Single-page apps don't trigger screen reader page load announcements. You need to announce route changes.

```tsx
// React Router
function RouteAnnouncer() {
  const location = useLocation();
  const [announcement, setAnnouncement] = useState('');

  useEffect(() => {
    // Use the page title or h1 as the announcement
    const title = document.title || 'Page loaded';
    setAnnouncement(title);
  }, [location.pathname]);

  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      className="sr-only"
    >
      {announcement}
    </div>
  );
}
```

**Next.js**: Built-in route announcer since Next.js 11+ (no action needed).

## Step 3: Keyboard Interaction Patterns Reference

| Widget | Enter/Space | Arrow Keys | Tab | Escape |
|--------|-------------|------------|-----|--------|
| Button | Activates | — | Moves to/from | — |
| Link | Follows | — | Moves to/from | — |
| Checkbox | Toggles | — | Moves to/from | — |
| Radio group | Selects | Move selection | Into/out of group | — |
| Tab list | Activates tab | Move between tabs | Into panel | — |
| Menu | Selects item | Navigate items | — | Closes menu |
| Dialog | — | — | Cycles within | Closes dialog |
| Combobox | Selects option | Navigate options | Into/out | Closes listbox |
| Accordion | Toggles panel | Between headers | Move to/from | — |
| Tree | Toggles/selects | Navigate nodes | Into/out | — |
| Slider | — | Change value | Into/out | — |

## Step 4: Testing

After implementing fixes:

1. **Tab through the entire page** — every interactive element should receive focus in logical order
2. **Activate everything by keyboard** — buttons with Enter/Space, links with Enter
3. **Test modals** — focus traps, Escape closes, focus returns
4. **Test dropdowns** — arrow key navigation, Escape closes
5. **Check focus visibility** — every focused element should have a visible indicator
6. **Test without mouse** — unplug it. Can you use the entire app?