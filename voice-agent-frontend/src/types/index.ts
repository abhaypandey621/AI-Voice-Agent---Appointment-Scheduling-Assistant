// User types
export interface User {
  id: string;
  phone_number: string;
  name?: string;
  email?: string;
  created_at: string;
  updated_at: string;
}

// Appointment types
export interface Appointment {
  id: string;
  user_phone: string;
  user_name?: string;
  date_time: string;
  duration: number;
  purpose?: string;
  status: 'booked' | 'cancelled' | 'completed';
  notes?: string;
  created_at: string;
  updated_at: string;
}

// Time slot types
export interface TimeSlot {
  date_time: string;
  time: string;
  available: boolean;
  duration: number;
}

// Call session types
export interface CallSession {
  id: string;
  room_name: string;
  user_phone?: string;
  user_name?: string;
  started_at: string;
  ended_at?: string;
  messages: ConversationMessage[];
  tool_calls: ToolCallRecord[];
  cost_breakdown?: CostBreakdown;
}

// Conversation message types
export interface ConversationMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
}

// Tool call types
export interface ToolCallRecord {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
  result?: unknown;
  timestamp: string;
}

export interface ToolCallPayload {
  id: string;
  name: string;
  arguments: Record<string, unknown>;
  status: 'pending' | 'executing' | 'completed' | 'failed';
  result?: unknown;
  error?: string;
}

export interface ToolResultPayload {
  id: string;
  name: string;
  result: unknown;
  error?: string;
}

// Call summary types
export interface CallSummary {
  id?: string;
  session_id: string;
  user_phone?: string;
  summary: string;
  appointments_booked: Appointment[];
  user_preferences: string[];
  key_topics: string[];
  duration: number;
  duration_seconds?: number;  // Backend sends this instead of duration
  created_at: string;
}

// Cost breakdown types
export interface CostBreakdown {
  stt_cost: number;
  tts_cost: number;
  llm_cost: number;
  avatar_cost: number;
  total_cost: number;
  stt_minutes: number;
  tts_characters: number;
  llm_tokens: number;
}

// WebSocket message types
export interface WSMessage<T = unknown> {
  type: string;
  payload: T;
}

export type WSMessageType =
  | 'connected'
  | 'transcript'
  | 'agent_response'
  | 'tool_call'
  | 'tool_result'
  | 'call_summary'
  | 'call_end'
  | 'error'
  | 'avatar_state'
  | 'cost_update'
  | 'session'
  | 'pong';

// Transcript payload
export interface TranscriptPayload {
  text: string;
  is_final: boolean;
}

// Connection payload
export interface ConnectionPayload {
  agent_id: string;
  room_name: string;
}

// Avatar session types
export interface AvatarSession {
  conversation_id: string;
  conversation_url: string;
  status: string;
}

// Token response
export interface TokenResponse {
  token: string;
  room_name: string;
  url: string;
  participant_name?: string;
}

// API error response
export interface APIError {
  error: string;
  message?: string;
}

// Call state for UI
export type CallState =
  | 'idle'
  | 'connecting'
  | 'connected'
  | 'speaking'
  | 'listening'
  | 'processing'
  | 'ending'
  | 'ended';

// Avatar state for UI
export type AvatarState =
  | 'idle'
  | 'speaking'
  | 'listening'
  | 'thinking';
