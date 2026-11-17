# RuSender Go SDK

**Русский** · [English](https://github.com/rusender/rusender-go/blob/main/README-EN.md)

Официальная библиотека для [публичного API RuSender](https://rusender.ru/docs/) на Go. Только стандартная
библиотека, Go 1.21 и новее.

> Разработка ведётся во внутреннем репозитории RuSender, GitHub — зеркало. Ошибки и пожелания — в Issues.

## Установка

```bash
go get github.com/rusender/rusender-go
```

## Быстрый старт

Создайте API-ключ в кабинете: **Интеграции → API**. Ключ передаётся в `NewClient`; если передать пустую
строку, он читается из `RUSENDER_API_KEY`. Пробелы, перевод строки и кавычки вокруг ключа отбрасываются.

```go
import "github.com/rusender/rusender-go"

client, err := rusender.NewClient(os.Getenv("RUSENDER_API_KEY"))
if err != nil {
	log.Fatal(err)
}

email, err := client.Emails.Send(ctx, &rusender.SendEmailParams{
	From:    rusender.Address{Email: "no-reply@shop.ru", Name: "Магазин"},
	To:      []rusender.Address{rusender.Addr("client@example.ru")},
	Subject: "Заказ оформлен",
	HTML:    "<p>Спасибо за заказ!</p>",
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(email.ID, email.Status) // 0199b2f4-…  queued
```

Клиент создайте один раз и переиспользуйте: он безопасен для одновременного использования из горутин.

Отправка по шаблону:

```go
_, err = client.Emails.Send(ctx, &rusender.SendEmailParams{
	From:          rusender.Addr("no-reply@shop.ru"),
	To:            []rusender.Address{{Email: "client@example.ru", Name: "Анна"}},
	Subject:       "Заказ №1234",
	TemplateAlias: "order-created",
	Variables:     map[string]any{"orderId": 1234, "total": "2 990 ₽"},
})
```

Вложения — байты; в base64 SDK кодирует сам:

```go
invoice, err := os.ReadFile("invoice.pdf")
// ...
_, err = client.Emails.Send(ctx, &rusender.SendEmailParams{
	From:        rusender.Addr("no-reply@shop.ru"),
	To:          []rusender.Address{rusender.Addr("client@example.ru")},
	Subject:     "Счёт",
	Text:        "Счёт во вложении",
	Attachments: []rusender.EmailAttachment{{FileName: "invoice.pdf", Content: invoice}},
})
```

Рассылка с вложениями: сначала загрузите файлы (до 20, всего до 5 МБ), затем укажите их `ID`:

```go
price, _ := os.Open("price.pdf")
defer price.Close()

files, err := client.Files.Upload(ctx, []rusender.UploadFile{{FileName: "price.pdf", Content: price}})
// ...
campaign, err := client.Campaigns.Create(ctx, &rusender.CreateCampaignParams{
	Name:        "Новый прайс",
	Subject:     "Прайс на ноябрь",
	TemplateID:  12,
	SenderID:    3,
	ListIDs:     []int{7},
	Attachments: []int{files[0].ID},
})
```

Статус письма и события по получателям:

```go
sent, err := client.Emails.Get(ctx, email.ID)
// ...
for _, recipient := range sent.Recipients {
	fmt.Println(recipient.Email, recipient.Status, recipient.Reason)
}
```

## Ресурсы

| Ресурс | Что умеет |
|---|---|
| `Emails` | отправка, список писем и доставок, статус письма, события получателя |
| `Contacts` | поиск, изменение, отписка и возврат подписки, статистика, удаление, выгрузка, история отправок |
| `Lists` | списки контактов; `Lists.Contacts` — добавление, перенос, импорт, выгрузка; `Lists.Variables` — переменные |
| `Campaigns` | рассылки: черновики, запуск, пауза и продолжение, A/B-тест, отправка частями, статистика, получатели |
| `Folders` | папки рассылок |
| `Templates` | шаблоны писем |
| `Senders`, `Domains` | адреса и домены отправителей, проверка DNS |
| `SendingKeys` | ключи отправки транзакционных писем и их статистика |
| `Segments` | сегменты контактов и условия |
| `Webhooks` | подписки на события: доставки, открытия, отписки, статусы рассылок; тестовое событие |
| `Files` | загрузка вложений для рассылок, список файлов, скачивание |
| `Me` | аккаунт и права текущего ключа |

Цепочки в SDK пока не входят.

## Параметры

Каждый метод принимает `context.Context` первым аргументом. Параметры — структура по указателю; `nil` —
без параметров.

Необязательные поля — обычные значения: пустая строка, `0` и `false` не отправляются. Указатель — только там,
где пустое значение осмысленно и его нужно уметь отправить: имя и телефон контакта или текстовая версия шаблона
в `Update` (`Phone: rusender.String("")` очищает телефон), даты (`*time.Time`). Для указателей есть хелперы
`rusender.String`, `rusender.Int`, `rusender.Bool`, `rusender.Time` и `rusender.Ptr`. Если API принимает `null`,
чтобы очистить значение, поле имеет тип `rusender.Nullable`:

```go
_, err = client.Templates.Update(ctx, 12, &rusender.UpdateTemplateParams{
	Name:  "Заказ оформлен",
	Alias: rusender.Null[string](), // удалить алиас; rusender.Value("order") — задать
})
```

В `Update` рассылки и сегмента пустой, но не `nil` срез (`ListIDs: []int{}`) отправляется как `[]` и очищает
значение; `nil` оставляет его как есть.

Перечисления — строковые типы с константами (`rusender.CampaignDraft`, `rusender.ContactActive`). Тип открыт:
новое значение из API не ломает разбор ответа.

## Ошибки

Любой ответ API не 2xx — ошибка `*rusender.Error` с полями `StatusCode`, `Code`, `Message`, `Details`,
`RequestID` и `RetryAfter`. Указывайте `RequestID` в обращениях в поддержку. Класс ошибки проверяется через
`errors.Is`:

```go
_, err := client.Emails.Send(ctx, params)
var apiErr *rusender.Error
switch {
case errors.Is(err, rusender.ErrValidation) && errors.As(err, &apiErr):
	log.Println(apiErr.Code, apiErr.Details) // [{Field: to, Reason: …}]
case errors.Is(err, rusender.ErrRateLimit) && errors.As(err, &apiErr):
	log.Printf("повторите через %s", apiErr.RetryAfter)
case err != nil:
	return err
}
```

| Значение для `errors.Is` | Когда |
|---|---|
| `ErrValidation` | 400, 422 — запрос не прошёл проверку, поля в `Details` |
| `ErrAuthentication` | 401 — ключ неверен или отозван |
| `ErrPaymentRequired` | 402 — не хватает средств или лимита тарифа |
| `ErrPermission` | 403 — у ключа нет нужного scope |
| `ErrNotFound` | 404 |
| `ErrConflict` | 409 — конфликт или `Idempotency-Key` повторён с другим телом |
| `ErrRateLimit` | 429, `RetryAfter` — сколько ждать до повтора |
| `*NetworkError` (`errors.As`) | запрос не дошёл или ответ не получен |
| `ErrTimeout` | ответ не пришёл за таймаут |
| `ErrUnexpectedResponse` | ответ вне контракта API: HTML-страница прокси с кодом 200, список без `items` |

Отмена контекста возвращает `ctx.Err()`: `context.Canceled` или `context.DeadlineExceeded`.

## Ответ целиком

Методы возвращают данные. Чтобы получить ещё `RequestID`, лимит запросов и сам `*http.Response`, передайте
`rusender.WithResponseInto`:

```go
var info rusender.ResponseInfo
list, err := client.Lists.Get(ctx, 7, rusender.WithResponseInto(&info))
// ...
fmt.Println(info.RequestID, info.RateLimit) // &{Limit:120 Remaining:119 ResetAt:…}
// подождать до сброса окна: time.Until(info.RateLimit.ResetAt)
```

`info.Body` — сырой JSON ответа, в том числе поля, которых SDK ещё не знает.

## Даты

Моменты времени в ответах — `CreatedAt`, `UpdatedAt`, `ScheduledAt`, `SentAt` и другие — имеют тип
`time.Time` (`*time.Time`, если значения может не быть). Календарные дни без времени (`Date` в активности
рассылки, `DateFrom`/`DateTo` в статистике ключа) остаются строками. В параметрах даты передаются как
`time.Time` и уходят в API в формате ISO 8601:

```go
start := time.Date(2026, 12, 1, 9, 0, 0, 0, time.FixedZone("MSK", 3*60*60))
_, err = client.Campaigns.Schedule(ctx, campaignID, &rusender.ScheduleCampaignParams{ScheduledStartDate: &start})
```

## Повторы и идемпотентность

SDK сам повторяет запрос при 502, 503, 504 и сетевых ошибках — по умолчанию 2 раза с экспоненциальной
задержкой. POST, PATCH и DELETE повторяются безопасно: SDK отправляет `Idempotency-Key`, и сервер не выполнит
запрос дважды. Свой ключ — например, номер заказа — передайте опцией:

```go
_, err = client.Emails.Send(ctx, params, rusender.WithIdempotencyKey(fmt.Sprintf("order-%d", order.ID)))
```

При превышении лимита запросов (429) SDK не ждёт и не повторяет запрос сам, а сразу возвращает ошибку
`ErrRateLimit`. Через сколько можно повторить — в `apiErr.RetryAfter`: подождите это время или поставьте задачу
в очередь. При повторе передайте тот же ключ идемпотентности.

## Пагинация

Списки возвращают одну страницу:

```go
page, err := client.Contacts.List(ctx, &rusender.ListContactsParams{
	PageParams: rusender.PageParams{Page: 1, Limit: 100},
	Status:     []rusender.ContactStatus{rusender.ContactActive},
})
// ...
fmt.Println(len(page.Items), page.Pagination.TotalPages)
```

Письма и доставки листаются курсором:

```go
params := &rusender.ListEmailsParams{Limit: 100}
for {
	page, err := client.Emails.List(ctx, params)
	if err != nil {
		return err
	}
	// page.Items
	if page.NextCursor == nil {
		break
	}
	params.Cursor = *page.NextCursor
}
```

## Выгрузки и файлы

`Contacts.Export`, `Lists.Contacts.Export` и `Files.Download` возвращают `*http.Response`: читайте или
копируйте `Body` и закройте его. Таймаут ограничивает ожидание заголовков ответа, чтение тела — только контекст.

```go
resp, err := client.Contacts.Export(ctx, &rusender.ExportContactsParams{Format: rusender.ExportCSV})
if err != nil {
	return err
}
defer resp.Body.Close()
_, err = io.Copy(file, resp.Body)
```

## Настройки клиента

```go
client, err := rusender.NewClient(apiKey,
	rusender.WithBaseURL("https://api.rusender.ru/api"),
	rusender.WithTimeout(30*time.Second), // на одну попытку
	rusender.WithMaxRetries(2),
	rusender.WithHeader("X-Team", "billing"),
	rusender.WithHTTPClient(&http.Client{Transport: myTransport}),
)
```

Опции отдельного запроса — последние аргументы любого метода: `WithIdempotencyKey`, `WithRequestID`
(уходит в `X-Request-ID`), `WithRequestTimeout`, `WithRequestMaxRetries`, `WithRequestHeader`,
`WithResponseInto`.

Полное руководство — в [документации](https://rusender.ru/docs/), справочник по типам — на
[pkg.go.dev](https://pkg.go.dev/github.com/rusender/rusender-go).

## Лицензия

MIT
