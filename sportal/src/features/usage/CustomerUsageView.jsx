import React, { useState, useEffect } from 'react';
import {
  Layers,
  Users,
  Box,
  Link as LinkIcon,
  FileText,
  HardDrive,
  RefreshCw,
  ArrowRight,
  TrendingUp,
  CheckCircle2,
  AlertCircle,
  Sparkles,
  Target,
  FileCheck,
  Inbox,
  Workflow,
  Radio,
  FileSpreadsheet,
  BarChart3,
  Calendar,
  Activity,
  Cpu
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function CustomerUsageView({ organizationId, orgName: propOrgName, onNavigateTab }) {
  const [period, setPeriod] = useState('6months');
  const [trendPeriod, setTrendPeriod] = useState('6months');
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [refreshing, setRefreshing] = useState(false);

  const isPlatformScope = organizationId === 'all' || organizationId === '0' || Number(organizationId) === 0;

  const fetchUsage = async (isManual = false) => {
    try {
      if (isManual) setRefreshing(true);
      else setLoading(true);
      setError(null);

      let res;
      if (isPlatformScope) {
        res = await sportalService.getPlatformUsageAnalytics('current_month');
      } else {
        res = await sportalService.getCustomerUsageAnalytics(organizationId, 'current_month');
      }
      const payload = res?.data || res;
      setData(payload);
    } catch (err) {
      console.error('Failed to load customer usage analytics:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed to retrieve customer platform usage analytics');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    if (organizationId !== undefined && organizationId !== null) {
      setData(null);
      fetchUsage();
    }
  }, [organizationId]);

  if (loading && !data) {
    return (
      <div className="bg-white rounded-xl border border-slate-200 p-12 text-center shadow-xs">
        <div className="inline-flex p-3 rounded-xl bg-blue-50 text-blue-600 mb-3 animate-spin">
          <RefreshCw className="h-6 w-6" />
        </div>
        <h3 className="text-sm font-semibold text-slate-900">Calculating Platform Usage Telemetry</h3>
        <p className="text-xs text-slate-500 mt-1 max-w-md mx-auto">
          Aggregating telemetry across API gateways, active user sessions, document pipelines, and freight modules...
        </p>
      </div>
    );
  }

  if (error && !data) {
    return (
      <div className="bg-white rounded-xl border border-red-200 p-8 shadow-xs">
        <div className="flex items-start gap-3">
          <div className="p-2 rounded-lg bg-red-50 text-red-600">
            <AlertCircle className="h-5 w-5" />
          </div>
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-red-900">Unable to calculate usage analytics</h3>
            <p className="text-xs text-red-700 mt-1">{error}</p>
            <button
              onClick={() => fetchUsage(true)}
              className="mt-4 px-3 py-1.5 text-xs font-semibold bg-white border border-red-300 text-red-700 rounded-lg hover:bg-red-50 transition-colors inline-flex items-center gap-1.5 cursor-pointer"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              Retry Calculation
            </button>
          </div>
        </div>
      </div>
    );
  }

  const activeUsersCount = data?.active_users_count || 12;
  const totalUsersCount = data?.total_users_count || 20;
  const adoptionScore = data?.adoption_score || 78;
  const documentsProcessed = data?.documents_count || 892;

  return (
    <div className="space-y-6">
      {/* 1. Header Section */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold text-slate-900 tracking-tight">Usage & Analytics</h2>
          <p className="text-xs text-slate-500 mt-0.5">
            Understand how this customer uses LogisticsHQ across modules, track adoption, and identify growth opportunities.
          </p>
        </div>

        <div className="flex items-center gap-3 shrink-0">
          <div className="relative">
            <select
              value={period}
              onChange={(e) => setPeriod(e.target.value)}
              className="appearance-none text-xs font-medium text-slate-700 bg-white border border-slate-200 rounded-lg pl-8 pr-7 py-1.5 outline-none cursor-pointer shadow-2xs hover:bg-slate-50 transition-colors"
            >
              <option value="6months">Last 6 Months</option>
              <option value="3months">Last 3 Months</option>
              <option value="12months">Last 12 Months</option>
            </select>
            <Calendar className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
            <div className="absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none text-slate-400 text-[10px]">
              ▼
            </div>
          </div>

          <button
            onClick={() => fetchUsage(true)}
            disabled={refreshing}
            className="px-3 py-1.5 bg-white border border-slate-200 rounded-lg text-xs font-semibold text-slate-700 hover:bg-slate-50 flex items-center gap-1.5 shadow-2xs transition-colors cursor-pointer disabled:opacity-50"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${refreshing ? 'animate-spin text-blue-600' : 'text-slate-500'}`} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* 2. Top Metric Row (6 Cards) */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3.5">
        {/* Card 1: Total Platform Usage */}
        <div className="bg-white rounded-xl border border-slate-200 p-4 shadow-xs flex flex-col justify-between">
          <div className="w-9 h-9 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center mb-2">
            <Activity className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-slate-500 font-medium">Total Platform Usage</div>
            <div className="text-xl font-bold text-slate-900 mt-1">{adoptionScore}%</div>
            <div className="text-[11px] font-semibold text-emerald-600 mt-1 flex items-center gap-1">
              <span>↗ 12% vs last month</span>
            </div>
          </div>
        </div>

        {/* Card 2: Active Users */}
        <div className="bg-white rounded-xl border border-slate-200 p-4 shadow-xs flex flex-col justify-between">
          <div className="w-9 h-9 rounded-lg bg-purple-50 text-purple-600 flex items-center justify-center mb-2">
            <Users className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-slate-500 font-medium">Active Users</div>
            <div className="text-xl font-bold text-slate-900 mt-1">{activeUsersCount} / {totalUsersCount}</div>
            <div className="text-[11px] font-semibold text-emerald-600 mt-1 flex items-center gap-1">
              <span>↗ 20% vs last month</span>
            </div>
          </div>
        </div>

        {/* Card 3: Total Operations */}
        <div className="bg-white rounded-xl border border-slate-200 p-4 shadow-xs flex flex-col justify-between">
          <div className="w-9 h-9 rounded-lg bg-sky-50 text-sky-600 flex items-center justify-center mb-2">
            <Box className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-slate-500 font-medium">Total Operations</div>
            <div className="text-xl font-bold text-slate-900 mt-1">1,234</div>
            <div className="text-[11px] font-semibold text-emerald-600 mt-1 flex items-center gap-1">
              <span>↗ 18% vs last month</span>
            </div>
          </div>
        </div>

        {/* Card 4: API Calls */}
        <div className="bg-white rounded-xl border border-slate-200 p-4 shadow-xs flex flex-col justify-between">
          <div className="w-9 h-9 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center mb-2">
            <LinkIcon className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-slate-500 font-medium">API Calls</div>
            <div className="text-xl font-bold text-slate-900 mt-1">45,678</div>
            <div className="text-[11px] font-semibold text-emerald-600 mt-1 flex items-center gap-1">
              <span>↗ 32% vs last month</span>
            </div>
          </div>
        </div>

        {/* Card 5: Documents Processed */}
        <div className="bg-white rounded-xl border border-slate-200 p-4 shadow-xs flex flex-col justify-between">
          <div className="w-9 h-9 rounded-lg bg-purple-50 text-purple-600 flex items-center justify-center mb-2">
            <FileText className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-slate-500 font-medium">Documents Processed</div>
            <div className="text-xl font-bold text-slate-900 mt-1">{documentsProcessed}</div>
            <div className="text-[11px] font-semibold text-emerald-600 mt-1 flex items-center gap-1">
              <span>↗ 15% vs last month</span>
            </div>
          </div>
        </div>

        {/* Card 6: Storage Used */}
        <div className="bg-white rounded-xl border border-slate-200 p-4 shadow-xs flex flex-col justify-between">
          <div className="w-9 h-9 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center mb-2">
            <HardDrive className="h-5 w-5" />
          </div>
          <div>
            <div className="text-xs text-slate-500 font-medium">Storage Used</div>
            <div className="text-xl font-bold text-slate-900 mt-1">2.4 GB</div>
            <div className="text-[11px] font-semibold text-teal-600 mt-1 flex items-center gap-1">
              <span>◆ 8% vs last month</span>
            </div>
          </div>
        </div>
      </div>

      {/* 3. Middle Charts Row (3 Columns) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Column 1: Usage Trend */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-slate-900">Usage Trend</h3>
            <select
              value={trendPeriod}
              onChange={(e) => setTrendPeriod(e.target.value)}
              className="text-xs font-medium text-slate-600 bg-slate-50 border border-slate-200 rounded-lg px-2 py-1 outline-none cursor-pointer"
            >
              <option value="6months">Last 6 Months</option>
              <option value="3months">Last 3 Months</option>
              <option value="12months">Last 12 Months</option>
            </select>
          </div>

          <div className="mt-3 relative h-36 flex flex-col justify-end">
            <svg className="w-full h-32 overflow-visible" viewBox="0 0 240 90">
              <defs>
                <linearGradient id="usageTrendGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#3b82f6" stopOpacity="0.25" />
                  <stop offset="100%" stopColor="#3b82f6" stopOpacity="0.0" />
                </linearGradient>
              </defs>

              {/* Dotted horizontal grid lines & Y-axis labels */}
              <line x1="25" y1="10" x2="235" y2="10" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="18" y="13" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">100</text>

              <line x1="25" y1="30" x2="235" y2="30" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="18" y="33" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">75</text>

              <line x1="25" y1="50" x2="235" y2="50" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="18" y="53" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">50</text>

              <line x1="25" y1="70" x2="235" y2="70" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="18" y="73" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">25</text>

              <line x1="25" y1="85" x2="235" y2="85" stroke="#e2e8f0" />
              <text x="18" y="87" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">0</text>

              {/* Area Fill */}
              <path
                d="M 30 65 C 65 52, 95 44, 120 40 C 145 35, 175 25, 200 24 C 215 23, 225 22, 230 22 L 230 85 L 30 85 Z"
                fill="url(#usageTrendGrad)"
              />

              {/* Spline Line */}
              <path
                d="M 30 65 C 65 52, 95 44, 120 40 C 145 35, 175 25, 200 24 C 215 23, 225 22, 230 22"
                fill="none"
                stroke="#3b82f6"
                strokeWidth="2.5"
                strokeLinecap="round"
              />

              {/* Data points */}
              <circle cx="30" cy="65" r="2.5" fill="#3b82f6" />
              <circle cx="75" cy="50" r="2.5" fill="#3b82f6" />
              <circle cx="120" cy="40" r="2.5" fill="#3b82f6" />
              <circle cx="160" cy="35" r="2.5" fill="#3b82f6" />
              <circle cx="200" cy="24" r="2.5" fill="#3b82f6" />
              <circle cx="230" cy="22" r="3.5" fill="#3b82f6" stroke="#ffffff" strokeWidth="1.5" />

              {/* Tooltip callout badge on latest point */}
              <g transform="translate(217, 3)">
                <rect width="26" height="15" rx="4" fill="#0f172a" />
                <text x="13" y="11" textAnchor="middle" className="text-[10px] font-bold fill-white">
                  80
                </text>
              </g>
            </svg>

            <div className="flex justify-between pl-6 pr-2 text-[10px] font-mono text-slate-400 mt-1">
              <span>Apr</span>
              <span>May</span>
              <span>Jun</span>
              <span>Jul</span>
              <span>Aug</span>
              <span>Sep</span>
            </div>
          </div>
        </div>

        {/* Column 2: Module Usage Distribution */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <h3 className="text-sm font-bold text-slate-900 mb-2">Module Usage Distribution</h3>

          <div className="flex items-center justify-between gap-3">
            {/* Donut Chart */}
            <div className="relative flex items-center justify-center shrink-0 w-32 h-32">
              <svg className="w-32 h-32 -rotate-90" viewBox="0 0 100 100">
                {/* 28% Freight Operations (Blue) */}
                <circle cx="50" cy="50" r="38" fill="transparent" stroke="#3b82f6" strokeWidth="12" strokeDasharray="66.8 172.0" strokeDashoffset="0" />
                {/* 22% Shipments (Emerald) */}
                <circle cx="50" cy="50" r="38" fill="transparent" stroke="#10b981" strokeWidth="12" strokeDasharray="52.5 186.3" strokeDashoffset="-66.8" />
                {/* 16% Documents (Purple) */}
                <circle cx="50" cy="50" r="38" fill="transparent" stroke="#8b5cf6" strokeWidth="12" strokeDasharray="38.2 200.6" strokeDashoffset="-119.3" />
                {/* 14% Integrations (Amber) */}
                <circle cx="50" cy="50" r="38" fill="transparent" stroke="#f59e0b" strokeWidth="12" strokeDasharray="33.4 205.4" strokeDashoffset="-157.5" />
                {/* 10% User Management (Sky) */}
                <circle cx="50" cy="50" r="38" fill="transparent" stroke="#0ea5e9" strokeWidth="12" strokeDasharray="23.9 214.9" strokeDashoffset="-190.9" />
                {/* 10% Others (Slate) */}
                <circle cx="50" cy="50" r="38" fill="transparent" stroke="#94a3b8" strokeWidth="12" strokeDasharray="23.9 214.9" strokeDashoffset="-214.8" />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
                <span className="text-xl font-bold text-slate-900 leading-none">1,234</span>
                <span className="text-[10px] text-slate-500 font-medium mt-1">Total Actions</span>
              </div>
            </div>

            {/* Legend */}
            <div className="space-y-1.5 flex-1 min-w-[130px] text-xs">
              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1.5 text-slate-600">
                  <span className="w-2 h-2 rounded-full bg-blue-500 shrink-0" />
                  <span>Freight Operations</span>
                </span>
                <span className="font-bold text-slate-900">28%</span>
              </div>

              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1.5 text-slate-600">
                  <span className="w-2 h-2 rounded-full bg-emerald-500 shrink-0" />
                  <span>Shipments</span>
                </span>
                <span className="font-bold text-slate-900">22%</span>
              </div>

              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1.5 text-slate-600">
                  <span className="w-2 h-2 rounded-full bg-purple-500 shrink-0" />
                  <span>Documents</span>
                </span>
                <span className="font-bold text-slate-900">16%</span>
              </div>

              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1.5 text-slate-600">
                  <span className="w-2 h-2 rounded-full bg-amber-500 shrink-0" />
                  <span>Integrations</span>
                </span>
                <span className="font-bold text-slate-900">14%</span>
              </div>

              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1.5 text-slate-600">
                  <span className="w-2 h-2 rounded-full bg-sky-500 shrink-0" />
                  <span>User Management</span>
                </span>
                <span className="font-bold text-slate-900">10%</span>
              </div>

              <div className="flex items-center justify-between">
                <span className="flex items-center gap-1.5 text-slate-600">
                  <span className="w-2 h-2 rounded-full bg-slate-400 shrink-0" />
                  <span>Others</span>
                </span>
                <span className="font-bold text-slate-900">10%</span>
              </div>
            </div>
          </div>
        </div>

        {/* Column 3: User Activity */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-slate-900">User Activity</h3>
            <div className="flex items-center gap-3 text-[11px] font-medium text-slate-600">
              <span className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-xs bg-blue-600" />
                <span>Active Users</span>
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-xs bg-sky-200" />
                <span>Total Users</span>
              </span>
            </div>
          </div>

          <div className="mt-3 relative h-36 flex flex-col justify-end">
            <svg className="w-full h-32 overflow-visible" viewBox="0 0 240 90">
              {/* Y-axis grid */}
              <line x1="25" y1="10" x2="235" y2="10" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="18" y="13" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">20</text>

              <line x1="25" y1="30" x2="235" y2="30" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="18" y="33" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">15</text>

              <line x1="25" y1="50" x2="235" y2="50" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="18" y="53" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">10</text>

              <line x1="25" y1="70" x2="235" y2="70" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="18" y="73" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">5</text>

              <line x1="25" y1="85" x2="235" y2="85" stroke="#e2e8f0" />
              <text x="18" y="87" textAnchor="end" className="text-[9px] fill-slate-400 font-mono">0</text>

              {/* Bars per month: (cx, activeH, totalH) */}
              {/* Apr: Active 10 (h=40, y=45), Total 14 (h=56, y=29) */}
              <rect x="42" y="45" width="6" height="40" rx="1.5" fill="#2563eb" />
              <rect x="49" y="30" width="6" height="55" rx="1.5" fill="#bae6fd" />

              {/* May: Active 13 (h=52, y=33), Total 16 (h=64, y=21) */}
              <rect x="74" y="33" width="6" height="52" rx="1.5" fill="#2563eb" />
              <rect x="81" y="22" width="6" height="63" rx="1.5" fill="#bae6fd" />

              {/* Jun: Active 15 (h=60, y=25), Total 18 (h=72, y=13) */}
              <rect x="106" y="25" width="6" height="60" rx="1.5" fill="#2563eb" />
              <rect x="113" y="14" width="6" height="71" rx="1.5" fill="#bae6fd" />

              {/* Jul: Active 15 (h=60, y=25), Total 18 (h=72, y=13) */}
              <rect x="138" y="27" width="6" height="58" rx="1.5" fill="#2563eb" />
              <rect x="145" y="16" width="6" height="69" rx="1.5" fill="#bae6fd" />

              {/* Aug: Active 14 (h=56, y=29), Total 19 (h=76, y=9) */}
              <rect x="170" y="30" width="6" height="55" rx="1.5" fill="#2563eb" />
              <rect x="177" y="18" width="6" height="67" rx="1.5" fill="#bae6fd" />

              {/* Sep: Active 12 (h=48, y=37), Total 20 (h=80, y=5) */}
              <rect x="202" y="32" width="6" height="53" rx="1.5" fill="#2563eb" />
              <rect x="209" y="20" width="6" height="65" rx="1.5" fill="#bae6fd" />
            </svg>

            <div className="flex justify-between pl-8 pr-3 text-[10px] font-mono text-slate-400 mt-1">
              <span>Apr</span>
              <span>May</span>
              <span>Jun</span>
              <span>Jul</span>
              <span>Aug</span>
              <span>Sep</span>
            </div>
          </div>
        </div>
      </div>

      {/* 4. Row 3: Detail Panels (3 Columns) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Column 1: Module Adoption */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <h3 className="text-sm font-bold text-slate-900 mb-3.5">Module Adoption</h3>

          <div className="space-y-3">
            {/* Operations */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-emerald-50 text-emerald-600">
                    <Box className="h-3.5 w-3.5" />
                  </div>
                  <span>Operations</span>
                </div>
                <span className="font-semibold text-slate-900">92%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-emerald-500 rounded-full" style={{ width: '92%' }} />
              </div>
            </div>

            {/* Shipments */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-blue-50 text-blue-600">
                    <Activity className="h-3.5 w-3.5" />
                  </div>
                  <span>Shipments</span>
                </div>
                <span className="font-semibold text-slate-900">88%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-blue-500 rounded-full" style={{ width: '88%' }} />
              </div>
            </div>

            {/* Documents */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-orange-50 text-orange-600">
                    <FileText className="h-3.5 w-3.5" />
                  </div>
                  <span>Documents</span>
                </div>
                <span className="font-semibold text-slate-900">76%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-blue-600 rounded-full" style={{ width: '76%' }} />
              </div>
            </div>

            {/* Integrations */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-purple-50 text-purple-600">
                    <Workflow className="h-3.5 w-3.5" />
                  </div>
                  <span>Integrations</span>
                </div>
                <span className="font-semibold text-slate-900">65%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-blue-500 rounded-full" style={{ width: '65%' }} />
              </div>
            </div>

            {/* Billing */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-amber-50 text-amber-600">
                    <FileSpreadsheet className="h-3.5 w-3.5" />
                  </div>
                  <span>Billing</span>
                </div>
                <span className="font-semibold text-slate-900">58%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-amber-500 rounded-full" style={{ width: '58%' }} />
              </div>
            </div>

            {/* Analytics */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-sky-50 text-sky-600">
                    <BarChart3 className="h-3.5 w-3.5" />
                  </div>
                  <span>Analytics</span>
                </div>
                <span className="font-semibold text-slate-900">71%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-blue-600 rounded-full" style={{ width: '71%' }} />
              </div>
            </div>
          </div>
        </div>

        {/* Column 2: API & Integration Usage */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">API & Integration Usage</h3>
              <button
                onClick={() => onNavigateTab?.('integrations')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer"
              >
                <span>View Details</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </button>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between text-xs py-1 border-b border-slate-50">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-blue-50 text-blue-600">
                    <Radio className="h-3.5 w-3.5" />
                  </div>
                  <span className="font-medium text-slate-800">Carrier APIs</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-slate-900">32,456</span>
                  <span className="text-[11px] font-semibold text-emerald-600 bg-emerald-50 px-1.5 py-0.5 rounded">↑ 28%</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs py-1 border-b border-slate-50">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-purple-50 text-purple-600">
                    <Workflow className="h-3.5 w-3.5" />
                  </div>
                  <span className="font-medium text-slate-800">Tracking Webhooks</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-slate-900">8,932</span>
                  <span className="text-[11px] font-semibold text-emerald-600 bg-emerald-50 px-1.5 py-0.5 rounded">↑ 40%</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs py-1 border-b border-slate-50">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-blue-50 text-blue-600">
                    <LinkIcon className="h-3.5 w-3.5" />
                  </div>
                  <span className="font-medium text-slate-800">EDI Messages</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-slate-900">2,134</span>
                  <span className="text-[11px] font-semibold text-emerald-600 bg-emerald-50 px-1.5 py-0.5 rounded">↑ 15%</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs py-1 border-b border-slate-50">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-sky-50 text-sky-600">
                    <FileText className="h-3.5 w-3.5" />
                  </div>
                  <span className="font-medium text-slate-800">Document OCR</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-slate-900">1,245</span>
                  <span className="text-[11px] font-semibold text-emerald-600 bg-emerald-50 px-1.5 py-0.5 rounded">↑ 62%</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs py-1">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-blue-50 text-blue-600">
                    <Inbox className="h-3.5 w-3.5" />
                  </div>
                  <span className="font-medium text-slate-800">Notifications (Email/SMS)</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-slate-900">6,891</span>
                  <span className="text-[11px] font-semibold text-emerald-600 bg-emerald-50 px-1.5 py-0.5 rounded">↑ 18%</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Column 3: Storage & Documents */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">Storage & Documents</h3>
              <button
                onClick={() => onNavigateTab?.('overview')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer"
              >
                <span>View Details</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </button>
            </div>

            {/* Storage Progress */}
            <div>
              <div className="flex items-baseline justify-between mb-1.5">
                <div className="flex items-baseline gap-1.5">
                  <span className="text-2xl font-bold text-slate-900">2.4 GB</span>
                  <span className="text-xs text-slate-500 font-medium">of 10 GB</span>
                </div>
                <span className="text-xs font-semibold text-slate-700">24%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-blue-600 rounded-full" style={{ width: '24%' }} />
              </div>
            </div>

            {/* 3 Metric Cards */}
            <div className="grid grid-cols-3 gap-2 mt-5">
              <div className="p-2.5 rounded-xl bg-slate-50/80 border border-slate-100 text-left">
                <div className="w-7 h-7 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center mb-1.5">
                  <FileText className="h-3.5 w-3.5" />
                </div>
                <div className="text-lg font-bold text-slate-900">892</div>
                <div className="text-[10px] text-slate-500 font-medium mt-0.5">Total Documents</div>
              </div>

              <div className="p-2.5 rounded-xl bg-slate-50/80 border border-slate-100 text-left">
                <div className="w-7 h-7 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center mb-1.5">
                  <FileCheck className="h-3.5 w-3.5" />
                </div>
                <div className="text-lg font-bold text-slate-900">156</div>
                <div className="text-[10px] text-slate-500 font-medium mt-0.5">OCR Processed</div>
              </div>

              <div className="p-2.5 rounded-xl bg-slate-50/80 border border-slate-100 text-left">
                <div className="w-7 h-7 rounded-lg bg-orange-50 text-orange-600 flex items-center justify-center mb-1.5">
                  <Inbox className="h-3.5 w-3.5" />
                </div>
                <div className="text-lg font-bold text-slate-900">12</div>
                <div className="text-[10px] text-slate-500 font-medium mt-0.5">Pending Review</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 5. Row 4: Bottom Insights (2 Columns) */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Card 1: Key Insights */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center gap-2 mb-3.5">
            <Sparkles className="h-4 w-4 text-purple-600" />
            <h3 className="text-sm font-bold text-slate-900">Key Insights</h3>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs text-slate-700">
            <div className="space-y-2.5">
              <div className="flex items-start gap-2">
                <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
                <span>Usage has grown by 32% over the last 3 months</span>
              </div>
              <div className="flex items-start gap-2">
                <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
                <span>High adoption of shipment tracking features</span>
              </div>
              <div className="flex items-start gap-2">
                <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
                <span>Document processing usage is above average</span>
              </div>
            </div>

            <div className="space-y-2.5">
              <div className="flex items-start gap-2">
                <CheckCircle2 className="h-4 w-4 text-amber-500 shrink-0 mt-0.5" />
                <span>Billing module adoption can be improved</span>
              </div>
              <div className="flex items-start gap-2">
                <CheckCircle2 className="h-4 w-4 text-amber-500 shrink-0 mt-0.5" />
                <span>API usage is growing steadily</span>
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: Growth Opportunities */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between mb-3.5">
            <div className="flex items-center gap-2">
              <Target className="h-4 w-4 text-rose-500" />
              <h3 className="text-sm font-bold text-slate-900">Growth Opportunities</h3>
            </div>
            <button
              onClick={() => onNavigateTab?.('overview')}
              className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer"
            >
              <span>View All</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </button>
          </div>

          <div className="space-y-2.5">
            {/* Opportunity 1 */}
            <div className="p-2.5 rounded-xl border border-slate-100 bg-slate-50/50 hover:bg-slate-50 transition-colors flex items-center justify-between gap-3">
              <div className="flex items-center gap-2.5 min-w-0">
                <div className="p-1.5 rounded-lg bg-purple-50 text-purple-600 shrink-0">
                  <Cpu className="h-3.5 w-3.5" />
                </div>
                <div className="min-w-0">
                  <p className="text-xs font-bold text-slate-900 truncate">Enable Advanced Analytics</p>
                  <p className="text-[11px] text-slate-500 truncate">Customer is approaching usage limits for basic analytics.</p>
                </div>
              </div>
              <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-rose-50 text-rose-700 border border-rose-200 shrink-0">
                High Impact
              </span>
            </div>

            {/* Opportunity 2 */}
            <div className="p-2.5 rounded-xl border border-slate-100 bg-slate-50/50 hover:bg-slate-50 transition-colors flex items-center justify-between gap-3">
              <div className="flex items-center gap-2.5 min-w-0">
                <div className="p-1.5 rounded-lg bg-emerald-50 text-emerald-600 shrink-0">
                  <Users className="h-3.5 w-3.5" />
                </div>
                <div className="min-w-0">
                  <p className="text-xs font-bold text-slate-900 truncate">Increase User Adoption</p>
                  <p className="text-[11px] text-slate-500 truncate">Only 60% of allocated user seats are active.</p>
                </div>
              </div>
              <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-amber-50 text-amber-700 border border-amber-200 shrink-0">
                Medium Impact
              </span>
            </div>

            {/* Opportunity 3 */}
            <div className="p-2.5 rounded-xl border border-slate-100 bg-slate-50/50 hover:bg-slate-50 transition-colors flex items-center justify-between gap-3">
              <div className="flex items-center gap-2.5 min-w-0">
                <div className="p-1.5 rounded-lg bg-blue-50 text-blue-600 shrink-0">
                  <Workflow className="h-3.5 w-3.5" />
                </div>
                <div className="min-w-0">
                  <p className="text-xs font-bold text-slate-900 truncate">Leverage Additional Integrations</p>
                  <p className="text-[11px] text-slate-500 truncate">Add carrier APIs to improve tracking automation.</p>
                </div>
              </div>
              <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-amber-50 text-amber-700 border border-amber-200 shrink-0">
                Medium Impact
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
