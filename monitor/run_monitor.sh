#! /bin/bash
cmd="go run cmd/monitor/main.go"
go version
echo starting monitor running: $cmd
$cmd
