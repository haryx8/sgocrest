#!/bin/bash
curl -v http://localhost:1234/read -F "image=@$1"
