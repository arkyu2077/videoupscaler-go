package videoupscaler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestUpscale(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/upscale" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer vu_test_key" {
			t.Fatalf("unexpected Authorization header: %q", got)
		}
		if got := r.Header.Get("Idempotency-Key"); got == "" {
			t.Fatal("expected non-empty Idempotency-Key header")
		}

		var body UpscaleRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.VideoURL != "https://example.com/sample.mp4" {
			t.Fatalf("unexpected video_url: %q", body.VideoURL)
		}
		if body.Resolution != Resolution4K {
			t.Fatalf("unexpected resolution: %q", body.Resolution)
		}
		if body.Ops == nil || body.Ops.Denoise != 0.5 || !body.Ops.Stabilize {
			t.Fatalf("unexpected ops: %+v", body.Ops)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Task{
			ID:          "task_123",
			Status:      StatusQueued,
			Resolution:  body.Resolution,
			VideoURL:    body.VideoURL,
			CreditsUsed: 12,
		})
	}))
	defer srv.Close()

	client := NewClient("vu_test_key", WithBaseURL(srv.URL))
	task, err := client.Upscale(context.Background(), UpscaleRequest{
		VideoURL:   "https://example.com/sample.mp4",
		Resolution: Resolution4K,
		Ops: &UpscaleOps{
			Denoise:   0.5,
			Stabilize: true,
		},
	})
	if err != nil {
		t.Fatalf("Upscale error: %v", err)
	}
	if task.ID != "task_123" || task.Status != StatusQueued {
		t.Fatalf("unexpected task: %+v", task)
	}
	if task.IdempotencyKey == "" {
		t.Fatal("expected generated IdempotencyKey on returned task")
	}
}

func TestGetTaskAndWaitForCompletion(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/tasks/task_456" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			_ = json.NewEncoder(w).Encode(Task{
				ID:     "task_456",
				Status: StatusProcessing,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(Task{
			ID:        "task_456",
			Status:    StatusCompleted,
			OutputURL: "https://cdn.videoupscaler.co/outputs/task_456_4k.mp4",
		})
	}))
	defer srv.Close()

	client := NewClient("vu_test_key",
		WithBaseURL(srv.URL),
		WithPollInterval(10*time.Millisecond),
		WithWaitChunkSeconds(5),
	)

	t1, err := client.GetTask(context.Background(), "task_456", 10)
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	if t1.Status != StatusProcessing {
		t.Fatalf("expected processing, got %q", t1.Status)
	}

	completed, err := client.WaitForCompletion(context.Background(), "task_456")
	if err != nil {
		t.Fatalf("WaitForCompletion error: %v", err)
	}
	if completed.Status != StatusCompleted || completed.OutputURL == "" {
		t.Fatalf("unexpected completed task: %+v", completed)
	}
}

func TestGetAccountAndListTasks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/account":
			_ = json.NewEncoder(w).Encode(Account{
				Credits:          240.5,
				CreditsPerSecond: 1.5,
				Plan:             "pro",
			})
		case "/tasks":
			if r.URL.Query().Get("page") != "1" || r.URL.Query().Get("limit") != "10" {
				t.Fatalf("unexpected query params: %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(TaskList{
				Tasks: []Task{{ID: "task_1", Status: StatusCompleted}},
				Page:  1,
				Limit: 10,
				Total: 1,
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewClient("vu_test_key", WithBaseURL(srv.URL))

	acct, err := client.GetAccount(context.Background())
	if err != nil {
		t.Fatalf("GetAccount error: %v", err)
	}
	if acct.Credits != 240.5 || acct.CreditsPerSecond != 1.5 {
		t.Fatalf("unexpected account: %+v", acct)
	}

	list, err := client.ListTasks(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListTasks error: %v", err)
	}
	if len(list.Tasks) != 1 || list.Tasks[0].ID != "task_1" {
		t.Fatalf("unexpected task list: %+v", list)
	}
}
