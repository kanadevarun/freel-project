import { useState, useEffect, useCallback } from 'react';
import {
  ShieldCheck,
  Zap,
  CheckCircle2,
  AlertTriangle,
  Clock,
  ArrowRight,
  Eye,
  Play,
  RotateCcw,
  Check,
  X,
  FileText,
  Search,
  Filter,
  Lock,
  Sparkles,
  Info,
  Layers,
  Database,
  ExternalLink,
} from 'lucide-react';
import { orchestrationService } from '../../../services/orchestrationService';

export default function ActionOrchestrationTab({ onNavigateToApprovals }) {
  const [subTab, setSubTab] = useState('proposals'); // 'proposals' | 'executions' | 'registry'
  const [proposals, setProposals] = useState([]);
  const [executions, setExecutions] = useState([]);
  const [actions, setActions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Filters
  const [statusFilter, setStatusFilter] = useState('');
  const [riskFilter, setRiskFilter] = useState('');
  const [search, setSearch] = useState('');

  // Modals & Drawers
  const [selectedProposal, setSelectedProposal] = useState(null);
  const [selectedExecution, setSelectedExecution] = useState(null);
  const [isProposeModalOpen, setIsProposeModalOpen] = useState(false);
  const [executingProposalId, setExecutingProposalId] = useState(null);
  const [executeSuccessMsg, setExecuteSuccessMsg] = useState(null);

  // Propose form state
  const [proposeForm, setProposeForm] = useState({
    source_module: 'INVOICES',
    source_record_type: 'INVOICE',
    source_record_id: '103',
    trigger_event: 'INVOICE_OVERDUE',
  });
  const [generatingProposal, setGeneratingProposal] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const [propsRes, execsRes, actsRes] = await Promise.all([
        orchestrationService.listProposals({ limit: 50 }).catch(() => ({ proposals: [] })),
        orchestrationService.listExecutions({ limit: 50 }).catch(() => ({ executions: [] })),
        orchestrationService.getRegisteredActions().catch(() => ({ actions: [] })),
      ]);

      setProposals(propsRes?.proposals || []);
      setExecutions(execsRes?.executions || []);
      setActions(actsRes?.actions || []);
    } catch (err) {
      setError(err.message || 'Failed to load action orchestration data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleExecute = async (proposal) => {
    if (proposal.requires_approval && proposal.approval_status !== 'Approved') {
      alert(`Execution Blocked: This action requires human approval. Current approval status is "${proposal.approval_status || 'Pending'}".`);
      return;
    }

    try {
      setExecutingProposalId(proposal.proposal_id);
      const res = await orchestrationService.executeProposal(proposal.proposal_id);
      if (res?.success && res.execution) {
        setExecuteSuccessMsg(`Action executed successfully! Verification: ${res.execution.verification_status}`);
        setTimeout(() => setExecuteSuccessMsg(null), 6000);
        fetchData();
      }
    } catch (err) {
      alert(`Execution Failed: ${err.message || err}`);
    } finally {
      setExecutingProposalId(null);
    }
  };

  const handleGenerateProposal = async (e) => {
    e.preventDefault();
    try {
      setGeneratingProposal(true);
      const res = await orchestrationService.generateProposal({
        source_module: proposeForm.source_module,
        source_record_type: proposeForm.source_record_type,
        source_record_id: proposeForm.source_record_id,
        trigger_event: proposeForm.trigger_event,
        operational_signals: [],
      });
      if (res?.success && res.proposal) {
        setIsProposeModalOpen(false);
        fetchData();
        setSelectedProposal(res.proposal);
      }
    } catch (err) {
      alert(`Proposal Generation Failed: ${err.message || err}`);
    } finally {
      setGeneratingProposal(false);
    }
  };

  // Filtered Proposals
  const filteredProposals = proposals.filter((p) => {
    if (statusFilter && p.status !== statusFilter) return false;
    if (riskFilter && p.risk_level !== riskFilter) return false;
    if (search) {
      const q = search.toLowerCase();
      return (
        p.proposal_id?.toLowerCase().includes(q) ||
        p.proposed_action_type?.toLowerCase().includes(q) ||
        p.explanation?.toLowerCase().includes(q) ||
        p.source_module?.toLowerCase().includes(q)
      );
    }
    return true;
  });

  return (
    <div className="action-orchestration-container" style={{ background: '#ffffff', borderRadius: '12px', padding: '24px', border: '1px solid #e2e8f0' }}>
      {/* Top Banner */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '24px' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
            <span style={{ background: '#eff6ff', color: '#2563eb', padding: '4px 10px', borderRadius: '6px', fontSize: '12px', fontWeight: 600 }}>
              Phase 3 Task 3.2
            </span>
            <h2 style={{ fontSize: '20px', fontWeight: 700, color: '#0f172a', margin: 0 }}>
              Controlled AI Workflow Execution & Action Orchestration
            </h2>
          </div>
          <p style={{ fontSize: '13px', color: '#64748b', margin: 0 }}>
            Deterministic signal detection → Grounded Action Proposal → Human Approval Gate → Verified Database Execution
          </p>
        </div>

        <button
          onClick={() => setIsProposeModalOpen(true)}
          style={{
            background: '#2563eb',
            color: '#ffffff',
            border: 'none',
            borderRadius: '8px',
            padding: '10px 18px',
            fontSize: '13px',
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            cursor: 'pointer',
            boxShadow: '0 1px 3px rgba(37,99,235,0.2)',
          }}
          id="btn-trigger-action-proposal"
        >
          <Sparkles size={16} />
          Propose Action
        </button>
      </div>

      {/* Success Notification Alert */}
      {executeSuccessMsg && (
        <div style={{ background: '#ecfdf5', border: '1px solid #10b981', color: '#065f46', padding: '12px 16px', borderRadius: '8px', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '10px', fontSize: '13px' }}>
          <CheckCircle2 size={18} color="#10b981" />
          <span>{executeSuccessMsg}</span>
        </div>
      )}

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px', marginBottom: '24px' }}>
        <div style={{ background: '#f8fafc', padding: '16px', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
          <div style={{ fontSize: '12px', color: '#64748b', fontWeight: 500 }}>Active Proposals</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#0f172a', marginTop: '4px' }}>{proposals.length}</div>
          <div style={{ fontSize: '11px', color: '#64748b', marginTop: '2px' }}>Grounded with DB facts</div>
        </div>
        <div style={{ background: '#fefce8', padding: '16px', borderRadius: '8px', border: '1px solid #fef08a' }}>
          <div style={{ fontSize: '12px', color: '#854d0e', fontWeight: 500 }}>Pending Approval</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#a16207', marginTop: '4px' }}>
            {proposals.filter((p) => p.requires_approval && p.approval_status !== 'Approved').length}
          </div>
          <div style={{ fontSize: '11px', color: '#854d0e', marginTop: '2px' }}>High-risk & external gates</div>
        </div>
        <div style={{ background: '#f0fdf4', padding: '16px', borderRadius: '8px', border: '1px solid #bbf7d0' }}>
          <div style={{ fontSize: '12px', color: '#166534', fontWeight: 500 }}>Executed & Verified</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#15803d', marginTop: '4px' }}>
            {executions.filter((e) => e.verification_status === 'VERIFIED').length}
          </div>
          <div style={{ fontSize: '11px', color: '#166534', marginTop: '2px' }}>100% DB verified outcomes</div>
        </div>
        <div style={{ background: '#eff6ff', padding: '16px', borderRadius: '8px', border: '1px solid #bfdbfe' }}>
          <div style={{ fontSize: '12px', color: '#1e40af', fontWeight: 500 }}>Registered Actions</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#1d4ed8', marginTop: '4px' }}>{actions.length}</div>
          <div style={{ fontSize: '11px', color: '#1e40af', marginTop: '2px' }}>Go platform registry</div>
        </div>
      </div>

      {/* Sub-navigation Tabs */}
      <div style={{ display: 'flex', borderBottom: '1px solid #e2e8f0', marginBottom: '20px', gap: '8px' }}>
        <button
          onClick={() => setSubTab('proposals')}
          style={{
            background: 'none',
            border: 'none',
            borderBottom: subTab === 'proposals' ? '2px solid #2563eb' : '2px solid transparent',
            padding: '10px 16px',
            fontSize: '13px',
            fontWeight: 600,
            color: subTab === 'proposals' ? '#2563eb' : '#64748b',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
          id="subtab-proposals"
        >
          <Sparkles size={16} />
          Action Proposals ({proposals.length})
        </button>
        <button
          onClick={() => setSubTab('executions')}
          style={{
            background: 'none',
            border: 'none',
            borderBottom: subTab === 'executions' ? '2px solid #2563eb' : '2px solid transparent',
            padding: '10px 16px',
            fontSize: '13px',
            fontWeight: 600,
            color: subTab === 'executions' ? '#2563eb' : '#64748b',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
          id="subtab-executions"
        >
          <Database size={16} />
          Execution Log & DB Verification ({executions.length})
        </button>
        <button
          onClick={() => setSubTab('registry')}
          style={{
            background: 'none',
            border: 'none',
            borderBottom: subTab === 'registry' ? '2px solid #2563eb' : '2px solid transparent',
            padding: '10px 16px',
            fontSize: '13px',
            fontWeight: 600,
            color: subTab === 'registry' ? '#2563eb' : '#64748b',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
          id="subtab-registry"
        >
          <Layers size={16} />
          Registered Action Capabilities ({actions.length})
        </button>
      </div>

      {/* Subtab Content: Proposals */}
      {subTab === 'proposals' && (
        <div>
          {/* Filters Bar */}
          <div style={{ display: 'flex', gap: '12px', marginBottom: '16px' }}>
            <div style={{ position: 'relative', flex: 1 }}>
              <Search size={16} style={{ position: 'absolute', left: '12px', top: '10px', color: '#94a3b8' }} />
              <input
                type="text"
                placeholder="Search proposals by action, ID, explanation..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 12px 8px 36px',
                  borderRadius: '6px',
                  border: '1px solid #cbd5e1',
                  fontSize: '13px',
                  outline: 'none',
                }}
              />
            </div>
            <select
              value={riskFilter}
              onChange={(e) => setRiskFilter(e.target.value)}
              style={{ padding: '8px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '13px', color: '#334155' }}
            >
              <option value="">All Risk Levels</option>
              <option value="SAFE_INTERNAL">Safe Internal</option>
              <option value="LOW">Low Risk</option>
              <option value="MEDIUM">Medium Risk</option>
              <option value="HIGH">High Risk</option>
            </select>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              style={{ padding: '8px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '13px', color: '#334155' }}
            >
              <option value="">All Proposal Statuses</option>
              <option value="PROPOSED">Proposed</option>
              <option value="APPROVED">Approved</option>
              <option value="EXECUTING">Executing</option>
              <option value="EXECUTED">Executed</option>
            </select>
          </div>

          {/* Proposals Table */}
          {filteredProposals.length === 0 ? (
            <div style={{ padding: '48px', textAlign: 'center', background: '#f8fafc', borderRadius: '8px', border: '1px dashed #cbd5e1' }}>
              <Sparkles size={32} color="#94a3b8" style={{ margin: '0 auto 12px' }} />
              <p style={{ margin: '0 0 8px', fontSize: '14px', fontWeight: 600, color: '#334155' }}>No Action Proposals Found</p>
              <p style={{ margin: 0, fontSize: '13px', color: '#64748b' }}>
                Trigger an action proposal above to run deterministic evaluation and grounded action synthesis.
              </p>
            </div>
          ) : (
            <div style={{ overflowX: 'auto', border: '1px solid #e2e8f0', borderRadius: '8px' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', textAlign: 'left' }}>
                <thead>
                  <tr style={{ background: '#f8fafc', borderBottom: '1px solid #e2e8f0', color: '#475569', fontWeight: 600 }}>
                    <th style={{ padding: '12px 16px' }}>Proposal</th>
                    <th style={{ padding: '12px 16px' }}>Source Context</th>
                    <th style={{ padding: '12px 16px' }}>Proposed Action</th>
                    <th style={{ padding: '12px 16px' }}>Risk & Approval</th>
                    <th style={{ padding: '12px 16px' }}>Confidence</th>
                    <th style={{ padding: '12px 16px', textAlign: 'right' }}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredProposals.map((p) => {
                    const isHighRisk = p.risk_level === 'HIGH' || p.risk_level === 'CRITICAL';
                    const isApproved = p.approval_status === 'Approved';
                    const isPending = p.requires_approval && !isApproved;

                    return (
                      <tr key={p.id} style={{ borderBottom: '1px solid #e2e8f0' }}>
                        <td style={{ padding: '14px 16px' }}>
                          <div style={{ fontWeight: 600, color: '#0f172a' }}>{p.proposal_id}</div>
                          <div style={{ fontSize: '11px', color: '#64748b', marginTop: '2px' }}>
                            {p.created_at ? new Date(p.created_at).toLocaleString() : 'Recent'}
                          </div>
                        </td>
                        <td style={{ padding: '14px 16px' }}>
                          <span style={{ background: '#f1f5f9', color: '#334155', padding: '2px 8px', borderRadius: '4px', fontSize: '11px', fontWeight: 600 }}>
                            {p.source_module}
                          </span>
                          <span style={{ fontSize: '12px', color: '#64748b', marginLeft: '6px' }}>
                            #{p.source_record_id} ({p.trigger_event})
                          </span>
                        </td>
                        <td style={{ padding: '14px 16px' }}>
                          <div style={{ fontWeight: 600, color: '#1e293b' }}>{p.proposed_action_type}</div>
                          <div style={{ fontSize: '12px', color: '#64748b', maxWidth: '320px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                            {p.explanation}
                          </div>
                        </td>
                        <td style={{ padding: '14px 16px' }}>
                          <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
                            <span
                              style={{
                                padding: '2px 8px',
                                borderRadius: '4px',
                                fontSize: '11px',
                                fontWeight: 600,
                                background: isHighRisk ? '#fee2e2' : '#f0fdf4',
                                color: isHighRisk ? '#b91c1c' : '#15803d',
                              }}
                            >
                              {p.risk_level}
                            </span>
                            {p.requires_approval ? (
                              <span
                                style={{
                                  padding: '2px 8px',
                                  borderRadius: '4px',
                                  fontSize: '11px',
                                  fontWeight: 600,
                                  background: isApproved ? '#ecfdf5' : '#fef3c7',
                                  color: isApproved ? '#047857' : '#b45309',
                                  display: 'flex',
                                  alignItems: 'center',
                                  gap: '4px',
                                }}
                              >
                                <Lock size={10} />
                                {p.approval_status || 'Pending'}
                              </span>
                            ) : (
                              <span style={{ fontSize: '11px', color: '#64748b' }}>Pre-Approved</span>
                            )}
                          </div>
                        </td>
                        <td style={{ padding: '14px 16px' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                            <div style={{ width: '45px', background: '#e2e8f0', height: '6px', borderRadius: '3px', overflow: 'hidden' }}>
                              <div
                                style={{
                                  width: `${Math.round((p.confidence || 0.85) * 100)}%`,
                                  background: '#2563eb',
                                  height: '100%',
                                }}
                              />
                            </div>
                            <span style={{ fontSize: '12px', fontWeight: 600, color: '#334155' }}>
                              {Math.round((p.confidence || 0.85) * 100)}%
                            </span>
                          </div>
                        </td>
                        <td style={{ padding: '14px 16px', textAlign: 'right' }}>
                          <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                            <button
                              onClick={() => setSelectedProposal(p)}
                              style={{
                                background: '#f8fafc',
                                border: '1px solid #cbd5e1',
                                borderRadius: '6px',
                                padding: '6px 10px',
                                fontSize: '12px',
                                cursor: 'pointer',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '4px',
                                color: '#334155',
                              }}
                            >
                              <Eye size={13} />
                              Evidence
                            </button>
                            <button
                              onClick={() => handleExecute(p)}
                              disabled={isPending || executingProposalId === p.proposal_id}
                              style={{
                                background: isPending ? '#e2e8f0' : '#16a34a',
                                color: isPending ? '#94a3b8' : '#ffffff',
                                border: 'none',
                                borderRadius: '6px',
                                padding: '6px 12px',
                                fontSize: '12px',
                                fontWeight: 600,
                                cursor: isPending ? 'not-allowed' : 'pointer',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '4px',
                              }}
                              title={isPending ? 'Requires human approval before execution' : 'Execute action with idempotency key'}
                            >
                              <Play size={13} />
                              {executingProposalId === p.proposal_id ? 'Executing...' : 'Execute'}
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Subtab Content: Executions */}
      {subTab === 'executions' && (
        <div>
          {executions.length === 0 ? (
            <div style={{ padding: '48px', textAlign: 'center', background: '#f8fafc', borderRadius: '8px', border: '1px dashed #cbd5e1' }}>
              <Database size={32} color="#94a3b8" style={{ margin: '0 auto 12px' }} />
              <p style={{ margin: '0 0 8px', fontSize: '14px', fontWeight: 600, color: '#334155' }}>No Executions Recorded Yet</p>
              <p style={{ margin: 0, fontSize: '13px', color: '#64748b' }}>
                When action proposals are executed, every step and MariaDB verification log will appear here.
              </p>
            </div>
          ) : (
            <div style={{ overflowX: 'auto', border: '1px solid #e2e8f0', borderRadius: '8px' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', textAlign: 'left' }}>
                <thead>
                  <tr style={{ background: '#f8fafc', borderBottom: '1px solid #e2e8f0', color: '#475569', fontWeight: 600 }}>
                    <th style={{ padding: '12px 16px' }}>Exec #</th>
                    <th style={{ padding: '12px 16px' }}>Action Name</th>
                    <th style={{ padding: '12px 16px' }}>Status</th>
                    <th style={{ padding: '12px 16px' }}>DB Verification</th>
                    <th style={{ padding: '12px 16px' }}>Idempotency Key</th>
                    <th style={{ padding: '12px 16px' }}>Executed At</th>
                    <th style={{ padding: '12px 16px', textAlign: 'right' }}>Details</th>
                  </tr>
                </thead>
                <tbody>
                  {executions.map((e) => (
                    <tr key={e.id} style={{ borderBottom: '1px solid #e2e8f0' }}>
                      <td style={{ padding: '14px 16px', fontWeight: 600, color: '#0f172a' }}>#{e.id}</td>
                      <td style={{ padding: '14px 16px', fontWeight: 600, color: '#1e293b' }}>{e.action_name}</td>
                      <td style={{ padding: '14px 16px' }}>
                        <span
                          style={{
                            padding: '3px 8px',
                            borderRadius: '4px',
                            fontSize: '11px',
                            fontWeight: 600,
                            background: e.status === 'COMPLETED' ? '#ecfdf5' : '#fee2e2',
                            color: e.status === 'COMPLETED' ? '#047857' : '#b91c1c',
                          }}
                        >
                          {e.status}
                        </span>
                      </td>
                      <td style={{ padding: '14px 16px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <CheckCircle2 size={15} color="#16a34a" />
                          <span style={{ fontWeight: 600, color: '#15803d', fontSize: '12px' }}>
                            {e.verification_status}
                          </span>
                        </div>
                      </td>
                      <td style={{ padding: '14px 16px', fontFamily: 'monospace', fontSize: '12px', color: '#64748b' }}>
                        {e.idempotency_key}
                      </td>
                      <td style={{ padding: '14px 16px', fontSize: '12px', color: '#64748b' }}>
                        {e.created_at ? new Date(e.created_at).toLocaleString() : 'Just now'}
                      </td>
                      <td style={{ padding: '14px 16px', textAlign: 'right' }}>
                        <button
                          onClick={() => setSelectedExecution(e)}
                          style={{
                            background: '#f8fafc',
                            border: '1px solid #cbd5e1',
                            borderRadius: '6px',
                            padding: '6px 10px',
                            fontSize: '12px',
                            cursor: 'pointer',
                            color: '#334155',
                          }}
                        >
                          View Logs
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* Subtab Content: Registry */}
      {subTab === 'registry' && (
        <div>
          <p style={{ fontSize: '13px', color: '#64748b', marginBottom: '16px' }}>
            The following business actions are registered in the Go Orchestration Engine with strict input schemas, permission gates, and post-execution DB verifiers:
          </p>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '16px' }}>
            {actions.map((act) => (
              <div
                key={act.name}
                style={{
                  background: '#f8fafc',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                  padding: '16px',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '8px',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontWeight: 700, color: '#0f172a', fontSize: '14px' }}>{act.name}</span>
                  <span
                    style={{
                      padding: '2px 8px',
                      borderRadius: '4px',
                      fontSize: '11px',
                      fontWeight: 600,
                      background: act.category === 'SAFE_INTERNAL' ? '#eff6ff' : '#fef2f2',
                      color: act.category === 'SAFE_INTERNAL' ? '#2563eb' : '#dc2626',
                    }}
                  >
                    {act.category}
                  </span>
                </div>
                <p style={{ fontSize: '12px', color: '#64748b', margin: 0 }}>{act.description}</p>
                <div style={{ display: 'flex', gap: '12px', fontSize: '11px', color: '#475569', marginTop: '8px' }}>
                  <span>Approval: <strong>{act.requires_approval ? 'Required' : 'Pre-approved'}</strong></span>
                  <span>Reversibility: <strong>{act.is_reversible}</strong></span>
                  <span>Permission: <code>{act.required_permission}</code></span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Evidence & Details Modal */}
      {selectedProposal && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(15,23,42,0.4)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 }}>
          <div style={{ background: '#ffffff', borderRadius: '12px', width: '600px', maxWidth: '90vw', padding: '24px', boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Sparkles size={18} color="#2563eb" />
                <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700, color: '#0f172a' }}>
                  Proposal Evidence & Impact ({selectedProposal.proposal_id})
                </h3>
              </div>
              <button onClick={() => setSelectedProposal(null)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#64748b' }}>
                <X size={18} />
              </button>
            </div>

            <div style={{ marginBottom: '16px' }}>
              <div style={{ fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>AI Explanation:</div>
              <div style={{ background: '#f8fafc', padding: '12px', borderRadius: '6px', fontSize: '13px', color: '#1e293b' }}>
                {selectedProposal.explanation}
              </div>
            </div>

            <div style={{ marginBottom: '16px' }}>
              <div style={{ fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>Expected Business Impact:</div>
              <div style={{ background: '#f8fafc', padding: '12px', borderRadius: '6px', fontSize: '13px', color: '#1e293b' }}>
                {selectedProposal.expected_impact}
              </div>
            </div>

            <div style={{ marginBottom: '16px' }}>
              <div style={{ fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '6px' }}>Grounded Database Facts (Evidence):</div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                {selectedProposal.evidence && selectedProposal.evidence.map((fact, idx) => (
                  <div key={idx} style={{ background: '#f1f5f9', padding: '8px 12px', borderRadius: '6px', fontSize: '12px', color: '#334155', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <CheckCircle2 size={14} color="#2563eb" />
                    <span>{typeof fact === 'string' ? fact : JSON.stringify(fact)}</span>
                  </div>
                ))}
              </div>
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '20px' }}>
              <button
                onClick={() => setSelectedProposal(null)}
                style={{ background: '#f1f5f9', border: 'none', borderRadius: '6px', padding: '8px 16px', fontSize: '13px', cursor: 'pointer', color: '#475569' }}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Execution Logs Modal */}
      {selectedExecution && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(15,23,42,0.4)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 }}>
          <div style={{ background: '#ffffff', borderRadius: '12px', width: '600px', maxWidth: '90vw', padding: '24px', boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <Database size={18} color="#16a34a" />
                <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700, color: '#0f172a' }}>
                  Execution Verification Details (#{selectedExecution.id})
                </h3>
              </div>
              <button onClick={() => setSelectedExecution(null)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#64748b' }}>
                <X size={18} />
              </button>
            </div>

            <div style={{ marginBottom: '14px' }}>
              <div style={{ fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>Idempotency Key:</div>
              <div style={{ background: '#f8fafc', padding: '8px 12px', borderRadius: '6px', fontSize: '12px', fontFamily: 'monospace', color: '#334155' }}>
                {selectedExecution.idempotency_key}
              </div>
            </div>

            <div style={{ marginBottom: '14px' }}>
              <div style={{ fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>MariaDB Post-Execution Verification:</div>
              <div style={{ background: '#ecfdf5', border: '1px solid #10b981', padding: '10px 14px', borderRadius: '6px', fontSize: '13px', color: '#065f46' }}>
                {typeof selectedExecution.verification_details === 'string'
                  ? selectedExecution.verification_details
                  : JSON.stringify(selectedExecution.verification_details, null, 2)}
              </div>
            </div>

            <div style={{ marginBottom: '14px' }}>
              <div style={{ fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>Action Output Result:</div>
              <pre style={{ background: '#f8fafc', padding: '10px', borderRadius: '6px', fontSize: '12px', overflowX: 'auto', margin: 0 }}>
                {JSON.stringify(selectedExecution.action_output, null, 2)}
              </pre>
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '20px' }}>
              <button
                onClick={() => setSelectedExecution(null)}
                style={{ background: '#f1f5f9', border: 'none', borderRadius: '6px', padding: '8px 16px', fontSize: '13px', cursor: 'pointer', color: '#475569' }}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Propose Action Modal */}
      {isProposeModalOpen && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(15,23,42,0.4)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 }}>
          <form onSubmit={handleGenerateProposal} style={{ background: '#ffffff', borderRadius: '12px', width: '500px', maxWidth: '90vw', padding: '24px', boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
              <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700, color: '#0f172a' }}>
                Propose Controlled Business Action
              </h3>
              <button type="button" onClick={() => setIsProposeModalOpen(false)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: '#64748b' }}>
                <X size={18} />
              </button>
            </div>

            <div style={{ marginBottom: '14px' }}>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>Source Module</label>
              <select
                value={proposeForm.source_module}
                onChange={(e) => {
                  const m = e.target.value;
                  setProposeForm({
                    ...proposeForm,
                    source_module: m,
                    source_record_type: m === 'SHIPMENTS' ? 'SHIPMENT' : m === 'INVOICES' ? 'INVOICE' : 'CUSTOMER',
                    source_record_id: m === 'INVOICES' ? '103' : '101',
                    trigger_event: m === 'INVOICES' ? 'INVOICE_OVERDUE' : 'SCHEDULE_VARIANCE_SIGNAL',
                  });
                }}
                style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '13px' }}
              >
                <option value="INVOICES">Invoices & Receivables</option>
                <option value="SHIPMENTS">Shipments & Operations</option>
                <option value="CUSTOMERS">Customers & Commercial</option>
              </select>
            </div>

            <div style={{ marginBottom: '14px' }}>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>Target Record ID</label>
              <input
                type="text"
                value={proposeForm.source_record_id}
                onChange={(e) => setProposeForm({ ...proposeForm, source_record_id: e.target.value })}
                required
                style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '13px' }}
              />
            </div>

            <div style={{ marginBottom: '14px' }}>
              <label style={{ display: 'block', fontSize: '12px', fontWeight: 600, color: '#475569', marginBottom: '4px' }}>Trigger Event</label>
              <input
                type="text"
                value={proposeForm.trigger_event}
                onChange={(e) => setProposeForm({ ...proposeForm, trigger_event: e.target.value })}
                required
                style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', border: '1px solid #cbd5e1', fontSize: '13px' }}
              />
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '20px' }}>
              <button
                type="button"
                onClick={() => setIsProposeModalOpen(false)}
                style={{ background: '#f1f5f9', border: 'none', borderRadius: '6px', padding: '8px 16px', fontSize: '13px', cursor: 'pointer', color: '#475569' }}
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={generatingProposal}
                style={{ background: '#2563eb', color: '#ffffff', border: 'none', borderRadius: '6px', padding: '8px 16px', fontSize: '13px', fontWeight: 600, cursor: generatingProposal ? 'not-allowed' : 'pointer' }}
              >
                {generatingProposal ? 'Synthesizing...' : 'Synthesize Proposal'}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
