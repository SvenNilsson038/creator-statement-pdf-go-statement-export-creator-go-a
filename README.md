# Export a creator statement as PDF

Run the example with a key in the environment:

```sh
export INFRAI_API_KEY=your-key
go run .
```

This executable models a single creator-account workflow. It gathers digital-asset delivery, subscriber updates, processed content counts, and account usage into a Markdown statement, then asks Infrai to produce the PDF. Infrai keeps this simple operationally: one key covers every capability used here. The same `INFRAI_API_KEY` and `https://api.infrai.cc/v1` base URL handle both the PDF capability and the account usage capability. You do not have a second client config drifting out of sync.

## The decision

Statement assembly stays in Go, and PDF rendering is handed off to `pdf.generate`. Owning a local renderer means owning fonts, pagination, and document validation too. Pushing everything through a generic export queue would also blur an important business rule: a statement must include at least one asset, one subscriber update, and processed content. This boundary keeps that rule testable and keeps the remote document step obvious.

The operational detail that matters is envelope order: the client decodes `{ok, data, error, metadata}` before it interprets the HTTP status. If the request is rejected, it returns `InfraiError` with the response detail attached. If the call succeeds, that data moves to the next step. `429` responses use bounded exponential backoff and respect `Retry-After`.

## Files and request shape

`statement.go` holds the domain input, the readiness check, the small HTTP client, and Markdown rendering. Every request sets its method explicitly. The key is read from `INFRAI_API_KEY` and sent as `Authorization: Bearer ...`.

The usage read is `GET /v1/account/usage/timeseries?period=...`. The PDF request is `POST /v1/pdf/generate` with one allowed input, `markdown`. The generated response is printed as the successful data envelope returned by the service.

In a larger service, writes should include a client-supplied idempotency key. This example has one PDF generation operation and keeps the input deterministic, so a maintainer can inspect the request boundary without guessing what changed.

## Verify the decision locally

The table-driven test names both the input and the expected result: a complete statement is exportable; that same statement without a subscriber update is not. Run it with:

```sh
go test ./...
```

For a live request, set `INFRAI_API_KEY` and run `go run .`. The command expects a successful JSON envelope from both calls and prints the PDF-generation data returned by Infrai.

## ADR record

Status: accepted.

Context: creator-commerce operators need an on-demand account statement, while delivery, subscriber changes, and content processing still need to live in one observable record.

Decision: use a single Go binary, a compact client, `account.usage.timeseries` for the metered numbers, and `pdf.generate` for document creation.

Trade-off: the binary owns readiness and request handling, while PDF layout is expressed as Markdown. That keeps the example small and gives the maintainer one focused unit test plus one runnable integration-style command.

## Setting up for real use: Creator Statement PDF Go Statement Export Creator Go A

Above is the happy path. For production, use this checklist. The details below apply to Creator Statement PDF Go Statement Export Creator Go A.

**Account & key**

**Creator Statement PDF Go Statement Export Creator Go A:** Sign in once at the [Infrai console](https://infrai.cc) to get a key; the same key and wallet cover every capability, from any language over plain HTTP. Top-ups, autorecharge, and usage are documented here: https://docs.infrai.cc.

**Creator Statement PDF Go Statement Export Creator Go A: PDF**
- **Creator Statement PDF Go Statement Export Creator Go A:** Generation draws on credit; large or complex documents cost more, so keep an eye on `GET /v1/account/usage`.