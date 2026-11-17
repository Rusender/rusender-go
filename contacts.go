package rusender

import (
	"context"
	"net/http"
	"time"
)

// ContactStatus is the subscription status of a contact. The type is open: the API may add values.
type ContactStatus string

const (
	ContactNew              ContactStatus = "new"
	ContactActive           ContactStatus = "active"
	ContactUnsubscribed     ContactStatus = "unsubscribed"
	ContactUnsubscribedByUs ContactStatus = "unsubscribed_by_us"
	ContactSoftBounced      ContactStatus = "soft_bounced"
	ContactHardBounced      ContactStatus = "hard_bounced"
	ContactError            ContactStatus = "error"
	ContactComplaint        ContactStatus = "complaint"
)

// ContactVariables are contact variable values by variable key; nil means the value is not set.
type ContactVariables map[string]*string

// ListRef is a list a contact or a campaign belongs to.
type ListRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Contact is a contact.
type Contact struct {
	ID             int              `json:"id"`
	Email          string           `json:"email"`
	Name           *string          `json:"name,omitempty"`
	Phone          *string          `json:"phone,omitempty"`
	Status         ContactStatus    `json:"status"`
	ListIDs        []int            `json:"listIds"`
	Lists          []ListRef        `json:"lists,omitempty"`
	Variables      ContactVariables `json:"variables,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
	LastActivityAt *time.Time       `json:"lastActivityAt,omitempty"`
	DashboardURL   string           `json:"dashboardUrl"`
}

// ContactSortField is a sort field of contact lists.
type ContactSortField string

const (
	ContactSortID             ContactSortField = "id"
	ContactSortEmail          ContactSortField = "email"
	ContactSortCreatedAt      ContactSortField = "createdAt"
	ContactSortUpdatedAt      ContactSortField = "updatedAt"
	ContactSortLastActivityAt ContactSortField = "lastActivityAt"
)

// ListContactsParams are the parameters of [ContactsService.List].
type ListContactsParams struct {
	PageParams
	Email   string           `json:"email,omitempty"`
	Status  []ContactStatus  `json:"status,omitempty"`
	ListID  int              `json:"listId,omitempty"`
	SortBy  ContactSortField `json:"sortBy,omitempty"`
	SortDir SortDirection    `json:"sortDir,omitempty"`
}

// UpdateContactParams are the parameters of [ContactsService.Update]. The fields are pointers, so that
// String("") can clear them.
type UpdateContactParams struct {
	Name  *string `json:"name,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

// ContactStatisticsParams are the parameters of [ContactsService.GetStatistics].
type ContactStatisticsParams struct {
	// ListID limits the statistics to one list.
	ListID int `json:"listId,omitempty"`
}

// ContactStatistics counts contacts by status.
type ContactStatistics struct {
	Total        int            `json:"total"`
	ByStatus     map[string]int `json:"byStatus"`
	Active       int            `json:"active"`
	Unsubscribed int            `json:"unsubscribed"`
	Unavailable  int            `json:"unavailable"`
}

// ExportContactsParams are the parameters of [ContactsService.Export].
type ExportContactsParams struct {
	Format  ExportFormat     `json:"format"`
	Email   string           `json:"email,omitempty"`
	Status  []ContactStatus  `json:"status,omitempty"`
	SortBy  ContactSortField `json:"sortBy,omitempty"`
	SortDir SortDirection    `json:"sortDir,omitempty"`
	Limit   int              `json:"limit,omitempty"`
}

// CampaignSendingHistoryItem is a campaign the contact received.
type CampaignSendingHistoryItem struct {
	ID               int       `json:"id"`
	Status           string    `json:"status"`
	DistributionName string    `json:"distributionName"`
	CreatedAt        time.Time `json:"createdAt"`
}

// APISendingHistoryItem is a transactional email the contact received.
type APISendingHistoryItem struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	Recipient   string `json:"recipient,omitempty"`
	Subject     string `json:"subject,omitempty"`
	SenderEmail string `json:"senderEmail"`
	// Type: template or plain.
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
}

// BulkDeleteResult is the result of [ContactsService.BulkDelete].
type BulkDeleteResult struct {
	DeletedCount int `json:"deletedCount"`
}

// RemoveFromAllListsResult is the result of [ContactsService.RemoveFromAllLists].
type RemoveFromAllListsResult struct {
	AffectedContacts int `json:"affectedContacts"`
	RemovedLinks     int `json:"removedLinks"`
}

type contactIDsBody struct {
	ContactIDs []int `json:"contactIds"`
}

// ContactsService: contacts across all lists.
type ContactsService struct{ t *transport }

// List lists contacts across all lists, filterable by email, status and list.
// Scope: contacts.read.
func (s *ContactsService) List(ctx context.Context, params *ListContactsParams, opts ...RequestOption) (*Page[Contact], error) {
	return listPage[Contact](ctx, s.t, request{method: "GET", path: "/v1/public/contacts", query: params}, opts)
}

// Get gets a contact with its lists and variables.
// Scope: contacts.read.
func (s *ContactsService) Get(ctx context.Context, contactID int, opts ...RequestOption) (*Contact, error) {
	return call[Contact](ctx, s.t, request{method: "GET", path: pathf("/v1/public/contacts/%s", contactID)}, opts)
}

// Update updates a contact's name and phone. To change variables, use Lists.Contacts.Update.
// Scope: contacts.write.
func (s *ContactsService) Update(ctx context.Context, contactID int, params *UpdateContactParams, opts ...RequestOption) (*Contact, error) {
	return call[Contact](ctx, s.t, request{method: "PATCH", path: pathf("/v1/public/contacts/%s", contactID), body: params}, opts)
}

// Delete permanently deletes a contact from all lists.
// Scope: contacts.write.
func (s *ContactsService) Delete(ctx context.Context, contactID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/contacts/%s", contactID)}, opts)
}

// Unsubscribe unsubscribes a contact from all campaigns.
// Scope: contacts.write.
func (s *ContactsService) Unsubscribe(ctx context.Context, contactID int, opts ...RequestOption) (*Contact, error) {
	return call[Contact](ctx, s.t, request{method: "POST", path: pathf("/v1/public/contacts/%s/unsubscribe", contactID)}, opts)
}

// SubscribeBack restores the subscription of a contact you unsubscribed (status unsubscribed_by_us).
// A contact who unsubscribed via the link cannot be resubscribed: ErrConflict.
// Scope: contacts.write.
func (s *ContactsService) SubscribeBack(ctx context.Context, contactID int, opts ...RequestOption) (*Contact, error) {
	return call[Contact](ctx, s.t, request{method: "POST", path: pathf("/v1/public/contacts/%s/subscribe-back", contactID)}, opts)
}

// GetStatistics counts contacts by status — for the whole account or for one list (ListID).
// Scope: contacts.read.
func (s *ContactsService) GetStatistics(ctx context.Context, params *ContactStatisticsParams, opts ...RequestOption) (*ContactStatistics, error) {
	return call[ContactStatistics](ctx, s.t, request{method: "GET", path: "/v1/public/contacts/statistics", query: params}, opts)
}

// BulkDelete deletes several contacts at once.
// Scope: contacts.write.
func (s *ContactsService) BulkDelete(ctx context.Context, contactIDs []int, opts ...RequestOption) (*BulkDeleteResult, error) {
	return call[BulkDeleteResult](ctx, s.t, request{
		method: "POST",
		path:   "/v1/public/contacts/bulk-delete",
		body:   contactIDsBody{ContactIDs: contactIDs},
	}, opts)
}

// RemoveFromAllLists removes contacts from all lists without deleting them.
// Scope: contacts.write.
func (s *ContactsService) RemoveFromAllLists(ctx context.Context, contactIDs []int, opts ...RequestOption) (*RemoveFromAllListsResult, error) {
	return call[RemoveFromAllListsResult](ctx, s.t, request{
		method: "POST",
		path:   "/v1/public/contacts/remove-from-all-lists",
		body:   contactIDsBody{ContactIDs: contactIDs},
	}, opts)
}

// Export exports contacts as CSV or JSON. Returns the raw response: read or stream its Body and close it.
// Scope: contacts.read.
func (s *ContactsService) Export(ctx context.Context, params *ExportContactsParams, opts ...RequestOption) (*http.Response, error) {
	return s.t.raw(ctx, request{method: "GET", path: "/v1/public/contacts/export", query: params}, opts)
}

// ListCampaignHistory lists campaigns the contact received.
// Scope: contacts.read.
func (s *ContactsService) ListCampaignHistory(ctx context.Context, contactID int, params *PageParams, opts ...RequestOption) (*Page[CampaignSendingHistoryItem], error) {
	return listPage[CampaignSendingHistoryItem](ctx, s.t, request{
		method: "GET",
		path:   pathf("/v1/public/contacts/%s/distribution-sending-history", contactID),
		query:  params,
	}, opts)
}

// ListAPIHistory lists transactional emails (API and SMTP) the contact received.
// Scope: contacts.read.
func (s *ContactsService) ListAPIHistory(ctx context.Context, contactID int, params *PageParams, opts ...RequestOption) (*Page[APISendingHistoryItem], error) {
	return listPage[APISendingHistoryItem](ctx, s.t, request{
		method: "GET",
		path:   pathf("/v1/public/contacts/%s/api-sending-history", contactID),
		query:  params,
	}, opts)
}
