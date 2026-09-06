import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { 
  Link2, Plus, Trash2, ExternalLink, Activity, Users, 
  FileText, CheckCircle, DollarSign, Calendar, Tag, 
  Layers, ArrowUpRight, Zap, Shield, HelpCircle, X,
  TrendingUp, Truck, Building2, ChevronDown, Hash
} from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import ConfirmModal from '../Settings/ConfirmModal';
import toast from 'react-hot-toast';
import './ContractLinkedRecordsPanel.css';

export default function ContractLinkedRecordsPanel({ 
  contractId, 
  contractStatus, 
  contract, 
  onLinksChanged 
}) {
  const activeContractId = contractId || contract?.id;
  const activeContractStatus = contractStatus || contract?.status;

  const [summary, setSummary] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [unlinkingLinkId, setUnlinkingLinkId] = useState(null);
  const [isUnlinking, setIsUnlinking] = useState(false);

  const fetchLinks = useCallback(async () => {
    if (!activeContractId) {
      setIsLoading(false);
      return;
    }
    setIsLoading(true);
    try {
      const res = await contractsService.getContractRelationships(activeContractId);
      setSummary(res.data?.relationships || {
        active_links: [],
        total_count: 0,
        by_type: {}
      });
    } catch (err) {
      console.error('Failed to load contract relationship summary', err);
      toast.error('Failed to load linked records');
    } finally {
      setIsLoading(false);
    }
  }, [activeContractId]);

  useEffect(() => {
    fetchLinks();
  }, [fetchLinks]);

  const handleUnlink = (linkId) => {
    setUnlinkingLinkId(linkId);
  };

  const handleConfirmUnlink = async () => {
    if (!unlinkingLinkId) return;
    setIsUnlinking(true);
    try {
      await contractsService.removeContractLink(activeContractId, unlinkingLinkId);
      toast.success('Link removed successfully');
      setUnlinkingLinkId(null);
      fetchLinks();
      if (onLinksChanged) onLinksChanged();
    } catch (err) {
      toast.error(err.message || 'Failed to remove link');
    } finally {
      setIsUnlinking(false);
    }
  };

  if (isLoading) {
    return (
      <div className="linked-records-loading">
        <Activity size={24} className="spin-animate text-muted" />
        <p>Loading connected commercial records...</p>
      </div>
    );
  }

  const parties = summary?.parties || [];
  const rates = summary?.commercial_rates || [];
  const quotations = summary?.quotations || [];
  const spots = summary?.spot_rate_activity || [];

  const totalLinksCount = parties.length + rates.length + quotations.length + spots.length;
  const isEditable = contractStatus !== 'ARCHIVED';

  return (
    <div className="linked-records-workspace">
      
      {/* Panel Action Header */}
      <div className="linked-panel-header">
        <div className="linked-header-left">
          <h3>Commercial Relationships</h3>
          <span className="linked-count-badge">{totalLinksCount} Connected</span>
        </div>
        {isEditable && (
          <button className="btn-link-record" onClick={() => setIsAddModalOpen(true)}>
            <Plus size={15} /> Link Record
          </button>
        )}
      </div>

      {/* ── Empty State ── */}
      {totalLinksCount === 0 ? (
        <div className="linked-empty-card">
          <div className="empty-link-icon-wrap">
            <Link2 size={36} />
          </div>
          <h4>No Records Linked Yet</h4>
          <p>
            Connect this contract with rates, quotations, customers, carriers, or spot requests to build a complete commercial relationship graph.
          </p>
          {isEditable && (
            <button className="btn-empty-link-action" onClick={() => setIsAddModalOpen(true)}>
              <Plus size={16} /> Link Commercial Record
            </button>
          )}
        </div>
      ) : (
        <div className="linked-groups-container">
          
          {/* 1. Commercial Rates */}
          {rates.length > 0 && (
            <div className="linked-group-block">
              <div className="linked-group-title">
                <div className="group-title-left">
                  <span className="group-type-icon icon-rates"><Layers size={14} /></span>
                  <h5>Commercial Rates & Tariffs</h5>
                </div>
                <span className="group-counter">{rates.length}</span>
              </div>

              <div className="linked-cards-grid">
                {rates.map(link => (
                  <div className="record-card" key={link.id}>
                    <div className="record-card-top">
                      <div className="record-ref-group">
                        <span className="record-ref">{link.reference_name || `RATE-#${link.linked_entity_id}`}</span>
                        {link.is_primary ? <span className="primary-pill">Primary Rate</span> : null}
                      </div>
                      <div className="record-actions-group">
                        <Link 
                          to={`/dashboard/rate-management`} 
                          className="btn-card-action"
                          title="Open in Rate Management"
                        >
                          <ArrowUpRight size={14} />
                        </Link>
                        {isEditable && (
                          <button 
                            className="btn-card-action text-danger" 
                            onClick={() => handleUnlink(link.id)}
                            title="Unlink Record"
                          >
                            <Trash2 size={14} />
                          </button>
                        )}
                      </div>
                    </div>

                    <div className="record-details-grid">
                      <div className="record-detail-item">
                        <span className="rd-label">Source Module</span>
                        <span className="rd-value">Rate Management</span>
                      </div>
                      <div className="record-detail-item">
                        <span className="rd-label">Link Type</span>
                        <span className="rd-value">{link.link_type}</span>
                      </div>
                    </div>

                    {link.notes && <p className="record-notes-line">{link.notes}</p>}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 2. Quotations */}
          {quotations.length > 0 && (
            <div className="linked-group-block">
              <div className="linked-group-title">
                <div className="group-title-left">
                  <span className="group-type-icon icon-quote"><FileText size={14} /></span>
                  <h5>Quotations & Lane Pricing</h5>
                </div>
                <span className="group-counter">{quotations.length}</span>
              </div>

              <div className="linked-cards-grid">
                {quotations.map(link => (
                  <div className="record-card" key={link.id}>
                    <div className="record-card-top">
                      <div className="record-ref-group">
                        <span className="record-ref text-indigo">{link.reference_name || `QUOTE-#${link.linked_entity_id}`}</span>
                        {link.is_primary ? <span className="primary-pill">Primary Commercial Basis</span> : null}
                      </div>
                      <div className="record-actions-group">
                        <Link 
                          to={`/dashboard/quotations`} 
                          className="btn-card-action"
                          title="Open in Quotations"
                        >
                          <ArrowUpRight size={14} />
                        </Link>
                        {isEditable && (
                          <button 
                            className="btn-card-action text-danger" 
                            onClick={() => handleUnlink(link.id)}
                            title="Unlink Record"
                          >
                            <Trash2 size={14} />
                          </button>
                        )}
                      </div>
                    </div>

                    <div className="record-details-grid">
                      <div className="record-detail-item">
                        <span className="rd-label">Source Module</span>
                        <span className="rd-value">Quotations Desk</span>
                      </div>
                      <div className="record-detail-item">
                        <span className="rd-label">Link Type</span>
                        <span className="rd-value">{link.link_type}</span>
                      </div>
                    </div>

                    {link.notes && <p className="record-notes-line">{link.notes}</p>}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 3. Parties (Customers & Carriers) */}
          {parties.length > 0 && (
            <div className="linked-group-block">
              <div className="linked-group-title">
                <div className="group-title-left">
                  <span className="group-type-icon icon-party"><Users size={14} /></span>
                  <h5>Associated Parties & Counterparties</h5>
                </div>
                <span className="group-counter">{parties.length}</span>
              </div>

              <div className="linked-cards-grid">
                {parties.map(link => (
                  <div className="record-card" key={link.id}>
                    <div className="record-card-top">
                      <div className="record-ref-group">
                        <span className="record-ref text-emerald">{link.reference_name || `PARTY-#${link.linked_entity_id}`}</span>
                        {link.is_primary ? <span className="primary-pill">Primary Counterparty</span> : null}
                      </div>
                      <div className="record-actions-group">
                        <Link 
                          to={`/dashboard/customers`} 
                          className="btn-card-action"
                          title="Open Directory Profile"
                        >
                          <ArrowUpRight size={14} />
                        </Link>
                        {isEditable && (
                          <button 
                            className="btn-card-action text-danger" 
                            onClick={() => handleUnlink(link.id)}
                            title="Unlink Record"
                          >
                            <Trash2 size={14} />
                          </button>
                        )}
                      </div>
                    </div>

                    <div className="record-details-grid">
                      <div className="record-detail-item">
                        <span className="rd-label">Entity Category</span>
                        <span className="rd-value">{link.linked_entity_type}</span>
                      </div>
                      <div className="record-detail-item">
                        <span className="rd-label">Link Type</span>
                        <span className="rd-value">{link.link_type}</span>
                      </div>
                    </div>

                    {link.notes && <p className="record-notes-line">{link.notes}</p>}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 4. Spot Rate Activity */}
          {spots.length > 0 && (
            <div className="linked-group-block">
              <div className="linked-group-title">
                <div className="group-title-left">
                  <span className="group-type-icon icon-spot"><Zap size={14} /></span>
                  <h5>Spot Rate Requests & Responses</h5>
                </div>
                <span className="group-counter">{spots.length}</span>
              </div>

              <div className="linked-cards-grid">
                {spots.map(link => (
                  <div className="record-card" key={link.id}>
                    <div className="record-card-top">
                      <div className="record-ref-group">
                        <span className="record-ref text-amber">{link.reference_name || `SPOT-#${link.linked_entity_id}`}</span>
                        {link.is_primary ? <span className="primary-pill">Primary Spot</span> : null}
                      </div>
                      <div className="record-actions-group">
                        <Link 
                          to={`/dashboard/rate-management`} 
                          className="btn-card-action"
                          title="Open in Spot Rates"
                        >
                          <ArrowUpRight size={14} />
                        </Link>
                        {isEditable && (
                          <button 
                            className="btn-card-action text-danger" 
                            onClick={() => handleUnlink(link.id)}
                            title="Unlink Record"
                          >
                            <Trash2 size={14} />
                          </button>
                        )}
                      </div>
                    </div>

                    <div className="record-details-grid">
                      <div className="record-detail-item">
                        <span className="rd-label">Source Module</span>
                        <span className="rd-value">Spot Desk</span>
                      </div>
                      <div className="record-detail-item">
                        <span className="rd-label">Link Type</span>
                        <span className="rd-value">{link.link_type}</span>
                      </div>
                    </div>

                    {link.notes && <p className="record-notes-line">{link.notes}</p>}
                  </div>
                ))}
              </div>
            </div>
          )}

        </div>
      )}

      {/* ── Add Link Modal ── */}
      {isAddModalOpen && (
        <AddLinkModal 
          contractId={activeContractId} 
          onClose={() => setIsAddModalOpen(false)}
          onSuccess={() => {
            setIsAddModalOpen(false);
            fetchLinks();
            if (onLinksChanged) onLinksChanged();
          }}
        />
      )}

      {/* ── Unlink Confirmation Modal ── */}
      <ConfirmModal
        isOpen={Boolean(unlinkingLinkId)}
        title="Unlink Commercial Record"
        message="Are you sure you want to unlink this record? The source document will remain intact in its home module."
        confirmText="Unlink Record"
        confirmStyle="danger"
        isLoading={isUnlinking}
        onConfirm={handleConfirmUnlink}
        onCancel={() => setUnlinkingLinkId(null)}
      />
    </div>
  );
}

function AddLinkModal({ contractId, onClose, onSuccess }) {
  const [formData, setFormData] = useState({
    linked_entity_type: 'QUOTATION',
    linked_entity_id: '',
    link_type: 'QUOTATION',
    is_primary: false,
    notes: ''
  });
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const MODULE_OPTIONS = [
    {
      value: 'QUOTATION',
      label: 'Quotation Desk',
      sublabel: 'Connect customer price quotation or bid proposal',
      icon: <FileText size={16} className="text-indigo-600" />,
      colorClass: 'mod-indigo',
      badge: 'QUOTATIONS',
      defaultLinkType: 'QUOTATION'
    },
    {
      value: 'CARRIER_RATE_CONTRACT',
      label: 'Carrier Rate Contract',
      sublabel: 'Backing baseline ocean, air, or drayage rate contract',
      icon: <TrendingUp size={16} className="text-blue-600" />,
      colorClass: 'mod-blue',
      badge: 'RATES',
      defaultLinkType: 'RATE_CONTRACT'
    },
    {
      value: 'MANAGED_RATE',
      label: 'Managed Rate Tariff',
      sublabel: 'Specific lane tariff sheet or buy rate line item',
      icon: <TrendingUp size={16} className="text-emerald-600" />,
      colorClass: 'mod-emerald',
      badge: 'TARIFFS',
      defaultLinkType: 'RATE'
    },
    {
      value: 'CUSTOMER',
      label: 'Customer Account',
      sublabel: 'Directly bind commercial client profile to contract',
      icon: <Users size={16} className="text-sky-600" />,
      colorClass: 'mod-sky',
      badge: 'CLIENT',
      defaultLinkType: 'CUSTOMER'
    },
    {
      value: 'CARRIER',
      label: 'Carrier Account',
      sublabel: 'Service provider or shipping line agreement partner',
      icon: <Truck size={16} className="text-amber-600" />,
      colorClass: 'mod-amber',
      badge: 'CARRIER',
      defaultLinkType: 'CARRIER'
    },
    {
      value: 'VENDOR',
      label: 'Vendor / Partner',
      sublabel: 'Warehouse, customs broker, or 3PL logistics provider',
      icon: <Building2 size={16} className="text-purple-600" />,
      colorClass: 'mod-purple',
      badge: 'VENDOR',
      defaultLinkType: 'VENDOR'
    },
    {
      value: 'SPOT_RATE_REQUEST',
      label: 'Spot Sourcing Request',
      sublabel: 'Ad-hoc spot inquiry or dynamic quote request',
      icon: <Zap size={16} className="text-rose-600" />,
      colorClass: 'mod-rose',
      badge: 'SPOT',
      defaultLinkType: 'SPOT_RATE_REQUEST'
    }
  ];

  const selectedOption = MODULE_OPTIONS.find(o => o.value === formData.linked_entity_type) || MODULE_OPTIONS[0];

  const handleSelectModule = (opt) => {
    setFormData({
      ...formData,
      linked_entity_type: opt.value,
      link_type: opt.defaultLinkType
    });
    setIsDropdownOpen(false);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.linked_entity_id) {
      toast.error('Please enter a valid entity ID or reference number');
      return;
    }

    setIsSubmitting(true);
    try {
      await contractsService.addContractLink(contractId, {
        ...formData,
        linked_entity_id: Number(formData.linked_entity_id)
      });
      toast.success('Commercial link established successfully');
      onSuccess();
    } catch (err) {
      toast.error(err?.response?.data?.error?.message || err.message || 'Failed to add link');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content link-modal-window premium-modal" onClick={e => e.stopPropagation()}>
        
        {/* Header with gradient subtle glow */}
        <div className="modal-header">
          <div className="modal-title-group">
            <div className="modal-icon-badge">
              <Link2 size={19} />
            </div>
            <div>
              <div className="modal-header-tag">CROSS-MODULE CONNECTIVITY</div>
              <h3>Link Commercial Record</h3>
              <p className="modal-subtitle">Establish traceable bidirectional relations between this contract and operational logistics modules.</p>
            </div>
          </div>
          <button className="close-btn" onClick={onClose} title="Close window">
            <X size={18} />
          </button>
        </div>

        <form className="modal-form" onSubmit={handleSubmit}>
          <div className="modal-body">
            
            {/* Custom Interactive Module Dropdown */}
            <div className="form-group relative">
              <label className="form-label">
                <span>Target Module & Record Category</span>
                <span className="label-req">*</span>
              </label>

              {/* Custom Trigger */}
              <div 
                className={`custom-select-trigger ${isDropdownOpen ? 'open' : ''}`}
                onClick={() => setIsDropdownOpen(!isDropdownOpen)}
              >
                <div className="trigger-left">
                  <div className={`opt-icon-box ${selectedOption.colorClass}`}>
                    {selectedOption.icon}
                  </div>
                  <div className="trigger-info">
                    <span className="trigger-main-title">{selectedOption.label}</span>
                    <span className="trigger-sub-title">{selectedOption.sublabel}</span>
                  </div>
                </div>
                <div className="trigger-right">
                  <span className="trigger-badge">{selectedOption.badge}</span>
                  <ChevronDown size={16} className={`trigger-arrow ${isDropdownOpen ? 'rotate' : ''}`} />
                </div>
              </div>

              {/* Dropdown Menu Popover */}
              {isDropdownOpen && (
                <div className="custom-dropdown-menu">
                  <div className="dropdown-menu-header">Select source module</div>
                  <div className="dropdown-menu-list">
                    {MODULE_OPTIONS.map((opt) => {
                      const isSelected = opt.value === formData.linked_entity_type;
                      return (
                        <div
                          key={opt.value}
                          className={`dropdown-option-item ${isSelected ? 'selected' : ''}`}
                          onClick={() => handleSelectModule(opt)}
                        >
                          <div className="opt-item-left">
                            <div className={`opt-icon-box ${opt.colorClass}`}>
                              {opt.icon}
                            </div>
                            <div className="opt-text-group">
                              <span className="opt-name">{opt.label}</span>
                              <span className="opt-desc">{opt.sublabel}</span>
                            </div>
                          </div>
                          <div className="opt-item-right">
                            <span className="opt-badge">{opt.badge}</span>
                            {isSelected && <CheckCircle size={15} className="text-blue-600" />}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>

            {/* Record ID Input */}
            <div className="form-group">
              <label className="form-label">
                <span>Record ID or Number</span>
                <span className="label-req">*</span>
              </label>
              <div className="input-with-icon">
                <Hash size={16} className="input-leading-icon" />
                <input
                  type="number"
                  min="1"
                  className="modal-input has-icon"
                  placeholder="e.g. 1, 25, 1042"
                  value={formData.linked_entity_id}
                  onChange={e => setFormData({ ...formData, linked_entity_id: e.target.value })}
                  required
                />
              </div>
              <span className="field-hint">Enter the target database ID from the selected module.</span>
            </div>

            {/* Primary Backing Record Toggle */}
            <div 
              className={`primary-toggle-card ${formData.is_primary ? 'active-primary' : ''}`}
              onClick={() => setFormData({ ...formData, is_primary: !formData.is_primary })}
            >
              <div className="toggle-left">
                <div className={`custom-switch ${formData.is_primary ? 'on' : 'off'}`}>
                  <div className="switch-handle"></div>
                </div>
                <div className="toggle-info">
                  <div className="toggle-title-row">
                    <span className="toggle-title">Primary Commercial Backing Agreement</span>
                    {formData.is_primary && <span className="primary-active-tag">PRIMARY GOVERNING</span>}
                  </div>
                  <span className="toggle-sub">Designates this record as the governing rate & SLA provider for automatic quotation lookups.</span>
                </div>
              </div>
            </div>

            {/* Notes Field */}
            <div className="form-group">
              <label className="form-label">
                <span>Relationship Notes / Commercial Context</span>
                <span className="label-opt">(Optional)</span>
              </label>
              <textarea
                className="modal-textarea"
                placeholder="e.g. Tier-1 Transpacific contract governing spot overflow & fuel surcharges..."
                rows="2"
                value={formData.notes}
                onChange={e => setFormData({ ...formData, notes: e.target.value })}
              />
            </div>

          </div>

          {/* Footer */}
          <div className="modal-footer">
            <button type="button" className="btn-modal-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-modal-primary" disabled={isSubmitting}>
              <Link2 size={15} />
              <span>{isSubmitting ? 'Establishing Link...' : 'Establish Commercial Link'}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
