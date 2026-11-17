package rusender

import "context"

// APIKeyScope is a permission of an API key, such as contacts.read or email.send. The legacy scopes
// external_mail.send, external_mail.read and external_mail.write are superseded by email.send, email.read
// and sending_keys.*.
type APIKeyScope string

// Me is the account and the scopes of the current API key.
type Me struct {
	AccountID    int    `json:"accountId"`
	AccountEmail string `json:"accountEmail"`
	// Scopes of the current API key.
	Scopes []APIKeyScope `json:"scopes"`
}

// MeService: the account and scopes of the current API key.
type MeService struct{ t *transport }

// Get gets the account and the scopes of the current API key. Handy to check the key at startup.
// No specific scope is required.
func (s *MeService) Get(ctx context.Context, opts ...RequestOption) (*Me, error) {
	return call[Me](ctx, s.t, request{method: "GET", path: "/v1/public/me"}, opts)
}
