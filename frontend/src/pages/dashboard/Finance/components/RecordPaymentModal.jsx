import React, { useState, useEffect } from 'react';
import { X, DollarSign, Calendar, CreditCard, FileText, AlertCircle } from 'lucide-react';
import api from '../../../../services/api';
import './RecordPaymentModal.css';

export default function RecordPaymentModal({ isOpen, invoice, onClose, onSuccess, showToast }) {
  if (!isOpen || !invoice) return null;

  const todayStr = new Date().toISOString().split('T')[0];
  const remainingBalance = Number(invoice.balance || invoice.balance_due || 0);
  const totalAmount = Number(invoice.amount || invoice.total_amount || 0);
  const paidAmount = Number(invoice.amountPaid || invoice.paid_amount || 0);

  const [amount, setAmount] = useState(remainingBalance > 0 ? remainingBalance : 0);
  const [paymentDate, setPaymentDate] = useState(todayStr);
  const [paymentMethod, setPaymentMethod] = useState('Wire Transfer');
  const [paymentRef, setPaymentRef] = useState(`PAY-${new Date().getFullYear()}-${Math.floor(1000 + Math.random() * 9000)}`);
  const [notes, setNotes] = useState('');

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState(null);

  useEffect(() => {
    if (invoice) {
      const bal = Number(invoice.balance || invoice.balance_due || 0);
      setAmount(bal > 0 ? bal : 0);
      setPaymentRef(`PAY-${new Date().getFullYear()}-${Math.floor(1000 + Math.random() * 9000)}`);
      setErrorMsg(null);
    }
  }, [invoice]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setErrorMsg(null);

    const numAmount = Number(amount);
    if (isNaN(numAmount) || numAmount <= 0) {
      setErrorMsg('Payment amount must be greater than zero.');
      return;
    }

    if (numAmount > remainingBalance + 0.01) {
      setErrorMsg(`Payment amount ($${numAmount.toFixed(2)}) cannot exceed the remaining balance ($${remainingBalance.toFixed(2)}).`);
      return;
    }

    try {
      setIsSubmitting(true);
      const payload = {
        amount: numAmount,
        payment_date: paymentDate,
        payment_method: paymentMethod,
        payment_ref: paymentRef,
        notes: notes
      };

      const updatedInv = await api.post(`/api/v1/invoices/${invoice.id}/payments`, payload);

      if (showToast) {
        showToast(`Recorded payment of $${numAmount.toFixed(2)} for ${invoice.invoiceNumber || ''}`, 'success');
      }

      if (onSuccess) {
        onSuccess(updatedInv);
      }

      onClose();
    } catch (err) {
      console.error('Failed to record payment:', err);
      setErrorMsg(err.message || 'Failed to record payment on server.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="record-payment-modal-overlay">
      <div className="record-payment-modal-container">
        {/* Header */}
        <div className="record-payment-header">
          <div className="header-title-flex">
            <div className="rpm-icon-badge">
              <DollarSign size={20} className="icon-green" />
            </div>
            <div>
              <h2>Record Payment</h2>
              <p>Allocate customer payment to Invoice #{invoice.invoiceNumber}</p>
            </div>
          </div>
          <button className="close-btn" onClick={onClose} disabled={isSubmitting}>
            <X size={18} />
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="record-payment-body">
          {errorMsg && (
            <div className="payment-alert-banner">
              <AlertCircle size={16} />
              <span>{errorMsg}</span>
            </div>
          )}

          {/* Financial Summary Box */}
          <div className="payment-summary-box">
            <div className="summary-col">
              <span className="col-label">Invoice Total</span>
              <span className="col-val">${totalAmount.toLocaleString('en-US', { minimumFractionDigits: 2 })}</span>
            </div>
            <div className="summary-col">
              <span className="col-label">Paid to Date</span>
              <span className="col-val text-green">${paidAmount.toLocaleString('en-US', { minimumFractionDigits: 2 })}</span>
            </div>
            <div className="summary-col bold-col">
              <span className="col-label">Balance Due</span>
              <span className="col-val text-red">${remainingBalance.toLocaleString('en-US', { minimumFractionDigits: 2 })}</span>
            </div>
          </div>

          <div className="form-grid-2col">
            <div className="form-group">
              <label>Payment Amount ({invoice.currency || 'USD'}) *</label>
              <div className="input-prefix-wrapper">
                <span className="prefix-symbol">$</span>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  max={remainingBalance}
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  placeholder="0.00"
                  className="form-control-input prefix-padding"
                  required
                />
              </div>
            </div>

            <div className="form-group">
              <label>Payment Date *</label>
              <input
                type="date"
                value={paymentDate}
                onChange={(e) => setPaymentDate(e.target.value)}
                className="form-control-input"
                required
              />
            </div>

            <div className="form-group">
              <label>Payment Method *</label>
              <select
                value={paymentMethod}
                onChange={(e) => setPaymentMethod(e.target.value)}
                className="form-control-select"
              >
                <option value="Wire Transfer">Wire Transfer</option>
                <option value="ACH Transfer">ACH Transfer</option>
                <option value="Credit Card">Credit Card</option>
                <option value="Check">Check</option>
                <option value="Bank Transfer">Bank Transfer</option>
                <option value="Cash">Cash</option>
              </select>
            </div>

            <div className="form-group">
              <label>Payment Reference / Txn ID *</label>
              <input
                type="text"
                value={paymentRef}
                onChange={(e) => setPaymentRef(e.target.value)}
                placeholder="e.g. WT-2026-9901"
                className="form-control-input"
                required
              />
            </div>
          </div>

          <div className="form-group margin-top-12">
            <label>Internal Payment Notes</label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Optional remarks, bank confirmation numbers, or check details..."
              className="form-control-textarea"
              rows={3}
            />
          </div>

          {/* Modal Footer */}
          <div className="record-payment-footer">
            <button
              type="button"
              className="btn-cancel"
              onClick={onClose}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-submit-payment"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Recording Payment...' : `Record Payment ($${Number(amount || 0).toFixed(2)})`}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
