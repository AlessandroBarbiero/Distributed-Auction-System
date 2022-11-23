#!/bin/bash

set -euo pipefail

filename="./client/DNS_Cache.info" 

while read line 
do 
  echo "$TERM -e go run server/server.go "$line" &"
   $TERM -e go run server/server.go "$line" &
done < "$filename"
