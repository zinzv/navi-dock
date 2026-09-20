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

**NAS / bind mount:** `PUID` / `PGID` default to `1000`; override in `.env` if needed to match the host data directory owner (`id <user>`).

Open: http://localhost:7530

### Build with a version

`APP_VERSION` is passed into the image as the Docker build arg `VERSION` (baked into the binary). If unset, the Dockerfile falls back to the `VERSION` file.

PowerShell:

```powershell
$env:APP_VERSION = "v0.1.0"
docker compose build
```

bash:

```bash
APP_VERSION=v0.1.0 docker compose build
```

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

## Automatic network detection

Automatic access mode requires an explicit environment variable:

```env
LAN_PROBE_DOMAIN=lan.zeven.site
```

This must be set in the deployment environment (for example `.env` or Docker Compose
`environment`). It cannot be configured from the settings UI.

Only the domain is configurable. The probe path is fixed to `/api/network/ping`, and
the browser uses the same HTTP/HTTPS protocol as the current NaviDock page with a
built-in 1000ms timeout.

Configure split DNS so the probe hostname resolves to this NaviDock instance only
inside the LAN. The probe endpoint is public, returns `204 No Content`, and disables
caching. When NaviDock uses HTTPS, the probe hostname also needs a valid certificate.

How modes work:

- **Auto**: stay in auto mode; probe result only chooses internal vs external URL
- **Internal / External**: force the corresponding URL regardless of probe result
