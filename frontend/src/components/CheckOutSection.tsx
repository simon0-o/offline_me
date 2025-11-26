'use client';

import { useState, useEffect } from 'react';
import { formatTime, showNotification, formatDateTime } from '@/lib/utils';
import { api } from '@/lib/api';

interface CheckOutSectionProps {
  onSuccess: () => void;
}

export default function CheckOutSection({ onSuccess }: CheckOutSectionProps) {
  const [checkOutTime, setCheckOutTime] = useState('');
  const [manuallyChanged, setManuallyChanged] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!manuallyChanged) {
      setCheckOutTime(formatTime(new Date()));
    }
  }, [manuallyChanged]);

  const handleCheckOut = async () => {
    if (!checkOutTime) {
      alert('Please select a check-out time');
      return;
    }

    setLoading(true);
    try {
      const today = new Date();
      const [hours, minutes] = checkOutTime.split(':');
      const checkOutDate = new Date(
        today.getFullYear(),
        today.getMonth(),
        today.getDate(),
        parseInt(hours),
        parseInt(minutes)
      );

      const result = await api.checkOut({
        check_out_time: checkOutDate.toISOString(),
      });

      const overtimeHours = Math.floor(Math.abs(result.overtime_minutes) / 60);
      const overtimeMinutes = Math.abs(result.overtime_minutes) % 60;

      let message = `Check-out time: ${formatDateTime(result.check_out_time)}`;
      if (result.overtime_minutes > 0) {
        message += `\nOvertime: ${overtimeHours}h ${overtimeMinutes}m`;
      } else if (result.overtime_minutes < 0) {
        message += `\nUnder 10 hours by: ${overtimeHours}h ${overtimeMinutes}m`;
      } else {
        message += `\nExactly 10 hours worked`;
      }

      setManuallyChanged(false);
      showNotification('Check-out Successful!', message);
      onSuccess();
    } catch (error) {
      console.error('Failed to check out:', error);
      alert('Failed to check out');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mb-8">
      <h3 className="text-lg font-semibold mb-3">Check Out</h3>
      <p className="mb-3 text-sm text-gray-600">
        Click to record your check-out time. You can update it multiple times.
      </p>
      <div className="flex flex-wrap items-center gap-2">
        <label htmlFor="checkOutTime" className="text-sm">
          Check-out time:
        </label>
        <input
          type="time"
          id="checkOutTime"
          value={checkOutTime}
          onChange={(e) => {
            setCheckOutTime(e.target.value);
            setManuallyChanged(true);
          }}
          className="px-3 py-2 border border-gray-300 rounded-md text-base"
        />
        <button
          onClick={handleCheckOut}
          disabled={loading}
          className="bg-red-600 text-white px-5 py-2 rounded-md text-base cursor-pointer transition-colors hover:bg-red-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          {loading ? 'Checking out...' : 'Check Out'}
        </button>
      </div>
    </div>
  );
}
