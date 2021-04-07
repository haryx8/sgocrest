#!/bin/bash
curl -v http://localhost:1234/read/image -F "image=@$1"
