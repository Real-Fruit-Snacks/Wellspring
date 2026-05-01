<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-light.svg">
  <img alt="Wellspring" src="https://raw.githubusercontent.com/Real-Fruit-Snacks/Wellspring/main/docs/assets/logo-dark.svg" width="100%">
</picture>

> [!IMPORTANT]
> **Payload delivery server for authorized red team engagements.** Single Go binary with token-gated TLS delivery, 12 loader methods, AES-256-GCM encryption at rest, memory zeroing on shutdown, and decoy nginx infrastructure.

> *A wellspring is an abundant natural source that flows continuously, providing steady sustenance to those who need it. Perfect metaphor for a payload delivery server—a reliable, controlled source that provides the tools red teams need for authorized security testing operations.*

---

## §1 / Premise

Wellspring is a **payload delivery server** designed for authorized red team engagements with secure, token-gated access control. Deploy a single Go binary that serves payloads via 12 different loader methods with dual TLS listeners and comprehensive anti-forensics measures.

The server implements AES-256-GCM encryption at rest, HMAC-SHA256 token validation, configurable TTL and IP constraints, memory zeroing on shutdown, and nginx decoy infrastructure. All operations use Go standard library with zero external dependencies.

**Authorization Required**: Designed exclusively for authorized security testing with explicit written permission.

---

## §2 / Specs

| KEY             | VALUE                                                                       |
|-----------------|-----------------------------------------------------------------------------|
| **DELIVERY**    | **12 loader methods** — socat TLS/TCP, curl, wget, bash, python, memfd |
| **SECURITY**    | **Token-gated access** with HMAC-SHA256, TTL, IP locks, max-use limits |
| **CRYPTO**      | **AES-256-GCM encryption** at rest with memory zeroing on shutdown |
| **LISTENERS**   | **Dual TLS servers** — HTTPS :443 and raw TLS :4443 with TLS 1.3 minimum |
| **STEALTH**     | **Nginx decoy infrastructure** with identical 404 pages and headers |
| **BINARY**      | **Single Go executable** with zero external dependencies |
| **FILELESS**    | **memfd_create execution** entirely in memory with no disk writes |
| **PLATFORM**    | **Linux deployment** with Kernel 3.17+ for memfd support |

---

## §3 / Quickstart

**Prerequisites:** Go 1.21+, Linux deployment target

```bash
# Build static binary
git clone https://github.com/Real-Fruit-Snacks/Wellspring
cd Wellspring && make build

# Generate TLS certificate with SNI decoy
./build/wellspring -gencert -sni cloudflare-dns.com

# Start token-gated payload server
./build/wellspring -cert server.crt -key server.key

# Interactive console operations
# payload add ./implant --name undertow
# generate p1 all --host 10.0.0.1 --port 443
```

---

## §4 / Reference

```bash
# SERVER STARTUP
wellspring -cert server.crt -key server.key              # Dual TLS listeners
wellspring -listen :443 -raw-listen :4443                # Custom ports
wellspring -decoy ./custom-404.html                      # Custom nginx decoy

# PAYLOAD MANAGEMENT
payload add ./implant --name undertow                     # Load binary with auto-detect
payload list                                              # Show loaded payloads
payload remove p1                                         # Remove and zero memory

# TOKEN-GATED GENERATION
generate p1 socat-tls --host 10.0.0.1 --port 4443        # Raw TLS delivery
generate p1 curl --host 10.0.0.1 --port 443             # HTTPS delivery
generate p1 memfd --host 10.0.0.1 --port 443             # Fileless execution
generate p1 all --host 10.0.0.1 --port 443              # All compatible loaders

# ACCESS CONTROL
generate p1 curl --ttl 1h --single-use                   # TTL and use limits
generate p1 wget --max-uses 5 --source-lock 10.0.0.0/24 # IP/CIDR restrictions

# OPERATIONS
tokens                                                    # List active tokens
token revoke <value>                                      # Revoke specific token
callbacks                                                 # Show delivery events
cheatsheet --host 10.0.0.1 --port 443                   # Technique reference
```

---

## §5 / Architecture

**Three-Layer Design**: HTTP/TLS server layer → Payload manager with encryption → Interactive CLI console

```
cmd/wellspring/
├── main.go         # Entry point, flags, server+CLI init
└── internal/
    ├── server/     # Dual TLS listener infrastructure
    │   ├── server.go    # HTTPS :443 and raw TLS :4443 setup
    │   ├── handlers.go  # /p/<token> delivery, decoy nginx responses
    │   └── tls.go       # ECDSA P-256 self-signed cert generation
    ├── payload/    # Encrypted payload management
    │   ├── manager.go   # Add/remove/list with AES-256-GCM at rest
    │   ├── token.go     # HMAC-SHA256 validation, TTL, IP constraints
    │   └── encrypt.go   # Memory encryption with zeroing on exit
    ├── loader/     # 12 delivery method implementations
    │   ├── registry.go  # Auto-registration via init() functions
    │   └── *.go         # socat, curl, wget, bash, python, memfd
    ├── antiforensics/
    │   └── expiry.go    # Background TTL enforcer, memory zeroing
    └── cli/        # Interactive operator console
        ├── cli.go       # Command processing with Catppuccin Mocha
        └── banner.go    # ASCII art banner and colors
```

**Token Security**: HMAC-SHA256 keyed hashing prevents timing attacks, configurable TTL and max-use limits, source IP/CIDR locking with immediate revocation capability.

---

## §6 / Authorization

Wellspring is designed for **authorized red team engagements** with explicit written permission. The tool generates and delivers payloads that should only be used against systems you own or have proper authorization to test.

Security vulnerabilities should be reported via [GitHub Security Advisories](https://github.com/Real-Fruit-Snacks/Wellspring/security/advisories) with responsible disclosure.

**Wellspring does not**: provide command and control functionality, exploit any vulnerability, provide persistence mechanisms, bypass endpoint detection, or exfiltrate data—it's purely a payload delivery mechanism.

---

**Real-Fruit-Snacks** — [All projects](https://real-fruit-snacks.github.io/) · [Security](https://github.com/Real-Fruit-Snacks/Wellspring/security/advisories) · [License](LICENSE)