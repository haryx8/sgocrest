#!/bin/bash
curl -v http://localhost:1234/read/file -F "file=@$1"
