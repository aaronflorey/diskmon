#!/bin/sh
set -e

if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
  case "$1" in
    0|remove)
      systemctl disable --now diskmon.service || true
      ;;
    *)
      ;;
  esac
fi
