# Namecheck — Full Project Audit

Audit date: 2026-07-09 · Commit audited: `29e50fe` · Toolchain used: Go 1.24.7, golangci-lint v2.5.0

## Verdict at a glance

The project builds cleanly, `go vet` is silent, all tests pass (including with `-race`), and the
supply-chain posture (gosec, CodeQL, Sonar, Dependabot, SBOMs, cosign signing) is well above
average for a project of this size. However, the HTTP server has a **goroutine leak on every
request**, the Tinder checker **queries the wrong URL**, two checkers have **dead validation
logic**, and the entire CI toolchain is pinned to **end-of-life Go versions**, which means
released binaries ship with unpatched stdlib CVEs.

Overall test coverage: **19.6 %** of statements.

---

## 1. Correctness bugs

### 1.1 [HIGH] Goroutine leak on every `/check` request — `cmd/server/main.go:140,185`

`sem` is created as a buffered channel but **nothing ever sends into it**; the only operation is
the deferred receive:

```go
sem := make(chan struct{}, maxOutstanding)   // line 140 — no send anywhere
...
defer func() { <-sem }()                     // line 185 — blocks forever
```

Since `defer wg.Done()` is registered after the semaphore release, it runs *first* (LIFO), so the
handler completes and responds normally — masking the fact that **all 6 checker goroutines per
request remain blocked forever on `<-sem>`**. Under load this is unbounded memory growth and
eventual OOM. Either implement the semaphore properly (`sem <- struct{}{}` before work, `<-sem`
on defer) or remove it.

### 1.2 [HIGH] Tinder checker queries the wrong username — `tinder/tinder.go:40`

```go
endpoint := fmt.Sprintf("https://tinder.com/@marta3%s", url.PathEscape(username))
```

A hardcoded `marta3` is concatenated in front of the username, so the checker reports the
availability of `@marta3<username>`, never the requested name. Every Tinder result the server
returns is wrong.

### 1.3 [HIGH] Dead validation: `legalPattern` never used — `instagram/instagram.go:28`, `tinder/tinder.go:28`

Both packages compile a `legalPattern` regexp that no code calls (confirmed by golangci-lint
`unused`). `IsValid` for Instagram/Tinder only checks length, so usernames with characters those
platforms forbid pass validation and produce meaningless availability probes.

### 1.4 [MEDIUM] `nil` cause stored in error on non-200 responses — `instagram/instagram.go:66-73`, `tinder/tinder.go:66-73`

In the `resp.StatusCode != http.StatusOK` branch, `err` is necessarily `nil` at that point, yet it
is stored as `Cause: err`. The real information (the status code) is discarded and the error
renders as `... : <nil>`. Store a synthesized error such as
`fmt.Errorf("unexpected status code %d", resp.StatusCode)`.

### 1.5 [MEDIUM] Each request runs every checker 3 times — `cmd/server/main.go:113`, `cmd/cli/main.go:40`

`for i := 0; i < 3; i++` appends three identical GitHub and Tinder checkers, so one `/check` call
fires 6 outbound HTTP requests (4 of them redundant) and triples the goroutine leak of §1.1. This
looks like leftover load-testing scaffolding from a workshop; the loop should be removed.

### 1.6 [MEDIUM] Server `WriteTimeout: 1s` vs. slow upstream checks — `cmd/server/main.go:57`

`/check` performs live HTTP calls to github.com/tinder.com before writing anything. If they take
longer than 1 s, the connection is torn down mid-response. There is also no timeout on the
outbound calls themselves (`http.DefaultClient` has none) — only the inbound `r.Context()`
bounds them. Raise `WriteTimeout` and/or use a client with an explicit `Timeout`.

### 1.7 [LOW] CLI abandons remaining results on first error — `cmd/cli/main.go:63-88`

The `select` loop `return`s on the first value from `errc`, discarding results from checkers that
succeeded and leaking their goroutines (unbuffered send to `results` after the reader is gone is
avoided only because both channels are buffered to `len(checkers)` — the goroutines finish, but
their answers are silently dropped).

### 1.8 [LOW] Redundant handler registration — `cmd/server/main.go:52`

`http.Handle("/", r)` registers the router on `http.DefaultServeMux`, but the server is started
with `Handler: r`, so the DefaultServeMux registration is dead code (and would become an
accidental exposure if anyone ever started the default mux, e.g. by re-enabling the commented
`net/http/pprof` import — see the warning comment on line 11).

---

## 2. Security

### 2.1 [MEDIUM] Unbounded, attacker-controlled map growth — `cmd/server/main.go:42,108-110`

`m[username]++` stores every distinct `username` query-param value forever. A public endpoint
that grows a map on attacker-chosen keys with no size cap, TTL, or length limit on the key is a
memory-exhaustion vector. Combined with §1.1, resource exhaustion is trivial. Cap the map or use
an LRU; also cap accepted username length before counting.

### 2.2 [MEDIUM] `/stats` and `/details` publicly leak all queried usernames — `cmd/server/main.go:70-85`

Every username anyone has ever checked is exposed to any caller, with counts. That is a privacy
leak of user search history. These endpoints should be authenticated, or removed.

### 2.3 [MEDIUM] All CI/release Go versions are end-of-life

| Where | Version | Status |
|---|---|---|
| `go.mod` | go 1.18 | EOL since 2023-02 |
| `go-test.yml` | 1.20.x | EOL since 2024-02 |
| `golangci-lint.yml` | 1.19 | EOL since 2023-09 |
| `goreleaser.yml` (release builds!) | 1.19 | EOL since 2023-09 |

Release binaries are compiled with Go 1.19 and therefore carry every `net/http`, `crypto/*` etc.
stdlib CVE fixed since then. Bump `go.mod` and all workflows to a supported toolchain (1.24/1.25)
— the code builds cleanly on 1.24 already (verified). Consider `go-version: 'stable'` in
workflows so Dependabot-style drift stops recurring.

*(Note: `govulncheck` could not reach vuln.go.dev from this sandbox; the assessment above is
based on toolchain age. The only module dependency, `gorilla/mux v1.8.1`, is the latest release
and has no known CVEs.)*

### 2.4 [LOW] No rate limiting on the server

`/check` triggers outbound requests to third parties on behalf of any caller — an amplification
primitive. Add per-IP rate limiting if this is ever exposed publicly.

### 2.5 [LOW] Dockerfile: unpinned base image — `Dockerfile:1`

`FROM alpine:latest` is unpinned (non-reproducible cert layer). Pin a version or digest. The
final `FROM scratch` stage is good; consider adding a numeric `USER 65534` so the process does
not run as uid 0.

### 2.6 Positives

CodeQL + gosec + Sonar scanning with scheduled runs, minimal workflow permissions, SBOM
generation and cosign image signing in goreleaser, a real `SECURITY.md`, and Dependabot for both
gomod and actions — this is a strong baseline. (The `SECURITY.md` supported-versions table lists
5.1.x/4.0.x, which doesn't match this repo's actual versioning — worth updating.)

---

## 3. Tests

### 3.1 [MEDIUM] Instagram/Tinder tests hit the live internet — `instagram/instagram_test.go:29-57`, `tinder/tinder_test.go`

`TestUsernameAvailable`/`TestUsernameNotAvailable` use `http.DefaultClient` against real
instagram.com/tinder.com. They are slow, flaky, fail offline, and depend on third-party HTML.
The repo already has a perfectly good `stub` package (used by the GitHub tests) — use it here.

### 3.2 [MEDIUM] Assertion logic masks failures in those same tests

```go
if got != want && err == nil { t.Errorf(...) }
```

If the live call errors (blocked, rate-limited, page moved), `err != nil` and the test silently
passes. These tests can essentially never fail in CI, which is why the brittle scraping
heuristics (§3.4) can rot undetected.

### 3.3 [LOW] Coverage is 19.6 %; `IsAvailable` paths mostly untested

`cmd/*` (the bulk of the logic, including the buggy handler in §1.1) has 0 % coverage. The
`handleCheck` flow is testable via `httptest` + the `stub` client.

### 3.4 [LOW] Scraping heuristics are brittle and likely already stale

Instagram availability = absence of `"noarchive, noimageindex"`; Tinder = absence of
`"</path></svg>Log in to like me</div>"`. Both break silently whenever the sites change markup
(Instagram now typically requires login, so results are probably already wrong). At minimum,
detect the "signal string absent for a known-taken account" case as an error rather than
reporting "available".

### 3.5 [LOW] Dead test code

The whole `twitter` package (source, tests, benchmarks, examples) is commented out. The
benchmark CI job (`go-test.yml`) runs `go test -bench=. .` against the root package, which has no
benchmarks — the job measures nothing. Either restore Twitter behind its Cloud-Function API or
delete the package; git history keeps it.

---

## 4. Code quality (golangci-lint v2.5.0: 3 issues; plus review findings)

- `cmd/server/main.go:72` — `fmt.Fprint` return value unchecked (errcheck).
- `instagram/instagram.go:28`, `tinder/tinder.go:28` — unused `legalPattern` (see §1.3).
- `instagram/instagram.go:75-83`, `tinder/tinder.go:75-83` — pointless `io.TeeReader` into a
  `bytes.Buffer` that is never read; just `io.ReadAll(resp.Body)`.
- `github/github.go:69` etc. — library packages print to **stdout** on body-close errors
  ("Error closing file" — it's not a file). Libraries shouldn't print; drop or surface the error.
- `namecheck.go:34` — typo "avaibility" → "availability" (user-visible error text);
  `cmd/cli/main.go:76` — "errrors".
- `cmd/server/main.go:41,92,102` — plain `uint64` + `atomic.AddUint64`; prefer `atomic.Uint64`
  (Go ≥1.19).
- Large blocks of commented-out code in `cmd/server/main.go`, `cmd/cli/main.go`, `twitter/` —
  delete; git remembers.
- `falser`/`truer` carry an unused `Client` field.

---

## 5. CI / release engineering

- **[HIGH] `docker-releaser.yml` pushes to a placeholder image** — `tags: user/app:latest`
  (line 38). It also builds from the raw repo context, but the `Dockerfile` expects a prebuilt
  `namecheck` binary (`COPY namecheck ...`), so the build fails anyway. This workflow is broken
  twice over; goreleaser already builds/pushes real images — delete or fix this workflow.
- **[MEDIUM] Lint pin mismatch** — `golangci-lint.yml` uses `golangci-lint-action@v9` with
  `version: v1.50.1` (2022). Action v9 targets golangci-lint v2; this combination is broken or
  silently stale. Repin to a current v2.x and fix the 3 findings above.
- **[MEDIUM] goreleaser configs use removed/deprecated fields** — `archives.replacements`
  (removed in modern goreleaser) and `brews.tap` (now `repository`) in both `.goreleaser*.yaml`;
  a current goreleaser will refuse the config.
- **[LOW] Benchmark job is a no-op** and on plain pushes `github.base_ref` is empty, making the
  "previous" checkout ill-defined; gate the job on `pull_request` only.
- **[LOW] Dependabot auto-merge never fires** — the enable-auto-merge step is gated on
  `contains(dependency-names, 'my-dependency')`, a template placeholder (the approve step *does*
  run). Intentional? If not, fix the condition; if yes, delete the dead step.
- README's Maintenance badge is hardcoded to `2023`.

---

## 6. Recommended fix order

1. Fix the semaphore goroutine leak (or drop `sem`) — `cmd/server/main.go` (§1.1).
2. Fix the Tinder URL (§1.2) and wire `legalPattern` into Instagram/Tinder `IsValid` (§1.3).
3. Remove the ×3 checker loops (§1.5).
4. Bump Go to a supported version in `go.mod` + all four workflows; repin golangci-lint (§2.3, §5).
5. Replace live-network tests with the `stub` client and fix the `&& err == nil` assertions (§3.1–3.2).
6. Cap/limit the visits map and gate `/stats`+`/details` (§2.1–2.2).
7. Fix or delete `docker-releaser.yml`; modernize goreleaser configs (§5).
8. Sweep the small stuff from §4.
