# VideoUpscaler Go SDK (`videoupscaler-go`)

[![Go Reference](https://pkg.go.dev/badge/github.com/arkyu2077/videoupscaler-go.svg)](https://pkg.go.dev/github.com/arkyu2077/videoupscaler-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/arkyu2077/videoupscaler-go)](https://goreportcard.com/report/github.com/arkyu2077/videoupscaler-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Official, zero-dependency Go client SDK for the [AI Video Upscaler](https://videoupscaler.co) REST API. Upscale and enhance videos to **1080p, 2K, and 4K Ultra HD** directly from Go services, media pipelines, worker queues, and CLI tools.

Whether you are building an automated [4K Video Upscaler](https://videoupscaler.co/4k-video-upscaler) microservice, running a high-throughput [Batch Video Upscaler](https://videoupscaler.co/batch-video-upscaler) worker pool, or chaining enhancement stages with the [AI Video Denoiser](https://videoupscaler.co/denoise-video), [60fps AI Frame Interpolation](https://videoupscaler.co/frame-interpolation), and [AI Video Stabilizer](https://videoupscaler.co/stabilize-video), `videoupscaler-go` provides a clean, context-aware, idiomatic Go interface.

---

## Features

- **Zero External Dependencies**: Built strictly on the Go 1.21+ standard library (`net/http`, `context`, `encoding/json`).
- **4K AI Super-Resolution**: Upscale `.mp4`, `.mov`, and `.webm` footage to `1080p`, `2k`, or `4k` via the [4K Video Upscaler](https://videoupscaler.co/4k-video-upscaler) engine.
- **Multi-Stage AI Enhancement**: Combine super-resolution with [AI Video Denoiser](https://videoupscaler.co/denoise-video), [60fps AI Frame Interpolation](https://videoupscaler.co/frame-interpolation), [AI Video Stabilizer](https://videoupscaler.co/stabilize-video), deinterlacing, deflickering, face restoration, and HDR output (`hdr10`, `hlg`).
- **Built-in Long Polling & Idempotency**: Automatic UUIDv4 `Idempotency-Key` generation on `POST /upscale` and server-side long-polling (`?wait=45`) in `WaitForCompletion` to minimize network overhead.
- **Concurrency-Ready**: Designed for goroutine worker pools and [Batch Video Upscaler](https://videoupscaler.co/batch-video-upscaler) workflows.

---

## Installation

Requires Go 1.21 or newer:

```bash
go get github.com/arkyu2077/videoupscaler-go@v0.1.0
```

Import the package in your Go code:

```go
import "github.com/arkyu2077/videoupscaler-go"
```

---

## Authentication

1. Sign in to [VideoUpscaler](https://videoupscaler.co) and open **Settings → API Keys**.
2. Create an API key (`vu_live_...`) and export it in your environment:

```bash
export VIDEOUPSCALER_API_KEY="vu_live_..."
```

For full endpoint specifications, webhook payload schemas, and rate limits, see the [API & Workflow Documentation](https://videoupscaler.co/docs/workflows).

---

## Quickstart: Upscale a Video to 4K

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arkyu2077/videoupscaler-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	client := videoupscaler.NewClient(os.Getenv("VIDEOUPSCALER_API_KEY"))

	// 1. Check account credit balance
	account, err := client.GetAccount(ctx)
	if err != nil {
		log.Fatalf("get account failed: %v", err)
	}
	fmt.Printf("Available credits: %.1f (%.1f credits/sec)\n", account.Credits, account.CreditsPerSecond)

	// 2. Submit a 4K upscaling + enhancement task
	task, err := client.Upscale(ctx, videoupscaler.UpscaleRequest{
		VideoURL:   "https://cdn.example.com/archive/clip-720p.mp4",
		Resolution: videoupscaler.Resolution4K,
		Ops: &videoupscaler.UpscaleOps{
			Engine:        "flashvsr",
			Denoise:       0.4,
			Stabilize:     true,
			FPSMultiplier: 2, // 60fps AI frame interpolation
		},
	})
	if err != nil {
		log.Fatalf("submit upscale failed: %v", err)
	}
	fmt.Printf("Submitted task %s (status: %s)\n", task.ID, task.Status)

	// 3. Wait for completion using built-in long-polling
	completed, err := client.WaitForCompletion(ctx, task.ID)
	if err != nil {
		log.Fatalf("upscale failed: %v", err)
	}
	fmt.Printf("4K Output Ready: %s\n", completed.OutputURL)
}
```

---

## Concurrent Batch 4K Upscaling Example

Go's concurrency primitives pair naturally with the [Batch Video Upscaler](https://videoupscaler.co/batch-video-upscaler) API. Below is a worker-bounded batch pipeline that upscales multiple videos to 4K concurrently:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/arkyu2077/videoupscaler-go"
)

func upscaleBatch(ctx context.Context, urls []string, concurrency int) {
	client := videoupscaler.NewClient(os.Getenv("VIDEOUPSCALER_API_KEY"))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, videoURL := range urls {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			task, err := client.Upscale(ctx, videoupscaler.UpscaleRequest{
				VideoURL:   src,
				Resolution: videoupscaler.Resolution4K,
			})
			if err != nil {
				fmt.Printf("[FAIL] %s: %v\n", src, err)
				return
			}

			done, err := client.WaitForCompletion(ctx, task.ID)
			if err != nil {
				fmt.Printf("[FAIL] %s (task %s): %v\n", src, task.ID, err)
				return
			}
			fmt.Printf("[DONE] %s -> %s\n", src, done.OutputURL)
		}(videoURL)
	}

	wg.Wait()
}
```

---

## API Overview

| Method | Description | REST Endpoint |
| --- | --- | --- |
| `NewClient(apiKey, opts...)` | Creates a configured `*Client` instance | — |
| `client.Upscale(ctx, req)` | Submits a video for 1080p, 2K, or 4K AI upscaling | `POST /api/v1/upscale` |
| `client.GetTask(ctx, taskID, waitSec)` | Gets task status with optional server long-polling (`0–45s`) | `GET /api/v1/tasks/{id}` |
| `client.WaitForCompletion(ctx, taskID)` | Polls until `completed` or `failed` | `GET /api/v1/tasks/{id}?wait=45` |
| `client.ListTasks(ctx, page, limit)` | Lists recent upscaling tasks (paginated) | `GET /api/v1/tasks` |
| `client.GetAccount(ctx)` | Returns remaining credit balance and pricing rate | `GET /api/v1/account` |

---

## Official Ecosystem & Integrations

Alongside the Go SDK, VideoUpscaler offers official packages and no-code connectors across the ecosystem:

- **Web Platform**: [AI Video Upscaler](https://videoupscaler.co)
  - [4K Video Upscaler](https://videoupscaler.co/4k-video-upscaler) — Upscale low-res footage to crisp 1080p, 2K, and 4K.
  - [Batch Video Upscaler](https://videoupscaler.co/batch-video-upscaler) — Process entire folders and multi-clip queues in parallel.
  - [AI Video Denoiser](https://videoupscaler.co/denoise-video) — Remove low-light grain, sensor noise, and compression artifacts.
  - [60fps AI Frame Interpolation](https://videoupscaler.co/frame-interpolation) — Convert 24/30fps footage into silky-smooth 60fps or slow-motion.
  - [AI Video Stabilizer](https://videoupscaler.co/stabilize-video) — Eliminate handheld camera shake without heavy cropping.
- **Workflow & Automation Hub**: [n8n, Make, Zapier & MCP Integrations](https://videoupscaler.co/integrations)
- **Developer Docs**: [API & Workflow Documentation](https://videoupscaler.co/docs/workflows)
- **npm Packages**:
  - [`videoupscaler` on npm](https://www.npmjs.com/package/videoupscaler) — Node.js SDK & CLI
  - [`videoupscaler-mcp` on npm](https://www.npmjs.com/package/videoupscaler-mcp) — Model Context Protocol (MCP) server for Claude Desktop, Cursor, and AI agents
  - [`n8n-nodes-videoupscaler` on npm](https://www.npmjs.com/package/n8n-nodes-videoupscaler) — Official community node for n8n workflow automation

---

## License

Released under the [MIT License](./LICENSE).
