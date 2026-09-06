import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  Plus, Search, Filter, ChevronDown, X, Eye, Edit2, Send,
  Download, MoreHorizontal, ArrowRight, RefreshCw, FileText,
  ChevronLeft, ChevronRight, User, MapPin, Calendar, DollarSign,
  Clock, CheckCircle, XCircle, AlertCircle, TrendingUp,
  Trash2, Layers, Percent, Sparkles, Check, Info, ShieldAlert,
  Lock, Bookmark, CreditCard, FileCheck, Copy, Shield, Settings,
  Ship, Plane, Truck, Box, Globe, CheckCircle2, ThumbsUp,
  ThumbsDown, RotateCcw, FileSignature, History, UserCheck,
  Printer, ExternalLink, Ban, Share2, CheckSquare, Briefcase,
  CheckCheck, ArrowUpRight, Compass, ShieldCheck, BarChart3, GitBranch
} from 'lucide-react';
import quotationService from '../../../services/quotationService';
import QuotationAnalyticsWorkspace from './QuotationAnalyticsWorkspace';
import RateSelectionSection from './RateSelectionSection';
import ModuleHeroEmptyState from '../../../components/dashboard/ModuleHeroEmptyState';
import './QuotationsPage.css';

// ─── Constants ────────────────────────────────────────────────────────────────

const STATUS_TABS = [
  'All',
  'Draft',
  'Ready For Review',
  'Approved',
  'Changes Requested',
  'Sent',
  'Viewed',
  'Accepted',
  'Declined',
  'Expired',
  'Cancelled',
];

const STATUS_CONFIG = {
  DRAFT:             { label: 'DRAFT',             bg: '#F1F5F9', color: '#475569', dotColor: '#94A3B8' },
  READY_FOR_REVIEW:  { label: 'READY FOR REVIEW',  bg: '#FEF3C7', color: '#D97706', dotColor: '#F59E0B' },
  APPROVED:          { label: 'APPROVED',          bg: '#E0E7FF', color: '#4338CA', dotColor: '#6366F1' },
  CHANGES_REQUESTED: { label: 'CHANGES REQUESTED', bg: '#FEE2E2', color: '#B91C1C', dotColor: '#EF4444' },
  SENT:              { label: 'SENT',              bg: '#DBEAFE', color: '#1D4ED8', dotColor: '#3B82F6' },
  VIEWED:            { label: 'VIEWED',            bg: '#EDE9FE', color: '#7C3AED', dotColor: '#8B5CF6' },
  ACCEPTED:          { label: 'ACCEPTED',          bg: '#DCFCE7', color: '#15803D', dotColor: '#22C55E' },
  DECLINED:          { label: 'DECLINED',          bg: '#FEE2E2', color: '#DC2626', dotColor: '#EF4444' },
  REJECTED:          { label: 'REJECTED',          bg: '#FEE2E2', color: '#DC2626', dotColor: '#EF4444' },
  EXPIRED:           { label: 'EXPIRED',           bg: '#FEF3C7', color: '#92400E', dotColor: '#F59E0B' },
  CANCELLED:         { label: 'CANCELLED',         bg: '#F3F4F6', color: '#6B7280', dotColor: '#9CA3AF' },
};

const KPI_CONFIG = [
  { key: 'total_quotations', label: 'Total Quotations', icon: FileText,   iconBg: 'rgba(37, 99, 235, 0.1)',  iconColor: '#2563EB', trend: '+18%', badge: 'Total' },
  { key: 'draft_count',      label: 'Draft',            icon: Edit2,      iconBg: 'rgba(100, 116, 139, 0.1)', iconColor: '#475569', trend: '+8%', badge: 'WIP' },
  { key: 'sent_count',       label: 'Sent',             icon: Send,       iconBg: 'rgba(99, 102, 241, 0.1)',  iconColor: '#6366F1', trend: '+12%', badge: 'Active' },
  { key: 'viewed_count',     label: 'Viewed',           icon: Eye,        iconBg: 'rgba(217, 119, 6, 0.1)',   iconColor: '#D97706', trend: '+5%', badge: 'Seen' },
  { key: 'accepted_count',   label: 'Accepted',         icon: CheckCircle,iconBg: 'rgba(22, 163, 74, 0.1)',   iconColor: '#16A34A', trend: '+25%', badge: 'Won' },
  { key: 'expired_count',    label: 'Expired',          icon: Clock,      iconBg: 'rgba(220, 38, 38, 0.1)',   iconColor: '#DC2626', trend: '-3%', negative: true, badge: 'Overdue' },
];

const SERVICE_TYPE_OPTIONS = [
  { value: 'FCL', label: 'FCL', sublabel: 'Full Container Load', icon: Layers },
  { value: 'LCL', label: 'LCL', sublabel: 'Less than Container Load', icon: Box },
  { value: 'Air Freight', label: 'Air Freight', sublabel: 'Fast Air Cargo', icon: Plane },
  { value: 'Rail', label: 'Rail Freight', sublabel: 'Intermodal Rail Cargo', icon: Layers },
  { value: 'Road', label: 'Road Transport', sublabel: 'FTL / LTL Trucking', icon: Truck },
  { value: 'Multimodal', label: 'Multimodal', sublabel: 'Combined Multi-Leg Freight', icon: Globe },
];

const TRANSPORT_MODE_OPTIONS = [
  { value: 'Ocean Freight', label: 'Ocean Freight', sublabel: 'Sea Vessel / Container Line', icon: Ship },
  { value: 'Air Freight', label: 'Air Freight', sublabel: 'Commercial & Cargo Airway', icon: Plane },
  { value: 'Rail', label: 'Rail', sublabel: 'Rail Network Logistics', icon: Layers },
  { value: 'Road', label: 'Road', sublabel: 'Truck & Highway Transport', icon: Truck },
  { value: 'Multimodal', label: 'Multimodal', sublabel: 'Multi-Leg Combination', icon: Globe },
];

const CURRENCY_OPTIONS = [
  { value: 'USD', label: 'USD ($)', sublabel: 'US Dollar', badge: '$' },
  { value: 'EUR', label: 'EUR (€)', sublabel: 'Euro', badge: '€' },
  { value: 'GBP', label: 'GBP (£)', sublabel: 'British Pound', badge: '£' },
  { value: 'INR', label: 'INR (₹)', sublabel: 'Indian Rupee', badge: '₹' },
  { value: 'SGD', label: 'SGD (S$)', sublabel: 'Singapore Dollar', badge: 'S$' },
  { value: 'AED', label: 'AED (د.إ)', sublabel: 'UAE Dirham', badge: 'AED' },
];

const CHARGE_CATEGORIES = [
  { value: 'FREIGHT',       label: 'Freight', sublabel: 'Base Ocean / Air / Land freight rate', icon: Layers },
  { value: 'ORIGIN',        label: 'Origin Charges', sublabel: 'Origin terminal handling, CFS, cartage', icon: MapPin },
  { value: 'DESTINATION',   label: 'Destination Charges', sublabel: 'Destination terminal handling, delivery', icon: MapPin },
  { value: 'SURCHARGE',     label: 'Surcharges', sublabel: 'Fuel (BAF), currency (CAF), peak season', icon: Percent },
  { value: 'DOCUMENTATION', label: 'Documentation', sublabel: 'B/L issuance, export/import clearance', icon: FileText },
  { value: 'CUSTOMS',       label: 'Customs Brokerage', sublabel: 'Customs entry & duty processing', icon: Shield },
  { value: 'INSURANCE',     label: 'Marine Cargo Insurance', sublabel: 'All-risk cargo coverage', icon: ShieldAlert },
  { value: 'TAX',           label: 'Taxes & Levies', sublabel: 'Government fees, port taxes, GST/VAT', icon: DollarSign },
  { value: 'OTHER',         label: 'Other Auxiliary', sublabel: 'Inspection, palletization, special handling', icon: Sparkles },
];

const CALC_BASES = [
  { value: 'FLAT',          label: 'Flat / Lump Sum', sublabel: 'Fixed charge per entire quotation' },
  { value: 'PER_CONTAINER', label: 'Per Container', sublabel: 'Multiplied by total container units' },
  { value: 'PER_SHIPMENT',  label: 'Per Shipment', sublabel: 'Single fixed fee per shipment' },
  { value: 'PER_WEIGHT',    label: 'Per Weight (kg/ton)', sublabel: 'Charge based on chargeable gross weight' },
  { value: 'PER_VOLUME',    label: 'Per Volume (CBM)', sublabel: 'Charge based on total cubic meters' },
  { value: 'PER_UNIT',      label: 'Per Package / Unit', sublabel: 'Charge per individual carton or pallet' },
  { value: 'PERCENTAGE',    label: 'Percentage (%)', sublabel: 'Calculated as percentage of line/base value' },
];

const DISCOUNT_TYPE_OPTIONS = [
  { value: 'NONE',       label: 'No Discount', sublabel: 'Full price applies' },
  { value: 'PERCENTAGE', label: 'Percentage (%)', sublabel: 'Deducts percentage from item total' },
  { value: 'FIXED',      label: 'Fixed Amount', sublabel: 'Deducts fixed currency sum' },
];

const PAYMENT_TERMS_OPTIONS = [
  { value: 'PREPAID',        label: 'Prepaid (Origin Paid)', sublabel: 'Shipper/Origin pays freight charges', icon: CreditCard },
  { value: 'COLLECT',        label: 'Collect (Destination Paid)', sublabel: 'Consignee pays at destination', icon: CreditCard },
  { value: 'NET_15',         label: 'Net 15 Days', sublabel: 'Payment due 15 days from invoice', icon: Calendar },
  { value: 'NET_30',         label: 'Net 30 Days', sublabel: 'Standard commercial 30-day term', icon: Calendar },
  { value: 'NET_60',         label: 'Net 60 Days', sublabel: 'Extended 60-day commercial credit', icon: Calendar },
  { value: 'DUE_ON_RECEIPT', label: 'Due on Receipt', sublabel: 'Immediate payment upon receiving invoice', icon: Clock },
  { value: 'CUSTOM',         label: 'Custom Terms', sublabel: 'Tailored payment schedule as agreed', icon: Settings },
];

// ─── Helpers ─────────────────────────────────────────────────────────────────

function fmtDate(d) {
  if (!d) return '—';
  return new Date(d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
}

function fmtAmount(amount, currency = 'USD') {
  if (amount === undefined || amount === null || isNaN(amount)) return '—';
  return `${currency} ${Number(amount).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function daysUntilExpiry(validUntil) {
  if (!validUntil) return null;
  const diff = Math.ceil((new Date(validUntil) - new Date()) / (1000 * 60 * 60 * 24));
  return diff;
}

function expiryLabel(validUntil) {
  if (!validUntil) return null;
  const days = daysUntilExpiry(validUntil);
  if (days < 0) return { text: 'Expired', color: '#DC2626', bg: '#FEE2E2', status: 'EXPIRED' };
  if (days === 0) return { text: 'Expires today', color: '#EA580C', bg: '#FFF7ED', status: 'EXPIRING_SOON' };
  if (days <= 7)  return { text: `${days} days left`, color: '#EA580C', bg: '#FFF7ED', status: 'EXPIRING_SOON' };
  if (days <= 14) return { text: `${days} days left`, color: '#D97706', bg: '#FFFBEB', status: 'ACTIVE' };
  return { text: `${days} days left`, color: '#15803D', bg: '#F0FDF4', status: 'ACTIVE' };
}

function timeAgo(dateStr) {
  if (!dateStr) return '—';
  const diff = Math.floor((Date.now() - new Date(dateStr)) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 604800) return `${Math.floor(diff / 86400)}d ago`;
  return fmtDate(dateStr);
}

function StatusBadge({ status }) {
  const cfg = STATUS_CONFIG[status] || STATUS_CONFIG.DRAFT;
  return (
    <span className="qt-status-badge" style={{ background: cfg.bg, color: cfg.color }}>
      <span className="qt-status-dot" style={{ background: cfg.dotColor }} />
      {cfg.label}
    </span>
  );
}

function MarginBadge({ health, marginPct }) {
  let bg = '#DCFCE7';
  let color = '#15803D';
  let label = 'HEALTHY';

  if (health === 'NEGATIVE' || marginPct < 0) {
    bg = '#FEE2E2';
    color = '#DC2626';
    label = 'NEGATIVE';
  } else if (health === 'LOW' || (marginPct >= 0 && marginPct < 15)) {
    bg = '#FEF3C7';
    color = '#B45309';
    label = 'LOW';
  }

  return (
    <span className="qt-margin-badge" style={{ background: bg, color }}>
      <span className="qt-margin-badge-dot" style={{ background: color }} />
      {label} ({Number(marginPct || 0).toFixed(1)}%)
    </span>
  );
}

function ValidityBadge({ validUntil }) {
  const exp = expiryLabel(validUntil);
  if (!exp) {
    return <span className="qt-validity-badge not-set">Validity Not Set</span>;
  }
  return (
    <span className="qt-validity-badge" style={{ background: exp.bg, color: exp.color }}>
      <span className="qt-validity-dot" style={{ background: exp.color }} />
      {exp.status === 'EXPIRED' ? 'EXPIRED' : exp.status === 'EXPIRING_SOON' ? 'EXPIRING SOON' : 'ACTIVE'} ({exp.text})
    </span>
  );
}

function renderSelectIcon(Icon, size = 14) {
  if (!Icon) return null;
  if (React.isValidElement(Icon)) return Icon;
  if (typeof Icon === 'function' || (typeof Icon === 'object' && (Icon.$$typeof || Icon.render))) {
    const Component = Icon;
    return <Component size={size} />;
  }
  if (typeof Icon === 'string') return <span>{Icon}</span>;
  return null;
}

// ─── Custom Select Component ──────────────────────────────────────────────────

function CustomSelect({
  value,
  onChange,
  options = [],
  placeholder = 'Select...',
  disabled = false,
  className = '',
  size = 'md',
  icon: PrefixIcon,
  ariaLabel,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [dropUp, setDropUp] = useState(false);
  const dropdownRef = useRef(null);

  const normalizedOptions = options.map(opt => {
    if (typeof opt === 'string' || typeof opt === 'number') {
      return { value: String(opt), label: String(opt) };
    }
    return { ...opt, value: String(opt.value) };
  });

  const selectedOption = normalizedOptions.find(opt => String(opt.value) === String(value));

  const toggleOpen = () => {
    if (disabled) return;
    if (!isOpen && dropdownRef.current) {
      const rect = dropdownRef.current.getBoundingClientRect();
      const spaceBelow = window.innerHeight - rect.bottom;
      if (spaceBelow < 250 && rect.top > 250) {
        setDropUp(true);
      } else {
        setDropUp(false);
      }
    }
    setIsOpen(prev => !prev);
  };

  useEffect(() => {
    if (!isOpen) return;
    const handleClickOutside = (e) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  const handleSelect = (optVal) => {
    onChange(optVal);
    setIsOpen(false);
  };

  const SelectedIcon = selectedOption?.icon || PrefixIcon;

  return (
    <div
      ref={dropdownRef}
      className={`qt-custom-select-container ${isOpen ? 'is-open' : ''} ${disabled ? 'is-disabled' : ''} ${className}`}
    >
      <button
        type="button"
        className={`qt-custom-select-trigger qt-custom-select-trigger--${size} ${isOpen ? 'active' : ''}`}
        onClick={toggleOpen}
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        aria-label={ariaLabel}
      >
        <div className="qt-custom-select-trigger-content">
          {SelectedIcon && (
            <span className="qt-custom-select-trigger-icon">
              {renderSelectIcon(SelectedIcon, 14)}
            </span>
          )}
          {selectedOption?.badge && (
            <span className="qt-custom-select-trigger-badge">{selectedOption.badge}</span>
          )}
          <span className={`qt-custom-select-trigger-label ${!selectedOption ? 'is-placeholder' : ''}`}>
            {selectedOption ? selectedOption.label : placeholder}
          </span>
        </div>
        <ChevronDown size={14} className={`qt-custom-select-chevron ${isOpen ? 'is-rotated' : ''}`} />
      </button>

      {isOpen && (
        <div className={`qt-custom-select-dropdown ${dropUp ? 'is-dropup' : ''}`} role="listbox">
          <div className="qt-custom-select-options">
            {normalizedOptions.map((opt) => {
              const isSelected = String(opt.value) === String(value);
              const OptIcon = opt.icon;
              return (
                <div
                  key={opt.value}
                  role="option"
                  aria-selected={isSelected}
                  className={`qt-custom-select-option ${isSelected ? 'is-selected' : ''}`}
                  onClick={() => handleSelect(opt.value)}
                >
                  <div className="qt-custom-select-option-main">
                    {OptIcon && (
                      <span className="qt-custom-select-option-icon">
                        {renderSelectIcon(OptIcon, 14)}
                      </span>
                    )}
                    {opt.badge && (
                      <span className="qt-custom-select-option-badge">{opt.badge}</span>
                    )}
                    <div className="qt-custom-select-option-text">
                      <span className="qt-custom-select-option-label">{opt.label}</span>
                      {opt.sublabel && (
                        <span className="qt-custom-select-option-sublabel">{opt.sublabel}</span>
                      )}
                    </div>
                  </div>
                  {isSelected && (
                    <Check size={14} className="qt-custom-select-option-check" />
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Add / Edit Charge Modal ──────────────────────────────────────────────────

function AddEditChargeModal({ quotation, initialData, onClose, onSaved }) {
  const isEdit = Boolean(initialData?.id);
  const [form, setForm] = useState({
    charge_code: initialData?.charge_code || '',
    charge_name: initialData?.charge_name || '',
    charge_category: initialData?.charge_category || 'FREIGHT',
    charge_type: initialData?.charge_type || 'SELL',
    calculation_basis: initialData?.calculation_basis || 'PER_CONTAINER',
    quantity: initialData?.quantity ?? 1,
    unit_price: initialData?.unit_price ?? 0,
    cost_amount: initialData?.cost_amount ?? 0,
    currency: initialData?.currency || quotation?.currency || 'USD',
    tax_rate: initialData?.tax_rate ?? 0,
    discount_type: initialData?.discount_type || 'NONE',
    discount_value: initialData?.discount_value ?? 0,
    is_optional: initialData?.is_optional ?? false,
    notes: initialData?.notes || '',
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const set = (k, v) => setForm(f => ({ ...f, [k]: v }));

  const qty = Number(form.quantity) || 1;
  const unitPrice = Number(form.unit_price) || 0;
  const cost = Number(form.cost_amount) || 0;
  const taxRate = Number(form.tax_rate) || 0;
  const discVal = Number(form.discount_value) || 0;

  let baseSell = form.calculation_basis === 'PERCENTAGE'
    ? qty * (unitPrice / 100)
    : qty * unitPrice;

  let discountAmt = 0;
  if (form.discount_type === 'PERCENTAGE') {
    discountAmt = baseSell * (discVal / 100);
  } else if (form.discount_type === 'FIXED') {
    discountAmt = discVal;
  }
  discountAmt = Math.min(discountAmt, baseSell);

  const sellAfterDiscount = Math.max(0, baseSell - discountAmt);
  const taxAmt = sellAfterDiscount * (taxRate / 100);
  const totalSell = sellAfterDiscount + taxAmt;
  const totalCost = qty * cost;
  const profit = totalSell - totalCost;
  const marginPct = totalSell > 0 ? (profit / totalSell) * 100 : 0;

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!form.charge_name.trim()) {
      setError('Charge name is required.');
      return;
    }
    if (form.quantity <= 0) {
      setError('Quantity must be greater than 0.');
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const payload = {
        charge_code: form.charge_code || form.charge_category.substring(0, 3),
        charge_name: form.charge_name,
        charge_category: form.charge_category,
        charge_type: form.charge_type,
        calculation_basis: form.calculation_basis,
        quantity: Number(form.quantity),
        unit_price: Number(form.unit_price),
        cost_amount: Number(form.cost_amount),
        currency: form.currency,
        tax_rate: Number(form.tax_rate),
        discount_type: form.discount_type,
        discount_value: Number(form.discount_value),
        is_optional: Boolean(form.is_optional),
        notes: form.notes,
      };

      let res;
      if (isEdit) {
        res = await quotationService.updateQuotationCharge(quotation.id, initialData.id, payload);
      } else {
        res = await quotationService.addQuotationCharge(quotation.id, payload);
      }
      onSaved(res?.data || res);
    } catch (err) {
      setError(err?.message || 'Failed to save charge line item.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="qt-modal-overlay" onClick={onClose}>
      <div className="qt-modal qt-modal--wide" onClick={e => e.stopPropagation()}>
        <div className="qt-modal-header">
          <div>
            <h2 className="qt-modal-title">{isEdit ? 'Edit Line Item Charge' : 'Add Charge Line Item'}</h2>
            <p className="qt-modal-sub">
              {quotation?.quotation_number} • Configure commercial pricing & cost breakdown
            </p>
          </div>
          <button className="qt-modal-close" onClick={onClose}><X size={20} /></button>
        </div>

        <form className="qt-modal-body" onSubmit={handleSubmit}>
          {error && <div className="qt-modal-error">{error}</div>}

          <div className="qt-form-row">
            <div className="qt-form-group" style={{ flex: 2 }}>
              <label>Charge Name *</label>
              <input
                placeholder="e.g. Ocean Freight 40GP, Origin THC, BAF"
                value={form.charge_name}
                onChange={e => set('charge_name', e.target.value)}
                required
              />
            </div>
            <div className="qt-form-group">
              <label>Code</label>
              <input
                placeholder="BAS, THC, BAF"
                value={form.charge_code}
                onChange={e => set('charge_code', e.target.value.toUpperCase())}
              />
            </div>
            <div className="qt-form-group">
              <label>Category</label>
              <CustomSelect
                value={form.charge_category}
                onChange={v => set('charge_category', v)}
                options={CHARGE_CATEGORIES}
              />
            </div>
          </div>

          <div className="qt-form-row">
            <div className="qt-form-group">
              <label>Calculation Basis</label>
              <CustomSelect
                value={form.calculation_basis}
                onChange={v => set('calculation_basis', v)}
                options={CALC_BASES}
              />
            </div>
            <div className="qt-form-group qt-form-group--sm">
              <label>Quantity</label>
              <input
                type="number"
                step="any"
                min="0.0001"
                value={form.quantity}
                onChange={e => set('quantity', e.target.value)}
              />
            </div>
            <div className="qt-form-group">
              <label>Unit Sell Price ({form.currency}) *</label>
              <input
                type="number"
                step="any"
                min="0"
                value={form.unit_price}
                onChange={e => set('unit_price', e.target.value)}
                required
              />
            </div>
            <div className="qt-form-group">
              <label>Unit Buy Cost ({form.currency})</label>
              <input
                type="number"
                step="any"
                min="0"
                placeholder="0.00"
                value={form.cost_amount}
                onChange={e => set('cost_amount', e.target.value)}
              />
            </div>
          </div>

          <div className="qt-form-row">
            <div className="qt-form-group">
              <label>Discount Type</label>
              <CustomSelect
                value={form.discount_type}
                onChange={v => set('discount_type', v)}
                options={DISCOUNT_TYPE_OPTIONS}
              />
            </div>
            {form.discount_type !== 'NONE' && (
              <div className="qt-form-group">
                <label>Discount Value {form.discount_type === 'PERCENTAGE' ? '(%)' : `(${form.currency})`}</label>
                <input
                  type="number"
                  step="any"
                  min="0"
                  value={form.discount_value}
                  onChange={e => set('discount_value', e.target.value)}
                />
              </div>
            )}
            <div className="qt-form-group">
              <label>Tax Rate (%)</label>
              <input
                type="number"
                step="any"
                min="0"
                max="100"
                placeholder="0.0"
                value={form.tax_rate}
                onChange={e => set('tax_rate', e.target.value)}
              />
            </div>
            <div className="qt-form-group qt-form-group--sm">
              <label>Currency</label>
              <CustomSelect
                value={form.currency}
                onChange={v => set('currency', v)}
                options={CURRENCY_OPTIONS}
                size="sm"
              />
            </div>
          </div>

          <div className="qt-form-row qt-form-row--align-center">
            <label className="qt-checkbox-label">
              <input
                type="checkbox"
                checked={form.is_optional}
                onChange={e => set('is_optional', e.target.checked)}
              />
              <span>Optional Charge (Customer add-on, excluded from base subtotal)</span>
            </label>
          </div>

          <div className="qt-form-group">
            <label>Notes / Operational Conditions</label>
            <input
              placeholder="e.g. Free time 14 days, subject to carrier GRI"
              value={form.notes}
              onChange={e => set('notes', e.target.value)}
            />
          </div>

          {/* Live Calculation Preview Banner */}
          <div className="qt-charge-preview-card">
            <div className="qt-preview-title">
              <Sparkles size={14} /> Live Commercial Calculation Preview
            </div>
            <div className="qt-preview-grid">
              <div>
                <span className="qt-preview-lbl">Total Sell:</span>
                <span className="qt-preview-val">{fmtAmount(totalSell, form.currency)}</span>
              </div>
              <div>
                <span className="qt-preview-lbl">Total Cost:</span>
                <span className="qt-preview-val">{fmtAmount(totalCost, form.currency)}</span>
              </div>
              <div>
                <span className="qt-preview-lbl">Gross Profit:</span>
                <span className="qt-preview-val" style={{ color: profit >= 0 ? '#15803D' : '#DC2626' }}>
                  {fmtAmount(profit, form.currency)}
                </span>
              </div>
              <div>
                <span className="qt-preview-lbl">Gross Margin:</span>
                <span className="qt-preview-val">
                  <MarginBadge health={profit < 0 ? 'NEGATIVE' : marginPct < 15 ? 'LOW' : 'HEALTHY'} marginPct={marginPct} />
                </span>
              </div>
            </div>
          </div>

          <div className="qt-modal-footer">
            <button type="button" className="qt-btn qt-btn--ghost" onClick={onClose}>Cancel</button>
            <button type="submit" className="qt-btn qt-btn--primary" disabled={loading}>
              {loading ? <><RefreshCw size={14} className="qt-spin" /> Saving…</> : <><Check size={14} /> {isEdit ? 'Save Changes' : 'Add Charge'}</>}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ─── Rate Candidates Drawer (Import from Rate Management) ──────────────────────

function RateCandidatesDrawer({ quotation, onClose, onImported }) {
  const [candidates, setCandidates] = useState([]);
  const [loading, setLoading] = useState(true);
  const [importingId, setImportingId] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    quotationService.getQuotationRateCandidates(quotation.id)
      .then(res => {
        const d = res?.data || res;
        setCandidates(Array.isArray(d) ? d : []);
      })
      .catch(err => {
        setError(err?.message || 'Failed to load rate candidates.');
      })
      .finally(() => setLoading(false));
  }, [quotation.id]);

  const handleImport = async (rateId) => {
    try {
      setImportingId(rateId);
      setError(null);
      const res = await quotationService.importRateCharges(quotation.id, { rate_id: rateId });
      onImported(res?.data || res);
    } catch (err) {
      setError(err?.message || 'Failed to import rate charges.');
    } finally {
      setImportingId(null);
    }
  };

  return (
    <div className="qt-modal-overlay" onClick={onClose}>
      <div className="qt-drawer" onClick={e => e.stopPropagation()}>
        <div className="qt-drawer-header">
          <div>
            <h2 className="qt-drawer-title">Import Rate from Rate Intelligence</h2>
            <p className="qt-drawer-sub">
              Available live carrier spot & contract rates for {quotation.origin || quotation.origin_code} → {quotation.destination || quotation.destination_code}
            </p>
          </div>
          <button className="qt-modal-close" onClick={onClose}><X size={20} /></button>
        </div>

        <div className="qt-drawer-body">
          {error && <div className="qt-modal-error">{error}</div>}

          {loading ? (
            <div className="qt-drawer-loading">
              <RefreshCw size={24} className="qt-spin" />
              <p>Searching confirmed rates for this lane…</p>
            </div>
          ) : candidates.length === 0 ? (
            <div className="qt-empty-state" style={{ padding: '40px 20px' }}>
              <div className="qt-empty-icon"><FileText size={40} /></div>
              <h3>No matching rates available</h3>
              <p>No confirmed rates were found for lane {quotation.origin} → {quotation.destination}.</p>
              <p className="qt-cell-sub">You can add custom line items directly using "+ Add Charge" or apply a template.</p>
            </div>
          ) : (
            <div className="qt-candidates-list">
              {candidates.map(c => (
                <div key={c.id} className="qt-candidate-card">
                  <div className="qt-candidate-top">
                    <div>
                      <div className="qt-candidate-carrier">
                        <strong>{c.carrier_name || c.carrier_scac}</strong>
                        <span className="qt-candidate-tag">{c.equipment_type || '40GP'}</span>
                        <span className={`qt-candidate-source ${c.source === 'CONTRACT_PDF' ? 'contract' : 'spot'}`}>
                          {c.source === 'CONTRACT_PDF' ? 'Contract Rate' : 'Spot Rate'}
                        </span>
                      </div>
                      <div className="qt-candidate-route">
                        {c.origin_port} → {c.destination_port}
                        {c.transit_days && <span> • {c.transit_days} days transit</span>}
                      </div>
                    </div>
                    <div className="qt-candidate-price-box">
                      <div className="qt-candidate-price">{fmtAmount(c.total_buy_price, c.currency)}</div>
                      <div className="qt-candidate-price-sub">Total Buy Cost / container</div>
                    </div>
                  </div>

                  <div className="qt-candidate-breakdown">
                    <div className="qt-breakdown-pill">
                      <span>Ocean Freight:</span> <strong>{fmtAmount(c.ocean_freight, c.currency)}</strong>
                    </div>
                    {c.origin_charges > 0 && (
                      <div className="qt-breakdown-pill">
                        <span>Origin:</span> <strong>{fmtAmount(c.origin_charges, c.currency)}</strong>
                      </div>
                    )}
                    {c.destination_charges > 0 && (
                      <div className="qt-breakdown-pill">
                        <span>Dest:</span> <strong>{fmtAmount(c.destination_charges, c.currency)}</strong>
                      </div>
                    )}
                    {c.surcharges?.length > 0 && (
                      <div className="qt-breakdown-pill">
                        <span>Surcharges:</span> <strong>{c.surcharges.length} items</strong>
                      </div>
                    )}
                  </div>

                  <div className="qt-candidate-footer">
                    <span className="qt-candidate-note">
                      ⚡ Rate components will be imported as persisted snapshots with commercial margin.
                    </span>
                    <button
                      className="qt-btn qt-btn--primary qt-btn--sm"
                      disabled={importingId === c.id}
                      onClick={() => handleImport(c.id)}
                    >
                      {importingId === c.id ? (
                        <><RefreshCw size={13} className="qt-spin" /> Importing…</>
                      ) : (
                        <><Sparkles size={13} /> Import This Rate</>
                      )}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Apply Template Modal (Task 18.3) ─────────────────────────────────────────

function ApplyTemplateModal({ quotation, onClose, onApplied }) {
  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedId, setSelectedId] = useState(null);
  const [overrideCharges, setOverrideCharges] = useState(true);
  const [applying, setApplying] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    quotationService.listTemplates({ active_only: true })
      .then(res => {
        const d = res?.data || res;
        const list = Array.isArray(d) ? d : [];
        setTemplates(list);
        if (list.length > 0) setSelectedId(list[0].id);
      })
      .catch(err => {
        setError(err?.message || 'Failed to load reusable templates.');
      })
      .finally(() => setLoading(false));
  }, []);

  const handleApply = async () => {
    if (!selectedId) return;
    try {
      setApplying(true);
      setError(null);
      const res = await quotationService.applyTemplate(quotation.id, {
        template_id: selectedId,
        override_charges: overrideCharges,
      });
      onApplied(res?.data || res);
    } catch (err) {
      setError(err?.message || 'Failed to apply template.');
    } finally {
      setApplying(false);
    }
  };

  const activeTmpl = templates.find(t => t.id === selectedId);

  return (
    <div className="qt-modal-overlay" onClick={onClose}>
      <div className="qt-modal qt-modal--wide" onClick={e => e.stopPropagation()}>
        <div className="qt-modal-header">
          <div>
            <h2 className="qt-modal-title">Apply Quotation Template</h2>
            <p className="qt-modal-sub">
              {quotation?.quotation_number} • Copy standardized charges, payment terms, and commercial terms
            </p>
          </div>
          <button className="qt-modal-close" onClick={onClose}><X size={20} /></button>
        </div>

        <div className="qt-modal-body">
          {error && <div className="qt-modal-error">{error}</div>}

          {loading ? (
            <div className="qt-drawer-loading">
              <RefreshCw size={24} className="qt-spin" />
              <p>Loading reusable quotation templates…</p>
            </div>
          ) : templates.length === 0 ? (
            <div className="qt-empty-state" style={{ padding: '30px 20px' }}>
              <div className="qt-empty-icon"><Bookmark size={36} /></div>
              <h3>No templates available</h3>
              <p>Create a template first from the Templates Drawer or save this quotation as a template.</p>
            </div>
          ) : (
            <>
              <div className="qt-template-select-list">
                {templates.map(t => (
                  <div
                    key={t.id}
                    className={`qt-template-card-option ${selectedId === t.id ? 'active' : ''}`}
                    onClick={() => setSelectedId(t.id)}
                  >
                    <div className="qt-template-card-top">
                      <div className="qt-template-card-name">
                        <Bookmark size={14} className="qt-template-icon" />
                        <strong>{t.name}</strong>
                      </div>
                      <span className="qt-badge-mode">{t.shipment_mode || 'Standard'}</span>
                    </div>
                    {t.description && <div className="qt-template-card-desc">{t.description}</div>}
                    <div className="qt-template-card-meta">
                      <span>📦 {t.charge_count || 0} charge items</span>
                      <span>📅 {t.validity_days || 30} days validity</span>
                      <span>💳 {t.payment_terms || 'PREPAID'}</span>
                      <span>💲 {t.currency || 'USD'}</span>
                    </div>
                  </div>
                ))}
              </div>

              {activeTmpl && (
                <div className="qt-template-preview-box">
                  <div className="qt-preview-title">
                    <Info size={13} /> Snapshot Preview for "{activeTmpl.name}"
                  </div>
                  <p className="qt-template-preview-text">
                    Applying this template will snapshot copy <strong>{activeTmpl.charge_count || 0} charge items</strong> into your quotation, configure payment terms to <strong>{activeTmpl.payment_terms}</strong>, and set validity for <strong>{activeTmpl.validity_days} days</strong>.
                  </p>
                </div>
              )}

              <div className="qt-form-row qt-form-row--align-center" style={{ marginTop: '8px' }}>
                <label className="qt-checkbox-label">
                  <input
                    type="checkbox"
                    checked={overrideCharges}
                    onChange={e => setOverrideCharges(e.target.checked)}
                  />
                  <span>Replace existing charges on this quotation (Uncheck to append)</span>
                </label>
              </div>
            </>
          )}

          <div className="qt-modal-footer">
            <button type="button" className="qt-btn qt-btn--ghost" onClick={onClose}>Cancel</button>
            <button
              type="button"
              className="qt-btn qt-btn--primary"
              disabled={applying || !selectedId}
              onClick={handleApply}
            >
              {applying ? <><RefreshCw size={14} className="qt-spin" /> Applying…</> : <><Copy size={14} /> Apply Template</>}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Save As Template Modal (Task 18.3) ───────────────────────────────────────

function SaveAsTemplateModal({ quotation, onClose, onSaved }) {
  const [name, setName] = useState(
    quotation ? `${quotation.service_type || 'Standard'} ${quotation.transport_mode || 'Freight'} Template` : ''
  );
  const [description, setDescription] = useState(
    quotation ? `Reusable template based on ${quotation.quotation_number} (${quotation.origin || ''} → ${quotation.destination || ''})` : ''
  );
  const [validityDays, setValidityDays] = useState(30);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSave = async (e) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Template name is required.');
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const res = await quotationService.saveAsTemplate(quotation.id, {
        name: name.trim(),
        description: description.trim(),
        validity_days: Number(validityDays) || 30,
      });
      onSaved(res?.data || res);
    } catch (err) {
      setError(err?.message || 'Failed to save template.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="qt-modal-overlay" onClick={onClose}>
      <div className="qt-modal" onClick={e => e.stopPropagation()}>
        <div className="qt-modal-header">
          <div>
            <h2 className="qt-modal-title">Save as Reusable Template</h2>
            <p className="qt-modal-sub">
              Save current charges, payment terms, and notes as an organizational template
            </p>
          </div>
          <button className="qt-modal-close" onClick={onClose}><X size={20} /></button>
        </div>

        <form className="qt-modal-body" onSubmit={handleSave}>
          {error && <div className="qt-modal-error">{error}</div>}

          <div className="qt-form-group">
            <label>Template Name *</label>
            <input
              placeholder="e.g. Standard Ocean FCL Export"
              value={name}
              onChange={e => setName(e.target.value)}
              required
            />
          </div>

          <div className="qt-form-group">
            <label>Description</label>
            <textarea
              rows={2}
              placeholder="Describe when to use this template..."
              value={description}
              onChange={e => setDescription(e.target.value)}
            />
          </div>

          <div className="qt-form-group">
            <label>Default Validity (Days)</label>
            <input
              type="number"
              min="1"
              max="365"
              value={validityDays}
              onChange={e => setValidityDays(e.target.value)}
            />
          </div>

          <div className="qt-template-preview-box">
            <div className="qt-preview-title">
              <FileCheck size={13} /> Captured Template Contents
            </div>
            <ul className="qt-template-features-list">
              <li>All configured charge items and calculation bases</li>
              <li>Payment terms: <strong>{quotation?.payment_terms || 'PREPAID'}</strong></li>
              <li>Standard commercial terms & conditions</li>
              <li>Customer notes & internal operational notes</li>
            </ul>
          </div>

          <div className="qt-modal-footer">
            <button type="button" className="qt-btn qt-btn--ghost" onClick={onClose}>Cancel</button>
            <button type="submit" className="qt-btn qt-btn--primary" disabled={loading}>
              {loading ? <><RefreshCw size={14} className="qt-spin" /> Saving…</> : <><Bookmark size={14} /> Save Template</>}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ─── Templates Management Drawer (Task 18.3) ──────────────────────────────────

function TemplatesManagementDrawer({ onClose }) {
  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [form, setForm] = useState({
    name: '', description: '', shipment_mode: 'FCL', transport_mode: 'Ocean Freight',
    currency: 'USD', validity_days: 30, payment_terms: 'PREPAID',
    commercial_terms: '', customer_notes: '', internal_notes: '',
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);

  const loadTemplates = useCallback(() => {
    setLoading(true);
    quotationService.listTemplates({ active_only: false })
      .then(res => {
        const d = res?.data || res;
        setTemplates(Array.isArray(d) ? d : []);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    loadTemplates();
  }, [loadTemplates]);

  const handleStartCreate = () => {
    setSelectedTemplate(null);
    setForm({
      name: '', description: '', shipment_mode: 'FCL', transport_mode: 'Ocean Freight',
      currency: 'USD', validity_days: 30, payment_terms: 'PREPAID',
      commercial_terms: '', customer_notes: '', internal_notes: '',
    });
    setIsEditing(true);
  };

  const handleStartEdit = (t) => {
    setSelectedTemplate(t);
    setForm({
      name: t.name || '', description: t.description || '',
      shipment_mode: t.shipment_mode || 'FCL', transport_mode: t.transport_mode || 'Ocean Freight',
      currency: t.currency || 'USD', validity_days: t.validity_days || 30,
      payment_terms: t.payment_terms || 'PREPAID',
      commercial_terms: t.commercial_terms || '',
      customer_notes: t.customer_notes || '',
      internal_notes: t.internal_notes || '',
    });
    setIsEditing(true);
  };

  const handleSave = async (e) => {
    e.preventDefault();
    if (!form.name.trim()) return;
    try {
      setSaving(true);
      setError(null);
      if (selectedTemplate?.id) {
        await quotationService.updateTemplate(selectedTemplate.id, form);
      } else {
        await quotationService.createTemplate(form);
      }
      setIsEditing(false);
      loadTemplates();
    } catch (err) {
      setError(err?.message || 'Failed to save template.');
    } finally {
      setSaving(false);
    }
  };

  const handleArchive = async (id) => {
    if (!window.confirm('Are you sure you want to archive this template?')) return;
    try {
      await quotationService.deleteTemplate(id);
      loadTemplates();
    } catch (err) {
      alert(err?.message || 'Failed to archive template.');
    }
  };

  return (
    <div className="qt-modal-overlay" onClick={onClose}>
      <div className="qt-drawer qt-drawer--wide" onClick={e => e.stopPropagation()}>
        <div className="qt-drawer-header">
          <div>
            <h2 className="qt-drawer-title">Quotation Templates Library</h2>
            <p className="qt-drawer-sub">
              Manage standardized quotation configurations, terms, and charges
            </p>
          </div>
          <div className="qt-header-right">
            {!isEditing && (
              <button className="qt-btn qt-btn--primary qt-btn--sm" onClick={handleStartCreate}>
                <Plus size={13} /> New Template
              </button>
            )}
            <button className="qt-modal-close" onClick={onClose}><X size={20} /></button>
          </div>
        </div>

        <div className="qt-drawer-body">
          {error && <div className="qt-modal-error">{error}</div>}

          {isEditing ? (
            <form className="qt-template-edit-form" onSubmit={handleSave}>
              <div className="qt-form-group">
                <label>Template Name *</label>
                <input
                  placeholder="e.g. Standard Air Freight Import"
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  required
                />
              </div>
              <div className="qt-form-group">
                <label>Description</label>
                <input
                  placeholder="Brief description of this template..."
                  value={form.description}
                  onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                />
              </div>
              <div className="qt-form-row">
                <div className="qt-form-group">
                  <label>Shipment Mode</label>
                  <CustomSelect
                    value={form.shipment_mode}
                    onChange={v => setForm(f => ({ ...f, shipment_mode: v }))}
                    options={SERVICE_TYPE_OPTIONS}
                  />
                </div>
                <div className="qt-form-group">
                  <label>Transport Mode</label>
                  <CustomSelect
                    value={form.transport_mode}
                    onChange={v => setForm(f => ({ ...f, transport_mode: v }))}
                    options={TRANSPORT_MODE_OPTIONS}
                  />
                </div>
                <div className="qt-form-group qt-form-group--sm">
                  <label>Currency</label>
                  <CustomSelect
                    value={form.currency}
                    onChange={v => setForm(f => ({ ...f, currency: v }))}
                    options={CURRENCY_OPTIONS}
                    size="sm"
                  />
                </div>
              </div>
              <div className="qt-form-row">
                <div className="qt-form-group">
                  <label>Payment Terms</label>
                  <CustomSelect
                    value={form.payment_terms}
                    onChange={v => setForm(f => ({ ...f, payment_terms: v }))}
                    options={PAYMENT_TERMS_OPTIONS}
                  />
                </div>
                <div className="qt-form-group qt-form-group--sm">
                  <label>Validity (Days)</label>
                  <input
                    type="number"
                    min="1"
                    value={form.validity_days}
                    onChange={e => setForm(f => ({ ...f, validity_days: e.target.value }))}
                  />
                </div>
              </div>
              <div className="qt-form-group">
                <label>Commercial Terms & Conditions</label>
                <textarea
                  rows={3}
                  placeholder="Standard contractual clauses, carrier GRI caveats..."
                  value={form.commercial_terms}
                  onChange={e => setForm(f => ({ ...f, commercial_terms: e.target.value }))}
                />
              </div>
              <div className="qt-form-group">
                <label>Customer-Facing Notes</label>
                <textarea
                  rows={2}
                  placeholder="Notes shown on customer quotation..."
                  value={form.customer_notes}
                  onChange={e => setForm(f => ({ ...f, customer_notes: e.target.value }))}
                />
              </div>
              <div className="qt-form-group">
                <label>Private Internal Notes (Hidden from customer)</label>
                <textarea
                  rows={2}
                  placeholder="Private internal operational instructions..."
                  value={form.internal_notes}
                  onChange={e => setForm(f => ({ ...f, internal_notes: e.target.value }))}
                />
              </div>

              <div className="qt-modal-footer" style={{ margin: '8px -20px -16px -20px' }}>
                <button type="button" className="qt-btn qt-btn--ghost" onClick={() => setIsEditing(false)}>Cancel</button>
                <button type="submit" className="qt-btn qt-btn--primary" disabled={saving}>
                  {saving ? <><RefreshCw size={14} className="qt-spin" /> Saving…</> : <><Check size={14} /> Save Template</>}
                </button>
              </div>
            </form>
          ) : loading ? (
            <div className="qt-drawer-loading">
              <RefreshCw size={24} className="qt-spin" />
              <p>Loading templates…</p>
            </div>
          ) : templates.length === 0 ? (
            <div className="qt-empty-state" style={{ padding: '40px 20px' }}>
              <div className="qt-empty-icon"><Bookmark size={40} /></div>
              <h3>No templates found</h3>
              <p>Create templates to streamline your quotation workflow.</p>
              <button className="qt-btn qt-btn--primary" onClick={handleStartCreate}>
                <Plus size={14} /> Create First Template
              </button>
            </div>
          ) : (
            <div className="qt-templates-list">
              {templates.map(t => (
                <div key={t.id} className="qt-template-manage-card">
                  <div className="qt-template-manage-header">
                    <div>
                      <div className="qt-template-manage-title">
                        <strong>{t.name}</strong>
                        <span className="qt-badge-mode">{t.shipment_mode || 'FCL'}</span>
                        {!t.is_active && <span className="qt-archived-tag">Archived</span>}
                      </div>
                      {t.description && <div className="qt-template-manage-desc">{t.description}</div>}
                    </div>
                    <div className="qt-row-actions-group">
                      <button className="qt-icon-btn" title="Edit Template" onClick={() => handleStartEdit(t)}>
                        <Edit2 size={14} />
                      </button>
                      {t.is_active && (
                        <button className="qt-icon-btn qt-icon-btn--danger" title="Archive Template" onClick={() => handleArchive(t.id)}>
                          <Trash2 size={14} />
                        </button>
                      )}
                    </div>
                  </div>

                  <div className="qt-template-card-meta">
                    <span>📦 {t.charge_count || 0} charge items</span>
                    <span>📅 {t.validity_days || 30} days validity</span>
                    <span>💳 {t.payment_terms || 'PREPAID'}</span>
                    <span>💲 {t.currency || 'USD'}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Commercial Terms & Notes Tab (Task 18.3) ─────────────────────────────────

function CommercialTermsTab({ quotation, onTermsUpdated }) {
  const [form, setForm] = useState({
    valid_from: quotation?.valid_from ? quotation.valid_from.substring(0, 10) : '',
    valid_until: quotation?.valid_until ? quotation.valid_until.substring(0, 10) : '',
    payment_terms: quotation?.payment_terms || 'PREPAID',
    commercial_terms: quotation?.commercial_terms || '',
    customer_notes: quotation?.customer_notes || quotation?.notes || '',
    internal_notes: quotation?.internal_notes || '',
  });
  const [saving, setSaving] = useState(false);
  const [savedSuccess, setSavedSuccess] = useState(false);
  const [error, setError] = useState(null);

  const isEditable = quotation?.status === 'DRAFT' || quotation?.status === 'CHANGES_REQUESTED';

  useEffect(() => {
    if (quotation) {
      setForm({
        valid_from: quotation.valid_from ? quotation.valid_from.substring(0, 10) : '',
        valid_until: quotation.valid_until ? quotation.valid_until.substring(0, 10) : '',
        payment_terms: quotation.payment_terms || 'PREPAID',
        commercial_terms: quotation.commercial_terms || '',
        customer_notes: quotation.customer_notes || quotation.notes || '',
        internal_notes: quotation.internal_notes || '',
      });
    }
  }, [quotation]);

  const handleSave = async (e) => {
    e.preventDefault();
    try {
      setSaving(true);
      setError(null);
      setSavedSuccess(false);
      const res = await quotationService.updateCommercialTerms(quotation.id, {
        valid_from: form.valid_from || null,
        valid_until: form.valid_until || null,
        payment_terms: form.payment_terms || 'PREPAID',
        commercial_terms: form.commercial_terms || '',
        customer_notes: form.customer_notes || '',
        internal_notes: form.internal_notes || '',
      });
      const updated = res?.data || res;
      setSavedSuccess(true);
      if (onTermsUpdated) onTermsUpdated(updated);
      setTimeout(() => setSavedSuccess(false), 3000);
    } catch (err) {
      setError(err?.message || 'Failed to update commercial terms.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <form className="qt-terms-tab" onSubmit={handleSave}>
      <div className="qt-terms-header">
        <div>
          <h3 className="qt-section-heading">Commercial Terms, Validity & Operational Notes</h3>
          <p className="qt-section-sub">Configure customer contract terms, payment conditions, and private internal instructions.</p>
        </div>
        {isEditable && (
          <button type="submit" className="qt-btn qt-btn--primary qt-btn--sm" disabled={saving}>
            {saving ? <><RefreshCw size={13} className="qt-spin" /> Saving…</> : <><Check size={13} /> Save Terms</>}
          </button>
        )}
      </div>

      {savedSuccess && (
        <div className="qt-success-banner">
          <CheckCircle size={14} /> Commercial terms and validity saved successfully.
        </div>
      )}

      {error && (
        <div className="qt-error-banner">
          <AlertCircle size={14} /> {error}
        </div>
      )}

      {/* Validity & Payment Section */}
      <div className="qt-terms-card">
        <div className="qt-terms-card-title">
          <Calendar size={14} /> Commercial Validity & Payment Method
        </div>
        <div className="qt-form-row">
          <div className="qt-form-group">
            <label>Valid From</label>
            <input
              type="date"
              disabled={!isEditable}
              value={form.valid_from}
              onChange={e => setForm(f => ({ ...f, valid_from: e.target.value }))}
            />
          </div>
          <div className="qt-form-group">
            <label>Valid Until</label>
            <input
              type="date"
              disabled={!isEditable}
              value={form.valid_until}
              onChange={e => setForm(f => ({ ...f, valid_until: e.target.value }))}
            />
          </div>
          <div className="qt-form-group">
            <label>Validity Status</label>
            <div style={{ paddingTop: '6px' }}>
              <ValidityBadge validUntil={form.valid_until} />
            </div>
          </div>
        </div>

        <div className="qt-form-row">
          <div className="qt-form-group">
            <label>Payment Terms</label>
            <CustomSelect
              disabled={!isEditable}
              value={form.payment_terms}
              onChange={v => setForm(f => ({ ...f, payment_terms: v }))}
              options={PAYMENT_TERMS_OPTIONS}
            />
          </div>
        </div>
      </div>

      {/* Customer Commercial Terms */}
      <div className="qt-terms-card">
        <div className="qt-terms-card-title">
          <FileText size={14} /> Standard Commercial Terms & Conditions (Customer-Facing)
        </div>
        <p className="qt-card-desc">Printed on client quotation document. Include demurrage/detention free time, carrier GRI terms, and exclusions.</p>
        <textarea
          rows={3}
          disabled={!isEditable}
          placeholder="e.g. Rate is valid for general cargo only. Free time: 7 days at POD. Subject to carrier space and equipment availability..."
          value={form.commercial_terms}
          onChange={e => setForm(f => ({ ...f, commercial_terms: e.target.value }))}
        />
      </div>

      {/* Customer-Facing Notes */}
      <div className="qt-terms-card">
        <div className="qt-terms-card-title">
          <User size={14} /> Customer-Facing Notes & Greetings
        </div>
        <p className="qt-card-desc">Included in customer proposal email and PDF header.</p>
        <textarea
          rows={2}
          disabled={!isEditable}
          placeholder="e.g. Thank you for requesting a rate. We have secured priority vessel space for your shipment."
          value={form.customer_notes}
          onChange={e => setForm(f => ({ ...f, customer_notes: e.target.value }))}
        />
      </div>

      {/* Private Internal Notes */}
      <div className="qt-terms-card qt-terms-card--internal">
        <div className="qt-terms-card-title" style={{ color: '#92400E' }}>
          <Lock size={14} /> Private Internal Notes (Hidden from Customer)
        </div>
        <div className="qt-internal-lock-banner">
          <Shield size={13} />
          <span>🔒 <strong>Internal only:</strong> These notes are visible strictly to your operations & pricing team. They are never exported or sent to the customer.</span>
        </div>
        <textarea
          rows={3}
          disabled={!isEditable}
          placeholder="e.g. Negotiated 10% sub-agent discount on destination THC. Must assign booking via carrier EDI portal."
          value={form.internal_notes}
          onChange={e => setForm(f => ({ ...f, internal_notes: e.target.value }))}
        />
      </div>
    </form>
  );
}

// ─── Pricing & Charges Workspace Tab ──────────────────────────────────────────

function PricingChargesTab({ quotation, onPricingUpdated, onOpenApplyTemplate, onOpenSaveTemplate }) {
  const [pricing, setPricing] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const [showAddModal, setShowAddModal] = useState(false);
  const [editItem, setEditItem] = useState(null);
  const [showRateDrawer, setShowRateDrawer] = useState(false);
  const [deletingId, setDeletingId] = useState(null);

  const isEditable = quotation?.status === 'DRAFT' || quotation?.status === 'CHANGES_REQUESTED';

  const loadPricing = useCallback(() => {
    if (!quotation?.id) return;
    setLoading(true);
    setError(null);
    quotationService.getQuotationPricing(quotation.id)
      .then(res => {
        const d = res?.data || res;
        setPricing(d);
      })
      .catch(err => {
        setError(err?.message || 'Failed to load pricing details.');
      })
      .finally(() => setLoading(false));
  }, [quotation?.id]);

  useEffect(() => {
    loadPricing();
  }, [loadPricing]);

  const handleDelete = async (chargeId) => {
    if (!window.confirm('Are you sure you want to remove this charge line item?')) return;
    try {
      setDeletingId(chargeId);
      const res = await quotationService.deleteQuotationCharge(quotation.id, chargeId);
      const updatedPricing = res?.data || res;
      setPricing(updatedPricing);
      if (onPricingUpdated) onPricingUpdated(updatedPricing);
    } catch (err) {
      alert(err?.message || 'Failed to delete charge.');
    } finally {
      setDeletingId(null);
    }
  };

  const handleSavedOrImported = (updatedPricing) => {
    setShowAddModal(false);
    setEditItem(null);
    setShowRateDrawer(false);
    setPricing(updatedPricing);
    if (onPricingUpdated) onPricingUpdated(updatedPricing);
  };

  const items = pricing?.charge_items || [];
  const summary = pricing?.summary;

  const grouped = items.reduce((acc, item) => {
    const cat = item.charge_category || 'OTHER';
    if (!acc[cat]) acc[cat] = [];
    acc[cat].push(item);
    return acc;
  }, {});

  return (
    <div className="qt-pricing-tab">
      {/* Action Header */}
      <div className="qt-pricing-actions-header">
        <div className="qt-pricing-actions-left">
          <h3 className="qt-section-heading">Commercial Charges & Pricing Engine</h3>
          {!isEditable && (
            <span className="qt-locked-pill">
              🔒 Locked ({quotation?.status}) — charges can only be modified in DRAFT status
            </span>
          )}
        </div>
        <div className="qt-pricing-actions-right">
          <button
            className="qt-btn qt-btn--ghost qt-btn--sm"
            onClick={loadPricing}
            title="Refresh pricing"
          >
            <RefreshCw size={14} className={loading ? 'qt-spin' : ''} />
          </button>
          {isEditable && (
            <>
              <button
                className="qt-btn qt-btn--ghost qt-btn--sm"
                onClick={onOpenApplyTemplate}
                title="Apply Reusable Template"
              >
                <Bookmark size={14} /> Apply Template
              </button>
              <button
                className="qt-btn qt-btn--secondary qt-btn--sm"
                onClick={() => setShowRateDrawer(true)}
              >
                <Sparkles size={14} /> Import Rate
              </button>
              <button
                className="qt-btn qt-btn--primary qt-btn--sm"
                onClick={() => { setEditItem(null); setShowAddModal(true); }}
              >
                <Plus size={14} /> Add Charge
              </button>
            </>
          )}
        </div>
      </div>

      {/* Multi-currency warning */}
      {summary?.multi_currency_warning && (
        <div className="qt-warning-banner">
          <ShieldAlert size={15} />
          <span>Charges contain mixed currencies without a fixed conversion rate. Consolidated total may differ.</span>
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="qt-error-banner">
          <AlertCircle size={14} /> {error}
          <button onClick={loadPricing}>Retry</button>
        </div>
      )}

      {/* Loading */}
      {loading ? (
        <div className="qt-charges-skeleton">
          {[1,2,3].map(i => <div key={i} className="qt-skel-line" style={{ height: '32px', marginBottom: '8px' }} />)}
        </div>
      ) : items.length === 0 ? (
        <div className="qt-empty-charges">
          <div className="qt-empty-icon"><DollarSign size={40} /></div>
          <h4>No pricing items added yet</h4>
          <p>Build your quotation by adding line-item freight charges, surcharges, or import confirmed carrier rates.</p>
          {isEditable && (
            <div className="qt-empty-charges-btns">
              <button className="qt-btn qt-btn--primary" onClick={() => setShowAddModal(true)}>
                <Plus size={14} /> Add Charge Manually
              </button>
              <button className="qt-btn qt-btn--secondary" onClick={onOpenApplyTemplate}>
                <Bookmark size={14} /> Apply Template
              </button>
              <button className="qt-btn qt-btn--ghost" onClick={() => setShowRateDrawer(true)}>
                <Sparkles size={14} /> Import Rate from Rate Management
              </button>
            </div>
          )}
        </div>
      ) : (
        <div className="qt-charges-table-wrap">
          <table className="qt-charges-table">
            <thead>
              <tr>
                <th>CHARGE ITEM</th>
                <th>BASIS</th>
                <th style={{ textAlign: 'right' }}>QTY</th>
                <th style={{ textAlign: 'right' }}>UNIT SELL</th>
                <th style={{ textAlign: 'right' }}>DISC / TAX</th>
                <th style={{ textAlign: 'right' }}>SELL AMOUNT</th>
                <th style={{ textAlign: 'right' }}>UNIT COST</th>
                <th style={{ textAlign: 'center' }}>MARGIN</th>
                {isEditable && <th style={{ textAlign: 'center' }}>ACTIONS</th>}
              </tr>
            </thead>
            <tbody>
              {Object.keys(grouped).map(catKey => (
                <React.Fragment key={catKey}>
                  <tr className="qt-cat-header-row">
                    <td colSpan={isEditable ? 9 : 8}>
                      <span className="qt-cat-tag">
                        {CHARGE_CATEGORIES.find(c => c.value === catKey)?.label || catKey}
                      </span>
                    </td>
                  </tr>
                  {grouped[catKey].map(item => {
                    const profit = item.total_sell - item.total_cost;
                    const margin = item.total_sell > 0 ? (profit / item.total_sell) * 100 : 0;
                    return (
                      <tr key={item.id} className="qt-charge-row">
                        <td>
                          <div className="qt-charge-name">
                            {item.charge_name}
                            {item.is_optional && <span className="qt-optional-badge">Optional</span>}
                          </div>
                          <div className="qt-charge-code">{item.charge_code} {item.notes ? `• ${item.notes}` : ''}</div>
                        </td>
                        <td>
                          <span className="qt-basis-badge">{item.calculation_basis}</span>
                        </td>
                        <td style={{ textAlign: 'right', fontWeight: 500 }}>
                          {item.quantity}
                        </td>
                        <td style={{ textAlign: 'right' }}>
                          {fmtAmount(item.unit_price, item.currency)}
                        </td>
                        <td style={{ textAlign: 'right', fontSize: '12px' }}>
                          {item.discount_amount > 0 && (
                            <div style={{ color: '#DC2626' }}>-{fmtAmount(item.discount_amount, item.currency)}</div>
                          )}
                          {item.tax_amount > 0 && (
                            <div style={{ color: '#64748B' }}>+{fmtAmount(item.tax_amount, item.currency)} tax</div>
                          )}
                          {!item.discount_amount && !item.tax_amount && '—'}
                        </td>
                        <td style={{ textAlign: 'right', fontWeight: 600, color: '#0F172A' }}>
                          {fmtAmount(item.total_sell, item.currency)}
                        </td>
                        <td style={{ textAlign: 'right', color: '#64748B' }}>
                          {item.cost_amount > 0 ? fmtAmount(item.cost_amount, item.currency) : '—'}
                        </td>
                        <td style={{ textAlign: 'center' }}>
                          {item.cost_amount > 0 ? (
                            <MarginBadge health={margin < 0 ? 'NEGATIVE' : margin < 15 ? 'LOW' : 'HEALTHY'} marginPct={margin} />
                          ) : (
                            <span className="qt-cell-empty">—</span>
                          )}
                        </td>
                        {isEditable && (
                          <td style={{ textAlign: 'center' }}>
                            <div className="qt-row-actions-group">
                              <button
                                className="qt-icon-btn"
                                title="Edit Charge"
                                onClick={() => { setEditItem(item); setShowAddModal(true); }}
                              >
                                <Edit2 size={13} />
                              </button>
                              <button
                                className="qt-icon-btn qt-icon-btn--danger"
                                title="Delete Charge"
                                disabled={deletingId === item.id}
                                onClick={() => handleDelete(item.id)}
                              >
                                <Trash2 size={13} />
                              </button>
                            </div>
                          </td>
                        )}
                      </tr>
                    );
                  })}
                </React.Fragment>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Pricing Summary Card */}
      {summary && (
        <div className="qt-pricing-summary-card">
          <div className="qt-summary-col">
            <h4 className="qt-summary-title">Customer Commercial Pricing</h4>
            <div className="qt-sum-row">
              <span>Freight Base</span>
              <span>{fmtAmount(summary.freight_amount, summary.currency)}</span>
            </div>
            {summary.origin_charges > 0 && (
              <div className="qt-sum-row">
                <span>Origin Handling / THC</span>
                <span>{fmtAmount(summary.origin_charges, summary.currency)}</span>
              </div>
            )}
            {summary.destination_charges > 0 && (
              <div className="qt-sum-row">
                <span>Destination Handling / DTHC</span>
                <span>{fmtAmount(summary.destination_charges, summary.currency)}</span>
              </div>
            )}
            {summary.surcharges > 0 && (
              <div className="qt-sum-row">
                <span>Surcharges (BAF/CAF/PSS)</span>
                <span>{fmtAmount(summary.surcharges, summary.currency)}</span>
              </div>
            )}
            {summary.documentation_charges > 0 && (
              <div className="qt-sum-row">
                <span>Documentation</span>
                <span>{fmtAmount(summary.documentation_charges, summary.currency)}</span>
              </div>
            )}
            {summary.customs_charges > 0 && (
              <div className="qt-sum-row">
                <span>Customs Clearance</span>
                <span>{fmtAmount(summary.customs_charges, summary.currency)}</span>
              </div>
            )}
            {summary.insurance_charges > 0 && (
              <div className="qt-sum-row">
                <span>Marine Cargo Insurance</span>
                <span>{fmtAmount(summary.insurance_charges, summary.currency)}</span>
              </div>
            )}
            {summary.other_charges > 0 && (
              <div className="qt-sum-row">
                <span>Other Line Items</span>
                <span>{fmtAmount(summary.other_charges, summary.currency)}</span>
              </div>
            )}
            {summary.discounts > 0 && (
              <div className="qt-sum-row" style={{ color: '#DC2626' }}>
                <span>Discounts Applied</span>
                <span>-{fmtAmount(summary.discounts, summary.currency)}</span>
              </div>
            )}
            <div className="qt-sum-row">
              <span>Subtotal (pre-tax)</span>
              <span>{fmtAmount(summary.subtotal, summary.currency)}</span>
            </div>
            <div className="qt-sum-row">
              <span>Taxes & Duties</span>
              <span>{fmtAmount(summary.taxes, summary.currency)}</span>
            </div>
            <div className="qt-sum-row qt-sum-total">
              <span>Total Quotation Amount</span>
              <span>{fmtAmount(summary.total_amount, summary.currency)}</span>
            </div>
          </div>

          <div className="qt-summary-col qt-summary-col--internal">
            <h4 className="qt-summary-title">Internal Commercial Margin</h4>
            <div className="qt-sum-row">
              <span>Total Buy Cost</span>
              <span>{fmtAmount(summary.total_cost, summary.currency)}</span>
            </div>
            <div className="qt-sum-row">
              <span>Total Sell Revenue</span>
              <span>{fmtAmount(summary.total_amount, summary.currency)}</span>
            </div>
            <div className="qt-sum-row qt-sum-total">
              <span>Gross Profit</span>
              <span style={{ color: summary.gross_profit >= 0 ? '#15803D' : '#DC2626' }}>
                {fmtAmount(summary.gross_profit, summary.currency)}
              </span>
            </div>
            <div className="qt-margin-health-row">
              <span>Commercial Margin Health:</span>
              <MarginBadge health={summary.margin_health} marginPct={summary.gross_margin_percentage} />
            </div>
          </div>
        </div>
      )}

      {/* Add / Edit Modal */}
      {showAddModal && (
        <AddEditChargeModal
          quotation={quotation}
          initialData={editItem}
          onClose={() => { setShowAddModal(false); setEditItem(null); }}
          onSaved={handleSavedOrImported}
        />
      )}

      {/* Rate Candidates Drawer */}
      {showRateDrawer && (
        <RateCandidatesDrawer
          quotation={quotation}
          onClose={() => setShowRateDrawer(false)}
          onImported={handleSavedOrImported}
        />
      )}
    </div>
  );
}

// ─── New Quotation Modal ──────────────────────────────────────────────────────

function NewQuotationModal({ onClose, onCreated }) {
  const [form, setForm] = useState({
    origin: '', origin_code: '', destination: '', destination_code: '',
    service_type: 'FCL', transport_mode: 'Ocean Freight',
    currency: 'USD', payment_terms: 'PREPAID',
    valid_from: '', valid_until: '', notes: '',
    template_id: '',
  });
  const [templates, setTemplates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    quotationService.listTemplates({ active_only: true })
      .then(res => {
        const d = res?.data || res;
        setTemplates(Array.isArray(d) ? d : []);
      })
      .catch(console.error);
  }, []);

  const set = (k, v) => setForm(f => ({ ...f, [k]: v }));

  const handleTemplateSelect = (templateIdStr) => {
    set('template_id', templateIdStr);
    if (!templateIdStr) return;
    const t = templates.find(item => String(item.id) === templateIdStr);
    if (t) {
      if (t.shipment_mode) set('service_type', t.shipment_mode);
      if (t.transport_mode) set('transport_mode', t.transport_mode);
      if (t.currency) set('currency', t.currency);
      if (t.payment_terms) set('payment_terms', t.payment_terms);
      if (t.origin) set('origin', t.origin);
      if (t.destination) set('destination', t.destination);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!form.origin || !form.destination) {
      setError('Origin and destination are required.');
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const payload = {
        origin: form.origin, origin_code: form.origin_code,
        destination: form.destination, destination_code: form.destination_code,
        service_type: form.service_type, transport_mode: form.transport_mode,
        currency: form.currency, payment_terms: form.payment_terms,
        notes: form.notes,
        ...(form.valid_from && { valid_from: form.valid_from }),
        ...(form.valid_until && { valid_until: form.valid_until }),
        ...(form.template_id && { template_id: Number(form.template_id) }),
      };
      const res = await quotationService.createQuotation(payload);
      const created = res?.data || res;
      onCreated(created);
    } catch (err) {
      setError(err?.message || 'Failed to create quotation.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="qt-modal-overlay" onClick={onClose}>
      <div className="qt-modal" onClick={e => e.stopPropagation()}>
        <div className="qt-modal-header">
          <div>
            <h2 className="qt-modal-title">New Quotation</h2>
            <p className="qt-modal-sub">Create a new customer quotation or initialize from a template</p>
          </div>
          <button className="qt-modal-close" onClick={onClose}><X size={20} /></button>
        </div>

        <form className="qt-modal-body" onSubmit={handleSubmit}>
          {error && <div className="qt-modal-error">{error}</div>}

          {/* Template Selection */}
          {templates.length > 0 && (
            <div className="qt-form-card">
              <div className="qt-form-card-title">
                <Bookmark size={14} className="qt-text-primary" /> Template & Baseline Configuration
              </div>
              <div className="qt-form-group">
                <label>Start from Template (Optional)</label>
                <CustomSelect
                  value={form.template_id}
                  onChange={v => handleTemplateSelect(v)}
                  placeholder="-- Blank Quotation (Start from scratch) --"
                  options={[
                    { value: '', label: 'Blank Quotation', sublabel: 'Start from scratch without pre-filled charges' },
                    ...templates.map(t => ({
                      value: String(t.id),
                      label: t.name,
                      sublabel: `${t.shipment_mode || 'FCL'} · ${t.currency || 'USD'} · ${t.charge_count || 0} charge items · ${t.validity_days || 30}d validity`,
                      badge: t.is_default ? 'Default' : undefined,
                    }))
                  ]}
                />
              </div>
            </div>
          )}

          {/* Route & Transport Section */}
          <div className="qt-form-card">
            <div className="qt-form-card-title">
              <MapPin size={14} className="qt-text-primary" /> Route & Transport Details
            </div>
            <div className="qt-form-row">
              <div className="qt-form-group" style={{ flex: '2 1 0' }}>
                <label>Origin Port / City *</label>
                <input placeholder="e.g. Nhava Sheva (INNSA)" value={form.origin} onChange={e => set('origin', e.target.value)} required />
              </div>
              <div className="qt-form-group" style={{ flex: '1 1 0' }}>
                <label>Origin Code</label>
                <input placeholder="INNSA" value={form.origin_code} onChange={e => set('origin_code', e.target.value)} />
              </div>
            </div>
            <div className="qt-form-row">
              <div className="qt-form-group" style={{ flex: '2 1 0' }}>
                <label>Destination Port / City *</label>
                <input placeholder="e.g. Rotterdam (NLRTM)" value={form.destination} onChange={e => set('destination', e.target.value)} required />
              </div>
              <div className="qt-form-group" style={{ flex: '1 1 0' }}>
                <label>Destination Code</label>
                <input placeholder="NLRTM" value={form.destination_code} onChange={e => set('destination_code', e.target.value)} />
              </div>
            </div>
            <div className="qt-form-row">
              <div className="qt-form-group">
                <label>Service Type</label>
                <CustomSelect
                  value={form.service_type}
                  onChange={v => set('service_type', v)}
                  options={SERVICE_TYPE_OPTIONS}
                />
              </div>
              <div className="qt-form-group">
                <label>Transport Mode</label>
                <CustomSelect
                  value={form.transport_mode}
                  onChange={v => set('transport_mode', v)}
                  options={TRANSPORT_MODE_OPTIONS}
                />
              </div>
            </div>
          </div>

          {/* Commercial & Financial Conditions */}
          <div className="qt-form-card">
            <div className="qt-form-card-title">
              <DollarSign size={14} className="qt-text-primary" /> Commercial & Billing Terms
            </div>
            <div className="qt-form-row">
              <div className="qt-form-group">
                <label>Currency</label>
                <CustomSelect
                  value={form.currency}
                  onChange={v => set('currency', v)}
                  options={CURRENCY_OPTIONS}
                />
              </div>
              <div className="qt-form-group">
                <label>Payment Terms</label>
                <CustomSelect
                  value={form.payment_terms}
                  onChange={v => set('payment_terms', v)}
                  options={PAYMENT_TERMS_OPTIONS}
                />
              </div>
              <div className="qt-form-group">
                <label>Valid Until</label>
                <input type="date" value={form.valid_until} onChange={e => set('valid_until', e.target.value)} />
              </div>
            </div>
            <div className="qt-form-group">
              <label>Internal Notes (Optional)</label>
              <textarea rows={2} placeholder="Add operational caveats or customer notes..." value={form.notes} onChange={e => set('notes', e.target.value)} />
            </div>
          </div>

          <div className="qt-modal-footer">
            <button type="button" className="qt-btn qt-btn--ghost" onClick={onClose}>Cancel</button>
            <button type="submit" className="qt-btn qt-btn--primary" disabled={loading}>
              {loading ? <><RefreshCw size={14} className="qt-spin" /> Creating…</> : <><Plus size={14} /> Create Quotation</>}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ─── Detail Panel ─────────────────────────────────────────────────────────────

function QuotationDetailPanel({ quotationId, onClose, onQuotationUpdated, onOpenTemplatesManager }) {
  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('pricing');

  const [showApplyTmplModal, setShowApplyTmplModal] = useState(false);
  const [showSaveTmplModal, setShowSaveTmplModal] = useState(false);
  const [showConvertModal, setShowConvertModal] = useState(false);
  const [conversionSuccessResult, setConversionSuccessResult] = useState(null);

  const fetchDetail = useCallback(() => {
    if (!quotationId) return;
    setLoading(true);
    quotationService.getQuotation(quotationId)
      .then(res => {
        const d = res?.data || res;
        setDetail(d);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [quotationId]);

  const handleConversionSuccess = (result) => {
    setShowConvertModal(false);
    setConversionSuccessResult(result);
    fetchDetail();
    if (onQuotationUpdated) onQuotationUpdated();
  };

  useEffect(() => {
    fetchDetail();
  }, [fetchDetail]);

  if (!quotationId) return null;

  const q = detail?.quotation || (detail?.id ? detail : null);
  const customer = detail?.customer || (detail?.customer_name ? { name: detail.customer_name } : null);
  const activity = detail?.activity || [];

  const exp = q?.valid_until ? expiryLabel(q.valid_until) : null;
  const isDraft = q?.status === 'DRAFT' || !q?.status;

  const handlePricingUpdated = (updatedPricing) => {
    setDetail(prev => ({
      ...prev,
      pricing: updatedPricing,
      quotation: {
        ...prev.quotation,
        total_amount: updatedPricing.summary?.total_amount ?? prev.quotation.total_amount,
        subtotal: updatedPricing.summary?.subtotal ?? prev.quotation.subtotal,
        surcharges: updatedPricing.summary?.surcharges ?? prev.quotation.surcharges,
        taxes: updatedPricing.summary?.taxes ?? prev.quotation.taxes,
        total_cost: updatedPricing.summary?.total_cost ?? prev.quotation.total_cost,
        gross_profit: updatedPricing.summary?.gross_profit ?? prev.quotation.gross_profit,
        gross_margin_pct: updatedPricing.summary?.gross_margin_percentage ?? prev.quotation.gross_margin_pct,
      }
    }));
    if (onQuotationUpdated) onQuotationUpdated();
  };

  const handleTermsUpdated = (updatedDetail) => {
    setDetail(updatedDetail);
    if (onQuotationUpdated) onQuotationUpdated();
  };

  const handleTemplateApplied = (updatedPricing) => {
    setShowApplyTmplModal(false);
    fetchDetail();
    if (onQuotationUpdated) onQuotationUpdated();
  };

  return (
    <div className="qt-detail-panel">
      <div className="qt-detail-header">
        <div className="qt-detail-title-row">
          <div>
            <h2 className="qt-detail-number">{loading ? '…' : (q?.quotation_number || '—')}</h2>
            {exp && !loading && (
              <p className="qt-detail-validity">
                Valid Till: {fmtDate(q?.valid_until)}&nbsp;
                <span style={{ color: exp.color }}>({exp.text})</span>
              </p>
            )}
          </div>
          <div className="qt-detail-header-right">
            {q?.status && <StatusBadge status={q.status} />}
            <button className="qt-detail-close" onClick={onClose}><X size={16} /></button>
          </div>
        </div>

        {/* Quick action bar */}
        <div className="qt-detail-actions-bar">
          {isDraft && (
            <button className="qt-action-btn qt-action-btn--highlight" onClick={() => setShowApplyTmplModal(true)}>
              <Bookmark size={14} />
              <span>Apply Template</span>
            </button>
          )}
          {q?.status === 'ACCEPTED' && !q?.converted_booking_id && (
            <button className="qt-action-btn qt-action-btn--highlight" style={{ background: '#15803D', color: '#FFF' }} onClick={() => setShowConvertModal(true)}>
              <Briefcase size={14} />
              <span>Convert to Booking</span>
            </button>
          )}
          <button className="qt-action-btn" onClick={() => setShowSaveTmplModal(true)}>
            <Copy size={14} />
            <span>Save as Template</span>
          </button>
          {[
            { icon: Eye,      label: 'View' },
            { icon: Send,     label: 'Send' },
            { icon: Download, label: 'Download' },
          ].map(({ icon: Icon, label }) => (
            <button key={label} className="qt-action-btn">
              <Icon size={14} />
              <span>{label}</span>
            </button>
          ))}
        </div>

        {/* Tabs */}
        <div className="qt-detail-tabs">
          {[
            { id: 'pricing',  label: 'Pricing & Charges' },
            { id: 'terms',    label: 'Commercial Terms' },
            { id: 'handover', label: q?.status === 'ACCEPTED' || q?.converted_booking_id ? 'Commercial Handover ⚡' : 'Handover' },
            { id: 'overview', label: 'Overview' },
            { id: 'activity', label: 'Activity' },
          ].map(t => (
            <button
              key={t.id}
              className={`qt-detail-tab ${activeTab === t.id ? 'active' : ''}`}
              onClick={() => setActiveTab(t.id)}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>

      <div className="qt-detail-body">
        {loading ? (
          <div className="qt-detail-skeleton">
            {[80, 60, 90, 50, 70].map((w, i) => (
              <div key={i} className="qt-skel-line" style={{ width: `${w}%` }} />
            ))}
          </div>
        ) : activeTab === 'pricing' ? (
          <PricingChargesTab
            quotation={q}
            onPricingUpdated={handlePricingUpdated}
            onOpenApplyTemplate={() => setShowApplyTmplModal(true)}
            onOpenSaveTemplate={() => setShowSaveTmplModal(true)}
          />
        ) : activeTab === 'terms' ? (
          <CommercialTermsTab
            quotation={q}
            onTermsUpdated={handleTermsUpdated}
          />
        ) : activeTab === 'handover' ? (
          <CommercialHandoverTab
            quotation={q}
            onConverted={handleConversionSuccess}
            onOpenConvertModal={() => setShowConvertModal(true)}
          />
        ) : activeTab === 'overview' ? (
          <>
            {/* Customer */}
            {customer && (
              <div className="qt-detail-section">
                <div className="qt-detail-section-title">
                  <User size={14} />
                  Customer
                </div>
                <div className="qt-detail-card">
                  <div className="qt-cust-name-row">
                    <span className="qt-cust-avatar">{customer.name?.[0] || '?'}</span>
                    <div>
                      <div className="qt-cust-name">{customer.name}</div>
                      <div className="qt-cust-code">{customer.customer_code}</div>
                    </div>
                  </div>
                  {customer.contact_phone && (
                    <div className="qt-cust-row">📞 {customer.contact_phone}</div>
                  )}
                  {customer.contact_email && (
                    <div className="qt-cust-row">✉️ {customer.contact_email}</div>
                  )}
                </div>
              </div>
            )}

            {/* Route / Service */}
            {(q?.origin || q?.destination) && (
              <div className="qt-detail-section">
                <div className="qt-detail-section-title">
                  <MapPin size={14} />
                  Route / Service
                </div>
                <div className="qt-detail-card">
                  <div className="qt-route-display">
                    <span className="qt-route-port">
                      {q.origin}{q.origin_code ? ` (${q.origin_code})` : ''}
                    </span>
                    <ArrowRight size={14} className="qt-route-arrow" />
                    <span className="qt-route-port">
                      {q.destination}{q.destination_code ? ` (${q.destination_code})` : ''}
                    </span>
                  </div>
                  {q.service_type && (
                    <div className="qt-route-service">
                      {q.service_type} • {q.transport_mode}
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Quotation Summary */}
            <div className="qt-detail-section">
              <div className="qt-detail-section-title">
                <DollarSign size={14} />
                Quotation Summary
              </div>
              <div className="qt-detail-card">
                <div className="qt-summary-row">
                  <span>Subtotal</span>
                  <span>{fmtAmount(q?.subtotal, q?.currency)}</span>
                </div>
                <div className="qt-summary-row">
                  <span>Surcharges</span>
                  <span>{fmtAmount(q?.surcharges, q?.currency)}</span>
                </div>
                <div className="qt-summary-row">
                  <span>Taxes</span>
                  <span>{fmtAmount(q?.taxes, q?.currency)}</span>
                </div>
                <div className="qt-summary-row qt-summary-total">
                  <span>Total Amount</span>
                  <span>{fmtAmount(q?.total_amount, q?.currency)}</span>
                </div>
                {q?.total_cost > 0 && (
                  <div className="qt-summary-row" style={{ marginTop: '8px', paddingTop: '8px', borderTop: '1px dashed #E2E8F0' }}>
                    <span>Gross Margin</span>
                    <MarginBadge health={q.gross_margin_pct < 15 ? 'LOW' : 'HEALTHY'} marginPct={q.gross_margin_pct} />
                  </div>
                )}
                <div className="qt-summary-currency">Currency: {q?.currency || 'USD'} • Payment: {q?.payment_terms || 'PREPAID'}</div>
              </div>
            </div>

            {/* Status Timeline */}
            {activity.length > 0 && (
              <div className="qt-detail-section">
                <div className="qt-detail-section-title">
                  <Clock size={14} />
                  Status Timeline
                </div>
                <div className="qt-timeline">
                  {activity.map((act, idx) => (
                    <div key={act.id || idx} className="qt-timeline-item">
                      <div className="qt-timeline-dot" />
                      <div className="qt-timeline-content">
                        <div className="qt-timeline-event">{act.description || act.activity_type}</div>
                        <div className="qt-timeline-meta">
                          {fmtDate(act.created_at)}{act.actor ? ` by ${act.actor}` : ''}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </>
        ) : (
          <div className="qt-timeline">
            {activity.length === 0 ? (
              <div className="qt-empty-small">No activity recorded.</div>
            ) : activity.map((act, idx) => (
              <div key={act.id || idx} className="qt-timeline-item">
                <div className="qt-timeline-dot" />
                <div className="qt-timeline-content">
                  <div className="qt-timeline-event">{act.description || act.activity_type}</div>
                  <div className="qt-timeline-meta">
                    {fmtDate(act.created_at)}{act.actor ? ` by ${act.actor}` : ''}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Apply Template Modal */}
      {showApplyTmplModal && (
        <ApplyTemplateModal
          quotation={q}
          onClose={() => setShowApplyTmplModal(false)}
          onApplied={handleTemplateApplied}
        />
      )}

      {/* Save as Template Modal */}
      {showSaveTmplModal && (
        <SaveAsTemplateModal
          quotation={q}
          onClose={() => setShowSaveTmplModal(false)}
          onSaved={() => { setShowSaveTmplModal(false); fetchDetail(); }}
        />
      )}

      {/* Convert to Booking Modal */}
      {showConvertModal && (
        <ConvertToBookingModal
          quotation={q}
          onClose={() => setShowConvertModal(false)}
          onConverted={handleConversionSuccess}
        />
      )}

      {/* Conversion Success Modal */}
      {conversionSuccessResult && (
        <ConversionSuccessModal
          result={conversionSuccessResult}
          onClose={() => {
            setConversionSuccessResult(null);
            fetchDetail();
          }}
        />
      )}
    </div>
  );
}

// ─── Commercial Handover & Operations Traceability Workspace (Task 18.7) ──────

function CommercialHandoverTab({ quotation, onConverted, onOpenConvertModal }) {
  const [handoverData, setHandoverData] = useState(null);
  const [preview, setPreview] = useState(null);
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [confirming, setConfirming] = useState(false);
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [confirmNotes, setConfirmNotes] = useState('');

  const fetchHandoverData = useCallback(() => {
    if (!quotation?.id) return;
    setLoading(true);
    setError(null);
    Promise.all([
      quotationService.getOperationalHandover(quotation.id).catch(err => {
        console.warn('Operational handover warning:', err);
        return null;
      }),
      quotationService.getConversionPreview(quotation.id).catch(err => {
        console.warn('Preview load warning:', err);
        return null;
      }),
      quotationService.getConversionHistory(quotation.id).catch(err => {
        console.warn('History load warning:', err);
        return [];
      }),
    ]).then(([handoverRes, prevRes, histRes]) => {
      if (handoverRes) {
        setHandoverData(handoverRes?.data || handoverRes);
      }
      if (prevRes) {
        setPreview(prevRes?.data || prevRes);
      }
      const histData = histRes?.data || histRes;
      setHistory(Array.isArray(histData) ? histData : []);
    }).catch(err => {
      const errMsg = err?.response?.data?.error?.message || err?.message || 'Failed to load handover details.';
      setError(errMsg);
    }).finally(() => setLoading(false));
  }, [quotation?.id]);

  useEffect(() => {
    fetchHandoverData();
  }, [fetchHandoverData]);

  const handleConfirmHandover = async () => {
    try {
      setConfirming(true);
      await quotationService.confirmOperationalHandover(quotation.id, {
        confirmation_notes: confirmNotes,
        notify_operations: true,
      });
      setShowConfirmModal(false);
      fetchHandoverData();
      if (onConverted) onConverted();
    } catch (err) {
      alert(err?.response?.data?.error?.message || err?.message || 'Failed to confirm operational handover.');
    } finally {
      setConfirming(false);
    }
  };

  const isAccepted = quotation?.status === 'ACCEPTED';
  const isConverted = quotation?.conversion_status === 'CONVERTED' || !!quotation?.converted_booking_id || !!handoverData?.booking_id;
  const canConvert = preview?.can_convert && isAccepted && !isConverted;

  const lineage = handoverData?.lineage_chain || [];
  const changes = handoverData?.operational_changes || [];
  const snapshot = handoverData?.commercial_snapshot || preview;
  const isHandoverConfirmed = handoverData?.handover_status === 'BOOKING_CONFIRMED' || handoverData?.handover_status === 'COMPLETED';

  return (
    <div className="qt-handover-tab">
      {/* Header */}
      <div className="qt-handover-header">
        <div>
          <h3 className="qt-section-heading">Operational Lineage & Handover Workspace</h3>
          <p className="qt-section-sub">Trace end-to-end chain from Accepted Quote → Booking → Shipment → Tracking with full commercial snapshot integrity.</p>
        </div>
        <button className="qt-btn qt-btn--ghost qt-btn--sm" onClick={fetchHandoverData} title="Refresh Status">
          <RefreshCw size={13} className={loading ? 'qt-spin' : ''} />
        </button>
      </div>

      {error && (
        <div className="qt-error-banner">
          <AlertCircle size={14} /> {error}
        </div>
      )}

      {/* ── Interactive Rate Selection & Commercial Snapshot ── */}
      <RateSelectionSection
        quotation={quotation}
        canEdit={['DRAFT', 'CHANGES_REQUESTED'].includes(quotation?.status)}
        onRateUpdated={fetchHandoverData}
      />

      {/* ── Interactive Lineage Chain (Quote → Booking → Shipment → Tracking) ── */}
      <div className="qt-lineage-card">
        <div className="qt-lineage-header">
          <div className="qt-lineage-title">
            <GitBranch size={15} /> End-to-End Operational Lineage
          </div>
          {handoverData?.handover_status && (
            <span className={`qt-handover-status-tag ${handoverData.handover_status.toLowerCase()}`}>
              Status: {handoverData.handover_status.replace(/_/g, ' ')}
            </span>
          )}
        </div>

        <div className="qt-lineage-bar">
          {lineage.length > 0 ? (
            lineage.map((step, idx) => {
              const isCompleted = step.status === 'COMPLETED';
              const isActive = step.status === 'ACTIVE';
              return (
                <div key={step.step_id || idx} className={`qt-lineage-node ${isCompleted ? 'completed' : isActive ? 'active' : 'pending'}`}>
                  <div className="qt-node-badge">
                    {isCompleted ? <CheckCircle size={14} /> : isActive ? <span className="qt-node-pulse" /> : <span className="qt-node-dot" />}
                  </div>
                  <div className="qt-node-body">
                    <div className="qt-node-name">{step.name}</div>
                    <div className="qt-node-ref">
                      {step.url ? (
                        <a href={step.url} className="qt-node-link">
                          {step.ref_number || step.status} <ArrowUpRight size={10} />
                        </a>
                      ) : (
                        <span>{step.ref_number || step.status}</span>
                      )}
                    </div>
                  </div>
                  {idx < lineage.length - 1 && <div className="qt-node-connector" />}
                </div>
              );
            })
          ) : (
            <div className="qt-lineage-node-list">
              {[
                { name: 'Quotation', ref: quotation?.quotation_number || 'DRAFT', status: 'COMPLETED' },
                { name: 'Customer Acceptance', ref: isAccepted ? 'ACCEPTED' : 'PENDING', status: isAccepted ? 'COMPLETED' : 'ACTIVE' },
                { name: 'Operational Booking', ref: isConverted ? 'CONVERTED' : 'PENDING', status: isConverted ? 'COMPLETED' : 'PENDING' },
                { name: 'Active Shipment', ref: isHandoverConfirmed ? 'CONFIRMED' : 'PENDING', status: isHandoverConfirmed ? 'COMPLETED' : 'PENDING' },
                { name: 'Milestone Tracking', ref: 'PENDING', status: 'PENDING' },
              ].map((step, idx, arr) => {
                const isCompleted = step.status === 'COMPLETED';
                const isActive = step.status === 'ACTIVE';
                return (
                  <div key={idx} className={`qt-lineage-node ${isCompleted ? 'completed' : isActive ? 'active' : 'pending'}`}>
                    <div className="qt-node-badge">
                      {isCompleted ? <CheckCircle size={14} /> : isActive ? <span className="qt-node-pulse" /> : <span className="qt-node-dot" />}
                    </div>
                    <div className="qt-node-body">
                      <div className="qt-node-name">{step.name}</div>
                      <div className="qt-node-ref">
                        <span>{step.ref}</span>
                      </div>
                    </div>
                    {idx < arr.length - 1 && <div className="qt-node-connector" />}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* ── Conversion & Handover Status Controls ── */}
      {isConverted ? (
        <div className="qt-converted-banner">
          <div className="qt-converted-left">
            <div className="qt-converted-icon"><CheckCircle2 size={24} /></div>
            <div>
              <div className="qt-converted-title">
                ✓ Converted to Operational Booking {handoverData?.booking_number || quotation.converted_booking_id}
              </div>
              <div className="qt-converted-sub">
                {isHandoverConfirmed ? (
                  <span style={{ color: '#15803D', fontWeight: 600 }}>
                    ⚡ Operational handover confirmed by {handoverData?.booking_confirmed_by || 'Operations'} on {fmtDate(handoverData?.booking_confirmed_at)}.
                  </span>
                ) : (
                  <span>
                    Booking created. Commercial terms are locked. Operator confirmation is pending.
                  </span>
                )}
              </div>
            </div>
          </div>
          <div className="qt-converted-actions">
            {!isHandoverConfirmed && handoverData?.can_confirm_handover && (
              <button className="qt-btn qt-btn--primary qt-btn--sm" onClick={() => setShowConfirmModal(true)}>
                <ShieldCheck size={14} /> Confirm Handover
              </button>
            )}
            <a href="/dashboard/bookings" className="qt-btn qt-btn--secondary qt-btn--sm">
              <ArrowUpRight size={13} /> Open Bookings
            </a>
            {(quotation.converted_shipment_id || handoverData?.shipment_id) && (
              <a href="/dashboard/shipments" className="qt-btn qt-btn--ghost qt-btn--sm">
                <ArrowUpRight size={13} /> Open Shipments
              </a>
            )}
          </div>
        </div>
      ) : !isAccepted ? (
        <div className="qt-blocked-banner">
          <div className="qt-blocked-header">
            <Lock size={16} />
            <strong>Commercial Acceptance Required Before Operational Conversion</strong>
          </div>
          <p>
            Current Status: <strong>{quotation?.status}</strong>. Only quotations in <strong>ACCEPTED</strong> status are eligible for operational booking conversion.
          </p>
        </div>
      ) : (
        <div className="qt-readiness-banner">
          <div className="qt-readiness-top">
            <div className="qt-readiness-badge">
              <span className="qt-status-dot" style={{ background: '#22C55E' }} />
              ✓ QUOTATION ACCEPTED — {preview?.can_convert ? 'READY FOR OPERATIONAL HANDOVER' : 'MISSING OPERATIONAL REQUIREMENTS'}
            </div>
            {canConvert && (
              <button className="qt-btn qt-btn--primary qt-btn--sm" onClick={onOpenConvertModal}>
                <Briefcase size={14} /> Convert to Booking
              </button>
            )}
          </div>

          {preview?.blocking_reasons?.length > 0 && (
            <div className="qt-blocking-reasons-box">
              <div className="qt-blocking-title">Before creating a booking:</div>
              <ul className="qt-blocking-list">
                {preview.blocking_reasons.map((reason, idx) => (
                  <li key={idx}>• {reason}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {/* ── Operational Drift & Changes Card ── */}
      {isConverted && (
        <div className="qt-handover-card">
          <div className="qt-handover-card-header">
            <Sliders size={15} /> Operational Details vs Commercial Quote Baseline
          </div>
          {changes.length === 0 ? (
            <div className="qt-no-changes-box">
              <ShieldCheck size={16} color="#16A34A" />
              <span>All operational parameters align perfectly with the original commercial quotation baseline. No operational drift detected.</span>
            </div>
          ) : (
            <div className="qt-changes-table-wrap">
              <div className="qt-changes-banner-notice">
                <AlertCircle size={14} />
                <span>Notice: Operations updated schedule/routing details below. The accepted quotation remains permanently locked and preserved as the commercial source record.</span>
              </div>
              <table className="qt-changes-table">
                <thead>
                  <tr>
                    <th>OPERATIONAL FIELD</th>
                    <th>CATEGORY</th>
                    <th>COMMERCIAL BASELINE</th>
                    <th>CURRENT OPS VALUE</th>
                    <th>UPDATED BY</th>
                  </tr>
                </thead>
                <tbody>
                  {changes.map((ch, idx) => (
                    <tr key={idx}>
                      <td><strong>{ch.field}</strong></td>
                      <td><span className="qt-tag-category">{ch.category}</span></td>
                      <td className="qt-baseline-val">{ch.baseline_value || '—'}</td>
                      <td className="qt-current-val"><strong>{ch.current_value || '—'}</strong></td>
                      <td className="qt-changed-meta">{ch.changed_by} • {fmtDate(ch.changed_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* ── Frozen Commercial Snapshot ── */}
      {snapshot && (
        <div className="qt-handover-card">
          <div className="qt-handover-card-header">
            <FileCheck size={15} /> Preserved Commercial Terms & Conditions
          </div>
          <div className="qt-handover-grid">
            <div className="qt-handover-field">
              <span className="qt-meta-label">Commercial Quote No</span>
              <span className="qt-meta-val">{snapshot.quotation_number || '—'}</span>
            </div>
            <div className="qt-handover-field">
              <span className="qt-meta-label">Customer Name</span>
              <span className="qt-meta-val">{snapshot.customer_name || '—'}</span>
            </div>
            <div className="qt-handover-field">
              <span className="qt-meta-label">Commercial Route</span>
              <span className="qt-meta-val">{snapshot.origin || snapshot.origin_code || '—'} → {snapshot.destination || snapshot.destination_code || '—'}</span>
            </div>
            <div className="qt-handover-field">
              <span className="qt-meta-label">Transport & Service</span>
              <span className="qt-meta-val">{snapshot.transport_mode || '—'} • {snapshot.service_type || '—'}</span>
            </div>
            <div className="qt-handover-field">
              <span className="qt-meta-label">Total Commercial Value</span>
              <span className="qt-meta-val qt-meta-val--highlight">{fmtAmount(snapshot.total_amount, snapshot.currency)}</span>
            </div>
            <div className="qt-handover-field">
              <span className="qt-meta-label">Payment Terms</span>
              <span className="qt-meta-val">{snapshot.payment_terms || 'PREPAID'}</span>
            </div>
          </div>

          <div className="qt-guarantee-note">
            <ShieldCheck size={14} />
            <span>🔒 Commercial Snapshot Integrity: Operations modifications will never silently overwrite this accepted price quotation.</span>
          </div>
        </div>
      )}

      {/* ── Handover Audit Trail ── */}
      <div className="qt-handover-history-card">
        <div className="qt-handover-card-header">
          <History size={15} /> Handover & Lifecycle Audit Trail
        </div>
        {history.length === 0 ? (
          <div className="qt-empty-small">No conversion attempts recorded yet.</div>
        ) : (
          <div className="qt-handover-history-list">
            {history.map((h, idx) => (
              <div key={h.id || idx} className="qt-handover-history-item">
                <div className={`qt-history-status-indicator ${h.status === 'SUCCESS' ? 'success' : 'failed'}`} />
                <div className="qt-history-content">
                  <div className="qt-history-title-row">
                    <strong>{h.action}</strong>
                    <span className={`qt-history-status-tag ${h.status === 'SUCCESS' ? 'success' : 'failed'}`}>
                      {h.status}
                    </span>
                  </div>
                  <div className="qt-history-msg">{h.message}</div>
                  <div className="qt-history-meta">
                    {fmtDate(h.created_at)} • Performed by {h.performed_by || 'Operations'}
                    {h.booking_id && ` • Booking ID: ${h.booking_id}`}
                    {h.shipment_id && ` • Shipment ID: ${h.shipment_id}`}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ── Confirm Operational Handover Modal ── */}
      {showConfirmModal && (
        <div className="qt-modal-backdrop">
          <div className="qt-modal qt-modal--sm">
            <div className="qt-modal-header">
              <h3>Confirm Commercial Handover to Operations</h3>
              <button className="qt-modal-close" onClick={() => setShowConfirmModal(false)}><X size={16} /></button>
            </div>
            <div className="qt-modal-body">
              <div className="qt-info-box">
                <ShieldCheck size={18} color="#2563EB" />
                <div>
                  <strong>Operational Confirmation:</strong> This confirms that the accepted commercial quotation has been successfully handed over to Operations. The original quotation will remain unchanged.
                </div>
              </div>
              <div className="qt-form-group" style={{ marginTop: '14px' }}>
                <label className="qt-form-label">Confirmation & Handover Notes (Optional)</label>
                <textarea
                  className="qt-form-input"
                  rows={3}
                  placeholder="e.g. Carrier space confirmed. Assigned to Operations dispatch desk."
                  value={confirmNotes}
                  onChange={e => setConfirmNotes(e.target.value)}
                />
              </div>
            </div>
            <div className="qt-modal-footer">
              <button className="qt-btn qt-btn--ghost" onClick={() => setShowConfirmModal(false)}>Cancel</button>
              <button className="qt-btn qt-btn--primary" onClick={handleConfirmHandover} disabled={confirming}>
                {confirming ? 'Confirming...' : 'Confirm Handover'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Convert to Booking Modal (Task 18.6) ─────────────────────────────────────

function ConvertToBookingModal({ quotation, onClose, onConverted }) {
  const [form, setForm] = useState({
    booking_number: '',
    carrier_name: 'Maersk Line',
    carrier_scac: 'MAEU',
    vessel_name: 'MAERSK MC-KINNEY MOLLER',
    voyage_number: '2609E',
    etd: '',
    eta: '',
    cargo_summary: `Commercial Quote ${quotation?.quotation_number} — ${quotation?.customer_name} (${quotation?.origin} → ${quotation?.destination})`,
    special_instructions: `Terms: ${quotation?.commercial_terms || 'Standard'}. Payment: ${quotation?.payment_terms || 'PREPAID'}.`,
    operational_notes: '',
    create_shipment_immediately: true,
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const CARRIER_OPTIONS = [
    { value: 'MAEU', label: 'Maersk (MAEU)', name: 'Maersk Line' },
    { value: 'MSCU', label: 'MSC (MSCU)', name: 'Mediterranean Shipping Co' },
    { value: 'CMDU', label: 'CMA CGM (CMDU)', name: 'CMA CGM Group' },
    { value: 'HLCU', label: 'Hapag-Lloyd (HLCU)', name: 'Hapag-Lloyd AG' },
    { value: 'COSU', label: 'COSCO (COSU)', name: 'COSCO Shipping' },
    { value: 'EGLV', label: 'Evergreen (EGLV)', name: 'Evergreen Marine' },
    { value: 'OTHER', label: 'Other Freight Carrier', name: 'Other' },
  ];

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!quotation?.id) return;
    try {
      setLoading(true);
      setError(null);
      const res = await quotationService.convertToBooking(quotation.id, {
        booking_number: form.booking_number ? form.booking_number.trim() : null,
        carrier_name: form.carrier_name,
        carrier_scac: form.carrier_scac,
        vessel_name: form.vessel_name ? form.vessel_name.trim() : null,
        voyage_number: form.voyage_number ? form.voyage_number.trim() : null,
        etd: form.etd ? new Date(form.etd).toISOString() : null,
        eta: form.eta ? new Date(form.eta).toISOString() : null,
        cargo_summary: form.cargo_summary ? form.cargo_summary.trim() : null,
        special_instructions: form.special_instructions ? form.special_instructions.trim() : null,
        operational_notes: form.operational_notes ? form.operational_notes.trim() : '',
        create_shipment_immediately: form.create_shipment_immediately,
      });
      const result = res?.data || res;
      onConverted(result);
    } catch (err) {
      setError(err?.message || 'Failed to convert quotation to booking.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="qt-modal-overlay" onClick={onClose}>
      <div className="qt-modal qt-modal--lg" onClick={e => e.stopPropagation()}>
        <div className="qt-modal-header">
          <div>
            <h2 className="qt-modal-title">Convert Quotation to Operational Booking</h2>
            <p className="qt-modal-sub">Create an active operational booking from accepted commercial terms for {quotation.quotation_number}</p>
          </div>
          <button className="qt-modal-close" onClick={onClose}><X size={20} /></button>
        </div>

        <form onSubmit={handleSubmit} className="qt-modal-body">
          {error && <div className="qt-modal-error">{error}</div>}

          {/* Commercial Source (Read-only) */}
          <div className="qt-modal-section-title">
            <FileText size={14} /> Commercial Source (Read-Only)
          </div>
          <div className="qt-commercial-summary-box">
            <div className="qt-comm-row">
              <div><span>Quote Number:</span> <strong>{quotation.quotation_number}</strong></div>
              <div><span>Customer:</span> <strong>{quotation.customer_name}</strong></div>
              <div><span>Route:</span> <strong>{quotation.origin} → {quotation.destination}</strong></div>
            </div>
            <div className="qt-comm-row">
              <div><span>Service / Mode:</span> <strong>{quotation.service_type} • {quotation.transport_mode}</strong></div>
              <div><span>Payment Terms:</span> <strong>{quotation.payment_terms || 'PREPAID'}</strong></div>
              <div><span>Total Amount:</span> <strong style={{ color: '#15803D' }}>{fmtAmount(quotation.total_amount, quotation.currency)}</strong></div>
            </div>
          </div>

          {/* Operational Setup (Editable) */}
          <div className="qt-modal-section-title" style={{ marginTop: '16px' }}>
            <Compass size={14} /> Operational Booking Configuration
          </div>

          <div className="qt-form-row">
            <div className="qt-form-group">
              <label>Carrier *</label>
              <CustomSelect
                value={form.carrier_scac}
                onChange={v => {
                  const opt = CARRIER_OPTIONS.find(c => c.value === v);
                  setForm(f => ({ ...f, carrier_scac: v, carrier_name: opt ? opt.name : f.carrier_name }));
                }}
                options={CARRIER_OPTIONS}
              />
            </div>
            <div className="qt-form-group">
              <label>Carrier Name Override</label>
              <input
                value={form.carrier_name}
                onChange={e => setForm(f => ({ ...f, carrier_name: e.target.value }))}
                placeholder="e.g. Maersk Line"
                required
              />
            </div>
            <div className="qt-form-group">
              <label>Booking Number Override (Optional)</label>
              <input
                value={form.booking_number}
                onChange={e => setForm(f => ({ ...f, booking_number: e.target.value }))}
                placeholder="Auto-generated if blank"
              />
            </div>
          </div>

          <div className="qt-form-row">
            <div className="qt-form-group">
              <label>Vessel / Flight Name</label>
              <input
                value={form.vessel_name}
                onChange={e => setForm(f => ({ ...f, vessel_name: e.target.value }))}
                placeholder="e.g. MAERSK MC-KINNEY MOLLER"
              />
            </div>
            <div className="qt-form-group">
              <label>Voyage / Flight Number</label>
              <input
                value={form.voyage_number}
                onChange={e => setForm(f => ({ ...f, voyage_number: e.target.value }))}
                placeholder="e.g. 2609E"
              />
            </div>
          </div>

          <div className="qt-form-row">
            <div className="qt-form-group">
              <label>Estimated Departure (ETD)</label>
              <input
                type="date"
                value={form.etd}
                onChange={e => setForm(f => ({ ...f, etd: e.target.value }))}
              />
            </div>
            <div className="qt-form-group">
              <label>Estimated Arrival (ETA)</label>
              <input
                type="date"
                value={form.eta}
                onChange={e => setForm(f => ({ ...f, eta: e.target.value }))}
              />
            </div>
          </div>

          <div className="qt-form-group">
            <label>Cargo Summary & Description</label>
            <input
              value={form.cargo_summary}
              onChange={e => setForm(f => ({ ...f, cargo_summary: e.target.value }))}
            />
          </div>

          <div className="qt-form-group">
            <label>Internal Operational Handover Notes</label>
            <textarea
              rows={2}
              placeholder="Instructions for operations dispatcher, origin port agent, or customs team..."
              value={form.operational_notes}
              onChange={e => setForm(f => ({ ...f, operational_notes: e.target.value }))}
            />
          </div>

          <div className="qt-form-row qt-form-row--align-center">
            <label className="qt-checkbox-label">
              <input
                type="checkbox"
                checked={form.create_shipment_immediately}
                onChange={e => setForm(f => ({ ...f, create_shipment_immediately: e.target.checked }))}
              />
              <span><strong>Auto-Generate Operations Shipment:</strong> Automatically create and link tracking milestones in Shipments workflow.</span>
            </label>
          </div>

          <div className="qt-conversion-warning-box">
            <ShieldCheck size={16} />
            <div>
              <strong>Commercial Record Protection:</strong>
              <div>You are about to create an operational booking from this accepted quotation. The quotation will remain unchanged as the immutable commercial record.</div>
            </div>
          </div>

          <div className="qt-modal-footer">
            <button type="button" className="qt-btn qt-btn--ghost" onClick={onClose}>Cancel</button>
            <button type="submit" className="qt-btn qt-btn--primary" disabled={loading}>
              {loading ? <><RefreshCw size={14} className="qt-spin" /> Converting…</> : <><Briefcase size={14} /> Create Booking</>}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ─── Conversion Success Modal (Task 18.6) ─────────────────────────────────────

function ConversionSuccessModal({ result, onClose }) {
  if (!result) return null;
  return (
    <div className="qt-modal-overlay" onClick={onClose}>
      <div className="qt-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '520px' }}>
        <div className="qt-success-modal-header">
          <div className="qt-success-badge-icon"><CheckCircle2 size={36} /></div>
          <h2>✓ BOOKING CREATED SUCCESSFULLY</h2>
          <p>Commercial quotation was converted into an operational booking.</p>
        </div>

        <div className="qt-success-details-card">
          <div className="qt-success-detail-row">
            <span>Booking Number:</span>
            <strong style={{ fontSize: '15px', color: '#0F172A' }}>{result.booking_number || `BK-${result.booking_id}`}</strong>
          </div>
          {result.shipment_id && (
            <div className="qt-success-detail-row">
              <span>Linked Shipment ID:</span>
              <strong>Shipment #{result.shipment_id}</strong>
            </div>
          )}
          <div className="qt-success-detail-row">
            <span>Source Quotation:</span>
            <strong>{result.quotation_number}</strong>
          </div>
          <div className="qt-success-detail-row">
            <span>Conversion Status:</span>
            <span className="qt-status-badge" style={{ background: '#DCFCE7', color: '#15803D' }}>CONVERTED</span>
          </div>
          <div className="qt-success-detail-row">
            <span>Timestamp:</span>
            <span>{new Date(result.converted_at || Date.now()).toLocaleString()}</span>
          </div>
        </div>

        <div className="qt-success-modal-actions">
          <a href="/dashboard/bookings" className="qt-btn qt-btn--primary">
            <ArrowUpRight size={14} /> Open Bookings Workspace
          </a>
          {result.shipment_id && (
            <a href="/dashboard/shipments" className="qt-btn qt-btn--secondary">
              <ArrowUpRight size={14} /> Open Shipments
            </a>
          )}
          <button type="button" className="qt-btn qt-btn--ghost" onClick={onClose}>
            Back to Quotation
          </button>
        </div>
      </div>
    </div>
  );
}


// ─── Main Page ────────────────────────────────────────────────────────────────

export default function QuotationsPage() {
  const [summary, setSummary] = useState(null);
  const [quotations, setQuotations] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const [activeTab, setActiveTab] = useState('All');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [validityFilter, setValidityFilter] = useState('ALL');
  const [page, setPage] = useState(1);
  const LIMIT = 10;

  const [viewMode, setViewMode] = useState('workspace'); // 'workspace' | 'analytics'
  const [selectedId, setSelectedId] = useState(null);
  const [showNewModal, setShowNewModal] = useState(false);
  const [showTemplatesDrawer, setShowTemplatesDrawer] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const searchRef = useRef(null);

  const tabToStatus = useCallback((tab) => {
    switch (tab) {
      case 'Draft':             return 'DRAFT';
      case 'Ready For Review':  return 'READY_FOR_REVIEW';
      case 'Approved':          return 'APPROVED';
      case 'Changes Requested': return 'CHANGES_REQUESTED';
      case 'Sent':              return 'SENT';
      case 'Viewed':            return 'VIEWED';
      case 'Accepted':          return 'ACCEPTED';
      case 'Declined':          return 'DECLINED';
      case 'Expired':           return 'EXPIRED';
      case 'Cancelled':         return 'CANCELLED';
      default:                  return 'ALL';
    }
  }, []);

  const fetchSummary = useCallback(async () => {
    try {
      const res = await quotationService.getSummary();
      setSummary(res.data || res);
    } catch (e) {
      console.warn('Failed to load quotation summary', e);
    }
  }, []);

  const fetchQuotations = useCallback(async () => {
    setLoading(true);
    setError(null);
    const status = tabToStatus(activeTab);
    quotationService.listQuotations({
      search, status, validity: validityFilter, page, limit: LIMIT,
    }).then(res => {
      const d = res?.data || res;
      setQuotations(d?.quotations || []);
      setTotal(d?.total || 0);
    }).catch(err => {
      setError('Failed to load quotations. Please try again.');
      console.error(err);
    }).finally(() => setLoading(false));
  }, [activeTab, search, validityFilter, page, tabToStatus]);

  useEffect(() => {
    fetchSummary();
  }, [fetchSummary, refreshKey]);

  useEffect(() => {
    fetchQuotations();
  }, [fetchQuotations, refreshKey]);

  const totalPages = Math.ceil(total / LIMIT) || 1;

  const handleCreated = (q) => {
    setShowNewModal(false);
    setRefreshKey(k => k + 1);
    if (q?.id) setSelectedId(q.id);
  };

  const handleRowClick = (id) => {
    setSelectedId(prev => prev === id ? null : id);
  };

  const handleSearchChange = (e) => {
    setSearch(e.target.value);
    setPage(1);
  };

  const handleTabChange = (tab) => {
    setActiveTab(tab);
    setPage(1);
  };

  const getPageNumbers = () => {
    const pages = [];
    if (totalPages <= 6) {
      for (let i = 1; i <= totalPages; i++) pages.push(i);
    } else {
      pages.push(1);
      if (page > 3) pages.push('…');
      for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) pages.push(i);
      if (page < totalPages - 2) pages.push('…');
      pages.push(totalPages);
    }
    return pages;
  };

  return (
    <div className={`qt-page ${selectedId && viewMode === 'workspace' ? 'qt-page--split' : ''}`}>
      {/* ── Left Panel ── */}
      <div className="qt-main">
        {/* Header with Top-Level Workspace Switcher */}
        <div className="qt-header">
          <div className="qt-header-left">
            <div className="qt-top-nav-switcher">
              <button
                className={`qt-top-nav-btn ${viewMode === 'workspace' ? 'active' : ''}`}
                onClick={() => setViewMode('workspace')}
              >
                <FileText size={14} /> Quotations Workspace
              </button>
              <button
                className={`qt-top-nav-btn ${viewMode === 'analytics' ? 'active' : ''}`}
                onClick={() => setViewMode('analytics')}
              >
                <BarChart3 size={14} /> Performance & Analytics
              </button>
            </div>
            <p className="qt-subtitle" style={{ marginTop: '6px' }}>
              {viewMode === 'workspace'
                ? 'Manage customer quotations, rate pricing, templates, and commercial handover.'
                : 'Management intelligence on quotation volume, win rates, conversion, and margin health.'}
            </p>
          </div>
          <div className="qt-header-right">
            {viewMode === 'workspace' && (
              <>
                <button
                  className="qt-btn qt-btn--ghost"
                  onClick={() => setShowTemplatesDrawer(true)}
                >
                  <Bookmark size={14} /> Templates Library
                </button>
                <button
                  className="qt-btn qt-btn--primary"
                  onClick={() => setShowNewModal(true)}
                >
                  <Plus size={15} /> New Quotation
                </button>
              </>
            )}
          </div>
        </div>

        {/* ── Render Analytics Workspace or Operations Table ── */}
        {viewMode === 'analytics' ? (
          <QuotationAnalyticsWorkspace
            onSelectQuotation={(id) => {
              setViewMode('workspace');
              setSelectedId(id);
            }}
          />
        ) : (
          <>
            {/* KPI Strip */}
            <div className="qt-kpi-strip">
          {KPI_CONFIG.map(kpi => {
            const IconComp = kpi.icon;
            const val = summary ? (summary[kpi.key] ?? 0) : 0;
            const totalQuotes = summary ? (summary.total_quotations ?? 0) : 0;
            const pct = totalQuotes > 0 ? Math.round((val / totalQuotes) * 100) : 0;

            return (
              <div key={kpi.key} className="qt-kpi-card">
                <div className="qt-kpi-top-row">
                  <div className="qt-kpi-icon" style={{ background: kpi.iconBg, color: kpi.iconColor }}>
                    <IconComp size={18} />
                  </div>
                  <span className="qt-kpi-tag" style={{ color: kpi.iconColor, background: kpi.iconBg }}>
                    {kpi.badge}
                  </span>
                </div>
                <div className="qt-kpi-body">
                  <div className="qt-kpi-value">{summary ? val : '—'}</div>
                  <div className="qt-kpi-label">{kpi.label}</div>
                  <div className={`qt-kpi-trend ${kpi.negative && val > 0 ? 'negative' : ''}`}>
                    <span>{kpi.negative && val > 0 ? '↓' : '↑'} {totalQuotes > 0 ? `${pct}%` : '0'}</span>
                    <span className="qt-kpi-trend-sub">{totalQuotes > 0 ? 'of total' : 'no active quotes'}</span>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Status Tabs */}
        <div className="qt-tabs">
          {STATUS_TABS.map(tab => (
            <button
              key={tab}
              className={`qt-tab ${activeTab === tab ? 'active' : ''}`}
              onClick={() => handleTabChange(tab)}
            >
              {tab}
            </button>
          ))}
        </div>

        {/* Search + Filters */}
        <div className="qt-filter-bar">
          <div className="qt-search-wrap">
            <Search size={14} className="qt-search-icon" />
            <input
              ref={searchRef}
              className="qt-search-input"
              placeholder="Search by quotation no., customer, route..."
              value={search}
              onChange={handleSearchChange}
            />
          </div>

          <div className="qt-filter-select-group">
            <div className="qt-filter-select">
              <span className="qt-filter-label">Status</span>
              <CustomSelect
                value={statusFilter}
                onChange={v => { setStatusFilter(v); setPage(1); }}
                options={[
                  { value: 'ALL', label: 'All Statuses' },
                  ...Object.keys(STATUS_CONFIG).map(s => ({
                    value: s,
                    label: STATUS_CONFIG[s].label,
                  }))
                ]}
                size="sm"
              />
            </div>
            <div className="qt-filter-select">
              <span className="qt-filter-label">Valid Till</span>
              <CustomSelect
                value={validityFilter}
                onChange={v => { setValidityFilter(v); setPage(1); }}
                options={[
                  { value: 'ALL', label: 'All' },
                  { value: 'VALID', label: 'Valid / Active' },
                  { value: 'EXPIRING_SOON', label: 'Expiring Soon' },
                  { value: 'EXPIRED', label: 'Expired' },
                ]}
                size="sm"
              />
            </div>
          </div>

          <button className="qt-filter-btn">
            <Filter size={13} /> Filters
          </button>
        </div>

        {/* Error Banner */}
        {error && (
          <div className="qt-error-banner">
            <AlertCircle size={14} /> {error}
            <button onClick={() => setRefreshKey(k => k + 1)}>Retry</button>
          </div>
        )}

        {/* Table */}
        {loading ? (
          <div className="qt-table-wrap">
            <table className="qt-table">
              <thead>
                <tr>
                  <th>QUOTATION NO.</th><th>CUSTOMER</th><th>ROUTE / SERVICE</th>
                  <th>AMOUNT</th><th>STATUS</th><th>VALID TILL</th><th>UPDATED</th><th>ACTIONS</th>
                </tr>
              </thead>
              <tbody>
                {[1,2,3,4,5,6].map(i => (
                  <tr key={i} className="qt-skel-row">
                    <td colSpan={8}><div className="qt-skel-bar" /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : quotations.length === 0 && activeTab === 'All' && !search ? (
          <ModuleHeroEmptyState
            icon={<FileText size={28} />}
            badgeTheme="indigo"
            title="No Freight Quotations Created Yet"
            description="Generate professional freight quotes with multi-modal rate calculation, automated markup controls, PDF generation, and customer email sharing."
            primaryAction={{
              label: 'Create Quotation',
              icon: <Plus size={15} />,
              onClick: () => setShowNewModal(true),
            }}
            secondaryAction={{
              label: 'Browse Inbound RFQs',
              icon: <ArrowRight size={15} />,
              onClick: () => navigate('/dashboard/rfqs'),
            }}
            features={[
              {
                icon: <Sparkles size={18} />,
                iconBg: '#eff6ff',
                iconColor: '#2563eb',
                title: 'Multi-Leg Rate Calculation',
                desc: 'Automatically combine ocean/air base freights with origin and destination local haulage and customs.',
              },
              {
                icon: <Percent size={18} />,
                iconBg: '#ecfdf5',
                iconColor: '#059669',
                title: 'Margin & Profit Protection',
                desc: 'Enforce target margin percentages and manager approval triggers for discounted rates.',
              },
              {
                icon: <Send size={18} />,
                iconBg: '#f5f3ff',
                iconColor: '#7c3aed',
                title: 'Instant Customer Portal & PDF',
                desc: 'Generate branded PDF quotes with validity timers, digital acceptance, and live viewing telemetry.',
              },
            ]}
          />
        ) : quotations.length === 0 ? (
          <div className="qt-empty-state">
            <div className="qt-empty-icon"><FileText size={48} /></div>
            <h3>No quotations found</h3>
            <p>
              Try adjusting your search or filter criteria.
            </p>
          </div>
        ) : (
          <div className="qt-table-wrap">
            <table className="qt-table">
              <thead>
                <tr>
                  <th>QUOTATION NO.</th>
                  <th>CUSTOMER</th>
                  <th>ROUTE / SERVICE</th>
                  <th>AMOUNT</th>
                  <th>STATUS</th>
                  <th>VALID TILL</th>
                  <th>UPDATED</th>
                  <th>ACTIONS</th>
                </tr>
              </thead>
              <tbody>
                {quotations.map(q => {
                  const exp = expiryLabel(q.valid_until);
                  const isSelected = selectedId === q.id;
                  return (
                    <tr
                      key={q.id}
                      className={`qt-row ${isSelected ? 'qt-row--selected' : ''}`}
                      onClick={() => handleRowClick(q.id)}
                    >
                      <td>
                        <div className="qt-cell-main">{q.quotation_number}</div>
                        <div className="qt-cell-sub">{q.payment_terms || 'PREPAID'}</div>
                      </td>
                      <td>
                        <div className="qt-cell-main">{q.customer_name || '—'}</div>
                        {q.customer_id && (
                          <div className="qt-cell-sub">CUST-{String(q.customer_id).padStart(4, '0')}</div>
                        )}
                      </td>
                      <td>
                        {q.origin || q.destination ? (
                          <>
                            <div className="qt-cell-route">
                              <span>{q.origin || '—'}{q.origin_code ? ` (${q.origin_code})` : ''}</span>
                              <ArrowRight size={11} />
                              <span>{q.destination || '—'}{q.destination_code ? ` (${q.destination_code})` : ''}</span>
                            </div>
                            {q.service_type && (
                              <div className="qt-cell-sub">{q.service_type} • {q.transport_mode}</div>
                            )}
                          </>
                        ) : <span className="qt-cell-empty">—</span>}
                      </td>
                      <td>
                        <div className="qt-cell-amount">
                          {fmtAmount(q.total_amount, q.currency)}
                        </div>
                        {q.gross_margin_pct !== undefined && q.gross_margin_pct > 0 && (
                          <div className="qt-cell-sub" style={{ color: '#15803D' }}>
                            Margin: {Number(q.gross_margin_pct).toFixed(1)}%
                          </div>
                        )}
                      </td>
                      <td>
                        <StatusBadge status={q.status} />
                        {q.updated_at && (
                          <div className="qt-cell-sub">{timeAgo(q.updated_at)}</div>
                        )}
                      </td>
                      <td>
                        {q.valid_until ? (
                          <>
                            <div className="qt-cell-main">{fmtDate(q.valid_until)}</div>
                            {exp && (
                              <div className="qt-cell-expiry" style={{ color: exp.color }}>
                                {exp.text}
                              </div>
                            )}
                          </>
                        ) : <span className="qt-cell-empty">—</span>}
                      </td>
                      <td>
                        {q.updated_at ? (
                          <>
                            <div className="qt-cell-main">{fmtDate(q.updated_at)}</div>
                            <div className="qt-cell-sub">
                              {new Date(q.updated_at).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                            </div>
                          </>
                        ) : <span className="qt-cell-empty">—</span>}
                      </td>
                      <td onClick={e => e.stopPropagation()}>
                        <div className="qt-row-actions-cell">
                          <button
                            className="qt-btn qt-btn--ghost qt-btn--xs qt-view-details-btn"
                            onClick={() => handleRowClick(q.id)}
                            title="View Details"
                          >
                            <Eye size={13} />
                            <span>View Details</span>
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {!loading && quotations.length > 0 && (
          <div className="qt-pagination">
            <span className="qt-pagination-info">
              Showing {(page - 1) * LIMIT + 1} to {Math.min(page * LIMIT, total)} of {total} quotations
            </span>
            <div className="qt-pagination-controls">
              <button
                className="qt-page-btn"
                onClick={() => setPage(1)}
                disabled={page === 1}
              >«</button>
              <button
                className="qt-page-btn"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
              >‹</button>
              {getPageNumbers().map((p, i) =>
                p === '…' ? (
                  <span key={`ellipsis-${i}`} className="qt-page-ellipsis">…</span>
                ) : (
                  <button
                    key={p}
                    className={`qt-page-btn ${page === p ? 'active' : ''}`}
                    onClick={() => setPage(p)}
                  >{p}</button>
                )
              )}
              <button
                className="qt-page-btn"
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
              >›</button>
              <button
                className="qt-page-btn"
                onClick={() => setPage(totalPages)}
                disabled={page === totalPages}
              >»</button>
            </div>
          </div>
        )}
        </>
        )}
      </div>

      {/* ── Right Detail Panel ── */}
      {selectedId && (
        <QuotationDetailPanel
          quotationId={selectedId}
          onClose={() => setSelectedId(null)}
          onQuotationUpdated={() => setRefreshKey(k => k + 1)}
          onOpenTemplatesManager={() => setShowTemplatesDrawer(true)}
        />
      )}

      {/* ── New Quotation Modal ── */}
      {showNewModal && (
        <NewQuotationModal
          onClose={() => setShowNewModal(false)}
          onCreated={handleCreated}
        />
      )}

      {/* ── Reusable Templates Library Drawer ── */}
      {showTemplatesDrawer && (
        <TemplatesManagementDrawer
          onClose={() => setShowTemplatesDrawer(false)}
        />
      )}
    </div>
  );
}
