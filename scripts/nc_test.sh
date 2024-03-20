#!/usr/bin/env bash

if [ -z "${SERVER_PORT}" ]
then
    SERVER_PORT=12345
fi

TEST_MESSAGES=("echo" "server" "this" "should" "be" "the" "same")

for m in ${TEST_MESSAGES[@]}; do
    RECEIVED=$(printf "%s" "$m" | nc server ${SERVER_PORT} | tr -d '\n')
    NETCAT_EXIT_CODE=$?
    if [ $NETCAT_EXIT_CODE -ne 0 ]
    then
        printf "An error occurred while reaching the server, exit code: %d\n" $NETCAT_EXIT_CODE
        exit $NETCAT_EXIT_CODE
    fi
    if [ "${RECEIVED}" != "$m" ]
    then
        printf "Expected %s got %s\n" "$m" "$RECEIVED"
        exit 1
    fi
    echo "sent: \"${m}\", received: \"${RECEIVED}\""        
done

echo "all tests passed"

exit 0
