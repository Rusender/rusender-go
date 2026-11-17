package rusender

import (
	"context"
	"time"
)

// Address is an email address with an optional display name.
type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Addr returns an Address without a name.
func Addr(email string) Address { return Address{Email: email} }

// EmailAttachment is a file attached to a transactional email: either Content with FileName, or FileID of a
// file already uploaded to RuSender (attached under its own name; FileName must be empty).
type EmailAttachment struct {
	FileName string `json:"fileName,omitempty"`
	// Content is the file contents; it is encoded to base64 automatically.
	Content []byte `json:"content,omitempty"`
	FileID  int    `json:"fileId,omitempty"`
}

// SendEmailParams are the parameters of [EmailsService.Send].
type SendEmailParams struct {
	From Address `json:"from"`
	// To: up to 50 recipients together with Cc and Bcc.
	To  []Address `json:"to"`
	Cc  []Address `json:"cc,omitempty"`
	Bcc []Address `json:"bcc,omitempty"`
	// ReplyTo: up to 5 addresses.
	ReplyTo     []Address `json:"replyTo,omitempty"`
	Subject     string    `json:"subject"`
	PreviewText string    `json:"previewText,omitempty"`
	HTML        string    `json:"html,omitempty"`
	Text        string    `json:"text,omitempty"`
	// TemplateID: template by id instead of HTML/Text.
	TemplateID int `json:"templateId,omitempty"`
	// TemplateAlias: template by alias instead of HTML/Text.
	TemplateAlias string `json:"templateAlias,omitempty"`
	// Variables: template variable values — string, number, bool or nil.
	Variables map[string]any    `json:"variables,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	// Attachments: up to 20.
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	// KeyID is the sending key; the domain's default key is used when zero.
	KeyID int `json:"keyId,omitempty"`
}

// EmailStatus is the status of an email. The type is open: the API may add values.
type EmailStatus string

const (
	EmailQueued    EmailStatus = "queued"
	EmailSending   EmailStatus = "sending"
	EmailDelivered EmailStatus = "delivered"
	EmailBounced   EmailStatus = "bounced"
	EmailFailed    EmailStatus = "failed"
)

// RecipientStatus is the delivery status of one recipient.
type RecipientStatus string

const (
	RecipientQueued     RecipientStatus = "queued"
	RecipientSending    RecipientStatus = "sending"
	RecipientDelivered  RecipientStatus = "delivered"
	RecipientBounced    RecipientStatus = "bounced"
	RecipientFailed     RecipientStatus = "failed"
	RecipientSuppressed RecipientStatus = "suppressed"
)

// RecipientType is the address field of a recipient.
type RecipientType string

const (
	RecipientTo  RecipientType = "to"
	RecipientCc  RecipientType = "cc"
	RecipientBcc RecipientType = "bcc"
)

// EmailEventType is the type of a recipient event.
type EmailEventType string

const (
	EmailEventSuppressed   EmailEventType = "suppressed"
	EmailEventDelivered    EmailEventType = "delivered"
	EmailEventBounced      EmailEventType = "bounced"
	EmailEventFailed       EmailEventType = "failed"
	EmailEventOpened       EmailEventType = "opened"
	EmailEventClicked      EmailEventType = "clicked"
	EmailEventUnsubscribed EmailEventType = "unsubscribed"
	EmailEventComplained   EmailEventType = "complained"
)

// EmailReasonCode explains why a recipient was not delivered, e.g. RECIPIENT_UNSUBSCRIBED,
// DELIVERY_MAILBOX_NOT_FOUND, DELIVERY_MAILBOX_FULL, INTERNAL_ERROR.
type EmailReasonCode string

// EmailReason is the reason of a recipient status.
type EmailReason struct {
	Code EmailReasonCode `json:"code"`
	// BounceType: permanent, transient or undetermined.
	BounceType   string `json:"bounceType,omitempty"`
	Message      string `json:"message"`
	SMTPResponse string `json:"smtpResponse,omitempty"`
}

// AcceptedEmailRecipient is a recipient of a just accepted email.
type AcceptedEmailRecipient struct {
	ID     string          `json:"id"`
	Email  string          `json:"email"`
	Name   *string         `json:"name"`
	Type   RecipientType   `json:"type"`
	Status RecipientStatus `json:"status"`
	Reason *EmailReason    `json:"reason"`
}

// AcceptedEmail is the result of [EmailsService.Send].
type AcceptedEmail struct {
	ID         string                   `json:"id"`
	Status     EmailStatus              `json:"status"`
	CreatedAt  time.Time                `json:"createdAt"`
	Recipients []AcceptedEmailRecipient `json:"recipients"`
}

// EmailRecipient is a recipient of an email with send and completion times.
type EmailRecipient struct {
	AcceptedEmailRecipient
	SentAt      *time.Time `json:"sentAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

// EmailEvent is an event of a recipient.
type EmailEvent struct {
	Type       EmailEventType `json:"type"`
	OccurredAt time.Time      `json:"occurredAt"`
	URL        string         `json:"url,omitempty"`
}

// EmailRecipientWithEvents is a recipient with its events.
type EmailRecipientWithEvents struct {
	EmailRecipient
	Events []EmailEvent `json:"events"`
}

// EmailSender is the sender address of an email.
type EmailSender struct {
	Email string  `json:"email"`
	Name  *string `json:"name"`
}

// EmailSummary is an email in a list.
type EmailSummary struct {
	ID          string           `json:"id"`
	Status      EmailStatus      `json:"status"`
	KeyID       int              `json:"keyId"`
	From        EmailSender      `json:"from"`
	Subject     string           `json:"subject"`
	TemplateID  *int             `json:"templateId"`
	CreatedAt   time.Time        `json:"createdAt"`
	SentAt      *time.Time       `json:"sentAt"`
	CompletedAt *time.Time       `json:"completedAt"`
	Recipients  []EmailRecipient `json:"recipients"`
}

// Email is an email with the status, reason and events of every recipient.
type Email struct {
	ID          string                     `json:"id"`
	Status      EmailStatus                `json:"status"`
	KeyID       int                        `json:"keyId"`
	From        EmailSender                `json:"from"`
	Subject     string                     `json:"subject"`
	TemplateID  *int                       `json:"templateId"`
	CreatedAt   time.Time                  `json:"createdAt"`
	SentAt      *time.Time                 `json:"sentAt"`
	CompletedAt *time.Time                 `json:"completedAt"`
	Recipients  []EmailRecipientWithEvents `json:"recipients"`
}

// EmailRecipientDetails is one recipient of an email with its events.
type EmailRecipientDetails struct {
	EmailRecipientWithEvents
	EmailID      string `json:"emailId"`
	SMTPResponse string `json:"smtpResponse,omitempty"`
	ClickCount   int    `json:"clickCount"`
}

// Delivery is one recipient of any email, see [EmailsService.ListRecipients].
type Delivery struct {
	ID          *string         `json:"id"`
	EmailID     *string         `json:"emailId"`
	Email       string          `json:"email"`
	Name        *string         `json:"name"`
	Type        *RecipientType  `json:"type"`
	Status      RecipientStatus `json:"status"`
	Reason      *EmailReason    `json:"reason"`
	KeyID       int             `json:"keyId"`
	From        EmailSender     `json:"from"`
	Subject     *string         `json:"subject"`
	CreatedAt   time.Time       `json:"createdAt"`
	SentAt      *time.Time      `json:"sentAt"`
	CompletedAt *time.Time      `json:"completedAt"`
}

// EmailRecipientEvent is an event of a recipient with the device and client of opens and clicks.
type EmailRecipientEvent struct {
	EmailEvent
	// Device: desktop, phone or tablet.
	Device *string `json:"device,omitempty"`
	// Client: outlook, chrome, firefox, safari, apple_mail or other.
	Client *string `json:"client,omitempty"`
}

// ListEmailsParams are the parameters of [EmailsService.List].
type ListEmailsParams struct {
	// Limit: 1–100, 20 by default.
	Limit    int        `json:"limit,omitempty"`
	Cursor   string     `json:"cursor,omitempty"`
	KeyID    int        `json:"keyId,omitempty"`
	DateFrom *time.Time `json:"dateFrom,omitempty"`
	DateTo   *time.Time `json:"dateTo,omitempty"`
	From     string     `json:"from,omitempty"`
}

// ListEmailRecipientsParams are the parameters of [EmailsService.ListRecipients].
type ListEmailRecipientsParams struct {
	ListEmailsParams
	Status []RecipientStatus `json:"status,omitempty"`
	To     string            `json:"to,omitempty"`
}

// ListRecipientEventsParams are the parameters of [EmailsService.ListRecipientEvents].
type ListRecipientEventsParams struct {
	PageParams
	Type []EmailEventType `json:"type,omitempty"`
}

// EmailsService: transactional emails.
type EmailsService struct{ t *transport }

// Send sends a transactional email. The email is accepted for delivery (status queued); track delivery
// with Get or webhooks. Attachment contents are encoded to base64 automatically.
// Scope: email.send.
func (s *EmailsService) Send(ctx context.Context, params *SendEmailParams, opts ...RequestOption) (*AcceptedEmail, error) {
	return call[AcceptedEmail](ctx, s.t, request{method: "POST", path: "/v1/public/emails", body: params}, opts)
}

// List lists sent emails, newest first, with cursor pagination: request the next page with Cursor set to
// NextCursor.
// Scope: email.read.
func (s *EmailsService) List(ctx context.Context, params *ListEmailsParams, opts ...RequestOption) (*CursorPage[EmailSummary], error) {
	return listCursorPage[EmailSummary](ctx, s.t, request{method: "GET", path: "/v1/public/emails", query: params}, opts)
}

// Get gets an email with the status and reason of every recipient.
// Scope: email.read.
func (s *EmailsService) Get(ctx context.Context, emailID string, opts ...RequestOption) (*Email, error) {
	return call[Email](ctx, s.t, request{method: "GET", path: pathf("/v1/public/emails/%s", emailID)}, opts)
}

// ListRecipients lists deliveries across all emails — one item per recipient, filterable by status and
// address. Cursor pagination: request the next page with Cursor set to NextCursor.
// Scope: email.read.
func (s *EmailsService) ListRecipients(ctx context.Context, params *ListEmailRecipientsParams, opts ...RequestOption) (*CursorPage[Delivery], error) {
	return listCursorPage[Delivery](ctx, s.t, request{method: "GET", path: "/v1/public/emails/recipients", query: params}, opts)
}

// GetRecipient gets one recipient of an email with its status, reason and events.
// Scope: email.read.
func (s *EmailsService) GetRecipient(ctx context.Context, emailID, recipientID string, opts ...RequestOption) (*EmailRecipientDetails, error) {
	return call[EmailRecipientDetails](ctx, s.t, request{
		method: "GET",
		path:   pathf("/v1/public/emails/%s/recipients/%s", emailID, recipientID),
	}, opts)
}

// ListRecipientEvents lists events of an email recipient: delivery, bounces, opens, clicks, unsubscribes.
// Scope: email.read.
func (s *EmailsService) ListRecipientEvents(ctx context.Context, emailID, recipientID string, params *ListRecipientEventsParams, opts ...RequestOption) (*Page[EmailRecipientEvent], error) {
	return listPage[EmailRecipientEvent](ctx, s.t, request{
		method: "GET",
		path:   pathf("/v1/public/emails/%s/recipients/%s/events", emailID, recipientID),
		query:  params,
	}, opts)
}
