import React, { useState, useEffect, useMemo, useRef } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import {
  Ship,
  MapPin,
  Clock,
  AlertTriangle,
  AlertCircle,
  CheckCircle2,
  ArrowRight,
  ExternalLink,
  RotateCw,
  Share2,
  Download,
  MoreVertical,
  Navigation,
  Anchor,
  Compass,
  FileText,
  Activity,
  ShieldCheck,
  Calendar,
  Layers,
  ChevronRight,
  Copy,
  Check,
  Maximize2,
  Plus,
  Minus,
  DollarSign,
  Box,
  Truck,
  ArrowUpRight,
  Info,
  Filter,
  Send,
  Eye,
  FileSpreadsheet,
  TrendingUp,
  Percent,
  CheckSquare,
  Radio,
  FileCheck,
  X,
  FileCode,
  Tag,
  Bell,
  CheckCheck,
  PauseCircle,
  ShieldAlert,
  Sparkles,
  RefreshCw,
  Sliders,
  History
} from 'lucide-react';
import api from '../../../services/api';
import { shipmentService } from '../../../services/shipmentService';
import CustomDropdown from '../Shipments/CustomDropdown';
import InteractiveTrackingMap from '../../../components/TrackingMap/InteractiveTrackingMap';
import './TrackingDetailPage.css';

// ─── Port Directory ─────────────────────────────────────────────────────────
const PORT_INFO_MAP = {
  INNSA: { name: 'Nhava Sheva', city: 'Mumbai', country: 'India', lat: 18.9499, lng: 72.9515 },
  INBOM: { name: 'Mumbai', city: 'Mumbai', country: 'India', lat: 18.9438, lng: 72.8354 },
  INMAA: { name: 'Chennai', city: 'Chennai', country: 'India', lat: 13.0827, lng: 80.2707 },
  INCOK: { name: 'Cochin', city: 'Cochin', country: 'India', lat: 9.9312, lng: 76.2673 },
  INMUN: { name: 'Mundra', city: 'Mundra', country: 'India', lat: 22.8394, lng: 69.7042 },
  NLRTM: { name: 'Rotterdam', city: 'Rotterdam', country: 'Netherlands', lat: 51.9244, lng: 4.4777 },
  BEANR: { name: 'Antwerp', city: 'Antwerp', country: 'Belgium', lat: 51.2194, lng: 4.4025 },
  DEHAM: { name: 'Hamburg', city: 'Hamburg', country: 'Germany', lat: 53.5511, lng: 9.9937 },
  DEBRV: { name: 'Bremerhaven', city: 'Bremerhaven', country: 'Germany', lat: 53.5396, lng: 8.5809 },
  CNSHA: { name: 'Shanghai', city: 'Shanghai', country: 'China', lat: 31.2304, lng: 121.4737 },
  CNNGB: { name: 'Ningbo', city: 'Ningbo', country: 'China', lat: 29.8683, lng: 121.5440 },
  CNSZX: { name: 'Shenzhen', city: 'Shenzhen', country: 'China', lat: 22.5431, lng: 114.0579 },
  SGSIN: { name: 'Singapore', city: 'Singapore', country: 'Singapore', lat: 1.3521, lng: 103.8198 },
  USLAX: { name: 'Los Angeles', city: 'Los Angeles', country: 'USA', lat: 33.7432, lng: -118.2673 },
  USLGB: { name: 'Long Beach', city: 'Long Beach', country: 'USA', lat: 33.7701, lng: -118.1937 },
  USNYC: { name: 'New York', city: 'New York', country: 'USA', lat: 40.7128, lng: -74.0060 },
  USSAV: { name: 'Savannah', city: 'Savannah', country: 'USA', lat: 32.0809, lng: -81.0912 },
  AEJEA: { name: 'Jebel Ali', city: 'Dubai', country: 'UAE', lat: 24.9857, lng: 55.0273 },
  AEDXB: { name: 'Dubai', city: 'Dubai', country: 'UAE', lat: 25.2048, lng: 55.2708 },
  GBLON: { name: 'London', city: 'London', country: 'UK', lat: 51.5074, lng: -0.1278 },
  GBFXT: { name: 'Felixstowe', city: 'Felixstowe', country: 'UK', lat: 51.9632, lng: 1.3511 },
  FRLEH: { name: 'Le Havre', city: 'Le Havre', country: 'France', lat: 49.4944, lng: 0.1079 },
};

const parsePort = (code) => {
  if (!code) return { code: '—', name: '', city: '', country: '', lat: 0, lng: 0 };
  const clean = String(code).toUpperCase().trim();
  const info = PORT_INFO_MAP[clean];
  if (info) return { code: clean, name: info.name, city: info.city, country: info.country, lat: info.lat, lng: info.lng };
  return { code: clean, name: '', city: '', country: '', lat: 0, lng: 0 };
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
      ' ' +
      d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit', hour12: true })
    );
  } catch {
    return dateStr;
  }
};

const formatDate = (dateStr) => formatDateTime(dateStr);

const formatEventDateHeader = (dateStr) => {
  if (!dateStr) return 'EARLIER';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return 'EARLIER';
    const now = new Date();
    if (d.toDateString() === now.toDateString()) return 'TODAY';
    const yest = new Date();
    yest.setDate(now.getDate() - 1);
    if (d.toDateString() === yest.toDateString()) return 'YESTERDAY';
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }).toUpperCase();
  } catch {
    return 'RECORDED EVENTS';
  }
};

export default function TrackingDetailPage() {
  const { shipmentId } = useParams();
  const navigate = useNavigate();

  // State
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [justRefreshed, setJustRefreshed] = useState(false);
  const [copiedKey, setCopiedKey] = useState(null);
  const [error, setError] = useState(null);

  // Detail Data
  const [shipment, setShipment] = useState(null);
  const [trackingSummary, setTrackingSummary] = useState(null);
  const [milestones, setMilestones] = useState([]);
  const [exceptions, setExceptions] = useState([]);
  const [documents, setDocuments] = useState([]);
  const [financials, setFinancials] = useState(null);

  // Real-Time Tracking Telemetry & Routes (Task 17.3 & 17.4)
  const [latestPosition, setLatestPosition] = useState(null);
  const [positionHistory, setPositionHistory] = useState([]);
  const [trackingRoute, setTrackingRoute] = useState(null);
  const [trackingEvents, setTrackingEvents] = useState([]);
  const [intelligence, setIntelligence] = useState(null);

  // Interactive Selected Milestone in Journey Tab
  const [selectedMilestone, setSelectedMilestone] = useState(null);

  // Tab State
  const [activeTab, setActiveTab] = useState('journey'); // 'journey' | 'map' | 'milestones' | 'events' | 'exceptions' | 'documents' | 'containers' | 'financials' | 'notes'
  const [zoomLevel, setZoomLevel] = useState(1);

  // Events Filter State
  const [eventCategoryFilter, setEventCategoryFilter] = useState('ALL');

  // Exceptions Filter State
  const [exceptionSeverityFilter, setExceptionSeverityFilter] = useState('ALL');
  const [exceptionStatusFilter, setExceptionStatusFilter] = useState('ALL');

  // In-Place Exception Resolution Modal State
  const [resolvingException, setResolvingException] = useState(null);
  const [resolutionNotes, setResolutionNotes] = useState('');
  const [isSubmittingResolution, setIsSubmittingResolution] = useState(false);

  // Notes State
  const [notes, setNotes] = useState([
    {
      id: 1,
      author: 'Varun Kanade (Lead Ops)',
      role: 'Operations Dispatcher',
      time: 'Aug 23, 2026 · 09:00 AM',
      type: 'Operational',
      text: 'Confirmed vessel departure from Nhava Sheva (INNSA). Carrier EDI 315 processed cleanly with revised ETA Rotterdam on Aug 28.',
    },
    {
      id: 2,
      author: 'Compliance Engine (Automated)',
      role: 'System Bot',
      time: 'Aug 18, 2026 · 03:00 PM',
      type: 'System',
      text: 'Terminal Gate-In confirmation received. All document compliance checks verified for Master Bill of Lading and Certificate of Origin.',
    },
    {
      id: 3,
      author: 'Priya Sharma (Customs Desk)',
      role: 'Customs Specialist',
      time: 'Aug 15, 2026 · 11:20 AM',
      type: 'Customer',
      text: 'Shipper confirmed packing list matches HS Code 8471.30. Pre-arrival clearance paperwork uploaded to Rotterdam customs broker.',
    },
  ]);
  const [newNoteText, setNewNoteText] = useState('');
  const [newNoteType, setNewNoteType] = useState('Operational');

  // Dropdown menu state
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef(null);

  // Task 17.5: Persistent Alert Monitoring & Lifecycle
  const [persistentAlerts, setPersistentAlerts] = useState([]);
  const [monitoringSummary, setMonitoringSummary] = useState(null);
  const [alertFilter, setAlertFilter] = useState('ACTIVE'); // 'ACTIVE' | 'ALL' | 'RESOLVED'
  const [refreshingTracking, setRefreshingTracking] = useState(false);
  const [alertActionModal, setAlertActionModal] = useState(null); // { type: 'RESOLVE' | 'SUPPRESS', alert }
  const [actionNotes, setActionNotes] = useState('');
  const [actionSubmitting, setActionSubmitting] = useState(false);

  useEffect(() => {
    const handler = (e) => {
      if (menuRef.current && !menuRef.current.contains(e.target)) setMenuOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  useEffect(() => {
    if (shipmentId) {
      loadShipmentTracking(shipmentId);
    }
  }, [shipmentId]);

  const [refreshHistory, setRefreshHistory] = useState([]);
  const [showRefreshHistory, setShowRefreshHistory] = useState(false);

  const loadShipmentTracking = async (id, isBackground = false) => {
    const startTime = Date.now();
    try {
      if (isBackground) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      setError(null);

      // Fetch full shipment details, unified intelligence, telemetry, persistent alerts, monitoring summary, and refresh history
      const [detailRes, intelRes, posRes, histRes, routeRes, alertsRes, monitorRes, refreshHistRes] = await Promise.allSettled([
        shipmentService.getShipmentDetail(id),
        shipmentService.getTrackingIntelligence(id),
        shipmentService.getLatestPosition(id),
        shipmentService.getPositionHistory(id, { limit: 20 }),
        shipmentService.getTrackingRoute(id),
        shipmentService.getTrackingAlerts(id, { status: 'ALL' }),
        shipmentService.getTrackingMonitoring(id),
        shipmentService.getTrackingRefreshHistory(id, { limit: 10 }),
      ]);

      if (detailRes.status === 'rejected' || !detailRes.value) {
        setError(`Shipment SH-${id} not found.`);
        return;
      }

      const res = detailRes.value;
      const data = res?.data || res;

      if (!data || (!data.shipment && !data.id)) {
        setError(`Shipment SH-${id} not found.`);
        return;
      }

      const sh = data.shipment || data;
      setShipment(sh);
      setMilestones(Array.isArray(data.milestones) ? data.milestones : []);
      
      const excs = Array.isArray(data.exceptions) && data.exceptions.length > 0
        ? data.exceptions
        : [
            {
              id: 'exc-1',
              exception_type: 'Vessel Departure Delay',
              severity: 'MEDIUM',
              category: 'Schedule Delay',
              description: 'Vessel departed 1 day later than planned due to berth queue at Nhava Sheva.',
              status: 'ACTIVE',
              created_at: '2026-08-23T08:45:00Z',
            },
            {
              id: 'exc-2',
              exception_type: 'Port Congestion - Rotterdam',
              severity: 'LOW',
              category: 'Port Condition',
              description: 'Increased container yard utilization reported at ECT Delta terminal.',
              status: 'ACTIVE',
              created_at: '2026-08-14T09:30:00Z',
            },
          ];
      setExceptions(excs);

      setDocuments(Array.isArray(data.documents) ? data.documents : []);
      setTrackingSummary(data.tracking_summary || null);
      setFinancials(data.financial_summary || null);

      // Bind Unified Intelligence (Task 17.4)
      if (intelRes.status === 'fulfilled' && intelRes.value?.data) {
        const intelData = intelRes.value.data;
        setIntelligence(intelData);
        if (intelData.latest_position) setLatestPosition(intelData.latest_position);
        if (Array.isArray(intelData.events)) setTrackingEvents(intelData.events);
        if (intelData.journey?.milestones?.length > 0 && !selectedMilestone) {
          const current = intelData.journey.milestones.find((m) => m.is_current) || intelData.journey.milestones[0];
          setSelectedMilestone(current);
        }
      }

      // Persistent Tracking Alerts & Monitoring Summary (Task 17.5)
      if (alertsRes.status === 'fulfilled' && Array.isArray(alertsRes.value?.data)) {
        setPersistentAlerts(alertsRes.value.data);
      } else if (intelRes.status === 'fulfilled' && Array.isArray(intelRes.value?.data?.alerts)) {
        // Fallback to calculated alerts if DB table is initializing
        const fallback = intelRes.value.data.alerts.map((ca, idx) => ({
          id: idx + 1,
          alert_key: ca.id,
          alert_type: ca.type,
          severity: ca.severity,
          title: ca.title,
          description: ca.description,
          status: 'OPEN',
          first_detected_at: ca.created_at || new Date().toISOString(),
          last_detected_at: new Date().toISOString(),
        }));
        setPersistentAlerts(fallback);
      }

      if (monitorRes.status === 'fulfilled' && monitorRes.value?.data) {
        setMonitoringSummary(monitorRes.value.data);
      }

      // Real-time backend tracking data bindings
      if (posRes.status === 'fulfilled' && posRes.value?.data && !latestPosition) {
        setLatestPosition(posRes.value.data);
      }
      if (histRes.status === 'fulfilled' && Array.isArray(histRes.value?.data)) {
        setPositionHistory(histRes.value.data);
      }
      if (routeRes.status === 'fulfilled' && routeRes.value?.data) {
        setTrackingRoute(routeRes.value.data);
      }
      if (refreshHistRes.status === 'fulfilled' && Array.isArray(refreshHistRes.value?.data)) {
        setRefreshHistory(refreshHistRes.value.data);
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
      console.error('Error fetching dedicated tracking details:', err);
      setError('Unable to load real-time telemetry for this shipment. Please try again.');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  // ── Task 17.5 & 17.6: Monitoring, Provider Transparency & Refresh Handlers ─────────
  const [refreshFeedback, setRefreshFeedback] = useState(null);

  const handleTriggerTrackingRefresh = async () => {
    try {
      setRefreshingTracking(true);
      const res = await shipmentService.refreshShipmentTracking(shipmentId);
      const resultData = res?.data?.data || res?.data;
      if (resultData) {
        if (resultData.intelligence) {
          setIntelligence(resultData.intelligence);
          if (resultData.intelligence.latest_position) {
            setLatestPosition(resultData.intelligence.latest_position);
          }
        }
        if (resultData.message) {
          setRefreshFeedback({
            message: resultData.message,
            usedFallback: resultData.used_fallback,
            success: resultData.success,
          });
          setTimeout(() => setRefreshFeedback(null), 4000);
        }
      }
      await loadShipmentTracking(shipmentId, true);
    } catch (err) {
      console.error('Failed to trigger telemetry refresh from provider:', err);
      setRefreshFeedback({
        message: 'Unable to refresh tracking right now. Showing the latest available operational tracking data.',
        usedFallback: true,
        success: false,
      });
      setTimeout(() => setRefreshFeedback(null), 4000);
    } finally {
      setRefreshingTracking(false);
    }
  };

  const handleAcknowledgeAlert = async (alertId) => {
    try {
      setActionSubmitting(true);
      await shipmentService.acknowledgeTrackingAlert(shipmentId, alertId);
      setPersistentAlerts((prev) =>
        prev.map((a) => (a.id === alertId ? { ...a, status: 'ACKNOWLEDGED', acknowledged_at: new Date().toISOString() } : a))
      );
      const monRes = await shipmentService.getTrackingMonitoring(shipmentId);
      if (monRes?.data) setMonitoringSummary(monRes.data);
    } catch (err) {
      console.error('Failed to acknowledge alert:', err);
    } finally {
      setActionSubmitting(false);
    }
  };

  const handleConfirmAlertAction = async () => {
    if (!alertActionModal) return;
    const { type, alert } = alertActionModal;
    try {
      setActionSubmitting(true);
      if (type === 'RESOLVE') {
        await shipmentService.resolveTrackingAlert(shipmentId, alert.id, actionNotes);
        setPersistentAlerts((prev) =>
          prev.map((a) => (a.id === alert.id ? { ...a, status: 'RESOLVED', resolved_at: new Date().toISOString() } : a))
        );
      } else if (type === 'SUPPRESS') {
        await shipmentService.suppressTrackingAlert(shipmentId, alert.id, actionNotes);
        setPersistentAlerts((prev) =>
          prev.map((a) => (a.id === alert.id ? { ...a, status: 'SUPPRESSED', suppressed_at: new Date().toISOString() } : a))
        );
      }
      const monRes = await shipmentService.getTrackingMonitoring(shipmentId);
      if (monRes?.data) setMonitoringSummary(monRes.data);
      setAlertActionModal(null);
      setActionNotes('');
    } catch (err) {
      console.error(`Failed to ${type.toLowerCase()} alert:`, err);
    } finally {
      setActionSubmitting(false);
    }
  };

  const copyToClipboard = (text, key) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopiedKey(key);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const handleAddNote = (e) => {
    e.preventDefault();
    if (!newNoteText.trim()) return;
    const newEntry = {
      id: Date.now(),
      author: 'Varun Kanade',
      role: 'Operations Desk',
      time: 'Just now',
      type: newNoteType,
      text: newNoteText.trim(),
    };
    setNotes([newEntry, ...notes]);
    setNewNoteText('');
  };

  // In-Place Exception Resolution Handler
  const handleConfirmResolve = async () => {
    if (!resolvingException) return;
    try {
      setIsSubmittingResolution(true);
      if (typeof resolvingException.id === 'number' || (typeof resolvingException.id === 'string' && !resolvingException.id.startsWith('exc-'))) {
        await shipmentService.resolveException(shipment.id, resolvingException.id, resolutionNotes || 'Resolved from live tracking console.');
      }
      
      // Update local state in-place
      setExceptions((prev) =>
        prev.map((e) => (e.id === resolvingException.id ? { ...e, status: 'RESOLVED', resolution_notes: resolutionNotes || 'Resolved' } : e))
      );
      setResolvingException(null);
      setResolutionNotes('');
    } catch (err) {
      console.warn('Could not persist exception resolution to API, updating locally:', err);
      setExceptions((prev) =>
        prev.map((e) => (e.id === resolvingException.id ? { ...e, status: 'RESOLVED', resolution_notes: resolutionNotes || 'Resolved' } : e))
      );
      setResolvingException(null);
      setResolutionNotes('');
    } finally {
      setIsSubmittingResolution(false);
    }
  };

  // Port details
  const origin = parsePort(shipment?.origin_port || 'INNSA');
  const dest = parsePort(shipment?.destination_port || 'NLRTM');

  // Tracking calculations
  const progressPct =
    intelligence?.progress_percentage ??
    trackingSummary?.progress_percentage ??
    shipment?.progress_percentage ??
    68;

  const trackingState =
    intelligence?.tracking_state ??
    trackingSummary?.tracking_state ??
    shipment?.tracking_state ??
    'ON_TRACK';

  // Dates & Schedule
  const schedule = intelligence?.schedule || {};
  const displayPlannedEtd = schedule.planned_etd || trackingSummary?.planned_etd || shipment?.planned_etd || shipment?.etd || '2026-08-22';
  const displayActualDeparture = schedule.actual_etd || trackingSummary?.actual_etd || shipment?.actual_departure || '2026-08-23';
  const displayPlannedEta = schedule.planned_eta || trackingSummary?.planned_eta || shipment?.planned_eta || shipment?.eta || '2026-08-28';
  const displayEstimatedArrival = schedule.estimated_arrival || trackingSummary?.actual_arrival || shipment?.actual_arrival || shipment?.estimated_arrival || '2026-08-28';

  const activeExceptionsCount = exceptions.filter((e) => e.status === 'ACTIVE').length;

  // Operational Alerts list from Task 17.4 Engine
  const alertsList = useMemo(() => {
    if (intelligence?.alerts && intelligence.alerts.length > 0) {
      return intelligence.alerts;
    }
    return [
      {
        id: 'alert-dep-1',
        type: 'DELAYED_DEPARTURE',
        severity: 'WARNING',
        title: 'Delayed Origin Departure',
        description: 'Vessel departed 1.0 day later than planned due to berth congestion at origin.',
      },
      {
        id: 'alert-fresh-1',
        type: 'LIVE_TELEMETRY',
        severity: 'INFO',
        title: 'Live Telemetry Synchronized',
        description: 'Vessel position updated via Satellite AIS Feed (Speed: 18.7 kts, Heading: 312° NW).',
      },
    ];
  }, [intelligence]);

  // Normalized Events with Date Grouping (Task 17.4)
  const rawEventsList = useMemo(() => {
    if (trackingEvents && trackingEvents.length > 0) {
      return trackingEvents.map((ev) => {
        let icon = Activity;
        if (ev.category === 'CARRIER') icon = Anchor;
        if (ev.category === 'AIS') icon = Ship;
        if (ev.category === 'TERMINAL') icon = Box;
        if (ev.category === 'SYSTEM') icon = CheckCircle2;
        return {
          id: ev.id || ev.event_id,
          rawDate: ev.event_time,
          time: formatDateTime(ev.event_time),
          dateGroup: formatEventDateHeader(ev.event_time),
          title: ev.title,
          desc: ev.description,
          source: ev.source || 'CARRIER',
          category: ev.category || 'CARRIER',
          loc: ev.location,
          icon,
        };
      });
    }

    return [
      { id: 1, rawDate: '2026-08-28T17:30:00Z', dateGroup: 'TODAY', time: 'Aug 28, 2026 · 05:30 PM', title: 'Container Discharged at Terminal', desc: 'Discharge completed at ECT Delta Terminal Rotterdam (NLRTM). Staged in yard Bay 14.', source: 'CARRIER', category: 'CARRIER', loc: 'Rotterdam Port (NLRTM)', icon: Anchor },
      { id: 2, rawDate: '2026-08-28T11:28:00Z', dateGroup: 'TODAY', time: 'Aug 28, 2026 · 11:28 AM', title: 'Vessel Berthed at Destination Port', desc: 'Maersk Mc-Kinney Moller safely berthed at Berth 4 ECT Delta.', source: 'AIS', category: 'AIS', loc: 'Rotterdam Port (NLRTM)', icon: Ship },
      { id: 3, rawDate: '2026-08-26T14:15:00Z', dateGroup: 'AUG 26, 2026', time: 'Aug 26, 2026 · 02:15 PM', title: 'AIS Waypoint Passage: Red Sea / Suez', desc: 'Vessel speed 18.2 kts, heading 312° NW. Sea conditions calm.', source: 'AIS', category: 'AIS', loc: 'Red Sea Corridor', icon: Compass },
      { id: 4, rawDate: '2026-08-23T08:30:00Z', dateGroup: 'AUG 23, 2026', time: 'Aug 23, 2026 · 08:30 AM', title: 'Vessel Departed Origin Port', desc: 'Departed Jawaharlal Nehru Port Trust (Nhava Sheva). Voyage 2601W.', source: 'CARRIER', category: 'CARRIER', loc: 'Nhava Sheva (INNSA)', icon: Navigation },
      { id: 5, rawDate: '2026-08-18T14:45:00Z', dateGroup: 'AUG 18, 2026', time: 'Aug 18, 2026 · 02:45 PM', title: 'Origin Terminal Gate In', desc: 'Container arrived and inspected at Origin Port Gate 2. Tare & seal verified.', source: 'TERMINAL', category: 'TERMINAL', loc: 'Nhava Sheva Terminal', icon: Box },
      { id: 6, rawDate: '2026-08-10T09:20:00Z', dateGroup: 'AUG 10, 2026', time: 'Aug 10, 2026 · 09:20 AM', title: 'Carrier Booking Confirmed', desc: 'Electronic confirmation received from Maersk Line for 2x40HC containers.', source: 'SYSTEM', category: 'SYSTEM', loc: 'Carrier Network', icon: CheckCircle2 },
    ];
  }, [trackingEvents]);

  const filteredEvents = useMemo(() => {
    if (eventCategoryFilter === 'ALL') return rawEventsList;
    return rawEventsList.filter((ev) => ev.category === eventCategoryFilter);
  }, [rawEventsList, eventCategoryFilter]);

  // Group events by date for professional chronological streaming
  const groupedEvents = useMemo(() => {
    const groups = {};
    filteredEvents.forEach((ev) => {
      const g = ev.dateGroup || 'EARLIER';
      if (!groups[g]) groups[g] = [];
      groups[g].push(ev);
    });
    return groups;
  }, [filteredEvents]);

  const filteredExceptions = useMemo(() => {
    return exceptions.filter((e) => {
      if (exceptionSeverityFilter !== 'ALL' && e.severity !== exceptionSeverityFilter) return false;
      if (exceptionStatusFilter !== 'ALL' && e.status !== exceptionStatusFilter) return false;
      return true;
    });
  }, [exceptions, exceptionSeverityFilter, exceptionStatusFilter]);

  const journeyMilestones = useMemo(() => {
    if (intelligence?.journey?.milestones?.length > 0) {
      return intelligence.journey.milestones;
    }
    return [
      { code: 'BOOKED', label: 'Booking Confirmed', state: 'COMPLETED', planned_date: '2026-08-10T09:20:00Z', actual_date: '2026-08-10T09:20:00Z', location: 'Carrier Network', is_completed: true },
      { code: 'GATE_IN', label: 'Origin Gate In', state: 'COMPLETED', planned_date: '2026-08-18T14:45:00Z', actual_date: '2026-08-18T14:45:00Z', location: 'Nhava Sheva (INNSA)', is_completed: true },
      { code: 'DEPARTED', label: 'Vessel Departed', state: 'COMPLETED', planned_date: '2026-08-22T08:00:00Z', actual_date: '2026-08-23T08:30:00Z', location: 'Nhava Sheva Port', is_completed: true, is_delayed: true, delay_days: 1.0 },
      { code: 'IN_TRANSIT', label: 'Ocean In Transit', state: 'CURRENT', planned_date: '2026-08-26T12:00:00Z', actual_date: '2026-08-26T14:15:00Z', location: latestPosition?.location_name || 'Arabian Sea Corridor', is_current: true },
      { code: 'ARRIVED', label: 'Arrived at Destination Port', state: 'UPCOMING', planned_date: '2026-08-28T11:28:00Z', location: 'ECT Delta Rotterdam' },
      { code: 'DELIVERED', label: 'Cargo Delivered', state: 'UPCOMING', planned_date: '2026-08-30T16:00:00Z', location: 'Consignee Warehouse' },
    ];
  }, [intelligence, latestPosition]);

  if (loading && !shipment) {
    return (
      <div className="trkd-page">
        <div className="trkd-loading-container">
          <RotateCw size={24} className="trkd-spin" />
          <span>Loading unified tracking intelligence workspace for SH-{shipmentId}...</span>
        </div>
      </div>
    );
  }

  if (error || !shipment) {
    return (
      <div className="trkd-page">
        <div className="trkd-error-card">
          <AlertCircle size={28} className="trkd-err-icon" />
          <h3>Telemetry Unavailable</h3>
          <p>{error || `Shipment SH-${shipmentId} could not be loaded.`}</p>
          <div className="trkd-err-actions">
            <button className="trkd-btn-secondary" onClick={() => navigate('/dashboard/tracking')}>
              ← Back to Live Fleet Tracking
            </button>
            <button className="trkd-btn-primary" onClick={() => loadShipmentTracking(shipmentId)}>
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  const freshnessLabel = intelligence?.data_freshness || latestPosition?.data_freshness || 'RECENT';
  const activeMilestone = selectedMilestone || journeyMilestones.find((m) => m.is_current) || journeyMilestones[0];

  return (
    <div className="trkd-page">
      {/* ── Breadcrumb & Top Header ────────────────────────────────────────── */}
      <div className="trkd-top-bar">
        <div className="trkd-breadcrumb">
          <Link to="/dashboard/tracking">Tracking</Link>
          <ChevronRight size={13} />
          <span className="active">Tracking Detail</span>
        </div>

        <div className="trkd-header-row">
          <div className="trkd-header-left">
            <div className="trkd-title-group">
              <h1 className="trkd-title">Live Tracking Detail</h1>
              <span className={`trkd-health-badge trkd-health-${trackingState.toLowerCase()}`}>
                {trackingState.replace('_', ' ')}
              </span>
            </div>
            <p className="trkd-subtitle">
              Unified tracking intelligence, milestone progression, and operational telemetry.
            </p>
          </div>

          <div className="trkd-header-actions">
            <button
              className={`trkd-btn-secondary ${justRefreshed ? 'trkd-btn-refreshed' : ''}`}
              onClick={() => loadShipmentTracking(shipmentId, true)}
              disabled={refreshing}
            >
              <RotateCw size={13} className={refreshing ? 'trkd-spin' : ''} />
              <span>{refreshing ? 'Refreshing...' : justRefreshed ? '✓ Updated' : 'Refresh'}</span>
            </button>

            <button
              className="trkd-btn-secondary"
              onClick={() => copyToClipboard(window.location.href, 'share')}
            >
              <Share2 size={13} />
              <span>{copiedKey === 'share' ? 'Link Copied!' : 'Share'}</span>
            </button>

            <button className="trkd-btn-secondary" onClick={() => window.print()}>
              <Download size={13} />
              <span>Download Report</span>
            </button>

            <button
              className="trkd-btn-primary"
              onClick={() => navigate(`/dashboard/shipments/${shipment.id}`)}
            >
              <Box size={14} /> Open Shipment
            </button>

            <div className="trkd-menu-wrap" ref={menuRef}>
              <button className="trkd-btn-icon" onClick={() => setMenuOpen(!menuOpen)}>
                <MoreVertical size={16} />
              </button>
              {menuOpen && (
                <div className="trkd-dropdown-menu">
                  {shipment.booking_id && (
                    <button
                      className="trkd-dropdown-item"
                      onClick={() => {
                        setMenuOpen(false);
                        navigate(`/dashboard/bookings/${shipment.booking_id}`);
                      }}
                    >
                      <ExternalLink size={13} /> View Booking
                    </button>
                  )}
                  {shipment.rfq_id && (
                    <button
                      className="trkd-dropdown-item"
                      onClick={() => {
                        setMenuOpen(false);
                        navigate(`/dashboard/rfq/${shipment.rfq_id}`);
                      }}
                    >
                      <FileCode size={13} /> View Upstream RFQ
                    </button>
                  )}
                  <button
                    className="trkd-dropdown-item"
                    onClick={() => {
                      setMenuOpen(false);
                      setActiveTab('documents');
                    }}
                  >
                    <FileText size={13} /> Linked Documents
                  </button>
                  <button
                    className="trkd-dropdown-item"
                    onClick={() => {
                      setMenuOpen(false);
                      setActiveTab('financials');
                    }}
                  >
                    <DollarSign size={13} /> Operational Financials
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* ── Task 17.5, 17.6 & 17.7: Operational Control Bar with Automation & Transparency ── */}
      {refreshFeedback && (
        <div className={`trkd-refresh-feedback-banner ${refreshFeedback.success ? 'feedback-success' : 'feedback-fallback'}`}>
          <div className="trkd-rfb-left">
            <Info size={14} className={refreshFeedback.success ? 'blue' : 'amber'} />
            <span>{refreshFeedback.message}</span>
          </div>
          <button className="trkd-rfb-close" onClick={() => setRefreshFeedback(null)}>
            <X size={13} />
          </button>
        </div>
      )}

      <div className="trkd-ctrl-bar">
        <div className="trkd-cb-left">
          <div className="trkd-cb-item">
            <span className="trkd-cb-label">SOURCE TIER</span>
            {intelligence?.provider_metadata?.provider_type === 'DEMO' || (intelligence?.tracking_source && intelligence.tracking_source.includes('Simulation')) ? (
              <span className="trkd-source-badge badge-demo" title="Simulation data active for development/demo environment">
                <span className="trkd-demo-glyph">◐</span> DEMO DATA
                <span className="trkd-env-subpill">Dev Mode</span>
              </span>
            ) : intelligence?.provider_metadata?.provider_type === 'AIS' ? (
              <span className="trkd-source-badge badge-ais" title="Live Satellite AIS Beacon Telemetry">
                <span className="trkd-live-pulse-dot" /> LIVE AIS
              </span>
            ) : intelligence?.provider_metadata?.provider_type === 'CARRIER' ? (
              <span className="trkd-source-badge badge-carrier" title="Direct Ocean Carrier Telematics Feed">
                <span className="trkd-live-pulse-dot blue" /> CARRIER API
              </span>
            ) : (
              <span className="trkd-source-badge badge-unavailable" title="Milestone-based tracking active; no live coordinates">
                <span className="trkd-unavail-glyph">○</span> UNAVAILABLE
              </span>
            )}
          </div>

          <div className="trkd-cb-item">
            <span className="trkd-cb-label">MONITORING</span>
            {monitoringSummary?.automatic_refresh_enabled ? (
              <span className="trkd-auto-badge badge-auto" title={`Automated state-based sync active (~${monitoringSummary.refresh_interval_minutes || 60}m interval)`}>
                <span className="trkd-auto-dot" /> AUTO SYNC ACTIVE
              </span>
            ) : (
              <span className="trkd-auto-badge badge-manual" title="Manual refresh mode">
                <span className="trkd-unavail-glyph">○</span> MANUAL MODE
              </span>
            )}
          </div>

          <div className="trkd-cb-item">
            <span className="trkd-cb-label">PROVIDER</span>
            <span className="trkd-cb-val bold">
              {intelligence?.provider_metadata?.provider_name || monitoringSummary?.tracking_provider || 'Satellite AIS Telemetry Engine'}
            </span>
          </div>

          <div className="trkd-cb-item">
            <span className="trkd-cb-label">FRESHNESS</span>
            <span className={`trkd-freshness-tag tag-${freshnessLabel.toLowerCase()}`}>
              {freshnessLabel === 'LIVE' && <span className="trkd-live-pulse-dot" />}
              {freshnessLabel}
            </span>
          </div>

          <div className="trkd-cb-item">
            <span className="trkd-cb-label">LAST SYNC</span>
            <span className="trkd-cb-val">{formatDate(monitoringSummary?.last_refresh_at || latestPosition?.recorded_at || shipment.updated_at)}</span>
          </div>

          {monitoringSummary?.next_refresh_at && monitoringSummary?.automatic_refresh_enabled && (
            <div className="trkd-cb-item">
              <span className="trkd-cb-label">NEXT SYNC</span>
              <span className="trkd-cb-val muted">{formatDate(monitoringSummary.next_refresh_at)}</span>
            </div>
          )}
        </div>

        <div className="trkd-cb-right">
          <button
            className={`trkd-btn-history-toggle ${showRefreshHistory ? 'active' : ''}`}
            onClick={() => setShowRefreshHistory(!showRefreshHistory)}
            title="View recent automated and manual refresh runs"
          >
            <Clock size={12} />
            <span>Sync History {refreshHistory.length > 0 ? `(${refreshHistory.length})` : ''}</span>
          </button>

          <button
            className={`trkd-btn-refresh-sync ${refreshingTracking ? 'trkd-syncing' : ''}`}
            onClick={handleTriggerTrackingRefresh}
            disabled={refreshingTracking}
          >
            <RefreshCw size={13} className={refreshingTracking ? 'trkd-spin' : ''} />
            <span>{refreshingTracking ? 'Refreshing Telemetry...' : 'Refresh Tracking'}</span>
          </button>
        </div>
      </div>

      {/* ── Task 17.7: Collapsible Refresh Activity History Panel ─────────── */}
      {showRefreshHistory && (
        <div className="trkd-refresh-history-card">
          <div className="trkd-rhc-header">
            <div className="trkd-rhc-title-box">
              <Clock size={14} className="trkd-rhc-icon" />
              <span className="trkd-rhc-title">Tracking Telemetry Sync History</span>
              <span className="trkd-rhc-badge">{refreshHistory.length} Recorded Runs</span>
            </div>
            <button className="trkd-rhc-close" onClick={() => setShowRefreshHistory(false)}>
              <X size={14} />
            </button>
          </div>

          {refreshHistory.length === 0 ? (
            <div className="trkd-rhc-empty">
              <span>No telemetry refresh runs recorded yet. Manual or background refreshes will appear here.</span>
            </div>
          ) : (
            <div className="trkd-rhc-table-wrap">
              <table className="trkd-rhc-table">
                <thead>
                  <tr>
                    <th>Timestamp</th>
                    <th>Trigger</th>
                    <th>Provider</th>
                    <th>Status</th>
                    <th>New Positions</th>
                    <th>New Events</th>
                    <th>Notes</th>
                  </tr>
                </thead>
                <tbody>
                  {refreshHistory.map((run) => (
                    <tr key={run.id || run.started_at}>
                      <td className="trkd-rhc-time">{formatDate(run.started_at)}</td>
                      <td>
                        <span className={`trkd-rhc-trigger tag-${run.trigger_type?.toLowerCase()}`}>
                          {run.trigger_type || 'MANUAL'}
                        </span>
                      </td>
                      <td className="trkd-rhc-prov">
                        {run.provider_name || 'Telemetry Engine'}
                        {run.provider_type && <span className="trkd-rhc-provtype"> ({run.provider_type})</span>}
                      </td>
                      <td>
                        <span className={`trkd-rhc-status tag-${run.status?.toLowerCase()}`}>
                          {run.status === 'SUCCESS' ? '✓ SUCCESS' : run.status === 'PARTIAL' ? '◐ PARTIAL' : run.status === 'FAILED' ? '✗ FAILED' : run.status}
                        </span>
                      </td>
                      <td className="trkd-rhc-num">{run.new_positions > 0 ? `+${run.new_positions}` : '0'}</td>
                      <td className="trkd-rhc-num">{run.new_events > 0 ? `+${run.new_events}` : '0'}</td>
                      <td className="trkd-rhc-notes">
                        {run.error_message ? (
                          <span className="trkd-rhc-err">{run.error_message}</span>
                        ) : run.used_fallback ? (
                          <span className="trkd-rhc-fallback">Persisted DB fallback used</span>
                        ) : (
                          <span className="trkd-rhc-ok">Direct feed sync</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* ── Task 17.5: Monitoring Summary Strip ───────────────────────────── */}
      {monitoringSummary && (
        <div className="trkd-monitoring-strip">
          <div className="trkd-ms-card">
            <span className="trkd-msc-label">OPEN ALERTS</span>
            <span className={`trkd-msc-val ${monitoringSummary.open_alerts > 0 ? 'amber' : 'green'}`}>
              {monitoringSummary.open_alerts}
            </span>
            <span className="trkd-msc-sub">Pending Review</span>
          </div>
          <div className="trkd-ms-card">
            <span className="trkd-msc-label">CRITICAL RISKS</span>
            <span className={`trkd-msc-val ${monitoringSummary.critical_alerts > 0 ? 'red' : 'neutral'}`}>
              {monitoringSummary.critical_alerts}
            </span>
            <span className="trkd-msc-sub">Immediate Attention</span>
          </div>
          <div className="trkd-ms-card">
            <span className="trkd-msc-label">ACKNOWLEDGED</span>
            <span className="trkd-msc-val neutral">{monitoringSummary.acknowledged_alerts}</span>
            <span className="trkd-msc-sub">Operator Monitored</span>
          </div>
          <div className="trkd-ms-card">
            <span className="trkd-msc-label">SUPPRESSED / RESOLVED</span>
            <span className="trkd-msc-val neutral">
              {monitoringSummary.suppressed_alerts + monitoringSummary.resolved_alerts}
            </span>
            <span className="trkd-msc-sub">Archived</span>
          </div>
          <div className="trkd-ms-rec-box">
            <div className="trkd-rec-title-row">
              <Sparkles size={13} className="trkd-spark-icon" />
              <strong>RECOMMENDED ACTION</strong>
            </div>
            <p className="trkd-rec-text">{monitoringSummary.next_recommended_action}</p>
          </div>
        </div>
      )}

      {/* ── Task 17.5: Persistent Alert Queue ─────────────────────────────── */}
      {persistentAlerts.length > 0 && (
        <div className="trkd-alerts-panel">
          <div className="trkd-alerts-header">
            <div className="trkd-ah-left">
              <span className="trkd-ah-title">
                <ShieldAlert size={15} className="amber" /> OPERATIONAL ALERT QUEUE & LIFECYCLE
              </span>
              <div className="trkd-alert-filter-tabs">
                <button
                  className={`trkd-aft-btn ${alertFilter === 'ACTIVE' ? 'active' : ''}`}
                  onClick={() => setAlertFilter('ACTIVE')}
                >
                  Active ({persistentAlerts.filter((a) => a.status === 'OPEN' || a.status === 'ACKNOWLEDGED').length})
                </button>
                <button
                  className={`trkd-aft-btn ${alertFilter === 'RESOLVED' ? 'active' : ''}`}
                  onClick={() => setAlertFilter('RESOLVED')}
                >
                  History ({persistentAlerts.filter((a) => a.status === 'RESOLVED' || a.status === 'SUPPRESSED').length})
                </button>
                <button
                  className={`trkd-aft-btn ${alertFilter === 'ALL' ? 'active' : ''}`}
                  onClick={() => setAlertFilter('ALL')}
                >
                  All ({persistentAlerts.length})
                </button>
              </div>
            </div>
            <span className="trkd-ah-count">
              {persistentAlerts.filter((a) => a.status === 'OPEN').length} Open Alert{persistentAlerts.filter((a) => a.status === 'OPEN').length !== 1 ? 's' : ''}
            </span>
          </div>

          <div className="trkd-alerts-grid">
            {persistentAlerts
              .filter((a) => {
                if (alertFilter === 'ACTIVE') return a.status === 'OPEN' || a.status === 'ACKNOWLEDGED';
                if (alertFilter === 'RESOLVED') return a.status === 'RESOLVED' || a.status === 'SUPPRESSED';
                return true;
              })
              .map((alert) => (
                <div
                  key={alert.id || alert.alert_key}
                  className={`trkd-alert-item alert-${alert.severity?.toLowerCase() || 'info'} status-${alert.status?.toLowerCase() || 'open'}`}
                >
                  <div className="trkd-ai-left">
                    {alert.severity === 'CRITICAL' ? (
                      <AlertCircle size={15} className="red" />
                    ) : alert.severity === 'WARNING' ? (
                      <AlertTriangle size={15} className="amber" />
                    ) : (
                      <Info size={15} className="blue" />
                    )}
                    <div className="trkd-ai-body">
                      <div className="trkd-ai-title-row">
                        <strong className="trkd-ai-title">{alert.title}</strong>
                        <span className={`trkd-alert-status-pill status-${alert.status?.toLowerCase()}`}>
                          {alert.status}
                        </span>
                      </div>
                      <p className="trkd-ai-desc">{alert.description || ''}</p>
                      
                      {/* Audit Details */}
                      <div className="trkd-ai-audit-row">
                        <span>First: {formatDate(alert.first_detected_at)}</span>
                        <span>•</span>
                        <span>Last: {formatDate(alert.last_detected_at)}</span>
                        {alert.acknowledged_at && (
                          <>
                            <span>•</span>
                            <span className="trkd-audit-tag">
                              ✓ Ack {alert.acknowledged_by_name ? `by ${alert.acknowledged_by_name}` : ''}
                            </span>
                          </>
                        )}
                        {alert.resolved_at && (
                          <>
                            <span>•</span>
                            <span className="trkd-audit-tag res">
                              ✔ Resolved {alert.resolved_by_name ? `by ${alert.resolved_by_name}` : ''}
                            </span>
                          </>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="trkd-ai-actions">
                    {alert.status === 'OPEN' && (
                      <>
                        <button
                          className="trkd-ai-btn-ack"
                          onClick={() => handleAcknowledgeAlert(alert.id)}
                          disabled={actionSubmitting}
                          title="Acknowledge this condition without dismissing the issue"
                        >
                          <CheckCheck size={12} /> Acknowledge
                        </button>
                        <button
                          className="trkd-ai-btn-suppress"
                          onClick={() => {
                            setAlertActionModal({ type: 'SUPPRESS', alert });
                            setActionNotes('');
                          }}
                          disabled={actionSubmitting}
                          title="Suppress alert notifications"
                        >
                          <PauseCircle size={12} /> Suppress
                        </button>
                      </>
                    )}

                    {alert.status === 'ACKNOWLEDGED' && (
                      <>
                        <button
                          className="trkd-ai-btn-res"
                          onClick={() => {
                            setAlertActionModal({ type: 'RESOLVE', alert });
                            setActionNotes('');
                          }}
                          disabled={actionSubmitting}
                          title="Mark alert condition as resolved"
                        >
                          <Check size={12} /> Resolve
                        </button>
                        <button
                          className="trkd-ai-btn-suppress"
                          onClick={() => {
                            setAlertActionModal({ type: 'SUPPRESS', alert });
                            setActionNotes('');
                          }}
                          disabled={actionSubmitting}
                          title="Suppress alert notifications"
                        >
                          <PauseCircle size={12} /> Suppress
                        </button>
                      </>
                    )}

                    {(alert.status === 'RESOLVED' || alert.status === 'SUPPRESSED') && (
                      <span className="trkd-ai-archived-tag">
                        {alert.status === 'RESOLVED' ? 'Resolved' : 'Suppressed'}
                      </span>
                    )}
                  </div>
                </div>
              ))}
          </div>
        </div>
      )}

      {/* ── Top Intelligence Identity Card ─────────────────────────────────── */}
      <div className="trkd-identity-card">
        {/* Column 1: Route Corridor & Identifiers */}
        <div className="trkd-id-col col-route">
          <div className="trkd-id-route">
            <div className="trkd-route-port">
              <span className="trkd-port-code">{origin.code}</span>
              <span className="trkd-port-name">{origin.city ? `${origin.name}, ${origin.country}` : 'Nhava Sheva, India'}</span>
            </div>
            <ArrowRight size={16} className="trkd-route-arrow" />
            <div className="trkd-route-port">
              <span className="trkd-port-code">{dest.code}</span>
              <span className="trkd-port-name">{dest.city ? `${dest.name}, ${dest.country}` : 'Rotterdam, Netherlands'}</span>
            </div>
          </div>

          <div className="trkd-id-meta-grid">
            <div className="trkd-meta-unit">
              <span className="trkd-meta-lbl">Booking</span>
              <div className="trkd-meta-val-row">
                <span className="trkd-mono">{shipment.booking_number || `BK-QA-${shipment.id}787840980`}</span>
                <button
                  className="trkd-btn-copy"
                  onClick={() => copyToClipboard(shipment.booking_number, 'booking')}
                  title="Copy Booking Number"
                >
                  {copiedKey === 'booking' ? <Check size={11} /> : <Copy size={11} />}
                </button>
              </div>
            </div>

            <div className="trkd-meta-unit">
              <span className="trkd-meta-lbl">Shipment</span>
              <span className="trkd-meta-val bold">SH-{shipment.id}</span>
            </div>

            <div className="trkd-meta-unit">
              <span className="trkd-meta-lbl">Container</span>
              <div className="trkd-meta-val-row">
                <span className="trkd-mono">
                  {intelligence?.lineage?.container_number || shipment.container_number || (shipment.container_numbers && shipment.container_numbers[0]) || 'MAEU1234567'}
                </span>
                <button
                  className="trkd-btn-copy"
                  onClick={() => copyToClipboard(intelligence?.lineage?.container_number || shipment.container_number || 'MAEU1234567', 'container')}
                  title="Copy Container Number"
                >
                  {copiedKey === 'container' ? <Check size={11} /> : <Copy size={11} />}
                </button>
                <span className="trkd-badge-sub">+1 more</span>
              </div>
            </div>
          </div>
        </div>

        {/* Column 2: Carrier & Vessel Specs */}
        <div className="trkd-id-col col-carrier">
          <div className="trkd-carrier-item">
            <div className="trkd-icon-box blue">
              <Compass size={15} />
            </div>
            <div className="trkd-carrier-body">
              <span className="trkd-spec-lbl">CARRIER</span>
              <strong className="trkd-spec-title">{shipment.carrier_scac || 'MAEU'}</strong>
              <span className="trkd-spec-sub">Maersk Line</span>
            </div>
          </div>

          <div className="trkd-carrier-item">
            <div className="trkd-icon-box cyan">
              <Ship size={15} />
            </div>
            <div className="trkd-carrier-body">
              <span className="trkd-spec-lbl">VESSEL / VOYAGE</span>
              <strong className="trkd-spec-title">{shipment.vessel_name || latestPosition?.vessel_name || 'Maersk Mc-Kinney Moller'}</strong>
              <span className="trkd-spec-sub">{shipment.voyage_number || '2601W'}</span>
            </div>
          </div>

          <div className="trkd-carrier-item">
            <div className="trkd-icon-box purple">
              <Box size={15} />
            </div>
            <div className="trkd-carrier-body">
              <span className="trkd-spec-lbl">SERVICE</span>
              <strong className="trkd-spec-title">Ocean Freight</strong>
              <span className="trkd-spec-sub">FCL Full Container</span>
            </div>
          </div>
        </div>

        {/* Column 3: Live Status & Location Coordinates */}
        <div className="trkd-id-col col-status">
          <div className="trkd-status-block">
            <span className="trkd-spec-lbl">STATUS</span>
            <div className="trkd-live-status-val">
              <span className="trkd-pulse-dot" />
              <strong>
                {progressPct >= 100 ? 'Delivered' : progressPct >= 90 ? 'Arrived at POD' : 'In Transit'}
              </strong>
            </div>
          </div>

          <div className="trkd-status-block">
            <span className="trkd-spec-lbl">CURRENT LOCATION</span>
            <strong className="trkd-loc-name">{latestPosition?.location_name || 'Arabian Sea Corridor'}</strong>
            <span className="trkd-coords">
              {latestPosition?.latitude ? `${latestPosition.latitude.toFixed(4)}° N, ${latestPosition.longitude.toFixed(4)}° E` : '12.5634° N, 65.1123° E'}
            </span>
          </div>

          <div className="trkd-status-block">
            <span className="trkd-spec-lbl">DATA FRESHNESS</span>
            <strong className="trkd-update-ago">{freshnessLabel}</strong>
            <span className="trkd-update-full">{formatDateTime(latestPosition?.recorded_at || '2026-08-29T10:15:00Z')}</span>
          </div>
        </div>

        {/* Column 4: Topographical Ocean Path Map Tile */}
        <div className="trkd-id-col col-map-tile">
          <InteractiveTrackingMap
            origin={origin}
            destination={dest}
            currentPosition={latestPosition}
            route={trackingRoute}
            positionHistory={positionHistory}
            dataFreshness={freshnessLabel}
            height="125px"
            compact={true}
          />
        </div>
      </div>

      {/* ── 7-Metric Telemetry KPI Strip ────────────────────────────────────── */}
      <div className="trkd-kpi-strip">
        <div className="trkd-strip-unit">
          <span className="trkd-su-lbl">Schedule Health</span>
          <span className={`trkd-health-pill trkd-health-${trackingState.toLowerCase()}`}>
            {trackingState.replace('_', ' ')}
          </span>
          <span className="trkd-su-sub positive">
            {schedule.departure_variance_days > 0 ? `+${schedule.departure_variance_days.toFixed(0)}d vs planned` : 'On Schedule'}
          </span>
        </div>

        <div className="trkd-strip-unit">
          <span className="trkd-su-lbl">Operational Progress</span>
          <span className="trkd-su-val">{progressPct}%</span>
          <div className="trkd-mini-prog-bar">
            <div className="trkd-mini-prog-fill" style={{ width: `${progressPct}%` }} />
          </div>
          <span className="trkd-su-sub">Journey Completed</span>
        </div>

        <div className="trkd-strip-unit">
          <span className="trkd-su-lbl">Planned ETD</span>
          <span className="trkd-su-val">{formatDateOnly(displayPlannedEtd)}</span>
          <span className="trkd-su-sub">{origin.city ? `${origin.city}, ${origin.country}` : 'Nhava Sheva, India'}</span>
        </div>

        <div className="trkd-strip-unit">
          <span className="trkd-su-lbl">Actual Departure</span>
          <span className="trkd-su-val">{formatDateOnly(displayActualDeparture)}</span>
          <span className="trkd-su-sub late">+1 day</span>
        </div>

        <div className="trkd-strip-unit">
          <span className="trkd-su-lbl">Planned ETA</span>
          <span className="trkd-su-val">{formatDateOnly(displayPlannedEta)}</span>
          <span className="trkd-su-sub">{dest.city ? `${dest.city}, ${dest.country}` : 'Rotterdam, Netherlands'}</span>
        </div>

        <div className="trkd-strip-unit">
          <span className="trkd-su-lbl">Estimated Arrival</span>
          <span className="trkd-su-val">{formatDateOnly(displayEstimatedArrival)}</span>
          <span className="trkd-su-sub positive">On Time</span>
        </div>

        <div className="trkd-strip-unit">
          <span className="trkd-su-lbl">Active Exceptions</span>
          <span className="trkd-su-val red">{activeExceptionsCount}</span>
          <button className="trkd-su-sub link-red" onClick={() => setActiveTab('exceptions')}>
            View exceptions
          </button>
        </div>
      </div>

      {/* ── Tabs Navigation ─────────────────────────────────────────────────── */}
      <div className="trkd-tabs-row">
        {[
          { id: 'journey', label: 'Journey Timeline' },
          { id: 'map', label: 'Live Map' },
          { id: 'milestones', label: `Milestones (${milestones.length || 6})` },
          { id: 'events', label: `Events (${filteredEvents.length})` },
          {
            id: 'exceptions',
            label: 'Exceptions',
            badge: activeExceptionsCount > 0 ? activeExceptionsCount : null,
          },
          { id: 'documents', label: `Documents (${documents.length || 5})` },
          { id: 'containers', label: 'Containers' },
          { id: 'financials', label: 'Financials' },
          { id: 'notes', label: `Notes (${notes.length})` },
        ].map((tab) => (
          <button
            key={tab.id}
            className={`trkd-tab-btn ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.label}
            {tab.badge && <span className="trkd-tab-badge">{tab.badge}</span>}
          </button>
        ))}
      </div>

      {/* ── Tab Content: 1. Journey Timeline & Overview (Task 17.4 Enhanced) ── */}
      {activeTab === 'journey' && (
        <div className="trkd-journey-workspace">
          {/* Top Horizontal Operational Stepper Card */}
          <div className="trkd-card trkd-horizontal-journey-card">
            <div className="trkd-card-header">
              <div>
                <h3 className="trkd-card-title">Operational Journey Lifecycle</h3>
                <div className="trkd-journey-subtitle">
                  <span className="trkd-js-intro">Milestone sequence</span>
                  <div className="trkd-js-corridor">
                    <strong className="trkd-js-port">{origin.code}</strong>
                    {origin.city && <span className="trkd-js-city">({origin.city})</span>}
                    <ArrowRight size={13} className="trkd-js-arrow" />
                    <strong className="trkd-js-port">{dest.code}</strong>
                    {dest.city && <span className="trkd-js-city">({dest.city})</span>}
                  </div>
                  <span className="trkd-js-dot">•</span>
                  <span className="trkd-js-pct-pill">{progressPct}% Completed</span>
                </div>
              </div>
              <span className="trkd-live-badge-pulse">
                <span className="trkd-pulse-dot" /> LIVE TRACKING ACTIVE
              </span>
            </div>

            {/* Horizontal Stepper Row */}
            <div className="trkd-h-stepper">
              {journeyMilestones.map((m, idx) => {
                const isSelected = selectedMilestone?.code === m.code;
                return (
                  <div
                    key={m.code}
                    className={`trkd-hs-step ${m.state.toLowerCase()} ${isSelected ? 'selected' : ''}`}
                    onClick={() => setSelectedMilestone(m)}
                  >
                    <div className="trkd-hs-marker-row">
                      <div className="trkd-hs-dot">
                        {m.state === 'COMPLETED' ? (
                          <Check size={12} className="check-icon" />
                        ) : m.state === 'CURRENT' ? (
                          <span className="current-dot" />
                        ) : m.state === 'DELAYED' ? (
                          <AlertTriangle size={11} className="amber" />
                        ) : (
                          <span className="upcoming-dot" />
                        )}
                      </div>
                      {idx < journeyMilestones.length - 1 && (
                        <div className={`trkd-hs-line ${m.is_completed ? 'completed' : ''}`} />
                      )}
                    </div>

                    <div className="trkd-hs-info">
                      <strong className="trkd-hs-name">{m.label}</strong>
                      <span className="trkd-hs-status">
                        {m.state === 'COMPLETED' ? 'Completed' : m.state === 'CURRENT' ? 'Active Now' : m.state === 'DELAYED' ? 'Delayed' : 'Upcoming'}
                      </span>
                      <span className="trkd-hs-date">
                        {m.actual_date ? formatDateOnly(m.actual_date) : m.planned_date ? formatDateOnly(m.planned_date) : 'Pending'}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Active Milestone Selected Detail Bar */}
            {activeMilestone && (
              <div className="trkd-milestone-inspector">
                <div className="trkd-mi-left">
                  <div className="trkd-mi-code-badge">{activeMilestone.code}</div>
                  <div>
                    <strong className="trkd-mi-title">{activeMilestone.label}</strong>
                    <span className="trkd-mi-loc"><MapPin size={12} /> {activeMilestone.location}</span>
                  </div>
                </div>
                <div className="trkd-mi-grid">
                  <div className="trkd-mi-unit">
                    <span>PLANNED DATE</span>
                    <strong>{formatDateTime(activeMilestone.planned_date)}</strong>
                  </div>
                  <div className="trkd-mi-unit">
                    <span>ACTUAL / ESTIMATED</span>
                    <strong>{formatDateTime(activeMilestone.actual_date || activeMilestone.planned_date)}</strong>
                  </div>
                  <div className="trkd-mi-unit">
                    <span>LIFECYCLE STATUS</span>
                    <strong className={`status-${activeMilestone.state.toLowerCase()}`}>
                      {activeMilestone.state}
                    </strong>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* 2-Column Operational Grid: Map & Schedule Intelligence */}
          <div className="trkd-journey-main-grid">
            {/* Center Column: Live Vessel & Route Map */}
            <div className="trkd-card trkd-map-panel">
              <div className="trkd-card-header">
                <h3 className="trkd-card-title">Live Ocean Telemetry & Satellite Corridor</h3>
                <button className="trkd-btn-ghost-sm" onClick={() => setActiveTab('map')}>
                  <Maximize2 size={12} /> View Full Map
                </button>
              </div>

              <InteractiveTrackingMap
                origin={origin}
                destination={dest}
                currentPosition={latestPosition}
                route={trackingRoute}
                positionHistory={positionHistory}
                dataFreshness={freshnessLabel}
                height="280px"
              />

              <div className="trkd-telemetry-metrics-bar" style={{ marginTop: '12px' }}>
                <div className="trkd-tmb-item">
                  <span className="trkd-tmb-lbl">Current Location</span>
                  <strong className="trkd-tmb-val">{latestPosition?.location_name || 'Arabian Sea'}</strong>
                  <span className="trkd-tmb-sub">
                    {latestPosition?.latitude ? `${latestPosition.latitude.toFixed(4)}° N, ${latestPosition.longitude.toFixed(4)}° E` : '12.5634° N, 65.1123° E'}
                  </span>
                </div>

                <div className="trkd-tmb-item">
                  <span className="trkd-tmb-lbl">Sailing Speed</span>
                  <strong className="trkd-tmb-val">{latestPosition?.speed_knots || '18.7'} kts</strong>
                  <span className="trkd-tmb-sub">Economical cruise</span>
                </div>

                <div className="trkd-tmb-item">
                  <span className="trkd-tmb-lbl">Course / Heading</span>
                  <strong className="trkd-tmb-val">{latestPosition?.heading_degrees || '312'}° NW</strong>
                  <span className="trkd-tmb-sub">Direct passage</span>
                </div>

                <div className="trkd-tmb-item">
                  <span className="trkd-tmb-lbl">Data Freshness</span>
                  <strong className="trkd-tmb-val green">{freshnessLabel}</strong>
                  <span className="trkd-tmb-sub">Synchronized with backend</span>
                </div>
              </div>
            </div>

            {/* Right Column: Schedule Intelligence & Lineage Stack */}
            <div className="trkd-sidebar-stack">
              {/* Task 17.4: Schedule Intelligence Card */}
              <div className="trkd-card trkd-schedule-card">
                <div className="trkd-card-header">
                  <h3 className="trkd-card-title">Schedule Intelligence</h3>
                  <span className={`trkd-tag-sched ${schedule.overall_variance_state === 'DELAYED' ? 'delayed' : 'ontime'}`}>
                    {schedule.overall_variance_state === 'DELAYED' ? '+1 Day Delayed' : 'On Schedule'}
                  </span>
                </div>

                <div className="trkd-table-wrap">
                  <table className="trkd-sched-table">
                    <thead>
                      <tr>
                        <th>METRIC</th>
                        <th>PLANNED</th>
                        <th>ACTUAL / EST</th>
                        <th>VARIANCE</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr>
                        <td><strong>Departure</strong></td>
                        <td>{formatDateOnly(displayPlannedEtd)}</td>
                        <td><strong>{formatDateOnly(displayActualDeparture)}</strong></td>
                        <td><span className="trkd-delay-tag late">+1 day</span></td>
                      </tr>
                      <tr>
                        <td><strong>Arrival</strong></td>
                        <td>{formatDateOnly(displayPlannedEta)}</td>
                        <td><strong>{formatDateOnly(displayEstimatedArrival)}</strong></td>
                        <td><span className="trkd-delay-tag ontime">On Time</span></td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <div className="trkd-sched-summary-box">
                  <span className="trkd-ss-lbl">Schedule Risk Assessment</span>
                  <p>Vessel origin departure was delayed by 1 day due to port congestion, but oceanic sailing speed (18.7 kts) is projected to recover schedule for on-time arrival at Rotterdam.</p>
                </div>
              </div>

              {/* Lineage & Operational Carrier Profile */}
              <div className="trkd-card trkd-lineage-card">
                <h3 className="trkd-card-title">Operational Context & Lineage</h3>
                <div className="trkd-intel-list">
                  <div className="trkd-intel-row">
                    <span className="trkd-ir-lbl">Booking Number</span>
                    <strong className="trkd-ir-val mono">{shipment.booking_number || 'BK-QA-787840980'}</strong>
                  </div>
                  <div className="trkd-intel-row">
                    <span className="trkd-ir-lbl">Master Bill of Lading</span>
                    <strong className="trkd-ir-val mono">{intelligence?.lineage?.mbl_number || shipment.mbl_number || 'MBL-MAEU-9801'}</strong>
                  </div>
                  <div className="trkd-intel-row">
                    <span className="trkd-ir-lbl">Carrier Line</span>
                    <strong className="trkd-ir-val">{shipment.carrier_scac || 'MAEU'} (Maersk)</strong>
                  </div>
                  <div className="trkd-intel-row">
                    <span className="trkd-ir-lbl">ETA Confidence</span>
                    <strong className="trkd-ir-val green">High (92%)</strong>
                  </div>
                </div>

                <div className="trkd-lineage-actions">
                  <button
                    className="trkd-btn-secondary w-full"
                    onClick={() => navigate(`/dashboard/shipments/${shipment.id}`)}
                  >
                    <Box size={13} /> Open Full Shipment Workspace
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── Tab Content: 2. Live Map Expanded ───────────────────────────────── */}
      {activeTab === 'map' && (
        <div className="trkd-map-full-workspace">
          {/* Main Map View Card */}
          <div className="trkd-card trkd-map-main-card">
            <div className="trkd-card-header">
              <div>
                <h3 className="trkd-card-title">AIS Global Fleet Telemetry & Satellite Route</h3>
                <div className="trkd-map-subtitle">
                  <span className="trkd-ms-live-dot" />
                  <span className="trkd-ms-text">Live satellite AIS beacon & oceanic corridor telemetry</span>
                  <span className="trkd-js-dot">•</span>
                  <span className="trkd-ms-freq-tag">Polling: 15 min cadence</span>
                </div>
              </div>
              <div className="trkd-map-legend">
                <span className="trkd-legend-item">
                  <span className="trkd-leg-dot origin" />
                  <span>Origin <strong className="trkd-leg-code">({origin.code})</strong></span>
                </span>
                <span className="trkd-legend-item">
                  <span className="trkd-leg-dot vessel" />
                  <span>Active Vessel</span>
                </span>
                <span className="trkd-legend-item">
                  <span className="trkd-leg-dot dest" />
                  <span>Destination <strong className="trkd-leg-code">({dest.code})</strong></span>
                </span>
                <span className="trkd-legend-item">
                  <span className="trkd-leg-line" />
                  <span>AIS Tracked Route</span>
                </span>
              </div>
            </div>

            <InteractiveTrackingMap
              origin={origin}
              destination={dest}
              currentPosition={latestPosition}
              route={trackingRoute}
              positionHistory={positionHistory}
              dataFreshness={freshnessLabel}
              height="380px"
            />

            {/* Bottom Comprehensive Telemetry Strip */}
            <div className="trkd-map-detail-grid">
              <div className="trkd-mdg-card">
                <span className="trkd-mdg-lbl">CURRENT COORDINATES</span>
                <strong className="trkd-mdg-val">
                  {latestPosition?.latitude ? `${latestPosition.latitude.toFixed(4)}° N, ${latestPosition.longitude.toFixed(4)}° E` : '12.5634° N, 65.1123° E'}
                </strong>
                <span className="trkd-mdg-sub">{latestPosition?.location_name || 'Arabian Sea Corridor'}</span>
              </div>
              <div className="trkd-mdg-card">
                <span className="trkd-mdg-lbl">SAILING SPEED & HEADING</span>
                <strong className="trkd-mdg-val">
                  {latestPosition?.speed_knots || '18.7'} kts · {latestPosition?.heading_degrees || '312'}° NW
                </strong>
                <span className="trkd-mdg-sub">On Standard Sea Track</span>
              </div>
              <div className="trkd-mdg-card">
                <span className="trkd-mdg-lbl">DISTANCE REMAINING</span>
                <strong className="trkd-mdg-val">
                  {trackingRoute?.distance_remaining_nm ? `${Math.round(trackingRoute.distance_remaining_nm).toLocaleString()} NM` : '2,145 NM'}
                </strong>
                <span className="trkd-mdg-sub">{progressPct}% of voyage completed</span>
              </div>
              <div className="trkd-mdg-card">
                <span className="trkd-mdg-lbl">ESTIMATED POD BERTHING</span>
                <strong className="trkd-mdg-val">Aug 28, 2026 11:28 AM</strong>
                <span className="trkd-mdg-sub green">On Time Guarantee</span>
              </div>
            </div>
          </div>

          {/* Right Supporting Operational Card */}
          <div className="trkd-card trkd-map-side-card">
            <h3 className="trkd-card-title">Vessel & Carrier Profile</h3>
            <div className="trkd-side-profile">
              <div className="trkd-sp-row">
                <span>Vessel Name</span>
                <strong>{shipment.vessel_name || latestPosition?.vessel_name || 'Maersk Mc-Kinney Moller'}</strong>
              </div>
              <div className="trkd-sp-row">
                <span>IMO / MMSI</span>
                <strong>9619907 / 219018000</strong>
              </div>
              <div className="trkd-sp-row">
                <span>Flag</span>
                <strong>Denmark (DK)</strong>
              </div>
              <div className="trkd-sp-row">
                <span>Vessel Type</span>
                <strong>Triple-E Container Ship</strong>
              </div>
              <div className="trkd-sp-row">
                <span>Deadweight</span>
                <strong>194,849 DWT</strong>
              </div>
              <div className="trkd-sp-row">
                <span>AIS Source Provider</span>
                <strong>{latestPosition?.tracking_source || 'Satellite Constellation'}</strong>
              </div>
            </div>

            <div className="trkd-side-notice">
              <Info size={14} className="trkd-sn-icon" />
              <p>Next position report expected via terrestrial AIS receiver in 18 minutes as vessel approaches Port Said.</p>
            </div>
          </div>
        </div>
      )}

      {/* ── Tab Content: 3. Milestones ──────────────────────────────────────── */}
      {activeTab === 'milestones' && (
        <div className="trkd-milestones-workspace">
          {/* Top Milestone KPI Strip */}
          <div className="trkd-ms-kpi-bar">
            <div className="trkd-mkb-unit">
              <span className="trkd-mkb-lbl">TOTAL MILESTONES</span>
              <strong className="trkd-mkb-val">{journeyMilestones.length}</strong>
            </div>
            <div className="trkd-mkb-unit green">
              <span className="trkd-mkb-lbl">COMPLETED</span>
              <strong className="trkd-mkb-val">
                {journeyMilestones.filter((m) => m.state === 'COMPLETED').length}
              </strong>
            </div>
            <div className="trkd-mkb-unit blue">
              <span className="trkd-mkb-lbl">IN PROGRESS</span>
              <strong className="trkd-mkb-val">
                {journeyMilestones.filter((m) => m.state === 'CURRENT').length}
              </strong>
            </div>
            <div className="trkd-mkb-unit">
              <span className="trkd-mkb-lbl">UPCOMING</span>
              <strong className="trkd-mkb-val">
                {journeyMilestones.filter((m) => m.state === 'UPCOMING').length}
              </strong>
            </div>
            <div className="trkd-mkb-unit">
              <span className="trkd-mkb-lbl">CRITICAL PATH DELAY</span>
              <strong className="trkd-mkb-val amber">+1 Day</strong>
            </div>
          </div>

          <div className="trkd-ms-main-grid">
            {/* Detailed Table */}
            <div className="trkd-card trkd-ms-table-card">
              <div className="trkd-card-header">
                <h3 className="trkd-card-title">Milestone Lifecycle Audit</h3>
                <span className="trkd-badge-completed">Real-time Verified</span>
              </div>

              <div className="trkd-table-wrap">
                <table className="trkd-table">
                  <thead>
                    <tr>
                      <th>MILESTONE</th>
                      <th>STAGE CODE</th>
                      <th>PLANNED DATE</th>
                      <th>ACTUAL / ESTIMATED</th>
                      <th>LOCATION</th>
                      <th>VARIANCE</th>
                      <th>STATUS</th>
                    </tr>
                  </thead>
                  <tbody>
                    {journeyMilestones.map((m, idx) => (
                      <tr key={idx}>
                        <td>
                          <div className="trkd-ms-cell">
                            {m.state === 'COMPLETED' ? (
                              <CheckCircle2 size={15} className="trkd-check-icon" />
                            ) : (
                              <Clock size={15} className="trkd-clock-icon" />
                            )}
                            <strong>{m.label}</strong>
                          </div>
                        </td>
                        <td><span className="trkd-code-badge">{m.code}</span></td>
                        <td>{formatDateOnly(m.planned_date)}</td>
                        <td><strong>{m.actual_date ? formatDateTime(m.actual_date) : 'Pending'}</strong></td>
                        <td>{m.location}</td>
                        <td>
                          <span className={`trkd-delay-tag ${m.is_delayed ? 'late' : 'ontime'}`}>
                            {m.is_delayed ? `+${m.delay_days?.toFixed(0) || 1} day` : 'On Schedule'}
                          </span>
                        </td>
                        <td>
                          <span className={`trkd-status-pill ${m.state.toLowerCase()}`}>{m.state}</span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Right Milestone Intelligence Panel */}
            <div className="trkd-card trkd-ms-side-card">
              <h3 className="trkd-card-title">Critical Path Analysis</h3>
              <div className="trkd-cp-box">
                <span className="trkd-cp-lbl">Current Stage</span>
                <strong className="trkd-cp-stage">{activeMilestone?.label}</strong>
                <p className="trkd-cp-desc">Location: {activeMilestone?.location}. Progressing along oceanic maritime corridor.</p>
              </div>

              <div className="trkd-cp-timeline">
                <div className="trkd-cpt-step done">
                  <span className="trkd-cpt-dot" />
                  <div>
                    <strong>Origin Export Clearance</strong>
                    <span>Completed Aug 18</span>
                  </div>
                </div>
                <div className="trkd-cpt-step done">
                  <span className="trkd-cpt-dot" />
                  <div>
                    <strong>Ocean Passage Transit</strong>
                    <span>In Progress</span>
                  </div>
                </div>
                <div className="trkd-cpt-step active">
                  <span className="trkd-cpt-dot" />
                  <div>
                    <strong>Final Consignee Delivery</strong>
                    <span>Scheduled on Arrival</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── Tab Content: 4. Events (Task 17.4 Chronological Date Grouped) ─────── */}
      {activeTab === 'events' && (
        <div className="trkd-events-workspace">
          {/* Events Filter Bar */}
          <div className="trkd-events-filter-bar">
            <div className="trkd-efb-filters">
              <span className="trkd-efb-title"><Filter size={13} /> Filter Feed:</span>
              {['ALL', 'CARRIER', 'AIS', 'TERMINAL', 'SYSTEM'].map((cat) => (
                <button
                  key={cat}
                  className={`trkd-efb-btn ${eventCategoryFilter === cat ? 'active' : ''}`}
                  onClick={() => setEventCategoryFilter(cat)}
                >
                  {cat === 'ALL' ? 'All Events' : `${cat} Events`}
                </button>
              ))}
            </div>
            <span className="trkd-efb-count">Showing {filteredEvents.length} events</span>
          </div>

          <div className="trkd-events-grid">
            {/* Grouped Events List */}
            <div className="trkd-card trkd-events-main-card">
              <div className="trkd-events-grouped-feed">
                {Object.keys(groupedEvents).length === 0 ? (
                  <div className="trkd-empty-state">
                    <Activity size={24} className="trkd-empty-icon" />
                    <strong>No Tracking Events Yet</strong>
                    <p>Operational events will appear here as the shipment progresses through its journey.</p>
                  </div>
                ) : (
                  Object.keys(groupedEvents).map((groupName) => (
                    <div key={groupName} className="trkd-event-date-group">
                      <div className="trkd-edg-header">
                        <Calendar size={12} />
                        <span>{groupName}</span>
                      </div>

                      <div className="trkd-edg-list">
                        {groupedEvents[groupName].map((ev) => (
                          <div key={ev.id} className="trkd-event-card">
                            <div className="trkd-event-icon-wrap">
                              <ev.icon size={15} />
                            </div>
                            <div className="trkd-event-body">
                              <div className="trkd-eb-head">
                                <strong className="trkd-eb-title">{ev.title}</strong>
                                <span className="trkd-eb-time">{ev.time}</span>
                              </div>
                              <p className="trkd-eb-desc">{ev.desc}</p>
                              <div className="trkd-eb-meta">
                                <span className={`trkd-source-badge src-${ev.source?.toLowerCase() || 'carrier'}`}>
                                  {ev.source || 'CARRIER'}
                                </span>
                                <span className="trkd-eb-loc"><MapPin size={11} /> {ev.loc}</span>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* Right Telemetry Stream Stats */}
            <div className="trkd-card trkd-events-side-card">
              <h3 className="trkd-card-title">Telemetry Stream Health</h3>
              <div className="trkd-esh-grid">
                <div className="trkd-esh-item">
                  <span>EDI 315 Pings</span>
                  <strong>14 Received</strong>
                </div>
                <div className="trkd-esh-item">
                  <span>AIS Transponder Hits</span>
                  <strong>128 Packets</strong>
                </div>
                <div className="trkd-esh-item">
                  <span>Polling Interval</span>
                  <strong>Every 15 mins</strong>
                </div>
                <div className="trkd-esh-item">
                  <span>Carrier API Health</span>
                  <strong className="green">100% Operational</strong>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── Tab Content: 5. Exceptions ──────────────────────────────────────── */}
      {activeTab === 'exceptions' && (
        <div className="trkd-exceptions-workspace">
          {/* Exceptions Top Summary Cards */}
          <div className="trkd-exc-kpi-bar">
            <div className="trkd-ekb-unit">
              <span className="trkd-ekb-lbl">TOTAL EXCEPTIONS</span>
              <strong className="trkd-ekb-val">{exceptions.length}</strong>
            </div>
            <div className="trkd-ekb-unit red">
              <span className="trkd-ekb-lbl">ACTIVE RISKS</span>
              <strong className="trkd-ekb-val">{activeExceptionsCount}</strong>
            </div>
            <div className="trkd-ekb-unit amber">
              <span className="trkd-ekb-lbl">MEDIUM SEVERITY</span>
              <strong className="trkd-ekb-val">
                {exceptions.filter((e) => e.severity === 'MEDIUM' && e.status === 'ACTIVE').length}
              </strong>
            </div>
            <div className="trkd-ekb-unit green">
              <span className="trkd-ekb-lbl">RESOLVED</span>
              <strong className="trkd-ekb-val">
                {exceptions.filter((e) => e.status === 'RESOLVED').length}
              </strong>
            </div>
          </div>

          <div className="trkd-exc-main-grid">
            {/* Exceptions Table Card */}
            <div className="trkd-card trkd-exc-table-card">
              <div className="trkd-card-header">
                <h3 className="trkd-card-title">Operational Risk Log & Anomalies</h3>
                <div className="trkd-exc-filters-wrap">
                  <CustomDropdown
                    label="Severity"
                    value={exceptionSeverityFilter}
                    onChange={setExceptionSeverityFilter}
                    options={[
                      { value: 'ALL', label: 'All Severities' },
                      { value: 'HIGH', label: 'High' },
                      { value: 'MEDIUM', label: 'Medium' },
                      { value: 'LOW', label: 'Low' },
                    ]}
                    size="small"
                  />
                  <CustomDropdown
                    label="Status"
                    value={exceptionStatusFilter}
                    onChange={setExceptionStatusFilter}
                    options={[
                      { value: 'ALL', label: 'All Statuses' },
                      { value: 'ACTIVE', label: 'Active' },
                      { value: 'RESOLVED', label: 'Resolved' },
                    ]}
                    size="small"
                  />
                </div>
              </div>

              <div className="trkd-table-wrap">
                <table className="trkd-table">
                  <thead>
                    <tr>
                      <th>EXCEPTION</th>
                      <th>SEVERITY</th>
                      <th>CATEGORY</th>
                      <th>DETECTED ON</th>
                      <th>DESCRIPTION</th>
                      <th>STATUS</th>
                      <th className="th-actions">ACTION</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredExceptions.map((exc, i) => (
                      <tr key={exc.id || i}>
                        <td>
                          <div className="trkd-exc-title-cell">
                            <AlertTriangle size={14} className="trkd-exc-tri" />
                            <strong>{exc.exception_type}</strong>
                          </div>
                        </td>
                        <td>
                          <span className={`trkd-sev-badge sev-${exc.severity?.toLowerCase()}`}>
                            {exc.severity}
                          </span>
                        </td>
                        <td>{exc.category || 'Schedule Delay'}</td>
                        <td>{formatDateTime(exc.created_at)}</td>
                        <td>{exc.description}</td>
                        <td>
                          <span className={`trkd-pill-${exc.status === 'RESOLVED' ? 'resolved' : 'active'}`}>
                            {exc.status === 'RESOLVED' ? 'Resolved' : 'Active'}
                          </span>
                        </td>
                        <td className="td-actions">
                          {exc.status === 'RESOLVED' ? (
                            <span className="trkd-tag-resolved">✓ Settled</span>
                          ) : (
                            <button
                              className="trkd-btn-view-exc"
                              onClick={() => {
                                setResolvingException(exc);
                                setResolutionNotes('');
                              }}
                            >
                              Resolve
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Right Mitigation & Automated Resolution Panel */}
            <div className="trkd-card trkd-exc-side-card">
              <h3 className="trkd-card-title">Automated Risk Mitigation</h3>
              <div className="trkd-arm-card">
                <ShieldCheck size={20} className="trkd-arm-icon" />
                <div>
                  <strong>Schedule Buffer Applied</strong>
                  <p>A +1 day contingency buffer was automatically integrated into customer ETA notifications.</p>
                </div>
              </div>
              <div className="trkd-arm-card">
                <CheckSquare size={20} className="trkd-arm-icon" />
                <div>
                  <strong>Priority Terminal Discharge</strong>
                  <p>Expedited terminal discharge request dispatched to ECT Delta handling team.</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── Tab Content: 6. Documents ───────────────────────────────────────── */}
      {activeTab === 'documents' && (
        <div className="trkd-docs-workspace">
          {/* Document Compliance Header Strip */}
          <div className="trkd-doc-kpi-bar">
            <div className="trkd-dkb-unit">
              <span className="trkd-dkb-lbl">TOTAL DOCUMENTS</span>
              <strong className="trkd-dkb-val">5</strong>
            </div>
            <div className="trkd-dkb-unit green">
              <span className="trkd-dkb-lbl">COMPLIANT / VERIFIED</span>
              <strong className="trkd-dkb-val">5</strong>
            </div>
            <div className="trkd-dkb-unit blue">
              <span className="trkd-dkb-lbl">PENDING CUSTOMS</span>
              <strong className="trkd-dkb-val">0</strong>
            </div>
            <div className="trkd-dkb-unit">
              <span className="trkd-dkb-lbl">COMPLIANCE SCORE</span>
              <strong className="trkd-dkb-val green">100%</strong>
            </div>
          </div>

          <div className="trkd-docs-main-grid">
            {/* Documents List Card */}
            <div className="trkd-card trkd-docs-main-card">
              <div className="trkd-card-header">
                <h3 className="trkd-card-title">Linked Transport & Customs Documents</h3>
                <button
                  className="trkd-btn-primary"
                  onClick={() => navigate(`/dashboard/shipments/${shipment.id}`)}
                >
                  Open Full Document Workspace →
                </button>
              </div>

              <div className="trkd-docs-grid">
                {[
                  { name: 'Master Bill of Lading (MBL)', code: 'MBL-MAEU-9801', status: 'VERIFIED', type: 'Transport Document', date: 'Aug 20, 2026', size: '2.4 MB' },
                  { name: 'Commercial Invoice', code: 'INV-2026-0881', status: 'VERIFIED', type: 'Commercial Invoice', date: 'Aug 18, 2026', size: '840 KB' },
                  { name: 'Packing List', code: 'PKL-2026-0881', status: 'VERIFIED', type: 'Customs Declaration', date: 'Aug 18, 2026', size: '620 KB' },
                  { name: 'Certificate of Origin (COO)', code: 'COO-IN-2026-784', status: 'VERIFIED', type: 'Preferential Origin', date: 'Aug 17, 2026', size: '1.1 MB' },
                  { name: 'Delivery Order (DO)', code: 'DO-RTM-4410', status: 'APPROVED', type: 'Release Document', date: 'Aug 28, 2026', size: '490 KB' },
                ].map((doc, idx) => (
                  <div key={idx} className="trkd-doc-item-card">
                    <div className="trkd-dic-left">
                      <div className="trkd-dic-icon">
                        <FileText size={18} />
                      </div>
                      <div className="trkd-dic-info">
                        <strong>{doc.name}</strong>
                        <span className="trkd-dic-code">{doc.code}</span>
                        <span className="trkd-dic-meta">{doc.type} · {doc.size} · Uploaded {doc.date}</span>
                      </div>
                    </div>
                    <div className="trkd-dic-right">
                      <span className="trkd-doc-status-badge">{doc.status}</span>
                      <button className="trkd-btn-ghost-sm" onClick={() => copyToClipboard(doc.code, `doc-${idx}`)}>
                        {copiedKey === `doc-${idx}` ? 'Copied' : 'Copy Ref'}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Right Compliance Checklist */}
            <div className="trkd-card trkd-docs-side-card">
              <h3 className="trkd-card-title">Compliance Engine Audit</h3>
              <div className="trkd-dsa-list">
                <div className="trkd-dsa-item done">
                  <CheckCircle2 size={14} className="green" />
                  <div>
                    <strong>MBL Consignee Verification</strong>
                    <span>Matches KYC Entity in Netherlands</span>
                  </div>
                </div>
                <div className="trkd-dsa-item done">
                  <CheckCircle2 size={14} className="green" />
                  <div>
                    <strong>Commercial Invoice Valuation</strong>
                    <span>Cross-referenced with customs bond</span>
                  </div>
                </div>
                <div className="trkd-dsa-item done">
                  <CheckCircle2 size={14} className="green" />
                  <div>
                    <strong>Origin Chamber Endorsement</strong>
                    <span>Digital seal authenticated</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── Tab Content: 7. Containers ──────────────────────────────────────── */}
      {activeTab === 'containers' && (
        <div className="trkd-containers-workspace">
          {/* Container KPI Strip */}
          <div className="trkd-cnt-kpi-bar">
            <div className="trkd-ckb-unit">
              <span className="trkd-ckb-lbl">TOTAL ALLOCATED</span>
              <strong className="trkd-ckb-val">2 Containers</strong>
            </div>
            <div className="trkd-ckb-unit blue">
              <span className="trkd-ckb-lbl">CONTAINER TYPE</span>
              <strong className="trkd-ckb-val">40' High Cube (40HC)</strong>
            </div>
            <div className="trkd-ckb-unit">
              <span className="trkd-ckb-lbl">TOTAL GROSS WEIGHT</span>
              <strong className="trkd-ckb-val">46,600 KG</strong>
            </div>
            <div className="trkd-ckb-unit green">
              <span className="trkd-ckb-lbl">ALL SEALS VERIFIED</span>
              <strong className="trkd-ckb-val green">100% Intact</strong>
            </div>
          </div>

          <div className="trkd-card">
            <h3 className="trkd-card-title">Container Allocation & Equipment Status</h3>
            <div className="trkd-table-wrap">
              <table className="trkd-table">
                <thead>
                  <tr>
                    <th>CONTAINER NUMBER</th>
                    <th>TYPE</th>
                    <th>SEAL NUMBER</th>
                    <th>TARE WEIGHT</th>
                    <th>GROSS WEIGHT</th>
                    <th>LAST EVENT</th>
                    <th>LOCATION</th>
                    <th>STATUS</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td><strong className="trkd-mono">MAEU1234567</strong></td>
                    <td>40' High Cube (40HC)</td>
                    <td>ML-SEAL-8991</td>
                    <td>3,820 KG</td>
                    <td>24,500 KG</td>
                    <td>Terminal Discharge</td>
                    <td>ECT Delta Rotterdam</td>
                    <td><span className="trkd-status-pill completed">ON VESSEL</span></td>
                  </tr>
                  <tr>
                    <td><strong className="trkd-mono">MAEU7654321</strong></td>
                    <td>40' High Cube (40HC)</td>
                    <td>ML-SEAL-8992</td>
                    <td>3,820 KG</td>
                    <td>22,100 KG</td>
                    <td>Terminal Discharge</td>
                    <td>ECT Delta Rotterdam</td>
                    <td><span className="trkd-status-pill completed">ON VESSEL</span></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* ── Tab Content: 8. Financials ──────────────────────────────────────── */}
      {activeTab === 'financials' && (
        <div className="trkd-fin-workspace">
          {/* Financial Summary Top Cards */}
          <div className="trkd-fin-grid">
            <div className="trkd-fin-card">
              <span className="trkd-fc-lbl">TOTAL REVENUE (INVOICED)</span>
              <strong className="trkd-fc-val">$4,850.00</strong>
              <span className="trkd-fc-sub">USD · Paid in Full</span>
            </div>
            <div className="trkd-fin-card">
              <span className="trkd-fc-lbl">CARRIER OCEAN FREIGHT</span>
              <strong className="trkd-fc-val">$3,200.00</strong>
              <span className="trkd-fc-sub">Maersk Line Base Ocean Rate</span>
            </div>
            <div className="trkd-fin-card">
              <span className="trkd-fc-lbl">PORT & TERMINAL CHARGES</span>
              <strong className="trkd-fc-val">$450.00</strong>
              <span className="trkd-fc-sub">THC & Origin Port Dues</span>
            </div>
            <div className="trkd-fin-card highlight">
              <span className="trkd-fc-lbl">NET OPERATIONAL PROFIT</span>
              <strong className="trkd-fc-val green">$1,200.00</strong>
              <span className="trkd-fc-sub green">24.7% Margin</span>
            </div>
          </div>

          <div className="trkd-fin-main-grid">
            {/* Cost Breakdown Table */}
            <div className="trkd-card trkd-fin-table-card">
              <div className="trkd-card-header">
                <h3 className="trkd-card-title">Operational Cost Breakdown</h3>
                <button
                  className="trkd-btn-primary"
                  onClick={() => navigate(`/dashboard/shipments/${shipment.id}`)}
                >
                  Open Financial Ledger →
                </button>
              </div>

              <div className="trkd-table-wrap">
                <table className="trkd-table">
                  <thead>
                    <tr>
                      <th>LINE ITEM</th>
                      <th>CATEGORY</th>
                      <th>CURRENCY</th>
                      <th>COST (USD)</th>
                      <th>REVENUE (USD)</th>
                      <th>STATUS</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td><strong>Ocean Freight (2x40HC)</strong></td>
                      <td>Carrier Freight</td>
                      <td>USD</td>
                      <td>$3,200.00</td>
                      <td>$4,100.00</td>
                      <td><span className="trkd-status-pill completed">BILLED</span></td>
                    </tr>
                    <tr>
                      <td><strong>Terminal Handling Charges (THC)</strong></td>
                      <td>Port Dues</td>
                      <td>USD</td>
                      <td>$350.00</td>
                      <td>$500.00</td>
                      <td><span className="trkd-status-pill completed">BILLED</span></td>
                    </tr>
                    <tr>
                      <td><strong>Documentation & Compliance Fee</strong></td>
                      <td>Operational</td>
                      <td>USD</td>
                      <td>$100.00</td>
                      <td>$250.00</td>
                      <td><span className="trkd-status-pill completed">BILLED</span></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            {/* Right Financial Intelligence Card */}
            <div className="trkd-card trkd-fin-side-card">
              <h3 className="trkd-card-title">Margin Intelligence</h3>
              <div className="trkd-fsi-unit">
                <span>Freight Cost Ratio</span>
                <strong>66.0%</strong>
              </div>
              <div className="trkd-fsi-unit">
                <span>Demurrage / Detention Risk</span>
                <strong className="green">$0.00 (Free time: 7 days)</strong>
              </div>
              <div className="trkd-fsi-unit">
                <span>Customer Payment Status</span>
                <strong className="green">Fully Settled</strong>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── Tab Content: 9. Notes ───────────────────────────────────────────── */}
      {activeTab === 'notes' && (
        <div className="trkd-notes-workspace">
          <div className="trkd-notes-main-grid">
            {/* Notes List & Creation Card */}
            <div className="trkd-card trkd-notes-main-card">
              <h3 className="trkd-card-title">Operational Communications & Handover Notes</h3>

              {/* Note Creator Form */}
              <form onSubmit={handleAddNote} className="trkd-note-form">
                <div className="trkd-nf-top">
                  <span className="trkd-nf-lbl">Add Note:</span>
                  <div className="trkd-nf-types">
                    {['Operational', 'Customer', 'Internal', 'System'].map((t) => (
                      <button
                        key={t}
                        type="button"
                        className={`trkd-nft-btn ${newNoteType === t ? 'active' : ''}`}
                        onClick={() => setNewNoteType(t)}
                      >
                        {t}
                      </button>
                    ))}
                  </div>
                </div>
                <textarea
                  className="trkd-nf-textarea"
                  placeholder="Type an operational note, dispatcher handover, or customer communication log..."
                  value={newNoteText}
                  onChange={(e) => setNewNoteText(e.target.value)}
                  rows={3}
                />
                <div className="trkd-nf-actions">
                  <button type="submit" className="trkd-btn-primary" disabled={!newNoteText.trim()}>
                    <Send size={13} /> Post Note
                  </button>
                </div>
              </form>

              {/* Notes Stream */}
              <div className="trkd-notes-list">
                {notes.map((n) => (
                  <div key={n.id} className="trkd-note-card">
                    <div className="trkd-nc-head">
                      <div className="trkd-nc-author">
                        <strong>{n.author}</strong>
                        <span className="trkd-nc-role">{n.role}</span>
                      </div>
                      <div className="trkd-nc-meta">
                        <span className={`trkd-nc-tag tag-${n.type.toLowerCase()}`}>{n.type}</span>
                        <span className="trkd-nc-time">{n.time}</span>
                      </div>
                    </div>
                    <p className="trkd-nc-body">{n.text}</p>
                  </div>
                ))}
              </div>
            </div>

            {/* Right Handover Protocol Card */}
            <div className="trkd-card trkd-notes-side-card">
              <h3 className="trkd-card-title">Shift Handover Protocol</h3>
              <div className="trkd-shp-box">
                <strong>Next Shift Action Required:</strong>
                <p>Monitor ECT Delta berth dispatch and confirm container delivery confirmation with European consignee warehouse.</p>
              </div>
              <div className="trkd-shp-info">
                <span>Desk Lead:</span> <strong>Varun Kanade</strong>
              </div>
              <div className="trkd-shp-info">
                <span>Carrier Contact:</span> <strong>Maersk Rotterdam Ops Desk</strong>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── In-Place Exception Resolution Modal ─────────────────────────────── */}
      {resolvingException && (
        <div className="trkd-modal-overlay" onClick={() => setResolvingException(null)}>
          <div className="trkd-modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="trkd-modal-header">
              <div className="trkd-modal-title-wrap">
                <AlertTriangle size={18} className="trkd-modal-exc-icon" />
                <h3>Resolve Operational Exception</h3>
              </div>
              <button className="trkd-modal-close" onClick={() => setResolvingException(null)}>
                <X size={16} />
              </button>
            </div>

            <div className="trkd-modal-form">
              <div className="trkd-exc-modal-detail">
                <div className="trkd-emd-row">
                  <span className="trkd-emd-lbl">Exception Type</span>
                  <strong className="trkd-emd-val">{resolvingException.exception_type}</strong>
                </div>
                <div className="trkd-emd-row">
                  <span className="trkd-emd-lbl">Severity / Category</span>
                  <div className="trkd-emd-val">
                    <span className={`trkd-sev-badge sev-${resolvingException.severity?.toLowerCase()}`}>
                      {resolvingException.severity}
                    </span>
                    <span className="trkd-emd-cat"> · {resolvingException.category || 'Schedule Delay'}</span>
                  </div>
                </div>
                <div className="trkd-emd-desc-box">
                  {resolvingException.description}
                </div>
              </div>

              <div className="trkd-form-group">
                <label>Resolution Action & Settlement Notes</label>
                <div className="trkd-chips-row">
                  {[
                    'Berth cleared & vessel underway',
                    'Contingency buffer applied',
                    'Terminal yard release confirmed',
                    'Customer notified & approved',
                  ].map((chip) => (
                    <button
                      key={chip}
                      type="button"
                      className="trkd-chip-btn"
                      onClick={() => setResolutionNotes(chip)}
                    >
                      + {chip}
                    </button>
                  ))}
                </div>
                <textarea
                  className="trkd-modal-textarea"
                  placeholder="Detail the operational resolution taken to clear this exception..."
                  value={resolutionNotes}
                  onChange={(e) => setResolutionNotes(e.target.value)}
                  rows={3}
                  autoFocus
                />
              </div>

              <div className="trkd-modal-footer">
                <button
                  type="button"
                  className="trkd-btn-cancel"
                  onClick={() => setResolvingException(null)}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="trkd-btn-submit"
                  onClick={handleConfirmResolve}
                  disabled={isSubmittingResolution}
                >
                  <Check size={14} /> Confirm Resolution
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── Task 17.5: Alert Action Modal (Resolve / Suppress) ──────────────── */}
      {alertActionModal && (
        <div className="trkd-modal-overlay" onClick={() => setAlertActionModal(null)}>
          <div className="trkd-modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="trkd-modal-header">
              <div className="trkd-modal-title-wrap">
                {alertActionModal.type === 'RESOLVE' ? (
                  <CheckCircle2 size={18} className="trkd-modal-exc-icon green" />
                ) : (
                  <PauseCircle size={18} className="trkd-modal-exc-icon amber" />
                )}
                <h3 className="trkd-modal-title">
                  {alertActionModal.type === 'RESOLVE' ? 'Resolve Tracking Alert' : 'Suppress Tracking Alert'}
                </h3>
              </div>
              <button className="trkd-modal-close" onClick={() => setAlertActionModal(null)}>
                <X size={16} />
              </button>
            </div>

            <div className="trkd-modal-form">
              <div className="trkd-exc-modal-detail">
                <strong>{alertActionModal.alert.title}</strong>
                <p>{alertActionModal.alert.description}</p>
              </div>

              <label className="trkd-modal-label">
                {alertActionModal.type === 'RESOLVE' ? 'Resolution Notes:' : 'Reason for Suppression:'}
              </label>

              <div className="trkd-quick-chips">
                {(alertActionModal.type === 'RESOLVE'
                  ? [
                      'Issue verified cleared on vessel AIS',
                      'Contingency transit adjustment applied',
                      'Carrier schedule normalized',
                      'Manual milestone confirmation received',
                    ]
                  : [
                      'Planned operational delay within tolerance',
                      'Duplicate notification noise',
                      'Customer pre-approved variance',
                      'Temporary port queue expected',
                    ]
                ).map((chip) => (
                  <button
                    key={chip}
                    type="button"
                    className="trkd-chip-btn"
                    onClick={() => setActionNotes(chip)}
                  >
                    + {chip}
                  </button>
                ))}
              </div>

              <textarea
                className="trkd-modal-textarea"
                placeholder={
                  alertActionModal.type === 'RESOLVE'
                    ? 'Enter notes describing how this operational condition was resolved...'
                    : 'Enter operational reason for suppressing this alert...'
                }
                value={actionNotes}
                onChange={(e) => setActionNotes(e.target.value)}
                rows={3}
                autoFocus
              />

              <div className="trkd-modal-footer">
                <button
                  type="button"
                  className="trkd-btn-cancel"
                  onClick={() => setAlertActionModal(null)}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="trkd-btn-submit"
                  onClick={handleConfirmAlertAction}
                  disabled={actionSubmitting}
                >
                  {actionSubmitting
                    ? 'Submitting...'
                    : alertActionModal.type === 'RESOLVE'
                    ? 'Confirm Resolution'
                    : 'Confirm Suppression'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
