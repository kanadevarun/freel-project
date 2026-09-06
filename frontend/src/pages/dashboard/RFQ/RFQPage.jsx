import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { rfqService } from '../../../services/rfqService';
import RFQList from './RFQList';
import RFQBuilder from './RFQBuilder';
import RFQStatusLegend from './components/RFQStatusLegend';
import { calculateRFQCompleteness } from './utils/completeness';
import { RFQ_STAGES } from './constants';
import { useAuth } from '../../../context/AuthContext';
import './RFQPage.css';

export default function RFQPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [rfqs, setRfqs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('ALL');
  const [showBuilder, setShowBuilder] = useState(false);

  // Search & Filter state
  const [searchQuery, setSearchQuery] = useState('');
  const [modeFilter, setModeFilter] = useState('ALL');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [incotermFilter, setIncotermFilter] = useState('ALL');
  const [showFilterDropdowns, setShowFilterDropdowns] = useState(false);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 8;

  const fetchRFQs = async () => {
    setIsLoading(true);
    try {
      const res = await rfqService.listRFQs();
      const items = Array.isArray(res) ? res : (res?.data || res?.rfqs || []);
      setRfqs(items);
    } catch (error) {
      console.error('Failed to fetch RFQs:', error);
      toast.error('Failed to load RFQs');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchRFQs();
    const searchParams = new URLSearchParams(window.location.search);
    const paramRfqId = searchParams.get('rfqId') || searchParams.get('openRfq');
    if (paramRfqId) {
      navigate(`/dashboard/rfqs/${paramRfqId}`);
    }
  }, [navigate]);

  // Compute tab counts
  const tabCounts = useMemo(() => {
    const counts = {
      ALL: rfqs.length,
      DRAFT: 0,
      AWAITING_QUOTE: 0,
      WON: 0,
      LOST: 0,
    };

    rfqs.forEach(rfq => {
      const stage = rfq.stage;
      if (stage === RFQ_STAGES.STAGE_RFQ_CREATED || !stage || stage === 'DRAFT') {
        counts.DRAFT++;
      } else if (stage === RFQ_STAGES.STAGE_PRICING_ASSIGNED || stage === RFQ_STAGES.STAGE_QUOTE_GENERATED) {
        counts.AWAITING_QUOTE++;
      } else if (stage === RFQ_STAGES.STAGE_WON) {
        counts.WON++;
      } else if (stage === RFQ_STAGES.STAGE_LOST) {
        counts.LOST++;
      }
    });

    return counts;
  }, [rfqs]);

  // Filter logic
  const filteredRFQs = useMemo(() => {
    return rfqs.filter((rfq) => {
      // Tab filter
      if (activeTab === 'DRAFT') {
        if (rfq.stage !== RFQ_STAGES.STAGE_RFQ_CREATED && rfq.stage && rfq.stage !== 'DRAFT') return false;
      } else if (activeTab === 'AWAITING_QUOTE') {
        if (rfq.stage !== RFQ_STAGES.STAGE_PRICING_ASSIGNED && rfq.stage !== RFQ_STAGES.STAGE_QUOTE_GENERATED) return false;
      } else if (activeTab === 'WON') {
        if (rfq.stage !== RFQ_STAGES.STAGE_WON) return false;
      } else if (activeTab === 'LOST') {
        if (rfq.stage !== RFQ_STAGES.STAGE_LOST) return false;
      }

      // Search query filter
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const rfqNum = (rfq.rfq_number || '').toLowerCase();
        const cust = (rfq.customer_name || '').toLowerCase();
        const origin = (rfq.origin || '').toLowerCase();
        const dest = (rfq.destination || '').toLowerCase();
        const incoterm = (rfq.incoterms || '').toLowerCase();
        const email = (rfq.customer_email || '').toLowerCase();

        if (
          !rfqNum.includes(q) &&
          !cust.includes(q) &&
          !origin.includes(q) &&
          !dest.includes(q) &&
          !incoterm.includes(q) &&
          !email.includes(q)
        ) {
          return false;
        }
      }

      // Mode filter (default Ocean Freight)
      if (modeFilter !== 'ALL') {
        const mode = 'Ocean Freight'; // RFQ mode
        if (!mode.toLowerCase().includes(modeFilter.toLowerCase())) return false;
      }

      // Status filter
      if (statusFilter !== 'ALL') {
        const completeness = calculateRFQCompleteness(rfq);
        if (completeness.operationalStatus !== statusFilter && rfq.stage !== statusFilter) {
          return false;
        }
      }

      // Incoterm filter
      if (incotermFilter !== 'ALL') {
        if ((rfq.incoterms || '').toUpperCase() !== incotermFilter.toUpperCase()) {
          return false;
        }
      }

      return true;
    });
  }, [rfqs, activeTab, searchQuery, modeFilter, statusFilter, incotermFilter]);

  // Paginated RFQs
  const totalResults = filteredRFQs.length;
  const totalPages = Math.ceil(totalResults / pageSize) || 1;
  const paginatedRFQs = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filteredRFQs.slice(start, start + pageSize);
  }, [filteredRFQs, currentPage, pageSize]);

  const userName = user?.first_name || 'Varun';

  return (
    <div className="rfq-page-container">
      {/* 1. Header with Welcome Title & Action Buttons */}
      <div className="rfq-header-row">
        <div>
          <h1 className="rfq-page-title">Request for Quotations (RFQs)</h1>
          <p className="rfq-page-subtitle">Manage freight rate inquiries, route specifications, carrier tariff matching, and customer quote generation.</p>
        </div>

        <div className="rfq-header-actions">
          <button
            className="rfq-btn rfq-btn-outline"
            onClick={() => toast('CSV import wizard will open in the next update.', { icon: '📥' })}
          >
            <span>📥 Import CSV</span>
          </button>
          <button
            className="rfq-btn rfq-btn-primary"
            onClick={() => setShowBuilder(true)}
          >
            <span>+ New RFQ</span>
          </button>
        </div>
      </div>

      {/* 2. Operational Tabs */}
      <div className="rfq-tabs-row">
        {[
          { id: 'ALL', label: 'All RFQs', count: tabCounts.ALL },
          { id: 'DRAFT', label: 'New / Draft', count: tabCounts.DRAFT },
          { id: 'AWAITING_QUOTE', label: 'Awaiting Quote', count: tabCounts.AWAITING_QUOTE },
          { id: 'WON', label: 'Won', count: tabCounts.WON },
          { id: 'LOST', label: 'Lost', count: tabCounts.LOST },
        ].map((tab) => (
          <button
            key={tab.id}
            className={`rfq-tab-button ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => {
              setActiveTab(tab.id);
              setCurrentPage(1);
            }}
          >
            <span>{tab.label}</span>
            <span className="rfq-tab-count">
              {isLoading && rfqs.length === 0 ? '···' : `(${tab.count})`}
            </span>
          </button>
        ))}
      </div>

      {isLoading && rfqs.length === 0 && (
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          padding: '10px 16px',
          background: 'linear-gradient(90deg, #EFF6FF 0%, #EEF2FF 100%)',
          border: '1px solid #DBEAFE',
          borderRadius: '10px',
          marginBottom: '16px',
          fontSize: '12px',
          color: '#1E40AF',
          fontWeight: 600
        }}>
          <span style={{ animation: 'spin 1.5s linear infinite', display: 'inline-block' }}>🔄</span>
          <span>Loading shipment requests from backend...</span>
        </div>
      )}

      {/* 3. Search & Filter Bar */}
      <div className="rfq-filter-bar">
        <div className="rfq-search-box">
          <span className="rfq-search-icon">🔍</span>
          <input
            type="text"
            className="rfq-search-input"
            placeholder="Search by RFQ #, customer, route, mode..."
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
              setCurrentPage(1);
            }}
          />
          {searchQuery && (
            <button className="rfq-search-clear" onClick={() => setSearchQuery('')}>✕</button>
          )}
        </div>

        <div className="rfq-filter-controls">
          {/* Mode Dropdown */}
          <div className="rfq-filter-select-wrapper">
            <select
              className="rfq-filter-select"
              value={modeFilter}
              onChange={(e) => {
                setModeFilter(e.target.value);
                setCurrentPage(1);
              }}
            >
              <option value="ALL">Mode</option>
              <option value="Ocean">Ocean Freight</option>
              <option value="Air">Air Freight</option>
            </select>
          </div>

          {/* Status Dropdown */}
          <div className="rfq-filter-select-wrapper">
            <select
              className="rfq-filter-select"
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setCurrentPage(1);
              }}
            >
              <option value="ALL">Status</option>
              <option value="READY_FOR_QUOTATION">Ready for Quotation</option>
              <option value="QUOTE_GENERATED">Quote Generated</option>
              <option value="PRICING_ASSIGNED">Pricing Assigned</option>
              <option value="INFORMATION_REQUIRED">Information Required</option>
              <option value="STAGE_RFQ_CREATED">Draft</option>
            </select>
          </div>

          {/* Incoterms Dropdown */}
          <div className="rfq-filter-select-wrapper">
            <select
              className="rfq-filter-select"
              value={incotermFilter}
              onChange={(e) => {
                setIncotermFilter(e.target.value);
                setCurrentPage(1);
              }}
            >
              <option value="ALL">Incoterms</option>
              <option value="FOB">FOB</option>
              <option value="CIF">CIF</option>
              <option value="EXW">EXW</option>
              <option value="DDP">DDP</option>
              <option value="DAP">DAP</option>
              <option value="FCA">FCA</option>
            </select>
          </div>

          {/* Filters Button */}
          <button
            className={`rfq-filter-toggle-btn ${showFilterDropdowns ? 'active' : ''}`}
            onClick={() => {
              if (modeFilter !== 'ALL' || statusFilter !== 'ALL' || incotermFilter !== 'ALL') {
                setModeFilter('ALL');
                setStatusFilter('ALL');
                setIncotermFilter('ALL');
                toast.success('Filters cleared');
              } else {
                setShowFilterDropdowns(!showFilterDropdowns);
              }
            }}
          >
            <span>Filters</span>
            <span style={{ fontSize: '13px' }}>⚙️</span>
          </button>
        </div>
      </div>

      {/* 4. Table */}
      <RFQList
        rfqs={paginatedRFQs}
        isLoading={isLoading}
        onRowClick={(rfq) => navigate(`/dashboard/rfqs/${rfq.id}`)}
        onNewRFQ={() => setShowBuilder(true)}
      />

      {/* 5. Pagination */}
      {!isLoading && filteredRFQs.length > 0 && (
        <div className="rfq-pagination-row">
          <div className="rfq-pagination-info">
            Showing {Math.min((currentPage - 1) * pageSize + 1, totalResults)} to{' '}
            {Math.min(currentPage * pageSize, totalResults)} of {totalResults} results
          </div>

          <div className="rfq-pagination-controls">
            <button
              className="rfq-page-nav-btn"
              disabled={currentPage <= 1}
              onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
            >
              ‹
            </button>

            {Array.from({ length: totalPages }, (_, i) => i + 1).slice(0, 5).map((pg) => (
              <button
                key={pg}
                className={`rfq-page-num-btn ${currentPage === pg ? 'active' : ''}`}
                onClick={() => setCurrentPage(pg)}
              >
                {pg}
              </button>
            ))}

            {totalPages > 5 && (
              <>
                <span style={{ color: '#94A3B8', padding: '0 4px' }}>...</span>
                <button
                  className={`rfq-page-num-btn ${currentPage === totalPages ? 'active' : ''}`}
                  onClick={() => setCurrentPage(totalPages)}
                >
                  {totalPages}
                </button>
              </>
            )}

            <button
              className="rfq-page-nav-btn"
              disabled={currentPage >= totalPages}
              onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
            >
              ›
            </button>
          </div>
        </div>
      )}

      {/* 6. Status Legend & Workflow Intelligence */}
      <RFQStatusLegend
        rfqs={rfqs}
        activeTab={activeTab}
        onSelectTab={(tabKey) => {
          setActiveTab(tabKey);
          setCurrentPage(1);
        }}
      />

      {/* RFQ Builder Modal */}
      {showBuilder && (
        <RFQBuilder
          onClose={() => setShowBuilder(false)}
          onSuccess={() => {
            setShowBuilder(false);
            fetchRFQs();
          }}
        />
      )}
    </div>
  );
}
