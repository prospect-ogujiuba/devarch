package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/priz/devarch-ai/internal/embedding"
	"github.com/priz/devarch-ai/internal/llm"
	"github.com/priz/devarch-ai/internal/ramalama"
	"github.com/priz/devarch-ai/internal/respond"
)

type Conversation struct {
	Messages []llm.ChatMessage
	LastUsed time.Time
}

const (
	ragScoreThreshold = 0.3
	ragDefaultLimit   = 5
)

type Handler struct {
	manager        *ramalama.Manager
	client         *llm.Client
	contextBuilder *llm.ContextBuilder
	embedClient    *embedding.Client
	embedStore     *embedding.Store
	conversations  sync.Map
	mu             sync.Mutex
}

func NewHandler(ctx context.Context, manager *ramalama.Manager, client *llm.Client, contextBuilder *llm.ContextBuilder, embedClient *embedding.Client, embedStore *embedding.Store) *Handler {
	h := &Handler{
		manager:        manager,
		client:         client,
		contextBuilder: contextBuilder,
		embedClient:    embedClient,
		embedStore:     embedStore,
	}
	go h.pruneConversations(ctx)
	return h
}

func (h *Handler) pruneConversations(ctx context.Context) {
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

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.manager.GetStatus()
	if err != nil {
		respond.InternalError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, status)
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	if err := h.manager.Stop(); err != nil {
		respond.InternalError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "stopped", "message": "LLM container stopped"})
}

type generateRequest struct {
	Description string `json:"description"`
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, "invalid request body")
		return
	}
	if req.Description == "" {
		respond.BadRequest(w, "description is required")
		return
	}

	ctx := h.contextBuilder.ServiceAuthorContext()
	if rag := h.ragContext(req.Description); rag != "" {
		ctx += "\n\n" + rag
	}
	systemPrompt := llm.ServiceAuthorPrompt(ctx)

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Description},
	}

	resp, err := h.client.Complete(messages)
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	if len(resp.Choices) == 0 {
		respond.InternalError(w, fmt.Errorf("no response from LLM"))
		return
	}

	content := resp.Choices[0].Message.Content

	var service map[string]interface{}
	if err := json.Unmarshal([]byte(content), &service); err != nil {
		respond.JSON(w, http.StatusOK, map[string]string{"raw": content})
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{"service": service})
}

type chatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
	Stream         bool   `json:"stream,omitempty"`
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, "invalid request body")
		return
	}
	if req.Message == "" {
		respond.BadRequest(w, "message is required")
		return
	}

	convID := req.ConversationID
	if convID == "" {
		convID = fmt.Sprintf("conv-%d", time.Now().UnixNano())
	}

	conv := h.getOrCreateConversation(convID)

	userContent := req.Message
	if rag := h.ragContext(req.Message); rag != "" {
		userContent = req.Message + "\n\n---\n" + rag
	}

	h.mu.Lock()
	conv.Messages = append(conv.Messages, llm.ChatMessage{Role: "user", Content: userContent})
	conv.LastUsed = time.Now()
	messages := make([]llm.ChatMessage, len(conv.Messages))
	copy(messages, conv.Messages)
	h.mu.Unlock()

	if req.Stream {
		h.streamChat(w, conv, convID, messages)
		return
	}

	resp, err := h.client.Complete(messages)
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	if len(resp.Choices) == 0 {
		respond.InternalError(w, fmt.Errorf("no response from LLM"))
		return
	}

	assistantMsg := resp.Choices[0].Message
	h.mu.Lock()
	conv.Messages = append(conv.Messages, assistantMsg)
	h.mu.Unlock()
	h.conversations.Store(convID, conv)

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"message":         assistantMsg.Content,
		"conversation_id": convID,
	})
}

func (h *Handler) getOrCreateConversation(convID string) *Conversation {
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

func (h *Handler) streamChat(w http.ResponseWriter, conv *Conversation, convID string, messages []llm.ChatMessage) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		respond.InternalError(w, fmt.Errorf("streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

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

func (h *Handler) Diagnose(w http.ResponseWriter, r *http.Request) {
	var req diagnoseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, "invalid request body")
		return
	}
	if req.Target == "" {
		respond.BadRequest(w, "target is required")
		return
	}

	ctx := h.contextBuilder.DebugContext(req.Target)
	if rag := h.ragContext(req.Target); rag != "" {
		ctx += "\n\n" + rag
	}
	systemPrompt := llm.DebugPrompt(ctx)

	messages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Diagnose issues with: %s", req.Target)},
	}

	resp, err := h.client.Complete(messages)
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	if len(resp.Choices) == 0 {
		respond.InternalError(w, fmt.Errorf("no response from LLM"))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"target":    req.Target,
		"diagnosis": resp.Choices[0].Message.Content,
	})
}

func (h *Handler) PullModel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, "invalid request body")
		return
	}
	if req.Model == "" {
		respond.BadRequest(w, "model is required")
		return
	}

	output, err := h.manager.PullModel(req.Model)
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "pulled", "message": "Model pulled successfully", "output": output})
}

func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.manager.ListModels()
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{"models": models})
}

type embedRequest struct {
	Text string `json:"text"`
}

func (h *Handler) Embed(w http.ResponseWriter, r *http.Request) {
	var req embedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, "invalid request body")
		return
	}
	if req.Text == "" {
		respond.BadRequest(w, "text is required")
		return
	}

	vec, err := h.embedClient.EmbedSingle(req.Text)
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"embedding":  vec,
		"dimensions": len(vec),
	})
}

type indexRequest struct {
	SourceType string `json:"source_type"`
	SourceID   string `json:"source_id"`
	Content    string `json:"content"`
}

func (h *Handler) IndexDocument(w http.ResponseWriter, r *http.Request) {
	if h.embedStore == nil {
		respond.InternalError(w, fmt.Errorf("embedding store not configured"))
		return
	}

	var req indexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, "invalid request body")
		return
	}
	if req.SourceType == "" || req.SourceID == "" || req.Content == "" {
		respond.BadRequest(w, "source_type, source_id, and content are required")
		return
	}

	vec, err := h.embedClient.EmbedSingle(req.Content)
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	if err := h.embedStore.Upsert(req.SourceType, req.SourceID, req.Content, vec); err != nil {
		respond.InternalError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "indexed"})
}

type searchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

func (h *Handler) SearchDocuments(w http.ResponseWriter, r *http.Request) {
	if h.embedStore == nil {
		respond.InternalError(w, fmt.Errorf("embedding store not configured"))
		return
	}

	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, "invalid request body")
		return
	}
	if req.Query == "" {
		respond.BadRequest(w, "query is required")
		return
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}

	vec, err := h.embedClient.EmbedSingle(req.Query)
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	docs, err := h.embedStore.Search(vec, req.Limit)
	if err != nil {
		respond.InternalError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{"results": docs})
}

func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	if h.embedStore == nil {
		respond.InternalError(w, fmt.Errorf("embedding store not configured"))
		return
	}

	var req struct {
		SourceType string `json:"source_type"`
		SourceID   string `json:"source_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, "invalid request body")
		return
	}
	if req.SourceType == "" || req.SourceID == "" {
		respond.BadRequest(w, "source_type and source_id are required")
		return
	}

	if err := h.embedStore.DeleteBySource(req.SourceType, req.SourceID); err != nil {
		respond.InternalError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ragContext embeds the query, searches pgvector, and returns formatted context.
// Returns empty string if embedding store is unavailable or no relevant results found.
func (h *Handler) ragContext(query string) string {
	if h.embedStore == nil || h.embedClient == nil {
		return ""
	}

	vec, err := h.embedClient.EmbedSingle(query)
	if err != nil {
		return ""
	}

	docs, err := h.embedStore.Search(vec, ragDefaultLimit)
	if err != nil || len(docs) == 0 {
		return ""
	}

	var parts []string
	for _, doc := range docs {
		if doc.Score < ragScoreThreshold {
			continue
		}
		parts = append(parts, fmt.Sprintf("[%s/%s] (relevance: %.0f%%)\n%s", doc.SourceType, doc.SourceID, doc.Score*100, doc.Content))
	}
	if len(parts) == 0 {
		return ""
	}

	return "Relevant knowledge:\n" + strings.Join(parts, "\n\n")
}

func (h *Handler) IndexServices(w http.ResponseWriter, r *http.Request) {
	if h.embedStore == nil {
		respond.InternalError(w, fmt.Errorf("embedding store not configured"))
		return
	}

	services := h.contextBuilder.FetchServices()
	if len(services) == 0 {
		respond.JSON(w, http.StatusOK, map[string]interface{}{"indexed": 0, "message": "no services found"})
		return
	}

	indexed := 0
	var errors []string
	for _, svc := range services {
		content, _ := json.MarshalIndent(svc, "", "  ")
		name, _ := svc["name"].(string)
		if name == "" {
			continue
		}

		vec, err := h.embedClient.EmbedSingle(string(content))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: embed failed: %v", name, err))
			continue
		}

		if err := h.embedStore.Upsert("service", name, string(content), vec); err != nil {
			errors = append(errors, fmt.Sprintf("%s: store failed: %v", name, err))
			continue
		}
		indexed++
	}

	result := map[string]interface{}{"indexed": indexed, "total": len(services)}
	if len(errors) > 0 {
		result["errors"] = errors
	}
	respond.JSON(w, http.StatusOK, result)
}
