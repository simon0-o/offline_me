import { api } from '@/lib/api';
import HomeClient from '@/components/HomeClient';

export default async function Home() {
  // Fetch data on the server
  const [status, monthlyStats, config] = await Promise.all([
    api.getStatus().catch(() => null),
    api.getMonthlyStats().catch(() => null),
    api.getConfig().catch(() => ({
      work_hours: 480, // Default 8 hours
      auto_fetch_enabled: false,
      check_in_api_url: '',
      authorization: '',
      p_auth: '',
      p_rtoken: '',
      check_in_webhook_url: '',
      check_out_webhook_url: '',
    })),
  ]);

  return (
    <HomeClient
      initialStatus={status}
      initialMonthlyStats={monthlyStats}
      initialConfig={config}
    />
  );
}
