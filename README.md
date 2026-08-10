# NaviDock

Personal NAS Dashboard · Docker one-click deploy

## Layout

```
navi-dock/
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── README.md
├── backend/
├── frontend/
├── migrations/
├── configs/
├── scripts/
└── docs/
```

## Quick start

```bash
cp .env.example .env
docker compose up -d --build
```

**NAS / bind mount:** edit `.env` and set `PUID` / `PGID` to match the host data directory owner (`id <user>`).

Open: http://localhost:7530

## Features (MVP)

- Home dashboard with service icons
- Settings page (中文 / English)
- Custom browser tab title
- Theme: light / dark / system (header switcher)
- Network mode switcher
- Custom home background + opacity
- SQLite settings persistence

## Local development

### Backend

```bash
./scripts/dev-backend.sh
# Windows: .\scripts\dev-backend.ps1
```

### Frontend

```bash
./scripts/dev-frontend.sh
# Windows: .\scripts\dev-frontend.ps1
```

Vite proxies `/api` to `http://127.0.0.1:7530`.
