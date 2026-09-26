---
name: Samaneh Amn Setad Khabari Herasat - Hamadan
description: Secure public report portal and administrative intelligence system for the Security Department of Hamadan Education Directorate.
colors:
  primary: "#3b82f6"
  primary-hover: "#2563eb"
  primary-dark: "#1d4ed8"
  neutral-bg: "#0b1120"
  neutral-surface: "rgba(30, 41, 59, 0.72)"
  neutral-surface-solid: "#1e293b"
  neutral-elevated: "rgba(15, 23, 42, 0.65)"
  neutral-input: "rgba(15, 23, 42, 0.6)"
  border-subtle: "rgba(255, 255, 255, 0.08)"
  border-accent: "rgba(59, 130, 246, 0.4)"
  text-main: "#f8fafc"
  text-muted: "#94a3b8"
  text-subtle: "#64748b"
  status-success: "#10b981"
  status-danger: "#f43f5e"
  status-warning: "#f59e0b"
typography:
  display:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "26px"
    fontWeight: 700
    lineHeight: 1.4
    letterSpacing: "-0.5px"
  headline:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "20px"
    fontWeight: 700
    lineHeight: 1.5
  title:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "16px"
    fontWeight: 600
    lineHeight: 1.6
  body:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "14.5px"
    fontWeight: 400
    lineHeight: 1.7
  label:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "13px"
    fontWeight: 500
    lineHeight: 1.5
rounded:
  sm: "6px"
  md: "10px"
  lg: "16px"
  full: "9999px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.text-main}"
    rounded: "{rounded.sm}"
    padding: "14px 24px"
  button-primary-hover:
    backgroundColor: "{colors.primary-hover}"
  button-secondary:
    backgroundColor: "{colors.neutral-surface-solid}"
    textColor: "{colors.text-main}"
    rounded: "{rounded.sm}"
    padding: "14px 22px"
---

# Design System: سامانه دیده‌بان امین - ستاد خبری و دریافت گزارش‌های مردمی حراست همدان

## Overview

**Creative North Star: "The Sovereign Digital Sentinel - سامانه دیده‌بان امین"**

The visual language of the Herasat Portal bridges high-security institutional trust with ultra-modern digital craftsmanship. Engineered specifically for confidential intelligence and report handling, the interface emits calm authority, strict privacy reassurance, and uncompromising clarity. Deep obsidian and navy tones (`#0b1120`, `#1e293b`) establish a solid cryptographic foundation, accented with laser-sharp cobalt blue (`#3b82f6`) and emerald validation accents (`#10b981`).

The experience is divided into two distinct operational modes:
1. **Persuade & Reassure (Public Portal):** A frictionless 3-step reporting wizard with transparent anonymity guarantees, drag-and-drop cryptographic evidence upload, instant tracking receipt issuance, and two-way encrypted followup threads.
2. **Operate (Admin Intelligence Console):** High-density, scan-optimized situational awareness, multi-level access control, comprehensive audit trails, automated profanity/vulgarity defense, and instant print/export readiness.

### Key Characteristics:
- **Directional RTL Excellence:** Native Right-to-Left geometry with contextual navigation arrows (Forward = Left `←`, Backward = Right `→`).
- **Standardized Iconography:** Pure Lucide & Tabler SVG glyphs with standardized 2.0px stroke-weight.
- **Glassmorphism & Depth:** Frosted dark slate surfaces with 10px backdrop blur and delicate 1px boundary highlights.
- **Micro-tactile Feedback:** Dynamic hover translations, push states, and non-blocking toast notifications.

---

## Colors

The color palette centers on a dark-mode institutional system designed for high focus and zero visual fatigue.

### Primary
- **Cobalt Shield Blue** (`#3b82f6`): Primary action triggers, active step indicators, and interactive highlights.
- **Deep Slate Hover** (`#2563eb` / `#1d4ed8`): Focused and active button states.

### Neutral
- **Obsidian Deep Background** (`#0b1120`): Canvas background providing high-contrast base.
- **Frosted Slate Surface** (`rgba(30, 41, 59, 0.72)`): Container cards and modal overlays.
- **Elevated Navy Surface** (`#1e293b`): Solid inner components, dropdowns, and drawer panels.
- **Pure Crisp White** (`#f8fafc`): Primary headline and body typography.
- **Muted Platinum** (`#94a3b8`): Subtitles, helper text, and secondary details.
- **Subtle Slate** (`#64748b`): Disabled states, minor timestamps, and placeholder copy.

### Status & Accents
- **Cryptographic Emerald** (`#10b981`): Success notifications, completed wizard steps, and resolved status badges.
- **Security Rose** (`#f43f5e`): Destructive actions, file removal, profanity warnings, and validation errors.
- **Warning Amber** (`#f59e0b`): Character count limit warnings and pending status indicators.

### Named Rules
**The Single Accent Rule.** The primary blue accent (`#3b82f6`) is preserved exclusively for interactive forward actions and focal identifiers. Background surfaces never shift hues arbitrarily.

---

## Typography

**Primary Font:** Vazirmatn (with fallback: Tahoma, Arial, sans-serif)
**Direction:** RTL (Right-to-Left) with LTR preservation on cryptographic codes and phone numbers.

### Hierarchy
- **Display / Portal Header** (Bold 700, 26px, line-height 1.4, tracking -0.5px): Primary system title and security banner.
- **Headline / Section Header** (Bold 700, 20px, line-height 1.5): Step headings and success screens.
- **Title / Form Labels** (Medium 500 / Semi-Bold 600, 14.5px-16px): Input field labels and drawer titles.
- **Body / Descriptive Text** (Regular 400, 14.5px, line-height 1.7): Guideline explanations, legal notices, and messages.
- **Label / Micro-Badges** (Medium 500, 12px-13px): Character counters, status tags, and optional markers.

### Named Rules
**The Persian Digit Rule.** Numbers inside general Persian prose and step counters utilize localized Persian numerals (`۰-۹`), while tracking codes (`16-character alphanumeric`) and contact phone numbers enforce explicit LTR formatting.

---

## Layout

- **Container Model:** Maximum desktop constraint of `1020px` with central alignment and `35px 40px` internal padding.
- **Responsive Adaptability:** Fluid collapse from multi-column grid to 1-column mobile views at `< 768px`.
- **Viewpoint Stability:** Stable viewport sizing using `min-h-[100dvh]` to eliminate iOS/Android mobile address bar layout shifts.
- **RTL Geometry:** Spatial margins and padding flip strictly according to RTL flow (`margin-left` for leading icons, `margin-right` for trailing indicators).

---

## Elevation & Depth

Surfaces employ tonal layering with subtle 1px border outlines rather than harsh drop shadows.

### Shadow Vocabulary
- **Card Ambient** (`0 20px 40px -15px rgba(0, 0, 0, 0.5)`): Main container elevation over background artwork.
- **Accent Glow** (`0 0 20px rgba(59, 130, 246, 0.35)`): Focus states, step tracker pulses, and primary action hovers.
- **Inner Edge Border** (`1px solid rgba(255, 255, 255, 0.08)`): Subtle glass refraction edge on all dark panels.

---

## Shapes

- **Base Radius:** `6px` (`--radius-sm`) for inputs, buttons, and badges.
- **Panel Radius:** `10px` (`--radius-md`) for info cards, step containers, and dropzones.
- **Container Radius:** `16px` (`--radius-lg`) for the primary card container.
- **Pill Radius:** `9999px` (`--radius-full`) for circular step indicators and status chips.

---

## Components

### 1. Navigation & Pill Tab Bar
- Segmented dual-pill control (`ثبت گزارش جدید` and `پیگیری گزارش`) with sliding spring capsule icon animation (`.tab-icon-pill`) that traverses smoothly from right to left across the active tab.
- Drawer trigger button (`راهنما و سوالات متداول`) for opening the orthogonal-grid FAQ sliding drawer.

### 2. Multi-Step Form Tracker
- 3 distinct stages: `۱. هویت گزارش‌دهنده`, `۲. موضوع و شرح`, `۳. مستندات و ثبت`.
- Active circle pulses with glowing cobalt halo; completed steps illuminate with emerald checkmark (`✓`).

### 3. Glass Category Dropdown (`dropdown-menu`)
- Framer-motion inspired glassmorphic dropdown with Lucide categorized badge icons.
- Rotating ChevronDown (180°), blur-scale entry animation, keyboard Escape support, and auto-dismiss on outside click.

### 4. Centered Smart Evidence Upload (`be-ui-file-upload`)
- Centered drag-and-drop zone with animated cloud icon container and pill browse button.
- Real-time MIME & extension detection rendering dedicated SVG icons (Images, Archives/ZIPs, PDFs, Docs).
- Progress bar simulation (`#3b82f6` -> `#10b981`) and smart metadata chip.

### 5. FAQ Themed Accordion (`accordion-2`)
- Orthogonal background grid overlay (`32px x 32px` line matrix).
- Smooth collapsible items with rotating chevron (180°), dedicated category icons, and high-legibility response copy.

### 6. Admin 3D Sign-in Card (`sign-in-card-2`)
- Interactive 3D perspective card (`1200px`) tracking mouse coordinates (`rotateX`, `rotateY`).
- Quad traveling border light beams (top, right, bottom, left) with pulsing corner glows.
- Frosted crest container and password visibility eye toggling.

### 7. Action Buttons & Directional Arrows
- **Next Step Button (`مرحله بعد`):** Primary blue gradient, trailing arrow pointing Left (`←`) with dynamic `-4px` hover translation.
- **Previous Step Button (`مرحله قبل`):** Secondary slate gradient, leading arrow pointing Right (`→`) with dynamic `+4px` hover translation.
- **Submit Button (`ثبت گزارش`):** Primary gradient with integrated shield-check icon and asynchronous loading spinner.

### 8. Secure Tracking & Follow-up Thread
- 16-character monospace tracking code container with one-click clipboard copy and print receipt capabilities.
- Live follow-up conversation thread rendering reporter remarks vs official security responses.

### 9. Assistant Communication & Chat Language
- **Assistant Chat Language:** ALWAYS speak, respond, and chat in English with user in all conversations and chats, while preserving Persian for client-facing UI strings and portal content.

---

## Do's and Don'ts

### Do:
- **Do** maintain standard 2.0px stroke-width on all SVG glyphs.
- **Do** point forward navigation arrows to the Left (`←`) and backward arrows to the Right (`→`) in RTL mode.
- **Do** provide clear feedback cycles for every async action (spinner + disabled state).
- **Do** preserve LTR direction on tracking codes, phone numbers, and CAPTCHA fields.

### Don't:
- **Don't** mix multiple icon sets or introduce emoji icons inside critical UI buttons.
- **Don't** use pure black `#000000` for dark mode surfaces—use layered obsidian/slate tones.
- **Don't** allow form inputs or buttons to wrap text onto multiple lines at desktop breakpoints.
- **Don't** bypass validation checks or profanity screening before submission.
- **Don't** put comments in codes.

NO DEPENDANCIES FOR THIS PROJECT... EVERYTHING MUST BE LOCAL FOR SECURITY REASONS (BY EVERYTHING I MEAN ASSETS AND NOT PLUGINS).
PLUGINS OR EXTERNAL LIBRARIES ARE FORBIDDEN.