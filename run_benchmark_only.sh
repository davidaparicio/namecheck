#!/usr/bin/env bash

echo "Let's Bench"
go test -v -run=^$ -bench . -benchmem -benchtime=15s ./