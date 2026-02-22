# Architecture Deep Dive: How Databunker Protects Your Data

Understanding Databunker's security guarantees from first principles.

## The Problem: Why Traditional Encryption Fails

### Scenario: E-commerce Database Breach

**Traditional Setup:**
```sql
-- users table (encrypted at rest)
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255),       -- Encrypted on disk
  name VARCHAR(255),        -- Encrypted on disk
  credit_card VARCHAR(19),  -- Encrypted on disk
  ssn VARCHAR(11)          -- Encrypted on disk
);
```

**Attack Vector 1: SQL Injection**
```sql
-- Attacker injects malicious SQL
SELECT * FROM users WHERE email = 'admin@example.com' OR '1'='1';
-- Returns ALL users in PLAINTEXT (encryption happens at disk level)
```

**Attack Vector 2: Compromised API**
```graphql
# GraphQL injection
query {
  users {  # Bulk query
    email
    creditCard
    ssn
  }
}
# Leaks entire database in one request
```

**Attack Vector 3: Insider Threat**
```bash
# Database admin runs COPY command
psql -c "COPY users TO '/tmp/leak.csv' CSV HEADER"
# Entire database exported in plaintext
```

**Result:** Game over. Encryption-at-rest provides ZERO protection at API layer.

---

## Databunker's Solution: API-Level Encryption + Tokenization

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Your Application                         │
│  (stores UUID tokens instead of PII)                        │
└──────────────┬────────────────────────────────┬─────────────┘
               │                                │
       ┌───────▼────────┐              ┌────────▼──────────┐
       │  Main Database │              │   Databunker      │
       │  (PostgreSQL)  │              │   (Secure Vault)  │
       ├────────────────┤              ├───────────────────┤
       │ id: UUID       │              │ Encrypted Storage │
       │ user_token:    │              │ - AES-256-GCM     │
       │   a3b2c1...    │──lookup────→ │ - bcrypt hashes   │
       │ created_at     │              │ - No bulk export  │
       └────────────────┘              └───────────────────┘
```

**Key Difference:** Databunker sits BETWEEN your app and the sensitive data.

---

## Layer 1: Tokenization Engine

### How Tokens Work

**Step 1: Store User Data**
```bash
curl -X POST http://databunker:3000/v1/user \
  -H "X-Bunker-Token: MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "ssn": "123-45-6789",
    "credit_card": "4111111111111111"
  }'
```

**Response:**
```json
{
  "status": "ok",
  "token": "a3b2c1d4-e5f6-7890-abcd-ef1234567890"
}
```

**Step 2: Store Token in Your Database**
```sql
INSERT INTO orders (user_token, product_id, amount)
VALUES ('a3b2c1d4-e5f6-7890-abcd-ef1234567890', 'prod_123', 99.99);
```

**Step 3: Retrieve Data When Needed**
```bash
curl -H "X-Bunker-Token: MASTER_KEY" \
  http://databunker:3000/v1/user/token/a3b2c1d4-e5f6-7890-abcd-ef1234567890
```

**Why This Matters:**
- **SQL injection?** Only leaks tokens, not PII
- **GraphQL bulk query?** Your DB has no PII to leak
- **Database dump?** Only meaningless UUIDs

---

## Layer 2: Encrypted Storage

### Encryption Stack

```
┌─────────────────────────────────────────┐
│  User Input (JSON)                      │
│  {"email": "alice@example.com"}         │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  AES-256-GCM Encryption                 │
│  Key: PBKDF2(MASTER_KEY, salt, 100k)    │
│  Output: [nonce|ciphertext|tag]         │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  Storage (SQLite/PostgreSQL)            │
│  records table:                         │
│    token: a3b2c1...                     │
│    data: 0x3a7f2e... (encrypted blob)   │
└─────────────────────────────────────────┘
```

**Encryption Details:**
- **Algorithm:** AES-256-GCM (authenticated encryption)
- **Key Derivation:** PBKDF2-HMAC-SHA256 (100,000 rounds)
- **Per-Record Salt:** Unique salt for each record
- **Authentication:** GCM tag prevents tampering

**Code Walkthrough (Go):**
```go
// internal/crypto/encrypt.go
func EncryptRecord(plaintext []byte, masterKey string) ([]byte, error) {
    // Derive encryption key from master key
    salt := generateRandomSalt(32)
    key := pbkdf2.Key([]byte(masterKey), salt, 100000, 32, sha256.New)
    
    // Create AES-GCM cipher
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    
    // Generate random nonce
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    
    // Encrypt + authenticate
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    
    // Return: salt || ciphertext
    return append(salt, ciphertext...), nil
}
```

**Why This Matters:**
- **Database compromised?** Attacker sees encrypted blobs
- **No master key?** Data is unreadable
- **Tampering detected?** GCM tag verification fails

---

## Layer 3: Secure Indexing (Hash-Based Search)

### The Challenge: Searchable Encryption

**Problem:** You want to search users by email, but can't store emails in plaintext.

**Naive Solution (DON'T DO THIS):**
```sql
-- ❌ Searchable plaintext index
CREATE INDEX idx_email ON users(email);
-- SQL injection can still read the index!
```

**Databunker's Solution: Blind Indexing**

```
Input: alice@example.com
         │
         ▼
   bcrypt hash (cost=10)
         │
         ▼
$2a$10$3euPcmQFCiblsZeEu5s7p.z8zv... (60 chars)
         │
         ▼
  Store in 'emailhash' column
```

**Storage Schema:**
```sql
CREATE TABLE records (
    token TEXT PRIMARY KEY,        -- UUID
    data BLOB,                     -- Encrypted PII
    emailhash TEXT,                -- bcrypt(email)
    phonehash TEXT,                -- bcrypt(phone)
    loginhash TEXT                 -- bcrypt(login)
);
CREATE INDEX idx_emailhash ON records(emailhash);
```

**Search Query:**
```bash
# Lookup by email
curl -H "X-Bunker-Token: MASTER_KEY" \
  "http://databunker:3000/v1/user/email/alice@example.com"
```

**Databunker internally:**
```go
// 1. Hash the search term
emailHash := bcrypt.GenerateFromPassword([]byte("alice@example.com"), 10)

// 2. Query by hash
db.QueryRow("SELECT token, data FROM records WHERE emailhash = ?", emailHash)

// 3. Decrypt data
plaintext := decrypt(encryptedData, masterKey)

// 4. Return to API caller
```

**Why This Matters:**
- **Rainbow table attack?** bcrypt is too slow (can't precompute)
- **Dump emailhash column?** Hashes are one-way (can't reverse)
- **Brute-force search?** bcrypt makes it computationally infeasible

---

## Layer 4: API Security

### Preventing Bulk Data Leaks

**Problem:** Traditional APIs allow bulk retrieval.

```bash
# ❌ Traditional API - returns ALL users
GET /api/users
```

**Databunker's Approach: Granular Access Only**

```bash
# ✅ Single-user lookup by token
GET /v1/user/token/{uuid}

# ✅ Single-user lookup by email
GET /v1/user/email/{email}

# ❌ Bulk export disabled by default
GET /v1/users  # Returns 403 Forbidden
```

**Audit Trail:**
```json
{
  "timestamp": "2026-02-22T12:00:00Z",
  "action": "user.read",
  "token": "a3b2c1d4-e5f6-7890-abcd-ef1234567890",
  "ip": "192.168.1.100",
  "api_key": "admin_key_1"
}
```

Every access is logged with:
- **Who** (API key)
- **What** (user token + fields accessed)
- **When** (timestamp)
- **Where** (IP address)

---

## Layer 5: Injection Protection

### SQL Injection Defense

**Traditional Database (Vulnerable):**
```sql
-- Vulnerable query
query := "SELECT * FROM users WHERE email = '" + userInput + "'"
-- Injection: userInput = "admin@example.com' OR '1'='1"
```

**Databunker (Immune):**
```go
// 1. User searches for: admin@example.com' OR '1'='1
maliciousInput := "admin@example.com' OR '1'='1"

// 2. Databunker hashes the entire string
emailHash := bcrypt(maliciousInput)
// Result: $2a$10$abc123... (NOT a valid email hash)

// 3. Query by hash (parameterized)
db.QueryRow("SELECT data FROM records WHERE emailhash = ?", emailHash)
// Returns 0 rows (no match)
```

**Why It Works:**
- Injection payload is hashed as-is
- Hash doesn't match any real user's email hash
- Query fails safely without leaking data

---

## Layer 6: GDPR Compliance Engine

### Right to Erasure (Article 17)

**Traditional Approach:**
```sql
-- ❌ Hard delete - breaks foreign keys
DELETE FROM users WHERE id = '123';
-- ❌ Soft delete - still stored
UPDATE users SET deleted_at = NOW() WHERE id = '123';
```

**Databunker Approach:**
```bash
# Immediate erasure
DELETE /v1/user/token/{uuid}
```

**What Happens:**
1. Decrypt user data
2. Generate anonymized placeholder:
   ```json
   {
     "email": "deleted_user_123@example.com",
     "name": "Deleted User"
   }
   ```
3. Re-encrypt placeholder
4. Overwrite original data
5. Log erasure in audit trail

**Application Impact:**
```sql
-- Your database still has the token
SELECT user_token FROM orders WHERE id = 1;
-- Returns: a3b2c1d4-e5f6-7890-abcd-ef1234567890

-- But fetching data returns placeholder
GET /v1/user/token/a3b2c1d4-e5f6-7890-abcd-ef1234567890
-- Returns: {"email": "deleted_user_123@example.com"}
```

**Why This Matters:**
- Foreign keys stay intact
- Reports don't break
- GDPR compliance in seconds

---

## Layer 7: Consent Management

### GDPR Consent Tracking

**Databunker API:**
```bash
# Record consent
POST /v1/consent/{token}/newsletter
{
  "consented": true,
  "lawful_basis": "consent",
  "consent_method": "web_form",
  "consent_proof": "IP: 192.168.1.100, Timestamp: 2026-02-22T12:00:00Z"
}

# Check consent
GET /v1/consent/{token}/newsletter
# Returns: {"consented": true, "when": "2026-02-22T12:00:00Z"}

# Withdraw consent
DELETE /v1/consent/{token}/newsletter
```

**Application Integration:**
```python
# Before sending marketing email
consent = databunker.get_consent(user_token, "newsletter")
if consent["consented"]:
    send_email(user)
else:
    log("User opted out of newsletter")
```

---

## Performance Characteristics

### Latency Breakdown

**Single-User Lookup:**
```
┌────────────────────────────────────────┐
│ HTTP request parsing          │  0.1ms │
│ API key validation            │  0.2ms │
│ Hash generation (bcrypt)      │  8.0ms │ ← Dominant cost
│ Database query                │  1.5ms │
│ Decryption (AES-256-GCM)      │  0.3ms │
│ JSON serialization            │  0.2ms │
│ Total                         │ 10.3ms │
└────────────────────────────────────────┘
```

**Optimization: Cache Hash Lookups**
```go
// Cache bcrypt hashes (read-only, safe to cache)
var hashCache = cache.New(5*time.Minute, 10*time.Minute)

func lookupByEmail(email string) (User, error) {
    hash, found := hashCache.Get(email)
    if !found {
        hash = bcrypt(email)  // 8ms
        hashCache.Set(email, hash, cache.DefaultExpiration)
    }
    return db.Query("SELECT data FROM records WHERE emailhash = ?", hash)
}
```

**Result:** Subsequent lookups drop to ~2.5ms.

---

## Security Threat Model

### Attack 1: Compromised Application Server

**Attacker gains shell access to your app server.**

**What they can access:**
- ✅ Your main database (but only UUIDs, no PII)
- ❌ Databunker (requires API key)
- ❌ Master encryption key (stored separately)

**Mitigation:**
- Store Databunker API key in environment variable
- Use different API keys for read vs write
- Enable IP whitelisting for Databunker

### Attack 2: Compromised Database

**Attacker dumps your PostgreSQL database.**

**What they get:**
- ❌ PII (it's all in Databunker)
- ✅ UUIDs (useless without Databunker access)
- ✅ Business data (orders, products, etc.)

**Mitigation:**
- Rotate Databunker tokens periodically
- Use time-limited session tokens

### Attack 3: Stolen Databunker Backup

**Attacker steals Databunker's database file.**

**What they need to decrypt:**
1. Master encryption key (not in backup)
2. Brute-force AES-256 (impossible)
3. Brute-force PBKDF2 (100k rounds = very slow)

**Mitigation:**
- Store master key in hardware security module (HSM)
- Use key rotation
- Encrypt backups at rest

---

## Deployment Architectures

### Architecture 1: Single Instance

```
┌─────────────┐
│   Your App  │
└──────┬──────┘
       │
       ▼
┌──────────────┐     ┌─────────────────┐
│  Main DB     │     │   Databunker    │
│ (PostgreSQL) │     │ (Docker/Binary) │
└──────────────┘     └─────────────────┘
```

**Pros:** Simple, low latency  
**Cons:** Single point of failure  
**Use case:** Development, small apps

### Architecture 2: High Availability

```
┌─────────────┐
│   Your App  │
│  (3 nodes)  │
└──────┬──────┘
       │
       ▼
┌──────────────┐     ┌─────────────────────────┐
│  Main DB     │     │  Databunker Cluster     │
│ (Replicated) │     │  ┌─────┐  ┌─────┐       │
└──────────────┘     │  │ DB1 │  │ DB2 │       │
                     │  └─────┘  └─────┘       │
                     │  Load Balancer          │
                     └─────────────────────────┘
```

**Pros:** No downtime, horizontal scaling  
**Cons:** More complex  
**Use case:** Production, > 100k users

### Architecture 3: Multi-Region

```
┌──────────────────────────────────────────┐
│          Global Load Balancer            │
└──────┬───────────────────┬───────────────┘
       │                   │
   ┌───▼────┐         ┌────▼────┐
   │ US-East│         │ EU-West │
   │ Region │         │ Region  │
   ├────────┤         ├─────────┤
   │ App    │         │ App     │
   │ DB     │         │ DB      │
   │ Bunker │         │ Bunker  │
   └────────┘         └─────────┘
```

**Pros:** Low latency globally, GDPR data residency  
**Cons:** Most complex  
**Use case:** Global SaaS, > 1M users

---

## Code Example: Full Integration

```python
# app.py
import requests
from flask import Flask, request

app = Flask(__name__)

DATABUNKER_URL = "http://databunker:3000"
DATABUNKER_TOKEN = "your_master_key_here"

def databunker_headers():
    return {"X-Bunker-Token": DATABUNKER_TOKEN}

@app.route('/signup', methods=['POST'])
def signup():
    # 1. Store PII in Databunker
    user_data = {
        "email": request.json["email"],
        "name": request.json["name"],
        "phone": request.json["phone"]
    }
    resp = requests.post(
        f"{DATABUNKER_URL}/v1/user",
        headers=databunker_headers(),
        json=user_data
    )
    user_token = resp.json()["token"]
    
    # 2. Store token in your database
    db.execute(
        "INSERT INTO users (token, created_at) VALUES (?, NOW())",
        (user_token,)
    )
    
    return {"status": "ok", "user_token": user_token}

@app.route('/profile/<user_token>', methods=['GET'])
def get_profile(user_token):
    # Fetch PII from Databunker
    resp = requests.get(
        f"{DATABUNKER_URL}/v1/user/token/{user_token}",
        headers=databunker_headers()
    )
    return resp.json()

@app.route('/user/delete/<user_token>', methods=['DELETE'])
def delete_user(user_token):
    # GDPR erasure
    requests.delete(
        f"{DATABUNKER_URL}/v1/user/token/{user_token}",
        headers=databunker_headers()
    )
    # Keep token in DB (for audit trail)
    db.execute("UPDATE users SET deleted_at = NOW() WHERE token = ?", (user_token,))
    return {"status": "deleted"}
```

---

## Comparison: Traditional vs Databunker

| Security Aspect | Traditional DB | Databunker |
|----------------|----------------|------------|
| Encryption-at-rest | ✅ Disk-level | ✅ Field-level |
| SQL injection protection | ❌ Vulnerable | ✅ Immune |
| GraphQL bulk queries | ❌ Vulnerable | ✅ Blocked |
| Insider threat | ❌ DBA has full access | ✅ Audit trail required |
| GDPR erasure | ❌ Manual (slow) | ✅ Instant API call |
| Consent management | ❌ Roll your own | ✅ Built-in |
| Audit trail | ❌ External tool | ✅ Automatic |
| Backup security | ⚠️ Plaintext risk | ✅ Encrypted blobs |

---

## Further Reading

- **Official Docs:** https://databunker.org/doc/
- **API Reference:** [API.md](./API.md)
- **Installation Guide:** [INSTALLATION.md](./INSTALLATION.md)
- **Source Code:** https://github.com/securitybunker/databunker

---

**Questions?** Open an issue or ask in [Discussions](https://github.com/securitybunker/databunker/discussions)!
