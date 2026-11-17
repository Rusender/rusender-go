package rusender

import (
	"net/http"
	"os"
	"strings"
)

// Client is the RuSender public API client. Create it once and reuse it: it is safe for concurrent use.
type Client struct {
	// Emails: transactional emails — sending, status, history and recipient events.
	Emails *EmailsService
	// Contacts across all lists: search, update, subscription, export, sending history.
	Contacts *ContactsService
	// Lists: contact lists, their contacts (Lists.Contacts) and variables (Lists.Variables).
	Lists *ListsService
	// Campaigns: drafts, start, pause, A/B tests, chunked sending, statistics.
	Campaigns *CampaignsService
	// Folders: campaign folders.
	Folders *FoldersService
	// Templates: email templates.
	Templates *TemplatesService
	// Senders: sender addresses for campaigns.
	Senders *SendersService
	// Domains: sender domains and their DNS verification.
	Domains *DomainsService
	// Webhooks: subscriptions to email and campaign events.
	Webhooks *WebhooksService
	// Segments: contact segments built from conditions.
	Segments *SegmentsService
	// SendingKeys: sending keys used for transactional emails.
	SendingKeys *SendingKeysService
	// Files: account files, campaign attachments.
	Files *FilesService
	// Me: the account and scopes of the current API key.
	Me *MeService
}

// NewClient creates a client. apiKey is the API key token (rs_ck_…); when empty, it is read from the
// RUSENDER_API_KEY environment variable. Surrounding whitespace and quotes are trimmed. Returns ErrMissingAPIKey if there is no key, and an error if the
// timeout or the number of retries is invalid.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		apiKey = os.Getenv("RUSENDER_API_KEY")
	}
	// a key read from a secret file or an env file often carries a trailing newline or quotes: the server
	// would reject "Bearer rs_ck_…\n" with 401
	apiKey = strings.Trim(strings.TrimSpace(apiKey), `"'`)
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	cfg := clientConfig{
		baseURL:    DefaultBaseURL,
		timeout:    DefaultTimeout,
		maxRetries: DefaultMaxRetries,
		httpClient: http.DefaultClient,
		headers:    http.Header{},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.httpClient == nil {
		cfg.httpClient = http.DefaultClient
	}
	if err := validateSettings(cfg.timeout, cfg.maxRetries); err != nil {
		return nil, err
	}

	t := newTransport(apiKey, cfg)
	return &Client{
		Emails:      &EmailsService{t: t},
		Contacts:    &ContactsService{t: t},
		Lists:       &ListsService{t: t, Contacts: &ListContactsService{t: t}, Variables: &ListVariablesService{t: t}},
		Campaigns:   &CampaignsService{t: t},
		Folders:     &FoldersService{t: t},
		Templates:   &TemplatesService{t: t},
		Senders:     &SendersService{t: t},
		Domains:     &DomainsService{t: t},
		Webhooks:    &WebhooksService{t: t},
		Segments:    &SegmentsService{t: t},
		SendingKeys: &SendingKeysService{t: t},
		Files:       &FilesService{t: t},
		Me:          &MeService{t: t},
	}, nil
}

// String hides the API key: printing the client must not leak it into logs.
func (c *Client) String() string { return "rusender.Client{apiKey: [hidden]}" }

// GoString hides the API key for %#v.
func (c *Client) GoString() string { return c.String() }
