#!/bin/bash

# Script to test the API Endpoints

METHOD=$1
ID=$2

URL="http://localhost:3000/account"

if [ $METHOD == "help" ]; then
    echo "./test.sh <METHOD> <ID?>"
fi

# Get all contacts
if [[ $METHOD == 'get' && -z $ID ]]; then
    curl -X GET $URL
fi

# Get contact with ID
if [[ $METHOD == 'get' && -n $ID ]]; then
    if [ -z $ID ]; then 
        echo "ID is requried"
        exit 1
    fi

    curl -X GET $URL/$ID
fi

# Create a new contact
if [ $METHOD == 'post1' ]; then
    curl -X POST $URL \
        -H "Content-Type: application/json" \
        -d '{
            "firstName": "John",
            "lastName": "Doe"
        }'
fi

if [ $METHOD == 'post2' ]; then
    curl -X POST $URL \
        -H "Content-Type: application/json" \
        -d '{
            "firstName": "Jane",
            "lastName": "Hill"
        }'
fi

# Delete a contact
if [ $METHOD == 'delete' ]; then
    if [ -z $ID ]; then 
        echo "ID is requried"
        exit 1
    fi

    curl -X DELETE $URL/$ID
fi

exit 0