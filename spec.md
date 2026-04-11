# plain-cli Specification

A command-line interface for the Plain Android companion app, written in Go.

---

## Design Decisions

| Decision | Choice |
|---|---|
| Output/styling | charm.sh suite (lipgloss, bubbletea, bubbles) + `--output json\|table\|plain` everywhere |
| Config location | XDG-compliant (`~/.config/plain/config.toml`, respects `$XDG_CONFIG_HOME`) |
| Config format | TOML |
| Pagination | Fetch all by default; `--limit` and `--offset` optional |
| DataType enum | Lowercase CLI values (`image`, `feed-entry`, etc.) mapped to GraphQL uppercase |
| WebSocket | `github.com/coder/websocket` |

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/alecthomas/kong` | Command-line parsing |
| `github.com/charmbracelet/lipgloss` | Table styling and colored output |
| `github.com/charmbracelet/bubbletea` | Interactive TUI (prompts, pagers) |
| `github.com/charmbracelet/bubbles` | TUI components (spinner, text input, etc.) |
| `github.com/coder/websocket` | WebSocket transport (auth + event stream) |
| `github.com/BurntSushi/toml` | TOML config file parsing |
| `golang.org/x/crypto` | XChaCha20-Poly1305 encryption |

## Tooling

| Tool | Purpose |
|---|---|
| `gofumpt` | Strict Go formatter; all code must be `gofumpt`-clean |
| `golangci-lint` | Linter suite |

---

## Global Flags

Defined on the root Kong struct, available to every subcommand:

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--host` | `PLAIN_HOST` | (from config) | Plain server base URL |
| `--token` | `PLAIN_TOKEN` | (from config) | Auth token (base64) |
| `--client-id` | `PLAIN_CLIENT_ID` | (from config, auto-generated) | Stable client UUID |
| `--output` | `PLAIN_OUTPUT` | `table` | Output format: `table`, `json`, or `plain` |

---

## Config File

Path: `$XDG_CONFIG_HOME/plain/config.toml` (falls back to `~/.config/plain/config.toml`)

```toml
host      = "https://192.168.1.5:8080"
token     = "<base64_auth_token>"
client_id = "<uuid>"
```

- `client_id` is auto-generated as a random UUID on first run and written to config automatically.
- `token` is written by `plain auth login` after a successful authentication handshake.
- Flags and environment variables override config file values.

---

## Output Modes

All list and get commands support `--output table|json|plain`.

- **`table`** (default): lipgloss-styled table with colored headers, rendered to the terminal.
- **`json`**: raw JSON array or object, suitable for piping to `jq`.
- **`plain`**: one record per line, fields as `key: value` pairs separated by blank lines. Suitable for scripting without a JSON parser.

The `output.Printer` interface is defined in `internal/output/output.go` and injected into every command via Kong's `Bind()`.

---

## Pagination

List commands fetch **all results by default** by auto-paginating internally (repeated requests until exhausted).

Optional flags available on every list command:

| Flag | Description |
|---|---|
| `--limit <n>` | Return at most n results (disables auto-pagination) |
| `--offset <n>` | Skip the first n results |

---

## DataType Enum Mapping

| CLI value | GraphQL value |
|---|---|
| `image` | `IMAGE` |
| `video` | `VIDEO` |
| `audio` | `AUDIO` |
| `note` | `NOTE` |
| `feed-entry` | `FEED_ENTRY` |
| `call` | `CALL` |
| `contact` | `CONTACT` |
| `message` | `MESSAGE` |
| `bookmark` | `BOOKMARK` |

---

## Command Tree

```
plain
├── auth
│   ├── login               # WebSocket auth handshake; sends empty password hash,
│   │                       #   waits for device confirmation (PENDING flow)
│   ├── login --password    # same flow but prompts for password instead
│   └── status              # POST /init to verify current token is valid
│
├── device
│   ├── info                # DeviceInfo query
│   ├── battery             # battery query
│   ├── app                 # app query (version, ports, permissions, etc.)
│   ├── peers               # peers query
│   ├── mounts              # mounts query (storage with free/used bytes)
│   └── relaunch            # relaunchApp mutation
│
├── files
│   ├── ls [--root <path>] [--query <q>] [--sort <by>] [--limit n] [--offset n]
│   ├── recent
│   ├── info <path>
│   ├── mkdir <path>
│   ├── write <path> [--content <text>] [--overwrite]   # reads stdin if no --content
│   ├── rename <path> <name>
│   ├── copy <src> <dst> [--overwrite]
│   ├── move <src> <dst> [--overwrite]
│   ├── delete <paths...>
│   ├── download <path> [--out <local-path>]            # GET /fs, streams to file or stdout
│   ├── upload <local-path> <remote-path>               # chunked upload + mergeChunks
│   └── favorites
│       ├── ls
│       ├── add <root-path> <full-path>
│       ├── remove <full-path>
│       └── alias <full-path> <alias>
│
├── images
│   ├── ls [--query <q>] [--sort <by>] [--limit n] [--offset n]
│   ├── buckets
│   ├── download <id> [--out <local-path>]
│   ├── trash <query>
│   ├── restore <query>
│   ├── delete <query>
│   └── search
│       ├── status
│       ├── enable
│       ├── disable
│       └── index [--force]
│
├── videos
│   ├── ls [--query <q>] [--sort <by>] [--limit n] [--offset n]
│   ├── buckets
│   ├── download <id> [--out <local-path>]
│   ├── trash <query>
│   ├── restore <query>
│   └── delete <query>
│
├── audio
│   ├── ls [--query <q>] [--sort <by>] [--limit n] [--offset n]
│   ├── play <path>
│   ├── mode <order|shuffle|repeat|repeat-one>
│   ├── trash <query>
│   ├── restore <query>
│   ├── delete <query>
│   └── playlist
│       ├── ls
│       ├── add <query>
│       ├── remove <path>
│       ├── clear
│       └── reorder <paths...>
│
├── sms
│   ├── ls [--query <q>] [--limit n] [--offset n]
│   ├── conversations [--query <q>] [--limit n] [--offset n]
│   ├── send <number> <body>
│   └── send-mms <number> <body> --thread-id <id> [--attachments <paths...>]
│
├── contacts
│   ├── ls [--query <q>] [--limit n] [--offset n]
│   ├── sources
│   └── delete <query>
│
├── calls
│   ├── ls [--query <q>] [--limit n] [--offset n]
│   ├── call <number>
│   └── delete <query>
│
├── notes
│   ├── ls [--query <q>] [--limit n] [--offset n]
│   ├── get <id>
│   ├── save [--id <id>] --title <title> [--content <text>]  # reads stdin if no --content;
│   │                                                         #   omit --id to create new
│   ├── trash <query>
│   ├── restore <query>
│   ├── delete <query>
│   └── export <query>
│
├── feeds
│   ├── ls
│   ├── add <url> [--fetch-content]
│   ├── update <id> --name <name> [--fetch-content]
│   ├── delete <id>
│   ├── sync [--id <id>]           # omit --id to sync all feeds
│   ├── import <opml-file>
│   ├── export                     # prints OPML XML to stdout
│   └── entries
│       ├── ls [--query <q>] [--limit n] [--offset n]
│       ├── get <id>
│       ├── delete <query>
│       └── save-to-notes <query>
│
├── packages
│   ├── ls [--query <q>] [--sort <by>] [--limit n] [--offset n]
│   ├── install <device-path>
│   └── uninstall <ids...>
│
├── notifications
│   ├── ls
│   ├── cancel <ids...>
│   └── reply <id> --action-index <n> --text <text>
│
├── bookmarks
│   ├── ls
│   ├── add <urls...> --group-id <id>
│   ├── update <id> [--url <url>] [--title <title>] [--group-id <id>]
│   │            [--pinned] [--sort-order <n>]
│   ├── delete <ids...>
│   └── groups
│       ├── ls
│       ├── create <name>
│       ├── update <id> [--name <name>] [--collapsed] [--sort-order <n>]
│       └── delete <id>
│
├── chat
│   ├── channels
│   │   ├── ls
│   │   ├── create <name>
│   │   ├── update <id> <name>
│   │   ├── delete <id>
│   │   ├── leave <id>
│   │   ├── invite
│   │   │   ├── accept <id>
│   │   │   └── decline <id>
│   │   └── members
│   │       ├── add <channel-id> <peer-id>
│   │       └── remove <channel-id> <peer-id>
│   ├── messages <channel-id>
│   ├── send <to-id> <content>
│   └── delete <id>
│
├── tags
│   ├── ls --type <datatype>
│   ├── create --type <datatype> --name <name>
│   ├── update <id> --name <name>
│   ├── delete <id>
│   ├── add --type <datatype> --tags <id,...> --query <query>
│   └── remove --type <datatype> --tags <id,...> --query <query>
│
├── screen
│   ├── status
│   ├── start [--audio]
│   ├── stop
│   └── quality <auto|hd|smooth>
│
├── pomodoro
│   ├── status
│   ├── settings
│   ├── start [--time-left <seconds>]
│   ├── stop
│   └── pause
│
└── clipboard
    └── set <text>
```

---

## File Layout

```
plain-cli/
├── main.go                   # ~10 lines: Kong.Parse(), nothing else
├── go.mod
├── go.sum
├── spec.md
├── api.md
│
├── cmd/                      # Kong command structs, one file per domain
│   ├── root.go               # CLI root struct; global flags; AfterApply() loads config,
│   │                         #   constructs client.Client and output.Printer, binds both
│   │                         #   into Kong for dependency injection
│   ├── auth.go
│   ├── device.go
│   ├── files.go
│   ├── images.go
│   ├── videos.go
│   ├── audio.go
│   ├── sms.go
│   ├── contacts.go
│   ├── calls.go
│   ├── notes.go
│   ├── feeds.go
│   ├── packages.go
│   ├── notifications.go
│   ├── bookmarks.go
│   ├── chat.go
│   ├── tags.go
│   ├── screen.go
│   ├── pomodoro.go
│   └── clipboard.go
│
└── internal/
    ├── client/
    │   ├── client.go         # Client struct; GraphQL() method; REST helpers (GET, POST)
    │   ├── crypto.go         # XChaCha20-Poly1305 encrypt/decrypt; session key and login
    │   │                     #   key derivation; replay-protection wrapping for GraphQL
    │   ├── auth.go           # POST /init check; WebSocket password auth handshake
    │   ├── events.go         # WebSocket event stream; typed event dispatch;
    │   │                     #   reconnect with exponential backoff (1s→5s)
    │   ├── upload.go         # chunked upload orchestration: split, parallel POST
    │   │                     #   /upload_chunk (up to 3 in flight), resume via
    │   │                     #   uploadedChunks query, mergeChunks mutation
    │   └── fileid.go         # build encrypted fileId for /fs and /proxyfs endpoints
    │
    ├── api/
    │   ├── types.go          # Go structs mirroring every GraphQL type
    │   ├── queries.go        # GraphQL query string constants
    │   └── mutations.go      # GraphQL mutation string constants
    │
    ├── config/
    │   └── config.go         # Config struct; Load() / Save(); XDG path resolution;
    │                         #   client ID auto-generation and persistence
    │
    └── output/
        ├── output.go         # Format enum (Table/JSON/Plain); Printer interface
        ├── table.go          # lipgloss-based table renderer
        ├── json.go           # encoding/json renderer (pretty-printed)
        └── plain.go          # key: value line renderer for scripting
```

---

## Architecture Notes

### Dependency injection via Kong

`cmd/root.go` defines an `AfterApply()` hook that runs after all flags are parsed. It:

1. Loads `~/.config/plain/config.toml` and merges flag/env overrides.
2. Auto-generates and persists `client_id` if absent.
3. Constructs a `*client.Client`.
4. Constructs an `output.Printer` based on `--output`.
5. Calls `kong.Bind()` with both, making them available as injected parameters to every subcommand's `Run(*client.Client, output.Printer)` method.

### Encryption

All GraphQL request/response bodies and WebSocket frames are encrypted with **XChaCha20-Poly1305**.

Wire format: `[24-byte nonce][ciphertext + 16-byte Poly1305 tag]`

Key sources:
- **Session key**: first 32 bytes of `base64_decode(auth_token)`
- **Login key**: first 32 ASCII bytes of `sha512_hex(password)`

GraphQL payloads are replay-protected by wrapping before encryption:
```
"<unix_ms>|<16_char_hex_nonce>|<graphql_json>"
```

### File upload flow

1. Split local file into ~1 MB chunks.
2. Query `uploadedChunks(fileId)` to find already-received chunks (resume support).
3. POST remaining chunks to `/upload_chunk` with up to 3 parallel goroutines.
4. Call `mergeChunks` mutation to assemble on device.

### WebSocket event stream

- Connect to `ws[s]://<host>?cid=<client_id>` after authentication.
- Send one binary frame: `encrypt(session_key, unix_ms_timestamp_string)` to sync clocks.
- Each incoming frame: `[1-byte event type][ciphertext]` — decrypt to get JSON payload.
- On disconnect: reconnect with backoff starting at 1s, +1s per attempt, capped at 5s.
