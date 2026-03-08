# Frontend Rules

## Tech Stack
- **Styling:** Tailwind CSS v4 (`@tailwindcss/vite` plugin, no PostCSS config needed)
- **Design system:** Custom `@apply`-based component classes in `index.css`
- **Utilities:** `cn()` from `@/lib/utils` (clsx + tailwind-merge) for conditional classes
- **Path alias:** `@/*` maps to `src/*` (configured in vite.config.ts + tsconfig.app.json)

## Design Language
Clean, light, angular. Monochrome, flat, editorial. No gradients, no shadows, no border-radius.

- **Font:** DM Sans (Google Fonts). Bold uppercase for labels/buttons.
- **Colors:** White bg, black text (#2d2d2d), gray secondary (#767676), green accent (#018849), red error (#d4351c). All defined as `@theme` tokens in `index.css`.
- **Spacing:** Generous. Narrow content columns (~480px player forms, ~600px game, ~960px admin).

## Component Classes (defined in `index.css`)
Use these instead of writing long Tailwind class strings for common patterns:

| Class | Purpose |
|-------|---------|
| `.page` / `.page-md` / `.page-wide` | Container widths (480/600/960px) |
| `.btn` | Black filled button, uppercase |
| `.btn.btn-accent` | Green filled button — use sparingly for key CTAs (Join, Start, Submit) |
| `.btn-secondary` | White + black border |
| `.btn-ghost` | Transparent, subtle |
| `.btn-danger` | Red border, red fill on hover |
| `.btn-sm` | Smaller padding/font |
| `.input` | Text input with thin border |
| `.input-label` | Uppercase label above input |
| `.card` / `.card-header` | Bordered container, no shadow |
| `.admin-table` | Table with uppercase headers |
| `.text-feedback-error` / `.text-feedback-success` | Colored feedback text |
| `.spinner` / `.spinner-lg` | CSS-only loading spinner |

## Shared Components (`src/components/`)
- `Spinner.tsx` — `<Spinner />` and `<LoadingPage />` (replaces `aria-busy` patterns)
- `ErrorMessage.tsx` — `<ErrorMessage message="..." />`
- `PageContainer.tsx` — `<PageContainer size="sm|md|wide">` (replaces `className="container" style={{ maxWidth }}`)

## Coding Standards
- No inline `style={{}}` — use Tailwind utilities or component classes
- No Pico.css patterns (`aria-busy` on elements, `<hgroup>`, `<mark>`, `className="container"`, `className="grid"`, `className="outline"`)
- Proper label-input pattern: `<label className="input-label">` above `<input className="input">`
- Use `<Spinner />` inside buttons for loading states, not text changes
- Prefer component classes for repeated patterns, Tailwind utilities for one-offs

## Responsive Design
- Mobile-first: base styles for mobile, `sm:` / `md:` for larger
- Grid: `grid grid-cols-1 sm:grid-cols-2 gap-4` (not Pico's `className="grid"`)

## Accessibility
- Keyboard navigable interactive elements
- `role="alert"` on error messages
- Semantic HTML: `<nav>`, `<main>`, `<details>`, `<summary>`
- Visible focus: inputs get darker border on focus
