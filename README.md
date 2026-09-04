# daily-it-podcast

A personal daily IT-news podcast: generated automatically and listened to only by me. Not a commercial or public service.

## Shape

Generation and playback are separate systems. The only thing that connects them is a set of files on a personal Google Drive. Episodes are never generated through the UI.

```text
Generator (Go + GitHub Actions cron)
  fetch -> manuscript (Cursor Cloud Agents REST) -> speech (Gemini TTS) -> save
        |
  Personal Google Drive (audio + manuscript)
        |
Playback (Vite + TypeScript + React + Cloudflare)
  Access -> UI -> Workers (Hono, proxy for Drive reads)
```
[Runtime diagram](apps/diagrams/runtime.png)　

## Technology choices

| Role | Choice |
|------|--------|
| Playback UI | Vite + TypeScript + React + Pico.css (classless) |
| Behind the UI | Cloudflare Workers (Hono, Drive proxy) |
| Entry to the UI | Cloudflare Access (`DEPLOY.md`) |
| Generation | Go CLI + GitHub Actions cron |
| Fetch | Official APIs / RSS from several sources (HackerNews, Lobsters, ITmedia NEWS) |
| Storage | Personal Google Drive |
| Manuscript | Cursor Cloud Agents REST (Port `TextWriter`) |
| Speech | Google Gemini TTS |

## Repository

```text
apps/playback/contracts/ # web <-> worker HTTP
apps/playback/web/       # Vite UI
apps/playback/worker/    # BFF
apps/generator/          # Go CLI
apps/diagrams/           # runtime diagram (code-first)
contracts/               # representation on Drive (SSOT)
.github/workflows/
```

| What you want | Source of truth |
|------|------|
| Layers, dependencies, test layout | `DESIGN.md` |
| Deploy, Access, GHA operation, secret registration | `DEPLOY.md` |
| Drive file contracts | `contracts/` |
| Playback HTTP contracts | `apps/playback/contracts/` |
| Open-work index | `docs/tasks/todo/*-lane.md` |
| Recurring decisions | `docs/decisions/` |

## Branches

| Branch | Role |
|--------|------|
| `develop` | SSOT |
| `master` | release |

`feature/*` -> PR (base: `develop`) -> `master` is released by shim.

## Usage

1. **Playback:** enter through Access -> list -> play / show manuscript. Steps in `DEPLOY.md`.
2. **Generation:** GHA schedule / manual. Never started from the UI. Artifacts follow `contracts/`. Operation in `DEPLOY.md`.
3. **Playback local:** `cd apps/playback && npm ci && npm run dev` (Node version from `.nvmrc`).
4. **Generator:** Go module at `apps/generator/go.mod`. Put `golangci-lint` on PATH.
5. **Hooks:** `./scripts/install-hooks.sh`.
6. **Verification entry points:** `./scripts/check-static.sh` / `./scripts/test-unit.sh` / `./scripts/test-integration.sh` (details and thresholds in `DESIGN.md`; credentialed and scheduled E2E in `DEPLOY.md`).

## Constraints

- Private; no custom database; no multi-user.
- No unauthorized push to `master`. Production deploy policy is in `DEPLOY.md`.
