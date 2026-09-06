import React, { useState, useEffect, useCallback } from 'react';
import { 
  X, Calendar, User, FileText, CheckCircle, ShieldAlert, 
  Archive, Edit2, AlertCircle, Clock, ShieldCheck, 
  AlertTriangle, DollarSign, Truck, Plane, Ship, 
  Layers, Link2, Unlink, RefreshCw, FilePlus, Sparkles,
  ArrowRight, Info, Building2, ExternalLink, Hash, Check,
  Activity, CheckCircle2, XCircle, ArrowUpRight, TrendingUp,
  ChevronLeft, ChevronRight
} from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import ConfirmModal from '../Settings/ConfirmModal';
import toast from 'react-hot-toast';
import ContractLinkedRecordsPanel from './ContractLinkedRecordsPanel';
import ContractVersionsPanel from './ContractVersionsPanel';
import ContractAmendmentsPanel from './ContractAmendmentsPanel';
import ContractApprovalPanel from './ContractApprovalPanel';
import ContractDocumentsPanel from './ContractDocumentsPanel';
import ContractTermsPanel from './ContractTermsPanel';
import ContractObligationsPanel from './ContractObligationsPanel';
import ContractCompliancePanel from './ContractCompliancePanel';
import ContractPerformancePanel from './ContractPerformancePanel';
import './ContractDrawer.css';

export default function ContractDrawer({ contract, onClose, onEdit, onUpdate }) {
  const [activeTab, setActiveTab] = useState('OVERVIEW');
  const [events, setEvents] = useState([]);
  const [isLoadingEvents, setIsLoadingEvents] = useState(false);
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);
  const [statusConfirmModal, setStatusConfirmModal] = useState({
    isOpen: false,
    newStatus: '',
    actionName: '',
    title: '',
    message: '',
    confirmText: '',
    confirmStyle: 'primary'
  });

  // Task 20.3 States: Renewal, Commercial Impact, Risks
  const [renewalTracking, setRenewalTracking] = useState(null);
  const [commercialImpact, setCommercialImpact] = useState(null);
  const [riskEvents, setRiskEvents] = useState([]);
  const [isLoadingIntelligence, setIsLoadingIntelligence] = useState(false);

  // Renewal Form Modal State
  const [showRenewalModal, setShowRenewalModal] = useState(false);
  const [renewalForm, setRenewalForm] = useState({
    target_completion_date: '',
    owner: contract?.owner || '',
    notes: ''
  });

  // Successor Link Modal State
  const [showSuccessorModal, setShowSuccessorModal] = useState(false);
  const [successorIDInput, setSuccessorIDInput] = useState('');
  const [availableContracts, setAvailableContracts] = useState([]);

  // Resolve Risk Modal State
  const [resolvingRisk, setResolvingRisk] = useState(null);
  const [resolutionNotes, setResolutionNotes] = useState('');

  const fetchIntelligenceData = useCallback(async () => {
    if (!contract) return;
    setIsLoadingIntelligence(true);
    try {
      const [renewalRes, impactRes, risksRes] = await Promise.all([
        contractsService.getRenewalTracking(contract.id).catch(() => null),
        contractsService.getCommercialImpact(contract.id).catch(() => null),
        contractsService.getContractRisks(contract.id).catch(() => ({ data: [] }))
      ]);

      setRenewalTracking(renewalRes?.data || renewalRes || null);
      setCommercialImpact(impactRes?.data || impactRes || null);
      setRiskEvents(risksRes?.data || []);
    } catch (err) {
      console.error('Failed to load contract intelligence', err);
    } finally {
      setIsLoadingIntelligence(false);
    }
  }, [contract]);

  useEffect(() => {
    if (contract) {
      fetchIntelligenceData();
    }
  }, [contract, fetchIntelligenceData]);

  useEffect(() => {
    if (activeTab === 'TIMELINE' && contract) {
      fetchEvents();
    }
  }, [activeTab, contract]);

  const fetchEvents = async () => {
    setIsLoadingEvents(true);
    try {
      const [lifecycleEvents, intelEvents] = await Promise.all([
        contractsService.getContractLifecycleEvents(contract.id).catch(() => ({ data: [] })),
        contractsService.getLifecycleEvents({ contract_id: contract.id }).catch(() => ({ data: [] }))
      ]);

      const lItems = Array.isArray(lifecycleEvents) ? lifecycleEvents : (lifecycleEvents?.data?.data || lifecycleEvents?.data || []);
      const iItems = Array.isArray(intelEvents) ? intelEvents : (intelEvents?.data || []);

      // Merge and sort
      const combined = [
        ...lItems.map(item => ({ ...item, isIntel: false })),
        ...iItems.map(item => ({ ...item, isIntel: true }))
      ].sort((a, b) => new Date(b.created_at || b.event_date) - new Date(a.created_at || a.event_date));

      setEvents(combined);
    } catch (err) {
      console.error(err);
      toast.error('Failed to load contract timeline');
    } finally {
      setIsLoadingEvents(false);
    }
  };

  const handleStartRenewal = async () => {
    try {
      const res = await contractsService.startRenewal(contract.id, {
        target_completion_date: renewalForm.target_completion_date || undefined,
        owner: renewalForm.owner || undefined,
        notes: renewalForm.notes || undefined,
      });
      toast.success('Commercial renewal initiated successfully');
      setShowRenewalModal(false);
      fetchIntelligenceData();
      if (onUpdate) onUpdate();
    } catch (err) {
      toast.error(err.message || 'Failed to initiate renewal');
    }
  };

  const handleUpdateRenewalStatus = async (status) => {
    try {
      await contractsService.updateRenewalTracking(contract.id, {
        renewal_status: status
      });
      toast.success(`Renewal status updated to ${status}`);
      fetchIntelligenceData();
      if (onUpdate) onUpdate();
    } catch (err) {
      toast.error(err.message || 'Failed to update renewal status');
    }
  };

  const handleLinkSuccessor = async () => {
    const successorID = parseInt(successorIDInput, 10);
    if (isNaN(successorID) || successorID <= 0) {
      toast.error('Please enter a valid successor contract ID');
      return;
    }
    try {
      await contractsService.updateRenewalTracking(contract.id, {
        successor_contract_id: successorID,
        renewal_status: 'RENEWED'
      });
      toast.success('Successor contract linked and agreement marked as renewed');
      setShowSuccessorModal(false);
      fetchIntelligenceData();
      if (onUpdate) onUpdate();
    } catch (err) {
      toast.error(err.message || 'Failed to link successor contract');
    }
  };

  const handleResolveRisk = async () => {
    if (!resolvingRisk) return;
    try {
      await contractsService.resolveRisk(contract.id, resolvingRisk.id, {
        resolution_notes: resolutionNotes || undefined
      });
      toast.success('Risk event marked as resolved');
      setResolvingRisk(null);
      setResolutionNotes('');
      fetchIntelligenceData();
    } catch (err) {
      toast.error(err.message || 'Failed to resolve risk');
    }
  };

  const promptStatusUpdate = (newStatus, actionName) => {
    setStatusConfirmModal({
      isOpen: true,
      newStatus,
      actionName,
      title: `${actionName} Contract Agreement`,
      message: `Are you sure you want to ${actionName.toLowerCase()} "${contract.contract_name || contract.contract_reference}"? This will update the contract lifecycle and governance state.`,
      confirmText: `${actionName} Contract`,
      confirmStyle: newStatus === 'ARCHIVED' ? 'danger' : 'primary'
    });
  };

  const handleConfirmedStatusUpdate = async () => {
    const { newStatus, actionName } = statusConfirmModal;
    setIsUpdatingStatus(true);
    try {
      if (newStatus === 'ARCHIVED') {
        await contractsService.archiveContract(contract.id);
      } else {
        await contractsService.updateContractLifecycle(contract.id, {
          new_status: newStatus,
          description: `User manually ${actionName.toLowerCase()} contract`
        });
      }
      toast.success(`Contract ${actionName.toLowerCase()} successfully`);
      setStatusConfirmModal(prev => ({ ...prev, isOpen: false }));
      if (onUpdate) onUpdate();
      onClose();
    } catch (err) {
      console.error(err);
      toast.error(err.message || `Failed to ${actionName} contract`);
    } finally {
      setIsUpdatingStatus(false);
    }
  };

  if (!contract) return null;

  // Compute Contract Health Intelligence
  const computeHealth = () => {
    if (renewalTracking?.renewal_status === 'IN_PROGRESS') {
      return {
        label: 'Renewal In Progress',
        badgeClass: 'health-renewal',
        icon: <Sparkles size={15} />,
        daysText: 'Under Active Renewal',
        progress: 80,
        subtext: 'Commercial renegotiation or replacement contract in progress.'
      };
    }
    if (renewalTracking?.renewal_status === 'RENEWED') {
      return {
        label: 'Renewed Agreement',
        badgeClass: 'health-active',
        icon: <ShieldCheck size={15} />,
        daysText: 'Successor Linked',
        progress: 100,
        subtext: 'Successfully replaced by successor commercial agreement.'
      };
    }
    if (!contract.expiry_date || contract.status === 'DRAFT') {
      return {
        label: 'Draft Pipeline',
        badgeClass: 'health-draft',
        icon: <Clock size={15} />,
        daysText: 'Pending Activation',
        progress: 25,
        subtext: 'Agreement awaiting formal execution and commercial effect.'
      };
    }

    const today = new Date();
    const expiry = new Date(contract.expiry_date);
    const effective = contract.effective_date ? new Date(contract.effective_date) : new Date(today.getTime() - 180 * 86400000);
    const diffDays = Math.ceil((expiry.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

    const totalDuration = Math.max(1, expiry.getTime() - effective.getTime());
    const elapsed = Math.max(0, today.getTime() - effective.getTime());
    const progress = Math.min(100, Math.max(0, Math.round((elapsed / totalDuration) * 100)));

    if (contract.status === 'EXPIRED' || diffDays <= 0) {
      return {
        label: 'Expired Agreement',
        badgeClass: 'health-expired',
        icon: <AlertCircle size={15} />,
        daysText: `Expired (${Math.abs(diffDays)} Days Ago)`,
        progress: 100,
        subtext: 'Commercial validity window elapsed. Rates and SLAs require explicit renewal.'
      };
    }

    if (diffDays <= 7) {
      return {
        label: 'Critical Expiry Window',
        badgeClass: 'health-critical',
        icon: <AlertTriangle size={15} />,
        daysText: `${diffDays} Days Remaining`,
        progress,
        subtext: 'Impending expiry requires immediate commercial extension or successor execution.'
      };
    }

    if (diffDays <= 30) {
      return {
        label: 'Expiring Soon',
        badgeClass: 'health-warning',
        icon: <Clock size={15} />,
        daysText: `${diffDays} Days Remaining`,
        progress,
        subtext: 'Contract term concludes within 30 days. Priority renewal review recommended.'
      };
    }

    return {
      label: 'Active & In Effect',
      badgeClass: 'health-active',
      icon: <ShieldCheck size={15} />,
      daysText: `${diffDays} Days Remaining`,
      progress,
      subtext: 'Active commercial agreement operating within scheduled term.'
    };
  };

  const health = computeHealth();

  const getModeIcon = (mode) => {
    switch ((mode || '').toUpperCase()) {
      case 'AIR': return <Plane size={14} className="text-blue" />;
      case 'OCEAN': case 'SEA': return <Ship size={14} className="text-indigo" />;
      case 'ROAD': return <Truck size={14} className="text-emerald" />;
      default: return <Layers size={14} className="text-slate" />;
    }
  };

  const getEventIcon = (eventType) => {
    switch (eventType) {
      case 'CREATED': case 'CONTRACT_CREATED': return <FilePlus size={15} className="text-blue" />;
      case 'STATUS_CHANGED': case 'STATUS_CHANGE': return <RefreshCw size={15} className="text-indigo" />;
      case 'LINK_ADDED': return <Link2 size={15} className="text-emerald" />;
      case 'LINK_REMOVED': return <Unlink size={15} className="text-rose" />;
      case 'EXPIRY_APPROACHING': case 'CRITICAL_EXPIRY_RISK': return <AlertTriangle size={15} className="text-amber" />;
      case 'RENEWAL_STARTED': case 'RENEWAL_COMPLETED': return <Sparkles size={15} className="text-indigo" />;
      case 'RISK_RESOLVED': return <CheckCircle2 size={15} className="text-emerald" />;
      default: return <Clock size={15} className="text-slate" />;
    }
  };

  const activeRisksCount = riskEvents.filter(r => !r.is_resolved).length;

  return (
    <div className="contract-drawer-overlay" onClick={onClose}>
      <div className="contract-drawer" onClick={e => e.stopPropagation()}>
        
        {/* ── 1. Clean Top Header ── */}
        <div className="cd-header">
          <div className="cd-header-main">
            <div className="cd-header-left">
              <h2 className="cd-title">{contract.contract_name}</h2>
              <div className="cd-subtitle-row">
                <span className="cd-ref-tag">{contract.contract_reference || 'REF-TBD'}</span>
                <span className="cd-type-tag">
                  {contract.contract_type ? contract.contract_type.replace(/_/g, ' ') : 'AGREEMENT'}
                </span>
                {contract.transport_mode && (
                  <span className="cd-mode-tag">
                    {getModeIcon(contract.transport_mode)}
                    {contract.transport_mode}
                  </span>
                )}
                {contract.source_document_id && (
                  <span className="cd-source-pill" title="Created via AI Agreement Document Import">
                    <Sparkles size={11} /> AI Imported
                  </span>
                )}
                <span className="cd-party-separator">·</span>
                <span className="cd-party-inline">
                  <Building2 size={13} className="text-muted" />
                  <strong>{contract.party_name || 'Pending Party'}</strong>
                </span>
              </div>
            </div>

            <div className="cd-header-right">
              <div className={`cd-status-badge ${contract.status?.toLowerCase()}`}>
                <span className="cd-status-dot"></span>
                {contract.status}
              </div>
              <button className="cd-close-btn" onClick={onClose} title="Close">
                <X size={18} />
              </button>
            </div>
          </div>

          {/* ── Tab Navigation Strip with Horizontal Scroll Controls ── */}
          <div className="cd-tabs-wrapper">
            <button 
              type="button"
              className="cd-tab-scroll-arrow left" 
              onClick={() => {
                const el = document.querySelector('.cd-tabs-strip');
                if (el) el.scrollBy({ left: -140, behavior: 'smooth' });
              }}
              title="Scroll left"
            >
              <ChevronLeft size={15} />
            </button>

            <div className="cd-tabs-strip">
              {[
                { id: 'OVERVIEW', label: 'Overview' },
                { id: 'OBLIGATIONS', label: 'Obligations & SLAs' },
                { id: 'COMPLIANCE', label: 'Compliance & Risks' },
                { id: 'PERFORMANCE', label: 'Performance' },
                { id: 'TERMS', label: 'Key Terms' },
                { id: 'DOCUMENTS', label: 'Documents' },
                { id: 'VERSIONS', label: 'Versions' },
                { id: 'AMENDMENTS', label: 'Amendments' },
                { id: 'APPROVALS', label: 'Approvals' },
                { 
                  id: 'RENEWAL_RISKS', 
                  label: (
                    <span className="tab-label-with-badge">
                      Renewal Tracking
                      {activeRisksCount > 0 && <span className="tab-badge alert">{activeRisksCount}</span>}
                    </span>
                  ) 
                },
                { id: 'COMMERCIAL_IMPACT', label: 'Commercial Impact' },
                { id: 'LINKED_RECORDS', label: 'Connected Records' },
                { id: 'TIMELINE', label: 'Lifecycle Timeline' }
              ].map(tab => (
                <button
                  key={tab.id}
                  className={`cd-tab-btn ${activeTab === tab.id ? 'active' : ''}`}
                  onClick={(e) => {
                    setActiveTab(tab.id);
                    e.currentTarget.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' });
                  }}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            <button 
              type="button"
              className="cd-tab-scroll-arrow right" 
              onClick={() => {
                const el = document.querySelector('.cd-tabs-strip');
                if (el) el.scrollBy({ left: 140, behavior: 'smooth' });
              }}
              title="Scroll right"
            >
              <ChevronRight size={15} />
            </button>
          </div>
        </div>

        {/* ── 2. Scrollable Body ── */}
        <div className="cd-body">
          
          {/* TAB 1: OVERVIEW */}
          {activeTab === 'OVERVIEW' && (
            <div className="cd-overview-container">

              {/* Quick Action Toolbar */}
              <div className="cd-actions-bar">
                <button className="cd-btn cd-btn-secondary" onClick={onEdit} disabled={contract.status === 'ARCHIVED'}>
                  <Edit2 size={14} /> Edit Agreement
                </button>
                {contract.status === 'DRAFT' && (
                  <button 
                    className="cd-btn cd-btn-primary" 
                    onClick={() => promptStatusUpdate('ACTIVE', 'Activate')}
                    disabled={isUpdatingStatus}
                  >
                    <CheckCircle size={14} /> Activate Contract
                  </button>
                )}
                {contract.status === 'ACTIVE' && (
                  <button 
                    className="cd-btn cd-btn-warning" 
                    onClick={() => promptStatusUpdate('EXPIRED', 'Expire')}
                    disabled={isUpdatingStatus}
                  >
                    <Clock size={14} /> Mark Expired
                  </button>
                )}
                {contract.status !== 'ARCHIVED' && (
                  <button 
                    className="cd-btn cd-btn-danger-outline" 
                    onClick={() => promptStatusUpdate('ARCHIVED', 'Archive')}
                    disabled={isUpdatingStatus}
                  >
                    <Archive size={14} /> Archive
                  </button>
                )}
              </div>

              {/* 3-Stat Metric Strip */}
              <div className="cd-metric-strip">
                <div className="cd-metric-card">
                  <div className="cd-metric-icon icon-blue">
                    <DollarSign size={16} />
                  </div>
                  <div className="cd-metric-info">
                    <span className="cd-metric-label">Contract Value</span>
                    <span className="cd-metric-val">
                      {contract.contract_value 
                        ? `${contract.currency || 'USD'} ${Number(contract.contract_value).toLocaleString()}` 
                        : 'Custom Terms'}
                    </span>
                  </div>
                </div>

                <div className="cd-metric-card">
                  <div className="cd-metric-icon icon-emerald">
                    <Calendar size={16} />
                  </div>
                  <div className="cd-metric-info">
                    <span className="cd-metric-label">Validity Window</span>
                    <span className="cd-metric-val">
                      {contract.effective_date ? new Date(contract.effective_date).toLocaleDateString('en-US', { month: 'short', year: 'numeric' }) : 'Open'}
                      {' → '}
                      {contract.expiry_date ? new Date(contract.expiry_date).toLocaleDateString('en-US', { month: 'short', year: 'numeric' }) : 'Indefinite'}
                    </span>
                  </div>
                </div>

                <div className="cd-metric-card">
                  <div className="cd-metric-icon icon-purple">
                    {getModeIcon(contract.transport_mode)}
                  </div>
                  <div className="cd-metric-info">
                    <span className="cd-metric-label">Transport Mode</span>
                    <span className="cd-metric-val">{contract.transport_mode || 'Multimodal'}</span>
                  </div>
                </div>
              </div>

              {/* Lifecycle Health Intelligence Banner */}
              <div className={`cd-health-banner ${health.badgeClass}`}>
                <div className="cd-hb-top">
                  <div className="cd-hb-left">
                    <div className="cd-hb-icon">{health.icon}</div>
                    <div>
                      <h4 className="cd-hb-title">{health.label}</h4>
                      <p className="cd-hb-subtext">{health.subtext}</p>
                    </div>
                  </div>
                  <div className="cd-hb-timing">
                    <span className="cd-hb-days">{health.daysText}</span>
                  </div>
                </div>
                <div className="cd-hb-progress-bar">
                  <div className="cd-hb-progress-fill" style={{ width: `${health.progress}%` }}></div>
                </div>
              </div>

              {/* Renewal In-Progress Banner (if applicable) */}
              {renewalTracking && renewalTracking.renewal_status === 'IN_PROGRESS' && (
                <div className="cd-renewal-callout">
                  <div className="renewal-callout-header">
                    <Sparkles size={16} className="text-indigo" />
                    <strong>Commercial Renewal Active</strong>
                  </div>
                  <p className="renewal-callout-desc">
                    Assigned Owner: <strong>{renewalTracking.owner || 'Commercial Lead'}</strong> · Target: <strong>{renewalTracking.target_completion_date || 'TBD'}</strong>
                  </p>
                  <button className="cd-btn cd-btn-secondary btn-sm" onClick={() => setActiveTab('RENEWAL_RISKS')}>
                    View Renewal Workflow →
                  </button>
                </div>
              )}

              {/* Structured Key-Value Information Grid */}
              <div className="cd-details-grid">
                
                {/* Panel 1: Agreement & Parties */}
                <div className="cd-card">
                  <div className="cd-card-header">
                    <Building2 size={15} />
                    <span>Commercial Counterparty & SLA</span>
                  </div>
                  <div className="cd-card-body">
                    <div className="cd-kv-row">
                      <span className="cd-k">Counterparty</span>
                      <span className="cd-v font-bold">{contract.party_name || '—'}</span>
                    </div>
                    <div className="cd-kv-row">
                      <span className="cd-k">Party Category</span>
                      <span className="cd-v">{contract.contract_type || 'Commercial Partner'}</span>
                    </div>
                    <div className="cd-kv-row">
                      <span className="cd-k">Internal Owner</span>
                      <span className="cd-v">{contract.owner || 'Commercial Operations'}</span>
                    </div>
                    <div className="cd-kv-row">
                      <span className="cd-k">Lifecycle Status</span>
                      <span className="cd-v">{contract.status}</span>
                    </div>
                  </div>
                </div>

                {/* Panel 2: Term & Validity Dates */}
                <div className="cd-card">
                  <div className="cd-card-header">
                    <Calendar size={15} />
                    <span>Term & Validity Details</span>
                  </div>
                  <div className="cd-card-body">
                    <div className="cd-kv-row">
                      <span className="cd-k">Effective Date</span>
                      <span className="cd-v font-mono">{contract.effective_date || 'Immediate'}</span>
                    </div>
                    <div className="cd-kv-row">
                      <span className="cd-k">Expiration Date</span>
                      <span className="cd-v font-mono">{contract.expiry_date || 'Continuous / No Expiry'}</span>
                    </div>
                    <div className="cd-kv-row">
                      <span className="cd-k">Registered On</span>
                      <span className="cd-v">{contract.created_at ? new Date(contract.created_at).toLocaleDateString() : '—'}</span>
                    </div>
                    <div className="cd-kv-row">
                      <span className="cd-k">Last Modified</span>
                      <span className="cd-v">{contract.updated_at ? new Date(contract.updated_at).toLocaleDateString() : '—'}</span>
                    </div>
                  </div>
                </div>

              </div>

              {/* Commercial Scope & Description */}
              <div className="cd-card">
                <div className="cd-card-header">
                  <FileText size={15} />
                  <span>Commercial Scope & Internal Notes</span>
                </div>
                <div className="cd-card-body">
                  <p className="cd-desc-text">
                    {contract.description || 'No detailed scope description registered for this commercial agreement.'}
                  </p>
                  {contract.notes && (
                    <div className="cd-notes-callout">
                      <Info size={14} className="cd-notes-icon" />
                      <span>{contract.notes}</span>
                    </div>
                  )}
                </div>
              </div>

            </div>
          )}

          {/* TAB 2: RENEWAL & RISKS */}
          {activeTab === 'RENEWAL_RISKS' && (
            <div className="cd-renewal-container">
              
              {/* Renewal Workflow Card */}
              <div className="cd-card">
                <div className="cd-card-header">
                  <Sparkles size={15} />
                  <span>Commercial Renewal Workflow</span>
                </div>
                <div className="cd-card-body">
                  
                  {(!renewalTracking || renewalTracking.renewal_status === 'NOT_STARTED') && (
                    <div className="renewal-not-started-state">
                      <div className="renewal-prompt-icon">
                        <Clock size={28} />
                      </div>
                      <h4>Renewal Not Started</h4>
                      <p>
                        This agreement has not yet entered formal renegotiation. Initiating renewal tracks deadlines, targets, assigned owners, and successor contract linkages.
                      </p>
                      <button className="cd-btn cd-btn-primary" onClick={() => setShowRenewalModal(true)}>
                        <Sparkles size={14} /> Initiate Commercial Renewal
                      </button>
                    </div>
                  )}

                  {renewalTracking && renewalTracking.renewal_status === 'IN_PROGRESS' && (
                    <div className="renewal-in-progress-view">
                      <div className="renewal-status-bar in-progress">
                        <Sparkles size={16} />
                        <div>
                          <strong>Renewal Currently In Progress</strong>
                          <p>Target Date: {renewalTracking.target_completion_date || 'Not set'} · Owner: {renewalTracking.owner || 'Unassigned'}</p>
                        </div>
                      </div>

                      <div className="cd-kv-row">
                        <span className="cd-k">Renewal Start Date</span>
                        <span className="cd-v font-mono">{renewalTracking.renewal_start_date || '—'}</span>
                      </div>
                      <div className="cd-kv-row">
                        <span className="cd-k">Target Completion</span>
                        <span className="cd-v font-mono">{renewalTracking.target_completion_date || '—'}</span>
                      </div>
                      <div className="cd-kv-row">
                        <span className="cd-k">Renewal Notes</span>
                        <span className="cd-v">{renewalTracking.notes || 'None recorded'}</span>
                      </div>

                      <div className="renewal-action-buttons">
                        <button className="cd-btn cd-btn-primary btn-sm" onClick={() => setShowSuccessorModal(true)}>
                          <CheckCircle size={14} /> Link Successor Contract & Complete
                        </button>
                        <button className="cd-btn cd-btn-secondary btn-sm" onClick={() => handleUpdateRenewalStatus('ABANDONED')}>
                          <XCircle size={14} /> Abandon Renewal
                        </button>
                      </div>
                    </div>
                  )}

                  {renewalTracking && renewalTracking.renewal_status === 'RENEWED' && (
                    <div className="renewal-completed-view">
                      <div className="renewal-status-bar completed">
                        <CheckCircle2 size={16} />
                        <div>
                          <strong>Agreement Successfully Renewed</strong>
                          <p>
                            Successor Contract: <strong>{renewalTracking.successor_name || `Contract #${renewalTracking.successor_contract_id}`}</strong> ({renewalTracking.successor_reference || 'REF-LINKED'})
                          </p>
                        </div>
                      </div>
                    </div>
                  )}

                </div>
              </div>

              {/* Risk Events Section */}
              <div className="cd-card">
                <div className="cd-card-header">
                  <ShieldAlert size={15} />
                  <span>Commercial & Lifecycle Risks ({riskEvents.length})</span>
                </div>
                <div className="cd-card-body">
                  {riskEvents.length === 0 ? (
                    <div className="cd-empty-state-sm">
                      <CheckCircle2 size={24} className="text-emerald" />
                      <p>No unresolved commercial or lifecycle risk events detected.</p>
                    </div>
                  ) : (
                    <div className="risks-list">
                      {riskEvents.map(risk => (
                        <div key={risk.id} className={`risk-item-card ${risk.is_resolved ? 'resolved' : risk.severity.toLowerCase()}`}>
                          <div className="risk-item-top">
                            <div className="risk-item-badge">
                              <span className={`risk-severity-pill ${risk.severity.toLowerCase()}`}>
                                {risk.severity}
                              </span>
                              <span className="risk-type-text">{risk.risk_type.replace(/_/g, ' ')}</span>
                            </div>
                            {risk.is_resolved ? (
                              <span className="risk-resolved-tag">
                                <Check size={11} /> Resolved
                              </span>
                            ) : (
                              <button 
                                className="risk-resolve-btn"
                                onClick={() => setResolvingRisk(risk)}
                              >
                                Resolve Risk
                              </button>
                            )}
                          </div>
                          <p className="risk-item-desc">{risk.description}</p>
                          {risk.is_resolved && risk.resolution_notes && (
                            <p className="risk-res-notes">Resolution: {risk.resolution_notes}</p>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

            </div>
          )}

          {/* TAB 3: COMMERCIAL IMPACT */}
          {activeTab === 'COMMERCIAL_IMPACT' && (
            <div className="cd-impact-container">
              
              <div className="cd-impact-banner">
                <Info size={16} />
                <span>
                  <strong>Read-Only Impact Analysis:</strong> Downstream dependencies across Managed Rates, Quotations, and Spot Requests calculated deterministically. Historical rates & quotation snapshots are strictly preserved.
                </span>
              </div>

              <div className="cd-impact-grid">
                
                {/* 1. Rates Impact */}
                <div className="cd-impact-card">
                  <div className="impact-card-top">
                    <DollarSign size={18} className="text-blue" />
                    <span className="impact-card-tag">Managed Rates</span>
                  </div>
                  <div className="impact-card-value">
                    {commercialImpact?.active_rates_count ?? 0}
                    <span className="impact-val-sub">/ {commercialImpact?.linked_rates_count ?? 0} Active</span>
                  </div>
                  <p className="impact-card-desc">
                    Active commercial freight rates referencing this agreement's terms.
                  </p>
                </div>

                {/* 2. Quotations Impact */}
                <div className="cd-impact-card">
                  <div className="impact-card-top">
                    <FileText size={18} className="text-emerald" />
                    <span className="impact-card-tag">Quotations</span>
                  </div>
                  <div className="impact-card-value">
                    {commercialImpact?.accepted_quotations_count ?? 0}
                    <span className="impact-val-sub">Accepted ({commercialImpact?.draft_quotations_count ?? 0} Draft)</span>
                  </div>
                  <p className="impact-card-desc">
                    Accepted customer quotation proposals and active draft estimates.
                  </p>
                </div>

                {/* 3. Spot Sourcing */}
                <div className="cd-impact-card">
                  <div className="impact-card-top">
                    <Layers size={18} className="text-purple" />
                    <span className="impact-card-tag">Spot Activity</span>
                  </div>
                  <div className="impact-card-value">
                    {(commercialImpact?.spot_requests_count ?? 0) + (commercialImpact?.spot_responses_count ?? 0)}
                    <span className="impact-val-sub">Requests & Responses</span>
                  </div>
                  <p className="impact-card-desc">
                    Spot sourcing RFQs and carrier responses tied to this contract SLA.
                  </p>
                </div>

                {/* 4. Commercial Value Exposure */}
                <div className="cd-impact-card">
                  <div className="impact-card-top">
                    <TrendingUp size={18} className="text-amber" />
                    <span className="impact-card-tag">Exposure</span>
                  </div>
                  <div className="impact-card-value">
                    {contract.currency || 'USD'} {(commercialImpact?.total_commercial_exposure ?? 0).toLocaleString()}
                  </div>
                  <p className="impact-card-desc">
                    Cumulative quotation value operating under these contract terms.
                  </p>
                </div>

              </div>

            </div>
          )}

          {/* TAB 4: CONNECTED RECORDS */}
          {activeTab === 'LINKED_RECORDS' && (
            <div className="cd-linked-container">
              <ContractLinkedRecordsPanel 
                contractId={contract?.id}
                contractStatus={contract?.status}
                contract={contract} 
                onLinksChanged={() => {
                  fetchIntelligenceData();
                  if (onUpdate) onUpdate();
                }} 
              />
            </div>
          )}

          {/* TAB 5: LIFECYCLE TIMELINE */}
          {activeTab === 'TIMELINE' && (
            <div className="cd-timeline-container">
              {/* Header Overview Card */}
              <div className="timeline-header-card">
                <div className="timeline-header-main">
                  <div>
                    <h4 className="timeline-heading">Contract Lifecycle & Audit Trail</h4>
                    <p className="timeline-subheading">
                      Immutable sequence of version snapshots, status changes, and counterparty governance.
                    </p>
                  </div>
                  <div className="timeline-stats-group">
                    <span className="timeline-stat-pill">
                      <strong>{events.length}</strong> Audit Events
                    </span>
                    <span className={`timeline-status-pill ${contract.status?.toLowerCase() || 'active'}`}>
                      {contract.status || 'ACTIVE'}
                    </span>
                  </div>
                </div>
              </div>

              {isLoadingEvents ? (
                <div className="cd-loading-state">
                  <RefreshCw className="cd-spin" size={24} />
                  <span>Aggregating commercial audit timeline...</span>
                </div>
              ) : events.length === 0 ? (
                <div className="cd-empty-timeline">
                  <Clock size={32} className="text-muted" />
                  <h4>No Lifecycle Events Recorded</h4>
                  <p>All status transitions, amendments, and governance milestones will be audited here.</p>
                </div>
              ) : (
                <div className="cd-timeline-flow">
                  {events.map((ev, index) => {
                    const eventType = (ev.event_type || ev.action || 'EVENT').toUpperCase();
                    let badgeClass = 'type-blue';
                    if (eventType.includes('STATUS') || eventType.includes('ACTIVE') || eventType.includes('RESOLVED')) {
                      badgeClass = 'type-emerald';
                    } else if (eventType.includes('AMENDMENT') || eventType.includes('EDIT')) {
                      badgeClass = 'type-amber';
                    } else if (eventType.includes('OBLIGATION') || eventType.includes('SLA')) {
                      badgeClass = 'type-purple';
                    } else if (eventType.includes('DOC') || eventType.includes('UPLOAD')) {
                      badgeClass = 'type-indigo';
                    }

                    return (
                      <div key={ev.id || index} className={`cd-timeline-node ${badgeClass}`}>
                        <div className={`cd-node-icon-box ${badgeClass}`}>
                          {getEventIcon(ev.event_type || ev.action)}
                        </div>
                        <div className="cd-node-content">
                          <div className="cd-node-header">
                            <div className="cd-header-tags">
                              <span className={`cd-node-type ${badgeClass}`}>
                                {eventType.replace(/_/g, ' ')}
                              </span>
                              {ev.new_status && (
                                <span className="cd-status-trans-tag">
                                  {ev.previous_status ? `${ev.previous_status} → ` : ''}{ev.new_status}
                                </span>
                              )}
                            </div>
                            <span className="cd-node-date">
                              <Clock size={11} />
                              {ev.created_at || ev.event_date ? new Date(ev.created_at || ev.event_date).toLocaleString() : 'Recent'}
                            </span>
                          </div>

                          <p className="cd-node-desc">
                            {ev.description || (ev.action ? `${ev.action} ${ev.linked_entity_type || 'record'}` : 'Lifecycle record')}
                          </p>

                          <div className="cd-node-footer">
                            <span className="cd-node-user">
                              <User size={11} />
                              <span>{ev.performed_by ? (ev.performed_by.includes('@') ? ev.performed_by : `User: ${ev.performed_by}`) : 'System Automated'}</span>
                            </span>
                          </div>
                        </div>
                      </div>
                    );
                  })}

                  {/* Genesis Milestone Anchor */}
                  <div className="cd-timeline-genesis">
                    <div className="genesis-dot"></div>
                    <span className="genesis-text">Contract Created & Baseline Initialized</span>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* TAB: VERSIONS */}
          {activeTab === 'VERSIONS' && (
            <ContractVersionsPanel
              contractId={contract.id}
              contractStatus={contract.status}
              onVersionPromoted={() => {
                if (onUpdate) onUpdate();
                fetchIntelligenceData();
              }}
            />
          )}

          {/* TAB: AMENDMENTS */}
          {activeTab === 'AMENDMENTS' && (
            <ContractAmendmentsPanel
              contractId={contract.id}
              contractStatus={contract.status}
              onAmendmentImplemented={() => {
                if (onUpdate) onUpdate();
                fetchIntelligenceData();
              }}
            />
          )}

          {/* TAB: APPROVALS */}
          {activeTab === 'APPROVALS' && (
            <ContractApprovalPanel
              contractId={contract.id}
              contractStatus={contract.status}
              onApprovalCompleted={() => {
                if (onUpdate) onUpdate();
                fetchIntelligenceData();
              }}
            />
          )}

          {/* TAB: OBLIGATIONS */}
          {activeTab === 'OBLIGATIONS' && (
            <div className="cd-tab-panel-container">
              <ContractObligationsPanel contract={contract} />
            </div>
          )}

          {/* TAB: COMPLIANCE */}
          {activeTab === 'COMPLIANCE' && (
            <div className="cd-tab-panel-container">
              <ContractCompliancePanel contract={contract} />
            </div>
          )}

          {/* TAB: PERFORMANCE */}
          {activeTab === 'PERFORMANCE' && (
            <div className="cd-tab-panel-container">
              <ContractPerformancePanel contract={contract} />
            </div>
          )}

          {/* TAB: TERMS */}
          {activeTab === 'TERMS' && (
            <div className="cd-tab-panel-container">
              <ContractTermsPanel contract={contract} />
            </div>
          )}

          {/* TAB: DOCUMENTS */}
          {activeTab === 'DOCUMENTS' && (
            <div className="cd-tab-panel-container">
              <ContractDocumentsPanel contract={contract} />
            </div>
          )}

        </div>

      </div>

      {/* ── Modal: Initiate Renewal ── */}
      {showRenewalModal && (
        <div className="contract-form-backdrop" onClick={() => setShowRenewalModal(false)}>
          <div className="cd-modal-dialog" onClick={e => e.stopPropagation()}>
            <div className="cd-modal-header">
              <h3>Initiate Commercial Renewal</h3>
              <button className="cd-close-btn" onClick={() => setShowRenewalModal(false)}>
                <X size={16} />
              </button>
            </div>
            <div className="cd-modal-body">
              <div className="cf-field-group">
                <label className="cf-label">Target Completion Date</label>
                <input
                  type="date"
                  className="cf-input"
                  value={renewalForm.target_completion_date}
                  onChange={e => setRenewalForm({ ...renewalForm, target_completion_date: e.target.value })}
                />
              </div>
              <div className="cf-field-group">
                <label className="cf-label">Assigned Owner</label>
                <input
                  type="text"
                  className="cf-input"
                  placeholder="e.g. John Doe / Commercial Lead"
                  value={renewalForm.owner}
                  onChange={e => setRenewalForm({ ...renewalForm, owner: e.target.value })}
                />
              </div>
              <div className="cf-field-group">
                <label className="cf-label">Renewal Strategy / Notes</label>
                <textarea
                  className="cf-textarea"
                  rows={3}
                  placeholder="e.g. Renegotiate Q4 fuel surcharges and demurrage free time..."
                  value={renewalForm.notes}
                  onChange={e => setRenewalForm({ ...renewalForm, notes: e.target.value })}
                />
              </div>
            </div>
            <div className="cd-modal-footer">
              <button className="cd-btn cd-btn-secondary" onClick={() => setShowRenewalModal(false)}>
                Cancel
              </button>
              <button className="cd-btn cd-btn-primary" onClick={handleStartRenewal}>
                Start Renewal Workflow
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Modal: Link Successor Contract ── */}
      {showSuccessorModal && (
        <div className="contract-form-backdrop" onClick={() => setShowSuccessorModal(false)}>
          <div className="cd-modal-dialog" onClick={e => e.stopPropagation()}>
            <div className="cd-modal-header">
              <h3>Link Successor Contract</h3>
              <button className="cd-close-btn" onClick={() => setShowSuccessorModal(false)}>
                <X size={16} />
              </button>
            </div>
            <div className="cd-modal-body">
              <p className="cd-modal-sub">
                Linking a successor contract establishes permanent commercial traceability and transitions this agreement's renewal status to <strong>RENEWED</strong>.
              </p>
              <div className="cf-field-group">
                <label className="cf-label">Successor Contract ID</label>
                <input
                  type="number"
                  className="cf-input"
                  placeholder="Enter ID of newly executed contract (e.g. 2)"
                  value={successorIDInput}
                  onChange={e => setSuccessorIDInput(e.target.value)}
                />
              </div>
            </div>
            <div className="cd-modal-footer">
              <button className="cd-btn cd-btn-secondary" onClick={() => setShowSuccessorModal(false)}>
                Cancel
              </button>
              <button className="cd-btn cd-btn-primary" onClick={handleLinkSuccessor}>
                Complete Renewal & Link
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Modal: Resolve Risk ── */}
      {resolvingRisk && (
        <div className="contract-form-backdrop" onClick={() => setResolvingRisk(null)}>
          <div className="cd-modal-dialog" onClick={e => e.stopPropagation()}>
            <div className="cd-modal-header">
              <h3>Resolve Commercial Risk</h3>
              <button className="cd-close-btn" onClick={() => setResolvingRisk(null)}>
                <X size={16} />
              </button>
            </div>
            <div className="cd-modal-body">
              <p className="cd-modal-sub">
                Risk: <strong>{resolvingRisk.description}</strong>
              </p>
              <div className="cf-field-group">
                <label className="cf-label">Resolution Notes / Action Taken</label>
                <textarea
                  className="cf-textarea"
                  rows={3}
                  placeholder="e.g. Verified with Maersk commercial rep; extension addendum executed."
                  value={resolutionNotes}
                  onChange={e => setResolutionNotes(e.target.value)}
                />
              </div>
            </div>
            <div className="cd-modal-footer">
              <button className="cd-btn cd-btn-secondary" onClick={() => setResolvingRisk(null)}>
                Cancel
              </button>
              <button className="cd-btn cd-btn-primary" onClick={handleResolveRisk}>
                Confirm Resolution
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Status Confirmation Modal ── */}
      <ConfirmModal
        isOpen={statusConfirmModal.isOpen}
        title={statusConfirmModal.title}
        message={statusConfirmModal.message}
        confirmText={statusConfirmModal.confirmText}
        confirmStyle={statusConfirmModal.confirmStyle}
        isLoading={isUpdatingStatus}
        onConfirm={handleConfirmedStatusUpdate}
        onCancel={() => setStatusConfirmModal(prev => ({ ...prev, isOpen: false }))}
      />

    </div>
  );
}
