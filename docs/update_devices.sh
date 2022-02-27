#!/bin/sh

set -euo pipefail

if [ -x "$(command -v apk)" ]; then
  apk add wget gzip postgresql-client
fi

if [ -z "${DATABASE_URL}" ]; then
  echo "no env DATABASE_URL"
  exit 1
fi

psql $DATABASE_URL -f devices.sql
