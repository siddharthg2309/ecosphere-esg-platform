# EcoSphere — Design System Reference

> Visual language for the EcoSphere ESG Management Platform.
> Derived from the **Odoo** design language ([hackathon.odoo.com](https://hackathon.odoo.com/) and
> [github.com/odoo/design-themes](https://github.com/odoo/design-themes) — `theme_common` base), adapted for an
> ESG / sustainability product. The low-fidelity admin reference is [`wireframe.png`](wireframe.png).
> The tokens below are implemented verbatim in [`wireframes/assets/theme.css`](wireframes/assets/theme.css).

---

## 1. Principles

1. **Calm, data-dense, flat.** Odoo screens are information-rich but never noisy: white cards on a light gray
   canvas, thin borders, minimal shadow, generous whitespace between groups. Let data breathe.
2. **One accent per module.** The chrome (sidebar, top bar, primary buttons) is Odoo **plum**. Each ESG module
   carries its own accent (green / blue / violet / amber) so users always know which pillar they are in.
3. **Status is color-coded and consistent.** Green = good/approved/on-track, Amber = warning/pending/under-review,
   Red = danger/overdue/rejected, Gray = draft/neutral. Never invent a new status color.
4. **Typographic contrast.** Body is **Inter** (clean, legible). A single **Caveat** (handwritten) flourish is used
   sparingly for section personality — exactly as the hackathon site underlines its hero heading.
5. **Accessible by default.** Text meets WCAG AA (4.5:1), interactive targets ≥ 40px, visible focus rings, never
   color-only signalling (pair every status color with a label or icon).

---

## 2. Color Palette

### 2.1 Brand (chrome, primary actions) — Odoo plum

| Token | Hex | Usage |
| --- | --- | --- |
| `--brand-900` | `#1B1319` | Sidebar background (deepest plum-black) |
| `--brand-700` | `#583B51` | Active nav item, hover on plum |
| `--brand-600` | `#65435C` | Primary button hover |
| `--brand` | `#714B67` | **Primary** — buttons, links, active tab underline, logo |
| `--brand-100` | `#F1EAEF` | Plum tint — selected row, badge background |

### 2.2 Neutrals (surfaces & text)

| Token | Hex | Usage |
| --- | --- | --- |
| `--ink` | `#212529` | Primary text |
| `--ink-muted` | `#6C757D` | Secondary text, captions, table meta |
| `--line` | `#E9ECEF` | Card & table borders, dividers |
| `--line-strong` | `#DEE2E6` | Input borders, stronger separators |
| `--surface` | `#FFFFFF` | Card / panel background |
| `--canvas` | `#F8F9FA` | Page background, table header, footer strips |

### 2.3 Semantic / status

| Token | Hex | Meaning |
| --- | --- | --- |
| `--success` | `#1F9254` | Approved, Active, On-track, Completed, Resolved |
| `--success-bg` | `#E6F4EC` | Success pill background |
| `--warning` | `#C77700` | Pending, Under Review, Due-soon |
| `--warning-bg` | `#FBF0DD` | Warning pill background |
| `--danger` | `#C0392B` | Overdue, Rejected, High severity, Open issue |
| `--danger-bg` | `#FBEAE7` | Danger pill background |
| `--info` | `#2F6FED` | Informational, links inside content |
| `--info-bg` | `#E7EEFD` | Info pill background |
| `--neutral` | `#6C757D` | Draft, Archived, Inactive |
| `--neutral-bg` | `#EEF0F2` | Neutral pill background |

### 2.4 Module accents (one per pillar)

| Module | Token | Hex | Tint |
| --- | --- | --- | --- |
| Environmental | `--env` | `#1F9254` (green) | `#E6F4EC` |
| Social | `--social` | `#2F6FED` (blue) | `#E7EEFD` |
| Governance | `--gov` | `#6D28D9` (violet) | `#EDE7FB` |
| Gamification | `--game` | `#E8833A` (amber) | `#FBEEE1` |

### 2.5 ESG score bands (for scores, gauges, rankings)

| Band | Range | Hex |
| --- | --- | --- |
| Excellent | 80–100 | `#1F9254` |
| Good | 60–79 | `#7BAF3A` |
| Fair | 40–59 | `#C77700` |
| Poor | 0–39 | `#C0392B` |

---

## 3. Typography

**Fonts** (from the Odoo hackathon site): body **Inter**, display flourish **Caveat**. Load Inter 300/400/500/600/700
and Caveat 400/700. Fallback stack: `Inter, "Segoe UI", system-ui, -apple-system, sans-serif`.

| Role | Font / Weight | Size / Line | Notes |
| --- | --- | --- | --- |
| Display flourish | Caveat 700 | 40–52 / 1.1 | Hero, empty-state headline, section personality only |
| H1 (page title) | Inter 700 | 26 / 32 | One per screen |
| H2 (section) | Inter 700 | 20 / 28 | Card group headers |
| H3 (card title) | Inter 600 | 16 / 24 | Card / widget titles |
| Body | Inter 400 | 14 / 22 | Default |
| Body-strong | Inter 600 | 14 / 22 | Emphasis, key figures inline |
| Metric / KPI | Inter 700 | 30–36 / 1.1 | Score tiles, big numbers |
| Small / caption | Inter 400 | 12–13 / 18 | Table meta, helper text |
| Overline / label | Inter 600 | 11 / 16, `letter-spacing: .06em`, UPPERCASE | Field labels, tab group titles |

---

## 4. Spacing, Layout & Grid

- **Spacing scale (px):** `4 · 8 · 12 · 16 · 24 · 32 · 48 · 64`. Token names `--space-1`…`--space-8`.
- **App shell:** fixed **left sidebar 240px** (plum-black) + **top bar 56px** + fluid content on `--canvas`.
- **Content max-width:** 1280px, centered, 24–32px gutters.
- **Card grid:** 12-col fluid; KPI rows use 4 equal cards; module content uses auto-fit cards `minmax(260px, 1fr)`.
- **Card padding:** 20px (`--space-4` + 4). Group vertical rhythm: 24px between card rows.
- **Mobile (< 768px):** sidebar collapses to an icon rail / drawer; KPI cards stack 1-col; tables become horizontally
  scrollable inside an `overflow-x:auto` wrapper (never let the page scroll sideways).

---

## 5. Elevation, Radius, Borders

| Token | Value | Usage |
| --- | --- | --- |
| `--radius-sm` | 6px | Buttons, inputs, pills |
| `--radius` | 10px | Cards, panels, modals |
| `--radius-pill` | 999px | Status chips, avatars, date badges |
| `--shadow-sm` | `0 1px 2px rgba(16,24,40,.06)` | Resting card |
| `--shadow-md` | `0 4px 14px rgba(16,24,40,.10)` | Hover card, dropdown |
| `--shadow-lg` | `0 12px 32px rgba(16,24,40,.16)` | Modal / dialog |

Borders are the primary separation device (Odoo is low-shadow): `1px solid var(--line)` on every card and table.

---

## 6. Components

**Card (Odoo pattern).** White surface, 1px `--line` border, `--radius`, `--shadow-sm`. Optional colored top-accent
strip (3px, module accent). Anatomy from the Odoo Events card: **media/date badge → bold title → muted description →
meta row (icon + text) → light footer strip** (`--canvas`) for status/actions.

**Buttons.**
- Primary: plum fill `--brand`, white text, `--radius-sm`, 10×16 padding, hover `--brand-600`.
- Secondary: white fill, `--line-strong` border, `--ink` text.
- Ghost / tertiary: transparent, plum text.
- Danger: `--danger` fill (destructive confirmations only).
- Sizes: sm (32px), md (40px, default), icon (40×40).

**Inputs / selects.** White, 1px `--line-strong`, `--radius-sm`, 40px tall, 12px inset. Focus: 2px `--brand` ring
(`box-shadow: 0 0 0 3px rgba(113,75,103,.18)`). Labels are overline style above the field. Show inline validation text
in `--danger` below the field.

**Tabs (module sub-nav).** Horizontal, underline style. Active tab: `--ink` text + 2px `--brand` (or module accent)
underline. Inactive: `--ink-muted`. Matches the wireframe's per-module tab rows.

**Status pill / chip.** `--radius-pill`, 2×10 padding, 12px/600 text, `{semantic}-bg` background + `{semantic}` text.
Always carries a word ("Active", "Overdue"), never color alone.

**Score tile (KPI).** Large card, overline label, big metric (`--metric`), colored by ESG band, optional trend arrow
(▲ green / ▼ red) + delta. Used on the dashboard for Environmental / Social / Governance / Overall ESG.

**Progress bar.** 8px track `--line`, fill in module accent or ESG band; used for goal progress (Target vs Current CO₂)
and challenge progress.

**Data table.** Header row on `--canvas`, 11px overline labels; body rows 44px, 1px `--line` dividers, hover
`--brand-100`. Right-align numerics. Row actions (View/Edit/Delete) as ghost icon buttons revealed on hover.

**Sidebar nav.** Plum-black `--brand-900`, grouped by module with an icon + label; active item pill `--brand-700`.
Collapsible groups (Environmental, Social, Governance, Gamification, Reports, Settings).

**Top bar.** White, 1px bottom `--line`: breadcrumb / page title left; global search, notification bell (with count
badge), user avatar right.

**Toggle / switch.** Used heavily in Settings → ESG Configuration. 36×20 track, plum when on.

**Notification toast & bell dropdown.** Toast top-right, `--shadow-md`, accent left-border by type. Bell dropdown lists
the 4 notification types (compliance raised, approval decision, policy reminder, badge unlock) with icon + timestamp.

**Modal.** Centered, max 560px, `--radius`, `--shadow-lg`, header + body + right-aligned action footer.

**Empty state.** Centered icon, a Caveat one-liner, muted helper text, primary CTA.

---

## 7. Data Visualization

Chart palette (categorical, colorblind-safe, in order): `#714B67 · #1F9254 · #2F6FED · #E8833A · #6D28D9 · #7BAF3A`.
- **Emissions trend** = line/area, single series, `--env` with 12% fill.
- **Department ESG ranking** = horizontal/vertical bars colored by ESG band.
- **Diversity** = donut using the categorical palette.
- Axes/gridlines `--line`; labels `--ink-muted` 12px; always label units (t CO₂, %, XP).

---

## 8. Motion & Accessibility

- **Motion:** 120–200ms ease-out for hover/expand; respect `prefers-reduced-motion`. No decorative animation on data.
- **Contrast:** body ≥ 4.5:1, large text ≥ 3:1. Plum `#714B67` on white = 6.6:1 ✓.
- **Focus:** every interactive element has a visible `--brand` focus ring; logical tab order.
- **Targets:** ≥ 40×40px; table row actions expand hit-area to the full cell.
- **Semantics:** real `<button>`/`<a>`/`<table>`/`<label>`; status conveyed by text+icon, not color alone.
- **Dark mode (optional, v2):** invert canvas/surface to `#16181C` / `#1F2226`, keep accents; tokens already isolated
  so a `:root[data-theme="dark"]` block is the only change needed.

---

## 9. Token → CSS map

All tokens above are CSS custom properties on `:root` in
[`wireframes/assets/theme.css`](wireframes/assets/theme.css). Components in the hi-fi wireframes
(`wireframes/*.html`) consume only these variables — swap a token, restyle the whole app. When implemented in React,
these become the single source of truth for the design tokens layer (see `plan.md` → Frontend / MVVM).
