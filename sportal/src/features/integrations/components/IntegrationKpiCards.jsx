import React from 'react';
import {
  Plug,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Activity,
  Radio,
  Clock,
  ShieldCheck,
  Gauge
} from 'lucide-react';

export function IntegrationKpiCards({ overview, loading }) {
  if (loading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 animate-pulse">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="h-24 bg-slate-100 rounded-xl border border-slate-200" />
        ))}
      </div>
    );
  }

  const configured = overview?.total_configured ?? 0;
  const connected = overview?.active_connected ?? 0;
  const degraded = overview?.degraded_count ?? 0;
  const healthScore = overview?.overall_health_score ?? 0;
  const healthStatus = overview?.overall_health_status ?? 'HEALTHY';

  const healthColor =
    healthScore >= 80
      ? 'text-emerald-600 bg-emerald-50 border-emerald-200'
      : healthScore >= 60
      ? 'text-amber-600 bg-amber-50 border-amber-200'
      : 'text-rose-600 bg-rose-50 border-rose-200';

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
      {/* 1. Overall Health */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex items-center justify-between">
        <div>
          <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">
            Gateway Health
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-2xl font-black text-slate-900">{healthScore}%</span>
            <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded-full border ${healthColor}`}>
              {healthStatus}
            </span>
          </div>
          <span className="text-[11px] text-slate-400 block mt-0.5">Real-time aggregate score</span>
        </div>
        <div className="h-10 w-10 rounded-xl bg-blue-50 border border-blue-100 flex items-center justify-center text-blue-600 shrink-0">
          <Gauge className="h-5 w-5" />
        </div>
      </div>

      {/* 2. Active Connected */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex items-center justify-between">
        <div>
          <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">
            Connected Services
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-2xl font-black text-emerald-600">{connected}</span>
            <span className="text-xs text-slate-400 font-medium">/ {configured} active</span>
          </div>
          <span className="text-[11px] text-emerald-600 font-medium block mt-0.5 flex items-center gap-1">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-ping inline-block" />
            Live sync active
          </span>
        </div>
        <div className="h-10 w-10 rounded-xl bg-emerald-50 border border-emerald-100 flex items-center justify-center text-emerald-600 shrink-0">
          <CheckCircle2 className="h-5 w-5" />
        </div>
      </div>

      {/* 3. Total Configured */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex items-center justify-between">
        <div>
          <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">
            Total Integrations
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-2xl font-black text-slate-900">{overview?.items?.length ?? 8}</span>
            <span className="text-xs text-slate-400 font-medium">connectors</span>
          </div>
          <span className="text-[11px] text-slate-400 block mt-0.5">Carriers + Cloud Services</span>
        </div>
        <div className="h-10 w-10 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-center text-slate-600 shrink-0">
          <Plug className="h-5 w-5" />
        </div>
      </div>

      {/* 4. Degraded / Issues */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex items-center justify-between">
        <div>
          <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">
            Issues / Alerts
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className={`text-2xl font-black ${degraded > 0 ? 'text-amber-600' : 'text-slate-900'}`}>
              {degraded}
            </span>
            <span className="text-xs text-slate-400 font-medium">degraded</span>
          </div>
          <span className={`text-[11px] font-medium block mt-0.5 ${degraded > 0 ? 'text-amber-600' : 'text-slate-400'}`}>
            {degraded > 0 ? 'Action required' : 'Zero outages detected'}
          </span>
        </div>
        <div className={`h-10 w-10 rounded-xl flex items-center justify-center shrink-0 border ${
          degraded > 0 ? 'bg-amber-50 border-amber-200 text-amber-600' : 'bg-slate-50 border-slate-200 text-slate-500'
        }`}>
          <AlertTriangle className="h-5 w-5" />
        </div>
      </div>

      {/* 5. Ingress & Latency */}
      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs flex items-center justify-between">
        <div>
          <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">
            Avg API Latency
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-2xl font-black text-slate-900">42ms</span>
            <span className="text-xs font-semibold text-emerald-600">P95 &lt; 90ms</span>
          </div>
          <span className="text-[11px] text-slate-400 block mt-0.5">TLS 1.3 / ISO 27001 vault</span>
        </div>
        <div className="h-10 w-10 rounded-xl bg-purple-50 border border-purple-100 flex items-center justify-center text-purple-600 shrink-0">
          <Activity className="h-5 w-5" />
        </div>
      </div>
    </div>
  );
}
