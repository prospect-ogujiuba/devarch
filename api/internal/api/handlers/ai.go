package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/priz/devarch-api/internal/api/respond"
	"github.com/priz/devarch-api/internal/llm"
	"github.com/priz/devarch-api/internal/ramalama"
)

type Conversation struct {
	Messages []llm.ChatMessage
	LastUsed time.Time
}

type AIHandler struct {
	manager        *ramalama.Manager
	client         *llm.Client
	contextBuilder *llm.ContextBuilder
	conversations  sync.Map
	mu             sync.Mutex
}

func NewAIHandler(ctx context.Context, manager *ramalama.Manager, client *llm.Client, contextBuilder *llm.ContextBuilder) *AIHandler {
	h := &AIHandler{
		manager:        manager,
		client:         client,
		contextBuilder: contextBuilder,
	}
	go h.pruneConversations(ctx)
	return h
}

func (h *AIHandler) pruneConversations(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.conversations.Range(func(key, value any) bool {
				conv, ok := value.(*Conversation)
				if !ok {
					h.conversations.Delete(key)
					return true
				}
				if !conv.LastUsed.IsZero() && time.Since(conv.LastUsed) > 30*time.Minute {
					h.conversations.Delete(key)
				}
				return true
			})
		}
	}
}

func (h *AIHandler) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.manager.GetStatus()
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.JSON(w, r, http.StatusOK, status)
}

func (h *AIHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if err := h.manager.Stop(); err != nil {
		respond.InternalError(w, r, err)
		return
	}
	respond.Action(w, r, http.StatusOK, "stopped", respond.WithMessage("LLM container stopped"))
}

type generateRequest struct {
	Description string `json:"description"`
}

func (h *AIHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request body")
		return
	}
	if req.Description == "" {
		respond.BadRequest(w, r, "description is required")
		return
	}

	ctx := h.contextBuilder.ServiceAuthorContext()
	systemPrompt := llm.ServiceAuthorPrompt(ctx)

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Description},
	}

	resp, err := h.client.Complete(messages)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if len(resp.Choices) == 0 {
		respond.InternalError(w, r, fmt.Errorf("no response from LLM"))
		return
	}

	content := resp.Choices[0].Message.Content

	// Try to parse as JSON to validate
	var service map[string]interface{}
	if err := json.Unmarshal([]byte(content), &service); err != nil {
		// Return raw if not valid JSON
		respond.JSON(w, r, http.StatusOK, map[string]string{"raw": content})
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]interface{}{"service": service})
}

type chatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
	Stream         bool   `json:"stream,omitempty"`
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request body")
		return
	}
	if req.Message == "" {
		respond.BadRequest(w, r, "message is required")
		return
	}

	// Get or create conversation
	convID := req.ConversationID
	if convID == "" {
		convID = fmt.Sprintf("conv-%d", time.Now().UnixNano())
	}

	conv := h.getOrCreateConversation(convID)

	h.mu.Lock()
	conv.Messages = append(conv.Messages, llm.ChatMessage{Role: "user", Content: req.Message})
	conv.LastUsed = time.Now()
	messages := make([]llm.ChatMessage, len(conv.Messages))
	copy(messages, conv.Messages)
	h.mu.Unlock()

	if req.Stream {
		h.streamChat(w, r, conv, convID, messages)
		return
	}

	resp, err := h.client.Complete(messages)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if len(resp.Choices) == 0 {
		respond.InternalError(w, r, fmt.Errorf("no response from LLM"))
		return
	}

	assistantMsg := resp.Choices[0].Message
	h.mu.Lock()
	conv.Messages = append(conv.Messages, assistantMsg)
	h.mu.Unlock()
	h.conversations.Store(convID, conv)

	respond.JSON(w, r, http.StatusOK, map[string]interface{}{
		"message":         assistantMsg.Content,
		"conversation_id": convID,
	})
}

func (h *AIHandler) getOrCreateConversation(convID string) *Conversation {
	if existing, ok := h.conversations.Load(convID); ok {
		if conv, ok := existing.(*Conversation); ok {
			return conv
		}
	}
	ctx := h.contextBuilder.CLIAssistantContext()
	conv := &Conversation{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: llm.CLIAssistantPrompt(ctx)},
		},
		LastUsed: time.Now(),
	}
	h.conversations.Store(convID, conv)
	return conv
}

func (h *AIHandler) streamChat(w http.ResponseWriter, r *http.Request, conv *Conversation, convID string, messages []llm.ChatMessage) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		respond.InternalError(w, r, fmt.Errorf("streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Send conversation ID first
	fmt.Fprintf(w, "data: {\"conversation_id\":%q}\n\n", convID)
	flusher.Flush()

	var fullContent string
	err := h.client.StreamComplete(messages, func(chunk llm.StreamChunk) {
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			fullContent += delta
			data, _ := json.Marshal(map[string]string{"content": delta})
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	})

	if err != nil {
		fmt.Fprintf(w, "data: {\"error\":%q}\n\n", err.Error())
		flusher.Flush()
		return
	}

	h.mu.Lock()
	conv.Messages = append(conv.Messages, llm.ChatMessage{Role: "assistant", Content: fullContent})
	h.mu.Unlock()
	h.conversations.Store(convID, conv)

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

type diagnoseRequest struct {
	Target string `json:"target"`
}

func (h *AIHandler) Diagnose(w http.ResponseWriter, r *http.Request) {
	var req diagnoseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request body")
		return
	}
	if req.Target == "" {
		respond.BadRequest(w, r, "target is required")
		return
	}

	ctx := h.contextBuilder.DebugContext(req.Target)
	systemPrompt := llm.DebugPrompt(ctx)

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Diagnose issues with: %s", req.Target)},
	}

	resp, err := h.client.Complete(messages)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	if len(resp.Choices) == 0 {
		respond.InternalError(w, r, fmt.Errorf("no response from LLM"))
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]interface{}{
		"target":    req.Target,
		"diagnosis": resp.Choices[0].Message.Content,
	})
}

func (h *AIHandler) PullModel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, "invalid request body")
		return
	}
	if req.Model == "" {
		respond.BadRequest(w, r, "model is required")
		return
	}

	output, err := h.manager.PullModel(req.Model)
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.Action(w, r, http.StatusOK, "pulled", respond.WithMessage("Model pulled successfully"), respond.WithOutput(output))
}

func (h *AIHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.manager.ListModels()
	if err != nil {
		respond.InternalError(w, r, err)
		return
	}

	respond.JSON(w, r, http.StatusOK, map[string]interface{}{"models": models})
}
