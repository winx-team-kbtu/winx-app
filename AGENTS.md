
## Agent Skills

Use installed skills when relevant to the task. Read and follow
the selected skill's SKILL.md instructions before using it.
Do not invoke all skills for every task.

### Skill Selection

- `design-taste-frontend`
  Use for major frontend design tasks, landing pages, visual
  redesigns, and improvements to visual quality.
  Preserve the existing ReRead design system and branding.
  Do not redesign unrelated components.

- `web-design-guidelines`
  Use for UI/UX reviews, accessibility audits, responsive
  design checks, and design best-practice reviews.
  After significant frontend changes, review affected UI
  files against these guidelines.

- `image-to-code`
  Use when the task explicitly requires an image-first
  design workflow: generate visual references, analyze
  them, and implement the matching frontend.
  Requires image-generation capabilities.
  Do not invoke for ordinary frontend bug fixes.

- `playwright-cli`
  Use for browser-based testing, E2E checks, visual
  verification, and responsive testing.
  Prefer local or staging environments.
  Never perform destructive actions or modify production
  data without explicit permission.

### Frontend Workflow

For significant UI changes:
1. Inspect existing components and design conventions.
2. Apply `design-taste-frontend` when visual design is in scope.
3. Implement changes using the existing frontend stack.
4. Review affected UI with `web-design-guidelines`.
5. Verify behavior using `playwright-cli` when available.
6. Report which skills were used and what was tested.

### ReRead Design Constraints

- Maintain the existing brand identity and visual consistency.
- Prioritize mobile-first responsive design.
- Preserve Russian and Kazakh localization.
- Maintain existing functionality and API contracts.
- Avoid generic AI-generated visual patterns.
- Do not introduce unnecessary dependencies or animations.
- Never sacrifice accessibility or performance for aesthetics.

### General Rules

- Use skills selectively based on task relevance.
- Do not invoke design skills for backend-only tasks.
- Explicit user instructions take priority over skill preferences.
- Do not claim a test passed unless it was actually executed.
