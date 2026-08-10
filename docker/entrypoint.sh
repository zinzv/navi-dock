#!/bin/sh
set -e

# Common typo on NAS panels
if [ -z "${PGID}" ] && [ -n "${PGUID}" ]; then
  PGID="${PGUID}"
fi

if [ -z "${PUID}" ] || [ -z "${PGID}" ]; then
  echo "navi-dock: PUID and PGID are required." >&2
  echo "Set them to the owner of the bind-mounted /data directory (see: id <user>)." >&2
  exit 1
fi

echo "----------------------------------------"
echo "NaviDock"
echo "User UID:    ${PUID}"
echo "User GID:    ${PGID}"
echo "Entrypoint:  /entrypoint.sh"
echo "Data path:   /data"
echo "----------------------------------------"

ensure_data_dirs() {
  mkdir -p /data/database /data/assets/icons /data/assets/wallpapers
}

fix_ownership() {
  if [ "$(id -u)" != "0" ]; then
    return 0
  fi
  if [ "${PUID}" = "0" ] && [ "${PGID}" = "0" ]; then
    return 0
  fi
  chown -R "${PUID}:${PGID}" /data 2>/dev/null || true
}

require_writable_data() {
  if [ -w /data ]; then
    return 0
  fi
  echo "navi-dock: /data is not writable (uid=$(id -u) gid=$(id -g))." >&2
  echo "Check PUID/PGID and host directory permissions." >&2
  exit 1
}

ensure_data_dirs

if [ "$(id -u)" = "0" ]; then
  fix_ownership
  if [ "${PUID}" = "0" ] && [ "${PGID}" = "0" ]; then
    require_writable_data
    exec /app/navidock "$@"
  fi
  require_writable_data
  exec su-exec "${PUID}:${PGID}" /app/navidock "$@"
fi

require_writable_data
exec /app/navidock "$@"
