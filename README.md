# Autodidact

Autodidact is an informal, distraction-free **"Things to Learn"** queue engine built for self-taught learners. It organizes chaotic learning materials (books, papers, documentation, talks, and experiment ideas) into ordered sections so you can conquer them one at a time.

## Features

- **Compact Dashboard:**
  - Clean, high-density list of all learning topics and active queues with progress stats.
  - Direct duplication and deletion controls per list.
  - Overall progress bar and quick overview of active What Next targets.
- **What Next Priority Queue:**
  - Dedicated priority queue for the topics, books, or articles you want to tackle next.
  - Native **Drag & Drop** handle (`⠿`) to reorder priorities with persistent atomic ordering.
  - Live state badges (`🎯 In Next`) and quick toggle buttons (`+ What Next` / `✓ In What Next`) across item lists and detail workspaces.
  - Pick drawer with instant search across all study queues to pull items into What Next.
  - Finished log for timestamped completion history and reading reflections.
- **Unlimited Resources & Links:**
  - Attach unlimited reference links, papers, and book URLs to any learning item.
  - Clean, readable link formatting (`Link: https://... ↗`) in item views and the What Next queue.
  - Inline quick-add link input to attach additional links directly in the workspace.
  - Dynamic link management in edit modals with single-click removal (`✕`).
- **Direct Item Workspace:**
  - Status switcher pills (`To Do`, `Doing`, `Done`, `Skipped`).
  - Personal markdown notes and reflections scratchpad with raw edit and live preview modes, saved directly to SQLite.
  - Sequential bottom navigation bar (`Previous`, `Skip`, `Mark Done & Next`).
- **Continuous Quick Adding:**
  - Add books, articles, papers, or videos back-to-back: hitting <kbd>Enter</kbd> automatically retains cursor focus on the input so you can keep typing without mouse clicks.
  - Automatic URL detection and clean titling with dash/strip sanitizer.
- **Local Access Security:**
  - Bound to `127.0.0.1` by default to prevent LAN/network exposure.
  - Optional `AUTH_TOKEN` protection with in-app unlock screen.

## Tech Stack

- **Frontend:** SvelteKit (Svelte 5 runes mode), Tailwind CSS, Lucide icons, Marked (Markdown).
- **Backend:** Go 1.23+ with Chi router.
- **Database:** SQLite3 via pure Go `modernc.org/sqlite` (CGO-free, WAL mode enabled).
- **Orchestration:** Docker Compose multi-stage builds.

## Quick Start

### Running with Docker Compose

```bash
docker compose up --build
```

- **Frontend:** [http://localhost:3000](http://localhost:3000)
- **Default Security Key:** `autodidact-local-key`

### Running Locally Without Docker

```bash
# Terminal 1 - Backend
cd backend
go run ./cmd/api

# Terminal 2 - Frontend
cd frontend
npm run dev
```

Visit [http://localhost:5173](http://localhost:5173).

## License

[MIT](LICENSE)
