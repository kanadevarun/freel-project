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
      case 'Pending': return 'status-badge pending';
      case 'Approved': return 'status-badge approved';
      case 'Rejected': return 'status-badge rejected';
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

  return (
    <tr className="approval-row" onClick={() => onSelect(item)}>
      {/* REQUEST */}
      <td>
        <div className="request-cell">
          <div className="icon-wrapper">
            {getItemIcon(item.type)}
          </div>
          <div>
            <strong className="request-title">{item.title}</strong>
            <span className="request-id">{item.id}</span>
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
          {item.status !== 'Approved' && item.status !== 'Rejected' && (
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
