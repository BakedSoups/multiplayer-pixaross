# Nonogram Pictures

A small Go + Ebitengine vertical-slice prototype for a cozy, DS-era-inspired nonogram puzzle game. It loads puzzle JSON from `assets/`, generates clues from the solution, lets you draw with a right-side tool trigger, and reveals pixel art when solved.

<img width="1183" height="1376" alt="image" src="https://github.com/user-attachments/assets/2315be46-67d8-43d6-9ac9-fa36fb654cb8" />

## Run

```sh
go run ./cmd/game
```

Controls:

- Click or touch the right-side trigger to switch between fill and X-mark tools.
- Drag across cells to keep applying the selected tool.
- `F` selects fill, `X` or `M` selects X-mark, `Z` undoes, and `R` resets.

## Add A Puzzle

Create a folder in `levels/` named with the level number and title:

```text
levels/007-flower/
  art.png
```

The image can be named `art.png`, `art.webp`, `sheet.png`, or `sheet.webp`. If there is only one PNG or WebP in the folder, that file is used automatically.

The image must be a two-panel spritesheet: the left tile is the before/line art and puzzle solution, and the right tile is the colored reveal. A 10x10 level should be a 20x10 image, and a 15x15 level should be a 30x15 image. The generator infers the tile size from the image height.

The older flat-file format still works:

```text
levels/L3-Flower_16.png
```

In that format, the suffix is the tile size. A `_16` file should be 32x16. If the suffix is wrong but the image is still a two-panel sheet, the generator uses the image height as the tile size.

For opaque black-and-white line art, the generator treats the dominant white-ish or black-ish first-panel color as the empty background.

Generate puzzle JSON from every sheet in `levels/`:

```sh
go run ./cmd/genlevels
```

This writes self-contained `puzzle.json` files under `assets/puzzles/` and `internal/assets/embedded/assets/puzzles/`. No split skeleton/reveal images are generated.

## Community Editor

Open **Community > Create**. Guest drawings and packs are saved in browser local storage, and **My Art** keeps multiple editable drafts. Publishing asks for title, description, tags, and whether the creator wants the level inspected for possible inclusion in the main game. Main-game submission is optional, requires a rights confirmation, and does not guarantee approval.

### Import From Aseprite

For the most reliable multi-level import, export the sprite sheet as PNG and enable Aseprite's JSON data export. Name paired frames with the same base name:

```text
flower_before
flower_after
lion_before
lion_after
```

Select the PNG and JSON together from **Community > Create > Import Sprite Sheet**. Frames must be square and 8, 10, 15, or 20 pixels. The importer turns every matching `_before` / `_after` pair into a separate draft. Before art is normalized to black and determines the nonogram solution; After art keeps its colors.

For a regular PNG without JSON, arrange any number of pairs horizontally (`Before, After`) or vertically (`Before` above `After`). The image dimensions must be exact multiples of the selected tile size. A single 1x2 pair works through the same importer.

### Community Backend

The app remains fully usable for guests without a backend. To enable accounts and publishing:

1. Create a Supabase project and apply `supabase/migrations/001_community.sql`.
2. Set `SUPABASE_URL` and `SUPABASE_ANON_KEY` in your shell or in a local `.env` file. `SUPABASE_URL` should be the project URL (`https://...supabase.co`); copied Data API URLs ending in `/rest/v1` are normalized by the config script. You can copy `.env.example` as a starting point. Do not use a service-role key.
3. Generate the ignored browser config:

```sh
scripts/write-web-config.sh
```

`scripts/dev-web.sh` runs this automatically before every web build.
4. Add the game URL to Supabase Auth redirect URLs and enable email magic links.
5. Deploy `supabase/functions/notify-official-submission` and set `WEBHOOK_SECRET`, `RESEND_API_KEY`, `REVIEW_EMAIL`, `REVIEW_FROM_EMAIL`, and `ADMIN_REVIEW_URL`.
6. Create a Supabase Database Webhook for inserts on `official_submissions`, targeting that Edge Function with the matching `x-webhook-secret` header.
7. Set reviewer profiles to `moderator` or `admin` with a trusted SQL/admin operation. Review submissions at `/admin.html` after signing in through the game.

Published level versions are immutable. Packs reference exact versions, user data is protected by row-level security, and review emails contain an admin link rather than artwork attachments.

## Checks

Local checks:

```sh
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

GitHub Actions runs formatting, tests, vet, build, and `gocritic` duplicate-code checks on every push and pull request.

## Web And Mobile

Ebitengine can target WebAssembly, Android, and iOS later. A future web build can use:

```sh
GOOS=js GOARCH=wasm go build -o static/game.wasm ./cmd/game
```

For a local web dev loop, install `watchexec`, then run:

```sh
PORT=8000 ./scripts/dev-web.sh
```

<img width="813" height="1174" alt="image" src="https://github.com/user-attachments/assets/5a22b50c-164d-4e11-b366-419a6ed7ea4b" />

That serves `static/`, rebuilds `static/game.wasm` when Go/assets change, and reloads the browser on localhost after the WASM file changes.

Android and iOS should be added once the MVP input, layout, and asset loading choices are stable.
