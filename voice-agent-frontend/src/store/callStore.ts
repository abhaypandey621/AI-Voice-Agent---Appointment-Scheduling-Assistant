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

  setCallState: (callState) => {
    console.log('[CallStore] setCallState called with:', callState);
    return set({ callState });
  },

  setAvatarState: (avatarState) => set({ avatarState }),

  setConnected: (isConnected) =>
    set((state) => {
      const newCallState = isConnected
        ? 'connected'
        : state.callSummary
        ? 'ended'
        : 'idle';
      console.log('[CallStore] setConnected called:', isConnected, 'hasSummary:', !!state.callSummary, 'newCallState:', newCallState);
      return {
        isConnected,
        callState: newCallState,
      };
    }),

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
          ? {
              ...tc,
              status: result.error ? 'failed' : 'completed',
              result: result.result,
              error: result.error,
            }
          : tc
      ),
      activeToolCall:
        state.activeToolCall?.id === result.id
          ? {
              ...state.activeToolCall,
              status: result.error ? 'failed' : 'completed',
              result: result.result,
              error: result.error,
            }
          : state.activeToolCall,
    })),

  setActiveToolCall: (activeToolCall) => set({ activeToolCall }),

  setCallSummary: (callSummary) => {
    console.log('[CallStore] setCallSummary called, setting callState to ended');
    return set({ callSummary, callState: 'ended' });
  },

  setCostBreakdown: (costBreakdown) => set({ costBreakdown }),

  setAvatarUrl: (avatarUrl) => set({ avatarUrl }),

  setAvatarConversationId: (avatarConversationId) => set({ avatarConversationId }),

  reset: () => {
    console.log('[CallStore] reset called - clearing all state');
    console.trace('[CallStore] reset call stack');
    return set(initialState);
  },
}));

export default useCallStore;
