# Project Rules

## Tech Stack
- **Styling:** Tailwind CSS
- **Components:** shadcn/ui
- **Animation:** Framer Motion
- **Icons:** Lucide React
- **Utilities:** clsx + tailwind-merge for class composition, cva for variants

## Coding Standards
- Use `cn()` from `@/lib/utils` (or equivalent) for merging Tailwind classes
- Use `cva` from class-variance-authority for component variant definitions
- Prefer shadcn/ui as the base — extend with Tailwind, don't recreate
- Co-locate related files (component + test + styles)
- absolute minimal styling - simplicity and readability over beauty

## Responsive Design
- Mobile-first: base styles for mobile, `sm:` / `md:` / `lg:` / `xl:` for larger
- Test at: 320px, 768px, 1024px, 1440px
- Responsive grid: `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6`

## Frontend Aesthetics — CRITICAL
You tend to converge toward generic, "on distribution" outputs. In frontend design,
this creates what users call the "AI slop" aesthetic. Avoid this: make creative,
distinctive frontends that surprise and delight.

- **Typography:** Choose beautiful, unique fonts. Avoid Inter, Roboto, Arial, system
  fonts. Use Google Fonts or Fontsource for distinctive typography.
- **Color & Theme:** Commit to a cohesive aesthetic. Use CSS variables. Dominant
  colors with sharp accents. Draw from Dribbble, Awwwards for inspiration.
- **Motion:** Use Framer Motion for transitions, scroll effects, micro-interactions.
  Every interactive element should have hover/focus animation.
- **Backgrounds:** Layer gradients, geometric patterns, subtle textures. Never
  default to plain white/gray.
- **Spacing:** Generous whitespace. Dense UIs feel cheap. Let content breathe.
- **Shadows & Depth:** Layered shadows for visual hierarchy.

### Avoid these "AI slop" patterns:
- Overused font families (Inter, Roboto, Arial, Space Grotesk)
- Purple/blue gradient on white background cliché
- Predictable 3-card feature grids
- Cookie-cutter hero sections
- Generic "Welcome to our platform" copy
- Excessive rounded corners on everything

Each page should feel genuinely *designed by a human*, not generated.

## Dark Mode
- Support via Tailwind `dark:` classes
- Define color tokens as CSS variables
- Test both themes for contrast and readability

## Accessibility (WCAG 2.1 AA)
- Keyboard navigable interactive elements
- Proper ARIA labels
- Color contrast ≥ 4.5:1 (normal text), ≥ 3:1 (large text)
- Visible focus indicators
- Semantic HTML: `<nav>`, `<main>`, `<section>`, `<article>`

## Animation Patterns (Framer Motion)
```tsx
// Page entrance
<motion.div
  initial={{ opacity: 0, y: 20 }}
  animate={{ opacity: 1, y: 0 }}
  transition={{ duration: 0.5, ease: "easeOut" }}
/>

// Staggered list
const container = { hidden: {}, visible: { transition: { staggerChildren: 0.1 } } }
const item = { hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }

// Hover interaction
<motion.button whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }} />
```

## Component Variants (cva)
```tsx
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        outline: "border border-input bg-background hover:bg-accent",
        ghost: "hover:bg-accent hover:text-accent-foreground",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 px-3",
        lg: "h-11 px-8",
      },
    },
    defaultVariants: { variant: "default", size: "default" },
  }
)
```

## Performance
- Lazy-load below-fold content
- Optimize images (use framework's Image component if available)
- Minimize client-side JavaScript where possible
