import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import ManualUpdatePanel from './ManualUpdatePanel';
import DocumentWorkspace from './DocumentWorkspace';
import FinanceWorkspace from './FinanceWorkspace';
import BillingWorkspace from './BillingWorkspace';
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

  const token = localStorage.getItem('token') || 'test-token';
  const isDev = import.meta.env.DEV || window.location.hostname === 'localhost';

  useEffect(() => {
    fetchShipmentDetails();
  }, [id]);

  const fetchShipmentDetails = async () => {
    try {
      const response = await fetch(`/api/v1/shipments/${id}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (!response.ok) {
        throw new Error(`Error: ${response.status} ${response.statusText}`);
      }
      const resData = await response.json();
      setShipment(resData.data.shipment);
      setMilestones(resData.data.milestones || []);
      setExceptions(resData.data.exceptions || []);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch shipment details:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleResolveException = async (excId) => {
    try {
      const response = await fetch(`/api/v1/shipments/exceptions/${excId}/resolve`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (response.ok) {
        fetchShipmentDetails(); // Refresh details after resolution
      } else {
        alert('Failed to resolve exception.');
      }
    } catch (err) {
      console.error('Error resolving exception:', err);
    }
  };

  const handleSubmitCarrierUpdate = async (e) => {
    e.preventDefault();
    if (!rawUpdate.trim()) return;

    try {
      setSubmittingUpdate(true);
      setUpdateMsg('');
      const response = await fetch(`/api/v1/shipments/${id}/carrier-update`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          event_id: `WEB-MANUAL-${Date.now()}`,
          description: rawUpdate
        })
      });

      if (response.ok) {
        setUpdateMsg('✅ Update received. Processing task enqueued...');
        setRawUpdate('');
        
        // Start polling interval - poll 3 times, every 3 seconds (Group 6 fix)
        let attempts = 0;
        const intervalId = setInterval(async () => {
          attempts += 1;
          await fetchShipmentDetails();
          if (attempts >= 3) {
            clearInterval(intervalId);
          }
        }, 3000);

      } else {
        const errData = await response.json();
        setUpdateMsg(`❌ Failed to process: ${errData.message || 'Unknown error'}`);
      }
    } catch (err) {
      console.error('Error posting carrier update:', err);
      setUpdateMsg(`❌ Connection failure: ${err.message}`);
    } finally {
      setSubmittingUpdate(false);
    }
  };

  const getSeverityBadgeClass = (severity) => {
    switch (severity) {
      case 'CRITICAL': return 'exc-critical';
      case 'WARNING': return 'exc-warning';
      case 'INFO': return 'exc-info';
      default: return 'exc-info';
    }
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

  if (loading) {
    return (
      <div className="loading-container">
        <div className="auth-spinner" />
        <p>Loading operational timeline control...</p>
      </div>
    );
  }

  if (error || !shipment) {
    return (
      <div className="shipments-container">
        <div className="error-banner">
          ⚠️ <strong>Failed to retrieve shipment:</strong> {error || 'Record not found.'}
        </div>
        <button className="btn-secondary" onClick={() => navigate('/dashboard/shipments')}>
          ← Back to fleet list
        </button>
      </div>
    );
  }

  const activeExceptionCount = exceptions.filter(e => !e.resolved).length;

  return (
    <div className="shipment-detail-container radial-glow-top">
      <div className="detail-header">
        <div>
          <button className="btn-back" onClick={() => navigate('/dashboard/shipments')}>
            ← Back to Fleet
          </button>
          <div className="detail-title-row">
            <h1 className="detail-title">Shipment: {shipment.booking_number || `SH-${shipment.id}`}</h1>
            <span className={`status-badge detail-status ${shipment.status === 'EXCEPTION' ? 'badge-exception' : 'badge-booked'}`}>
              {shipment.status.replace('_', ' ')}
            </span>
          </div>
          <p className="detail-subtitle">Route: {shipment.origin_port} ➔ {shipment.destination_port} | Carrier: {shipment.carrier_scac}</p>
        </div>
      </div>

      {activeExceptionCount > 0 && (
        <div className="critical-exception-alert-banner">
          <div className="alert-emoji">⚠️</div>
          <div>
            <h4>{activeExceptionCount} Exception(s) Require Actions</h4>
            <p>Shipment schedule or compliance hold detected. Resolve anomalies to restore green route status.</p>
          </div>
        </div>
      )}

      <div className="detail-grid">
        {/* Left Panel: Milestone Gantt & Control */}
        <div className="left-panel">
          <div className="glass-card panel-card">
            <h3>Milestone Gantt Timeline</h3>
            <p className="panel-desc">Planned vs. Actual milestones evaluated by the Operations Agent.</p>

            <div className="milestones-timeline">
              {milestones.map((m, index) => {
                const isCompleted = m.status === 'COMPLETED';
                const isPending = m.status === 'PLANNED';
                return (
                  <div key={m.id} className={`timeline-node ${isCompleted ? 'completed' : 'pending'}`}>
                    <div className="timeline-marker">
                      <div className="marker-dot">
                        {isCompleted ? '✓' : '•'}
                      </div>
                      {index < milestones.length - 1 && <div className="marker-line" />}
                    </div>
                    <div className="timeline-content">
                      <div className="timeline-header">
                        <h4>{m.milestone_code.replace('_', ' ')}</h4>
                        <span className={`node-badge ${isCompleted ? 'badge-completed' : 'badge-pending'}`}>
                          {m.status}
                        </span>
                      </div>
                      <p className="node-desc">{m.description || 'No description available.'}</p>
                      <div className="timeline-dates">
                        <div>
                          <small>Planned Date:</small>
                          <span>{formatDate(m.planned_date)}</span>
                        </div>
                        {isCompleted && (
                          <div>
                            <small>Actual Date:</small>
                            <span className="actual-date-val">{formatDate(m.actual_date)}</span>
                          </div>
                        )}
                      </div>
                      {m.location && (
                        <div className="node-location">
                          📍 <span>{m.location}</span>
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
          
          {/* Document Compliance Panel */}
          <DocumentWorkspace 
            shipmentId={id} 
            token={token} 
            onRefreshShipment={fetchShipmentDetails} 
          />

          {/* Phase 5: Finance & Reconciliation Panel */}
          <FinanceWorkspace 
            shipmentId={id} 
            token={token} 
            onRefreshShipment={fetchShipmentDetails} 
          />

          {/* Gap-Closure: Customer Invoicing & Profitability Panel */}
          <BillingWorkspace 
            shipmentId={id} 
            token={token} 
            onRefreshShipment={fetchShipmentDetails} 
          />
        </div>

        {/* Right Panel: Metadata & Exceptions Control */}
        <div className="right-panel">
          {/* Metadata Card */}
          <div className="glass-card panel-card metadata-card">
            <h3>Shipment Metadata</h3>
            <div className="metadata-grid">
              <div className="meta-item">
                <label>Carrier SCAC</label>
                <span>{shipment.carrier_scac}</span>
              </div>
              <div className="meta-item">
                <label>Vessel Name</label>
                <span>🚢 {shipment.vessel_name || 'Unassigned'}</span>
              </div>
              <div className="meta-item">
                <label>Voyage Number</label>
                <span>{shipment.voyage_number || 'Unassigned'}</span>
              </div>
              <div className="meta-item">
                <label>Containers</label>
                <span className="container-numbers-list">
                  {shipment.container_numbers && shipment.container_numbers.length > 0
                    ? shipment.container_numbers.join(', ')
                    : 'MSKU1234567 (Assigned)'}
                </span>
              </div>
            </div>
          </div>

          {/* Exceptions Panel */}
          <div className="glass-card panel-card">
            <h3>Anomalies & Exceptions</h3>
            <p className="panel-desc">Exceptions flagged automatically by the Operations Agent.</p>

            {exceptions.length === 0 ? (
              <div className="no-exceptions">
                <span className="success-emoji">💚</span>
                <p>No exceptions logged. Shipment route is operating normally.</p>
              </div>
            ) : (
              <div className="exceptions-list">
                {exceptions.map(e => (
                  <div key={e.id} className={`exception-card ${e.resolved ? 'resolved' : 'unresolved'}`}>
                    <div className="exception-header">
                      <span className={`severity-badge ${getSeverityBadgeClass(e.severity)}`}>
                        {e.severity}
                      </span>
                      {e.resolved ? (
                        <span className="resolved-status">✓ Resolved</span>
                      ) : (
                        <button
                          className="btn-resolve"
                          onClick={() => handleResolveException(e.id)}
                        >
                          Resolve Exception
                        </button>
                      )}
                    </div>
                    <h4>{e.title}</h4>
                    <p className="exc-desc">{e.description}</p>
                    {e.ai_summary && (
                      <div className="exc-ai-summary">
                        <strong>AI Analysis:</strong> {e.ai_summary}
                      </div>
                    )}
                    <div className="exc-date">
                      <small>Logged at: {formatDate(e.created_at)}</small>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Simulate Webhook / Update Panel - development only (Group 6 fix) */}
          {isDev && (
            <ManualUpdatePanel
              onSubmit={handleSubmitCarrierUpdate}
              submittingUpdate={submittingUpdate}
              rawUpdate={rawUpdate}
              setRawUpdate={setRawUpdate}
              updateMsg={updateMsg}
            />
          )}
        </div>
      </div>
    </div>
  );
}
