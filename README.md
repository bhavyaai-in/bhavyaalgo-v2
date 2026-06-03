# BhavyaAI Algo Trading Platform

Algorithmic trading platform with Angel One broker integration. Live market data streaming, watchlist management, and order placement.

## Tech Stack

- **Frontend**: Vue 3 (Vite), Pinia
- **Backend**: Go 1.23+
- **Database**: SQLite (with WAL mode)
- **Market Data**: Angel One Smart API WebSocket
- **Broker**: Angel One (via proxy)

## Project Structure

```
├── frontend/                # Vue 3 SPA
│   └── src/
│       ├── components/      # UI components (Watchlist, Orders, OrderPad)
│       ├── composables/     # WebSocket, broker data composables
│       ├── stores/          # Pinia stores (auth, broker data)
│       ├── views/           # Dashboard, Brokers, Login pages
│       └── utils/           # API utility
├── backend/                 # Go server
│   ├── blueprints/          # Route handlers (watchlist, broker, orders)
│   ├── brokers/angel/       # Angel One API client
│   ├── db/                  # SQLite queries & generated code
│   ├── internal/            # Config, session store, seed data
│   └── ws/                  # WebSocket hub for live ticks
├── run_dev.py               # Dev runner (starts both frontend & backend)
└── test_order.py            # Order placement test script
```

## Getting Started

### Prerequisites

- Go 1.23+
- Node.js 20+
- SQLite3

### Setup

```bash
# Backend
cd backend
cp .env.example .env    # configure ADMIN_EMAIL / ADMIN_PASSWORD
cp seed.sample.json seed.json   # configure your broker credentials
go mod tidy
go run .                # starts on :8080

# Frontend (separate terminal)
cd frontend
npm install
npm run dev             # starts on :5173

# Or run both together:
python run_dev.py
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ADMIN_EMAIL` | Login email |
| `ADMIN_PASSWORD` | Login password |

### Login

Credentials are set via `ADMIN_EMAIL` / `ADMIN_PASSWORD` in `.env`.

## Features

- **Watchlist**: Create/manage symbol watchlists with drag-drop reorder
- **Live Ticks**: Real-time LTP via Angel One Smart API WebSocket
- **Order Pad**: Buy/Sell order placement with product type & quantity
- **Order Book**: View order status with rejection reason tooltips
- **Broker Connect**: Angel One authentication flow

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/login` | User login |
| GET/POST | `/api/brokers` | List/Create brokers |
| POST | `/api/connect-broker` | Connect broker |
| POST | `/api/broker-place-order` | Place order |
| GET | `/api/watchlists` | List watchlists |
| POST | `/api/watchlists/{id}/items` | Add symbol to watchlist |

## Broker

Angel One API is accessed through a proxy (`clipx.bhavyaai.com/rqfarward`).  
Broker-specific constants are defined in `backend/brokers/angel/constants.go`.
