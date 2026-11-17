package rusender

import (
	"context"
	"encoding/json"
	"time"
)

// SegmentField is the field a segment condition checks. The type is open: the API may add values.
type SegmentField string

const (
	SegmentContactName          SegmentField = "contactName"
	SegmentContactEmail         SegmentField = "contactEmail"
	SegmentContactPhone         SegmentField = "contactPhone"
	SegmentContactStatus        SegmentField = "contactStatus"
	SegmentContactDateCreatedAt SegmentField = "contactDateCreatedAt"
	SegmentContactInList        SegmentField = "contactInList"
	SegmentActiveInCampaign     SegmentField = "activeInCampaign"
	SegmentEmailProvider        SegmentField = "emailProvider"
	SegmentVariable             SegmentField = "variable"
)

// SegmentOperator is the comparison of a segment condition: equal, notEqual, contain, notContain, thereIs,
// missing, participated, notParticipated, opened, notOpened, clicked, notClicked, before, after,
// findInList, notFindInList, sentTo.
type SegmentOperator string

// ConditionsJoin joins the conditions of a segment: and or or.
type ConditionsJoin string

const (
	JoinAnd ConditionsJoin = "and"
	JoinOr  ConditionsJoin = "or"
)

// SegmentConditionInput is a condition in a request.
type SegmentConditionInput struct {
	// ID, in Update: id of an existing condition of this segment to keep it and change it in place; conditions
	// without ID are added. An unknown id → ErrNotFound with SEGMENT_CONDITION_NOT_FOUND, a repeated id →
	// ErrValidation with SEGMENT_CONDITION_DUPLICATE_ID. Ignored in Create.
	ID    int          `json:"id,omitempty"`
	Field SegmentField `json:"field"`
	// VariableID is for Field SegmentVariable.
	VariableID int             `json:"variableId,omitempty"`
	Operator   SegmentOperator `json:"operator"`
	Value      string          `json:"value,omitempty"`
}

// SegmentCondition is a condition of a segment.
type SegmentCondition struct {
	ID         int             `json:"id"`
	Field      SegmentField    `json:"field"`
	VariableID *int            `json:"variableId,omitempty"`
	Operator   SegmentOperator `json:"operator"`
	Value      *string         `json:"value,omitempty"`
}

// Segment is a contact segment built from conditions.
type Segment struct {
	ID                  int                `json:"id"`
	Name                string             `json:"name"`
	ApplyConditionsType ConditionsJoin     `json:"applyConditionsType"`
	Conditions          []SegmentCondition `json:"conditions"`
	CreatedAt           time.Time          `json:"createdAt"`
	UpdatedAt           time.Time          `json:"updatedAt"`
	DashboardURL        string             `json:"dashboardUrl"`
}

// ListSegmentsParams are the parameters of [SegmentsService.List].
type ListSegmentsParams struct {
	PageParams
	// Name searches by name.
	Name string `json:"name,omitempty"`
}

// CreateSegmentParams are the parameters of [SegmentsService.Create].
type CreateSegmentParams struct {
	Name                string                  `json:"name"`
	ApplyConditionsType ConditionsJoin          `json:"applyConditionsType,omitempty"`
	Conditions          []SegmentConditionInput `json:"conditions,omitempty"`
}

// UpdateSegmentParams are the parameters of [SegmentsService.Update]. Conditions, when non-nil, becomes the
// new set; an empty, non-nil slice removes all conditions.
type UpdateSegmentParams struct {
	Name                string                  `json:"name,omitempty"`
	ApplyConditionsType ConditionsJoin          `json:"applyConditionsType,omitempty"`
	Conditions          []SegmentConditionInput `json:"conditions,omitempty"`
}

// MarshalJSON sends an empty, non-nil Conditions as [] instead of leaving it out.
func (p UpdateSegmentParams) MarshalJSON() ([]byte, error) {
	type plain UpdateSegmentParams
	return json.Marshal(struct {
		plain
		Conditions *[]SegmentConditionInput `json:"conditions,omitempty"`
	}{plain(p), keepEmpty(p.Conditions)})
}

// SegmentFilterOperator is an operator a filter field supports.
type SegmentFilterOperator struct {
	Operator SegmentOperator `json:"operator"`
	// ValueType: string, number, date or none.
	ValueType     string `json:"valueType"`
	ValueRequired bool   `json:"valueRequired"`
}

// SegmentFilter is a field available for conditions.
type SegmentFilter struct {
	Field      SegmentField            `json:"field"`
	VariableID *int                    `json:"variableId,omitempty"`
	Label      string                  `json:"label,omitempty"`
	Operators  []SegmentFilterOperator `json:"operators"`
}

// SegmentFilterValue is a selectable value of a filter field.
type SegmentFilterValue struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// SegmentsService: contact segments.
type SegmentsService struct{ t *transport }

// List lists segments with their conditions, searchable by name.
// Scope: segments.read.
func (s *SegmentsService) List(ctx context.Context, params *ListSegmentsParams, opts ...RequestOption) (*Page[Segment], error) {
	return listPage[Segment](ctx, s.t, request{method: "GET", path: "/v1/public/segments", query: params}, opts)
}

// Create creates a segment from conditions joined with ApplyConditionsType (and / or).
// Scope: segments.write.
func (s *SegmentsService) Create(ctx context.Context, params *CreateSegmentParams, opts ...RequestOption) (*Segment, error) {
	return call[Segment](ctx, s.t, request{method: "POST", path: "/v1/public/segments", body: params}, opts)
}

// Get gets a segment with all its conditions.
// Scope: segments.read.
func (s *SegmentsService) Get(ctx context.Context, segmentID int, opts ...RequestOption) (*Segment, error) {
	return call[Segment](ctx, s.t, request{method: "GET", path: pathf("/v1/public/segments/%s", segmentID)}, opts)
}

// Update updates a segment. If Conditions is non-nil, it becomes the new set: conditions with ID are kept
// and changed, without ID — added, the ones not listed — removed; an empty slice removes all. With nil
// Conditions they stay as is.
// Scope: segments.write.
func (s *SegmentsService) Update(ctx context.Context, segmentID int, params *UpdateSegmentParams, opts ...RequestOption) (*Segment, error) {
	return call[Segment](ctx, s.t, request{method: "PATCH", path: pathf("/v1/public/segments/%s", segmentID), body: params}, opts)
}

// ListFilters lists the fields available for conditions, including list variables, with the operators each
// field supports.
// Scope: segments.read.
func (s *SegmentsService) ListFilters(ctx context.Context, opts ...RequestOption) ([]SegmentFilter, error) {
	return listItems[SegmentFilter](ctx, s.t, request{method: "GET", path: "/v1/public/segments/filters"}, opts)
}

// ListFilterValues lists selectable values of a field with a fixed set of values — currently
// SegmentEmailProvider (mail providers).
// Scope: segments.read.
func (s *SegmentsService) ListFilterValues(ctx context.Context, field SegmentField, opts ...RequestOption) ([]SegmentFilterValue, error) {
	return listItems[SegmentFilterValue](ctx, s.t, request{method: "GET", path: pathf("/v1/public/segments/filters/%s/values", field)}, opts)
}
