package rusender

import (
	"context"
	"time"
)

// SendingKeyStatus is the status of a sending key: enabled, disabled, banned or error.
type SendingKeyStatus string

// SendingKeyDomain is the domain of a sending key.
type SendingKeyDomain struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	IsVerified bool   `json:"isVerified"`
}

// SendingKey is a key for sending transactional emails.
type SendingKey struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	Status       SendingKeyStatus `json:"status"`
	IsThrottled  bool             `json:"isThrottled"`
	IsDefault    bool             `json:"isDefault"`
	Domain       SendingKeyDomain `json:"domain"`
	DashboardURL string           `json:"dashboardUrl"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

// SendingKeyTotals are the counts of a sending key statistics.
type SendingKeyTotals struct {
	All          int `json:"all"`
	Delivered    int `json:"delivered"`
	Opened       int `json:"opened"`
	Clicked      int `json:"clicked"`
	Error        int `json:"error"`
	SoftBounced  int `json:"softBounced"`
	HardBounced  int `json:"hardBounced"`
	Unsubscribed int `json:"unsubscribed"`
	Complaint    int `json:"complaint"`
}

// SendingKeyRates are the rates of a sending key statistics.
type SendingKeyRates struct {
	OpenRate                    float64 `json:"openRate"`
	ClickRate                   float64 `json:"clickRate"`
	ErrorRate                   float64 `json:"errorRate"`
	UnsubscribeAndComplaintRate float64 `json:"unsubscribeAndComplaintRate"`
}

// SendingKeyStats is the sending statistics of a key for a period.
type SendingKeyStats struct {
	// DateFrom and DateTo are calendar days, YYYY-MM-DD.
	DateFrom string           `json:"dateFrom"`
	DateTo   string           `json:"dateTo"`
	Totals   SendingKeyTotals `json:"totals"`
	Rates    SendingKeyRates  `json:"rates"`
}

// CreateSendingKeyParams are the parameters of [SendingKeysService.Create].
type CreateSendingKeyParams struct {
	Name     string `json:"name"`
	DomainID int    `json:"domainId"`
}

// UpdateSendingKeyParams are the parameters of [SendingKeysService.Update].
type UpdateSendingKeyParams struct {
	Name string `json:"name,omitempty"`
	// IsDefault: true makes the key the default key of its domain; false is not accepted.
	IsDefault bool `json:"isDefault,omitempty"`
}

// SendingKeyStatsParams are the parameters of [SendingKeysService.GetStatistics].
type SendingKeyStatsParams struct {
	// DateFrom and DateTo are calendar days, YYYY-MM-DD.
	DateFrom string `json:"dateFrom,omitempty"`
	DateTo   string `json:"dateTo,omitempty"`
}

// SendingKeysService: sending keys for transactional emails.
type SendingKeysService struct{ t *transport }

// List lists sending keys for transactional emails.
// Scope: sending_keys.read.
func (s *SendingKeysService) List(ctx context.Context, params *PageParams, opts ...RequestOption) (*Page[SendingKey], error) {
	return listPage[SendingKey](ctx, s.t, request{method: "GET", path: "/v1/public/sending-keys", query: params}, opts)
}

// Create creates a sending key for a verified domain. The key starts disabled and becomes enabled after
// moderation.
// Scope: sending_keys.write.
func (s *SendingKeysService) Create(ctx context.Context, params *CreateSendingKeyParams, opts ...RequestOption) (*SendingKey, error) {
	return call[SendingKey](ctx, s.t, request{method: "POST", path: "/v1/public/sending-keys", body: params}, opts)
}

// Get gets a sending key.
// Scope: sending_keys.read.
func (s *SendingKeysService) Get(ctx context.Context, keyID int, opts ...RequestOption) (*SendingKey, error) {
	return call[SendingKey](ctx, s.t, request{method: "GET", path: pathf("/v1/public/sending-keys/%s", keyID)}, opts)
}

// Update renames a sending key or, with IsDefault, makes it the default key of its domain.
// Scope: sending_keys.write.
func (s *SendingKeysService) Update(ctx context.Context, keyID int, params *UpdateSendingKeyParams, opts ...RequestOption) (*SendingKey, error) {
	return call[SendingKey](ctx, s.t, request{method: "PATCH", path: pathf("/v1/public/sending-keys/%s", keyID), body: params}, opts)
}

// Delete deletes a sending key. Sending with it stops immediately.
// Scope: sending_keys.write.
func (s *SendingKeysService) Delete(ctx context.Context, keyID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/sending-keys/%s", keyID)}, opts)
}

// GetStatistics gets sending statistics of a key for a period; the last 30 days when DateFrom and DateTo are
// empty.
// Scope: sending_keys.read.
func (s *SendingKeysService) GetStatistics(ctx context.Context, keyID int, params *SendingKeyStatsParams, opts ...RequestOption) (*SendingKeyStats, error) {
	return call[SendingKeyStats](ctx, s.t, request{method: "GET", path: pathf("/v1/public/sending-keys/%s/stats", keyID), query: params}, opts)
}
