#!/bin/sh

DATABUNKER_APIKEY=$1
if [ -z $DATABUNKER_APIKEY ]; then
  echo "missing api key parameter"
  exit
fi

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
echo "User token is $TOKEN"

RESULT=`curl -s http://localhost:3000/v1/userapp/token/$TOKEN/shipping \
  -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json" \
  -d '{"country":"Israel","address":"Allenby 1","postcode":"12345","status":"active"}' | jq ".status" | tr -d '"'`
echo "User shipping record created, status $RESULT"

RESULT=`curl -s http://localhost:3000/v1/userapp/token/$TOKEN \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
echo "View list of all user apps $RESULT"

RESULT=`curl -s http://localhost:3000/v1/userapps \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
echo "View list of all apps $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/token/$TOKEN/send-sms -XPOST \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
echo "Enable consent send-sms for user by token: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/email/test@paranoidguy.com/send-sms2 -XPOST \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
echo "Enable consent send-sms for user by email: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/phone/4444/send-sms2 -XDELETE \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
echo "Disabke consent send-sms for user by phone: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/token/$TOKEN/send-sms \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
echo "View this specific consent only: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/token/$TOKEN \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
echo "View all user consents: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consents/send-sms \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
echo "View all users with send-sms consent on: $RESULT"
