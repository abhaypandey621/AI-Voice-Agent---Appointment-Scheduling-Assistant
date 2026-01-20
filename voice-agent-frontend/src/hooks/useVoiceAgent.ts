import { useCallback, useEffect, useRef } from 'react';
import { useCallStore } from '../store/callStore';
import { wsService } from '../services/websocket';
import { api } from '../services/api';
import { useAudio, float32ToInt16, int16ToArrayBuffer } from './useAudio';
import type {
  ConnectionPayload,
  TranscriptPayload,
  ToolCallPayload,
  ToolResultPayload,
  CallSummary,
  CostBreakdown,
  CallState,
  AvatarState,
  ConversationMessage,
} from '../types';

interface UseVoiceAgentOptions {
  autoConnect?: boolean;
  roomName?: string;
}

interface UseVoiceAgentReturn {
  isConnected: boolean;
  isRecording: boolean;
  callState: CallState;
  avatarState: AvatarState;
  error: string | null;
  messages: ConversationMessage[];
  currentTranscript: string;
  toolCalls: ToolCallPayload[];
  activeToolCall: ToolCallPayload | null;
  callSummary: CallSummary | null;
  costBreakdown: CostBreakdown | null;
  audioLevel: number;
  connect: () => Promise<void>;
  disconnect: () => void;
  startRecording: () => Promise<void>;
  stopRecording: () => void;
  sendText: (text: string) => void;
  endCall: () => void;
}

export function useVoiceAgent(
  options: UseVoiceAgentOptions = {}
): UseVoiceAgentReturn {
  const { autoConnect = false, roomName } = options;

  const store = useCallStore();
  const audioContextRef = useRef<AudioContext | null>(null);
  const audioQueueRef = useRef<ArrayBuffer[]>([]);
  const isPlayingRef = useRef(false);

  // Handle audio data from microphone
  const handleAudioData = useCallback((data: Float32Array) => {
    if (wsService.isConnected()) {
      const int16Data = float32ToInt16(data);
      const buffer = int16ToArrayBuffer(int16Data);
      wsService.sendBinary(buffer);
    }
  }, []);

  const {
    isRecording,
    isSupported,
    startRecording: startMicRecording,
    stopRecording: stopMicRecording,
    audioLevel,
  } = useAudio({ onAudioData: handleAudioData });

  // Play received audio
  const playAudio = useCallback(async (audioData: ArrayBuffer) => {
    audioQueueRef.current.push(audioData);

    if (isPlayingRef.current) return;
    isPlayingRef.current = true;

    while (audioQueueRef.current.length > 0) {
      const data = audioQueueRef.current.shift();
      if (!data) continue;

      try {
        if (!audioContextRef.current) {
          audioContextRef.current = new AudioContext({ sampleRate: 24000 });
        }

        const audioContext = audioContextRef.current;

        // Convert Int16 PCM to Float32
        const int16Array = new Int16Array(data);
        const float32Array = new Float32Array(int16Array.length);
        for (let i = 0; i < int16Array.length; i++) {
          float32Array[i] = int16Array[i] / 32768;
        }

        // Create audio buffer
        const audioBuffer = audioContext.createBuffer(
          1,
          float32Array.length,
          24000
        );
        audioBuffer.getChannelData(0).set(float32Array);

        // Play the audio
        const source = audioContext.createBufferSource();
        source.buffer = audioBuffer;
        source.connect(audioContext.destination);

        await new Promise<void>((resolve) => {
          source.onended = () => resolve();
          source.start();
        });
      } catch (error) {
        console.error('Failed to play audio:', error);
      }
    }

    isPlayingRef.current = false;
  }, []);

  // WebSocket connection
  const connect = useCallback(async () => {
    store.setCallState('connecting');
    store.setError(null);

    try {
      // Create avatar session if available
      try {
        const avatarSession = await api.createAvatarSession();
        if (avatarSession?.conversation_url) {
          store.setAvatarUrl(avatarSession.conversation_url);
          store.setAvatarConversationId(avatarSession.conversation_id);
        }
      } catch (error) {
        // Avatar is optional, continue without it
        console.warn('Failed to create avatar session:', error);
      }

      await wsService.connect(roomName, {
        onConnect: (payload: ConnectionPayload) => {
          store.setConnection(payload.agent_id, payload.room_name);
          store.setCallState('connected');
        },
        onTranscript: (payload: TranscriptPayload) => {
          store.setTranscript(payload.text, payload.is_final);
          if (payload.is_final && payload.text) {
            store.addMessage({
              role: 'user',
              content: payload.text,
              timestamp: new Date().toISOString(),
            });
            store.setAvatarState('thinking');
          } else {
            store.setCallState('listening');
          }
        },
        onAgentResponse: (text: string) => {
          store.addMessage({
            role: 'assistant',
            content: text,
            timestamp: new Date().toISOString(),
          });
          store.setAvatarState('speaking');
          store.setCallState('speaking');
        },
        onToolCall: (payload: ToolCallPayload) => {
          store.addToolCall(payload);
          store.setCallState('processing');
        },
        onToolResult: (payload: ToolResultPayload) => {
          store.updateToolCall(payload);
        },
        onCallSummary: (summary: CallSummary, cost: CostBreakdown) => {
          store.setCallSummary(summary);
          store.setCostBreakdown(cost);
        },
        onCallEnd: () => {
          store.setCallState('ended');
          store.setConnected(false);
          stopMicRecording();
        },
        onError: (error: string) => {
          store.setError(error);
        },
        onAudioData: (data: ArrayBuffer) => {
          playAudio(data);
        },
        onDisconnect: () => {
          store.setConnected(false);
          store.setCallState('idle');
        },
      });
    } catch (error) {
      store.setError(error instanceof Error ? error.message : 'Connection failed');
      store.setCallState('idle');
    }
  }, [roomName, store, playAudio, stopMicRecording]);

  const disconnect = useCallback(() => {
    stopMicRecording();

    // End avatar session if active
    const conversationId = store.avatarConversationId;
    if (conversationId) {
      api.endAvatarSession(conversationId).catch((err) => {
        console.warn('Failed to end avatar session:', err);
      });
    }

    wsService.disconnect();
    store.reset();

    if (audioContextRef.current) {
      audioContextRef.current.close();
      audioContextRef.current = null;
    }
  }, [stopMicRecording, store]);

  const startRecording = useCallback(async () => {
    if (!isSupported) {
      store.setError('Audio recording is not supported in this browser');
      return;
    }

    if (!wsService.isConnected()) {
      await connect();
    }

    await startMicRecording();
    store.setCallState('listening');
    store.setAvatarState('listening');
  }, [isSupported, connect, startMicRecording, store]);

  const stopRecording = useCallback(() => {
    stopMicRecording();
    if (store.callState === 'listening') {
      store.setCallState('connected');
    }
    store.setAvatarState('idle');
  }, [stopMicRecording, store]);

  const sendText = useCallback((text: string) => {
    if (!wsService.isConnected()) {
      store.setError('Not connected');
      return;
    }

    wsService.sendTextInput(text);
    store.addMessage({
      role: 'user',
      content: text,
      timestamp: new Date().toISOString(),
    });
    store.setCallState('processing');
    store.setAvatarState('thinking');
  }, [store]);

  const endCall = useCallback(() => {
    wsService.endCall();
    store.setCallState('ending');
  }, [store]);

  // Auto-connect if option is set
  useEffect(() => {
    if (autoConnect) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [autoConnect]); // eslint-disable-line react-hooks/exhaustive-deps

  return {
    isConnected: store.isConnected,
    isRecording,
    callState: store.callState,
    avatarState: store.avatarState,
    error: store.error,
    messages: store.messages,
    currentTranscript: store.currentTranscript,
    toolCalls: store.toolCalls,
    activeToolCall: store.activeToolCall,
    callSummary: store.callSummary,
    costBreakdown: store.costBreakdown,
    audioLevel,
    connect,
    disconnect,
    startRecording,
    stopRecording,
    sendText,
    endCall,
  };
}

export default useVoiceAgent;
