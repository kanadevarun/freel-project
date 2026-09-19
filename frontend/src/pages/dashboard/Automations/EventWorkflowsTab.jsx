import { useState, useEffect, useCallback } from 'react';
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  CheckCircle2,
  Clock,
  ExternalLink,
  Eye,
  Filter,
  Layers,
  Play,
  RefreshCw,
  RotateCcw,
  Search,
  ShieldAlert,
  ShieldCheck,
  Sparkles,
  X,
  XCircle,
  FileText,
  Mail,
} from 'lucide-react';
import { eventWorkflowsService } from '../../../services/eventWorkflowsService';
import './EventWorkflowsTab.css';

export default function EventWorkflowsTab({ onNavigateToApprovals }) {
  const [subTab, setSubTab] = useState('workflows'); // 'workflows' | 'events'
  const [overview, setOverview] = useState(null);
  const [workflows, setWorkflows] = useState([]);
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionSuccess, setActionSuccess] = useState(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [search, setSearch] = useState('');

  // Selected for drawer
  const [selectedWorkflow, setSelectedWorkflow] = useState(null);
  const [selectedEvent, setSelectedEvent] = useState(null);

  // Simulation Modal
  const [isSimulateModalOpen, setIsSimulateModalOpen] = useState(false);
  const [simulating, setSimulating] = useState(false);
  const [simulateTemplate, setSimulateTemplate] = useState('shipment_delay');

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const [ovRes, wfRes, evRes] = await Promise.all([
        eventWorkflowsService.getOverview().catch(() => null),
        eventWorkflowsService.listWorkflows({ limit: 50 }).catch(() => ({ workflows: [] })),
        eventWorkflowsService.listEvents({ limit: 50 }).catch(() => ({ events: [] })),
      ]);

      if (ovRes) setOverview(ovRes);
      if (wfRes?.workflows) setWorkflows(wfRes.workflows);
      if (evRes?.events) setEvents(evRes.events);
    } catch (err) {
      console.error('Failed to load event workflow data:', err);
      setError('Unable to load event-driven workflows from the system.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleRetry = async (wfId) => {
    try {
      const updated = await eventWorkflowsService.retryWorkflow(wfId);
      setActionSuccess(`Workflow #${wfId} successfully re-evaluated by AI.`);
      fetchData();
      if (selectedWorkflow && selectedWorkflow.id === wfId) {
        setSelectedWorkflow(updated);
      }
    } catch (err) {
      alert(`Retry failed: ${err.message}`);
    }
  };

  const handleCancel = async (wfId) => {
    if (!window.confirm(`Cancel workflow #${wfId}?`)) return;
    try {
      const updated = await eventWorkflowsService.cancelWorkflow(wfId);
      setActionSuccess(`Workflow #${wfId} has been cancelled.`);
      fetchData();
      if (selectedWorkflow && selectedWorkflow.id === wfId) {
        setSelectedWorkflow(updated);
      }
    } catch (err) {
      alert(`Cancellation failed: ${err.message}`);
    }
  };

  const handleSimulate = async () => {
    setSimulating(true);
    try {
      let payload = {};
      let eventType = '';
      let sourceModule = '';
      let sourceRecordType = '';
      let sourceRecordId = '101';

      if (simulateTemplate === 'shipment_delay') {
        eventType = 'shipment.milestone_missed';
        sourceModule = 'SHIPMENTS';
        sourceRecordType = 'SHIPMENT';
        payload = {
          milestone_code: 'BERTH_ARRIVAL',
          severity: 'HIGH',
          delay_hours: 36,
          exception_type: 'Congestion Delay',
          revised_eta: '2026-09-15T18:00:00Z',
        };
      } else if (simulateTemplate === 'invoice_overdue') {
        eventType = 'invoice.overdue';
        sourceModule = 'FINANCE';
        sourceRecordType = 'INVOICE';
        payload = {
          outstanding_amount: 18450.0,
          days_overdue: 21,
          is_high_value: true,
        };
      } else if (simulateTemplate === 'contract_expiry') {
        eventType = 'contract.approaching_expiry';
        sourceModule = 'CONTRACTS';
        sourceRecordType = 'CONTRACT';
        payload = {
          days_until_expiry: 14,
          annual_contract_value: 120000.0,
          requires_compliance_audit: true,
        };
      } else if (simulateTemplate === 'lead_qualified') {
        eventType = 'lead.qualified';
        sourceModule = 'LEADS';
        sourceRecordType = 'LEAD';
        payload = {
          lead_score: 92,
          budget_usd: 85000,
          estimated_lanes: 4,
        };
      }

      const res = await eventWorkflowsService.simulateEvent({
        event_type: eventType,
        source_module: sourceModule,
        source_record_type: sourceRecordType,
        source_record_id: sourceRecordId,
        payload: payload,
      });

      setIsSimulateModalOpen(false);
      setActionSuccess(`Simulated event '${eventType}' ingested. Workflow #${res.id} created.`);
      fetchData();
      setSelectedWorkflow(res);
    } catch (err) {
      alert(`Simulation failed: ${err.message}`);
    } finally {
      setSimulating(false);
    }
  };

  // Filter workflows
  const filteredWorkflows = workflows.filter((w) => {
    if (statusFilter && w.status !== statusFilter) return false;
    if (typeFilter && w.workflow_type !== typeFilter) return false;
    if (search) {
      const q = search.toLowerCase();
      return (
        w.workflow_type.toLowerCase().includes(q) ||
        w.trigger_event_type.toLowerCase().includes(q) ||
        w.source_record_id.toLowerCase().includes(q) ||
        (w.ai_summary && w.ai_summary.toLowerCase().includes(q))
      );
    }
    return true;
  });

  const getUrgencyBadge = (urgency) => {
    const u = (urgency || 'MEDIUM').toUpperCase();
    let cls = 'ew-badge-urgency-medium';
    if (u === 'CRITICAL') cls = 'ew-badge-urgency-critical';
    if (u === 'HIGH') cls = 'ew-badge-urgency-high';
    if (u === 'LOW') cls = 'ew-badge-urgency-low';
    return <span className={`ew-badge ${cls}`}>{u}</span>;
  };

  const getStatusBadge = (status) => {
    const s = (status || 'RUNNING').toUpperCase();
    let cls = 'ew-badge-status-running';
    if (s === 'AWAITING_APPROVAL') cls = 'ew-badge-status-awaiting';
    if (s === 'COMPLETED') cls = 'ew-badge-status-completed';
    if (s === 'FAILED') cls = 'ew-badge-status-failed';
    if (s === 'CANCELLED') cls = 'ew-badge-status-cancelled';
    return <span className={`ew-badge ${cls}`}>{s.replace('_', ' ')}</span>;
  };

  return (
    <div className="event-workflows-tab">
      {/* Toast Alert */}
      {actionSuccess && (
        <div className="auto-toast-banner" role="status">
          <CheckCircle2 size={16} />
          <span>{actionSuccess}</span>
          <button className="btn-close-toast" onClick={() => setActionSuccess(null)}>
            <X size={14} />
          </button>
        </div>
      )}

      {/* KPI Overview Cards */}
      <div className="ew-kpi-grid">
        <div className="ew-kpi-card">
          <div className="ew-kpi-icon-wrap ew-kpi-icon-blue">
            <Activity size={22} />
          </div>
          <div className="ew-kpi-body">
            <span className="ew-kpi-val">{overview?.total_events_ingested || events.length}</span>
            <span className="ew-kpi-label">Ingested Domain Events</span>
          </div>
        </div>

        <div className="ew-kpi-card">
          <div className="ew-kpi-icon-wrap ew-kpi-icon-amber">
            <Clock size={22} />
          </div>
          <div className="ew-kpi-body">
            <span className="ew-kpi-val">{overview?.active_workflows_count || workflows.filter(w => w.status === 'RUNNING' || w.status === 'AWAITING_APPROVAL').length}</span>
            <span className="ew-kpi-label">Active AI Workflows</span>
          </div>
        </div>

        <div className="ew-kpi-card">
          <div className="ew-kpi-icon-wrap ew-kpi-icon-purple">
            <ShieldCheck size={22} />
          </div>
          <div className="ew-kpi-body">
            <span className="ew-kpi-val">{overview?.awaiting_approval_count || workflows.filter(w => w.status === 'AWAITING_APPROVAL').length}</span>
            <span className="ew-kpi-label">Awaiting Human Approval</span>
          </div>
        </div>

        <div className="ew-kpi-card">
          <div className="ew-kpi-icon-wrap ew-kpi-icon-green">
            <Sparkles size={22} />
          </div>
          <div className="ew-kpi-body">
            <span className="ew-kpi-val">{overview?.deduplicated_events_count || 0}</span>
            <span className="ew-kpi-label">Deduplicated & Prevented</span>
          </div>
        </div>
      </div>

      {/* Controls Bar */}
      <div className="ew-controls-bar">
        <div className="ew-subnav">
          <button
            className={`ew-subnav-btn ${subTab === 'workflows' ? 'active' : ''}`}
            onClick={() => setSubTab('workflows')}
          >
            <Layers size={15} />
            Cross-Module Workflows ({filteredWorkflows.length})
          </button>
          <button
            className={`ew-subnav-btn ${subTab === 'events' ? 'active' : ''}`}
            onClick={() => setSubTab('events')}
          >
            <Activity size={15} />
            Event Store Ledger ({events.length})
          </button>
        </div>

        <div className="ew-actions-group">
          <input
            type="text"
            className="ew-search-input"
            placeholder="Search workflows or events..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />

          {subTab === 'workflows' && (
            <>
              <select
                className="ew-filter-select"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
              >
                <option value="">All Statuses</option>
                <option value="RUNNING">Running</option>
                <option value="AWAITING_APPROVAL">Awaiting Approval</option>
                <option value="COMPLETED">Completed</option>
                <option value="FAILED">Failed</option>
                <option value="CANCELLED">Cancelled</option>
              </select>

              <select
                className="ew-filter-select"
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
              >
                <option value="">All Workflow Types</option>
                <option value="SHIPMENT_EXCEPTION_RESPONSE">Shipment Exception</option>
                <option value="INVOICE_COLLECTION_ESCALATION">Invoice Collection</option>
                <option value="CONTRACT_COMPLIANCE_RENEWAL">Contract & Compliance</option>
                <option value="LEAD_FOLLOWUP">Lead Follow-up</option>
                <option value="RFQ_QUOTATION_DISPATCH">RFQ to Quotation</option>
              </select>
            </>
          )}

          <button
            className="btn-secondary-light"
            onClick={fetchData}
            title="Refresh event data"
          >
            <RefreshCw size={14} className={loading ? 'spin-icon' : ''} />
            Refresh
          </button>

          <button
            className="btn-primary-dark"
            onClick={() => setIsSimulateModalOpen(true)}
            id="btn-simulate-domain-event"
          >
            <Play size={14} />
            Simulate Event
          </button>
        </div>
      </div>

      {/* Main Content Area */}
      {subTab === 'workflows' ? (
        <div className="ew-table-wrapper">
          <table className="ew-table">
            <thead>
              <tr>
                <th>Workflow ID & Type</th>
                <th>Triggering Event</th>
                <th>Source Record</th>
                <th>Urgency</th>
                <th>AI Summary</th>
                <th>Status & Governance</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredWorkflows.length === 0 ? (
                <tr>
                  <td colSpan="7" className="ew-empty-state">
                    No matching cross-module event workflows found.
                  </td>
                </tr>
              ) : (
                filteredWorkflows.map((w) => (
                  <tr key={w.id}>
                    <td>
                      <strong>#{w.id}</strong>
                      <div style={{ fontSize: '0.75rem', color: '#64748b' }}>
                        {w.workflow_type}
                      </div>
                    </td>
                    <td>
                      <code>{w.trigger_event_type}</code>
                    </td>
                    <td>
                      <span className="ew-module-tag">{w.source_module}</span>
                      <div style={{ fontSize: '0.8125rem', marginTop: '2px' }}>
                        {w.source_record_type} #{w.source_record_id}
                      </div>
                    </td>
                    <td>{getUrgencyBadge(w.urgency)}</td>
                    <td style={{ maxWidth: '280px', lineHeight: '1.4' }}>
                      <span title={w.ai_summary}>
                        {w.ai_summary.length > 85 ? `${w.ai_summary.substring(0, 85)}...` : w.ai_summary}
                      </span>
                    </td>
                    <td>
                      <div>{getStatusBadge(w.status)}</div>
                      {w.approval_id && (
                        <div style={{ marginTop: '4px' }}>
                          <span
                            className="ew-approval-link"
                            onClick={() => onNavigateToApprovals && onNavigateToApprovals()}
                            title="View in Approvals Center"
                          >
                            Approval #{w.approval_id}
                          </span>
                        </div>
                      )}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '6px' }}>
                        <button
                          className="btn-secondary-light"
                          style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                          onClick={() => setSelectedWorkflow(w)}
                          title="View complete analysis & recommendations"
                        >
                          <Eye size={13} />
                          Details
                        </button>
                        {w.status === 'FAILED' && (
                          <button
                            className="btn-secondary-light"
                            style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                            onClick={() => handleRetry(w.id)}
                            title="Retry workflow"
                          >
                            <RotateCcw size={13} />
                          </button>
                        )}
                        {(w.status === 'RUNNING' || w.status === 'AWAITING_APPROVAL') && (
                          <button
                            className="btn-secondary-light"
                            style={{ padding: '4px 8px', fontSize: '0.75rem', color: '#b91c1c' }}
                            onClick={() => handleCancel(w.id)}
                            title="Cancel workflow"
                          >
                            <X size={13} />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      ) : (
        /* Event Store Ledger */
        <div className="ew-table-wrapper">
          <table className="ew-table">
            <thead>
              <tr>
                <th>Event ID</th>
                <th>Event Type</th>
                <th>Source Record</th>
                <th>Correlation ID</th>
                <th>Deduplication Key</th>
                <th>Status</th>
                <th>Ingested At</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {events.length === 0 ? (
                <tr>
                  <td colSpan="8" className="ew-empty-state">
                    No domain events ingested yet.
                  </td>
                </tr>
              ) : (
                events.map((e) => (
                  <tr key={e.id}>
                    <td><strong>#{e.id}</strong></td>
                    <td><code>{e.event_type}</code></td>
                    <td>
                      <span className="ew-module-tag">{e.source_module}</span> {e.source_record_type} #{e.source_record_id}
                    </td>
                    <td style={{ fontSize: '0.75rem', color: '#64748b' }}>
                      {e.correlation_id}
                    </td>
                    <td style={{ fontSize: '0.75rem', color: '#475569', maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={e.dedup_key}>
                      {e.dedup_key}
                    </td>
                    <td>
                      <span className={`ew-badge ${e.status === 'PROCESSED' ? 'ew-badge-status-completed' : e.status === 'IGNORED' ? 'ew-badge-status-cancelled' : 'ew-badge-status-running'}`}>
                        {e.status}
                      </span>
                    </td>
                    <td style={{ fontSize: '0.8125rem' }}>
                      {new Date(e.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                    </td>
                    <td>
                      <button
                        className="btn-secondary-light"
                        style={{ padding: '4px 8px', fontSize: '0.75rem' }}
                        onClick={() => setSelectedEvent(e)}
                      >
                        <Eye size={13} />
                        Payload
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Workflow Detail Drawer */}
      {selectedWorkflow && (
        <div className="ew-drawer-overlay" onClick={() => setSelectedWorkflow(null)}>
          <div className="ew-drawer-panel" onClick={(e) => e.stopPropagation()}>
            <div className="ew-drawer-header">
              <div>
                <div className="ew-drawer-title">
                  Cross-Module Workflow #{selectedWorkflow.id}
                </div>
                <div className="ew-drawer-sub">
                  Type: {selectedWorkflow.workflow_type} | Trigger: {selectedWorkflow.trigger_event_type}
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setSelectedWorkflow(null)}>
                <X size={18} />
              </button>
            </div>

            <div className="ew-drawer-content">
              {/* Approval Gating Banner */}
              {selectedWorkflow.status === 'AWAITING_APPROVAL' && (
                <div className="ew-approval-banner">
                  <ShieldAlert size={20} style={{ flexShrink: 0, marginTop: '2px' }} />
                  <div>
                    <strong>Human Approval Required (HITL Enforcement)</strong>
                    <p style={{ margin: '4px 0 6px 0', fontSize: '0.8125rem' }}>
                      Action '{selectedWorkflow.action_name}' involves significant operational or financial impact. Go has generated Approval Request #{selectedWorkflow.approval_id}.
                    </p>
                    {onNavigateToApprovals && (
                      <button
                        className="btn-primary-dark"
                        style={{ padding: '4px 10px', fontSize: '0.8125rem' }}
                        onClick={() => {
                          setSelectedWorkflow(null);
                          onNavigateToApprovals();
                        }}
                      >
                        Open in Approvals Center <ArrowRight size={13} />
                      </button>
                    )}
                  </div>
                </div>
              )}

              {/* Confirmed Event Data */}
              <div className="ew-section-card">
                <div className="ew-section-title">
                  <Activity size={16} /> Confirmed Event Data (Source of Truth)
                </div>
                <div className="ew-fact-grid">
                  <div className="ew-fact-item">
                    <span className="ew-fact-label">Event Type</span>
                    <span className="ew-fact-val">{selectedWorkflow.trigger_event_type}</span>
                  </div>
                  <div className="ew-fact-item">
                    <span className="ew-fact-label">Source Module</span>
                    <span className="ew-fact-val">{selectedWorkflow.source_module} ({selectedWorkflow.source_record_type} #{selectedWorkflow.source_record_id})</span>
                  </div>
                  <div className="ew-fact-item">
                    <span className="ew-fact-label">Status</span>
                    <span className="ew-fact-val">{selectedWorkflow.status}</span>
                  </div>
                  <div className="ew-fact-item">
                    <span className="ew-fact-label">Assessed Urgency</span>
                    <span className="ew-fact-val">{selectedWorkflow.urgency}</span>
                  </div>
                  <div className="ew-fact-item" style={{ gridColumn: 'span 2' }}>
                    <span className="ew-fact-label">Correlation ID</span>
                    <span className="ew-fact-val" style={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>{selectedWorkflow.correlation_id}</span>
                  </div>
                </div>
              </div>

              {/* AI Analysis & Synthesis */}
              <div className="ew-section-card">
                <div className="ew-section-title">
                  <Sparkles size={16} /> AI Cross-Module Synthesis (Python Sidecar)
                </div>
                <div className="ew-ai-summary-box">
                  {selectedWorkflow.ai_summary}
                </div>
              </div>

              {/* Recommended Next-Step Actions */}
              <div className="ew-section-card">
                <div className="ew-section-title">
                  <ShieldCheck size={16} /> Recommended Actions & Governance
                </div>
                {(() => {
                  let recs = [];
                  try {
                    recs = typeof selectedWorkflow.recommendations === 'string'
                      ? JSON.parse(selectedWorkflow.recommendations)
                      : selectedWorkflow.recommendations;
                  } catch (err) {
                    recs = [];
                  }
                  if (!Array.isArray(recs) || recs.length === 0) {
                    return <div style={{ fontSize: '0.8125rem', color: '#64748b' }}>No automated recommendations generated.</div>;
                  }
                  return recs.map((r, idx) => (
                    <div key={idx} className="ew-rec-card">
                      <div className="ew-rec-header">
                        <span className="ew-rec-title">{r.title || r.action_name}</span>
                        <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
                          <span className={`ew-badge ${r.risk_level === 'HIGH' || r.risk_level === 'CRITICAL' ? 'ew-badge-urgency-high' : 'ew-badge-urgency-low'}`}>
                            {r.risk_level} RISK
                          </span>
                          {r.requires_approval && (
                            <span className="ew-badge ew-badge-status-awaiting">
                              APPROVAL REQUIRED
                            </span>
                          )}
                        </div>
                      </div>
                      <div className="ew-rec-desc">{r.description}</div>
                    </div>
                  ));
                })()}
              </div>

              {/* Workflow Actions */}
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
                {selectedWorkflow.status === 'FAILED' && (
                  <button
                    className="btn-secondary-light"
                    onClick={() => handleRetry(selectedWorkflow.id)}
                  >
                    <RotateCcw size={14} /> Retry Workflow
                  </button>
                )}
                {(selectedWorkflow.status === 'RUNNING' || selectedWorkflow.status === 'AWAITING_APPROVAL') && (
                  <button
                    className="btn-secondary-light"
                    style={{ color: '#b91c1c' }}
                    onClick={() => handleCancel(selectedWorkflow.id)}
                  >
                    <X size={14} /> Cancel Workflow
                  </button>
                )}
                <button
                  className="btn-primary-dark"
                  onClick={() => setSelectedWorkflow(null)}
                >
                  Close Drawer
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Event Payload Modal */}
      {selectedEvent && (
        <div className="ew-drawer-overlay" onClick={() => setSelectedEvent(null)}>
          <div className="modal-content-card" style={{ maxWidth: '600px', margin: 'auto', background: '#ffffff', borderRadius: '8px', padding: '24px' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
              <h3 style={{ margin: 0, fontSize: '1.125rem' }}>Domain Event #{selectedEvent.id}</h3>
              <button className="btn-close-modal" onClick={() => setSelectedEvent(null)}>
                <X size={18} />
              </button>
            </div>
            <div style={{ fontSize: '0.8125rem', display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '16px' }}>
              <div><strong>Type:</strong> <code>{selectedEvent.event_type}</code></div>
              <div><strong>Module:</strong> {selectedEvent.source_module} ({selectedEvent.source_record_type} #{selectedEvent.source_record_id})</div>
              <div><strong>Deduplication Key:</strong> <code>{selectedEvent.dedup_key}</code></div>
              <div><strong>Correlation ID:</strong> <code>{selectedEvent.correlation_id}</code></div>
            </div>
            <div style={{ fontSize: '0.875rem', fontWeight: 600, marginBottom: '6px' }}>Event Payload:</div>
            <pre style={{ background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: '6px', padding: '12px', fontSize: '0.75rem', maxHeight: '250px', overflowY: 'auto' }}>
              {typeof selectedEvent.payload === 'string' ? selectedEvent.payload : JSON.stringify(selectedEvent.payload, null, 2)}
            </pre>
            <div style={{ textAlign: 'right', marginTop: '16px' }}>
              <button className="btn-primary-dark" onClick={() => setSelectedEvent(null)}>Close</button>
            </div>
          </div>
        </div>
      )}

      {/* Event Simulation Modal */}
      {isSimulateModalOpen && (
        <div className="ew-drawer-overlay" onClick={() => setIsSimulateModalOpen(false)}>
          <div className="modal-content-card" style={{ maxWidth: '520px', margin: 'auto', background: '#ffffff', borderRadius: '8px', padding: '24px' }} onClick={(e) => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
              <h3 style={{ margin: 0, fontSize: '1.125rem' }}>Simulate Domain Event Ingestion</h3>
              <button className="btn-close-modal" onClick={() => setIsSimulateModalOpen(false)}>
                <X size={18} />
              </button>
            </div>
            <p style={{ fontSize: '0.875rem', color: '#64748b', marginBottom: '16px' }}>
              Select a domain event scenario to test real Go event validation, deduplication, AI sidecar cross-module reasoning, and HITL governance creation:
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', marginBottom: '20px' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '0.875rem', cursor: 'pointer' }}>
                <input
                  type="radio"
                  name="sim_template"
                  value="shipment_delay"
                  checked={simulateTemplate === 'shipment_delay'}
                  onChange={(e) => setSimulateTemplate(e.target.value)}
                />
                <div>
                  <strong>Shipment Milestone Missed</strong>
                  <div style={{ fontSize: '0.75rem', color: '#64748b' }}>36h berth congestion delay requiring delay notices and customer update.</div>
                </div>
              </label>

              <label style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '0.875rem', cursor: 'pointer' }}>
                <input
                  type="radio"
                  name="sim_template"
                  value="invoice_overdue"
                  checked={simulateTemplate === 'invoice_overdue'}
                  onChange={(e) => setSimulateTemplate(e.target.value)}
                />
                <div>
                  <strong>Invoice Overdue Escalation</strong>
                  <div style={{ fontSize: '0.75rem', color: '#64748b' }}>$18,450 commercial invoice 21 days past due requiring dunning reminder.</div>
                </div>
              </label>

              <label style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '0.875rem', cursor: 'pointer' }}>
                <input
                  type="radio"
                  name="sim_template"
                  value="contract_expiry"
                  checked={simulateTemplate === 'contract_expiry'}
                  onChange={(e) => setSimulateTemplate(e.target.value)}
                />
                <div>
                  <strong>Contract Approaching Expiry</strong>
                  <div style={{ fontSize: '0.75rem', color: '#64748b' }}>14 days to MSA expiration; initiates SLA review & compliance audit.</div>
                </div>
              </label>

              <label style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '0.875rem', cursor: 'pointer' }}>
                <input
                  type="radio"
                  name="sim_template"
                  value="lead_qualified"
                  checked={simulateTemplate === 'lead_qualified'}
                  onChange={(e) => setSimulateTemplate(e.target.value)}
                />
                <div>
                  <strong>High-Value Lead Qualified</strong>
                  <div style={{ fontSize: '0.75rem', color: '#64748b' }}>Lead score 92, $85k budget; recommends executive sales engagement.</div>
                </div>
              </label>
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
              <button
                className="btn-secondary-light"
                onClick={() => setIsSimulateModalOpen(false)}
                disabled={simulating}
              >
                Cancel
              </button>
              <button
                className="btn-primary-dark"
                onClick={handleSimulate}
                disabled={simulating}
              >
                {simulating ? (
                  <>
                    <RefreshCw size={14} className="spin-icon" /> Ingesting & Analyzing...
                  </>
                ) : (
                  <>
                    <Play size={14} /> Ingest & Execute Workflow
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
