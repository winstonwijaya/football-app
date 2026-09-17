# football-app

Backend API for managing football teams under a parent company: teams, players,
match scheduling, and match results. Built with Go, Gin, GORM, and PostgreSQL.
JSON responses, consumed by an Android client.

---

## Quick start (docker-compose)

One command brings up Postgres, applies migrations, seeds an admin user, and
starts the API:

```bash
docker compose up --build
```

That's it — the API is now listening on `http://localhost:8080`, with a
working login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin12345"}'
```

Default seeded credentials (`admin` / `admin12345`) come from `.env.example`'s
`SEED_USERNAME`/`SEED_PASSWORD` — copy that file to `.env` and change them for
anything beyond local dev, same as `JWT_SECRET`.

Tear down: `docker compose down` (add `-v` to also drop the Postgres volume).

### What docker-compose actually runs

- `db` — Postgres 16
- `migrate` — applies all migrations via the official `migrate/migrate` image, then exits
- `seed` — creates the admin user if it doesn't already exist, then exits (idempotent — safe to `docker compose up` again)
- `app` — the API itself

---

## Running locally without Docker

Requires Go 1.25+ and a reachable Postgres instance.

```bash
# 1. copy and adjust env vars (see .env.example for every variable)
cp .env.example .env

# 2. the app reads real environment variables, not .env directly — source it:
set -a; source .env; set +a

# 3. apply migrations (uses the dockerized migrate/migrate image — no local
#    `migrate` CLI install needed; only Docker is required for this step)
make migrate-up

# 4. seed the first user (there is no registration endpoint — see Assumptions)
make seed SEED_USERNAME=admin SEED_PASSWORD=changeme

# 5. run the API
make run
```

## Makefile targets

```
make run            # go run ./cmd/api
make build          # builds ./bin/api and ./bin/seed
make test           # go test ./... -v
make fmt / vet / tidy
make migrate-up     # apply all migrations (dockerized migrate/migrate)
make migrate-down   # roll back all migrations
make seed           # provision the first user
make docker-up      # docker compose up --build -d
make docker-down    # docker compose down
make docker-logs    # docker compose logs -f
```

---

## API collection

`postman_collection.json` — import into Postman. Run **Auth > Login** first;
its test script saves the returned token into the `access_token` collection
variable automatically, so every other request just works via Bearer auth.

Folders mirror the resources 1:1 (Teams, Players, Matches, Reports), each
self-contained with its own List/Create/Get/Update/Delete in the order the
endpoints are listed in the spec — this is a reference/exploration
collection, not a scripted demo. Requests assume IDs from a fresh database
(team/player/match = 1, a second team = 2), and running an entire folder or
the whole collection straight through isn't guaranteed to work cleanly —
e.g. `Teams > Delete Team` will break a later `Players > Create Player` that
references it. Run what you need, substituting IDs from earlier responses
as you go. The one exception: `Matches` is ordered so `Report Match Result`
runs before `Delete Match`, so running that folder top-to-bottom
deliberately demonstrates the immutability rule — deleting a `Played` match
correctly returns `409` instead of a confusing `404`.

---

## Testing

```bash
make test
```

Unit tests use in-memory fakes for the repository interfaces — no database
required. Coverage focuses on the business-logic-heavy paths per the brief's
ask: result reporting (reconciliation, own-goal crediting, already-reported
conflict) and report calculation (outcome derivation, top scorer, cumulative
wins), plus auth and the request-ID/auth middleware.

---

## Security

This is an internal admin tool, not a public-facing API — there is
deliberately no self-service registration. Given that framing:

- **Auth**: JWT (HS256), issued by `POST /auth/login`, required on every
  other endpoint. Token TTL configurable via `JWT_EXPIRY_MINUTES` (default 60
  min).
- **Passwords**: bcrypt-hashed, never stored or logged in plaintext.
- **No registration endpoint**: users are provisioned directly via
  `cmd/seed`, matching the single-user-role assumption below — there's
  nothing to self-serve into.
- **Rate limiting**: a general per-IP limiter (10 req/s, burst 20) on the
  whole API, plus a stricter one specifically on `/auth/login` (~5
  attempts/min, burst 5) as a brute-force guard. In-memory, per-process —
  would need a shared store (e.g. Redis) behind multiple replicas.
- **Error responses never leak internals**: unexpected errors (DB failures,
  etc.) return a generic `"internal server error"` to the client; the real
  cause is only logged server-side, tagged with a request ID for
  correlation.
- **Audit trail**: every mutation records `created_by`/`updated_by`/
  `deleted_by` from the authenticated user's JWT — never client-supplied.
- **`JWT_SECRET`**: falls back to an insecure default for local dev only,
  with a loud startup warning. Must be set explicitly (and kept out of git)
  for anything beyond local use.
- **`DB_SSLMODE=disable`** is fine for the Docker-internal network used
  here; a deployment reaching a real managed Postgres instance over the
  network should use `require` or `verify-full` instead.
- **RBAC-ready, not RBAC-implemented**: single role today (see Assumptions)
  — the schema and JWT claims are structured so a `role` claim could be
  added without a breaking schema change.

---

## Assumptions

The case study left some things unspecified; these are the assumptions made
to fill those gaps, plus the reasoning behind judgment calls made while
implementing.

**Roles & auth**
- Single user role assumed. If a future requirement adds RBAC, the logic
  would be updated then — the JWT/middleware structure doesn't need to
  change, just what's checked.
- No registration endpoint by design — an internal tool's users are
  provisioned by an admin/ops process, not self-service. The first user is
  created via `cmd/seed`, a small idempotent CLI (see Quick start).

**Teams**
- Team logo is assumed to be a URL/public URL pointing at an image hosted
  elsewhere (`logo_url`) — no file upload handling.

**Players**
- Positions: the brief lists 4 in Indonesian (penyerang/gelandang/bertahan/
  penjaga gawang) — implemented as the English equivalents in Title Case
  (`Forward`/`Midfielder`/`Defender`/`Goalkeeper`), for consistency with the
  match status enum's casing style.
- There's no dedicated "player transfer" log. Moving a player to a different
  team is just a normal field update (`team_id`, and `squad_number` if it
  changes) via `PUT /players/:id` — full-replace semantics, same as any
  other field.
- `match_logs.team_id` is stored on the log row itself (not derived from the
  player's *current* team) specifically so a later transfer never corrupts
  a historical match's recorded result.

**Matches**
- `match_datetime` itself is timezone-aware (clients should send a full
  ISO-8601 timestamp with an explicit offset, e.g.
  `2026-10-01T06:00:00+07:00`) and is stored/compared as an absolute
  instant — no ambiguity there.
- The `from`/`to` filters on `GET /matches` (and the "one match per team per
  day" rule below) are different: they only take a bare `YYYY-MM-DD` date,
  which is inherently ambiguous without an agreed timezone — there's no
  universal frontend convention for what calendar day a bare date means.
  **This API assumes WIB (Asia/Jakarta, UTC+7)** for both: `from`/`to` cover
  that calendar day's full 00:00:00–23:59:59.999999999 WIB range, and the
  per-day conflict check groups matches by their WIB calendar day, not UTC.
  A match at 06:00 WIB is `23:00 UTC` the *previous* day — without this,
  both the list filter and the conflict rule would silently use the wrong
  day for any match outside UTC daytime hours.
- A team can play at most one match per calendar day (WIB, per above). A
  **cancelled** match doesn't count against this limit — a team can be
  rescheduled into a day where they previously had a match that got
  cancelled.
- Match date and time are stored in a single column (`match_datetime`), not
  split into separate date/time columns.
- `PUT /matches/:id` and result reporting are deliberately split: `PUT` can
  reschedule (team/datetime) and toggle status between `Scheduled` and
  `Cancelled` — it can never set scores or transition directly to `Played`.
  That only happens through `POST /matches/:id/result`.
- Once a match's result has been reported (`status = Played`), it becomes
  immutable: `PUT`/`DELETE` both return `409`. The score is now tied to
  specific `match_logs` rows, so editing the schedule out from under them
  isn't allowed.
- Match reports (`/reports/matches...`) only exist for `Played` matches — a
  scheduled, cancelled, or nonexistent match ID all return the same `404`.

**Match logs / result reporting**
- `match_logs` is a general match-event table by design — only the `GOAL`
  action is implemented for now, but the shape (action as free text, no
  rigid `CHECK` constraint) leaves room for cards, substitutions, etc. later
  without a migration.
- Log rows can only be added, never edited or deleted — append-only,
  matching the table's lack of update/delete audit columns.
- Result reporting is a **post-match recap**, not live/incremental scoring:
  the admin submits the final score and full goal list *after* the match
  already happened, in one payload, not goal-by-goal as it happens.
- In a goal entry, `team_id` is the side **credited** with the goal;
  `player_id` is the actual scorer. For a normal goal these are the same
  team; for an own goal they're opposite teams (the scorer belongs to the
  team that did *not* benefit).
- Submitted goals must reconcile exactly with the stated `home_score`/
  `away_score` (per team, excluding own goals from the scorer's own tally)
  or the whole report is rejected with `422` — no partial acceptance.
- Reporting a result twice on the same match returns `409`.

**Reports**
- `home_team_cumulative_wins`/`away_team_cumulative_wins` are each team's
  **overall** win total as of that match's date — not wins restricted to
  that specific home/away position. A team that won once at home and once
  away has a cumulative count of 2 either way; the field name only
  indicates *which team* (whoever is currently in that slot for *this*
  match), not that the count itself is position-specific. Verified against
  a scenario built specifically to distinguish the two readings: a team
  that won as home in an earlier match still carries that win into its
  cumulative total when it later appears in the away slot.
- Top scorer excludes own goals from a player's personal tally — an own
  goal counts toward the match score but isn't credited as that player
  "scoring."
- A 0-0 draw, or a match decided entirely by own goals, has no top scorer
  (`top_scorer` is omitted from the response, not a zero-goal entry).

**API design (beyond the literal endpoint list)**
- Pagination (`page`/`limit`) was extended to the Players, Matches, and
  Reports list endpoints for consistency, even though the brief's endpoint
  list only shows it explicitly for Teams.
- Every mutating endpoint requires authentication — there are no public
  write endpoints.

**Database**
- All mutable tables carry `created_at/by`, `updated_at/by`, `deleted_at/by`.
- Soft delete throughout: every delete is an `UPDATE deleted_at/deleted_by`,
  never a real `DELETE FROM`. Verified there's no direct GORM `.Delete()`
  call anywhere in the codebase that could bypass this.
- Squad numbers are unique within a team, enforced via a partial unique
  index that survives soft deletes (a re-added player can reuse a number a
  soft-deleted player held).
- `height`/`weight` are the actual DB column names (not `height_cm`/
  `weight_kg`) — units are documented via `COMMENT ON COLUMN` instead. The
  JSON API still exposes them as `height_cm`/`weight_kg` for clarity to
  consumers; this is a DB-naming choice, not an API contract change.

---

## Code layout

```
cmd/api/main.go     # composition root — wires everything by hand, no DI framework
cmd/seed/main.go     # provisions the first user (no registration endpoint)
internal/
  config/            # env config, DB connection
  handler/            # gin handlers: bind, call service, render
  service/            # business rules, transactions
  repository/         # SQL / GORM only
  model/              # GORM structs
  middleware/          # auth, request ID, rate limiting, error handling
  dto/                 # request/response structs
pkg/
  response/            # JSON envelope helpers
  apperror/             # typed service-layer errors -> HTTP status mapping
  hash/ token/          # bcrypt / JWT primitives
migrations/            # golang-migrate SQL files
```

Handlers never touch the database; services never touch `*gin.Context`. The
service layer returns a typed `*apperror.Error` carrying a code; one
middleware (`ErrorHandler`) maps that to an HTTP status and the JSON
envelope — no scattered `c.JSON(400, ...)` calls in handlers.
