// Package videoupscaler provides the official Go client SDK for the VideoUpscaler AI API (https://videoupscaler.co).
//
// VideoUpscaler enables developers to upscale and enhance videos up to 1080p, 2K, and 4K resolution
// with AI-powered super-resolution, denoising, frame interpolation, and video stabilization.
//
// # Key Capabilities & Documentation
//
//   - AI Video Upscaler Platform: https://videoupscaler.co
//   - 4K Video Upscaler (1080p, 2K & 4K AI Super-Resolution): https://videoupscaler.co/4k-video-upscaler
//   - Batch Video Upscaler (Concurrent Multi-Clip Pipelines): https://videoupscaler.co/batch-video-upscaler
//   - Workflow Integrations (n8n, Make, Zapier & MCP): https://videoupscaler.co/integrations
//   - API & Workflow Documentation: https://videoupscaler.co/docs/workflows
//
// # Quickstart
//
//	client := videoupscaler.NewClient(os.Getenv("VIDEOUPSCALER_API_KEY"))
//	task, err := client.Upscale(ctx, videoupscaler.UpscaleRequest{
//		VideoURL:   "https://example.com/input.mp4",
//		Resolution: videoupscaler.Resolution4K,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	completed, err := client.WaitForCompletion(ctx, task.ID)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("4K Video URL:", completed.OutputURL)
package videoupscaler

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default endpoint for the VideoUpscaler v1 REST API.
	DefaultBaseURL = "https://videoupscaler.co/api/v1"

	// DefaultPollInterval is the default interval between polling requests in WaitForCompletion
	// when long-polling returns immediately.
	DefaultPollInterval = 2 * time.Second

	// DefaultWaitChunkSeconds is the default server-side long-poll duration per GetTask call during WaitForCompletion.
	DefaultWaitChunkSeconds = 45

	// Resolution1080p upscales video to Full HD (1920x1080).
	Resolution1080p = "1080p"
	// Resolution2K upscales video to 2K QHD (2560x1440).
	Resolution2K = "2k"
	// Resolution4K upscales video to 4K Ultra HD (3840x2160).
	Resolution4K = "4k"

	// StatusQueued indicates the task is waiting for a GPU worker.
	StatusQueued = "queued"
	// StatusProcessing indicates the task is actively being processed.
	StatusProcessing = "processing"
	// StatusCompleted indicates the task finished and OutputURL is ready.
	StatusCompleted = "completed"
	// StatusFailed indicates the task failed; check Task.Error for details.
	StatusFailed = "failed"
)

// Client is an idiomatic, zero-dependency HTTP client for the VideoUpscaler AI API (https://videoupscaler.co).
type Client struct {
	apiKey       string
	baseURL      string
	httpClient   *http.Client
	userAgent    string
	pollInterval time.Duration
	waitChunkSec int
}

// Option configures a Client instance.
type Option func(*Client)

// WithBaseURL overrides the default API base URL (https://videoupscaler.co/api/v1).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient provides a custom *http.Client for API requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithUserAgent sets a custom User-Agent header on outgoing requests.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// WithPollInterval configures the delay between poll attempts in WaitForCompletion.
func WithPollInterval(d time.Duration) Option {
	return func(c *Client) {
		if d >= 0 {
			c.pollInterval = d
		}
	}
}

// WithWaitChunkSeconds configures the long-poll wait parameter (0-45 seconds) used by WaitForCompletion.
func WithWaitChunkSeconds(seconds int) Option {
	return func(c *Client) {
		if seconds >= 0 && seconds <= 45 {
			c.waitChunkSec = seconds
		}
	}
}

// NewClient creates a new VideoUpscaler API client.
// Obtain an API key from https://videoupscaler.co/settings/apikeys.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:       strings.TrimSpace(apiKey),
		baseURL:      DefaultBaseURL,
		httpClient:   &http.Client{Timeout: 5 * time.Minute},
		userAgent:    "videoupscaler-go/0.1.0 (+https://github.com/arkyu2077/videoupscaler-go)",
		pollInterval: DefaultPollInterval,
		waitChunkSec: DefaultWaitChunkSeconds,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// UpscaleOps specifies optional AI video enhancement filters to run alongside upscaling.
// Learn more at https://videoupscaler.co/docs/workflows.
type UpscaleOps struct {
	// Engine selects the AI model ("flashvsr", "real-esrgan", " anime"). Defaults to "flashvsr".
	Engine string `json:"engine,omitempty"`
	// Denoise strength from 0.0 to 1.0 (https://videoupscaler.co/denoise-video).
	Denoise float64 `json:"denoise,omitempty"`
	// Deinterlace removes interlacing artifacts from legacy footage.
	Deinterlace bool `json:"deinterlace,omitempty"`
	// Deflicker smooths temporal luminance flicker.
	Deflicker bool `json:"deflicker,omitempty"`
	// Stabilize applies AI motion stabilization (https://videoupscaler.co/stabilize-video).
	Stabilize bool `json:"stabilize,omitempty"`
	// FaceRestore enhances facial details during upscaling.
	FaceRestore bool `json:"faceRestore,omitempty"`
	// FPSMultiplier increases frame rate (e.g. 2 for 60fps interpolation: https://videoupscaler.co/frame-interpolation).
	FPSMultiplier int `json:"fpsMultiplier,omitempty"`
	// SlowMotion outputs interpolated frames as smooth slow-motion video.
	SlowMotion bool `json:"slowMotion,omitempty"`
	// HDR tone-mapping or PQ/HLG output mode ("none", "hdr10", "hlg").
	HDR string `json:"hdr,omitempty"`
}

// UpscaleRequest defines parameters for submitting a video upscaling job via POST /upscale.
type UpscaleRequest struct {
	// VideoURL is the publicly accessible HTTP(S) URL of the source video (.mp4, .mov, .webm).
	VideoURL string `json:"video_url"`
	// Resolution is the target output resolution: "1080p", "2k", or "4k" (default "4k").
	Resolution string `json:"resolution"`
	// Ops configures optional AI enhancement stages (denoise, frame interpolation, stabilization, etc.).
	Ops *UpscaleOps `json:"ops,omitempty"`
	// WebhookURL is an optional HTTPS endpoint that receives a POST callback when the task finishes.
	WebhookURL string `json:"webhook_url,omitempty"`
	// IdempotencyKey prevents duplicate billing if a request is retried. Auto-generated if empty.
	IdempotencyKey string `json:"-"`
}

// Task represents an asynchronous video upscaling job.
type Task struct {
	ID             string      `json:"id"`
	Status         string      `json:"status"`
	Resolution     string      `json:"resolution,omitempty"`
	VideoURL       string      `json:"video_url,omitempty"`
	OutputURL      string      `json:"output_url,omitempty"`
	CreditsUsed    float64     `json:"credits_used,omitempty"`
	DurationSec    float64     `json:"duration_seconds,omitempty"`
	Error          string      `json:"error,omitempty"`
	CreatedAt      string      `json:"created_at,omitempty"`
	CompletedAt    string      `json:"completed_at,omitempty"`
	Ops            *UpscaleOps `json:"ops,omitempty"`
	IdempotencyKey string      `json:"idempotency_key,omitempty"`
}

// IsTerminal reports whether the task has reached a final state ("completed" or "failed").
func (t *Task) IsTerminal() bool {
	return t != nil && (t.Status == StatusCompleted || t.Status == StatusFailed)
}

// TaskList contains a paginated list of tasks returned by ListTasks.
type TaskList struct {
	Tasks []Task `json:"tasks"`
	Page  int    `json:"page,omitempty"`
	Limit int    `json:"limit,omitempty"`
	Total int    `json:"total,omitempty"`
}

// Account represents the current user's credit balance and pricing tier.
type Account struct {
	Credits          float64 `json:"credits"`
	CreditsPerSecond float64 `json:"credits_per_second"`
	Plan             string  `json:"plan,omitempty"`
}

// APIError is returned when the VideoUpscaler API responds with a non-2xx HTTP status.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("videoupscaler: HTTP %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("videoupscaler: HTTP %d: %s", e.StatusCode, e.Message)
}

// Upscale submits a new video upscaling job (POST /upscale).
// If req.Resolution is empty, it defaults to "4k" (https://videoupscaler.co/4k-video-upscaler).
// If req.IdempotencyKey is empty, a random UUIDv4 key is generated automatically.
func (c *Client) Upscale(ctx context.Context, req UpscaleRequest) (*Task, error) {
	if strings.TrimSpace(req.VideoURL) == "" {
		return nil, errors.New("videoupscaler: VideoURL is required")
	}
	if req.Resolution == "" {
		req.Resolution = Resolution4K
	} else {
		req.Resolution = strings.ToLower(strings.TrimSpace(req.Resolution))
	}
	switch req.Resolution {
	case Resolution1080p, Resolution2K, Resolution4K:
	default:
		return nil, fmt.Errorf("videoupscaler: invalid resolution %q (must be 1080p, 2k, or 4k)", req.Resolution)
	}

	idemKey := strings.TrimSpace(req.IdempotencyKey)
	if idemKey == "" {
		idemKey = newUUID()
	}

	headers := map[string]string{
		"Idempotency-Key": idemKey,
	}

	var task Task
	if err := c.doJSON(ctx, http.MethodPost, "/upscale", req, headers, &task); err != nil {
		return nil, err
	}
	if task.IdempotencyKey == "" {
		task.IdempotencyKey = idemKey
	}
	return &task, nil
}

// GetTask retrieves the current status of a task by ID (GET /tasks/{id}).
// When waitSeconds > 0 (up to 45), the server long-polls until the task completes or the timeout expires.
func (c *Client) GetTask(ctx context.Context, taskID string, waitSeconds int) (*Task, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errors.New("videoupscaler: taskID is required")
	}
	path := "/tasks/" + url.PathEscape(taskID)
	if waitSeconds > 0 {
		if waitSeconds > 45 {
			waitSeconds = 45
		}
		path += "?wait=" + strconv.Itoa(waitSeconds)
	}
	var task Task
	if err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// WaitForCompletion polls GetTask until the task reaches StatusCompleted or StatusFailed,
// or until ctx is canceled. If the task fails, an error is returned alongside the failed *Task.
func (c *Client) WaitForCompletion(ctx context.Context, taskID string) (*Task, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		task, err := c.GetTask(ctx, taskID, c.waitChunkSec)
		if err != nil {
			return nil, err
		}
		switch task.Status {
		case StatusCompleted:
			return task, nil
		case StatusFailed:
			msg := task.Error
			if msg == "" {
				msg = "video upscaling task failed"
			}
			return task, fmt.Errorf("videoupscaler: task %s failed: %s", task.ID, msg)
		}

		if c.pollInterval > 0 {
			timer := time.NewTimer(c.pollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}
	}
}

// ListTasks returns a paginated list of recent upscaling tasks (GET /tasks?page=&limit=).
func (c *Client) ListTasks(ctx context.Context, page, limit int) (*TaskList, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	q := url.Values{}
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))

	var list TaskList
	if err := c.doJSON(ctx, http.MethodGet, "/tasks?"+q.Encode(), nil, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetAccount retrieves the credit balance and per-second rate for the authenticated account (GET /account).
func (c *Client) GetAccount(ctx context.Context) (*Account, error) {
	var acct Account
	if err := c.doJSON(ctx, http.MethodGet, "/account", nil, nil, &acct); err != nil {
		return nil, err
	}
	return &acct, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, headers map[string]string, out any) error {
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("videoupscaler: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("videoupscaler: create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("videoupscaler: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("videoupscaler: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, respBytes)
	}

	if out != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("videoupscaler: decode response: %w", err)
		}
	}
	return nil
}

func parseAPIError(statusCode int, data []byte) error {
	apiErr := &APIError{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
	}
	if len(data) == 0 {
		return apiErr
	}
	var envelope struct {
		Error any    `json:"error"`
		Code  string `json:"code"`
		Msg   string `json:"message"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil {
		switch v := envelope.Error.(type) {
		case string:
			if v != "" {
				apiErr.Message = v
			}
			if envelope.Code != "" {
				apiErr.Code = envelope.Code
			}
		case map[string]any:
			if c, ok := v["code"].(string); ok {
				apiErr.Code = c
			}
			if m, ok := v["message"].(string); ok && m != "" {
				apiErr.Message = m
			}
		}
		if apiErr.Message == http.StatusText(statusCode) && envelope.Msg != "" {
			apiErr.Message = envelope.Msg
		}
	} else {
		trimmed := strings.TrimSpace(string(data))
		if trimmed != "" {
			apiErr.Message = trimmed
		}
	}
	return apiErr
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("vu-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
