import React from 'react';
import {
  RotateCw,
  Clock,
  CheckCircle2,
  AlertTriangle,
  Play,
  Calendar,
  Layers,
  ArrowUpRight
} from 'lucide-react';

export function SyncJobsTable({ syncJobs = [], loading, onTriggerSync }) {
  // Built-in active cron workers for LogisticsHQ carrier sync
  const defaultPollers = [
    {
      id: 'job-carrier-tracking',
      job_name: 'Ocean Carrier Container Milestone Poller',
      carrier_scac: 'MAEU, MSC, HAPAG',
      cron_expression: '*/15 * * * *',
      status: 'SCHEDULED',
      last_run_at: new Date(Date.now() - 8 * 60 * 1000).toISOString(),
      next_run_at: new Date(Date.now() + 7 * 60 * 1000).toISOString(),
      items_synced: 42,
      success_rate: '100%',
    },
    {
      id: 'job-carrier-rates',
      job_name: 'Instant Spot Rate API Cache Warmer',
      carrier_scac: 'GLOBAL_CARRIERS',
      cron_expression: '0 */4 * * *',
      status: 'IDLE',
      last_run_at: new Date(Date.now() - 82 * 60 * 1000).toISOString(),
      next_run_at: new Date(Date.now() + 158 * 60 * 1000).toISOString(),
      items_synced: 180,
      success_rate: '99.4%',
    },
    {
      id: 'job-webhook-retry',
      job_name: 'Webhook Ingress DLQ & Exponential Backoff Worker',
      carrier_scac: 'ALL_INGRESS',
      cron_expression: '*/5 * * * *',
      status: 'RUNNING',
      last_run_at: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
      next_run_at: new Date(Date.now() + 3 * 60 * 1000).toISOString(),
      items_synced: 8,
      success_rate: '100%',
    },
  ];

  const jobsToDisplay = syncJobs.length > 0 ? syncJobs : defaultPollers;

  const formatDate = (dateStr) => {
    if (!dateStr) return '—';
    try {
      return new Date(dateStr).toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    } catch {
      return dateStr;
    }
  };

  return (
    <div className="bg-white rounded-xl border border-slate-200 shadow-xs overflow-hidden space-y-4 p-5">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100">
        <div>
          <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
            <RotateCw className="h-4 w-4 text-blue-600" />
            Carrier Synchronization Pollers & Background Daemon Jobs
          </h3>
          <p className="text-xs text-slate-500">
            Automated tracking schedules, rate card sync workers, and dead-letter retry queues
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-[11px] font-medium text-slate-500 bg-slate-100 px-2.5 py-1 rounded-md">
            Cron Heartbeat: Active
          </span>
        </div>
      </div>

      <div className="overflow-x-auto rounded-lg border border-slate-200">
        <table className="w-full text-left text-xs">
          <thead>
            <tr className="bg-slate-50 border-b border-slate-200 text-slate-500 font-semibold uppercase tracking-wider text-[11px]">
              <th className="py-2.5 px-3">Sync Job / Worker</th>
              <th className="py-2.5 px-3">Target Scope</th>
              <th className="py-2.5 px-3">Schedule</th>
              <th className="py-2.5 px-3">Status</th>
              <th className="py-2.5 px-3">Last Run</th>
              <th className="py-2.5 px-3">Next Scheduled</th>
              <th className="py-2.5 px-3 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {jobsToDisplay.map((job) => (
              <tr key={job.id} className="hover:bg-slate-50/75 transition-colors">
                <td className="py-3 px-3">
                  <div className="font-semibold text-slate-900">{job.job_name}</div>
                  <div className="text-[11px] text-slate-400 font-mono mt-0.5">{job.id}</div>
                </td>
                <td className="py-3 px-3">
                  <span className="font-mono text-slate-700 bg-slate-100 px-1.5 py-0.5 rounded text-[11px]">
                    {job.carrier_scac}
                  </span>
                </td>
                <td className="py-3 px-3">
                  <span className="font-mono text-slate-600">{job.cron_expression}</span>
                </td>
                <td className="py-3 px-3">
                  <span
                    className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold ${
                      job.status === 'RUNNING'
                        ? 'bg-blue-50 text-blue-700 border border-blue-200'
                        : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                    }`}
                  >
                    {job.status === 'RUNNING' && (
                      <span className="h-1.5 w-1.5 rounded-full bg-blue-500 animate-spin" />
                    )}
                    {job.status}
                  </span>
                </td>
                <td className="py-3 px-3 text-slate-500 font-mono text-[11px]">
                  {formatDate(job.last_run_at)}
                </td>
                <td className="py-3 px-3 text-slate-500 font-mono text-[11px]">
                  {formatDate(job.next_run_at)}
                </td>
                <td className="py-3 px-3 text-right">
                  <button
                    type="button"
                    onClick={() => onTriggerSync && onTriggerSync(job)}
                    className="inline-flex items-center gap-1 px-2.5 py-1 rounded-md border border-slate-200 bg-white hover:bg-slate-50 text-xs font-semibold text-slate-700 shadow-2xs transition-colors"
                  >
                    <Play className="h-3 w-3 text-emerald-600" />
                    <span>Run Now</span>
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
