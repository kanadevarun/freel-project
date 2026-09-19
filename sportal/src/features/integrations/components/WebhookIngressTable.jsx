import React, { useState } from 'react';
import {
  Webhook,
  CheckCircle2,
  AlertTriangle,
  Clock,
  ChevronRight,
  Eye,
  RefreshCw,
  Code,
  Check,
  Copy,
  Search,
  Filter
} from 'lucide-react';

export function WebhookIngressTable({ webhooks = [], loading, onRefresh }) {
  const [selectedWebhook, setSelectedWebhook] = useState(null);
  const [copied, setCopied] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [eventFilter, setEventFilter] = useState('ALL');

  const filtered = webhooks.filter((w) => {
    const matchesSearch =
      !searchQuery ||
      w.carrier_scac?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      w.event_type?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      w.reference_number?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesFilter =
      eventFilter === 'ALL' || w.event_type === eventFilter || w.status === eventFilter;
    return matchesSearch && matchesFilter;
  });

  const handleCopy = (text) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '—';
    try {
      return new Date(dateStr).toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
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
      {/* Table Header & Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100">
        <div>
          <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
            <Webhook className="h-4 w-4 text-emerald-600" />
            Live Webhook Ingress & Delivery Stream
          </h3>
          <p className="text-xs text-slate-500">
            Real-time asynchronous carrier and gateway webhooks received via LogisticsHQ webhook endpoints
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="h-3.5 w-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="Search reference, SCAC..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-8 pl-8 pr-3 text-xs rounded-lg border border-slate-200 bg-white placeholder:text-slate-400 focus:outline-none focus:border-navy-900 w-48"
            />
          </div>

          <select
            value={eventFilter}
            onChange={(e) => setEventFilter(e.target.value)}
            className="h-8 px-2 text-xs rounded-lg border border-slate-200 bg-white text-slate-700 focus:outline-none focus:border-navy-900"
          >
            <option value="ALL">All Events</option>
            <option value="EQUIPMENT_GATE_IN">Gate In</option>
            <option value="VESSEL_DEPARTURE">Vessel Departure</option>
            <option value="VESSEL_ARRIVAL">Vessel Arrival</option>
            <option value="CONTAINER_DISCHARGED">Discharged</option>
            <option value="CUSTOMS_CLEARED">Customs Cleared</option>
            <option value="PROCESSED">Processed</option>
          </select>

          <button
            type="button"
            onClick={onRefresh}
            disabled={loading}
            className="h-8 px-2.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 text-slate-600 text-xs font-medium flex items-center gap-1.5 transition-colors"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin text-blue-600' : ''}`} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto rounded-lg border border-slate-200">
        <table className="w-full text-left text-xs">
          <thead>
            <tr className="bg-slate-50 border-b border-slate-200 text-slate-500 font-semibold uppercase tracking-wider text-[11px]">
              <th className="py-2.5 px-3">Carrier / Provider</th>
              <th className="py-2.5 px-3">Event Type</th>
              <th className="py-2.5 px-3">Reference / Container</th>
              <th className="py-2.5 px-3">Status</th>
              <th className="py-2.5 px-3">Received At</th>
              <th className="py-2.5 px-3 text-right">Payload</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {loading && webhooks.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-8 text-center text-slate-400">
                  Loading webhook stream...
                </td>
              </tr>
            ) : filtered.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-8 text-center text-slate-500">
                  {webhooks.length === 0
                    ? 'Webhook has not received an event yet.'
                    : 'No webhook events match the selected criteria.'}
                </td>
              </tr>
            ) : (
              filtered.map((wh) => (
                <tr key={wh.id} className="hover:bg-slate-50/75 transition-colors">
                  <td className="py-2.5 px-3 font-semibold text-slate-900">
                    <span className="font-mono bg-slate-100 text-slate-700 px-1.5 py-0.5 rounded text-[11px] mr-1.5">
                      {wh.carrier_scac || 'GENERIC'}
                    </span>
                  </td>
                  <td className="py-2.5 px-3">
                    <span className="font-mono text-slate-800 font-medium">{wh.event_type}</span>
                  </td>
                  <td className="py-2.5 px-3 font-mono text-slate-600 font-medium">
                    {wh.reference_number || '—'}
                  </td>
                  <td className="py-2.5 px-3">
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">
                      <CheckCircle2 className="h-3 w-3 text-emerald-600" />
                      {wh.status || 'PROCESSED'}
                    </span>
                  </td>
                  <td className="py-2.5 px-3 text-slate-500 font-mono text-[11px]">
                    {formatDate(wh.received_at)}
                  </td>
                  <td className="py-2.5 px-3 text-right">
                    <button
                      type="button"
                      onClick={() => setSelectedWebhook(wh)}
                      className="inline-flex items-center gap-1 text-xs font-semibold text-navy-900 hover:text-blue-700 transition-colors"
                    >
                      <Eye className="h-3.5 w-3.5" />
                      <span>Inspect</span>
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Payload Inspection Modal */}
      {selectedWebhook && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-xs p-4">
          <div className="bg-white rounded-xl shadow-xl border border-slate-200 w-full max-w-xl overflow-hidden flex flex-col max-h-[85vh] animate-in fade-in zoom-in-95 duration-150">
            <div className="p-4 border-b border-slate-200 flex items-center justify-between bg-slate-50">
              <div>
                <h4 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                  <Code className="h-4 w-4 text-navy-900" />
                  Webhook Payload — {selectedWebhook.event_type}
                </h4>
                <p className="text-[11px] text-slate-500 mt-0.5 font-mono">
                  ID: #{selectedWebhook.id} • Carrier: {selectedWebhook.carrier_scac} • Received:{' '}
                  {formatDate(selectedWebhook.received_at)}
                </p>
              </div>
              <button
                type="button"
                onClick={() => setSelectedWebhook(null)}
                className="text-slate-400 hover:text-slate-700 text-sm font-bold p-1"
              >
                ✕
              </button>
            </div>

            <div className="p-4 overflow-y-auto flex-1 bg-slate-950 font-mono text-xs text-emerald-400">
              <pre className="whitespace-pre-wrap leading-relaxed">
                {(() => {
                  try {
                    const raw = selectedWebhook.payload_json || selectedWebhook.payload_preview || '{}';
                    const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw;
                    return JSON.stringify(parsed, null, 2);
                  } catch {
                    return selectedWebhook.payload_preview || selectedWebhook.payload_json || '{}';
                  }
                })()}
              </pre>
            </div>

            <div className="p-3 border-t border-slate-200 bg-slate-50 flex items-center justify-between">
              <button
                type="button"
                onClick={() =>
                  handleCopy(
                    typeof selectedWebhook.payload_json === 'string'
                      ? selectedWebhook.payload_json
                      : JSON.stringify(selectedWebhook.payload_json, null, 2)
                  )
                }
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 text-xs font-medium text-slate-700 transition-colors"
              >
                {copied ? <Check className="h-3.5 w-3.5 text-emerald-600" /> : <Copy className="h-3.5 w-3.5" />}
                <span>{copied ? 'Copied' : 'Copy Payload JSON'}</span>
              </button>

              <button
                type="button"
                onClick={() => setSelectedWebhook(null)}
                className="px-4 py-1.5 rounded-lg bg-navy-900 hover:bg-navy-800 text-white text-xs font-semibold transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
