export interface WorkSession {
  id: number;
  check_in_time: string;
  check_out_time?: string;
  expected_check_out_time: string;
  work_hours: number;
  overtime_minutes: number;
  date: string;
}

export interface StatusResponse {
  has_checked_in: boolean;
  check_in_time?: string;
  check_out_time?: string;
  expected_check_out_time?: string;
  current_time: string;
  is_check_out_time: boolean;
  work_hours: number;
  overtime_minutes: number;
}

export interface WorkConfig {
  work_hours: number;
  auto_fetch_enabled: boolean;
  check_in_api_url: string;
  authorization: string;
  p_auth: string;
  p_rtoken: string;
  check_in_webhook_url: string;
  check_out_webhook_url: string;
}

export interface CheckInRequest {
  check_in_time: string;
}

export interface CheckInResponse {
  check_in_time: string;
  expected_check_out_time: string;
}

export interface CheckOutRequest {
  check_out_time: string;
}

export interface CheckOutResponse {
  check_out_time: string;
  overtime_minutes: number;
}

export interface TodayCheckInRequest {
  date: string;
  re_check_in?: boolean;
}

export interface TodayCheckInResponse {
  has_checked_in: boolean;
  check_in_time?: string;
  can_auto_fetch: boolean;
  auto_fetch_enabled: boolean;
  api_error?: string;
}

export interface MonthStats {
  year_month: string;
  total_days: number;
  checked_out_days: number;
  overtime_minutes: number;
}

export interface MonthlyStatsResponse {
  current_month: MonthStats;
  last_month: MonthStats;
}
