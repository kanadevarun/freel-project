import React, { useState } from 'react';
import {
  FileText, Check, X, MoreVertical, Link as LinkIcon,
  Building, DollarSign, Truck
} from 'lucide-react';

export default function ApprovalRow({ item, onSelect, onApprove, onOpenRejectModal }) {
  const [approving, setApproving] = useState(false);

  const getTypeBadgeClass = (type) => {
    switch (type) {
      case 'Document Approval': return 'type-badge document';
      case 'Commercial Approval': return 'type-badge commercial';
      case 'Operations Approval': return 'type-badge operations';
      case 'Finance Approval': return 'type-badge finance';
      default: return 'type-badge default';
    }
  };

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'Overdue': return 'status-badge overdue';
      case 'Pending':
      case 'PENDING_APPROVAL': return 'status-badge pending';
      case 'Returned for Changes':
      case 'RETURNED_FOR_CHANGES': return 'status-badge in-progress';
      case 'Approved':
      case 'APPROVED':
      case 'Completed':
      case 'COMPLETED': return 'status-badge approved';
      case 'Rejected':
      case 'REJECTED': return 'status-badge rejected';
      case 'Cancelled':
      case 'CANCELLED': return 'status-badge cancelled';
      case 'Expired':
      case 'EXPIRED': return 'status-badge overdue';
      case 'Executing':
      case 'EXECUTING': return 'status-badge in-progress';
      case 'Failed':
      case 'FAILED': return 'status-badge rejected';
      default: return 'status-badge pending';
    }
  };

  const getItemIcon = (type) => {
    switch (type) {
      case 'Document Approval': return <FileText size={18} className="req-icon red" />;
      case 'Commercial Approval': return <Building size={18} className="req-icon blue" />;
      case 'Operations Approval': return <Truck size={18} className="req-icon orange" />;
      case 'Finance Approval': return <DollarSign size={18} className="req-icon green" />;
      default: return <FileText size={18} className="req-icon blue" />;
    }
  };

  const handleQuickApprove = async (e) => {
    e.stopPropagation();
    if (approving) return;
    try {
      setApproving(true);
      await onApprove(item, 'Quick approved from table row');
      setApproving(false);
    } catch (err) {
      console.error(err);
      setApproving(false);
    }
  };

  const handleQuickReject = (e) => {
    e.stopPropagation();
    onOpenRejectModal(item);
  };

  const isTerminal = ['Approved', 'APPROVED', 'Rejected', 'REJECTED', 'Cancelled', 'CANCELLED', 'Expired', 'EXPIRED', 'Completed', 'COMPLETED'].includes(item.status);

  return (
    <tr className="approval-row" onClick={() => onSelect(item)}>
      {/* REQUEST */}
      <td>
        <div className="request-cell">
          <div className="icon-wrapper">
            {getItemIcon(item.type)}
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
              <strong className="request-title">{item.title}</strong>
              {item.actorType === 'AI_AGENT' && (
                <span style={{ fontSize: '0.68rem', fontWeight: 800, padding: '1px 6px', borderRadius: 4, background: '#EDE9FE', color: '#6D28D9' }}>
                  🤖 AI
                </span>
              )}
              {item.riskLevel && (
                <span style={{
                  fontSize: '0.68rem',
                  fontWeight: 800,
                  padding: '1px 6px',
                  borderRadius: 4,
                  background: item.riskLevel === 'CRITICAL' || item.riskLevel === 'HIGH_RISK' ? '#FEE2E2' : '#EFF6FF',
                  color: item.riskLevel === 'CRITICAL' || item.riskLevel === 'HIGH_RISK' ? '#B91C1C' : '#1D4ED8',
                }}>
                  {item.riskLevel}
                </span>
              )}
              {item.externalCommunication && (
                <span style={{ fontSize: '0.68rem', fontWeight: 700, padding: '1px 5px', borderRadius: 4, background: '#EFF6FF', color: '#1E40AF' }}>
                  ✉ Outbound
                </span>
              )}
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 2 }}>
              <span className="request-id">{item.id}</span>
              {item.actionName && (
                <span style={{ fontSize: '0.72rem', color: '#64748B', fontFamily: 'monospace' }}>
                  • {item.actionName}
                </span>
              )}
            </div>
          </div>
        </div>
      </td>

      {/* TYPE */}
      <td>
        <span className={getTypeBadgeClass(item.type)}>
          {item.type}
        </span>
      </td>

      {/* RELATED TO */}
      <td>
        <div className="related-cell">
          <div className="related-ref">
            <LinkIcon size={12} className="link-icon" />
            <span>{item.relatedRef}</span>
          </div>
          <div className="customer-subtext">{item.customerName}</div>
        </div>
      </td>

      {/* REQUESTED BY */}
      <td>
        <div className="requester-cell">
          <div className="avatar-circle">{item.avatar}</div>
          <div>
            <div className="requester-name">{item.requesterName}</div>
            <div className="department-subtext">{item.department}</div>
          </div>
        </div>
      </td>

      {/* DUE DATE */}
      <td>
        <div className="duedate-cell">
          <div className="due-date">{item.dueDate}</div>
          <div className={`due-subtext ${item.dueText === 'Overdue' ? 'overdue-text' : ''}`}>
            {item.dueText}
          </div>
        </div>
      </td>

      {/* STATUS */}
      <td>
        <span className={getStatusBadgeClass(item.status)}>
          {item.status}
        </span>
      </td>

      {/* ACTIONS */}
      <td style={{ textAlign: 'right' }} onClick={(e) => e.stopPropagation()}>
        <div className="row-actions-group">
          {!isTerminal && (
            <>
              <button
                type="button"
                className="action-icon-btn check-btn"
                title="Approve Request"
                disabled={approving}
                onClick={handleQuickApprove}
              >
                {approving ? '...' : <Check size={14} />}
              </button>

              <button
                type="button"
                className="action-icon-btn x-btn"
                title="Reject Request"
                onClick={handleQuickReject}
              >
                <X size={14} />
              </button>
            </>
          )}

          <button
            type="button"
            className="action-icon-btn more-btn"
            title="View Details"
            onClick={() => onSelect(item)}
          >
            <MoreVertical size={14} />
          </button>
        </div>
      </td>
    </tr>
  );
}
