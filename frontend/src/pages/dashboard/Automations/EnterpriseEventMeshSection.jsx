import React, { useState, useEffect, useCallback } from 'react';
import {
  Activity,
  AlertOctagon,
  ArrowRight,
  CheckCircle2,
  Clock,
  ExternalLink,
  Filter,
  Layers,
  Play,
  RefreshCw,
  RotateCcw,
  Search,
  ShieldAlert,
  ShieldCheck,
  Sparkles,
  Zap,
  Radio,
  FileText,
  AlertTriangle,
  Send
} from 'lucide-react';
import { enterpriseService } from '../../../services/enterpriseService';

export default function EnterpriseEventMeshSection() {
  const [overview, setOverview] = useState(null);
  const [events, setEvents] = useState([]);
  const [deadLetters, setDeadLetters] = useState([]);
  const [rules, setRules] = useState([]);
  const [activeSubTab, setActiveSubTab] = useState('events'); // 'events' | 'dead_letters' | 'rules' | 'simulate'
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState('ALL');
  const [actionSuccess, setActionSuccess] = useState(null);
  const [selectedEvent, setSelectedEvent] = useState(null);

  // Simulation form
  const [simulateType, setSimulateType] = useState('SHIPMENT_DELAY_DETECTED');
  const [simulateEntity, setSimulateEntity] = useState('SH-9082');
  const [simulatePayload, setSimulatePayload] = useState('{"delay_hours": 18, "reason": "Port congestion at USLAX"}');
  const [isSimulating, setIsSimulating] = useState(false);

  const loadData = useCallback(async () => {
    try {
      setIsLoading(true);
      const [ovRes, evRes, dlRes, rlRes] = await Promise.all([
        enterpriseService.getEventMeshOverview().catch(() => null),
        enterpriseService.listMeshEvents({ limit: 50 }).catch(() => ({ events: [], total: 0 })),
        enterpriseService.getMeshDeadLetters(20).catch(() => ({ data: [] })),
        enterpriseService.listMeshRoutingRules().catch(() => ({ data: [] })),
      ]);

      if (ovRes?.data) setOverview(ovRes.data);
      if (evRes?.events) setEvents(evRes.events);
      if (dlRes?.data) setDeadLetters(dlRes.data);
      if (rlRes?.data) setRules(rlRes.data);
    } catch (err) {
      console.error('Failed loading event mesh data', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleSimulateEvent = async (e) => {
    e.preventDefault();
    try {
      setIsSimulating(true);
      let parsed = {};
      try {
        parsed = JSON.parse(simulatePayload);
      } catch (err) {
        alert('Invalid JSON in payload: ' + err.message);
        setIsSimulating(false);
        return;
      }

      const res = await enterpriseService.ingestMeshEvent({
        event_id: `sim-evt-${Date.now()}`,
        event_type: simulateType,
        source_module: 'simulation_control',
        entity_type: simulateType.includes('SHIPMENT') ? 'SHIPMENT' : (simulateType.includes('RFQ') ? 'RFQ' : 'INVOICE'),
        entity_id: simulateEntity,
        payload: parsed,
      });

      if (res?.data) {
        setActionSuccess(`Event ingested successfully! Routed to workflow: ${res.data.workflow?.workflow_type || 'COMPLETED'}`);
        setTimeout(() => setActionSuccess(null), 6000);
        await loadData();
      }
    } catch (err) {
      alert(`Simulation failed: ${err.message}`);
    } finally {
      setIsSimulating(false);
    }
  };

  const handleReplayDeadLetter = async (eventId) => {
    try {
      const res = await enterpriseService.replayMeshDeadLetter(eventId);
      if (res?.data) {
        setActionSuccess(`Dead-letter event ${eventId} replayed and processed!`);
        setTimeout(() => setActionSuccess(null), 5000);
        await loadData();
      }
    } catch (err) {
      alert(`Replay failed: ${err.message}`);
    }
  };

  const filteredEvents = events.filter((ev) => {
    if (filterType !== 'ALL' && ev.processing_status !== filterType) return false;
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      ev.event_id?.toLowerCase().includes(q) ||
      ev.event_type?.toLowerCase().includes(q) ||
      ev.entity_id?.toLowerCase().includes(q)
    );
  });

  return (
    <div className="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden my-6">
      {/* ── Header Strip ── */}
      <div className="px-6 py-4 border-b border-slate-200 bg-slate-50 flex items-center justify-between flex-wrap gap-3">
        <div className="flex items-center space-x-3">
          <div className="p-2 bg-indigo-600 text-white rounded-lg shadow-xs">
            <Radio size={18} className="animate-pulse" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <h2 className="text-sm font-bold text-slate-900 tracking-tight">
                Enterprise Event Mesh & Autonomous Workflow Engine
              </h2>
              <span className="text-[10px] px-2 py-0.5 rounded-full font-bold bg-indigo-100 text-indigo-800 border border-indigo-200">
                PHASE 7.8 UNIFIED BUS
              </span>
            </div>
            <p className="text-xs text-slate-500 mt-0.5">
              Deterministic routing, deduplication, feedback loop suppression, and dead-letter governance across 7 operational domains.
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <button
            type="button"
            onClick={loadData}
            disabled={isLoading}
            className="px-3 py-1.5 bg-white border border-slate-300 hover:bg-slate-50 text-slate-700 rounded-lg text-xs font-semibold flex items-center space-x-1.5 shadow-xs transition-colors"
          >
            <RefreshCw size={13} className={isLoading ? 'animate-spin' : ''} />
            <span>Refresh Mesh</span>
          </button>
        </div>
      </div>

      {/* ── Success Alert Banner ── */}
      {actionSuccess && (
        <div className="px-6 py-2.5 bg-emerald-50 border-b border-emerald-200 text-emerald-800 text-xs font-medium flex items-center space-x-2">
          <CheckCircle2 size={15} className="text-emerald-600" />
          <span>{actionSuccess}</span>
        </div>
      )}

      {/* ── KPI Overview Strip ── */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-3 p-4 bg-slate-50/50 border-b border-slate-200 text-center">
        <div className="p-3 bg-white border border-slate-200 rounded-lg">
          <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">Total Received</div>
          <div className="text-lg font-bold text-slate-900 mt-0.5">
            {overview?.total_events_received || events.length}
          </div>
          <div className="text-[10px] text-slate-400 mt-0.5">Validated Schema</div>
        </div>

        <div className="p-3 bg-white border border-slate-200 rounded-lg">
          <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">Workflows Triggered</div>
          <div className="text-lg font-bold text-indigo-600 mt-0.5">
            {overview?.workflows_triggered || 0}
          </div>
          <div className="text-[10px] text-indigo-500 font-medium mt-0.5">Autonomous Operations</div>
        </div>

        <div className="p-3 bg-white border border-slate-200 rounded-lg">
          <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">Deduplicated</div>
          <div className="text-lg font-bold text-blue-600 mt-0.5">
            {overview?.deduplicated_count || 0}
          </div>
          <div className="text-[10px] text-slate-400 mt-0.5">Idempotency Guard</div>
        </div>

        <div className="p-3 bg-white border border-slate-200 rounded-lg">
          <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">Loops Suppressed</div>
          <div className="text-lg font-bold text-emerald-600 mt-0.5">
            {overview?.loops_suppressed_count || 0}
          </div>
          <div className="text-[10px] text-emerald-600 font-medium mt-0.5">Feedback Guard Active</div>
        </div>

        <div className="p-3 bg-white border border-slate-200 rounded-lg">
          <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">Dead-Letter Queue</div>
          <div className="text-lg font-bold text-rose-600 mt-0.5">
            {overview?.dead_letter_count || deadLetters.length}
          </div>
          <div className="text-[10px] text-slate-400 mt-0.5">Triage Required</div>
        </div>
      </div>

      {/* ── Sub Tabs ── */}
      <div className="flex border-b border-slate-200 px-6 bg-white gap-6">
        <button
          type="button"
          onClick={() => setActiveSubTab('events')}
          className={`py-3 text-xs font-bold border-b-2 transition-colors flex items-center gap-1.5 ${
            activeSubTab === 'events'
              ? 'border-indigo-600 text-indigo-600'
              : 'border-transparent text-slate-500 hover:text-slate-900'
          }`}
        >
          <Activity size={14} />
          <span>Ingested Events ({events.length})</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveSubTab('dead_letters')}
          className={`py-3 text-xs font-bold border-b-2 transition-colors flex items-center gap-1.5 ${
            activeSubTab === 'dead_letters'
              ? 'border-rose-600 text-rose-600'
              : 'border-transparent text-slate-500 hover:text-slate-900'
          }`}
        >
          <AlertOctagon size={14} />
          <span>Dead-Letter Queue ({deadLetters.length})</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveSubTab('rules')}
          className={`py-3 text-xs font-bold border-b-2 transition-colors flex items-center gap-1.5 ${
            activeSubTab === 'rules'
              ? 'border-indigo-600 text-indigo-600'
              : 'border-transparent text-slate-500 hover:text-slate-900'
          }`}
        >
          <Layers size={14} />
          <span>Routing Rules ({rules.length})</span>
        </button>

        <button
          type="button"
          onClick={() => setActiveSubTab('simulate')}
          className={`py-3 text-xs font-bold border-b-2 transition-colors flex items-center gap-1.5 ${
            activeSubTab === 'simulate'
              ? 'border-indigo-600 text-indigo-600'
              : 'border-transparent text-slate-500 hover:text-slate-900'
          }`}
        >
          <Send size={14} />
          <span>Simulate Real Event</span>
        </button>
      </div>

      {/* ── Sub Tab Content ── */}
      <div className="p-6">
        {/* TAB 1: INGESTED EVENTS */}
        {activeSubTab === 'events' && (
          <div>
            {/* Filter / Search Bar */}
            <div className="flex items-center justify-between gap-3 mb-4 flex-wrap">
              <div className="relative flex-1 min-w-[220px]">
                <Search size={14} className="absolute left-3 top-2.5 text-slate-400" />
                <input
                  type="text"
                  placeholder="Filter by Event ID, Type, or Entity..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-9 pr-3 py-1.5 bg-slate-50 border border-slate-200 rounded-lg text-xs text-slate-800 focus:outline-hidden focus:ring-1 focus:ring-indigo-500"
                />
              </div>

              <div className="flex items-center gap-2">
                <Filter size={13} className="text-slate-400" />
                <select
                  value={filterType}
                  onChange={(e) => setFilterType(e.target.value)}
                  className="bg-slate-50 border border-slate-200 rounded-lg text-xs px-2.5 py-1.5 text-slate-700"
                >
                  <option value="ALL">All Statuses</option>
                  <option value="WORKFLOW_ACTIVE">Active Workflow</option>
                  <option value="DEDUPLICATED">Deduplicated</option>
                  <option value="LOOP_SUPPRESSED">Loop Suppressed</option>
                  <option value="COMPLETED">Completed</option>
                </select>
              </div>
            </div>

            {/* Table */}
            {filteredEvents.length === 0 ? (
              <div className="text-center py-8 text-slate-400 text-xs">
                No events found matching current criteria.
              </div>
            ) : (
              <div className="border border-slate-200 rounded-lg overflow-x-auto">
                <table className="w-full text-left border-collapse text-xs">
                  <thead>
                    <tr className="bg-slate-50 border-b border-slate-200 text-slate-600 font-semibold">
                      <th className="py-2.5 px-3">Event ID</th>
                      <th className="py-2.5 px-3">Event Type</th>
                      <th className="py-2.5 px-3">Entity</th>
                      <th className="py-2.5 px-3">Target Workflow</th>
                      <th className="py-2.5 px-3">Status</th>
                      <th className="py-2.5 px-3">Occurred At</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {filteredEvents.map((ev) => (
                      <tr key={ev.event_id} className="hover:bg-slate-50/70 transition-colors">
                        <td className="py-2.5 px-3 font-mono text-slate-700 font-medium">
                          {ev.event_id}
                        </td>
                        <td className="py-2.5 px-3">
                          <span className="font-bold text-slate-900">{ev.event_type}</span>
                          <span className="block text-[10px] text-slate-400">{ev.source_module}</span>
                        </td>
                        <td className="py-2.5 px-3">
                          <span className="px-1.5 py-0.5 rounded-sm bg-slate-100 font-mono text-slate-700 text-[10px]">
                            {ev.entity_type} #{ev.entity_id}
                          </span>
                        </td>
                        <td className="py-2.5 px-3">
                          {ev.workflow_type ? (
                            <span className="text-indigo-700 font-semibold">{ev.workflow_type}</span>
                          ) : (
                            <span className="text-slate-400">—</span>
                          )}
                        </td>
                        <td className="py-2.5 px-3">
                          <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                            ev.processing_status === 'WORKFLOW_ACTIVE'
                              ? 'bg-indigo-100 text-indigo-800'
                              : ev.processing_status === 'DEDUPLICATED'
                              ? 'bg-blue-100 text-blue-800'
                              : ev.processing_status === 'LOOP_SUPPRESSED'
                              ? 'bg-emerald-100 text-emerald-800'
                              : 'bg-slate-100 text-slate-600'
                          }`}>
                            {ev.processing_status}
                          </span>
                        </td>
                        <td className="py-2.5 px-3 text-slate-500 text-[11px]">
                          {new Date(ev.occurred_at).toLocaleTimeString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* TAB 2: DEAD-LETTER QUEUE */}
        {activeSubTab === 'dead_letters' && (
          <div>
            <div className="mb-4 text-xs text-slate-500">
              Events that failed schema validation, contained malicious prompt injections, or encountered unrecoverable domain errors.
            </div>

            {deadLetters.length === 0 ? (
              <div className="p-8 text-center bg-slate-50 border border-dashed border-slate-200 rounded-xl text-slate-400 text-xs">
                <CheckCircle2 size={24} className="mx-auto mb-2 text-emerald-500" />
                Dead-letter queue is completely empty. All events processed normally.
              </div>
            ) : (
              <div className="space-y-3">
                {deadLetters.map((dl) => (
                  <div key={dl.event_id} className="p-3.5 bg-rose-50/40 border border-rose-200 rounded-lg flex items-center justify-between">
                    <div>
                      <div className="flex items-center space-x-2">
                        <span className="font-mono text-xs font-bold text-rose-900">{dl.event_id}</span>
                        <span className="text-[10px] px-2 py-0.2 rounded-full font-bold bg-rose-100 text-rose-800">
                          {dl.event_type}
                        </span>
                      </div>
                      <p className="text-xs text-rose-700 mt-1 font-medium">
                        Reason: {dl.failure_reason}
                      </p>
                      <span className="text-[10px] text-slate-400 mt-0.5 block">
                        Received: {new Date(dl.dead_lettered_at).toLocaleString()}
                      </span>
                    </div>

                    <button
                      type="button"
                      onClick={() => handleReplayDeadLetter(dl.event_id)}
                      className="px-3 py-1.5 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-xs font-bold flex items-center gap-1 shadow-xs transition-colors"
                    >
                      <RotateCcw size={12} />
                      <span>Replay</span>
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* TAB 3: ROUTING RULES */}
        {activeSubTab === 'rules' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {rules.map((r, i) => (
              <div key={i} className="p-4 border border-slate-200 rounded-lg bg-slate-50/50">
                <div className="flex items-center justify-between">
                  <span className="font-bold text-xs text-slate-900">{r.event_type}</span>
                  <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-indigo-100 text-indigo-800">
                    {r.target_workflow_type}
                  </span>
                </div>
                <p className="text-xs text-slate-600 mt-1.5 leading-relaxed">{r.description}</p>
                <div className="mt-3 pt-2 border-t border-slate-200 flex items-center justify-between text-[11px] text-slate-500">
                  <span>Priority: <strong className="text-slate-700">{r.default_priority}</strong></span>
                  <span>Min Autonomy: <strong className="text-slate-700">{r.min_autonomy_level}</strong></span>
                  <span>Cooldown: <strong className="text-slate-700">{r.cooldown_seconds}s</strong></span>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* TAB 4: SIMULATE EVENT */}
        {activeSubTab === 'simulate' && (
          <form onSubmit={handleSimulateEvent} className="max-w-xl space-y-4">
            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">Select Event Type</label>
              <select
                value={simulateType}
                onChange={(e) => setSimulateType(e.target.value)}
                className="w-full bg-white border border-slate-200 rounded-lg text-xs px-3 py-2 text-slate-800"
              >
                <option value="SHIPMENT_DELAY_DETECTED">SHIPMENT_DELAY_DETECTED (Shipment Recovery)</option>
                <option value="TEMPERATURE_EXCURSION">TEMPERATURE_EXCURSION (Cold Chain Exception)</option>
                <option value="RFQ_CREATED">RFQ_CREATED (Commercial Quote-to-Cash)</option>
                <option value="INVOICE_OVERDUE">INVOICE_OVERDUE (Financial Collection)</option>
                <option value="MARGIN_EROSION_DETECTED">MARGIN_EROSION_DETECTED (Revenue Optimization)</option>
                <option value="CONTRACT_EXPIRATION_APPROACHING">CONTRACT_EXPIRATION_APPROACHING (Contract & Risk)</option>
                <option value="CUSTOMS_REGULATORY_HOLD">CUSTOMS_REGULATORY_HOLD (Compliance Risk)</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">Target Entity Identifier</label>
              <input
                type="text"
                value={simulateEntity}
                onChange={(e) => setSimulateEntity(e.target.value)}
                className="w-full bg-white border border-slate-200 rounded-lg text-xs px-3 py-2 text-slate-800 font-mono"
                placeholder="e.g. SH-9082, RFQ-401, CTR-2026"
              />
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-700 mb-1">Payload JSON (Untrusted External Data)</label>
              <textarea
                value={simulatePayload}
                onChange={(e) => setSimulatePayload(e.target.value)}
                rows={3}
                className="w-full bg-white border border-slate-200 rounded-lg text-xs px-3 py-2 text-slate-800 font-mono"
              />
            </div>

            <button
              type="submit"
              disabled={isSimulating}
              className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-bold flex items-center space-x-2 shadow-xs transition-colors"
            >
              <Send size={14} className={isSimulating ? 'animate-pulse' : ''} />
              <span>{isSimulating ? 'Ingesting Event...' : 'Ingest & Trigger Workflow'}</span>
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
