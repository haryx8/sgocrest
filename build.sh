#!/bin/bash
env GOOS=linux GOARCH=amd64 go build -v sgocrest.go
# env GOOS=darwin GOARCH=amd64 go build -v sgocrest.go
ls -lah|grep sgocrest
upx sgocrest
ls -lah|grep sgocrest