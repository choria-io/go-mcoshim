#!/bin/bash

if [ "$3" != "--parse-compound" ]
then
    echo '{"statuscode": 1, "statusmsg": "missing --parse-compound"}'
    exit 1
fi

read request

if [ "$request" != "systemd=true and staging_http_get=crl" ]
then
    echo '{"statuscode": 1, "statusmsg": "incorrect stdin"}'
    exit 1
fi

echo '[{"statement":"systemd=true"},{"and":"and"},{"statement":"staging_http_get=crl"}]'

exit 0
