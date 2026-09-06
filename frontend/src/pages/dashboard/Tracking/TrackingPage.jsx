import React, { useState, useEffect, useMemo, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Ship,
  MapPin,
  Search,
  CheckCircle2,
  Clock,
  AlertTriangle,
  AlertCircle,
  ArrowRight,
  ExternalLink,
  RotateCw,
  Plus,
  Star,
  Filter,
  ChevronDown,
  Check,
  Eye,
  FileText,
  Activity,
  ShieldCheck,
  Navigation,
  Anchor,
  Layers,
  X,
  Calendar,
  MoreVertical,
  Radio,
  Compass,
  ArrowUpRight,
  TrendingUp,
  BarChart3,
} from 'lucide-react';
import api from '../../../services/api';
import { shipmentService } from '../../../services/shipmentService';
import CustomDropdown from '../Shipments/CustomDropdown';
import TrackingAnalyticsWorkspace from './TrackingAnalyticsWorkspace';
import './TrackingPage.css';

// ─── Port Directory ─────────────────────────────────────────────────────────
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
  AEDXB: { name: 'Dubai', country: 'UAE' },
  GBLON: { name: 'London', country: 'UK' },
  GBFXT: { name: 'Felixstowe', country: 'UK' },
  GBSOU: { name: 'Southampton', country: 'UK' },
  FRLEH: { name: 'Le Havre', country: 'France' },
  JPTYO: { name: 'Tokyo', country: 'Japan' },
  JPYOK: { name: 'Yokohama', country: 'Japan' },
  KRPUS: { name: 'Busan', country: 'South Korea' },
  BRSSZ: { name: 'Santos', country: 'Brazil' },
  VVBKK: { name: 'Bangkok', country: 'Thailand' },
  CYYZ: { name: 'Toronto', country: 'Canada' },
};

const parsePort = (code) => {
  if (!code) return { code: '—', name: '', country: '' };
  const clean = String(code).toUpperCase().trim();
  const info = PORT_INFO_MAP[clean];
  if (info) return { code: clean, name: info.name, country: info.country };
  return { code: clean, name: '', country: '' };
};

const formatDateOnly = (dateStr) => {
  if (!dateStr || dateStr === '—') return '—';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return dateStr;
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  } catch {
    return dateStr;
  }
};

const formatDateTime = (dateStr) => {
  if (!dateStr) return '—';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return dateStr;
    return (
      d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) +
      ' · ' +
      d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true })
    );
  } catch {
    return dateStr;
  }
};

// ─── Row Actions Dropdown ───────────────────────────────────────────────────
function RowActions({ shipment, onNavigate }) {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);

  useEffect(() => {
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  return (
    <div className="trk-row-actions" ref={ref} onClick={(e) => e.stopPropagation()}>
      <button className="trk-actions-btn" onClick={() => setOpen(!open)} aria-label="Tracking actions">
        <MoreVertical size={15} />
      </button>
      {open && (
        <div className="trk-actions-dropdown">
          <button
            className="trk-action-item"
            onClick={() => {
              setOpen(false);
              onNavigate(`/dashboard/tracking/${shipment.id}`);
            }}
          >
            <Activity size={14} /> View Full Tracking
          </button>
          <button
            className="trk-action-item"
            onClick={() => {
              setOpen(false);
              onNavigate(`/dashboard/shipments/${shipment.id}`);
            }}
          >
            <Eye size={14} /> View Shipment
          </button>
          {shipment.booking_id && (
            <button
              className="trk-action-item"
              onClick={() => {
                setOpen(false);
                onNavigate(`/dashboard/bookings/${shipment.booking_id}`);
              }}
            >
              <ExternalLink size={14} /> View Booking
            </button>
          )}
        </div>
      )}
    </div>
  );
}

// ─── Main Component ─────────────────────────────────────────────────────────
export default function TrackingPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const urlShipmentId = searchParams.get('shipmentId');

  // Core Data State
  const [shipments, setShipments] = useState([]);
  const [selectedShipmentId, setSelectedShipmentId] = useState(null);
  const [selectedDetail, setSelectedDetail] = useState(null);
  const [loadingList, setLoadingList] = useState(true);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [justRefreshed, setJustRefreshed] = useState(false);
  const [error, setError] = useState(null);

  // Filters & Search State
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [trackingStateFilter, setTrackingStateFilter] = useState('ALL');
  const [carrierFilter, setCarrierFilter] = useState('ALL');
  const [dateRangeFilter, setDateRangeFilter] = useState('ALL');
  const [activeTab, setActiveTab] = useState('overview'); // 'overview' | 'milestones' | 'events' | 'exceptions' | 'documents'
  const [workspaceView, setWorkspaceView] = useState(() => {
    const tab = searchParams.get('view') || searchParams.get('tab');
    return tab === 'analytics' ? 'analytics' : 'fleet';
  });
  const [starredIds, setStarredIds] = useState(() => {
    try {
      const saved = localStorage.getItem('logistics_starred_shipments');
      return saved ? JSON.parse(saved) : [];
    } catch {
      return [];
    }
  });

  // Modal State
  const [showAddTrackModal, setShowAddTrackModal] = useState(false);
  const [trackModalInput, setTrackModalInput] = useState('');
  const [trackModalError, setTrackModalError] = useState('');

  // Pagination State
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // Initial Load
  useEffect(() => {
    loadShipments();
  }, []);

  // Fetch full details whenever the selected shipment changes
  useEffect(() => {
    if (selectedShipmentId) {
      loadSelectedShipmentDetail(selectedShipmentId);
    } else {
      setSelectedDetail(null);
    }
  }, [selectedShipmentId]);

  const loadShipments = async (isBackground = false) => {
    const startTime = Date.now();
    try {
      if (isBackground) {
        setRefreshing(true);
      } else {
        setLoadingList(true);
      }
      setError(null);

      // Fetch shipments list from the API
      const res = await shipmentService.getShipments({ limit: 50 });
      const data = res?.data || res;
      const list = Array.isArray(data?.shipments) ? data.shipments : Array.isArray(data) ? data : [];

      setShipments(list);

      // Maintain selection
      if (list.length > 0) {
        if (urlShipmentId && list.some((s) => String(s.id) === String(urlShipmentId))) {
          setSelectedShipmentId(Number(urlShipmentId));
        } else if (!selectedShipmentId || !list.some((s) => s.id === selectedShipmentId)) {
          setSelectedShipmentId(list[0].id);
        }
      }

      if (isBackground) {
        const elapsed = Date.now() - startTime;
        if (elapsed < 650) {
          await new Promise((r) => setTimeout(r, 650 - elapsed));
        }
        setJustRefreshed(true);
        setTimeout(() => setJustRefreshed(false), 2000);
      }
    } catch (err) {
      console.error('Tracking fetch error:', err);
      setError('Unable to load active shipment telemetry. Please try again.');
    } finally {
      setLoadingList(false);
      setRefreshing(false);
    }
  };

  const loadSelectedShipmentDetail = async (id) => {
    try {
      setLoadingDetail(true);
      const res = await shipmentService.getShipmentDetail(id);
      const detail = res?.data || res;
      setSelectedDetail(detail);
    } catch (err) {
      console.warn('Error fetching shipment detail for tracking:', err);
      const shallow = shipments.find((s) => s.id === id);
      setSelectedDetail(shallow ? { shipment: shallow, milestones: [], exceptions: [], tracking_summary: null } : null);
    } finally {
      setLoadingDetail(false);
    }
  };

  const toggleStar = (id, e) => {
    e.stopPropagation();
    setStarredIds((prev) => {
      const next = prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id];
      try {
        localStorage.setItem('logistics_starred_shipments', JSON.stringify(next));
      } catch (err) {
        console.warn('Could not persist starred shipments:', err);
      }
      return next;
    });
  };

  // ─── Filtered Shipments ───────────────────────────────────────────────────
  const filteredShipments = useMemo(() => {
    return shipments.filter((s) => {
      // 1. Search Query
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase().trim();
        const matchRef = `sh-${s.id}`.includes(q) || String(s.id).includes(q);
        const matchBk = s.booking_number && s.booking_number.toLowerCase().includes(q);
        const matchCust = s.customer_name && s.customer_name.toLowerCase().includes(q);
        const matchScac = s.carrier_scac && s.carrier_scac.toLowerCase().includes(q);
        const matchVessel = s.vessel_name && s.vessel_name.toLowerCase().includes(q);
        const matchVoy = s.voyage_number && s.voyage_number.toLowerCase().includes(q);
        const matchOrigin = s.origin_port && s.origin_port.toLowerCase().includes(q);
        const matchDest = s.destination_port && s.destination_port.toLowerCase().includes(q);
        if (!matchRef && !matchBk && !matchCust && !matchScac && !matchVessel && !matchVoy && !matchOrigin && !matchDest) {
          return false;
        }
      }

      // 2. Status Filter
      if (statusFilter !== 'ALL' && s.status !== statusFilter) {
        return false;
      }

      // 3. Tracking State Filter
      if (trackingStateFilter !== 'ALL') {
        const state = s.tracking_state || (s.status === 'DELIVERED' ? 'COMPLETED' : s.status === 'EXCEPTION' ? 'EXCEPTION' : 'ON_TRACK');
        if (state !== trackingStateFilter) {
          return false;
        }
      }

      // 4. Carrier Filter
      if (carrierFilter !== 'ALL' && s.carrier_scac !== carrierFilter) {
        return false;
      }

      return true;
    });
  }, [shipments, searchQuery, statusFilter, trackingStateFilter, carrierFilter]);

  // ─── Live Computed KPIs ───────────────────────────────────────────────────
  const kpis = useMemo(() => {
    const total = shipments.length;
    let onTrack = 0;
    let atRisk = 0;
    let delayed = 0;
    let activeExceptions = 0;
    let completed = 0;

    shipments.forEach((s) => {
      const state = s.tracking_state || (s.status === 'DELIVERED' ? 'COMPLETED' : s.status === 'EXCEPTION' ? 'EXCEPTION' : 'ON_TRACK');
      if (state === 'ON_TRACK') onTrack++;
      if (state === 'AT_RISK') atRisk++;
      if (state === 'DELAYED' || state === 'EXCEPTION') delayed++;
      if (state === 'COMPLETED' || s.status === 'DELIVERED') completed++;
      if (s.active_exceptions_count > 0 || s.status === 'EXCEPTION') activeExceptions++;
    });

    return { total, onTrack, atRisk, delayed, activeExceptions, completed };
  }, [shipments]);

  // Selected item references
  const selectedShipment = selectedDetail?.shipment || shipments.find((s) => s.id === selectedShipmentId) || null;
  const trackingSummary = selectedDetail?.tracking_summary || null;
  const milestones = selectedDetail?.milestones || [];
  const exceptions = selectedDetail?.exceptions || [];
  const documents = selectedDetail?.documents || [];

  const selectedOrigin = parsePort(selectedShipment?.origin_port || 'INNSA');
  const selectedDest = parsePort(selectedShipment?.destination_port || 'NLRTM');

  const progressPct =
    trackingSummary?.progress_percentage ??
    selectedShipment?.progress_percentage ??
    (() => {
      switch (selectedShipment?.status) {
        case 'BOOKED':
          return 30;
        case 'DEPARTED':
          return 50;
        case 'IN_TRANSIT':
          return 70;
        case 'ARRIVED':
          return 90;
        case 'DELIVERED':
          return 100;
        default:
          return 60;
      }
    })();

  const trackingState =
    trackingSummary?.tracking_state ||
    selectedShipment?.tracking_state ||
    (selectedShipment?.status === 'DELIVERED'
      ? 'COMPLETED'
      : selectedShipment?.status === 'EXCEPTION'
        ? 'EXCEPTION'
        : 'ON_TRACK');

  // Robust Date Resolution
  const displayEtd =
    trackingSummary?.actual_etd ||
    trackingSummary?.planned_etd ||
    selectedShipment?.actual_departure ||
    selectedShipment?.planned_etd ||
    selectedShipment?.etd ||
    selectedShipment?.estimated_departure ||
    milestones.find((m) => m.milestone_code === 'DEPARTURE' || m.milestone_code === 'VESSEL_DEPARTURE')?.actual_date ||
    milestones.find((m) => m.milestone_code === 'DEPARTURE' || m.milestone_code === 'VESSEL_DEPARTURE')?.planned_date ||
    '2026-08-22';

  const displayEta =
    trackingSummary?.actual_arrival ||
    trackingSummary?.planned_eta ||
    selectedShipment?.actual_arrival ||
    selectedShipment?.planned_eta ||
    selectedShipment?.eta ||
    selectedShipment?.estimated_arrival ||
    milestones.find((m) => m.milestone_code === 'ARRIVAL' || m.milestone_code === 'VESSEL_ARRIVAL' || m.milestone_code === 'DISCHARGE')?.actual_date ||
    milestones.find((m) => m.milestone_code === 'ARRIVAL' || m.milestone_code === 'VESSEL_ARRIVAL' || m.milestone_code === 'DISCHARGE')?.planned_date ||
    '2026-08-28';

  // Pagination Slice
  const paginatedShipments = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filteredShipments.slice(start, start + pageSize);
  }, [filteredShipments, currentPage, pageSize]);

  const totalPages = Math.ceil(filteredShipments.length / pageSize) || 1;

  // Handle Add/Track Modal Submit
  const handleTrackSubmit = (e) => {
    e.preventDefault();
    if (!trackModalInput.trim()) {
      setTrackModalError('Please enter a booking number, shipment ID, or container reference.');
      return;
    }
    const clean = trackModalInput.trim().toLowerCase();
    const found = shipments.find(
      (s) =>
        `sh-${s.id}`.toLowerCase() === clean ||
        String(s.id) === clean ||
        (s.booking_number && s.booking_number.toLowerCase().includes(clean))
    );

    if (found) {
      setSelectedShipmentId(found.id);
      setShowAddTrackModal(false);
      setTrackModalInput('');
      setTrackModalError('');
    } else {
      setTrackModalError(`No active shipment found matching "${trackModalInput}". Try another reference.`);
    }
  };

  return (
    <div className="trk-page">
      {/* ── Top Header ──────────────────────────────────────────────────────── */}
      <div className="trk-header-row">
        <div>
          <h1 className="trk-header-title">Tracking & Live Fleet Intelligence</h1>
          <p className="trk-header-sub">
            Real-time visibility into your shipments, carrier milestones, and global container movements.
          </p>
        </div>
        <div className="trk-header-actions">
          {/* Top View Switcher */}
          <div className="trk-view-switcher">
            <button
              className={`trk-view-btn ${workspaceView === 'fleet' ? 'active' : ''}`}
              onClick={() => setWorkspaceView('fleet')}
            >
              <Radio size={13} /> Fleet Monitor
            </button>
            <button
              className={`trk-view-btn ${workspaceView === 'analytics' ? 'active' : ''}`}
              onClick={() => setWorkspaceView('analytics')}
            >
              <TrendingUp size={13} /> Performance & Analytics
            </button>
          </div>

          {workspaceView === 'fleet' && (
            <>
              <button
                className={`trk-btn-secondary ${justRefreshed ? 'trk-btn-refreshed' : ''}`}
                onClick={() => loadShipments(true)}
                disabled={refreshing}
                title="Refresh real-time tracking telemetry"
              >
                <RotateCw size={13} className={refreshing ? 'trk-spin' : ''} />
                <span>{refreshing ? 'Refreshing...' : justRefreshed ? '✓ Updated' : 'Refresh'}</span>
              </button>
              <button className="trk-btn-primary" onClick={() => setShowAddTrackModal(true)}>
                <Plus size={14} /> Add / Track Shipment
              </button>
            </>
          )}
        </div>
      </div>

      {workspaceView === 'analytics' ? (
        <TrackingAnalyticsWorkspace onNavigateShipment={(id) => navigate(`/dashboard/tracking/${id}`)} />
      ) : (
        <>
          {/* ── KPI Summary Cards (6 Cards) ─────────────────────────────────────── */}
          <div className="trk-kpi-grid">
            <button
              className={`trk-kpi-card ${trackingStateFilter === 'ALL' && statusFilter === 'ALL' ? 'active' : ''}`}
              onClick={() => {
                setTrackingStateFilter('ALL');
                setStatusFilter('ALL');
              }}
            >
              <div className="trk-kpi-top">
                <span className="trk-kpi-label">TOTAL TRACKED</span>
                <div className="trk-kpi-icon-wrap icon-blue">
                  <Compass size={14} />
                </div>
              </div>
              <div className="trk-kpi-val-row">
                <span className="trk-kpi-val">{kpis.total}</span>
                <span className="trk-kpi-sub">{kpis.total === 1 ? 'Active shipment' : 'Active shipments'}</span>
              </div>
              <div className="trk-kpi-footer">
                <span className={kpis.total > 0 ? 'trk-trend-up' : 'trk-trend-neutral'}>
                  {kpis.total > 0 ? '● Real-time telemetry' : 'No shipments tracked'}
                </span>
              </div>
            </button>

            <button
              className={`trk-kpi-card ${trackingStateFilter === 'ON_TRACK' ? 'active' : ''}`}
              onClick={() => setTrackingStateFilter(trackingStateFilter === 'ON_TRACK' ? 'ALL' : 'ON_TRACK')}
            >
              <div className="trk-kpi-top">
                <span className="trk-kpi-label">ON TRACK</span>
                <div className="trk-kpi-icon-wrap icon-green">
                  <CheckCircle2 size={14} />
                </div>
              </div>
              <div className="trk-kpi-val-row">
                <span className="trk-kpi-val">{kpis.onTrack}</span>
                <span className="trk-kpi-sub">
                  {kpis.total > 0 ? `${Math.round((kpis.onTrack / kpis.total) * 100)}% of total` : '0% of total'}
                </span>
              </div>
              <div className="trk-kpi-footer">
                <span className={kpis.onTrack > 0 ? 'trk-trend-up' : 'trk-trend-neutral'}>
                  {kpis.onTrack > 0 ? '✓ Normal progression' : '0 on-track loads'}
                </span>
              </div>
            </button>

            <button
              className={`trk-kpi-card ${trackingStateFilter === 'AT_RISK' ? 'active' : ''}`}
              onClick={() => setTrackingStateFilter(trackingStateFilter === 'AT_RISK' ? 'ALL' : 'AT_RISK')}
            >
              <div className="trk-kpi-top">
                <span className="trk-kpi-label">AT RISK</span>
                <div className="trk-kpi-icon-wrap icon-yellow">
                  <Clock size={14} />
                </div>
              </div>
              <div className="trk-kpi-val-row">
                <span className="trk-kpi-val">{kpis.atRisk}</span>
                <span className="trk-kpi-sub">
                  {kpis.total > 0 ? `${Math.round((kpis.atRisk / kpis.total) * 100)}% of total` : '0% of total'}
                </span>
              </div>
              <div className="trk-kpi-footer">
                <span className={kpis.atRisk > 0 ? 'trk-trend-neutral' : 'trk-trend-up'}>
                  {kpis.atRisk > 0 ? '⚠️ Requires monitoring' : '0 loads at risk'}
                </span>
              </div>
            </button>

            <button
              className={`trk-kpi-card ${trackingStateFilter === 'DELAYED' ? 'active' : ''}`}
              onClick={() => setTrackingStateFilter(trackingStateFilter === 'DELAYED' ? 'ALL' : 'DELAYED')}
            >
              <div className="trk-kpi-top">
                <span className="trk-kpi-label">DELAYED</span>
                <div className="trk-kpi-icon-wrap icon-red">
                  <AlertCircle size={14} />
                </div>
              </div>
              <div className="trk-kpi-val-row">
                <span className="trk-kpi-val">{kpis.delayed}</span>
                <span className="trk-kpi-sub">
                  {kpis.total > 0 ? `${Math.round((kpis.delayed / kpis.total) * 100)}% of total` : '0% of total'}
                </span>
              </div>
              <div className="trk-kpi-footer">
                <span className={kpis.delayed > 0 ? 'trk-trend-down' : 'trk-trend-up'}>
                  {kpis.delayed > 0 ? '🚨 ETA breached' : '0 delayed loads'}
                </span>
              </div>
            </button>

            <button
              className={`trk-kpi-card ${statusFilter === 'EXCEPTION' ? 'active' : ''}`}
              onClick={() => setStatusFilter(statusFilter === 'EXCEPTION' ? 'ALL' : 'EXCEPTION')}
            >
              <div className="trk-kpi-top">
                <span className="trk-kpi-label">ACTIVE EXCEPTIONS</span>
                <div className="trk-kpi-icon-wrap icon-purple">
                  <AlertTriangle size={14} />
                </div>
              </div>
              <div className="trk-kpi-val-row">
                <span className="trk-kpi-val">{kpis.activeExceptions}</span>
                <span className="trk-kpi-sub">Shipment holds</span>
              </div>
              <div className="trk-kpi-footer">
                <span className={kpis.activeExceptions > 0 ? 'trk-trend-down' : 'trk-trend-up'}>
                  {kpis.activeExceptions > 0 ? '⚡ Action required' : '0 active holds'}
                </span>
              </div>
            </button>

            <button
              className={`trk-kpi-card ${trackingStateFilter === 'COMPLETED' || statusFilter === 'DELIVERED' ? 'active' : ''}`}
              onClick={() => {
                setTrackingStateFilter('COMPLETED');
                setStatusFilter('DELIVERED');
              }}
            >
              <div className="trk-kpi-top">
                <span className="trk-kpi-label">COMPLETED</span>
                <div className="trk-kpi-icon-wrap icon-cyan">
                  <ShieldCheck size={14} />
                </div>
              </div>
              <div className="trk-kpi-val-row">
                <span className="trk-kpi-val">{kpis.completed}</span>
                <span className="trk-kpi-sub">Delivered</span>
              </div>
              <div className="trk-kpi-footer">
                <span className={kpis.completed > 0 ? 'trk-trend-up' : 'trk-trend-neutral'}>
                  {kpis.completed > 0 ? '✓ Successfully closed' : '0 delivered'}
                </span>
              </div>
            </button>
          </div>

      {/* ── Search & Filter Controls ────────────────────────────────────────── */}
      <div className="trk-filter-bar">
        <div className="trk-search-wrap">
          <Search size={14} className="trk-search-icon" />
          <input
            type="text"
            className="trk-search-input"
            placeholder="Search by shipment, booking, container, vessel, carrier..."
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
              setCurrentPage(1);
            }}
          />
          {searchQuery && (
            <button className="trk-search-clear" onClick={() => setSearchQuery('')}>
              <X size={13} />
            </button>
          )}
        </div>

        <div className="trk-filters-group">
          <CustomDropdown
            label="Status"
            value={statusFilter}
            onChange={(val) => {
              setStatusFilter(val);
              setCurrentPage(1);
            }}
            options={[
              { value: 'ALL', label: 'All Statuses' },
              { value: 'BOOKED', label: 'Booked' },
              { value: 'DEPARTED', label: 'Departed' },
              { value: 'IN_TRANSIT', label: 'In Transit' },
              { value: 'ARRIVED', label: 'Arrived' },
              { value: 'DELIVERED', label: 'Delivered' },
              { value: 'EXCEPTION', label: 'Exception' },
            ]}
            size="small"
          />

          <CustomDropdown
            label="Tracking Health"
            value={trackingStateFilter}
            onChange={(val) => {
              setTrackingStateFilter(val);
              setCurrentPage(1);
            }}
            options={[
              { value: 'ALL', label: 'All States' },
              { value: 'ON_TRACK', label: 'On Track' },
              { value: 'AT_RISK', label: 'At Risk' },
              { value: 'DELAYED', label: 'Delayed' },
              { value: 'COMPLETED', label: 'Completed' },
            ]}
            size="small"
          />

          <CustomDropdown
            label="Carrier"
            value={carrierFilter}
            onChange={(val) => {
              setCarrierFilter(val);
              setCurrentPage(1);
            }}
            options={[
              { value: 'ALL', label: 'All Carriers' },
              { value: 'MAEU', label: 'MAEU (Maersk)' },
              { value: 'COSCO', label: 'COSCO' },
              { value: 'MSC', label: 'MSC' },
              { value: 'HMM', label: 'HMM' },
              { value: 'CMA-CGM', label: 'CMA CGM' },
            ]}
            size="small"
          />

          <CustomDropdown
            label="Date Range"
            value={dateRangeFilter}
            onChange={setDateRangeFilter}
            options={[
              { value: 'ALL', label: 'All Time' },
              { value: 'THIS_WEEK', label: 'Aug 9 - Aug 15, 2026' },
              { value: 'THIS_MONTH', label: 'August 2026' },
            ]}
            size="small"
          />
        </div>
      </div>

      {/* ── Error Banner ────────────────────────────────────────────────────── */}
      {error && (
        <div className="trk-error-banner">
          <AlertTriangle size={15} />
          <span>{error}</span>
          <button onClick={() => loadShipments(false)} className="trk-btn-retry">
            Retry
          </button>
        </div>
      )}

      {/* ── Main Workspace ─────────────────────────────────────────────────── */}
      {!loadingList && shipments.length === 0 ? (
        /* ── Full-Width Empty State for New User / Zero Shipments ── */
        <div className="trk-empty-workspace-card">
          <div className="trk-ew-hero">
            <div className="trk-ew-radar-wrap">
              <div className="trk-ew-radar-ring ring-3" />
              <div className="trk-ew-radar-ring ring-2" />
              <div className="trk-ew-radar-ring ring-1" />
              <div className="trk-ew-radar-sweep" />
              <div className="trk-ew-ship-icon">
                <Ship size={28} />
              </div>
            </div>

            <h3 className="trk-ew-title">No Active Fleet Shipments Tracked</h3>
            <p className="trk-ew-desc">
              Real-time vessel AIS positions, carrier milestone telemetry, predictive ETAs, and automated exception detection will appear here once your first shipment is booked.
            </p>

            <div className="trk-ew-btn-row">
              <button className="trk-btn-primary" onClick={() => navigate('/dashboard/shipments')}>
                <Plus size={14} /> Create First Shipment
              </button>
              <button className="trk-btn-secondary" onClick={() => setShowAddTrackModal(true)}>
                <Search size={14} /> Track by Reference / Booking #
              </button>
            </div>
          </div>

          <div className="trk-ew-features-grid">
            <div className="trk-ew-feat-card">
              <div className="trk-ew-feat-icon icon-blue">
                <Compass size={18} />
              </div>
              <div className="trk-ew-feat-content">
                <h4>Live AIS Vessel Positions</h4>
                <p>Track global GPS positions, origin-to-destination ocean routes, and live port coordinates.</p>
              </div>
            </div>

            <div className="trk-ew-feat-card">
              <div className="trk-ew-feat-icon icon-green">
                <Clock size={18} />
              </div>
              <div className="trk-ew-feat-content">
                <h4>Predictive ETA & Milestones</h4>
                <p>Automated carrier milestone synchronizations with weather routing and delay forecasts.</p>
              </div>
            </div>

            <div className="trk-ew-feat-card">
              <div className="trk-ew-feat-icon icon-purple">
                <AlertTriangle size={18} />
              </div>
              <div className="trk-ew-feat-content">
                <h4>Automated Exception Guard</h4>
                <p>Instant proactive alerts on demurrage risks, transshipment rollovers, and customs holds.</p>
              </div>
            </div>
          </div>
        </div>
      ) : (
        /* ── Populated 2-Column Fleet Command Workspace ── */
        <div className="trk-workspace-grid">
          {/* Left Column: Tracked Shipments List Table */}
          <div className="trk-table-card">
            <div className="trk-table-header-row">
              <h3 className="trk-table-title">Tracked Shipments ({filteredShipments.length})</h3>
              <div className="trk-table-actions">
                {(statusFilter !== 'ALL' || trackingStateFilter !== 'ALL' || carrierFilter !== 'ALL' || searchQuery) && (
                  <button
                    className="trk-btn-clear-filter"
                    onClick={() => {
                      setStatusFilter('ALL');
                      setTrackingStateFilter('ALL');
                      setCarrierFilter('ALL');
                      setSearchQuery('');
                    }}
                  >
                    <Filter size={12} /> Clear Filters
                  </button>
                )}
              </div>
            </div>

            <div className="trk-table-wrap">
              <table className="trk-table">
                <thead>
                  <tr>
                    <th style={{ width: 34 }}>⭐</th>
                    <th style={{ minWidth: 120 }}>SHIPMENT</th>
                    <th style={{ minWidth: 120 }}>ROUTE</th>
                    <th style={{ minWidth: 130 }}>CARRIER / VESSEL</th>
                    <th style={{ minWidth: 100 }}>PROGRESS</th>
                    <th style={{ minWidth: 90 }}>ETD</th>
                    <th style={{ minWidth: 90 }}>ETA</th>
                    <th style={{ minWidth: 90 }}>HEALTH</th>
                    <th style={{ minWidth: 44, textAlign: 'center' }}>EXC.</th>
                    <th className="th-actions" style={{ width: 44 }}>ACTIONS</th>
                  </tr>
                </thead>
                <tbody>
                  {loadingList ? (
                    [...Array(5)].map((_, i) => (
                      <tr key={i} className="trk-skel-row">
                        {[30, 110, 110, 120, 80, 70, 70, 70, 30, 30].map((w, j) => (
                          <td key={j}>
                            <span className="trk-skel-box" style={{ width: w }} />
                          </td>
                        ))}
                      </tr>
                    ))
                  ) : paginatedShipments.length === 0 ? (
                    <tr>
                      <td colSpan={10}>
                        <div className="trk-empty-box">
                          <Ship size={28} className="trk-empty-icon" />
                          <h4>No shipments match criteria</h4>
                          <p>Try clearing filters or search queries to view tracked shipments.</p>
                          <button
                            className="trk-btn-secondary"
                            style={{ marginTop: 8 }}
                            onClick={() => {
                              setStatusFilter('ALL');
                              setTrackingStateFilter('ALL');
                              setCarrierFilter('ALL');
                              setSearchQuery('');
                            }}
                          >
                            Reset Filters
                          </button>
                        </div>
                      </td>
                    </tr>
                  ) : (
                    paginatedShipments.map((s) => {
                      const isSelected = selectedShipmentId === s.id;
                      const isStarred = starredIds.includes(s.id);
                      const origin = parsePort(s.origin_port || 'INNSA');
                      const dest = parsePort(s.destination_port || 'NLRTM');
                      const itemProgress =
                        s.progress_percentage ??
                        (() => {
                          switch (s.status) {
                            case 'BOOKED':
                              return 30;
                            case 'DEPARTED':
                              return 50;
                            case 'IN_TRANSIT':
                              return 70;
                            case 'ARRIVED':
                              return 90;
                            case 'DELIVERED':
                              return 100;
                            default:
                              return 60;
                          }
                        })();
                      const itemState =
                        s.tracking_state ||
                        (s.status === 'DELIVERED' ? 'COMPLETED' : s.status === 'EXCEPTION' ? 'EXCEPTION' : 'ON_TRACK');

                      return (
                        <tr
                          key={s.id}
                          className={`trk-row ${isSelected ? 'trk-row-selected' : ''}`}
                          onClick={() => setSelectedShipmentId(s.id)}
                        >
                          {/* 1. Star / Favorite */}
                          <td onClick={(e) => toggleStar(s.id, e)}>
                            <button
                              className={`trk-star-btn ${isStarred ? 'trk-star-active' : ''}`}
                              title={isStarred ? 'Unstar shipment' : 'Star shipment'}
                            >
                              <Star size={13} fill={isStarred ? '#f59e0b' : 'transparent'} />
                            </button>
                          </td>

                          {/* 2. Shipment Reference */}
                          <td>
                            <div className="trk-cell-shipment">
                              <span className="trk-ship-ref">SH-{s.id}</span>
                              <span className="trk-ship-booking">{s.booking_number || `BK-QA-${s.id}78784`}</span>
                              <span className="trk-ship-type">Ocean Freight</span>
                            </div>
                          </td>

                          {/* 3. Route Corridor */}
                          <td>
                            <div className="trk-cell-route">
                              <div className="trk-route-line origin">
                                <span className="trk-route-dot origin" />
                                <span className="trk-port-code">{origin.code}</span>
                                <span className="trk-port-city">{origin.name}</span>
                              </div>
                              <div className="trk-route-conn">
                                <span className="trk-route-arrow">↓</span>
                              </div>
                              <div className="trk-route-line dest">
                                <span className="trk-route-dot dest" />
                                <span className="trk-port-code">{dest.code}</span>
                                <span className="trk-port-city">{dest.name}</span>
                              </div>
                            </div>
                          </td>

                          {/* 4. Carrier / Vessel */}
                          <td>
                            <div className="trk-cell-carrier">
                              <span className="trk-scac-badge">{s.carrier_scac || 'MAEU'}</span>
                              <span className="trk-vessel-name">{s.vessel_name || 'Maersk Mc-Kinney Moller'}</span>
                              <span className="trk-voyage-no">Voy: {s.voyage_number || '2601W'}</span>
                            </div>
                          </td>

                          {/* 5. Progress */}
                          <td>
                            <div className="trk-cell-progress">
                              <div className="trk-prog-pct-row">
                                <span className="trk-prog-pct">{itemProgress}%</span>
                                <span className="trk-prog-sub">
                                  {itemProgress >= 100 ? 'Delivered' : itemProgress >= 90 ? 'Arrived' : 'In Transit'}
                                </span>
                              </div>
                              <div className="trk-prog-track">
                                <div
                                  className="trk-prog-bar"
                                  style={{
                                    width: `${itemProgress}%`,
                                    background:
                                      itemState === 'DELAYED' || itemState === 'EXCEPTION'
                                        ? '#dc2626'
                                        : itemState === 'AT_RISK'
                                          ? '#d97706'
                                          : 'linear-gradient(90deg, #3b82f6, #2563eb)',
                                  }}
                                />
                              </div>
                            </div>
                          </td>

                          {/* 6. ETD */}
                          <td>
                            <div className="trk-cell-date">
                              <span className="trk-date-val">{formatDateOnly(s.actual_departure || s.planned_etd || s.etd || '2026-08-22')}</span>
                              <span className="trk-date-sub actual">✓ Actual</span>
                            </div>
                          </td>

                          {/* 7. ETA */}
                          <td>
                            <div className="trk-cell-date">
                              <span className="trk-date-val">{formatDateOnly(s.actual_arrival || s.planned_eta || s.eta || '2026-08-28')}</span>
                              <span className={`trk-date-sub ${itemState === 'DELAYED' ? 'delayed' : 'ontime'}`}>
                                {itemState === 'DELAYED' ? '+3 days' : '+0.1d late'}
                              </span>
                            </div>
                          </td>

                          {/* 8. Health Pill */}
                          <td>
                            <div className="trk-cell-health-wrap">
                              <span className={`trk-health-pill trk-health-${itemState.toLowerCase()}`}>
                                {itemState.replace('_', ' ')}
                              </span>
                            </div>
                          </td>

                          {/* 9. Exceptions Count */}
                          <td style={{ textAlign: 'center' }}>
                            <span
                              className={`trk-exc-badge ${
                                s.active_exceptions_count > 0 ? 'exc-active' : 'exc-none'
                              }`}
                            >
                              {s.active_exceptions_count || 0}
                            </span>
                          </td>

                          {/* 10. Row Actions */}
                          <td className="td-actions">
                            <RowActions
                              shipment={s}
                              onNavigate={navigate}
                            />
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination Controls */}
            {filteredShipments.length > 0 && (
              <div className="trk-pagination-row">
                <span className="trk-pag-info">
                  Showing {Math.min((currentPage - 1) * pageSize + 1, filteredShipments.length)} to{' '}
                  {Math.min(currentPage * pageSize, filteredShipments.length)} of {filteredShipments.length} shipments
                </span>
                <div className="trk-pag-buttons">
                  <button
                    className="trk-pag-btn"
                    onClick={() => setCurrentPage(1)}
                    disabled={currentPage <= 1}
                  >
                    «
                  </button>
                  <button
                    className="trk-pag-btn"
                    onClick={() => setCurrentPage((p) => Math.max(p - 1, 1))}
                    disabled={currentPage <= 1}
                  >
                    ‹
                  </button>
                  {[...Array(totalPages)].map((_, i) => (
                    <button
                      key={i + 1}
                      className={`trk-pag-btn ${currentPage === i + 1 ? 'active' : ''}`}
                      onClick={() => setCurrentPage(i + 1)}
                    >
                      {i + 1}
                    </button>
                  ))}
                  <button
                    className="trk-pag-btn"
                    onClick={() => setCurrentPage((p) => Math.min(p + 1, totalPages))}
                    disabled={currentPage >= totalPages}
                  >
                    ›
                  </button>
                  <button
                    className="trk-pag-btn"
                    onClick={() => setCurrentPage(totalPages)}
                    disabled={currentPage >= totalPages}
                  >
                    »
                  </button>
                </div>
              </div>
            )}
          </div>

        {/* Right Column: Selected Shipment Intelligence Workspace */}
        <div className="trk-detail-panel">
          {loadingDetail && !selectedShipment ? (
            <div className="trk-detail-skel">
              <div className="trk-skel-header" />
              <div className="trk-skel-body" />
            </div>
          ) : selectedShipment ? (
            <>
              {/* Selected Shipment Header Card */}
              <div className="trk-detail-header-card">
                <div className="trk-dh-top">
                  <div>
                    <span className="trk-dh-label">SELECTED SHIPMENT</span>
                    <div className="trk-dh-title-row">
                      <h2 className="trk-dh-title">SH-{selectedShipment.id}</h2>
                      <span className={`trk-health-pill trk-health-${trackingState.toLowerCase()}`}>
                        {trackingState.replace('_', ' ')}
                      </span>
                    </div>
                    <div className="trk-dh-sub">
                      Booking: <strong>{selectedShipment.booking_number || `BK-QA-${selectedShipment.id}78784`}</strong>
                    </div>
                    <div className="trk-dh-meta-line">
                      <span>{selectedShipment.carrier_scac || 'MAEU'}</span>
                      <span>•</span>
                      <span>{selectedShipment.vessel_name || 'Maersk Mc-Kinney Moller'}</span>
                      {selectedShipment.voyage_number && (
                        <>
                          <span>•</span>
                          <span>{selectedShipment.voyage_number}</span>
                        </>
                      )}
                    </div>
                  </div>
                </div>

                {/* Tabs Navigation */}
                <div className="trk-dh-tabs">
                  <button
                    className={`trk-tab-btn ${activeTab === 'overview' ? 'active' : ''}`}
                    onClick={() => setActiveTab('overview')}
                  >
                    Overview
                  </button>
                  <button
                    className={`trk-tab-btn ${activeTab === 'milestones' ? 'active' : ''}`}
                    onClick={() => setActiveTab('milestones')}
                  >
                    Milestones ({milestones.length || 5})
                  </button>
                  <button
                    className={`trk-tab-btn ${activeTab === 'events' ? 'active' : ''}`}
                    onClick={() => setActiveTab('events')}
                  >
                    Events
                  </button>
                  <button
                    className={`trk-tab-btn ${activeTab === 'exceptions' ? 'active' : ''}`}
                    onClick={() => setActiveTab('exceptions')}
                  >
                    Exceptions {exceptions.length > 0 && <span className="trk-tab-badge">{exceptions.length}</span>}
                  </button>
                  <button
                    className={`trk-tab-btn ${activeTab === 'documents' ? 'active' : ''}`}
                    onClick={() => setActiveTab('documents')}
                  >
                    Documents
                  </button>
                </div>
              </div>

              {/* Tab 1: Overview */}
              {activeTab === 'overview' && (
                <div className="trk-tab-content">
                  {/* Route Visualizer Card */}
                  <div className="trk-card trk-route-card">
                    <div className="trk-route-viz-header">
                      <div className="trk-rv-port left">
                        <span className="trk-rv-code">{selectedOrigin.code}</span>
                        <span className="trk-rv-name">{selectedOrigin.name || 'Nhava Sheva, India'}</span>
                      </div>
                      <div className="trk-rv-port right">
                        <span className="trk-rv-code">{selectedDest.code}</span>
                        <span className="trk-rv-name">{selectedDest.name || 'Rotterdam, Netherlands'}</span>
                      </div>
                    </div>

                    <div className="trk-rv-track-container">
                      <div className="trk-rv-line-bg" />
                      <div className="trk-rv-line-fill" style={{ width: `${progressPct}%` }} />
                      <div
                        className="trk-rv-ship-marker"
                        style={{ left: `${Math.min(Math.max(progressPct, 5), 95)}%` }}
                      >
                        <Ship size={13} />
                      </div>
                    </div>

                    <div className="trk-rv-dates-grid">
                      <div className="trk-rv-date-unit">
                        <span className="trk-rv-date-lbl">ETD</span>
                        <span className="trk-rv-date-val">{formatDateOnly(displayEtd)}</span>
                        <span className="trk-rv-date-tag actual">Actual</span>
                      </div>
                      <div className="trk-rv-date-unit center">
                        <span className="trk-rv-date-lbl">ETA</span>
                        <span className="trk-rv-date-val">{formatDateOnly(displayEta)}</span>
                        <span className="trk-rv-date-tag late">+0.1 day</span>
                      </div>
                      <div className="trk-rv-date-unit right">
                        <span className="trk-rv-date-lbl">Transit Time</span>
                        <span className="trk-rv-date-val">6 days</span>
                        <span className="trk-rv-date-tag ocean">Direct Ocean</span>
                      </div>
                    </div>
                  </div>

                  {/* Tracking Progress Card */}
                  <div className="trk-card trk-progress-card">
                    <div className="trk-pc-header">
                      <span className="trk-pc-title">Tracking Progress</span>
                    </div>
                    <div className="trk-pc-val-row">
                      <span className="trk-pc-pct">{progressPct}%</span>
                      <span className="trk-pc-sub">Journey Completed</span>
                    </div>
                    <div className="trk-pc-bar-track">
                      <div className="trk-pc-bar-fill" style={{ width: `${progressPct}%` }} />
                    </div>

                    <div className="trk-pc-status-box">
                      <div className="trk-pc-status-item">
                        <span className="trk-pc-item-lbl">Current Status</span>
                        <div className="trk-pc-item-val">
                          <span className="trk-live-dot" />
                          <strong>
                            {progressPct >= 100
                              ? 'Delivered'
                              : progressPct >= 90
                                ? 'Arrived at POD'
                                : 'In Transit'}
                          </strong>
                        </div>
                        <span className="trk-pc-sub-desc">
                          {progressPct >= 100
                            ? 'Cargo handed over to consignee'
                            : 'Vessel arrived at destination port of discharge'}
                        </span>
                      </div>

                      <div className="trk-pc-status-item" style={{ marginTop: 10 }}>
                        <span className="trk-pc-item-lbl">Telemetry Source</span>
                        <div className="trk-pc-loc-val">
                          <span className="trk-compact-source-badge">◐ DEMO SIMULATION</span>
                        </div>
                        <span className="trk-pc-sub-desc">Simulated AIS Corridor (Development Mode)</span>
                      </div>
                    </div>
                  </div>

                  {/* Quick Actions Card */}
                  <div className="trk-card trk-actions-card">
                    <span className="trk-card-title">Quick Actions</span>
                    <div className="trk-qa-container">
                      <div className="trk-qa-buttons-grid">
                        <button
                          className="trk-qa-btn secondary"
                          onClick={() => navigate(`/dashboard/shipments/${selectedShipment.id}`)}
                        >
                          <Eye size={13} /> View Shipment
                        </button>
                        <button
                          className="trk-qa-btn secondary"
                          onClick={() => {
                            if (selectedShipment.booking_id) {
                              navigate(`/dashboard/bookings/${selectedShipment.booking_id}`);
                            } else {
                              navigate(`/dashboard/bookings`);
                            }
                          }}
                        >
                          <ExternalLink size={13} /> View Booking
                        </button>
                      </div>
                      <button
                        className="trk-qa-btn-full primary"
                        onClick={() => navigate(`/dashboard/tracking/${selectedShipment.id}`)}
                      >
                        <Activity size={14} /> View Full Tracking <ArrowUpRight size={14} />
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 2: Milestones */}
              {activeTab === 'milestones' && (
                <div className="trk-tab-content">
                  <div className="trk-card trk-milestones-card">
                    <div className="trk-card-header-row">
                      <span className="trk-card-title">Milestone Progress</span>
                      <span className="trk-badge-completed">5 of 5 Completed</span>
                    </div>

                    <div className="trk-ms-timeline">
                      {[
                        { code: 'BOOKED', label: 'Booking Confirmed', desc: 'Carrier booking confirmed by shipping line', date: 'Aug 20, 2026' },
                        { code: 'GATE_IN', label: 'Origin Gate In', desc: 'Container arrived at Nhava Sheva Terminal', date: 'Aug 21, 2026' },
                        { code: 'DEPARTED', label: 'Vessel Departed', desc: 'Vessel departed Nhava Sheva Port (INNSA)', date: 'Aug 22, 2026' },
                        { code: 'ARRIVED', label: 'Arrived at POD', desc: 'Vessel berthed at Rotterdam Port (NLRTM)', date: 'Aug 28, 2026' },
                        { code: 'DELIVERED', label: 'Cargo Delivered', desc: 'Final container delivery completed', date: 'Aug 28, 2026' },
                      ].map((m, idx) => (
                        <div key={m.code} className="trk-ms-item completed">
                          <div className="trk-ms-node">
                            <CheckCircle2 size={15} />
                            {idx < 4 && <div className="trk-ms-line" />}
                          </div>
                          <div className="trk-ms-info">
                            <div className="trk-ms-name-row">
                              <span className="trk-ms-name">{m.label}</span>
                              <span className="trk-ms-date">{m.date}</span>
                            </div>
                            <span className="trk-ms-desc">{m.desc}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 3: Events (Telemetry Log) */}
              {activeTab === 'events' && (
                <div className="trk-tab-content">
                  <div className="trk-card">
                    <span className="trk-card-title">Telemetry & AIS Event History</span>
                    <div className="trk-events-list">
                      {[
                        { time: 'Aug 28, 2026 · 05:30 PM', title: 'Container Discharged at Terminal', desc: 'Discharge completed at ECT Delta Terminal Rotterdam', icon: Anchor },
                        { time: 'Aug 28, 2026 · 11:28 AM', title: 'Vessel Berthed at POD', desc: 'Maersk Mc-Kinney Moller berthed at Berth 4', icon: Ship },
                        { time: 'Aug 26, 2026 · 02:15 PM', title: 'AIS Waypoint Passage: English Channel', desc: 'Vessel speed 18.2 kts, heading 042°', icon: Compass },
                        { time: 'Aug 22, 2026 · 05:00 PM', title: 'Vessel Departed Origin Port', desc: 'Departed Jawaharlal Nehru Port Trust (Nhava Sheva)', icon: Navigation },
                        { time: 'Aug 20, 2026 · 01:00 PM', title: 'Carrier Booking Confirmed', desc: 'EDI 301 confirmation received from MAEU', icon: CheckCircle2 },
                      ].map((ev, i) => (
                        <div key={i} className="trk-event-item">
                          <div className="trk-event-icon">
                            <ev.icon size={13} />
                          </div>
                          <div className="trk-event-content">
                            <div className="trk-event-title-row">
                              <span className="trk-event-title">{ev.title}</span>
                              <span className="trk-event-time">{ev.time}</span>
                            </div>
                            <span className="trk-event-desc">{ev.desc}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 4: Exceptions */}
              {activeTab === 'exceptions' && (
                <div className="trk-tab-content">
                  <div className="trk-card">
                    <span className="trk-card-title">Operational Exceptions & Risk Audit</span>
                    {exceptions.length > 0 ? (
                      <div className="trk-exc-list">
                        {exceptions.map((exc) => (
                          <div key={exc.id} className="trk-exc-item">
                            <AlertTriangle size={15} className="trk-exc-icon" />
                            <div className="trk-exc-body">
                              <div className="trk-exc-head">
                                <span className="trk-exc-type">{exc.exception_type}</span>
                                <span className={`trk-exc-sev sev-${exc.severity?.toLowerCase()}`}>
                                  {exc.severity}
                                </span>
                              </div>
                              <p className="trk-exc-desc">{exc.description}</p>
                              <span className="trk-exc-time">Raised: {formatDateTime(exc.created_at)}</span>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="trk-positive-state">
                        <ShieldCheck size={28} className="trk-pos-icon" />
                        <h4>No active operational risks detected</h4>
                        <p>All carrier milestones and schedule telemetry are executing within expected limits.</p>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Tab 5: Documents */}
              {activeTab === 'documents' && (
                <div className="trk-tab-content">
                  <div className="trk-card">
                    <span className="trk-card-title">Linked Shipment Documents</span>
                    <div className="trk-docs-list">
                      {[
                        { title: 'Master Bill of Lading (MBL)', code: 'MBL-MAEU-9801', status: 'VERIFIED' },
                        { title: 'Commercial Invoice & Packing List', code: 'INV-2026-0881', status: 'VERIFIED' },
                        { title: 'Certificate of Origin', code: 'COO-IN-2026', status: 'VERIFIED' },
                        { title: 'Delivery Order (DO)', code: 'DO-RTM-4410', status: 'APPROVED' },
                      ].map((doc, i) => (
                        <div key={i} className="trk-doc-row">
                          <FileText size={14} className="trk-doc-icon" />
                          <div className="trk-doc-info">
                            <span className="trk-doc-title">{doc.title}</span>
                            <span className="trk-doc-code">{doc.code}</span>
                          </div>
                          <span className="trk-doc-badge">{doc.status}</span>
                        </div>
                      ))}
                    </div>
                    <button
                      className="trk-btn-view-docs"
                      onClick={() => navigate(`/dashboard/shipments/${selectedShipment.id}`)}
                    >
                      Open Full Document Workspace →
                    </button>
                  </div>
                </div>
              )}
            </>
          ) : (
            <div className="trk-empty-detail">
              <Compass size={32} />
              <h4>No Shipment Selected</h4>
              <p>Select a shipment from the list on the left to inspect its live fleet tracking telemetry.</p>
            </div>
          )}
        </div>
      </div>
      )}
    </>
  )}

  {/* ── Modal: Add / Track Shipment ──────────────────────────────────────── */}
      {showAddTrackModal && (
        <div className="trk-modal-overlay" onClick={() => setShowAddTrackModal(false)}>
          <div className="trk-modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="trk-modal-header">
              <div className="trk-modal-title-wrap">
                <Ship size={18} className="trk-modal-ship-icon" />
                <h3>Add / Track Shipment Telemetry</h3>
              </div>
              <button className="trk-modal-close" onClick={() => setShowAddTrackModal(false)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleTrackSubmit} className="trk-modal-form">
              <p className="trk-modal-desc">
                Enter an existing Booking Reference (e.g. <code>BK-QA-178740980</code>), Shipment ID (e.g. <code>SH-26</code>), or Container Number to activate live route tracking.
              </p>

              <div className="trk-form-group">
                <label>Shipment Reference / Booking # / Container #</label>
                <input
                  type="text"
                  placeholder="e.g. SH-26 or BK-QA-178740980"
                  value={trackModalInput}
                  onChange={(e) => {
                    setTrackModalInput(e.target.value);
                    setTrackModalError('');
                  }}
                  autoFocus
                />
                {trackModalError && <span className="trk-form-error">{trackModalError}</span>}
              </div>

              <div className="trk-modal-footer">
                <button
                  type="button"
                  className="trk-btn-cancel"
                  onClick={() => setShowAddTrackModal(false)}
                >
                  Cancel
                </button>
                <button type="submit" className="trk-btn-submit">
                  <Search size={14} /> Find & Track
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
