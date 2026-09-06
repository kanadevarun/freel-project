import React, { useState, useEffect, useCallback } from 'react';
import {
  FileText, Plus, RefreshCw, Calendar, Clock, AlertTriangle, CheckCircle,
  ChevronRight, Search, Filter, ArrowUpRight, Shield, Layers, MoreVertical,
  Building2, Hash, Edit2, Archive, X
} from 'lucide-react';
import { rateService } from '../../../services/rateService';
import ConfirmModal from '../Settings/ConfirmModal';
import toast from 'react-hot-toast';
import './RateContractsWorkspace.css';

const CONTRACT_TYPES = [
  { value: 'ALL', label: 'All Contract Types' },
  { value: 'ANNUAL_SERVICE', label: 'Annual Service Agreement' },
  { value: 'NAMED_ACCOUNT', label: 'Named Account Contract' },
  { value: 'TIERED_VOLUME', label: 'Tiered Volume Agreement' },
  { value: 'SPOT_COMMITMENT', label: 'Spot Commitment Contract' },
  { value: 'FRAMEWORK', label: 'Framework Tariff' },
];

const CONTRACT_STATUSES = [
  { value: 'ALL', label: 'All Statuses' },
  { value: 'ACTIVE', label: 'Active' },
  { value: 'EXPIRING_SOON', label: 'Expiring Soon' },
  { value: 'EXPIRED', label: 'Expired' },
  { value: 'DRAFT', label: 'Draft' },
  { value: 'ARCHIVED', label: 'Archived' },
];

const RENEWAL_STATUSES = [
  { value: 'ALL', label: 'All Renewals' },
  { value: 'NOT_STARTED', label: 'Not Started' },
  { value: 'IN_PROGRESS', label: 'In Progress' },
  { value: 'RENEWED', label: 'Renewed' },
  { value: 'NOT_RENEWING', label: 'Not Renewing' },
];

export default function RateContractsWorkspace({ CustomSelect, onOpenRateDetail }) {
  const [contracts, setContracts] = useState([]);
  const [summary, setSummary] = useState({
    total_contracts: 0,
    active_contracts: 0,
    expiring_soon_contracts: 0,
    expired_contracts: 0,
    renewal_required: 0,
    total_linked_rates: 0,
    expiring_soon_list: [],
  });
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [selectedType, setSelectedType] = useState('ALL');
  const [selectedStatus, setSelectedStatus] = useState('ALL');
  const [selectedRenewal, setSelectedRenewal] = useState('ALL');

  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showRenewModal, setShowRenewModal] = useState(false);
  const [activeContractForRenew, setActiveContractForRenew] = useState(null);
  const [renewExpiryDate, setRenewExpiryDate] = useState('');
  const [renewStatus, setRenewStatus] = useState('RENEWED');
  const [renewNotes, setRenewNotes] = useState('');

  // New Contract Form State
  const [newContract, setNewContract] = useState({
    contract_reference: '',
    carrier_name: '',
    carrier_code: '',
    contract_name: '',
    contract_type: 'ANNUAL_SERVICE',
    transport_mode: 'Ocean FCL',
    currency: 'USD',
    effective_date: '',
    expiry_date: '',
    renewal_owner: '',
    notes: '',
  });

  const fetchContractsData = useCallback(async () => {
    try {
      setLoading(true);
      const [contractsRes, summaryRes] = await Promise.all([
        rateService.getRateContracts({
          search,
          contract_type: selectedType,
          status: selectedStatus,
          renewal_status: selectedRenewal,
        }),
        rateService.getRateContractSummary(),
      ]);

      const contractsData = contractsRes?.data?.contracts || contractsRes?.contracts || contractsRes?.data || (Array.isArray(contractsRes) ? contractsRes : []);
      setContracts(Array.isArray(contractsData) ? contractsData : []);

      const summaryData = summaryRes?.data || summaryRes || {};
      if (summaryData && typeof summaryData === 'object') {
        setSummary({
          total_contracts: summaryData.total_contracts || 0,
          active_contracts: summaryData.active_contracts || 0,
          expiring_soon_contracts: summaryData.expiring_soon_contracts || 0,
          expired_contracts: summaryData.expired_contracts || 0,
          renewal_required: summaryData.renewal_required || 0,
          total_linked_rates: summaryData.total_linked_rates || 0,
          expiring_soon_list: summaryData.expiring_soon_list || [],
        });
      }
    } catch (err) {
      console.error('Failed to load rate contracts:', err);
      toast.error('Failed to load carrier rate contracts');
    } finally {
      setLoading(false);
    }
  }, [search, selectedType, selectedStatus, selectedRenewal]);

  useEffect(() => {
    fetchContractsData();
  }, [fetchContractsData]);

  const handleCreateContract = async (e) => {
    e.preventDefault();
    if (!newContract.carrier_name || !newContract.contract_name || !newContract.effective_date || !newContract.expiry_date) {
      toast.error('Please fill in all required fields');
      return;
    }
    try {
      await rateService.createRateContract(newContract);
      toast.success('Carrier rate contract created successfully');
      setShowCreateModal(false);
      setNewContract({
        contract_reference: '',
        carrier_name: '',
        carrier_code: '',
        contract_name: '',
        contract_type: 'ANNUAL_SERVICE',
        transport_mode: 'Ocean FCL',
        currency: 'USD',
        effective_date: '',
        expiry_date: '',
        renewal_owner: '',
        notes: '',
      });
      fetchContractsData();
    } catch (err) {
      toast.error(err?.message || err?.details?.error?.message || 'Failed to create contract');
    }
  };

  const handleRenewContract = async (e) => {
    e.preventDefault();
    if (!activeContractForRenew || !renewExpiryDate) {
      toast.error('Please specify the new contract expiry date');
      return;
    }
    try {
      await rateService.renewRateContract(activeContractForRenew.id, {
        new_expiry_date: renewExpiryDate,
        renewal_status: renewStatus,
        notes: renewNotes || undefined,
      });
      toast.success('Contract renewed successfully');
      setShowRenewModal(false);
      setActiveContractForRenew(null);
      setRenewExpiryDate('');
      fetchContractsData();
    } catch (err) {
      toast.error(err?.message || err?.details?.error?.message || 'Failed to renew contract');
    }
  };

  // Archive Confirmation Modal State
  const [archiveModalOpen, setArchiveModalOpen] = useState(false);
  const [contractToArchive, setContractToArchive] = useState(null);
  const [archiveLoading, setArchiveLoading] = useState(false);

  const triggerArchiveContract = (contract) => {
    setContractToArchive(contract);
    setArchiveModalOpen(true);
  };

  const handleConfirmArchive = async () => {
    if (!contractToArchive) return;
    setArchiveLoading(true);
    try {
      await rateService.archiveRateContract(contractToArchive.id);
      toast.success(`Contract ${contractToArchive.contract_reference} archived`);
      setArchiveModalOpen(false);
      setContractToArchive(null);
      fetchContractsData();
    } catch (err) {
      toast.error(err?.message || 'Failed to archive contract');
    } finally {
      setArchiveLoading(false);
    }
  };

  const openQuickRenew = (contract) => {
    setActiveContractForRenew(contract);
    // Suggest +1 year from current expiry
    if (contract.expiry_date) {
      const parts = contract.expiry_date.split('-');
      if (parts.length === 3) {
        const nextYear = parseInt(parts[0], 10) + 1;
        setRenewExpiryDate(`${nextYear}-${parts[1]}-${parts[2]}`);
      }
    }
    setRenewStatus('RENEWED');
    setShowRenewModal(true);
  };

  return (
    <div className="rc-workspace">
      {/* ── KPI Summary Strip ────────────────────────────────────────────── */}
      <div className="rc-kpi-grid">
        <div className="rc-kpi-card">
          <div className="rc-kpi-header">
            <span className="rc-kpi-title">Total Contracts</span>
            <div className="rc-kpi-icon" style={{ background: '#EFF6FF', color: '#2563EB' }}>
              <FileText size={16} />
            </div>
          </div>
          <div className="rc-kpi-val">{summary.total_contracts}</div>
          <span className="rc-kpi-sub">Commercial Agreements</span>
        </div>

        <div className="rc-kpi-card">
          <div className="rc-kpi-header">
            <span className="rc-kpi-title">Active</span>
            <div className="rc-kpi-icon" style={{ background: '#ECFDF5', color: '#059669' }}>
              <CheckCircle size={16} />
            </div>
          </div>
          <div className="rc-kpi-val" style={{ color: '#059669' }}>{summary.active_contracts}</div>
          <span className="rc-kpi-sub">Valid & In-Force</span>
        </div>

        <div className="rc-kpi-card">
          <div className="rc-kpi-header">
            <span className="rc-kpi-title">Expiring Soon</span>
            <div className="rc-kpi-icon" style={{ background: '#FFFBEB', color: '#D97706' }}>
              <Clock size={16} />
            </div>
          </div>
          <div className="rc-kpi-val" style={{ color: '#D97706' }}>{summary.expiring_soon_contracts}</div>
          <span className="rc-kpi-sub">Within 30 Days</span>
        </div>

        <div className="rc-kpi-card">
          <div className="rc-kpi-header">
            <span className="rc-kpi-title">Expired</span>
            <div className="rc-kpi-icon" style={{ background: '#FEF2F2', color: '#DC2626' }}>
              <AlertTriangle size={16} />
            </div>
          </div>
          <div className="rc-kpi-val" style={{ color: '#DC2626' }}>{summary.expired_contracts}</div>
          <span className="rc-kpi-sub">Needs Renewal</span>
        </div>

        <div className="rc-kpi-card">
          <div className="rc-kpi-header">
            <span className="rc-kpi-title">Renewal Action</span>
            <div className="rc-kpi-icon" style={{ background: '#F5F3FF', color: '#7C3AED' }}>
              <RefreshCw size={16} />
            </div>
          </div>
          <div className="rc-kpi-val" style={{ color: '#7C3AED' }}>{summary.renewal_required}</div>
          <span className="rc-kpi-sub">Pending Negotiation</span>
        </div>

        <div className="rc-kpi-card">
          <div className="rc-kpi-header">
            <span className="rc-kpi-title">Linked Rates</span>
            <div className="rc-kpi-icon" style={{ background: '#F0FDFA', color: '#0D9488' }}>
              <Layers size={16} />
            </div>
          </div>
          <div className="rc-kpi-val" style={{ color: '#0D9488' }}>{summary.total_linked_rates}</div>
          <span className="rc-kpi-sub">Tariff Lines</span>
        </div>
      </div>

      {/* ── Main Contracts & Renewal Layout ──────────────────────────────── */}
      <div className="rc-layout-grid">
        {/* Left Column: Contracts Table */}
        <div className="rc-card">
          <div className="rc-card-header">
            <div className="rc-card-title-row">
              <h2 className="rc-card-title">Carrier Rate Contracts</h2>
              <span className="rm-table-count-badge">{contracts.length} records</span>
            </div>
            <div className="rc-card-actions">
              <button className="rc-btn-primary" onClick={() => setShowCreateModal(true)}>
                <Plus size={15} /> Add Carrier Contract
              </button>
            </div>
          </div>

          {/* Filter Bar */}
          <div className="rc-filter-bar">
            <div className="rc-search-input-wrap">
              <Search size={14} />
              <input
                type="text"
                placeholder="Search contract ref, carrier, agreement name..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="rc-search-input"
              />
            </div>

            <CustomSelect
              value={selectedType}
              onChange={setSelectedType}
              options={CONTRACT_TYPES}
              placeholder="Contract Type"
            />

            <CustomSelect
              value={selectedStatus}
              onChange={setSelectedStatus}
              options={CONTRACT_STATUSES}
              placeholder="Status"
            />

            <CustomSelect
              value={selectedRenewal}
              onChange={setSelectedRenewal}
              options={RENEWAL_STATUSES}
              placeholder="Renewal Status"
            />
          </div>

          {/* Table */}
          <div className="rc-table-wrap">
            <table className="rc-table">
              <thead>
                <tr>
                  <th>Contract Ref</th>
                  <th>Carrier</th>
                  <th>Contract Name</th>
                  <th>Transport Mode</th>
                  <th>Effective Date</th>
                  <th>Expiry Date</th>
                  <th>Days Left</th>
                  <th>Rates</th>
                  <th>Renewal</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {contracts.length === 0 ? (
                  <tr>
                    <td colSpan="11" style={{ textAlign: 'center', padding: '48px 16px', color: '#64748B' }}>
                      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '8px' }}>
                        <FileText size={32} style={{ color: '#94A3B8' }} />
                        <span style={{ fontWeight: 700, color: '#0F172A' }}>No carrier contracts yet</span>
                        <span style={{ fontSize: '12.5px' }}>Create your first carrier agreement to manage contract rates, validity periods, and renewals.</span>
                      </div>
                    </td>
                  </tr>
                ) : (
                  contracts.map((c) => (
                    <tr key={c.id}>
                      <td>
                        <span className="rc-contract-ref">{c.contract_reference}</span>
                      </td>
                      <td>
                        <div className="rc-carrier-badge">
                          <Building2 size={13} style={{ color: '#64748B' }} />
                          <span>{c.carrier_name}</span>
                        </div>
                      </td>
                      <td style={{ fontWeight: 600, color: '#0F172A' }}>{c.contract_name}</td>
                      <td>{c.transport_mode}</td>
                      <td style={{ color: '#64748B' }}>{c.effective_date}</td>
                      <td style={{ fontWeight: 600 }}>{c.expiry_date}</td>
                      <td>
                        {c.days_until_expiry <= 0 ? (
                          <span className="rc-days-badge rc-days-critical">Expired</span>
                        ) : c.days_until_expiry <= 30 ? (
                          <span className="rc-days-badge rc-days-warning">{c.days_until_expiry}d left</span>
                        ) : (
                          <span style={{ color: '#64748B' }}>{c.days_until_expiry}d</span>
                        )}
                      </td>
                      <td>
                        <span style={{ fontWeight: 750, color: '#2563EB' }}>{c.linked_rate_count}</span>
                      </td>
                      <td>
                        <span className={`rc-renewal-badge rc-renewal-${c.renewal_status.toLowerCase().replace(/_/g, '-')}`}>
                          {c.renewal_status.replace(/_/g, ' ')}
                        </span>
                      </td>
                      <td>
                        <span className={`rc-badge-pill rc-badge-${c.status.toLowerCase()}`}>
                          {c.status.replace(/_/g, ' ')}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <div style={{ display: 'inline-flex', gap: '6px' }}>
                          <button
                            className="rc-btn-secondary"
                            style={{ padding: '4px 8px', fontSize: '12px' }}
                            title="Renew Contract"
                            onClick={() => openQuickRenew(c)}
                          >
                            <RefreshCw size={12} /> Renew
                          </button>
                          <button
                            className="rc-btn-secondary"
                            style={{ padding: '4px 8px', fontSize: '12px', color: '#DC2626' }}
                            title="Archive Contract"
                            onClick={() => triggerArchiveContract(c)}
                          >
                            <Archive size={12} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Right Column: Expiry & Renewal Panel */}
        <div className="rc-sidebar-panel">
          <div className="rc-expiring-card">
            <div className="rc-expiring-header">
              <span className="rc-expiring-title">
                <AlertTriangle size={15} style={{ color: '#D97706' }} /> Expiring & Critical Contracts
              </span>
              <span style={{ fontSize: '11px', fontWeight: 700, color: '#64748B' }}>
                {summary.expiring_soon_list?.length || 0} Listed
              </span>
            </div>

            {(!summary.expiring_soon_list || summary.expiring_soon_list.length === 0) ? (
              <div style={{ padding: '24px 12px', textAlign: 'center', color: '#94A3B8', fontSize: '12.5px' }}>
                No carrier contracts currently expiring within 30 days.
              </div>
            ) : (
              summary.expiring_soon_list.map((item) => (
                <div key={item.id} className="rc-expiring-item">
                  <div className="rc-expiring-top">
                    <span className="rc-expiring-carrier">{item.carrier_name}</span>
                    {item.days_until_expiry <= 0 ? (
                      <span className="rc-days-badge rc-days-critical">EXPIRED</span>
                    ) : (
                      <span className="rc-days-badge rc-days-warning">{item.days_until_expiry}d left</span>
                    )}
                  </div>
                  <div className="rc-expiring-meta">
                    <span>{item.contract_reference}</span>
                    <span>Exp: {item.expiry_date}</span>
                  </div>
                  <div className="rc-expiring-actions">
                    <button className="rc-btn-quick-renew" onClick={() => openQuickRenew(item)}>
                      <RefreshCw size={11} style={{ display: 'inline', marginRight: '4px' }} /> Quick Renew
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      {/* ── Modal: Create Carrier Contract ────────────────────────────────── */}
      {showCreateModal && (
        <div className="rc-modal-overlay">
          <div className="rc-modal">
            <div className="rc-modal-header">
              <h3 className="rc-modal-title">Create Carrier Rate Contract</h3>
              <button className="rc-modal-close" onClick={() => setShowCreateModal(false)}>
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleCreateContract}>
              <div className="rc-modal-body">
                <div className="rc-modal-grid">
                  <div className="rm-field-group">
                    <label>Contract Reference *</label>
                    <input
                      type="text"
                      className="rm-input"
                      placeholder="e.g. MSC-2026-US-IND"
                      value={newContract.contract_reference}
                      onChange={(e) => setNewContract({ ...newContract, contract_reference: e.target.value })}
                      required
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Carrier Name *</label>
                    <input
                      type="text"
                      className="rm-input"
                      placeholder="e.g. MSC Mediterranean Shipping"
                      value={newContract.carrier_name}
                      onChange={(e) => setNewContract({ ...newContract, carrier_name: e.target.value })}
                      required
                    />
                  </div>
                </div>

                <div className="rm-field-group">
                  <label>Contract / Agreement Name *</label>
                  <input
                    type="text"
                    className="rm-input"
                    placeholder="e.g. 2026 Transpacific Annual Tier-1 Service Agreement"
                    value={newContract.contract_name}
                    onChange={(e) => setNewContract({ ...newContract, contract_name: e.target.value })}
                    required
                  />
                </div>

                <div className="rc-modal-grid">
                  <div className="rm-field-group">
                    <label>Contract Type</label>
                    <CustomSelect
                      value={newContract.contract_type}
                      onChange={(val) => setNewContract({ ...newContract, contract_type: val })}
                      options={CONTRACT_TYPES.filter(t => t.value !== 'ALL')}
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Transport Mode</label>
                    <CustomSelect
                      value={newContract.transport_mode}
                      onChange={(val) => setNewContract({ ...newContract, transport_mode: val })}
                      options={['Ocean FCL', 'Ocean LCL', 'Air Freight', 'Road Freight', 'Rail Freight']}
                    />
                  </div>
                </div>

                <div className="rc-modal-grid">
                  <div className="rm-field-group">
                    <label>Effective Date *</label>
                    <input
                      type="date"
                      className="rm-input"
                      value={newContract.effective_date}
                      onChange={(e) => setNewContract({ ...newContract, effective_date: e.target.value })}
                      required
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Expiry Date *</label>
                    <input
                      type="date"
                      className="rm-input"
                      value={newContract.expiry_date}
                      onChange={(e) => setNewContract({ ...newContract, expiry_date: e.target.value })}
                      required
                    />
                  </div>
                </div>

                <div className="rm-field-group">
                  <label>Renewal Owner / Commercial Contact</label>
                  <input
                    type="text"
                    className="rm-input"
                    placeholder="e.g. John Doe (Commercial Procurement)"
                    value={newContract.renewal_owner}
                    onChange={(e) => setNewContract({ ...newContract, renewal_owner: e.target.value })}
                  />
                </div>

                <div className="rm-field-group">
                  <label>Notes & Terms</label>
                  <textarea
                    className="rm-input"
                    rows="2"
                    placeholder="Key contract conditions, demurrage agreements, or volume tiers..."
                    value={newContract.notes}
                    onChange={(e) => setNewContract({ ...newContract, notes: e.target.value })}
                  />
                </div>
              </div>
              <div className="rc-modal-footer">
                <button type="button" className="rc-btn-secondary" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="rc-btn-primary">
                  Save Carrier Contract
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Modal: Renew Carrier Contract ────────────────────────────────── */}
      {showRenewModal && activeContractForRenew && (
        <div className="rc-modal-overlay">
          <div className="rc-modal" style={{ maxWidth: '520px' }}>
            <div className="rc-modal-header">
              <h3 className="rc-modal-title">Renew Carrier Contract</h3>
              <button className="rc-modal-close" onClick={() => setShowRenewModal(false)}>
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleRenewContract}>
              <div className="rc-modal-body">
                <div style={{ background: '#F8FAFC', padding: '12px 14px', borderRadius: '8px', fontSize: '13px' }}>
                  <div style={{ fontWeight: 700, color: '#0F172A' }}>{activeContractForRenew.carrier_name}</div>
                  <div style={{ color: '#64748B', marginTop: '2px' }}>{activeContractForRenew.contract_reference} • {activeContractForRenew.contract_name}</div>
                </div>

                <div className="rm-field-group">
                  <label>New Expiry Date *</label>
                  <input
                    type="date"
                    className="rm-input"
                    value={renewExpiryDate}
                    onChange={(e) => setRenewExpiryDate(e.target.value)}
                    required
                  />
                </div>

                <div className="rm-field-group">
                  <label>Renewal Status</label>
                  <CustomSelect
                    value={renewStatus}
                    onChange={setRenewStatus}
                    options={[
                      { value: 'RENEWED', label: 'Renewed (Contract Active)' },
                      { value: 'IN_PROGRESS', label: 'In Progress (Under Negotiation)' },
                      { value: 'NOT_RENEWING', label: 'Not Renewing (Allow to Expire)' },
                    ]}
                  />
                </div>

                <div className="rm-field-group">
                  <label>Renewal Notes</label>
                  <textarea
                    className="rm-input"
                    rows="2"
                    placeholder="Extension agreement reference, revised bunker formulas, etc."
                    value={renewNotes}
                    onChange={(e) => setRenewNotes(e.target.value)}
                  />
                </div>
              </div>
              <div className="rc-modal-footer">
                <button type="button" className="rc-btn-secondary" onClick={() => setShowRenewModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="rc-btn-primary">
                  Confirm Renewal
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Modal: Archive Contract Confirmation ── */}
      <ConfirmModal
        isOpen={archiveModalOpen}
        title="Archive Carrier Contract"
        message={`Are you sure you want to archive contract "${contractToArchive?.contract_reference}" (${contractToArchive?.carrier_name})? This will preserve historical records but mark the contract as archived.`}
        confirmText="Archive Contract"
        confirmStyle="danger"
        isLoading={archiveLoading}
        onConfirm={handleConfirmArchive}
        onCancel={() => {
          setArchiveModalOpen(false);
          setContractToArchive(null);
        }}
      />
    </div>
  );
}
