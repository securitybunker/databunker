# Data Bunker API


**Data Bunker is an information tokenization and storage service build to comply with GDPR and CCPA privacy requirements.**

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
### `GET /v1/user/{token,login,email,phone}/{address}`

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
### `PUT /v1/user/{token,login,email,phone}/{address}`

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
### `DELETE /v1/user/{token,login,email,phone}/{address}`

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

Sesion generation API is flexible and you can push any data you wish to save in session record. It can be:
user ip, mobile device info, user agent, etc...

Each session record has an expiration period. When the record it is expired, it is automatically deleted.


| Resource / HTTP method       | POST (create)      | GET (read)     | PUT (update)   | DELETE (delete) |
| ---------------------------- | ------------------ | -------------- | -------------- | --------------- |
| /v1/session/phone/{phone}    | Create new session | Get sessions   | Error          | Error           |
| /v1/session/email/{email}    | Create new session | Get sessions   | Error          | Error           |
| /v1/session/token/{token}    | Create new session | Get sessions   | Error          | Error           |
| /v1/session/session/:session | Error              | Get session    | Error          | Error           |



## Create user session record
### `POST /v1/session/{token,login,email,phone}/{address}`

### Explanation

This API is used to create new user session and if the request is successful it returns new `{session}` token.

You can send the data as JSON POST or as regular POST parameters.

Additional parameter is **expiration** specifies TTL for this session record.

### Example:

```
curl -s http://localhost:3000/v1/session/email/test@paranoidguy.com -XPOST \
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


## Get all session records by user address.
### `GET /v1/session/{token,login,email,phone}/{address}`

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
email marketing information. Data bunker provides an API for user consent management. 

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
### `POST /v1/consent/{token,login,email,phone}/{address}/{brief}`

### Explanation

This API is used to store user consent.

### POST Body Format

POST Body can contain regular form data or JSON. Expected values are `message`, `status`.

| Parameter   | Required  | Description                                                                 |
| ----------- | --------- | --------------------------------------------------------------------------- |
| status      | No        | Consent status. Default value is **accept**. Allowed values: cancel/accept. |
| message     | No        | Optional text message describing consent.                                   |


### Example:

Create consent by posting JSON:

```
curl -s http://localhost:3000/v1/consent/email/test@paranoidguy.com/send-sms -XPOST \
  -H "X-Bunker-Token: $XTOKEN" \
  -H "Content-Type: application/json" \
  -d '{"mesasge":"Optional long text here."}'
{"status":"ok"}
```

Create consent by POSTing user key/value fiels as POST fields:

```
curl -s http://localhost:3000/v1/consent/email/test@paranoidguy.com/send-sms -XPOST \
  -H "X-Bunker-Token: $XTOKEN"
  -d 'mesasge=optional+text'
{"status":"ok"}
```

---

## Withdraw consent record
### `DELETE /v1/consent/{token,login,email,phone}/{address}/{brief}`

### Explanation

This API is used to withdraw user consent.

### Example:

Withdraw consent:

```
curl -s http://localhost:3000/v1/consent/email/test@paranoidguy.com/send-sms -XDELETE \
  -H "X-Bunker-Token: $XTOKEN"
{"status":"ok"}
```

---

## Get a list of all user consent records
### `GET /v1/consent/{token,login,email,phone}/{address}`

### Explanation
This API returns an array of all user consent records. No pagination is supported.

### Example:

Fetch by user email:

```
curl --header "X-Bunker-Token: $XTOKEN" -XGET \
   https://localhost:3000/v1/consent/email/test@paranoidguy.com
{"status":"ok","total":2,"rows":[
   {"brief":"send-email-mailgun-on-login","message":"send-email-mailgun-on-login","status":"accept",
   "token":"254d2abf-e927-bdcf-9cb2-f43c3cb7a8fa","mode":"email","when":1576154130,"who":"test@paranoidguy.com"},
   {"brief":"send-sms-twilio-on-login","message":"send-sms-twilio-on-login","status":"accept",
   "token":"254d2abf-e927-bdcf-9cb2-f43c3cb7a8fa","mode":"phone","when":1576174683,"who":"4444"}]}
```

---

## Passwordless tokens API

| Resource / HTTP method | POST (create)     | GET (read)    | PUT (update)     | DELETE (delete)  |
| ---------------------- | ----------------- | ------------- | ---------------- | ---------------- |
| /v1/xtoken/{token}     | Create new record | Error         | Error            | Error            |
| /v1/xtoken/:xtoken     | Error             | Get data      | Error            | Error            |

	router.POST("/v1/xtoken/{token}", e.userNewToken)
	router.GET("/v1/xtoken/:xtoken", e.userCheckToken)


---


## Shareable record identity API

| Resource / HTTP method | POST (create)     | GET (read)    | PUT (update)     | DELETE (delete)  |
| ---------------------- | ----------------- | ------------- | ---------------- | ---------------- |
| /v1/record/{token}     | Create new record | Error         | Error            | Error            |
| /v1/record/{record}    | Error             | Get data      | Error            | Error            |


---


### 3rd party logging

Instead of maintaining internal logs, a lot of companies are using 3rd party logging facility like logz or coralogix or something else.
To improve adherence to GDPR, we build a special feature - generate specific session id for such 3rd party service.

When using these uuids in external systems, you basically **pseudonymise personal data**. In addition, in accordance with GDPR Article 5:
**Principles relating to processing of personal data**. Personal data shall be: (c) 
adequate, relevant and limited to what is necessary in relation to the purposes for which they are processed (‘**data minimisation**’);

Here is a command to do it:

```
curl -d 'ip=user@example.com' \
     -d 'user-agent=mozila' \
     -d 'partner=coralogix' \
     -d 'expiration=7d'\
   https://bunker.company.com/gensession/DAD2474A-E9A7-4BA7-BFC2-C4506880198E
```

It will generate a new token, that you can now pass to 3rd party system as a user id.
