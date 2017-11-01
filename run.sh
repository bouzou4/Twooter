#!/bin/bash

trap 'kill $(jobs -p)' EXIT
trap 'kill $(jobs -p)' SIGINT

BASEDIR=$(dirname "$0")
cd "$BASEDIR"
bin/appserver | sed "s/^/[appserver] /" & bin/webserver | sed "s/^/[webserver] /"


