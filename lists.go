package rusender

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// MailingList is a contact list.
type MailingList struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Status        Status    `json:"status"`
	ContactsCount int       `json:"contactsCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	DashboardURL  string    `json:"dashboardUrl"`
}

// ListListsParams are the parameters of [ListsService.List].
type ListListsParams struct {
	PageParams
	// Q searches by name.
	Q      string `json:"q,omitempty"`
	Status Status `json:"status,omitempty"`
	// SortBy: id, name, createdAt or updatedAt.
	SortBy  string        `json:"sortBy,omitempty"`
	SortDir SortDirection `json:"sortDir,omitempty"`
}

// CreateListParams are the parameters of [ListsService.Create].
type CreateListParams struct {
	Name string `json:"name"`
}

// UpdateListParams are the parameters of [ListsService.Update].
type UpdateListParams struct {
	Name   string `json:"name,omitempty"`
	Status Status `json:"status,omitempty"`
}

// ListStatistics is the statistics of a list. Rates are shares from 0 to 1.
type ListStatistics struct {
	ListID                 int     `json:"listId"`
	CountActiveContacts    int     `json:"countActiveContacts"`
	CountNotActiveContacts int     `json:"countNotActiveContacts"`
	CountContacts          int     `json:"countContacts"`
	CountUnsubscribe       int     `json:"countUnsubscribe"`
	CountSending           int     `json:"countSending"`
	CountOpen              int     `json:"countOpen"`
	CountError             int     `json:"countError"`
	CountComplaint         int     `json:"countComplaint"`
	OpenRate               float64 `json:"openRate"`
	ErrorRate              float64 `json:"errorRate"`
	UnsubscribeRate        float64 `json:"unsubscribeRate"`
	ComplaintRate          float64 `json:"complaintRate"`
}

// ListsService: contact lists, their contacts and variables.
type ListsService struct {
	t *transport
	// Contacts: contacts of a list.
	Contacts *ListContactsService
	// Variables: variables of a list.
	Variables *ListVariablesService
}

// List lists contact lists, searchable by name.
// Scope: contacts.read.
func (s *ListsService) List(ctx context.Context, params *ListListsParams, opts ...RequestOption) (*Page[MailingList], error) {
	return listPage[MailingList](ctx, s.t, request{method: "GET", path: "/v1/public/lists", query: params}, opts)
}

// Create creates a contact list.
// Scope: contacts.write.
func (s *ListsService) Create(ctx context.Context, params *CreateListParams, opts ...RequestOption) (*MailingList, error) {
	return call[MailingList](ctx, s.t, request{method: "POST", path: "/v1/public/lists", body: params}, opts)
}

// Get gets a list with its contact count.
// Scope: contacts.read.
func (s *ListsService) Get(ctx context.Context, listID int, opts ...RequestOption) (*MailingList, error) {
	return call[MailingList](ctx, s.t, request{method: "GET", path: pathf("/v1/public/lists/%s", listID)}, opts)
}

// Update renames a list or changes its status.
// Scope: contacts.write.
func (s *ListsService) Update(ctx context.Context, listID int, params *UpdateListParams, opts ...RequestOption) (*MailingList, error) {
	return call[MailingList](ctx, s.t, request{method: "PATCH", path: pathf("/v1/public/lists/%s", listID), body: params}, opts)
}

// Delete deletes a list together with contacts that belong only to this list.
// Scope: contacts.write.
func (s *ListsService) Delete(ctx context.Context, listID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/lists/%s", listID)}, opts)
}

// GetStatistics gets list statistics: contacts by status, sends, opens, errors, unsubscribes, complaints.
// Scope: contacts.read.
func (s *ListsService) GetStatistics(ctx context.Context, listID int, opts ...RequestOption) (*ListStatistics, error) {
	return call[ListStatistics](ctx, s.t, request{method: "GET", path: pathf("/v1/public/lists/%s/statistics", listID)}, opts)
}

// CreateListContactParams are the parameters of [ListContactsService.Create].
type CreateListContactParams struct {
	Email     string           `json:"email"`
	Name      string           `json:"name,omitempty"`
	Phone     string           `json:"phone,omitempty"`
	Variables ContactVariables `json:"variables,omitempty"`
}

// ContactUpsertResult is the result of [ListContactsService.Create].
type ContactUpsertResult struct {
	Contact
	// Created is true when the contact was created, false when an existing contact was reused.
	Created bool `json:"created"`
}

// UpdateListContactParams are the parameters of [ListContactsService.Update]. Name and Phone are pointers, so
// that String("") can clear them.
type UpdateListContactParams struct {
	Name      *string          `json:"name,omitempty"`
	Phone     *string          `json:"phone,omitempty"`
	Variables ContactVariables `json:"variables,omitempty"`
}

// ListContactsInListParams are the parameters of [ListContactsService.List].
type ListContactsInListParams struct {
	PageParams
	Email   string           `json:"email,omitempty"`
	Status  []ContactStatus  `json:"status,omitempty"`
	SortBy  ContactSortField `json:"sortBy,omitempty"`
	SortDir SortDirection    `json:"sortDir,omitempty"`
}

// ImportContactRow is one contact of an import.
type ImportContactRow struct {
	Email     string           `json:"email,omitempty"`
	Name      string           `json:"name,omitempty"`
	Phone     string           `json:"phone,omitempty"`
	Variables ContactVariables `json:"variables,omitempty"`
}

// ImportVariable is a list variable created before an import.
type ImportVariable struct {
	Key  string       `json:"key"`
	Name string       `json:"name"`
	Type VariableType `json:"type,omitempty"`
}

// ImportContactsParams are the parameters of [ListContactsService.Import].
type ImportContactsParams struct {
	Rows []ImportContactRow `json:"rows"`
	// Variables to create in the list before the import.
	Variables []ImportVariable `json:"variables,omitempty"`
	// UpdateExisting: whether to update existing contacts, false by default.
	UpdateExisting bool `json:"updateExisting,omitempty"`
}

// ImportError is a sample error of an import row.
type ImportError struct {
	Row     int    `json:"row"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ImportContactsResult is the result of [ListContactsService.Import].
type ImportContactsResult struct {
	Processed      int           `json:"processed"`
	AddedCount     int           `json:"addedCount"`
	UpdatedCount   int           `json:"updatedCount"`
	DuplicateCount int           `json:"duplicateCount"`
	ErrorCount     int           `json:"errorCount"`
	ErrorsSample   []ImportError `json:"errorsSample,omitempty"`
}

// ExportListContactsParams are the parameters of [ListContactsService.Export].
type ExportListContactsParams struct {
	Format  ExportFormat     `json:"format"`
	Email   string           `json:"email,omitempty"`
	Status  []ContactStatus  `json:"status,omitempty"`
	SortBy  ContactSortField `json:"sortBy,omitempty"`
	SortDir SortDirection    `json:"sortDir,omitempty"`
	Limit   int              `json:"limit,omitempty"`
}

// MoveContactsParams are the parameters of [ListContactsService.Move].
type MoveContactsParams struct {
	ToListID   int   `json:"toListId"`
	ContactIDs []int `json:"contactIds"`
}

// ListContactsResult is the result of adding contacts to a list or removing them.
type ListContactsResult struct {
	ListID           int `json:"listId"`
	AffectedContacts int `json:"affectedContacts"`
}

// MoveContactsResult is the result of [ListContactsService.Move].
type MoveContactsResult struct {
	MovedContacts int `json:"movedContacts"`
}

// ListContactsService: contacts of a list.
type ListContactsService struct{ t *transport }

// List lists contacts in a list, filterable by email and status.
// Scope: contacts.read.
func (s *ListContactsService) List(ctx context.Context, listID int, params *ListContactsInListParams, opts ...RequestOption) (*Page[Contact], error) {
	return listPage[Contact](ctx, s.t, request{method: "GET", path: pathf("/v1/public/lists/%s/contacts", listID), query: params}, opts)
}

// Create adds a contact to a list. If a contact with this email exists, it is reused: Name and Phone from
// the request are ignored, Variables are applied. Created in the result tells the two apart.
// Scope: contacts.write.
func (s *ListContactsService) Create(ctx context.Context, listID int, params *CreateListContactParams, opts ...RequestOption) (*ContactUpsertResult, error) {
	return call[ContactUpsertResult](ctx, s.t, request{method: "POST", path: pathf("/v1/public/lists/%s/contacts", listID), body: params}, opts)
}

// Update updates a contact in a list: name, phone and list variables.
// Scope: contacts.write.
func (s *ListContactsService) Update(ctx context.Context, listID, contactID int, params *UpdateListContactParams, opts ...RequestOption) (*Contact, error) {
	return call[Contact](ctx, s.t, request{
		method: "PATCH",
		path:   pathf("/v1/public/lists/%s/contacts/%s", listID, contactID),
		body:   params,
	}, opts)
}

// Add adds existing contacts to a list.
// Scope: contacts.write.
func (s *ListContactsService) Add(ctx context.Context, listID int, contactIDs []int, opts ...RequestOption) (*ListContactsResult, error) {
	return call[ListContactsResult](ctx, s.t, request{
		method: "POST",
		path:   pathf("/v1/public/lists/%s/contacts/add", listID),
		body:   contactIDsBody{ContactIDs: contactIDs},
	}, opts)
}

// Remove removes contacts from a list. The contacts are neither deleted nor unsubscribed.
// Scope: contacts.write.
func (s *ListContactsService) Remove(ctx context.Context, listID int, contactIDs []int, opts ...RequestOption) (*ListContactsResult, error) {
	return call[ListContactsResult](ctx, s.t, request{
		method: "POST",
		path:   pathf("/v1/public/lists/%s/contacts/remove", listID),
		body:   contactIDsBody{ContactIDs: contactIDs},
	}, opts)
}

// Move moves contacts from this list to another one (ToListID).
// Scope: contacts.write.
func (s *ListContactsService) Move(ctx context.Context, listID int, params *MoveContactsParams, opts ...RequestOption) (*MoveContactsResult, error) {
	return call[MoveContactsResult](ctx, s.t, request{method: "POST", path: pathf("/v1/public/lists/%s/contacts/move", listID), body: params}, opts)
}

// Import imports contacts as rows. Runs synchronously; counts and sample errors are in the result.
// Pass Variables to create list variables before the import.
// Scope: contacts.write.
func (s *ListContactsService) Import(ctx context.Context, listID int, params *ImportContactsParams, opts ...RequestOption) (*ImportContactsResult, error) {
	return call[ImportContactsResult](ctx, s.t, request{method: "POST", path: pathf("/v1/public/lists/%s/contacts/import", listID), body: params}, opts)
}

// Export exports the contacts of a list as CSV or JSON. Returns the raw response: read or stream its Body
// and close it.
// Scope: contacts.read.
func (s *ListContactsService) Export(ctx context.Context, listID int, params *ExportListContactsParams, opts ...RequestOption) (*http.Response, error) {
	return s.t.raw(ctx, request{method: "GET", path: pathf("/v1/public/lists/%s/contacts/export", listID), query: params}, opts)
}

// VariableType is the type of a list variable: string or phone.
type VariableType string

const (
	VariableString VariableType = "string"
	VariablePhone  VariableType = "phone"
)

// Variable is a list variable.
type Variable struct {
	ID           int          `json:"id"`
	Key          string       `json:"key"`
	Name         string       `json:"name"`
	Type         VariableType `json:"type"`
	DefaultValue *string      `json:"defaultValue,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

// CreateVariableParams are the parameters of [ListVariablesService.Create].
type CreateVariableParams struct {
	Key          string       `json:"key"`
	Name         string       `json:"name"`
	Type         VariableType `json:"type,omitempty"`
	DefaultValue string       `json:"defaultValue,omitempty"`
}

// UpdateVariableParams are the parameters of [ListVariablesService.Update].
type UpdateVariableParams struct {
	Name string `json:"name,omitempty"`
	// DefaultValue: Null clears the default value.
	DefaultValue Nullable[string] `json:"defaultValue"`
}

// MarshalJSON leaves out an unset DefaultValue.
func (p UpdateVariableParams) MarshalJSON() ([]byte, error) {
	type plain UpdateVariableParams
	return json.Marshal(struct {
		plain
		DefaultValue *Nullable[string] `json:"defaultValue,omitempty"`
	}{plain(p), p.DefaultValue.sent()})
}

// ListVariablesService: variables of a list.
type ListVariablesService struct{ t *transport }

// List lists the variables of a list.
// Scope: contacts.read.
func (s *ListVariablesService) List(ctx context.Context, listID int, opts ...RequestOption) ([]Variable, error) {
	return listItems[Variable](ctx, s.t, request{method: "GET", path: pathf("/v1/public/lists/%s/variables", listID)}, opts)
}

// Create creates a list variable. Key is used in Variables of contacts and in templates.
// Scope: contacts.write.
func (s *ListVariablesService) Create(ctx context.Context, listID int, params *CreateVariableParams, opts ...RequestOption) (*Variable, error) {
	return call[Variable](ctx, s.t, request{method: "POST", path: pathf("/v1/public/lists/%s/variables", listID), body: params}, opts)
}

// Update updates the name or default value of a list variable.
// Scope: contacts.write.
func (s *ListVariablesService) Update(ctx context.Context, listID, variableID int, params *UpdateVariableParams, opts ...RequestOption) (*Variable, error) {
	return call[Variable](ctx, s.t, request{
		method: "PATCH",
		path:   pathf("/v1/public/lists/%s/variables/%s", listID, variableID),
		body:   params,
	}, opts)
}

// Delete deletes a list variable together with its values in contacts.
// Scope: contacts.write.
func (s *ListVariablesService) Delete(ctx context.Context, listID, variableID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/lists/%s/variables/%s", listID, variableID)}, opts)
}
