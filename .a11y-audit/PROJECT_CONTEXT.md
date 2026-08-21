# Accessibility audit context

Project: Agentlink
Base URL: https://agentlink.run/
Local artifact URL: http://127.0.0.1:8088/
Routes: `/`
Standards: WCAG 2.1 AA
Output mode: markdown+json

The canonical automated gate is `.github/workflows/a11y.yml`. It serves the tracked `docs/` artifact and fails on every axe-core violation. Baseline suppression is intentionally disabled because the single page is expected to pass with zero automated violations.

Manual review remains required for keyboard order, visible focus, screen-reader reading order, zoom and reflow, reduced motion, and timing behavior that axe-core cannot establish.
