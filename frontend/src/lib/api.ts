import type {
  StatusResponse,
  WorkConfig,
  CheckInRequest,
  CheckInResponse,
  CheckOutRequest,
  CheckOutResponse,
  TodayCheckInRequest,
  TodayCheckInResponse,
  MonthlyStatsResponse,
} from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.statusText}`);
  }

  return response.json();
}

export const api = {
  async getStatus(): Promise<StatusResponse> {
    return fetchApi<StatusResponse>('/api/status');
  },

  async getConfig(): Promise<WorkConfig> {
    return fetchApi<WorkConfig>('/api/config');
  },

  async updateConfig(config: Partial<WorkConfig>): Promise<void> {
    return fetchApi('/api/config', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  },

  async checkIn(data: CheckInRequest): Promise<CheckInResponse> {
    return fetchApi<CheckInResponse>('/api/checkin', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async checkOut(data: CheckOutRequest): Promise<CheckOutResponse> {
    return fetchApi<CheckOutResponse>('/api/checkout', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async getTodayCheckIn(data: TodayCheckInRequest): Promise<TodayCheckInResponse> {
    return fetchApi<TodayCheckInResponse>('/api/today-checkin', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async getMonthlyStats(): Promise<MonthlyStatsResponse> {
    return fetchApi<MonthlyStatsResponse>('/api/monthly-stats');
  },
};
