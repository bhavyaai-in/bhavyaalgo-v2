# Project Rules

## Code Style
- No comments in code unless absolutely necessary
- Capitalize first letter of each word in all user-facing text (names, labels, titles)
- All text must be properly cased

## Frontend
- Modal components go in `frontend/src/modals/brokers/`
- Each modal takes `show: Boolean` and `broker: Object` props, emits `close`
- Page components for full tabs go in `frontend/src/views/`
- Reusable display components use `PageDisplay` suffix (e.g. `OrdersPageDisplay.vue`)
- Modal components use `Modal` suffix (e.g. `OrdersModal.vue`)
- Broker card action buttons use chip-style grid (4 columns × 2 rows, icon + label)

## Backend
- Blueprint handlers in `backend/blueprints/` with `Register*Routes(mux)` pattern
- Broker-specific implementations in `backend/angel/`, `backend/brokers/`
- API errors from external providers (Angel One) must be checked via `errorcode` field, not just HTTP status
- `.env` excluded from git — credentials via env vars or `ADMIN_EMAIL`/`ADMIN_PASSWORD`

## Data Flow
- Broker card buttons → dedicated Modal component → fetches own data via API → displays in popup
- Dashboard tabs → PageDisplay component → fetches data on mount/watch → displays full page
- Auth via JWT token in `Authorization` header, stored in `localStorage`
