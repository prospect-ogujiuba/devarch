package podman

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"
)

type EventFilter struct {
	Container []string
	Event     []string
	Image     []string
	Type      []string
	Volume    []string
}

type EventCallback func(event *Event, err error) bool

func (c *Client) StreamEvents(ctx context.Context, since time.Time, filters *EventFilter, callback EventCallback) error {
	params := url.Values{}
	params.Set("stream", "true")

	if !since.IsZero() {
		params.Set("since", since.Format(time.RFC3339))
	}

	if filters != nil {
		filterJSON := buildFilterJSON(filters)
		if filterJSON != "" {
			params.Set("filters", filterJSON)
		}
	}

	path := "/libpod/events?" + params.Encode()

	resp, err := c.get(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("events stream failed %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if !callback(nil, err) {
				return nil
			}
			continue
		}

		if len(line) == 0 {
			continue
		}

		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			if !callback(nil, err) {
				return nil
			}
			continue
		}

		if !callback(&event, nil) {
			return nil
		}
	}
}

func buildFilterJSON(f *EventFilter) string {
	filters := make(map[string][]string)

	if len(f.Container) > 0 {
		filters["container"] = f.Container
	}
	if len(f.Event) > 0 {
		filters["event"] = f.Event
	}
	if len(f.Image) > 0 {
		filters["image"] = f.Image
	}
	if len(f.Type) > 0 {
		filters["type"] = f.Type
	}
	if len(f.Volume) > 0 {
		filters["volume"] = f.Volume
	}

	if len(filters) == 0 {
		return ""
	}

	data, _ := json.Marshal(filters)
	return string(data)
}

type EventHandler struct {
	client    *Client
	callbacks []EventCallback
	filters   *EventFilter
}

func NewEventHandler(client *Client) *EventHandler {
	return &EventHandler{
		client:    client,
		callbacks: make([]EventCallback, 0),
	}
}

func (h *EventHandler) OnContainerEvent(callback EventCallback) *EventHandler {
	h.callbacks = append(h.callbacks, callback)
	if h.filters == nil {
		h.filters = &EventFilter{}
	}
	h.filters.Type = append(h.filters.Type, "container")
	return h
}

func (h *EventHandler) OnImageEvent(callback EventCallback) *EventHandler {
	h.callbacks = append(h.callbacks, callback)
	if h.filters == nil {
		h.filters = &EventFilter{}
	}
	h.filters.Type = append(h.filters.Type, "image")
	return h
}

func (h *EventHandler) OnVolumeEvent(callback EventCallback) *EventHandler {
	h.callbacks = append(h.callbacks, callback)
	if h.filters == nil {
		h.filters = &EventFilter{}
	}
	h.filters.Type = append(h.filters.Type, "volume")
	return h
}

func (h *EventHandler) WithContainerFilter(containers ...string) *EventHandler {
	if h.filters == nil {
		h.filters = &EventFilter{}
	}
	h.filters.Container = append(h.filters.Container, containers...)
	return h
}

func (h *EventHandler) Start(ctx context.Context) error {
	return h.client.StreamEvents(ctx, time.Time{}, h.filters, func(event *Event, err error) bool {
		if err != nil {
			return true
		}
		for _, cb := range h.callbacks {
			if !cb(event, nil) {
				return false
			}
		}
		return true
	})
}

type ContainerEventType string

const (
	EventStart   ContainerEventType = "start"
	EventStop    ContainerEventType = "stop"
	EventDie     ContainerEventType = "die"
	EventKill    ContainerEventType = "kill"
	EventPause   ContainerEventType = "pause"
	EventUnpause ContainerEventType = "unpause"
	EventCreate  ContainerEventType = "create"
	EventRemove  ContainerEventType = "remove"
	EventRename  ContainerEventType = "rename"
	EventRestart ContainerEventType = "restart"
	EventHealth  ContainerEventType = "health_status"
)

func IsContainerEvent(event *Event, eventType ContainerEventType) bool {
	return event.Type == "container" && event.Action == string(eventType)
}

func GetContainerName(event *Event) string {
	if event.Actor.Attributes != nil {
		if name, ok := event.Actor.Attributes["name"]; ok {
			return name
		}
	}
	return event.Actor.ID
}
