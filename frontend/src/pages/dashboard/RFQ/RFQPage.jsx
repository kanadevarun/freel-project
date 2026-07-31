import React, { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { rfqService } from '../../../services/rfqService';
import PageHeader from '../../../components/dashboard/PageHeader';
import RFQList from './RFQList';
import RFQBuilder from './RFQBuilder';
import RFQDetail from './RFQDetail';
import './RFQPage.css';
import { RFQ_STAGES } from './constants';

export default function RFQPage() {
  const [rfqs, setRfqs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('ALL');
  const [showBuilder, setShowBuilder] = useState(false);
  const [selectedRfqId, setSelectedRfqId] = useState(null);

  const fetchRFQs = async () => {
    setIsLoading(true);
    try {
      const res = await rfqService.listRFQs();
      setRfqs(res.data.rfqs || []);
    } catch (error) {
      console.error('Failed to fetch RFQs:', error);
      toast.error('Failed to load RFQs');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchRFQs();
  }, []);

  const handleRowClick = (rfq) => {
    setSelectedRfqId(rfq.id);
  };

  // Filter logic
  const filteredRFQs = rfqs.filter((rfq) => {
    if (activeTab === 'ALL') return true;
    if (activeTab === 'DRAFT') return rfq.stage === RFQ_STAGES.STAGE_RFQ_CREATED;
    if (activeTab === 'AWAITING_QUOTE') return rfq.stage === RFQ_STAGES.STAGE_PRICING_ASSIGNED;
    if (activeTab === 'WON_LOST') return rfq.stage === RFQ_STAGES.STAGE_WON || rfq.stage === RFQ_STAGES.STAGE_LOST;
    return true;
  });

  return (
    <div className="rfq-page-container">
      <div className="page-header-wrapper">
        <PageHeader 
          title="Shipment Requests (RFQs)"
          subtitle="Manage customer quotes and initiate shipments"
          onAdd={() => setShowBuilder(true)}
          addButtonText="New RFQ"
        />
      </div>

      <div className="rfq-tabs">
        <div 
          className={`rfq-tab ${activeTab === 'ALL' ? 'active' : ''}`}
          onClick={() => setActiveTab('ALL')}
        >
          All Requests
        </div>
        <div 
          className={`rfq-tab ${activeTab === 'DRAFT' ? 'active' : ''}`}
          onClick={() => setActiveTab('DRAFT')}
        >
          New / Draft
        </div>
        <div 
          className={`rfq-tab ${activeTab === 'AWAITING_QUOTE' ? 'active' : ''}`}
          onClick={() => setActiveTab('AWAITING_QUOTE')}
        >
          Awaiting Quote
        </div>
        <div 
          className={`rfq-tab ${activeTab === 'WON_LOST' ? 'active' : ''}`}
          onClick={() => setActiveTab('WON_LOST')}
        >
          Won/Lost
        </div>
      </div>

      <RFQList 
        rfqs={filteredRFQs} 
        isLoading={isLoading} 
        onRowClick={handleRowClick}
      />

      {showBuilder && (
        <RFQBuilder 
          onClose={() => setShowBuilder(false)}
          onSuccess={() => {
            setShowBuilder(false);
            fetchRFQs();
          }}
        />
      )}

      {selectedRfqId && (
        <RFQDetail 
          rfqId={selectedRfqId}
          onClose={() => {
            setSelectedRfqId(null);
            fetchRFQs(); // Refresh list to get updated stage
          }}
        />
      )}
    </div>
  );
}
