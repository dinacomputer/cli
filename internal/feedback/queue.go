// Package feedback persists feedback submissions that couldn't be sent to the
// Signals API and exposes a retry path. The queue lives under
// ~/.config/dina/feedback-queue/ as one JSON file per item, named with a
// time-sortable ID.
package feedback

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/dinacomputer/cli/internal/api"
)

const queueDirName = "feedback-queue"

// Kind discriminates queued items.
type Kind string

const (
	KindBug      Kind = "bug"
	KindFeature  Kind = "feature"
	KindFeedback Kind = "feedback"
)

// Item is one queued submission.
type Item struct {
	ID       string          `json:"-"`
	Kind     Kind            `json:"kind"`
	QueuedAt time.Time       `json:"queued_at"`
	Body     json.RawMessage `json:"body"`
}

// queueDir returns the directory where items are stored.
func queueDir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "dina", queueDirName), nil
}

// Enqueue persists a feedback submission for later retry. body must be one of
// api.BugBody, api.FeatureRequestBody, or api.FeedbackBody.
func Enqueue(kind Kind, body any) error {
	dir, err := queueDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return err
	}

	id, err := newID()
	if err != nil {
		return err
	}

	item := Item{
		Kind:     kind,
		QueuedAt: time.Now().UTC(),
		Body:     bodyJSON,
	}
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, id+".json"), data, 0o600)
}

// newID returns a sortable, unique ID like "20260427T100000Z-abc123".
func newID() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return time.Now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(b[:]), nil
}

// List returns every queued item, oldest first.
func List() ([]Item, error) {
	dir, err := queueDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var items []Item
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var item Item
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}
		item.ID = e.Name()[:len(e.Name())-len(".json")]
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].QueuedAt.Before(items[j].QueuedAt)
	})
	return items, nil
}

// Remove deletes the queued item with the given ID. Missing files are not an
// error.
func Remove(id string) error {
	dir, err := queueDir()
	if err != nil {
		return err
	}
	err = os.Remove(filepath.Join(dir, id+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Clear deletes every queued item. Returns the number of items removed and
// any error encountered while removing. Stops at the first error.
func Clear() (int, error) {
	items, err := List()
	if err != nil {
		return 0, err
	}
	for i, item := range items {
		if err := Remove(item.ID); err != nil {
			return i, err
		}
	}
	return len(items), nil
}

// Submit retries one queued item via the given API client. On success the
// item is removed from the queue.
func Submit(client *api.Client, item Item) (*api.SubmitResponseBody, error) {
	switch item.Kind {
	case KindBug:
		var body api.BugBody
		if err := json.Unmarshal(item.Body, &body); err != nil {
			return nil, fmt.Errorf("decoding queued bug: %w", err)
		}
		resp, err := client.SubmitBug(body)
		if err != nil {
			return nil, err
		}
		return resp, Remove(item.ID)
	case KindFeature:
		var body api.FeatureRequestBody
		if err := json.Unmarshal(item.Body, &body); err != nil {
			return nil, fmt.Errorf("decoding queued feature request: %w", err)
		}
		resp, err := client.SubmitFeatureRequest(body)
		if err != nil {
			return nil, err
		}
		return resp, Remove(item.ID)
	case KindFeedback:
		var body api.FeedbackBody
		if err := json.Unmarshal(item.Body, &body); err != nil {
			return nil, fmt.Errorf("decoding queued feedback: %w", err)
		}
		resp, err := client.SubmitFeedback(body)
		if err != nil {
			return nil, err
		}
		return resp, Remove(item.ID)
	default:
		return nil, fmt.Errorf("unknown queued item kind %q", item.Kind)
	}
}
