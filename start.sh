#!/bin/sh

: ${URL:=https://gitlab.com/api/v3/}
: ${TOKEN:=}

: ${INTERVAL:=5}
: ${PORT:=7070}

: ${VERIFY_SSL:=1}

ADDR=":$PORT"

exec app -addr="$ADDR" -url="$URL" -token="$TOKEN" -verify-ssl=$VERIFY_SSL -interval="$INTERVAL"
