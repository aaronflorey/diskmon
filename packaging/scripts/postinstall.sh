#!/bin/sh
set -e

install -d -m 0755 /var/lib/diskmon
install -d -m 0755 /etc/diskmon

if command -v systemd-tmpfiles >/dev/null 2>&1; then
  systemd-tmpfiles --create /usr/lib/tmpfiles.d/diskmon.conf >/dev/null 2>&1 || true
fi

if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
  systemctl daemon-reload || true

  case "$1" in
    1|configure)
      systemctl enable --now diskmon.service || true
      ;;
    *)
      systemctl try-restart diskmon.service || true
      ;;
  esac
fi
