package rusender

import (
	"context"
	"time"
)

// DNSRecord is a DNS record to add to the domain's zone.
type DNSRecord struct {
	IsVerified bool   `json:"isVerified"`
	Type       string `json:"type"`
	Host       string `json:"host"`
	Data       string `json:"data"`
}

// DomainDNSRecords are the records that verify a domain.
type DomainDNSRecords struct {
	DKIM  DNSRecord `json:"dkim"`
	SPF   DNSRecord `json:"spf"`
	DMARC DNSRecord `json:"dmarc"`
}

// Domain is a sending domain.
type Domain struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	Status       Status           `json:"status"`
	IsVerified   bool             `json:"isVerified"`
	DNSRecords   DomainDNSRecords `json:"dnsRecords"`
	CheckedAt    time.Time        `json:"checkedAt"`
	DashboardURL string           `json:"dashboardUrl"`
}

// DomainVerification is a domain after a DNS check.
type DomainVerification struct {
	Domain
	// Warnings: SPF_MULTIPLE_RECORDS, DKIM_DUPLICATE_RECORDS.
	Warnings []string `json:"warnings"`
	// SuggestedSPFRecord is a merged SPF record to use if the current one is not valid.
	SuggestedSPFRecord *string `json:"suggestedSpfRecord,omitempty"`
}

// CreateDomainParams are the parameters of [DomainsService.Create].
type CreateDomainParams struct {
	Name string `json:"name"`
}

// DomainsService: sending domains and their DNS verification.
type DomainsService struct{ t *transport }

// List lists sending domains with their DNS records and verification status.
// Scope: domains.read.
func (s *DomainsService) List(ctx context.Context, params *PageParams, opts ...RequestOption) (*Page[Domain], error) {
	return listPage[Domain](ctx, s.t, request{method: "GET", path: "/v1/public/domains", query: params}, opts)
}

// Create adds a sending domain. Add the DNS records from DNSRecords (DKIM, SPF, DMARC) to the domain's
// zone, then call Verify.
// Scope: domains.write.
func (s *DomainsService) Create(ctx context.Context, params *CreateDomainParams, opts ...RequestOption) (*Domain, error) {
	return call[Domain](ctx, s.t, request{method: "POST", path: "/v1/public/domains", body: params}, opts)
}

// Get gets a domain with its DNS records.
// Scope: domains.read.
func (s *DomainsService) Get(ctx context.Context, domainID int, opts ...RequestOption) (*Domain, error) {
	return call[Domain](ctx, s.t, request{method: "GET", path: pathf("/v1/public/domains/%s", domainID)}, opts)
}

// Verify checks the domain's DKIM, SPF and DMARC records now and refreshes the verification status.
// May return Warnings and a SuggestedSPFRecord.
// Scope: domains.write.
func (s *DomainsService) Verify(ctx context.Context, domainID int, opts ...RequestOption) (*DomainVerification, error) {
	return call[DomainVerification](ctx, s.t, request{method: "POST", path: pathf("/v1/public/domains/%s/verify", domainID)}, opts)
}

// Delete deletes an unverified domain. Verified domains cannot be deleted via the API.
// Scope: domains.write.
func (s *DomainsService) Delete(ctx context.Context, domainID int, opts ...RequestOption) (*DeleteResult, error) {
	return call[DeleteResult](ctx, s.t, request{method: "DELETE", path: pathf("/v1/public/domains/%s", domainID)}, opts)
}
