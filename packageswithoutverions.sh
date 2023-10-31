#!/bin/bash

echo "Packages with just master versions"
export MASTER="$(cat ./index/*/*.json | grep -o \"versions\":\\\[\"master\"\\\] | wc -l)"
export MAIN="$(cat ./index/*/*.json | grep -o \"versions\":\\\[\"main\"\\\] | wc -l)"

echo $(($MAIN + $MASTER))

