#!/usr/bin/env sh

if [ -z "${SERVER_PORT}" ]
then
    SERVER_PORT=12345
fi


NETWORK="tp0_testing_net"
MESSAGE="echo server: this should be the same"

RECEIVED=$(docker run --rm --network tp0_testing_net alpine sh -c "echo ${MESSAGE} | nc server:${SERVER_PORT}")

if [ "${RECEIVED}" != "${MESSAGE}" ]
then
    echo "action: test_echo_server | result: fail"
else
    echo "action: test_echo_server | result: success"
fi
