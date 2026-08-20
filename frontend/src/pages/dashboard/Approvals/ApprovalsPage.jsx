import React, { useState, useEffect } from 'react';
import { CheckSquare, CheckCircle2, XCircle, AlertCircle, ShieldAlert, ArrowRight, ExternalLink } from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import PageHeader from '../../../components/dashboard/PageHeader';
import ReviewModal from '../Contracts/ReviewModal';
import './ApprovalsPage.css';

export default function ApprovalsPage() {
  const [reviewItems, setReviewItems] = useState([]);
  const [activeTab, setActiveTab] = useState('PENDING');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedItem, setSelectedItem] = useState(null);
  const [actionSuccess, setActionSuccess] = useState(null);

  useEffect(() => {
    fetchReviewItems(activeTab);
  }, [activeTab]);

  const fetchReviewItems = async (status) => {
    try {
      setLoading(true);
      setError(null);
      const items = await contractsService.listReviewItems(status);
      setReviewItems(items || []);
    } catch (err) {
      console.error('Failed to load review items:', err);
      setError('Unable to load approval queue items. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async (id, correctedData, notes) => {
    try {
      await contractsService.approveReviewItem(id, correctedData, notes);
      setActionSuccess('Item approved and live rate catalog updated.');
      setSelectedItem(null);
      fetchReviewItems(activeTab);
      setTimeout(() => setActionSuccess(null), 4000);
    } catch (err) {
      console.error(err);
      setError(err?.message || 'Failed to approve item.');
    }
  };

  const handleReject = async (id, notes) => {
    try {
      await contractsService.rejectReviewItem(id, notes);
      setActionSuccess('Item rejected from ingestion catalog.');
      setSelectedItem(null);
      fetchReviewItems(activeTab);
      setTimeout(() => setActionSuccess(null), 4000);
    } catch (err) {
      console.error(err);
      setError(err?.message || 'Failed to reject item.');
    }
  };

  return (
    <div className="approvals-page">
      <PageHeader
        title="Approvals & Compliance Queue"
        subtitle="Human-in-the-loop review queue for AI rate extractions, margin anomalies, and compliance flags"
      />

      {/* Tabs */}
      <div className="approvals-tabs-row">
        {['PENDING', 'APPROVED', 'REJECTED'].map((tab) => (
          <button
            key={tab}
            className={`tab-pill-btn ${activeTab === tab ? 'active' : ''}`}
            onClick={() => setActiveTab(tab)}
          >
            {tab === 'PENDING' ? 'Pending Review' : tab === 'APPROVED' ? 'Approved History' : 'Rejected'}
          </button>
        ))}
      </div>

      {actionSuccess && (
        <div className="approvals-success-banner">
          <CheckCircle2 size={16} />
          <span>{actionSuccess}</span>
        </div>
      )}

      {error && (
        <div className="approvals-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={() => fetchReviewItems(activeTab)} className="btn-retry">Retry</button>
        </div>
      )}

      {loading ? (
        <div className="approvals-skeleton">
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton-approval-card" />
          ))}
        </div>
      ) : reviewItems.length === 0 ? (
        <div className="approvals-empty-state">
          <div className="empty-icon-box">✅</div>
          <h3>All clear — No {activeTab.toLowerCase()} items</h3>
          <p>
            {activeTab === 'PENDING'
              ? 'There are no pending contract rate extractions or invoice reconciliation items requiring manual human review.'
              : `No items currently classified under ${activeTab.toLowerCase()} status.`}
          </p>
        </div>
      ) : (
        <div className="approvals-list-grid">
          {reviewItems.map((item) => {
            const data = item.extracted_data || {};
            const flags = item.review_flags || [];

            return (
              <div key={item.id} className="approval-item-card">
                <div className="approval-card-top">
                  <div className="type-badge-row">
                    <span className="scac-tag">{data.carrier_scac || 'CARRIER'}</span>
                    <span className="lane-text">
                      {data.origin_port || 'ORIGIN'} → {data.destination_port || 'DEST'}
                    </span>
                  </div>
                  <div className="confidence-pill">
                    <span>Confidence: </span>
                    <strong>{item.confidence_score || 70}%</strong>
                  </div>
                </div>

                {/* Flags Row */}
                {flags.length > 0 && (
                  <div className="flags-row">
                    {flags.map((fl, idx) => (
                      <span key={idx} className="flag-tag">
                        <ShieldAlert size={12} />
                        {fl.replace('_', ' ')}
                      </span>
                    ))}
                  </div>
                )}

                {/* Extracted Details */}
                <div className="extracted-fields-grid">
                  <div className="field-group">
                    <span className="field-label">Ocean Freight</span>
                    <span className="field-val">${Number(data.ocean_freight || 0).toLocaleString()}</span>
                  </div>
                  <div className="field-group">
                    <span className="field-label">Total Buy Price</span>
                    <span className="field-val">${Number(data.total_buy_price || 0).toLocaleString()}</span>
                  </div>
                  <div className="field-group">
                    <span className="field-label">Validity</span>
                    <span className="field-val">
                      {data.valid_until ? new Date(data.valid_until).toLocaleDateString() : 'N/A'}
                    </span>
                  </div>
                </div>

                {/* Action Buttons */}
                {activeTab === 'PENDING' && (
                  <div className="approval-actions-row">
                    <button
                      className="btn-review-item"
                      onClick={() => setSelectedItem(item)}
                    >
                      <span>Review & Resolve</span>
                      <ArrowRight size={14} />
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Review Modal */}
      {selectedItem && (
        <ReviewModal
          item={selectedItem}
          onClose={() => setSelectedItem(null)}
          onApprove={handleApprove}
          onReject={handleReject}
        />
      )}
    </div>
  );
}
