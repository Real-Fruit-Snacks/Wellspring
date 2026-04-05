<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-light.svg">
  <img alt="Wellspring" src="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-dark.svg" width="520">
</picture>

![Go](https://img.shields.io/badge/language-Go-00ADD8.svg)
![Platform](https://img.shields.io/badge/platform-Linux-lightgrey)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

**Payload delivery server for authorized red team engagements**

Single Go binary, zero external dependencies. Token-gated TLS delivery with 12 loader methods, AES-256-GCM encryption at rest, memory zeroing on shutdown, and decoy nginx infrastructure. Socat is the primary delivery method with fallbacks for curl, wget, python, netcat, perl, and fileless (memfd) execution.

> **Authorization Required**: This tool is designed exclusively for authorized security testing with explicit written permission. Unauthorized access to computer systems is illegal and may result in criminal prosecution.

[Quick Start](#quick-start) • [Delivery Methods](#delivery-methods) • [Architecture](#architecture) • [Security](#security)

</div>

---

## Highlights

<table>
<tr>
<td width="50%">

**Token-Gated Access**
Every payload requires a valid token with configurable TTL, max uses, and source IP/CIDR lock. Tokens are HMAC-SHA256 keyed to prevent timing attacks on map lookup. Expired and exhausted tokens are purged automatically with sensitive field zeroing.

</td>
<td width="50%">

**12 Delivery Methods**
Socat TLS, socat TCP, socat memfd, socat PTY, curl, wget, bash /dev/tcp, python urllib, netcat, perl IO::Socket::SSL, memfd_create fileless, and /dev/shm tmpfs staging. Every loader validates inputs against a strict per-field character whitelist.

</td>
</tr>
<tr>
<td width="50%">

**Encrypt at Rest**
All payloads stored with AES-256-GCM encryption in memory. Plaintext is zeroed after delivery with `runtime.KeepAlive` to prevent compiler elision. Encryption keys and HMAC keys are zeroed on exit.

</td>
<td width="50%">

**Dual TLS Listeners**
HTTPS on `:443` serves payloads at `GET /p/<token>` for HTTP-based loaders. Raw TLS on `:4443` reads a token from the first line and writes raw payload bytes for `socat OPENSSL:` connections. TLS 1.3 minimum, no downgrade.

</td>
</tr>
<tr>
<td width="50%">

**Fileless Execution**
`memfd_create` loaders execute payloads entirely in memory -- no file touches disk. `/dev/shm` staging uses tmpfs for environments where memfd is unavailable. Both methods leave zero filesystem artifacts.

</td>
<td width="50%">

**Anti-Forensics**
SIGINT/SIGTERM trigger immediate zeroing of all payloads, tokens, and cryptographic keys before exit. Token values are zeroed in-place via `unsafe.StringData` on revocation, expiry, and shutdown. Background goroutine enforces TTL purging.

</td>
</tr>
<tr>
<td width="50%">

**Decoy Infrastructure**
All HTTP responses use `Server: nginx/1.24.0` header. Invalid tokens and unknown paths return an identical nginx 404 decoy page. The server is indistinguishable from a default nginx installation to casual inspection.

</td>
<td width="50%">

**53 Tests**
Payload encryption, token lifecycle, arch detection, shell injection prevention, HTTPS delivery, raw TLS delivery, server header consistency, ring buffer compaction, and anti-forensics coverage with race detection.

</td>
</tr>
</table>

---

## Quick Start

### Prerequisites

<table>
<tr>
<th>Requirement</th>
<th>Version</th>
<th>Purpose</th>
</tr>
<tr>
<td>Go</td>
<td>1.21+</td>
<td>Build toolchain</td>
</tr>
<tr>
<td>Make</td>
<td>Any</td>
<td>Build automation</td>
</tr>
</table>

### Build

```bash
# Clone
git clone https://github.com/Real-Fruit-Snacks/Wellspring
cd Wellspring

# Build static binary
make build    # -> build/wellspring
```

### Verification

```bash
# Generate TLS cert
./build/wellspring -gencert -sni cloudflare-dns.com

# Start server
./build/wellspring -cert server.crt -key server.key

# Interactive console opens:
#   payload add ./implant --name undertow
#   generate p1 all --host 10.0.0.1 --port 443
```

---

## Usage

### Startup Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-listen` | `:443` | HTTPS listener address |
| `-raw-listen` | `:4443` | Raw TLS listener address |
| `-cert` / `-key` | (auto-gen) | TLS certificate and key paths |
| `-sni` | `cloudflare-dns.com` | SNI/CN for auto-generated cert |
| `-decoy` | (built-in) | Path to custom HTML decoy page |
| `-gencert` | | Generate TLS cert and exit |

### Console Commands

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
  cheatsheet --host IP --port PORT    Delivery technique reference

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
| `bash-devtcp` | bash | /dev/tcp raw pull (no external tools, no TLS) |
| `python` | python3 | urllib HTTPS pull + exec |
| `netcat` | nc | Raw pull (no TLS) |
| `perl` | perl | IO::Socket::SSL HTTPS pull + exec |
| `memfd` | curl, python3 | curl + memfd_create fileless execution |
| `devshm` | curl | /dev/shm staging (tmpfs, no disk write) |

---

## Stager Generation

Generate stagers for a loaded payload:

```bash
# Single loader
generate p1 curl --host 10.0.0.1 --port 443

# All compatible loaders at once
generate p1 all --host 10.0.0.1 --port 443

# With access restrictions
generate p1 socat-tls --host 10.0.0.1 --port 4443 --ttl 1h --single-use --source-lock 10.0.0.0/24

# Multiple uses with CIDR lock
generate p1 memfd --host 10.0.0.1 --port 443 --max-uses 5 --source-lock 192.168.1.0/24
```

### Cheatsheet

Generate a standalone technique reference card -- no payloads, no tokens, no server required:

```bash
cheatsheet --host 10.0.0.1 --port 443
```

Outputs all 12 delivery techniques as ready-to-use one-liners organized by category (Socat TLS, Socat TCP, HTTP/HTTPS, Raw TCP, Fileless) with host and port substituted in.

---

## Architecture

### Dual TLS Listeners

- **HTTPS** (`:443`) -- Serves payloads at `GET /p/<token>` for HTTP-based loaders (curl, wget, python). All other paths return a decoy nginx 404 page.
- **Raw TLS** (`:4443`) -- After TLS handshake, reads token from first line, writes raw payload bytes. For `socat OPENSSL:` connections.

### Request Flow

```
Operator                    Wellspring                    Target
   |                           |                            |
   |-- payload add ----------> |                            |
   |-- generate p1 curl -----> |                            |
   |   <-- one-liner stager -- |                            |
   |                           |                            |
   |   (operator sends stager to target via C2)             |
   |                           |                            |
   |                           | <-- GET /p/<token> ------- |
   |                           | -- validate token -->      |
   |                           | -- decrypt payload -->     |
   |                           | -- deliver + zero -------->|
   |                           |                            |
```

### Project Structure

```
cmd/wellspring/main.go          Entry point, flags, server+CLI init
internal/
  server/
    server.go                   Dual TLS listener setup
    handlers.go                 /p/<token> delivery, raw TLS handler, decoy
    tls.go                      ECDSA P-256 self-signed cert generation
  payload/
    payload.go                  ELF/PE arch detection
    manager.go                  Add/remove/list/get with encrypt-at-rest
    token.go                    Token generation, validation, TTL, source-lock
    encrypt.go                  AES-256-GCM encryption
  loader/
    loader.go                   Loader interface, input validation
    registry.go                 Auto-registration via init()
    *.go                        12 delivery method implementations
  callback/tracker.go           In-memory delivery event log
  antiforensics/expiry.go       Background TTL enforcer, memory zeroing
  cli/
    cli.go                      Interactive operator console
    banner.go                   ASCII art banner
    colors.go                   Catppuccin Mocha palette
  theme/theme.go                Shared ANSI color constants
```

---

## Configuration

### Build Targets

| Target | Command | Output |
|--------|---------|--------|
| Static binary | `make build` | `build/wellspring` |
| Run tests | `make test` | Race-detected test suite |
| Clean | `make clean` | Remove build artifacts |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Clean shutdown (SIGINT/SIGTERM or `exit` command) |
| `1` | Startup failure (missing cert, port in use) |

---

## Platform Support

<table>
<tr>
<th>Feature</th>
<th>Linux</th>
<th>Notes</th>
</tr>
<tr>
<td>HTTPS listener</td>
<td>Any</td>
<td>TLS 1.3 minimum, ECDSA P-256</td>
</tr>
<tr>
<td>Raw TLS listener</td>
<td>Any</td>
<td>Token-from-first-line protocol</td>
</tr>
<tr>
<td>AES-256-GCM encryption</td>
<td>Any</td>
<td>Go stdlib crypto/aes + crypto/cipher</td>
</tr>
<tr>
<td>memfd_create loaders</td>
<td>Kernel 3.17+</td>
<td>Fileless execution requires memfd support</td>
</tr>
<tr>
<td>/dev/shm staging</td>
<td>Any</td>
<td>tmpfs available on all standard Linux</td>
</tr>
<tr>
<td>Memory zeroing</td>
<td>Any</td>
<td>runtime.KeepAlive prevents elision</td>
</tr>
</table>

---

## Testing

```bash
go test ./... -race -count=1 -v
```

53 tests covering payload encryption, token lifecycle, arch detection, shell injection prevention, HTTPS delivery, raw TLS delivery, server header consistency, ring buffer compaction, and anti-forensics. Requires Go 1.21+.

---

## Security

### Vulnerability Reporting

Do **not** open public issues for security vulnerabilities. See [SECURITY.md](SECURITY.md) for responsible disclosure instructions.

### Threat Model

Wellspring protects payloads during staging and delivery:

- **Token-gated access** -- Payloads require valid tokens with configurable TTL, max uses, and source IP/CIDR lock.
- **Encrypt at rest** -- All payloads stored with AES-256-GCM encryption in memory.
- **Memory zeroing** -- Plaintext zeroed after delivery with `runtime.KeepAlive` to prevent compiler elision. Tokens zeroed via `unsafe.StringData` on revocation, expiry, and shutdown.
- **Anti-timing** -- Token store uses HMAC-SHA256 keyed hashing to prevent timing attacks on map lookup.
- **Shell injection prevention** -- All loader inputs validated against a strict per-field character whitelist.
- **TLS 1.3 minimum** -- No downgrade to TLS 1.2. Session tickets disabled.
- **HTTP hardening** -- Read/write/idle timeouts prevent slowloris. Max header size capped at 64KB.
- **Connection limiting** -- Raw TLS listener capped at 128 concurrent connections via semaphore.
- **Consistent fingerprint** -- All HTTP responses use `Server: nginx/1.24.0` header. Invalid tokens return identical decoy page.
- **Signal handling** -- SIGINT/SIGTERM trigger immediate zeroing of all payloads, tokens, and keys.
- **Auto-expiry** -- Background goroutine purges expired/exhausted tokens with sensitive field zeroing.
- **Payload size limit** -- 100MB maximum enforced on load; only regular files accepted.

### What Wellspring Does NOT Do

- Does not provide command and control functionality
- Does not exploit any vulnerability
- Does not provide persistence mechanisms
- Does not bypass endpoint detection
- Does not exfiltrate data

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## Resources

- [Releases](https://github.com/Real-Fruit-Snacks/Wellspring/releases)
- [Issues](https://github.com/Real-Fruit-Snacks/Wellspring/issues)
- [Security Policy](https://github.com/Real-Fruit-Snacks/Wellspring/security/policy)
- [Contributing](CONTRIBUTING.md)

---

<div align="center">

**Part of the Real-Fruit-Snacks water-themed security toolkit**

[Aquifer](https://github.com/Real-Fruit-Snacks/Aquifer) • [Cascade](https://github.com/Real-Fruit-Snacks/Cascade) • [Conduit](https://github.com/Real-Fruit-Snacks/Conduit) • [Deadwater](https://github.com/Real-Fruit-Snacks/Deadwater) • [Deluge](https://github.com/Real-Fruit-Snacks/Deluge) • [Depth](https://github.com/Real-Fruit-Snacks/Depth) • [Dew](https://github.com/Real-Fruit-Snacks/Dew) • [Droplet](https://github.com/Real-Fruit-Snacks/Droplet) • [Fathom](https://github.com/Real-Fruit-Snacks/Fathom) • [Flux](https://github.com/Real-Fruit-Snacks/Flux) • [Grotto](https://github.com/Real-Fruit-Snacks/Grotto) • [HydroShot](https://github.com/Real-Fruit-Snacks/HydroShot) • [Maelstrom](https://github.com/Real-Fruit-Snacks/Maelstrom) • [Rapids](https://github.com/Real-Fruit-Snacks/Rapids) • [Ripple](https://github.com/Real-Fruit-Snacks/Ripple) • [Riptide](https://github.com/Real-Fruit-Snacks/Riptide) • [Runoff](https://github.com/Real-Fruit-Snacks/Runoff) • [Seep](https://github.com/Real-Fruit-Snacks/Seep) • [Shallows](https://github.com/Real-Fruit-Snacks/Shallows) • [Siphon](https://github.com/Real-Fruit-Snacks/Siphon) • [Slipstream](https://github.com/Real-Fruit-Snacks/Slipstream) • [Spillway](https://github.com/Real-Fruit-Snacks/Spillway) • [Surge](https://github.com/Real-Fruit-Snacks/Surge) • [Tidemark](https://github.com/Real-Fruit-Snacks/Tidemark) • [Tidepool](https://github.com/Real-Fruit-Snacks/Tidepool) • [Undercurrent](https://github.com/Real-Fruit-Snacks/Undercurrent) • [Undertow](https://github.com/Real-Fruit-Snacks/Undertow) • [Vapor](https://github.com/Real-Fruit-Snacks/Vapor) • [Wellspring](https://github.com/Real-Fruit-Snacks/Wellspring) • [Whirlpool](https://github.com/Real-Fruit-Snacks/Whirlpool)

*Remember: With great power comes great responsibility.*

</div>
