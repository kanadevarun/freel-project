import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { rfqService } from '../../../services/rfqService';
import UniversalTimeline from '../../../components/dashboard/UniversalTimeline';
import PricingWorkspace from './PricingWorkspace';
import StatusBadge from '../../../components/dashboard/StatusBadge';
import { STAGE_CONFIG } from './constants';
import './RFQPage.css'; // Shared styles

/**
 * RFQDetail is a slide-out drawer or full-page view for a single RFQ.
 * Features a HubSpot-style layout: Details/Comms on Left, Timeline in Center, AI/Pricing on Right.
 */
export default function RFQDetail({ rfqId, onClose }) {
  const navigate = useNavigate();
  const [rfq, setRfq] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('NOTES'); // NOTES, EMAILS, FILES
  const [isRequestExpanded, setIsRequestExpanded] = useState(true);

  const [timelineEvents, setTimelineEvents] = useState([]);

  const fetchDetail = async () => {
    setIsLoading(true);
    try {
      const [rfqRes, timelineRes] = await Promise.all([
        rfqService.getRFQ(rfqId),
        rfqService.getTimeline(rfqId)
      ]);
      const fetchedRfq = rfqRes.data || rfqRes;
      setRfq(fetchedRfq);
      
      let events = timelineRes.data || timelineRes || [];
      // Ensure origin creation timeline marker is included (Part 6)
      if (fetchedRfq && fetchedRfq.lead_id) {
        const hasOriginEvent = events.some(e => e.action === 'RFQ_CREATED' || (e.description && e.description.includes('Created from customer email')));
        if (!hasOriginEvent) {
          events = [
            {
              id: 'origin-created-event',
              action: 'RFQ_CREATED',
              description: '✓ RFQ Created — Created from customer email conversation.',
              timestamp: fetchedRfq.created_at || new Date().toISOString(),
              actor: 'Email Parser AI'
            },
            ...events
          ];
        }
      }
      setTimelineEvents(events);
    } catch (error) {
      console.error('Failed to load RFQ:', error);
      toast.error('Failed to load details');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (rfqId) {
      fetchDetail();
    }
  }, [rfqId]);

  if (isLoading) {
    return (
      <div className="rfq-detail-drawer loading">
        <div className="spinner-large"></div>
      </div>
    );
  }

  if (!rfq) return null;

  const stageConfig = STAGE_CONFIG[rfq.stage] || { label: rfq.stage, color: 'gray' };
  
  const getHealthColor = (score) => {
    if (score >= 90) return '#10b981';
    if (score >= 60) return '#f59e0b';
    return '#ef4444';
  };

  return (
    <div className="rfq-detail-drawer open">
      <div className="drawer-overlay" onClick={onClose}></div>
      
      <div className="drawer-content-large">
        <div className="drawer-header">
          <div className="drawer-title-group">
            <h2>{rfq.rfq_number}</h2>
            <StatusBadge status={stageConfig.label} color={stageConfig.color} />
            <div className="health-score-badge" style={{ borderColor: getHealthColor(rfq.health_score), color: getHealthColor(rfq.health_score) }}>
              Health: {rfq.health_score}%
            </div>
          </div>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>

        <div className="drawer-body hubspot-layout">
          
          {/* Left Panel: Details & Comms */}
          <div className="layout-left">
            {/* Part 1 — RFQ Origin Information */}
            {rfq.lead_id && (
              <div className="info-card origin-info-card" style={{
                backgroundColor: '#f8fafc',
                border: '1px solid #cbd5e1',
                borderRadius: '10px',
                padding: '12px 14px',
                marginBottom: '16px'
              }}>
                <div style={{ fontSize: '11px', fontWeight: '700', color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '8px' }}>
                  Created from customer inquiry
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px 12px', fontSize: '12.5px', marginBottom: '10px' }}>
                  <div><span style={{ color: '#64748b' }}>Customer:</span> <strong style={{ color: '#0f172a' }}>{rfq.customer_name || rfq.customer_id || 'Inquiry Contact'}</strong></div>
                  <div><span style={{ color: '#64748b' }}>Source:</span> <strong style={{ color: '#0f172a' }}>Email conversation</strong></div>
                  <div><span style={{ color: '#64748b' }}>Lead:</span> <strong style={{ color: '#0f172a' }}>Lead #{rfq.lead_id}</strong></div>
                </div>
                <button
                  className="btn btn-secondary btn-sm"
                  onClick={() => {
                    if (onClose) onClose();
                    navigate(`/dashboard/leads?leadId=${rfq.lead_id}&tab=email`);
                  }}
                  style={{
                    fontSize: '12px',
                    fontWeight: '600',
                    padding: '6px 12px',
                    backgroundColor: '#eff6ff',
                    color: '#2563eb',
                    border: '1px solid #bfdbfe',
                    borderRadius: '6px',
                    cursor: 'pointer',
                    width: '100%'
                  }}
                >
                  [ View Lead Conversation → ]
                </button>
              </div>
            )}

            {/* Part 2 — Preserve Conversation Context in RFQ (Collapsible Customer Request) */}
            <div className="info-card customer-request-card" style={{
              backgroundColor: '#ffffff',
              border: '1px solid #cbd5e1',
              borderRadius: '10px',
              padding: '12px 14px',
              marginBottom: '16px'
            }}>
              <div 
                onClick={() => setIsRequestExpanded(!isRequestExpanded)}
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer' }}
              >
                <h3 style={{ margin: 0, fontSize: '13px', fontWeight: '700', color: '#0f172a' }}>
                  Customer Request ({rfq.cargo_description || rfq.rfq_number})
                </h3>
                <span style={{ fontSize: '11px', color: '#64748b', fontWeight: '600' }}>
                  {isRequestExpanded ? '▲ Hide' : '▼ Expand'}
                </span>
              </div>

              {isRequestExpanded && (
                <div style={{ marginTop: '10px', paddingTop: '10px', borderTop: '1px solid #f1f5f9' }}>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px 12px', fontSize: '12px' }}>
                    <div><span style={{ color: '#64748b' }}>Origin:</span> <strong>{rfq.origin || 'N/A'}</strong></div>
                    <div><span style={{ color: '#64748b' }}>Destination:</span> <strong>{rfq.destination || 'N/A'}</strong></div>
                    <div><span style={{ color: '#64748b' }}>Cargo:</span> <strong>{rfq.cargo_description || 'General Cargo'}</strong></div>
                    <div><span style={{ color: '#64748b' }}>Ready Date:</span> <strong>{rfq.target_date || rfq.ready_date || 'N/A'}</strong></div>
                    <div><span style={{ color: '#64748b' }}>Weight:</span> <strong>{rfq.cargo_weight ? `${rfq.cargo_weight} kg` : 'N/A'}</strong></div>
                    <div><span style={{ color: '#64748b' }}>Volume:</span> <strong>{rfq.cargo_volume ? `${rfq.cargo_volume} CBM` : 'N/A'}</strong></div>
                    <div><span style={{ color: '#64748b' }}>Incoterms:</span> <strong>{rfq.incoterms || 'FOB'}</strong></div>
                  </div>
                  {rfq.lead_id && (
                    <div style={{ marginTop: '10px' }}>
                      <button
                        onClick={() => {
                          if (onClose) onClose();
                          navigate(`/dashboard/leads?leadId=${rfq.lead_id}&tab=email`);
                        }}
                        style={{
                          fontSize: '11.5px',
                          color: '#2563eb',
                          background: 'none',
                          border: 'none',
                          padding: 0,
                          cursor: 'pointer',
                          fontWeight: '600',
                          textDecoration: 'underline'
                        }}
                      >
                        [ View email conversation ]
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>

            <div className="info-card">
              <h3>Shipment Details</h3>
              <div className="info-grid">
                <div className="info-item">
                  <label>Customer ID</label>
                  <div>{rfq.customer_id}</div>
                </div>
                <div className="info-item">
                  <label>Route</label>
                  <div>{rfq.origin || '-'} → {rfq.destination || '-'}</div>
                </div>
                <div className="info-item">
                  <label>Incoterms</label>
                  <div>{rfq.incoterms || '-'}</div>
                </div>
              </div>
            </div>

            <div className="comms-panel">
              <div className="comms-tabs">
                <button className={activeTab === 'NOTES' ? 'active' : ''} onClick={() => setActiveTab('NOTES')}>Notes</button>
                <button className={activeTab === 'EMAILS' ? 'active' : ''} onClick={() => setActiveTab('EMAILS')}>Emails</button>
                <button className={activeTab === 'FILES' ? 'active' : ''} onClick={() => setActiveTab('FILES')}>Files</button>
              </div>
              <div className="comms-content">
                {activeTab === 'NOTES' && (
                  <div className="note-composer">
                    <textarea placeholder="Leave a note about this RFQ..." />
                    <button className="btn-secondary btn-sm">Save Note</button>
                  </div>
                )}
                {activeTab === 'EMAILS' && (
                  <div className="empty-state-sm">No emails synced yet.</div>
                )}
                {activeTab === 'FILES' && (
                  <div className="file-upload-zone">
                    <p>Drag & drop BL, MSDS, or Packing Lists here</p>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Center Panel: Timeline */}
          <div className="layout-center">
            <h3>Activity Timeline</h3>
            <UniversalTimeline events={timelineEvents} />
          </div>

          {/* Right Panel: AI & Pricing */}
          <div className="layout-right">
            
            <div className="ai-suggestions-panel">
              <div className="panel-header">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                </svg>
                AI Insights
              </div>
              <ul className="insights-list">
                <li className="warning">⚠ Missing HS Code</li>
                <li className="info">ℹ️ Customer usually accepts within 2 hrs</li>
                <li className="success">⭐ Recommended Margin: 18%</li>
              </ul>
            </div>

            <PricingWorkspace 
              rfq={rfq} 
              quotes={rfq.quotes} 
              onQuoteSubmitted={fetchDetail} 
            />

          </div>

        </div>
      </div>
    </div>
  );
}
