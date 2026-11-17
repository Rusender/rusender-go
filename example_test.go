package rusender_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/rusender/rusender-go"
)

func Example() {
	client, err := rusender.NewClient(os.Getenv("RUSENDER_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	email, err := client.Emails.Send(context.Background(), &rusender.SendEmailParams{
		From:    rusender.Address{Email: "no-reply@shop.ru", Name: "Shop"},
		To:      []rusender.Address{rusender.Addr("client@example.ru")},
		Subject: "Order placed",
		HTML:    "<p>Thank you for your order!</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(email.ID, email.Status)
}

func ExampleEmailsService_Send_template() {
	client, _ := rusender.NewClient("")
	ctx := context.Background()

	_, err := client.Emails.Send(ctx, &rusender.SendEmailParams{
		From:          rusender.Addr("no-reply@shop.ru"),
		To:            []rusender.Address{{Email: "client@example.ru", Name: "Anna"}},
		Subject:       "Order #1234",
		TemplateAlias: "order-created",
		Variables:     map[string]any{"orderId": 1234, "total": "2 990 ₽"},
	}, rusender.WithIdempotencyKey("order-1234"))
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleError() {
	client, _ := rusender.NewClient("")

	_, err := client.Contacts.Get(context.Background(), 42)
	var apiErr *rusender.Error
	switch {
	case errors.Is(err, rusender.ErrNotFound):
		fmt.Println("no such contact")
	case errors.Is(err, rusender.ErrRateLimit) && errors.As(err, &apiErr):
		fmt.Println("retry after", apiErr.RetryAfter)
	case errors.As(err, &apiErr):
		fmt.Println(apiErr.Code, apiErr.Details, apiErr.RequestID)
	case err != nil:
		log.Fatal(err)
	}
}

func ExampleWithResponseInto() {
	client, _ := rusender.NewClient("")

	var info rusender.ResponseInfo
	list, err := client.Lists.Get(context.Background(), 7, rusender.WithResponseInto(&info))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list.Name, info.RequestID)
	if info.RateLimit != nil {
		fmt.Println(info.RateLimit.Remaining, "requests left until", info.RateLimit.ResetAt)
	}
}

func ExampleEmailsService_List() {
	client, _ := rusender.NewClient("")
	ctx := context.Background()

	params := &rusender.ListEmailsParams{Limit: 100}
	for {
		page, err := client.Emails.List(ctx, params)
		if err != nil {
			log.Fatal(err)
		}
		for _, email := range page.Items {
			fmt.Println(email.ID, email.Status)
		}
		if page.NextCursor == nil {
			break
		}
		params.Cursor = *page.NextCursor
	}
}

func ExampleFilesService_Upload() {
	client, _ := rusender.NewClient("")
	ctx := context.Background()

	price, err := os.Open("price.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer price.Close()

	files, err := client.Files.Upload(ctx, []rusender.UploadFile{{FileName: "price.pdf", Content: price}})
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Campaigns.Create(ctx, &rusender.CreateCampaignParams{
		Name:        "New price list",
		Attachments: []int{files[0].ID},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleTemplatesService_Update_nullable() {
	client, _ := rusender.NewClient("")

	// Null clears the alias; a field left at its zero value is not sent and stays as is.
	_, err := client.Templates.Update(context.Background(), 12, &rusender.UpdateTemplateParams{
		Name:  "Order created",
		Alias: rusender.Null[string](),
	})
	if err != nil {
		log.Fatal(err)
	}
}
