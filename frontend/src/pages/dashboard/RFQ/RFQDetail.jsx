import React, { useState, useEffect } from 'react';
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
  const [rfq, setRfq] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('NOTES'); // NOTES, EMAILS, FILES

  const [timelineEvents, setTimelineEvents] = useState([]);

  const fetchDetail = async () => {
    setIsLoading(true);
    try {
      const [rfqRes, timelineRes] = await Promise.all([
        rfqService.getRFQ(rfqId),
        rfqService.getTimeline(rfqId)
      ]);
      setRfq(rfqRes.data);
      setTimelineEvents(timelineRes.data || []);
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
