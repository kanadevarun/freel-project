import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
import './Shipments.css';

export default function ShipmentsPage({ mode = 'shipments', defaultStatus = 'ALL' }) {
  const [shipments, setShipments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filterStatus, setFilterStatus] = useState(defaultStatus);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const isBookingsMode = mode === 'bookings';
  const pageTitle = isBookingsMode ? 'Carrier Bookings' : 'Shipments Tracker';
  const pageSubtitle = isBookingsMode
    ? 'Manage confirmed carrier bookings, release orders, and equipment allocations.'
    : 'Real-time container tracking, milestone Gantt monitoring, and automated exception detection.';

  const fetchShipments = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.get('/api/v1/shipments');
      const items = Array.isArray(res) ? res : (res?.data || []);
      setShipments(items);
    } catch (err) {
      console.error('Failed to fetch shipments:', err);
      setError(err?.message || 'We couldn\'t load your freight operations data.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchShipments();
  }, [fetchShipments]);

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'BOOKING_PENDING': return 'badge-booking-pending';
      case 'BOOKED': return 'badge-booked';
      case 'DEPARTED': return 'badge-departed';
      case 'IN_TRANSIT': return 'badge-intransit';
      case 'ARRIVED': return 'badge-arrived';
      case 'DELIVERED': return 'badge-delivered';
      case 'EXCEPTION': return 'badge-exception';
      default: return 'badge-default';
    }
  };

  const formatPort = (portCode) => {
    if (!portCode) return '—';
    return portCode.toUpperCase();
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '—';
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch {
      return dateStr;
    }
  };

  const filteredShipments = shipments.filter(s => {
    if (isBookingsMode) {
      return s.status === 'BOOKED' || s.status === 'BOOKING_PENDING';
    }
    if (filterStatus === 'ALL') return true;
    if (filterStatus === 'EXCEPTION') return s.status === 'EXCEPTION';
    return s.status === filterStatus;
  });

  return (
    <div className="shipments-container radial-glow-top">
      {/* Header */}
      <div className="shipments-header">
        <div>
          <span className="section-label section-label-blue">Operations Control</span>
          <h1 className="shipments-title">{pageTitle}</h1>
          <p className="shipments-subtitle">{pageSubtitle}</p>
        </div>
        <button className="btn-primary" onClick={fetchShipments}>
          🔄 Refresh
        </button>
      </div>

      {/* Error state banner with retry */}
      {error && (
        <div className="error-banner" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            ⚠️ <strong>Notice:</strong> {error}
          </div>
          <button
            onClick={fetchShipments}
            style={{
              background: '#EF4444',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              padding: '4px 12px',
              cursor: 'pointer',
              fontWeight: 600,
            }}
          >
            Try Again
          </button>
        </div>
      )}

      {/* Filters (only for shipments view) */}
      {!isBookingsMode && (
        <div className="filter-bar">
          {['ALL', 'BOOKING_PENDING', 'BOOKED', 'DEPARTED', 'IN_TRANSIT', 'ARRIVED', 'DELIVERED', 'EXCEPTION'].map(status => (
            <button
              key={status}
              className={`filter-tab ${filterStatus === status ? 'active' : ''}`}
              onClick={() => setFilterStatus(status)}
            >
              {status.replace('_', ' ')}
            </button>
          ))}
        </div>
      )}

      {/* Loading Skeleton */}
      {loading ? (
        <div className="loading-container">
          <div className="auth-spinner" />
          <p>Retrieving {isBookingsMode ? 'carrier bookings' : 'fleet shipments'}...</p>
        </div>
      ) : filteredShipments.length === 0 ? (
        /* Enterprise Empty State */
        <div className="empty-state" style={{
          background: '#FFFFFF',
          border: '1px solid #E2E8F0',
          borderRadius: '16px',
          padding: '48px 32px',
          textAlign: 'center',
          boxShadow: '0 1px 3px rgba(15, 23, 42, 0.03)',
        }}>
          <div style={{
            width: '56px',
            height: '56px',
            borderRadius: '14px',
            background: '#EFF6FF',
            border: '1px solid #DBEAFE',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '1.6rem',
            margin: '0 auto 16px auto',
          }}>
            {isBookingsMode ? '📦' : '🚢'}
          </div>

          <h3 style={{ fontSize: '1.2rem', fontWeight: 800, color: '#0F172A', marginBottom: '8px' }}>
            {isBookingsMode ? 'No bookings yet' : 'No shipments yet'}
          </h3>

          <p style={{ fontSize: '0.85rem', color: '#64748B', maxWidth: '440px', margin: '0 auto 24px auto', lineHeight: 1.5 }}>
            {isBookingsMode
              ? 'Confirmed carrier bookings will appear here once an RFQ quote is won and booked.'
              : 'Once you win an RFQ and confirm a carrier booking, your live shipment will appear here.'}
          </p>

          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', marginBottom: '32px' }}>
            <button
              className="btn-primary"
              onClick={() => navigate('/dashboard/rfqs')}
              style={{
                background: 'linear-gradient(135deg, #2563EB 0%, #4F46E5 100%)',
                color: '#FFFFFF',
                border: 'none',
                borderRadius: '9px',
                padding: '10px 22px',
                fontSize: '0.82rem',
                fontWeight: 700,
                cursor: 'pointer',
              }}
            >
              {isBookingsMode ? 'View Won RFQs' : 'View RFQs'}
            </button>
          </div>

          {/* Shipment Lifecycle Diagram */}
          <div style={{
            borderTop: '1px solid #F1F5F9',
            paddingTop: '24px',
            maxWidth: '620px',
            margin: '0 auto',
          }}>
            <div style={{ fontSize: '0.72rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: '#94A3B8', marginBottom: '14px' }}>
              Freight Forwarding Lifecycle
            </div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '8px' }}>
              {[
                { step: '1', name: 'RFQ' },
                { step: '2', name: 'Quote' },
                { step: '3', name: 'Booking' },
                { step: '4', name: 'Shipment' },
                { step: '5', name: 'Delivered' },
              ].map((item, idx, arr) => (
                <React.Fragment key={item.step}>
                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                    <div style={{
                      width: '24px',
                      height: '24px',
                      borderRadius: '50%',
                      background: '#F1F5F9',
                      color: '#2563EB',
                      fontSize: '0.72rem',
                      fontWeight: 800,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      marginBottom: '4px',
                    }}>
                      {item.step}
                    </div>
                    <span style={{ fontSize: '0.75rem', fontWeight: 700, color: '#334155' }}>{item.name}</span>
                  </div>
                  {idx < arr.length - 1 && (
                    <span style={{ color: '#CBD5E1', fontSize: '0.85rem', fontWeight: 700 }}>→</span>
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>
        </div>
      ) : (
        /* Data Table */
        <div className="table-responsive glass-card">
          <table className="shipments-table">
            <thead>
              <tr>
                <th>Booking Ref</th>
                <th>Route</th>
                <th>Carrier SCAC</th>
                <th>Vessel / Voyage</th>
                <th>ETD</th>
                <th>ETA</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredShipments.map(s => (
                <tr
                  key={s.id}
                  className="shipment-row"
                  onClick={() => navigate(`/dashboard/shipments/${s.id}`)}
                >
                  <td className="booking-cell">
                    <strong>{s.booking_number || `SH-${s.id}`}</strong>
                    {s.container_numbers && s.container_numbers.length > 0 && (
                      <span className="container-count-badge">
                        📦 {s.container_numbers.length} Container(s)
                      </span>
                    )}
                  </td>
                  <td className="route-cell">
                    <span className="port-badge">{formatPort(s.origin_port)}</span>
                    <span className="route-arrow">➔</span>
                    <span className="port-badge">{formatPort(s.destination_port)}</span>
                  </td>
                  <td className="carrier-cell">
                    <strong>{s.carrier_scac}</strong>
                  </td>
                  <td>
                    {s.vessel_name ? (
                      <div className="vessel-info">
                        <span>🚢 {s.vessel_name}</span>
                        {s.voyage_number && <span className="voyage-num">Voy: {s.voyage_number}</span>}
                      </div>
                    ) : (
                      <span className="unassigned-text">Unassigned</span>
                    )}
                  </td>
                  <td>{formatDate(s.etd)}</td>
                  <td>{formatDate(s.eta)}</td>
                  <td>
                    <span className={`status-badge ${getStatusBadgeClass(s.status)}`}>
                      {s.status.replace('_', ' ')}
                    </span>
                  </td>
                  <td>
                    <button className="btn-action">
                      View Control →
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
