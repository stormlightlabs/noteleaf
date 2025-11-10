---
title: Design System
sidebar_label: Design System
description: Color palette guidance that keeps the terminal UI cohesive.
sidebar_position: 4
---

# Design System

## Color Palette and Design System

Noteleaf uses a carefully chosen color palette defined in `internal/ui/palette.go`:

**Brand Colors**:

- **Malibu** (`#00A4FF`): Primary blue for accents and highlights
- **Julep** (`#00FFB2`): Success green for completed items
- **Sriracha** (`#EB4268`): Warning red for urgent/error states
- **Tang** (`#FF985A`): Orange for warnings and attention
- **Lichen** (`#5CDFEA`): Teal for informational elements

**Neutral Palette** (Dark to Light):

- **Pepper** (`#201F26`): Dark background
- **BBQ** (`#2d2c35`): Secondary background
- **Charcoal** (`#3A3943`): Tertiary background
- **Iron** (`#4D4C57`): Borders and subtle elements
- **Oyster** (`#605F6B`): Muted text
- **Smoke** (`#BFBCC8`): Secondary text
- **Salt** (`#F1EFEF`): Primary text in dark mode
- **Butter** (`#FFFAF1`): Light background

This palette ensures consistency across all UI components and provides excellent contrast for readability in terminal environments.
