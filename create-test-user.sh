#!/bin/sh

XTOKEN=$1
if [ -z $XTOKEN ]; then
  echo "missing api key parameter"
  exit
fi

RESULT=`curl -s http://localhost:3000/v1/consent/email/test@paranoidguy.com/contract-approval -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Create anonymous consent by email 1: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/email/test@paranoidguy.com/contract-approval -XPOST \
   -H "X-Bunker-Token: $XTOKEN"` \
   -d '{"lawfulbasis":"contract"}'
echo "Create anonymous consent by email 2: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/phone/4444/contract-approval -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Create anonymous consent by phone 1: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/phone/4444/contract-approval -XPOST \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "Create anonymous consent by phone 2: $RESULT"

echo "Creating user."
RESULT=`curl -s http://localhost:3000/v1/user \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"fname":"Test","lname":"Account","email":"test@paranoidguy.com","phone":"4444"}'`
STATUS=`echo $RESULT | jq ".status" | tr -d '"'`
if [ "$STATUS" = "error" ]; then
  echo "Error to create user, trying to lookup by email."
  RESULT=`curl -s http://localhost:3000/v1/user/email/test@paranoidguy.com -H "X-Bunker-Token: $XTOKEN"`
  STATUS=`echo $RESULT | jq ".status" | tr -d '"'`
fi
if [ "$STATUS" = "error" ]; then
  echo "Failed to fetch user by email, got: $RESULT"
  exit
fi

TOKEN=`echo $RESULT | jq ".token" | tr -d '"'`
echo "User token is $TOKEN"

RESULT=`curl -s http://localhost:3000/v1/userapp/token/$TOKEN/shipping \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"country":"UK","city":"London","address":"221B Baker Street","postcode":"12345","status":"new"}'`
echo "User shipping record created, status $RESULT"

RESULT=`curl -s http://localhost:3000/v1/userapp/token/$TOKEN/shipping -XPUT \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"status":"delivered"}'`
echo "User shipping record updated, status $RESULT"

RESULT=`curl -s http://localhost:3000/v1/userapp/token/$TOKEN/shipping \
  -H "X-Bunker-Token: $XTOKEN"`
echo "User shipping record ready, status $RESULT"


RESULT=`curl -s http://localhost:3000/v1/sharedrecord/token/$TOKEN \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"app":"shipping","fields":"address"}'`
echo "Shared record created, status $RESULT"
RECORD=`echo $RESULT | jq ".record" | tr -d '"'`
echo $RECORD

RESULT=`curl -s http://localhost:3000/v1/get/$RECORD`
echo "Get shared record (no password/access token): $RESULT"

RESULT=`curl -s http://localhost:3000/v1/userapp/token/$TOKEN \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "View list of all user apps $RESULT"

RESULT=`curl -s http://localhost:3000/v1/userapps \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "View list of all apps $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/token/$TOKEN/send-sms -XPOST \
   -H "X-Bunker-Token: $XTOKEN" -d "expiration=30s"`
echo "Enable consent send-sms for user by token: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/email/test@paranoidguy.com/send-sms2 -XPOST \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Enable consent send-sms for user by email: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/phone/4444/send-sms2 -XDELETE \
   -H "X-Bunker-Token: $XTOKEN"`
echo "Withdraw consent send-sms for user by phone: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/token/$TOKEN/send-sms \
   -H "X-Bunker-Token: $XTOKEN"`
echo "View this specific consent only: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consent/token/$TOKEN \
   -H "X-Bunker-Token: $XTOKEN"`
echo "View all user consents: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/consents/send-sms \
   -H "X-Bunker-Token: $XTOKEN"`
echo "View all users with send-sms consent on: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/session/token/$TOKEN -XPOST \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
   -d '{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2"}'`
echo "Create session 1: $RESULT"

SESSION=`echo $RESULT | jq ".session" | tr -d '"'`
echo $SESSION

RESULT=`curl -s http://localhost:3000/v1/session/email/test@paranoidguy.com -XPOST \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
   -d '{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2","info":"email"}'`
echo "Create session 2: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/session/session/$SESSION \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "Get session 1: $RESULT"

RESULT=`curl -s http://localhost:3000/v1/session/phone/4444 \
   -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json"`
echo "Get sessions: $RESULT"
