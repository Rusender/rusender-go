package rusender

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"reflect"
	"strings"
	"testing"
	"time"
)

func assertJSON(t *testing.T, got map[string]any, want string) {
	t.Helper()
	var expected map[string]any
	if err := json.Unmarshal([]byte(want), &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, expected) {
		gotJSON, _ := json.Marshal(got)
		t.Errorf("body\n got %s\nwant %s", gotJSON, want)
	}
}

var acceptedEmail = map[string]any{"id": "em_1", "status": "queued", "createdAt": "2026-11-12T12:00:00Z", "recipients": []any{}}

func TestSendEmail(t *testing.T) {
	client, s := newStubClient(t, []step{okStatus(202, acceptedEmail)})
	email, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:          Address{Email: "no-reply@shop.ru", Name: "Shop"},
		To:            []Address{Addr("a@example.ru"), {Email: "b@example.ru", Name: "B"}},
		Cc:            []Address{Addr("c@example.ru")},
		Subject:       "Order",
		TemplateAlias: "order-created",
		Variables:     map[string]any{"orderId": 42},
	})
	if err != nil {
		t.Fatal(err)
	}
	if email.ID != "em_1" || email.Status != EmailQueued {
		t.Errorf("email %+v", email)
	}
	req := s.request(t, 0)
	if req.Method != "POST" || req.URL.Path != "/api/v1/public/emails" || req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("%s %s %s", req.Method, req.URL.Path, req.Header.Get("Content-Type"))
	}
	assertJSON(t, req.data(t), `{
		"from": {"email": "no-reply@shop.ru", "name": "Shop"},
		"to": [{"email": "a@example.ru"}, {"email": "b@example.ru", "name": "B"}],
		"cc": [{"email": "c@example.ru"}],
		"subject": "Order",
		"templateAlias": "order-created",
		"variables": {"orderId": 42}
	}`)
}

func TestSendEmailAttachments(t *testing.T) {
	client, s := newStubClient(t, []step{okStatus(202, acceptedEmail)})
	_, err := client.Emails.Send(context.Background(), &SendEmailParams{
		From:    Addr("no-reply@shop.ru"),
		To:      []Address{Addr("a@example.ru")},
		Subject: "Invoice",
		Text:    "See attached",
		Attachments: []EmailAttachment{
			{FileName: "a.txt", Content: []byte("hello")},
			{FileID: 101},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	attachments := s.request(t, 0).data(t)["attachments"]
	want := []any{
		map[string]any{"fileName": "a.txt", "content": "aGVsbG8="},
		map[string]any{"fileId": float64(101)},
	}
	if !reflect.DeepEqual(attachments, want) {
		t.Errorf("attachments %v", attachments)
	}
}

func TestPage(t *testing.T) {
	pagination := map[string]any{"page": 2, "limit": 20, "totalItems": 21, "totalPages": 2}
	client, s := newStubClient(t, []step{ok(map[string]any{"items": []any{map[string]any{"id": 1, "name": "Clients"}}},
		map[string]any{"pagination": pagination})})
	page, err := client.Lists.List(context.Background(), &ListListsParams{Q: "Cli", PageParams: PageParams{Page: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "Clients" || page.Pagination != (Pagination{2, 20, 21, 2}) {
		t.Errorf("page %+v", page)
	}
	if got := s.request(t, 0).URL.RawQuery; got != "page=2&q=Cli" {
		t.Errorf("query %s", got)
	}
}

func TestCursorPage(t *testing.T) {
	client, _ := newStubClient(t, []step{ok(map[string]any{
		"items": []any{map[string]any{"id": "em_1", "status": "delivered"}}, "hasMore": true, "nextCursor": "c2",
	})})
	page, err := client.Emails.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || !page.HasMore || page.NextCursor == nil || *page.NextCursor != "c2" {
		t.Errorf("page %+v", page)
	}
}

func TestPlainList(t *testing.T) {
	client, _ := newStubClient(t, []step{ok(map[string]any{"items": []any{map[string]any{"id": 1, "key": "city"}}})})
	variables, err := client.Lists.Variables.List(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(variables) != 1 || variables[0].Key != "city" {
		t.Errorf("variables %+v", variables)
	}
}

func TestAddContactsToList(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"listId": 3, "affectedContacts": 2})})
	result, err := client.Lists.Contacts.Add(context.Background(), 3, []int{10, 11})
	if err != nil {
		t.Fatal(err)
	}
	if result.AffectedContacts != 2 {
		t.Errorf("result %+v", result)
	}
	req := s.request(t, 0)
	if req.URL.Path != "/api/v1/public/lists/3/contacts/add" {
		t.Errorf("path %s", req.URL.Path)
	}
	assertJSON(t, req.data(t), `{"contactIds": [10, 11]}`)
}

func TestNilParamsSendEmptyObject(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"id": 5, "name": "Draft"})})
	if _, err := client.Campaigns.Create(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if got := string(s.request(t, 0).body); got != `{"data":{}}` {
		t.Errorf("body %s", got)
	}
}

func TestPostWithoutBody(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"id": 5})})
	if _, err := client.Campaigns.Pause(context.Background(), 5); err != nil {
		t.Fatal(err)
	}
	req := s.request(t, 0)
	if len(req.body) != 0 || req.Header.Get("Content-Type") != "" {
		t.Errorf("body %q, content type %q", req.body, req.Header.Get("Content-Type"))
	}
	if req.Header.Get("Idempotency-Key") == "" {
		t.Error("pause must send Idempotency-Key")
	}
}

func TestNullable(t *testing.T) {
	cases := []struct {
		params UpdateTemplateParams
		want   string
	}{
		{UpdateTemplateParams{Name: "Order"}, `{"name": "Order"}`},
		{UpdateTemplateParams{Alias: Null[string]()}, `{"alias": null}`},
		{UpdateTemplateParams{Alias: Value("order")}, `{"alias": "order"}`},
	}
	for _, c := range cases {
		client, s := newStubClient(t, []step{ok(map[string]any{"id": 1})})
		if _, err := client.Templates.Update(context.Background(), 1, &c.params); err != nil {
			t.Fatal(err)
		}
		assertJSON(t, s.request(t, 0).data(t), c.want)
	}

	client, s := newStubClient(t, []step{ok(map[string]any{"id": 1})})
	if _, err := client.Lists.Variables.Update(context.Background(), 1, 2, &UpdateVariableParams{DefaultValue: Null[string]()}); err != nil {
		t.Fatal(err)
	}
	assertJSON(t, s.request(t, 0).data(t), `{"defaultValue": null}`)
}

func TestEmptySliceClearsInUpdate(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"id": 1}), ok(map[string]any{"id": 1}), ok(map[string]any{"id": 1})})
	ctx := context.Background()
	if _, err := client.Segments.Update(ctx, 1, &UpdateSegmentParams{Conditions: []SegmentConditionInput{}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Segments.Update(ctx, 1, &UpdateSegmentParams{Name: "VIP"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Campaigns.Update(ctx, 1, &UpdateCampaignParams{Attachments: []int{}, Subject: "Hi"}); err != nil {
		t.Fatal(err)
	}
	assertJSON(t, s.request(t, 0).data(t), `{"conditions": []}`)
	assertJSON(t, s.request(t, 1).data(t), `{"name": "VIP"}`)
	assertJSON(t, s.request(t, 2).data(t), `{"subject": "Hi", "attachments": []}`)
}

func TestScheduleSendsDateAsISO(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"id": 5})})
	start := time.Date(2026, 12, 1, 9, 0, 0, 0, time.UTC)
	if _, err := client.Campaigns.Schedule(context.Background(), 5, &ScheduleCampaignParams{ScheduledStartDate: &start}); err != nil {
		t.Fatal(err)
	}
	assertJSON(t, s.request(t, 0).data(t), `{"scheduledStartDate": "2026-12-01T09:00:00Z"}`)
}

func TestTemplatesListUsesV2(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"items": []any{}}, testPagination)})
	if _, err := client.Templates.List(context.Background(), &ListTemplatesParams{Name: "Order"}); err != nil {
		t.Fatal(err)
	}
	if got := s.request(t, 0).URL.Path; got != "/api/v2/public/templates" {
		t.Errorf("path %s", got)
	}
}

func TestUploadFiles(t *testing.T) {
	client, s := newStubClient(t, []step{okStatus(201, map[string]any{"files": []any{
		map[string]any{"id": 1, "fileName": "a.pdf", "size": 4, "createdAt": "2026-11-12T12:00:00Z"},
		map[string]any{"id": 2, "fileName": "b.png", "size": 3, "createdAt": "2026-11-12T12:00:00Z"},
	}})})
	recordSleeps(client)
	files, err := client.Files.Upload(context.Background(), []UploadFile{
		{FileName: "a.pdf", Content: strings.NewReader("%PDF")},
		{FileName: "b.png", Content: bytes.NewReader([]byte{1, 2, 3})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[1].ID != 2 {
		t.Errorf("files %+v", files)
	}

	req := s.request(t, 0)
	if req.Header.Get("Idempotency-Key") != "" {
		t.Error("upload must not send Idempotency-Key")
	}
	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("content type %q", req.Header.Get("Content-Type"))
	}
	reader := multipart.NewReader(bytes.NewReader(req.body), params["boundary"])
	var names []string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		content, _ := io.ReadAll(part)
		if part.FormName() != "file" {
			t.Errorf("field %q", part.FormName())
		}
		names = append(names, part.FileName()+":"+string(content))
	}
	if strings.Join(names, "|") != "a.pdf:%PDF|b.png:\x01\x02\x03" {
		t.Errorf("parts %q", names)
	}
}

func TestUploadIsNotRetried(t *testing.T) {
	client, s := newStubClient(t, []step{fail(503, "UNAVAILABLE", "Down")})
	recordSleeps(client)
	if _, err := client.Files.Upload(context.Background(), []UploadFile{{FileName: "a.pdf", Content: strings.NewReader("x")}}); err == nil {
		t.Fatal("no error")
	}
	if s.count() != 1 {
		t.Errorf("%d requests", s.count())
	}
}

// Zero values of optional fields are not sent; a pointer field set to String("") is sent and clears the value.
func TestOptionalFields(t *testing.T) {
	client, s := newStubClient(t, []step{okStatus(202, acceptedEmail), ok(map[string]any{"id": 1}), ok(map[string]any{"id": 1})})
	ctx := context.Background()
	_, err := client.Emails.Send(ctx, &SendEmailParams{
		From: Addr("no-reply@shop.ru"), To: []Address{Addr("a@example.ru")}, Subject: "Hi", HTML: "<p>Hi</p>",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertJSON(t, s.request(t, 0).data(t), `{
		"from": {"email": "no-reply@shop.ru"}, "to": [{"email": "a@example.ru"}], "subject": "Hi", "html": "<p>Hi</p>"
	}`)

	if _, err := client.Contacts.Update(ctx, 1, &UpdateContactParams{Phone: String("")}); err != nil {
		t.Fatal(err)
	}
	assertJSON(t, s.request(t, 1).data(t), `{"phone": ""}`)

	if _, err := client.Campaigns.List(ctx, &ListCampaignsParams{WithStats: true}); err == nil {
		t.Fatal("the stub answers without items, want an error")
	}
	if got := s.request(t, 2).URL.RawQuery; got != "withStats=true" {
		t.Errorf("query %s", got)
	}
}
