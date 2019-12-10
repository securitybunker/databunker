# Paranoid Guy Data Bunker

Other documents: [API LIST](API.md)

**Data Bunker is advanced personal information tokenization and storage service build to comply with GDPR.**

This project, when deployed, can replace all user personal records scattered in the organization's different
internal databases with one user token generated and managed by Data Bunker service.

By deploying this project and moving all personal information to one place, you will comply with the following
GDPR statement: *Personal data should be processed in a manner that ensures appropriate security and 
confidentiality of the  personal data, including for preventing unauthorized access to or use of personal
data and the equipment used for the processing.*

**NOTE**: Implementing this project does not make you fully compliant with GDPR requirements and you still
need to consult with an attorney specializing in privacy.

#### Diagram of old-style solution.

![picture](images/old-style-solution.png)

#### Diagram of Solution with Paranoid Guy Data Bunker
![picture](images/new-style-solution.png)


---

# This product stands many GDPR requirements

## Right to be forgotten / Right to erasure

When your customer asks for his **right to be forgotten** legal right, his private records will be
wiped out of the Data Bunker database, giving you the possibility to leave all internal databases unchanged.

**NOTE**: You just need to make sure that you do not have any user identifiable information in your other databases,
logs, files, etc...

## Right of access

We build in passwordless login into the data bunker service. So, your customer/user can log in into his personal account
at Data Bunker and view all information collected by Data Bunker in connection to his profile.

## Right to rectification

Your customer/user can log in to his personal account at Data Bunker and change his records. If needed, Bunker will
send you a notification request about the change.

## Right to restrict processing / Right to object

Data Bunker can work as management for all user consents. User can cancel specific consent in his personal account at
Bunker, for example, to block sending him emails. Your backend can work with Data Bunker using API to add, or cancel
consents and we will send you a notification about user actions.

## Right to data portability (partial)

Your customer/user can log in to his personal account at Data Bunker and view and extract all his records stored at
Data Bunker.

**NOTE**: You need to provide your customers with a way to extract data from your internal databases.

## Data minimisation

Basically, when you clean up your databases from personal records and use Data Bunker token instead, you
are already minimizing the personal information you store in different systems. In addition, when sending
you customer data to 3rd party systems Data Bunker provides you with purposely build *shareable identity*
that is time-bound.

## Data Accuracy

We allow the customer to change the records that are stored in Data Bunker. This way we achieve data accuracy.

## Transparency

All operations with personal records are saved in the audit log. Your customer can log in to his account at Data Bunker
and view the audit trail.

## Integrity and confidentiality

All personal data is encrypted. Only relevant personnel can access the data. We audit all operations with personal records.
All-access to Data Bunker API is done using an HTTPS SSL certificate. Enterprise version supports Shamir's Secret Sharing
algorithm to split the master key to a number of keys. A number of keys (that can be saved in different hands in the
organization) are required to bring up the system.

## Accountability principle

Each one, connected to Data Bunker must provide an access token to do any operation in Data Bunker or the user needs to
login to access his own account. All operations are saved in the audit log.

## Privacy by design

This product, from the architecture level was build to comply with strick privacy laws. Deploying this or similar
architecture, can make your company privacy by design compliant.

## NOTE

Implementing this project does not make you fully compliant with GDPR requirements and you still need to
consult with an attorney specializing in privacy.

---

# Data Bunker usecases

## Personal Information tokenization and storage

This is already covered deeply above. Here I can add that Data Bunker has a layer of application
level personal information storage and each user in our database can be linked to a number of
application records (saved in Data Bunker).

## Audit of all operations with personal records

This is already covered above.

## GDPR compliant logging

Data Bunker supports a number of API that can help you to store user information in logs in
GDPR compliant way and work with cloud logging companies.

## Consent management, i.e. withdawal

According to GDPR, if you want to send your customer SMS using 3rd party gateway,
you must show to your customer a detailed notification message that you will send
his phone number to a specific SMS gateway company and the user needs to confirm that.

You need to store these confirmations and Data Bunker can help you with that.

Consent must be freely given, specific, informed and unambiguous. From GDPR, Article 7, item 3:

* **The data subject shall have the right to withdraw his or her consent at any time.**
* **It shall be as easy to withdraw as to give consent.**

In Data Bunker:

* Your customers can log in to his Data Bunker account and view all consents he gave.
* Users can also discharge consents and we will send you a notification message.


## User signup and sign-in

When implementing signup and sign-in in your customer-facing applications, we recommend you to
store all signup records in the Data Bunker database. We support 3 types of indexes, index
by login, index by email and index by phone. So you can easily implement login logic with
our service.

Index by email and index by phone allow us to give your customers passwordless access to their
personal profile at Data Bunker. We send your user a one-time login code by SMS or email to
give him access to his account at Data Bunker.


---

# Questions

## Why Open Source?

I am a big fan of the open-source movement. After a lot of thoughts and consultations,
the main Data Bunker product will be open source.

We are doing this to give our customers a steady base to continue using this solution in case
the company is closed.

Enterprise version will be closed source.

## What is considered PII?

Following it a partial list.

* Name
* Address
* IP address
* Browsing history
* Political opinion
* Sexual orientation
* Social Security Number
* Financial data
* Banking data
* Cookie data
* Contacts
* Mobile device ID
* Passport data
* Driving license
* ID number
* Health / medical data
* RFID
* Genetic info
* Ethnic and racial information

## Technology stack?

I am a big fan of go language and I use it extensively.


## Project technical features:

* [Encrypted storage for personal information](#personal-information-tokanization)
* [Application data separation](#application-data-separation)
* [Time-limited passwordless access to personal information](#time-limited-passwordless-access-to-personal-information)
* [Web and mobile app session data storage](#web-and-mobile-app-session-data-storage)
* [Time-limited passwordless access to web and app session data](#web-and-mobile-app-session-data-storage)
* [Share user identiy with 3rd party services](#share-user-identity-with-3rd-parties)
* [User consent management, storage & withdrawal](#user-consent-management)
* [Audit of all operations](#audit)
* [Customer UI](#user-ui)
* [User passwordless authentication](#custom-user-index)

## Enterprise features
* [Split master key with Shamir's Secret Sharing algo](#master-key-split-in-enterprise-version)
* [Advaned role management, ACL](#advanced-acl)
* [Support Hashicorp Vault](#hashicorp-vault-integration)

---

## Encryption in motion and encryption in storage

All access to Data Bunker API is done using HTTPS SSL certificate. All records that have user personal information
are encrypted or securely hashed in the databases. All user records are encrypted with a 32 byte key comprizing of
System Master key (24 bytes, stored in memory, not on disk) and user record key (8 bytes, stored on disk).

### Master key split in Enterprise version

Upon initial start, the **Enterprise version** generates a secret master key and 5 keys out of it.
These 5 keys are generated using Shamir's Secret Sharing algorithm. Combining 3 of any of the keys,
ejects original master key and that can be used to decrypt all records.

The Master key is kept in RAM and is never stored to disk. You will need to provide 3 kits to unlock the application.
It is possible to save these keys in the AWS secret store and other vault services.

---

## Data Bunker internal tables

Information inside Data Bunker is saved in multiple tables in encrypted format. Here is a diagram of tables.

Detailed usecase for each table is covered bellow.


![picture](images/data-bunker-tables.png)

---

## Personal information tokanization

User information, or PII, received in HTML POST key/value format of or JSON format is serialized, encrypted
with a 32 byte key and saved in database. You will get a user token to use in internal databases. Afterwords,
you can query the Data Bunker service to receive personal information, saving audit trail.

![picture](images/create-user-token-flow.png)

---

## Application data separation

When creating application, I suppose you do not want to mix your customer data with data from other applications.
In addition to personal information record, Data Bunker provides you a way to store your app user information in a
specific type of record for that. So, you can retreave only your app' user personal information.

![picture](images/create-user-app-record.png)

---

## Web and mobile app session data storage

Web or mobile application session data is very similar. They contain customer IP address, browser information,
web server headers, logged-in user info, etc... Many systems, including popular webservers, like Nginx, Apache
simply store this information in logs. This information, according to GDPR is considered personal identifiable
information and must be secured and controlled.

So, you can not save user ip or browser information in logs now. Insead, Data Bunker will generate you a token to
save in logs. Data Bunker provides you an API to retreave this info out of Data Bunker without additional password
for a limited time as in GDPR. For example one month.

![picture](images/create-user-session-flow.png)

---

## Time-limited passwordless access to personal information

Sometimes you want to share user, app or session private information in less trusted systems without providing
access to system root token.

Data Bunker has an API that allows you to generate temprorary access token to access specific fields in the
user personal record or application level data or a session record for a limited time only.

Your partner can retrieve this information and only specific fields during this specific timeframe.
Afterward, access will be blocked.

**IMAGE**

---

## Shareable user identity for 3rd parties

When sharing data with 3rd party services like web analytics, logging, intelligence, etc... sometimes we need to
share user id, for example, customer original IP address or email address. All these pieces of information
are considred user identifiable information and must be minimized when sending to 3rd paty systems.

***Do not share your customer user name, IP, emails, etc... because they look nice in reports!***

According to GDPR: *The personal data should be adequate, relevant and **limited to what is necessary** for the
purposes for which they are processed.*

Our system can generate you time-limited shareable identity token that you can share with 3rd parties as an identity.
This identity, can link bacck to the user personal record or user app record or to specific user session.

Optionally, Data Bunker can incorporate partner name in identity so, you track this identity usage.

**IMAGE**

---

## User consent management

Consent in GDPR terms is clear approval for example to share user information with 3rd party, for example with SMS
gateway company to send him urgent notifications.

Consent must be freely given, specific, informed and unambiguous. From GDPR, Article 7, item 3:

* **The data subject shall have the right to withdraw his or her consent at any time.**
* **It shall be as easy to withdraw as to give consent.**

To comply with this requirement, we support storage and management of user consent by API level and in user UI.

---

## Audit

Data Bunker saves audit events on all API operation. For example, new personal record added or changed; personal information
record retreaved, etc...

By providing Audit of events, in relation to personal data, provides response to GDRP Article 15 requirement:
*Right of access by the data subject*.

Special features:

* Personal information in audit event is encrypted.
* User can view his own records only.

Each audit record consists of:

* Date and time
* Operation title
* Operation status
* Operation description
* Change before and after if applicable
* User session info if available: IP address, headers, etc...

**IMAGE**

Example from google: https://console.cloud.google.com/home/activity

---

## User UI

Internal UI is build to allow users to login with their email and access all data collected by this system.
You can easily change it to your requirements.

According to GDPR, controller must provide Data subject with:

* Right of access by the data subject (Article 15)
* Right to rectification (Article 16)
* Right to be forgotten (Article 17)
* Right to restriction of processing (Article 18)
* Notification obligation regarding rectification or erasure of personal data or restriction of processing (Article 19)
* Right to data portability (Article 20)
* Right to object (Article 21)

---

# Enterprise features

## Advanced role management, ACL

By default, all access to Data Bunker is done with one root token or with **Time-limited passwordless access tokens**
that allow to read data from specific user record only.

For more granular control, Data Bunker supports the notion of custom roles. For example, you can create a role
to view all records or another role to add and change any user records; view sessions, view all audit events, etc...

After you define a role, the system allow you to generate access token for this role (you will need to have root token
for all these operations).

Data Bunker have an API for all these operations.

## Support Hashicorp Vault

Hashicorp Vault, is a great piece of new generation of security product, has a notion of session accounts/passwords.
Hashicorp Vault can store root access token to Paranoid Guy Data Bunker, and when your application wants to open
session and access Data Bunker, it will talk with Bunker to issue a temp token with specified role.
When your application session is closed with Data Bunker, Hashicorp Vault will connect to Data Bunker and revoke access token.

This architecture is done to minimize the chance that if the attacker breakes into your application server,
he will not get a full controll over the Data Bunker service as root token will not be saved in your
application server.

This is all done with the help of custom plugin we build for Hashicorp Vault.

---

Other documents: [API LIST](API.md)