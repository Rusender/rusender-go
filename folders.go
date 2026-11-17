package rusender

import (
	"context"
	"time"
)

// Folder is a campaign folder.
type Folder struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FolderCampaignsResult is the result of adding campaigns to a folder or removing them.
type FolderCampaignsResult struct {
	FolderID          int `json:"folderId"`
	AffectedCampaigns int `json:"affectedCampaigns"`
}

// ListFoldersParams are the parameters of [FoldersService.List].
type ListFoldersParams struct {
	PageParams
	// Q searches by name.
	Q string `json:"q,omitempty"`
}

// CreateFolderParams are the parameters of [FoldersService.Create].
type CreateFolderParams struct {
	Name string `json:"name"`
}

// UpdateFolderParams are the parameters of [FoldersService.Update].
type UpdateFolderParams struct {
	Name string `json:"name,omitempty"`
}

type campaignIDsBody struct {
	CampaignIDs []int `json:"campaignIds"`
}

// FoldersService: campaign folders.
type FoldersService struct{ t *transport }

// List lists campaign folders, searchable by name.
// Scope: campaigns.read.
func (s *FoldersService) List(ctx context.Context, params *ListFoldersParams, opts ...RequestOption) (*Page[Folder], error) {
	return listPage[Folder](ctx, s.t, request{method: "GET", path: "/v1/public/folders", query: params}, opts)
}

// Create creates a campaign folder.
// Scope: campaigns.write.
func (s *FoldersService) Create(ctx context.Context, params *CreateFolderParams, opts ...RequestOption) (*Folder, error) {
	return call[Folder](ctx, s.t, request{method: "POST", path: "/v1/public/folders", body: params}, opts)
}

// Get gets a campaign folder.
// Scope: campaigns.read.
func (s *FoldersService) Get(ctx context.Context, folderID int, opts ...RequestOption) (*Folder, error) {
	return call[Folder](ctx, s.t, request{method: "GET", path: pathf("/v1/public/folders/%s", folderID)}, opts)
}

// Update renames a campaign folder.
// Scope: campaigns.write.
func (s *FoldersService) Update(ctx context.Context, folderID int, params *UpdateFolderParams, opts ...RequestOption) (*Folder, error) {
	return call[Folder](ctx, s.t, request{method: "PATCH", path: pathf("/v1/public/folders/%s", folderID), body: params}, opts)
}

// Delete deletes a campaign folder. The campaigns in it are not deleted, only taken out of the folder.
// Scope: campaigns.write.
func (s *FoldersService) Delete(ctx context.Context, folderID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/folders/%s", folderID)}, opts)
}

// AddCampaigns puts campaigns into a folder. A campaign belongs to one folder, so it moves from its current
// folder. Unknown ids are ignored; AffectedCampaigns is the number of campaigns actually added.
// Scope: campaigns.write.
func (s *FoldersService) AddCampaigns(ctx context.Context, folderID int, campaignIDs []int, opts ...RequestOption) (*FolderCampaignsResult, error) {
	return call[FolderCampaignsResult](ctx, s.t, request{
		method: "POST",
		path:   pathf("/v1/public/folders/%s/campaigns/add", folderID),
		body:   campaignIDsBody{CampaignIDs: campaignIDs},
	}, opts)
}

// RemoveCampaigns takes campaigns out of a folder without deleting them. Unknown ids are ignored;
// AffectedCampaigns is the number of campaigns actually removed.
// Scope: campaigns.write.
func (s *FoldersService) RemoveCampaigns(ctx context.Context, folderID int, campaignIDs []int, opts ...RequestOption) (*FolderCampaignsResult, error) {
	return call[FolderCampaignsResult](ctx, s.t, request{
		method: "POST",
		path:   pathf("/v1/public/folders/%s/campaigns/remove", folderID),
		body:   campaignIDsBody{CampaignIDs: campaignIDs},
	}, opts)
}
