# Strata Studio Design System

This document details the extracted design tokens, color palette, typography, layout structure, and component styling rules for the **Strata Studio** application, derived from the Stitch design system.

---

## 🎨 Color Palette

The interface follows a refined developer-centric aesthetic using a warm neutral foundation paired with a vibrant primary indigo accent.

### Core Brand & Neutrals
*   **Primary (Soft Indigo):** `#4648d4` (Used for primary actions, headers, and active states)
*   **Primary Accent (Hover):** `#6063ee` / `#645efb`
*   **Secondary (Bright Purple/Indigo):** `#4b41e1`
*   **Background / Surface (Warm Neutral):** `#f8f9fa` (Reduces eye strain compared to clinical white)
*   **Surface Bright:** `#f8f9fa`
*   **Surface Dim:** `#d9dadb`

### Surface Container Hierarchy
To establish clean visual hierarchy, the following container shades are utilized:
*   **Lowest Container (Card Backgrounds, Modals):** `#ffffff` (Pure White)
*   **Low Container (Sidebar):** `#f3f4f5`
*   **Standard Container:** `#edeeef`
*   **High Container:** `#e7e8e9`
*   **Highest Container:** `#e1e3e4`

### Text & Outlines
*   **On-Surface (Primary Text):** `#191c1d`
*   **On-Surface Variant (Secondary Text):** `#464554`
*   **Outline (Borders):** `#767586`
*   **Outline Variant (Dividers):** `#c7c4d7`

### Semantic Colors
*   **Error:** `#ba1a1a`
*   **Error Container:** `#ffdad6`
*   **On-Error:** `#ffffff`
*   **Success (Emerald-like, for status):** Green/Emerald theme mappings.

---

## 📐 Spacing & Layout Structure

The layout is built around a **4px baseline grid** to maintain visual alignment and rhythm.

### Grid & Breakpoints
*   **Sidebar Navigation:** Fixed width of `256px` (`w-64`) on Desktop.
*   **Desktop Layout:** Left sidebar + fluid 12-column grid content area with `32px` (`p-margin-desktop`) padding.
*   **Mobile Layout:** Single-column layout with top-bar navigation/drawer and `16px` (`p-margin-mobile`) padding.

### Spacing Scale
*   **Base / xs:** `4px`
*   **sm:** `8px`
*   **md:** `16px`
*   **lg:** `24px` (Standard gutter / card padding)
*   **xl:** `32px` (Section gaps)

---

## 🔤 Typography

The application uses **Plus Jakarta Sans** for UI and labels, and **JetBrains Mono** for code fragments, JSON payloads, and technical details.

| Style Name | Font Family | Size | Weight | Line Height | Letter Spacing |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Headline LG** | Plus Jakarta Sans | `30px` | `700` (Bold) | `38px` | `-0.02em` |
| **Headline LG Mobile** | Plus Jakarta Sans | `24px` | `700` (Bold) | `32px` | `-0.02em` |
| **Headline MD** | Plus Jakarta Sans | `20px` | `600` (Semi-Bold) | `28px` | `-0.01em` |
| **Body LG** | Plus Jakarta Sans | `16px` | `400` (Regular) | `24px` | Normal |
| **Body MD** | Plus Jakarta Sans | `14px` | `400` (Regular) | `20px` | Normal |
| **Label MD** | Plus Jakarta Sans | `12px` | `600` (Semi-Bold) | `16px` | `0.05em` |
| **Code SM** | JetBrains Mono | `13px` | `400` (Regular) | `18px` | Normal |

---

## 💎 Elevation, Depth & Shapes

### Corner Radii (Roundness)
*   **Standard Inputs & Buttons:** `8px` (`0.5rem` / `rounded-lg`)
*   **Cards & Main Containers:** `12px` (`0.75rem` / `rounded-xl`)
*   **Inner elements / nested buttons:** `4px` to `6px` (`rounded` or `rounded-md`)
*   **Pills / Status Badges:** `9999px` (`rounded-full`)

### Elevation & Shadows
*   **Level 0 (Floor):** `#f8f9fa` background (no shadow).
*   **Level 1 (Cards):** White background with subtle shadow: `shadow-sm` (`0 1px 2px 0 rgba(0, 0, 0, 0.05)`).
*   **Level 2 (Dropdowns, Modals):** White background with `shadow-md` or `shadow-2xl` (`0 4px 6px -1px rgba(0,0,0,0.1)`).

---

## 🧩 Components Design Guidelines

### Buttons
*   **Primary Button:** Solid Indigo (`#4648d4`) background, White text, `8px` border radius. On hover: opacity `90%` or primary-container (`#6063ee`).
*   **Secondary Button:** Container-lowest/bright background, Outline border (`#c7c4d7`), On-surface text.
*   **Ghost Button:** Transparent background, primary indigo text.

### Inputs & Textareas
*   **Background:** `#ffffff`
*   **Border:** `1px` solid `#c7c4d7`
*   **Focus State:** Changes to Primary Indigo (`#4648d4`) with a `3px` outer glow at `10%` opacity (`ring-[3px] ring-primary/10`).

### Cards & Panels
*   **Background:** `#ffffff`
*   **Radius:** `12px` (`rounded-xl`)
*   **Shadow:** `shadow-sm`
*   **Slide-over Side Panels:** Translate from right, full height, width max-width `md` (`448px`), `shadow-2xl`, with white backdrop-blur overlay over dashboard content.
