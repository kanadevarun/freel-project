import React, { useState } from 'react';
import {
  Ship,
  MessageSquare,
  Mail,
  HardDrive,
  FileSearch,
  Webhook,
  CheckCircle2,
  AlertCircle,
  XCircle,
  Activity,
  Play,
  RotateCw,
  Settings,
  ShieldCheck,
  ExternalLink,
  Zap,
  Lock
} from 'lucide-react';

const CATEGORY_ICONS = {
  CARRIER: Ship,
  SMS: MessageSquare,
  EMAIL: Mail,
  STORAGE: HardDrive,
  TEXTRACT: FileSearch,
  WEBHOOK: Webhook,
};

const CATEGORY_COLORS = {
  CARRIER: 'text-blue-600 bg-blue-50 border-blue-200',
  SMS: 'text-amber-600 bg-amber-50 border-amber-200',
  EMAIL: 'text-indigo-600 bg-indigo-50 border-indigo-200',
  STORAGE: 'text-cyan-600 bg-cyan-50 border-cyan-200',
  TEXTRACT: 'text-purple-600 bg-purple-50 border-purple-200',
  WEBHOOK: 'text-emerald-600 bg-emerald-50 border-emerald-200',
};

export function IntegrationCard({
  integration,
  onToggle,
  onTestConnection,
  onConfigure,
  onViewDetails,
  canManage = true,
}) {
  const [toggling, setToggling] = useState(false);
  const [testing, setTesting] = useState(false);

  const isCarrier = integration.category === 'CARRIER';
  const IconComponent = CATEGORY_ICONS[integration.category] || Activity;
  const categoryBadgeClass = CATEGORY_COLORS[integration.category] || 'text-slate-600 bg-slate-50 border-slate-200';

  const rawStatus = (integration.status || '').toUpperCase();
  const isConnected = rawStatus === 'CONNECTED' || rawStatus === 'HEALTHY' || rawStatus === 'ACTIVE';
  const isDegraded = rawStatus === 'DEGRADED';
  const isFailed = rawStatus === 'FAILED' || rawStatus === 'ERROR';
  const isDisabled = rawStatus === 'DISABLED';
  const isPending = rawStatus === 'PENDING';
  const isUnavailable = rawStatus === 'UNAVAILABLE';
  const isNotConfigured = rawStatus === 'NOT_CONFIGURED' || rawStatus === 'CONFIGURATION_REQUIRED' || !rawStatus;

  let statusBadge;
  if (isConnected) {
    statusBadge = (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 shrink-0">
        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
        Connected
      </span>
    );
  } else if (isDegraded) {
    statusBadge = (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200 shrink-0">
        <AlertCircle className="h-3 w-3 text-amber-600" />
        Degraded
      </span>
    );
  } else if (isFailed) {
    statusBadge = (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200 shrink-0">
        <XCircle className="h-3 w-3 text-rose-600" />
        Failed
      </span>
    );
  } else if (isDisabled) {
    statusBadge = (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-600 border border-slate-200 shrink-0">
        <span className="h-1.5 w-1.5 rounded-full bg-slate-400" />
        Disabled
      </span>
    );
  } else if (isPending) {
    statusBadge = (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200 shrink-0">
        <RotateCw className="h-3 w-3 text-blue-600 animate-spin" />
        Pending
      </span>
    );
  } else if (isUnavailable) {
    statusBadge = (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-500 border border-slate-200 shrink-0">
        <XCircle className="h-3 w-3 text-slate-400" />
        Unavailable
      </span>
    );
  } else {
    statusBadge = (
      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200 shrink-0">
        <AlertCircle className="h-3 w-3 text-amber-500" />
        Not Configured
      </span>
    );
  }

  const handleToggle = async (e) => {
    e.stopPropagation();
    if (!canManage || toggling) return;
    try {
      setToggling(true);
      await onToggle(integration, !integration.is_enabled);
    } finally {
      setToggling(false);
    }
  };

  const handleTest = async (e) => {
    e.stopPropagation();
    if (testing) return;
    try {
      setTesting(true);
      await onTestConnection(integration);
    } finally {
      setTesting(false);
    }
  };

  const handleCardClick = () => {
    if (onViewDetails) {
      onViewDetails(integration);
    }
  };

  const scac = (integration.scac || integration.provider_name || '').toUpperCase();
  const capabilities = integration.supported_capabilities || integration.dependent_workflows || [];

  return (
    <div
      onClick={handleCardClick}
      className="bg-white rounded-2xl border border-slate-200/90 hover:border-blue-400/80 transition-all duration-200 shadow-xs hover:shadow-md flex flex-col justify-between overflow-hidden group cursor-pointer"
    >
      {/* Header & Body */}
      <div className="p-5 space-y-3.5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className={`h-11 w-11 rounded-xl flex items-center justify-center font-bold text-xs group-hover:scale-105 transition-transform shrink-0 ${
              isCarrier
                ? scac === 'MAEU' ? 'bg-sky-500 text-white shadow-sm'
                : scac === 'HLCU' ? 'bg-orange-500 text-white shadow-sm'
                : scac === 'MSCU' ? 'bg-amber-900 text-amber-100 shadow-sm'
                : scac === 'ONE' ? 'bg-pink-600 text-white shadow-sm'
                : 'bg-blue-700 text-white shadow-sm'
                : 'bg-slate-50 border border-slate-100 text-slate-700'
            }`}>
              {isCarrier ? scac.substring(0, 4) : <IconComponent className="h-5 w-5 text-navy-900" />}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h4 className="text-sm font-bold text-slate-900 leading-snug group-hover:text-blue-600 transition-colors">
                  {integration.display_name}
                </h4>
              </div>
              <div className="flex items-center gap-2 mt-0.5">
                <span className={`text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border ${categoryBadgeClass}`}>
                  {integration.category}
                </span>
                <span className="text-[11px] font-mono font-bold text-slate-500">
                  {isCarrier ? `SCAC: ${scac}` : integration.provider_name}
                </span>
                {integration.environment && (
                  <span className="text-[10px] font-mono text-slate-400 uppercase">
                    • {integration.environment}
                  </span>
                )}
              </div>
            </div>
          </div>
          {statusBadge}
        </div>

        {/* Description */}
        <p className="text-xs text-slate-600 line-clamp-2 leading-relaxed">
          {integration.description || 'Enterprise integration connector for LogisticsHQ freight operations.'}
        </p>

        {/* Capabilities Chips */}
        {capabilities.length > 0 && (
          <div className="flex flex-wrap gap-1.5 pt-1">
            {capabilities.slice(0, 3).map((cap) => (
              <span
                key={cap}
                className="text-[10px] font-medium px-2 py-0.5 rounded-md bg-slate-100 text-slate-600 border border-slate-200/80"
              >
                ✓ {cap}
              </span>
            ))}
            {capabilities.length > 3 && (
              <span className="text-[10px] font-bold px-1.5 py-0.5 rounded bg-blue-50 text-blue-600 border border-blue-200">
                +{capabilities.length - 3} more
              </span>
            )}
          </div>
        )}

        {/* Metrics Grid */}
        <div className="pt-3 border-t border-slate-100 grid grid-cols-3 gap-2 text-left">
          {isCarrier ? (
            <>
              <div className="bg-slate-50 p-2 rounded-lg border border-slate-100">
                <span className="text-[9px] uppercase font-bold text-slate-400 block tracking-wider">Containers</span>
                <span className="text-xs font-black text-slate-900 mt-0.5 block">
                  {integration.active_containers_count ? `${integration.active_containers_count} TEU` : '32 TEU'}
                </span>
              </div>
              <div className="bg-slate-50 p-2 rounded-lg border border-slate-100">
                <span className="text-[9px] uppercase font-bold text-slate-400 block tracking-wider">Bookings</span>
                <span className="text-xs font-black text-slate-900 mt-0.5 block">
                  {integration.active_bookings_count ? `${integration.active_bookings_count} Active` : '4 Active'}
                </span>
              </div>
              <div className="bg-slate-50 p-2 rounded-lg border border-slate-100">
                <span className="text-[9px] uppercase font-bold text-slate-400 block tracking-wider">Gateway Ping</span>
                <span className="text-xs font-black text-emerald-600 mt-0.5 block">
                  {integration.latency_ms ? `${integration.latency_ms}ms` : '142ms'}
                </span>
              </div>
            </>
          ) : (
            <>
              <div>
                <span className="text-[10px] uppercase font-semibold text-slate-400 block">Health Score</span>
                <span className={`text-xs font-bold ${
                  (integration.health_score ?? 0) >= 80 ? 'text-emerald-600' : 'text-slate-700'
                }`}>
                  {integration.health_score ?? 50}%
                </span>
              </div>
              <div>
                <span className="text-[10px] uppercase font-semibold text-slate-400 block">Sync Status</span>
                <span className="text-xs font-semibold text-slate-700 mt-0.5 block">
                  {integration.sync_status || (integration.is_enabled ? 'IDLE' : 'STOPPED')}
                </span>
              </div>
              <div>
                <span className="text-[10px] uppercase font-semibold text-slate-400 block">Protocol</span>
                <span className="text-xs font-mono font-medium text-slate-600 mt-0.5 block">
                  {integration.connection_method || 'REST / TLS'}
                </span>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Footer Actions */}
      <div className="bg-slate-50/85 border-t border-slate-100 px-5 py-3 flex items-center justify-between gap-3">
        {/* Toggle Switch */}
        <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
          <label className="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              className="sr-only peer"
              checked={integration.is_enabled}
              disabled={!canManage || toggling}
              onChange={handleToggle}
            />
            <div className={`w-9 h-5 bg-slate-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all ${
              integration.is_enabled ? 'peer-checked:bg-slate-900' : ''
            } ${!canManage ? 'opacity-50 cursor-not-allowed' : ''}`} />
          </label>
          <span className="text-xs font-medium text-slate-600">
            {toggling ? 'Updating...' : integration.is_enabled ? 'Active' : 'Inactive'}
          </span>
        </div>

        {/* Buttons */}
        <div className="flex items-center gap-1.5" onClick={(e) => e.stopPropagation()}>
          {isCarrier ? (
            <button
              type="button"
              onClick={() => onViewDetails && onViewDetails(integration)}
              className="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg border border-blue-200 bg-blue-50/80 hover:bg-blue-100 text-blue-700 text-xs font-bold transition-colors cursor-pointer shadow-2xs"
            >
              <span>View Data</span>
              <ExternalLink className="h-3 w-3" />
            </button>
          ) : (
            <button
              type="button"
              onClick={handleTest}
              disabled={testing}
              title={isNotConfigured ? 'Test connection' : 'Test API handshake and latency'}
              className="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 text-slate-700 text-xs font-medium transition-colors disabled:opacity-50 cursor-pointer shadow-2xs"
            >
              {testing ? (
                <RotateCw className="h-3.5 w-3.5 animate-spin text-blue-600" />
              ) : (
                <Zap className="h-3.5 w-3.5 text-blue-600" />
              )}
              <span>{testing ? 'Testing...' : 'Test'}</span>
            </button>
          )}

          <button
            type="button"
            onClick={() => onConfigure(integration)}
            className="inline-flex items-center gap-1 p-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 text-slate-600 hover:text-slate-900 text-xs font-medium transition-colors cursor-pointer shadow-2xs"
            title="Configure Credentials & Settings"
          >
            <Settings className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </div>
  );
}
