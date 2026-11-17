package rusender

import (
	"context"
	"encoding/json"
	"time"
)

// TemplateSummary is a template in a list, without HTML.
type TemplateSummary struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Alias *string `json:"alias"`
	// Type: html or unlayer (built in the editor).
	Type         string    `json:"type"`
	Thumbnail    *string   `json:"thumbnail,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	DashboardURL string    `json:"dashboardUrl"`
}

// Template is a template with its HTML.
type Template struct {
	TemplateSummary
	HTML string  `json:"html"`
	Text *string `json:"text,omitempty"`
}

// ListTemplatesParams are the parameters of [TemplatesService.List].
type ListTemplatesParams struct {
	PageParams
	// Name searches by name.
	Name string `json:"name,omitempty"`
}

// CreateTemplateParams are the parameters of [TemplatesService.Create].
type CreateTemplateParams struct {
	Name string `json:"name"`
	HTML string `json:"html"`
	Text string `json:"text,omitempty"`
	// Alias for sending with TemplateAlias, up to 64 characters.
	Alias string `json:"alias,omitempty"`
}

// UpdateTemplateParams are the parameters of [TemplatesService.Update].
type UpdateTemplateParams struct {
	Name string `json:"name,omitempty"`
	HTML string `json:"html,omitempty"`
	// Text: a pointer, so that String("") can clear the plain-text version.
	Text *string `json:"text,omitempty"`
	// Alias: Null removes the alias.
	Alias Nullable[string] `json:"alias"`
}

// MarshalJSON leaves out an unset Alias.
func (p UpdateTemplateParams) MarshalJSON() ([]byte, error) {
	type plain UpdateTemplateParams
	return json.Marshal(struct {
		plain
		Alias *Nullable[string] `json:"alias,omitempty"`
	}{plain(p), p.Alias.sent()})
}

// TemplatesService: email templates.
type TemplatesService struct{ t *transport }

// List lists templates without HTML, searchable by name. Use Get for the full template.
// Scope: templates.read.
func (s *TemplatesService) List(ctx context.Context, params *ListTemplatesParams, opts ...RequestOption) (*Page[TemplateSummary], error) {
	return listPage[TemplateSummary](ctx, s.t, request{method: "GET", path: "/v2/public/templates", query: params}, opts)
}

// Create creates an HTML template.
// Scope: templates.write.
func (s *TemplatesService) Create(ctx context.Context, params *CreateTemplateParams, opts ...RequestOption) (*Template, error) {
	return call[Template](ctx, s.t, request{method: "POST", path: "/v1/public/templates", body: params}, opts)
}

// Get gets a template with its HTML.
// Scope: templates.read.
func (s *TemplatesService) Get(ctx context.Context, templateID int, opts ...RequestOption) (*Template, error) {
	return call[Template](ctx, s.t, request{method: "GET", path: pathf("/v1/public/templates/%s", templateID)}, opts)
}

// Update updates an HTML template. Builder templates cannot be edited via the API.
// Scope: templates.write.
func (s *TemplatesService) Update(ctx context.Context, templateID int, params *UpdateTemplateParams, opts ...RequestOption) (*Template, error) {
	return call[Template](ctx, s.t, request{method: "PATCH", path: pathf("/v1/public/templates/%s", templateID), body: params}, opts)
}

// Delete deletes a template. Campaigns keep their own copy and are not affected; transactional emails that
// use this template (TemplateID, TemplateAlias) stop being sent.
// Scope: templates.write.
func (s *TemplatesService) Delete(ctx context.Context, templateID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/templates/%s", templateID)}, opts)
}
