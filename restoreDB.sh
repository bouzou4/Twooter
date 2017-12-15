#!/bin/bash

BASEDIR=$(dirname "$0")

cd "$BASEDIR"

rm -rf Data
cp -rf DataBKUP "$BASEDIR/Data"