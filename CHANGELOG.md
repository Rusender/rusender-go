# Changelog

**Русский** · [English](https://github.com/rusender/rusender-go/blob/main/CHANGELOG-EN.md)

## 0.1.0

Первая версия.

- Клиент `rusender.NewClient` для публичного API: `Emails`, `Contacts`, `Lists` (`Contacts`, `Variables`),
  `Campaigns`, `Folders`, `Templates`, `Senders`, `Domains`, `SendingKeys`, `Segments`, `Webhooks`, `Files`, `Me`.
- `context.Context` первым аргументом каждого метода; опции клиента и запроса — functional options.
- Ошибки `*rusender.Error` с `Code`, `Details`, `RequestID`; класс по HTTP-статусу через `errors.Is`
  (`ErrValidation`, `ErrNotFound`, `ErrRateLimit` и другие); `*NetworkError`, `ErrTimeout`.
- Повторы при 502–504 и сетевых ошибках; автоматический `Idempotency-Key`. 429 не повторяется —
  `ErrRateLimit` с `RetryAfter`.
- Загрузка вложений `Files.Upload` для рассылок.
- `WithResponseInto` у любого метода: `RequestID`, лимит запросов, `*http.Response` и сырой JSON ответа.
- Ответ вне контракта API (не JSON-конверт, список без `items`) — `ErrUnexpectedResponse`, а не пустые данные.
- Поля, которые API может вернуть как `null`, — указатели; моменты времени — `time.Time`.
- `Nullable` для полей, где `null` очищает значение.
- Необязательные поля параметров — обычные значения, указатели только там, где пустое значение осмысленно.
- Пробелы, перевод строки и кавычки вокруг API-ключа отбрасываются.
- Только стандартная библиотека, Go 1.21+.
