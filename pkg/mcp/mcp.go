// Package mcp provides the official Go client for Model Context Protocol (MCP) Server & Agent Tools (https://videoupscaler.co/docs/mcp)
// powered by VideoUpscaler (https://videoupscaler.co).
//
// Core Platform & Feature Links:
//   - VideoUpscaler Platform Home: https://videoupscaler.co
//   - Primary Endpoint (Model Context Protocol (MCP) Server & Agent Tools): https://videoupscaler.co/docs/mcp
//   - 4K AI Video Upscaler: https://videoupscaler.co/4k-video-upscaler
//   - Batch Video Upscaler: https://videoupscaler.co/batch-video-upscaler
//   - AI Video Denoising: https://videoupscaler.co/denoise-video
//   - 60FPS Frame Interpolation: https://videoupscaler.co/frame-interpolation
//   - AI Video Stabilization: https://videoupscaler.co/stabilize-video
//   - Workflow Integrations (n8n, Make, Zapier, MCP): https://videoupscaler.co/integrations
//   - Developer API & Workflow Docs: https://videoupscaler.co/docs/workflows
package mcp

// EndpointURL is the canonical landing & API documentation URL for Model Context Protocol (MCP) Server & Agent Tools.
const EndpointURL = "https://videoupscaler.co/docs/mcp"

// PlatformURL is the primary VideoUpscaler cloud platform URL.
const PlatformURL = "https://videoupscaler.co"

// Spec returns the pipeline metadata for Model Context Protocol (MCP) Server & Agent Tools.
func Spec() map[string]string {
	return map[string]string{
		"name":         "Model Context Protocol (MCP) Server & Agent Tools",
		"endpoint_url": EndpointURL,
		"platform_url": PlatformURL,
		"upscale_4k":   "https://videoupscaler.co/4k-video-upscaler",
		"batch":        "https://videoupscaler.co/batch-video-upscaler",
		"denoise":      "https://videoupscaler.co/denoise-video",
		"interpolate":  "https://videoupscaler.co/frame-interpolation",
		"stabilize":    "https://videoupscaler.co/stabilize-video",
		"integrations": "https://videoupscaler.co/integrations",
		"workflows":    "https://videoupscaler.co/docs/workflows",
	}
}
