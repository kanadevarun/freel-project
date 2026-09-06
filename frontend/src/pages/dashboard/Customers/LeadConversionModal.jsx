import React, { useState, useEffect } from 'react';
import { X, AlertTriangle, CheckCircle, Building, UserCheck, PlusCircle } from 'lucide-react';
import customerService from '../../../services/customerService';
import './LeadConversionModal.css';

export default function LeadConversionModal({ isOpen, onClose, lead, onSuccess }) {
  const [checking, setChecking] = useState(true);
  const [duplicates, setDuplicates] = useState([]);
  const [selectedOption, setSelectedOption] = useState('NEW'); // 'NEW' or customerId
  const [customerType, setCustomerType] = useState('SHIPPER');
  const [taxId, setTaxId] = useState('');
  const [converting, setConverting] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (isOpen && lead) {
      runDuplicateCheck();
    }
  }, [isOpen, lead]);

  const runDuplicateCheck = async () => {
    try {
      setChecking(true);
      setError(null);
      const matches = await customerService.checkDuplicate({
        name: lead.company_name,
        email: lead.email,
        phone: lead.phone,
        lead_id: lead.id,
      });
      const list = Array.isArray(matches) ? matches : [];
      setDuplicates(list);
      if (list.length > 0) {
        setSelectedOption(list[0].customer_id);
      } else {
        setSelectedOption('NEW');
      }
    } catch (err) {
      console.error('Duplicate check error:', err);
      setDuplicates([]);
      setSelectedOption('NEW');
    } finally {
      setChecking(false);
    }
  };

  if (!isOpen || !lead) return null;

  const handleConvert = async () => {
    try {
      setConverting(true);
      setError(null);

      const payload = {
        lead_id: lead.id,
        customer_type: customerType,
        tax_id: taxId ? taxId : undefined,
        force_create_new: selectedOption === 'NEW',
        link_to_existing_customer_id: selectedOption !== 'NEW' ? Number(selectedOption) : undefined,
      };

      const res = await customerService.convertLeadToCustomer(payload);
      onSuccess(res);
      onClose();
    } catch (err) {
      console.error('Lead conversion failed:', err);
      setError(err?.response?.data?.message || 'Failed to convert lead. Please try again.');
    } finally {
      setConverting(false);
    }
  };

  return (
    <div className="conv-modal-backdrop" onClick={onClose}>
      <div className="conv-modal-container" onClick={(e) => e.stopPropagation()}>
        <div className="conv-modal-header">
          <div>
            <h2>Convert Sales Lead to Customer</h2>
            <p>
              Establishing commercial relationship for <strong>{lead.company_name}</strong>
            </p>
          </div>
          <button className="conv-modal-close" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {error && <div className="conv-modal-error">⚠️ {error}</div>}

        <div className="conv-modal-body">
          {/* Lead Summary Header Card */}
          <div className="conv-lead-summary">
            <div className="lead-sum-badge">LEAD #{lead.id}</div>
            <div className="lead-sum-details">
              <strong>{lead.company_name}</strong>
              <span>
                Contact: {lead.contact_name || 'N/A'} • {lead.email || 'No Email'} • {lead.phone || 'No Phone'}
              </span>
            </div>
          </div>

          {checking ? (
            <div className="conv-checking">
              <div className="conv-spinner" />
              <span>Scanning Customer Directory for duplicate accounts...</span>
            </div>
          ) : (
            <>
              {duplicates.length > 0 ? (
                <div className="conv-duplicate-warning">
                  <div className="warning-title">
                    <AlertTriangle size={18} />
                    <span>Potential Duplicate Customer Accounts Detected ({duplicates.length})</span>
                  </div>
                  <p className="warning-desc">
                    We found existing customer accounts matching company name or contact details. Linking prevents duplicate records while updating history.
                  </p>

                  <div className="conv-options-list">
                    {duplicates.map((dup) => (
                      <label
                        key={dup.customer_id}
                        className={`conv-option-card ${selectedOption === dup.customer_id ? 'selected' : ''}`}
                      >
                        <input
                          type="radio"
                          name="conversion_option"
                          value={dup.customer_id}
                          checked={selectedOption === dup.customer_id}
                          onChange={() => setSelectedOption(dup.customer_id)}
                        />
                        <div className="option-content">
                          <div className="option-header">
                            <span className="cust-code">{dup.customer_code}</span>
                            <span className="cust-name">{dup.customer_name}</span>
                            <span className="confidence-chip">{dup.confidence_score}% Match</span>
                          </div>
                          <div className="option-reason">
                            Matched on: <strong>{dup.match_reason}</strong>
                          </div>
                        </div>
                      </label>
                    ))}

                    <label className={`conv-option-card ${selectedOption === 'NEW' ? 'selected' : ''}`}>
                      <input
                        type="radio"
                        name="conversion_option"
                        value="NEW"
                        checked={selectedOption === 'NEW'}
                        onChange={() => setSelectedOption('NEW')}
                      />
                      <div className="option-content">
                        <div className="option-header">
                          <PlusCircle size={16} color="#2563eb" />
                          <strong style={{ color: '#0f172a' }}>Create Brand New Customer Account</strong>
                        </div>
                        <div className="option-reason">
                          Disregard duplicates and create a separate new commercial customer account entity.
                        </div>
                      </div>
                    </label>
                  </div>
                </div>
              ) : (
                <div className="conv-no-duplicate">
                  <CheckCircle size={20} color="#16a34a" />
                  <div>
                    <strong>No Duplicate Accounts Found</strong>
                    <p>This lead can be converted into a new Customer Account safely.</p>
                  </div>
                </div>
              )}

              {/* Conversion Parameters */}
              <div className="conv-params-grid">
                <div className="conv-field">
                  <label>Customer Account Type</label>
                  <select value={customerType} onChange={(e) => setCustomerType(e.target.value)}>
                    <option value="SHIPPER">Shipper</option>
                    <option value="IMPORTER">Importer</option>
                    <option value="EXPORTER">Exporter</option>
                    <option value="TRADER">Trader</option>
                    <option value="MANUFACTURER">Manufacturer</option>
                    <option value="LOGISTICS_CO">Logistics Co.</option>
                  </select>
                </div>

                <div className="conv-field">
                  <label>GSTIN / Tax ID (Optional)</label>
                  <input
                    type="text"
                    placeholder="e.g. 27AAAC01234B1Z5"
                    value={taxId}
                    onChange={(e) => setTaxId(e.target.value)}
                  />
                </div>
              </div>
            </>
          )}
        </div>

        <div className="conv-modal-footer">
          <button className="btn-cancel" onClick={onClose} disabled={converting}>
            Cancel
          </button>
          <button className="btn-submit" onClick={handleConvert} disabled={checking || converting}>
            {converting
              ? 'Converting...'
              : selectedOption === 'NEW'
              ? 'Create & Convert to Customer'
              : 'Link Lead to Customer Account'}
          </button>
        </div>
      </div>
    </div>
  );
}
