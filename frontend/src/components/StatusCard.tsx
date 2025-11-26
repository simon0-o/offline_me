'use client';

import { formatDateTime } from '@/lib/utils';
import type { StatusResponse } from '@/lib/types';

interface StatusCardProps {
  status: StatusResponse | null;
}

export default function StatusCard({ status }: StatusCardProps) {
  if (!status) {
    return (
      <div className="bg-gray-100 rounded-lg p-5 mb-5">
        <h3 className="text-lg font-semibold mb-3">Current Status</h3>
        <div>Loading...</div>
      </div>
    );
  }

  const hours = Math.floor(status.work_hours / 60);
  const minutes = status.work_hours % 60;
  const workTimeText = minutes > 0 ? `${hours} hours ${minutes} minutes` : `${hours} hours`;

  if (status.has_checked_in) {
    return (
      <div className="bg-gray-100 rounded-lg p-5 mb-5">
        <h3 className="text-lg font-semibold mb-3">Current Status</h3>
        <div className="text-green-600 font-semibold">✅ Checked in today</div>
        <div className="text-lg font-semibold my-2">
          Check-in: {formatDateTime(status.check_in_time!)}
        </div>

        {status.check_out_time ? (
          <>
            <div className="text-lg font-semibold my-2">
              Last Check-out: {formatDateTime(status.check_out_time)}
            </div>
            {status.overtime_minutes > 0 ? (
              <div className="text-lg font-semibold my-2 text-orange-500">
                ⚠️ Overtime: {Math.floor(Math.abs(status.overtime_minutes) / 60)}h{' '}
                {Math.abs(status.overtime_minutes) % 60}m
              </div>
            ) : status.overtime_minutes < 0 ? (
              <div className="text-lg font-semibold my-2 text-green-600">
                ✅ Under 10 hours by: {Math.floor(Math.abs(status.overtime_minutes) / 60)}h{' '}
                {Math.abs(status.overtime_minutes) % 60}m
              </div>
            ) : (
              <div className="text-lg font-semibold my-2 text-green-600">
                ✅ Exactly 10 hours worked
              </div>
            )}
          </>
        ) : (
          <>
            <div className="text-lg font-semibold my-2">
              Expected check-out: {formatDateTime(status.expected_check_out_time!)}
            </div>
            <div className="text-lg font-semibold my-2">
              Current time: {formatDateTime(status.current_time)}
            </div>
            <div className="text-lg font-semibold my-2">Work hours: {workTimeText}</div>
          </>
        )}
      </div>
    );
  }

  return (
    <div className="bg-gray-100 rounded-lg p-5 mb-5">
      <h3 className="text-lg font-semibold mb-3">Current Status</h3>
      <div className="text-orange-500 font-semibold">
        ⏰ Please check in to start tracking your work time
      </div>
      <div className="text-lg font-semibold my-2">
        Current time: {formatDateTime(status.current_time)}
      </div>
      <div className="text-lg font-semibold my-2">Configured work hours: {workTimeText}</div>
    </div>
  );
}
