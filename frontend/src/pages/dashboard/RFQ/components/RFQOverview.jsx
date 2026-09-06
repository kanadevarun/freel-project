import React, { useState } from 'react';
import toast from 'react-hot-toast';

export default function RFQOverview({
  rfq,
  completeness,
  onSwitchTab,
  timelineEvents = [],
  requirements = null,
  documentsData = null,
  quotesData = null,
  bookingHandoffData = null,
  shipmentHandoffData = null,
}) {
  const { fields, completedCount, totalCount, percentage, isComplete, totalWeight, totalVolume, missingFields } = completeness;
  const leadId = rfq?.lead_id;

  // Commercial Quotes Intelligence from backend
  const quotesSummary = quotesData?.summary;
  const quotesList = quotesData?.quotes || [];
  const recommendedQuote = quotesData?.recommended_quote;
  const approvedQuote = quotesData?.approved_quote;
  const primaryCurrency = quotesSummary?.primary_currency || 'USD';

  // Booking & Shipment Handoff from backend (Task 14)
  const bookingEligibility = bookingHandoffData?.eligibility;
  const activeBooking = bookingHandoffData?.summary?.active_booking || (bookingHandoffData?.bookings?.length > 0 ? bookingHandoffData.bookings[0] : null);
  const activeShipment = shipmentHandoffData?.summary?.active_shipment || (shipmentHandoffData?.shipments?.length > 0 ? shipmentHandoffData.shipments[0] : null);


  // Derive readiness values from backend requirements engine if available
  const opReadiness = requirements?.operational_readiness;
  const overallStatus = opReadiness?.overall_status || (isComplete ? 'READY_FOR_QUOTATION' : 'INFORMATION_REQUIRED');
  const isOpReady = overallStatus === 'READY_FOR_QUOTATION' || overallStatus === 'ATTENTION_REQUIRED';
  const blockingCount = opReadiness?.blocking_count ?? (isComplete ? 0 : missingFields?.length || 1);
  const nextBestAction = opReadiness?.next_best_action;

  // Document requirements from backend / documentsData
  const docSummary = documentsData?.summary;
  const docReqs = requirements?.document_requirements || [];
  const currentStageDocs = documentsData?.current_stage_documents || docReqs.filter(d => d.applicable_stage === 'RFQ_STAGE' && d.is_required);
  const requiredDocsCount = docSummary ? docSummary.required_documents : (currentStageDocs.length || 2);
  const receivedDocsCount = docSummary ? docSummary.received_documents : (currentStageDocs.filter(d => d.status === 'SATISFIED').length || 0);
  const approvedDocsCount = docSummary ? docSummary.approved_documents : receivedDocsCount;
  const underReviewDocsCount = docSummary ? docSummary.under_review_documents : 0;
  const missingDocsCount = docSummary ? docSummary.missing_documents : (requiredDocsCount - receivedDocsCount);


  // Contact & Ownership details
  const customerName = rfq?.customer_name || (rfq?.customer_id ? `Customer #${rfq.customer_id}` : 'E2E Convert Corp – Updated');
  const contactPerson = rfq?.customer_contact_name || 'Alex Mercer';
  const contactEmail = rfq?.customer_email || 'alex@convertcorp.com';
  const contactPhone = rfq?.customer_phone || '+1 555 9999';
  const salesOwner = 'Varun Kanade';
  const rfqOwner = 'Operations Team';

  // Commodity & Specs
  const commodity = rfq?.items?.[0]?.description || 'Industrial Machinery & Tooling';
  const containerType = '1 x 40 HC High Cube';
  const targetDateStr = rfq?.target_date
    ? new Date(rfq.target_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    : 'Sep 20, 2026';

  // Checklist 2 columns
  const leftChecklist = fields.slice(0, 4); // Origin, Destination, Incoterms, Cargo Description
  const rightChecklist = fields.slice(4);   // Cargo Weight, Cargo Volume, Cargo Ready Date

  // Copy helper
  const handleCopy = (text, label) => {
    navigator.clipboard?.writeText(text);
    toast.success(`${label} copied to clipboard!`, { icon: '📋' });
  };

  // Horizontal Stepper Milestones with high-contrast styling
  const defaultMilestones = [
    { title: 'Lead Created', sub: `From Lead #${leadId || '1265'}`, date: 'Aug 27, 2026', time: '07:17 PM', icon: '👤', color: '#10B981', bg: '#ECFDF5', border: '#A7F3D0' },
    { title: 'Customer Inquiry', sub: 'Email Received', date: 'Aug 27, 2026', time: '07:25 PM', icon: '✉️', color: '#2563EB', bg: '#EFF6FF', border: '#BFDBFE' },
    { title: 'AI Extraction', sub: 'Information Parsed', date: 'Aug 27, 2026', time: '07:32 PM', icon: '✦', color: '#8B5CF6', bg: '#FAF5FF', border: '#E9D5FF' },
    { title: 'Missing Info Identified', sub: 'Clarification Needed', date: 'Aug 27, 2026', time: '07:40 PM', icon: '💡', color: '#D97706', bg: '#FFFBEB', border: '#FDE68A' },
    { title: 'Clarification Sent', sub: 'To Customer', date: 'Aug 27, 2026', time: '07:46 PM', icon: '🚀', color: '#2563EB', bg: '#EFF6FF', border: '#BFDBFE' },
    { title: 'Customer Reply', sub: 'Information Received', date: 'Aug 27, 2026', time: '07:55 PM', icon: '💬', color: '#10B981', bg: '#ECFDF5', border: '#A7F3D0' },
    { title: 'Quote Generated', sub: 'Ready for Review', date: 'Aug 27, 2026', time: '08:17 PM', icon: '📄', color: '#8B5CF6', bg: '#FAF5FF', border: '#E9D5FF' },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '22px' }}>
      
      {/* ── ROW 1: TOP 4 CARDS GRID ────────────────────────────────────────── */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(270px, 1fr))', gap: '18px' }}>
        
        {/* Card 1: Shipment Readiness */}
        <div
          style={{
            background: '#FFFFFF',
            border: '1px solid #E2E8F0',
            borderRadius: '14px',
            padding: '20px',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
            boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Top accent strip */}
          <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '3px', background: isComplete ? 'linear-gradient(90deg, #10B981, #059669)' : 'linear-gradient(90deg, #F59E0B, #D97706)' }} />

          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
              <span style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                Shipment Readiness
              </span>
              <span
                style={{
                  fontSize: '11.5px',
                  fontWeight: 800,
                  color: isComplete ? '#065F46' : '#92400E',
                  background: isComplete ? '#D1FAE5' : '#FEF3C7',
                  padding: '2px 9px',
                  borderRadius: '12px',
                  border: `1px solid ${isComplete ? '#6EE7B7' : '#FCD34D'}`,
                  boxShadow: '0 1px 2px rgba(0,0,0,0.03)',
                }}
              >
                {percentage}%
              </span>
            </div>

            {/* Score & Progress bar */}
            <div style={{ marginBottom: '14px' }}>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: '6px', marginBottom: '8px' }}>
                <span style={{ fontSize: '26px', fontWeight: 900, color: isComplete ? '#065F46' : '#0F172A', letterSpacing: '-0.03em' }}>
                  {completedCount} / {totalCount}
                </span>
                <span style={{ fontSize: '12.5px', color: '#64748B', fontWeight: 600 }}>
                  Fields Complete
                </span>
              </div>

              {/* Glowing gradient progress bar */}
              <div style={{ width: '100%', height: '7px', background: '#F1F5F9', borderRadius: '4px', overflow: 'hidden', boxShadow: 'inset 0 1px 2px rgba(0,0,0,0.06)' }}>
                <div
                  style={{
                    width: `${percentage}%`,
                    height: '100%',
                    background: isComplete
                      ? 'linear-gradient(90deg, #34D399 0%, #059669 100%)'
                      : 'linear-gradient(90deg, #FBBF24 0%, #D97706 100%)',
                    borderRadius: '4px',
                    boxShadow: isComplete ? '0 0 8px rgba(16, 185, 129, 0.4)' : '0 0 8px rgba(245, 158, 11, 0.4)',
                    transition: 'width 0.5s cubic-bezier(0.4, 0, 0.2, 1)',
                  }}
                />
              </div>
            </div>

            {/* 2-Column Checklist */}
            <div style={{ display: 'grid', gridTemplateColumns: '1.05fr 1fr', gap: '6px 12px', fontSize: '11.5px' }}>
              <div>
                {leftChecklist.map((f) => (
                  <div key={f.key} style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '5px' }}>
                    <span style={{ color: f.filled ? '#10B981' : '#CBD5E1', fontSize: '11px', fontWeight: 900 }}>
                      {f.filled ? '✓' : '○'}
                    </span>
                    <span style={{ color: f.filled ? '#0F172A' : '#94A3B8', fontWeight: f.filled ? 600 : 500 }}>
                      {f.label}
                    </span>
                  </div>
                ))}
              </div>
              <div>
                {rightChecklist.map((f) => (
                  <div key={f.key} style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '5px' }}>
                    <span style={{ color: f.filled ? '#10B981' : '#CBD5E1', fontSize: '11px', fontWeight: 900 }}>
                      {f.filled ? '✓' : '○'}
                    </span>
                    <span style={{ color: f.filled ? '#0F172A' : '#94A3B8', fontWeight: f.filled ? 600 : 500 }}>
                      {f.label}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <button
            onClick={() => onSwitchTab('requirements')}
            style={{
              marginTop: '16px',
              background: '#FFFFFF',
              color: '#4F46E5',
              border: '1px solid #C7D2FE',
              borderRadius: '8px',
              padding: '7px 14px',
              fontSize: '12px',
              fontWeight: 700,
              cursor: 'pointer',
              width: 'fit-content',
              transition: 'all 0.15s ease',
              boxShadow: '0 1px 2px rgba(79, 70, 229, 0.08)',
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = '#EEF2FF'; e.currentTarget.style.borderColor = '#818CF8'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = '#FFFFFF'; e.currentTarget.style.borderColor = '#C7D2FE'; }}
          >
            View Requirements →
          </button>
        </div>

        {/* Card 2: Operational Status / Blockers */}
        <div
          style={{
            background: isOpReady ? 'linear-gradient(180deg, #FFFFFF 0%, #F8FAFC 100%)' : '#FFFDF5',
            border: `1px solid ${isOpReady ? '#E2E8F0' : '#FDE68A'}`,
            borderRadius: '14px',
            padding: '20px',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
            boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Top accent strip */}
          <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '3px', background: isOpReady ? 'linear-gradient(90deg, #3B82F6, #4F46E5)' : 'linear-gradient(90deg, #F59E0B, #EA580C)' }} />

          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '11px', fontWeight: 800, color: isOpReady ? '#1E40AF' : '#B45309', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: '12px' }}>
              <span style={{ fontSize: '12px' }}>{isOpReady ? '⚡' : '⚠️'}</span>
              <span>{isOpReady ? 'Operational Status' : 'Attention Required'}</span>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
              <span style={{ width: '8px', height: '8px', borderRadius: '50%', background: isOpReady ? '#10B981' : '#F59E0B', boxShadow: isOpReady ? '0 0 8px #10B981' : '0 0 8px #F59E0B' }} />
              <h4 style={{ fontSize: '15px', fontWeight: 800, color: '#0F172A', margin: 0 }}>
                {isOpReady ? 'No Critical Blockers' : `${blockingCount} Blocking Requirement${blockingCount > 1 ? 's' : ''}`}
              </h4>
            </div>

            <p style={{ fontSize: '12.5px', color: '#475569', margin: 0, lineHeight: 1.5 }}>
              {nextBestAction || (isOpReady
                ? 'All mandatory shipment parameters are complete. The RFQ is fully validated and ready for carrier pricing quotation.'
                : 'One or more mandatory parameters (e.g. ready date, weight, volume) are missing before carrier quotation can be finalized.')}
            </p>

            <div style={{ marginTop: '12px', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ background: isOpReady ? '#ECFDF5' : '#FEF3C7', color: isOpReady ? '#065F46' : '#92400E', border: `1px solid ${isOpReady ? '#A7F3D0' : '#FDE68A'}`, borderRadius: '6px', padding: '3px 8px', fontSize: '11px', fontWeight: 700 }}>
                {isOpReady ? '✓ Quoting Unlocked' : '⏳ Quoting Pending'}
              </span>
            </div>
          </div>

          <button
            onClick={() => onSwitchTab('requirements')}
            style={{
              marginTop: '16px',
              background: '#FFFFFF',
              color: '#334155',
              border: '1px solid #CBD5E1',
              borderRadius: '8px',
              padding: '7px 14px',
              fontSize: '12px',
              fontWeight: 700,
              cursor: 'pointer',
              width: 'fit-content',
              transition: 'all 0.15s ease',
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = '#F1F5F9'; e.currentTarget.style.color = '#0F172A'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = '#FFFFFF'; e.currentTarget.style.color = '#334155'; }}
          >
            View Details →
          </button>
        </div>

        {/* Card 3: AI Conversation Intelligence */}
        <div
          style={{
            background: 'linear-gradient(150deg, #FAF5FF 0%, #FFFFFF 60%, #F5F3FF 100%)',
            border: '1px solid #E9D5FF',
            borderRadius: '14px',
            padding: '20px',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
            boxShadow: '0 2px 8px rgba(139, 92, 246, 0.06)',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Top accent strip */}
          <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '3px', background: 'linear-gradient(90deg, #8B5CF6, #6366F1)' }} />

          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '11px', fontWeight: 800, color: '#6D28D9', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                <span style={{ fontSize: '12px' }}>✦</span>
                <span>AI Context</span>
              </div>
              <span style={{ fontSize: '11px', fontWeight: 700, background: '#EDE9FE', color: '#6D28D9', padding: '2px 8px', borderRadius: '10px' }}>
                Thread Active
              </span>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '10px' }}>
              <h4 style={{ fontSize: '15px', fontWeight: 800, color: '#0F172A', margin: 0 }}>
                {leadId ? `Source Lead #${leadId}` : 'Direct RFQ Creation'}
              </h4>
            </div>

            {/* Micro stats */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '12px' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: 'rgba(255, 255, 255, 0.7)', padding: '5px 8px', borderRadius: '6px', border: '1px solid #F3E8FF' }}>
                <span style={{ color: '#64748B' }}>💬 Inbound Emails Parsed</span>
                <strong style={{ color: '#0F172A', fontWeight: 800 }}>2 Received</strong>
              </div>

              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: 'rgba(255, 255, 255, 0.7)', padding: '5px 8px', borderRadius: '6px', border: '1px solid #F3E8FF' }}>
                <span style={{ color: '#64748B' }}>🔄 Information merged across replies</span>
                <strong style={{ color: '#059669', fontWeight: 800 }}>✓ Auto-Resolved</strong>
              </div>
            </div>
          </div>

          <button
            onClick={() => onSwitchTab('activity')}
            style={{
              marginTop: '16px',
              background: '#FAF5FF',
              color: '#7E22CE',
              border: '1px solid #D8B4FE',
              borderRadius: '8px',
              padding: '7px 14px',
              fontSize: '12px',
              fontWeight: 700,
              cursor: 'pointer',
              width: 'fit-content',
              transition: 'all 0.15s ease',
              boxShadow: '0 1px 3px rgba(126, 34, 206, 0.1)',
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = '#F3E8FF'; e.currentTarget.style.borderColor = '#C084FC'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = '#FAF5FF'; e.currentTarget.style.borderColor = '#D8B4FE'; }}
          >
            ✦ AI Conversation Summary
          </button>
        </div>

        {/* Card 4: Document Readiness */}
        <div
          style={{
            background: '#FFFFFF',
            border: '1px solid #E2E8F0',
            borderRadius: '14px',
            padding: '20px',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
            boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Top accent strip */}
          <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '3px', background: 'linear-gradient(90deg, #F59E0B, #E11D48)' }} />

          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '11px', fontWeight: 800, color: '#0F172A', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: '12px' }}>
              <span style={{ fontSize: '13px' }}>📄</span>
              <span>Document Readiness</span>
            </div>

            <div style={{ display: 'flex', alignItems: 'baseline', gap: '16px', marginBottom: '8px' }}>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: '4px' }}>
                <span style={{ fontSize: '24px', fontWeight: 900, color: '#0F172A', letterSpacing: '-0.02em' }}>{approvedDocsCount}</span>
                <span style={{ fontSize: '11px', color: '#059669', fontWeight: 700, textTransform: 'uppercase' }}>Approved</span>
              </div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: '4px' }}>
                <span style={{ fontSize: '24px', fontWeight: 900, color: missingDocsCount > 0 ? '#E11D48' : '#64748B', letterSpacing: '-0.02em' }}>{missingDocsCount}</span>
                <span style={{ fontSize: '11px', color: missingDocsCount > 0 ? '#E11D48' : '#64748B', fontWeight: 700, textTransform: 'uppercase' }}>Missing</span>
              </div>
            </div>

            <div style={{ fontSize: '11.5px', color: '#64748B', fontWeight: 600, marginBottom: '10px' }}>
              {approvedDocsCount} / {requiredDocsCount} Required Documents Complete
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '5px', fontSize: '11.5px' }}>
              {currentStageDocs.length > 0 ? (
                currentStageDocs.map(doc => {
                  const status = doc.document_status || (doc.status === 'SATISFIED' ? 'APPROVED' : doc.status);
                  const isApproved = status === 'APPROVED' || doc.status === 'SATISFIED';
                  const isReview = status === 'UNDER_REVIEW' || status === 'UPLOADED';
                  const bg = isApproved ? '#ECFDF5' : isReview ? '#FEF3C7' : '#FFF1F2';
                  const border = isApproved ? '#A7F3D0' : isReview ? '#FDE68A' : '#FECDD3';
                  const textColor = isApproved ? '#059669' : isReview ? '#D97706' : '#E11D48';
                  const label = isApproved ? '✓ Approved' : isReview ? '⏳ Under Review' : '⚠ Missing';

                  return (
                    <div key={doc.doc_type} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: bg, padding: '4px 8px', borderRadius: '6px', border: `1px solid ${border}` }}>
                      <span style={{ color: '#334155', fontWeight: 600 }}>📋 {doc.title}</span>
                      <span style={{ color: textColor, fontWeight: 800, fontSize: '10.5px' }}>
                        {label}
                      </span>
                    </div>
                  );
                })
              ) : (
                <>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#FFF1F2', padding: '4px 8px', borderRadius: '6px', border: '1px solid #FECDD3' }}>
                    <span style={{ color: '#334155', fontWeight: 600 }}>📋 Commercial Invoice</span>
                    <span style={{ color: '#E11D48', fontWeight: 800, fontSize: '10.5px' }}>⚠ Missing</span>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#FFF1F2', padding: '4px 8px', borderRadius: '6px', border: '1px solid #FECDD3' }}>
                    <span style={{ color: '#334155', fontWeight: 600 }}>📦 Packing List</span>
                    <span style={{ color: '#E11D48', fontWeight: 800, fontSize: '10.5px' }}>⚠ Missing</span>
                  </div>
                </>
              )}
            </div>
          </div>

          <button
            onClick={() => onSwitchTab('documents')}
            style={{
              marginTop: '16px',
              background: '#FFFFFF',
              color: '#4F46E5',
              border: '1px solid #C7D2FE',
              borderRadius: '8px',
              padding: '7px 14px',
              fontSize: '12px',
              fontWeight: 700,
              cursor: 'pointer',
              width: 'fit-content',
              transition: 'all 0.15s ease',
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = '#EEF2FF'; e.currentTarget.style.borderColor = '#818CF8'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = '#FFFFFF'; e.currentTarget.style.borderColor = '#C7D2FE'; }}
          >
            Manage Documents →
          </button>
        </div>
      </div>


      {/* ── ROW 2: MIDDLE SECTION (Shipment Overview + Customer Ownership + Latest Alerts) ── */}
      <div style={{ display: 'grid', gridTemplateColumns: '1.7fr 1fr 1fr', gap: '18px', alignItems: 'stretch' }}>
        
        {/* Card 1: Shipment Overview (Route Visual Graphic) */}
        <div
          style={{
            background: '#FFFFFF',
            border: '1px solid #E2E8F0',
            borderRadius: '14px',
            padding: '20px',
            boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
        >
          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '14px' }}>
              <span style={{ fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                Shipment Overview
              </span>
              <span style={{ background: '#EFF6FF', color: '#1D4ED8', border: '1px solid #BFDBFE', borderRadius: '8px', padding: '2px 8px', fontSize: '11px', fontWeight: 700 }}>
                FCL Ocean Route
              </span>
            </div>

            {/* Glowing Map & Vessel Trajectory */}
            <div
              style={{
                height: '140px',
                background: 'linear-gradient(135deg, #0F172A 0%, #1E293B 100%)',
                borderRadius: '10px',
                position: 'relative',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '0 32px',
                overflow: 'hidden',
                marginBottom: '16px',
                boxShadow: 'inset 0 1px 3px rgba(0,0,0,0.5)',
              }}
            >
              {/* Curved Waypoint Trajectory */}
              <svg style={{ position: 'absolute', top: 0, left: 0, width: '100%', height: '100%', pointerEvents: 'none' }} viewBox="0 0 500 140">
                <defs>
                  <linearGradient id="routeGlow" x1="0%" y1="0%" x2="100%" y2="0%">
                    <stop offset="0%" stopColor="#38BDF8" stopOpacity="0.9" />
                    <stop offset="50%" stopColor="#818CF8" stopOpacity="1" />
                    <stop offset="100%" stopColor="#C084FC" stopOpacity="0.9" />
                  </linearGradient>
                </defs>
                <path d="M 60,70 Q 250,15 440,70" fill="none" stroke="url(#routeGlow)" strokeWidth="2.5" strokeDasharray="6,6" />
              </svg>

              {/* Origin Node */}
              <div style={{ position: 'relative', zIndex: 2 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                  <span style={{ fontSize: '12px' }}>🇮🇳</span>
                  <span style={{ fontSize: '10px', color: '#38BDF8', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.04em' }}>Origin Port</span>
                </div>
                <div style={{ fontSize: '14px', fontWeight: 900, color: '#FFFFFF', marginTop: '2px', letterSpacing: '-0.01em' }}>
                  {rfq?.origin || 'Nhava Sheva'}
                </div>
                <div style={{ display: 'inline-block', background: 'rgba(56, 189, 248, 0.2)', color: '#38BDF8', border: '1px solid rgba(56, 189, 248, 0.4)', borderRadius: '4px', padding: '1px 5px', fontSize: '10px', fontWeight: 800, marginTop: '2px' }}>
                  INNSA · India
                </div>
              </div>

              {/* Center Floating Vessel Waypoint */}
              <div
                style={{
                  position: 'relative',
                  zIndex: 2,
                  width: '38px',
                  height: '38px',
                  borderRadius: '50%',
                  background: 'linear-gradient(135deg, #1E293B 0%, #0F172A 100%)',
                  border: '2px solid #818CF8',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '18px',
                  boxShadow: '0 0 16px rgba(129, 140, 248, 0.6)',
                }}
              >
                🚢
              </div>

              {/* Destination Node */}
              <div style={{ position: 'relative', zIndex: 2, textAlign: 'right' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: '4px' }}>
                  <span style={{ fontSize: '10px', color: '#C084FC', fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.04em' }}>Destination</span>
                  <span style={{ fontSize: '12px' }}>🇩🇪</span>
                </div>
                <div style={{ fontSize: '14px', fontWeight: 900, color: '#FFFFFF', marginTop: '2px', letterSpacing: '-0.01em' }}>
                  {rfq?.destination || 'Hamburg'}
                </div>
                <div style={{ display: 'inline-block', background: 'rgba(192, 132, 252, 0.2)', color: '#C084FC', border: '1px solid rgba(192, 132, 252, 0.4)', borderRadius: '4px', padding: '1px 5px', fontSize: '10px', fontWeight: 800, marginTop: '2px' }}>
                  DEHAM · Germany
                </div>
              </div>
            </div>
          </div>

          {/* Bottom 5 Structured Parameter Cards */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: '8px' }}>
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '8px 10px' }}>
              <div style={{ color: '#64748B', fontWeight: 700, fontSize: '10px', textTransform: 'uppercase' }}>🏷️ Incoterms</div>
              <div style={{ fontWeight: 800, color: '#0F172A', marginTop: '2px', fontSize: '12px' }}>{rfq?.incoterms || 'FOB'}</div>
              <div style={{ fontSize: '9.5px', color: '#64748B' }}>Free On Board</div>
            </div>

            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '8px 10px' }}>
              <div style={{ color: '#64748B', fontWeight: 700, fontSize: '10px', textTransform: 'uppercase' }}>🚢 Mode</div>
              <div style={{ fontWeight: 800, color: '#0F172A', marginTop: '2px', fontSize: '12px' }}>Ocean Freight</div>
              <div style={{ fontSize: '9.5px', color: '#64748B' }}>FCL Transport</div>
            </div>

            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '8px 10px' }}>
              <div style={{ color: '#64748B', fontWeight: 700, fontSize: '10px', textTransform: 'uppercase' }}>📅 Ready Date</div>
              <div style={{ fontWeight: 800, color: '#0F172A', marginTop: '2px', fontSize: '12px' }}>{targetDateStr}</div>
              <div style={{ fontSize: '9.5px', color: '#64748B' }}>Target Cargo Date</div>
            </div>

            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '8px 10px' }}>
              <div style={{ color: '#64748B', fontWeight: 700, fontSize: '10px', textTransform: 'uppercase' }}>📦 Commodity</div>
              <div style={{ fontWeight: 800, color: '#0F172A', marginTop: '2px', fontSize: '12px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }} title={commodity}>
                {commodity}
              </div>
              <div style={{ fontSize: '9.5px', color: '#64748B' }}>Standard Cargo</div>
            </div>

            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '8px 10px' }}>
              <div style={{ color: '#64748B', fontWeight: 700, fontSize: '10px', textTransform: 'uppercase' }}>📐 Container</div>
              <div style={{ fontWeight: 800, color: '#0F172A', marginTop: '2px', fontSize: '12px' }}>{containerType}</div>
              <div style={{ fontSize: '9.5px', color: '#64748B' }}>Dry Van Equipment</div>
            </div>
          </div>
        </div>

        {/* Card 2: Customer & Ownership */}
        <div
          style={{
            background: '#FFFFFF',
            border: '1px solid #E2E8F0',
            borderRadius: '14px',
            padding: '20px',
            boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
        >
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: '14px' }}>
              <span>👤</span>
              <span>Customer & Ownership</span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', fontSize: '12px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid #F1F5F9', paddingBottom: '7px' }}>
                <span style={{ color: '#64748B' }}>Customer</span>
                <strong style={{ color: '#0F172A', fontWeight: 800, textAlign: 'right' }}>{customerName}</strong>
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid #F1F5F9', paddingBottom: '7px' }}>
                <span style={{ color: '#64748B' }}>Contact Person</span>
                <strong style={{ color: '#0F172A', fontWeight: 700, textAlign: 'right' }}>{contactPerson}</strong>
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid #F1F5F9', paddingBottom: '7px' }}>
                <span style={{ color: '#64748B' }}>Email</span>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <strong style={{ color: '#0F172A', fontSize: '11.5px' }}>{contactEmail}</strong>
                  <button
                    onClick={() => handleCopy(contactEmail, 'Email')}
                    style={{ background: '#F1F5F9', border: '1px solid #CBD5E1', borderRadius: '4px', cursor: 'pointer', fontSize: '11px', padding: '1px 5px' }}
                    title="Copy email"
                  >
                    ✉️
                  </button>
                </div>
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid #F1F5F9', paddingBottom: '7px' }}>
                <span style={{ color: '#64748B' }}>Phone</span>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <strong style={{ color: '#0F172A' }}>{contactPhone}</strong>
                  <button
                    onClick={() => handleCopy(contactPhone, 'Phone')}
                    style={{ background: '#F1F5F9', border: '1px solid #CBD5E1', borderRadius: '4px', cursor: 'pointer', fontSize: '11px', padding: '1px 5px' }}
                    title="Copy phone"
                  >
                    📞
                  </button>
                </div>
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid #F1F5F9', paddingBottom: '7px' }}>
                <span style={{ color: '#64748B' }}>Sales Owner</span>
                <strong style={{ color: '#0F172A', fontWeight: 700 }}>{salesOwner}</strong>
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#64748B' }}>RFQ Owner</span>
                <strong style={{ color: '#0F172A', fontWeight: 700 }}>{rfqOwner}</strong>
              </div>
            </div>
          </div>
        </div>

        {/* Card 3: Latest Alerts */}
        <div
          style={{
            background: '#FFFFFF',
            border: '1px solid #E2E8F0',
            borderRadius: '14px',
            padding: '20px',
            boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
        >
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: '14px' }}>
              <span>🔔</span>
              <span>Latest Alerts</span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '9px', fontSize: '12px' }}>
              <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', borderRadius: '8px', padding: '8px 10px' }}>
                <div style={{ fontWeight: 800, color: '#B45309', display: 'flex', alignItems: 'center', gap: '4px' }}>
                  <span>⚠️</span>
                  <span>Packing List is missing</span>
                </div>
                <div style={{ fontSize: '11px', color: '#92400E', marginTop: '2px' }}>Required before carrier quotation</div>
              </div>

              <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', borderRadius: '8px', padding: '8px 10px' }}>
                <div style={{ fontWeight: 800, color: '#B45309', display: 'flex', alignItems: 'center', gap: '4px' }}>
                  <span>⚠️</span>
                  <span>Commercial Invoice is missing</span>
                </div>
                <div style={{ fontSize: '11px', color: '#92400E', marginTop: '2px' }}>Required before customs clearance</div>
              </div>

              <div style={{ background: '#ECFDF5', border: '1px solid #A7F3D0', borderRadius: '8px', padding: '8px 10px' }}>
                <div style={{ fontWeight: 800, color: '#065F46', display: 'flex', alignItems: 'center', gap: '4px' }}>
                  <span>✓</span>
                  <span>All shipment information complete</span>
                </div>
                <div style={{ fontSize: '11px', color: '#047857', marginTop: '2px' }}>Good to generate quotation</div>
              </div>
            </div>
          </div>

          <button
            onClick={() => onSwitchTab('requirements')}
            style={{
              marginTop: '16px',
              background: 'none',
              border: 'none',
              color: '#4F46E5',
              fontSize: '12px',
              fontWeight: 800,
              cursor: 'pointer',
              textAlign: 'left',
              padding: 0,
              display: 'inline-flex',
              alignItems: 'center',
              gap: '4px',
            }}
          >
            View All Alerts →
          </button>
        </div>

      </div>

      {/* ── ROW 3: CARRIER QUOTES & COMMERCIAL INTELLIGENCE SNAPSHOT ── */}
      <div
        style={{
          background: '#FFFFFF',
          border: '1px solid #E2E8F0',
          borderRadius: '14px',
          padding: '20px',
          boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '14px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span style={{ fontSize: '16px' }}>💰</span>
            <div>
              <h4 style={{ fontSize: '14px', fontWeight: 800, color: '#0F172A', margin: 0 }}>
                Carrier Quotations & Margin Intelligence
              </h4>
              <span style={{ fontSize: '11.5px', color: '#64748B' }}>
                {quotesList.length} carrier quote{quotesList.length === 1 ? '' : 's'} recorded · Direct multi-carrier comparison
              </span>
            </div>
          </div>

          <button
            onClick={() => onSwitchTab('quotes')}
            style={{
              background: '#4F46E5',
              color: '#FFFFFF',
              border: 'none',
              borderRadius: '8px',
              padding: '7px 14px',
              fontSize: '12px',
              fontWeight: 700,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
            }}
          >
            <span>Open Quotes Workspace →</span>
          </button>
        </div>

        {quotesList.length > 0 ? (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '12px' }}>
            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '10px', padding: '12px' }}>
              <div style={{ fontSize: '10.5px', fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>LOWEST BUY RATE</div>
              <div style={{ fontSize: '16px', fontWeight: 800, color: '#0F172A', marginTop: '2px' }}>
                {quotesSummary?.lowest_buy_amount ? `${primaryCurrency} ${Number(quotesSummary.lowest_buy_amount).toLocaleString()}` : '—'}
              </div>
              <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>Best liner ocean freight</div>
            </div>

            <div style={{ background: '#ECFDF5', border: '1px solid #A7F3D0', borderRadius: '10px', padding: '12px' }}>
              <div style={{ fontSize: '10.5px', fontWeight: 700, color: '#065F46', textTransform: 'uppercase' }}>BEST COMMERCIAL MARGIN</div>
              <div style={{ fontSize: '16px', fontWeight: 800, color: '#059669', marginTop: '2px' }}>
                {quotesSummary?.highest_margin_amount ? `+${primaryCurrency} ${Number(quotesSummary.highest_margin_amount).toLocaleString()} (${quotesSummary?.highest_margin_percentage}%)` : '—'}
              </div>
              <div style={{ fontSize: '11px', color: '#047857', marginTop: '2px' }}>Optimal commercial yield</div>
            </div>

            <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '10px', padding: '12px' }}>
              <div style={{ fontSize: '10.5px', fontWeight: 700, color: '#64748B', textTransform: 'uppercase' }}>RECOMMENDED CARRIER</div>
              <div style={{ fontSize: '14px', fontWeight: 800, color: '#4F46E5', marginTop: '2px' }}>
                {recommendedQuote ? recommendedQuote.carrier_name : approvedQuote ? approvedQuote.carrier_name : 'Selection Pending'}
              </div>
              <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
                {approvedQuote ? '✓ Operational Quote Approved' : recommendedQuote ? '★ Recommended by Pricing' : 'Awaiting recommendation'}
              </div>
            </div>
          </div>
        ) : (
          <div style={{ background: '#F8FAFC', border: '1px dashed #CBD5E1', borderRadius: '10px', padding: '16px', textAlign: 'center', fontSize: '12.5px', color: '#64748B' }}>
            <span>No carrier quotes submitted yet. When RFQ requirements are fulfilled, carrier rates will be compared here.</span>
          </div>
        )}
      </div>

      {/* ── ROW 3.5: DOWNSTREAM COMMERCIAL CLOSURE, BOOKING & SHIPMENT HANDOFF (Task 14) ── */}
      <div
        style={{
          background: '#FFFFFF',
          border: '1px solid #E2E8F0',
          borderRadius: '14px',
          padding: '20px',
          boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '14px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span style={{ fontSize: '16px' }}>🚢</span>
            <div>
              <h4 style={{ fontSize: '14px', fontWeight: 800, color: '#0F172A', margin: 0 }}>
                Downstream Commercial Closure & Execution Handoff
              </h4>
              <span style={{ fontSize: '11.5px', color: '#64748B' }}>
                End-to-end lineage: Commercial Quote Approval → Carrier Booking → Live Shipment Execution
              </span>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '8px' }}>
            <button
              onClick={() => onSwitchTab('booking')}
              style={{
                background: '#4F46E5',
                color: '#FFFFFF',
                border: 'none',
                borderRadius: '8px',
                padding: '6px 12px',
                fontSize: '11.5px',
                fontWeight: 700,
                cursor: 'pointer',
              }}
            >
              Booking Tab →
            </button>
            <button
              onClick={() => onSwitchTab('shipment')}
              style={{
                background: '#0F172A',
                color: '#FFFFFF',
                border: 'none',
                borderRadius: '8px',
                padding: '6px 12px',
                fontSize: '11.5px',
                fontWeight: 700,
                cursor: 'pointer',
              }}
            >
              Shipment Tab →
            </button>
          </div>
        </div>

        {/* 3 Columns: Commercial Closure -> Booking Handoff -> Shipment Execution */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '14px' }}>
          {/* Stage 1: Commercial Closure */}
          <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '10px', padding: '14px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
              <span style={{ fontSize: '10.5px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase' }}>1. Commercial Closure</span>
              <span style={{
                fontSize: '10px',
                fontWeight: 800,
                padding: '2px 6px',
                borderRadius: '4px',
                background: approvedQuote ? '#ECFDF5' : '#FFFBEB',
                color: approvedQuote ? '#065F46' : '#B45309',
                border: `1px solid ${approvedQuote ? '#A7F3D0' : '#FDE68A'}`,
              }}>
                {approvedQuote ? 'APPROVED' : 'PENDING'}
              </span>
            </div>
            {approvedQuote ? (
              <div>
                <div style={{ fontSize: '13px', fontWeight: 800, color: '#0F172A' }}>{approvedQuote.carrier_name}</div>
                <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
                  Sell: <strong style={{ color: '#059669' }}>{approvedQuote.currency} {approvedQuote.sell_price?.toLocaleString()}</strong>
                  {' '}(Margin: {approvedQuote.margin_percentage?.toFixed(1)}%)
                </div>
              </div>
            ) : (
              <div style={{ fontSize: '11.5px', color: '#64748B' }}>
                Quote requires commercial approval before booking can be generated.
              </div>
            )}
          </div>

          {/* Stage 2: Carrier Booking Handoff */}
          <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '10px', padding: '14px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
              <span style={{ fontSize: '10.5px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase' }}>2. Carrier Booking</span>
              <span style={{
                fontSize: '10px',
                fontWeight: 800,
                padding: '2px 6px',
                borderRadius: '4px',
                background: activeBooking?.status === 'CONFIRMED' ? '#ECFDF5' : activeBooking ? '#EFF6FF' : '#F1F5F9',
                color: activeBooking?.status === 'CONFIRMED' ? '#065F46' : activeBooking ? '#1E40AF' : '#64748B',
                border: `1px solid ${activeBooking?.status === 'CONFIRMED' ? '#A7F3D0' : activeBooking ? '#BFDBFE' : '#CBD5E1'}`,
              }}>
                {activeBooking ? activeBooking.status : 'NOT CREATED'}
              </span>
            </div>
            {activeBooking ? (
              <div>
                <div style={{ fontSize: '13px', fontWeight: 800, color: '#0F172A', fontFamily: 'monospace' }}>{activeBooking.booking_number}</div>
                <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
                  Carrier: <strong>{activeBooking.carrier_name}</strong> · Route: {activeBooking.origin_port} → {activeBooking.destination_port}
                </div>
              </div>
            ) : (
              <div style={{ fontSize: '11.5px', color: '#64748B' }}>
                {bookingEligibility?.is_eligible ? 'Eligible for booking creation.' : 'Prerequisites not yet met.'}
              </div>
            )}
          </div>

          {/* Stage 3: Shipment Execution */}
          <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '10px', padding: '14px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
              <span style={{ fontSize: '10.5px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase' }}>3. Shipment Execution</span>
              <span style={{
                fontSize: '10px',
                fontWeight: 800,
                padding: '2px 6px',
                borderRadius: '4px',
                background: activeShipment ? '#ECFDF5' : '#F1F5F9',
                color: activeShipment ? '#065F46' : '#64748B',
                border: `1px solid ${activeShipment ? '#A7F3D0' : '#CBD5E1'}`,
              }}>
                {activeShipment ? activeShipment.status : 'PENDING'}
              </span>
            </div>
            {activeShipment ? (
              <div>
                <div style={{ fontSize: '13px', fontWeight: 800, color: '#0F172A' }}>Shipment #{activeShipment.id}</div>
                <div style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
                  {activeShipment.carrier_name || activeShipment.carrier_scac} · Milestone: <strong style={{ color: '#0D9488' }}>{activeShipment.status}</strong>
                </div>
              </div>
            ) : (
              <div style={{ fontSize: '11.5px', color: '#64748B' }}>
                Execution record initiated once booking is confirmed.
              </div>
            )}
          </div>
        </div>
      </div>

      {/* ── ROW 4: BOTTOM RFQ ACTIVITY TIMELINE (Horizontal Milestone Stepper) ── */}
      <div
        style={{
          background: '#FFFFFF',
          border: '1px solid #E2E8F0',
          borderRadius: '14px',
          padding: '20px',
          boxShadow: '0 2px 8px rgba(15, 23, 42, 0.04)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '18px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '11px', fontWeight: 800, color: '#64748B', textTransform: 'uppercase', letterSpacing: '0.06em' }}>
            <span>⚡</span>
            <span>RFQ Activity Timeline</span>
          </div>
          <button
            onClick={() => onSwitchTab('activity')}
            style={{
              background: 'none',
              border: 'none',
              color: '#4F46E5',
              fontSize: '12.5px',
              fontWeight: 800,
              cursor: 'pointer',
              padding: 0,
            }}
          >
            View Full Activity →
          </button>
        </div>


        {/* Stepper with connected nodes */}
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', position: 'relative', overflowX: 'auto', paddingBottom: '6px' }}>
          
          {/* Connecting Line */}
          <div
            style={{
              position: 'absolute',
              top: '18px',
              left: '40px',
              right: '40px',
              height: '3px',
              background: 'linear-gradient(90deg, #10B981, #3B82F6, #8B5CF6, #F59E0B, #3B82F6, #10B981, #8B5CF6)',
              zIndex: 1,
              borderRadius: '2px',
            }}
          />

          {defaultMilestones.map((step, idx) => (
            <div
              key={idx}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                textAlign: 'center',
                minWidth: '125px',
                position: 'relative',
                zIndex: 2,
              }}
            >
              {/* Step Icon Node */}
              <div
                style={{
                  width: '36px',
                  height: '36px',
                  borderRadius: '50%',
                  background: step.bg,
                  border: `2px solid ${step.color}`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '14px',
                  marginBottom: '8px',
                  boxShadow: `0 2px 6px ${step.border}`,
                  transition: 'transform 0.15s ease',
                  cursor: 'pointer',
                }}
                onMouseEnter={(e) => { e.currentTarget.style.transform = 'scale(1.1)'; }}
                onMouseLeave={(e) => { e.currentTarget.style.transform = 'scale(1)'; }}
              >
                {step.icon}
              </div>

              {/* Title & Sub */}
              <div style={{ fontSize: '11.5px', fontWeight: 800, color: '#0F172A' }}>
                {step.title}
              </div>
              <div style={{ fontSize: '10.5px', color: '#64748B', marginTop: '2px', fontWeight: 500 }}>
                {step.sub}
              </div>

              {/* Timestamp */}
              <div style={{ fontSize: '10px', color: '#94A3B8', marginTop: '4px', fontWeight: 600 }}>
                {step.date}
              </div>
              <div style={{ fontSize: '9.5px', color: '#94A3B8' }}>
                {step.time}
              </div>
            </div>
          ))}
        </div>
      </div>

    </div>
  );
}
