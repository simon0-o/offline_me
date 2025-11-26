'use client';

import type { MonthlyStatsResponse } from '@/lib/types';

interface MonthlyStatsCardProps {
  stats: MonthlyStatsResponse | null;
}

export default function MonthlyStatsCard({ stats }: MonthlyStatsCardProps) {
  if (!stats) {
    return (
      <div className="bg-gray-100 rounded-lg p-5 mb-5">
        <h3 className="text-lg font-semibold mb-3">Monthly Overtime Summary</h3>
        <div>Loading...</div>
      </div>
    );
  }

  function formatOvertimeDisplay(minutes: number) {
    const hours = Math.floor(Math.abs(minutes) / 60);
    const mins = Math.abs(minutes) % 60;
    const timeStr = `${hours}h ${mins}m`;

    if (minutes > 0) {
      return <span className="text-orange-500">+{timeStr}</span>;
    } else if (minutes < 0) {
      return <span className="text-green-600">-{timeStr}</span>;
    } else {
      return <span className="text-green-600">0h 0m</span>;
    }
  }

  return (
    <div className="bg-gray-100 rounded-lg p-5 mb-5">
      <h3 className="text-lg font-semibold mb-3">Monthly Overtime Summary</h3>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        <div>
          <div className="font-semibold mb-2">
            📅 Current Month ({stats.current_month.year_month})
          </div>
          <div className="ml-2">
            <div>Check-in days: {stats.current_month.total_days}</div>
            <div>Checked-out days: {stats.current_month.checked_out_days}</div>
            <div className="font-semibold mt-1">
              Total Overtime: {formatOvertimeDisplay(stats.current_month.overtime_minutes)}
            </div>
          </div>
        </div>
        <div>
          <div className="font-semibold mb-2">
            📅 Last Month ({stats.last_month.year_month})
          </div>
          <div className="ml-2">
            <div>Check-in days: {stats.last_month.total_days}</div>
            <div>Checked-out days: {stats.last_month.checked_out_days}</div>
            <div className="font-semibold mt-1">
              Total Overtime: {formatOvertimeDisplay(stats.last_month.overtime_minutes)}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
