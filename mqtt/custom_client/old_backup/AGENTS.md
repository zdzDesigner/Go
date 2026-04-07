# OLD/ Directory Knowledge Base

**Generated:** 2026-02-19
**Purpose:** Legacy MQTT client implementations

## OVERVIEW
This directory contains historical MQTT client implementations that served as development prototypes. These files are kept for reference but should not be used in production.

## STRUCTURE
```
./old/
├── main.go      # Initial MQTT client implementation
├── main_1.go    # Enhanced version with improved buffering
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Basic MQTT impl | main.go | Simple connection, publish, subscribe |
| Enhanced impl | main_1.go | Better buffer management, TLS support |

## CONVENTIONS
- Uses separate MQTTClient struct in main_1.go vs direct methods in main.go
- Both implementations use similar packet encoding/decoding logic
- main_1.go has more robust error handling and connection management

## ANTI-PATTERNS (THIS DIRECTORY)
- Do NOT use these files as the primary implementation
- Do NOT copy the TLS implementation without careful review
- Do NOT expect these to match current main.go API

## UNIQUE STYLES
- More verbose error handling compared to main implementation
- Different approach to channel-based message buffering
- Alternative packet handling strategies for comparison

## COMMANDS
```bash
# Compare implementations
diff main.go main_1.go
```

## NOTES
- Contains duplicate constant definitions (MQTT message types) that conflict with main implementation
- Both files lack proper connection cleanup in some error scenarios
- Keep for reference but avoid using code directly in new development
- Some features from these implementations have been incorporated into the current main implementation

## Improvements Incorporated
- Enhanced length decoding function from main_1.go has been integrated into current implementation
- Asynchronous message processing concepts from main_1.go influenced current receiver design
- Robust buffer management techniques have been adapted for improved channel handling

## Documentation
- English version: AGENTS.md (this file)
- Chinese version: AGENTS_CN.md