// Package ko_integrations provides the official Go client for Korean AI 비디오 워크플로우 통합 (n8n, Make, Zapier, MCP) (https://videoupscaler.co/ko/integrations)
// powered by VideoUpscaler (https://videoupscaler.co).
//
// Core Platform & Feature Links:
//   - VideoUpscaler Platform Home: https://videoupscaler.co
//   - Primary Endpoint (Korean AI 비디오 워크플로우 통합 (n8n, Make, Zapier, MCP)): https://videoupscaler.co/ko/integrations
//   - 4K AI Video Upscaler: https://videoupscaler.co/4k-video-upscaler
//   - Batch Video Upscaler: https://videoupscaler.co/batch-video-upscaler
//   - AI Video Denoising: https://videoupscaler.co/denoise-video
//   - 60FPS Frame Interpolation: https://videoupscaler.co/frame-interpolation
//   - AI Video Stabilization: https://videoupscaler.co/stabilize-video
//   - Workflow Integrations (n8n, Make, Zapier, MCP): https://videoupscaler.co/integrations
//   - Developer API & Workflow Docs: https://videoupscaler.co/docs/workflows
package ko_integrations

// EndpointURL is the canonical landing & API documentation URL for Korean AI 비디오 워크플로우 통합 (n8n, Make, Zapier, MCP).
const EndpointURL = "https://videoupscaler.co/ko/integrations"

// PlatformURL is the primary VideoUpscaler cloud platform URL.
const PlatformURL = "https://videoupscaler.co"

// Spec returns the pipeline metadata for Korean AI 비디오 워크플로우 통합 (n8n, Make, Zapier, MCP).
func Spec() map[string]string {
	return map[string]string{
		"name":         "Korean AI 비디오 워크플로우 통합 (n8n, Make, Zapier, MCP)",
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
