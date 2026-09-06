import React, { useState, useEffect, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import bookingService from '../../../services/bookingService';
import {
  RotateCw, Plus, Search, Database, Clock, CheckCircle2,
  Ship, Layers, FileText, AlertCircle, Filter, X,
  ArrowRight, ShieldCheck, Download, Sparkles, Anchor, Compass
} from 'lucide-react';
import ModuleHeroEmptyState from '../../../components/dashboard/ModuleHeroEmptyState';
import './BookingsPage.css';

export default function BookingsPage() {
  const navigate = useNavigate();

  // State
  const [bookings, setBookings] = useState([]);
  const [kpis, setKpis] = useState({
    total_bookings: 0,
    draft: 0,
    requested: 0,
    pending_confirmation: 0,
    confirmed: 0,
    completed: 0,
    cancelled: 0,
    departing_soon: 0,
  });
  const [pagination, setPagination] = useState({
    current_page: 1,
    page_size: 10,
    total_items: 0,
    total_pages: 1,
  });

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Filters
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [carrierFilter, setCarrierFilter] = useState('');
  const [carriers, setCarriers] = useState([]);
  const [page, setPage] = useState(1);
  const limit = 10;

  // Create Booking Modal State
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalStep, setModalStep] = useState(1); // 1 = Pick RFQ, 2 = Enter Booking Details
  const [eligibleRFQs, setEligibleRFQs] = useState([]);
  const [loadingEligible, setLoadingEligible] = useState(false);
  const [selectedRFQ, setSelectedRFQ] = useState(null);
  const [bookingFormData, setBookingFormData] = useState({
    booking_number: '',
    carrier_name: '',
    carrier_scac: '',
    vessel_name: '',
    voyage_number: '',
    origin_port: '',
    destination_port: '',
    etd: '',
    eta: '',
    special_instructions: '',
  });
  const [submittingBooking, setSubmittingBooking] = useState(false);
  const [modalError, setModalError] = useState(null);

  // Quick Action Modal (Update Status / Confirm / Create Shipment)
  const [actionModalBooking, setActionModalBooking] = useState(null);
  const [actionType, setActionType] = useState(null); // 'STATUS' or 'SHIPMENT'
  const [actionStatusTarget, setActionStatusTarget] = useState('');
  const [actionNotes, setActionNotes] = useState('');
  const [containerNumberInput, setContainerNumberInput] = useState('');
  const [submittingAction, setSubmittingAction] = useState(false);

  // Load Bookings Workspace Data
  const loadBookings = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await bookingService.getBookings({
        page,
        limit,
        status: statusFilter,
        carrier: carrierFilter,
        search,
      });

      let bookingsList = [];
      let kpisData = {
        total_bookings: 0,
        draft: 0,
        requested: 0,
        pending_confirmation: 0,
        confirmed: 0,
        completed: 0,
        cancelled: 0,
        departing_soon: 0,
      };
      let paginationData = {
        current_page: 1,
        page_size: 10,
        total_items: 0,
        total_pages: 1,
      };
      let carriersList = [];

      if (res) {
        if (Array.isArray(res)) {
          bookingsList = res;
        } else {
          const root = res.data && !Array.isArray(res.data) ? res.data : res;
          bookingsList = Array.isArray(root.items) ? root.items : (Array.isArray(root.data) ? root.data : []);
          if (root.kpis) kpisData = root.kpis;
          if (root.pagination) paginationData = root.pagination;
          if (root.carriers) carriersList = root.carriers;
        }
      }

      setBookings(bookingsList);
      setKpis(kpisData);
      setPagination(paginationData);
      if (carriersList && carriersList.length > 0) {
        setCarriers(carriersList);
      }
    } catch (err) {
      console.error('Failed to load bookings workspace:', err);
      setError(err.response?.data?.error?.message || 'Unable to load bookings workspace. Please check your network.');
    } finally {
      setLoading(false);
    }
  }, [page, limit, statusFilter, carrierFilter, search]);

  useEffect(() => {
    loadBookings();
  }, [loadBookings]);

  // Load Eligible RFQs when opening Create Booking Modal
  const openCreateModal = async () => {
    setIsModalOpen(true);
    setModalStep(1);
    setSelectedRFQ(null);
    setModalError(null);
    try {
      setLoadingEligible(true);
      const res = await bookingService.getEligibleRFQs();
      let list = [];
      if (res) {
        if (Array.isArray(res)) {
          list = res;
        } else if (res.data) {
          list = Array.isArray(res.data) ? res.data : (Array.isArray(res.data.data) ? res.data.data : []);
        }
      }
      setEligibleRFQs(list);
    } catch (err) {
      console.error('Failed to load eligible RFQs:', err);
      setModalError('Could not load eligible RFQs with approved quotes.');
    } finally {
      setLoadingEligible(false);
    }
  };

  const handleSelectRFQ = (rfq) => {
    setSelectedRFQ(rfq);
    const today = new Date();
    const defaultETD = rfq.target_date ? rfq.target_date.split('T')[0] : today.toISOString().split('T')[0];
    const defaultETA = new Date(today.getTime() + (rfq.transit_time_days || 14) * 86400000).toISOString().split('T')[0];

    setBookingFormData({
      booking_number: `BKG-${rfq.carrier_scac || 'CAR'}-${Date.now().toString().slice(-6)}`,
      carrier_name: rfq.carrier_name || '',
      carrier_scac: rfq.carrier_scac || '',
      vessel_name: '',
      voyage_number: '',
      origin_port: rfq.origin_port || '',
      destination_port: rfq.destination_port || '',
      etd: defaultETD,
      eta: defaultETA,
      special_instructions: '',
    });
    setModalStep(2);
  };

  const handleCreateBookingSubmit = async (e) => {
    e.preventDefault();
    if (!selectedRFQ) return;
    try {
      setSubmittingBooking(true);
      setModalError(null);

      const payload = {
        quote_id: selectedRFQ.approved_quote_id,
        booking_number: bookingFormData.booking_number,
        carrier_name: bookingFormData.carrier_name,
        carrier_scac: bookingFormData.carrier_scac,
        origin_port: bookingFormData.origin_port,
        destination_port: bookingFormData.destination_port,
        vessel_name: bookingFormData.vessel_name || null,
        voyage_number: bookingFormData.voyage_number || null,
        etd: bookingFormData.etd ? new Date(bookingFormData.etd).toISOString() : null,
        eta: bookingFormData.eta ? new Date(bookingFormData.eta).toISOString() : null,
        special_instructions: bookingFormData.special_instructions || null,
      };

      await bookingService.createBooking(selectedRFQ.rfq_id, payload);
      setIsModalOpen(false);
      loadBookings();
    } catch (err) {
      console.error('Error creating booking:', err);
      setModalError(err.response?.data?.error?.message || 'Failed to create booking.');
    } finally {
      setSubmittingBooking(false);
    }
  };

  // Direct Status Update or Shipment Action Handler
  const handleQuickAction = (booking, type, targetStatus = '') => {
    setActionModalBooking(booking);
    setActionType(type);
    setActionStatusTarget(targetStatus);
    setActionNotes('');
    setContainerNumberInput(`MSKU${Date.now().toString().slice(-7)}`);
  };

  const handleSubmitQuickAction = async (e) => {
    e.preventDefault();
    if (!actionModalBooking) return;
    try {
      setSubmittingAction(true);
      if (actionType === 'STATUS') {
        await bookingService.updateBookingStatus(actionModalBooking.id, actionStatusTarget, actionNotes);
      } else if (actionType === 'SHIPMENT') {
        const containers = containerNumberInput
          ? containerNumberInput.split(',').map((s) => s.trim()).filter(Boolean)
          : [];
        await bookingService.createShipment(actionModalBooking.id, {
          container_numbers: containers,
          notes: actionNotes || 'Shipment handoff created from confirmed booking',
        });
      }
      setActionModalBooking(null);
      loadBookings();
    } catch (err) {
      alert(err.response?.data?.error?.message || 'Action failed.');
    } finally {
      setSubmittingAction(false);
    }
  };

  return (
    <div className="bookings-page-container">
      
      {/* ── 1. Modern Page Header ── */}
      <div className="bookings-page-header">
        <div className="bph-left">
          <div className="bph-badge-row">
            <span className="bph-tag">OPERATIONAL WORKSPACE</span>
            <span className="bph-dot">·</span>
            <span className="bph-status-live">
              <span className="bph-pulse"></span>
              Live Booking Operations
            </span>
          </div>
          <h1 className="bph-title">Carrier Bookings</h1>
          <p className="bph-description">
            Authoritative carrier space bookings, confirmation lifecycles, and shipment execution handoffs.
          </p>
        </div>
        <div className="bph-actions">
          <button className="bph-btn-outline" onClick={loadBookings} title="Refresh Database Records">
            <RotateCw size={14} />
            <span>Refresh</span>
          </button>
          <button className="bph-btn-primary" onClick={openCreateModal}>
            <Plus size={15} />
            <span>+ Create Booking</span>
          </button>
        </div>
      </div>

      {/* ── 2. Compact Operational Summary Bar ── */}
      <div className="bookings-portfolio-strip">
        <div className="portfolio-strip-left">
          <Database size={15} className="portfolio-strip-icon" />
          <span>
            <strong>Carrier Booking Operations</strong> — All records verified with commercial RFQ and Carrier Quote lineage.
          </span>
        </div>
        <div className="portfolio-strip-right">
          <span className="portfolio-live-dot"></span>
          <span className="portfolio-live-label">Total Active:</span>
          <strong>{loading ? '...' : kpis.total_bookings}</strong>
        </div>
      </div>

      {/* ── 3. 5-Metric Unified Workflow KPI Strip ── */}
      <div className="bookings-kpi-grid">
        
        {/* 1. Total Bookings */}
        <div
          className={`bookings-kpi-card kpi-indigo ${statusFilter === 'ALL' ? 'active' : ''}`}
          onClick={() => { setStatusFilter('ALL'); setPage(1); }}
        >
          <div className="kpi-watermark-icon">
            <FileText size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper icon-indigo">
              <FileText size={16} />
            </div>
            <span className="kpi-tag tag-indigo">Repository</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Total Bookings</span>
            <div className="kpi-card-number">{loading ? '...' : kpis.total_bookings}</div>
            <span className="kpi-card-subtitle">All active bookings</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-indigo"></span>
            <span className="kpi-footer-text">Operational baseline</span>
          </div>
        </div>

        {/* 2. Pending Action */}
        <div
          className={`bookings-kpi-card kpi-amber ${statusFilter === 'PENDING_ACTION' ? 'active' : ''}`}
          onClick={() => { setStatusFilter('PENDING_ACTION'); setPage(1); }}
        >
          <div className="kpi-watermark-icon">
            <Clock size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper icon-amber">
              <Clock size={16} />
            </div>
            <span className="kpi-tag tag-amber">Action Required</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Pending Action</span>
            <div className="kpi-card-number text-amber">
              {loading ? '...' : ((kpis.requested || 0) + (kpis.pending_confirmation || 0))}
            </div>
            <span className="kpi-card-subtitle">Req + Pending Space</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-amber"></span>
            <span className="kpi-footer-text">Carrier response needed</span>
          </div>
        </div>

        {/* 3. Confirmed Space */}
        <div
          className={`bookings-kpi-card kpi-emerald ${statusFilter === 'CONFIRMED' ? 'active' : ''}`}
          onClick={() => { setStatusFilter('CONFIRMED'); setPage(1); }}
        >
          <div className="kpi-watermark-icon">
            <CheckCircle2 size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper icon-emerald">
              <CheckCircle2 size={16} />
            </div>
            <span className="kpi-tag tag-emerald">Confirmed</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Confirmed Space</span>
            <div className="kpi-card-number text-emerald">{loading ? '...' : kpis.confirmed}</div>
            <span className="kpi-card-subtitle">Space allocation secured</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-emerald"></span>
            <span className="kpi-footer-text">Ready for execution</span>
          </div>
        </div>

        {/* 4. Departing Soon */}
        <div
          className="bookings-kpi-card kpi-blue"
          onClick={() => { setStatusFilter('CONFIRMED'); setPage(1); }}
        >
          <div className="kpi-watermark-icon">
            <Ship size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper icon-blue">
              <Ship size={16} />
            </div>
            <span className="kpi-tag tag-blue">ETD ≤ 7d</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Departing Soon</span>
            <div className="kpi-card-number text-blue">{loading ? '...' : kpis.departing_soon}</div>
            <span className="kpi-card-subtitle">Upcoming departures</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-blue"></span>
            <span className="kpi-footer-text">Vessel window active</span>
          </div>
        </div>

        {/* 5. Executed / Completed */}
        <div
          className={`bookings-kpi-card kpi-purple ${statusFilter === 'COMPLETED' ? 'active' : ''}`}
          onClick={() => { setStatusFilter('COMPLETED'); setPage(1); }}
        >
          <div className="kpi-watermark-icon">
            <Layers size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper icon-purple">
              <Layers size={16} />
            </div>
            <span className="kpi-tag tag-purple">Executed</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Executed / Handoff</span>
            <div className="kpi-card-number text-purple">{loading ? '...' : kpis.completed}</div>
            <span className="kpi-card-subtitle">Converted to shipments</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-purple"></span>
            <span className="kpi-footer-text">Full shipment lifecycle</span>
          </div>
        </div>

      </div>

      {/* ── 4. Search and Filter Workspace ── */}
      <div className="bookings-toolbar">
        <div className="search-input-wrapper">
          <Search size={15} className="search-icon-lucide" />
          <input
            type="text"
            className="search-input"
            placeholder="Search booking #, RFQ #, carrier, vessel, or ports..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          />
          {search && (
            <button className="search-clear-btn" onClick={() => { setSearch(''); setPage(1); }}>
              <X size={13} />
            </button>
          )}
        </div>

        <div className="filter-dropdowns">
          <div className="filter-select-wrapper">
            <Filter size={14} className="filter-select-icon" />
            <select
              className="filter-select"
              value={carrierFilter}
              onChange={(e) => { setCarrierFilter(e.target.value); setPage(1); }}
            >
              <option value="">All Carriers</option>
              {carriers.map((c) => (
                <option key={c.carrier_name} value={c.carrier_name}>
                  {c.carrier_name} {c.carrier_scac ? `(${c.carrier_scac})` : ''}
                </option>
              ))}
            </select>
          </div>

          {(search || carrierFilter || statusFilter !== 'ALL') && (
            <button
              className="btn-clear-filters"
              onClick={() => {
                setSearch('');
                setCarrierFilter('');
                setStatusFilter('ALL');
                setPage(1);
              }}
            >
              <X size={12} />
              <span>Reset Filters</span>
            </button>
          )}
        </div>
      </div>

      {/* Status Filter Tabs */}
      <div className="status-tabs-container">
        {[
          { id: 'ALL', label: 'All Bookings', count: kpis.total_bookings },
          { id: 'PENDING_ACTION', label: 'Pending Action', count: (kpis.requested || 0) + (kpis.pending_confirmation || 0) },
          { id: 'DRAFT', label: 'Draft', count: kpis.draft },
          { id: 'REQUESTED', label: 'Requested', count: kpis.requested },
          { id: 'PENDING_CONFIRMATION', label: 'Pending Space', count: kpis.pending_confirmation },
          { id: 'CONFIRMED', label: 'Confirmed', count: kpis.confirmed },
          { id: 'COMPLETED', label: 'Completed', count: kpis.completed },
          { id: 'CANCELLED', label: 'Cancelled', count: kpis.cancelled },
        ].map((tab) => (
          <button
            key={tab.id}
            className={`status-tab ${statusFilter === tab.id ? 'active' : ''}`}
            onClick={() => { setStatusFilter(tab.id); setPage(1); }}
          >
            <span>{tab.label}</span>
            <span className="status-tab-count">{tab.count}</span>
          </button>
        ))}
      </div>

      {/* Main Table Card */}
      <div className="bookings-table-card">
        {error ? (
          <div className="empty-bookings-state error-state">
            <div className="empty-icon" style={{ color: '#ef4444' }}>⚠️</div>
            <h3 className="empty-title" style={{ color: '#ef4444' }}>Unable to load bookings</h3>
            <p className="empty-desc">{error}</p>
            <button className="btn-refresh" onClick={loadBookings} style={{ display: 'inline-block', marginTop: '12px' }}>
              🔄 Retry
            </button>
          </div>
        ) : loading ? (
          <div className="skeleton-loading">
            <span className="skeleton-pulse">···</span>
            <p style={{ marginTop: '12px', fontWeight: 600 }}>Loading Carrier Bookings Workspace...</p>
          </div>
        ) : bookings.length === 0 ? (
          search || carrierFilter || statusFilter !== 'ALL' ? (
            <div className="empty-bookings-state filter-empty-state">
              <div className="empty-icon">🔍</div>
              <h3 className="empty-title">No bookings match your current filters</h3>
              <p className="empty-desc">Try adjusting search terms or resetting filters.</p>
              <button
                className="btn-clear-filters"
                style={{ marginTop: '12px' }}
                onClick={() => {
                  setSearch('');
                  setCarrierFilter('');
                  setStatusFilter('ALL');
                  setPage(1);
                }}
              >
                ✕ Clear Filters
              </button>
            </div>
          ) : (
            <ModuleHeroEmptyState
              icon={<Anchor size={28} />}
              badgeTheme="blue"
              title="No Carrier Bookings Allocated Yet"
              description="Secure ocean container allocations and air cargo space by reserving capacity with contracted shipping lines and carriers for approved freight quotes."
              primaryAction={{
                label: 'Create First Booking',
                icon: <Plus size={15} />,
                onClick: openCreateModal,
              }}
              secondaryAction={{
                label: 'Explore Quotations',
                icon: <ArrowRight size={15} />,
                onClick: () => navigate('/dashboard/quotations'),
              }}
              features={[
                {
                  icon: <Ship size={18} />,
                  iconBg: '#eff6ff',
                  iconColor: '#2563eb',
                  title: 'Carrier Booking Requests & EDI',
                  desc: 'Generate structured booking requests for ocean liners and air cargo partner carriers.',
                },
                {
                  icon: <Clock size={18} />,
                  iconBg: '#fef3c7',
                  iconColor: '#d97706',
                  title: 'Space Guard & Cut-Off Tracking',
                  desc: 'Track VGM submission deadlines, shipping instruction (SI) windows, and carrier confirmation.',
                },
                {
                  icon: <Layers size={18} />,
                  iconBg: '#f5f3ff',
                  iconColor: '#7c3aed',
                  title: 'Automated Shipment Provisioning',
                  desc: 'Convert confirmed bookings into live operational shipments with container tracking telemetry.',
                },
              ]}
            />
          )
        ) : (
          <>
            <div className="table-responsive">
              <table className="bookings-table">
                <thead>
                  <tr>
                    <th className="bookings-th-num">Booking #</th>
                    <th className="bookings-th-status">Status</th>
                    <th className="bookings-th-rfq">RFQ / Customer</th>
                    <th className="bookings-th-carrier">Carrier</th>
                    <th className="bookings-th-route">Route</th>
                    <th className="bookings-th-vessel">Vessel</th>
                    <th className="bookings-th-schedule">Schedule</th>
                    <th className="bookings-th-shipment">Execution</th>
                    <th className="bookings-th-actions">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {bookings.map((b) => (
                    <tr key={b.id}>
                      {/* Booking Number */}
                      <td>
                        <Link to={`/dashboard/bookings/${b.id}`} className="booking-num-link">
                          {b.booking_number}
                        </Link>
                        <div className="booking-created-sub">
                          Created {new Date(b.created_at).toLocaleDateString()}
                        </div>
                      </td>

                      {/* Status */}
                      <td>
                        <span className={`status-pill ${b.status.toLowerCase()}`}>
                          <span className="status-dot"></span>
                          {b.status.replace(/_/g, ' ')}
                        </span>
                      </td>

                      {/* RFQ & Customer */}
                      <td>
                        <div style={{ fontWeight: 700, color: '#0f172a', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }} title={b.customer_name}>{b.customer_name}</div>
                        <div style={{ fontSize: '0.8rem', marginTop: '2px' }}>
                          <Link to={`/dashboard/rfqs/${b.rfq_id}`} style={{ color: '#4f46e5', textDecoration: 'none', fontWeight: 600 }}>
                            {b.rfq_number}
                          </Link>
                          {b.quote_sell_price > 0 && (
                            <span style={{ color: '#16a34a', marginLeft: '6px', fontWeight: 700 }}>
                              (${b.quote_sell_price.toLocaleString()})
                            </span>
                          )}
                        </div>
                      </td>

                      {/* Carrier */}
                      <td>
                        <div className="carrier-info">
                          <span className="carrier-name-bold" style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', display: 'block' }} title={b.carrier_name}>{b.carrier_name}</span>
                          {b.carrier_scac && (
                            <span className="carrier-vessel-sub">SCAC: {b.carrier_scac}</span>
                          )}
                        </div>
                      </td>

                      {/* Route */}
                      <td className="bookings-cell-route">
                        <div className="route-cell" title={`${b.origin_port} ➔ ${b.destination_port}`}>
                          <span className="route-port-text" title={b.origin_port}>{b.origin_port}</span>
                          <span className="route-arrow">➔</span>
                          <span className="route-port-text" title={b.destination_port}>{b.destination_port}</span>
                        </div>
                      </td>

                      {/* Vessel / Voyage */}
                      <td>
                        {b.vessel_name ? (
                          <div className="carrier-info">
                            <span style={{ fontWeight: 600, color: '#1e293b' }}>{b.vessel_name}</span>
                            {b.voyage_number && (
                              <span className="carrier-vessel-sub">Voy: {b.voyage_number}</span>
                            )}
                          </div>
                        ) : (
                          <span style={{ color: '#94a3b8', fontStyle: 'italic' }}>TBA by carrier</span>
                        )}
                      </td>

                      {/* Schedule */}
                      <td>
                        <div className="schedule-info">
                          <div className="schedule-row">
                            <span className="schedule-label">ETD:</span>
                            <span className="schedule-val">{b.etd ? new Date(b.etd).toLocaleDateString() : 'TBA'}</span>
                          </div>
                          <div className="schedule-row">
                            <span className="schedule-label">ETA:</span>
                            <span className="schedule-val">{b.eta ? new Date(b.eta).toLocaleDateString() : 'TBA'}</span>
                          </div>
                        </div>
                      </td>

                      {/* Shipment Execution Link */}
                      <td>
                        {b.shipment_id ? (
                          <Link
                            to={`/dashboard/shipments/${b.shipment_id}`}
                            className="shipment-handoff-badge active"
                            style={{ textDecoration: 'none' }}
                          >
                            <span>🚢 Shipment #{b.shipment_id}</span>
                          </Link>
                        ) : b.status === 'CONFIRMED' ? (
                          <span className="shipment-handoff-badge" style={{ background: '#fef3c7', color: '#92400e', border: '1px solid #fde68a' }}>
                            Ready for Execution
                          </span>
                        ) : (
                          <span className="shipment-handoff-badge none">Not Created</span>
                        )}
                      </td>

                      {/* Contextual Actions */}
                      <td>
                        <div className="action-buttons">
                          {b.status === 'DRAFT' && (
                            <button
                              className="btn-action-primary"
                              onClick={() => handleQuickAction(b, 'STATUS', 'REQUESTED')}
                            >
                              Request Booking
                            </button>
                          )}
                          {b.status === 'REQUESTED' && (
                            <button
                              className="btn-action-primary"
                              style={{ background: '#059669' }}
                              onClick={() => handleQuickAction(b, 'STATUS', 'CONFIRMED')}
                            >
                              Confirm Space
                            </button>
                          )}
                          {b.status === 'PENDING_CONFIRMATION' && (
                            <button
                              className="btn-action-primary"
                              style={{ background: '#059669' }}
                              onClick={() => handleQuickAction(b, 'STATUS', 'CONFIRMED')}
                            >
                              Confirm Space
                            </button>
                          )}
                          {b.status === 'CONFIRMED' && !b.shipment_id && (
                            <button
                              className="btn-action-primary"
                              style={{ background: '#0284c7' }}
                              onClick={() => handleQuickAction(b, 'SHIPMENT')}
                            >
                              Create Shipment
                            </button>
                          )}
                          <Link to={`/dashboard/bookings/${b.id}`} className="btn-action-secondary" style={{ textDecoration: 'none' }}>
                            View Detail
                          </Link>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination Footer */}
            <div className="pagination-footer">
              <div className="pagination-info">
                Showing <strong>{bookings.length}</strong> of <strong>{pagination.total_items}</strong> bookings
              </div>
              <div className="pagination-controls">
                <button
                  className="btn-page"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  ← Previous
                </button>
                <span style={{ fontSize: '0.85rem', fontWeight: 600, color: '#475569' }}>
                  Page {pagination.current_page} of {pagination.total_pages || 1}
                </span>
                <button
                  className="btn-page"
                  disabled={page >= pagination.total_pages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Next →
                </button>
              </div>
            </div>
          </>
        )}
      </div>

      {/* ── CREATE BOOKING MODAL (2 STEPS) ── */}
      {isModalOpen && (
        <div className="modal-overlay" onClick={() => setIsModalOpen(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                {modalStep === 1 ? 'Select Eligible RFQ with Approved Quote' : 'Carrier Space Booking Details'}
              </h2>
              <button className="btn-close-modal" onClick={() => setIsModalOpen(false)}>✕</button>
            </div>

            <div className="modal-body">
              {modalError && (
                <div className="error-banner" style={{ marginBottom: '16px' }}>
                  {modalError}
                </div>
              )}

              {modalStep === 1 ? (
                <div>
                  {loadingEligible ? (
                    <div className="skeleton-loading">
                      <span className="skeleton-pulse">···</span>
                      <p>Loading eligible RFQs from database...</p>
                    </div>
                  ) : eligibleRFQs.length === 0 ? (
                    <div className="modal-eligible-empty-state">
                      <div className="modal-empty-header">
                        <div className="modal-empty-badge">
                          <AlertCircle size={13} />
                          <span>Prerequisite Workflow Required</span>
                        </div>
                        <div className="modal-empty-icon-glow">
                          <Ship size={32} className="modal-empty-icon-main" />
                        </div>
                        <h3 className="modal-empty-title">No RFQs Ready for Carrier Space Allocation</h3>
                        <p className="modal-empty-desc">
                          Carrier space bookings secure container allocations against approved commercial quotations. Currently, there are no active RFQs with approved quotes awaiting capacity reservation.
                        </p>
                      </div>

                      {/* 3-Step Lifecycle Visual Pipeline */}
                      <div className="modal-workflow-stepper">
                        <div className="workflow-step completed">
                          <div className="step-num">1</div>
                          <div className="step-content">
                            <span className="step-title">1. Create / Receive RFQ</span>
                            <span className="step-desc">Capture POL ➔ POD route, container types, and cargo weight.</span>
                          </div>
                        </div>
                        <div className="workflow-step-connector">➔</div>
                        <div className="workflow-step current">
                          <div className="step-num">2</div>
                          <div className="step-content">
                            <span className="step-title">2. Approve Quotation</span>
                            <span className="step-desc">Generate commercial quote and mark as approved / won by customer.</span>
                          </div>
                        </div>
                        <div className="workflow-step-connector">➔</div>
                        <div className="workflow-step next">
                          <div className="step-num">3</div>
                          <div className="step-content">
                            <span className="step-title">3. Carrier Allocation</span>
                            <span className="step-desc">Book container space with shipping line and trigger shipment lifecycle.</span>
                          </div>
                        </div>
                      </div>

                      {/* Action Cards */}
                      <div className="modal-empty-actions-grid">
                        <div
                          className="modal-action-card primary-card"
                          onClick={() => {
                            setIsModalOpen(false);
                            navigate('/dashboard/rfqs');
                          }}
                        >
                          <div className="card-icon-wrap">
                            <FileText size={20} />
                          </div>
                          <div className="card-text-wrap">
                            <div className="card-title">
                              <span>Go to RFQs Pipeline</span>
                              <ArrowRight size={14} className="card-arrow" />
                            </div>
                            <p className="card-desc">
                              View open inquiries, compute spot buy/sell margins, and approve quotes.
                            </p>
                          </div>
                        </div>

                        <div
                          className="modal-action-card secondary-card"
                          onClick={() => {
                            setIsModalOpen(false);
                            navigate('/dashboard/quotations');
                          }}
                        >
                          <div className="card-icon-wrap">
                            <CheckCircle2 size={20} />
                          </div>
                          <div className="card-text-wrap">
                            <div className="card-title">
                              <span>Manage Quotations</span>
                              <ArrowRight size={14} className="card-arrow" />
                            </div>
                            <p className="card-desc">
                              Review pending customer quotations and convert approved bids into bookings.
                            </p>
                          </div>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div>
                      <p style={{ fontSize: '0.9rem', color: '#64748b', margin: '0 0 12px 0' }}>
                        Choose an RFQ that has an approved commercial quotation and is ready for carrier space reservation:
                      </p>
                      <div className="eligible-rfq-list">
                        {eligibleRFQs.map((rfq) => (
                          <div
                            key={rfq.rfq_id}
                            className="eligible-rfq-item"
                            onClick={() => handleSelectRFQ(rfq)}
                          >
                            <div className="eligible-rfq-top">
                              <span className="eligible-rfq-num">{rfq.rfq_number} — {rfq.customer_name}</span>
                              <span className="eligible-rfq-price">
                                ${rfq.sell_price.toLocaleString()} {rfq.currency}
                              </span>
                            </div>
                            <div className="eligible-rfq-meta">
                              <span>🚢 Carrier: <strong>{rfq.carrier_name}</strong></span>
                              <span>📍 {rfq.origin_port} ➔ {rfq.destination_port}</span>
                              <span>📦 {rfq.cargo_description} ({rfq.total_weight_kg.toLocaleString()} kg)</span>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <form id="create-booking-form" onSubmit={handleCreateBookingSubmit}>
                  <div style={{ background: '#f8fafc', padding: '12px 16px', borderRadius: '8px', marginBottom: '16px', border: '1px solid #e2e8f0', fontSize: '0.85rem' }}>
                    Selected: <strong>{selectedRFQ?.rfq_number}</strong> ({selectedRFQ?.customer_name}) | Carrier: <strong>{selectedRFQ?.carrier_name}</strong>
                  </div>

                  <div className="form-grid">
                    <div className="form-group">
                      <label className="form-label">Carrier Booking Number *</label>
                      <input
                        type="text"
                        className="form-input"
                        required
                        value={bookingFormData.booking_number}
                        onChange={(e) => setBookingFormData({ ...bookingFormData, booking_number: e.target.value })}
                      />
                    </div>

                    <div className="form-group">
                      <label className="form-label">Carrier Name *</label>
                      <input
                        type="text"
                        className="form-input"
                        required
                        value={bookingFormData.carrier_name}
                        onChange={(e) => setBookingFormData({ ...bookingFormData, carrier_name: e.target.value })}
                      />
                    </div>

                    <div className="form-group">
                      <label className="form-label">Vessel Name</label>
                      <input
                        type="text"
                        className="form-input"
                        placeholder="e.g. Maersk Mc-Kinney"
                        value={bookingFormData.vessel_name}
                        onChange={(e) => setBookingFormData({ ...bookingFormData, vessel_name: e.target.value })}
                      />
                    </div>

                    <div className="form-group">
                      <label className="form-label">Voyage Number</label>
                      <input
                        type="text"
                        className="form-input"
                        placeholder="e.g. 2608E"
                        value={bookingFormData.voyage_number}
                        onChange={(e) => setBookingFormData({ ...bookingFormData, voyage_number: e.target.value })}
                      />
                    </div>

                    <div className="form-group">
                      <label className="form-label">Estimated Departure (ETD)</label>
                      <input
                        type="date"
                        className="form-input"
                        value={bookingFormData.etd}
                        onChange={(e) => setBookingFormData({ ...bookingFormData, etd: e.target.value })}
                      />
                    </div>

                    <div className="form-group">
                      <label className="form-label">Estimated Arrival (ETA)</label>
                      <input
                        type="date"
                        className="form-input"
                        value={bookingFormData.eta}
                        onChange={(e) => setBookingFormData({ ...bookingFormData, eta: e.target.value })}
                      />
                    </div>

                    <div className="form-group full-width">
                      <label className="form-label">Special Cargo Instructions</label>
                      <textarea
                        className="form-textarea"
                        rows="2"
                        placeholder="Container pickup instructions, temperature setpoints, or cut-off notices..."
                        value={bookingFormData.special_instructions}
                        onChange={(e) => setBookingFormData({ ...bookingFormData, special_instructions: e.target.value })}
                      />
                    </div>
                  </div>
                </form>
              )}
            </div>

            <div className="modal-footer">
              {modalStep === 2 ? (
                <>
                  <button type="button" className="btn-modal-secondary" onClick={() => setModalStep(1)}>
                    ← Back to RFQ Picker
                  </button>
                  <button
                    type="submit"
                    form="create-booking-form"
                    className="btn-modal-primary"
                    disabled={submittingBooking}
                  >
                    {submittingBooking ? 'Saving Booking...' : 'Confirm & Create Booking'}
                  </button>
                </>
              ) : eligibleRFQs.length === 0 ? (
                <>
                  <button type="button" className="btn-modal-secondary" onClick={() => setIsModalOpen(false)}>
                    Dismiss
                  </button>
                  <button
                    type="button"
                    className="btn-modal-primary"
                    onClick={() => {
                      setIsModalOpen(false);
                      navigate('/dashboard/rfqs');
                    }}
                  >
                    <Compass size={15} />
                    <span>Explore RFQs Pipeline</span>
                    <ArrowRight size={14} />
                  </button>
                </>
              ) : (
                <button type="button" className="btn-modal-secondary" onClick={() => setIsModalOpen(false)}>
                  Cancel
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ── QUICK ACTION MODAL (CONFIRM SPACE / CREATE SHIPMENT) ── */}
      {actionModalBooking && (
        <div className="modal-overlay" onClick={() => setActionModalBooking(null)}>
          <div className="modal-card" style={{ maxWidth: '520px' }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                {actionType === 'STATUS' ? `Update Status: ${actionStatusTarget}` : 'Handoff to Shipment Execution'}
              </h2>
              <button className="btn-close-modal" onClick={() => setActionModalBooking(null)}>✕</button>
            </div>

            <form onSubmit={handleSubmitQuickAction}>
              <div className="modal-body">
                <p style={{ fontSize: '0.88rem', color: '#475569', margin: '0 0 16px 0' }}>
                  Booking: <strong>{actionModalBooking.booking_number}</strong> | Carrier: <strong>{actionModalBooking.carrier_name}</strong>
                </p>

                {actionType === 'SHIPMENT' && (
                  <div className="form-group" style={{ marginBottom: '16px' }}>
                    <label className="form-label">Container Numbers (comma-separated) *</label>
                    <input
                      type="text"
                      className="form-input"
                      required
                      placeholder="e.g. MSKU9012345, MSKU9012346"
                      value={containerNumberInput}
                      onChange={(e) => setContainerNumberInput(e.target.value)}
                    />
                    <span style={{ fontSize: '0.75rem', color: '#64748b' }}>
                      Containers assigned for vessel execution.
                    </span>
                  </div>
                )}

                <div className="form-group">
                  <label className="form-label">Operational Notes / Reason</label>
                  <textarea
                    className="form-textarea"
                    rows="2"
                    placeholder="Enter notes for audit trail..."
                    value={actionNotes}
                    onChange={(e) => setActionNotes(e.target.value)}
                  />
                </div>
              </div>

              <div className="modal-footer">
                <button type="button" className="btn-modal-secondary" onClick={() => setActionModalBooking(null)}>
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-modal-primary"
                  style={{
                    background: actionType === 'SHIPMENT' 
                      ? 'linear-gradient(135deg, #0284c7 0%, #0369a1 100%)' 
                      : 'linear-gradient(135deg, #059669 0%, #047857 100%)'
                  }}
                  disabled={submittingAction}
                >
                  {submittingAction ? 'Processing...' : actionType === 'SHIPMENT' ? 'Create Shipment' : 'Confirm Space'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
