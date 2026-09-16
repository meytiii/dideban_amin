# Design Taste Frontend Rules

## Core Directives
- **Anti-Slop & Anti-Default:** Never default to generic AI purple gradients, generic glassmorphism, or centered heroes with equal cards.
- **RTL & Typography:** Respect Persian/RTL layout mechanics. Use proper font hierarchy (e.g. Vazirmatn / modern sans) with clean weights and line-heights.
- **Icon Discipline:**
  - Use recognized icon families (Lucide, Heroicons, Tabler Icons).
  - Never mix icon styles haphazardly; enforce consistent `stroke-width` (e.g., 2.0 / 2.2).
  - Respect RTL directionality for contextual arrows (Next points Left `←`, Previous points Right `→`).
- **Color Consistency Lock:** Single locked accent color with high-contrast neutral backgrounds.
- **Interactive UI & Motion:**
  - Micro-interactions on buttons, hover states, active tactile feedback (`scale(0.98)` or slight elevation shift).
  - Respect `prefers-reduced-motion`.
