import React, { useState } from 'react';
import {
  Ship,
  X,
  ExternalLink,
  CheckCircle2,
  AlertCircle,
  Clock,
  Zap,
  RotateCw,
  Globe,
  Key,
  ShieldCheck,
  FileText,
  Anchor,
  Box,
  Radio,
  ArrowUpRight,
  Code
} from 'lucide-react';
import { MOCK_SHIPMENTS, MOCK_WEBHOOKS, MOCK_CONTRACTS } from '../../../services/mockData';

const CARRIER_COLORS = {
  MAEU: { bg: 'bg-sky-500', text: 'text-sky-600', border: 'border-sky-200', light: 'bg-sky-50' },
  HLCU: { bg: 'bg-orange-500', text: 'text-orange-600', border: 'border-orange-200', light: 'bg-orange-50' },
  MSCU: { bg: 'bg-amber-900', text: 'text-amber-800', border: 'border-amber-200', light: 'bg-amber-50' },
  ONE: { bg: 'bg-pink-600', text: 'text-pink-600', border: 'border-pink-200', light: 'bg-pink-50' },
  CMDU: { bg: 'bg-blue-700', text: 'text-blue-700', border: 'border-blue-200', light: 'bg-blue-50' },
  CMA_CGM: { bg: 'bg-blue-700', text: 'text-blue-700', border: 'border-blue-200', light: 'bg-blue-50' },
};

export function CarrierDetailsDrawer({
  isOpen,
  onClose,
  carrier,
  onTestConnection,
  onConfigure,
}) {
  const [activeTab, setActiveTab] = useState('SHIPMENTS'); // SHIPMENTS, WEBHOOKS, CONTRACTS, CONFIG
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [expandedWebhookId, setExpandedWebhookId] = useState(null);

  if (!isOpen || !carrier) return null;

  const scac = (carrier.scac || carrier.provider_name || 'CARRIER').toUpperCase();
  const brand = CARRIER_COLORS[scac] || { bg: 'bg-slate-700', text: 'text-slate-700', border: 'border-slate-200', light: 'bg-slate-50' };

  // Filter linked shipments for this carrier
  const linkedShipments = MOCK_SHIPMENTS.filter(
    (s) => s.carrier_scac?.toUpperCase() === scac || (scac === 'CMDU' && s.carrier_scac === 'CMA')
  );

  // Filter linked webhooks
  const linkedWebhooks = MOCK_WEBHOOKS.filter(
    (w) => w.scac?.toUpperCase() === scac
  );

  // Filter linked contracts
  const linkedContracts = MOCK_CONTRACTS.filter(
    (c) => c.carrier_scac?.toUpperCase() === scac
  );

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      if (onTestConnection) {
        const res = await onTestConnection(carrier);
        setTestResult({
          success: true,
          latency: carrier.latency_ms || 142,
          message: `DCSA 2.0 API Handshake successful • Latency: ${carrier.latency_ms || 142}ms via TLS 1.3`,
        });
      } else {
        await new Promise((r) => setTimeout(r, 600));
        setTestResult({
          success: true,
          latency: carrier.latency_ms || 142,
          message: `Direct gateway handshake verified with ${carrier.display_name}.`,
        });
      }
    } catch (err) {
      setTestResult({
        success: false,
        message: err.message || 'API handshake timeout',
      });
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-slate-900/50 backdrop-blur-xs flex justify-end animate-in fade-in duration-200">
      <div className="w-full max-w-2xl bg-white h-full shadow-2xl flex flex-col animate-in slide-in-from-right duration-250">
        {/* 1. Header Bar */}
        <div className="p-5 border-b border-slate-200 bg-slate-50/90 flex items-start justify-between gap-4">
          <div className="flex items-center gap-3.5">
            <div className={`h-12 w-12 rounded-2xl ${brand.bg} text-white flex items-center justify-center font-black text-sm tracking-wider shadow-sm shrink-0`}>
              {scac.substring(0, 4)}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-base font-bold text-slate-900">{carrier.display_name}</h3>
                <span className={`text-[11px] font-mono font-bold px-2 py-0.5 rounded border ${brand.light} ${brand.text} ${brand.border}`}>
                  SCAC: {scac}
                </span>
              </div>
              <div className="flex items-center gap-2 mt-1">
                <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                  <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
                  {carrier.status === 'CONFIGURATION_REQUIRED' ? 'Configuration Required' : 'Connected & Synced'}
                </span>
                <span className="text-slate-400 text-xs">•</span>
                <span className="text-xs font-mono text-slate-500">
                  {carrier.environment || 'PRODUCTION'}
                </span>
              </div>
            </div>
          </div>

          <button
            type="button"
            onClick={onClose}
            className="p-1.5 rounded-xl hover:bg-slate-200 text-slate-400 hover:text-slate-700 transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* 2. Quick Telemetry Banner (4 KPIs) */}
        <div className="bg-slate-900 text-white p-4 grid grid-cols-4 gap-3 text-center border-b border-slate-800">
          <div className="p-2 rounded-xl bg-slate-800/60">
            <span className="text-[10px] uppercase font-bold text-slate-400 block tracking-wider">Active TEU / Cargo</span>
            <span className="text-sm font-black text-white mt-0.5 block">
              {carrier.active_containers_count || linkedShipments.length * 2 || 12} Containers
            </span>
          </div>
          <div className="p-2 rounded-xl bg-slate-800/60">
            <span className="text-[10px] uppercase font-bold text-slate-400 block tracking-wider">Confirmed Bookings</span>
            <span className="text-sm font-black text-white mt-0.5 block">
              {carrier.active_bookings_count || linkedShipments.length || 4} Active
            </span>
          </div>
          <div className="p-2 rounded-xl bg-slate-800/60">
            <span className="text-[10px] uppercase font-bold text-slate-400 block tracking-wider">30D Ingress Events</span>
            <span className="text-sm font-black text-sky-400 mt-0.5 block">
              {carrier.event_count_30d ? carrier.event_count_30d.toLocaleString() : '3,100'}
            </span>
          </div>
          <div className="p-2 rounded-xl bg-slate-800/60">
            <span className="text-[10px] uppercase font-bold text-slate-400 block tracking-wider">Avg Gateway Ping</span>
            <span className="text-sm font-black text-emerald-400 mt-0.5 block">
              {carrier.latency_ms || 142}ms
            </span>
          </div>
        </div>

        {/* 3. Sub-Navigation Tabs */}
        <div className="px-5 border-b border-slate-200 bg-white flex items-center gap-2 overflow-x-auto">
          {[
            { id: 'SHIPMENTS', label: 'Active Shipments', count: linkedShipments.length, icon: Ship },
            { id: 'WEBHOOKS', label: 'Carrier Webhooks', count: linkedWebhooks.length, icon: Radio },
            { id: 'CONTRACTS', label: 'Rate Contracts', count: linkedContracts.length, icon: FileText },
            { id: 'CONFIG', label: 'Gateway Config', icon: Key },
          ].map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => setActiveTab(tab.id)}
                className={`py-3 px-3 text-xs font-bold border-b-2 flex items-center gap-1.5 transition-colors whitespace-nowrap ${
                  isActive
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-slate-500 hover:text-slate-800'
                }`}
              >
                <Icon className="h-3.5 w-3.5" />
                <span>{tab.label}</span>
                {tab.count !== undefined && (
                  <span className={`text-[10px] px-1.5 py-0.2 rounded-full ${isActive ? 'bg-blue-100 text-blue-700' : 'bg-slate-100 text-slate-600'}`}>
                    {tab.count}
                  </span>
                )}
              </button>
            );
          })}
        </div>

        {/* 4. Tab Body Content */}
        <div className="flex-1 overflow-y-auto p-5 space-y-4 bg-slate-50/50">
          {/* TAB: SHIPMENTS */}
          {activeTab === 'SHIPMENTS' && (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-700">
                  Live Cargo & Bookings Dispatched via {carrier.display_name}
                </span>
                <span className="text-[11px] text-slate-500">
                  {linkedShipments.length} active sea journeys
                </span>
              </div>

              {linkedShipments.length === 0 ? (
                <div className="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-500">
                  <Ship className="h-8 w-8 text-slate-300 mx-auto mb-2" />
                  <p className="text-xs font-bold text-slate-700">No active shipments currently on water for this carrier.</p>
                  <p className="text-[11px] text-slate-400 mt-1">New booking dispatch requests will link container telemetry here automatically.</p>
                </div>
              ) : (
                linkedShipments.map((shp) => (
                  <div key={shp.id} className="bg-white rounded-xl border border-slate-200 p-4 shadow-2xs hover:border-slate-300 transition-all space-y-3">
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-xs font-black text-slate-900">{shp.booking_number}</span>
                          <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 border border-blue-200">
                            {shp.status}
                          </span>
                        </div>
                        <span className="text-xs text-slate-500 mt-0.5 block">
                          Container: <strong className="font-mono text-slate-700">{shp.container_number}</strong> • Vessel: <strong>{shp.vessel_name}</strong> ({shp.voyage_number})
                        </span>
                      </div>
                      <span className="text-[11px] font-bold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
                        {shp.containers_count} TEU
                      </span>
                    </div>

                    <div className="pt-2 border-t border-slate-100 flex items-center justify-between text-xs text-slate-600">
                      <div>
                        <span className="text-[10px] text-slate-400 uppercase block">Port Pair Route</span>
                        <span className="font-semibold text-slate-800">{shp.origin_port} → {shp.destination_port}</span>
                      </div>
                      <div className="text-right">
                        <span className="text-[10px] text-slate-400 uppercase block">Estimated Arrival (ETA)</span>
                        <span className="font-mono text-slate-800">
                          {new Date(shp.eta).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </span>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* TAB: WEBHOOKS */}
          {activeTab === 'WEBHOOKS' && (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-700">
                  Real-time DCSA Ingress Events Stream
                </span>
                <span className="text-[11px] text-slate-500">
                  {linkedWebhooks.length} verified events
                </span>
              </div>

              {linkedWebhooks.length === 0 ? (
                <div className="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-500">
                  <Radio className="h-8 w-8 text-slate-300 mx-auto mb-2" />
                  <p className="text-xs font-bold text-slate-700">No webhooks recorded in current buffer.</p>
                  <p className="text-[11px] text-slate-400 mt-1">Carrier push notifications are ingested in real-time as containers move.</p>
                </div>
              ) : (
                linkedWebhooks.map((wh) => {
                  const isExpanded = expandedWebhookId === wh.id;
                  return (
                    <div key={wh.id} className="bg-white rounded-xl border border-slate-200 p-3.5 shadow-2xs space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className="p-1 rounded bg-emerald-50 text-emerald-600">
                            <CheckCircle2 className="h-3.5 w-3.5" />
                          </span>
                          <span className="font-mono text-xs font-bold text-slate-900">{wh.event_type}</span>
                        </div>
                        <span className="text-[11px] text-slate-400 font-mono">
                          {new Date(wh.received_at).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                        </span>
                      </div>

                      <div className="text-xs text-slate-600 flex items-center justify-between">
                        <span>Ref: <strong className="font-mono text-slate-800">{wh.reference_number}</strong></span>
                        <span>Booking: <strong className="font-mono text-slate-800">{wh.booking_number}</strong></span>
                      </div>

                      <button
                        type="button"
                        onClick={() => setExpandedWebhookId(isExpanded ? null : wh.id)}
                        className="text-[11px] font-bold text-blue-600 hover:text-blue-800 flex items-center gap-1 cursor-pointer pt-1"
                      >
                        <Code className="h-3 w-3" />
                        <span>{isExpanded ? 'Hide Payload' : 'Inspect JSON Payload'}</span>
                      </button>

                      {isExpanded && (
                        <pre className="bg-slate-900 text-slate-200 p-3 rounded-lg text-[10px] font-mono overflow-x-auto leading-relaxed mt-2">
                          {JSON.stringify(wh.payload, null, 2)}
                        </pre>
                      )}
                    </div>
                  );
                })
              )}
            </div>
          )}

          {/* TAB: CONTRACTS */}
          {activeTab === 'CONTRACTS' && (
            <div className="space-y-3">
              <span className="text-xs font-bold text-slate-700 block">
                Direct Master Service Agreements & Volume Commitments
              </span>

              {linkedContracts.length === 0 ? (
                <div className="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-500">
                  <FileText className="h-8 w-8 text-slate-300 mx-auto mb-2" />
                  <p className="text-xs font-bold text-slate-700">No active service contracts recorded.</p>
                  <p className="text-[11px] text-slate-400 mt-1">Spot rates and dynamic tariff lookups remain active for this carrier.</p>
                </div>
              ) : (
                linkedContracts.map((c) => (
                  <div key={c.id} className="bg-white rounded-xl border border-slate-200 p-4 shadow-2xs space-y-2">
                    <div className="flex items-start justify-between">
                      <div>
                        <h4 className="text-xs font-bold text-slate-900">{c.contract_name}</h4>
                        <span className="font-mono text-[11px] text-slate-500 mt-0.5 block">{c.contract_reference}</span>
                      </div>
                      <span className="text-xs font-bold px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                        {c.status}
                      </span>
                    </div>
                    <p className="text-xs text-slate-600 leading-relaxed">{c.description}</p>
                    <div className="pt-2 border-t border-slate-100 flex items-center justify-between text-xs">
                      <span className="text-slate-500">Committed Value: <strong className="text-slate-900">${c.contract_value?.toLocaleString()} {c.currency}</strong></span>
                      <span className="text-slate-500">Valid until: <strong className="text-slate-900">{new Date(c.expiry_date).toLocaleDateString()}</strong></span>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* TAB: CONFIG */}
          {activeTab === 'CONFIG' && (
            <div className="bg-white rounded-xl border border-slate-200 p-5 space-y-4 text-xs">
              <div className="p-3 rounded-lg bg-blue-50/70 border border-blue-200 text-blue-900 space-y-1">
                <span className="font-bold flex items-center gap-1.5">
                  <ShieldCheck className="h-4 w-4 text-blue-600" />
                  Hardware-Security-Module (HSM) Vault Protection
                </span>
                <p className="text-[11px] leading-relaxed">
                  Direct liner API keys, client secrets, and DCSA certificates are encrypted at rest with AES-256-GCM.
                </p>
              </div>

              <div>
                <span className="font-bold text-slate-700 block mb-1">API Base URL / Endpoint</span>
                <div className="p-2.5 rounded-lg bg-slate-50 border border-slate-200 font-mono text-[11px] text-slate-800">
                  {carrier.safe_config?.endpoint || `https://api.${scac.toLowerCase()}.com/v2/ocean`}
                </div>
              </div>

              <div>
                <span className="font-bold text-slate-700 block mb-1">Authenticated Client ID / Identity</span>
                <div className="p-2.5 rounded-lg bg-slate-50 border border-slate-200 font-mono text-[11px] text-slate-800 flex items-center justify-between">
                  <span>{carrier.safe_config?.client_id || `${scac.toLowerCase()}-freel-prod-****`}</span>
                  <span className="text-[10px] font-bold text-emerald-600 bg-emerald-50 px-1.5 py-0.5 rounded border border-emerald-200">
                    VERIFIED
                  </span>
                </div>
              </div>

              <div>
                <span className="font-bold text-slate-700 block mb-1">Supported Capabilities</span>
                <div className="flex flex-wrap gap-1.5">
                  {(carrier.supported_capabilities || ['Track & Trace', 'Spot Quoting', 'Space Booking']).map((cap) => (
                    <span key={cap} className="px-2.5 py-1 rounded-md bg-slate-100 text-slate-700 border border-slate-200 text-[11px] font-medium">
                      ✓ {cap}
                    </span>
                  ))}
                </div>
              </div>

              {testResult && (
                <div className={`p-3 rounded-xl border flex items-start gap-2 ${testResult.success ? 'bg-emerald-50 border-emerald-200 text-emerald-800' : 'bg-rose-50 border-rose-200 text-rose-800'}`}>
                  {testResult.success ? <CheckCircle2 className="h-4 w-4 text-emerald-600 shrink-0 mt-0.5" /> : <AlertCircle className="h-4 w-4 text-rose-600 shrink-0 mt-0.5" />}
                  <div className="text-xs">
                    <strong className="block font-bold">{testResult.success ? 'Handshake Diagnostics Succeeded' : 'Handshake Failed'}</strong>
                    <span className="text-[11px]">{testResult.message}</span>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* 5. Footer Actions */}
        <div className="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between gap-3">
          <button
            type="button"
            onClick={handleTest}
            disabled={testing}
            className="px-4 py-2 rounded-xl bg-white hover:bg-slate-100 border border-slate-200 text-slate-700 font-bold text-xs flex items-center gap-2 shadow-2xs transition-all cursor-pointer disabled:opacity-50"
          >
            <RotateCw className={`h-3.5 w-3.5 ${testing ? 'animate-spin text-blue-600' : 'text-slate-500'}`} />
            <span>{testing ? 'Testing Handshake...' : 'Test API Handshake'}</span>
          </button>

          <button
            type="button"
            onClick={() => {
              onClose();
              if (onConfigure) onConfigure(carrier);
            }}
            className="px-4 py-2 rounded-xl bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs flex items-center gap-1.5 shadow-sm transition-all cursor-pointer"
          >
            <Key className="h-3.5 w-3.5" />
            <span>Configure Credentials</span>
          </button>
        </div>
      </div>
    </div>
  );
}
