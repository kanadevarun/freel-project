import React, { useState, useEffect, useCallback } from 'react';
import { 
  ShieldCheck, CheckCircle2, XCircle, Clock, AlertCircle, 
  RefreshCw, MessageSquare, User, FileText, ChevronRight
} from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import toast from 'react-hot-toast';
import './ContractApprovalPanel.css';

export default function ContractApprovalPanel({ contractId, contractStatus, onApprovalCompleted }) {
  const [approvals, setApprovals] = useState([]);
  const [isLoading, setIsLoading] = useState(false);

  // Decision Modal State
  const [activeDecision, setActiveDecision] = useState(null); // { request, action: 'APPROVE' | 'REJECT' }
  const [decisionNotes, setDecisionNotes] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);

  const fetchApprovals = useCallback(async () => {
    if (!contractId) return;
    setIsLoading(true);
    try {
      const res = await contractsService.getContractApprovals(contractId);
      setApprovals(res?.data || []);
    } catch (err) {
      console.error('Failed to load approvals', err);
      toast.error('Failed to load contract approval history');
    } finally {
      setIsLoading(false);
    }
  }, [contractId]);

  useEffect(() => {
    fetchApprovals();
  }, [fetchApprovals]);

  const handleDecisionSubmit = async (e) => {
    e.preventDefault();
    if (!activeDecision) return;

    setIsProcessing(true);
    try {
      if (activeDecision.action === 'APPROVE') {
        await contractsService.approveContractChange(contractId, activeDecision.request.id, {
          comment: decisionNotes
        });
        toast.success('Approval granted successfully');
      } else {
        await contractsService.rejectContractChange(contractId, activeDecision.request.id, {
          reason: decisionNotes || 'Rejected by authorized reviewer'
        });
        toast.success('Change request rejected');
      }

      setActiveDecision(null);
      setDecisionNotes('');
      await fetchApprovals();
      if (onApprovalCompleted) onApprovalCompleted();
    } catch (err) {
      console.error('Approval decision error', err);
      toast.error(err?.response?.data?.error?.message || 'Failed to submit decision');
    } finally {
      setIsProcessing(false);
    }
  };

  const pendingApprovals = approvals.filter(a => a.status === 'PENDING');
  const pastApprovals = approvals.filter(a => a.status !== 'PENDING');

  return (
    <div className="approval-panel-container">
      {/* Top Header */}
      <div className="approval-top-bar">
        <div className="approval-title-group">
          <div className="appr-icon-badge">
            <ShieldCheck size={16} />
          </div>
          <div>
            <h4>Approval Governance & Audit Trail</h4>
            <p>Formal approval decisions are recorded with immutable reviewer attribution and audit timestamps.</p>
          </div>
        </div>

        <button 
          className="appr-refresh-btn" 
          onClick={fetchApprovals} 
          title="Refresh Approvals"
          disabled={isLoading}
        >
          <RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />
        </button>
      </div>

      {isLoading ? (
        <div className="appr-loading">
          <RefreshCw size={24} className="spin-icon text-blue" />
          <span>Loading approval queue...</span>
        </div>
      ) : approvals.length === 0 ? (
        <div className="appr-empty-state">
          <ShieldCheck size={36} className="text-muted" />
          <h5>No Approval Requests</h5>
          <p>There are no active or historical approval workflows initiated for this contract.</p>
        </div>
      ) : (
        <div className="approval-sections">
          {/* Pending Queue */}
          {pendingApprovals.length > 0 && (
            <div className="appr-section">
              <div className="section-title">
                <Clock size={15} className="text-amber" />
                <h6>Pending Approval Actions ({pendingApprovals.length})</h6>
              </div>

              <div className="appr-cards-list">
                {pendingApprovals.map(req => (
                  <div key={req.id} className="appr-card pending">
                    <div className="appr-card-main">
                      <div className="appr-badge-row">
                        <span className="appr-type-badge">{req.approval_type}</span>
                        <span className="appr-status-badge pending">PENDING REVIEW</span>
                      </div>

                      <div className="appr-info-row">
                        <span><strong>Requested:</strong> {new Date(req.requested_at).toLocaleDateString()}</span>
                        {req.requested_by && <span><strong>By:</strong> User #{req.requested_by}</span>}
                        {req.assigned_to && <span><strong>Assigned:</strong> User #{req.assigned_to}</span>}
                      </div>
                    </div>

                    <div className="appr-card-actions">
                      <button 
                        className="btn-reject"
                        onClick={() => setActiveDecision({ request: req, action: 'REJECT' })}
                      >
                        <XCircle size={13} /> Reject
                      </button>
                      <button 
                        className="btn-approve"
                        onClick={() => setActiveDecision({ request: req, action: 'APPROVE' })}
                      >
                        <CheckCircle2 size={13} /> Approve
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Historical Decisions */}
          {pastApprovals.length > 0 && (
            <div className="appr-section">
              <div className="section-title">
                <FileText size={15} className="text-muted" />
                <h6>Decision History & Audit Trail</h6>
              </div>

              <div className="appr-cards-list">
                {pastApprovals.map(req => {
                  const isApproved = req.status === 'APPROVED';
                  const isRejected = req.status === 'REJECTED';
                  return (
                    <div key={req.id} className={`appr-card past ${req.status.toLowerCase()}`}>
                      <div className="appr-card-main">
                        <div className="appr-badge-row">
                          <span className="appr-type-badge">{req.approval_type}</span>
                          <span className={`appr-status-badge ${req.status.toLowerCase()}`}>
                            {req.status}
                          </span>
                        </div>

                        {req.decision_comment && (
                          <div className="appr-comment-box">
                            <MessageSquare size={12} className="text-muted" />
                            <span>"{req.decision_comment}"</span>
                          </div>
                        )}

                        <div className="appr-info-row">
                          {req.decided_at && <span><strong>Decided:</strong> {new Date(req.decided_at).toLocaleDateString()}</span>}
                          {req.decision_by && <span><strong>By:</strong> User #{req.decision_by}</span>}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Decision Modal */}
      {activeDecision && (
        <div className="appr-modal-overlay" onClick={() => setActiveDecision(null)}>
          <div className="appr-modal-card" onClick={e => e.stopPropagation()}>
            <div className="appr-modal-header">
              <h5>
                {activeDecision.action === 'APPROVE' ? 'Grant Approval' : 'Reject Change Request'}
              </h5>
              <button className="btn-close" onClick={() => setActiveDecision(null)}>×</button>
            </div>

            <form onSubmit={handleDecisionSubmit}>
              <div className="appr-modal-body">
                <p className="appr-modal-prompt">
                  {activeDecision.action === 'APPROVE' 
                    ? 'Confirming approval will advance this agreement or amendment to its next lifecycle state.'
                    : 'Please provide a clear reason for rejecting this change request.'
                  }
                </p>

                <div className="form-group">
                  <label>
                    {activeDecision.action === 'APPROVE' ? 'Decision Notes / Approval Reference (Optional)' : 'Rejection Reason *'}
                  </label>
                  <textarea 
                    rows={3}
                    required={activeDecision.action === 'REJECT'}
                    placeholder={activeDecision.action === 'APPROVE' ? 'e.g. Approved per commercial review...' : 'e.g. Rate increase exceeds allowable contract cap...'}
                    value={decisionNotes}
                    onChange={e => setDecisionNotes(e.target.value)}
                  />
                </div>
              </div>

              <div className="appr-modal-footer">
                <button type="button" className="btn-cancel" onClick={() => setActiveDecision(null)}>
                  Cancel
                </button>
                <button 
                  type="submit" 
                  className={activeDecision.action === 'APPROVE' ? 'btn-confirm-approve' : 'btn-confirm-reject'}
                  disabled={isProcessing}
                >
                  {isProcessing ? 'Recording...' : activeDecision.action === 'APPROVE' ? 'Confirm Approval' : 'Confirm Rejection'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
