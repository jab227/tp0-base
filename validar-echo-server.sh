#!/usr/bin/env bash
set -eu

: ${SERVER_PORT=}

if [ -z "${SERVER_PORT}" ]
then
    SERVER_PORT=12345
fi

TEST_MESSAGES=("echo" "server" "this" "should" "be" "the" "same")

for m in ${TEST_MESSAGES[@]}; do
    RECEIVED=$(printf "%s" "$m" | nc server ${SERVER_PORT} | tr -d '\n')
    if [ "${RECEIVED}" != "$m" ]
    then
        echo "action: test_echo_server | result: fail"
        exit 1
    fi
done
echo "action: test_echo_server | result: success"
exit 0
