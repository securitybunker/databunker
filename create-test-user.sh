#!/bin/bash

DATABUNKER_APIKEY='1ca9a727-ee10-9cc6-c1a2-72da05951e69'

echo "Creating user."
RESULT=`curl -s http://localhost:3000/v1/user \
  -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json" \
  -d '{"fname":"Test","lname":"Account","email":"test@paranoidguy.com","phone":"4444"}'`
STATUS=`echo $RESULT | jq ".status" | tr -d '"'`
if [ "$STATUS" == "error" ]; then
  echo "Error to create user, trying to lookup by email."
  RESULT=`curl -s http://localhost:3000/v1/user/email/test@paranoidguy.com -H "X-Bunker-Token: "$DATABUNKER_APIKEY`
  STATUS=`echo $RESULT | jq ".status" | tr -d '"'`
fi
if [ "$STATUS" == "error" ]; then
  echo "Failed to fetch user by email, got: $RESULT"
  exit
fi

TOKEN=`echo $RESULT | jq ".token" | tr -d '"'`
echo "User token is  $TOKEN"

RESULT=`curl -s http://localhost:3000/v1/userapp/token/$TOKEN/shiping \
  -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json" \
  -d '{"country":"Israel","address":"Allenby 1","postcode":"12345","status":"active"}' | jq ".status" | tr -d '"'`
echo "User shiping record created, status $RESULT"