import type {
  TokenResponse,
  Appointment,
  TimeSlot,
  CallSummary,
  AvatarSession,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

// Log the API URL being used (for debugging)
if (import.meta.env.DEV) {
  console.log('[API Service] Using API URL:', API_BASE_URL);
} else {
  console.log('[API Service] Production API URL configured');
} class APIService {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({
        error: 'Request failed',
      }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // Room management
  async createRoom(roomName?: string): Promise<TokenResponse> {
    return this.request<TokenResponse>('/api/rooms', {
      method: 'POST',
      body: JSON.stringify({ room_name: roomName }),
    });
  }

  async getToken(room: string, participant?: string): Promise<TokenResponse> {
    const params = new URLSearchParams({ room });
    if (participant) params.append('participant', participant);
    return this.request<TokenResponse>(`/api/token?${params}`);
  }

  // Avatar sessions
  async createAvatarSession(
    replicaId?: string,
    callbackUrl?: string
  ): Promise<AvatarSession> {
    return this.request<AvatarSession>('/api/avatar/session', {
      method: 'POST',
      body: JSON.stringify({
        replica_id: replicaId,
        callback_url: callbackUrl,
      }),
    });
  }

  async endAvatarSession(conversationId: string): Promise<void> {
    await this.request(`/api/avatar/session/${conversationId}/end`, {
      method: 'POST',
    });
  }

  async getAvatarReplicas(): Promise<{ replicas: Record<string, unknown>[] }> {
    return this.request('/api/avatar/replicas');
  }

  // Appointments
  async getAppointments(phone: string): Promise<{ appointments: Appointment[]; count: number }> {
    return this.request(`/api/appointments?phone=${encodeURIComponent(phone)}`);
  }

  async getAvailableSlots(date: string): Promise<{ date: string; slots: TimeSlot[] }> {
    return this.request(`/api/slots?date=${encodeURIComponent(date)}`);
  }

  // Call summaries
  async getCallSummaries(phone: string): Promise<{ summaries: CallSummary[]; count: number }> {
    return this.request(`/api/summaries?phone=${encodeURIComponent(phone)}`);
  }

  // Stats
  async getStats(): Promise<{ active_connections: number; timestamp: string }> {
    return this.request('/api/stats');
  }

  // Health check
  async healthCheck(): Promise<{ status: string; timestamp: string; version: string }> {
    return this.request('/health');
  }
}

export const api = new APIService();
export default api;
