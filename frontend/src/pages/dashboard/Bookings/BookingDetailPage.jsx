import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import {
  Ship, Zap, ShieldCheck, RefreshCw, ArrowRight, Anchor, MapPin, 
  Package, FileText, CheckCircle2, Clock, AlertCircle, Info, Lock, 
  Globe, ExternalLink, Sliders, ChevronRight, Check
} from 'lucide-react';
import bookingService from '../../../services/bookingService';
import './BookingDetailPage.css';
import './BookingsPage.css';

export default function BookingDetailPage() {
  const { bookingId } = useParams();
  const navigate = useNavigate();

  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Status Action Modal State
  const [statusModalOpen, setStatusModalOpen] = useState(false);
  const [targetStatus, setTargetStatus] = useState('');
  const [statusNotes, setStatusNotes] = useState('');
  const [submittingStatus, setSubmittingStatus] = useState(false);

  // Shipment Handoff Modal State
  const [shipmentModalOpen, setShipmentModalOpen] = useState(false);
  const [containerInput, setContainerInput] = useState('');
  const [shipmentNotes, setShipmentNotes] = useState('');
  const [submittingShipment, setSubmittingShipment] = useState(false);

  // Carrier Direct Booking & Sync State (Task 5 & UI Enhancement)
  const [carrierBookModalOpen, setCarrierBookModalOpen] = useState(false);
  const [carrierInstructions, setCarrierInstructions] = useState('');
  const [movementType, setMovementType] = useState('CY_CY');
  const [releaseType, setReleaseType] = useState('WAYBILL');
  const [autoProvisionTracking, setAutoProvisionTracking] = useState(true);
  const [submittingCarrierBook, setSubmittingCarrierBook] = useState(false);
  const [syncingCarrier, setSyncingCarrier] = useState(false);

  const loadBookingDetail = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await bookingService.getBookingDetail(bookingId);
      const dataObj = res?.data || res;
      if (dataObj) {
        setDetail(dataObj);
      }
    } catch (err) {
      console.error('Failed to load booking detail:', err);
      setError(err.response?.data?.error?.message || 'Unable to load booking details.');
    } finally {
      setLoading(false);
    }
  }, [bookingId]);

  useEffect(() => {
    loadBookingDetail();
  }, [loadBookingDetail]);

  const handleOpenCarrierBookModal = () => {
    setCarrierInstructions(detail?.booking?.special_instructions || '');
    setCarrierBookModalOpen(true);
  };

  const handleCarrierBookSubmit = async (e) => {
    e.preventDefault();
    try {
      setSubmittingCarrierBook(true);
      const combinedInstructions = [
        carrierInstructions?.trim(),
        `[Movement: ${movementType.replace('_', '/')}]`,
        `[Release: ${releaseType}]`,
        autoProvisionTracking ? '[Auto-Tracking: ENABLED]' : ''
      ].filter(Boolean).join(' | ');

      await bookingService.bookWithCarrier(bookingId, {
        special_instructions: combinedInstructions,
        force_retry: !!detail?.booking?.carrier_booking_error
      });
      setCarrierBookModalOpen(false);
      await loadBookingDetail();
    } catch (err) {
      alert(err.response?.data?.error?.message || 'Failed to submit carrier booking.');
    } finally {
      setSubmittingCarrierBook(false);
    }
  };

  const handleSyncCarrier = async () => {
    try {
      setSyncingCarrier(true);
      await bookingService.syncCarrierBooking(bookingId);
      await loadBookingDetail();
    } catch (err) {
      alert(err.response?.data?.error?.message || 'Failed to sync carrier booking status.');
    } finally {
      setSyncingCarrier(false);
    }
  };

  const handleOpenStatusModal = (status) => {
    setTargetStatus(status);
    setStatusNotes('');
    setStatusModalOpen(true);
  };

  const handleStatusSubmit = async (e) => {
    e.preventDefault();
    try {
      setSubmittingStatus(true);
      await bookingService.updateBookingStatus(bookingId, targetStatus, statusNotes);
      setStatusModalOpen(false);
      loadBookingDetail();
    } catch (err) {
      alert(err.response?.data?.error?.message || 'Status transition failed.');
    } finally {
      setSubmittingStatus(false);
    }
  };

  const handleOpenShipmentModal = () => {
    setContainerInput(`MSKU${Date.now().toString().slice(-7)}`);
    setShipmentNotes('');
    setShipmentModalOpen(true);
  };

  const handleShipmentSubmit = async (e) => {
    e.preventDefault();
    try {
      setSubmittingShipment(true);
      const containers = containerInput
        ? containerInput.split(',').map((s) => s.trim()).filter(Boolean)
        : [];
      const res = await bookingService.createShipment(bookingId, {
        container_numbers: containers,
        notes: shipmentNotes || 'Shipment handoff created from confirmed booking',
      });
      setShipmentModalOpen(false);
      const resData = res?.data?.data || res?.data || res;
      if (resData?.id) {
        navigate(`/dashboard/shipments/${resData.id}`);
      } else {
        loadBookingDetail();
      }
    } catch (err) {
      alert(err.response?.data?.error?.message || 'Failed to create shipment.');
    } finally {
      setSubmittingShipment(false);
    }
  };

  if (loading) {
    return (
      <div className="booking-detail-page">
        <div className="skeleton-loading" style={{ marginTop: '80px' }}>
          <span className="skeleton-pulse">···</span>
          <p style={{ marginTop: '16px', fontWeight: 600 }}>Loading Carrier Booking Workspace...</p>
        </div>
      </div>
    );
  }

  if (error || !detail) {
    return (
      <div className="booking-detail-page">
        <div className="error-banner" style={{ marginTop: '40px' }}>
          <h3>Unable to Load Booking</h3>
          <p>{error || 'Booking record not found.'}</p>
          <Link to="/dashboard/bookings" className="btn-refresh" style={{ display: 'inline-block', marginTop: '12px', textDecoration: 'none' }}>
            ← Back to Bookings Workspace
          </Link>
        </div>
      </div>
    );
  }

  const { booking, source_rfq, commercial_quote, cargo_summary, linked_shipment, activity_events, allowed_actions } = detail;

  // Determine Lifecycle Stage Index (1 to 4)
  let activeStageStep = 3;
  if (booking.status === 'DRAFT') activeStageStep = 2;
  if (booking.status === 'REQUESTED' || booking.status === 'PENDING_CONFIRMATION') activeStageStep = 2;
  if (booking.status === 'CONFIRMED' && !linked_shipment) activeStageStep = 3;
  if (linked_shipment || booking.status === 'COMPLETED') activeStageStep = 4;

  return (
    <div className="booking-detail-page">
      {/* Breadcrumbs */}
      <div className="booking-breadcrumbs">
        <Link to="/dashboard">Dashboard</Link>
        <span className="breadcrumb-sep">/</span>
        <Link to="/dashboard/bookings">Bookings Workspace</Link>
        <span className="breadcrumb-sep">/</span>
        <span style={{ color: '#0f172a', fontWeight: 700 }}>{booking.booking_number}</span>
      </div>

      {/* Header */}
      <div className="booking-detail-header">
        <div className="booking-header-left">
          <div className="booking-title-row">
            <h1 className="booking-detail-title">{booking.booking_number}</h1>
            <span className={`status-pill ${booking.status.toLowerCase()}`}>
              <span className="status-dot"></span>
              {booking.status.replace(/_/g, ' ')}
            </span>
            <div className="booking-route-pill">
              <span>📍 {booking.origin_port}</span>
              <span>➔</span>
              <span>📍 {booking.destination_port}</span>
            </div>
            {linked_shipment && (
              <Link
                to={`/dashboard/shipments/${linked_shipment.id}`}
                className="booking-route-pill"
                style={{ background: '#e0f2fe', borderColor: '#bae6fd', color: '#0369a1', textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: '4px' }}
              >
                <span>📦 Active Shipment #{linked_shipment.id} ({linked_shipment.status}{linked_shipment.closure_status && linked_shipment.closure_status !== 'ACTIVE' ? ` - ${linked_shipment.closure_status}` : ''})</span>
              </Link>
            )}
          </div>
          <div style={{ fontSize: '0.88rem', color: '#64748b' }}>
            Carrier: <strong>{booking.carrier_name}</strong> {booking.carrier_scac && `(${booking.carrier_scac})`} | Customer: <strong>{source_rfq.customer_name}</strong>
          </div>
        </div>

        {/* Header Action Buttons */}
        <div className="booking-header-actions">
          {/* Live Carrier Direct API Submission */}
          {(!booking.carrier_booking_reference || booking.carrier_booking_status === 'REJECTED') && booking.status !== 'CANCELLED' && booking.status !== 'COMPLETED' && (
            <button
              className="btn-detail-action"
              style={{ background: '#2563eb', color: '#ffffff', display: 'inline-flex', alignItems: 'center', gap: '6px' }}
              onClick={handleOpenCarrierBookModal}
            >
              ⚡ Book with Carrier (Direct API)
            </button>
          )}

          {/* Live Carrier Sync Button */}
          {booking.carrier_booking_reference && (
            <button
              className="btn-refresh"
              style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', fontWeight: 600 }}
              onClick={handleSyncCarrier}
              disabled={syncingCarrier}
              title="Synchronize status and allocation with carrier"
            >
              🔄 {syncingCarrier ? 'Syncing...' : 'Sync Carrier Booking'}
            </button>
          )}

          {booking.status === 'DRAFT' && (
            <button className="btn-detail-action" onClick={() => handleOpenStatusModal('REQUESTED')}>
              📤 Request Space (Manual)
            </button>
          )}

          {(booking.status === 'REQUESTED' || booking.status === 'PENDING_CONFIRMATION') && (
            <button className="btn-detail-action confirm" onClick={() => handleOpenStatusModal('CONFIRMED')}>
              ✅ Confirm Space (Manual)
            </button>
          )}

          {booking.status === 'CONFIRMED' && !linked_shipment && (
            <button className="btn-detail-action shipment" onClick={handleOpenShipmentModal}>
              🚢 Handoff to Shipment Execution
            </button>
          )}

          {linked_shipment && (
            <Link
              to={`/dashboard/shipments/${linked_shipment.id}`}
              className="btn-detail-action shipment"
              style={{ textDecoration: 'none' }}
            >
              View Active Shipment #{linked_shipment.id} →
            </Link>
          )}

          {booking.status !== 'CANCELLED' && booking.status !== 'COMPLETED' && (
            <button
              className="btn-refresh"
              style={{ color: '#b91c1c', borderColor: '#fecaca' }}
              onClick={() => handleOpenStatusModal('CANCELLED')}
            >
              Cancel Booking
            </button>
          )}
        </div>
      </div>

      {/* 4-Stage Operational Progress Tracker */}
      <div className="lifecycle-tracker-card">
        <div style={{ fontSize: '0.78rem', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.05em', color: '#64748b', marginBottom: '8px' }}>
          Operational Handoff Lifecycle
        </div>
        <div className="lifecycle-stages">
          <div className="lifecycle-connector-line"></div>

          {/* Stage 1 */}
          <div className={`lifecycle-stage-step completed`}>
            <div className="stage-step-circle">✓</div>
            <div className="stage-step-name">1. Commercial RFQ</div>
            <div className="stage-step-status">{source_rfq.rfq_number}</div>
          </div>

          {/* Stage 2 */}
          <div className={`lifecycle-stage-step ${activeStageStep >= 2 ? (activeStageStep > 2 ? 'completed' : 'active') : ''}`}>
            <div className="stage-step-circle">{activeStageStep > 2 ? '✓' : '2'}</div>
            <div className="stage-step-name">2. Carrier Quote</div>
            <div className="stage-step-status">{commercial_quote ? `${commercial_quote.carrier_name}` : 'Approved'}</div>
          </div>

          {/* Stage 3 */}
          <div className={`lifecycle-stage-step ${activeStageStep >= 3 ? (activeStageStep > 3 ? 'completed' : 'active') : ''}`}>
            <div className="stage-step-circle">{activeStageStep > 3 ? '✓' : '3'}</div>
            <div className="stage-step-name">3. Space Confirmed</div>
            <div className="stage-step-status">{booking.status}</div>
          </div>

          {/* Stage 4 */}
          <div className={`lifecycle-stage-step ${activeStageStep === 4 ? 'completed' : ''}`}>
            <div className="stage-step-circle">{activeStageStep === 4 ? '✓' : '4'}</div>
            <div className="stage-step-name">4. Shipment Execution</div>
            <div className="stage-step-status">{linked_shipment ? `Shipment #${linked_shipment.id}` : 'Pending Handoff'}</div>
          </div>
        </div>
      </div>

      {/* Two-Column Grid Layout */}
      <div className="booking-detail-grid">
        {/* LEFT COLUMN: Carrier & Voyage Particulars */}
        <div className="detail-grid-left">
          {/* Card 1: Carrier & Voyage Specifics */}
          <div className="detail-card">
            <div className="detail-card-header">
              <h3 className="detail-card-title">
                <span>🚢 Carrier Space & Voyage Particulars</span>
              </h3>
              <span style={{ fontSize: '0.8rem', color: '#64748b', fontWeight: 600 }}>
                SCAC: <strong>{booking.carrier_scac || 'N/A'}</strong>
              </span>
            </div>

            {/* Live Carrier Integration Confirmation Block */}
            {booking.carrier_booking_reference && (
              <div style={{ background: '#f0fdf4', border: '1px solid #bbf7d0', borderRadius: '8px', padding: '12px 16px', margin: '0 16px 16px 16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '12px' }}>
                <div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 800, textTransform: 'uppercase', color: '#166534', letterSpacing: '0.04em' }}>
                    ⚡ Carrier Direct API Confirmation
                  </div>
                  <div style={{ fontSize: '1rem', fontWeight: 800, color: '#14532d', fontFamily: 'monospace', marginTop: '2px' }}>
                    {booking.carrier_booking_reference}
                  </div>
                  <div style={{ fontSize: '0.78rem', color: '#15803d', marginTop: '2px' }}>
                    Carrier Status: <strong>{booking.carrier_booking_status || 'CONFIRMED'}</strong>
                    {booking.carrier_confirmation_reference && ` | Allotment: ${booking.carrier_confirmation_reference}`}
                    {booking.carrier_booked_at && ` | Confirmed: ${new Date(booking.carrier_booked_at).toLocaleString()}`}
                  </div>
                </div>
                <button
                  type="button"
                  className="btn-refresh"
                  style={{ background: '#ffffff', color: '#15803d', borderColor: '#86efac', fontWeight: 600, fontSize: '0.78rem', padding: '4px 10px' }}
                  onClick={handleSyncCarrier}
                  disabled={syncingCarrier}
                >
                  {syncingCarrier ? '🔄 Syncing...' : '🔄 Sync Allocation'}
                </button>
              </div>
            )}

            {/* Carrier Rejection / Error Alert */}
            {booking.carrier_booking_error && (
              <div style={{ background: '#fef2f2', border: '1px solid #fecaca', borderRadius: '8px', padding: '12px 16px', margin: '0 16px 16px 16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '12px' }}>
                <div>
                  <div style={{ fontSize: '0.75rem', fontWeight: 800, textTransform: 'uppercase', color: '#991b1b', letterSpacing: '0.04em' }}>
                    ⚠️ Carrier Direct API Exception
                  </div>
                  <div style={{ fontSize: '0.85rem', color: '#b91c1c', marginTop: '2px' }}>
                    {booking.carrier_booking_error}
                  </div>
                </div>
                <button
                  type="button"
                  className="btn-detail-action"
                  style={{ background: '#dc2626', color: '#ffffff', fontSize: '0.78rem', padding: '4px 10px' }}
                  onClick={handleOpenCarrierBookModal}
                >
                  ⚡ Retry Booking
                </button>
              </div>
            )}

            <div className="booking-meta-grid">
              <div className="booking-meta-item">
                <span className="booking-meta-label">Carrier Name</span>
                <span className="booking-meta-val">{booking.carrier_name}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Internal Booking Number</span>
                <span className="booking-meta-val highlight" style={{ fontFamily: 'monospace' }}>{booking.booking_number}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Vessel Name</span>
                <span className="booking-meta-val">{booking.vessel_name || <em style={{ color: '#94a3b8' }}>TBA by carrier</em>}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Voyage Number</span>
                <span className="booking-meta-val">{booking.voyage_number || <em style={{ color: '#94a3b8' }}>TBA by carrier</em>}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Port of Loading (Origin)</span>
                <span className="booking-meta-val">{booking.origin_port}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Port of Discharge (Destination)</span>
                <span className="booking-meta-val">{booking.destination_port}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Estimated Departure (ETD)</span>
                <span className="booking-meta-val">{booking.etd ? new Date(booking.etd).toLocaleDateString() : 'TBA'}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Estimated Arrival (ETA)</span>
                <span className="booking-meta-val">{booking.eta ? new Date(booking.eta).toLocaleDateString() : 'TBA'}</span>
              </div>

              {linked_shipment && (
                <div className="booking-meta-item full-width">
                  <span className="booking-meta-label">Assigned Containers</span>
                  <div className="container-tag-list" style={{ marginTop: '4px' }}>
                    {linked_shipment.container_numbers && linked_shipment.container_numbers.length > 0 ? (
                      linked_shipment.container_numbers.map((c, i) => (
                        <span key={i} className="container-tag" style={{ background: '#e0f2fe', color: '#0369a1', padding: '2px 8px', borderRadius: '4px', fontSize: '0.8rem', fontWeight: 600, marginRight: '6px' }}>{c}</span>
                      ))
                    ) : (
                      <span style={{ fontSize: '0.82rem', color: '#64748b' }}>No containers logged</span>
                    )}
                  </div>
                </div>
              )}

              {booking.special_instructions && (
                <div className="booking-meta-item full-width">
                  <span className="booking-meta-label">Special Instructions</span>
                  <span className="booking-meta-val" style={{ fontSize: '0.88rem', color: '#475569' }}>
                    {booking.special_instructions}
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* RIGHT COLUMN: Cargo & Upstream Quotation Details */}
        <div className="detail-grid-right">
          {/* Card 2: Cargo & Packaging */}
          <div className="detail-card">
            <div className="detail-card-header">
              <h3 className="detail-card-title">
                <span>📦 Cargo & Consignment Summary</span>
              </h3>
              <span style={{ fontSize: '0.8rem', color: '#64748b', fontWeight: 600 }}>
                {cargo_summary.cargo_type}
              </span>
            </div>

            <div className="booking-meta-grid">
              <div className="booking-meta-item">
                <span className="booking-meta-label">Commodity Description</span>
                <span className="booking-meta-val">{cargo_summary.commodity}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Packaging Type</span>
                <span className="booking-meta-val">{cargo_summary.packaging_type}</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Total Cargo Weight</span>
                <span className="booking-meta-val highlight">{cargo_summary.total_weight_kg.toLocaleString()} kg</span>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Total Volume</span>
                <span className="booking-meta-val highlight">{cargo_summary.total_volume_cbm.toLocaleString()} CBM</span>
              </div>
            </div>
          </div>

          {/* Card 3: Commercial Upstream Lineage */}
          <div className="detail-card">
            <div className="detail-card-header">
              <h3 className="detail-card-title">
                <span>🔗 Commercial Upstream Quotation</span>
              </h3>
            </div>

            <div className="booking-meta-grid">
              <div className="booking-meta-item">
                <span className="booking-meta-label">Source Commercial RFQ</span>
                <Link to={`/dashboard/rfqs/${source_rfq.id}`} className="booking-meta-val highlight" style={{ textDecoration: 'none' }}>
                  {source_rfq.rfq_number} ↗
                </Link>
              </div>

              <div className="booking-meta-item">
                <span className="booking-meta-label">Customer</span>
                <span className="booking-meta-val">{source_rfq.customer_name}</span>
              </div>
            </div>

            {commercial_quote ? (
              <div className="commercial-lineage-box">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                  <span style={{ fontSize: '0.85rem', fontWeight: 700, color: '#0f172a' }}>
                    Selected Quote: {commercial_quote.quote_reference || `Quote #${commercial_quote.id}`}
                  </span>
                  <span className="margin-pill">
                    +{commercial_quote.margin_percent.toFixed(1)}% Margin
                  </span>
                </div>

                <div className="booking-meta-grid" style={{ gap: '12px' }}>
                  <div className="booking-meta-item">
                    <span className="booking-meta-label">Carrier Buy Rate</span>
                    <span className="booking-meta-val" style={{ color: '#475569' }}>
                      ${commercial_quote.buy_price.toLocaleString()} {commercial_quote.currency}
                    </span>
                  </div>

                  <div className="booking-meta-item">
                    <span className="booking-meta-label">Customer Sell Price</span>
                    <span className="booking-meta-val" style={{ color: '#16a34a', fontWeight: 800 }}>
                      ${commercial_quote.sell_price.toLocaleString()} {commercial_quote.currency}
                    </span>
                  </div>

                  <div className="booking-meta-item">
                    <span className="booking-meta-label">Net Profit Margin</span>
                    <span className="booking-meta-val" style={{ color: '#047857' }}>
                      +${commercial_quote.margin_amount.toLocaleString()}
                    </span>
                  </div>

                  <div className="booking-meta-item">
                    <span className="booking-meta-label">Approved By</span>
                    <span className="booking-meta-val">
                      {commercial_quote.approved_by || 'Commercial Team'}
                    </span>
                  </div>
                </div>
              </div>
            ) : (
              <p style={{ fontSize: '0.85rem', color: '#94a3b8', fontStyle: 'italic', marginTop: '12px' }}>
                No commercial quotation reference attached to this booking.
              </p>
            )}
          </div>
        </div>
      </div>

      {/* ── UPDATE STATUS MODAL ── */}
      {statusModalOpen && (
        <div className="modal-overlay" onClick={() => setStatusModalOpen(false)}>
          <div className="modal-card" style={{ maxWidth: '480px' }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">Transition Booking Status: {targetStatus}</h3>
              <button className="btn-close-modal" onClick={() => setStatusModalOpen(false)}>✕</button>
            </div>

            <form onSubmit={handleStatusSubmit}>
              <div className="modal-body">
                <p style={{ fontSize: '0.88rem', color: '#475569', margin: '0 0 16px 0' }}>
                  Are you sure you want to transition booking <strong>{booking.booking_number}</strong> from <strong>{booking.status}</strong> to <strong>{targetStatus}</strong>?
                </p>

                <div className="form-group">
                  <label className="form-label">Notes for Audit Trail</label>
                  <textarea
                    className="form-textarea"
                    rows="3"
                    placeholder="Enter reason or carrier confirmation notes..."
                    value={statusNotes}
                    onChange={(e) => setStatusNotes(e.target.value)}
                  />
                </div>
              </div>

              <div className="modal-footer">
                <button type="button" className="btn-refresh" onClick={() => setStatusModalOpen(false)}>
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-create-booking"
                  style={{ background: targetStatus === 'CANCELLED' ? '#b91c1c' : '#059669' }}
                  disabled={submittingStatus}
                >
                  {submittingStatus ? 'Updating...' : `Confirm Transition`}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── SHIPMENT HANDOFF MODAL ── */}
      {shipmentModalOpen && (
        <div className="modal-overlay" onClick={() => setShipmentModalOpen(false)}>
          <div className="modal-card" style={{ maxWidth: '520px' }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">Initiate Shipment Execution</h3>
              <button className="btn-close-modal" onClick={() => setShipmentModalOpen(false)}>✕</button>
            </div>

            <form onSubmit={handleShipmentSubmit}>
              <div className="modal-body">
                <div style={{ background: '#f8fafc', padding: '12px', borderRadius: '8px', marginBottom: '16px', fontSize: '0.85rem' }}>
                  Booking: <strong>{booking.booking_number}</strong> | Carrier: <strong>{booking.carrier_name}</strong> ({booking.carrier_scac})<br/>
                  Route: <strong>{booking.origin_port}</strong> ➔ <strong>{booking.destination_port}</strong>
                </div>

                <div className="form-group" style={{ marginBottom: '14px' }}>
                  <label className="form-label">Container Numbers (comma-separated) *</label>
                  <input
                    type="text"
                    className="form-input"
                    required
                    placeholder="e.g. MSKU9012345, MSKU9012346"
                    value={containerInput}
                    onChange={(e) => setContainerInput(e.target.value)}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Handoff Instructions / Notes</label>
                  <textarea
                    className="form-textarea"
                    rows="2"
                    placeholder="Initial dispatch notes for freight operations..."
                    value={shipmentNotes}
                    onChange={(e) => setShipmentNotes(e.target.value)}
                  />
                </div>
              </div>

              <div className="modal-footer">
                <button type="button" className="btn-refresh" onClick={() => setShipmentModalOpen(false)}>
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-create-booking"
                  style={{ background: '#0284c7' }}
                  disabled={submittingShipment}
                >
                  {submittingShipment ? 'Creating Shipment...' : 'Create Shipment'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── LIVE CARRIER BOOKING MODAL (Enhanced with Operational Use Case & Dispatch Controls) ── */}
      {carrierBookModalOpen && (
        <div className="modal-overlay" onClick={() => setCarrierBookModalOpen(false)}>
          <div className="carrier-book-modal-card" onClick={(e) => e.stopPropagation()}>
            {/* Modal Header */}
            <div className="carrier-book-modal-header">
              <div className="carrier-book-header-left">
                <div className="carrier-book-icon-badge">
                  <Zap size={22} className="carrier-zap-icon" />
                </div>
                <div>
                  <h3 className="carrier-book-title">Direct Carrier API Booking</h3>
                  <p className="carrier-book-subtitle">
                    Reserve vessel space, lock container allocation & generate official carrier booking confirmation via DCSA API.
                  </p>
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setCarrierBookModalOpen(false)}>✕</button>
            </div>

            <form onSubmit={handleCarrierBookSubmit}>
              <div className="carrier-book-modal-body">
                {/* 1. WHY DIRECT CARRIER BOOKING? (Operational Use Case Box) */}
                <div className="carrier-usecase-banner">
                  <div className="carrier-usecase-header">
                    <div className="carrier-usecase-tag">
                      <Info size={14} />
                      <span>Why Direct Carrier API Booking is Required</span>
                    </div>
                    <span className="carrier-standard-badge">DCSA v2 / EDI Standard</span>
                  </div>
                  <p className="carrier-usecase-text">
                    Direct integration with shipping line booking engines connects your commercial quotation directly to the carrier's space management systems. It secures confirmed slot allocations, eliminates manual booking email delays, locks negotiated contract/spot tariffs to prevent invoice discrepancies, and automatically triggers end-to-end container tracking milestones.
                  </p>
                  <div className="carrier-usecase-grid">
                    <div className="carrier-usecase-item">
                      <div className="usecase-item-icon green">
                        <CheckCircle2 size={16} />
                      </div>
                      <div>
                        <strong>Instant Space Lock</strong>
                        <span>Authoritative allotment on scheduled vessel voyages</span>
                      </div>
                    </div>
                    <div className="carrier-usecase-item">
                      <div className="usecase-item-icon blue">
                        <Lock size={16} />
                      </div>
                      <div>
                        <strong>Tariff & Rate Protection</strong>
                        <span>Locks quoted prices against spot rate surges</span>
                      </div>
                    </div>
                    <div className="carrier-usecase-item">
                      <div className="usecase-item-icon purple">
                        <RefreshCw size={16} />
                      </div>
                      <div>
                        <strong>Automated Telemetry</strong>
                        <span>Auto-provisions container milestones upon confirmation</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* 2. DISPATCH OVERVIEW CARD */}
                <div className="carrier-dispatch-card">
                  <div className="carrier-dispatch-top">
                    <div className="carrier-brand-info">
                      <div className="carrier-logo-chip" style={{
                        background: booking.carrier_scac === 'MAEU' ? '#e0f2fe' : booking.carrier_scac === 'HLCU' ? '#fee2e2' : booking.carrier_scac === 'MSCU' ? '#fef3c7' : '#f1f5f9',
                        color: booking.carrier_scac === 'MAEU' ? '#0284c7' : booking.carrier_scac === 'HLCU' ? '#dc2626' : booking.carrier_scac === 'MSCU' ? '#b45309' : '#334155'
                      }}>
                        {booking.carrier_scac || 'CA'}
                      </div>
                      <div>
                        <div className="carrier-line-name">{booking.carrier_name}</div>
                        <div className="carrier-scac-text">SCAC: {booking.carrier_scac || 'N/A'} • Environment: PRODUCTION (API)</div>
                      </div>
                    </div>
                    <div className="carrier-security-pill">
                      <ShieldCheck size={14} />
                      <span>Encrypted API Dispatch</span>
                    </div>
                  </div>

                  {/* Route Visual Diagram */}
                  <div className="carrier-route-diagram">
                    <div className="route-port origin">
                      <span className="route-port-label">PORT OF LOADING (POL)</span>
                      <strong className="route-port-code">{booking.origin_port}</strong>
                      <span className="route-port-date">ETD: {booking.etd ? new Date(booking.etd).toLocaleDateString() : 'Immediate Schedule'}</span>
                    </div>
                    <div className="route-connector">
                      <div className="route-line-dashed"></div>
                      <div className="route-vessel-badge">
                        <Ship size={14} />
                        <span>Direct Ocean Transit</span>
                      </div>
                    </div>
                    <div className="route-port dest">
                      <span className="route-port-label">PORT OF DISCHARGE (POD)</span>
                      <strong className="route-port-code">{booking.destination_port}</strong>
                      <span className="route-port-date">ETA: {booking.eta ? new Date(booking.eta).toLocaleDateString() : 'Calculated on Dispatch'}</span>
                    </div>
                  </div>

                  {/* Metrics Specs Grid */}
                  <div className="carrier-specs-row">
                    <div className="carrier-spec-col">
                      <span className="spec-label">EQUIPMENT & PACKAGING</span>
                      <strong className="spec-val">{cargo_summary?.packaging_type || '1x 40HC High Cube Dry'}</strong>
                    </div>
                    <div className="carrier-spec-col">
                      <span className="spec-label">GROSS WEIGHT / VOLUME</span>
                      <strong className="spec-val">{cargo_summary?.total_weight_kg ? `${cargo_summary.total_weight_kg.toLocaleString()} kg` : 'Standard Weight'} • {cargo_summary?.total_volume_cbm ? `${cargo_summary.total_volume_cbm} CBM` : 'FCL'}</strong>
                    </div>
                    <div className="carrier-spec-col">
                      <span className="spec-label">COMMERCIAL RATE LOCK</span>
                      <strong className="spec-val highlight">
                        {commercial_quote ? (
                          `$${Number(commercial_quote.sell_price || commercial_quote.total_sell_price || commercial_quote.buy_price || 0).toLocaleString()} (${commercial_quote.quote_reference || `Quote #${commercial_quote.id}`})`
                        ) : (
                          'Agreed Contract Rate'
                        )}
                      </strong>
                    </div>
                  </div>
                </div>

                {/* 3. DISPATCH CONFIGURATION CONTROLS */}
                <div className="carrier-form-section">
                  <div className="carrier-form-row">
                    <div className="form-group flex-1">
                      <label className="form-label">Service & Movement Mode</label>
                      <select 
                        className="form-select"
                        value={movementType}
                        onChange={(e) => setMovementType(e.target.value)}
                      >
                        <option value="CY_CY">CY / CY (Container Yard to Container Yard)</option>
                        <option value="DOOR_CY">Door / CY (Factory Pickup to Container Yard)</option>
                        <option value="CY_DOOR">CY / Door (Container Yard to Consignee Door)</option>
                        <option value="CFS_CFS">CFS / CFS (Container Freight Station)</option>
                      </select>
                    </div>

                    <div className="form-group flex-1">
                      <label className="form-label">Bill of Lading Documentation Release</label>
                      <select 
                        className="form-select"
                        value={releaseType}
                        onChange={(e) => setReleaseType(e.target.value)}
                      >
                        <option value="WAYBILL">Express Sea Waybill (Instant Electronic Release)</option>
                        <option value="ORIGINAL_BL">Original Ocean Bill of Lading (3/3 Paper OBL)</option>
                        <option value="TELEX">Telex / Electronic Cargo Release (EDR)</option>
                      </select>
                    </div>
                  </div>

                  <div className="form-group" style={{ marginTop: '12px' }}>
                    <label className="form-label">Special Instructions & Handling Notes for Carrier</label>
                    <textarea
                      className="form-textarea"
                      rows="3"
                      placeholder="e.g. Standard dry cargo, clean sweep container, direct discharge at destination terminal..."
                      value={carrierInstructions}
                      onChange={(e) => setCarrierInstructions(e.target.value)}
                    />
                  </div>

                  <div className="carrier-checkbox-row">
                    <label className="carrier-checkbox-label">
                      <input 
                        type="checkbox"
                        checked={autoProvisionTracking}
                        onChange={(e) => setAutoProvisionTracking(e.target.checked)}
                      />
                      <span>
                        <strong>Auto-Provision Real-Time Tracking Telemetry:</strong> Automatically stream live milestone events (Gate-in, Vessel Departure, Transshipment, Discharge) into LogisticsHQ once confirmed.
                      </span>
                    </label>
                  </div>
                </div>

                {/* 4. TRUST & SECURITY PROTOCOL */}
                <div className="carrier-protocol-footer">
                  <Lock size={12} />
                  <span>DCSA v2 / EDI Standard Protocol • Idempotent Reservation • Zero Rate Leakage Guarantee</span>
                </div>
              </div>

              {/* Modal Footer Buttons */}
              <div className="carrier-book-modal-footer">
                <button type="button" className="btn-modal-cancel" onClick={() => setCarrierBookModalOpen(false)}>
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-submit-carrier-direct"
                  disabled={submittingCarrierBook}
                >
                  {submittingCarrierBook ? (
                    <>
                      <RefreshCw size={15} className="spin" />
                      <span>Submitting to Carrier API...</span>
                    </>
                  ) : (
                    <>
                      <Zap size={16} />
                      <span>Confirm & Dispatch Carrier Booking</span>
                    </>
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
