## gRPC PubSub Service
A gRPC-based publish-subscribe service implementation using the subpub package.

## Features
  - gRPC API for subscribing to and publishing events
  - Graceful shutdown
  - Dependency injection 
  - Logging
  - Thread-safe message handling

# Run directly
```bash
go run cmd/server/main.go #run server
```
## Run with docker-compose
```bash 
docker-compose -f deployments/docker-compose.yml up --build
```

# Api Documentation
```bash
go run cmd/client/main.go -mode sub -key test-key # run client in subscriber mode
go run cmd/client/main.go -mode pub -key test-key -message "Hello World!" # run client in publisher mode
go run cmd/client/main.go -mode both -key test-key -message "Hello World!" # run in client subscriber and publisher mode
```
## Flags
    -addr: Server address (default: localhost:50051)
    -mode: Client mode: sub, pub, or both (default: both)
    -key: Subscription/publish key (default: test-key)
    -message: Message to publish (default: Hello!)

# Configuration
```
{
    "server": {
        "host": "0.0.0.0",
        "port": "50051"
    },
    "log": {
        "level": "info",
        "file": "pubsub.log"
    }
}
```
