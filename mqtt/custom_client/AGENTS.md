# AGENTS.MD - Go MQTT Client Development Guidelines

## Project Overview
This is a custom MQTT client implementation in Go that handles MQTT protocol communication with support for TLS connections. The client implements core MQTT operations like connect, subscribe, publish, unsubscribe, and ping.

## Build Commands

### Basic Building
```bash
# Build the project
go build -o mqtt_client .

# Run directly
go run .
```

### Testing Commands
```bash
# Run all tests
go test ./...

# Run specific test file
go test -v packet_test.go

# Run individual test functions
go test -run TestConnectPacket

# Build test binary
go test -c
```

### Formatting and Linting
```bash
# Format Go files
go fmt ./...

# Vet code for common issues
go vet ./...

# Run linter if installed
golangci-lint run ./...

# Generate documentation
go doc
```

## Code Style Guidelines

### Imports
- Group imports with standard library first, then third-party, then project-local
- Use explicit import paths
- Use blank imports only for package initialization side effects

### Naming Conventions
- Use camelCase for function/method names
- Use PascalCase for exported functions/types
- Use SCREAMING_SNAKE_CASE for constants
- Use descriptive names; avoid abbreviations except well-known ones like "id", "err"

### Error Handling
- Always check errors and handle them appropriately
- Use wrapped errors with %w to maintain context
- Use errors.New() for simple errors and fmt.Errorf() for formatted errors
- Log errors with contextual information

### Type Definitions
- Define custom types for domain-specific concepts
- Use struct{} for empty structs
- Use type aliases sparingly
- Export types that are part of the public API

### Struct Organization
- Group related methods together
- Use pointer receivers for methods that modify the struct
- Use value receivers for methods that don't modify the struct
- Initialize struct fields with meaningful zero values when possible

### Formatting Standards
- Use 4-space indentation (not tabs)
- Limit line length to 100 characters where practical
- Group related constants together
- Comment exported functions/types
- Use // for inline comments and /* */ for block comments when necessary

### Concurrency Patterns
- Use channels for communication between goroutines
- Use sync.WaitGroup for coordinating goroutine completion
- Use mutexes for protecting shared data access
- Consider using context for cancellation and timeouts

### MQTT Protocol Specifics
- MQTT message types are defined as constants (CONNECT, CONNACK, PUBLISH, etc.)
- Packet IDs should be unique and avoid zero values
- Handle different QoS levels appropriately
- Implement proper length encoding/decoding for MQTT packets
- Maintain connection state properly

### TLS Security
- Use minimum TLS version 1.2
- Validate certificates in production (avoid InsecureSkipVerify in production)
- Set appropriate ServerName for certificate validation
- Implement proper TLS handshake verification

### Package Structure
- Main package contains client implementation and connection logic
- Related types and constants grouped logically
- Protocol-specific logic encapsulated in appropriate functions
- Error handling centralized where possible

## Testing Guidelines

### Unit Tests
- Test individual functions with table-driven tests where appropriate
- Test edge cases and error conditions
- Mock external dependencies where needed
- Ensure code coverage for critical paths

### Integration Tests
- Test actual MQTT broker connections in controlled environments
- Verify packet encoding/decoding correctness
- Test multiple concurrent subscriptions and publications
- Test connection lifecycle (connect, subscribe, publish, disconnect)

## Common Development Tasks

### Adding MQTT Features
1. Define new constants for MQTT message types if needed
2. Update packet processing logic
3. Add corresponding handler functions
4. Test thoroughly with real MQTT broker

### Updating Connection Logic
1. Modify connect/disconnect methods as needed
2. Update error handling for connection-related issues
3. Ensure proper cleanup and resource deallocation
4. Test with different broker configurations

### Modifying Packet Handling
1. Review MQTT specification for correct packet format
2. Update encoding/decoding logic accordingly
3. Update related tests
4. Verify compatibility with existing functionality

## Environment Setup
- Go 1.21+ required
- Access to MQTT broker for testing (local or remote)
- Optional: TLS certificates for secure connections
- Recommended: MQTT client tools for manual testing (mosquitto_pub/sub)

## Code Quality Checks
- Run go fmt before committing
- Verify tests pass before pushing
- Check for race conditions with "go run -race"
- Review error handling for completeness

## Known Issues & Maintenance
- Removed files from 'old/' directory due to duplicate constant and function declarations that conflicted with main implementation
- The packet_test.go file calls a non-existent method 'createConnectPacket' - this needs to be updated to use the correct method name
- Files were moved to 'old_backup/' directory to prevent compilation conflicts
- Ensure all test files properly import necessary functions and match current API
- Implemented: QOS/DUP/Retain handling for PUBLISH packets - was a previous TODO item that has now been completed

## Improvements Made
- Enhanced length decoding function in packet.go with improved error handling
- Added support for QoS 2 PUBLISH messages with PUBREC/PUBREL/PUBCOMP flow
- Improved message processing in receiver() with better buffering and channel handling
- Added missing MQTT constants (PUBREC, PUBREL, PUBCOMP) to support full QoS flows
- Enhanced PUBLISH packet creation to handle QoS, DUP, and RETAIN flags properly
- Improved asynchronous message handling to prevent blocking in receiver loop
- Integrated valuable features from old_backup/main_1.go: enhanced length decoding, improved buffer management

## Documentation
- English version: AGENTS.md (this file)
- Chinese version: AGENTS_CN.md
- Legacy code docs: old_backup/AGENTS.md, old_backup/AGENTS_CN.md