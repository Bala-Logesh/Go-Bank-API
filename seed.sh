#!/bin/bash

API_URL="http://localhost:3000/account"

JSON_DATA1='{
        "firstName": "John",
        "lastName": "Doe"
    }'

JSON_DATA2='{
        "firstName": "Jane",
        "lastName": "Hill"
    }'

curl -X POST "$API_URL" \
     -H "Content-Type: application/json" \
     -d "$JSON_DATA1"

curl -X POST "$API_URL" \
     -H "Content-Type: application/json" \
     -d "$JSON_DATA2"