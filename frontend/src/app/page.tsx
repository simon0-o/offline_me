'use client';

import { useState, useEffect, useCallback } from 'react';
import { api } from '@/lib/api';
import { requestNotificationPermission, showNotification } from '@/lib/utils';
import type { StatusResponse, MonthlyStatsResponse } from '@/lib/types';
import StatusCard from '@/components/StatusCard';
import MonthlyStatsCard from '@/components/MonthlyStatsCard';
import CheckInSection from '@/components/CheckInSection';
import ReCheckInSection from '@/components/ReCheckInSection';
import CheckOutSection from '@/components/CheckOutSection';
import ConfigSection from '@/components/ConfigSection';

export default function Home() {
  const [status, setStatus] = useState<StatusResponse | null>(null);
  const [monthlyStats, setMonthlyStats] = useState<MonthlyStatsResponse | null>(null);
  const [checkOutNotified, setCheckOutNotified] = useState(false);

  const loadStatus = useCallback(async () => {
    try {
      const data = await api.getStatus();
      setStatus(data);

      if (data.is_check_out_time && data.has_checked_in) {
        if (!checkOutNotified) {
          showNotification('Work Day Complete!', 'Time to check out!');
          setCheckOutNotified(true);
        }
      } else {
        setCheckOutNotified(false);
      }
    } catch (error) {
      console.error('Failed to load status:', error);
    }
  }, [checkOutNotified]);

  const loadMonthlyStats = useCallback(async () => {
    try {
      const data = await api.getMonthlyStats();
      setMonthlyStats(data);
    } catch (error) {
      console.error('Failed to load monthly stats:', error);
    }
  }, []);

  const handleRefresh = useCallback(() => {
    loadStatus();
    loadMonthlyStats();
  }, [loadStatus, loadMonthlyStats]);

  useEffect(() => {
    requestNotificationPermission();
    loadStatus();
    loadMonthlyStats();

    const statusInterval = setInterval(loadStatus, 30000);
    const statsInterval = setInterval(loadMonthlyStats, 60000);

    return () => {
      clearInterval(statusInterval);
      clearInterval(statsInterval);
    };
  }, [loadStatus, loadMonthlyStats]);

  // Auto-fetch check-in time on page load if not checked in
  useEffect(() => {
    if (status && !status.has_checked_in) {
      const tryAutoFetch = async () => {
        try {
          const today = new Date().toISOString().split('T')[0];
          await api.getTodayCheckIn({ date: today });
        } catch (error) {
          console.error('Failed to auto-fetch check-in time:', error);
        }
      };
      tryAutoFetch();
    }
  }, [status?.has_checked_in]);

  return (
    <div className="min-h-screen bg-gray-100 py-5 px-4">
      <div className="max-w-4xl mx-auto bg-white rounded-xl shadow-lg p-8">
        <h1 className="text-3xl font-bold text-gray-900 text-center mb-8">
          🕒 Work Time Tracker
        </h1>

        {status?.is_check_out_time && status.has_checked_in && (
          <div className="bg-green-600 text-white p-5 rounded-lg text-center text-lg font-semibold mb-5 animate-pulse">
            🎉 Time to check out! Your work day is complete!
          </div>
        )}

        <MonthlyStatsCard stats={monthlyStats} />
        <StatusCard status={status} />

        {status?.has_checked_in ? (
          <>
            <ReCheckInSection
              currentCheckInTime={status.check_in_time!}
              onSuccess={handleRefresh}
            />
            <CheckOutSection onSuccess={handleRefresh} />
          </>
        ) : (
          <CheckInSection onSuccess={handleRefresh} />
        )}

        <ConfigSection onConfigUpdate={handleRefresh} />
      </div>
    </div>
  );
}
