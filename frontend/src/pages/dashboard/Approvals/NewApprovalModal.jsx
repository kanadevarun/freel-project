import React, { useState, useEffect } from 'react';
import { X, Plus, AlertCircle, ShieldAlert, User, Calendar, Tag } from 'lucide-react';

export default function NewApprovalModal({
  isOpen,
  onClose,
  onSubmit,
  existingApprovals = [],
  initialCategory = 'DOCUMENTS',
  initialType = 'Document Approval',
  initialRelatedRef = '',
  initialCustomerName = '',
  initialShipmentId = 0,
  initialDocumentId = 0,
  initialCustomerId = 0,
}) {
  const [title, setTitle] = useState('');
  const [category, setCategory] = useState(initialCategory);
  const [type, setType] = useState(initialType);
  const [relatedRef, setRelatedRef] = useState(initialRelatedRef);
  const [customerName, setCustomerName] = useState(initialCustomerName);
  const [requesterName, setRequesterName] = useState('Varun Kanade');
  const [department, setDepartment] = useState('Operations');
  const [assignedTo, setAssignedTo] = useState('Arjun Singh (Operations Manager)');
  const [dueDate, setDueDate] = useState(() => {
    const d = new Date();
    d.setDate(d.getDate() + 7);
    return d.toISOString().split('T')[0];
  });
  const [priority, setPriority] = useState('HIGH');
  const [description, setDescription] = useState('');
  const [error, setError] = useState('');
  const [duplicateWarning, setDuplicateWarning] = useState('');

  const TEAM_APPROVERS = [
    { name: 'Arjun Singh', role: 'Operations Manager', dept: 'Operations' },
    { name: 'Neha Kapoor', role: 'Documentation Manager', dept: 'Documentation' },
    { name: 'Rohit Mehta', role: 'Sales Director', dept: 'Sales' },
    { name: 'Vikram Kumar', role: 'Finance Head', dept: 'Finance' },
    { name: 'Pooja Shah', role: 'Operations Lead', dept: 'Operations' },
  ];

  useEffect(() => {
    if (isOpen) {
      setCategory(initialCategory);
      setType(initialType);
      setRelatedRef(initialRelatedRef);
      setCustomerName(initialCustomerName);
    }
  }, [isOpen, initialCategory, initialType, initialRelatedRef, initialCustomerName]);

  if (!isOpen) return null;

  const handleRelatedRefChange = (val) => {
    setRelatedRef(val);
    // Duplicate check against pending items
    if (val.trim()) {
      const match = existingApprovals.find(
        (a) => a.status === 'Pending' && a.relatedRef && a.relatedRef.toLowerCase() === val.trim().toLowerCase()
      );
      if (match) {
        setDuplicateWarning(`Warning: Approval request (${match.id}) is already pending for "${val}".`);
      } else {
        setDuplicateWarning('');
      }
    } else {
      setDuplicateWarning('');
    }
  };

  const handleCategoryChange = (e) => {
    const val = e.target.value;
    setCategory(val);
    if (val === 'DOCUMENTS') {
      setType('Document Approval');
      setDepartment('Documentation');
      setAssignedTo('Neha Kapoor (Documentation Manager)');
    } else if (val === 'COMMERCIAL') {
      setType('Commercial Approval');
      setDepartment('Sales');
      setAssignedTo('Rohit Mehta (Sales Director)');
    } else if (val === 'OPERATIONS') {
      setType('Operations Approval');
      setDepartment('Operations');
      setAssignedTo('Arjun Singh (Operations Manager)');
    } else if (val === 'FINANCE') {
      setType('Finance Approval');
      setDepartment('Finance');
      setAssignedTo('Vikram Kumar (Finance Head)');
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!title.trim()) {
      setError('Please provide a title for the approval request.');
      return;
    }
    if (!relatedRef.trim()) {
      setError('Please specify a related shipment, document, customer, or invoice reference.');
      return;
    }

    const initials = requesterName
      .split(' ')
      .filter(Boolean)
      .map((p) => p[0])
      .join('')
      .toUpperCase() || 'VK';

    const newApprovalInput = {
      title: title.trim(),
      category,
      type,
      priority,
      relatedRef: relatedRef.trim(),
      customerName: customerName.trim() || 'Associated Customer',
      requesterName,
      department,
      assignedTo,
      dueDate,
      description: description.trim(),
      shipmentId: initialShipmentId,
      documentId: initialDocumentId,
      customerId: initialCustomerId,
      avatar: initials,
    };

    onSubmit(newApprovalInput);
    onClose();
  };

  return (
    <div className="approval-modal-overlay" onClick={onClose}>
      <div className="approval-modal-container" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 580 }}>
        <div className="approval-modal-header">
          <div>
            <h3 style={{ margin: 0, fontSize: '1.1rem', fontWeight: 800 }}>New Approval Request</h3>
            <p style={{ margin: '2px 0 0 0', fontSize: '0.78rem', color: '#64748B' }}>
              Submit a business approval request for document release, credit limits, or rate exceptions.
            </p>
          </div>
          <button type="button" className="modal-close-btn" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {error && (
          <div className="approval-modal-error" style={{ margin: '14px 24px 0 24px' }}>
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        {duplicateWarning && (
          <div style={{ background: '#FFFBEB', border: '1px solid #FDE68A', color: '#D97706', padding: '10px 14px', margin: '14px 24px 0 24px', borderRadius: 8, fontSize: '0.8rem', fontWeight: 650, display: 'flex', alignItems: 'center', gap: 8 }}>
            <ShieldAlert size={16} />
            <span>{duplicateWarning}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="approval-modal-form">
          <div className="form-group">
            <label className="form-label">Request Title *</label>
            <input
              type="text"
              className="form-input"
              placeholder="e.g. Rate Exception Sign-off - SHP-250812"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
            />
          </div>

          <div className="form-row-2">
            <div className="form-group">
              <label className="form-label">Category *</label>
              <select className="form-select" value={category} onChange={handleCategoryChange}>
                <option value="DOCUMENTS">Documents</option>
                <option value="COMMERCIAL">Commercial</option>
                <option value="OPERATIONS">Operations</option>
                <option value="FINANCE">Finance</option>
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Approval Type</label>
              <input type="text" className="form-input" value={type} readOnly style={{ background: '#F8FAFC' }} />
            </div>
          </div>

          <div className="form-row-2">
            <div className="form-group">
              <label className="form-label">Related Context Reference *</label>
              <input
                type="text"
                className="form-input"
                placeholder="e.g. Shipment: SHP-250812 or Customer: CUST-007"
                value={relatedRef}
                onChange={(e) => handleRelatedRefChange(e.target.value)}
                required
              />
            </div>

            <div className="form-group">
              <label className="form-label">Customer / Company</label>
              <input
                type="text"
                className="form-input"
                placeholder="e.g. ABC Exports Pvt Ltd"
                value={customerName}
                onChange={(e) => setCustomerName(e.target.value)}
              />
            </div>
          </div>

          <div className="form-row-2">
            <div className="form-group">
              <label className="form-label">Assign Approver *</label>
              <select className="form-select" value={assignedTo} onChange={(e) => setAssignedTo(e.target.value)}>
                {TEAM_APPROVERS.map((app) => (
                  <option key={app.name} value={`${app.name} (${app.role})`}>
                    {app.name} — {app.role}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Priority</label>
              <select className="form-select" value={priority} onChange={(e) => setPriority(e.target.value)}>
                <option value="URGENT">🔴 Urgent</option>
                <option value="HIGH">🟠 High</option>
                <option value="MEDIUM">🟡 Medium</option>
                <option value="LOW">🔵 Low</option>
              </select>
            </div>
          </div>

          <div className="form-row-2">
            <div className="form-group">
              <label className="form-label">Requester</label>
              <input type="text" className="form-input" value={`${requesterName} (${department})`} readOnly style={{ background: '#F8FAFC' }} />
            </div>

            <div className="form-group">
              <label className="form-label">Due Date *</label>
              <input
                type="date"
                className="form-input"
                value={dueDate}
                min={new Date().toISOString().split('T')[0]}
                onChange={(e) => setDueDate(e.target.value)}
                required
              />
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Description / Reason for Approval</label>
            <textarea
              className="form-textarea"
              rows={3}
              placeholder="Provide background details, operational context, or pricing exception notes for the approver..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="approval-modal-footer">
            <button type="button" className="btn-cancel" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-submit">
              <Plus size={16} /> Submit Approval Request
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
