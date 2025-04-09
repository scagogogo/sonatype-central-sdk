# Sonatype Central SDK for Go

[![Build Status](https://github.com/scagogogo/sonatype-central-sdk/actions/workflows/go-test.yml/badge.svg)](https://github.com/scagogogo/sonatype-central-sdk/actions/workflows/go-test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/scagogogo/sonatype-central-sdk)](https://goreportcard.com/report/github.com/scagogogo/sonatype-central-sdk)
[![GoDoc](https://godoc.org/github.com/scagogogo/sonatype-central-sdk?status.svg)](https://godoc.org/github.com/scagogogo/sonatype-central-sdk)
[![License](https://img.shields.io/github/license/scagogogo/sonatype-central-sdk)](https://github.com/scagogogo/sonatype-central-sdk/blob/main/LICENSE)

A comprehensive Go SDK for interacting with the Sonatype Central Repository API. This SDK provides a clean, type-safe interface to search and retrieve artifacts from Maven Central.

## Features

- Complete API coverage for all Sonatype Central search endpoints
- Rich searching capabilities (by GroupId, ArtifactId, SHA1, Class, etc.)
- Batch operations and concurrent processing
- Flexible client configuration
- Caching and retry mechanisms
- Async API support
- Security rating and metadata retrieval

## Installation

```bash
go get github.com/scagogogo/sonatype-central-sdk
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

func main() {
	// Create a new client
	client := api.NewClient()
	
	// Search for artifacts with GroupId "org.apache.commons"
	artifacts, err := client.SearchByGroupId(context.Background(), "org.apache.commons", 10)
	if err != nil {
		panic(err)
	}
	
	// Print results
	for _, artifact := range artifacts {
		fmt.Printf("GroupId: %s, ArtifactId: %s, Latest Version: %s\n",
			artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
	}
}
```

## Advanced Usage

See the [documentation](https://godoc.org/github.com/scagogogo/sonatype-central-sdk) for detailed API usage examples.

## License

MIT License - see [LICENSE](LICENSE) for details. 