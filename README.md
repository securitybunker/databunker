![Databunker solution](images/databunker-solution.png)

# Databunker

**Databunker is a Personally Identifiable Information (PII) Data Storage Service built to Comply with GDPR and CCPA Privacy Requirements.**

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

* Personal information tokenization and storage
* Critical data segregation
* Trace customer profile changes and access
* GDPR compliant logging : Web and mobile app session data storage
* Temporary customer/app/session identity for 3rd party services
* Data minimization and GDPR Scope reduction
* Consent management, i.e. withdawal
* Simplify user login
* GDPR user request workflow

---

# Contact us

For any questions, you can talk with us at: office@paranoidguy.com

---

Other documents: [API LIST](https://documenter.getpostman.com/view/11310294/Szmcbz32), [INSTALLATION](https://databunker.org/doc/install/)
