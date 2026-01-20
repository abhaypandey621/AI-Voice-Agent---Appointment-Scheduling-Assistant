# AI Voice Agent - Frontend

A modern React TypeScript frontend for the AI Voice Agent appointment scheduling assistant.

## 🚀 Features

- **Real-time Voice Interaction** with AI assistant
- **Visual Avatar** display with state indicators
- **Tool Call Visualization** showing AI actions in real-time
- **Conversation Transcript** panel
- **Call Summary** with cost breakdown
- **Multi-language Support** (7 languages)
- **Responsive Design** for desktop and mobile

## 📋 Prerequisites

- Node.js 18+
- npm or yarn
- Backend server running on port 8080

## 🛠️ Setup

### 1. Install Dependencies

```bash
cd voice-agent-frontend
npm install
```

### 2. Configure Environment

The `.env` file is already configured for local development:

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

For production, update these to your deployed backend URL.

### 3. Start Development Server

```bash
npm run dev
```

The app will open at `http://localhost:3000`

## 🧪 Testing Guide

### Prerequisites for Testing

1. **Backend must be running** on port 8080
2. **Backend must have valid API keys** configured in `.env`

### Step-by-Step Testing

#### 1. Start the Backend First

```bash
cd voice-agent-backend
# Make sure .env has valid API keys
make run
```

Verify backend is running:
```bash
curl http://localhost:8080/health
# Should return: {"status":"healthy",...}
```

#### 2. Start the Frontend

```bash
cd voice-agent-frontend
npm run dev
```

#### 3. Test the Voice Agent

1. Open `http://localhost:3000` in your browser
2. Click **"Start Call"** button
3. Wait for "Connected" status
4. Click the **microphone button** to start speaking
5. Say: "Hi, I'd like to book an appointment"
6. The AI will ask for your phone number
7. Provide a phone number (e.g., "My number is 555-123-4567")
8. Ask to see available slots or book an appointment
9. When done, say "Goodbye" or click "End Call"
10. View the call summary with cost breakdown

### Testing Without Voice (Text Input)

If you don't have a microphone or prefer text:

1. Start a call as above
2. Use the chat input field at the bottom right
3. Type messages and press Enter or click Send

### Testing Tool Calls

Try these phrases to trigger different tools:

| Say This | Tool Triggered |
|----------|----------------|
| "My phone number is 555-1234" | `identify_user` |
| "What times are available tomorrow?" | `fetch_slots` |
| "Book an appointment for 2pm" | `book_appointment` |
| "Show my appointments" | `retrieve_appointments` |
| "Cancel my appointment" | `cancel_appointment` |
| "Change my appointment to 3pm" | `modify_appointment` |
| "Goodbye" | `end_conversation` |

## 🏗️ Project Structure

```
voice-agent-frontend/
├── public/
├── src/
│   ├── components/
│   │   ├── Avatar/          # Avatar display component
│   │   ├── Call/            # Call controls
│   │   ├── Chat/            # Chat panel
│   │   ├── Summary/         # Call summary modal
│   │   ├── ToolDisplay/     # Tool execution display
│   │   └── UI/              # Shared UI components
│   ├── hooks/
│   │   ├── useAudio.ts      # Audio recording hook
│   │   └── useVoiceAgent.ts # Main voice agent hook
│   ├── services/
│   │   ├── api.ts           # REST API client
│   │   └── websocket.ts     # WebSocket client
│   ├── store/
│   │   └── callStore.ts     # Zustand state store
│   ├── styles/
│   │   └── index.css        # Tailwind CSS styles
│   ├── types/
│   │   └── index.ts         # TypeScript types
│   ├── utils/
│   │   └── i18n.ts          # Internationalization
│   ├── App.tsx              # Main app component
│   └── main.tsx             # Entry point
├── .env                     # Environment variables
├── package.json
├── tailwind.config.js
├── tsconfig.json
└── vite.config.ts
```

## 🔧 Troubleshooting

### "Connection Failed" Error

1. **Check backend is running:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check frontend .env file:**
   ```
   VITE_API_URL=http://localhost:8080
   VITE_WS_URL=ws://localhost:8080
   ```

3. **Check browser console** for detailed error messages

4. **Check backend logs** for any API key errors

### Microphone Not Working

1. Allow microphone access when prompted
2. Check browser permissions (click lock icon in URL bar)
3. Try using HTTPS (required for some browsers)

### No Audio Response

1. Check browser volume
2. Verify Cartesia API key is valid in backend
3. Check backend logs for TTS errors

## 🐳 Docker

```bash
# Build
docker build -t voice-agent-frontend .

# Run (ensure backend is accessible)
docker run -p 80:80 voice-agent-frontend
```

## 📦 Production Build

```bash
npm run build
# Output in dist/ folder
```

Deploy the `dist/` folder to any static hosting (Netlify, Vercel, etc.)

## 📄 License

MIT
