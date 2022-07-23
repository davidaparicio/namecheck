#!/usr/bin/env bash

FuzzFUNC="Fuzz" #"FuzzReverse"

echo "Let's Test"
go test -v ./... -coverprofile=coverage.out

echo "Let's Fuzz"
go test -fuzz ${FuzzFUNC} -fuzztime 15s

echo "Let's Bench"
go test -v -run=^$ -bench . -benchmem -benchtime=3s ./