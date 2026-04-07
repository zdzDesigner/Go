# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-19
**Commit:** 
**Branch:** 

## OVERVIEW
Go MQTT Client implementation with support for TLS connections and full MQTT protocol operations (connect, subscribe, publish, unsubscribe, ping). Includes proper QoS handling and connection management.

## STRUCTURE
```
./
├── main.go          # Main client implementation
├── packet.go        # MQTT packet encoding/decoding
├── topic.go         # Topic definitions and types
├── packet_test.go   # Unit tests
├── old/             # Legacy implementations (reference only)
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
└── AGENTS.md        # Current guidelines
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Core MQTT logic | main.go | Connect, subscribe, publish, unsubscribe |
| Packet handling | packet.go | Encoding/decoding, QoS handling |
| Topic management | topic.go | Topic definitions and struct |
| TLS implementation | main.go | TLS connection methods |
| Unit tests | packet_test.go | Tests for packet functions |
| Legacy code | old/ | Historical implementations |

## CODE MAP



## CONVENTIONS
- Use 4-space indentation
- Group imports: std lib → third-party → project-local
- MQTT message types as constants (CONNECT, CONNACK, PUBLISH, etc.)
- Use minimum TLS 1.2 for secure connections
- Packet IDs should be unique and avoid zero values
- Handle different QoS levels appropriately

## ANTI-PATTERNS (THIS PROJECT)
- Do NOT use files in 'old/' directory in production builds
- Do NOT set InsecureSkipVerify in production
- Do NOT ignore error returns from network operations
- Do NOT publish without proper topic validation

## UNIQUE STYLES
- Channel-based message handling for non-blocking operations
- Separate Packet struct for MQTT protocol handling
- Custom Topic struct with QoS, Dup, Retain fields

## COMMANDS
```bash
# Build project
go build -o mqtt_client .

# Run directly
go run .

# Run tests
go test ./...

# Format code
go fmt ./...

# Vet for issues
go vet ./...
```

## NOTES
- TLS connection uses InsecureSkipVerify (only for testing)
- QoS 2 not fully implemented in current version
- TODO: Add QOS/DUP/Retain handling for PUBLISH packets in packet.go
- Legacy files in 'old/' directory have conflicting constant declarations