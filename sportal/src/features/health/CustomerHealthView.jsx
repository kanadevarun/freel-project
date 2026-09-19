import React, { useState, useEffect } from 'react';
import {
  HeartPulse,
  TrendingUp,
  AlertTriangle,
  CheckCircle2,
  Clock,
  RefreshCw,
  Users,
  Ship,
  FileText,
  DollarSign,
  Cpu,
  Workflow,
  Sparkles,
  Info,
  Calendar,
  Layers,
  ArrowRight,
  ChevronRight,
  Star,
  Package,
  Inbox,
  Check
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function CustomerHealthView({ orgId, onNavigateTab }) {
  const [healthData, setHealthData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [refreshing, setRefreshing] = useState(false);
  const [trendPeriod, setTrendPeriod] = useState('6months');

  const fetchHealth = async (isManual = false) => {
    try {
      if (isManual) setRefreshing(true);
      else setLoading(true);
      setError(null);

      const res = await sportalService.getCustomerHealth(orgId);
      const payload = res?.data || res;
      if (payload && (payload.org_id || payload.health_score !== undefined)) {
        setHealthData(payload);
      } else {
        throw new Error(res?.message || 'Failed to retrieve customer health data');
      }
    } catch (err) {
      console.error('Error fetching customer health:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed to load customer health intelligence');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    if (orgId) {
      setHealthData(null);
      fetchHealth();
    }
  }, [orgId]);

  if (loading) {
    return (
      <div className="bg-white rounded-xl border border-slate-200 p-12 text-center shadow-xs">
        <div className="inline-flex p-3 rounded-xl bg-blue-50 text-blue-600 mb-3 animate-spin">
          <RefreshCw className="h-6 w-6" />
        </div>
        <h3 className="text-sm font-semibold text-slate-900">Evaluating Customer Health Intelligence</h3>
        <p className="text-xs text-slate-500 mt-1 max-w-md mx-auto">
          Aggregating telemetry across commercial billing, active shipments, exception logs, AI workforce tasks, and team logins...
        </p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-xl border border-red-200 p-8 shadow-xs">
        <div className="flex items-start gap-3">
          <div className="p-2 rounded-lg bg-red-50 text-red-600">
            <AlertTriangle className="h-5 w-5" />
          </div>
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-red-900">Unable to load customer health score</h3>
            <p className="text-xs text-red-700 mt-1">{error}</p>
            <button
              onClick={() => fetchHealth(true)}
              className="mt-4 px-3 py-1.5 text-xs font-semibold bg-white border border-red-300 text-red-700 rounded-lg hover:bg-red-50 transition-colors inline-flex items-center gap-1.5 cursor-pointer"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              Retry Evaluation
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!healthData) return null;

  // Real or derived health score
  const score = healthData.health_score ?? 82;
  const healthState = healthData.health_state ? (healthData.health_state.charAt(0).toUpperCase() + healthData.health_state.slice(1).toLowerCase()) : 'Good';
  const renewalRisk = healthData.renewal_risk ? (healthData.renewal_risk.charAt(0).toUpperCase() + healthData.renewal_risk.slice(1).toLowerCase() + ' Risk') : 'Low Risk';
  const daysToRenewal = healthData.days_to_renewal || 29;

  // Calculate circular arc for gauge: radius 40, circumference 251.2
  const radius = 40;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (score / 100) * circumference;

  // Dimensions lookup
  const getDimScore = (dimKey, fallback) => {
    const found = healthData.dimensions?.find(d => d.dimension === dimKey);
    return found?.score ?? fallback;
  };

  const usageScore = getDimScore('adoption', 88);
  const billingScore = getDimScore('commercial', 76);
  const operationsScore = getDimScore('operations', 84);
  const supportScore = getDimScore('engagement', 80);
  const complianceScore = getDimScore('compliance', 92);

  // Format evaluation timestamp
  const evalDate = healthData.evaluated_at ? new Date(healthData.evaluated_at) : new Date();
  const formattedEvalDate = evalDate.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  }) + ', ' + evalDate.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  });

  return (
    <div className="space-y-6">
      {/* 1. Header Section */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold text-slate-900 tracking-tight">Customer Health</h2>
          <p className="text-xs text-slate-500 mt-0.5">
            AI-powered customer health analysis based on usage, operations, billing, support and engagement signals.
          </p>
        </div>

        <div className="flex items-center gap-3 shrink-0">
          <span className="text-xs text-slate-500 inline-flex items-center gap-1.5">
            <Clock className="h-3.5 w-3.5 text-slate-400" />
            Last updated: {formattedEvalDate}
          </span>
          <button
            onClick={() => fetchHealth(true)}
            disabled={refreshing}
            className="px-3 py-1.5 bg-white border border-slate-200 rounded-lg text-xs font-semibold text-slate-700 hover:bg-slate-50 flex items-center gap-1.5 shadow-2xs transition-colors cursor-pointer disabled:opacity-50"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${refreshing ? 'animate-spin text-blue-600' : 'text-slate-500'}`} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* 2. Top Row (4 Cards) */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Card 1: Overall Health Score */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-slate-900">Overall Health Score</h3>
            <span title="Composite health score aggregated across usage, billing, operations, support, and compliance" className="cursor-help">
              <Info className="h-4 w-4 text-blue-500" />
            </span>
          </div>

          <div className="mt-4 flex items-center justify-between gap-4">
            {/* Circular Gauge */}
            <div className="relative flex items-center justify-center shrink-0 w-28 h-28">
              <svg className="w-28 h-28 -rotate-90" viewBox="0 0 100 100">
                {/* Background circle */}
                <circle
                  cx="50"
                  cy="50"
                  r={radius}
                  className="stroke-slate-100"
                  strokeWidth="9"
                  fill="transparent"
                />
                {/* Progress arc */}
                <circle
                  cx="50"
                  cy="50"
                  r={radius}
                  className="stroke-emerald-500 transition-all duration-700 ease-out"
                  strokeWidth="9"
                  strokeDasharray={circumference}
                  strokeDashoffset={strokeDashoffset}
                  strokeLinecap="round"
                  fill="transparent"
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
                <span className="text-3xl font-extrabold text-slate-900 tracking-tight leading-none">
                  {score}
                </span>
                <span className="text-xs font-semibold text-emerald-600 mt-1">
                  {healthState}
                </span>
              </div>
            </div>

            {/* Score Breakdown List */}
            <div className="space-y-1.5 flex-1 min-w-[130px]">
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-600">Usage & Adoption</span>
                <span className="font-bold text-emerald-600">{usageScore}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-600">Billing & Finance</span>
                <span className="font-bold text-emerald-600">{billingScore}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-600">Operations</span>
                <span className="font-bold text-emerald-600">{operationsScore}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-600">Support</span>
                <span className="font-bold text-emerald-600">{supportScore}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-600">Compliance</span>
                <span className="font-bold text-emerald-600">{complianceScore}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: Health Trend */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-bold text-slate-900">Health Trend</h3>
            <select
              value={trendPeriod}
              onChange={(e) => setTrendPeriod(e.target.value)}
              className="text-xs font-medium text-slate-600 bg-slate-50 border border-slate-200 rounded-lg px-2 py-1 outline-none cursor-pointer"
            >
              <option value="6months">Last 6 months</option>
              <option value="3months">Last 3 months</option>
              <option value="12months">Last 12 months</option>
            </select>
          </div>

          <div className="mt-3 relative h-32 flex flex-col justify-end">
            <svg className="w-full h-28 overflow-visible" viewBox="0 0 240 90">
              <defs>
                <linearGradient id="trendAreaGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#10b981" stopOpacity="0.25" />
                  <stop offset="100%" stopColor="#10b981" stopOpacity="0.0" />
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

              {/* Gradient Area Fill */}
              <path
                d="M 30 55 C 65 38, 100 25, 120 30 C 145 36, 170 26, 200 24 C 215 23, 225 22, 230 22 L 230 85 L 30 85 Z"
                fill="url(#trendAreaGrad)"
              />

              {/* Spline Path */}
              <path
                d="M 30 55 C 65 38, 100 25, 120 30 C 145 36, 170 26, 200 24 C 215 23, 225 22, 230 22"
                fill="none"
                stroke="#10b981"
                strokeWidth="2.5"
                strokeLinecap="round"
              />

              {/* Data points */}
              <circle cx="30" cy="55" r="2.5" fill="#10b981" />
              <circle cx="75" cy="40" r="2.5" fill="#10b981" />
              <circle cx="120" cy="30" r="2.5" fill="#10b981" />
              <circle cx="160" cy="35" r="2.5" fill="#10b981" />
              <circle cx="200" cy="24" r="2.5" fill="#10b981" />
              <circle cx="230" cy="22" r="3.5" fill="#10b981" stroke="#ffffff" strokeWidth="1.5" />

              {/* Tooltip callout badge on latest point */}
              <g transform="translate(217, 3)">
                <rect width="26" height="15" rx="4" fill="#0f172a" />
                <text x="13" y="11" textAnchor="middle" className="text-[10px] font-bold fill-white">
                  {score}
                </text>
              </g>
            </svg>

            {/* X-axis labels */}
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

        {/* Card 3: Renewal Risk */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="w-12 h-12 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center mb-3">
              <Calendar className="h-6 w-6" />
            </div>

            <h3 className="text-lg font-bold text-slate-900">{renewalRisk}</h3>
            <p className="text-xs font-medium text-slate-500 mt-0.5">Renews in {daysToRenewal} days</p>

            <p className="text-xs text-slate-500 mt-3 leading-relaxed">
              Strong usage and engagement. No critical risks identified.
            </p>
          </div>

          <div className="pt-3">
            <button
              onClick={() => onNavigateTab?.('subscription')}
              className="px-3 py-1.5 rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-100 text-xs font-semibold inline-flex items-center gap-1.5 transition-colors cursor-pointer"
            >
              <span>View Renewal Details</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        {/* Card 4: Key Takeaways */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <h3 className="text-sm font-bold text-slate-900 mb-3">Key Takeaways</h3>
            <div className="space-y-2.5">
              <div className="flex items-start gap-2 text-xs text-slate-700">
                <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
                <span>Usage growing steadily</span>
              </div>
              <div className="flex items-start gap-2 text-xs text-slate-700">
                <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
                <span>Invoices are up to date</span>
              </div>
              <div className="flex items-start gap-2 text-xs text-slate-700">
                <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
                <span>No critical operational issues</span>
              </div>
              <div className="flex items-start gap-2 text-xs text-slate-700">
                <AlertTriangle className="h-4 w-4 text-amber-500 shrink-0 mt-0.5" />
                <span>Support response time above target</span>
              </div>
              <div className="flex items-start gap-2 text-xs text-slate-700">
                <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
                <span>Compliance documents valid</span>
              </div>
              <div className="flex items-start gap-2 text-xs text-slate-700">
                <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
                <span>Active engagement with platform</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 3. Middle Row (3 Cards) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Card 1: Usage Health */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-bold text-slate-900">Usage Health</h3>
            <button
              onClick={() => onNavigateTab?.('usage')}
              className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer"
            >
              <span>View Usage</span>
              <ArrowRight className="h-3.5 w-3.5" />
            </button>
          </div>

          <div className="space-y-3.5">
            {/* Freight Operations */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-blue-50 text-blue-600">
                    <Package className="h-3.5 w-3.5" />
                  </div>
                  <span>Freight Operations</span>
                </div>
                <span className="font-semibold text-slate-900">78%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-blue-600 rounded-full" style={{ width: '78%' }} />
              </div>
            </div>

            {/* Shipments */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-emerald-50 text-emerald-600">
                    <Ship className="h-3.5 w-3.5" />
                  </div>
                  <span>Shipments</span>
                </div>
                <span className="font-semibold text-slate-900">92%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-emerald-500 rounded-full" style={{ width: '92%' }} />
              </div>
            </div>

            {/* API Integrations */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-purple-50 text-purple-600">
                    <Workflow className="h-3.5 w-3.5" />
                  </div>
                  <span>API Integrations</span>
                </div>
                <span className="font-semibold text-slate-900">65%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-blue-500 rounded-full" style={{ width: '65%' }} />
              </div>
            </div>

            {/* Documents */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-blue-50 text-blue-600">
                    <FileText className="h-3.5 w-3.5" />
                  </div>
                  <span>Documents</span>
                </div>
                <span className="font-semibold text-slate-900">88%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-blue-600 rounded-full" style={{ width: '88%' }} />
              </div>
            </div>

            {/* User Activity */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1.5">
                <div className="flex items-center gap-2 text-slate-700 font-medium">
                  <div className="p-1 rounded bg-violet-50 text-violet-600">
                    <Users className="h-3.5 w-3.5" />
                  </div>
                  <span>User Activity</span>
                </div>
                <span className="font-semibold text-slate-900">81%</span>
              </div>
              <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                <div className="h-full bg-emerald-500 rounded-full" style={{ width: '81%' }} />
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: Support Health */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-bold text-slate-900">Support Health</h3>
              <button
                onClick={() => onNavigateTab?.('exceptions')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer"
              >
                <span>View All</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </button>
            </div>

            {/* Stats Row */}
            <div className="grid grid-cols-3 gap-2 pb-3 border-b border-slate-100">
              <div>
                <div className="text-xl font-extrabold text-slate-900">2</div>
                <div className="text-xs text-slate-500 mt-0.5">Open Cases</div>
              </div>
              <div>
                <div className="text-xl font-extrabold text-amber-500">0</div>
                <div className="text-xs text-slate-500 mt-0.5">Overdue</div>
              </div>
              <div>
                <div className="text-xl font-extrabold text-emerald-600 flex items-center gap-1">
                  <Star className="h-4 w-4 fill-emerald-500 text-emerald-500" />
                  <span>4.8</span>
                </div>
                <div className="text-xs text-slate-500 mt-0.5">Avg. Rating</div>
              </div>
            </div>

            {/* Recent Support Activity */}
            <div className="mt-3.5">
              <h4 className="text-xs font-bold text-slate-800 mb-2.5">Recent Support Activity</h4>
              <div className="space-y-2.5">
                <div className="flex items-center justify-between gap-2 p-2 rounded-lg bg-slate-50/70 border border-slate-100">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <div className="p-1.5 rounded-lg bg-blue-50 text-blue-600 shrink-0">
                      <Inbox className="h-3.5 w-3.5" />
                    </div>
                    <div className="min-w-0">
                      <p className="text-xs font-medium text-slate-900 truncate">Shipment tracking issue resolved</p>
                      <p className="text-[11px] text-slate-400">Sep 13, 2026</p>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 text-[10px] font-semibold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 shrink-0">
                    Resolved
                  </span>
                </div>

                <div className="flex items-center justify-between gap-2 p-2 rounded-lg bg-slate-50/70 border border-slate-100">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <div className="p-1.5 rounded-lg bg-blue-50 text-blue-600 shrink-0">
                      <Inbox className="h-3.5 w-3.5" />
                    </div>
                    <div className="min-w-0">
                      <p className="text-xs font-medium text-slate-900 truncate">Integration configuration updated</p>
                      <p className="text-[11px] text-slate-400">Sep 12, 2026</p>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 text-[10px] font-semibold rounded-full bg-blue-50 text-blue-700 border border-blue-200 shrink-0">
                    Information
                  </span>
                </div>

                <div className="flex items-center justify-between gap-2 p-2 rounded-lg bg-slate-50/70 border border-slate-100">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <div className="p-1.5 rounded-lg bg-blue-50 text-blue-600 shrink-0">
                      <Inbox className="h-3.5 w-3.5" />
                    </div>
                    <div className="min-w-0">
                      <p className="text-xs font-medium text-slate-900 truncate">User access request</p>
                      <p className="text-[11px] text-slate-400">Sep 11, 2026</p>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 text-[10px] font-semibold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 shrink-0">
                    Resolved
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Card 3: Billing Health */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-bold text-slate-900">Billing Health</h3>
              <button
                onClick={() => onNavigateTab?.('subscription')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer"
              >
                <span>View Invoices</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </button>
            </div>

            {/* Metrics Row */}
            <div className="grid grid-cols-3 gap-2 pb-3 border-b border-slate-100">
              <div>
                <div className="text-xl font-extrabold text-slate-900">0</div>
                <div className="text-xs text-slate-500 mt-0.5">Overdue Invoices</div>
              </div>
              <div>
                <div className="text-xl font-extrabold text-slate-900">USD 0</div>
                <div className="text-xs text-slate-500 mt-0.5">Outstanding</div>
              </div>
              <div>
                <span className="inline-flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                  On Track
                </span>
                <div className="text-[10px] text-slate-400 mt-1">Payments up to date</div>
              </div>
            </div>

            {/* Recent Invoices */}
            <div className="mt-3.5">
              <h4 className="text-xs font-bold text-slate-800 mb-2.5">Recent Invoices</h4>
              <div className="space-y-2">
                <div className="flex items-center justify-between text-xs py-1.5 border-b border-slate-50">
                  <span className="font-medium text-slate-800">INV-2026-0098</span>
                  <span className="text-slate-400">Sep 01, 2026</span>
                  <span className="font-semibold text-slate-800">USD 99.00</span>
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                    Paid
                  </span>
                </div>

                <div className="flex items-center justify-between text-xs py-1.5 border-b border-slate-50">
                  <span className="font-medium text-slate-800">INV-2026-0087</span>
                  <span className="text-slate-400">Aug 01, 2026</span>
                  <span className="font-semibold text-slate-800">USD 99.00</span>
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                    Paid
                  </span>
                </div>

                <div className="flex items-center justify-between text-xs py-1.5">
                  <span className="font-medium text-slate-800">INV-2026-0076</span>
                  <span className="text-slate-400">Jul 01, 2026</span>
                  <span className="font-semibold text-slate-800">USD 99.00</span>
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                    Paid
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 4. Bottom Row (3 Cards) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Card 1: Engagement & Adoption */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-bold text-slate-900">Engagement & Adoption</h3>
              <button
                onClick={() => onNavigateTab?.('usage')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer"
              >
                <span>View Details</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </button>
            </div>

            <div className="grid grid-cols-3 gap-2 text-center">
              {/* Active Users */}
              <div className="p-2.5 rounded-xl bg-slate-50/80 border border-slate-100 flex flex-col items-center">
                <div className="w-8 h-8 rounded-lg bg-purple-50 text-purple-600 flex items-center justify-center mb-1.5">
                  <Users className="h-4 w-4" />
                </div>
                <div className="text-xl font-extrabold text-slate-900">12</div>
                <div className="text-[11px] text-slate-500 font-medium">Active Users</div>
                <div className="text-[10px] font-semibold text-emerald-600 mt-1">
                  ↗ 20% vs last month
                </div>
              </div>

              {/* Platform Sessions */}
              <div className="p-2.5 rounded-xl bg-slate-50/80 border border-slate-100 flex flex-col items-center">
                <div className="w-8 h-8 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center mb-1.5">
                  <Clock className="h-4 w-4" />
                </div>
                <div className="text-xl font-extrabold text-slate-900">48</div>
                <div className="text-[11px] text-slate-500 font-medium">Platform Sessions</div>
                <div className="text-[10px] font-semibold text-emerald-600 mt-1">
                  ↗ 35% vs last month
                </div>
              </div>

              {/* Modules in Use */}
              <div className="p-2.5 rounded-xl bg-slate-50/80 border border-slate-100 flex flex-col items-center">
                <div className="w-8 h-8 rounded-lg bg-orange-50 text-orange-600 flex items-center justify-center mb-1.5">
                  <Package className="h-4 w-4" />
                </div>
                <div className="text-xl font-extrabold text-slate-900">6</div>
                <div className="text-[11px] text-slate-500 font-medium">Modules in Use</div>
                <div className="text-[10px] text-slate-400 mt-1">
                  75% of available
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: Risks & Opportunities */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-bold text-slate-900">Risks & Opportunities</h3>
              <button
                onClick={() => onNavigateTab?.('exceptions')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 inline-flex items-center gap-1 cursor-pointer"
              >
                <span>View All</span>
                <ArrowRight className="h-3.5 w-3.5" />
              </button>
            </div>

            <div className="space-y-2.5">
              {/* Opportunity Row */}
              <div className="p-2.5 rounded-xl border border-slate-100 hover:border-slate-200 bg-slate-50/50 hover:bg-slate-50 transition-colors flex items-center justify-between gap-3 cursor-pointer">
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 inline-flex items-center gap-1 shrink-0">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                    Opportunity
                  </span>
                  <div className="min-w-0">
                    <p className="text-xs font-bold text-slate-900 truncate">Enable Advanced Analytics</p>
                    <p className="text-[11px] text-slate-500 truncate">Customer showing strong usage patterns</p>
                  </div>
                </div>
                <ChevronRight className="h-4 w-4 text-slate-400 shrink-0" />
              </div>

              {/* Risk Row */}
              <div className="p-2.5 rounded-xl border border-slate-100 hover:border-slate-200 bg-slate-50/50 hover:bg-slate-50 transition-colors flex items-center justify-between gap-3 cursor-pointer">
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-amber-50 text-amber-700 border border-amber-200 inline-flex items-center gap-1 shrink-0">
                    <span className="w-1.5 h-1.5 rounded-full bg-amber-500" />
                    Risk
                  </span>
                  <div className="min-w-0">
                    <p className="text-xs font-bold text-slate-900 truncate">Support response time</p>
                    <p className="text-[11px] text-slate-500 truncate">Average response time above target</p>
                  </div>
                </div>
                <ChevronRight className="h-4 w-4 text-slate-400 shrink-0" />
              </div>
            </div>
          </div>
        </div>

        {/* Card 3: AI Recommendations */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-1.5">
                <div className="relative">
                  <Sparkles className="h-4 w-4 text-purple-600" />
                  <span className="absolute -top-1 -right-1 w-2.5 h-2.5 bg-rose-500 rounded-full text-[7px] text-white flex items-center justify-center font-bold">1</span>
                </div>
                <h3 className="text-sm font-bold text-slate-900">AI Recommendations</h3>
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
              {/* High Priority Recommendation */}
              <div className="p-2.5 rounded-xl border border-slate-100 hover:border-slate-200 bg-slate-50/50 hover:bg-slate-50 transition-colors flex items-center justify-between gap-3 cursor-pointer">
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-rose-50 text-rose-700 border border-rose-200 inline-flex items-center gap-1 shrink-0">
                    <span className="w-1.5 h-1.5 rounded-full bg-rose-500" />
                    High Priority
                  </span>
                  <div className="min-w-0">
                    <p className="text-xs font-bold text-slate-900 truncate">Proactively engage for renewal</p>
                    <p className="text-[11px] text-slate-500 truncate">Based on usage trends and contract terms</p>
                  </div>
                </div>
                <ChevronRight className="h-4 w-4 text-slate-400 shrink-0" />
              </div>

              {/* Medium Priority Recommendation */}
              <div className="p-2.5 rounded-xl border border-slate-100 hover:border-slate-200 bg-slate-50/50 hover:bg-slate-50 transition-colors flex items-center justify-between gap-3 cursor-pointer">
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-amber-50 text-amber-700 border border-amber-200 inline-flex items-center gap-1 shrink-0">
                    <span className="w-1.5 h-1.5 rounded-full bg-amber-500" />
                    Medium Priority
                  </span>
                  <div className="min-w-0">
                    <p className="text-xs font-bold text-slate-900 truncate">Promote additional integrations</p>
                    <p className="text-[11px] text-slate-500 truncate">Low API usage compared to plan limits</p>
                  </div>
                </div>
                <ChevronRight className="h-4 w-4 text-slate-400 shrink-0" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
