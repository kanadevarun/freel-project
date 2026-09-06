import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { rfqService } from '../../../services/rfqService';
import { calculateRFQCompleteness } from './utils/completeness';
import RFQHeader from './components/RFQHeader';
import RFQOverview from './components/RFQOverview';
import RFQCargoItems from './components/RFQCargoItems';
import RFQRequirements from './components/RFQRequirements';
import RFQActivityTimeline from './components/RFQActivityTimeline';
import RFQDocuments from './components/RFQDocuments';
import RFQQuotes from './components/RFQQuotes';
import RFQBookingHandoff from './components/RFQBookingHandoff';
import RFQShipmentHandoff from './components/RFQShipmentHandoff';
import './RFQDetailPage.css';

export default function RFQDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const [rfq, setRfq] = useState(null);
  const [timelineEvents, setTimelineEvents] = useState([]);
  const [requirements, setRequirements] = useState(null);
  const [documentsData, setDocumentsData] = useState(null);
  const [quotesData, setQuotesData] = useState(null);
  const [bookingHandoffData, setBookingHandoffData] = useState(null);
  const [shipmentHandoffData, setShipmentHandoffData] = useState(null);
  const [activityData, setActivityData] = useState(null);
  const [isActivityLoading, setIsActivityLoading] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Active Tab state
  const initialTab = searchParams.get('tab') || 'overview';
  const [activeTab, setActiveTab] = useState(initialTab);

  const handleTabChange = (tabKey) => {
    setActiveTab(tabKey);
    setSearchParams({ tab: tabKey });
  };

  const fetchRFQData = useCallback(async () => {
    if (!id) return;
    setIsLoading(true);
    setError(null);

    try {
      const [rfqRes, timelineRes, requirementsRes, docsRes, quotesRes, bookingsRes, shipmentsRes] = await Promise.all([
        rfqService.getRFQ(id),
        rfqService.getTimeline(id).catch(err => {
          console.warn('Timeline retrieval failed or empty:', err);
          return { data: [] };
        }),
        rfqService.getRequirements(id).catch(err => {
          console.warn('Requirements evaluation failed:', err);
          return null;
        }),
        rfqService.getDocuments(id).catch(err => {
          console.warn('Documents retrieval failed:', err);
          return null;
        }),
        rfqService.getQuotes(id).catch(err => {
          console.warn('Quotes retrieval failed:', err);
          return null;
        }),
        rfqService.getRFQBookings(id).catch(err => {
          console.warn('Bookings handoff retrieval failed:', err);
          return null;
        }),
        rfqService.getRFQShipments(id).catch(err => {
          console.warn('Shipments handoff retrieval failed:', err);
          return null;
        })
      ]);

      const fetchedRfq = rfqRes?.data || rfqRes;
      if (!fetchedRfq || !fetchedRfq.id) {
        throw new Error('RFQ not found');
      }

      setRfq(fetchedRfq);
      const events = timelineRes?.data || timelineRes || [];
      setTimelineEvents(Array.isArray(events) ? events : []);

      const reqData = requirementsRes?.data || requirementsRes;
      setRequirements(reqData || null);

      const docs = docsRes?.data || docsRes;
      setDocumentsData(docs || null);

      const quotes = quotesRes?.data || quotesRes;
      setQuotesData(quotes || null);

      const bookings = bookingsRes?.data || bookingsRes;
      setBookingHandoffData(bookings || null);

      const shipments = shipmentsRes?.data || shipmentsRes;
      setShipmentHandoffData(shipments || null);
    } catch (err) {
      console.error('Failed to load RFQ workspace:', err);
      const status = err?.response?.status;
      if (status === 404 || status === 401 || status === 403) {
        setError('RFQ not found or you do not have permission to view it.');
      } else {
        setError(err.message || 'Failed to load RFQ details');
      }
      toast.error('Could not load RFQ workspace');
    } finally {
      setIsLoading(false);
    }
  }, [id]);

  const refreshAllData = useCallback(async () => {
    if (!id) return;
    try {
      const [rfqRes, reqRes, docsRes, quotesRes, bookingsRes, shipmentsRes, actRes] = await Promise.all([
        rfqService.getRFQ(id),
        rfqService.getRequirements(id),
        rfqService.getDocuments(id),
        rfqService.getQuotes(id).catch(() => null),
        rfqService.getRFQBookings(id).catch(() => null),
        rfqService.getRFQShipments(id).catch(() => null),
        rfqService.getActivity(id).catch(() => null)
      ]);
      if (rfqRes?.data) setRfq(rfqRes.data);
      if (reqRes?.data) setRequirements(reqRes.data);
      if (docsRes?.data) setDocumentsData(docsRes.data);
      if (quotesRes?.data || quotesRes) setQuotesData(quotesRes?.data || quotesRes);
      if (bookingsRes?.data || bookingsRes) setBookingHandoffData(bookingsRes?.data || bookingsRes);
      if (shipmentsRes?.data || shipmentsRes) setShipmentHandoffData(shipmentsRes?.data || shipmentsRes);
      if (actRes?.data) setActivityData(actRes.data);
    } catch (err) {
      console.error('Failed to refresh RFQ workspace data:', err);
    }
  }, [id]);

  useEffect(() => {
    fetchRFQData();
  }, [fetchRFQData]);

  // Lazy load Activity data when Activity tab is activated
  const fetchActivityData = useCallback(async (force = false) => {
    if (!id || (!force && activityData)) return;
    setIsActivityLoading(true);
    try {
      const actRes = await rfqService.getActivity(id);
      const data = actRes?.data || actRes;
      setActivityData(data);
    } catch (err) {
      console.warn('Activity retrieval failed:', err);
    } finally {
      setIsActivityLoading(false);
    }
  }, [id, activityData]);

  useEffect(() => {
    if (activeTab === 'activity') {
      fetchActivityData();
    }
  }, [activeTab, fetchActivityData]);

  // Loading skeleton state
  if (isLoading) {
    return (
      <div className="rfq-workspace-container animate-in fade-in duration-200">
        <div style={{
          background: '#FFFFFF',
          borderBottom: '1px solid #E2E8F0',
          padding: '24px 32px'
        }}>
          <div style={{ maxWidth: '1360px', margin: '0 auto' }}>
            <div style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '8px',
              padding: '6px 14px',
              background: 'linear-gradient(90deg, #EEF2FF 0%, #E0E7FF 100%)',
              border: '1px solid #C7D2FE',
              borderRadius: '20px',
              fontSize: '12px',
              color: '#3730A3',
              fontWeight: 700,
              marginBottom: '16px'
            }}>
              <span style={{ animation: 'spin 1.5s linear infinite', display: 'inline-block' }}>⚡</span>
              <span>Loading RFQ Workspace & syncing operational records...</span>
            </div>

            <div style={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
              <div className="rfq-skeleton-box" style={{ width: '280px', height: '32px', borderRadius: '8px' }} />
              <div className="rfq-skeleton-box" style={{ width: '120px', height: '24px', borderRadius: '12px' }} />
              <div className="rfq-skeleton-box" style={{ width: '100px', height: '24px', borderRadius: '12px' }} />
            </div>

            <div style={{ display: 'flex', gap: '24px', marginTop: '16px' }}>
              <div className="rfq-skeleton-box" style={{ width: '200px', height: '14px', borderRadius: '4px' }} />
              <div className="rfq-skeleton-box" style={{ width: '160px', height: '14px', borderRadius: '4px' }} />
              <div className="rfq-skeleton-box" style={{ width: '180px', height: '14px', borderRadius: '4px' }} />
            </div>
          </div>
        </div>
        <div className="rfq-workspace-content" style={{ maxWidth: '1360px', margin: '24px auto', padding: '0 32px' }}>
          <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '24px' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
              <div className="rfq-skeleton-box" style={{ height: '200px', borderRadius: '14px' }} />
              <div className="rfq-skeleton-box" style={{ height: '280px', borderRadius: '14px' }} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
              <div className="rfq-skeleton-box" style={{ height: '240px', borderRadius: '14px' }} />
              <div className="rfq-skeleton-box" style={{ height: '240px', borderRadius: '14px' }} />
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Error / Not Found state
  if (error || !rfq) {
    return (
      <div className="rfq-workspace-container">
        <div className="rfq-workspace-content" style={{ textAlign: 'center', paddingTop: '80px' }}>
          <div style={{ fontSize: '48px', marginBottom: '16px' }}>🔍</div>
          <h2 style={{ fontSize: '20px', fontWeight: 800, color: '#0F172A', marginBottom: '8px' }}>
            RFQ Not Found
          </h2>
          <p style={{ fontSize: '13.5px', color: '#64748B', maxWidth: '440px', margin: '0 auto 24px auto', lineHeight: 1.5 }}>
            {error || 'The requested shipment request could not be located in this organization.'}
          </p>
          <button
            onClick={() => navigate('/dashboard/rfqs')}
            style={{
              background: '#4F46E5',
              color: '#FFFFFF',
              border: 'none',
              borderRadius: '8px',
              padding: '10px 20px',
              fontSize: '13px',
              fontWeight: 700,
              cursor: 'pointer',
            }}
          >
            ← Back to All RFQs
          </button>
        </div>
      </div>
    );
  }

  const completeness = calculateRFQCompleteness(rfq);
  const quotesCount = quotesData?.quotes?.length ?? quotesData?.summary?.total_quotes ?? rfq?.quotes?.length ?? 0;
  const docAttentionCount = documentsData?.summary?.missing_documents ?? (requirements?.document_requirements?.filter(d => d.is_required && d.status !== 'SATISFIED').length ?? 0);

  // Requirements badge: show blocking count from backend (source of truth)
  const blockingCount = requirements?.operational_readiness?.blocking_count ?? 0;
  const missingRequired = requirements?.operational_readiness?.missing_required_count ?? 0;
  const reqBadgeCount = blockingCount + missingRequired;

  const bookingCount = bookingHandoffData?.summary?.total_bookings ?? bookingHandoffData?.bookings?.length ?? 0;
  const shipmentCount = shipmentHandoffData?.summary?.total_shipments ?? shipmentHandoffData?.shipments?.length ?? 0;

  return (
    <div className="rfq-workspace-container">
      
      {/* 1. RFQ Header */}
      <RFQHeader
        rfq={rfq}
        completeness={completeness}
        requirements={requirements}
        onTabChange={handleTabChange}
        onRefresh={fetchRFQData}
      />

      {/* 2. Workspace Body */}
      <div className="rfq-workspace-content">
        
        {/* Navigation Tabs (Sticky Header) */}
        <div className="rfq-workspace-tabs-container">
          <div className="rfq-workspace-tabs">
            <button
              className={`rfq-workspace-tab-btn ${activeTab === 'overview' ? 'active' : ''}`}
              onClick={() => handleTabChange('overview')}
            >
              <span>Overview</span>
            </button>

            <button
              className={`rfq-workspace-tab-btn ${activeTab === 'cargo' ? 'active' : ''}`}
              onClick={() => handleTabChange('cargo')}
            >
              <span>Cargo & Shipment</span>
            </button>

            <button
              className={`rfq-workspace-tab-btn ${activeTab === 'requirements' ? 'active' : ''}`}
              onClick={() => handleTabChange('requirements')}
            >
              <span>Requirements</span>
              {reqBadgeCount > 0 && (
                <span className="rfq-workspace-tab-count-badge orange">{reqBadgeCount}</span>
              )}
            </button>

            <button
              className={`rfq-workspace-tab-btn ${activeTab === 'documents' ? 'active' : ''}`}
              onClick={() => handleTabChange('documents')}
            >
              <span>Documents</span>
              {docAttentionCount > 0 && (
                <span className="rfq-workspace-tab-count-badge orange">{docAttentionCount}</span>
              )}
            </button>

            <button
              className={`rfq-workspace-tab-btn ${activeTab === 'activity' ? 'active' : ''}`}
              onClick={() => handleTabChange('activity')}
            >
              <span>Activity</span>
              {activityData?.summary?.total_events > 0 && (
                <span className="rfq-workspace-tab-count-badge">{activityData.summary.total_events}</span>
              )}
            </button>

            {/* Quotes Tab */}
            <button
              className={`rfq-workspace-tab-btn ${activeTab === 'quotes' ? 'active' : ''}`}
              onClick={() => handleTabChange('quotes')}
              data-testid="tab-quotes-btn"
            >
              <span>Quotes</span>
              {quotesCount > 0 && (
                <span className="rfq-workspace-tab-count-badge">{quotesCount}</span>
              )}
            </button>

            {/* Booking Tab (Task 14 Handoff) */}
            <button
              className={`rfq-workspace-tab-btn ${activeTab === 'booking' ? 'active' : ''}`}
              onClick={() => handleTabChange('booking')}
              data-testid="tab-booking-btn"
            >
              <span>Booking</span>
              {bookingCount > 0 && (
                <span className="rfq-workspace-tab-count-badge">{bookingCount}</span>
              )}
            </button>

            {/* Shipment Tab (Task 14 Handoff) */}
            <button
              className={`rfq-workspace-tab-btn ${activeTab === 'shipment' ? 'active' : ''}`}
              onClick={() => handleTabChange('shipment')}
              data-testid="tab-shipment-btn"
            >
              <span>Shipment</span>
              {shipmentCount > 0 && (
                <span className="rfq-workspace-tab-count-badge">{shipmentCount}</span>
              )}
            </button>
          </div>
        </div>


        {/* 3. Tab Content Panels */}
        {activeTab === 'overview' && (
          <RFQOverview
            rfq={rfq}
            completeness={completeness}
            onSwitchTab={handleTabChange}
            timelineEvents={timelineEvents}
            requirements={requirements}
            documentsData={documentsData}
            quotesData={quotesData}
            bookingHandoffData={bookingHandoffData}
            shipmentHandoffData={shipmentHandoffData}
          />
        )}

        {activeTab === 'cargo' && (
          <RFQCargoItems
            rfq={rfq}
            completeness={completeness}
            requirements={requirements}
            onSwitchTab={handleTabChange}
          />
        )}


        {activeTab === 'requirements' && (
          <RFQRequirements
            rfq={rfq}
            completeness={completeness}
            requirements={requirements}
            documentsData={documentsData}
            onSwitchTab={handleTabChange}
          />
        )}

        {activeTab === 'documents' && (
          <RFQDocuments
            rfq={rfq}
            documentsData={documentsData}
            onMutationSuccess={refreshAllData}
          />
        )}

        {activeTab === 'activity' && (
          <RFQActivityTimeline
            rfq={rfq}
            activityData={activityData}
            timelineEvents={timelineEvents}
            isLoading={isActivityLoading}
            onRefresh={() => fetchActivityData(true)}
          />
        )}

        {activeTab === 'quotes' && (
          <RFQQuotes
            rfq={rfq}
            quotesData={quotesData}
            requirements={requirements}
            onMutationSuccess={refreshAllData}
            onSwitchTab={handleTabChange}
            onRefresh={refreshAllData}
          />
        )}

        {activeTab === 'booking' && (
          <RFQBookingHandoff
            rfq={rfq}
            bookingHandoffData={bookingHandoffData}
            quotesData={quotesData}
            requirements={requirements}
            documentsData={documentsData}
            shipmentHandoffData={shipmentHandoffData}
            onSwitchTab={handleTabChange}
            onMutationSuccess={refreshAllData}
          />
        )}

        {activeTab === 'shipment' && (
          <RFQShipmentHandoff
            rfq={rfq}
            shipmentHandoffData={shipmentHandoffData}
            bookingHandoffData={bookingHandoffData}
            quotesData={quotesData}
            requirements={requirements}
            documentsData={documentsData}
            onSwitchTab={handleTabChange}
          />
        )}

      </div>

    </div>
  );
}


