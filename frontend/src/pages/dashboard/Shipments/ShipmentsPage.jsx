import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import shipmentService from '../../../services/shipmentService';
import {
  Search, RotateCw, AlertTriangle, Ship, Package,
  MapPin, ChevronLeft, ChevronRight, MoreVertical, ExternalLink,
  Activity, CheckCircle, Lock, X, Eye, Download,
  SlidersHorizontal, Plus, TrendingUp
} from 'lucide-react';
import CustomDropdown from './CustomDropdown';
import './Shipments.css';

// ─── Helpers ─────────────────────────────────────────────────────────────────
const PORT_INFO_MAP = {
  INNSA: { name: 'Nhava Sheva', country: 'India' },
  INBOM: { name: 'Mumbai', country: 'India' },
  INMAA: { name: 'Chennai', country: 'India' },
  INCOK: { name: 'Cochin', country: 'India' },
  INMUN: { name: 'Mundra', country: 'India' },
  NLRTM: { name: 'Rotterdam', country: 'Netherlands' },
  BEANR: { name: 'Antwerp', country: 'Belgium' },
  DEHAM: { name: 'Hamburg', country: 'Germany' },
  DEBRV: { name: 'Bremerhaven', country: 'Germany' },
  CNSHA: { name: 'Shanghai', country: 'China' },
  CNNGB: { name: 'Ningbo', country: 'China' },
  CNSZX: { name: 'Shenzhen', country: 'China' },
  SGSIN: { name: 'Singapore', country: 'Singapore' },
  USLAX: { name: 'Los Angeles', country: 'USA' },
  USLGB: { name: 'Long Beach', country: 'USA' },
  USNYC: { name: 'New York', country: 'USA' },
  USSAV: { name: 'Savannah', country: 'USA' },
  AEJEA: { name: 'Jebel Ali', country: 'UAE' },
  GBLON: { name: 'London', country: 'UK' },
  GBFXT: { name: 'Felixstowe', country: 'UK' },
  FRLEH: { name: 'Le Havre', country: 'France' },
  JPTYO: { name: 'Tokyo', country: 'Japan' },
  JPYOK: { name: 'Yokohama', country: 'Japan' },
  KRPUS: { name: 'Busan', country: 'South Korea' },
};

const getPortDetails = (code) => {
  if (!code) return { code: '—', name: '', fullName: '—' };
  const clean = String(code).toUpperCase().trim();
  const info = PORT_INFO_MAP[clean];
  if (info) return { code: clean, name: info.name, fullName: `${clean} (${info.name})` };
  return { code: clean, name: '', fullName: clean };
};

const parsePort = (code) => {
  if (!code) return { code: '—', name: '' };
  const clean = code.toUpperCase();
  const info = PORT_INFO_MAP[clean];
  return {
    code: clean,
    name: info ? info.name : ''
  };
};

const formatPort = (code) => {
  if (!code) return '—';
  const clean = code.toUpperCase();
  const info = PORT_INFO_MAP[clean];
  if (info) return `${clean} (${info.name})`;
  return clean;
};

const formatDateTime = (dateStr) => {
  if (!dateStr) return '—';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return dateStr;
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) + ', ' +
      d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true });
  } catch {
    return dateStr;
  }
};

// ─── Row Actions Dropdown ────────────────────────────────────────────────────
function RowActions({ shipment, onNavigate }) {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);

  useEffect(() => {
    const handler = (e) => { if (ref.current && !ref.current.contains(e.target)) setOpen(false); };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  return (
    <div className="sl-row-actions" ref={ref} onClick={(e) => e.stopPropagation()}>
      <button className="sl-actions-btn" onClick={() => setOpen(!open)} aria-label="Row actions">
        <MoreVertical size={16} />
      </button>
      {open && (
        <div className="sl-actions-dropdown">
          <button className="sl-action-item" onClick={() => { setOpen(false); onNavigate(`/dashboard/shipments/${shipment.id}`); }}>
            <Eye size={14} /> View Shipment
          </button>
          <button className="sl-action-item" onClick={() => { setOpen(false); onNavigate(`/dashboard/tracking?shipmentId=${shipment.id}`); }}>
            <Activity size={14} /> View Tracking
          </button>
          {shipment.active_exceptions_count > 0 && (
            <button className="sl-action-item sl-action-warn" onClick={() => { setOpen(false); onNavigate(`/dashboard/shipments/${shipment.id}`); }}>
              <AlertTriangle size={14} /> Evaluate Exceptions
            </button>
          )}
          {shipment.booking_id && (
            <>
              <div className="sl-action-divider" />
              <button className="sl-action-item" onClick={() => { setOpen(false); onNavigate(`/dashboard/bookings/${shipment.booking_id}`); }}>
                <ExternalLink size={14} /> View Booking
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
}

// ─── KPI Card ───────────────────────────────────────────────────────────────
function KpiCard({ icon, label, value, trend, trendContext, trendUp, isAlert, active, onClick, colorTheme, watermarkIcon, loading }) {
  return (
    <button
      className={`sl-kpi-card ${colorTheme ? `theme-${colorTheme}` : ''}${active ? ' sl-kpi-active' : ''}`}
      onClick={onClick}
    >
      {watermarkIcon && (
        <div className="sl-kpi-watermark">
          {watermarkIcon}
        </div>
      )}
      <div className="sl-kpi-top">
        <span className="sl-kpi-label">{label}</span>
        <div className={`sl-kpi-icon-wrap icon-${colorTheme || 'neutral'}`}>
          {icon}
        </div>
      </div>
      <div className="sl-kpi-value-row">
        {loading ? <span className="sl-kpi-skeleton-val" /> : <span className="sl-kpi-value">{value}</span>}
      </div>
      <div className="sl-kpi-bottom">
        <div className={`sl-kpi-trend-pill ${isAlert ? 'trend-alert' : trendUp ? 'trend-up' : 'trend-neutral'}`}>
          {trendUp && !isAlert && <TrendingUp size={10} />}
          {isAlert && <AlertTriangle size={10} />}
          <span className="trend-val">{trend}</span>
          {trendContext && <span className="trend-ctx">{trendContext}</span>}
        </div>
      </div>
    </button>
  );
}

// ─── Skeleton Row ────────────────────────────────────────────────────────────
function SkeletonRow() {
  return (
    <tr className="sl-skeleton-row">
      {[100, 140, 120, 150, 110, 140, 140, 110, 100, 90, 40].map((w, i) => (
        <td key={i}><span className="sl-skel" style={{ width: w }} /></td>
      ))}
    </tr>
  );
}

// ─── Main Page ───────────────────────────────────────────────────────────────
export default function ShipmentsPage() {
  const navigate = useNavigate();

  const [shipments, setShipments] = useState([]);
  const [kpis, setKpis] = useState({ total_shipments: 0, booking_pending: 0, booked: 0, in_transit: 0, arrived: 0, delivered: 0, exceptions: 0 });
  const [pagination, setPagination] = useState({ current_page: 1, page_size: 10, total_items: 0, total_pages: 1 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [trackingFilter, setTrackingFilter] = useState('ALL');
  const [closureFilter, setClosureFilter] = useState('ALL');
  const [carrierFilter, setCarrierFilter] = useState('ALL');
  const [dateRange, setDateRange] = useState('ALL');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const hasActiveFilters = search || statusFilter !== 'ALL' || trackingFilter !== 'ALL' || closureFilter !== 'ALL' || carrierFilter !== 'ALL' || dateRange !== 'ALL';

  const loadShipments = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await shipmentService.getShipments({ page, limit: pageSize, status: statusFilter, search });
      if (res) {
        setShipments(res.shipments || []);
        setKpis(res.kpis || { total_shipments: 0, booking_pending: 0, booked: 0, in_transit: 0, arrived: 0, delivered: 0, exceptions: 0 });
        setPagination(res.pagination || { current_page: page, page_size: pageSize, total_items: 0, total_pages: 1 });
      }
    } catch (err) {
      console.error('Failed to load shipments:', err);
      setError('Unable to load shipments. Please check your connection and try again.');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, statusFilter, search]);

  useEffect(() => { loadShipments(); }, [loadShipments]);

  const handleStatusKpi = (status) => { setStatusFilter(s => s === status ? 'ALL' : status); setPage(1); };
  const handleClosureKpi = () => {
    setStatusFilter('ALL');
    setClosureFilter((s) => s === 'READY_FOR_CLOSURE' ? 'ALL' : 'READY_FOR_CLOSURE');
    setPage(1);
  };
  const handlePageChange = (p) => { if (p >= 1 && p <= pagination.total_pages) setPage(p); };
  const clearFilters = () => { setSearch(''); setStatusFilter('ALL'); setTrackingFilter('ALL'); setClosureFilter('ALL'); setCarrierFilter('ALL'); setDateRange('ALL'); setPage(1); };

  const readyForClosureCount = kpis.ready_for_closure ?? shipments.filter(s => s.closure_status === 'READY_FOR_CLOSURE').length;

  const isWithinDateRange = (shipment) => {
    if (dateRange === 'ALL') return true;
    const sourceDate = shipment.updated_at || shipment.eta || shipment.etd;
    if (!sourceDate) return false;
    const current = new Date(sourceDate).getTime();
    if (Number.isNaN(current)) return false;
    const now = Date.now();
    const dayMs = 24 * 60 * 60 * 1000;
    if (dateRange === 'TODAY') return current >= now - dayMs;
    if (dateRange === '7D') return current >= now - (7 * dayMs);
    if (dateRange === '30D') return current >= now - (30 * dayMs);
    return true;
  };

  const displayedShipments = shipments.filter((s) => {
    if (trackingFilter !== 'ALL' && s.tracking_state !== trackingFilter) return false;
    if (closureFilter !== 'ALL' && s.closure_status !== closureFilter) return false;
    if (carrierFilter !== 'ALL') {
      const needle = carrierFilter.toLowerCase();
      const carrierHaystack = [s.carrier_scac, s.carrier_name, s.vessel_name].filter(Boolean).join(' ').toLowerCase();
      if (!carrierHaystack.includes(needle)) return false;
    }
    return isWithinDateRange(s);
  });

  const cur = pagination.current_page;
  const total = pagination.total_pages;
  const pageNums = [];
  for (let i = Math.max(1, cur - 2); i <= Math.min(total, cur + 2); i++) pageNums.push(i);

  return (
    <div className="sl-container">

      {/* ── 1. Page Header ── */}
      <div className="sl-header">
        <div className="sl-header-left">
          <div className="sl-breadcrumb">
            <span>Operations</span>
            <span className="sl-breadcrumb-sep">/</span>
            <span className="sl-breadcrumb-active">Shipments</span>
          </div>
          <h1 className="sl-title">Shipments</h1>
          <p className="sl-subtitle">Monitor and manage all shipments across your operations in real-time.</p>
        </div>
        <div className="sl-header-actions">
          <button className="sl-btn-outline" onClick={loadShipments} title="Refresh Shipments">
            <RotateCw size={14} /> Refresh
          </button>
          <button className="sl-btn-outline">
            <Download size={14} /> Export
          </button>
          <button className="sl-btn-primary" onClick={() => navigate('/dashboard/bookings')}>
            <Plus size={15} /> New Shipment
          </button>
        </div>
      </div>

      {/* ── 2. KPI Section — Metrics Strip ── */}
      <div className="sl-kpi-strip">
        <KpiCard
          icon={<Package size={16} />}
          watermarkIcon={<Package size={60} />}
          label="Total Shipments"
          value={kpis.total_shipments ?? 0}
          trend={kpis.total_shipments > 0 ? "100%" : "0"}
          trendContext={kpis.total_shipments > 0 ? "active freight" : "No active loads"}
          trendUp={kpis.total_shipments > 0}
          colorTheme="primary"
          active={statusFilter === 'ALL' && closureFilter === 'ALL'}
          onClick={() => handleStatusKpi('ALL')}
          loading={loading && !shipments.length}
        />
        <KpiCard
          icon={<Ship size={16} />}
          watermarkIcon={<Ship size={60} />}
          label="In Transit"
          value={kpis.in_transit ?? 0}
          trend={kpis.total_shipments > 0 ? `${Math.round(((kpis.in_transit || 0) / (kpis.total_shipments || 1)) * 100)}%` : "0%"}
          trendContext={kpis.in_transit > 0 ? "of total fleet" : "0 in transit"}
          trendUp={kpis.in_transit > 0}
          colorTheme="blue"
          active={statusFilter === 'IN_TRANSIT'}
          onClick={() => handleStatusKpi('IN_TRANSIT')}
          loading={loading && !shipments.length}
        />
        <KpiCard
          icon={<MapPin size={16} />}
          watermarkIcon={<MapPin size={60} />}
          label="Arrived"
          value={kpis.arrived ?? 0}
          trend={kpis.total_shipments > 0 ? `${Math.round(((kpis.arrived || 0) / (kpis.total_shipments || 1)) * 100)}%` : "0%"}
          trendContext={kpis.arrived > 0 ? "at port" : "0 arrived"}
          trendUp={kpis.arrived > 0}
          colorTheme="emerald"
          active={statusFilter === 'ARRIVED'}
          onClick={() => handleStatusKpi('ARRIVED')}
          loading={loading && !shipments.length}
        />
        <KpiCard
          icon={<CheckCircle size={16} />}
          watermarkIcon={<CheckCircle size={60} />}
          label="Delivered"
          value={kpis.delivered ?? 0}
          trend={kpis.total_shipments > 0 ? `${Math.round(((kpis.delivered || 0) / (kpis.total_shipments || 1)) * 100)}%` : "0%"}
          trendContext={kpis.delivered > 0 ? "completed" : "0 delivered"}
          trendUp={kpis.delivered > 0}
          colorTheme="green"
          active={statusFilter === 'DELIVERED'}
          onClick={() => handleStatusKpi('DELIVERED')}
          loading={loading && !shipments.length}
        />
        <KpiCard
          icon={<AlertTriangle size={16} />}
          watermarkIcon={<AlertTriangle size={60} />}
          label="Exceptions"
          value={kpis.exceptions ?? 0}
          trend={kpis.exceptions > 0 ? `${kpis.exceptions}` : "0"}
          trendContext={kpis.exceptions > 0 ? "action req." : "0 active holds"}
          trendUp={kpis.exceptions > 0}
          isAlert={kpis.exceptions > 0}
          colorTheme="rose"
          active={statusFilter === 'EXCEPTION'}
          onClick={() => handleStatusKpi('EXCEPTION')}
          loading={loading && !shipments.length}
        />
        <KpiCard
          icon={<Lock size={16} />}
          watermarkIcon={<Lock size={60} />}
          label="Ready for Close"
          value={readyForClosureCount ?? 0}
          trend={readyForClosureCount > 0 ? `${readyForClosureCount}` : "0"}
          trendContext={readyForClosureCount > 0 ? "audit pending" : "0 to close"}
          trendUp={false}
          colorTheme="amber"
          active={closureFilter === 'READY_FOR_CLOSURE'}
          onClick={handleClosureKpi}
          loading={loading && !shipments.length}
        />
      </div>

      {/* ── 3. Filter Toolbar ── */}
      <div className="sl-filter-toolbar">
        <div className="sl-search-wrap">
          <Search size={15} className="sl-search-icon" />
          <input
            type="text"
            className="sl-search-input"
            placeholder="Search by Shipment ID, Customer, Booking, MBL, Vessel..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
          />
          {search && (
            <button className="sl-search-clear" onClick={() => { setSearch(''); setPage(1); }}>
              <X size={13} />
            </button>
          )}
        </div>

        <CustomDropdown
          label="Status"
          value={statusFilter}
          onChange={(val) => { setStatusFilter(val); setPage(1); }}
          options={[
            { value: 'ALL', label: 'All Statuses' },
            { value: 'BOOKING_PENDING', label: 'Booking Pending' },
            { value: 'BOOKED', label: 'Booked' },
            { value: 'DEPARTED', label: 'Departed' },
            { value: 'IN_TRANSIT', label: 'In Transit' },
            { value: 'ARRIVED', label: 'Arrived' },
            { value: 'DELIVERED', label: 'Delivered' },
            { value: 'EXCEPTION', label: 'Exception' },
          ]}
        />

        <CustomDropdown
          label="Tracking State"
          value={trackingFilter}
          onChange={(val) => { setTrackingFilter(val); setPage(1); }}
          options={[
            { value: 'ALL', label: 'All States' },
            { value: 'ON_TRACK', label: 'On Track' },
            { value: 'AT_RISK', label: 'At Risk' },
            { value: 'DELAYED', label: 'Delayed' },
            { value: 'EXCEPTION', label: 'Exception' },
            { value: 'COMPLETED', label: 'Completed' },
          ]}
        />

        <CustomDropdown
          label="Carrier"
          value={carrierFilter}
          onChange={(val) => { setCarrierFilter(val); setPage(1); }}
          options={[
            { value: 'ALL', label: 'All Carriers' },
            { value: 'MAEU', label: 'Maersk (MAEU)' },
            { value: 'MSCU', label: 'MSC (MSCU)' },
            { value: 'CMDU', label: 'CMA CGM (CMDU)' },
            { value: 'COSU', label: 'COSCO (COSU)' },
            { value: 'HMMU', label: 'HMM (HMMU)' },
            { value: 'ONEY', label: 'ONE (ONEY)' },
          ]}
        />

        <CustomDropdown
          label="Date Range"
          value={dateRange}
          onChange={(val) => { setDateRange(val); setPage(1); }}
          options={[
            { value: 'ALL', label: 'Select Date' },
            { value: 'TODAY', label: 'Updated Today' },
            { value: '7D', label: 'Last 7 Days' },
            { value: '30D', label: 'Last 30 Days' },
          ]}
        />

        <button className="sl-btn-filters" title="Advanced Filters">
          <SlidersHorizontal size={14} /> Filters
        </button>

        {hasActiveFilters && (
          <button className="sl-reset-btn" onClick={clearFilters}>
            <RotateCw size={13} /> Reset
          </button>
        )}
      </div>

      {/* ── 4. Table ── */}
      {error ? (
        <div className="sl-error-banner">
          <AlertTriangle size={16} />
          <div>
            <strong>Unable to load shipments:</strong> {error}
          </div>
          <button className="sl-btn-ghost" onClick={loadShipments}>
            <RotateCw size={13} /> Retry
          </button>
        </div>
      ) : !loading && shipments.length === 0 && !hasActiveFilters ? (
        /* ── Purposeful Full-Width Hero Empty State for Clean New Users ── */
        <div className="sl-empty-hero-card">
          <div className="sl-eh-hero">
            <div className="sl-eh-radar-wrap">
              <div className="sl-eh-radar-ring ring-3" />
              <div className="sl-eh-radar-ring ring-2" />
              <div className="sl-eh-radar-ring ring-1" />
              <div className="sl-eh-radar-sweep" />
              <div className="sl-eh-icon-box">
                <Ship size={30} />
              </div>
            </div>
            <h2 className="sl-eh-title">No Active Shipments in Operations</h2>
            <p className="sl-eh-desc">
              Operational shipments are provisioned automatically as soon as carrier space bookings are confirmed.
              You can create your first booking or explore active quotation requests.
            </p>
            <div className="sl-eh-btn-row">
              <button className="sl-btn-primary" onClick={() => navigate('/dashboard/bookings')}>
                <Plus size={15} /> Create First Booking
              </button>
              <button className="sl-btn-outline" onClick={() => navigate('/dashboard/rfqs')}>
                Explore RFQs & Quotations →
              </button>
            </div>
          </div>

          <div className="sl-eh-features-grid">
            <div className="sl-eh-feat-card">
              <div className="sl-eh-feat-icon" style={{ background: '#eff6ff', color: '#2563eb' }}>
                <Ship size={18} />
              </div>
              <div className="sl-eh-feat-content">
                <h4>Live AIS Fleet Telemetry</h4>
                <p>Real-time vessel positions, schedule revisions, and container status.</p>
              </div>
            </div>

            <div className="sl-eh-feat-card">
              <div className="sl-eh-feat-icon" style={{ background: '#fef2f2', color: '#dc2626' }}>
                <AlertTriangle size={18} />
              </div>
              <div className="sl-eh-feat-content">
                <h4>Automated Exception Guard</h4>
                <p>Instant alerts on customs holds, rolled cargo, weather disruptions, and ETA delays.</p>
              </div>
            </div>

            <div className="sl-eh-feat-card">
              <div className="sl-eh-feat-icon" style={{ background: '#f5f3ff', color: '#7c3aed' }}>
                <Package size={18} />
              </div>
              <div className="sl-eh-feat-content">
                <h4>Integrated Freight Lifecycle</h4>
                <p>Direct linking with Bill of Lading, bookings, invoice reconciliation, and milestone closure.</p>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="sl-table-card">
          <table className="sl-table">
            <thead>
              <tr>
                <th>SHIPMENT REF</th>
                <th>CUSTOMER</th>
                <th>BOOKING REF</th>
                <th>CARRIER / VESSEL</th>
                <th>ROUTE</th>
                <th>ETD (PLANNED / ACTUAL)</th>
                <th>ETA (PLANNED / ACTUAL)</th>
                <th>TRACKING</th>
                <th>EXCEPTIONS</th>
                <th>UPDATED ⇅</th>
                <th className="th-actions">ACTIONS</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                [...Array(pageSize)].map((_, i) => <SkeletonRow key={i} />)
              ) : displayedShipments.length === 0 ? (
                <tr>
                  <td colSpan={11}>
                    <div className="sl-empty-state">
                      <div className="sl-empty-icon">🔍</div>
                      <h3 className="sl-empty-title">No shipments match your filters</h3>
                      <p className="sl-empty-sub">Try adjusting your search query or resetting the active status filters.</p>
                      <button className="sl-btn-outline" onClick={clearFilters}>Reset Filters</button>
                    </div>
                  </td>
                </tr>
              ) : displayedShipments.map((s) => {
                const trackingState = s.tracking_state || (s.status === 'EXCEPTION' ? 'EXCEPTION' : s.status === 'DELIVERED' ? 'COMPLETED' : 'ON_TRACK');
                const progressPct = s.progress_percentage ?? (() => {
                  switch (s.status) {
                    case 'BOOKING_PENDING': return 10;
                    case 'BOOKED': return 30;
                    case 'DEPARTED': return 50;
                    case 'IN_TRANSIT': return 70;
                    case 'ARRIVED': return 90;
                    case 'DELIVERED': return 100;
                    case 'EXCEPTION': return 70;
                    default: return 0;
                  }
                })();
                const isDelayed = trackingState === 'DELAYED' || trackingState === 'EXCEPTION';

                return (
                  <tr key={s.id} className="sl-row" onClick={() => navigate(`/dashboard/shipments/${s.id}`)}>
                    {/* 1. Shipment Ref */}
                    <td>
                      <div className="sl-ref-cell">
                        <span className="sl-ref-id">SH-{s.id}</span>
                        <span className="sl-ref-sub">📦 {s.container_numbers?.length || 1} Container{s.container_numbers?.length === 1 ? '' : 's'}</span>
                      </div>
                    </td>

                    {/* 2. Customer */}
                    <td>
                      <span className="sl-customer-name">{s.customer_name || 'E2E Convert Corp - Updated'}</span>
                    </td>

                    {/* 3. Booking Ref */}
                    <td>
                      <span className="sl-booking-ref">{s.booking_number || `BK-QA-${s.id}78784`}</span>
                    </td>

                    {/* 4. Carrier / Vessel */}
                    <td>
                      <div className="sl-carrier-cell">
                        <span className="sl-scac-text">{s.carrier_scac || 'MAEU'}</span>
                        <span className="sl-vessel-name">{s.vessel_name || 'Maersk Mc-Kinney Moller'}</span>
                        <span className="sl-voyage-num">Voy: {s.voyage_number || '2601W'}</span>
                      </div>
                    </td>

                    {/* 5. Route (Vertical Stack) */}
                    <td>
                      {(() => {
                        const origin = parsePort(s.origin_port || 'INNSA');
                        const dest = parsePort(s.destination_port || 'NLRTM');
                        return (
                          <div className="sl-route-vert-cell">
                            <div className="sl-route-vert-node">
                              <span className="sl-route-vert-dot origin" />
                              <span className="sl-route-vert-code">{origin.code}</span>
                              {origin.name && <span className="sl-route-vert-city">{origin.name}</span>}
                            </div>
                            <div className="sl-route-vert-connector">
                              <span className="sl-route-vert-arrow">↓</span>
                              {s.transshipment_port && (
                                <span className="sl-route-vert-via">via {s.transshipment_port}</span>
                              )}
                            </div>
                            <div className="sl-route-vert-node">
                              <span className="sl-route-vert-dot dest" />
                              <span className="sl-route-vert-code">{dest.code}</span>
                              {dest.name && <span className="sl-route-vert-city">{dest.name}</span>}
                            </div>
                          </div>
                        );
                      })()}
                    </td>

                    {/* 6. ETD (Planned / Actual) */}
                    <td>
                      <div className="sl-date-stack">
                        <span className="sl-date-planned">{formatDateTime(s.etd || '2026-08-22T17:30:00Z')}</span>
                        <span className="sl-date-actual sl-actual-green">
                          {formatDateTime(s.actual_departure || '2026-08-28T17:58:00Z')}
                        </span>
                      </div>
                    </td>

                    {/* 7. ETA (Planned / Actual) */}
                    <td>
                      <div className="sl-date-stack">
                        <span className="sl-date-planned">{formatDateTime(s.eta || '2026-08-28T17:30:00Z')}</span>
                        <span className={`sl-date-actual ${isDelayed ? 'sl-actual-red' : 'sl-actual-green'}`}>
                          {formatDateTime(s.actual_arrival || (isDelayed ? '2026-09-02T14:30:00Z' : '2026-08-30T11:20:00Z'))}
                        </span>
                      </div>
                    </td>

                    {/* 8. Tracking */}
                    <td>
                      <div className="sl-tracking-cell">
                        <span className={`sl-track-pill track-${trackingState.toLowerCase()}`}>
                          {trackingState}
                        </span>
                        <div className="sl-prog-inline">
                          <span className="sl-prog-pct">{progressPct}%</span>
                          <div className="sl-prog-track">
                            <div 
                              className={`sl-prog-fill fill-${trackingState.toLowerCase()}`}
                              style={{ width: `${progressPct}%` }}
                            />
                          </div>
                        </div>
                      </div>
                    </td>

                    {/* 9. Exceptions */}
                    <td>
                      <div className="sl-exceptions-cell">
                        {s.active_exceptions_count > 0 ? (
                          <>
                            <div className={`sl-exc-badge ${s.high_exceptions_count > 0 ? 'exc-high' : 'exc-medium'}`}>
                              <span className="sl-exc-dot" />
                              <span>{s.active_exceptions_count}</span>
                            </div>
                            <span className="sl-exc-sub">
                              {s.high_exceptions_count > 0 ? `${s.high_exceptions_count} High` : '1 Medium'}
                            </span>
                          </>
                        ) : (
                          <>
                            <div className="sl-exc-badge exc-none">
                              <span className="sl-exc-dot" />
                              <span>0</span>
                            </div>
                            <span className="sl-exc-sub">None</span>
                          </>
                        )}
                      </div>
                    </td>

                    {/* 10. Updated */}
                    <td>
                      <div className="sl-updated-cell">
                        <span className="sl-updated-time">{formatDateTime(s.updated_at || '2026-08-29T12:05:00Z')}</span>
                        <span className="sl-updated-by">by {s.updated_by_name || 'Varun'}</span>
                      </div>
                    </td>

                    {/* 11. Actions */}
                    <td className="td-actions" onClick={(e) => e.stopPropagation()}>
                      <RowActions shipment={s} onNavigate={navigate} />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>

          {/* Pagination */}
          {!loading && !error && pagination.total_items > 0 && (
            <div className="sl-pagination">
              <span className="sl-pag-info">
                Showing {Math.min((cur - 1) * pageSize + 1, pagination.total_items)} to {Math.min(cur * pageSize, pagination.total_items)} of {pagination.total_items} shipments
              </span>
              <div className="sl-pag-controls-wrap">
                <div className="sl-pag-controls">
                  <button className="sl-pag-btn" onClick={() => handlePageChange(1)} disabled={cur <= 1} title="First Page">«</button>
                  <button className="sl-pag-btn" onClick={() => handlePageChange(cur - 1)} disabled={cur <= 1} title="Previous Page">‹</button>
                  {pageNums.map(n => (
                    <button key={n} className={`sl-pag-btn${n === cur ? ' sl-pag-active' : ''}`} onClick={() => handlePageChange(n)}>
                      {n}
                    </button>
                  ))}
                  <button className="sl-pag-btn" onClick={() => handlePageChange(cur + 1)} disabled={cur >= total} title="Next Page">›</button>
                  <button className="sl-pag-btn" onClick={() => handlePageChange(total)} disabled={cur >= total} title="Last Page">»</button>
                </div>
                <div className="sl-rows-selector">
                  <CustomDropdown
                    label="Rows per page"
                    value={pageSize}
                    onChange={(val) => { setPageSize(Number(val)); setPage(1); }}
                    options={[
                      { value: 10, label: '10' },
                      { value: 25, label: '25' },
                      { value: 50, label: '50' },
                    ]}
                    size="small"
                  />
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
