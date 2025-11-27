'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import type { WorkConfig } from '@/lib/types';

interface ConfigSectionProps {
  onConfigUpdate: () => void;
  initialConfig?: WorkConfig;
}

export default function ConfigSection({ onConfigUpdate, initialConfig }: ConfigSectionProps) {
  const [workHours, setWorkHours] = useState(initialConfig ? Math.floor(initialConfig.work_hours / 60) : 8);
  const [workMinutes, setWorkMinutes] = useState(initialConfig ? initialConfig.work_hours % 60 : 0);
  const [autoFetchEnabled, setAutoFetchEnabled] = useState(initialConfig?.auto_fetch_enabled || false);
  const [checkInAPIURL, setCheckInAPIURL] = useState(initialConfig?.check_in_api_url || '');
  const [pAuth, setPAuth] = useState(initialConfig?.p_auth || '');
  const [pRToken, setPRToken] = useState(initialConfig?.p_rtoken || '');
  const [checkInWebhookURL, setCheckInWebhookURL] = useState(initialConfig?.check_in_webhook_url || '');
  const [checkOutWebhookURL, setCheckOutWebhookURL] = useState(initialConfig?.check_out_webhook_url || '');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!initialConfig) {
      loadConfig();
    }
  }, [initialConfig]);

  const loadConfig = async () => {
    try {
      const config = await api.getConfig();
      const totalMinutes = config.work_hours;
      const hours = Math.floor(totalMinutes / 60);
      const minutes = totalMinutes % 60;

      setWorkHours(hours);
      setWorkMinutes(minutes);
      setAutoFetchEnabled(config.auto_fetch_enabled || false);
      setCheckInAPIURL(config.check_in_api_url || '');
      setPAuth(config.p_auth || '');
      setPRToken(config.p_rtoken || '');
      setCheckInWebhookURL(config.check_in_webhook_url || '');
      setCheckOutWebhookURL(config.check_out_webhook_url || '');
    } catch (error) {
      console.error('Failed to load config:', error);
    }
  };

  const handleUpdateConfig = async () => {
    if (workHours < 0 || workHours > 24 || workMinutes < 0 || workMinutes > 59) {
      alert('Please enter valid work time (hours: 0-24, minutes: 0-59)');
      return;
    }

    if (workHours === 0 && workMinutes === 0) {
      alert('Work time must be greater than 0');
      return;
    }

    setLoading(true);
    try {
      await api.updateConfig({
        work_hours: workHours * 60 + workMinutes,
        auto_fetch_enabled: autoFetchEnabled,
        check_in_api_url: checkInAPIURL.trim(),
        p_auth: pAuth.trim(),
        p_rtoken: pRToken.trim(),
        check_in_webhook_url: checkInWebhookURL.trim(),
        check_out_webhook_url: checkOutWebhookURL.trim(),
      });

      alert('Configuration updated successfully!');
      onConfigUpdate();
    } catch (error) {
      console.error('Failed to update config:', error);
      alert('Failed to update configuration');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="border-t border-gray-300 pt-5">
      <h3 className="text-lg font-semibold mb-3">Work Configuration</h3>

      <div className="mb-5">
        <label className="block mb-2 text-sm">Work hours per day:</label>
        <div className="flex flex-wrap items-center gap-2">
          <input
            type="number"
            min="0"
            max="24"
            step="1"
            value={workHours}
            onChange={(e) => setWorkHours(parseInt(e.target.value) || 0)}
            className="w-20 px-3 py-2 border border-gray-300 rounded-md text-base"
          />
          <span className="text-sm">hours</span>
          <input
            type="number"
            min="0"
            max="59"
            step="1"
            value={workMinutes}
            onChange={(e) => setWorkMinutes(parseInt(e.target.value) || 0)}
            className="w-20 px-3 py-2 border border-gray-300 rounded-md text-base"
          />
          <span className="text-sm">minutes</span>
        </div>
      </div>

      <div className="mb-5">
        <h4 className="text-base font-semibold mb-2">HR API Integration</h4>
        <div className="mb-3">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={autoFetchEnabled}
              onChange={(e) => setAutoFetchEnabled(e.target.checked)}
              className="w-4 h-4"
            />
            <span className="text-sm">Enable auto-fetch check-in time from HR API</span>
          </label>
        </div>

        <div className="mb-3">
          <label htmlFor="checkInAPIURL" className="block mb-1 text-sm">
            API URL:
          </label>
          <input
            type="text"
            id="checkInAPIURL"
            value={checkInAPIURL}
            onChange={(e) => setCheckInAPIURL(e.target.value)}
            placeholder="your hr api url"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-base"
          />
        </div>

        <div className="mb-3">
          <label htmlFor="pAuth" className="block mb-1 text-sm">
            P-Auth Token:
          </label>
          <input
            type="text"
            id="pAuth"
            value={pAuth}
            onChange={(e) => setPAuth(e.target.value)}
            placeholder="your p-auth token"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-base"
          />
        </div>

        <div className="mb-3">
          <label htmlFor="pRToken" className="block mb-1 text-sm">
            P-RToken:
          </label>
          <input
            type="text"
            id="pRToken"
            value={pRToken}
            onChange={(e) => setPRToken(e.target.value)}
            placeholder="your p-rtoken"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-base"
          />
        </div>
      </div>

      <div className="mb-5">
        <h4 className="text-base font-semibold mb-2">Webhook Notifications</h4>
        <div className="mb-3">
          <label htmlFor="checkInWebhookURL" className="block mb-1 text-sm">
            Check-in Reminder Webhook (9:55 AM):
          </label>
          <input
            type="text"
            id="checkInWebhookURL"
            value={checkInWebhookURL}
            onChange={(e) => setCheckInWebhookURL(e.target.value)}
            placeholder="https://your-webhook-url.com/checkin"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-base"
          />
          <small className="block text-gray-600 mt-1 text-xs">
            Triggered at 9:55 AM daily if not checked in yet (skips holidays)
          </small>
        </div>

        <div className="mb-3">
          <label htmlFor="checkOutWebhookURL" className="block mb-1 text-sm">
            Check-out Reminder Webhook (8:30 PM & 9:30 PM):
          </label>
          <input
            type="text"
            id="checkOutWebhookURL"
            value={checkOutWebhookURL}
            onChange={(e) => setCheckOutWebhookURL(e.target.value)}
            placeholder="https://your-webhook-url.com/checkout"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-base"
          />
          <small className="block text-gray-600 mt-1 text-xs">
            Triggered at 8:30 PM and 9:30 PM daily if not checked out yet (skips holidays)
          </small>
        </div>
      </div>

      <button
        onClick={handleUpdateConfig}
        disabled={loading}
        className="bg-blue-600 text-white px-5 py-2 rounded-md text-base cursor-pointer transition-colors hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
      >
        {loading ? 'Updating...' : 'Update Configuration'}
      </button>
    </div>
  );
}
