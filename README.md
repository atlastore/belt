# belt


`belt` is a collection of robust, high-performance Go packages designed to be a "toolbelt" for building modern, scalable backend services. It provides modules for common tasks such as creating servers, structured logging, hashing, and more, all with a focus on ease of use and extensibility.

## Features

-   **`server`**: A powerful server package for creating gRPC, HTTP, or multiplexed (gRPC & HTTP on the same port) servers with graceful shutdown, TLS support, and automatic request logging.
-   **`logx`**: A structured logger built on `zap` that intercepts logs and forwards them to a configurable transport. It supports development/production modes and enriches logs with service context.
-   **`router`**: An intuitive HTTP router built on `gofiber/fiber`, allowing for easy definition of routes, route groups, and global/local middleware.
-   **`hashx`**: A comprehensive hashing library supporting a wide range of algorithms (SHA, BLAKE, HMAC, etc.). It provides a unified API for keyed and unkeyed hashing, 32/64-bit hashing, and concurrent hashing of multiple algorithms.
-   **`key`**: A versioned key generation and decoding system, perfect for creating unique identifiers that embed information like node and disk IDs.
-   **`io/disk`**: A simple utility for querying disk space usage (total, used, free).

## Installation

To install `belt` and its packages, use `go get`:

```sh
go get github.com/atlastore/belt
```

## Usage

Here are some examples of how to use the core packages in `belt`.

### Creating a Multiplexed Server

The `server` package can serve both gRPC and HTTP traffic on a single port using `cmux`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/atlastore/belt/logx"
	"github.com/atlastore/belt/router"
	"github.com/atlastore/belt/server"
	"github.com/atlastore/belt/server/http"
	"github.com/google/uuid"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// 1. Set up the logger
	log := logx.New(
		ctx,
		logx.NewConfig(logx.Development, "my-app", uuid.NewString()),
		&logx.ConsoleTransport{},
	)
	defer log.Close()

	// 2. Create an HTTP router
	r := router.New()
	r.Add("/hello", router.GET, func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// 3. Configure the server options
	// In a real application, you would also add your gRPC service registries here.
	// For example: `grpc.WithRegistry(myService, pb.RegisterMyServiceServer)`
	opts := []server.Option{
		http.WithRouter(r),
	}

	// 4. Create and start the multiplexed server
	muxServer := server.NewServer(server.MUX, log, opts...)

	log.Info("Starting server on localhost:8080", zap.String("type", "mux"))

	if err := muxServer.Start(ctx, ":8080"); err != nil {
		log.Fatal("Server failed to start", zap.Error(err))
	}
}
```

### Structured Logging with `logx`

The `logx` package provides structured, leveled logging that can be sent to custom transports.

```go
package main

import (
	"context"
	"errors"
	"time"

	"github.com/atlastore/belt/logx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure logger for a "worker" service in production
	cfg := logx.NewConfig(logx.Production, "worker-service", uuid.NewString())
	logger := logx.New(ctx, cfg, &logx.ConsoleTransport{})
	defer logger.Close()

	logger.Info("Worker starting up", zap.Int("num_goroutines", 5))

	// Logs include structured context
	logger.Error(
		"Failed to process job",
		zap.String("job_id", "job-12345"),
		zap.Error(errors.New("connection timed out")),
	)
	
	// Create a child logger with additional fields
	childLogger := logger.With(zap.String("component", "database"))
	childLogger.Debug("Executing query", zap.Duration("query_time", 150*time.Millisecond))
}
```

### Hashing Data with `hashx`

The `hashx` package provides a simple API for a wide variety of hashing algorithms.

```go
package main

import (
	"fmt"

	"github.com/atlastore/belt/hashx"
    "github.com/atlastore/belt/hashx/multi"
)

func main() {
	data := "this is some important data to be hashed"

	// Simple, unkeyed hash
	sha256sum, err := hashx.HashString(hashx.SHA256, data)
	if err != nil {
		panic(err)
	}
	fmt.Println("SHA256 Hash:", sha256sum.Encode()) // "sha256:..."

	// Keyed hash
	secretKey := []byte("a-very-secret-key-that-is-32b") // Blake3 requires a 32-byte key
	blake3sum, err := hashx.HashString(hashx.Blake3, data, secretKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("Blake3 Hash:", blake3sum.Encode()) // "blake3:..."

	// 64-bit integer hash
	fnv64hash := hashx.FNV64.HashString(data)
	fmt.Printf("FNV64 Hash: %d\n", fnv64hash)

	// Hash multiple algorithms at once
	hashes, err := multi.HashString(data, hashx.MD5, hashx.SHA1, hashx.XXHash)
	if err != nil {
		panic(err)
	}
	fmt.Println("Multi-hash results:", hashes)
}
```

### Versioned Keys with `key`

The `key` package allows you to create and parse versioned keys, which is useful for object storage or distributed systems.

```go
package main

import (
	"fmt"

	"github.com/atlastore/belt/hashx"
	"github.com/atlastore/belt/key"
)

func main() {
	// 1. Create a KeyFactory with unique IDs for your node and disk
	keyFactory := key.NewKeyFactory(key.KeyFactoryParams{
		NodeId: 1234567890,
		DiskID: 9876543210,
	})

	// 2. Define the data for a new key
	myKey := &key.KeyV1{
		IndexFileHash: hashx.FNV64.HashString("path/to/my/file.txt"),
		Identifier:    key.GenerateIdentifier(16), // A 16-byte random identifier
	}

	// 3. Encode the key into a hex string
	// This automatically injects the NodeID and DiskID from the factory.
	encodedKey, err := keyFactory.EncodeKey(myKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("Encoded Key:", encodedKey)

	// 4. Decode the key string back into a struct
	decodedKey, err := key.Decode[*key.KeyV1](encodedKey)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Decoded Key:\n")
	fmt.Printf("  - Version: %d\n", decodedKey.Version())
	fmt.Printf("  - Node ID: %d\n", decodedKey.NodeID)
	fmt.Printf("  - Disk ID: %d\n", decodedKey.DiskID)
	fmt.Printf("  - File Hash: %d\n", decodedKey.IndexFileHash)
}
```

## License

This repository is licensed under the [MIT License](LICENSE.txt).