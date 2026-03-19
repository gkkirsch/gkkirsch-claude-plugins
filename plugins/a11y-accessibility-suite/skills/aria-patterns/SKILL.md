---
name: aria-patterns
description: >
  Implement correct ARIA patterns for custom UI widgets. Provides copy-paste
  accessible implementations for modals, tabs, accordions, comboboxes, menus,
  tooltips, and more. Follows WAI-ARIA Authoring Practices 1.2.
  Triggers: "aria pattern", "accessible widget", "make this accessible",
  "screen reader support", "a11y for this component".
  NOT for: native HTML elements that don't need ARIA.
version: 1.0.0
argument-hint: "[widget type: modal, tabs, accordion, combobox, menu, etc.]"
allowed-tools: Read, Grep, Glob, Write, Edit
---

# ARIA Patterns

Implement accessible ARIA patterns for custom UI widgets.

## Step 1: Identify the Widget

If an argument was provided (e.g., "modal", "tabs"), jump to that pattern.
Otherwise, scan the codebase for custom widgets that need ARIA:

```
Grep: "Modal\|Dialog\|Popup\|Overlay" — dialog patterns
Grep: "Tab\|TabPanel\|TabList" — tab patterns
Grep: "Accordion\|Collapsible\|Expandable" — disclosure patterns
Grep: "Dropdown\|Select\|Combobox\|Autocomplete\|Typeahead" — combobox patterns
Grep: "Menu\|MenuBar\|ContextMenu" — menu patterns
Grep: "Tooltip\|Popover" — tooltip patterns
Grep: "Toast\|Notification\|Alert\|Banner" — alert patterns
Grep: "Carousel\|Slider\|Slideshow" — carousel patterns
Grep: "TreeView\|Tree\|FileTree" — tree patterns
```

## Step 2: Apply the Correct Pattern

### Modal Dialog Pattern

**Required ARIA:**
- `role="dialog"` on the container
- `aria-modal="true"` on the container
- `aria-labelledby` pointing to the title
- `aria-describedby` pointing to description (optional)

**Required behavior:**
1. Move focus to first focusable element (or the dialog) on open
2. Trap Tab/Shift+Tab inside the dialog
3. Close on Escape
4. Return focus to trigger on close
5. Set `inert` or `aria-hidden="true"` on background content

**React implementation:**
```tsx
import { useEffect, useRef, useCallback } from 'react';

function useDialogA11y(isOpen: boolean, onClose: () => void) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (isOpen) {
      triggerRef.current = document.activeElement as HTMLElement;
      // Focus first focusable element
      requestAnimationFrame(() => {
        const focusable = dialogRef.current?.querySelector<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        focusable?.focus();
      });
    } else {
      triggerRef.current?.focus();
    }
  }, [isOpen]);

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
      return;
    }
    if (e.key !== 'Tab') return;

    const focusable = dialogRef.current?.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    if (!focusable?.length) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first.focus();
    }
  }, [onClose]);

  useEffect(() => {
    if (!isOpen) return;
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, handleKeyDown]);

  return dialogRef;
}
```

### Tabs Pattern

**Required ARIA:**
- `role="tablist"` on the container
- `role="tab"` on each tab button
- `role="tabpanel"` on each panel
- `aria-selected="true"` on active tab
- `aria-controls` linking tab to panel
- `aria-labelledby` linking panel to tab

**Keyboard pattern (roving tabindex):**
- Left/Right arrows: move between tabs
- Home/End: jump to first/last tab
- Tab: move into the panel
- Only active tab has `tabindex="0"`, others have `tabindex="-1"`

```tsx
function useTabs(tabCount: number) {
  const [activeIndex, setActiveIndex] = useState(0);

  const handleKeyDown = (e: React.KeyboardEvent, index: number) => {
    let newIndex = index;
    switch (e.key) {
      case 'ArrowRight':
        newIndex = (index + 1) % tabCount;
        break;
      case 'ArrowLeft':
        newIndex = (index - 1 + tabCount) % tabCount;
        break;
      case 'Home':
        newIndex = 0;
        break;
      case 'End':
        newIndex = tabCount - 1;
        break;
      default:
        return;
    }
    e.preventDefault();
    setActiveIndex(newIndex);
    // Focus the new tab
    document.getElementById(`tab-${newIndex}`)?.focus();
  };

  return { activeIndex, setActiveIndex, handleKeyDown };
}
```

### Accordion Pattern

**Required ARIA:**
- Heading element wrapping a `<button>`
- `aria-expanded` on the button
- `aria-controls` linking to the panel
- `role="region"` + `aria-labelledby` on the panel

**Keyboard:**
- Enter/Space: toggle section
- Optional: Up/Down arrows between headers, Home/End

### Combobox (Autocomplete) Pattern

**Required ARIA:**
- `role="combobox"` on the input
- `aria-autocomplete="list"` (or `"both"` or `"inline"`)
- `aria-expanded` (true when listbox open)
- `aria-controls` linking to the listbox
- `aria-activedescendant` pointing to highlighted option
- `role="listbox"` on the dropdown list
- `role="option"` on each option

**Keyboard:**
- Down: open listbox, move to next option
- Up: move to previous option
- Enter: select highlighted option
- Escape: close listbox
- Type-ahead: filter options

### Toast/Notification Pattern

**Required ARIA:**
- `role="alert"` for errors/important (assertive)
- `role="status"` for success/info (polite)
- Container MUST exist in DOM before content is added

```tsx
// Accessible toast container — always mounted
function ToastContainer({ toasts }) {
  return (
    <div aria-live="polite" aria-atomic="false" className="toast-container">
      {toasts.map(toast => (
        <div key={toast.id} role={toast.type === 'error' ? 'alert' : 'status'}>
          {toast.message}
        </div>
      ))}
    </div>
  );
}
```

### Disclosure Pattern

**Required ARIA:**
- `aria-expanded` on the trigger button
- `aria-controls` linking to the content panel

This is the simplest ARIA pattern. A button toggles visibility.

```html
<button aria-expanded="false" aria-controls="details">Show details</button>
<div id="details" hidden>Details content...</div>
```

## Step 3: Implement and Verify

1. Apply the correct ARIA pattern to the identified widget
2. Add keyboard event handlers
3. Test: Tab to the widget, use arrow keys, use Enter/Space, use Escape
4. Verify in browser DevTools → Accessibility panel → check roles and properties