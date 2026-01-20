# AI Voice Agent - Backend

A production-ready Go backend for an AI-powered voice appointment scheduling assistant.

## 🚀 Features

- **Real-time Voice Communication** via WebSocket
- **Speech-to-Text** using Deepgram (Nova-2 model)
- **Text-to-Speech** using Cartesia
- **AI Conversations** powered by OpenAI GPT-4
- **Avatar Integration** with Tavus
- **Tool Calling System** for appointment management
- **Cost Tracking** per API call

## 📋 Prerequisites

- Go 1.22+
- API keys for:
  - [Deepgram](https://deepgram.com/) - STT (200 hours/month free)
  - [Cartesia](https://cartesia.ai/) - TTS
  - [OpenAI](https://platform.openai.com/) - LLM
  - [Tavus](https://www.tavus.io/) - Avatar (optional)
  - [Supabase](https://supabase.com/) - Database
  - [LiveKit](https://livekit.io/) - Real-time communication (optional)

## 🛠️ Setup

### 1. Clone and Install Dependencies

```bash
cd voice-agent-backend
go mod download
```

### 2. Configure Environment Variables

Copy the example env file and fill in your API keys:

```bash
cp .env.example .env
# Edit .env with your actual API keys
```

**Required Environment Variables:**

| Variable | Description | Get from |
|----------|-------------|----------|
| `DEEPGRAM_API_KEY` | Speech-to-Text API key | [Deepgram Console](https://console.deepgram.com/) |
| `CARTESIA_API_KEY` | Text-to-Speech API key | [Cartesia Dashboard](https://cartesia.ai/) |
| `LLM_API_KEY` | OpenAI API key | [OpenAI Platform](https://platform.openai.com/) |
| `SUPABASE_URL` | Supabase project URL | [Supabase Dashboard](https://supabase.com/) |
| `SUPABASE_API_KEY` | Supabase anon key | [Supabase Dashboard](https://supabase.com/) |

### 3. Set Up Database

Run the SQL migration in your Supabase SQL Editor:

```bash
# Copy contents of migrations/001_initial_schema.sql
# Paste into Supabase SQL Editor and run
```

### 4. Run the Server

```bash
# Development
make run
# or
go run ./cmd/server

# Production
make build
./bin/server
```

The server will start on `http://localhost:8080`

## 📡 API Endpoints

### REST API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/rooms` | Create a new room |
| GET | `/api/token` | Get access token for room |
| POST | `/api/avatar/session` | Create avatar session |
| GET | `/api/appointments` | Get user appointments |
| GET | `/api/slots` | Get available time slots |
| GET | `/api/summaries` | Get call summaries |

### WebSocket

Connect to `/ws` for real-time voice communication.

**Incoming Messages:**
- Binary: Audio data (16-bit PCM, 16kHz)
- `text_input`: Direct text input
- `end_call`: End the conversation
- `get_session`: Get current session state

**Outgoing Messages:**
- `connected`: Connection established
- `transcript`: STT result
- `agent_response`: AI response
- `tool_call`: Tool being executed
- `tool_result`: Tool result
- `call_summary`: Call summary at end
- Binary: TTS audio output

## 🧪 Testing

### Test Health Endpoint

```bash
curl http://localhost:8080/health
```

### Test WebSocket (using wscat)

```bash
npm install -g wscat
wscat -c ws://localhost:8080/ws
```

### Test with Frontend

1. Start the backend: `make run`
2. Start the frontend: `cd ../voice-agent-frontend && npm run dev`
3. Open `http://localhost:3000`
4. Click "Start Call"

## 🏗️ Project Structure

```
voice-agent-backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── agent/           # Voice agent orchestration
│   ├── config/          # Configuration management
│   ├── database/        # Supabase client
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── services/        # External service integrations
│   │   ├── avatar/      # Tavus integration
│   │   ├── cartesia/    # TTS service
│   │   ├── deepgram/    # STT service
│   │   ├── livekit/     # Real-time communication
│   │   └── llm/         # LLM service
│   ├── tools/           # Tool definitions and executor
│   └── websocket/       # WebSocket handler
├── migrations/          # Database migrations
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## 🐳 Docker

```bash
# Build image
docker build -t voice-agent-backend .

# Run with Docker Compose
docker-compose up -d
```

## 📝 Tool Calling System

The AI can call these tools during conversation:

| Tool | Description |
|------|-------------|
| `identify_user` | Identify user by phone number |
| `fetch_slots` | Get available appointment slots |
| `book_appointment` | Book a new appointment |
| `retrieve_appointments` | Get user's appointments |
| `cancel_appointment` | Cancel an appointment |
| `modify_appointment` | Modify appointment details |
| `end_conversation` | End the call |

## 📄 License

MIT
