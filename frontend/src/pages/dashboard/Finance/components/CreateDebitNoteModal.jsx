import React, { useState, useEffect, useCallback } from 'react';
import { X, Plus, Trash2, AlertCircle, FileWarning } from 'lucide-react';
import api from '../../../../services/api';
import './CreateDebitNoteModal.css';

const CURRENCIES = ['USD', 'EUR', 'GBP', 'INR', 'AED', 'SGD'];
const SERVICE_CATEGORIES = [
  'Ocean Freight', 'Air Freight', 'Road Transport', 'Customs Brokerage',
  'Port Charges', 'Demurrage', 'Detention', 'Handling', 'Documentation',
  'Insurance', 'Fuel Surcharge', 'THC', 'Other',
];

const emptyItem = () => ({ description: '', service_category: '', quantity: 1, unit_price: '' });

export default function CreateDebitNoteModal({ isOpen, onClose, showToast, onSuccess }) {
  const [form, setForm] = useState({
    customer_id: '',
    customer_name: '',
    shipment_number: '',
    invoice_number: '',
    reason: '',
    currency: 'USD',
    tax_amount: '',
    issue_date: '',
    due_date: '',
    notes: '',
    issue_immediately: false,
  });
  const [lineItems, setLineItems] = useState([emptyItem()]);
  const [customers, setCustomers] = useState([]);
  const [errors, setErrors] = useState({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    api.get('/api/v1/companies').then((res) => {
      const list = res?.data || res || [];
      setCustomers(Array.isArray(list) ? list : []);
    }).catch(() => {});
  }, [isOpen]);

  const resetForm = () => {
    setForm({
      customer_id: '', customer_name: '', shipment_number: '',
      invoice_number: '', reason: '', currency: 'USD',
      tax_amount: '', issue_date: '', due_date: '',
      notes: '', issue_immediately: false,
    });
    setLineItems([emptyItem()]);
    setErrors({});
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const subtotal = lineItems.reduce((sum, item) => {
    const qty = parseFloat(item.quantity) || 0;
    const price = parseFloat(item.unit_price) || 0;
    return sum + qty * price;
  }, 0);
  const tax = parseFloat(form.tax_amount) || 0;
  const total = subtotal + tax;

  const validate = () => {
    const errs = {};
    if (!form.customer_id && !form.customer_name.trim()) errs.customer = 'Customer is required';
    if (!form.reason.trim()) errs.reason = 'Reason is required';
    if (lineItems.length === 0) errs.items = 'At least one line item is required';
    lineItems.forEach((item, idx) => {
      if (!item.description.trim()) errs[`item_${idx}_desc`] = 'Description required';
      if (!item.unit_price || parseFloat(item.unit_price) <= 0) errs[`item_${idx}_price`] = 'Unit price required';
    });
    return errs;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const errs = validate();
    if (Object.keys(errs).length > 0) { setErrors(errs); return; }

    setSubmitting(true);
    try {
      const payload = {
        customer_id: parseInt(form.customer_id) || 0,
        customer_name: form.customer_name,
        shipment_number: form.shipment_number,
        invoice_number: form.invoice_number,
        reason: form.reason,
        currency: form.currency,
        tax_amount: parseFloat(form.tax_amount) || 0,
        issue_date: form.issue_date,
        due_date: form.due_date,
        notes: form.notes,
        issue_immediately: form.issue_immediately,
        line_items: lineItems.map((item) => ({
          description: item.description,
          service_category: item.service_category,
          quantity: parseFloat(item.quantity) || 1,
          unit_price: parseFloat(item.unit_price) || 0,
        })),
      };
      const res = await api.post('/api/v1/debit-notes', payload);
      const dn = res?.data || res;
      showToast(`Debit Note ${dn?.debit_note_number || ''} created successfully!`, 'success');
      onSuccess(dn);
      handleClose();
    } catch (err) {
      showToast(err?.message || 'Failed to create debit note', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const updateItem = (idx, field, value) => {
    setLineItems((prev) => prev.map((item, i) => i === idx ? { ...item, [field]: value } : item));
  };

  const addItem = () => setLineItems((prev) => [...prev, emptyItem()]);
  const removeItem = (idx) => setLineItems((prev) => prev.filter((_, i) => i !== idx));

  if (!isOpen) return null;

  return (
    <div className="dn-modal-overlay" onClick={handleClose}>
      <div className="dn-modal" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="dn-modal-header">
          <div className="dn-modal-title-block">
            <div className="dn-modal-icon">
              <FileWarning size={20} />
            </div>
            <div>
              <h2 className="dn-modal-title">New Debit Note</h2>
              <p className="dn-modal-subtitle">Issue additional charges to a customer</p>
            </div>
          </div>
          <button className="dn-modal-close" onClick={handleClose} aria-label="Close">
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="dn-modal-body">
          {/* Customer & Reference */}
          <div className="dn-form-section">
            <h3 className="dn-section-label">Customer & Reference</h3>
            <div className="dn-form-grid-2">
              <div className="dn-form-field">
                <label>Customer *</label>
                {customers.length > 0 ? (
                  <select
                    value={form.customer_id}
                    onChange={(e) => {
                      const selected = customers.find(c => String(c.id) === e.target.value);
                      setForm(f => ({
                        ...f,
                        customer_id: e.target.value,
                        customer_name: selected?.name || f.customer_name,
                      }));
                    }}
                    className={errors.customer ? 'error' : ''}
                  >
                    <option value="">Select customer...</option>
                    {customers.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                  </select>
                ) : (
                  <input
                    type="text"
                    placeholder="Customer name"
                    value={form.customer_name}
                    onChange={(e) => setForm(f => ({ ...f, customer_name: e.target.value }))}
                    className={errors.customer ? 'error' : ''}
                  />
                )}
                {errors.customer && <span className="dn-field-error"><AlertCircle size={12} />{errors.customer}</span>}
              </div>

              <div className="dn-form-field">
                <label>Currency</label>
                <select value={form.currency} onChange={(e) => setForm(f => ({ ...f, currency: e.target.value }))}>
                  {CURRENCIES.map(c => <option key={c} value={c}>{c}</option>)}
                </select>
              </div>

              <div className="dn-form-field">
                <label>Linked Invoice # <span className="dn-optional">(optional)</span></label>
                <input
                  type="text"
                  placeholder="e.g. INV-20260901-0012"
                  value={form.invoice_number}
                  onChange={(e) => setForm(f => ({ ...f, invoice_number: e.target.value }))}
                />
              </div>

              <div className="dn-form-field">
                <label>Shipment # <span className="dn-optional">(optional)</span></label>
                <input
                  type="text"
                  placeholder="e.g. SHP-001"
                  value={form.shipment_number}
                  onChange={(e) => setForm(f => ({ ...f, shipment_number: e.target.value }))}
                />
              </div>
            </div>

            <div className="dn-form-field dn-form-full">
              <label>Reason / Description *</label>
              <textarea
                rows={3}
                placeholder="Describe the reason for this debit note (e.g. demurrage charges, weight correction, additional port surcharge…)"
                value={form.reason}
                onChange={(e) => setForm(f => ({ ...f, reason: e.target.value }))}
                className={errors.reason ? 'error' : ''}
              />
              {errors.reason && <span className="dn-field-error"><AlertCircle size={12} />{errors.reason}</span>}
            </div>
          </div>

          {/* Line Items */}
          <div className="dn-form-section">
            <div className="dn-section-header-row">
              <h3 className="dn-section-label">Line Items</h3>
              <button type="button" className="dn-add-item-btn" onClick={addItem}>
                <Plus size={14} /> Add Line
              </button>
            </div>
            {errors.items && <span className="dn-field-error"><AlertCircle size={12} />{errors.items}</span>}

            <div className="dn-items-table">
              <div className="dn-items-header">
                <span>Description</span>
                <span>Category</span>
                <span>Qty</span>
                <span>Unit Price</span>
                <span>Total</span>
                <span></span>
              </div>
              {lineItems.map((item, idx) => {
                const lineTotal = (parseFloat(item.quantity) || 0) * (parseFloat(item.unit_price) || 0);
                return (
                  <div key={idx} className="dn-item-row">
                    <div className="dn-item-desc">
                      <input
                        type="text"
                        placeholder="e.g. Demurrage - Container XYZW"
                        value={item.description}
                        onChange={(e) => updateItem(idx, 'description', e.target.value)}
                        className={errors[`item_${idx}_desc`] ? 'error' : ''}
                      />
                    </div>
                    <div className="dn-item-cat">
                      <select
                        value={item.service_category}
                        onChange={(e) => updateItem(idx, 'service_category', e.target.value)}
                      >
                        <option value="">None</option>
                        {SERVICE_CATEGORIES.map(c => <option key={c} value={c}>{c}</option>)}
                      </select>
                    </div>
                    <div className="dn-item-qty">
                      <input
                        type="number"
                        min="0.01"
                        step="0.01"
                        value={item.quantity}
                        onChange={(e) => updateItem(idx, 'quantity', e.target.value)}
                      />
                    </div>
                    <div className="dn-item-price">
                      <input
                        type="number"
                        min="0"
                        step="0.01"
                        placeholder="0.00"
                        value={item.unit_price}
                        onChange={(e) => updateItem(idx, 'unit_price', e.target.value)}
                        className={errors[`item_${idx}_price`] ? 'error' : ''}
                      />
                    </div>
                    <div className="dn-item-total">
                      <span>{form.currency} {lineTotal.toFixed(2)}</span>
                    </div>
                    <div className="dn-item-remove">
                      {lineItems.length > 1 && (
                        <button type="button" onClick={() => removeItem(idx)} className="dn-remove-btn" aria-label="Remove">
                          <Trash2 size={14} />
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Totals */}
            <div className="dn-totals">
              <div className="dn-total-row">
                <span>Subtotal</span>
                <span>{form.currency} {subtotal.toFixed(2)}</span>
              </div>
              <div className="dn-total-row dn-tax-row">
                <span>Tax</span>
                <div className="dn-tax-input-wrap">
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    value={form.tax_amount}
                    onChange={(e) => setForm(f => ({ ...f, tax_amount: e.target.value }))}
                  />
                </div>
              </div>
              <div className="dn-total-row dn-grand-total">
                <span>Total</span>
                <strong>{form.currency} {total.toFixed(2)}</strong>
              </div>
            </div>
          </div>

          {/* Dates & Notes */}
          <div className="dn-form-section">
            <h3 className="dn-section-label">Dates & Notes</h3>
            <div className="dn-form-grid-2">
              <div className="dn-form-field">
                <label>Issue Date <span className="dn-optional">(optional)</span></label>
                <input type="date" value={form.issue_date} onChange={(e) => setForm(f => ({ ...f, issue_date: e.target.value }))} />
              </div>
              <div className="dn-form-field">
                <label>Due Date <span className="dn-optional">(optional)</span></label>
                <input type="date" value={form.due_date} onChange={(e) => setForm(f => ({ ...f, due_date: e.target.value }))} />
              </div>
            </div>
            <div className="dn-form-field dn-form-full">
              <label>Internal Notes <span className="dn-optional">(optional)</span></label>
              <textarea
                rows={2}
                placeholder="Internal notes or references..."
                value={form.notes}
                onChange={(e) => setForm(f => ({ ...f, notes: e.target.value }))}
              />
            </div>
          </div>

          {/* Issue Immediately toggle */}
          <div className="dn-issue-toggle">
            <label className="dn-toggle-label">
              <input
                type="checkbox"
                checked={form.issue_immediately}
                onChange={(e) => setForm(f => ({ ...f, issue_immediately: e.target.checked }))}
              />
              <span className="dn-toggle-track" />
              <span className="dn-toggle-text">Issue immediately (skip draft)</span>
            </label>
          </div>

          {/* Footer */}
          <div className="dn-modal-footer">
            <button type="button" className="dn-btn-secondary" onClick={handleClose} disabled={submitting}>
              Cancel
            </button>
            <button type="submit" className="dn-btn-primary" disabled={submitting}>
              {submitting ? 'Creating…' : form.issue_immediately ? 'Create & Issue' : 'Save as Draft'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
