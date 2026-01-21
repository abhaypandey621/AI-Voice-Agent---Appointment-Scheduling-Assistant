import type {
  WSMessage,
  WSMessageType,
  TranscriptPayload,
  ConnectionPayload,
  ToolCallPayload,
  ToolResultPayload,
  CallSummary,
  CostBreakdown,
} from '../types';

type MessageHandler = (message: WSMessage) => void;
type BinaryHandler = (data: ArrayBuffer) => void;

interface WSEventHandlers {
  onConnect?: (payload: ConnectionPayload) => void;
  onTranscript?: (payload: TranscriptPayload) => void;
  onAgentResponse?: (text: string) => void;
  onToolCall?: (payload: ToolCallPayload) => void;
  onToolResult?: (payload: ToolResultPayload) => void;
  onCallSummary?: (summary: CallSummary, cost: CostBreakdown) => void;
  onCallEnd?: () => void;
  onError?: (error: string) => void;
  onAudioData?: (data: ArrayBuffer) => void;
  onDisconnect?: () => void;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private handlers: WSEventHandlers = {};
  private messageHandlers: Map<WSMessageType, MessageHandler[]> = new Map();
  private binaryHandlers: BinaryHandler[] = [];
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isIntentionalClose = false;
  private pingInterval: number | null = null;

  constructor(url?: string) {
    if (url) {
      this.url = url;
    } else if (import.meta.env.VITE_WS_URL) {
      // Use configured WebSocket URL (should include /ws path)
      const baseUrl = import.meta.env.VITE_WS_URL;
      this.url = baseUrl.endsWith('/ws') ? baseUrl : `${baseUrl}/ws`;
    } else {
      // Fallback to same origin
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      this.url = `${wsProtocol}//${window.location.host}/ws`;
    }
    console.log('WebSocket URL configured:', this.url);
  }

  connect(roomName?: string, handlers?: WSEventHandlers): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      this.handlers = handlers || {};
      this.isIntentionalClose = false;

      const wsUrl = roomName ? `${this.url}?room=${encodeURIComponent(roomName)}` : this.url;

      try {
        this.ws = new WebSocket(wsUrl);
        this.ws.binaryType = 'arraybuffer';

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          this.startPingInterval();
          resolve();
        };

        this.ws.onmessage = (event) => {
          if (event.data instanceof ArrayBuffer) {
            // Binary message (audio data)
            this.binaryHandlers.forEach((handler) => handler(event.data));
            this.handlers.onAudioData?.(event.data);
          } else {
            // Text message (JSON)
            try {
              console.log('[WebSocket] Raw message received:', event.data);
              const message: WSMessage = JSON.parse(event.data);
              console.log('[WebSocket] Parsed message type:', message.type);
              this.handleMessage(message);
            } catch (e) {
              console.error('Failed to parse WebSocket message:', e, 'Raw data:', event.data);
            }
          }
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket closed:', event.code, event.reason);
          this.stopPingInterval();
          this.handlers.onDisconnect?.();

          if (!this.isIntentionalClose && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            setTimeout(() => {
              console.log(`Reconnecting... Attempt ${this.reconnectAttempts}`);
              this.connect(roomName, handlers);
            }, this.reconnectDelay * this.reconnectAttempts);
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect(): void {
    this.isIntentionalClose = true;
    this.stopPingInterval();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private handleMessage(message: WSMessage): void {
    const type = message.type as WSMessageType;

    // Call registered handlers for this message type
    const handlers = this.messageHandlers.get(type);
    if (handlers) {
      handlers.forEach((handler) => handler(message));
    }

    // Call specific event handlers
    switch (type) {
      case 'connected':
        this.handlers.onConnect?.(message.payload as ConnectionPayload);
        break;
      case 'transcript':
        this.handlers.onTranscript?.(message.payload as TranscriptPayload);
        break;
      case 'agent_response':
        this.handlers.onAgentResponse?.(message.payload as string);
        break;
      case 'tool_call':
        this.handlers.onToolCall?.(message.payload as ToolCallPayload);
        break;
      case 'tool_result':
        this.handlers.onToolResult?.(message.payload as ToolResultPayload);
        break;
      case 'call_summary': {
        console.log('[WebSocket] Received call_summary:', message.payload);
        const summaryPayload = message.payload as { summary: CallSummary; cost: CostBreakdown };
        console.log('[WebSocket] Parsed summary:', summaryPayload.summary);
        console.log('[WebSocket] Parsed cost:', summaryPayload.cost);
        if (this.handlers.onCallSummary) {
          console.log('[WebSocket] Calling onCallSummary handler');
          this.handlers.onCallSummary(summaryPayload.summary, summaryPayload.cost);
        } else {
          console.warn('[WebSocket] No onCallSummary handler registered!');
        }
        break;
      }
      case 'call_end':
        console.log('[WebSocket] Received call_end');
        this.handlers.onCallEnd?.();
        break;
      case 'error':
        this.handlers.onError?.(message.payload as string);
        break;
    }
  }

  on(type: WSMessageType, handler: MessageHandler): () => void {
    const handlers = this.messageHandlers.get(type) || [];
    handlers.push(handler);
    this.messageHandlers.set(type, handlers);

    // Return unsubscribe function
    return () => {
      const currentHandlers = this.messageHandlers.get(type) || [];
      const index = currentHandlers.indexOf(handler);
      if (index > -1) {
        currentHandlers.splice(index, 1);
        this.messageHandlers.set(type, currentHandlers);
      }
    };
  }

  onBinary(handler: BinaryHandler): () => void {
    this.binaryHandlers.push(handler);
    return () => {
      const index = this.binaryHandlers.indexOf(handler);
      if (index > -1) {
        this.binaryHandlers.splice(index, 1);
      }
    };
  }

  send(message: WSMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  sendBinary(data: ArrayBuffer | Uint8Array): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(data);
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  // Convenience methods
  sendTextInput(text: string): void {
    // Validate text input
    const cleanText = text?.trim() || '';

    if (!cleanText || cleanText === 'null' || cleanText === 'undefined') {
      console.warn('Invalid text input - cannot send empty or null values');
      return;
    }

    this.send({ type: 'text_input', payload: cleanText });
  }

  endCall(): void {
    this.send({ type: 'end_call', payload: null });
  }

  getSession(): void {
    this.send({ type: 'get_session', payload: null });
  }

  ping(): void {
    this.send({ type: 'ping', payload: null });
  }

  private startPingInterval(): void {
    this.pingInterval = window.setInterval(() => {
      this.ping();
    }, 30000);
  }

  private stopPingInterval(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export const wsService = new WebSocketService();
export default wsService;
