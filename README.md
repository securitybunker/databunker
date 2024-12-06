![Databunker solution](images/databunker-solution.png)

# Databunker

**Databunker is a self-hosted, GDPR compliant, secure database for personal records - PII/PHI/KYC: https://databunker.org/**

<div align="center">
 <p>
  <a href="https://github.com/securitybunker/databunker/stargazers" target="_blank"><img src="https://img.shields.io/github/stars/securitybunker/databunker.svg?logo=github&maxAge=86400" alt="Stars" /></a>
  <a href="https://github.com/securitybunker/databunker/actions?query=workflow%3ATests" target="_blank"><img src="https://github.com/securitybunker/databunker/workflows/Tests/badge.svg" alt="Tests" /></a>
  <a href="https://hub.docker.com/r/securitybunker/databunker"><img src="https://img.shields.io/docker/pulls/securitybunker/databunker?color=f02e65&style=flat-square" /></a>
 </p>
 <p>
  <a href="https://github.com/securitybunker/databunker-store"><img src="https://nodei.co/npm/@databunker/store.png?mini=true" alt="npm install @databunker/store" /></a>
  <a href="https://github.com/securitybunker/databunker-session-store"><img src="https://nodei.co/npm/@databunker/session-store.png?mini=true" alt="npm install @databunker/session-store" /></a>
 </p>
</div>

## 🚀 Quick Start (5 minutes)

```bash
# Pull and run Databunker container
docker pull securitybunker/databunker
docker run -p 3000:3000 -d --rm --name dbunker securitybunker/databunker demo

# Create user records
curl -s http://localhost:3000/v1/user -X POST \
  -H "X-Bunker-Token: DEMO" \
  -H "Content-Type: application/json" \
  -d '{"first":"John","last":"Doe","login":"john","email":"user@gmail.com"}'

# Get user by login, email, phone, or token
curl -s -H "X-Bunker-Token: DEMO" -X GET http://localhost:3000/v1/user/login/john

# Admin UI: http://localhost:3000
```

## 💡 What Problems Does Databunker Solve?

1. **Prevents Data Breaches**
   - Eliminates SQL injection vulnerabilities
   - Protects against GraphQL data exposure
   - Segregates sensitive data from your main database

2. **Simplifies Compliance**
   - GDPR, CCPA, HIPAA ready out of the box
   - Built-in consent management
   - Automated data minimization
   - Full audit trail of all operations

3. **Reduces Development Time**
   - Simple REST API for all operations
   - SDK available for popular languages
   - Drop-in replacement for your user table
   - Built-in session management

Project **demo** is available at: https://databunker.org/doc/demo/.

Please add a **star** if you like our project.

## 🔒 Key Security Features

- **Encrypted Storage**: All personal records are encrypted using AES-256
- **Secure API**: REST API with strong authentication
- **Tokenization**: Replace sensitive data with tokens in your main database
- **Access Control**: Fine-grained permissions and audit logging
- **Data Segregation**: Physical separation from your application database

## 🔌 Integration Examples

```javascript
// Node.js Example
const { Databunker } = require('databunker-sdk');
const db = new Databunker({
  url: 'http://localhost:3000',
  token: 'DEMO'
});

// Store user record
await db.users.create({
  email: 'user@example.com',
  name: 'John Doe',
  phone: '+1-415-555-0123'
});

// Retrieve user by email
const user = await db.users.findByEmail('user@example.com');
```

## 📊 Use Cases

- **User Profile Storage**: Secure storage for user personal data
- **Healthcare Records**: HIPAA-compliant patient data storage
- **Financial Services**: PCI DSS compliant customer records
- **Identity Management**: Secure user authentication and session storage
- **GDPR Compliance**: Built-in tools for data privacy regulations

## 💻 Available Editions

### Community Edition (Free)
- Full encryption capabilities
- MySQL/PostgreSQL support
- Basic audit logging
- MIT license

### Enterprise Edition (Startup-friendly pricing)
- Multi-tenancy support
- Advanced monitoring
- Priority support
- Custom deployment options
- Aurora PostgreSQL auto-scaling

## 🔧 Technical Specifications

- Written in Go for high performance
- Supports MySQL and PostgreSQL
- REST API with OpenAPI specification
- Containerized deployment
- Horizontal scaling support
- Automated backups
- High availability options

## 📚 Resources

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

## 📘 GDPR: Out of the box solution for:

1. [Right of access](#right-of-access)
1. [Right to restrict processing / Consent withdrawal](#right-to-restrict-processing--consent-withdrawal)
1. [Right to be forgotten](#right-to-be-forgotten)
1. [Right to rectification](#right-to-rectification)
1. [Right to data portability](#right-to-data-portability)


## ⚡ Databunker use cases

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

Help us to raise awareness. Please add a ⭐ **star** and share this project with your friends.
