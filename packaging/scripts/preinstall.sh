#!/bin/sh
set -e

if ! getent group diskmon >/dev/null 2>&1; then
  if command -v groupadd >/dev/null 2>&1; then
    groupadd --system diskmon >/dev/null 2>&1 || true
  elif command -v addgroup >/dev/null 2>&1; then
    addgroup --system diskmon >/dev/null 2>&1 || true
  fi
fi

if ! id -u diskmon >/dev/null 2>&1; then
  if command -v useradd >/dev/null 2>&1; then
    useradd --system --gid diskmon --home-dir /var/lib/diskmon --no-create-home --shell /usr/sbin/nologin diskmon >/dev/null 2>&1 || true
  elif command -v adduser >/dev/null 2>&1; then
    adduser --system --ingroup diskmon --home /var/lib/diskmon --no-create-home --shell /usr/sbin/nologin diskmon >/dev/null 2>&1 || true
  fi
fi
