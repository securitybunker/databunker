# Paranoid Guy Data Bunker

**Data Bunker is a Personally Identifiable Information (PII) Data Storage Service built to Comply with GDPR and CCPA Privacy Requirements.**

[![Slack](https://img.shields.io/badge/slack-join%20chat%20%E2%86%92-e01563.svg)](https://join.slack.com/t/paranoidguy/shared_invite/enQtODc2OTE1NjYyODM1LTI0MmM2ZmYwZDI0MzExMjJmOGQyMTY4Y2UzOTQ0ZDIwOTZjMmRkZDZkY2I3MzE1OWE3ZWVmNTY4MjIwMzNhZTQ)

Project **demo** is available at: [https://demo.databunker.org/](https://demo.databunker.org/) . Please add a **star** if you like our project.

We live in a world where our privacy of information is nonexistent, the EU has been working to remediate this fallacy with GDPR, and the US (California) follows with a first sparrow called CCPA.

Data Bunker Project is intended to ease the acceptance of GDPR and CCPA regulations while giving organizations an easy to implement API's, secure Database to store PII and privacy portal. This will give all of us, the real data owners, control of our data, and allow us to know who is using our data, what is he doing with it and have the freedom to decide if we agree to that or not.

This project, when deployed correctly, replaces all the customer's personal records (PII) scattered in the organization's different
internal databases and log files with a single, randomly generated token managed by the Data Bunker service.

By deploying this project and moving all personal information to one place, you will comply with the following
GDPR statement: *Personal data should be processed in a manner that ensures appropriate security and 
confidentiality of the  personal data, including for preventing unauthorized access to or use of personal
data and the equipment used for the processing.*

#### Diagram of old-style solution.

![picture](images/old-style-solution.png)

#### Diagram of Solution with Paranoid Guy Data Bunker
![picture](images/new-style-solution.png)

Other documents: [API LIST](API.md), [INSTALLATION](INSTALLATION.md)

## Demo

Project demo is available at: [https://demo.databunker.org/](https://demo.databunker.org/)

You can see management for **Natural person** (**data subject**) account access:

```
Phone: 4444
Code: 4444
```

```
Email: test@paranoidguy.com
Code: 4444
```

Demo Admin access token: ```DEMO```

---

# This project resolves most** of the GDPR requirements for you including:

**NOTE**: Implementing this project does not make you fully compliant with GDPR requirements and you still
need to consult with an attorney specializing in privacy.

**NOTE**: When we use the term "Customer" we mean the data of the end-user that his information is being stored, shared and deleted.

## Right of access

Data Bunker extracts **customer email**, **customer phone** values out of the customers' personal records granting
**passwordless** access for the customer into their Data bunker' personal account.
This is done by generating a random access key that Data Bunker sends to your customer by email or by SMS.
Your customer can login and can view all information collected and saved by Data Bunker in connection to his profile.

<p float="middle">
  <img align="top" style="vertical-align: top;" src="images/ui-login-form.png" alt="login form" />
  <img align="top" style="vertical-align: top;" src="images/ui-login-email.png" alt="login with email" /> 
  <img align="top" style="vertical-align: top;" src="images/ui-login-code.png"  alt="verify login with code" />
</p>

## Right to restrict processing / Right to object / Consent withdrawal

Data Bunker can manages all the customer's consents. A customer can **Withdraw/restrict/object to a specific consent** in his personal account at Data Bunker, for example, to restrict or block email. Your backend site can work with Data Bunker using our API to add, or cancel
consents and a callback operation will be fired when a customer's action takes place.

![Consent management](images/ui-consent-management.png)
![Consent withdrawal](images/ui-consent-withdrawal.png)

**NOTE**: Data bunker can call your backend script on a consent withdrawal (callback). You will have to handle these requests and remove
the customer records from other 3rd party processing companies. For example: web recording services, email gateways and etc...

## Privacy by design

This product, from the architecture level and down to code was built to comply with strict privacy laws such as GDPR and CCPA. Deploying this project can make your architecture **privacy by design** compliant.

## Transparency and Accountability principle

Any system or customer connecting to Data Bunker must provide an **access token** to authorize any operation, otherwise the operation will be aborted. An end customer can login to his profile with a random authorization code sent by email or SMS.

All operations with personal records are **saved in the audit log**.

Any customer can log in to his account at Data Bunker and view the **full audit of activities** performed on his profile.

![Forget me](images/ui-audit-log.png)

## Right to be forgotten / Right to erasure

When your customer requests to exercise his **right to be forgotten**, his private records will be wiped out of the Data Bunker database, giving you the possibility to leave all internal databases intact while not impacting any of your other systems.

Upon customer removal request, Data bunker can call your backend script (callback) with the customer details. You will have to handle these requests and remove other customer records from 3rd party processing companies. For example from web recording services, email gateways and etc...

![Forget me](images/ui-forget-me.png)

**NOTE**: You will need to make sure that you do not have any customer identifiable information (PII) in your other databases,
logs, files and etc.

## Right to rectification/ Data Accuracy

Your customer can log in to his personal account at Data Bunker and change his records, for example **change his Name**.
Data Bunker can fire a callback operation with a customer details, when a customer action takes place.

![Change profile](images/ui-profile-edit-and-save.png)


## Right to data portability

Your customer can log in to his personal account at Data Bunker and view and **extract all his records stored at Data Bunker.**

**NOTE**: You will need to provide your customers with a way to extract data from other internal databases.


## Integrity and confidentiality

**All personal data is encrypted**. An audit log is written for all operations with personal records.
All-access to Data Bunker API is done using an **HTTPS SSL certificate**. Enterprise version supports Shamir's Secret Sharing
algorithm to split the master key to a number of keys. A number of keys (that can be saved in different hands in the
organization) are required to bring up the system.


## NOTE

**Implementing this project does not make you fully compliant with GDPR requirements and you still need to
consult with an attorney specializing in privacy.**

---

# Databunker use cases

Detailed information can be found at: https://databunker.org/use-case/

## Personal information tokenization and storage

## Critical data segregation

## Trace customer profile changes and access

## GDPR compliant logging : Web and mobile app session data storage

## Temporary customer/app/session identity for 3rd party services

## Data minimization and GDPR Scope reduction

## Consent management, i.e. withdawal

## Simplify user login

## GDPR user request workflow

---

# Questions

## How do I search for all orders from a guy named John?

Data bunker supports customer record lookup by **login name** or **email address** or **phone number** or **token value**.
So, if you have one of these values, you can do the customer record lookup (using Data Bunker API) and get customer token.
After that you can find customer' orders from the **orders table**.

## How to backup Data Bunker database?

We have a special API call for that. You can run the following command to dump database in SQL format:

```
curl -s http://localhost:3000/v1/sys/backup -H "X-Bunker-Token: $TOKEN" -o backup.sql
```

## Does your product multi-master solution?

Multi-master solution or basically multiple instances of the databunker service is supported in **Data Bunker
Enterprise version** running on AWS cloud. The product is using AWS Aurora PostgreSQL database at the backend.

Open source version is using local **sqlite3** database that does not supports replication. You can easily backup it
using API call and restore. We are using sqlite3 as as it provides zero effort from customer to start using
our product.

## Can my DBA tune database performance characteristics?

Almost all Data Bunker requests are using database level indexes when performing API calls.
We would love your DBA to check product database schema for improvements. If we are missing something let us know.
We are using **sqlite3** in open source version and **Aurora PostgreSQL** in enterprive version. You can easily backup
sqlite3 database and view it's structure.

## What is the difference between tokenization solution XXX and Data Bunker?

Most of commercial tokenization solutions are used to tokenize one specific record, for example customer name or 
customer email, etc... These distinct records are not linked to one customer record. In our solution, we tokenize the 
whole customer record with all the details, that gives us many additional capabilities. So, in our system, the
**end customer** (**Natural person** or **data subject**) can "login" into his profile, change record or
manage his consents, or ask for **forget me**. In addition we provide many APIs to help with GDPR requirements.

## Why Open Source?

We are a big fan of the open-source movement. After a lot of thoughts and consultations,
the main Data Bunker product will be open source.

We are doing this to boost the adoption of a **privacy enabled world**.

Enterprise version will be closed source.

## What is considered PII or what information is recomended to store in Data Bunker?

Following is a partial list.

| PII                           | PII                       |
| ----------------------------- | ------------------------- |
| * Name                        | * RFID                    |
| * Address                     | * Contacts                |
| * IP address                  | * Genetic info            |
| * Cookie data                 | * Passport data           |
| * Banking info                | * Driving license         |
| * Financial data              | * Mobile device ID        |
| * Browsing history            | * Personal ID number      |
| * Political opinion           | * Ethnic information      |
| * Sexual orientation          | * Health / medical data   |
| * Social Security Number      | * Etc...                  |


# Technology stack?

We use golang/go to build the whole project, with 80% automatic test coverage. Open source version comes with internal
database (**sqlite3**) and Web UI as one executable file to make the project easy to deploy.

## Does the product has encryption in motion and encryption in storage?

All access to Data Bunker API is done using HTTPS SSL certificate. All records that have customer personal information
are encrypted or securely hashed in the databases. All customer records are encrypted with a 32 byte key comprizing of
**System Master key** (24 bytes, stored in memory, not on disk) and **customer record key** (8 bytes, stored on disk).
The **System Master key** is kept in RAM and is never stored to disk. Enterprise version supports **Master key split**. 

## Is databunker is end-user facing?

Yes. The end-user, according to GDPR must have control over the PII data. The user can change the personal data, give 
or withdraw consent, request forget-me. All user requests can be self - service (automatic) or with DPO / Admin approval.

## Is databunker is a wrapper for exisitng MySQL/PostgreSQL/SQL Server database?

This product is not a wrapper for existing database. It is a special database used to save personal informatin records
in a compliant way. The service provides a REST API to store and update user records in JSON format; and customer
facing web ui to perform user data requests.
 
## Data Bunker internal tables

Information inside Data Bunker is saved in multiple tables in encrypted format. Here is a diagram of tables.

Detailed use case for each table is covered bellow.


![picture](images/data-bunker-tables.png)

---

# Enterprise features (not an open source version)

## PosgreSQL backend

The Databunker open source works with a local database, while enterprise version works with PostgreSQL.
For example, AWS Autora PostgreSQL. The last one of Enterprise grade and is available in AWS cloud.

## Master key split

Upon initial start, the **Enterprise version** generates a secret master key and 5 keys out of it.
These 5 keys are generated using Shamir's Secret Sharing algorithm. Combining 3 of any of the keys,
ejects original master key and that can be used to decrypt all records.

The Master key is kept in RAM and is never stored to disk. You will need to provide 3 kits to unlock the application.
It is possible to save these keys in the AWS secret store and other vault services.

## Advanced role management, ACL

By default, all access to Data Bunker is done with one root token or with **Time-limited access tokens**
that allow to read data from specific customer record only.

For more granular control, Data Bunker supports the notion of custom roles. For example, you can create a role
to view all records or another role to add and change any customer records; view sessions, view all audit events, etc...

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

Hashicorp plugin support is in BETTA stage. Contact us for more info.


# Contact us

For any questions, you can talk with us at: office@paranoidguy.com

---

Other documents: [API LIST](API.md), [INSTALLATION](INSTALLATION.md)
