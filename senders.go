package rusender

import (
	"context"
	"time"
)

// Sender is a sender address for campaigns.
type Sender struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	// Verified is enabled when the address is verified.
	Verified     Status     `json:"verified"`
	DomainID     *int       `json:"domainId,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	VerifiedAt   *time.Time `json:"verifiedAt,omitempty"`
	DashboardURL string     `json:"dashboardUrl"`
}

// ListSendersParams are the parameters of [SendersService.List].
type ListSendersParams struct {
	PageParams
	Verified Status `json:"verified,omitempty"`
}

// CreateSenderParams are the parameters of [SendersService.Create].
type CreateSenderParams struct {
	Email string `json:"email"`
}

// ResendConfirmationResult is the result of [SendersService.ResendConfirmation].
type ResendConfirmationResult struct {
	Sent bool `json:"sent"`
}

// SendersService: sender addresses for campaigns.
type SendersService struct{ t *transport }

// List lists sender addresses, filterable by verification.
// Scope: senders.read.
func (s *SendersService) List(ctx context.Context, params *ListSendersParams, opts ...RequestOption) (*Page[Sender], error) {
	return listPage[Sender](ctx, s.t, request{method: "GET", path: "/v1/public/senders", query: params}, opts)
}

// Create adds a sender address. A confirmation email is sent to it; the sender is verified once the link in
// it is clicked.
// Scope: senders.write.
func (s *SendersService) Create(ctx context.Context, params *CreateSenderParams, opts ...RequestOption) (*Sender, error) {
	return call[Sender](ctx, s.t, request{method: "POST", path: "/v1/public/senders", body: params}, opts)
}

// Get gets a sender address.
// Scope: senders.read.
func (s *SendersService) Get(ctx context.Context, senderID int, opts ...RequestOption) (*Sender, error) {
	return call[Sender](ctx, s.t, request{method: "GET", path: pathf("/v1/public/senders/%s", senderID)}, opts)
}

// ResendConfirmation sends the confirmation email to an unverified sender address again.
// Scope: senders.write.
func (s *SendersService) ResendConfirmation(ctx context.Context, senderID int, opts ...RequestOption) (*ResendConfirmationResult, error) {
	return call[ResendConfirmationResult](ctx, s.t, request{method: "POST", path: pathf("/v1/public/senders/%s/resend-confirmation", senderID)}, opts)
}
