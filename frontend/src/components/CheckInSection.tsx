'use client';

import { useState } from 'react';
import { formatTime, showNotification, formatDateTime } from '@/lib/utils';
import { api } from '@/lib/api';

interface CheckInSectionProps {
  currentCheckInTime: string | null;
  onSuccess: () => void;
}

export default function CheckInSection({
  currentCheckInTime,
  onSuccess,
}: CheckInSectionProps) {
  const [reCheckInTime, setReCheckInTime] = useState(
    formatTime(new Date(currentCheckInTime || ''))
  );
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);

  const handleReCheckIn = async () => {
    if (!reCheckInTime) {
      alert('Please select a new check-in time');
      return;
    }

    setLoading(true);
    try {
      const today = new Date();
      const [hours, minutes] = reCheckInTime.split(':');
      const checkInDate = new Date(
        today.getFullYear(),
        today.getMonth(),
        today.getDate(),
        parseInt(hours),
        parseInt(minutes)
      );

      const result = await api.checkIn({
        check_in_time: checkInDate.toISOString(),
      });

      showNotification(
        'Re-Check-in Successful!',
        `New expected check-out: ${formatDateTime(result.expected_check_out_time)}`
      );
      onSuccess();
    } catch (error) {
      console.error('Failed to re-check in:', error);
      alert('Failed to re-check in');
    } finally {
      setLoading(false);
    }
  };

  const handleFetchFromAPI = async () => {
    setFetching(true);
    try {
      const today = new Date().toISOString().split('T')[0];
      const data = await api.getTodayCheckIn({
        date: today,
        re_check_in: true,
      });

      if (data.has_checked_in && data.check_in_time) {
        const checkInTime = new Date(data.check_in_time);
        alert(
          `✅ Successfully fetched and updated check-in time from HR API: ${formatTime(checkInTime)}`
        );
        onSuccess();
      } else if (data.check_in_time) {
        const checkInTime = new Date(data.check_in_time);
        setReCheckInTime(formatTime(checkInTime));
        alert(`✅ Fetched check-in time from HR API: ${formatTime(checkInTime)}`);
      } else if (data.api_error) {
        alert(`⚠️ HR API Error: ${data.api_error}`);
      } else {
        alert(
          '⚠️ Could not fetch check-in time from HR API. Please check your API configuration.'
        );
      }
    } catch (error) {
      console.error('Failed to auto-fetch for re-check-in:', error);
      alert('❌ Error fetching from HR API');
    } finally {
      setFetching(false);
    }
  };

  return (
    <div className="mb-8">
      <h3 className="text-lg font-semibold mb-3">Update Check-in Time</h3>
      <p className="mb-3 text-sm text-gray-600">
        Already checked in today. You can update your check-in time if needed:
      </p>
      <div className="flex flex-wrap items-center gap-2">
        <label htmlFor="reCheckInTime" className="text-sm">
          New check-in time:
        </label>
        <input
          type="time"
          id="reCheckInTime"
          value={reCheckInTime}
          onChange={(e) => setReCheckInTime(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-base"
        />
        <button
          onClick={handleReCheckIn}
          disabled={loading}
          className="bg-orange-500 text-white px-5 py-2 rounded-md text-base cursor-pointer transition-colors hover:bg-orange-600 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          {loading ? 'Updating...' : 'Re-Check In'}
        </button>
        <button
          onClick={handleFetchFromAPI}
          disabled={fetching}
          className="bg-green-600 text-white px-5 py-2 rounded-md text-base cursor-pointer transition-colors hover:bg-green-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          {fetching ? 'Fetching...' : '🔄 Fetch from HR API'}
        </button>
      </div>
    </div>
  );
}
