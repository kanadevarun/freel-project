import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import ManualUpdatePanel from './ManualUpdatePanel';
import DocumentWorkspace from './DocumentWorkspace';
import FinanceWorkspace from './FinanceWorkspace';
import BillingWorkspace from './BillingWorkspace';
import CustomDropdown from './CustomDropdown';
import api from '../../../services/api';
import { shipmentService } from '../../../services/shipmentService';
import {
  Ship,
  Package,
  Calendar,
  MapPin,
  User,
  AlertTriangle,
  ArrowLeft,
  Link as LinkIcon,
  CheckCircle2,
  ExternalLink,
  RefreshCw,
  Zap,
  MoreVertical,
  ChevronDown,
  Clock,
  FileText,
  DollarSign,
  ShieldAlert,
  Info,
  Check,
  AlertCircle,
  Radio,
  Activity
} from 'lucide-react';
import './Shipments.css';

export default function ShipmentDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [shipment, setShipment] = useState(null);
  const [milestones, setMilestones] = useState([]);
  const [exceptions, setExceptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [rawUpdate, setRawUpdate] = useState('');
  const [submittingUpdate, setSubmittingUpdate] = useState(false);
  const [updateMsg, setUpdateMsg] = useState('');

  // Header Dropdown State
  const [headerMenuOpen, setHeaderMenuOpen] = useState(false);
  const menuRef = useRef(null);
  const rightColRef = useRef(null);
  const [canScrollRightMore, setCanScrollRightMore] = useState(false);

  // Exception Interaction States
  const [resolvingExcId, setResolvingExcId] = useState(null);
  const [resolutionNotesInput, setResolutionNotesInput] = useState('');
  const [showRaiseForm, setShowRaiseForm] = useState(false);
  const [showResolvedHistory, setShowResolvedHistory] = useState(false);
  const [newExcType, setNewExcType] = useState('SCHEDULE_DELAY');
  const [newExcSeverity, setNewExcSeverity] = useState('MEDIUM');
  const [newExcTitle, setNewExcTitle] = useState('');
  const [newExcDesc, setNewExcDesc] = useState('');
  const [submittingException, setSubmittingException] = useState(false);
  const [raiseError, setRaiseError] = useState('');

  // Milestone Update States
  const [activeMilestoneForm, setActiveMilestoneForm] = useState(null);
  const [actualDate, setActualDate] = useState('');
  const [milestoneLocation, setMilestoneLocation] = useState('');
  const [milestoneNotes, setMilestoneNotes] = useState('');
  const [submittingMilestone, setSubmittingMilestone] = useState(false);

  const [trackingSummary, setTrackingSummary] = useState(null);
  const [submittingClosure, setSubmittingClosure] = useState(false);
  const [runningDiagnostics, setRunningDiagnostics] = useState(false);

  // Check if right column has more to scroll
  const checkRightColScroll = () => {
    const el = rightColRef.current;
    if (!el) return;
    const isScrollable = el.scrollHeight > el.clientHeight + 10;
    const isAtBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 25;
    setCanScrollRightMore(isScrollable && !isAtBottom);
  };

  useEffect(() => {
    // Initial and periodic check on render
    const timer = setTimeout(checkRightColScroll, 100);
    window.addEventListener('resize', checkRightColScroll);
    return () => {
      clearTimeout(timer);
      window.removeEventListener('resize', checkRightColScroll);
    };
  }, [shipment, trackingSummary, exceptions]);

  const handleScrollRightColToBottom = () => {
    if (rightColRef.current) {
      rightColRef.current.scrollTo({
        top: rightColRef.current.scrollHeight,
        behavior: 'smooth'
      });
    }
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setHeaderMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleOpenMilestoneForm = (m) => {
    setActiveMilestoneForm(m);
    const tzoffset = (new Date()).getTimezoneOffset() * 60000;
    const localISOTime = (new Date(Date.now() - tzoffset)).toISOString().slice(0, 16);
    setActualDate(localISOTime);
    setMilestoneLocation(m.location || '');
    setMilestoneNotes(m.notes || '');
  };

  const handleUpdateMilestone = async (e) => {
    e.preventDefault();
    if (!activeMilestoneForm) return;

    try {
      setSubmittingMilestone(true);
      const payload = {
        milestone_code: activeMilestoneForm.milestone_code,
        actual_date: new Date(actualDate).toISOString(),
        location: milestoneLocation ? milestoneLocation : undefined,
        notes: milestoneNotes ? milestoneNotes : undefined
      };

      await api.put(`/api/v1/shipments/${id}/milestones`, payload);
      setActiveMilestoneForm(null);
      await fetchShipmentDetails(false);
    } catch (err) {
      console.error('Error manually updating milestone:', err);
      alert('Failed to update milestone: ' + (err.message || 'Unknown error'));
    } finally {
      setSubmittingMilestone(false);
    }
  };

  const isDev = import.meta.env.DEV || window.location.hostname === 'localhost';

  useEffect(() => {
    fetchShipmentDetails(true);
  }, [id]);

  const fetchShipmentDetails = async (isInitial = false) => {
    try {
      if (isInitial || !shipment) {
        setLoading(true);
      }
      setError(null);
      const resData = await api.get(`/api/v1/shipments/${id}`);
      if (resData) {
        setShipment(resData.shipment);
        setMilestones(resData.milestones || []);
        setExceptions(resData.exceptions || []);

        try {
          const trackSummary = await api.get(`/api/v1/shipments/${id}/tracking`);
          setTrackingSummary(trackSummary);
        } catch (tErr) {
          console.error("Failed to load tracking summary:", tErr);
        }
      }
    } catch (err) {
      console.error('Failed to fetch shipment details:', err);
      if (!shipment) {
        setError(err.message || 'Access denied or resource not found.');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleAcknowledgeException = async (excId) => {
    try {
      await api.post(`/api/v1/shipments/${id}/exceptions/${excId}/acknowledge`);
      await fetchShipmentDetails(false);
    } catch (err) {
      console.error('Error acknowledging exception:', err);
      alert('Failed to acknowledge exception: ' + (err.message || 'Unknown error'));
    }
  };

  const handleResolveExceptionSubmit = async (e, excId) => {
    e.preventDefault();
    if (!resolutionNotesInput.trim()) {
      alert('Please enter resolution notes to document actions taken.');
      return;
    }
    try {
      await api.post(`/api/v1/shipments/${id}/exceptions/${excId}/resolve`, {
        resolution_notes: resolutionNotesInput,
      });
      setResolvingExcId(null);
      setResolutionNotesInput('');
      await fetchShipmentDetails(false);
    } catch (err) {
      console.error('Error resolving exception:', err);
      alert('Failed to resolve exception: ' + (err.message || 'Unknown error'));
    }
  };

  const handleDismissException = async (excId) => {
    if (!window.confirm('Are you sure you want to dismiss this operational exception? This will archive the alert without documentation.')) {
      return;
    }
    try {
      await api.post(`/api/v1/shipments/${id}/exceptions/${excId}/dismiss`);
      await fetchShipmentDetails(false);
    } catch (err) {
      console.error('Error dismissing exception:', err);
      alert('Failed to dismiss exception: ' + (err.message || 'Unknown error'));
    }
  };

  const handleRaiseException = async (e) => {
    e.preventDefault();
    setRaiseError('');

    // If title is left blank, auto-generate a clean professional title from the category / description
    let titleToSubmit = newExcTitle.trim();
    if (!titleToSubmit) {
      if (newExcDesc.trim()) {
        titleToSubmit = newExcDesc.trim().slice(0, 48);
        if (newExcDesc.trim().length > 48) titleToSubmit += '...';
      } else {
        const typeLabels = {
          SCHEDULE_DELAY: 'Schedule Delay Alert',
          VESSEL_ROLLOVER: 'Vessel Rollover Incident',
          PORT_CONGESTION: 'Port Congestion Delay',
          CUSTOMS_HOLD: 'Customs Clearance Hold',
          DOCUMENT_ISSUE: 'Documentation Discrepancy',
          CARRIER_DELAY: 'Carrier Transit Delay',
          ROUTE_DEVIATION: 'Vessel Route Deviation',
          CONTAINER_ISSUE: 'Container Equipment Issue',
          OTHER: 'Operational Exception Alert',
        };
        titleToSubmit = typeLabels[newExcType] || `${newExcType.replace(/_/g, ' ')} Alert`;
      }
    }

    try {
      setSubmittingException(true);
      await api.post(`/api/v1/shipments/${id}/exceptions`, {
        exception_type: newExcType,
        severity: newExcSeverity,
        title: titleToSubmit,
        description: newExcDesc.trim(),
      });
      setNewExcTitle('');
      setNewExcDesc('');
      setShowRaiseForm(false);
      await fetchShipmentDetails(false);
    } catch (err) {
      console.error('Error raising exception:', err);
      setRaiseError(err.message || 'Failed to create exception alert. Please try again.');
    } finally {
      setSubmittingException(false);
    }
  };

  const handleEvaluateExceptions = async () => {
    try {
      setRunningDiagnostics(true);
      await api.post(`/api/v1/shipments/${id}/exceptions/evaluate`);
      await fetchShipmentDetails(false);
    } catch (err) {
      console.error('Error evaluating exceptions:', err);
      alert('Failed to run exceptions diagnostics: ' + (err.message || 'Unknown error'));
    } finally {
      setRunningDiagnostics(false);
    }
  };

  const [syncingTracking, setSyncingTracking] = useState(false);
  const [syncFeedback, setSyncFeedback] = useState(null);

  const handleSyncTracking = async () => {
    try {
      setSyncingTracking(true);
      const res = await shipmentService.refreshShipmentTracking(id);
      const resultData = res?.data?.data || res?.data;
      if (resultData?.message) {
        setSyncFeedback({
          message: resultData.message,
          success: resultData.success,
          usedFallback: resultData.used_fallback,
        });
        setTimeout(() => setSyncFeedback(null), 5000);
      }
      await fetchShipmentDetails(false);
    } catch (err) {
      console.error('Failed to sync tracking:', err);
      setSyncFeedback({
        message: 'Unable to sync tracking from carrier right now. Showing existing operational tracking data.',
        success: false,
        usedFallback: true,
      });
      setTimeout(() => setSyncFeedback(null), 5000);
    } finally {
      setSyncingTracking(false);
    }
  };

  const handleSubmitCarrierUpdate = async (e) => {
    e.preventDefault();
    if (!rawUpdate.trim()) return;

    try {
      setSubmittingUpdate(true);
      setUpdateMsg('');
      await api.post(`/api/v1/shipments/${id}/carrier-update`, {
        event_id: `WEB-MANUAL-${Date.now()}`,
        description: rawUpdate
      });

      setUpdateMsg('✅ Update received. Processing task enqueued...');
      setRawUpdate('');
      
      let attempts = 0;
      const intervalId = setInterval(async () => {
        attempts += 1;
        await fetchShipmentDetails(false);
        if (attempts >= 3) {
          clearInterval(intervalId);
        }
      }, 3000);

    } catch (err) {
      console.error('Error posting carrier update:', err);
      setUpdateMsg(`❌ Failed to process: ${err.message || 'Connection failure'}`);
    } finally {
      setSubmittingUpdate(false);
    }
  };

  const handleEvaluateClosure = async () => {
    try {
      setSubmittingClosure(true);
      await api.post(`/api/v1/shipments/${id}/evaluate`);
      await fetchShipmentDetails(false);
    } catch (err) {
      alert("Failed to evaluate closure eligibility: " + (err.message || err));
    } finally {
      setSubmittingClosure(false);
    }
  };

  const handleCompleteShipment = async () => {
    try {
      setSubmittingClosure(true);
      await api.post(`/api/v1/shipments/${id}/complete`);
      await fetchShipmentDetails(false);
    } catch (err) {
      alert("Failed to complete shipment: " + (err.message || err));
    } finally {
      setSubmittingClosure(false);
    }
  };

  const handleReopenShipment = async () => {
    try {
      setSubmittingClosure(true);
      await api.post(`/api/v1/shipments/${id}/reopen`);
      await fetchShipmentDetails(false);
    } catch (err) {
      alert("Failed to reopen shipment: " + (err.message || err));
    } finally {
      setSubmittingClosure(false);
    }
  };

  const PORT_INFO_MAP = {
    INNSA: { name: 'Nhava Sheva (Mumbai)', country: 'India', flag: '🇮🇳' },
    INBOM: { name: 'Mumbai Port', country: 'India', flag: '🇮🇳' },
    INMAA: { name: 'Chennai Port', country: 'India', flag: '🇮🇳' },
    INCOK: { name: 'Cochin Port', country: 'India', flag: '🇮🇳' },
    INMUN: { name: 'Mundra Port', country: 'India', flag: '🇮🇳' },
    NLRTM: { name: 'Rotterdam', country: 'Netherlands', flag: '🇳🇱' },
    BEANR: { name: 'Antwerp', country: 'Belgium', flag: '🇧🇪' },
    DEHAM: { name: 'Hamburg', country: 'Germany', flag: '🇩🇪' },
    DEBRV: { name: 'Bremerhaven', country: 'Germany', flag: '🇩🇪' },
    CNSHA: { name: 'Shanghai', country: 'China', flag: '🇨🇳' },
    CNNGB: { name: 'Ningbo-Zhoushan', country: 'China', flag: '🇨🇳' },
    CNSZX: { name: 'Shenzhen', country: 'China', flag: '🇨🇳' },
    SGSIN: { name: 'Singapore', country: 'Singapore', flag: '🇸🇬' },
    USLAX: { name: 'Los Angeles', country: 'USA', flag: '🇺🇸' },
    USLGB: { name: 'Long Beach', country: 'USA', flag: '🇺🇸' },
    USNYC: { name: 'New York / NJ', country: 'USA', flag: '🇺🇸' },
    USSAV: { name: 'Savannah', country: 'USA', flag: '🇺🇸' },
    AEJEA: { name: 'Jebel Ali (Dubai)', country: 'UAE', flag: '🇦🇪' },
    GBLON: { name: 'London Gateway', country: 'UK', flag: '🇬🇧' },
    GBFXT: { name: 'Felixstowe', country: 'UK', flag: '🇬🇧' },
    FRLEH: { name: 'Le Havre', country: 'France', flag: '🇫🇷' },
    JPTYO: { name: 'Tokyo', country: 'Japan', flag: '🇯🇵' },
    JPYOK: { name: 'Yokohama', country: 'Japan', flag: '🇯🇵' },
    KRPUS: { name: 'Busan', country: 'South Korea', flag: '🇰🇷' },
  };

  const getPortDetails = (portCode) => {
    if (!portCode) return { code: '—', name: '', country: '', flag: '', fullName: '—' };
    const cleanCode = String(portCode).toUpperCase().trim();
    const info = PORT_INFO_MAP[cleanCode];
    if (info) {
      return {
        code: cleanCode,
        name: info.name,
        country: info.country,
        flag: info.flag,
        fullName: `${info.name} (${cleanCode})`
      };
    }
    return { code: cleanCode, name: '', country: '', flag: '📍', fullName: cleanCode };
  };

  const formatPort = (portCode) => {
    if (!portCode) return '—';
    const clean = portCode.toUpperCase();
    const info = PORT_INFO_MAP[clean];
    if (info) return `${clean} · ${info.name}`;
    return clean;
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '—';
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch {
      return dateStr;
    }
  };

  const formatDateOnly = (dateStr) => {
    if (!dateStr) return '—';
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch {
      return dateStr;
    }
  };

  // ── Computed state ─────────────────────────────────────────────────────────
  const activeExceptionCount = exceptions.filter(e => e.status !== 'RESOLVED' && e.status !== 'DISMISSED').length;
  const criticalExcCount = exceptions.filter(e => e.severity === 'CRITICAL' && e.status !== 'RESOLVED' && e.status !== 'DISMISSED').length;
  const resolvedExcCount = exceptions.filter(e => e.status === 'RESOLVED' || e.status === 'DISMISSED').length;

  if (loading && !shipment) {
    return (
      <div className="sd-container" style={{ padding: '32px', display: 'flex', flexDirection: 'column', gap: '20px' }}>
        <div className="sd-skel" style={{ height: '32px', width: '200px' }} />
        <div className="sd-skel" style={{ height: '100px' }} />
        <div className="sd-skel" style={{ height: '60px' }} />
        <div className="sd-skel" style={{ height: '300px' }} />
      </div>
    );
  }

  if (error || !shipment) {
    return (
      <div className="sd-container">
        <div className="sd-error-banner">
          <AlertTriangle size={18} />
          <div>
            <strong>Failed to load shipment:</strong> {error || 'Record not found.'}
          </div>
          <button className="sd-back-btn" onClick={() => navigate('/dashboard/shipments')}>
            <ArrowLeft size={14} /> Back to Shipments
          </button>
        </div>
      </div>
    );
  }

  const trackingState = trackingSummary?.tracking_state;
  const closureStatus = trackingSummary?.closure_status || 'ACTIVE';
  const shipmentStatus = shipment.status || 'UNKNOWN';
  const progressPct = trackingSummary?.progress_percentage ?? (() => {
    switch (shipmentStatus) {
      case 'BOOKING_PENDING': return 10;
      case 'BOOKED': return 30;
      case 'DEPARTED': return 50;
      case 'IN_TRANSIT': return 70;
      case 'ARRIVED': return 90;
      case 'DELIVERED': return 100;
      default: return 0;
    }
  })();

  const containers = Array.isArray(shipment.container_numbers)
    ? shipment.container_numbers
    : typeof shipment.container_numbers === 'string'
      ? (() => { try { return JSON.parse(shipment.container_numbers); } catch { return shipment.container_numbers ? [shipment.container_numbers] : []; } })()
      : [];

  const containerCount = containers.length;

  const trackingStateConfig = {
    ON_TRACK:  { label: 'On Track',  bg: '#eff6ff', fg: '#1d4ed8', border: '#bfdbfe' },
    AT_RISK:   { label: 'At Risk',   bg: '#fffbeb', fg: '#b45309', border: '#fde68a' },
    DELAYED:   { label: 'Delayed',   bg: '#fef2f2', fg: '#dc2626', border: '#fecaca' },
    EXCEPTION: { label: 'Exception', bg: '#fef2f2', fg: '#dc2626', border: '#fecaca' },
    COMPLETED: { label: 'Completed', bg: '#f0fdf4', fg: '#15803d', border: '#bbf7d0' },
  };
  const tsCfg = trackingStateConfig[trackingState] || { label: trackingState || 'TRACKED', bg: '#f1f5f9', fg: '#475569', border: '#e2e8f0' };

  const statusColorMap = {
    BOOKING_PENDING: { bg: '#fffbeb', fg: '#92400e', dot: '#d97706' },
    BOOKED:          { bg: '#eff6ff', fg: '#1e40af', dot: '#2563eb' },
    DEPARTED:        { bg: '#f0f9ff', fg: '#0369a1', dot: '#0284c7' },
    IN_TRANSIT:      { bg: '#f5f3ff', fg: '#5b21b6', dot: '#6d28d9' },
    ARRIVED:         { bg: '#ecfdf5', fg: '#065f46', dot: '#10b981' },
    DELIVERED:       { bg: '#f0fdf4', fg: '#166534', dot: '#15803d' },
    EXCEPTION:       { bg: '#fef2f2', fg: '#dc2626', dot: '#ef4444' },
  };
  const sCfg = statusColorMap[shipmentStatus] || { bg: '#f1f5f9', fg: '#475569', dot: '#94a3b8' };

  const closureColorMap = {
    ACTIVE:               { label: 'Active',           bg: '#dbeafe', fg: '#1d4ed8', border: '#bfdbfe' },
    READY_FOR_CLOSURE:    { label: 'Ready for Closure', bg: '#fef3c7', fg: '#d97706', border: '#fde68a' },
    CLOSED:               { label: 'Closed',           bg: '#dcfce7', fg: '#15803d', border: '#a7f3d0' },
    BLOCKED_BY_EXCEPTION: { label: 'Blocked',          bg: '#fee2e2', fg: '#dc2626', border: '#fecaca' },
  };
  const cCfg = closureColorMap[closureStatus] || closureColorMap.ACTIVE;

  const originPort = getPortDetails(shipment.origin_port);
  const destPort = getPortDetails(shipment.destination_port);

  return (
    <div className="sd-container">

      {/* ── 1. HERO SHIPMENT IDENTITY HEADER ───────────────────────────────── */}
      <div className="sd-hero">
        <div className="sd-hero-top-bar">
          <div className="sd-hero-nav">
            <button className="sd-back-btn" onClick={() => navigate('/dashboard/shipments')}>
              <ArrowLeft size={14} /> Shipments
            </button>
            <span className="sd-breadcrumb-sep">›</span>
            <span className="sd-breadcrumb-cur">SH-{shipment.id}</span>
          </div>

          <div className="sd-hero-actions">
            {shipment.booking_id && (
              <Link to={`/dashboard/bookings/${shipment.booking_id}`} className="sd-btn sd-btn-ghost">
                <ExternalLink size={14} /> View Booking
              </Link>
            )}

            <button
              className="sd-btn sd-btn-outline"
              onClick={handleSyncTracking}
              disabled={syncingTracking}
              title="Sync live tracking milestones from carrier integration"
            >
              <RefreshCw size={14} className={syncingTracking ? 'sd-spin' : ''} />
              {syncingTracking ? 'Syncing...' : 'Sync Tracking'}
            </button>

            <button
              className="sd-btn sd-btn-ghost"
              onClick={() => {
                const firstPending = milestones.find(m => m.status !== 'COMPLETED');
                if (firstPending) {
                  handleOpenMilestoneForm(firstPending);
                  document.getElementById('milestones-panel')?.scrollIntoView({ behavior: 'smooth' });
                } else if (milestones.length > 0) {
                  handleOpenMilestoneForm(milestones[milestones.length - 1]);
                  document.getElementById('milestones-panel')?.scrollIntoView({ behavior: 'smooth' });
                }
              }}
            >
              <RefreshCw size={14} /> Update Milestone
            </button>

            <button
              className="sd-btn sd-btn-ghost"
              onClick={handleEvaluateExceptions}
              disabled={runningDiagnostics}
            >
              <Zap size={14} className={runningDiagnostics ? 'sd-spin' : ''} />
              {runningDiagnostics ? 'Evaluating...' : 'Run Diagnostics'}
            </button>

            <div className="sd-actions-menu-wrap" ref={menuRef}>
              <button
                className="sd-btn sd-btn-outline"
                onClick={() => setHeaderMenuOpen(!headerMenuOpen)}
              >
                More Actions <ChevronDown size={14} />
              </button>
              {headerMenuOpen && (
                <div className="sd-header-dropdown">
                  {closureStatus === 'ACTIVE' && (
                    <button
                      className="sd-dropdown-item"
                      onClick={() => { setHeaderMenuOpen(false); handleEvaluateClosure(); }}
                    >
                      <Check size={14} /> Evaluate Closure Eligibility
                    </button>
                  )}
                  {closureStatus === 'READY_FOR_CLOSURE' && (
                    <button
                      className="sd-dropdown-item sd-item-success"
                      onClick={() => { setHeaderMenuOpen(false); handleCompleteShipment(); }}
                    >
                      <CheckCircle2 size={14} /> Complete & Close File
                    </button>
                  )}
                  {closureStatus === 'CLOSED' && (
                    <button
                      className="sd-dropdown-item"
                      onClick={() => { setHeaderMenuOpen(false); handleReopenShipment(); }}
                    >
                      <RefreshCw size={14} /> Reopen Shipment File
                    </button>
                  )}
                  <button
                    className="sd-dropdown-item"
                    onClick={() => { setHeaderMenuOpen(false); setShowRaiseForm(true); document.getElementById('exceptions-panel')?.scrollIntoView({ behavior: 'smooth' }); }}
                  >
                    <AlertCircle size={14} /> Raise Manual Exception
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>

        {syncFeedback && (
          <div
            style={{
              margin: '12px 24px',
              padding: '12px 18px',
              borderRadius: '8px',
              fontSize: '13px',
              fontWeight: '500',
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
              background: syncFeedback.success && !syncFeedback.usedFallback ? '#ecfdf5' : '#fffbeb',
              border: syncFeedback.success && !syncFeedback.usedFallback ? '1px solid #10b981' : '1px solid #f59e0b',
              color: syncFeedback.success && !syncFeedback.usedFallback ? '#065f46' : '#92400e',
              boxShadow: '0 1px 3px rgba(0,0,0,0.05)',
            }}
          >
            {syncFeedback.success && !syncFeedback.usedFallback ? <CheckCircle2 size={16} /> : <AlertTriangle size={16} />}
            <span>{syncFeedback.message}</span>
          </div>
        )}

        {/* ── 1. HERO SHIPMENT IDENTITY HEADER ───────────────────────────────── */}
        <div className="sd-hero-title-row">
          <div className="sd-hero-identity">
            <div className="sd-hero-icon"><Ship size={24} /></div>
            <div>
              <div className="sd-hero-title-wrap">
                <h1 className="sd-hero-title">SH-{shipment.id}</h1>
                <div className="sd-hero-badges">
                  {/* Primary Authoritative Operational Status Badge */}
                  <span className="sd-status-pill" style={{ background: sCfg.bg, color: sCfg.fg, border: `1px solid ${sCfg.border || 'transparent'}` }}>
                    <span className="sd-status-dot" style={{ background: sCfg.dot }} />
                    {shipmentStatus.replace(/_/g, ' ')}
                  </span>

                  {/* Operational Exception/Delay Warning Tag (Only shown if tracking is delayed/exception) */}
                  {(trackingState === 'DELAYED' || trackingState === 'EXCEPTION' || trackingState === 'AT_RISK') && (
                    <span className="sd-track-pill" style={{ background: tsCfg.bg, color: tsCfg.fg, border: `1px solid ${tsCfg.border}` }}>
                      <AlertTriangle size={11} /> {tsCfg.label}
                    </span>
                  )}

                  {/* Financial Closure Tag (Only shown once closed & archived) */}
                  {closureStatus === 'CLOSED' && (
                    <span className="sd-closure-pill" style={{ background: cCfg.bg, color: cCfg.fg, border: `1px solid ${cCfg.border}` }}>
                      <CheckCircle2 size={11} /> Closed & Archived
                    </span>
                  )}
                </div>
              </div>

              {/* Clean Structured Metadata Chips */}
              <div className="sd-hero-meta-chips">
                <span className="sd-meta-chip">
                  <Package size={12} /> {containerCount} Container{containerCount !== 1 ? 's' : ''}
                </span>
                {shipment.customer_name && (
                  <span className="sd-meta-chip sd-chip-customer" title={shipment.customer_name}>
                    <User size={12} /> Customer: <strong>{shipment.customer_name}</strong>
                  </span>
                )}
                {shipment.carrier_scac && (
                  <span className="sd-meta-chip sd-chip-carrier">
                    <Ship size={12} /> Carrier: <strong>{shipment.carrier_scac}</strong>
                  </span>
                )}
                <span className="sd-meta-chip sd-chip-date">
                  <Calendar size={12} /> Created: <strong>{formatDateOnly(shipment.created_at)}</strong>
                </span>
                {shipment.updated_at && (
                  <span className="sd-meta-chip sd-chip-date">
                    <Clock size={12} /> Updated: <strong>{formatDate(shipment.updated_at)}</strong>
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* ── 2. OPERATIONAL SUMMARY STRIP ───────────────────────────────────── */}
        <div className="sd-summary-strip">
          <div className="sd-sum-item sd-sum-progress-item">
            <span className="sd-sum-label">PROGRESS</span>
            <div className="sd-sum-progress-wrap">
              <div
                className="sd-sum-progress-bar"
                style={{
                  width: `${progressPct}%`,
                  background: trackingState === 'DELAYED' || trackingState === 'EXCEPTION'
                    ? '#dc2626'
                    : trackingState === 'AT_RISK'
                      ? '#d97706'
                      : 'linear-gradient(90deg, #3b82f6, #2563eb)'
                }}
              />
            </div>
            <span className="sd-sum-val sd-sum-pct">{progressPct}%</span>
          </div>

          <div className="sd-sum-divider" />

          {/* Clean Unified Route Corridor */}
          <div className="sd-sum-item sd-sum-route-item">
            <span className="sd-sum-label">ROUTE CORRIDOR</span>
            <div className="sd-sum-route-row">
              <div className="sd-sum-port-unit">
                <span className="sd-sum-port-code origin">{originPort.code}</span>
                <span className="sd-sum-port-city">{originPort.name || 'Origin'}</span>
              </div>
              <span className="sd-sum-route-arrow">→</span>
              <div className="sd-sum-port-unit">
                <span className="sd-sum-port-code dest">{destPort.code}</span>
                <span className="sd-sum-port-city">{destPort.name || 'Destination'}</span>
              </div>
            </div>
          </div>

          <div className="sd-sum-divider" />

          <div className="sd-sum-item">
            <span className="sd-sum-label">ETD</span>
            <span className="sd-sum-val">
              {formatDateOnly(
                trackingSummary?.actual_etd ||
                shipment.actual_departure ||
                trackingSummary?.planned_etd ||
                shipment.planned_etd ||
                shipment.etd ||
                milestones.find(m => m.milestone_code === 'DEPARTURE' || m.milestone_code === 'VESSEL_DEPARTURE')?.actual_date ||
                milestones.find(m => m.milestone_code === 'DEPARTURE' || m.milestone_code === 'VESSEL_DEPARTURE')?.planned_date
              )}
            </span>
          </div>

          <div className="sd-sum-item">
            <span className="sd-sum-label">ETA</span>
            <span className="sd-sum-val">
              {formatDateOnly(
                trackingSummary?.actual_arrival ||
                shipment.actual_arrival ||
                trackingSummary?.planned_eta ||
                shipment.planned_eta ||
                shipment.eta ||
                milestones.find(m => m.milestone_code === 'ARRIVAL' || m.milestone_code === 'VESSEL_ARRIVAL' || m.milestone_code === 'DISCHARGE')?.actual_date ||
                milestones.find(m => m.milestone_code === 'ARRIVAL' || m.milestone_code === 'VESSEL_ARRIVAL' || m.milestone_code === 'DISCHARGE')?.planned_date
              )}
            </span>
          </div>

          <div className="sd-sum-divider" />

          <div className="sd-sum-item">
            <span className="sd-sum-label">EXCEPTIONS</span>
            {activeExceptionCount > 0 ? (
              <span className="sd-sum-val sd-sum-exc-active">
                <AlertTriangle size={12} /> {activeExceptionCount} Active
                {criticalExcCount > 0 && <span className="sd-sum-crit"> · {criticalExcCount} Critical</span>}
              </span>
            ) : (
              <span className="sd-sum-val sd-sum-ok">✓ None</span>
            )}
          </div>

          {trackingSummary?.schedule_variance != null && (
            <>
              <div className="sd-sum-divider" />
              <div className="sd-sum-item">
                <span className="sd-sum-label">VARIANCE</span>
                <span className={`sd-sum-val ${trackingSummary.schedule_variance > 0 ? 'sd-sum-late' : trackingSummary.schedule_variance < 0 ? 'sd-sum-early' : 'sd-sum-ok'}`}>
                  {trackingSummary.schedule_variance > 0
                    ? `+${trackingSummary.schedule_variance.toFixed(1)}d late`
                    : trackingSummary.schedule_variance < 0
                      ? `${Math.abs(trackingSummary.schedule_variance).toFixed(1)}d early`
                      : 'On Schedule'}
                </span>
              </div>
            </>
          )}

          <div className="sd-sum-divider" />

          <div
            className="sd-sum-item"
            style={{ cursor: 'pointer' }}
            onClick={() => document.getElementById('finance-workspace')?.scrollIntoView({ behavior: 'smooth' })}
            title="Jump to Financial Operations Workspace"
          >
            <span className="sd-sum-label">FINANCIALS</span>
            <span className="sd-sum-val" style={{ color: '#2563eb', display: 'flex', alignItems: 'center', gap: '3px' }}>
              <DollarSign size={12} /> Cost & Margins →
            </span>
          </div>
        </div>
      </div>

      {/* ── 3. LINEAGE / OPERATIONAL METADATA CARDS ──────────────────────────── */}
      <div className="sd-lineage-row">
        <div className="sd-lineage-card">
          <div className="sd-lineage-icon" style={{ background: '#eff6ff', color: '#2563eb' }}>
            <Package size={16} />
          </div>
          <div className="sd-lineage-body">
            <div className="sd-lineage-label">ASSOCIATED BOOKING</div>
            <div className="sd-lineage-val">
              <strong>{shipment.booking_number || '—'}</strong>
              {shipment.booking_id && (
                <Link to={`/dashboard/bookings/${shipment.booking_id}`} className="sd-lineage-link">
                  <LinkIcon size={11} /> View
                </Link>
              )}
            </div>
          </div>
        </div>

        <div className="sd-lineage-card">
          <div className="sd-lineage-icon" style={{ background: '#f5f3ff', color: '#7c3aed' }}>
            <User size={16} />
          </div>
          <div className="sd-lineage-body">
            <div className="sd-lineage-label">CUSTOMER & RFQ</div>
            <div className="sd-lineage-val">
              <strong title={shipment.customer_name}>{shipment.customer_name || '—'}</strong>
              {shipment.rfq_id && (
                <Link to={`/dashboard/rfqs/${shipment.rfq_id}`} className="sd-lineage-link">
                  <LinkIcon size={11} /> {shipment.rfq_number || 'RFQ'}
                </Link>
              )}
            </div>
          </div>
        </div>

        <div className="sd-lineage-card">
          <div className="sd-lineage-icon" style={{ background: '#ecfdf5', color: '#059669' }}>
            <Ship size={16} />
          </div>
          <div className="sd-lineage-body">
            <div className="sd-lineage-label">VESSEL & VOYAGE</div>
            <div className="sd-lineage-val">
              <strong>{shipment.vessel_name || 'Unassigned'}</strong>
              {shipment.voyage_number && <span className="sd-lineage-sub">Voy: {shipment.voyage_number}</span>}
            </div>
          </div>
        </div>

        <div className="sd-lineage-card">
          <div className="sd-lineage-icon" style={{ background: '#fffbeb', color: '#d97706' }}>
            <FileText size={16} />
          </div>
          <div className="sd-lineage-body">
            <div className="sd-lineage-label">B/L NUMBERS</div>
            <div className="sd-lineage-val">
              {shipment.mbl_number ? <strong>MBL: {shipment.mbl_number}</strong> : <span className="sd-lineage-empty">MBL: —</span>}
              {shipment.hbl_number ? <span className="sd-lineage-sub">HBL: {shipment.hbl_number}</span> : <span className="sd-lineage-empty">HBL: —</span>}
            </div>
          </div>
        </div>
      </div>

      {/* ── ACTIVE EXCEPTIONS BANNER (When critical/active) ──────────────────── */}
      {activeExceptionCount > 0 && (
        <div className="sd-exc-alert-banner">
          <div className="sd-exc-alert-icon">⚠️</div>
          <div className="sd-exc-alert-content">
            <strong>{activeExceptionCount} Active Exception{activeExceptionCount !== 1 ? 's' : ''}</strong>
            {criticalExcCount > 0 && <span className="sd-exc-alert-crit"> · {criticalExcCount} Critical</span>}
            <p>Review and resolve exceptions in the panel below to unblock operations and ensure on-time delivery.</p>
          </div>
        </div>
      )}

      {/* ── 4. TRANSIT JOURNEY VISUALIZATION ─────────────────────────────────── */}
      <div className="sd-journey-card">
        <div className="sd-journey-header">
          <div>
            <h3 className="sd-journey-title">Transit Journey</h3>
            <p className="sd-journey-sub">
              {shipment.carrier_scac && <strong>{shipment.carrier_scac}</strong>}
              {shipment.vessel_name && <> · {shipment.vessel_name}</>}
              {shipment.voyage_number && <> (Voyage {shipment.voyage_number})</>}
              {' · Direct Ocean Route'}
            </p>
          </div>
          <div className="sd-journey-meta">
            {progressPct >= 100 ? (
              <span className="sd-journey-status-badge sd-status-completed">
                <CheckCircle2 size={13} /> Destination Reached (100%)
              </span>
            ) : (
              <span className="sd-journey-status-badge sd-status-active">
                <Ship size={13} /> In Transit · {progressPct}%
              </span>
            )}
          </div>
        </div>

        {/* ── Continuous Route Flow Container ── */}
        <div className="sd-journey-route-box">
          {/* Left: Origin Port */}
          <div className="sd-jport-side sd-jport-origin">
            <div className="sd-jport-code-row">
              <span className="sd-jport-dot origin" />
              <span className="sd-jport-code">{originPort.code}</span>
              {originPort.country && <span className="sd-jport-country">{originPort.country.toUpperCase()}</span>}
            </div>
            <div className="sd-jport-city">{originPort.name || 'Origin Port'}</div>
            <div className="sd-jport-date">
              <span className="sd-jport-date-lbl">
                {trackingSummary?.actual_etd || shipment.actual_departure ? 'Departed:' : 'Planned ETD:'}
              </span>{' '}
              <strong className="sd-jport-date-val">
                {formatDateOnly(
                  trackingSummary?.actual_etd ||
                  shipment.actual_departure ||
                  trackingSummary?.planned_etd ||
                  shipment.planned_etd ||
                  shipment.etd ||
                  milestones.find(m => m.milestone_code === 'DEPARTURE' || m.milestone_code === 'VESSEL_DEPARTURE')?.actual_date ||
                  milestones.find(m => m.milestone_code === 'DEPARTURE' || m.milestone_code === 'VESSEL_DEPARTURE')?.planned_date
                )}
              </strong>
            </div>
          </div>

          {/* Center: Transit Ocean Corridor */}
          <div className="sd-jcorridor-center">
            <div className="sd-jcorridor-vessel">
              <Ship size={14} className="sd-vessel-icon" />
              <span className="sd-vessel-text">{shipment.vessel_name || 'Vessel Assigned'}</span>
              {shipment.voyage_number && <span className="sd-voy-sub">· Voy {shipment.voyage_number}</span>}
            </div>

            <div className="sd-jtrack-line-container">
              <div className="sd-jtrack-bg-line" />
              <div
                className="sd-jtrack-fill-line"
                style={{
                  width: `${progressPct}%`,
                  background: trackingState === 'DELAYED' || trackingState === 'EXCEPTION'
                    ? '#dc2626'
                    : trackingState === 'AT_RISK'
                      ? '#d97706'
                      : 'linear-gradient(90deg, #3b82f6 0%, #2563eb 100%)'
                }}
              />
              <div
                className="sd-jtrack-ship-marker"
                style={{ left: `${Math.min(Math.max(progressPct, 4), 96)}%` }}
                title={`Vessel Position: ${progressPct}%`}
              >
                <Ship size={12} />
              </div>
            </div>

            <div className="sd-jcorridor-footer">
              <span>🌊 Direct Ocean Corridor</span>
              <span>·</span>
              <span>{progressPct >= 100 ? 'Port Discharged' : 'In Transit'}</span>
              {trackingSummary?.schedule_variance != null && (
                <>
                  <span>·</span>
                  <span className={trackingSummary.schedule_variance > 0 ? 'sd-var-late' : 'sd-var-ok'}>
                    {trackingSummary.schedule_variance > 0
                      ? `+${trackingSummary.schedule_variance.toFixed(1)}d late`
                      : 'On Schedule'}
                  </span>
                </>
              )}
            </div>
          </div>

          {/* Right: Destination Port */}
          <div className="sd-jport-side sd-jport-dest">
            <div className="sd-jport-code-row">
              <span className="sd-jport-dot dest" />
              <span className="sd-jport-code">{destPort.code}</span>
              {destPort.country && <span className="sd-jport-country">{destPort.country.toUpperCase()}</span>}
            </div>
            <div className="sd-jport-city">{destPort.name || 'Destination Port'}</div>
            <div className="sd-jport-date">
              <span className="sd-jport-date-lbl">
                {trackingSummary?.actual_arrival || shipment.actual_arrival ? 'Arrived:' : 'Planned ETA:'}
              </span>{' '}
              <strong className="sd-jport-date-val">
                {formatDateOnly(
                  trackingSummary?.actual_arrival ||
                  shipment.actual_arrival ||
                  trackingSummary?.planned_eta ||
                  shipment.planned_eta ||
                  shipment.eta ||
                  milestones.find(m => m.milestone_code === 'ARRIVAL' || m.milestone_code === 'VESSEL_ARRIVAL' || m.milestone_code === 'DISCHARGE')?.actual_date ||
                  milestones.find(m => m.milestone_code === 'ARRIVAL' || m.milestone_code === 'VESSEL_ARRIVAL' || m.milestone_code === 'DISCHARGE')?.planned_date
                )}
              </strong>
            </div>
          </div>
        </div>
      </div>

      {/* ── 5. MAIN TWO-COLUMN OPERATIONAL LAYOUT ────────────────────────────── */}
      <div className="sd-main-grid">

        {/* ── LEFT COLUMN: WORKSPACES ─────────────────────────────────────────── */}
        <div className="sd-left-col">

          {/* 1. Operational Milestones */}
          <div className="sd-panel" id="milestones-panel">
            <div className="sd-panel-header">
              <div className="sd-panel-title-row">
                <div>
                  <h3 className="sd-panel-title">Operational Milestones</h3>
                  <p className="sd-panel-desc">Real-time carrier milestone monitoring from electronic tracking updates.</p>
                </div>
                {milestones.length > 0 && (
                  <span className="sd-milestone-count-badge">
                    {milestones.filter(m => m.status === 'COMPLETED').length} of {milestones.length} Completed
                  </span>
                )}
              </div>
            </div>

            <div className="sd-milestones-list">
              {milestones.length === 0 ? (
                <div className="sd-panel-empty-box">
                  <Calendar size={22} className="sd-empty-icon-muted" />
                  <div className="sd-empty-title">Milestones Not Available</div>
                  <p className="sd-empty-text">Milestone information will appear when carrier updates are received for this shipment.</p>
                </div>
              ) : (
                milestones.map((m, idx) => {
                  const done = m.status === 'COMPLETED';
                  const isCurrent = !done && (idx === 0 || milestones[idx - 1]?.status === 'COMPLETED');
                  const isDelayed = !done && m.planned_date && new Date(m.planned_date) < new Date();

                  return (
                    <div
                      key={m.id}
                      className={`sd-milestone-node ${
                        done
                          ? 'sd-ms-done'
                          : isCurrent
                            ? 'sd-ms-active-node'
                            : 'sd-ms-pending'
                      }`}
                    >
                      <div className="sd-ms-connector">
                        <div
                          className={`sd-ms-dot ${
                            done
                              ? 'sd-ms-dot-done'
                              : isCurrent
                                ? 'sd-ms-dot-active'
                                : 'sd-ms-dot-pending'
                          }`}
                        >
                          {done ? <CheckCircle2 size={14} /> : <span>{idx + 1}</span>}
                        </div>
                        {idx < milestones.length - 1 && (
                          <div className={`sd-ms-line ${done ? 'sd-line-done' : ''}`} />
                        )}
                      </div>

                      <div className="sd-ms-content">
                        <div className="sd-ms-header">
                          <span className="sd-ms-code">{m.milestone_code.replace(/_/g, ' ')}</span>
                          <span
                            className={`sd-ms-badge ${
                              done
                                ? 'sd-ms-badge-done'
                                : isDelayed
                                  ? 'sd-ms-badge-delayed'
                                  : isCurrent
                                    ? 'sd-ms-badge-active'
                                    : 'sd-ms-badge-pending'
                            }`}
                          >
                            {done ? 'COMPLETED' : isDelayed ? 'DELAYED' : isCurrent ? 'ACTIVE' : 'PENDING'}
                          </span>

                          {!done && (
                            <button
                              className="sd-ms-mark-btn"
                              onClick={() => handleOpenMilestoneForm(m)}
                            >
                              Mark Complete
                            </button>
                          )}
                        </div>

                        {m.description && <p className="sd-ms-desc">{m.description}</p>}

                        <div className="sd-ms-dates">
                          <span className="sd-ms-planned-date">
                            Planned: <strong>{formatDate(m.planned_date)}</strong>
                          </span>
                          {done && (
                            <span className="sd-ms-actual-date">
                              ✓ Actual: <strong>{formatDate(m.actual_date)}</strong>
                            </span>
                          )}
                          {m.location && (
                            <span className="sd-ms-loc">
                              <MapPin size={11} /> {getPortDetails(m.location).fullName || m.location}
                            </span>
                          )}
                        </div>

                        {activeMilestoneForm?.id === m.id && (
                          <form onSubmit={handleUpdateMilestone} className="sd-ms-form">
                            <div className="sd-ms-form-header">
                              <strong>Complete Milestone: {m.milestone_code.replace(/_/g, ' ')}</strong>
                            </div>
                            <div className="sd-ms-form-grid">
                              <div className="sd-form-field">
                                <label>Actual Date & Time *</label>
                                <input
                                  type="datetime-local"
                                  required
                                  value={actualDate}
                                  onChange={(e) => setActualDate(e.target.value)}
                                />
                              </div>
                              <div className="sd-form-field">
                                <label>Location</label>
                                <input
                                  type="text"
                                  placeholder="e.g. INNSA"
                                  value={milestoneLocation}
                                  onChange={(e) => setMilestoneLocation(e.target.value)}
                                />
                              </div>
                            </div>
                            <div className="sd-form-field">
                              <label>Operational Notes</label>
                              <textarea
                                rows={2}
                                placeholder="Carrier logs or verification notes..."
                                value={milestoneNotes}
                                onChange={(e) => setMilestoneNotes(e.target.value)}
                              />
                            </div>
                            <div className="sd-ms-form-actions">
                              <button
                                type="button"
                                className="sd-form-btn-cancel"
                                onClick={() => setActiveMilestoneForm(null)}
                              >
                                Cancel
                              </button>
                              <button
                                type="submit"
                                className="sd-form-btn-save"
                                disabled={submittingMilestone}
                              >
                                {submittingMilestone ? 'Saving...' : 'Save Milestone'}
                              </button>
                            </div>
                          </form>
                        )}
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>

          {/* 2. Document & Compliance Workspace */}
          <DocumentWorkspace shipmentId={id} shipment={shipment} onRefreshShipment={() => fetchShipmentDetails(false)} />

          {/* 3. Finance & Reconciliation Workspaces */}
          <FinanceWorkspace shipmentId={id} shipment={shipment} onRefreshShipment={() => fetchShipmentDetails(false)} />
          <BillingWorkspace shipmentId={id} onRefreshShipment={() => fetchShipmentDetails(false)} />
        </div>

        {/* ── RIGHT COLUMN: OPERATIONAL INTELLIGENCE SIDEBAR ────────────────── */}
        <div className="sd-right-col-wrapper">
          <div
            className="sd-right-col"
            ref={rightColRef}
            onScroll={checkRightColScroll}
          >

            {/* A. Tracking Intelligence */}
            <div className="sd-panel sd-panel-tracking">
              <div className="sd-panel-header">
                <div className="sd-panel-title-row">
                  <h3 className="sd-panel-title">Tracking Intelligence</h3>
                  <span
                    className="sd-track-state-badge"
                    style={{ background: tsCfg.bg, color: tsCfg.fg, border: `1px solid ${tsCfg.border}` }}
                  >
                    {tsCfg.label}
                  </span>
                </div>
                <p className="sd-panel-desc">Real-time container journey telemetry and schedule variance metrics.</p>
              </div>

              {trackingSummary ? (
                <>
                  <div className="sd-track-progress-section">
                    <div className="sd-track-prog-header">
                      <span>Journey Progress</span>
                      <strong>{progressPct}%</strong>
                    </div>
                    <div className="sd-track-prog-track">
                      <div
                        className="sd-track-prog-fill"
                        style={{
                          width: `${progressPct}%`,
                          background:
                            trackingState === 'DELAYED' || trackingState === 'EXCEPTION'
                              ? '#dc2626'
                              : trackingState === 'AT_RISK'
                                ? '#d97706'
                                : 'linear-gradient(90deg, #3b82f6, #2563eb)'
                        }}
                      />
                    </div>
                  </div>

                  <div className="sd-track-dates">
                    <div className="sd-track-date-row">
                      <span className="sd-track-date-lbl">PLANNED ETD</span>
                      <span className="sd-track-date-val">{formatDateOnly(trackingSummary.planned_etd)}</span>
                    </div>
                    <div className="sd-track-date-row">
                      <span className="sd-track-date-lbl">ACTUAL DEPARTURE</span>
                      <span
                        className={`sd-track-date-val ${
                          trackingSummary.actual_etd > trackingSummary.planned_etd ? 'sd-date-late' : 'sd-date-ok'
                        }`}
                      >
                        {formatDateOnly(trackingSummary.actual_etd)}
                      </span>
                    </div>
                    <div className="sd-track-date-row">
                      <span className="sd-track-date-lbl">PLANNED ETA</span>
                      <span className="sd-track-date-val">{formatDateOnly(trackingSummary.planned_eta)}</span>
                    </div>
                    <div className="sd-track-date-row">
                      <span className="sd-track-date-lbl">ACTUAL ARRIVAL</span>
                      <span
                        className={`sd-track-date-val ${
                          trackingSummary.actual_arrival && trackingSummary.actual_arrival > trackingSummary.planned_eta
                            ? 'sd-date-late'
                            : 'sd-date-ok'
                        }`}
                      >
                        {formatDateOnly(trackingSummary.actual_arrival)}
                      </span>
                    </div>

                    {trackingSummary.schedule_variance != null && (
                      <div className="sd-track-date-row sd-track-variance">
                        <span className="sd-track-date-lbl">SCHEDULE VARIANCE</span>
                        <span
                          className={`sd-track-date-val ${
                            trackingSummary.schedule_variance > 0
                              ? 'sd-date-late'
                              : trackingSummary.schedule_variance < 0
                                ? 'sd-date-early'
                                : 'sd-date-ok'
                          }`}
                        >
                          {trackingSummary.schedule_variance > 0
                            ? `+${trackingSummary.schedule_variance.toFixed(1)} days late`
                            : trackingSummary.schedule_variance < 0
                              ? `${Math.abs(trackingSummary.schedule_variance).toFixed(1)} days early`
                              : 'On Schedule'}
                        </span>
                      </div>
                    )}
                  </div>
                </>
              ) : (
                <div className="sd-panel-empty-box">
                  <Clock size={20} className="sd-empty-icon-muted" />
                  <div className="sd-empty-title">Tracking Information Unavailable</div>
                  <p className="sd-empty-text">Tracking updates have not been received for this shipment yet.</p>
                </div>
              )}
            </div>

            {/* B. Closure Lifecycle */}
            {trackingSummary && (
              <div className="sd-panel sd-panel-closure" style={{ borderLeft: `4px solid ${cCfg.fg}` }}>
                <div className="sd-panel-header">
                  <div className="sd-panel-title-row">
                    <h3 className="sd-panel-title">Closure Lifecycle</h3>
                    <span
                      className="sd-closure-state-badge"
                      style={{ background: cCfg.bg, color: cCfg.fg, border: `1px solid ${cCfg.border}` }}
                    >
                      {cCfg.label}
                    </span>
                  </div>
                  <p className="sd-panel-desc">Controlled operational completion with automated audit readiness.</p>
                </div>

                <div className="sd-closure-actions">
                  {closureStatus === 'ACTIVE' && (
                    <button
                      className="sd-action-btn sd-action-primary"
                      onClick={handleEvaluateClosure}
                      disabled={submittingClosure}
                    >
                      {submittingClosure ? 'Evaluating...' : 'Evaluate Closure Eligibility'}
                    </button>
                  )}

                  {closureStatus === 'READY_FOR_CLOSURE' && (
                    <button
                      className="sd-action-btn sd-action-success"
                      onClick={handleCompleteShipment}
                      disabled={submittingClosure}
                    >
                      {submittingClosure ? 'Processing...' : '✓ Complete & Close File'}
                    </button>
                  )}

                  {closureStatus === 'BLOCKED_BY_EXCEPTION' && (
                    <div className="sd-closure-blocked">
                      <div className="sd-closure-blocked-msg">
                        ⚠️ <strong>Closure Blocked</strong> — Resolve all critical exceptions first.
                      </div>
                      <button
                        className="sd-action-btn sd-action-primary"
                        onClick={handleEvaluateClosure}
                        disabled={submittingClosure}
                      >
                        {submittingClosure ? 'Evaluating...' : 'Re-Evaluate Eligibility'}
                      </button>
                    </div>
                  )}

                  {closureStatus === 'CLOSED' && (
                    <button
                      className="sd-action-btn sd-action-secondary"
                      onClick={handleReopenShipment}
                      disabled={submittingClosure}
                    >
                      {submittingClosure ? 'Processing...' : 'Reopen Shipment File'}
                    </button>
                  )}
                </div>
              </div>
            )}

            {/* C. Container & Cargo Summary */}
            <div className="sd-panel">
              <div className="sd-panel-header">
                <h3 className="sd-panel-title">Container & Cargo Summary</h3>
                <p className="sd-panel-desc">Physical carrier manifest and equipment assignments.</p>
              </div>
              <div className="sd-metadata-grid">
                <div className="sd-meta-item">
                  <div className="sd-meta-label">CARRIER SCAC</div>
                  <div className="sd-meta-val"><span className="sd-scac-tag">{shipment.carrier_scac || '—'}</span></div>
                </div>
                <div className="sd-meta-item">
                  <div className="sd-meta-label">VESSEL</div>
                  <div className="sd-meta-val">🚢 {shipment.vessel_name || 'Unassigned'}</div>
                </div>
                <div className="sd-meta-item">
                  <div className="sd-meta-label">VOYAGE NUMBER</div>
                  <div className="sd-meta-val">{shipment.voyage_number || '—'}</div>
                </div>
                <div className="sd-meta-item">
                  <div className="sd-meta-label">MBL NUMBER</div>
                  <div className="sd-meta-val">{shipment.mbl_number || '—'}</div>
                </div>
                <div className="sd-meta-item">
                  <div className="sd-meta-label">HBL NUMBER</div>
                  <div className="sd-meta-val">{shipment.hbl_number || '—'}</div>
                </div>
                <div className="sd-meta-item sd-meta-full">
                  <div className="sd-meta-label">CONTAINERS ASSIGNED</div>
                  <div className="sd-containers-wrap">
                    {containers.length > 0 ? (
                      containers.map((c) => (
                        <span key={c} className="sd-container-tag">📦 {c}</span>
                      ))
                    ) : (
                      <span className="sd-meta-empty">No container equipment assigned yet.</span>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* D. Anomalies & Exceptions */}
            <div className="sd-panel" id="exceptions-panel">
              <div className="sd-panel-header">
                <div className="sd-panel-title-row">
                  <h3 className="sd-panel-title">Anomalies & Exceptions</h3>
                  <button
                    className="sd-diagnostics-btn"
                    onClick={handleEvaluateExceptions}
                    disabled={runningDiagnostics}
                  >
                    <Zap size={12} className={runningDiagnostics ? 'sd-spin' : ''} />
                    {runningDiagnostics ? 'Checking...' : 'Run Diagnostics'}
                  </button>
                </div>

                <div className="sd-exc-pills">
                  {activeExceptionCount > 0 && (
                    <span className="sd-exc-pill sd-pill-active">{activeExceptionCount} Active</span>
                  )}
                  {criticalExcCount > 0 && (
                    <span className="sd-exc-pill sd-pill-critical">{criticalExcCount} Critical</span>
                  )}
                  {resolvedExcCount > 0 && (
                    <span className="sd-exc-pill sd-pill-resolved">{resolvedExcCount} Resolved</span>
                  )}
                  {exceptions.length === 0 && (
                    <span className="sd-exc-pill sd-pill-clean">✓ All Clear</span>
                  )}
                </div>

                <p className="sd-panel-desc">Operational risk queue from telemetry and automated delay detection.</p>
              </div>

              {/* Raise Manual Exception drawer */}
              <div className="sd-raise-exc-wrap">
                <button className="sd-raise-toggle-btn" onClick={() => setShowRaiseForm(!showRaiseForm)}>
                  <span>🚨 Raise Manual Exception</span>
                  <span>{showRaiseForm ? '▲ Close' : '▼ Expand'}</span>
                </button>
                {showRaiseForm && (
                  <form onSubmit={handleRaiseException} className="sd-raise-form">
                    <div className="sd-raise-form-grid">
                      <div className="sd-form-field">
                        <label>Category</label>
                        <CustomDropdown
                          value={newExcType}
                          onChange={(val) => setNewExcType(val)}
                          options={[
                            { value: 'SCHEDULE_DELAY', label: 'Schedule Delay' },
                            { value: 'VESSEL_ROLLOVER', label: 'Vessel Rollover' },
                            { value: 'PORT_CONGESTION', label: 'Port Congestion' },
                            { value: 'CUSTOMS_HOLD', label: 'Customs Hold' },
                            { value: 'DOCUMENT_ISSUE', label: 'Document Issue' },
                            { value: 'CARRIER_DELAY', label: 'Carrier Delay' },
                            { value: 'ROUTE_DEVIATION', label: 'Route Deviation' },
                            { value: 'CONTAINER_ISSUE', label: 'Container Issue' },
                            { value: 'OTHER', label: 'Other' },
                          ]}
                        />
                      </div>
                      <div className="sd-form-field">
                        <label>Severity</label>
                        <CustomDropdown
                          value={newExcSeverity}
                          onChange={(val) => setNewExcSeverity(val)}
                          options={[
                            { value: 'LOW', label: 'Low' },
                            { value: 'MEDIUM', label: 'Medium' },
                            { value: 'HIGH', label: 'High' },
                            { value: 'CRITICAL', label: 'Critical' },
                          ]}
                        />
                      </div>
                    </div>
                    <div className="sd-form-field">
                      <label>
                        Exception Title <small style={{ color: '#94a3b8', fontWeight: 500 }}>(Optional · auto-generated if left blank)</small>
                      </label>
                      <input
                        type="text"
                        value={newExcTitle}
                        onChange={(e) => setNewExcTitle(e.target.value)}
                        placeholder="e.g. Customs Clearance Hold at POD"
                      />
                    </div>
                    <div className="sd-form-field">
                      <label>Description</label>
                      <textarea
                        value={newExcDesc}
                        onChange={(e) => setNewExcDesc(e.target.value)}
                        rows={3}
                        placeholder="Enter operational notes or root cause details..."
                      />
                    </div>

                    {raiseError && (
                      <div className="sd-error-banner" style={{ margin: '4px 0 8px 0', fontSize: '0.74rem' }}>
                        <AlertCircle size={13} />
                        <span>{raiseError}</span>
                      </div>
                    )}

                    <button
                      type="submit"
                      className="sd-action-btn sd-action-danger"
                      disabled={submittingException}
                      style={{ opacity: submittingException ? 0.7 : 1 }}
                    >
                      {submittingException ? 'Creating Alert...' : 'Create Exception Alert'}
                    </button>
                  </form>
                )}
              </div>

              {/* Exceptions Queue */}
              {exceptions.length === 0 ? (
                <div className="sd-exc-empty">
                  <span className="sd-exc-empty-icon">💚</span>
                  <div className="sd-empty-title">No Active Exceptions</div>
                  <p>Carrier tracking is operating normally for this shipment.</p>
                </div>
              ) : (
                <div className="sd-exc-list">
                  {/* 1. Active Exceptions First */}
                  {exceptions
                    .filter((e) => e.status !== 'RESOLVED' && e.status !== 'DISMISSED')
                    .map((exc) => {
                      const isAcknowledged = exc.status === 'ACKNOWLEDGED';
                      const sevPillMap = {
                        CRITICAL: { bg: '#fee2e2', fg: '#dc2626', border: '#fecaca' },
                        HIGH:     { bg: '#fef3c7', fg: '#b45309', border: '#fde68a' },
                        MEDIUM:   { bg: '#dbeafe', fg: '#1d4ed8', border: '#bfdbfe' },
                        LOW:      { bg: '#f1f5f9', fg: '#475569', border: '#e2e8f0' },
                      };
                      const sev = sevPillMap[exc.severity] || sevPillMap.LOW;

                      return (
                        <div key={exc.id} className="sd-exc-card sd-exc-card-active">
                          <div className="sd-exc-card-header">
                            <span
                              className="sd-exc-sev"
                              style={{ background: sev.bg, color: sev.fg, border: `1px solid ${sev.border}` }}
                            >
                              {exc.severity}
                            </span>
                            <span
                              className="sd-exc-status"
                              style={{
                                background: isAcknowledged ? '#fef3c7' : '#fee2e2',
                                color: isAcknowledged ? '#b45309' : '#dc2626'
                              }}
                            >
                              {exc.status}
                            </span>
                            {exc.exception_type && (
                              <span className="sd-exc-type-tag">
                                {exc.exception_type.replace(/_/g, ' ')}
                              </span>
                            )}
                          </div>

                          <h4 className="sd-exc-title">{exc.title}</h4>
                          {exc.description && <p className="sd-exc-desc">{exc.description}</p>}

                          {exc.ai_summary && (
                            <div className="sd-exc-ai">
                              <strong>AI Analysis:</strong> {exc.ai_summary}
                            </div>
                          )}

                          <div className="sd-exc-footer">
                            <small>Logged: {formatDate(exc.created_at)}</small>
                            <div className="sd-exc-actions">
                              {!isAcknowledged && (
                                <button
                                  className="sd-exc-btn sd-exc-ack"
                                  onClick={() => handleAcknowledgeException(exc.id)}
                                >
                                  Acknowledge
                                </button>
                              )}
                              <button
                                className="sd-exc-btn sd-exc-resolve"
                                onClick={() => setResolvingExcId(exc.id)}
                              >
                                Resolve
                              </button>
                              <button
                                className="sd-exc-btn sd-exc-dismiss"
                                onClick={() => handleDismissException(exc.id)}
                              >
                                Dismiss
                              </button>
                            </div>
                          </div>

                          {resolvingExcId === exc.id && (
                            <form
                              onSubmit={(evt) => handleResolveExceptionSubmit(evt, exc.id)}
                              className="sd-resolve-form"
                            >
                              <label>Resolution Documentation Notes</label>
                              <textarea
                                value={resolutionNotesInput}
                                onChange={(evt) => setResolutionNotesInput(evt.target.value)}
                                placeholder="Enter actions taken to resolve this operational exception..."
                                rows={2}
                                required
                              />
                              <div className="sd-resolve-actions">
                                <button
                                  type="button"
                                  className="sd-form-btn-cancel"
                                  onClick={() => {
                                    setResolvingExcId(null);
                                    setResolutionNotesInput('');
                                  }}
                                >
                                  Cancel
                                </button>
                                <button type="submit" className="sd-form-btn-save">
                                  Submit Resolution
                                </button>
                              </div>
                            </form>
                          )}
                        </div>
                      );
                    })}

                  {/* 2. Resolved / Dismissed Exceptions Collapsible Section */}
                  {exceptions.filter((e) => e.status === 'RESOLVED' || e.status === 'DISMISSED').length > 0 && (
                    <div className="sd-resolved-exc-section">
                      <button
                        type="button"
                        className="sd-resolved-toggle-btn"
                        onClick={() => {
                          setShowResolvedHistory(!showResolvedHistory);
                          setTimeout(checkRightColScroll, 50);
                        }}
                      >
                        <span className="sd-resolved-toggle-title">
                          {showResolvedHistory ? '▾' : '▸'} Resolved Exceptions ({exceptions.filter((e) => e.status === 'RESOLVED' || e.status === 'DISMISSED').length})
                        </span>
                        <span className="sd-resolved-toggle-hint">
                          {showResolvedHistory ? 'Hide history' : 'View resolution notes'}
                        </span>
                      </button>

                      {showResolvedHistory && (
                        <div className="sd-resolved-exc-list">
                          {exceptions
                            .filter((e) => e.status === 'RESOLVED' || e.status === 'DISMISSED')
                            .map((exc) => {
                              const isResolved = exc.status === 'RESOLVED';
                              const sevPillMap = {
                                CRITICAL: { bg: '#fee2e2', fg: '#dc2626', border: '#fecaca' },
                                HIGH:     { bg: '#fef3c7', fg: '#b45309', border: '#fde68a' },
                                MEDIUM:   { bg: '#dbeafe', fg: '#1d4ed8', border: '#bfdbfe' },
                                LOW:      { bg: '#f1f5f9', fg: '#475569', border: '#e2e8f0' },
                              };
                              const sev = sevPillMap[exc.severity] || sevPillMap.LOW;

                              return (
                                <div key={exc.id} className="sd-exc-card sd-exc-card-inactive">
                                  <div className="sd-exc-card-header">
                                    <span
                                      className="sd-exc-sev"
                                      style={{ background: sev.bg, color: sev.fg, border: `1px solid ${sev.border}` }}
                                    >
                                      {exc.severity}
                                    </span>
                                    <span
                                      className="sd-exc-status"
                                      style={{
                                        background: isResolved ? '#dcfce7' : '#f1f5f9',
                                        color: isResolved ? '#15803d' : '#475569'
                                      }}
                                    >
                                      {exc.status}
                                    </span>
                                    {exc.exception_type && (
                                      <span className="sd-exc-type-tag">
                                        {exc.exception_type.replace(/_/g, ' ')}
                                      </span>
                                    )}
                                  </div>

                                  <h4 className="sd-exc-title">{exc.title}</h4>
                                  {exc.description && <p className="sd-exc-desc">{exc.description}</p>}

                                  {isResolved && exc.resolution_notes && (
                                    <div className="sd-exc-resolution">
                                      <strong>Resolution Notes:</strong> {exc.resolution_notes}
                                      <div className="sd-exc-by">Resolved by: {exc.resolved_by || 'Operator'}</div>
                                    </div>
                                  )}

                                  <div className="sd-exc-footer">
                                    <small>Logged: {formatDate(exc.created_at)}</small>
                                  </div>
                                </div>
                              );
                            })}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )}
            </div>

            {/* Dev: Manual Carrier Update Panel */}
            {isDev && (
              <ManualUpdatePanel
                onSubmit={handleSubmitCarrierUpdate}
                submittingUpdate={submittingUpdate}
                rawUpdate={rawUpdate}
                setRawUpdate={setRawUpdate}
                updateMsg={updateMsg}
              />
            )}

            {/* End of Section Marker */}
            <div className="sd-right-col-end-marker">
              <Check size={13} className="sd-end-icon" />
              <span>All Operational Modules Loaded</span>
            </div>
          </div>

          {/* Floating Scroll Down Indicator Pill */}
          {canScrollRightMore && (
            <button
              type="button"
              className="sd-scroll-down-pill"
              onClick={handleScrollRightColToBottom}
              aria-label="Scroll down for more information"
            >
              <ChevronDown size={14} className="sd-bounce-arrow" />
              <span>Scroll down for more</span>
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

