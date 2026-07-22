package models

type ConversationMessage struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type RequestContext struct {
	Title        string                `json:"title,omitempty"`
	Conversation []ConversationMessage `json:"conversation,omitempty"`
	Custom       map[string]any        `json:"custom,omitempty"`
}

type RequestOptions struct {
	Signature string `json:"signature,omitempty"`
	Length    string `json:"length,omitempty"`
	Model     string `json:"model,omitempty"`
}

type RequestMetadata struct {
	Application string `json:"application,omitempty"`
	Module      string `json:"module,omitempty"`
	TenantID    string `json:"tenant_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Department  string `json:"department,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
}

type WriteRequest struct {
	Profile  string          `json:"profile" binding:"required"`
	Action   string          `json:"action" binding:"required"`
	Tone     string          `json:"tone,omitempty"`
	Language string          `json:"language,omitempty"`
	Text     string          `json:"text" binding:"required"`
	Context  RequestContext  `json:"context,omitempty"`
	Options  RequestOptions  `json:"options,omitempty"`
	Metadata RequestMetadata `json:"metadata,omitempty"`
}

type WriteResponseData struct {
	Result       string `json:"result"`
	Profile      string `json:"profile"`
	Action       string `json:"action"`
	Model        string `json:"model"`
	ProcessingMS int64  `json:"processing_ms"`
}

type ApiResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
	Error   any  `json:"error,omitempty"`
}

type ApiErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}
