package rusender

import (
	"context"
	"encoding/json"
	"time"
)

// CampaignType is the kind of a campaign. The type is open: the API may add values.
type CampaignType string

const (
	CampaignRegular       CampaignType = "regular"
	CampaignABBasic       CampaignType = "ab_basic"
	CampaignABVariant     CampaignType = "ab_variant"
	CampaignFollowUpBasic CampaignType = "follow_up_basic"
	CampaignFollowUp      CampaignType = "follow_up"
	CampaignChunked       CampaignType = "chunked"
)

// CampaignStatus is the status of a campaign.
type CampaignStatus string

const (
	CampaignDraft      CampaignStatus = "draft"
	CampaignBuilt      CampaignStatus = "built"
	CampaignScheduled  CampaignStatus = "scheduled"
	CampaignInProgress CampaignStatus = "in_progress"
	CampaignPaused     CampaignStatus = "paused"
	CampaignCompleted  CampaignStatus = "completed"
	CampaignBanned     CampaignStatus = "banned"
	CampaignOnHold     CampaignStatus = "on_hold"
)

// CampaignModerationStatus is the moderation status of a campaign: not_ready_to_moderate, ready_to_moderate,
// moderating, approved, banned or paused.
type CampaignModerationStatus string

// CampaignUTM are the UTM tags added to campaign links.
type CampaignUTM struct {
	Source   string `json:"utm_source,omitempty"`
	Medium   string `json:"utm_medium,omitempty"`
	Campaign string `json:"utm_campaign,omitempty"`
}

// StatMetric is a count with its rate.
type StatMetric struct {
	Count int `json:"count"`
	// Rate is a share from 0 to 1, e.g. 0.9 is 90%. Sending, delivery and errors are relative to all
	// recipients; opens, unsubscribes and complaints — to delivered; clicks — to opens.
	Rate float64 `json:"rate"`
}

// CampaignStats is the statistics of a campaign.
type CampaignStats struct {
	Total       int        `json:"total"`
	Sending     StatMetric `json:"sending"`
	Delivered   StatMetric `json:"delivered"`
	Open        StatMetric `json:"open"`
	Click       StatMetric `json:"click"`
	Error       StatMetric `json:"error"`
	Unsubscribe StatMetric `json:"unsubscribe"`
	Complaint   StatMetric `json:"complaint"`
}

// CampaignAttachment is a file attached to a campaign.
type CampaignAttachment struct {
	ID       int    `json:"id"`
	FileName string `json:"fileName"`
	Size     int    `json:"size"`
}

// Campaign is a campaign.
type Campaign struct {
	ID             int                      `json:"id"`
	Name           string                   `json:"name"`
	Subject        *string                  `json:"subject,omitempty"`
	PreviewTitle   *string                  `json:"previewTitle,omitempty"`
	Type           CampaignType             `json:"type"`
	Status         CampaignStatus           `json:"status"`
	StatusModerate CampaignModerationStatus `json:"statusModerate"`
	FolderID       *int                     `json:"folderId,omitempty"`
	IsArchived     bool                     `json:"isArchived"`
	DashboardURL   string                   `json:"dashboardUrl"`
	Attachments    []CampaignAttachment     `json:"attachments"`
	// TemplateID is the campaign's own copy of the template, not the id you passed: the campaign keeps a copy.
	TemplateID  *int           `json:"templateId,omitempty"`
	SenderID    *int           `json:"senderId,omitempty"`
	SenderEmail string         `json:"senderEmail,omitempty"`
	SenderName  *string        `json:"senderName,omitempty"`
	Lists       []ListRef      `json:"lists"`
	UTM         *CampaignUTM   `json:"utm,omitempty"`
	ScheduledAt *time.Time     `json:"scheduledAt,omitempty"`
	StartedAt   *time.Time     `json:"startedAt,omitempty"`
	FinishedAt  *time.Time     `json:"finishedAt,omitempty"`
	CreatedAt   *time.Time     `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time     `json:"updatedAt,omitempty"`
	Stats       *CampaignStats `json:"stats,omitempty"`
}

// CampaignPart is an A/B test variant or a follow-up in a campaign list.
type CampaignPart struct {
	ID int `json:"id"`
	// Role: ab_variant or follow_up.
	Role         string         `json:"role"`
	Name         string         `json:"name"`
	Subject      *string        `json:"subject,omitempty"`
	PreviewTitle *string        `json:"previewTitle,omitempty"`
	Status       CampaignStatus `json:"status"`
}

// CampaignListItem is a campaign in a list, with its parts.
type CampaignListItem struct {
	Campaign
	// Parts: A/B test variants and follow-ups.
	Parts []CampaignPart `json:"parts,omitempty"`
}

// ABTestVariant is a variant of an A/B campaign.
type ABTestVariant struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Subject      *string        `json:"subject,omitempty"`
	PreviewTitle *string        `json:"previewTitle,omitempty"`
	Status       CampaignStatus `json:"status"`
	SplitPercent float64        `json:"splitPercent"`
	IsWinner     *bool          `json:"isWinner,omitempty"`
	Stats        *CampaignStats `json:"stats,omitempty"`
}

// CampaignABTest describes the A/B test of a campaign.
type CampaignABTest struct {
	BasicID     int     `json:"basicId"`
	TestPercent float64 `json:"testPercent"`
	// ExperimentType: subject or template.
	ExperimentType   string          `json:"experimentType"`
	IsWinnerSelected bool            `json:"isWinnerSelected"`
	WinnerID         *int            `json:"winnerId,omitempty"`
	Variants         []ABTestVariant `json:"variants"`
}

// CampaignFollowUpInfo is the follow-up of a campaign.
type CampaignFollowUpInfo struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Subject      *string        `json:"subject,omitempty"`
	PreviewTitle *string        `json:"previewTitle,omitempty"`
	Status       CampaignStatus `json:"status"`
	OffsetHours  int            `json:"offsetHours"`
	Stats        *CampaignStats `json:"stats,omitempty"`
}

// CampaignParent is the base campaign of a part.
type CampaignParent struct {
	ID           int            `json:"id"`
	Type         CampaignType   `json:"type"`
	Name         string         `json:"name"`
	Subject      *string        `json:"subject,omitempty"`
	PreviewTitle *string        `json:"previewTitle,omitempty"`
	Status       CampaignStatus `json:"status"`
	Stats        *CampaignStats `json:"stats,omitempty"`
}

// CampaignDetails is a campaign with its A/B test, follow-up or parent.
type CampaignDetails struct {
	Campaign
	ABTest   *CampaignABTest       `json:"abTest,omitempty"`
	FollowUp *CampaignFollowUpInfo `json:"followUp,omitempty"`
	Parent   *CampaignParent       `json:"parent,omitempty"`
	// TotalStats: aggregated statistics of the campaign and its parts.
	TotalStats *CampaignStats `json:"totalStats,omitempty"`
}

// ABTestVariantParams is a variant of an A/B test in a request.
type ABTestVariantParams struct {
	Subject      string `json:"subject"`
	PreviewTitle string `json:"previewTitle,omitempty"`
}

// ABTestParams configure the A/B test of a campaign.
type ABTestParams struct {
	// Type: subject.
	Type     string                `json:"type,omitempty"`
	Variants []ABTestVariantParams `json:"variants"`
	// Percent is the share of the audience used for the test, %.
	Percent int `json:"percent,omitempty"`
}

// CreateCampaignParams are the parameters of [CampaignsService.Create].
type CreateCampaignParams struct {
	Name         string `json:"name,omitempty"`
	Subject      string `json:"subject,omitempty"`
	PreviewTitle string `json:"previewTitle,omitempty"`
	TemplateID   int    `json:"templateId,omitempty"`
	SenderID     int    `json:"senderId,omitempty"`
	SenderName   string `json:"senderName,omitempty"`
	ListIDs      []int  `json:"listIds,omitempty"`
	SegmentID    int    `json:"segmentId,omitempty"`
	// Attachments: ids of attached files, see Files.Upload.
	Attachments []int         `json:"attachments,omitempty"`
	UTM         *CampaignUTM  `json:"utm,omitempty"`
	ABTest      *ABTestParams `json:"abTest,omitempty"`
}

// UpdateCampaignParams are the parameters of [CampaignsService.Update]; only the fields you set change.
// An empty, non-nil ListIDs or Attachments clears it.
type UpdateCampaignParams CreateCampaignParams

// MarshalJSON sends an empty, non-nil ListIDs or Attachments as [] instead of leaving it out.
func (p UpdateCampaignParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CreateCampaignParams
		ListIDs     *[]int `json:"listIds,omitempty"`
		Attachments *[]int `json:"attachments,omitempty"`
	}{CreateCampaignParams(p), keepEmpty(p.ListIDs), keepEmpty(p.Attachments)})
}

// ListCampaignsParams are the parameters of [CampaignsService.List].
type ListCampaignsParams struct {
	PageParams
	Name   string           `json:"name,omitempty"`
	Status []CampaignStatus `json:"status,omitempty"`
	// WithStats includes statistics.
	WithStats bool `json:"withStats,omitempty"`
	// Archived lists archived campaigns instead of active ones.
	Archived bool `json:"archived,omitempty"`
}

// GetCampaignParams are the parameters of [CampaignsService.Get].
type GetCampaignParams struct {
	WithStats bool `json:"withStats,omitempty"`
}

// CampaignFollowUpParams add a follow-up email to recipients who did not open the campaign.
type CampaignFollowUpParams struct {
	Subject      string `json:"subject"`
	OffsetHours  int    `json:"offsetHours"`
	PreviewTitle string `json:"previewTitle,omitempty"`
}

// ChunkSettings configure chunked sending, one chunk per day.
type ChunkSettings struct {
	InitialChunkSize int `json:"initialChunkSize"`
	// IncreaseChunkSizePercent grows each next chunk by this percent; zero sends equal chunks.
	IncreaseChunkSizePercent int `json:"increaseChunkSizePercent,omitempty"`
}

// ScheduleCampaignParams are the parameters of [CampaignsService.Schedule].
type ScheduleCampaignParams struct {
	// ScheduledStartDate is a deferred start. Without it, sending starts after moderation.
	ScheduledStartDate *time.Time `json:"scheduledStartDate,omitempty"`
	// FollowUp: a follow-up email to recipients who did not open the campaign.
	FollowUp *CampaignFollowUpParams `json:"followUp,omitempty"`
	// ChunkSettings: chunked sending, one chunk per day.
	ChunkSettings *ChunkSettings `json:"chunkSettings,omitempty"`
}

// ArchiveResult is the result of archiving or unarchiving a campaign.
type ArchiveResult struct {
	Archived    bool  `json:"archived"`
	CampaignIDs []int `json:"campaignIds"`
}

// PreviewABTestParams are the parameters of [CampaignsService.PreviewABTest].
type PreviewABTestParams struct {
	Percent int `json:"percent,omitempty"`
}

// AudienceShare is a number of contacts with its percent.
type AudienceShare struct {
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

// ABTestPreview is how an A/B test would split the audience.
type ABTestPreview struct {
	TotalContacts  int           `json:"totalContacts"`
	Percent        int           `json:"percent"`
	VariantsCount  int           `json:"variantsCount"`
	PerVariant     AudienceShare `json:"perVariant"`
	WinnerAudience AudienceShare `json:"winnerAudience"`
	Warning        string        `json:"warning,omitempty"`
	Error          string        `json:"error,omitempty"`
	// AllowedMinPercent is nil when the audience is empty.
	AllowedMinPercent *int `json:"allowedMinPercent,omitempty"`
}

// PreviewChunksParams are the parameters of [CampaignsService.PreviewChunks].
type PreviewChunksParams struct {
	InitialChunkSize         int        `json:"initialChunkSize"`
	IncreaseChunkSizePercent int        `json:"increaseChunkSizePercent,omitempty"`
	ScheduledStartDate       *time.Time `json:"scheduledStartDate,omitempty"`
}

// Chunk is one chunk of a chunked sending plan.
type Chunk struct {
	Index             int       `json:"index"`
	RecipientsCount   int       `json:"recipientsCount"`
	ScheduledSendDate time.Time `json:"scheduledSendDate"`
}

// ChunksPreview is a chunked sending plan.
type ChunksPreview struct {
	TotalContacts int     `json:"totalContacts"`
	ChunksCount   int     `json:"chunksCount"`
	Chunks        []Chunk `json:"chunks"`
}

// CampaignActivity is the opens and clicks of a campaign on one day.
type CampaignActivity struct {
	// Date is the day as DD.MM.YYYY.
	Date   string `json:"date"`
	Opens  int    `json:"opens"`
	Clicks int    `json:"clicks"`
}

// LinkClicks is the number of clicks on a campaign link.
type LinkClicks struct {
	Link        string `json:"link"`
	TotalClicks int    `json:"totalClicks"`
}

// CampaignDomainStat is the statistics of a campaign for one recipient mailbox domain.
type CampaignDomainStat struct {
	GroupName   string     `json:"groupName"`
	GroupSlug   *string    `json:"groupSlug,omitempty"`
	IconBase64  *string    `json:"iconBase64,omitempty"`
	Domain      string     `json:"domain"`
	Total       int        `json:"total"`
	Open        StatMetric `json:"open"`
	Click       StatMetric `json:"click"`
	Error       StatMetric `json:"error"`
	Unsubscribe StatMetric `json:"unsubscribe"`
}

// CampaignRecipientStatus is the delivery and engagement status of a campaign recipient: delivered, opened,
// clicked, unsubscribed, complaint, error, soft_bounced, hard_bounced, error_infrastructure or not_delivered.
type CampaignRecipientStatus string

// CampaignRecipient is a recipient of a campaign.
type CampaignRecipient struct {
	ContactID    int                     `json:"contactId"`
	Email        string                  `json:"email"`
	Status       CampaignRecipientStatus `json:"status"`
	SentAt       *time.Time              `json:"sentAt"`
	ResponseSMTP string                  `json:"responseSmtp,omitempty"`
}

// CampaignRecipientsPage is a page of campaign recipients.
type CampaignRecipientsPage struct {
	Items []CampaignRecipient `json:"items"`
	// HasMore tells whether there is a next page.
	HasMore bool `json:"hasMore"`
}

// ListCampaignRecipientsParams are the parameters of [CampaignsService.ListRecipients].
type ListCampaignRecipientsParams struct {
	PageParams
	Status []CampaignRecipientStatus `json:"status,omitempty"`
}

// CampaignsService: campaigns.
type CampaignsService struct{ t *transport }

// List lists campaigns, filterable by name and status. Set WithStats to include statistics, Archived to list
// archived campaigns.
// Scope: campaigns.read.
func (s *CampaignsService) List(ctx context.Context, params *ListCampaignsParams, opts ...RequestOption) (*Page[CampaignListItem], error) {
	return listPage[CampaignListItem](ctx, s.t, request{method: "GET", path: "/v1/public/campaigns", query: params}, opts)
}

// Create creates a campaign draft. All fields are optional and can be set later with Update; a template,
// a sender, a subject and at least one list are required to start it.
// Scope: campaigns.write.
func (s *CampaignsService) Create(ctx context.Context, params *CreateCampaignParams, opts ...RequestOption) (*Campaign, error) {
	return call[Campaign](ctx, s.t, request{method: "POST", path: "/v1/public/campaigns", body: params}, opts)
}

// Get gets a campaign. For A/B and follow-up campaigns also returns ABTest, FollowUp or Parent; statistics
// (Stats, TotalStats) only with WithStats. Parts as a flat Parts list come with List.
// Scope: campaigns.read.
func (s *CampaignsService) Get(ctx context.Context, campaignID int, params *GetCampaignParams, opts ...RequestOption) (*CampaignDetails, error) {
	return call[CampaignDetails](ctx, s.t, request{method: "GET", path: pathf("/v1/public/campaigns/%s", campaignID), query: params}, opts)
}

// Update updates a campaign draft; only the fields you set change. For A/B campaigns ABTest.Variants
// replaces the whole set of variants.
// Scope: campaigns.write.
func (s *CampaignsService) Update(ctx context.Context, campaignID int, params *UpdateCampaignParams, opts ...RequestOption) (*Campaign, error) {
	return call[Campaign](ctx, s.t, request{method: "PATCH", path: pathf("/v1/public/campaigns/%s", campaignID), body: params}, opts)
}

// Delete deletes a campaign draft. Only drafts can be deleted.
// Scope: campaigns.write.
func (s *CampaignsService) Delete(ctx context.Context, campaignID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/campaigns/%s", campaignID)}, opts)
}

// Schedule starts a campaign draft. Sends real emails. The campaign goes to moderation and is sent after
// approval, or at ScheduledStartDate. Optionally adds a follow-up (FollowUp) or sends in daily chunks
// (ChunkSettings); neither is available for A/B campaigns.
// Not enough balance → ErrPaymentRequired with code CAMPAIGN_INSUFFICIENT_BALANCE: top up and retry, the same
// idempotency key is fine. A failed start leaves the draft unchanged, ScheduledStartDate included.
// Scope: campaigns.write.
func (s *CampaignsService) Schedule(ctx context.Context, campaignID int, params *ScheduleCampaignParams, opts ...RequestOption) (*Campaign, error) {
	return call[Campaign](ctx, s.t, request{method: "POST", path: pathf("/v1/public/campaigns/%s/schedule", campaignID), body: params}, opts)
}

// Archive moves a completed or banned campaign to the archive. A/B and follow-up campaigns are archived as
// a whole.
// Scope: campaigns.write.
func (s *CampaignsService) Archive(ctx context.Context, campaignID int, opts ...RequestOption) (*ArchiveResult, error) {
	return call[ArchiveResult](ctx, s.t, request{method: "POST", path: pathf("/v1/public/campaigns/%s/archive", campaignID)}, opts)
}

// Unarchive restores a campaign from the archive.
// Scope: campaigns.write.
func (s *CampaignsService) Unarchive(ctx context.Context, campaignID int, opts ...RequestOption) (*ArchiveResult, error) {
	return call[ArchiveResult](ctx, s.t, request{method: "POST", path: pathf("/v1/public/campaigns/%s/unarchive", campaignID)}, opts)
}

// Pause pauses a campaign that is being sent. Emails already handed over for delivery are not recalled;
// for an A/B campaign the whole experiment is paused. Returns ErrValidation with code CAMPAIGN_NOT_PAUSABLE
// if the campaign is not being sent, and with CAMPAIGN_IS_AB_VARIANT for an A/B test variant — pause the
// base A/B campaign instead.
// Scope: campaigns.write.
func (s *CampaignsService) Pause(ctx context.Context, campaignID int, opts ...RequestOption) (*Campaign, error) {
	return call[Campaign](ctx, s.t, request{method: "POST", path: pathf("/v1/public/campaigns/%s/pause", campaignID)}, opts)
}

// Resume resumes a paused campaign approved by moderation. Sends real emails; the campaign is checked again
// before sending. Returns ErrValidation with code CAMPAIGN_NOT_RESUMABLE if the campaign cannot be resumed,
// and with CAMPAIGN_IS_AB_VARIANT for an A/B test variant — resume the base A/B campaign instead.
// Scope: campaigns.write.
func (s *CampaignsService) Resume(ctx context.Context, campaignID int, opts ...RequestOption) (*Campaign, error) {
	return call[Campaign](ctx, s.t, request{method: "POST", path: pathf("/v1/public/campaigns/%s/resume", campaignID)}, opts)
}

// PreviewABTest calculates how an A/B test would split the audience for the given percent: contacts per
// variant and the share that receives the winner. Does not change the campaign.
// Scope: campaigns.read.
func (s *CampaignsService) PreviewABTest(ctx context.Context, campaignID int, params *PreviewABTestParams, opts ...RequestOption) (*ABTestPreview, error) {
	return call[ABTestPreview](ctx, s.t, request{method: "GET", path: pathf("/v1/public/campaigns/%s/ab-test/preview", campaignID), query: params}, opts)
}

// PreviewChunks calculates the chunked sending plan: number of chunks, recipients and send date of each
// chunk. Does not change the campaign; pass the same ChunkSettings to Schedule.
// Scope: campaigns.read.
func (s *CampaignsService) PreviewChunks(ctx context.Context, campaignID int, params *PreviewChunksParams, opts ...RequestOption) (*ChunksPreview, error) {
	return call[ChunksPreview](ctx, s.t, request{method: "GET", path: pathf("/v1/public/campaigns/%s/chunks/preview", campaignID), query: params}, opts)
}

// GetActivity gets opens and clicks by day, Date as DD.MM.YYYY: 11 days from the campaign start, for
// chunked sending at least the number of chunks.
// Scope: campaigns.read.
func (s *CampaignsService) GetActivity(ctx context.Context, campaignID int, opts ...RequestOption) ([]CampaignActivity, error) {
	return listItems[CampaignActivity](ctx, s.t, request{method: "GET", path: pathf("/v1/public/campaigns/%s/activity", campaignID)}, opts)
}

// GetLinkClicks gets click counts per link.
// Scope: campaigns.read.
func (s *CampaignsService) GetLinkClicks(ctx context.Context, campaignID int, opts ...RequestOption) ([]LinkClicks, error) {
	return listItems[LinkClicks](ctx, s.t, request{method: "GET", path: pathf("/v1/public/campaigns/%s/link-clicks", campaignID)}, opts)
}

// GetStatsByDomains gets campaign statistics broken down by recipient mailbox domain.
// Scope: campaigns.read.
func (s *CampaignsService) GetStatsByDomains(ctx context.Context, campaignID int, opts ...RequestOption) ([]CampaignDomainStat, error) {
	return listItems[CampaignDomainStat](ctx, s.t, request{method: "GET", path: pathf("/v1/public/campaigns/%s/stats-by-domains", campaignID)}, opts)
}

// ListRecipients lists campaign recipients with their delivery and engagement status, filterable by status.
// Scope: campaigns.read.
func (s *CampaignsService) ListRecipients(ctx context.Context, campaignID int, params *ListCampaignRecipientsParams, opts ...RequestOption) (*CampaignRecipientsPage, error) {
	req := request{method: "GET", path: pathf("/v1/public/campaigns/%s/recipients", campaignID), query: params}
	res, err := s.t.do(ctx, req, opts)
	if err != nil {
		return nil, err
	}
	items, err := itemsOf[CampaignRecipient](req, res.data)
	if err != nil {
		return nil, err
	}
	page, err := decode[CampaignRecipientsPage](req, res.data)
	if err != nil {
		return nil, err
	}
	page.Items = items
	return &page, nil
}
