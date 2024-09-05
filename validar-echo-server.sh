#!/usr/bin/env bash
set -eu

: ${SERVER_PORT=}

if [ -z "${SERVER_PORT}" ]
then
    SERVER_PORT=12345
fi

MESSAGE="echo server: this should be the same"

RECEIVED=$(printf "%s" ${MESSAGE} | nc 0.0.0.0 ${SERVER_PORT} | tr -d '\n')
if [ "${RECEIVED}" -ne "${MESSAGE}" ]
then
    echo "action: test_echo_server | result: fail"
    exit 1
fi
echo "action: test_echo_server | result: success"
exit 0
