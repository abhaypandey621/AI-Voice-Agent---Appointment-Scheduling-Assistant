# 🎙️ AI Voice Agent - Appointment Scheduling Assistant

A full-stack AI voice agent application that enables natural voice conversations for booking and managing appointments. Built with Go backend and React frontend, featuring real-time speech-to-text, text-to-speech, visual avatar integration, and intelligent appointment management.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Setup Instructions](#setup-instructions)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Frontend Features](#frontend-features)
- [Deployment](#deployment)
- [Development](#development)

## 🎯 Overview

This project is an AI-powered voice assistant designed for appointment scheduling. Users can interact with the agent through natural voice conversations to:

- **Identify themselves** by phone number
- **Book appointments** with automatic slot availability checking
- **Retrieve existing appointments**
- **Modify or cancel appointments**
- **Receive call summaries** at the end of conversations

The system uses advanced AI services for speech recognition, natural language understanding, speech synthesis, and visual avatar rendering.

## ✨ Features

### Core Functionality

1. **Voice Conversation**
   - Real-time speech-to-text using Deepgram
   - Natural text-to-speech using Cartesia
   - Context-aware conversations with LLM (OpenAI/Compatible)
   - <3 second response latency
   - Support for 5+ back-and-forth exchanges

2. **Visual Avatar Integration**
   - Tavus/Beyond Presence avatar support
   - Real-time avatar synchronization with speech
   - Smooth video streaming throughout conversation

3. **Intelligent Tool Calling**
   - Automatic intent recognition and tool selection
   - 7 specialized tools for appointment management:
     - `identify_user` - User identification by phone
     - `fetch_slots` - Available time slot checking
     - `book_appointment` - Appointment booking with conflict prevention
     - `retrieve_appointments` - Appointment history retrieval
     - `cancel_appointment` - Appointment cancellation
     - `modify_appointment` - Appointment modification
     - `end_conversation` - Graceful conversation termination

4. **Call Summary & Analytics**
   - Automatic summary generation at call end
   - Lists booked appointments
   - Captures user preferences
   - Cost breakdown with per-service pricing (bonus feature)

5. **Real-time UI Updates**
   - Live transcript display
   - Tool call visualization with status indicators
   - Avatar state synchronization
   - Cost tracking and display

## 🛠️ Tech Stack

### Backend

- **Language**: Go 1.24+
- **Web Framework**: Gin
- **WebSocket**: Gorilla WebSocket
- **Database**: Supabase (PostgreSQL)
- **Speech-to-Text**: Deepgram API
- **Text-to-Speech**: Cartesia API
- **LLM**: OpenAI API (or compatible)
- **Avatar**: Tavus API
- **Real-time**: WebSocket for bidirectional communication

### Frontend

- **Framework**: React 18.3+
- **Build Tool**: Vite 5
- **Language**: TypeScript 5.4+
- **State Management**: Zustand
- **Styling**: Tailwind CSS 3.4+
- **UI Components**: Lucide React icons
- **Date Handling**: date-fns

## 🏗️ Architecture

### System Architecture

```
┌─────────────────┐
│   React Frontend │
│   (Vite + TS)    │
└────────┬─────────┘
         │
         │ WebSocket / HTTP
         │
┌────────▼─────────────────────────────────────────┐
│              Go Backend Server                    │
│  ┌────────────────────────────────────────────┐  │
│  │  HTTP API (Gin Router)                     │  │
│  │  - REST endpoints                          │  │
│  │  - WebSocket handler                       │  │
│  └────────────────────────────────────────────┘  │
│                                                   │
│  ┌────────────────────────────────────────────┐  │
│  │  Voice Agent Service                       │  │
│  │  - Manages conversation state              │  │
│  │  - Coordinates STT/TTS/LLM                 │  │
│  │  - Handles tool execution                  │  │
│  └────────────────────────────────────────────┘  │
│                                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────┐│
│  │  Deepgram    │  │  Cartesia    │  │  OpenAI ││
│  │  (STT)       │  │  (TTS)       │  │  (LLM)  ││
│  └──────────────┘  └──────────────┘  └─────────┘│
│                                                   │
│  ┌────────────────────────────────────────────┐  │
│  │  Supabase (PostgreSQL)                     │  │
│  │  - Users, Appointments, Call Summaries     │  │
│  └────────────────────────────────────────────┘  │
│                                                   │
│  ┌────────────────────────────────────────────┐  │
│  │  Tavus API                                 │  │
│  │  - Avatar video streaming                  │  │
│  └────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────┘
```

### Data Flow

1. **Voice Input**:
   ```
   User Microphone → Frontend Audio Capture → WebSocket Binary → Backend → Deepgram API
   ```

2. **Processing**:
   ```
   Transcript → LLM Service → Tool Execution → Database Operations → LLM Response
   ```

3. **Voice Output**:
   ```
   LLM Response → Cartesia API → Audio Stream → WebSocket Binary → Frontend → Audio Playback
   ```

4. **Avatar**:
   ```
   Backend → Tavus API → Conversation URL → Frontend iframe → Avatar Video Stream
   ```

### Component Architecture

#### Backend Structure

```
voice-agent-backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── agent/          # Voice agent orchestration
│   ├── config/         # Configuration management
│   ├── database/       # Supabase client
│   ├── handlers/       # HTTP/WebSocket handlers
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   ├── services/       # External service integrations
│   │   ├── avatar/     # Tavus API client
│   │   ├── cartesia/   # Cartesia TTS client
│   │   ├── deepgram/   # Deepgram STT client
│   │   ├── llm/        # OpenAI LLM client
│   │   └── livekit/    # LiveKit integration
│   ├── tools/          # Tool definitions & executor
│   └── websocket/      # WebSocket connection manager
├── migrations/         # Database migrations
└── pkg/               # Public packages
```

#### Frontend Structure

```
voice-agent-frontend/
├── src/
│   ├── components/     # React components
│   │   ├── Avatar/     # Avatar display component
│   │   ├── Call/       # Call controls & status
│   │   ├── Chat/       # Chat panel & messages
│   │   ├── Summary/    # Call summary modal
│   │   └── ToolDisplay/ # Tool call visualization
│   ├── hooks/          # Custom React hooks
│   │   ├── useAudio.ts      # Audio capture & playback
│   │   └── useVoiceAgent.ts # Main voice agent hook
│   ├── services/       # API clients
│   │   ├── api.ts      # REST API client
│   │   └── websocket.ts # WebSocket client
│   ├── store/          # Zustand state management
│   ├── types/          # TypeScript type definitions
│   └── styles/         # Global styles
└── public/             # Static assets
```

## 📁 Project Structure

### Backend Key Files

- **`cmd/server/main.go`**: Application entry point, server setup
- **`internal/agent/agent.go`**: Voice agent state machine
- **`internal/websocket/handler.go`**: WebSocket connection handling
- **`internal/tools/executor.go`**: Tool execution logic
- **`internal/services/llm/llm.go`**: LLM integration with tool calling
- **`internal/database/supabase.go`**: Database operations

### Frontend Key Files

- **`src/App.tsx`**: Main application component
- **`src/hooks/useVoiceAgent.ts`**: Core voice agent integration
- **`src/components/Avatar/Avatar.tsx`**: Avatar display
- **`src/components/ToolDisplay/ToolDisplay.tsx`**: Tool call UI
- **`src/services/websocket.ts`**: WebSocket service

## 🚀 Setup Instructions

### Prerequisites

- Go 1.24+ installed
- Node.js 18+ and npm/yarn
- Supabase account and project
- API keys for:
  - Deepgram
  - Cartesia
  - OpenAI (or compatible LLM)
  - Tavus (optional, for avatar)

### Backend Setup

1. **Clone and navigate to backend**:
   ```bash
   cd voice-agent-backend
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up environment variables**:
   ```bash
   cp .env.example .env  # Create from template
   ```

4. **Configure `.env`** (see Configuration section)

5. **Run database migrations**:
   ```bash
   # Apply migrations to Supabase
   # Use Supabase dashboard or CLI
   ```

6. **Start the server**:
   ```bash
   go run cmd/server/main.go
   # Or
   make run
   ```

Server runs on `http://localhost:8080` by default.

### Frontend Setup

1. **Navigate to frontend**:
   ```bash
   cd voice-agent-frontend
   ```

2. **Install dependencies**:
   ```bash
   npm install
   # or
   yarn install
   ```

3. **Set up environment variables**:
   ```bash
   # Create .env file
   VITE_API_URL=http://localhost:8080
   VITE_WS_URL=ws://localhost:8080
   ```

4. **Start development server**:
   ```bash
   npm run dev
   # or
   yarn dev
   ```

Frontend runs on `http://localhost:3000` by default.

## ⚙️ Configuration

### Backend Environment Variables

Create a `.env` file in `voice-agent-backend/`:

```env
# Server
PORT=8080
ENVIRONMENT=development

# LiveKit (optional, for room management)
LIVEKIT_URL=
LIVEKIT_API_KEY=
LIVEKIT_API_SECRET=

# Deepgram (Speech-to-Text)
DEEPGRAM_API_KEY=your_deepgram_api_key

# Cartesia (Text-to-Speech)
CARTESIA_API_KEY=your_cartesia_api_key
CARTESIA_VOICE_ID=a0e99841-438c-4a64-b679-ae501e7d6091

# LLM (OpenAI or compatible)
LLM_PROVIDER=openai
LLM_API_KEY=your_openai_api_key
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4o

# Avatar (Tavus)
AVATAR_PROVIDER=tavus
AVATAR_API_KEY=your_tavus_api_key
AVATAR_ID=your_replica_id

# Supabase
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_API_KEY=your_supabase_anon_key

# Pricing (optional, for cost tracking)
DEEPGRAM_PRICE_PER_MIN=0.0043
CARTESIA_PRICE_PER_CHAR=0.000015
LLM_PRICE_PER_TOKEN=0.00003
```

### Frontend Environment Variables

Create a `.env` file in `voice-agent-frontend/`:

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

For production, update these to your deployed backend URLs.

## 📡 API Documentation

### REST Endpoints

#### Health Check
```http
GET /health
```
Returns server health status.

#### Create Room
```http
POST /api/rooms
Content-Type: application/json

{
  "room_name": "optional-room-name"
}
```

#### Get Token
```http
GET /api/token?room=room_name&participant=participant_name
```

#### Get Available Slots
```http
GET /api/slots?date=2024-01-15
```

#### Get Appointments
```http
GET /api/appointments?phone=+1234567890
```

#### Get Call Summaries
```http
GET /api/summaries?phone=+1234567890
```

#### Avatar Endpoints
```http
POST /api/avatar/session
Content-Type: application/json

{
  "replica_id": "optional-replica-id",
  "callback_url": "optional-callback-url"
}
```

```http
POST /api/avatar/session/:id/end
```

### WebSocket Protocol

**Connection**: `ws://localhost:8080/ws?room=room-name`

#### Client → Server Messages

1. **Binary Message**: Audio data (PCM 16-bit, 16kHz)
2. **Text Input**:
   ```json
   {
     "type": "text_input",
     "payload": "User text message"
   }
   ```
3. **End Call**:
   ```json
   {
     "type": "end_call",
     "payload": null
   }
   ```

#### Server → Client Messages

1. **Binary Message**: Audio data (TTS output)
2. **Connected**:
   ```json
   {
     "type": "connected",
     "payload": {
       "agent_id": "...",
       "room_name": "..."
     }
   }
   ```
3. **Transcript**:
   ```json
   {
     "type": "transcript",
     "payload": {
       "text": "User speech",
       "is_final": true
     }
   }
   ```
4. **Agent Response**:
   ```json
   {
     "type": "agent_response",
     "payload": "Agent's text response"
   }
   ```
5. **Tool Call**:
   ```json
   {
     "type": "tool_call",
     "payload": {
       "id": "...",
       "name": "book_appointment",
       "arguments": {...},
       "status": "executing"
     }
   }
   ```
6. **Tool Result**:
   ```json
   {
     "type": "tool_result",
     "payload": {
       "id": "...",
       "name": "book_appointment",
       "result": {...},
       "error": null
     }
   }
   ```
7. **Call Summary**:
   ```json
   {
     "type": "call_summary",
     "payload": {
       "summary": {...},
       "cost": {...}
     }
   }
   ```

## 🎨 Frontend Features

### UI Components

1. **Avatar Component**
   - Visual avatar display (Tavus iframe or fallback)
   - State indicators (speaking, listening, thinking)
   - Animated visual feedback

2. **Call Controls**
   - Start/End call buttons
   - Microphone toggle
   - Audio level visualization
   - Connection status

3. **Chat Panel**
   - Conversation history
   - Real-time transcript display
   - Text input for testing
   - Message timestamps

4. **Tool Display**
   - Real-time tool call visualization
   - Status indicators (executing, completed, failed)
   - Tool arguments display
   - Color-coded tool types

5. **Call Summary Modal**
   - Conversation summary
   - Booked appointments list
   - User preferences
   - Cost breakdown
   - Key topics discussed

### State Management

- **Zustand Store**: Centralized state for:
  - Connection status
  - Conversation messages
  - Tool calls
  - Avatar state
  - Call summary

### Audio Handling

- **Microphone Input**: Web Audio API for capture
- **Audio Streaming**: Real-time WebSocket binary transmission
- **Audio Playback**: Web Audio API for TTS output
- **Audio Level**: Visual feedback with waveform animation

## 🐳 Deployment

### Backend Deployment

#### Using Docker

```bash
cd voice-agent-backend
docker build -t voice-agent-backend .
docker run -p 8080:8080 --env-file .env voice-agent-backend
```

#### Using Docker Compose

```bash
docker-compose up -d
```

### Frontend Deployment

#### Build for Production

```bash
cd voice-agent-frontend
npm run build
```

Output is in `dist/` directory.

#### Deploy to Vercel/Netlify

1. Connect your repository
2. Set environment variables
3. Deploy

The frontend is configured for static hosting on Vercel, Netlify, or similar platforms.

## 💻 Development

### Running in Development

**Backend**:
```bash
cd voice-agent-backend
go run cmd/server/main.go
```

**Frontend**:
```bash
cd voice-agent-frontend
npm run dev
```

### Database Migrations

Supabase migrations are in `voice-agent-backend/migrations/`. Apply them through:
- Supabase Dashboard (SQL Editor)
- Supabase CLI

### Code Structure Guidelines

- **Backend**: Follow Go best practices, use interfaces for testability
- **Frontend**: Component-based architecture, custom hooks for logic
- **Error Handling**: Comprehensive error handling with user-friendly messages
- **Type Safety**: Full TypeScript coverage in frontend

### Testing

```bash
# Backend
cd voice-agent-backend
go test ./...

# Frontend
cd voice-agent-frontend
npm run test
```

## 📊 Database Schema

### Tables

- **users**: User information (phone, name)
- **appointments**: Appointment records (date, time, status)
- **call_summaries**: Call summary records

See `voice-agent-backend/migrations/001_initial_schema.sql` for full schema.

## 🔐 Security Considerations

- API keys stored in environment variables
- CORS configured for frontend origin
- Input validation on all endpoints
- SQL injection prevention via parameterized queries
- WebSocket origin checking

## 📝 License

This project is part of the SuperBryn AI Engineer Task assignment.

## 🤝 Contributing

This is an assignment project. For questions or issues, please refer to the project requirements.

---

**Built with ❤️ using Go, React, and modern AI services**
