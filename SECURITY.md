# Security Policy

## Supported Versions

Only the latest release of Wellspring is supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest | :x:               |

## Reporting a Vulnerability

**Do NOT open public issues for security vulnerabilities.**

If you discover a security vulnerability in Wellspring, please report it responsibly:

1. **Preferred:** Use [GitHub Security Advisories](https://github.com/Real-Fruit-Snacks/Wellspring/security/advisories/new) to create a private report.
2. **Alternative:** Email the maintainers directly with details of the vulnerability.

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Affected versions
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgment:** Within 48 hours of receipt
- **Assessment:** Within 7 days
- **Fix & Disclosure:** Within 90 days (coordinated responsible disclosure)

We follow a 90-day responsible disclosure timeline. If a fix is not released within 90 days, the reporter may disclose the vulnerability publicly.

## What is NOT a Vulnerability

Wellspring is a payload delivery server designed for authorized red team engagements. The following behaviors are **features, not bugs**:

- Token-gated payload delivery over TLS
- AES-256-GCM encryption of payloads at rest in memory
- Memory zeroing of payloads, tokens, and cryptographic keys on shutdown
- Decoy nginx responses for invalid requests
- Fileless execution via memfd_create and /dev/shm staging
- HMAC-SHA256 keyed token hashing to prevent timing attacks
- Stager generation for 12 delivery methods
- Raw TLS listener for socat-based delivery

These capabilities exist by design for legitimate security testing. Reports that simply describe Wellspring working as intended will be closed.

## Responsible Use

Wellspring is intended for authorized penetration testing, security research, and educational purposes only. Users are responsible for ensuring they have proper authorization before using this tool against any systems.
