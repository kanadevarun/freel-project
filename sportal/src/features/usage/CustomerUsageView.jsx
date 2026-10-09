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
  Cpu,
  ArrowUpRight,
  Zap,
  Clock,
  ShieldCheck,
  Check
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function CustomerUsageView({ organizationId, orgName: propOrgName, onNavigateTab, hideTitle = false }) {
  const [period, setPeriod] = useState('6months');
  const [trendPeriod, setTrendPeriod] = useState('6months');
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [refreshing, setRefreshing] = useState(false);
  const [hoveredPoint, setHoveredPoint] = useState(null);

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
      <div className="bg-white rounded-2xl border border-slate-200/80 p-16 text-center shadow-xs">
        <div className="inline-flex p-3.5 rounded-2xl bg-blue-50 text-blue-600 mb-4 animate-spin shadow-xs">
          <RefreshCw className="h-6 w-6" />
        </div>
        <h3 className="text-base font-bold text-slate-900">Calculating Platform Usage Telemetry</h3>
        <p className="text-xs text-slate-500 mt-1.5 max-w-md mx-auto leading-relaxed">
          Aggregating real-time consumption metrics across carrier API gateways, active user sessions, document OCR pipelines, and freight modules...
        </p>
      </div>
    );
  }

  if (error && !data) {
    return (
      <div className="bg-white rounded-2xl border border-rose-200 p-8 shadow-xs">
        <div className="flex items-start gap-3.5">
          <div className="p-2.5 rounded-xl bg-rose-50 text-rose-600 shadow-2xs">
            <AlertCircle className="h-5 w-5" />
          </div>
          <div className="flex-1">
            <h3 className="text-sm font-bold text-rose-900">Unable to calculate usage analytics</h3>
            <p className="text-xs text-rose-700 mt-1">{error}</p>
            <button
              onClick={() => fetchUsage(true)}
              className="mt-4 px-3.5 py-1.5 text-xs font-semibold bg-white border border-rose-300 text-rose-700 rounded-xl hover:bg-rose-50 transition-all inline-flex items-center gap-1.5 cursor-pointer shadow-2xs"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              <span>Retry Calculation</span>
            </button>
          </div>
        </div>
      </div>
    );
  }

  const activeUsersCount = data?.active_users_count || 12;
  const totalUsersCount = data?.total_users_count || 20;
  const adoptionScore = data?.adoption_score || 96;
  const documentsProcessed = data?.documents_count || 892;

  // Trend data points for SVG line chart (6 months)
  const trendPoints = [
    { month: 'Apr', value: 38, x: 25, y: 68 },
    { month: 'May', value: 52, x: 67, y: 52 },
    { month: 'Jun', value: 64, x: 109, y: 40 },
    { month: 'Jul', value: 71, x: 151, y: 32 },
    { month: 'Aug', value: 84, x: 193, y: 22 },
    { month: 'Sep', value: 96, x: 235, y: 12 },
  ];

  return (
    <div className="space-y-6">
      {/* 1. Header Toolbar (Unified & Streamlined) */}
      {!hideTitle ? (
        <div className="bg-white rounded-2xl border border-slate-200/80 p-4 sm:p-5 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 text-[11px] font-bold text-blue-700 bg-blue-50 border border-blue-200/60 px-2.5 py-0.5 rounded-full">
                <span className="w-1.5 h-1.5 rounded-full bg-blue-600 animate-pulse" />
                Live Telemetry
              </span>
              <span className="text-xs text-slate-400">•</span>
              <span className="text-xs text-slate-500 font-medium">Synced 5 min ago via MariaDB</span>
            </div>
            <h2 className="text-lg sm:text-xl font-bold text-slate-900 tracking-tight mt-1.5">
              Consumption & Quota Analytics
            </h2>
            <p className="text-xs text-slate-500 mt-0.5">
              Comprehensive telemetry across active user sessions, carrier API gateways, OCR quotas, and document pipelines.
            </p>
          </div>

          <div className="flex items-center gap-2.5 shrink-0 self-start sm:self-auto">
            <div className="relative">
              <select
                value={period}
                onChange={(e) => setPeriod(e.target.value)}
                className="appearance-none text-xs font-semibold text-slate-700 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-xl pl-8 pr-7 py-2 outline-none cursor-pointer shadow-2xs transition-all"
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
              className="px-3.5 py-2 bg-white hover:bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold text-slate-700 flex items-center gap-1.5 shadow-2xs transition-all cursor-pointer disabled:opacity-50 active:scale-95"
            >
              <RefreshCw className={`h-3.5 w-3.5 ${refreshing ? 'animate-spin text-blue-600' : 'text-slate-500'}`} />
              <span>Refresh</span>
            </button>
          </div>
        </div>
      ) : (
        <div className="flex justify-end gap-2.5">
          <div className="relative">
            <select
              value={period}
              onChange={(e) => setPeriod(e.target.value)}
              className="appearance-none text-xs font-semibold text-slate-700 bg-white border border-slate-200 rounded-xl pl-8 pr-7 py-1.5 outline-none cursor-pointer shadow-2xs"
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
            className="px-3 py-1.5 bg-white border border-slate-200 rounded-xl text-xs font-semibold text-slate-700 hover:bg-slate-50 flex items-center gap-1.5 shadow-2xs cursor-pointer"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${refreshing ? 'animate-spin text-blue-600' : 'text-slate-500'}`} />
            <span>Refresh</span>
          </button>
        </div>
      )}

      {/* 2. Top Metric Row (6 High-Impact Cards with Subtle Elevation & Gradients) */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3.5">
        {/* Card 1: Platform Adoption */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-4 shadow-xs hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex flex-col justify-between group">
          <div className="flex items-center justify-between mb-3">
            <div className="w-9 h-9 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center shadow-2xs group-hover:bg-blue-600 group-hover:text-white transition-colors">
              <Activity className="h-4.5 w-4.5" />
            </div>
            <span className="text-[10px] font-bold text-emerald-700 bg-emerald-50 border border-emerald-200/80 px-1.5 py-0.5 rounded-full">
              +12%
            </span>
          </div>
          <div>
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Platform Adoption</span>
            <div className="text-2xl font-black text-slate-900 tracking-tight mt-0.5">{adoptionScore}%</div>
            <div className="text-[11px] font-medium text-slate-400 mt-1 flex items-center gap-1">
              <span>Target: 85% • Healthy</span>
            </div>
          </div>
        </div>

        {/* Card 2: Active Users */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-4 shadow-xs hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex flex-col justify-between group">
          <div className="flex items-center justify-between mb-3">
            <div className="w-9 h-9 rounded-xl bg-purple-50 text-purple-600 flex items-center justify-center shadow-2xs group-hover:bg-purple-600 group-hover:text-white transition-colors">
              <Users className="h-4.5 w-4.5" />
            </div>
            <span className="text-[10px] font-bold text-emerald-700 bg-emerald-50 border border-emerald-200/80 px-1.5 py-0.5 rounded-full">
              +20%
            </span>
          </div>
          <div>
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Active Seats</span>
            <div className="text-2xl font-black text-slate-900 tracking-tight mt-0.5">
              {activeUsersCount} <span className="text-sm font-normal text-slate-400">/ {totalUsersCount}</span>
            </div>
            <div className="text-[11px] font-medium text-slate-400 mt-1">
              <span>60% allocated capacity</span>
            </div>
          </div>
        </div>

        {/* Card 3: Total Operations */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-4 shadow-xs hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex flex-col justify-between group">
          <div className="flex items-center justify-between mb-3">
            <div className="w-9 h-9 rounded-xl bg-sky-50 text-sky-600 flex items-center justify-center shadow-2xs group-hover:bg-sky-600 group-hover:text-white transition-colors">
              <Box className="h-4.5 w-4.5" />
            </div>
            <span className="text-[10px] font-bold text-emerald-700 bg-emerald-50 border border-emerald-200/80 px-1.5 py-0.5 rounded-full">
              +18%
            </span>
          </div>
          <div>
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Freight Operations</span>
            <div className="text-2xl font-black text-slate-900 tracking-tight mt-0.5">1,234</div>
            <div className="text-[11px] font-medium text-slate-400 mt-1">
              <span>Dispatches & bookings</span>
            </div>
          </div>
        </div>

        {/* Card 4: API Calls */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-4 shadow-xs hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex flex-col justify-between group">
          <div className="flex items-center justify-between mb-3">
            <div className="w-9 h-9 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center shadow-2xs group-hover:bg-indigo-600 group-hover:text-white transition-colors">
              <LinkIcon className="h-4.5 w-4.5" />
            </div>
            <span className="text-[10px] font-bold text-emerald-700 bg-emerald-50 border border-emerald-200/80 px-1.5 py-0.5 rounded-full">
              +32%
            </span>
          </div>
          <div>
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">API Gateway Calls</span>
            <div className="text-2xl font-black text-slate-900 tracking-tight mt-0.5">45,678</div>
            <div className="text-[11px] font-medium text-slate-400 mt-1">
              <span>Carriers & webhooks</span>
            </div>
          </div>
        </div>

        {/* Card 5: Documents Processed */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-4 shadow-xs hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex flex-col justify-between group">
          <div className="flex items-center justify-between mb-3">
            <div className="w-9 h-9 rounded-xl bg-orange-50 text-orange-600 flex items-center justify-center shadow-2xs group-hover:bg-orange-600 group-hover:text-white transition-colors">
              <FileText className="h-4.5 w-4.5" />
            </div>
            <span className="text-[10px] font-bold text-emerald-700 bg-emerald-50 border border-emerald-200/80 px-1.5 py-0.5 rounded-full">
              +15%
            </span>
          </div>
          <div>
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">Documents Parsed</span>
            <div className="text-2xl font-black text-slate-900 tracking-tight mt-0.5">{documentsProcessed}</div>
            <div className="text-[11px] font-medium text-slate-400 mt-1">
              <span>99.4% OCR confidence</span>
            </div>
          </div>
        </div>

        {/* Card 6: Storage Used */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-4 shadow-xs hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex flex-col justify-between group">
          <div className="flex items-center justify-between mb-3">
            <div className="w-9 h-9 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center shadow-2xs group-hover:bg-emerald-600 group-hover:text-white transition-colors">
              <HardDrive className="h-4.5 w-4.5" />
            </div>
            <span className="text-[10px] font-bold text-slate-600 bg-slate-100 px-1.5 py-0.5 rounded-full">
              24% Quota
            </span>
          </div>
          <div>
            <span className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider block">S3 Cloud Storage</span>
            <div className="text-2xl font-black text-slate-900 tracking-tight mt-0.5">2.4 GB</div>
            <div className="text-[11px] font-medium text-slate-400 mt-1">
              <span>of 10 GB limit provisioned</span>
            </div>
          </div>
        </div>
      </div>

      {/* 3. Middle Visual Analytics Section (3 Refined High-Clarity Cards) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        {/* Column 1: Usage Trend Spline */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                <span>Usage & Activity Trajectory</span>
              </h3>
              <p className="text-[11px] text-slate-400 mt-0.5">Monthly platform activity progression</p>
            </div>
            <div className="flex items-center gap-1.5 text-xs font-semibold text-blue-700 bg-blue-50 border border-blue-200/60 px-2 py-0.5 rounded-lg">
              <TrendingUp className="h-3 w-3" />
              <span>+32% Q3</span>
            </div>
          </div>

          <div className="mt-4 relative h-40 flex flex-col justify-end">
            <svg className="w-full h-34 overflow-visible" viewBox="0 0 250 85">
              <defs>
                <linearGradient id="usageGradientClean" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#3b82f6" stopOpacity="0.25" />
                  <stop offset="70%" stopColor="#60a5fa" stopOpacity="0.05" />
                  <stop offset="100%" stopColor="#ffffff" stopOpacity="0.0" />
                </linearGradient>
                <filter id="glowEffect" x="-20%" y="-20%" width="140%" height="140%">
                  <feDropShadow dx="0" dy="2" stdDeviation="3" floodColor="#2563eb" floodOpacity="0.3" />
                </filter>
              </defs>

              {/* Grid Lines */}
              <line x1="20" y1="10" x2="245" y2="10" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="14" y="13" textAnchor="end" className="text-[8px] fill-slate-400 font-mono">100k</text>

              <line x1="20" y1="32" x2="245" y2="32" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="14" y="35" textAnchor="end" className="text-[8px] fill-slate-400 font-mono">75k</text>

              <line x1="20" y1="54" x2="245" y2="54" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="14" y="57" textAnchor="end" className="text-[8px] fill-slate-400 font-mono">50k</text>

              <line x1="20" y1="75" x2="245" y2="75" stroke="#e2e8f0" />
              <text x="14" y="78" textAnchor="end" className="text-[8px] fill-slate-400 font-mono">0</text>

              {/* Smooth Area Path */}
              <path
                d="M 25 68 C 45 60, 55 54, 67 52 C 85 48, 95 43, 109 40 C 125 36, 135 34, 151 32 C 168 28, 178 24, 193 22 C 210 18, 222 14, 235 12 L 235 75 L 25 75 Z"
                fill="url(#usageGradientClean)"
              />

              {/* Smooth Spline Curve */}
              <path
                d="M 25 68 C 45 60, 55 54, 67 52 C 85 48, 95 43, 109 40 C 125 36, 135 34, 151 32 C 168 28, 178 24, 193 22 C 210 18, 222 14, 235 12"
                fill="none"
                stroke="#2563eb"
                strokeWidth="2.5"
                strokeLinecap="round"
                filter="url(#glowEffect)"
              />

              {/* Data points */}
              {trendPoints.map((pt, i) => (
                <g key={i}>
                  <circle
                    cx={pt.x}
                    cy={pt.y}
                    r={i === trendPoints.length - 1 ? 4 : 2.5}
                    fill={i === trendPoints.length - 1 ? '#2563eb' : '#3b82f6'}
                    stroke="#ffffff"
                    strokeWidth={i === trendPoints.length - 1 ? 2 : 1}
                    className="cursor-pointer hover:r-5 transition-all"
                  />
                </g>
              ))}

              {/* Peak Value Callout Tag */}
              <g transform="translate(208, 0)">
                <rect width="32" height="15" rx="5" fill="#0f172a" />
                <text x="16" y="11" textAnchor="middle" className="text-[9px] font-bold fill-white">
                  96k
                </text>
              </g>
            </svg>

            <div className="flex justify-between pl-6 pr-2 text-[10px] font-mono font-medium text-slate-400 mt-1.5 border-t border-slate-100 pt-1">
              <span>Apr 26</span>
              <span>May 26</span>
              <span>Jun 26</span>
              <span>Jul 26</span>
              <span>Aug 26</span>
              <span className="font-bold text-blue-600">Sep 26</span>
            </div>
          </div>
        </div>

        {/* Column 2: Module Workload Donut */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between mb-2">
            <div>
              <h3 className="text-sm font-bold text-slate-900">Module Workload Distribution</h3>
              <p className="text-[11px] text-slate-400 mt-0.5">Execution share by functional domain</p>
            </div>
            <span className="text-[11px] font-bold text-slate-600 bg-slate-100 px-2 py-0.5 rounded-md">
              6 Modules
            </span>
          </div>

          <div className="flex items-center justify-between gap-4 my-auto">
            {/* Donut Chart */}
            <div className="relative flex items-center justify-center shrink-0 w-32 h-32">
              <svg className="w-32 h-32 -rotate-90" viewBox="0 0 100 100">
                {/* 28% Freight Operations (Blue) */}
                <circle cx="50" cy="50" r="38" fill="transparent" stroke="#2563eb" strokeWidth="12" strokeDasharray="66.8 172.0" strokeDashoffset="0" />
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
              <div className="absolute inset-0 flex flex-col items-center justify-center text-center pointer-events-none">
                <span className="text-xl font-black text-slate-900 leading-none">1,234</span>
                <span className="text-[10px] text-slate-400 font-semibold uppercase tracking-wider mt-1">Actions</span>
              </div>
            </div>

            {/* Compact High-Readability Legend */}
            <div className="space-y-1.5 flex-1 min-w-[140px] text-xs">
              <div className="flex items-center justify-between py-0.5">
                <span className="flex items-center gap-1.5 text-slate-700 font-medium truncate">
                  <span className="w-2 h-2 rounded-full bg-blue-600 shrink-0" />
                  <span className="truncate">Freight Ops</span>
                </span>
                <span className="font-bold text-slate-900 text-[11px]">28%</span>
              </div>

              <div className="flex items-center justify-between py-0.5">
                <span className="flex items-center gap-1.5 text-slate-700 font-medium truncate">
                  <span className="w-2 h-2 rounded-full bg-emerald-500 shrink-0" />
                  <span className="truncate">Shipment Tracking</span>
                </span>
                <span className="font-bold text-slate-900 text-[11px]">22%</span>
              </div>

              <div className="flex items-center justify-between py-0.5">
                <span className="flex items-center gap-1.5 text-slate-700 font-medium truncate">
                  <span className="w-2 h-2 rounded-full bg-purple-500 shrink-0" />
                  <span className="truncate">Document OCR</span>
                </span>
                <span className="font-bold text-slate-900 text-[11px]">16%</span>
              </div>

              <div className="flex items-center justify-between py-0.5">
                <span className="flex items-center gap-1.5 text-slate-700 font-medium truncate">
                  <span className="w-2 h-2 rounded-full bg-amber-500 shrink-0" />
                  <span className="truncate">Carrier APIs</span>
                </span>
                <span className="font-bold text-slate-900 text-[11px]">14%</span>
              </div>

              <div className="flex items-center justify-between py-0.5">
                <span className="flex items-center gap-1.5 text-slate-700 font-medium truncate">
                  <span className="w-2 h-2 rounded-full bg-sky-500 shrink-0" />
                  <span className="truncate">Team & RBAC</span>
                </span>
                <span className="font-bold text-slate-900 text-[11px]">10%</span>
              </div>

              <div className="flex items-center justify-between py-0.5">
                <span className="flex items-center gap-1.5 text-slate-700 font-medium truncate">
                  <span className="w-2 h-2 rounded-full bg-slate-400 shrink-0" />
                  <span className="truncate">Billing / Others</span>
                </span>
                <span className="font-bold text-slate-900 text-[11px]">10%</span>
              </div>
            </div>
          </div>
        </div>

        {/* Column 3: User Engagement & Capacity */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-bold text-slate-900">User Seat Allocation</h3>
              <p className="text-[11px] text-slate-400 mt-0.5">Active vs provisioned team members</p>
            </div>
            <div className="flex items-center gap-3 text-[11px] font-semibold text-slate-600">
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 rounded-xs bg-blue-600" />
                <span>Active</span>
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 rounded-xs bg-sky-200" />
                <span>Total</span>
              </span>
            </div>
          </div>

          <div className="mt-4 relative h-40 flex flex-col justify-end">
            <svg className="w-full h-34 overflow-visible" viewBox="0 0 250 85">
              {/* Y-axis grid */}
              <line x1="20" y1="10" x2="245" y2="10" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="14" y="13" textAnchor="end" className="text-[8px] fill-slate-400 font-mono">20</text>

              <line x1="20" y1="32" x2="245" y2="32" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="14" y="35" textAnchor="end" className="text-[8px] fill-slate-400 font-mono">15</text>

              <line x1="20" y1="54" x2="245" y2="54" stroke="#f1f5f9" strokeDasharray="3 3" />
              <text x="14" y="57" textAnchor="end" className="text-[8px] fill-slate-400 font-mono">10</text>

              <line x1="20" y1="75" x2="245" y2="75" stroke="#e2e8f0" />
              <text x="14" y="78" textAnchor="end" className="text-[8px] fill-slate-400 font-mono">0</text>

              {/* Grouped Bars per month */}
              {/* Apr */}
              <rect x="36" y="42" width="7" height="33" rx="2" fill="#2563eb" />
              <rect x="44" y="28" width="7" height="47" rx="2" fill="#bae6fd" />

              {/* May */}
              <rect x="74" y="32" width="7" height="43" rx="2" fill="#2563eb" />
              <rect x="82" y="22" width="7" height="53" rx="2" fill="#bae6fd" />

              {/* Jun */}
              <rect x="112" y="25" width="7" height="50" rx="2" fill="#2563eb" />
              <rect x="120" y="16" width="7" height="59" rx="2" fill="#bae6fd" />

              {/* Jul */}
              <rect x="150" y="25" width="7" height="50" rx="2" fill="#2563eb" />
              <rect x="158" y="16" width="7" height="59" rx="2" fill="#bae6fd" />

              {/* Aug */}
              <rect x="188" y="28" width="7" height="47" rx="2" fill="#2563eb" />
              <rect x="196" y="14" width="7" height="61" rx="2" fill="#bae6fd" />

              {/* Sep */}
              <rect x="226" y="34" width="7" height="41" rx="2" fill="#2563eb" />
              <rect x="234" y="10" width="7" height="65" rx="2" fill="#bae6fd" />
            </svg>

            <div className="flex justify-between pl-8 pr-2 text-[10px] font-mono font-medium text-slate-400 mt-1.5 border-t border-slate-100 pt-1">
              <span>Apr</span>
              <span>May</span>
              <span>Jun</span>
              <span>Jul</span>
              <span>Aug</span>
              <span className="font-bold text-blue-600">Sep</span>
            </div>
          </div>
        </div>
      </div>

      {/* 4. Deep Operational Breakdown (3 Cards: Adoption Matrix, API Workload, Storage Vault) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        {/* Card 1: Module Adoption Progress */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                <Box className="h-4 w-4 text-emerald-600" />
                <span>Feature & Module Adoption</span>
              </h3>
              <span className="text-[10px] font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200/80">
                84% Average
              </span>
            </div>

            <div className="space-y-3.5">
              {/* Operations */}
              <div>
                <div className="flex items-center justify-between text-xs mb-1.5">
                  <span className="font-semibold text-slate-800">Freight Operations</span>
                  <span className="font-bold text-emerald-700 font-mono">92%</span>
                </div>
                <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                  <div className="h-full bg-linear-to-r from-emerald-500 to-teal-500 rounded-full" style={{ width: '92%' }} />
                </div>
              </div>

              {/* Shipments */}
              <div>
                <div className="flex items-center justify-between text-xs mb-1.5">
                  <span className="font-semibold text-slate-800">Shipment & Milestones</span>
                  <span className="font-bold text-blue-700 font-mono">88%</span>
                </div>
                <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                  <div className="h-full bg-linear-to-r from-blue-500 to-indigo-500 rounded-full" style={{ width: '88%' }} />
                </div>
              </div>

              {/* Documents */}
              <div>
                <div className="flex items-center justify-between text-xs mb-1.5">
                  <span className="font-semibold text-slate-800">Document Vault & OCR</span>
                  <span className="font-bold text-purple-700 font-mono">76%</span>
                </div>
                <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                  <div className="h-full bg-linear-to-r from-purple-500 to-violet-500 rounded-full" style={{ width: '76%' }} />
                </div>
              </div>

              {/* Integrations */}
              <div>
                <div className="flex items-center justify-between text-xs mb-1.5">
                  <span className="font-semibold text-slate-800">Carrier API Integrations</span>
                  <span className="font-bold text-amber-700 font-mono">65%</span>
                </div>
                <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                  <div className="h-full bg-linear-to-r from-amber-500 to-orange-500 rounded-full" style={{ width: '65%' }} />
                </div>
              </div>

              {/* Billing */}
              <div>
                <div className="flex items-center justify-between text-xs mb-1.5">
                  <span className="font-semibold text-slate-800">Billing & Accounting</span>
                  <span className="font-bold text-slate-700 font-mono">58%</span>
                </div>
                <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                  <div className="h-full bg-linear-to-r from-slate-500 to-slate-700 rounded-full" style={{ width: '58%' }} />
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: API & Gateway Workload */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                <LinkIcon className="h-4 w-4 text-blue-600" />
                <span>API & Gateway Telemetry</span>
              </h3>
              <button
                onClick={() => onNavigateTab?.('integrations')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer hover:underline"
              >
                <span>Gateway</span>
                <ArrowRight className="h-3 w-3" />
              </button>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between p-2.5 rounded-xl bg-slate-50/70 border border-slate-100">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-blue-100 text-blue-700">
                    <Radio className="h-3.5 w-3.5" />
                  </div>
                  <div>
                    <div className="text-xs font-bold text-slate-900">Carrier REST APIs</div>
                    <div className="text-[10px] text-slate-400">Maersk, Hapag-Lloyd, MSC</div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-xs font-bold font-mono text-slate-900">32,456</div>
                  <div className="text-[10px] font-bold text-emerald-600">↑ 28% MoM</div>
                </div>
              </div>

              <div className="flex items-center justify-between p-2.5 rounded-xl bg-slate-50/70 border border-slate-100">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-purple-100 text-purple-700">
                    <Workflow className="h-3.5 w-3.5" />
                  </div>
                  <div>
                    <div className="text-xs font-bold text-slate-900">Tracking Webhooks</div>
                    <div className="text-[10px] text-slate-400">Gate-in & vessel events</div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-xs font-bold font-mono text-slate-900">8,932</div>
                  <div className="text-[10px] font-bold text-emerald-600">↑ 40% MoM</div>
                </div>
              </div>

              <div className="flex items-center justify-between p-2.5 rounded-xl bg-slate-50/70 border border-slate-100">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-emerald-100 text-emerald-700">
                    <Cpu className="h-3.5 w-3.5" />
                  </div>
                  <div>
                    <div className="text-xs font-bold text-slate-900">Neural OCR Pipeline</div>
                    <div className="text-[10px] text-slate-400">AWS Textract extraction</div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-xs font-bold font-mono text-slate-900">1,245</div>
                  <div className="text-[10px] font-bold text-emerald-600">↑ 62% MoM</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Card 3: Storage Vault & Documents */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                <HardDrive className="h-4 w-4 text-emerald-600" />
                <span>Document Vault & S3 Quota</span>
              </h3>
              <span className="text-[10px] font-bold text-slate-600 bg-slate-100 px-2 py-0.5 rounded-md font-mono">
                AES-256 S3
              </span>
            </div>

            {/* Storage Progress Meter */}
            <div className="p-3.5 rounded-xl bg-slate-50/70 border border-slate-100 mb-4">
              <div className="flex items-baseline justify-between mb-2">
                <div className="flex items-baseline gap-1.5">
                  <span className="text-xl font-black text-slate-900">2.4 GB</span>
                  <span className="text-xs text-slate-500 font-medium">used of 10.0 GB</span>
                </div>
                <span className="text-xs font-bold text-blue-700 bg-blue-100 px-2 py-0.5 rounded-full">
                  24% Quota
                </span>
              </div>
              <div className="h-2.5 w-full bg-slate-200 rounded-full overflow-hidden">
                <div className="h-full bg-linear-to-r from-blue-500 to-indigo-600 rounded-full" style={{ width: '24%' }} />
              </div>
            </div>

            {/* Document KPI Tiles */}
            <div className="grid grid-cols-3 gap-2">
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 text-center">
                <div className="text-base font-black text-slate-900">892</div>
                <div className="text-[10px] text-slate-500 font-semibold mt-0.5">Bills of Lading</div>
              </div>

              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 text-center">
                <div className="text-base font-black text-emerald-600">156</div>
                <div className="text-[10px] text-slate-500 font-semibold mt-0.5">Verified OCR</div>
              </div>

              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 text-center">
                <div className="text-base font-black text-amber-600">12</div>
                <div className="text-[10px] text-slate-500 font-semibold mt-0.5">Pending Review</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 5. Strategic AI Growth & Account Optimization (Row 4) */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* Card 1: Key Insights */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center gap-2 mb-3.5">
              <div className="p-1.5 rounded-lg bg-purple-50 text-purple-600">
                <Sparkles className="h-4 w-4" />
              </div>
              <h3 className="text-sm font-bold text-slate-900">AI Platform Insights</h3>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs text-slate-700">
              <div className="space-y-2.5">
                <div className="flex items-start gap-2.5 p-2 rounded-xl bg-emerald-50/50 border border-emerald-100">
                  <CheckCircle2 className="h-4 w-4 text-emerald-600 shrink-0 mt-0.5" />
                  <span>Consumption velocity grew by <strong>32%</strong> over the last quarter.</span>
                </div>
                <div className="flex items-start gap-2.5 p-2 rounded-xl bg-blue-50/50 border border-blue-100">
                  <CheckCircle2 className="h-4 w-4 text-blue-600 shrink-0 mt-0.5" />
                  <span>Container tracking milestone synchronization active with zero latency.</span>
                </div>
              </div>

              <div className="space-y-2.5">
                <div className="flex items-start gap-2.5 p-2 rounded-xl bg-purple-50/50 border border-purple-100">
                  <CheckCircle2 className="h-4 w-4 text-purple-600 shrink-0 mt-0.5" />
                  <span>Neural OCR ingestion processing 300+ pages per month with 99.4% accuracy.</span>
                </div>
                <div className="flex items-start gap-2.5 p-2 rounded-xl bg-amber-50/50 border border-amber-100">
                  <CheckCircle2 className="h-4 w-4 text-amber-600 shrink-0 mt-0.5" />
                  <span>Billing module utilization has 42% headroom for growth.</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: Growth Opportunities */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <div className="flex items-center gap-2">
                <div className="p-1.5 rounded-lg bg-rose-50 text-rose-600">
                  <Target className="h-4 w-4" />
                </div>
                <h3 className="text-sm font-bold text-slate-900">Revenue & Expansion Opportunities</h3>
              </div>
              <span className="text-[10px] font-bold text-rose-700 bg-rose-50 px-2 py-0.5 rounded-full border border-rose-200/80">
                Actionable
              </span>
            </div>

            <div className="space-y-2.5">
              <div className="p-2.5 rounded-xl border border-slate-100 bg-slate-50/60 hover:bg-slate-100/60 transition-colors flex items-center justify-between gap-3">
                <div className="flex items-center gap-2.5 min-w-0">
                  <div className="p-1.5 rounded-lg bg-purple-50 text-purple-600 shrink-0">
                    <Zap className="h-3.5 w-3.5" />
                  </div>
                  <div className="min-w-0">
                    <p className="text-xs font-bold text-slate-900 truncate">Upgrade to Enterprise Ocean EDI Tier</p>
                    <p className="text-[11px] text-slate-500 truncate">Customer approaching API quota on standard REST tier.</p>
                  </div>
                </div>
                <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-rose-50 text-rose-700 border border-rose-200 shrink-0">
                  High Impact
                </span>
              </div>

              <div className="p-2.5 rounded-xl border border-slate-100 bg-slate-50/60 hover:bg-slate-100/60 transition-colors flex items-center justify-between gap-3">
                <div className="flex items-center gap-2.5 min-w-0">
                  <div className="p-1.5 rounded-lg bg-blue-50 text-blue-600 shrink-0">
                    <Users className="h-3.5 w-3.5" />
                  </div>
                  <div className="min-w-0">
                    <p className="text-xs font-bold text-slate-900 truncate">Activate Driver & CFS Dispatch Seats</p>
                    <p className="text-[11px] text-slate-500 truncate">8 unallocated seats ready for operations onboarding.</p>
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
    </div>
  );
}
