import React, { useState } from 'react';
import { X, Filter, RotateCcw } from 'lucide-react';
import './InvoiceFilterDrawer.css';

export default function InvoiceFilterDrawer({
  isOpen,
  onClose,
  filters,
  onApplyFilters,
  onResetFilters
}) {
  const [localFilters, setLocalFilters] = useState(filters);

  if (!isOpen) return null;

  const handleChange = (field, value) => {
    setLocalFilters((prev) => ({ ...prev, [field]: value }));
  };

  const handleApply = () => {
    onApplyFilters(localFilters);
    onClose();
  };

  const handleReset = () => {
    const emptyFilters = {
      customer: '',
      shipmentId: '',
      status: 'All',
      dateFrom: '',
      dateTo: '',
      minAmount: '',
      maxAmount: '',
      currency: ''
    };
    setLocalFilters(emptyFilters);
    onResetFilters();
  };

  return (
    <div className="filter-drawer-overlay" onClick={onClose}>
      <div className="filter-drawer-container" onClick={(e) => e.stopPropagation()}>
        <div className="filter-drawer-header">
          <div className="filter-title-group">
            <Filter size={18} className="filter-header-icon" />
            <h3>Filter Invoices</h3>
          </div>
          <button className="filter-close-btn" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        <div className="filter-drawer-body">
          {/* Status */}
          <div className="filter-field-group">
            <label>Invoice Status</label>
            <select
              value={localFilters.status || 'All'}
              onChange={(e) => handleChange('status', e.target.value)}
              className="filter-select"
            >
              <option value="All">All Statuses</option>
              <option value="Draft">Draft</option>
              <option value="Pending Approval">Pending Approval</option>
              <option value="Issued">Issued</option>
              <option value="Partially Paid">Partially Paid</option>
              <option value="Paid">Paid</option>
              <option value="Overdue">Overdue</option>
              <option value="Cancelled">Cancelled</option>
            </select>
          </div>

          {/* Customer */}
          <div className="filter-field-group">
            <label>Customer Name</label>
            <input
              type="text"
              placeholder="e.g. Global Traders"
              value={localFilters.customer || ''}
              onChange={(e) => handleChange('customer', e.target.value)}
              className="filter-input"
            />
          </div>

          {/* Shipment ID */}
          <div className="filter-field-group">
            <label>Shipment / Reference #</label>
            <input
              type="text"
              placeholder="e.g. SH-2026-00124"
              value={localFilters.shipmentId || ''}
              onChange={(e) => handleChange('shipmentId', e.target.value)}
              className="filter-input"
            />
          </div>

          {/* Date Range */}
          <div className="filter-field-group">
            <label>Invoice Date Range</label>
            <div className="filter-date-row">
              <input
                type="date"
                value={localFilters.dateFrom || ''}
                onChange={(e) => handleChange('dateFrom', e.target.value)}
                className="filter-input"
              />
              <span className="date-sep">to</span>
              <input
                type="date"
                value={localFilters.dateTo || ''}
                onChange={(e) => handleChange('dateTo', e.target.value)}
                className="filter-input"
              />
            </div>
          </div>

          {/* Amount Range */}
          <div className="filter-field-group">
            <label>Amount Range ($)</label>
            <div className="filter-amount-row">
              <input
                type="number"
                placeholder="Min"
                value={localFilters.minAmount || ''}
                onChange={(e) => handleChange('minAmount', e.target.value)}
                className="filter-input"
              />
              <span className="date-sep">-</span>
              <input
                type="number"
                placeholder="Max"
                value={localFilters.maxAmount || ''}
                onChange={(e) => handleChange('maxAmount', e.target.value)}
                className="filter-input"
              />
            </div>
          </div>
        </div>

        <div className="filter-drawer-footer">
          <button className="btn-reset-filter" onClick={handleReset}>
            <RotateCcw size={14} /> Reset
          </button>
          <div className="footer-right-btns">
            <button className="btn-cancel-filter" onClick={onClose}>
              Cancel
            </button>
            <button className="btn-apply-filter" onClick={handleApply}>
              Apply Filters
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
