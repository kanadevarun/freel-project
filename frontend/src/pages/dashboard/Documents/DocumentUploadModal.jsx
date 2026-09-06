import React, { useState, useEffect } from 'react';
import { X, Upload, FileText, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';
import api from '../../../services/api';

export default function DocumentUploadModal({
  isOpen,
  onClose,
  onUploadSuccess,
  initialDocType = 'HBL',
  initialCustomerId = '',
  initialShipmentId = '',
  initialLeadId = '',
  initialBookingId = ''
}) {
  const [selectedFile, setSelectedFile] = useState(null);
  const [docType, setDocType] = useState(initialDocType);
  const [shipmentId, setShipmentId] = useState(initialShipmentId);
  const [customerId, setCustomerId] = useState(initialCustomerId);
  const [leadId, setLeadId] = useState(initialLeadId);
  const [bookingId, setBookingId] = useState(initialBookingId);

  const [shipmentsOptions, setShipmentsOptions] = useState([]);
  const [customersOptions, setCustomersOptions] = useState([]);
  const [leadsOptions, setLeadsOptions] = useState([]);
  const [bookingsOptions, setBookingsOptions] = useState([]);

  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState('');
  const [isDragOver, setIsDragOver] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setDocType(initialDocType || 'HBL');
      setCustomerId(initialCustomerId || '');
      setShipmentId(initialShipmentId || '');
      setLeadId(initialLeadId || '');
      setBookingId(initialBookingId || '');
      fetchRelationalOptions();
    }
  }, [isOpen, initialDocType, initialCustomerId, initialShipmentId, initialLeadId, initialBookingId]);

  const fetchRelationalOptions = async () => {
    try {
      // Fetch Shipments
      const shipmentsRes = await api.get('/api/v1/shipments').catch(() => []);
      const sList = Array.isArray(shipmentsRes) ? shipmentsRes : (shipmentsRes?.data || shipmentsRes?.shipments || []);
      setShipmentsOptions(sList);

      // Fetch Customers
      const customersRes = await api.get('/api/v1/customers').catch(() => []);
      const cList = Array.isArray(customersRes) ? customersRes : (customersRes?.data?.customers || customersRes?.data || []);
      setCustomersOptions(cList);

      // Fetch Leads
      const leadsRes = await api.get('/api/v1/leads').catch(() => []);
      const lList = Array.isArray(leadsRes) ? leadsRes : (leadsRes?.data?.leads || leadsRes?.data || []);
      setLeadsOptions(lList);

      // Fetch Bookings/RFQs
      const bookingsRes = await api.get('/api/v1/rfqs').catch(() => []);
      const bList = Array.isArray(bookingsRes) ? bookingsRes : (bookingsRes?.data || []);
      setBookingsOptions(bList);
    } catch (err) {
      console.warn('Could not load relationship dropdown options:', err);
    }
  };

  if (!isOpen) return null;

  const handleFileChange = (e) => {
    if (e.target.files && e.target.files[0]) {
      setSelectedFile(e.target.files[0]);
      setError('');
    }
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragOver(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      setSelectedFile(e.dataTransfer.files[0]);
      setError('');
    }
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = () => {
    setIsDragOver(false);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!selectedFile) {
      setError('Please select a file to upload.');
      return;
    }

    try {
      setIsUploading(true);
      setError('');

      const formData = new FormData();
      formData.append('file', selectedFile);
      formData.append('doc_type', docType);
      if (shipmentId) formData.append('shipment_id', shipmentId);
      if (customerId) formData.append('customer_id', customerId);
      if (leadId) formData.append('lead_id', leadId);
      if (bookingId) formData.append('booking_id', bookingId);

      const res = await api.post('/api/v1/documents/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      const uploadedDoc = res?.data || res;
      if (onUploadSuccess) {
        onUploadSuccess(uploadedDoc);
      }

      // Reset form
      setSelectedFile(null);
      setDocType('HBL');
      setShipmentId('');
      setCustomerId('');
      setLeadId('');
      setBookingId('');
      onClose();
    } catch (err) {
      console.error('Document upload failed:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed to upload document. Please check your connection and try again.');
    } finally {
      setIsUploading(false);
    }
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  return (
    <div className="doc-upload-modal-overlay">
      <div className="doc-upload-modal-container">
        {/* Header */}
        <div className="doc-upload-modal-header">
          <div>
            <h3>
              {docType === 'HBL' || docType === 'MBL' || initialDocType === 'HBL' || initialDocType === 'MBL'
                ? 'Upload Bill of Lading'
                : docType === 'COMMERCIAL_INVOICE' || docType === 'PACKING_LIST' || initialDocType === 'COMMERCIAL_INVOICE' || initialDocType === 'PACKING_LIST'
                  ? 'Upload Invoice / Packing List'
                  : 'Upload Document'}
            </h3>
            <p>Upload shipment documents, bills of lading, or invoices to attach to records.</p>
          </div>
          <button className="doc-upload-modal-close" onClick={onClose} disabled={isUploading}>
            <X size={18} />
          </button>
        </div>

        {error && (
          <div className="doc-upload-modal-error">
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="doc-upload-modal-body">
          {/* File Upload Drag & Drop Dropzone */}
          <div
            className={`doc-upload-dropzone ${isDragOver ? 'drag-over' : ''} ${selectedFile ? 'has-file' : ''}`}
            onDrop={handleDrop}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
          >
            <input
              type="file"
              id="documentFileInput"
              onChange={handleFileChange}
              style={{ display: 'none' }}
              accept=".pdf,.png,.jpg,.jpeg,.doc,.docx,.xls,.xlsx"
            />

            {!selectedFile ? (
              <label htmlFor="documentFileInput" className="dropzone-label">
                <div className="dropzone-icon-box">
                  <Upload size={24} />
                </div>
                <span className="dropzone-title">Click to select or drag and drop file</span>
                <span className="dropzone-subtitle">PDF, PNG, JPG, DOCX, XLSX (Max 25MB)</span>
              </label>
            ) : (
              <div className="dropzone-selected-file">
                <FileText size={28} className="file-icon" />
                <div className="file-info">
                  <span className="file-name">{selectedFile.name}</span>
                  <span className="file-size">{formatFileSize(selectedFile.size)}</span>
                </div>
                <label htmlFor="documentFileInput" className="btn-change-file">
                  Change
                </label>
              </div>
            )}
          </div>

          {/* Document Type Dropdown */}
          <div className="doc-upload-form-group">
            <label>Document Category / Type <span className="req">*</span></label>
            <select value={docType} onChange={(e) => setDocType(e.target.value)} required>
              <option value="HBL">HBL — House Bill of Lading</option>
              <option value="MBL">MBL — Master Bill of Lading</option>
              <option value="COMMERCIAL_INVOICE">Commercial Invoice</option>
              <option value="PACKING_LIST">Packing List</option>
              <option value="CONTRACT">Carrier / Service Contract</option>
              <option value="QUOTATION">Quotation / Rate Sheet</option>
              <option value="OTHER">Other / General Document</option>
            </select>
          </div>

          {/* Relational Linkage Section */}
          <div className="doc-upload-relational-section">
            <h4>Link to Business Context (Optional)</h4>

            <div className="doc-upload-form-grid">
              {/* Shipment Link */}
              <div className="doc-upload-form-group">
                <label>Shipment</label>
                <select value={shipmentId} onChange={(e) => setShipmentId(e.target.value)}>
                  <option value="">-- None (Standalone Document) --</option>
                  {shipmentsOptions.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.shipment_number || `SH-${s.id}`} {s.origin_port && s.destination_port ? `(${s.origin_port} → ${s.destination_port})` : ''}
                    </option>
                  ))}
                </select>
              </div>

              {/* Customer Link */}
              <div className="doc-upload-form-group">
                <label>Customer</label>
                <select value={customerId} onChange={(e) => setCustomerId(e.target.value)}>
                  <option value="">-- None --</option>
                  {customersOptions.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.company_name || c.name || `Customer #${c.id}`}
                    </option>
                  ))}
                </select>
              </div>

              {/* Lead Link */}
              <div className="doc-upload-form-group">
                <label>Sales Lead</label>
                <select value={leadId} onChange={(e) => setLeadId(e.target.value)}>
                  <option value="">-- None --</option>
                  {leadsOptions.map((l) => (
                    <option key={l.id} value={l.id}>
                      {l.company_name || l.contact_name || `Lead #${l.id}`}
                    </option>
                  ))}
                </select>
              </div>

              {/* Booking / RFQ Link */}
              <div className="doc-upload-form-group">
                <label>Booking / RFQ</label>
                <select value={bookingId} onChange={(e) => setBookingId(e.target.value)}>
                  <option value="">-- None --</option>
                  {bookingsOptions.map((b) => (
                    <option key={b.id} value={b.id}>
                      {b.rfq_number || b.title || `RFQ / Booking #${b.id}`}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* Modal Footer Actions */}
          <div className="doc-upload-modal-footer">
            <button type="button" className="btn-cancel" onClick={onClose} disabled={isUploading}>
              Cancel
            </button>
            <button type="submit" className="btn-submit" disabled={isUploading || !selectedFile}>
              {isUploading ? (
                <>
                  <Loader2 size={16} className="spin-icon" /> Uploading Document...
                </>
              ) : (
                <>
                  <Upload size={16} /> Save Document
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
