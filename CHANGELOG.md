# Changelog

All notable changes to Wellspring will be documented in this file.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-04-04

### Added
- Dual TLS listeners: HTTPS on :443 and raw TLS on :4443
- Token-gated payload access with configurable TTL, max uses, and source IP/CIDR lock
- HMAC-SHA256 keyed token hashing to prevent timing attacks
- AES-256-GCM encryption of all payloads at rest in memory
- 12 delivery methods: socat-tls, socat-tcp, socat-memfd, socat-pty, curl, wget, bash-devtcp, python, netcat, perl, memfd, devshm
- Stager generation with per-loader one-liner output
- Fileless execution via memfd_create and /dev/shm staging
- Interactive operator console with Catppuccin Mocha theme
- Memory zeroing of payloads, tokens, and keys on SIGINT/SIGTERM
- Decoy nginx/1.24.0 responses for invalid requests
- Background TTL enforcer with automatic token expiry
- Shell injection prevention via per-field character whitelist
- TLS 1.3 minimum with session tickets disabled
- Connection limiting via semaphore (128 concurrent raw TLS)
- Cheatsheet command for standalone technique reference
- 53 tests covering encryption, tokens, delivery, and anti-forensics
