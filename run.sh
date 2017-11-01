#!/bin/bash

trap 'kill $(jobs -p)' EXIT
trap 'kill $(jobs -p)' SIGINT

BASEDIR=$(dirname "$0")

cd "$BASEDIR/bin"
rm *
go build -o appserver "$BASEDIR/src/appserver/app.go"
go build -o webserver "$BASEDIR/src/webserver/web.go"

cd "$BASEDIR"
bin/appserver | sed "s/^/[appserver] /" & bin/webserver | sed "s/^/[webserver] /"


