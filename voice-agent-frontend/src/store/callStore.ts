import { create } from 'zustand';
import type {
  CallState,
  AvatarState,
  ConversationMessage,
  ToolCallPayload,
  ToolResultPayload,
  CallSummary,
  CostBreakdown,
} from '../types';

interface CallStore {
  // Connection state
  agentId: string | null;
  roomName: string | null;
  callState: CallState;
  avatarState: AvatarState;
  isConnected: boolean;
  error: string | null;

  // Conversation
  messages: ConversationMessage[];
  currentTranscript: string;
  isTranscriptFinal: boolean;

  // Tool calls
  toolCalls: ToolCallPayload[];
  activeToolCall: ToolCallPayload | null;

  // Summary and cost
  callSummary: CallSummary | null;
  costBreakdown: CostBreakdown | null;

  // Avatar
  avatarUrl: string | null;
  avatarConversationId: string | null;

  // Actions
  setConnection: (agentId: string, roomName: string) => void;
  setCallState: (state: CallState) => void;
  setAvatarState: (state: AvatarState) => void;
  setConnected: (connected: boolean) => void;
  setError: (error: string | null) => void;

  addMessage: (message: ConversationMessage) => void;
  setTranscript: (text: string, isFinal: boolean) => void;

  addToolCall: (toolCall: ToolCallPayload) => void;
  updateToolCall: (result: ToolResultPayload) => void;
  setActiveToolCall: (toolCall: ToolCallPayload | null) => void;

  setCallSummary: (summary: CallSummary) => void;
  setCostBreakdown: (cost: CostBreakdown) => void;

  setAvatarUrl: (url: string | null) => void;
  setAvatarConversationId: (id: string | null) => void;

  reset: () => void;
}

const initialState = {
  agentId: null,
  roomName: null,
  callState: 'idle' as CallState,
  avatarState: 'idle' as AvatarState,
  isConnected: false,
  error: null,
  messages: [],
  currentTranscript: '',
  isTranscriptFinal: false,
  toolCalls: [],
  activeToolCall: null,
  callSummary: null,
  costBreakdown: null,
  avatarUrl: null,
  avatarConversationId: null,
};

export const useCallStore = create<CallStore>((set) => ({
  ...initialState,

  setConnection: (agentId, roomName) =>
    set({ agentId, roomName, isConnected: true, callState: 'connected' }),

  setCallState: (callState) => set({ callState }),

  setAvatarState: (avatarState) => set({ avatarState }),

  setConnected: (isConnected) =>
    set({ isConnected, callState: isConnected ? 'connected' : 'idle' }),

  setError: (error) => set({ error }),

  addMessage: (message) =>
    set((state) => ({
      messages: [...state.messages, message],
      currentTranscript: '',
      isTranscriptFinal: false,
    })),

  setTranscript: (text, isFinal) =>
    set({ currentTranscript: text, isTranscriptFinal: isFinal }),

  addToolCall: (toolCall) =>
    set((state) => ({
      toolCalls: [...state.toolCalls, toolCall],
      activeToolCall: toolCall,
    })),

  updateToolCall: (result) =>
    set((state) => ({
      toolCalls: state.toolCalls.map((tc) =>
        tc.id === result.id
          ? { ...tc, status: result.error ? 'failed' : 'completed' }
          : tc
      ),
      activeToolCall:
        state.activeToolCall?.id === result.id
          ? { ...state.activeToolCall, status: result.error ? 'failed' : 'completed' }
          : state.activeToolCall,
    })),

  setActiveToolCall: (activeToolCall) => set({ activeToolCall }),

  setCallSummary: (callSummary) => set({ callSummary, callState: 'ended' }),

  setCostBreakdown: (costBreakdown) => set({ costBreakdown }),

  setAvatarUrl: (avatarUrl) => set({ avatarUrl }),

  setAvatarConversationId: (avatarConversationId) => set({ avatarConversationId }),

  reset: () => set(initialState),
}));

export default useCallStore;
