![Databunker](https://databunker.org/img/databunker.png)

**Databunker is a special self-hosted encrypted database for critical personal data, PII, and PHI. It is built with privacy by design methodology. You get GDPR compliance out of the box. More info https://databunker.org/**

 <p>
  <a href="https://github.com/securitybunker/databunker/actions?query=workflow%3ATests" target="_blank"><img src="https://github.com/securitybunker/databunker/workflows/Tests/badge.svg" alt="Tests"></a>
  <a href="https://join.slack.com/t/databunker/shared_invite/zt-b6ukxzw3-JCxv8NJDESL40haM45RNIA"><img src="https://img.shields.io/badge/slack-join%20chat%20%E2%86%92-e01563.svg" alt="Join Databunker Slack channel" /></a>
  <a href="https://hub.docker.com/r/securitybunker/databunker"><img src="https://img.shields.io/docker/pulls/securitybunker/databunker?color=f02e65&style=flat" /></a>

## ⭐⭐⭐ Spread a word to make a world a bit safer
Help us to raise awareness. Please add a ⭐ **star** and share this project with your friends.

![Databunker solution](images/databunker-solution.png)

## Intro

We live in a world where the privacy of our information is nonexistent. The EU has been working to remediate this fallacy with GDPR, and the US (California) follows with a first sparrow called CCPA.

Databunker project is intended to ease the GDPR and CPRA compliance. It gives organizations easy-to-implement APIs and secure vault to store PII, and a privacy portal.

Databunker allows to know who is using your data, what is happening with your personal details and gives you the freedom to decide how to process your personal data.

Databunker, when deployed correctly, replaces all the customer's personal records (PII) scattered in the organization's different
internal databases and log files with a single randomly generated token managed by the Databunker service.

By deploying this project and moving all personal information to one place, you will comply with the following
GDPR statement: *Personal data should be processed in a manner that ensures appropriate security and 
confidentiality of the  personal data, including for preventing unauthorized access to or use of personal
data and the equipment used for the processing.*

#### Diagram of old-style solution.

![picture](images/old-style-solution.png)

#### Diagram of Solution with Databunker
![picture](images/new-style-solution.png)

Getting started guide: https://databunker.org/doc/start/

Databunker installation guide: https://databunker.org/doc/install/

## 🚀 Demo

Project demo is available at: [https://demo.databunker.org/](https://demo.databunker.org/)

You can access the demo UI using the following account credentials:

```
Phone: 4444
Captcha: type as displayed
Access code: 4444
```

```
Email: test@securitybunker.io
Captcha: type as displayed
Access code: 4444
```

Demo root token: ```DEMO```

---

## 🛠️ Node.js Examples

1. Node.js example implementing passwordless login using Databunker:
https://github.com/securitybunker/databunker-nodejs-passwordless-login

2. Node.js example with Passport.js, Magic.Link and Databunker:
https://github.com/securitybunker/databunker-nodejs-example

3. Secure Session Storage for Node.js apps:
https://databunker.org/use-case/secure-session-storage/#databunker-support-for-nodejs

## 🛠️ Node.JS modules

1. `@databunker/store` from https://github.com/securitybunker/databunker-store

2. `@databunker/session-store` from https://github.com/securitybunker/databunker-session-store

## ⚡ Databunker benchmark results:

https://databunker.org/doc/benchmark/

## ⚡ Production deployments

* Backend at https://privacybunker.io/
* Backend at https://bitbaza.io/

🚩 **Send us a note** if you are running Databunker in production mode, so we can add your wesbite to the list.

## Privacy by design

This product, from the architecture level and down to code was built to comply with strict privacy laws such as GDPR and CCPA. Deploying this project can make your architecture **privacy by design** compliant. For more info, check out the following article:

https://databunker.org/use-case/privacy-by-design-default/

## Transparency and Accountability principle

Any system or customer connecting to Databunker must provide an **access token** to authorize any operation, otherwise, the operation will be aborted. An end customer can login to his profile with a random authorization code sent by email or SMS.

All operations with personal records are **saved in the audit log**.

Any customer can log in to his account at Data Bunker and view the **full audit of activities** performed on his profile.

![Forget me](images/ui-audit-log.png)

## Integrity and confidentiality

**All personal data is encrypted**. An audit log is written for all operations with personal records.
Any request using Databunker API is done with **HTTPS SSL certificate**. The enterprise version supports Shamir's Secret Sharing
algorithm to split the master key into a number of keys. A number of keys (that can be saved in different hands in the
organization) are required to bring the system up.

---

## 🚀 Databunker quick start guide

Follow this [article](https://databunker.org/doc/start/).

---

# This projects provides an instant solution for GDPR compliance and user rights:

1. [Right of access](#right-of-access)
1. [Right to restrict processing / Consent withdrawal](#right-to-restrict-processing--consent-withdrawal)
1. [Right to be forgotten](#right-to-be-forgotten)
1. [Right to rectification](#right-to-rectification)
1. [Right to data portability](#right-to-data-portability)

🚩 **NOTE**: Implementing this project does not make you fully compliant with GDPR requirements and you still
need to consult with an attorney specializing in privacy.

🚩 **NOTE**: When we use the term "Customer" we mean the data of the end-user that his information is being stored, shared, and deleted.


## Right of access

Databunker extracts **customer email** and **customer phone** values out of the customers' personal records. It gives your customer **passwordless** access to his data stored under his account. This is done by generating a random access key sent by email or by SMS. Your customer can sign-in into Databunker, view information stored by Databunker, and make changes in compliance with a company's policy.

<p float="middle">
  <img align="top" style="vertical-align: top;" src="images/ui-login-form.png" alt="login form" />
  <img align="top" style="vertical-align: top;" src="images/ui-login-email.png" alt="login with email" /> 
  <img align="top" style="vertical-align: top;" src="images/ui-login-code.png"  alt="verify login with code" />
</p>

## Right to restrict processing / Consent withdrawal

Databunker can manage all of the customer's consents and agreements in one place. Your customer can **withdraw consent** and as a result **restrict processing** in his personal portal at Databunker. For example, your customer can block newsletter service. Your backend system can use Databunker as a collection of all agreements collected using the Databunker API.

![Consent management](images/ui-consent-management.png)
![Consent withdrawal](images/ui-consent-withdrawal.png)

🚩 **NOTE**: Databunker can call your backend script on a consent withdrawal (callback). You will have to handle these requests and remove
the customer records from other 3rd party processing companies. For example from email newsletter service, etc...

## Right to be forgotten

When your customer requests to execute his **right to be forgotten**, his private records will be wiped out of the Databunker database, giving you the possibility to leave all internal databases intact while not impacting any of your other systems.

Upon customer removal request, Databunker can call your backend script (callback) with the customer details. You will have to handle these requests and remove other customer records from the 3rd party processing companies. For example from browsing recording services, etc...

![Forget me](images/ui-forget-me.png)

🚩 **NOTE**: You will need to make sure that you do not have any customer identifiable information (PII) in your other databases,
logs files, etc...

## Right to rectification

**Right to rectification** is also known as **data accuracy** requirement. Your customer can sign-in to his personal account at Databunker and change his records, for example, **change his name**.
Databunker can fire a callback operation with customer's details when a customer operation takes place.

![Change profile](images/ui-profile-edit-and-save.png)


## Right to data portability

Your customer can sign in to his personal account at Databunker and view and **extract all his records stored at Databunker.**

🚩 **NOTE**: You will need to provide your customers with a way to extract data from other internal databases.


## 🚩 NOTE

**Implementing this project does not make you fully compliant with GDPR requirements and you still need to
consult with an attorney specializing in privacy.**

---

# Databunker use cases

Detailed information can be found at https://databunker.org/use-case/

* Personal information tokenization and storage https://databunker.org/use-case/customer-profile-storage-tokenization/
* Pseudonymized user identity for cross-border information transfer https://databunker.org/use-case/schrems-ii-compliance/
* Critical data segregation https://databunker.org/use-case/critical-data-segregation/
* Personal Data minimization https://databunker.org/use-case/data-minimization/
* Trace customer profile changes and access https://databunker.org/use-case/trace-profile-access-change/
* Temporary customer/app/session identity for 3rd party services https://databunker.org/use-case/temporary-record-identity/
* Encrypted session storage https://databunker.org/use-case/secure-session-storage/
* GDPR compliant logging https://databunker.org/use-case/gdpr-compliant-logging/ 
* User privacy portal https://databunker.org/use-case/user-privacy-controls/
* Consent management, i.e. withdrawal
* Passport.js support
* DPO friendly service
 
---

# Blog posts, articles, or other resources that talk about Databunker:

1. https://privacybunker.io/blog/gdpr-guide-for-startup-founders/
1. https://dbweekly.com/issues/348
1. https://www.freecodecamp.org/news/how-to-stay-gdpr-compliant-with-access-logs/
1. https://news.ycombinator.com/item?id=26690279
1. https://stackshare.io/databunker
1. https://hackernoon.com/data-leak-prevention-with-databunker-xnn33u9
1. https://anchor.fm/techandmain/episodes/Huawei--Microsoft-and-DataBunker--Yuli-Stremovsky-evl385
1. https://github.com/expressjs/session
1. https://databunker.org/

If you published an article about Databunker send us a link at yuli@privacybunker.io

---

# ⭐ We always strive to make our projects better

Your feedback is very important for us.

Give us a ⭐ **star** if you like our product.

If you have any questions, you can contact the development team at office@privacybunker.io.

Join the project slack channel to talk with developers: [https://databunker.slack.com/](https://join.slack.com/t/databunker/shared_invite/zt-b6ukxzw3-JCxv8NJDESL40haM45RNIA)
