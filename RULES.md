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

## Seed / First-Time Setup
- Seed data lives in `backend/seed.json` as the single source of truth
- Each table has a top-level key (e.g. `broker_list`, `broker_columns`) containing an array of records
- To add seed data for a new table:
  1. Add `CREATE TABLE IF NOT EXISTS` SQL constant in `main.go`
  2. Add sqlc queries in `db/query.sql`
  3. Run `sqlc generate` from `backend/`
  4. Add the key + records to `seed.json`
  5. Parse the new key in `seedFromFile()` and call the insert query
- Seed only runs if the table is empty (first-time setup only)

## UI / UX
- All API calls go through `utils/api.js` which automatically manages the global loader
- Loader is a thin animated bar at the top of the page (handled by `AppLoader.vue` + `ui` store)
- Custom confirm dialog via `import { confirm } from '../utils/api.js'` — returns a boolean Promise, renders `ConfirmModal.vue`
- Never use browser native `confirm()` — always use the custom one

## Data Flow
- Broker card buttons → dedicated Modal component → fetches own data via API → displays in popup
- Dashboard tabs → PageDisplay component → fetches data on mount/watch → displays full page
- Auth via JWT token in `Authorization` header, stored in `localStorage`
