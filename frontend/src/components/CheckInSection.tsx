'use client';

import { useState } from 'react';
import { showNotification, formatDateTime } from '@/lib/utils';
import { api } from '@/lib/api';

interface CheckInSectionProps {
  onSuccess: () => void;
}

export default function CheckInSection({ onSuccess }: CheckInSectionProps) {
  const [checkInTime, setCheckInTime] = useState('');
  const [loading, setLoading] = useState(false);

  const handleCheckIn = async () => {
    if (!checkInTime) {
      alert('Please select a check-in time');
      return;
    }

    setLoading(true);
    try {
      const today = new Date();
      const [hours, minutes] = checkInTime.split(':');
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
        'Check-in Successful!',
        `Expected check-out: ${formatDateTime(result.expected_check_out_time)}`
      );
      onSuccess();
    } catch (error) {
      console.error('Failed to check in:', error);
      alert('Failed to check in');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mb-8">
      <h3 className="text-lg font-semibold mb-3">Check In</h3>
      <div className="flex flex-wrap items-center gap-2">
        <label htmlFor="checkInTime" className="text-sm">
          Check-in time:
        </label>
        <input
          type="time"
          id="checkInTime"
          value={checkInTime}
          onChange={(e) => setCheckInTime(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-base"
        />
        <button
          onClick={handleCheckIn}
          disabled={loading}
          className="bg-blue-600 text-white px-5 py-2 rounded-md text-base cursor-pointer transition-colors hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          {loading ? 'Checking in...' : 'Check In'}
        </button>
      </div>
    </div>
  );
}
