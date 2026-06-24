package sqs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nerdmenot/doze/internal/awslocal"
	"github.com/nerdmenot/doze/internal/engine"
)

// Admin: expose each declared queue's depth and the data actions the dash/CLI run
// against the running backend, reusing the JSON-1.0 wire path the Converger uses.

// Actions reports the data operations doze offers for SQS queues.
func (Driver) Actions() []engine.Action {
	return []engine.Action{
		{ID: "peek", Label: "Peek", Kind: "queue"},
		{ID: "send", Label: "Send", Kind: "queue", InputHint: "message body"},
		{ID: "purge", Label: "Purge", Kind: "queue", Destructive: true},
	}
}

// Resources lists declared queues with a live depth/in-flight status line.
func (Driver) Resources(ctx context.Context, inst engine.Instance, ep engine.Endpoint) ([]engine.Resource, error) {
	cfg, ok := inst.Spec.(*Config)
	if !ok || cfg == nil {
		return nil, nil
	}
	client := awslocal.UnixHTTPClient(ep.Backend)
	out := make([]engine.Resource, 0, len(cfg.Queues))
	for _, q := range cfg.Queues {
		var r struct {
			Attributes map[string]string `json:"Attributes"`
		}
		// A queue that hasn't converged yet just shows an empty status.
		_ = sqsCall(ctx, client, "GetQueueAttributes",
			map[string]any{"QueueName": q.Name, "AttributeNames": []string{"All"}}, &r)
		out = append(out, engine.Resource{
			Kind: "queue", Name: q.Name,
			Status: queueStatus(r.Attributes), Info: queueInfo(r.Attributes),
		})
	}
	return out, nil
}

// Run performs an SQS data action and returns a human result line.
func (Driver) Run(ctx context.Context, _ engine.Instance, ep engine.Endpoint, action, resource, input string) (string, error) {
	client := awslocal.UnixHTTPClient(ep.Backend)
	switch action {
	case "purge":
		if err := sqsCall(ctx, client, "PurgeQueue", map[string]any{"QueueName": resource}, nil); err != nil {
			return "", err
		}
		return "purged " + resource, nil
	case "send":
		if strings.TrimSpace(input) == "" {
			return "", fmt.Errorf("a message body is required")
		}
		if err := sqsCall(ctx, client, "SendMessage",
			map[string]any{"QueueName": resource, "MessageBody": input}, nil); err != nil {
			return "", err
		}
		return "sent 1 message to " + resource, nil
	case "peek":
		// VisibilityTimeout 0 keeps the messages immediately visible — a true
		// non-destructive peek rather than a receive that hides them.
		var r struct {
			Messages []struct {
				Body string `json:"Body"`
			} `json:"Messages"`
		}
		if err := sqsCall(ctx, client, "ReceiveMessage", map[string]any{
			"QueueName": resource, "MaxNumberOfMessages": 10, "VisibilityTimeout": 0,
		}, &r); err != nil {
			return "", err
		}
		if len(r.Messages) == 0 {
			return "(no visible messages)", nil
		}
		var b strings.Builder
		for i, m := range r.Messages {
			fmt.Fprintf(&b, "%d. %s\n", i+1, m.Body)
		}
		return strings.TrimRight(b.String(), "\n"), nil
	}
	return "", fmt.Errorf("unknown sqs action %q", action)
}

func queueStatus(a map[string]string) string {
	if a == nil {
		return ""
	}
	depth := a["ApproximateNumberOfMessages"]
	if depth == "" {
		depth = "0"
	}
	s := depth + " msgs"
	if f := a["ApproximateNumberOfMessagesNotVisible"]; f != "" && f != "0" {
		s += " · " + f + " in-flight"
	}
	return s
}

func queueInfo(a map[string]string) map[string]string {
	if a == nil {
		return nil
	}
	info := map[string]string{}
	if a["FifoQueue"] == "true" {
		info["fifo"] = "true"
	}
	if rp := a["RedrivePolicy"]; rp != "" {
		info["redrive"] = rp
	}
	if len(info) == 0 {
		return nil
	}
	return info
}

// sqsCall posts a JSON-1.0 SQS request (X-Amz-Target) and, when out is non-nil,
// decodes the JSON response into it. Mirrors the Converger's wire format.
func sqsCall(ctx context.Context, c *http.Client, target string, payload map[string]any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix/", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	req.Header.Set("X-Amz-Target", "AmazonSQS."+target)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s: %s", target, resp.Status, strings.TrimSpace(string(msg)))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
