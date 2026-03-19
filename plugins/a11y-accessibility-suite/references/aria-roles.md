# ARIA Roles, States, and Properties Reference

Complete reference for WAI-ARIA 1.2 roles, states, and properties.

---

## Document Structure Roles

These roles define the structure of the page. Most are redundant with semantic HTML — prefer native elements.

| Role | HTML Equivalent | When to Use ARIA |
|------|----------------|------------------|
| `banner` | `<header>` (top-level) | Only if `<header>` isn't available |
| `complementary` | `<aside>` | Only if `<aside>` isn't available |
| `contentinfo` | `<footer>` (top-level) | Only if `<footer>` isn't available |
| `form` | `<form>` | Only if `<form>` can't be used |
| `main` | `<main>` | Only if `<main>` isn't available |
| `navigation` | `<nav>` | Only if `<nav>` isn't available |
| `region` | `<section>` with label | Generic labeled landmark |
| `search` | `<search>` (HTML5.2) | Search functionality container |
| `article` | `<article>` | Independent content piece |
| `definition` | `<dfn>` | Term definition |
| `directory` | — | Table of contents, deprecated |
| `document` | — | Document content region |
| `group` | `<fieldset>` | Grouping of related elements |
| `heading` | `<h1>`-`<h6>` | With `aria-level` if not using native |
| `img` | `<img>` | For non-`<img>` images (SVG, CSS) |
| `list` | `<ul>`, `<ol>` | Only if native lists not used |
| `listitem` | `<li>` | Child of `list` |
| `math` | `<math>` | Mathematical expression |
| `note` | — | Parenthetical or ancillary content |
| `presentation` / `none` | — | Remove element's semantics |
| `separator` | `<hr>` (non-focusable) | Visual divider |
| `table` | `<table>` | Only if `<table>` can't be used |
| `row` | `<tr>` | Table/grid row |
| `rowgroup` | `<thead>`, `<tbody>`, `<tfoot>` | Group of rows |
| `cell` | `<td>` | Table cell |
| `columnheader` | `<th scope="col">` | Column header |
| `rowheader` | `<th scope="row">` | Row header |

---

## Widget Roles

Interactive components. These REQUIRE keyboard handling and appropriate states.

| Role | Purpose | Required States/Properties | Keyboard |
|------|---------|---------------------------|----------|
| `alert` | Important message | — | Auto-announces |
| `alertdialog` | Alert requiring response | `aria-labelledby` or `aria-label` | Focus trapped |
| `button` | Clickable action | `aria-pressed` (toggle), `aria-expanded` (disclosure) | Enter, Space |
| `checkbox` | Toggle option | `aria-checked` (required) | Space |
| `combobox` | Text input + listbox | `aria-expanded`, `aria-controls`, `aria-autocomplete` | See combobox pattern |
| `dialog` | Modal dialog | `aria-modal`, `aria-labelledby` | Esc closes, focus trapped |
| `grid` | Interactive table | — | Arrow keys |
| `gridcell` | Cell in grid | — | Part of grid |
| `link` | Hyperlink | — | Enter |
| `listbox` | List of selectable options | `aria-multiselectable` | Up/Down, Home/End |
| `menu` | Menu of actions | — | Arrow keys, Enter |
| `menubar` | Horizontal menu | — | Left/Right, Enter |
| `menuitem` | Item in menu | — | Enter |
| `menuitemcheckbox` | Checkable menu item | `aria-checked` | Space |
| `menuitemradio` | Radio menu item | `aria-checked` | Space |
| `option` | Option in listbox | `aria-selected` | Part of listbox |
| `progressbar` | Progress indicator | `aria-valuenow`, `aria-valuemin`, `aria-valuemax` | — |
| `radio` | Radio button | `aria-checked` | Arrow keys in group |
| `radiogroup` | Group of radios | — | Arrow keys |
| `scrollbar` | Scroll control | `aria-valuenow`, `aria-orientation` | Arrow keys |
| `searchbox` | Search input | — | Type + Enter |
| `slider` | Range selector | `aria-valuenow`, `aria-valuemin`, `aria-valuemax` | Arrow keys |
| `spinbutton` | Numeric input | `aria-valuenow`, `aria-valuemin`, `aria-valuemax` | Up/Down |
| `status` | Status message | — | Auto-announces (polite) |
| `switch` | On/off toggle | `aria-checked` (required) | Space |
| `tab` | Tab in tablist | `aria-selected`, `aria-controls` | Arrow keys in tablist |
| `tablist` | Tab container | `aria-orientation` | Arrow keys |
| `tabpanel` | Content for tab | `aria-labelledby` | Tab from tablist |
| `textbox` | Text input | `aria-multiline`, `aria-readonly` | Type |
| `toolbar` | Toolbar | `aria-orientation` | Arrow keys |
| `tooltip` | Tooltip | — | Shows on focus + hover |
| `tree` | Tree view | — | Arrow keys |
| `treeitem` | Item in tree | `aria-expanded`, `aria-selected` | Arrow keys, Enter |
| `treegrid` | Tree + grid | — | Arrow keys |

---

## ARIA States and Properties

### Global States (Apply to Any Role)

| Attribute | Type | Values | Purpose |
|-----------|------|--------|---------|
| `aria-atomic` | Property | `true`, `false` | Whether live region re-reads entire content on change |
| `aria-busy` | State | `true`, `false` | Region is being updated, wait before announcing |
| `aria-controls` | Property | ID reference(s) | Element(s) controlled by this element |
| `aria-current` | State | `page`, `step`, `location`, `date`, `time`, `true`, `false` | Current item in a set |
| `aria-describedby` | Property | ID reference(s) | Element(s) that describe this element |
| `aria-description` | Property | String | Accessible description text |
| `aria-details` | Property | ID reference | Element providing extended description |
| `aria-disabled` | State | `true`, `false` | Element is disabled (prefer `disabled` attribute) |
| `aria-dropeffect` | Property | `copy`, `execute`, `link`, `move`, `none`, `popup` | Drag-and-drop effect (deprecated) |
| `aria-errormessage` | Property | ID reference | Element containing error message for this element |
| `aria-flowto` | Property | ID reference(s) | Next element(s) in reading order |
| `aria-grabbed` | State | `true`, `false` | Element is grabbed for drag (deprecated) |
| `aria-haspopup` | Property | `true`, `menu`, `listbox`, `tree`, `grid`, `dialog`, `false` | Popup type |
| `aria-hidden` | State | `true`, `false` | Element hidden from accessibility API |
| `aria-invalid` | State | `true`, `false`, `grammar`, `spelling` | Input value is invalid |
| `aria-keyshortcuts` | Property | String | Keyboard shortcuts for this element |
| `aria-label` | Property | String | Accessible name (when no visible text) |
| `aria-labelledby` | Property | ID reference(s) | Element(s) that label this element |
| `aria-live` | Property | `assertive`, `polite`, `off` | Live region announcement priority |
| `aria-owns` | Property | ID reference(s) | Elements logically owned (for DOM order fix) |
| `aria-relevant` | Property | `additions`, `removals`, `text`, `all` | What changes to announce in live region |
| `aria-roledescription` | Property | String | Human-readable role description |

### Widget States

| Attribute | Type | Roles | Values | Purpose |
|-----------|------|-------|--------|---------|
| `aria-autocomplete` | Property | combobox, textbox | `inline`, `list`, `both`, `none` | Type of autocomplete |
| `aria-checked` | State | checkbox, radio, switch, menuitemcheckbox, menuitemradio, option | `true`, `false`, `mixed` | Checked state |
| `aria-expanded` | State | button, combobox, link, menuitem, row, tab, treeitem | `true`, `false` | Expanded/collapsed |
| `aria-level` | Property | heading, listitem, row, treeitem | Integer | Hierarchical level |
| `aria-modal` | Property | dialog, alertdialog | `true`, `false` | Element is modal |
| `aria-multiline` | Property | textbox | `true`, `false` | Multi-line input |
| `aria-multiselectable` | Property | grid, listbox, tablist, tree | `true`, `false` | Multiple selection allowed |
| `aria-orientation` | Property | listbox, menu, menubar, radiogroup, scrollbar, separator, slider, tablist, toolbar | `horizontal`, `vertical` | Element orientation |
| `aria-placeholder` | Property | textbox, searchbox | String | Placeholder hint (prefer HTML) |
| `aria-posinset` | Property | article, listitem, menuitem, option, radio, row, tab, treeitem | Integer | Position in set |
| `aria-pressed` | State | button | `true`, `false`, `mixed` | Toggle button state |
| `aria-readonly` | Property | checkbox, combobox, grid, gridcell, listbox, radiogroup, slider, spinbutton, textbox | `true`, `false` | Read-only |
| `aria-required` | Property | checkbox, combobox, gridcell, listbox, radiogroup, spinbutton, textbox, tree | `true`, `false` | Required field |
| `aria-selected` | State | gridcell, option, row, tab, treeitem | `true`, `false` | Selected state |
| `aria-setsize` | Property | article, listitem, menuitem, option, radio, row, tab, treeitem | Integer | Total items in set |
| `aria-sort` | Property | columnheader, rowheader | `ascending`, `descending`, `none`, `other` | Sort direction |
| `aria-valuemax` | Property | meter, progressbar, scrollbar, separator, slider, spinbutton | Number | Maximum value |
| `aria-valuemin` | Property | meter, progressbar, scrollbar, separator, slider, spinbutton | Number | Minimum value |
| `aria-valuenow` | Property | meter, progressbar, scrollbar, separator, slider, spinbutton | Number | Current value |
| `aria-valuetext` | Property | meter, progressbar, scrollbar, separator, slider, spinbutton | String | Human-readable value text |

---

## Required Owned Elements (Parent-Child Requirements)

Some roles REQUIRE specific child roles:

| Parent Role | Required Children |
|-------------|-------------------|
| `list` | `listitem` |
| `menu` | `menuitem`, `menuitemcheckbox`, `menuitemradio` |
| `menubar` | `menuitem`, `menuitemcheckbox`, `menuitemradio` |
| `radiogroup` | `radio` |
| `tablist` | `tab` |
| `tree` | `treeitem` |
| `grid` | `row` → `gridcell` |
| `table` | `row` → `cell` |
| `rowgroup` | `row` |
| `row` (in grid) | `gridcell`, `columnheader`, `rowheader` |
| `row` (in table) | `cell`, `columnheader`, `rowheader` |

**If a role requires specific children, wrapping in `role="none"` / `role="presentation"` is necessary for intermediate elements:**

```html
<!-- List inside a styled wrapper -->
<ul role="list">
  <div role="none">  <!-- Intermediary must be role="none" -->
    <li role="listitem">Item</li>
  </div>
</ul>
```

---

## Naming and Describing

### How Accessible Names Are Computed (Priority Order)

1. `aria-labelledby` (references visible text from another element)
2. `aria-label` (string attribute, invisible to sighted users)
3. Native label (`<label for="id">`, `<caption>`, `<legend>`, `<figcaption>`)
4. Element content (text inside `<button>`, `<a>`, etc.)
5. `title` attribute (tooltip text, unreliable — avoid for primary name)
6. `placeholder` attribute (NOT a name — supplementary hint only)

### Rules

- Every interactive element MUST have an accessible name
- `aria-labelledby` overrides everything (even `<label>`)
- `aria-label` overrides native label and content
- `aria-describedby` is supplementary (read after the name)
- Descriptions are optional; names are required

### Examples

```html
<!-- Name from content -->
<button>Save Document</button>
<!-- SR: "Save Document, button" -->

<!-- Name from aria-label -->
<button aria-label="Close dialog">×</button>
<!-- SR: "Close dialog, button" -->

<!-- Name from aria-labelledby -->
<h2 id="section-title">Billing Address</h2>
<form aria-labelledby="section-title">
<!-- SR: "Billing Address, form" -->

<!-- Name + Description -->
<input aria-label="Password" aria-describedby="pw-hint" />
<p id="pw-hint">Must be at least 8 characters</p>
<!-- SR: "Password, edit. Must be at least 8 characters" -->
```

---

## Screen Reader Behavior Quick Reference

| ARIA | Screen Reader Announces |
|------|------------------------|
| `role="alert"` | Immediately interrupts with content |
| `role="status"` | At next pause, reads content |
| `aria-live="polite"` | At next pause, reads changed content |
| `aria-live="assertive"` | Immediately interrupts with changed content |
| `aria-expanded="true"` | "expanded" |
| `aria-expanded="false"` | "collapsed" |
| `aria-checked="true"` | "checked" |
| `aria-checked="false"` | "not checked" |
| `aria-checked="mixed"` | "partially checked" |
| `aria-selected="true"` | "selected" |
| `aria-pressed="true"` | "pressed" (toggle button) |
| `aria-current="page"` | "current page" |
| `aria-invalid="true"` | "invalid" |
| `aria-required="true"` | "required" |
| `aria-disabled="true"` | "dimmed" / "unavailable" |
| `aria-hidden="true"` | (not announced at all) |
| `aria-busy="true"` | (region changes not announced) |
| `aria-sort="ascending"` | "sorted ascending" |
| `aria-haspopup="true"` | "has popup" |
| `aria-modal="true"` | Content outside dialog hidden from AT |
