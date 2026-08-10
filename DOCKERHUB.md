# NaviDock

Lightweight personal NAS homepage / navigation dock.

Group your self-hosted apps, switch internal/external links, customize theme and wallpaper — one Docker container, SQLite included.

## Quick start

**Required on NAS:** set `PUID` and `PGID` to the owner of the bind-mounted data directory.

```bash
id zeven   # example: uid=1026 gid=100
```

```yaml
name: navi-dock

services:
  navi-dock:
    image: zevenz/navi-dock:latest
    container_name: navi-dock
    ports:
      - "7530:7530"
    environment:
      PUID: <your-puid>
      PGID: <your-pgid>
    volumes:
      - /volume1/docker/navi:/data
    restart: unless-stopped
```

**Synology Container Manager:** add environment variables `PUID` and `PGID`, and set **Entrypoint** to `/entrypoint.sh`.

**NAS permission fix (run once on the host, via SSH):**

```bash
mkdir -p /volume1/docker/navi
chown -R zeven:users /volume1/docker/navi
chmod 775 /volume1/docker/navi
id zeven   # use uid/gid for PUID/PGID below
```

If bind mount still fails on Synology, remove any anonymous `/data` volume in Container Manager and mount host path directly: `/volume1/docker/navi:/data`.

```bash
docker compose up -d
```

Open: http://localhost:7530

## Features

- App groups & icons (Iconify / image / text)
- Internal / external / auto network mode
- Light / dark / system theme
- Custom site title, favicon, wallpaper & opacity
- Icon & wallpaper gallery
- Import / export navigation data
- 中文 / English
- SQLite persistence under `/data`

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PUID` | **Yes** | UID of the host user that owns the bind-mounted `/data` directory |
| `PGID` | **Yes** | GID of the host user that owns the bind-mounted `/data` directory |

Find values on the NAS host:

```bash
id zeven
# uid=1026(zeven) gid=100(users)  →  PUID=1026  PGID=100
```

Startup logs will print the configured `PUID` / `PGID`.

## Data

Persist everything with one volume mount:

| Path | Purpose |
|------|---------|
| `/data/database` | SQLite database |
| `/data/assets/icons` | Uploaded icons |
| `/data/assets/wallpapers` | Uploaded wallpapers |

No `privileged` mode needed. Set `PUID` / `PGID` to match the host `./data` owner on NAS.

**Synology:** In Container Manager → Execution command, **Entrypoint must be** `/entrypoint.sh` (not `/app/navidock`). Verify:

```bash
docker inspect zevenz/navi-dock:latest --format 'Entrypoint={{json .Config.Entrypoint}}'
# Expected: Entrypoint=["/entrypoint.sh"]
```

If it shows `/app/navidock`, pull the latest image or rebuild and push.

## Source

https://git.zeven.site/zeven/navi-dock
