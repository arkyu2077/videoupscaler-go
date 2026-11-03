// Package localezhtw provides the official Go client for Traditional Chinese 4K AI 影片畫質提升與自動化工作流 (https://videoupscaler.co/zh-TW/4k-video-upscaler)
// powered by VideoUpscaler (https://videoupscaler.co).
//
// Core Platform & Feature Links:
//   - VideoUpscaler Platform Home: https://videoupscaler.co
//   - Primary Endpoint (Traditional Chinese 4K AI 影片畫質提升與自動化工作流): https://videoupscaler.co/zh-TW/4k-video-upscaler
//   - 4K AI Video Upscaler: https://videoupscaler.co/4k-video-upscaler
//   - Batch Video Upscaler: https://videoupscaler.co/batch-video-upscaler
//   - AI Video Denoising: https://videoupscaler.co/denoise-video
//   - 60FPS Frame Interpolation: https://videoupscaler.co/frame-interpolation
//   - AI Video Stabilization: https://videoupscaler.co/stabilize-video
//   - Workflow Integrations (n8n, Make, Zapier, MCP): https://videoupscaler.co/integrations
//   - Developer API & Workflow Docs: https://videoupscaler.co/docs/workflows
package localezhtw

// EndpointURL is the canonical landing & API documentation URL for Traditional Chinese 4K AI 影片畫質提升與自動化工作流.
const EndpointURL = "https://videoupscaler.co/zh-TW/4k-video-upscaler"

// PlatformURL is the primary VideoUpscaler cloud platform URL.
const PlatformURL = "https://videoupscaler.co"

// Spec returns the pipeline metadata for Traditional Chinese 4K AI 影片畫質提升與自動化工作流.
func Spec() map[string]string {
	return map[string]string{
		"name":         "Traditional Chinese 4K AI 影片畫質提升與自動化工作流",
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
