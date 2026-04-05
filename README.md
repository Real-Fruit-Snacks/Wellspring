<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-light.svg">
  <img alt="Wellspring" src="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-dark.svg" width="520">
</picture>

![Go](https://img.shields.io/badge/language-Go-00ADD8.svg)
![Platform](https://img.shields.io/badge/platform-Linux-lightgrey)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

**Payload delivery server for authorized red team engagements.**

Single Go binary with zero external dependencies. Token-gated TLS delivery with 12 loader methods, AES-256-GCM encryption at rest, memory zeroing on shutdown, and decoy nginx infrastructure.

> **Authorization Required**: Designed exclusively for authorized security testing with explicit written permission.

</div>

---

## Quick Start

### Prerequisites

- **Go** 1.21+ (build toolchain)
- **Make** (build automation)

### Build

```bash
git clone https://github.com/Real-Fruit-Snacks/Wellspring
cd Wellspring
make build    # -> build/wellspring
```

### Verify

```bash
# Generate TLS cert
./build/wellspring -gencert -sni cloudflare-dns.com

# Start server
./build/wellspring -cert server.crt -key server.key

# Interactive console:
#   payload add ./implant --name undertow
#   generate p1 all --host 10.0.0.1 --port 443
```

---

## Features

### Token-Gated Access

Every payload requires a valid token with configurable TTL, max uses, and source IP/CIDR lock. Tokens use HMAC-SHA256 keyed hashing to prevent timing attacks.

```
generate p1 socat-tls --host 10.0.0.1 --port 4443 --ttl 1h --single-use --source-lock 10.0.0.0/24
```

### 12 Delivery Methods

Socat TLS, socat TCP, socat memfd, socat PTY, curl, wget, bash /dev/tcp, python urllib, netcat, perl IO::Socket::SSL, memfd_create fileless, and /dev/shm tmpfs staging.

```
generate p1 all --host 10.0.0.1 --port 443
```

### Encrypt at Rest

All payloads stored with AES-256-GCM encryption in memory. Plaintext is zeroed after delivery with `runtime.KeepAlive` to prevent compiler elision.

```go
// Encryption keys and HMAC keys are zeroed on exit
```

### Dual TLS Listeners

HTTPS on `:443` serves payloads at `GET /p/<token>` for HTTP-based loaders. Raw TLS on `:4443` reads a token from the first line and writes raw payload bytes for socat connections.

```
-listen :443        # HTTPS listener
-raw-listen :4443   # Raw TLS listener (TLS 1.3 minimum)
```

### Fileless Execution

`memfd_create` loaders execute payloads entirely in memory with no file touching disk. `/dev/shm` staging uses tmpfs for environments where memfd is unavailable.

```
generate p1 memfd --host 10.0.0.1 --port 443
generate p1 devshm --host 10.0.0.1 --port 443
```

### Anti-Forensics

SIGINT/SIGTERM trigger immediate zeroing of all payloads, tokens, and cryptographic keys before exit. Background goroutine enforces TTL purging with sensitive field zeroing.

```
token revoke <value>   # Zeros token in-place via unsafe.StringData
exit                   # Shutdown + zero all payloads, tokens, keys
```

### Decoy Infrastructure

All HTTP responses use `Server: nginx/1.24.0` header. Invalid tokens and unknown paths return an identical nginx 404 page, making the server indistinguishable from a default nginx installation.

```
-decoy ./custom-404.html   # Optional custom decoy page
```

### Stager Generation

Generate ready-to-use one-liner stagers for any loaded payload, or use the cheatsheet command for a standalone technique reference card.

```
generate p1 curl --host 10.0.0.1 --port 443 --max-uses 5
cheatsheet --host 10.0.0.1 --port 443
```

---

## Architecture

```
cmd/wellspring/main.go            Entry point, flags, server+CLI init
internal/
  server/
    server.go                     Dual TLS listener setup
    handlers.go                   /p/<token> delivery, raw TLS, decoy
    tls.go                        ECDSA P-256 self-signed cert generation
  payload/
    payload.go                    ELF/PE arch detection
    manager.go                    Add/remove/list with encrypt-at-rest
    token.go                      Token generation, validation, TTL
    encrypt.go                    AES-256-GCM encryption
  loader/
    loader.go                     Loader interface, input validation
    registry.go                   Auto-registration via init()
    *.go                          12 delivery method implementations
  callback/tracker.go             In-memory delivery event log
  antiforensics/expiry.go         Background TTL enforcer, memory zeroing
  cli/
    cli.go                        Interactive operator console
    banner.go                     ASCII art banner
    colors.go                     Catppuccin Mocha palette
  theme/theme.go                  Shared ANSI color constants
```

Single Go binary with no external dependencies. All cryptographic operations use Go stdlib (`crypto/aes`, `crypto/cipher`, `crypto/ecdsa`). The server runs two concurrent TLS listeners with a shared payload manager and token store.

---

## Platform Support

| Feature | Requirement | Notes |
|---------|-------------|-------|
| HTTPS listener | Any Linux | TLS 1.3 minimum, ECDSA P-256 |
| Raw TLS listener | Any Linux | Token-from-first-line protocol |
| AES-256-GCM encryption | Any Linux | Go stdlib crypto |
| memfd_create loaders | Kernel 3.17+ | Fileless execution |
| /dev/shm staging | Any Linux | tmpfs available on all standard Linux |
| Memory zeroing | Any Linux | runtime.KeepAlive prevents elision |

---

## Console Commands

```
Payload Management
  payload add <path> [--name N]       Load a binary (auto-detects OS/arch)
  payload list                        Show loaded payloads
  payload remove <id>                 Remove and zero memory

Stager Generation
  generate <id> <loader> [flags]      Generate one-liner stager
  generate <id> all [flags]           Generate ALL compatible stagers
    --host IP  --port PORT            Server address
    --ttl 1h  --single-use            Token constraints
    --max-uses N  --source-lock CIDR  Access restrictions

Information
  loaders                             List delivery methods
  tokens                              List active tokens
  token revoke <value>                Revoke a token
  callbacks                           Show delivery log
  cheatsheet --host IP --port PORT    Technique reference

Control
  exit                                Shutdown + zero all payloads
```

---

## Delivery Methods

| Loader | Requires | Description |
|--------|----------|-------------|
| `socat-tls` | socat | OPENSSL pull + exec (raw TLS) |
| `socat-tcp` | socat | TCP pipe (no TLS) |
| `socat-memfd` | socat, python3 | socat + memfd_create (fileless) |
| `socat-pty` | socat | PTY allocation for interactive bootstrap |
| `curl` | curl | HTTPS pull + exec |
| `wget` | wget | HTTPS pull + exec |
| `bash-devtcp` | bash | /dev/tcp raw pull (no external tools) |
| `python` | python3 | urllib HTTPS pull + exec |
| `netcat` | nc | Raw pull (no TLS) |
| `perl` | perl | IO::Socket::SSL HTTPS pull + exec |
| `memfd` | curl, python3 | curl + memfd_create fileless execution |
| `devshm` | curl | /dev/shm staging (tmpfs, no disk write) |

---

## Security

Report vulnerabilities via [SECURITY.md](SECURITY.md) -- do not open public issues.

Wellspring does **not**:

- Provide command and control functionality
- Exploit any vulnerability
- Provide persistence mechanisms
- Bypass endpoint detection
- Exfiltrate data

---

## License

[MIT](LICENSE) -- Copyright 2026 Real-Fruit-Snacks
