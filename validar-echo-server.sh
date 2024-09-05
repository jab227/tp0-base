#!/usr/bin/env bash
set -eu

: ${SERVER_PORT=}

if [ -z "${SERVER_PORT}" ]
then
    SERVER_PORT=12345
fi


NETWORK="tp0_testing_net"
MESSAGE="echo server: this should be the same"

COMMAND=$(printf "%s" ${MESSAGE} | nc server ${SERVER_PORT})
RECEIVED=$(docker run --rm --network ${NETWORK} alpine sh -c ${COMMAND})

if [ "${RECEIVED}" -ne "${MESSAGE}" ]
then
    echo "action: test_echo_server | result: fail"
    exit 1
fi
echo "action: test_echo_server | result: success"
exit 0
