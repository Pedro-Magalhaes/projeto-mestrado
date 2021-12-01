#! /bin/bash
cmd="go run cmd/consumer/main.go"
go version
echo starting monitor running: $cmd
$cmd
