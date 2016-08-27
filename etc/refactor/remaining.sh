#!/bin/bash

if [ -z $1 ]
then
    echo "Usage: remainder.sh <path>"
    exit 1
fi

grep -R "func Test" $1 | grep -v vendor | grep -v "func TestMain" | grep -v "RF("
