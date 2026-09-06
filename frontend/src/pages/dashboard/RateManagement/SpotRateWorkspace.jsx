import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  Send, Plus, RefreshCw, Layers, CheckCircle2, Clock, AlertTriangle,
  ArrowRight, Search, Filter, Building2, Calendar, DollarSign,
  TrendingDown, Zap, Award, Check, Eye, Trash2, X, ChevronRight,
  ShieldCheck, ArrowUpRight, Scale, Tag, ChevronDown
} from 'lucide-react';
import { rateService } from '../../../services/rateService';
import ConfirmModal from '../Settings/ConfirmModal';
import toast from 'react-hot-toast';
import './SpotRateWorkspace.css';

// ── Custom Select Component ──────────────────────────────────────────────────
function CustomSelect({
  value,
  onChange,
  options = [],
  placeholder = 'Select...',
  disabled = false,
  className = '',
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [dropUp, setDropUp] = useState(false);
  const dropdownRef = useRef(null);

  const selectedOption = options.find((opt) => opt.value === value);

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleToggle = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (disabled) return;
    if (!isOpen && dropdownRef.current) {
      const rect = dropdownRef.current.getBoundingClientRect();
      const spaceBelow = window.innerHeight - rect.bottom;
      setDropUp(spaceBelow < 200 && rect.top > 200);
    }
    setIsOpen((prev) => !prev);
  };

  return (
    <div
      ref={dropdownRef}
      className={`rm-custom-select-container ${isOpen ? 'is-open' : ''} ${disabled ? 'is-disabled' : ''} ${className}`}
      style={{ width: '100%', position: 'relative', zIndex: isOpen ? 1050 : 1 }}
    >
      <button
        type="button"
        className={`rm-custom-select-trigger ${isOpen ? 'active' : ''}`}
        onClick={handleToggle}
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        <div className="rm-custom-select-trigger-content">
          <span className={`rm-custom-select-trigger-label ${!selectedOption ? 'is-placeholder' : ''}`}>
            {selectedOption ? selectedOption.label : placeholder}
          </span>
        </div>
        <ChevronDown
          size={14}
          className={`rm-custom-select-chevron ${isOpen ? 'is-rotated' : ''}`}
        />
      </button>

      {isOpen && (
        <div 
          className={`rm-custom-select-dropdown ${dropUp ? 'is-dropup' : ''}`} 
          role="listbox"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="rm-custom-select-options">
            {options.map((opt) => (
              <div
                key={opt.value}
                className={`rm-custom-select-option ${opt.value === value ? 'is-selected' : ''}`}
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  onChange(opt.value);
                  setIsOpen(false);
                }}
                role="option"
                aria-selected={opt.value === value}
              >
                <span>{opt.label}</span>
                {opt.value === value && <Check size={14} className="rm-custom-select-option-check" />}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default function SpotRateWorkspace() {
  const [requests, setRequests] = useState([]);
  const [summary, setSummary] = useState({
    total_requests: 0,
    open_requests: 0,
    awaiting_responses: 0,
    fully_responded: 0,
    selected_requests: 0,
    expiring_soon: 0,
    recent_spot_requests: [],
  });
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [modeFilter, setModeFilter] = useState('ALL');

  // Active Request & Modals
  const [selectedRequest, setSelectedRequest] = useState(null);
  const [comparison, setComparison] = useState(null);
  const [responses, setResponses] = useState([]);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showAddResponseModal, setShowAddResponseModal] = useState(false);
  const [showDetailDrawer, setShowDetailDrawer] = useState(false);
  const [drawerTab, setDrawerTab] = useState('comparison'); // 'comparison' | 'responses' | 'overview'

  // Confirm Modal state
  const [confirmModal, setConfirmModal] = useState({
    isOpen: false,
    title: '',
    message: '',
    confirmText: '',
    confirmStyle: 'danger',
    action: null,
  });

  // Create Request Form
  const [newRequest, setNewRequest] = useState({
    request_reference: '',
    customer_name: '',
    origin_port: '',
    destination_port: '',
    transport_mode: 'Ocean FCL',
    service_type: 'FCL',
    equipment_type: '40GP',
    commodity: 'General Cargo',
    container_quantity: 1,
    ready_date: new Date().toISOString().split('T')[0],
    required_by_date: new Date(Date.now() + 7 * 86400000).toISOString().split('T')[0],
    target_currency: 'USD',
    notes: '',
  });

  // Add Carrier Response Form
  const [newResponse, setNewResponse] = useState({
    carrier_name: '',
    carrier_code: '',
    supplier_name: '',
    currency: 'USD',
    base_amount: '',
    transit_days: '',
    free_days_origin: 7,
    free_days_destination: 14,
    valid_from: new Date().toISOString().split('T')[0],
    valid_until: new Date(Date.now() + 14 * 86400000).toISOString().split('T')[0],
    routing_notes: 'Direct service',
    response_notes: '',
    charges: [
      { charge_category: 'FREIGHT', charge_name: 'Base Ocean Freight', calculation_basis: 'FLAT', quantity: 1, unit_price: 0, currency: 'USD', display_order: 0 }
    ],
  });

  const fetchWorkspaceData = useCallback(async () => {
    setLoading(true);
    try {
      const [reqRes, sumRes] = await Promise.all([
        rateService.getSpotRateRequests({
          search,
          status: statusFilter,
          transport_mode: modeFilter,
          limit: 30,
        }),
        rateService.getSpotRateSummary(),
      ]);

      const reqData = reqRes?.data || reqRes || {};
      const sumData = sumRes?.data || sumRes || {};

      setRequests(Array.isArray(reqData.requests) ? reqData.requests : (Array.isArray(reqData) ? reqData : []));
      setSummary(sumData.summary || sumData || {
        total_requests: 0,
        open_requests: 0,
        awaiting_responses: 0,
        fully_responded: 0,
        selected_requests: 0,
        expiring_soon: 0,
      });
    } catch (err) {
      toast.error('Failed to load spot rate workspace data');
    } finally {
      setLoading(false);
    }
  }, [search, statusFilter, modeFilter]);

  useEffect(() => {
    fetchWorkspaceData();
  }, [fetchWorkspaceData]);

  const loadRequestDetail = async (reqId) => {
    try {
      const [reqRes, respRes, compRes] = await Promise.all([
        rateService.getSpotRateRequest(reqId),
        rateService.getSpotRateResponses(reqId),
        rateService.compareSpotRates(reqId),
      ]);

      setSelectedRequest(reqRes?.data || reqRes);
      const respList = respRes?.data?.responses || respRes?.responses || respRes || [];
      setResponses(Array.isArray(respList) ? respList : []);
      setComparison(compRes?.data || compRes);
      setShowDetailDrawer(true);
    } catch (err) {
      toast.error('Failed to load spot request details');
    }
  };

  const handleCreateRequest = async (e) => {
    e.preventDefault();
    if (!newRequest.origin_port || !newRequest.destination_port) {
      toast.error('Please enter origin and destination ports');
      return;
    }
    try {
      await rateService.createSpotRateRequest({
        ...newRequest,
        container_quantity: parseInt(newRequest.container_quantity, 10) || 1,
      });
      toast.success('Spot rate request created and broadcasted');
      setShowCreateModal(false);
      fetchWorkspaceData();
    } catch (err) {
      toast.error(err?.message || 'Failed to create spot request');
    }
  };

  const handleAddResponse = async (e) => {
    e.preventDefault();
    if (!selectedRequest?.id) return;
    if (!newResponse.carrier_name || !newResponse.base_amount) {
      toast.error('Please enter carrier name and base freight rate');
      return;
    }

    try {
      const base = parseFloat(newResponse.base_amount) || 0;
      await rateService.createSpotRateResponse(selectedRequest.id, {
        ...newResponse,
        spot_rate_request_id: selectedRequest.id,
        base_amount: base,
        transit_days: newResponse.transit_days ? parseInt(newResponse.transit_days, 10) : undefined,
        free_days_origin: parseInt(newResponse.free_days_origin, 10) || 0,
        free_days_destination: parseInt(newResponse.free_days_destination, 10) || 0,
        charges: newResponse.charges.map((c, i) => ({
          ...c,
          quantity: parseFloat(c.quantity) || 1,
          unit_price: i === 0 ? base : (parseFloat(c.unit_price) || 0),
          display_order: i,
        })),
      });

      toast.success(`Spot response from ${newResponse.carrier_name} recorded`);
      setShowAddResponseModal(false);
      loadRequestDetail(selectedRequest.id);
      fetchWorkspaceData();
    } catch (err) {
      toast.error(err?.message || 'Failed to add carrier response');
    }
  };

  const handleSelectPreferred = (response) => {
    setConfirmModal({
      isOpen: true,
      title: 'Select Preferred Carrier Rate',
      message: `Are you sure you want to select ${response.carrier_name} (${response.currency} ${Number(response.total_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}) as the preferred commercial rate for this lane? All other carrier quotes will remain intact and historically preserved.`,
      confirmText: 'Select Preferred Rate',
      confirmStyle: 'primary',
      action: async () => {
        try {
          await rateService.selectPreferredSpotRate(selectedRequest.id, response.id, {
            selection_notes: `Preferred rate selected via rate comparison matrix on ${new Date().toLocaleDateString()}`,
          });
          toast.success(`${response.carrier_name} selected as preferred carrier rate!`);
          setConfirmModal((prev) => ({ ...prev, isOpen: false }));
          loadRequestDetail(selectedRequest.id);
          fetchWorkspaceData();
        } catch (err) {
          toast.error(err?.message || 'Failed to select preferred rate');
        }
      },
    });
  };

  const handleCancelRequest = (req) => {
    setConfirmModal({
      isOpen: true,
      title: 'Cancel Spot Rate Request',
      message: `Are you sure you want to cancel spot request ${req.request_reference}?`,
      confirmText: 'Cancel Request',
      confirmStyle: 'danger',
      action: async () => {
        try {
          await rateService.cancelSpotRateRequest(req.id);
          toast.success(`Request ${req.request_reference} cancelled`);
          setConfirmModal((prev) => ({ ...prev, isOpen: false }));
          fetchWorkspaceData();
          if (selectedRequest?.id === req.id) {
            setShowDetailDrawer(false);
          }
        } catch (err) {
          toast.error('Failed to cancel request');
        }
      },
    });
  };

  return (
    <div className="sr-workspace">
      {/* ── KPI Strip ── */}
      <div className="sr-kpi-grid">
        <div className="sr-kpi-card">
          <div className="sr-kpi-header">
            <span className="sr-kpi-title">Total Requests</span>
            <div className="sr-kpi-icon sr-icon-blue"><Layers size={16} /></div>
          </div>
          <div className="sr-kpi-val">{summary.total_requests}</div>
          <div className="sr-kpi-sub">Commercial spot lanes</div>
        </div>

        <div className="sr-kpi-card">
          <div className="sr-kpi-header">
            <span className="sr-kpi-title">Awaiting Responses</span>
            <div className="sr-kpi-icon sr-icon-amber"><Clock size={16} /></div>
          </div>
          <div className="sr-kpi-val">{summary.awaiting_responses}</div>
          <div className="sr-kpi-sub">Carriers contacted</div>
        </div>

        <div className="sr-kpi-card">
          <div className="sr-kpi-header">
            <span className="sr-kpi-title">Fully Responded</span>
            <div className="sr-kpi-icon sr-icon-emerald"><CheckCircle2 size={16} /></div>
          </div>
          <div className="sr-kpi-val">{summary.fully_responded}</div>
          <div className="sr-kpi-sub">Ready for comparison</div>
        </div>

        <div className="sr-kpi-card">
          <div className="sr-kpi-header">
            <span className="sr-kpi-title">Selected / Booked</span>
            <div className="sr-kpi-icon sr-icon-purple"><Award size={16} /></div>
          </div>
          <div className="sr-kpi-val">{summary.selected_requests}</div>
          <div className="sr-kpi-sub">Preferred carrier chosen</div>
        </div>

        <div className="sr-kpi-card">
          <div className="sr-kpi-header">
            <span className="sr-kpi-title">Expiring Soon</span>
            <div className="sr-kpi-icon sr-icon-rose"><AlertTriangle size={16} /></div>
          </div>
          <div className="sr-kpi-val">{summary.expiring_soon}</div>
          <div className="sr-kpi-sub">Within next 3 days</div>
        </div>
      </div>

      {/* ── Search & Action Bar ── */}
      <div className="sr-filter-card">
        <div className="sr-filter-row">
          <div className="sr-search-box">
            <Search size={16} style={{ color: '#94A3B8' }} />
            <input
              type="text"
              placeholder="Search request ref, customer, origin, destination..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          <div className="sr-filter-group">
            <div className="sr-select-wrap">
              <CustomSelect
                value={statusFilter}
                onChange={setStatusFilter}
                options={[
                  { value: 'ALL', label: 'All Statuses' },
                  { value: 'SENT', label: 'Sent (Awaiting)' },
                  { value: 'PARTIALLY_RESPONDED', label: 'Partially Responded' },
                  { value: 'RESPONDED', label: 'Fully Responded' },
                  { value: 'SELECTED', label: 'Selected' },
                  { value: 'EXPIRED', label: 'Expired' },
                ]}
              />
            </div>

            <div className="sr-select-wrap">
              <CustomSelect
                value={modeFilter}
                onChange={setModeFilter}
                options={[
                  { value: 'ALL', label: 'All Modes' },
                  { value: 'Ocean FCL', label: 'Ocean FCL' },
                  { value: 'Ocean LCL', label: 'Ocean LCL' },
                  { value: 'Air Freight', label: 'Air Freight' },
                ]}
              />
            </div>

            <button
              type="button"
              className="sr-btn-primary"
              onClick={() => setShowCreateModal(true)}
            >
              <Plus size={15} />
              <span>New Spot Request</span>
            </button>
          </div>
        </div>
      </div>

      {/* ── Main Layout Grid ── */}
      <div className="sr-main-grid">
        {/* Left Column: Spot Requests Table */}
        <div className="sr-table-card">
          <div className="sr-table-header">
            <div className="sr-table-title">Spot Rate Requests ({requests.length})</div>
            <button className="sr-btn-secondary" style={{ padding: '6px 12px' }} onClick={fetchWorkspaceData}>
              <RefreshCw size={13} /> Refresh
            </button>
          </div>

          <div className="sr-table-wrap">
            <table className="sr-table">
              <thead>
                <tr>
                  <th>Request Ref</th>
                  <th>Customer</th>
                  <th>Origin → Destination</th>
                  <th>Mode / EQ</th>
                  <th>Required By</th>
                  <th>Responses</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={8} style={{ textAlign: 'center', padding: '40px', color: '#64748B' }}>
                      <RefreshCw className="animate-spin" size={20} style={{ margin: '0 auto 8px' }} />
                      Loading spot rate sourcing requests...
                    </td>
                  </tr>
                ) : requests.length === 0 ? (
                  <tr>
                    <td colSpan={8} style={{ textAlign: 'center', padding: '40px', color: '#64748B' }}>
                      No spot rate requests found. Click "New Spot Request" to source carrier rates for a specific shipment lane.
                    </td>
                  </tr>
                ) : (
                  requests.map((r) => (
                    <tr key={r.id} onClick={() => loadRequestDetail(r.id)}>
                      <td>
                        <span style={{ fontFamily: 'monospace', fontWeight: 750, color: '#2563EB' }}>
                          {r.request_reference}
                        </span>
                      </td>
                      <td>
                        <strong>{r.customer_name || 'General Spot Query'}</strong>
                      </td>
                      <td>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontWeight: 650 }}>
                          <span>{r.origin_port}</span>
                          <ArrowRight size={12} style={{ color: '#94A3B8' }} />
                          <span>{r.destination_port}</span>
                        </div>
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: '4px', alignItems: 'center' }}>
                          <span>{r.transport_mode}</span>
                          <span style={{ background: '#F1F5F9', padding: '2px 6px', borderRadius: '4px', fontSize: '11px', fontWeight: 700 }}>
                            {r.equipment_type || '40GP'}
                          </span>
                        </div>
                      </td>
                      <td>
                        <div style={{ fontSize: '12.5px' }}>
                          <div>{r.required_by_date}</div>
                          {r.days_until_required !== undefined && (
                            <span style={{ fontSize: '11px', color: r.days_until_required <= 2 ? '#DC2626' : '#64748B', fontWeight: 600 }}>
                              {r.days_until_required <= 0 ? 'Expired' : `${r.days_until_required}d left`}
                            </span>
                          )}
                        </div>
                      </td>
                      <td>
                        <span style={{ fontWeight: 800, color: r.response_count > 0 ? '#059669' : '#64748B' }}>
                          {r.response_count} quotes
                        </span>
                        {r.has_preferred && (
                          <span style={{ marginLeft: '6px', color: '#15803D' }} title="Preferred carrier selected">★</span>
                        )}
                      </td>
                      <td>
                        <span className={`sr-badge sr-badge-${r.status.toLowerCase().replace(/_/g, '-')}`}>
                          {r.status.replace(/_/g, ' ')}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }} onClick={(e) => e.stopPropagation()}>
                        <div style={{ display: 'inline-flex', gap: '6px' }}>
                          <button
                            type="button"
                            className="rm-btn-view-details"
                            onClick={() => loadRequestDetail(r.id)}
                          >
                            <Scale size={13} />
                            <span>Compare</span>
                          </button>
                          <button
                            type="button"
                            className="rm-action-icon-btn rm-action-icon-btn--delete"
                            title="Cancel Request"
                            onClick={() => handleCancelRequest(r)}
                          >
                            <Trash2 size={13} />
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

        {/* Right Column: Sourcing Intelligence & Tips */}
        <div className="sr-side-panel">
          <div className="sr-side-card">
            <div className="sr-side-card-title">
              <span>Spot Sourcing Rules</span>
              <ShieldCheck size={16} style={{ color: '#2563EB' }} />
            </div>
            <div style={{ fontSize: '12.5px', color: '#475569', display: 'flex', flexDirection: 'column', gap: '10px' }}>
              <div>
                <strong>Immutable Comparison:</strong> Carrier rate quotes are logged independently. Selecting a preferred option does not overwrite competing quotes.
              </div>
              <div>
                <strong>Multi-Currency Safe:</strong> Cross-currency totals are highlighted with conversion safety notices.
              </div>
              <div>
                <strong>Downstream Linked:</strong> Preferred quotes can be attached directly to customer quotations and operational handovers.
              </div>
            </div>
          </div>

          <div className="sr-side-card" style={{ background: '#EFF6FF', borderColor: '#BFDBFE' }}>
            <div className="sr-side-card-title" style={{ color: '#1E40AF' }}>
              <span>Need Spot Rates from MSC/Maersk?</span>
            </div>
            <p style={{ fontSize: '12.5px', color: '#1E3A8A', margin: 0, lineHeight: 1.4 }}>
              Click "+ New Spot Request" to record requests across multiple ocean & air carriers. Side-by-side pricing updates in real-time.
            </p>
          </div>
        </div>
      </div>

      {/* ── Slide-over Detail & Comparison Drawer (70% Right Full Panel) ── */}
      {showDetailDrawer && selectedRequest && (
        <div className="rm-drawer-overlay" onClick={() => setShowDetailDrawer(false)}>
          <div className="rm-drawer srd-drawer" onClick={(e) => e.stopPropagation()}>

            {/* ── Drawer Header ── */}
            <div className="srd-header">
              <div className="srd-header-meta">
                <div className="srd-header-top">
                  <div className="srd-ref-badge">{selectedRequest.request_reference}</div>
                  <span className={`sr-badge sr-badge-${selectedRequest.status.toLowerCase().replace(/_/g, '-')}`}>
                    {selectedRequest.status.replace(/_/g, ' ')}
                  </span>
                  {responses.length > 0 && (
                    <span className="srd-response-count-badge">
                      {responses.length} Carrier {responses.length === 1 ? 'Quote' : 'Quotes'}
                    </span>
                  )}
                </div>
                <div className="srd-header-route">
                  <span className="srd-route-port">{selectedRequest.origin_port}</span>
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" style={{ opacity: 0.4, flexShrink: 0 }}>
                    <path d="M3 8h10M9 4l4 4-4 4" stroke="#0F172A" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  <span className="srd-route-port">{selectedRequest.destination_port}</span>
                  <span className="srd-route-chip">{selectedRequest.transport_mode}</span>
                  <span className="srd-route-chip srd-route-chip-eq">{selectedRequest.equipment_type || '40GP'}</span>
                  {selectedRequest.customer_name && (
                    <span className="srd-route-customer">• {selectedRequest.customer_name}</span>
                  )}
                </div>
              </div>
              <div className="srd-header-actions">
                <button
                  type="button"
                  className="sr-btn-primary"
                  style={{ padding: '7px 14px', fontSize: '13px' }}
                  onClick={() => setShowAddResponseModal(true)}
                >
                  <Plus size={14} /> Add Carrier Quote
                </button>
                <button className="rm-drawer-close" onClick={() => setShowDetailDrawer(false)}>
                  <X size={18} />
                </button>
              </div>
            </div>

            {/* ── Drawer Tabs ── */}
            <div className="srd-tabs">
              <button className={`srd-tab ${drawerTab === 'comparison' ? 'is-active' : ''}`}
                onClick={() => setDrawerTab('comparison')}>
                <Scale size={14} />
                Side-by-Side Comparison
                <span className="srd-tab-count">{responses.length}</span>
              </button>
              <button className={`srd-tab ${drawerTab === 'responses' ? 'is-active' : ''}`}
                onClick={() => setDrawerTab('responses')}>
                <Building2 size={14} />
                Carrier Quotes & Charges
              </button>
              <button className={`srd-tab ${drawerTab === 'overview' ? 'is-active' : ''}`}
                onClick={() => setDrawerTab('overview')}>
                <Tag size={14} />
                Request Overview
              </button>
            </div>

            {/* ── Tab Body ── */}
            <div className="srd-body">

              {/* ═══════════════════════════════════════════════════════════
                  TAB 1 — Side-by-Side Comparison
              ═══════════════════════════════════════════════════════════ */}
              {drawerTab === 'comparison' && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>

                  {/* Banner */}
                  {comparison?.is_multi_currency ? (
                    <div className="sr-matrix-banner is-warning">
                      <AlertTriangle size={16} />
                      <div>
                        <strong>Multi-Currency Warning:</strong> Carrier quotes contain multiple currencies ({comparison.multi_currency_currencies?.join(', ')}). Price rankings are filtered against target currency <strong>{comparison.target_currency}</strong>.
                      </div>
                    </div>
                  ) : comparison?.comparison_summary_note ? (
                    <div className="sr-matrix-banner">
                      <Zap size={16} />
                      <div>{comparison.comparison_summary_note}</div>
                    </div>
                  ) : null}

                  {/* Recommendation Badges — 4-up grid */}
                  <div className="srd-rec-grid">
                    <div className="srd-rec-card srd-rec-cheapest">
                      <div className="srd-rec-label">🏷 Cheapest Rate</div>
                      <div className="srd-rec-carrier">{comparison?.cheapest_carrier_name || '—'}</div>
                      <div className="srd-rec-value" style={{ color: '#15803D' }}>
                        {comparison?.cheapest_amount ? `${comparison.target_currency} ${Number(comparison.cheapest_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}` : '—'}
                      </div>
                    </div>
                    <div className="srd-rec-card srd-rec-fastest">
                      <div className="srd-rec-label">⚡ Fastest Transit</div>
                      <div className="srd-rec-carrier">{comparison?.fastest_carrier_name || '—'}</div>
                      <div className="srd-rec-value" style={{ color: '#1D4ED8' }}>
                        {comparison?.fastest_transit_days ? `${comparison.fastest_transit_days} Days` : '—'}
                      </div>
                    </div>
                    <div className="srd-rec-card srd-rec-value-card">
                      <div className="srd-rec-label">⭐ Best Value</div>
                      <div className="srd-rec-carrier">{comparison?.best_value_carrier_name || '—'}</div>
                      <div className="srd-rec-value" style={{ color: '#7E22CE', fontSize: 11 }}>Balanced price & transit</div>
                    </div>
                    <div className="srd-rec-card srd-rec-preferred">
                      <div className="srd-rec-label">✓ Preferred Choice</div>
                      <div className="srd-rec-carrier">{comparison?.preferred_carrier_name || 'None Selected'}</div>
                      <div className="srd-rec-value" style={{ color: '#92400E', fontSize: 11 }}>
                        {comparison?.preferred_carrier_name ? 'Locked for Quotation/Booking' : 'Click "Select" below'}
                      </div>
                    </div>
                  </div>

                  {/* Quote Cards — replaces the cramped 7-column table */}
                  {responses.length === 0 ? (
                    <div className="srd-empty">
                      <Building2 size={32} style={{ color: '#CBD5E1' }} />
                      <div style={{ fontWeight: 700, color: '#475569', marginTop: 8 }}>No carrier quotes yet</div>
                      <div style={{ fontSize: 13, color: '#94A3B8', marginTop: 4 }}>Click "Add Carrier Quote" to log carrier rates for comparison.</div>
                    </div>
                  ) : (
                    <div className="srd-quote-cards">
                      {responses.map((resp, idx) => {
                        const compItem = comparison?.responses?.find((c) => c.response_id === resp.id);
                        const isPreferred = resp.is_preferred;
                        return (
                          <div key={resp.id} className={`srd-quote-card ${isPreferred ? 'srd-quote-card--preferred' : ''}`}>
                            {/* Card header row */}
                            <div className="srd-qc-header">
                              <div className="srd-qc-header-left">
                                <div className="srd-qc-rank">#{idx + 1}</div>
                                <div>
                                  <div className="srd-qc-carrier">
                                    {resp.carrier_name}
                                    {isPreferred && <span className="srd-preferred-chip">★ SELECTED</span>}
                                  </div>
                                  {resp.supplier_name && (
                                    <div className="srd-qc-supplier">via {resp.supplier_name}</div>
                                  )}
                                </div>
                              </div>
                              {/* Recommendation tags */}
                              <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap', justifyContent: 'flex-end' }}>
                                {compItem?.recommendation_tags?.map((t) => (
                                  <span key={t} className={`srd-rec-tag srd-rec-tag-${t.toLowerCase()}`}>{t}</span>
                                ))}
                              </div>
                            </div>

                            {/* Metrics row — 5 pillars */}
                            <div className="srd-qc-metrics">
                              <div className="srd-qc-metric srd-qc-metric--rate">
                                <div className="srd-qc-metric-label">Total Rate</div>
                                <div className="srd-qc-metric-val">
                                  {resp.currency} {Number(resp.total_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                                </div>
                                <div className="srd-qc-metric-sub">
                                  Base: {resp.currency} {Number(resp.base_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                                  {resp.charges?.length > 0 ? ` + ${resp.charges.length} charges` : ''}
                                </div>
                              </div>

                              <div className="srd-qc-metric">
                                <div className="srd-qc-metric-label">Transit Time</div>
                                <div className="srd-qc-metric-val srd-qc-metric-val--transit">
                                  {resp.transit_days ? `${resp.transit_days}d` : '—'}
                                </div>
                                <div className="srd-qc-metric-sub">sailing days</div>
                              </div>

                              <div className="srd-qc-metric">
                                <div className="srd-qc-metric-label">Free Days — Dest</div>
                                <div className="srd-qc-metric-val srd-qc-metric-val--freedays">
                                  {resp.free_days_destination ?? 14}d
                                </div>
                                <div className="srd-qc-metric-sub">at destination port</div>
                              </div>

                              <div className="srd-qc-metric">
                                <div className="srd-qc-metric-label">Free Days — Origin</div>
                                <div className="srd-qc-metric-val srd-qc-metric-val--freedays">
                                  {resp.free_days_origin ?? 7}d
                                </div>
                                <div className="srd-qc-metric-sub">at origin port</div>
                              </div>

                              <div className="srd-qc-metric">
                                <div className="srd-qc-metric-label">Validity</div>
                                <div className="srd-qc-metric-val srd-qc-metric-val--validity">
                                  {resp.valid_until || '—'}
                                </div>
                                <div className="srd-qc-metric-sub">
                                  {resp.valid_from ? `from ${resp.valid_from}` : 'validity date'}
                                </div>
                              </div>
                            </div>

                            {/* Routing note + Action */}
                            <div className="srd-qc-footer">
                              {resp.routing_notes && (
                                <span className="srd-qc-routing">🛳 {resp.routing_notes}</span>
                              )}
                              <div style={{ marginLeft: 'auto' }}>
                                {isPreferred ? (
                                  <span className="srd-selected-tick">
                                    <Check size={13} /> Preferred Carrier Selected
                                  </span>
                                ) : (
                                  <button
                                    type="button"
                                    className="sr-btn-primary srd-select-btn"
                                    onClick={() => handleSelectPreferred(resp)}
                                  >
                                    Select as Preferred
                                  </button>
                                )}
                              </div>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              )}

              {/* ═══════════════════════════════════════════════════════════
                  TAB 2 — Carrier Quotes & Itemized Charges
              ═══════════════════════════════════════════════════════════ */}
              {drawerTab === 'responses' && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                  {responses.length === 0 ? (
                    <div className="srd-empty">
                      <DollarSign size={32} style={{ color: '#CBD5E1' }} />
                      <div style={{ fontWeight: 700, color: '#475569', marginTop: 8 }}>No carrier quotes recorded</div>
                      <div style={{ fontSize: 13, color: '#94A3B8', marginTop: 4 }}>Add carrier quotes to see itemized charge breakdowns here.</div>
                    </div>
                  ) : responses.map((resp) => (
                    <div key={resp.id} className={`srd-cq-card ${resp.is_preferred ? 'srd-cq-card--preferred' : ''}`}>
                      {/* Carrier name + rate header */}
                      <div className="srd-cq-header">
                        <div className="srd-cq-header-left">
                          <div className="srd-cq-carrier-name">
                            {resp.carrier_name}
                            {resp.is_preferred && <span className="srd-preferred-chip">★ PREFERRED</span>}
                          </div>
                          <div className="srd-cq-meta">
                            Validity: <strong>{resp.valid_from}</strong> to <strong>{resp.valid_until}</strong>
                            &nbsp;•&nbsp; Routing: {resp.routing_notes || 'Direct service'}
                          </div>
                        </div>
                        <div className="srd-cq-total">
                          <div className="srd-cq-total-amount">
                            {resp.currency} {Number(resp.total_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                          </div>
                          <div className="srd-cq-total-label">Total Commercial Rate</div>
                        </div>
                      </div>

                      {/* Info pills */}
                      <div className="srd-cq-pills">
                        <div className="srd-pill">
                          <span className="srd-pill-label">Transit</span>
                          <span className="srd-pill-val">{resp.transit_days ? `${resp.transit_days} days` : '—'}</span>
                        </div>
                        <div className="srd-pill">
                          <span className="srd-pill-label">Free Days (Dest)</span>
                          <span className="srd-pill-val">{resp.free_days_destination ?? 14}d</span>
                        </div>
                        <div className="srd-pill">
                          <span className="srd-pill-label">Free Days (Origin)</span>
                          <span className="srd-pill-val">{resp.free_days_origin ?? 7}d</span>
                        </div>
                        <div className="srd-pill">
                          <span className="srd-pill-label">Base Rate</span>
                          <span className="srd-pill-val">{resp.currency} {Number(resp.base_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}</span>
                        </div>
                      </div>

                      {/* Itemized charges */}
                      <div className="srd-cq-charges">
                        <div className="srd-cq-charges-title">
                          Itemized Charges
                          <span className="srd-cq-charge-count">{resp.charges?.length || 0} items</span>
                        </div>
                        {resp.charges && resp.charges.length > 0 ? (
                          <div className="srd-cq-charge-list">
                            {resp.charges.map((c, i) => (
                              <div key={c.id || i} className="srd-cq-charge-row">
                                <div>
                                  <span className="srd-cq-charge-name">{c.charge_name}</span>
                                  <span className="srd-cq-charge-cat">{c.charge_category}</span>
                                </div>
                                <div className="srd-cq-charge-calc">
                                  {c.quantity > 1 && <span className="srd-cq-charge-qty">{c.quantity} × {c.currency} {Number(c.unit_price).toLocaleString(undefined, { minimumFractionDigits: 2 })}</span>}
                                </div>
                                <div className="srd-cq-charge-total">
                                  {c.currency} {Number(c.quantity * c.unit_price).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                                </div>
                              </div>
                            ))}
                            <div className="srd-cq-charge-row srd-cq-charge-total-row">
                              <div style={{ fontWeight: 700 }}>Total</div>
                              <div />
                              <div style={{ fontWeight: 800, color: '#0F172A' }}>
                                {resp.currency} {Number(resp.total_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                              </div>
                            </div>
                          </div>
                        ) : (
                          <div className="srd-cq-no-charges">
                            Base Rate only ({resp.currency} {Number(resp.base_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}) — no additional surcharges recorded
                          </div>
                        )}
                      </div>

                      {/* Notes */}
                      {resp.response_notes && (
                        <div className="srd-cq-notes">
                          <span style={{ fontWeight: 600 }}>Carrier Notes:</span> {resp.response_notes}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}

              {/* ═══════════════════════════════════════════════════════════
                  TAB 3 — Request Overview
              ═══════════════════════════════════════════════════════════ */}
              {drawerTab === 'overview' && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>

                  {/* Main details grid */}
                  <div className="srd-ov-card">
                    <div className="srd-ov-card-title">
                      <Tag size={15} style={{ color: '#6366F1' }} />
                      Commercial Sourcing Details
                    </div>
                    <div className="srd-ov-grid">
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Origin Port</div>
                        <div className="srd-ov-field-val">{selectedRequest.origin_port || '—'}</div>
                      </div>
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Destination Port</div>
                        <div className="srd-ov-field-val">{selectedRequest.destination_port || '—'}</div>
                      </div>
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Transport Mode</div>
                        <div className="srd-ov-field-val">{selectedRequest.transport_mode || '—'}</div>
                      </div>
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Equipment Type</div>
                        <div className="srd-ov-field-val">{selectedRequest.equipment_type || '40GP'}</div>
                      </div>
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Container Quantity</div>
                        <div className="srd-ov-field-val">{selectedRequest.container_quantity ?? 1} units</div>
                      </div>
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Target Currency</div>
                        <div className="srd-ov-field-val">{selectedRequest.target_currency || 'USD'}</div>
                      </div>
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Cargo Ready Date</div>
                        <div className="srd-ov-field-val">{selectedRequest.ready_date || '—'}</div>
                      </div>
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Required Response By</div>
                        <div className="srd-ov-field-val">{selectedRequest.required_by_date || '—'}</div>
                      </div>
                      {selectedRequest.commodity && (
                        <div className="srd-ov-field">
                          <div className="srd-ov-field-label">Commodity</div>
                          <div className="srd-ov-field-val">{selectedRequest.commodity}</div>
                        </div>
                      )}
                      {selectedRequest.service_type && (
                        <div className="srd-ov-field">
                          <div className="srd-ov-field-label">Service Type</div>
                          <div className="srd-ov-field-val">{selectedRequest.service_type}</div>
                        </div>
                      )}
                      {selectedRequest.customer_name && (
                        <div className="srd-ov-field">
                          <div className="srd-ov-field-label">Customer / Account</div>
                          <div className="srd-ov-field-val">{selectedRequest.customer_name}</div>
                        </div>
                      )}
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Status</div>
                        <div className="srd-ov-field-val">
                          <span className={`sr-badge sr-badge-${selectedRequest.status.toLowerCase().replace(/_/g, '-')}`}>
                            {selectedRequest.status.replace(/_/g, ' ')}
                          </span>
                        </div>
                      </div>
                    </div>
                    {selectedRequest.notes && (
                      <div className="srd-ov-notes">
                        <div className="srd-ov-notes-label">📝 Notes</div>
                        <div className="srd-ov-notes-text">{selectedRequest.notes}</div>
                      </div>
                    )}
                  </div>

                  {/* Sourcing Progress */}
                  <div className="srd-ov-card">
                    <div className="srd-ov-card-title">
                      <Award size={15} style={{ color: '#6366F1' }} />
                      Sourcing Progress
                    </div>
                    <div className="srd-ov-progress-row">
                      <div className="srd-ov-stat">
                        <div className="srd-ov-stat-val" style={{ color: '#2563EB' }}>{responses.length}</div>
                        <div className="srd-ov-stat-label">Carrier Quotes Received</div>
                      </div>
                      <div className="srd-ov-stat">
                        <div className="srd-ov-stat-val" style={{ color: '#16A34A' }}>
                          {responses.filter(r => r.is_preferred).length > 0 ? '✓' : '—'}
                        </div>
                        <div className="srd-ov-stat-label">Preferred Rate Selected</div>
                      </div>
                      <div className="srd-ov-stat">
                        <div className="srd-ov-stat-val" style={{ color: '#7C3AED' }}>
                          {responses.length > 0 ? `${(responses.reduce((a,b) => a + (b.transit_days || 0), 0) / responses.length).toFixed(0)}d` : '—'}
                        </div>
                        <div className="srd-ov-stat-label">Avg Transit (All Quotes)</div>
                      </div>
                      <div className="srd-ov-stat">
                        <div className="srd-ov-stat-val" style={{ color: '#D97706' }}>
                          {responses.length > 1 ? (
                            (() => {
                              const amounts = responses.map(r => r.total_amount).filter(Boolean);
                              const spread = Math.max(...amounts) - Math.min(...amounts);
                              return `${responses[0]?.currency || 'USD'} ${Number(spread).toLocaleString(undefined, { maximumFractionDigits: 0 })}`;
                            })()
                          ) : '—'}
                        </div>
                        <div className="srd-ov-stat-label">Price Spread (High−Low)</div>
                      </div>
                    </div>
                  </div>

                  {/* Reference info */}
                  <div className="srd-ov-card">
                    <div className="srd-ov-card-title">
                      <Send size={15} style={{ color: '#6366F1' }} />
                      Request Reference
                    </div>
                    <div className="srd-ov-grid">
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Request ID</div>
                        <div className="srd-ov-field-val srd-ov-field-mono">{selectedRequest.id}</div>
                      </div>
                      <div className="srd-ov-field">
                        <div className="srd-ov-field-label">Reference No.</div>
                        <div className="srd-ov-field-val srd-ov-field-mono">{selectedRequest.request_reference}</div>
                      </div>
                      {selectedRequest.created_at && (
                        <div className="srd-ov-field">
                          <div className="srd-ov-field-label">Created At</div>
                          <div className="srd-ov-field-val">
                            {new Date(selectedRequest.created_at).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })}
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              )}

            </div>
          </div>
        </div>
      )}

      {/* ── Modal: Create Spot Rate Request ── */}
      {showCreateModal && (
        <div className="sr-modal-overlay">
          <div className="sr-modal">
            <div className="sr-modal-header">
              <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 800 }}>Create Spot Rate Sourcing Request</h3>
              <button className="rc-modal-close" onClick={() => setShowCreateModal(false)}>
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleCreateRequest}>
              <div className="sr-modal-body">
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                  <div className="rm-field-group">
                    <label>Origin Port *</label>
                    <input
                      type="text"
                      className="rm-input"
                      placeholder="e.g. USNYC, New York"
                      value={newRequest.origin_port}
                      onChange={(e) => setNewRequest({ ...newRequest, origin_port: e.target.value })}
                      required
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Destination Port *</label>
                    <input
                      type="text"
                      className="rm-input"
                      placeholder="e.g. INNSA, Nhava Sheva"
                      value={newRequest.destination_port}
                      onChange={(e) => setNewRequest({ ...newRequest, destination_port: e.target.value })}
                      required
                    />
                  </div>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px' }}>
                  <div className="rm-field-group">
                    <label>Transport Mode</label>
                    <CustomSelect
                      value={newRequest.transport_mode}
                      onChange={(val) => setNewRequest({ ...newRequest, transport_mode: val })}
                      options={[
                        { value: 'Ocean FCL', label: 'Ocean FCL' },
                        { value: 'Ocean LCL', label: 'Ocean LCL' },
                        { value: 'Air Freight', label: 'Air Freight' },
                      ]}
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Equipment Type</label>
                    <CustomSelect
                      value={newRequest.equipment_type}
                      onChange={(val) => setNewRequest({ ...newRequest, equipment_type: val })}
                      options={[
                        { value: '20GP', label: '20GP' },
                        { value: '40GP', label: '40GP' },
                        { value: '40HC', label: '40HC' },
                        { value: '45HC', label: '45HC' },
                        { value: 'Reefer 40', label: 'Reefer 40' },
                      ]}
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Quantity</label>
                    <input
                      type="number"
                      min="1"
                      className="rm-input"
                      value={newRequest.container_quantity}
                      onChange={(e) => setNewRequest({ ...newRequest, container_quantity: e.target.value })}
                    />
                  </div>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                  <div className="rm-field-group">
                    <label>Cargo Ready Date *</label>
                    <input
                      type="date"
                      className="rm-input"
                      value={newRequest.ready_date}
                      onChange={(e) => setNewRequest({ ...newRequest, ready_date: e.target.value })}
                      required
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Quotes Required By *</label>
                    <input
                      type="date"
                      className="rm-input"
                      value={newRequest.required_by_date}
                      onChange={(e) => setNewRequest({ ...newRequest, required_by_date: e.target.value })}
                      required
                    />
                  </div>
                </div>

                <div className="rm-field-group">
                  <label>Customer Name (Optional)</label>
                  <input
                    type="text"
                    className="rm-input"
                    placeholder="e.g. Apex Industrial Solutions"
                    value={newRequest.customer_name}
                    onChange={(e) => setNewRequest({ ...newRequest, customer_name: e.target.value })}
                  />
                </div>

                <div className="rm-field-group">
                  <label>Commercial Sourcing Notes</label>
                  <textarea
                    className="rm-input"
                    rows="2"
                    placeholder="Specify target rate, cargo sensitivities, or routing requirements..."
                    value={newRequest.notes}
                    onChange={(e) => setNewRequest({ ...newRequest, notes: e.target.value })}
                  />
                </div>
              </div>
              <div className="sr-modal-footer">
                <button type="button" className="sr-btn-secondary" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="sr-btn-primary">
                  Broadcast Spot Request
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Modal: Add Carrier Response ── */}
      {showAddResponseModal && selectedRequest && (
        <div className="sr-modal-overlay">
          <div className="sr-modal">
            <div className="sr-modal-header">
              <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 800 }}>
                Log Carrier Spot Quote ({selectedRequest.origin_port} → {selectedRequest.destination_port})
              </h3>
              <button className="rc-modal-close" onClick={() => setShowAddResponseModal(false)}>
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleAddResponse}>
              <div className="sr-modal-body">
                <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '12px' }}>
                  <div className="rm-field-group">
                    <label>Carrier Name *</label>
                    <input
                      type="text"
                      className="rm-input"
                      placeholder="e.g. MSC, Maersk, CMA CGM, Hapag-Lloyd"
                      value={newResponse.carrier_name}
                      onChange={(e) => setNewResponse({ ...newResponse, carrier_name: e.target.value })}
                      required
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Carrier Code (SCAC)</label>
                    <input
                      type="text"
                      className="rm-input"
                      placeholder="e.g. MSCU"
                      value={newResponse.carrier_code}
                      onChange={(e) => setNewResponse({ ...newResponse, carrier_code: e.target.value })}
                    />
                  </div>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '12px' }}>
                  <div className="rm-field-group">
                    <label>Currency</label>
                    <CustomSelect
                      value={newResponse.currency}
                      onChange={(val) => setNewResponse({ ...newResponse, currency: val })}
                      options={[
                        { value: 'USD', label: 'USD ($)' },
                        { value: 'EUR', label: 'EUR (€)' },
                        { value: 'GBP', label: 'GBP (£)' },
                        { value: 'INR', label: 'INR (₹)' },
                      ]}
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Base Freight Rate *</label>
                    <input
                      type="number"
                      step="0.01"
                      className="rm-input"
                      placeholder="e.g. 2150.00"
                      value={newResponse.base_amount}
                      onChange={(e) => setNewResponse({ ...newResponse, base_amount: e.target.value })}
                      required
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Transit Days</label>
                    <input
                      type="number"
                      className="rm-input"
                      placeholder="e.g. 22"
                      value={newResponse.transit_days}
                      onChange={(e) => setNewResponse({ ...newResponse, transit_days: e.target.value })}
                    />
                  </div>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                  <div className="rm-field-group">
                    <label>Free Days Origin</label>
                    <input
                      type="number"
                      className="rm-input"
                      value={newResponse.free_days_origin}
                      onChange={(e) => setNewResponse({ ...newResponse, free_days_origin: e.target.value })}
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Free Days Destination</label>
                    <input
                      type="number"
                      className="rm-input"
                      value={newResponse.free_days_destination}
                      onChange={(e) => setNewResponse({ ...newResponse, free_days_destination: e.target.value })}
                    />
                  </div>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                  <div className="rm-field-group">
                    <label>Valid From</label>
                    <input
                      type="date"
                      className="rm-input"
                      value={newResponse.valid_from}
                      onChange={(e) => setNewResponse({ ...newResponse, valid_from: e.target.value })}
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Valid Until</label>
                    <input
                      type="date"
                      className="rm-input"
                      value={newResponse.valid_until}
                      onChange={(e) => setNewResponse({ ...newResponse, valid_until: e.target.value })}
                    />
                  </div>
                </div>

                <div className="rm-field-group">
                  <label>Routing Notes</label>
                  <input
                    type="text"
                    className="rm-input"
                    placeholder="e.g. Direct via Suez / Transshipment at Singapore"
                    value={newResponse.routing_notes}
                    onChange={(e) => setNewResponse({ ...newResponse, routing_notes: e.target.value })}
                  />
                </div>
              </div>
              <div className="sr-modal-footer">
                <button type="button" className="sr-btn-secondary" onClick={() => setShowAddResponseModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="sr-btn-primary">
                  Save Carrier Quote
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Custom Confirmation Modal ── */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        title={confirmModal.title}
        message={confirmModal.message}
        confirmText={confirmModal.confirmText}
        confirmStyle={confirmModal.confirmStyle}
        onConfirm={confirmModal.action}
        onCancel={() => setConfirmModal((prev) => ({ ...prev, isOpen: false }))}
      />
    </div>
  );
}
