![Databunker solution](images/databunker-solution.png)

# Databunker

**Databunker is a network-based, self-hosted, GDPR compliant, secure vault for personal data or PII: https://databunker.org/**

<div align="center">
 <p>
  <a href="https://github.com/securitybunker/databunker/stargazers" target="_blank"><img src="https://img.shields.io/github/stars/securitybunker/databunker.svg?logo=github&maxAge=86400" alt="Stars" /></a>
  <a href="https://github.com/securitybunker/databunker/actions?query=workflow%3ATests" target="_blank"><img src="https://github.com/securitybunker/databunker/workflows/Tests/badge.svg" alt="Tests" /></a>
  <a href="https://join.slack.com/t/databunker/shared_invite/zt-b6ukxzw3-JCxv8NJDESL40haM45RNIA"><img src="https://img.shields.io/badge/slack-join%20chat%20%E2%86%92-e01563.svg" alt="Join Databunker Slack channel" /></a>
  <a href="https://hub.docker.com/r/securitybunker/databunker"><img src="https://img.shields.io/docker/pulls/securitybunker/databunker?color=f02e65&style=flat-square" /></a>
 </p>
 <p>
  <a href="https://github.com/securitybunker/databunker-store"><img src="https://nodei.co/npm/@databunker/store.png?mini=true" alt="npm install @databunker/store" /></a>
  <a href="https://github.com/securitybunker/databunker-session-store"><img src="https://nodei.co/npm/@databunker/session-store.png?mini=true" alt="npm install @databunker/session-store" /></a>
 </p>
</div>

Project **demo** is available at: [https://demo.databunker.org/](https://demo.databunker.org/). Please add a **star** if you like our project.

⚠️ Here is a simple truth: <b>traditional database encryption often provides a false sense of security</b>.

What are the risks of traditional database security solutions?

* **Data encryption is not enough:** Most cloud and security vendors provide only data or disk encryption
* **Unfiltered GraphQL Queries:** Attackers can retrieve unencrypted data via incorrectly filtered queries
* **SQL Injection Attacks:** Cybercriminals can easily access plain text data through SQL injection

#### Introducing Databunker

Databunker is a specialized system for secure storage, data tokenization, and consent management, designed to protect:
* Personally Identifiable Information (PII)
* Protected Health Information (PHI)
* Payment Card Industry (PCI) data
* Know Your Customer (KYC) records

#### Key Features:
* **Open-Source:** Fully available under the commercially friendly MIT license
* **GDPR Compliant:** Built with privacy regulations in mind
* **Superior Protection:** Goes beyond standard database encryption offered by major vendors

#### How Databunker Reinvents Data Security:
Databunker introduces a new approach to customer data protection:
1. **Secure Indexing:** Utilizes hash-based indexing for all search indexes
1. **No Clear Text Storage:** Ensures all information is encrypted, enhancing overall security
1. **Restricted Bulk Retrieval:** Bulk retrieval is disabled by default, adding an extra layer of defense
1. **API-Based Communication:** Backend interacts with Databunker through API calls, similar to NoSQL solutions
1. **Record Token:** Databunker creates a secured version of your data object - an object UUID token that is safe to use in your database

Don't let your sensitive data become the next breach headline

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

* Backend at https://cloudrevive.com/
* Backend at https://metal8.cloud/
* Backend at https://bitbaza.io/

🚩 **Send us a note** if you are running Databunker in production mode, so we can add your website to the list.

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

* [A perfect backend for a KYC system for a crypto startup](https://databunker.org/success-story/kyc-backend-for-crypto-startup/)
* [Temporary record identities for secure data exchange](https://databunker.org/use-case/temporary-record-identity/)
* [Audit trail and tracing customer profile changes](https://databunker.org/use-case/trace-profile-access-change/)
* [Critical Data Segregation: Implementation Guide](https://databunker.org/use-case/critical-data-segregation/)
* [Continuous Data Protection for PII/PHI records](https://databunker.org/use-case/continuous-data-protection/)
* [Custom Privacy-Enhancing Technology - PET](https://databunker.org/use-case/privacy-enhancing-technology/)
* [User rights and privacy controls](https://databunker.org/use-case/user-privacy-controls/)
* [PII/PHI storage and tokenization](https://databunker.org/use-case/customer-profile-storage-tokenization/)
* [Automatic log retention policy](https://databunker.org/use-case/gdpr-compliant-logging/)
* [Privacy by Design Compliance](https://databunker.org/use-case/privacy-by-design-default/)
* [Simplify user login backend](https://databunker.org/use-case/simplify-user-login-backend/)
* [Consent Management Platform](https://databunker.org/use-case/consent-management-platform/)
* [Personal Data minimization](https://databunker.org/use-case/data-minimization/)
* [Secure session storage](https://databunker.org/use-case/secure-session-storage/)
* [GDPR request workflow](https://databunker.org/use-case/gdpr-user-request-workflow/)
* [DPO Management Portal](https://databunker.org/use-case/dpo-management-portal/)
* [User privacy portal](https://databunker.org/use-case/privacy-portal-for-customers/)
* [ISO27001 Compliance](https://databunker.org/use-case/iso27001-compliance/)
* [HIPAA Compliance](https://databunker.org/use-case/hipaa-compliance/)
* [GDPR Compliance](https://databunker.org/use-case/gdpr-compliance/)
* [SOC2 Compliance](https://databunker.org/use-case/soc2-compliance/)
* [Pseudonymization](https://databunker.org/use-case/pseudonymization-vs-anonymization/)
* Passport.js support
 
---

# Blog posts, articles, or other resources that talk about Databunker:

1. GDPR compliance and Databunker introduction video https://www.youtube.com/watch?v=QESOuL3LMj0
1. https://oppetmoln.se/20220223/databunker-en-oppen-losning-for-gdpr-saker-lagring-av-kundinformation/
1. https://www.freecodecamp.org/news/how-to-stay-gdpr-compliant-with-access-logs/
1. https://news.ycombinator.com/item?id=26690279
1. https://hackernoon.com/data-leak-prevention-with-databunker-xnn33u9
1. https://anchor.fm/techandmain/episodes/Huawei--Microsoft-and-DataBunker--Yuli-Stremovsky-evl385
1. https://nocomplexity.com/documents/simplifyprivacy/databunker.html
1. https://ipv6.rs/tutorial/FreeBSD_Latest/Databunker/
1. https://selfhostedworld.com/software/databunker
1. https://ipv6.rs/tutorial/Void_Linux/Databunker/
1. https://slashdot.org/software/p/Databunker/
1. https://github.com/expressjs/session
1. https://stackshare.io/databunker
1. https://dbweekly.com/issues/348
1. https://databunker.org/

If you published an article about Databunker send us a link at yuli@databunker.org

---

## 🚀 We are constantly working to improve this project

Your feedback is very important for us. Give us a ⭐ **star** if you like our product.

If you have any questions, you can contact the development team at hello@databunker.org.

Join the project slack channel to talk with developers: [https://databunker.slack.com/](https://join.slack.com/t/databunker/shared_invite/zt-b6ukxzw3-JCxv8NJDESL40haM45RNIA)

## ⭐ Spread a word to make a world a bit safer with Databunker
![Databunker](https://databunker.org/img/databunker.png)

Help us to raise awareness. Please add a ⭐ **star** and share this project with your friends.
