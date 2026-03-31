# Saudi Meet Server

A self-hosted, AI-powered video conferencing platform — fork of [plugNmeet](https://github.com/mynaparrot/plugNmeet-server) built on [LiveKit](https://github.com/livekit/livekit-server) WebRTC infrastructure.

---

## Architecture

```
┌─────────────┐       ┌──────────────────┐       ┌───────────────┐
│   Browser   │◄─────►│ plugnmeet-client │◄─────►│  plugnmeet-api│
│  (React UI) │  WS   │  Vite / React    │ HTTP  │  Go / Fiber   │
└─────────────┘       └──────────────────┘       └───────┬───────┘
                                                         │
                      ┌──────────────────────────────────┼──────────────────┐
                      │                                  │                  │
               ┌──────▼──────┐  ┌────────┐   ┌───────────▼──┐  ┌───────────▼──┐
               │   LiveKit   │  │  NATS   │  │   MariaDB    │  │    Redis     │
               │  (WebRTC)   │  │JetStream│  │  (persistent)│  │   (cache)    │
               └──────┬──────┘  └────┬────┘  └──────────────┘  └──────────────┘
                      │              │
          ┌───────────┼──────┐       │
          │           │      │       │
   ┌──────▼─-─┐ ┌──────▼─┐ ┌─▼───┐ ┌─▼───────--───┐  ┌──────────┐
   │ Ingress  │ │  SIP   │ │TURN │ │  Recorder    │  │ Etherpad │
   │(RTMP/WHIP│ │Gateway │ │     │ │(Chrome+FFmpeg│  │(notepad) │
   └──────────┘ └────────┘ └─────┘ └─────────────-┘  └──────────┘
```

**Data flow:** Clients connect via WebSocket to the React UI, which communicates with the Go API server over HTTP. Real-time media flows through LiveKit. NATS JetStream handles inter-service messaging (recorder commands, auth callouts). MariaDB stores rooms, recordings, and user data. Redis provides caching and pub/sub for LiveKit.

---

## Repository Structure

```
saudi-meet-server/
├── docker-compose.yaml            # Unified dev compose (all services)
│
├── plugNmeet-server/              # Go API server (Fiber framework)
│   ├── main.go                    # Entrypoint — config → factory → router
│   ├── pkg/
│   │   ├── config/                # App configuration & defaults
│   │   ├── controllers/           # HTTP handlers (room, auth, recording, polls, etc.)
│   │   ├── models/                # Business logic layer
│   │   ├── services/              # External integrations
│   │   │   ├── db/                # MariaDB queries
│   │   │   ├── nats/              # NATS JetStream messaging
│   │   │   ├── redis/             # Redis caching
│   │   │   ├── livekit/           # LiveKit API calls
│   │   │   └── insights/          # AI services (transcription, translation, summarization)
│   │   ├── factory/               # Dependency injection (wire)
│   │   ├── dbmodels/              # Database models
│   │   ├── routers/               # Fiber route registration
│   │   └── turn/                  # TURN server providers (coturn, cloudflare)
│   ├── config_sample.yaml         # Reference configuration
│   ├── livekit_sample.yaml        # LiveKit server config
│   ├── nats_server_sample.conf    # NATS JetStream + auth callout config
│   ├── sql_dump/install.sql       # Database schema init
│   └── docker-build/              # Dockerfile (prod) + Dockerfile.dev (air hot-reload)
│
├── plugNmeet-client/              # React frontend (Vite + Tailwind)
│   ├── src/                       # Application source
│   ├── vite.config.ts             # Build config (port 3000)
│   ├── Dockerfile.dev             # Dev container (pnpm + Vite HMR)
│   └── package.json               # Dependencies (React 19, LiveKit client, Redux)
│
├── plugNmeet-recorder/            # Go recorder service
│   ├── main.go                    # Entrypoint
│   ├── pkg/                       # Recorder + transcoder logic
│   ├── config_sample.yaml         # Reference configuration
│   └── docker-build/              # Dockerfile (Chrome + FFmpeg + PulseAudio)
│
└── recording_files/               # Shared volume for recordings (server ↔ recorder)
```

---

## Getting Started

### Prerequisites

- **Docker** and **Docker Compose** v2+
- Ports available: `3000`, `4222`, `5060`, `6379`, `7880–7882`, `8080`, `8089`, `8222`, `9001`

### 1. Clone

```bash
git clone <repo-url> saudi-meet-server
cd saudi-meet-server
```

### 2. Generate Config Files

Copy all sample configs and adjust for Docker networking:

```bash
# Server configs
cp plugNmeet-server/config_sample.yaml    plugNmeet-server/config.yaml
cp plugNmeet-server/livekit_sample.yaml   plugNmeet-server/livekit.yaml
cp plugNmeet-server/nats_server_sample.conf plugNmeet-server/nats_server.conf
cp plugNmeet-server/ingress_sample.yaml   plugNmeet-server/ingress.yaml
cp plugNmeet-server/sip_sample.yaml       plugNmeet-server/sip.yaml

# Recorder config
cp plugNmeet-recorder/config_sample.yaml  plugNmeet-recorder/config.yaml

# Client config
cp plugNmeet-client/src/assets/config_sample.js plugNmeet-client/src/assets/config.js
```

Update the recorder config (`plugNmeet-recorder/config.yaml`):
- `plugNmeet_info.host` → `http://plugnmeet-api:8080`
- `plugNmeet_info.api_key` → `plugnmeet`
- `plugNmeet_info.api_secret` → value from `plugNmeet-server/config.yaml` → `client.secret`
- `nats_info.nats_urls` → `nats://nats:4222`
- `copy_to_path.main_path` → `/recording_files`

### 3. Start

```bash
docker compose up --build -d
```

### 4. Verify

```bash
docker compose ps
```

All 10 services should be running:

| Service | Port | Description |
|---|---|---|
| **plugnmeet-api** | `8080` | Go API server (air hot-reload) |
| **plugnmeet-client** | `3000` | React UI (Vite HMR) |
| **livekit** | `7880` | WebRTC SFU |
| **livekit-ingress** | `1935`, `8089` | RTMP/WHIP ingress |
| **livekit-sip** | `5060` | SIP gateway |
| **nats** | `4222`, `8222` | JetStream messaging + WebSocket |
| **redis** | `6379` | Cache |
| **db** | `3306` (internal) | MariaDB |
| **etherpad** | `9001` | Shared notepad |
| **recorder** | — | Chrome-based recording + transcoding |

> **Note:** The recorder may exit on first start if NATS isn't ready yet. Run `docker compose restart recorder` if needed.

---

## Configuration

### Server (`plugNmeet-server/config.yaml`)

| Section | Purpose |
|---|---|
| `client` | API port, api_key/secret, webhook, prometheus |
| `livekit_info` | LiveKit connection (host, api_key, secret) |
| `redis_info` | Redis host, auth, sentinel support |
| `database_info` | MariaDB connection, replicas |
| `nats_info` | NATS URLs, WebSocket URLs, JetStream subjects, auth keys |
| `insights` | AI providers (Azure Speech, Google Gemini) for transcription, translation, summarization |
| `shared_notepad` | Etherpad hosts |
| `recorder_info` | Recording files path, token validity |

### Recorder (`plugNmeet-recorder/config.yaml`)

| Section | Purpose |
|---|---|
| `recorder` | ID, mode (both/recorderOnly/transcoderOnly), resolution, limits |
| `plugNmeet_info` | Server connection (host, api_key, api_secret) |
| `nats_info` | NATS connection for job dispatch |
| `ffmpeg_settings` | Codec presets for recording, post-processing, RTMP |

### Client (`plugNmeet-client/src/assets/config.js`)

Sets `window.plugNmeetConfig` with server URL, codec preferences, resolution defaults, and UI customization.

---

## Production Notes

- **Do not use the dev compose in production.** Build production images using the non-dev Dockerfiles.
- **Secrets:** Rotate all default keys/secrets in `config.yaml`, `livekit.yaml`, and `nats_server.conf` before deploying.
- **TLS:** Place a reverse proxy (nginx, Caddy, Traefik) in front for HTTPS termination.
- **NATS WebSocket:** The `nats_ws_urls` in server config must be reachable by end-user browsers (public URL with TLS).
- **Recordings:** Both `plugnmeet-api` and `recorder` must share the same `recording_files` volume/path.
- **Scaling:** LiveKit, NATS, and Redis all support clustering. The server supports MariaDB read replicas.
- **AI Insights:** Requires Azure Speech SDK credentials (transcription/translation) and Google Gemini API key (AI chat/summarization). Configure in `config.yaml` → `insights`.

---

## API Responsibilities

The Go server exposes a REST API on port `8080`. Key controller groups:

| Controller | Endpoints | Purpose |
|---|---|---|
| `room` | Create, join, end, fetch rooms | Room lifecycle management |
| `recording` | Start, stop, list, download | MP4 recording & RTMP broadcast |
| `user` | Mute, remove, switch presenter | In-room user management |
| `polls` | Create, vote, close | Live polling |
| `analytics` | Fetch session analytics | Post-meeting engagement data |
| `artifact` | List, download, delete | AI-generated summaries & transcripts |
| `insights` | Transcription, translation, synthesis | Real-time AI features |
| `auth` | Token generation, validation | JWT-based room access |
| `bbb` | BigBlueButton-compatible endpoints | BBB API compatibility layer |
| `webhook` | LiveKit event relay | Server-to-server event delivery |
| `nats_auth` | Auth callout handler | NATS client authentication |

Full API documentation: [plugnmeet.org/docs/api](https://www.plugnmeet.org/docs/api/intro)

---

## Contributing

1. Fork this repository
2. Create a feature branch from `main`
3. Make changes — the dev compose provides hot-reload for both server (Go/air) and client (Vite HMR)
4. Run existing tests: `cd plugNmeet-server && go test ./...`
5. Submit a pull request

---

## License

MIT — see [LICENSE](plugNmeet-server/LICENSE)
