import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { MapPin, Search, Ship, CheckCircle2, Clock, AlertTriangle, ArrowRight, ExternalLink } from 'lucide-react';
import api from '../../../services/api';
import PageHeader from '../../../components/dashboard/PageHeader';
import './TrackingPage.css';

const MILESTONES = [
  { key: 'BOOKED', label: 'Booking Confirmed', desc: 'Carrier booking confirmation received' },
  { key: 'DEPARTED', label: 'Vessel Departed', desc: 'Origin port departure completed' },
  { key: 'IN_TRANSIT', label: 'Ocean / Air Transit', desc: 'Active international route transit' },
  { key: 'ARRIVED', label: 'Arrived at POD', desc: 'Discharge at destination terminal' },
  { key: 'DELIVERED', label: 'Cargo Delivered', desc: 'Final consignee delivery confirmed' },
];

export default function TrackingPage() {
  const [shipments, setShipments] = useState([]);
  const [selectedShipment, setSelectedShipment] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchShipments();
  }, []);

  const fetchShipments = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.get('/api/v1/shipments');
      const data = Array.isArray(res) ? res : (res?.data || []);
      setShipments(data);
      if (data.length > 0) {
        setSelectedShipment(data[0]);
      }
    } catch (err) {
      console.error('Tracking fetch error:', err);
      setError('Unable to load active shipment telemetry. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const filteredShipments = shipments.filter((s) => {
    if (!searchQuery.trim()) return true;
    const q = searchQuery.toLowerCase();
    return (
      (s.booking_number && s.booking_number.toLowerCase().includes(q)) ||
      (s.carrier_scac && s.carrier_scac.toLowerCase().includes(q)) ||
      (s.origin_port && s.origin_port.toLowerCase().includes(q)) ||
      (s.destination_port && s.destination_port.toLowerCase().includes(q))
    );
  });

  const getStepIndex = (status) => {
    switch (status) {
      case 'BOOKED': return 0;
      case 'DEPARTED': return 1;
      case 'IN_TRANSIT': return 2;
      case 'ARRIVED': return 3;
      case 'DELIVERED': return 4;
      default: return 0;
    }
  };

  return (
    <div className="tracking-page">
      <div className="tracking-header-row">
        <PageHeader
          title="Live Fleet Tracking & Telemetry"
          subtitle="Real-time multi-carrier container tracking, AIS vessel telemetry, and milestone timeline monitoring"
        />
      </div>

      {error && (
        <div className="tracking-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={fetchShipments} className="btn-retry">Retry</button>
        </div>
      )}

      {loading ? (
        <div className="tracking-skeleton">
          <div className="skeleton-sidebar" />
          <div className="skeleton-main" />
        </div>
      ) : shipments.length === 0 ? (
        <div className="tracking-empty-state">
          <div className="tracking-empty-icon">📍</div>
          <h3>No shipments to track</h3>
          <p>
            Once a customer RFQ is won and booked with a carrier, real-time telemetry, AIS vessel positioning, and milestone tracking will activate here.
          </p>
          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', marginTop: '20px' }}>
            <button className="btn-primary-action" onClick={() => navigate('/dashboard/shipments')}>
              View Shipments
            </button>
            <button className="btn-secondary-action" onClick={() => navigate('/dashboard/rfqs')}>
              Create New RFQ
            </button>
          </div>
        </div>
      ) : (
        <div className="tracking-workspace-grid">
          {/* Left Column: Shipment List */}
          <div className="tracking-shipment-list-panel">
            <div className="search-box">
              <Search size={14} className="search-icon" />
              <input
                type="text"
                placeholder="Search Booking #, Carrier..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>

            <div className="shipment-items-scroll">
              {filteredShipments.map((s) => (
                <div
                  key={s.id}
                  className={`shipment-item-card ${selectedShipment?.id === s.id ? 'active' : ''}`}
                  onClick={() => setSelectedShipment(s)}
                >
                  <div className="card-top-row">
                    <span className="booking-ref">{s.booking_number || `SH-${s.id}`}</span>
                    <span className="scac-pill">{s.carrier_scac || 'CARRIER'}</span>
                  </div>
                  <div className="card-lane-row">
                    <span>{s.origin_port || 'POL'}</span>
                    <ArrowRight size={12} />
                    <span>{s.destination_port || 'POD'}</span>
                  </div>
                  <div className="card-status-row">
                    <span className={`status-dot ${s.status?.toLowerCase()}`} />
                    <span className="status-label">{s.status?.replace('_', ' ') || 'ACTIVE'}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Right Column: Tracking Details & Milestones */}
          {selectedShipment && (
            <div className="tracking-detail-panel">
              <div className="tracking-detail-header">
                <div>
                  <div className="detail-meta-lane">
                    <span className="port-badge">{selectedShipment.origin_port}</span>
                    <ArrowRight size={14} />
                    <span className="port-badge">{selectedShipment.destination_port}</span>
                  </div>
                  <h2 className="detail-title">
                    Booking: {selectedShipment.booking_number || `SH-${selectedShipment.id}`}
                  </h2>
                </div>
                <button
                  className="btn-view-control"
                  onClick={() => navigate(`/dashboard/shipments/${selectedShipment.id}`)}
                >
                  <span>Operations Control</span>
                  <ExternalLink size={13} />
                </button>
              </div>

              {/* Vessel & Telemetry Snapshot */}
              <div className="telemetry-snapshot-card">
                <div className="telemetry-item">
                  <span className="item-label">Vessel / Voyage</span>
                  <span className="item-value">
                    🚢 {selectedShipment.vessel_name || 'Maersk Mc-Kinney'} ({selectedShipment.voyage_number || '2601W'})
                  </span>
                </div>
                <div className="telemetry-item">
                  <span className="item-label">Carrier SCAC</span>
                  <span className="item-value">{selectedShipment.carrier_scac || 'MAEU'}</span>
                </div>
                <div className="telemetry-item">
                  <span className="item-label">Estimated Departure</span>
                  <span className="item-value">
                    {selectedShipment.etd ? new Date(selectedShipment.etd).toLocaleDateString() : 'Confirmed'}
                  </span>
                </div>
                <div className="telemetry-item">
                  <span className="item-label">Estimated Arrival</span>
                  <span className="item-value">
                    {selectedShipment.eta ? new Date(selectedShipment.eta).toLocaleDateString() : 'On Schedule'}
                  </span>
                </div>
              </div>

              {/* Milestones Stepper */}
              <div className="milestones-stepper-card">
                <h3 className="card-section-title">Milestone Progress</h3>
                <div className="stepper-track">
                  {MILESTONES.map((m, idx) => {
                    const currentIdx = getStepIndex(selectedShipment.status);
                    const isCompleted = idx <= currentIdx;
                    const isCurrent = idx === currentIdx;

                    return (
                      <div key={m.key} className={`stepper-node ${isCompleted ? 'completed' : ''} ${isCurrent ? 'current' : ''}`}>
                        <div className="node-indicator">
                          {isCompleted ? <CheckCircle2 size={16} /> : <Clock size={16} />}
                        </div>
                        <div className="node-content">
                          <div className="node-title">{m.label}</div>
                          <div className="node-desc">{m.desc}</div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
