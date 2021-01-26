#!/bin/bash
defaults write org.python.python ApplePersistenceIgnoreState NO
# cd /Users/hari/Documents/Projects/gocr
/usr/local/opt/python\@3.8/bin/python3 ocv.py -i $1
defaults write org.python.python ApplePersistenceIgnoreState YES
