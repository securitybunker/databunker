# Data Bunker API


**ParanoidGuy Databunker is a special GDPR compliant database for personal & private information.**


Latest API is available On Postman:

https://documenter.getpostman.com/view/11310294/Szmcbz32

---

## User Api

User API is used to store, change and retrieve user personal information out of Databunker.

| Resource / HTTP method | POST (create)    | GET (read)    | PUT (update)     | DELETE (delete)  |
| ---------------------- | ---------------- | ------------- | ---------------- | ---------------- |
| /v1/user               | Create new user  | Error         | Error            | Error            |
| /v1/user/token/{token} | Error            | Get user      | Update user      | Delete user PII  |
| /v1/user/login/{login} | Error            | Get user      | Update user      | Delete user PII  |
| /v1/user/email/{email} | Error            | Get user      | Update user      | Delete user PII  |
| /v1/user/phone/{phone} | Error            | Get user      | Update user      | Delete user PII  |


## Create user record
### `POST /v1/user`

### Explanation

This API is used to create new user record. If the request is successful it returns new user `{token}`.
On the database level, each records is encrypted with it's own key.


### POST Body Format

POST Body can contain regular form data or JSON. Data Bunker extracts `{login}`, `{phone}` and `{email}` out of
POST data or from JSON root level and builds additional hashed indexes for user record.

The `{login}`, `{phone}`, `{email}` values must be unique, otherwise you will get a duplicate user record error.
So, for example, you can not create two user records with the same email address.

The following content type supported:

* **application/json**
* **application/x-www-form-urlencoded**


### Example:

Create user by posting data in JSON format:

```
curl -s http://localhost:3000/v1/user -XPOST \
  -H "X-Bunker-Token: $XTOKEN" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "John","lastName":"Doe","email":"user@gmail.com"}'
{"status":"ok","token":"db80789b-0ad7-0690-035a-fd2c42531e87"}
```

Create user record by posting key/value fiels as POST parameters:

```
curl -s http://localhost:3000/v1/user -XPOST \
  -H "X-Bunker-Token: $XTOKEN" \
  -d 'firstName=John' \
  -d 'lastName=Doe' \
  -d 'email=user2@gmail.com'
{"status":"ok","token":"db80789b-0ad7-0690-035a-fd2c42531e87"}
```

**NOTE**: Keep this user `{token}` privately as it is an additional user identifier.

For work with semi-trusted environments or 3rd party companies, use **shareable record** instead of 
user `{token}`. Shareable `{record}` is time bounded.


## Get user record
### `GET /v1/user/{token,login,email,phone}/{identity}`

### Explanation
This API is used to get user record stored in Databunker. You can lookup user record by `{token}`,
`{email}`, `{phone}` or `{login}` values.

### Example:

Fetch user record by `{token}` value:

```
curl --header "X-Bunker-Token: $XTOKEN" -XGET \
   https://localhost:3000/v1/user/token/DAD2474A-E9A7-4BA7-BFC2-C4506880198E
{"status":"ok","token":"DAD2474A-E9A7-4BA7-BFC2-C4506880198E",
"data":{"fname":"paranoid","lname":"guy","login":"user1123"}}
```

Fetch user record by `{login}` name:

```
curl --header "X-Bunker-Token: $XTOKEN" -XGET \
   https://localhost:3000/v1/user/login/user1123
{"status":"ok","token":"DAD2474A-E9A7-4BA7-BFC2-C4506880198E",
"data":{"fname":"paranoid","lname":"guy","login":"user1123"}}
```


## Update user record
### `PUT /v1/user/{token,login,email,phone}/{identity}`

### Explanation

This API is used to update user record. User record can be identified by `{token}`, 
`{email}`, `{phone}` or `{login}`.
On success, this API call returns success status or error message on error.

### POST Body Format

POST Body can contain regular form POST data or JSON. When using JSON, you can remove the record from
user profile by setting it's value to null. For example `{"key-to-delete":null}`.

### Example:

The following command will change user name to "Alex". An Audit event will be generated saving
previous and new value.

```
curl --header "X-Bunker-Token: $XTOKEN" -d 'name=Alex' -XPUT \
   https://localhost:3000/v1/user/token/DAD2474A-E9A7-4BA7-BFC2-C4506880198E
{"status":"ok","token":"DAD2474A-E9A7-4BA7-BFC2-C4506880198E"}
```


## Delete user record
### `DELETE /v1/user/{token,login,email,phone}/{identity}`

This command will remove all user records from the database, leaving only user `{token}` for refference.
This API is used to fullfull the customer' **right to forget**.

In Databunker **enterprise version**, user record deletion can be delayed as defined by the company policy.

### Example:

```
curl -header "X-Bunker-Token: $XTOKEN" -XDELETE \
  https://localhost:3000/v1/user/token/DAD2474A-E9A7-4BA7-BFC2-C4506880198E
{"status":"ok","result":"done"}
```

---

## User App Api

The User App API is used when you want to store additional information about the user and do not want to 
mix is with profile data. For example shipping information.

| Resource / HTTP method              | POST (create)       | GET (read)        | PUT (update)  | DELETE  |
| ----------------------------------- | ------------------- | ----------------- | ------------- | ------- |
| /v1/userapp/token/{token}/{appname} | New user app record | Get record        | Change record | Delete  |
| /v1/userapp/token/{token}           | Error               | Get user app list | Error         | Error   |
| /v1/userapps                        | Error               | Get all app list  | Error         | Error   |


## Create user app record
### `POST /v1/userapp/token/{token}/{appname}`

### Explanation

This API is used to create new user app record and if the request is successful it returns `{"status":"ok"}`.
Subminiting several times for the same user and app will overwrite previous value.

You can send app data as JSON POST or as regular POST parameters.

### Example:

```
curl -s http://localhost:3000/v1/userapp/token/$TOKEN/shipping \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"country":"UK","city":"London","address":"221B Baker Street","postcode":"12345","status":"new"}'
{"status":"ok","token":"$TOKEN"}
```

## Update user app record
### `PUT /v1/userapp/token/{token}/{appname}`

### Explanation

Update user app record with new values.

### Example:

```
curl -s http://localhost:3000/v1/userapp/token/$TOKEN/shipping -XPUT \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"status":"delivered"}'
{"status":"ok","token":"$TOKEN"}
```

## Get user app record
### `GET /v1/userapp/token/{token}/{appname}`

### Explanation

Returns user app record JSON.

### Example:

```
curl -s http://localhost:3000/v1/userapp/token/$TOKEN/shipping \
  -H "X-Bunker-Token: $XTOKEN"
{"status":"ok","token":"94d12078-18c5-e973-54db-f9aa92790f3f",
  "data":{"address":"221B Baker Street","city":"London","country":"UK","postcode":"12345","status":"delivered"}}
```

---

## User Session Api

You can use **Session API** to store and manage user sessions. For example sessions for Web and Mobile applications.
You can use **session token** generated in your application logs instead of clear text user IP, cookies, etc...
This information is is considered now as PII.

Session generation API is flexible and you can push any data you wish to save in session record. It can be:
user ip, mobile device info, user agent, etc...

Each session record has an expiration period. When the record it is expired, it is automatically deleted.


| Resource / HTTP method       | POST (create)      | GET (read)     | PUT (update)   | DELETE (delete) |
| ---------------------------- | ------------------ | -------------- | -------------- | --------------- |
| /v1/session/phone/{phone}    | Create new session | Get sessions   | Error          | Error           |
| /v1/session/email/{email}    | Create new session | Get sessions   | Error          | Error           |
| /v1/session/token/{token}    | Create new session | Get sessions   | Error          | Error           |
| /v1/session/session/:session | Error              | Get session    | Error          | Error           |



## Create user session record
### `POST /v1/session/{token,login,email,phone}/{identity}`

### Explanation

This API is used to create new user session and if the request is successful it returns new `{session}` token.

You can send the data as JSON POST or as regular POST parameters.

Additional parameter is **expiration** specifies TTL for this session record.

### Example:

```
curl -s http://localhost:3000/v1/session/email/test@securitybunker.io -XPOST \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json" \
   -d '{"expiration":"3d","clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2"}'`
{"status":"ok","session":"7a77ffad-2010-4e47-abbe-bcd04509f784"}
```

## Get user session record
### `GET /v1/session/session/:session`

### Explanation

This API returns session data.

### Example:

```
curl -s http://localhost:3000/v1/session/session/7a77ffad-2010-4e47-abbe-bcd04509f784 \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
{"status":"ok","session":"7a77ffad-2010-4e47-abbe-bcd04509f784","when":1576526253,
 "data":{"clientip":"1.1.1.1","info":"email","x-forwarded-for":"2.2.2.2"}}
```


## Get all session records by user identity.
### `GET /v1/session/{token,login,email,phone}/{identity}`

### Explanation

This API returns an array of session records for the same user. This command supports paging
arguments **offset** and **limit** (in URL request).

### Example:

```
curl -s http://localhost:3000/v1/session/phone/4444 \
   -H "X-Bunker-Token: "$DATABUNKER_APIKEY -H "Content-Type: application/json"`
{"status":"ok","count":"20","rows":[
   {"when":1576525605,"data":{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2"}},
   {"when":1576525605,"data":{"clientip":"1.1.1.1","info":"email","x-forwarded-for":"2.2.2.2"}},
   {"when":1576525660,"data":{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2"}},
   {"when":1576525660,"data":{"clientip":"1.1.1.1","info":"email","x-forwarded-for":"2.2.2.2"}},
   {"when":1576526129,"data":{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2"}},
   {"when":1576526130,"data":{"clientip":"1.1.1.1","info":"email","x-forwarded-for":"2.2.2.2"}},
   {"when":1576526253,"data":{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2"}},
   {"when":1576526253,"data":{"clientip":"1.1.1.1","info":"email","x-forwarded-for":"2.2.2.2"}},
   {"when":1576526291,"data":{"clientip":"1.1.1.1","x-forwarded-for":"2.2.2.2"}},
   {"when":1576526291,"data":{"clientip":"1.1.1.1","info":"email","x-forwarded-for":"2.2.2.2"}}]}
```

---

## User consent management API

One of the GDPR requirements is the storage of user consent. For example, your customer must approve to receive
email marketing information. Data bunker provides an API for user consent storage and management. 

When working with consent, Data bunker is using `brief` value as a consent name.
It is unique per user, short consent id. Allowed chars are [a-z0-9\-] . Max 64 chars.


| Resource / HTTP method            | POST (create)     | GET (read)    | DELETE (delete)    |
| --------------------------------- | ----------------- | ------------- | ----------------- |
| /v1/consent/token/{token}/{brief} | Create / Approve  | Get record    | Withdraw consent  |
| /v1/consent/login/{login}/{brief} | Create / Approve  | Get record    | Withdraw consent  |
| /v1/consent/email/{email}/{brief} | Create / Approve  | Get record    | Withdraw consent  |
| /v1/consent/phone/{phone}/{brief} | Create / Approve  | Get record    | Withdraw consent  |
| /v1/consent/token/{token}         | N/A               | Get records   | N/A               |
| /v1/consent/login/{login}         | N/A               | Get records   | N/A               |
| /v1/consent/email/{email}         | N/A               | Get records   | N/A               |
| /v1/consent/phone/{phone}         | N/A               | Get records   | N/A               |
| /v1/consents/{brief}              | N/A               | Get all users | N/A               |


## Create consent record
### `POST /v1/consent/{token,login,email,phone}/{identity}/{brief}`

### Explanation

This API is used to store user consent.

### POST Body Format

POST Body can contain regular form data or JSON. Here is a table with list of expected parameters.

| Parameter (required)  | Description                                                                    |
| --------------------- | ------------------------------------------------------------------------------ |
| status (no)           | Consent status. Default value is **accept**. Allowed values: cancel/accept.    |
| message (no)          | Text message describing consent. If empty **brief** is displayed.              |
| freetext (no)         | Free text, used for internal usage.                                            |
| starttime (no)        | Date & time to automatically enable this consent. Expected value is in UNIX time format or kind of 10d or 1m, etc...|
| expiration (no)       | Consent expiration date. Expected value is in UNIX time format or kind of 10d or 1m, etc...|
| lawfulbasis (no)      | Default is **consent**. It can be: **contract-agreement**, **legal-obligations**, etc...|
| consentmethod (no)    | Default is **api**. It can be: **phone-consent**, **contract**, **app-consent**, **web-consent**, **email-consent**, etc...|
| referencecode (no)    | This can be used as an id of your internal document, contract, etc...          |
| lastmodifiedby (no)   | Name of the person that last modified this record or **customer**.             |

When consent is expired, the status value is changed to **expired**.

### Example:

Create consent by posting JSON:

```
curl -s http://localhost:3000/v1/consent/email/test@securitybunker.io/send-sms -XPOST \
  -H "X-Bunker-Token: $XTOKEN" \
  -H "Content-Type: application/json" \
  -d '{"mesasge":"Optional long text here."}'
{"status":"ok"}
```

Create consent by POSTing user key/value fiels as POST fields:

```
curl -s http://localhost:3000/v1/consent/email/test@securitybunker.io/send-sms -XPOST \
  -H "X-Bunker-Token: $XTOKEN"
  -d 'mesasge=optional+text'
{"status":"ok"}
```

---

## Withdraw consent record
### `DELETE /v1/consent/{token,login,email,phone}/{identity}/{brief}`

### Explanation

This API is used to withdraw user consent.

### Example:

Withdraw consent:

```
curl -s http://localhost:3000/v1/consent/email/test@securitybunker.io/send-sms -XDELETE \
  -H "X-Bunker-Token: $XTOKEN"
{"status":"ok"}
```

---

## Get a list of all user consent records
### `GET /v1/consent/{token,login,email,phone}/{identity}`

### Explanation
This API returns an array of all user consent records. No pagination is supported.

### Example:

Fetch by user email:

```
curl --header "X-Bunker-Token: $XTOKEN" -XGET \
   https://localhost:3000/v1/consent/email/test@securitybunker.io
{"status":"ok","total":2,"rows":[
   {"brief":"send-email-mailgun-on-login","message":"send-email-mailgun-on-login","status":"accept",
   "token":"254d2abf-e927-bdcf-9cb2-f43c3cb7a8fa","mode":"email","when":1576154130,"who":"test@securitybunker.io"},
   {"brief":"send-sms-twilio-on-login","message":"send-sms-twilio-on-login","status":"accept",
   "token":"254d2abf-e927-bdcf-9cb2-f43c3cb7a8fa","mode":"phone","when":1576174683,"who":"4444"}]}
```

---

## Shareable record API

Sometimes you want to share part of user profile, part fo app data or a part of session details
with a partner or to save in logs. Data Bunker provides an easy API to do it.

Each record created has an expiration time. Access to the record data is open to anyone without
any access token or without a password.

| Resource / HTTP method         | POST (create)     | GET (read)   | PUT (update)  | DELETE (delete)  |
| ------------------------------ | ----------------- | ------------ | ------------- | ---------------- |
| /v1/sharedrecord/token/{token} | Create new record | Error        | Error         | Error            |
| /v1/get/{record}               | Error             | Get data     | Error         | Error            |


## Create shared record request
### `POST /v1/sharedrecord/token/{token}`

### Explanation

This API is used to create shared record that is referencing other object in the database (user, user-app, session).
It returns `{record}` token on success.

You can user this token, for example when you need to 3rd party system as a user identity.

### POST Body Format

POST Body can contain regular form data or JSON. Expected optional values are `message`, `status`.

| Parameter   | Required  | Description                                          |
| ----------- | --------- | ---------------------------------------------------- |
| app         | No        | Application name.                                    |
| partner     | No        | Partner name. For example coralogix.                 |
| expiration  | No        | Record expiration time.                              |
| session     | No        | Session record token.                                |
| fields      | No        | List of records to extract. Separated by commas.     |

### Example:

```
curl -s http://localhost:3000/v1/sharedrecord/token/$TOKEN \
  -H "X-Bunker-Token: $XTOKEN" -H "Content-Type: application/json" \
  -d '{"app":"shipping","fields":"address"}'
{"status":"ok","record":"db90efc7-48fe-9709-891d-a8b295881a9a"}
```

## Get record value.
### `GET /v1/get/{record}`

Return record value.

### Example:

```
curl -s http://localhost:3000/v1/get/$RECORD
{"status":"ok","app":"shipping","data":{"address":"221B Baker Street"}}
```
