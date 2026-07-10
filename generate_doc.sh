#!/usr/bin/env bash

go test -v ./...
godoc -http=:6060 -play
open "http://127.0.0.1:6060"