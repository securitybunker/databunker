#!/bin/sh

# check if Databunker is using SSL port
if [ -n "$SSL_CERTIFICATE" ]; then
  if [ -f "$SSL_CERTIFICATE" ]; then
    /bin/busybox wget --tries=1 --spider --no-check-certificate -q https://localhost:3000/status || exit 1
    exit 0
  fi
fi

if [ -f "/databunker/certs/server.cer" ]; then
  /bin/busybox wget --tries=1 --spider --no-check-certificate -q https://localhost:3000/status || exit 1
  exit 0
fi

CERT_FILE=`/bin/busybox sed -n -e '/ssl_certificate:/ s/ .*ssl_certificate: "\(.*\)"/\1/p' /databunker/conf/databunker.yaml`
if [ -n "$CERT_FILE" ]; then
  if [ -f "$CERT_FILE"  ]; then
    /bin/busybox wget --tries=1 --spider --no-check-certificate -q https://localhost:3000/status || exit 1
    exit 0
  fi
fi

# by default check http port
/bin/busybox wget --tries=1 --spider -q http://localhost:3000/status || exit 1
