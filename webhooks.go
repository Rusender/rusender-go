package rusender

import (
	"context"
	"time"
)

// WebhookTrigger is an event a webhook subscribes to. The type is open: the API may add values.
type WebhookTrigger string

const (
	TriggerEmailSuppressed   WebhookTrigger = "email.suppressed"
	TriggerEmailDelivered    WebhookTrigger = "email.delivered"
	TriggerEmailBounced      WebhookTrigger = "email.bounced"
	TriggerEmailFailed       WebhookTrigger = "email.failed"
	TriggerEmailOpened       WebhookTrigger = "email.opened"
	TriggerEmailClicked      WebhookTrigger = "email.clicked"
	TriggerEmailUnsubscribed WebhookTrigger = "email.unsubscribed"
	TriggerEmailComplained   WebhookTrigger = "email.complained"

	TriggerCampaignStarted    WebhookTrigger = "mail_distribution.started"
	TriggerCampaignCompleted  WebhookTrigger = "mail_distribution.completed"
	TriggerCampaignBanned     WebhookTrigger = "mail_distribution.banned"
	TriggerCampaignModerating WebhookTrigger = "mail_distribution.moderating"
	TriggerCampaignApproved   WebhookTrigger = "mail_distribution.approved"
	TriggerCampaignRejected   WebhookTrigger = "mail_distribution.rejected"

	TriggerCampaignEmailDelivered   WebhookTrigger = "data_mail_distribution.delivered"
	TriggerCampaignEmailOpen        WebhookTrigger = "data_mail_distribution.open"
	TriggerCampaignEmailClick       WebhookTrigger = "data_mail_distribution.click"
	TriggerCampaignEmailHardBounced WebhookTrigger = "data_mail_distribution.hard_bounced"
	TriggerCampaignEmailSoftBounced WebhookTrigger = "data_mail_distribution.soft_bounced"
	TriggerCampaignEmailError       WebhookTrigger = "data_mail_distribution.error"
	TriggerCampaignEmailComplaint   WebhookTrigger = "data_mail_distribution.complaint"
	TriggerCampaignEmailUnsubscribe WebhookTrigger = "data_mail_distribution.unsubscribe"

	// Transactional emails of the legacy route (external-mails/send); emails sent with Emails.Send use TriggerEmail*.
	TriggerExternalMailDelivered   WebhookTrigger = "external_mail.delivered"
	TriggerExternalMailOpen        WebhookTrigger = "external_mail.open"
	TriggerExternalMailClick       WebhookTrigger = "external_mail.click"
	TriggerExternalMailHardBounced WebhookTrigger = "external_mail.hard_bounced"
	TriggerExternalMailSoftBounced WebhookTrigger = "external_mail.soft_bounced"
	TriggerExternalMailError       WebhookTrigger = "external_mail.error"
	TriggerExternalMailComplaint   WebhookTrigger = "external_mail.complaint"
	TriggerExternalMailUnsubscribe WebhookTrigger = "external_mail.unsubscribe"

	TriggerContactAdded           WebhookTrigger = "contact.added"
	TriggerContactRemoved         WebhookTrigger = "contact.removed"
	TriggerSendingKeyStateChanged WebhookTrigger = "sending_key.state_changed"
)

// Webhook is a subscription to events.
type Webhook struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	URL          string           `json:"url"`
	Triggers     []WebhookTrigger `json:"triggers"`
	Status       Status           `json:"status"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	DashboardURL string           `json:"dashboardUrl"`
}

// CreateWebhookParams are the parameters of [WebhooksService.Create].
type CreateWebhookParams struct {
	Name string `json:"name"`
	// URL of your handler.
	URL      string           `json:"url"`
	Triggers []WebhookTrigger `json:"triggers"`
	Status   Status           `json:"status,omitempty"`
}

// UpdateWebhookParams are the parameters of [WebhooksService.Update].
type UpdateWebhookParams struct {
	Name     string           `json:"name,omitempty"`
	URL      string           `json:"url,omitempty"`
	Triggers []WebhookTrigger `json:"triggers,omitempty"`
	Status   Status           `json:"status,omitempty"`
}

// WebhookTestResult is the result of [WebhooksService.Test].
type WebhookTestResult struct {
	Queued bool `json:"queued"`
}

// WebhooksService: webhook subscriptions to email and campaign events.
type WebhooksService struct{ t *transport }

// List lists webhooks.
// Scope: webhooks.read.
func (s *WebhooksService) List(ctx context.Context, params *PageParams, opts ...RequestOption) (*Page[Webhook], error) {
	return listPage[Webhook](ctx, s.t, request{method: "GET", path: "/v1/public/webhooks", query: params}, opts)
}

// Create creates a webhook: events from Triggers are sent to URL.
// Scope: webhooks.write.
func (s *WebhooksService) Create(ctx context.Context, params *CreateWebhookParams, opts ...RequestOption) (*Webhook, error) {
	return call[Webhook](ctx, s.t, request{method: "POST", path: "/v1/public/webhooks", body: params}, opts)
}

// Get gets a webhook.
// Scope: webhooks.read.
func (s *WebhooksService) Get(ctx context.Context, webhookID int, opts ...RequestOption) (*Webhook, error) {
	return call[Webhook](ctx, s.t, request{method: "GET", path: pathf("/v1/public/webhooks/%s", webhookID)}, opts)
}

// Update updates a webhook; only the fields you set change. Set Status to StatusDisabled to pause it.
// Scope: webhooks.write.
func (s *WebhooksService) Update(ctx context.Context, webhookID int, params *UpdateWebhookParams, opts ...RequestOption) (*Webhook, error) {
	return call[Webhook](ctx, s.t, request{method: "PATCH", path: pathf("/v1/public/webhooks/%s", webhookID), body: params}, opts)
}

// Delete deletes a webhook. Events stop being sent to its URL at once; to pause it, set Status to
// StatusDisabled instead.
// Scope: webhooks.write.
func (s *WebhooksService) Delete(ctx context.Context, webhookID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/webhooks/%s", webhookID)}, opts)
}

// Test sends a test event to the webhook URL to check your handler. Real data is not changed.
// Returns ErrValidation with code WEBHOOK_DISABLED if the webhook is disabled: enable it first.
// Not retried on failure: a retry would send a second event.
// Scope: webhooks.write.
func (s *WebhooksService) Test(ctx context.Context, webhookID int, opts ...RequestOption) (*WebhookTestResult, error) {
	return call[WebhookTestResult](ctx, s.t, request{
		method:        "POST",
		path:          pathf("/v1/public/webhooks/%s/test", webhookID),
		notIdempotent: true,
	}, opts)
}
