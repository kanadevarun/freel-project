import React, { useState, useEffect } from 'react';
import { X, Building2, User, MapPin, CreditCard, ShieldCheck } from 'lucide-react';
import customerService from '../../../services/customerService';
import './CustomerModal.css';

const CUSTOMER_TYPES = [
  { value: 'SHIPPER', label: 'Shipper' },
  { value: 'IMPORTER', label: 'Importer' },
  { value: 'EXPORTER', label: 'Exporter' },
  { value: 'TRADER', label: 'Trader' },
  { value: 'MANUFACTURER', label: 'Manufacturer' },
  { value: 'LOGISTICS_CO', label: 'Logistics Co.' },
  { value: 'CONSIGNEE', label: 'Consignee' },
  { value: 'BUYER', label: 'Buyer' },
  { value: 'SELLER', label: 'Seller' },
  { value: 'OTHER', label: 'Other' },
];

export default function CustomerModal({ isOpen, onClose, onSuccess, initialData = null }) {
  const [formData, setFormData] = useState({
    name: '',
    trading_name: '',
    customer_type: 'SHIPPER',
    industry: '',
    domain: '',
    website: '',
    tax_id: '',
    pan_number: '',
    eori_number: '',
    currency: 'USD',
    payment_terms: 'NET30',
    credit_limit: 0,
    country: '',
    city: '',
    contact_name: '',
    contact_email: '',
    contact_phone: '',
    notes: '',
  });

  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (initialData) {
      setFormData({
        name: initialData.name || '',
        trading_name: initialData.trading_name || '',
        customer_type: initialData.customer_type || 'SHIPPER',
        industry: initialData.industry || '',
        domain: initialData.domain || '',
        website: initialData.website || '',
        tax_id: initialData.tax_id || '',
        pan_number: initialData.pan_number || '',
        eori_number: initialData.eori_number || '',
        currency: initialData.currency || 'USD',
        payment_terms: initialData.payment_terms || 'NET30',
        credit_limit: initialData.credit_limit || 0,
        country: initialData.country || '',
        city: initialData.city || '',
        contact_name: initialData.contact_name || '',
        contact_email: initialData.contact_email || '',
        contact_phone: initialData.contact_phone || '',
        notes: initialData.notes || '',
      });
    } else {
      setFormData({
        name: '',
        trading_name: '',
        customer_type: 'SHIPPER',
        industry: '',
        domain: '',
        website: '',
        tax_id: '',
        pan_number: '',
        eori_number: '',
        currency: 'USD',
        payment_terms: 'NET30',
        credit_limit: 0,
        country: '',
        city: '',
        contact_name: '',
        contact_email: '',
        contact_phone: '',
        notes: '',
      });
    }
    setError(null);
  }, [initialData, isOpen]);

  if (!isOpen) return null;

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.name.trim()) {
      setError('Company Name is required.');
      return;
    }

    try {
      setSaving(true);
      setError(null);
      if (initialData) {
        await customerService.updateCustomer(initialData.id, formData);
      } else {
        await customerService.createCustomer(formData);
      }
      onSuccess();
      onClose();
    } catch (err) {
      console.error('Failed to save customer:', err);
      setError(err?.response?.data?.message || 'Failed to save customer. Please check fields.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="cust-modal-backdrop" onClick={onClose}>
      <div className="cust-modal-container" onClick={(e) => e.stopPropagation()}>
        <div className="cust-modal-header">
          <div>
            <h2>{initialData ? 'Edit Customer Account' : 'New Customer Account'}</h2>
            <p>Establish commercial counterparty profile, tax identifiers, and credit parameters</p>
          </div>
          <button className="cust-modal-close" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {error && <div className="cust-modal-error">⚠️ {error}</div>}

        <form onSubmit={handleSubmit} className="cust-modal-body">
          {/* Section 1: Company Profile */}
          <div className="cust-form-section">
            <div className="cust-section-title">
              <Building2 size={16} /> Company Profile
            </div>
            <div className="cust-form-grid">
              <div className="cust-form-field full-width">
                <label>Company Legal Name *</label>
                <input
                  type="text"
                  name="name"
                  value={formData.name}
                  onChange={handleChange}
                  placeholder="e.g. Oceanic Exports Pvt. Ltd."
                  required
                />
              </div>

              <div className="cust-form-field">
                <label>Trading Name (DBA)</label>
                <input
                  type="text"
                  name="trading_name"
                  value={formData.trading_name}
                  onChange={handleChange}
                  placeholder="e.g. Oceanic Exports"
                />
              </div>

              <div className="cust-form-field">
                <label>Account Type / Role</label>
                <select name="customer_type" value={formData.customer_type} onChange={handleChange}>
                  {CUSTOMER_TYPES.map((t) => (
                    <option key={t.value} value={t.value}>
                      {t.label}
                    </option>
                  ))}
                </select>
              </div>

              <div className="cust-form-field">
                <label>Industry Vertical</label>
                <input
                  type="text"
                  name="industry"
                  value={formData.industry}
                  onChange={handleChange}
                  placeholder="e.g. Electronics, Textiles"
                />
              </div>

              <div className="cust-form-field">
                <label>Website URL</label>
                <input
                  type="text"
                  name="website"
                  value={formData.website}
                  onChange={handleChange}
                  placeholder="e.g. www.oceanicexports.com"
                />
              </div>
            </div>
          </div>

          {/* Section 2: Location & Primary Contact */}
          <div className="cust-form-section">
            <div className="cust-section-title">
              <MapPin size={16} /> Location & Primary Contact
            </div>
            <div className="cust-form-grid">
              <div className="cust-form-field">
                <label>City</label>
                <input
                  type="text"
                  name="city"
                  value={formData.city}
                  onChange={handleChange}
                  placeholder="e.g. Mumbai"
                />
              </div>

              <div className="cust-form-field">
                <label>Country</label>
                <input
                  type="text"
                  name="country"
                  value={formData.country}
                  onChange={handleChange}
                  placeholder="e.g. India, UAE, Sweden"
                />
              </div>

              <div className="cust-form-field">
                <label>Primary Contact Name</label>
                <input
                  type="text"
                  name="contact_name"
                  value={formData.contact_name}
                  onChange={handleChange}
                  placeholder="e.g. Rajesh Mehta"
                />
              </div>

              <div className="cust-form-field">
                <label>Primary Contact Email</label>
                <input
                  type="email"
                  name="contact_email"
                  value={formData.contact_email}
                  onChange={handleChange}
                  placeholder="e.g. rajesh@oceanicexports.com"
                />
              </div>

              <div className="cust-form-field">
                <label>Primary Contact Phone</label>
                <input
                  type="text"
                  name="contact_phone"
                  value={formData.contact_phone}
                  onChange={handleChange}
                  placeholder="e.g. +91 98765 43210"
                />
              </div>
            </div>
          </div>

          {/* Section 3: Commercial & Tax Parameters */}
          <div className="cust-form-section">
            <div className="cust-section-title">
              <ShieldCheck size={16} /> Commercial & Tax Identifiers
            </div>
            <div className="cust-form-grid">
              <div className="cust-form-field">
                <label>GSTIN / Tax ID</label>
                <input
                  type="text"
                  name="tax_id"
                  value={formData.tax_id}
                  onChange={handleChange}
                  placeholder="e.g. 27AAAC01234B1Z5"
                />
              </div>

              <div className="cust-form-field">
                <label>PAN Number</label>
                <input
                  type="text"
                  name="pan_number"
                  value={formData.pan_number}
                  onChange={handleChange}
                  placeholder="e.g. AAAC01234B"
                />
              </div>

              <div className="cust-form-field">
                <label>EORI Number (EU)</label>
                <input
                  type="text"
                  name="eori_number"
                  value={formData.eori_number}
                  onChange={handleChange}
                  placeholder="e.g. GB123456789000"
                />
              </div>

              <div className="cust-form-field">
                <label>Billing Currency</label>
                <select name="currency" value={formData.currency} onChange={handleChange}>
                  <option value="USD">USD ($)</option>
                  <option value="EUR">EUR (€)</option>
                  <option value="INR">INR (₹)</option>
                  <option value="GBP">GBP (£)</option>
                  <option value="AED">AED (AED)</option>
                  <option value="SGD">SGD (S$)</option>
                </select>
              </div>

              <div className="cust-form-field">
                <label>Payment Terms</label>
                <select name="payment_terms" value={formData.payment_terms} onChange={handleChange}>
                  <option value="NET15">NET 15 Days</option>
                  <option value="NET30">NET 30 Days</option>
                  <option value="NET45">NET 45 Days</option>
                  <option value="NET60">NET 60 Days</option>
                  <option value="CIA">Cash In Advance (CIA)</option>
                  <option value="CAD">Cash Against Documents (CAD)</option>
                </select>
              </div>

              <div className="cust-form-field">
                <label>Credit Limit ($)</label>
                <input
                  type="number"
                  name="credit_limit"
                  value={formData.credit_limit}
                  onChange={handleChange}
                  placeholder="50000"
                />
              </div>
            </div>
          </div>

          <div className="cust-modal-footer">
            <button type="button" className="btn-cancel" onClick={onClose} disabled={saving}>
              Cancel
            </button>
            <button type="submit" className="btn-submit" disabled={saving}>
              {saving ? 'Saving Account...' : initialData ? 'Save Changes' : 'Create Customer Account'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
