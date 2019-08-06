package kubernetes

import (
	"net/http"
)

type WebhookHandler struct {
	Assets http.FileSystem
	Config WebhookHandlerConfig
}

type WebhookHandlerConfig struct {
	AnnotationNamespace string
}

func NewWebhookHandler(fs http.FileSystem, c *WebhookHandlerConfig) *WebhookHandler {
	if c == nil {
		c = &WebhookHandlerConfig{}
	}
	return &WebhookHandler{
		Assets: fs,
		Config: *c,
	}
}
