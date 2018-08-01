#!/bin/bash

if [ "$3" == "--validate-compound" ]
then
    read request

    if [ "$request" != '[{"statement":"systemd=true"},{"and":"and"},{"statement":"staging_http_get=crl"}]' || "$request" != '[{"statement":"systemd=false"},{"and":"and"},{"statement":"staging_http_get=crl"}]']
    then
        echo '{"matched": false}'
        exit 1
    fi


    if [ "$request" == '[{"statement":"systemd=true"},{"and":"and"},{"statement":"staging_http_get=crl"}]' ]
    then
        echo '{"matched": true}'
    else
        echo '{"matched": false}'
    fi

    exit 0
fi

if [ "$3" == "--parse-compound" ]
then
    read request

    if [ "$request" == "systemd=true and staging_http_get=crl" ]
    then
        echo '[{"statement":"systemd=true"},{"and":"and"},{"statement":"staging_http_get=crl"}]'
        exit 0
    fi

    if [ "$request" == "systemd=false and staging_http_get=crl" ]
    then
        echo '[{"statement":"systemd=false"},{"and":"and"},{"statement":"staging_http_get=crl"}]'
        exit 0
    fi

    echo '{"statuscode": 1, "statusmsg": "incorrect stdin"}'
    exit 1
fi

