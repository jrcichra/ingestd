#!/bin/bash
curl -i --header "Content-Type: application/json" \
  --request POST \
  --data "$1" \
  http://localhost:8080/blah/tab
