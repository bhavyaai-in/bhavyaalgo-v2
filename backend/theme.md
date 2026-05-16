# Design Tokens — BhavyaAI.com

## Fonts

| Usage | Font | Source |
|---|---|---|
| Body | **Satoshi** (300, 400, 500, 700, 900) | `https://api.fontshare.com/v2/css?f[]=satoshi@1,900,700,500,300,400&display=swap` |
| Headings | **Satoshi** | same as body |
| Code | **Source Code Pro**, monospace | Google Fonts |

### Secondary / Variant Fonts

| Font | Used In |
|---|---|
| Urbanist | Home Two variant |
| DM Serif Display | Home Two, Three, Five |
| Caveat | Home CRM variant |
| Montserrat Alternates | Home CRM variant (headings) |

### CSS Variable Font Tokens

```css
--heading-font: "Satoshi", sans-serif;
--body-font: "Satoshi", sans-serif;
--font-xs: 0.75rem;
--font-sm: 0.875rem;
--font-base: 1rem;
--font-lg: 1.125rem;
--font-xl: 1.25rem;
--font-2xl: 1.5rem;
--font-3xl: 1.875rem;
```

### Fluid Heading Sizes

```css
--heading-one:   clamp(2.125rem, -0.0733rem + 5.5vw,   4.4375rem);
--heading-two:   clamp(1.875rem,  0.7133rem + 2.8846vw, 3.75rem);
--heading-three: clamp(1.75rem,   0.3353rem + 2.1661vw, 3rem);
--heading-four:  clamp(1.5rem,    0.5569rem + 1.444vw,  1.875rem);
--heading-five:  clamp(1.125rem,  1.2rem + 0.722vw,     1.5rem);
--heading-six:   clamp(1rem,      0.769rem + 0.6813vw,  1.25rem);
```

---

## Brand Colors

### Primary Palette

| Token | HSL | Hex (approx) | Role |
|---|---|---|---|
| `--primary` / `--main` | `186 92% 42%` | `#0993A0` | Brand primary (teal/cyan) |
| `--secondary` / `--main-two` | `205 100% 20%` | `#003366` | Brand secondary (dark navy) |

### Light Theme (`:root`)

| CSS Variable | HSL | Role |
|---|---|---|
| `--background` | `210 14% 96%` | Page background |
| `--foreground` | `0 0% 43%` | Body text |
| `--card` | `210 20% 98%` | Card surface |
| `--card-foreground` | `0 0% 43%` | Card text |
| `--popover` | `210 20% 98%` | Popover surface |
| `--popover-foreground` | `0 0% 43%` | Popover text |
| `--primary` | `186 92% 42%` | Primary brand |
| `--primary-foreground` | `0 0% 100%` | On-primary text |
| `--secondary` | `205 100% 20%` | Secondary brand |
| `--secondary-foreground` | `0 0% 100%` | On-secondary text |
| `--muted` | `214 17% 91%` | Muted surface |
| `--muted-foreground` | `215 14% 47%` | Muted text |
| `--accent` | `186 92% 88.4%` | Accent surface (light teal) |
| `--accent-foreground` | `186 92% 33.6%` | Accent text (dark teal) |
| `--destructive` | `0 84.2% 60.2%` | Destructive/error red |
| `--destructive-foreground` | `0 0% 98%` | On-destructive text |
| `--border` | `215 20% 84%` | Borders |
| `--input` | `215 20% 84%` | Input borders |
| `--ring` | `186 92% 42%` | Focus ring |
| `--radius` | `0.5rem` | Border radius |

### Dark Theme (`.dark`)

| CSS Variable | HSL | Role |
|---|---|---|
| `--background` | `222.2 84% 4.9%` | Page background |
| `--foreground` | `210 14% 96%` | Body text |
| `--card` | `222.2 84% 6.9%` | Card surface |
| `--card-foreground` | `210 14% 96%` | Card text |
| `--popover` | `222.2 84% 4.9%` | Popover surface |
| `--popover-foreground` | `210 14% 96%` | Popover text |
| `--primary` | `186 92% 42%` | Primary brand (same) |
| `--primary-foreground` | `0 0% 10%` | On-primary text |
| `--secondary` | `205 100% 20%` | Secondary brand (same) |
| `--secondary-foreground` | `0 0% 98%` | On-secondary text |
| `--muted` | `217.2 32.6% 17.5%` | Muted surface |
| `--muted-foreground` | `215 20.2% 65.1%` | Muted text |
| `--accent` | `186 92% 88.4%` | Accent surface (same) |
| `--accent-foreground` | `186 92% 33.6%` | Accent text (same) |
| `--destructive` | `0 62.8% 30.6%` | Destructive red |
| `--destructive-foreground` | `0 0% 98%` | On-destructive text |
| `--border` | `217.2 32.6% 17.5%` | Borders |
| `--input` | `217.2 32.6% 17.5%` | Input borders |
| `--ring` | `186 92% 42%` | Focus ring |

---

## Extended Color Palette

### Split HSL Components (for dynamic use)

```css
--main-h: 186;  --main-s: 92%;   --main-l: 42%;
--main-two-h: 205; --main-two-s: 100%; --main-two-l: 20%;
--yellow-h: 40;  --yellow-s: 100%; --yellow-l: 54%;
--spring-green-h: 144; --spring-green-s: 80%; --spring-green-l: 55%;
--pink-h: 340; --pink-s: 82%; --pink-l: 60%;
--pink-dark-h: 340; --pink-dark-s: 70%; --pink-dark-l: 45%;
--pink-lighter-h: 340; --pink-lighter-s: 100%; --pink-lighter-l: 95%;
--purple-h: 270; --purple-s: 70%; --purple-l: 65%;
--purple-light-h: 270; --purple-light-s: 100%; --purple-light-l: 96%;
--paste-h: 170; --paste-s: 50%; --paste-l: 60%;
--paste-light-h: 170; --paste-light-s: 70%; --paste-light-l: 96%;
```

Usage: `hsl(var(--spring-green-h) var(--spring-green-s) var(--spring-green-l))`

### Derived Shades (main color, 50–900)

10 shades each for `--main` (teal) and `--main-two` (navy), e.g. `--main-50`, `--main-100`, ..., `--main-900`.

### Static Tokens

| Token | Value | Role |
|---|---|---|
| `--white` | `0 0% 100%` | White |
| `--black` | `249 63% 15%` | Near-black for headings |
| `--body` | `0 0% 43%` | Body text color |
| `--orange` | `18 100% 55%` | Orange accent |
| `--blue-color` | `#1351D8` | Blue accent |
| `--heading-color` | `var(--black)` | Heading color |

### Neutral Scale

| Token | Hex |
|---|---|
| `--neutral-50` | `#f9fafb` |
| `--neutral-100` | `#f3f4f6` |
| `--neutral-200` | `#e5e7eb` |
| `--neutral-300` | `#d1d5db` |
| `--neutral-400` | `#9ca3af` |
| `--neutral-500` | `#6b7280` |
| `--neutral-600` | `#4b5563` |
| `--neutral-700` | `#374151` |
| `--neutral-800` | `#1f2937` |
| `--neutral-900` | `#111827` |
| `--neutral-950` | `#030712` |

### Toast / Notification Colors

| Token | Hex | Role |
|---|---|---|
| `--success-600` | `#16A34A` | Success |
| `--danger-600` | `#DC2626` | Danger |
| `--warning-600` | `#D97706` | Warning |
| `--info-600` | `#0284C7` | Info |

---

## Chart Colors (shadcn/ui style)

| Token | Light HSL | Dark HSL |
|---|---|---|
| `--chart-1` | `12 76% 61%` | `220 70% 50%` |
| `--chart-2` | `173 58% 39%` | `160 60% 45%` |
| `--chart-3` | `197 37% 24%` | `30 80% 55%` |
| `--chart-4` | `43 74% 66%` | `280 65% 60%` |
| `--chart-5` | `27 87% 67%` | `340 75% 55%` |

---

## Border Radius

```css
--radius: 0.5rem;          /* lg */
--radius - 2px: 0.375rem;   /* md (calc) */
--radius - 4px: 0.25rem;    /* sm (calc) */
circle: 50%;
```

---

## Tailwind Config (reusable)

```ts
// tailwind.config.ts
fontFamily: {
  body: ["Satoshi", "sans-serif"],
  headline: ["Satoshi", "sans-serif"],
  code: ['"Source Code Pro"', "monospace"],
},
borderRadius: {
  lg: "var(--radius)",
  md: "calc(var(--radius) - 2px)",
  sm: "calc(var(--radius) - 4px)",
  circle: "50%",
},
colors: {
  background: "hsl(var(--background))",
  foreground: "hsl(var(--foreground))",
  primary: { DEFAULT: "hsl(var(--primary))", foreground: "hsl(var(--primary-foreground))" },
  secondary: { DEFAULT: "hsl(var(--secondary))", foreground: "hsl(var(--secondary-foreground))" },
  muted: { DEFAULT: "hsl(var(--muted))", foreground: "hsl(var(--muted-foreground))" },
  accent: { DEFAULT: "hsl(var(--accent))", foreground: "hsl(var(--accent-foreground))" },
  destructive: { DEFAULT: "hsl(var(--destructive))", foreground: "hsl(var(--destructive-foreground))" },
  border: "hsl(var(--border))",
  input: "hsl(var(--input))",
  ring: "hsl(var(--ring))",
  pink: { DEFAULT: "hsl(var(--pink-h) var(--pink-s) var(--pink-l))", dark: "hsl(var(--pink-dark-h) ...)", lighter: "hsl(var(--pink-lighter-h) ...)" },
  purple: { DEFAULT: "hsl(var(--purple-h) ...)", light: "hsl(var(--purple-light-h) ...)" },
  paste: { DEFAULT: "hsl(var(--paste-h) ...)", light: "hsl(var(--paste-light-h) ...)" },
  "spring-green": "hsl(var(--spring-green))",
},
```

---