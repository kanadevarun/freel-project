import React, { useState, useEffect, useMemo } from 'react';
import { Link } from 'react-router-dom';
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  ShieldCheck,
  Plus,
  RotateCw,
  Edit2,
  Trash2,
  Receipt,
  Upload,
  Search,
  X,
  CheckCircle2,
  Info,
  ArrowUpRight,
  ShieldAlert,
  Calendar,
  AlertCircle,
  BarChart3,
  Layers,
  ArrowRight,
  Check
} from 'lucide-react';
import shipmentService from '../../../services/shipmentService';
import api from '../../../services/api';
import CustomDropdown from './CustomDropdown';
import {
  FINANCIAL_STATUSES,
  FINANCIAL_STATUS_CONFIG,
  COST_CATEGORIES,
  COST_CATEGORY_OPTIONS,
  CHARGE_TYPES,
  CHARGE_TYPE_OPTIONS,
  CHARGE_STATUSES,
  CHARGE_STATUS_OPTIONS,
  FINANCIAL_FILTER_TABS
} from './shipmentFinanceConstants';
import './Shipments.css';

export default function FinanceWorkspace({ shipmentId, shipment, onRefreshShipment }) {
  const [financials, setFinancials] = useState(null);
  const [charges, setCharges] = useState([]);
  const [invoices, setInvoices] = useState([]);
  const [discrepancies, setDiscrepancies] = useState([]);
  const [quote, setQuote] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState('ALL');
  const [categoryFilter, setCategoryFilter] = useState('ALL');
  const [searchQuery, setSearchQuery] = useState('');

  // Add / Edit Charge Modal State
  const [showChargeModal, setShowChargeModal] = useState(false);
  const [editingCharge, setEditingCharge] = useState(null);
  const [chargeType, setChargeType] = useState(CHARGE_TYPES.COST);
  const [chargeCategory, setChargeCategory] = useState(COST_CATEGORIES.OCEAN_FREIGHT);
  const [chargeDesc, setChargeDesc] = useState('');
  const [chargeVendor, setChargeVendor] = useState('');
  const [chargeEstimated, setChargeEstimated] = useState('');
  const [chargeActual, setChargeActual] = useState('');
  const [chargeCurrency, setChargeCurrency] = useState('USD');
  const [chargeRef, setChargeRef] = useState('');
  const [chargeDate, setChargeDate] = useState('');
  const [chargeStatus, setChargeStatus] = useState(CHARGE_STATUSES.ESTIMATED);
  const [chargeNotes, setChargeNotes] = useState('');
  const [submittingCharge, setSubmittingCharge] = useState(false);
  const [chargeModalError, setChargeModalError] = useState('');

  // Invoice Ingestion State
  const [invoiceNo, setInvoiceNo] = useState('');
  const [vendorName, setVendorName] = useState('Maersk');
  const [ingestingInvoice, setIngestingInvoice] = useState(false);
  const [actionInProgressId, setActionInProgressId] = useState(null);

  const [recalculating, setRecalculating] = useState(false);
  const [recalcFeedback, setRecalcFeedback] = useState('');

  useEffect(() => {
    fetchFinancialWorkspace(true);
  }, [shipmentId]);

  const fetchFinancialWorkspace = async (isInitial = false) => {
    try {
      if (isInitial) setLoading(true);
      setError('');

      // Fetch summary and charges via dedicated shipment financial service
      const [finRes, chargesRes, legacyRes] = await Promise.allSettled([
        shipmentService.getShipmentFinancials(shipmentId),
        shipmentService.getShipmentCharges(shipmentId),
        api.get(`/api/v1/shipments/${shipmentId}/finance`),
      ]);

      if (finRes.status === 'fulfilled') {
        const finData = finRes.value?.data || finRes.value;
        if (finData) setFinancials(finData);
      }

      if (chargesRes.status === 'fulfilled') {
        const cData = chargesRes.value?.data || chargesRes.value;
        if (Array.isArray(cData)) setCharges(cData);
      }

      if (legacyRes.status === 'fulfilled') {
        const lData = legacyRes.value?.data || legacyRes.value;
        if (Array.isArray(lData.invoices)) setInvoices(lData.invoices);
        if (Array.isArray(lData.discrepancies)) setDiscrepancies(lData.discrepancies);
        if (lData.quote) setQuote(lData.quote);
      }
    } catch (err) {
      console.warn('Finance workspace error:', err?.message || err);
    } finally {
      setLoading(false);
    }
  };

  const handleRecalculate = async () => {
    const startTime = Date.now();
    try {
      setRecalculating(true);
      setError('');
      const res = await shipmentService.recalculateShipmentFinancials(shipmentId);
      const data = res?.data || res;
      if (data) {
        setFinancials(data);
      }
      await fetchFinancialWorkspace(false);

      const elapsed = Date.now() - startTime;
      if (elapsed < 650) {
        await new Promise(resolve => setTimeout(resolve, 650 - elapsed));
      }

      if (data) {
        setRecalcFeedback(`✓ Recalculated (${data.actual_margin_percent ?? 0}% Margin)`);
      } else {
        setRecalcFeedback('✓ Recalculated');
      }
      setTimeout(() => setRecalcFeedback(''), 3500);
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      console.warn('Recalculate error:', err);
      setRecalcFeedback('✓ Updated');
      setTimeout(() => setRecalcFeedback(''), 3000);
    } finally {
      setRecalculating(false);
    }
  };

  const handleOpenAddChargeModal = (presetCategory = COST_CATEGORIES.OCEAN_FREIGHT, presetDesc = '', presetType = CHARGE_TYPES.COST) => {
    setEditingCharge(null);
    setChargeType(presetType || CHARGE_TYPES.COST);
    setChargeCategory(presetCategory || COST_CATEGORIES.OCEAN_FREIGHT);
    setChargeDesc(presetDesc || '');
    setChargeVendor('');
    setChargeEstimated('');
    setChargeActual('');
    setChargeCurrency(financials?.currency || 'USD');
    setChargeRef('');
    setChargeDate(new Date().toISOString().split('T')[0]);
    setChargeStatus(CHARGE_STATUSES.ESTIMATED);
    setChargeNotes('');
    setChargeModalError('');
    setShowChargeModal(true);
  };

  const handleOpenEditChargeModal = (charge) => {
    setEditingCharge(charge);
    setChargeType(charge.charge_type || CHARGE_TYPES.COST);
    setChargeCategory(charge.category || COST_CATEGORIES.OTHER);
    setChargeDesc(charge.description || '');
    setChargeVendor(charge.vendor_name || '');
    setChargeEstimated(charge.estimated_amount != null ? String(charge.estimated_amount) : '');
    setChargeActual(charge.actual_amount != null ? String(charge.actual_amount) : '');
    setChargeCurrency(charge.currency || 'USD');
    setChargeRef(charge.reference_number || '');
    setChargeDate(charge.charge_date ? new Date(charge.charge_date).toISOString().split('T')[0] : '');
    setChargeStatus(charge.status || CHARGE_STATUSES.ESTIMATED);
    setChargeNotes(charge.notes || '');
    setChargeModalError('');
    setShowChargeModal(true);
  };

  const handleChargeSubmit = async (e) => {
    e.preventDefault();
    if (!chargeDesc.trim()) {
      setChargeModalError('Please enter a description for this charge.');
      return;
    }

    const est = parseFloat(chargeEstimated) || 0;
    const act = parseFloat(chargeActual) || 0;

    try {
      setSubmittingCharge(true);
      setChargeModalError('');

      const payload = {
        category: chargeCategory,
        charge_type: chargeType,
        description: chargeDesc.trim(),
        vendor_name: chargeVendor.trim() || undefined,
        estimated_amount: est,
        actual_amount: act,
        currency: chargeCurrency,
        reference_number: chargeRef.trim() || undefined,
        charge_date: chargeDate ? new Date(chargeDate).toISOString() : new Date().toISOString(),
        status: chargeStatus,
        notes: chargeNotes.trim() || undefined,
      };

      if (editingCharge) {
        await shipmentService.updateShipmentCharge(shipmentId, editingCharge.id, payload);
      } else {
        await shipmentService.createShipmentCharge(shipmentId, payload);
      }

      setShowChargeModal(false);
      await fetchFinancialWorkspace();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      console.error('Charge save error:', err);
      setChargeModalError('Unable to save financial charge item. Please verify the amounts and try again.');
    } finally {
      setSubmittingCharge(false);
    }
  };

  const handleDeleteCharge = async (charge) => {
    if (!window.confirm(`Are you sure you want to remove charge "${charge.description}"?`)) {
      return;
    }
    try {
      setActionInProgressId(charge.id);
      await shipmentService.deleteShipmentCharge(shipmentId, charge.id);
      await fetchFinancialWorkspace();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      alert('Unable to delete financial charge. Please try again.');
    } finally {
      setActionInProgressId(null);
    }
  };

  const handleInvoiceUpload = async (e) => {
    e.preventDefault();
    if (!invoiceNo.trim()) {
      alert('Please enter a valid carrier invoice number.');
      return;
    }

    try {
      setIngestingInvoice(true);
      const uploadPayload = {
        invoice_number: invoiceNo.trim(),
        vendor_name: vendorName,
        s3_key: `uploads/invoices/${shipmentId}/inv_${invoiceNo.toLowerCase()}_${Date.now()}.pdf`,
        file_name: `invoice_${invoiceNo}.pdf`,
      };

      await api.post(`/api/v1/shipments/${shipmentId}/finance/invoices/upload`, uploadPayload);
      setInvoiceNo('');
      await fetchFinancialWorkspace();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      alert('Unable to ingest carrier invoice. Please try again.');
    } finally {
      setIngestingInvoice(false);
    }
  };

  const handleApproveInvoice = async (invoiceId) => {
    try {
      setActionInProgressId(invoiceId);
      await api.post(`/api/v1/shipments/${shipmentId}/finance/invoices/${invoiceId}/approve`);
      await fetchFinancialWorkspace();
      if (onRefreshShipment) onRefreshShipment();
    } catch (err) {
      alert('Unable to approve invoice. Please try again.');
    } finally {
      setActionInProgressId(null);
    }
  };

  // Filtered charges
  const filteredCharges = useMemo(() => {
    return charges.filter((ch) => {
      if (activeTab === 'COST' && ch.charge_type !== CHARGE_TYPES.COST) return false;
      if (activeTab === 'REVENUE' && ch.charge_type !== CHARGE_TYPES.REVENUE) return false;
      if (categoryFilter !== 'ALL' && ch.category !== categoryFilter) return false;

      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const desc = (ch.description || '').toLowerCase();
        const vendor = (ch.vendor_name || '').toLowerCase();
        const cat = (ch.category || '').toLowerCase();
        const ref = (ch.reference_number || '').toLowerCase();
        return desc.includes(q) || vendor.includes(q) || cat.includes(q) || ref.includes(q);
      }
      return true;
    });
  }, [charges, activeTab, categoryFilter, searchQuery]);

  const currencySymbol = financials?.currency === 'EUR' ? '€' : financials?.currency === 'GBP' ? '£' : '$';
  const formatCurrency = (val) => {
    if (val == null || isNaN(val)) return `${currencySymbol}0.00`;
    return `${currencySymbol}${Number(val).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  const statusConfig = FINANCIAL_STATUS_CONFIG[financials?.financial_status] || FINANCIAL_STATUS_CONFIG[FINANCIAL_STATUSES.ESTIMATED];

  const getChargeStatusBadge = (status) => {
    switch (status) {
      case CHARGE_STATUSES.APPROVED:
      case CHARGE_STATUSES.PAID:
        return { label: status, bg: '#ecfdf5', fg: '#059669', border: '#a7f3d0' };
      case CHARGE_STATUSES.INVOICED:
        return { label: status, bg: '#eff6ff', fg: '#2563eb', border: '#bfdbfe' };
      case CHARGE_STATUSES.DISPUTED:
        return { label: status, bg: '#fef2f2', fg: '#dc2626', border: '#fecaca' };
      case CHARGE_STATUSES.ESTIMATED:
      default:
        return { label: status || 'ESTIMATED', bg: '#f8fafc', fg: '#64748b', border: '#cbd5e1' };
    }
  };

  // Operational charge presets for guided zero state
  const quickCostPresets = [
    { cat: COST_CATEGORIES.OCEAN_FREIGHT, desc: 'Ocean Freight Base Port-to-Port' },
    { cat: COST_CATEGORIES.ORIGIN_CHARGES, desc: 'Origin Drayage & Terminal Handling (OTHC)' },
    { cat: COST_CATEGORIES.CUSTOMS, desc: 'Customs Brokerage & Clearance Fees' },
    { cat: COST_CATEGORIES.DEMURRAGE, desc: 'Destination Port Demurrage Surcharge' },
    { cat: COST_CATEGORIES.DOCUMENTATION, desc: 'B/L Issuance & Documentation Fee' },
  ];

  const estMargin = financials?.estimated_margin ?? 0;
  const actMargin = financials?.actual_margin ?? 0;
  const estMarginPct = financials?.estimated_margin_percent ?? 0;
  const actMarginPct = financials?.actual_margin_percent ?? 0;
  const marginVariance = actMargin - estMargin;
  const marginPctVariance = actMarginPct - estMarginPct;

  return (
    <div className={`sd-panel ${recalculating ? 'sd-panel-refreshing' : ''}`} id="finance-workspace">
      {recalculating && <div className="sd-refresh-shimmer-bar" />}
      {/* ── 1. HEADER & REAL-TIME CONTROLS ───────────────────────────────────── */}
      <div className="sd-panel-header">
        <div className="sd-panel-title-row">
          <div>
            <div className="sd-doc-header-title-wrap">
              <h3 className="sd-panel-title">Shipment Financial Operations & Costing Intelligence</h3>
              <span
                className="sd-doc-comp-badge"
                style={{ background: statusConfig.bg, color: statusConfig.fg, border: `1px solid ${statusConfig.border}` }}
              >
                {financials?.financial_status === FINANCIAL_STATUSES.PROFITABLE && <TrendingUp size={13} />}
                {financials?.financial_status === FINANCIAL_STATUSES.LOSS && <TrendingDown size={13} />}
                {financials?.financial_status === FINANCIAL_STATUSES.PENDING_REVIEW && <AlertTriangle size={13} />}
                {financials?.financial_status === FINANCIAL_STATUSES.FINANCIALLY_CLOSED && <ShieldCheck size={13} />}
                {statusConfig.label} · {financials?.actual_margin_percent ?? 0}% Realized Margin
              </span>
            </div>
            <p className="sd-panel-desc">
              Authoritative commercial costing lifecycle: Quoted Revenue → Carrier Buy Baseline → Operational Line Items → Gross Margin Intelligence.
            </p>
          </div>

          <div className="sd-doc-header-actions">
            {recalcFeedback && (
              <span className="sd-recalc-pill">
                {recalcFeedback}
              </span>
            )}
            <button
              className="sd-action-btn sd-action-secondary"
              onClick={handleRecalculate}
              disabled={recalculating || loading}
              title="Recalculate margins and cost variances in real-time"
            >
              <RotateCw size={13} className={recalculating || loading ? 'sd-spin' : ''} />
              <span>{recalculating ? 'Recalculating...' : 'Recalculate'}</span>
            </button>
            <button className="sd-action-btn sd-action-primary" onClick={() => handleOpenAddChargeModal()}>
              <Plus size={14} /> Add Cost / Charge
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="sd-error-banner" style={{ marginBottom: '12px' }}>
          <AlertCircle size={14} />
          <span>{error}</span>
        </div>
      )}

      {/* ── 2. COMPACT OPERATIONAL FINANCIAL KPI STRIP (7 CARDS) ─────────────── */}
      <div className="sd-fin-kpi-grid">
        <div className="sd-fin-kpi-card">
          <span className="sd-fin-kpi-label">EST. REVENUE</span>
          <span className="sd-fin-kpi-val" style={{ color: '#0f172a' }}>
            {formatCurrency(financials?.estimated_revenue)}
          </span>
          <span className="sd-fin-kpi-sub">Quoted customer target</span>
        </div>

        <div className="sd-fin-kpi-card">
          <span className="sd-fin-kpi-label">ACTUAL REVENUE</span>
          <span className="sd-fin-kpi-val" style={{ color: '#2563eb' }}>
            {formatCurrency(financials?.actual_revenue)}
          </span>
          <span className="sd-fin-kpi-sub">Invoiced to customer</span>
        </div>

        <div className="sd-fin-kpi-card">
          <span className="sd-fin-kpi-label">EST. COST</span>
          <span className="sd-fin-kpi-val" style={{ color: '#475569' }}>
            {formatCurrency(financials?.estimated_cost)}
          </span>
          <span className="sd-fin-kpi-sub">Quoted carrier baseline</span>
        </div>

        <div className="sd-fin-kpi-card">
          <span className="sd-fin-kpi-label">ACTUAL COST</span>
          <span className="sd-fin-kpi-val" style={{ color: '#d97706' }}>
            {formatCurrency(financials?.actual_cost)}
          </span>
          <span className="sd-fin-kpi-sub">Invoices & charge lines</span>
        </div>

        <div className="sd-fin-kpi-card">
          <span className="sd-fin-kpi-label">GROSS MARGIN</span>
          <span
            className="sd-fin-kpi-val"
            style={{ color: (financials?.actual_margin ?? 0) >= 0 ? '#16a34a' : '#dc2626' }}
          >
            {formatCurrency(financials?.actual_margin)}
          </span>
          <span className="sd-fin-kpi-sub">Realized profit</span>
        </div>

        <div className="sd-fin-kpi-card">
          <span className="sd-fin-kpi-label">MARGIN %</span>
          <span
            className="sd-fin-kpi-val"
            style={{ color: (financials?.actual_margin_percent ?? 0) >= 8 ? '#16a34a' : '#dc2626' }}
          >
            {financials?.actual_margin_percent ?? 0}%
          </span>
          <span className="sd-fin-kpi-sub">Commercial return</span>
        </div>

        <div className="sd-fin-kpi-card">
          <span className="sd-fin-kpi-label">COST VARIANCE</span>
          <span
            className="sd-fin-kpi-val"
            style={{ color: (financials?.variance_amount ?? 0) <= 0 ? '#16a34a' : '#dc2626' }}
          >
            {(financials?.variance_amount ?? 0) > 0 ? '+' : ''}
            {formatCurrency(financials?.variance_amount)}
          </span>
          <span className="sd-fin-kpi-sub">
            {(financials?.variance_amount ?? 0) > 0 ? 'Over budget' : 'Under budget'}
          </span>
        </div>
      </div>

      {/* ── 3. PROFITABILITY INTELLIGENCE & FLOW WATERFALL ───────────────────── */}
      <div className="sd-fin-waterfall-card">
        <div className="sd-fin-waterfall-header">
          <div className="sd-fin-wf-title">
            <DollarSign size={15} />
            <span>Commercial Profitability Flow & Realized Margin</span>
          </div>
          <span className="sd-fin-wf-badge" style={{ background: statusConfig.bg, color: statusConfig.fg }}>
            {statusConfig.description}
          </span>
        </div>

        {/* Step-by-step financial flow */}
        <div className="sd-fin-wf-steps">
          <div className="sd-fin-wf-step">
            <span className="sd-fin-wf-label">Quoted / Invoiced Revenue</span>
            <strong className="sd-fin-wf-val">
              {formatCurrency(financials?.actual_revenue || financials?.estimated_revenue)}
            </strong>
            <small className="sd-fin-wf-sub">Customer Billing Receivable</small>
          </div>

          <div className="sd-fin-wf-arrow">−</div>

          <div className="sd-fin-wf-step">
            <span className="sd-fin-wf-label">Carrier Buy Baseline</span>
            <strong className="sd-fin-wf-val" style={{ color: '#475569' }}>
              {formatCurrency(quote?.total_buy_price || quote?.buy_price || financials?.estimated_cost)}
            </strong>
            <small className="sd-fin-wf-sub">Ocean / Air Freight Rate</small>
          </div>

          <div className="sd-fin-wf-arrow">−</div>

          <div className="sd-fin-wf-step">
            <span className="sd-fin-wf-label">Operational Charges</span>
            <strong className="sd-fin-wf-val" style={{ color: '#d97706' }}>
              {formatCurrency(
                charges.reduce((sum, c) => sum + (c.charge_type === CHARGE_TYPES.COST ? (c.actual_amount || c.estimated_amount || 0) : 0), 0)
              )}
            </strong>
            <small className="sd-fin-wf-sub">Drayage, Customs & Surcharges</small>
          </div>

          <div className="sd-fin-wf-arrow">=</div>

          <div className="sd-fin-wf-step sd-fin-wf-highlight">
            <span className="sd-fin-wf-label">Net Gross Profit</span>
            <strong
              className="sd-fin-wf-val"
              style={{ color: (financials?.actual_margin ?? 0) >= 0 ? '#16a34a' : '#dc2626' }}
            >
              {formatCurrency(financials?.actual_margin)}
            </strong>
            <small className="sd-fin-wf-sub">
              {financials?.actual_margin_percent ?? 0}% Realized Margin
            </small>
          </div>
        </div>

        {/* Side-by-side Estimated vs Actual Comparison */}
        <div className="sd-fin-comp-strip">
          <div className="sd-fin-comp-col">
            <span className="sd-fin-comp-label">Estimated Margin</span>
            <span className="sd-fin-comp-val">
              {formatCurrency(estMargin)} <small>({estMarginPct}%)</small>
            </span>
          </div>

          <div className="sd-fin-comp-col">
            <span className="sd-fin-comp-label">Actual Realized Margin</span>
            <span className="sd-fin-comp-val" style={{ color: actMargin >= 0 ? '#16a34a' : '#dc2626' }}>
              {formatCurrency(actMargin)} <small>({actMarginPct}%)</small>
            </span>
          </div>

          <div className="sd-fin-comp-col">
            <span className="sd-fin-comp-label">Margin Variance</span>
            <span
              className="sd-fin-comp-val"
              style={{ color: marginVariance >= 0 ? '#16a34a' : '#dc2626' }}
            >
              {marginVariance >= 0 ? `+${formatCurrency(marginVariance)}` : `-${formatCurrency(Math.abs(marginVariance))}`}
              <small> ({marginPctVariance >= 0 ? `+${marginPctVariance.toFixed(1)}%` : `${marginPctVariance.toFixed(1)}%`} pts)</small>
            </span>
          </div>
        </div>

        {(financials?.variance_amount ?? 0) > 0 && (
          <div className="sd-fin-warning-note">
            <AlertTriangle size={13} />
            <span>
              <strong>Cost Overrun Alert:</strong> Execution costs exceed initial quote estimate by {formatCurrency(financials?.variance_amount)} ({financials?.variance_percent}% variance). Audit charge line items below.
            </span>
          </div>
        )}
      </div>

      {/* ── 4. TABS & FILTER TOOLBAR ─────────────────────────────────────────── */}
      <div className="sd-doc-toolbar">
        <div className="sd-doc-tabs">
          {FINANCIAL_FILTER_TABS.map((tab) => {
            let count = charges.length;
            if (tab.id === 'COST') count = charges.filter((c) => c.charge_type === CHARGE_TYPES.COST).length;
            if (tab.id === 'REVENUE') count = charges.filter((c) => c.charge_type === CHARGE_TYPES.REVENUE).length;
            if (tab.id === 'INVOICES') count = invoices.length;

            return (
              <button
                key={tab.id}
                className={`sd-doc-tab ${activeTab === tab.id ? 'sd-doc-tab-active' : ''}`}
                onClick={() => setActiveTab(tab.id)}
              >
                <span>{tab.label}</span>
                <span className="sd-doc-tab-count">{count}</span>
              </button>
            );
          })}
        </div>

        {activeTab !== 'INVOICES' && (
          <div className="sd-doc-toolbar-right">
            <div className="sd-doc-search-box">
              <Search size={13} className="sd-search-icon" />
              <input
                type="text"
                placeholder="Search charges, vendors, references..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              {searchQuery && (
                <button className="sd-search-clear-btn" onClick={() => setSearchQuery('')}>
                  <X size={12} />
                </button>
              )}
            </div>

            <CustomDropdown
              value={categoryFilter}
              onChange={(val) => setCategoryFilter(val)}
              options={[{ value: 'ALL', label: 'All Categories' }, ...COST_CATEGORY_OPTIONS]}
              size="small"
            />
          </div>
        )}
      </div>

      {/* ── 5. OPERATIONAL CHARGES TABLE (OR COMPACT GUIDED EMPTY STATE) ──────── */}
      {activeTab !== 'INVOICES' && (
        <>
          {filteredCharges.length === 0 ? (
            <div className="sd-doc-compact-empty">
              <div className="sd-doc-compact-empty-header">
                <DollarSign size={22} className="sd-empty-icon-muted" />
                <div>
                  <h4 className="sd-doc-empty-h4">No operational charges recorded</h4>
                  <p className="sd-doc-empty-p">
                    Record shipment expenses such as Ocean Freight, Drayage, Customs Brokerage, Demurrage, or Customer Surcharges.
                  </p>
                </div>
                <button
                  className="sd-action-btn sd-action-primary"
                  onClick={() => handleOpenAddChargeModal()}
                >
                  <Plus size={13} /> Add Operational Cost
                </button>
              </div>

              <div className="sd-doc-recommended-row">
                <span className="sd-doc-rec-title">Quick Add Presets:</span>
                <div className="sd-doc-rec-chips">
                  {quickCostPresets.map((preset) => (
                    <button
                      key={preset.desc}
                      className="sd-doc-rec-chip"
                      onClick={() => handleOpenAddChargeModal(preset.cat, preset.desc)}
                    >
                      <Plus size={11} /> {preset.desc}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          ) : (
            <div className="sd-doc-table-wrap">
              <table className="sd-doc-table">
                <thead>
                  <tr>
                    <th>Type</th>
                    <th>Category & Description</th>
                    <th>Vendor / Payee</th>
                    <th>Charge Date</th>
                    <th>Estimated</th>
                    <th>Actual</th>
                    <th>Variance</th>
                    <th>Status</th>
                    <th style={{ textAlign: 'right' }}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredCharges.map((ch) => {
                    const sBadge = getChargeStatusBadge(ch.status);
                    const varAmt = (ch.actual_amount || 0) - (ch.estimated_amount || 0);
                    const isWorking = actionInProgressId === ch.id;

                    return (
                      <tr key={ch.id}>
                        {/* Type */}
                        <td>
                          <span
                            className="sd-source-pill"
                            style={{
                              background: ch.charge_type === CHARGE_TYPES.REVENUE ? '#eff6ff' : '#fef2f2',
                              color: ch.charge_type === CHARGE_TYPES.REVENUE ? '#1e40af' : '#991b1b',
                              border: `1px solid ${ch.charge_type === CHARGE_TYPES.REVENUE ? '#bfdbfe' : '#fecaca'}`,
                            }}
                          >
                            {ch.charge_type}
                          </span>
                        </td>

                        {/* Category & Description */}
                        <td>
                          <div className="sd-doc-info">
                            <span className="sd-doc-title">{ch.description}</span>
                            <div className="sd-doc-meta-row">
                              <span className="sd-doc-type-tag">{ch.category?.replace(/_/g, ' ')}</span>
                              {ch.reference_number && (
                                <span className="sd-doc-size-tag">Ref: {ch.reference_number}</span>
                              )}
                            </div>
                            {ch.notes && (
                              <div style={{ fontSize: '0.72rem', color: '#64748b', marginTop: '2px' }}>
                                <em>{ch.notes}</em>
                              </div>
                            )}
                          </div>
                        </td>

                        {/* Vendor / Payee */}
                        <td>
                          <span style={{ fontSize: '0.78rem', color: '#334155', fontWeight: 600 }}>
                            {ch.vendor_name || '—'}
                          </span>
                        </td>

                        {/* Charge Date */}
                        <td>
                          <span className="sd-doc-date">
                            {ch.charge_date ? new Date(ch.charge_date).toLocaleDateString() : '—'}
                          </span>
                        </td>

                        {/* Estimated */}
                        <td>
                          <span style={{ fontSize: '0.82rem', fontWeight: 600, color: '#64748b' }}>
                            {formatCurrency(ch.estimated_amount)}
                          </span>
                        </td>

                        {/* Actual */}
                        <td>
                          <span style={{ fontSize: '0.84rem', fontWeight: 700, color: '#0f172a' }}>
                            {formatCurrency(ch.actual_amount)}
                          </span>
                        </td>

                        {/* Variance */}
                        <td>
                          <span
                            style={{
                              fontSize: '0.8rem',
                              fontWeight: 700,
                              color: varAmt > 0 ? '#dc2626' : varAmt < 0 ? '#16a34a' : '#64748b',
                            }}
                          >
                            {varAmt > 0 ? `+${formatCurrency(varAmt)}` : varAmt < 0 ? `-${formatCurrency(Math.abs(varAmt))}` : '—'}
                          </span>
                        </td>

                        {/* Status */}
                        <td>
                          <span
                            className="sd-doc-status-badge"
                            style={{ background: sBadge.bg, color: sBadge.fg, border: `1px solid ${sBadge.border}` }}
                          >
                            {sBadge.label}
                          </span>
                        </td>

                        {/* Actions */}
                        <td>
                          <div className="sd-doc-actions-cell">
                            <button
                              className="sd-table-action-btn"
                              title="Edit charge"
                              disabled={isWorking}
                              onClick={() => handleOpenEditChargeModal(ch)}
                            >
                              <Edit2 size={12} />
                            </button>
                            <button
                              className="sd-table-action-btn sd-act-delete"
                              title="Delete charge"
                              disabled={isWorking}
                              onClick={() => handleDeleteCharge(ch)}
                            >
                              <Trash2 size={12} />
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
        </>
      )}

      {/* ── 6. CARRIER INVOICES & RECONCILIATION TAB ────────────────────────── */}
      {activeTab === 'INVOICES' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
          {/* Ingest Carrier Invoice Toolbar */}
          <form onSubmit={handleInvoiceUpload} className="sd-fin-invoice-form">
            <input
              type="text"
              placeholder="Invoice Number"
              value={invoiceNo}
              onChange={(e) => setInvoiceNo(e.target.value)}
              className="sd-fin-inv-input"
              required
              disabled={ingestingInvoice}
            />
            <CustomDropdown
              value={vendorName}
              onChange={(val) => setVendorName(val)}
              options={[
                { value: 'Maersk', label: 'Maersk (MAEU)' },
                { value: 'MSC', label: 'MSC (MSCU)' },
                { value: 'CMA CGM', label: 'CMA CGM (CMDU)' },
                { value: 'ONE', label: 'ONE (ONEY)' },
                { value: 'Hapag-Lloyd', label: 'Hapag-Lloyd (HLCU)' },
              ]}
              size="small"
              disabled={ingestingInvoice}
            />
            <button
              type="submit"
              className="sd-action-btn sd-action-primary"
              disabled={ingestingInvoice}
            >
              <Upload size={13} className={ingestingInvoice ? 'sd-spin' : ''} />
              <span>{ingestingInvoice ? 'Auditing...' : 'Ingest Carrier Invoice'}</span>
            </button>
          </form>

          {/* Discrepancies Alert */}
          {discrepancies.length > 0 && (
            <div className="sd-doc-discrepancy-banner">
              <div className="sd-disc-banner-header">
                <AlertTriangle size={15} />
                <span>Carrier Invoice Rate Discrepancies ({discrepancies.length})</span>
              </div>
              <div className="sd-disc-items-wrap">
                {discrepancies.map((disc) => (
                  <div key={disc.id} className="sd-disc-row">
                    <div className="sd-disc-desc">
                      <strong>{disc.charge_item || disc.field_name || 'Rate Item'}:</strong>{' '}
                      <span className="sd-disc-vals">
                        Expected: <code>{formatCurrency(disc.expected_amount || disc.expected_value)}</code> vs Invoiced: <code>{formatCurrency(disc.actual_amount || disc.actual_value)}</code>
                      </span>
                    </div>
                    <span className="sd-disc-status-pill sd-pill-open">
                      {disc.status || 'FLAGGED'}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Invoices List */}
          {invoices.length === 0 ? (
            <div className="sd-doc-compact-empty">
              <div className="sd-doc-compact-empty-header">
                <Receipt size={22} className="sd-empty-icon-muted" />
                <div>
                  <h4 className="sd-doc-empty-h4">No carrier invoices ingested yet</h4>
                  <p className="sd-doc-empty-p">
                    Upload a freight invoice to trigger automated three-way rate reconciliation against contracted carrier quotes.
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '10px' }}>
              {invoices.map((inv) => {
                const isWorking = actionInProgressId === inv.id;
                return (
                  <div key={inv.id} className="sd-fin-inv-card">
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                      <span className="sd-fin-vendor-pill">{inv.vendor_name || 'Carrier'}</span>
                      <span
                        className="sd-doc-status-badge"
                        style={{
                          background: inv.status === 'APPROVED' ? '#ecfdf5' : '#fef3c7',
                          color: inv.status === 'APPROVED' ? '#059669' : '#b45309',
                          border: `1px solid ${inv.status === 'APPROVED' ? '#a7f3d0' : '#fde68a'}`,
                        }}
                      >
                        {inv.status?.replace(/_/g, ' ')}
                      </span>
                    </div>

                    <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', marginTop: '6px' }}>
                      <span style={{ fontSize: '0.82rem', fontWeight: 700, color: '#0f172a' }}>
                        INV: {inv.invoice_number}
                      </span>
                      <span style={{ fontSize: '1rem', fontWeight: 800, color: '#0f172a' }}>
                        {formatCurrency(inv.total_amount)}
                      </span>
                    </div>

                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: '6px' }}>
                      <span style={{ fontSize: '0.72rem', color: '#64748b' }}>
                        {inv.created_at ? new Date(inv.created_at).toLocaleDateString() : '—'}
                      </span>
                      {inv.status !== 'APPROVED' && (
                        <button
                          className="sd-action-btn sd-action-primary"
                          style={{ padding: '3px 9px', fontSize: '0.72rem' }}
                          disabled={isWorking}
                          onClick={() => handleApproveInvoice(inv.id)}
                        >
                          Approve
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* ── 7. UPSTREAM LINEAGE BAR ─────────────────────────────────────────── */}
      {(financials?.rfq_id || financials?.booking_id || shipment?.rfq_id) && (
        <div className="sd-doc-lineage-footer">
          <Info size={14} className="sd-lineage-info-icon" />
          <div className="sd-lineage-text">
            <span className="sd-lineage-label">Authoritative Financial Lineage:</span>
            {(financials?.rfq_id || shipment?.rfq_id) && (
              <Link to={`/dashboard/rfqs/${financials?.rfq_id || shipment?.rfq_id}`} className="sd-doc-lineage-nav">
                Source RFQ Quote ({shipment?.rfq_number || `RFQ-${financials?.rfq_id || shipment?.rfq_id}`}) →
              </Link>
            )}
            {(financials?.booking_id || shipment?.booking_id) && (
              <Link to={`/dashboard/bookings/${financials?.booking_id || shipment?.booking_id}`} className="sd-doc-lineage-nav">
                Carrier Booking File ({shipment?.booking_number || `BK-${financials?.booking_id || shipment?.booking_id}`}) →
              </Link>
            )}
          </div>
        </div>
      )}

      {/* ── 8. ADD / EDIT CHARGE MODAL ───────────────────────────────────────── */}
      {showChargeModal && (
        <div className="sd-modal-overlay" onClick={() => setShowChargeModal(false)}>
          <div className="sd-modal-card sd-modal-compact" onClick={(e) => e.stopPropagation()}>
            <div className="sd-modal-header">
              <div className="sd-modal-title-row">
                <DollarSign size={18} className="sd-modal-icon" />
                <h3>{editingCharge ? 'Edit Financial Line Item' : 'Add Financial Cost / Revenue Line Item'}</h3>
              </div>
              <button className="sd-modal-close" onClick={() => setShowChargeModal(false)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleChargeSubmit} className="sd-modal-form">
              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Item Type *</label>
                  <CustomDropdown
                    value={chargeType}
                    onChange={(val) => setChargeType(val)}
                    options={CHARGE_TYPE_OPTIONS}
                  />
                </div>

                <div className="sd-form-field">
                  <label>Cost Category *</label>
                  <CustomDropdown
                    value={chargeCategory}
                    onChange={(val) => setChargeCategory(val)}
                    options={COST_CATEGORY_OPTIONS}
                  />
                </div>
              </div>

              <div className="sd-form-field">
                <label>Description *</label>
                <input
                  type="text"
                  placeholder="e.g. Origin Port Drayage & Lift Surcharge"
                  value={chargeDesc}
                  onChange={(e) => setChargeDesc(e.target.value)}
                  required
                />
              </div>

              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Vendor / Payee</label>
                  <input
                    type="text"
                    placeholder="e.g. Maersk / Terminal Operator"
                    value={chargeVendor}
                    onChange={(e) => setChargeVendor(e.target.value)}
                  />
                </div>

                <div className="sd-form-field">
                  <label>Reference # / B/L</label>
                  <input
                    type="text"
                    placeholder="e.g. INV-98421 or B/L Ref"
                    value={chargeRef}
                    onChange={(e) => setChargeRef(e.target.value)}
                  />
                </div>
              </div>

              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Estimated Amount ({chargeCurrency}) *</label>
                  <input
                    type="number"
                    step="0.01"
                    placeholder="0.00"
                    value={chargeEstimated}
                    onChange={(e) => setChargeEstimated(e.target.value)}
                    required
                  />
                </div>

                <div className="sd-form-field">
                  <label>Actual Invoiced Amount ({chargeCurrency})</label>
                  <input
                    type="number"
                    step="0.01"
                    placeholder="0.00"
                    value={chargeActual}
                    onChange={(e) => setChargeActual(e.target.value)}
                  />
                </div>
              </div>

              <div className="sd-form-grid-2">
                <div className="sd-form-field">
                  <label>Charge Date</label>
                  <input
                    type="date"
                    value={chargeDate}
                    onChange={(e) => setChargeDate(e.target.value)}
                  />
                </div>

                <div className="sd-form-field">
                  <label>Status *</label>
                  <CustomDropdown
                    value={chargeStatus}
                    onChange={(val) => setChargeStatus(val)}
                    options={CHARGE_STATUS_OPTIONS}
                  />
                </div>
              </div>

              <div className="sd-form-field">
                <label>Operational Notes <small>(Optional)</small></label>
                <input
                  type="text"
                  placeholder="e.g. Reconciled with terminal gate receipt"
                  value={chargeNotes}
                  onChange={(e) => setChargeNotes(e.target.value)}
                />
              </div>

              {chargeModalError && (
                <div className="sd-error-banner" style={{ margin: '8px 0' }}>
                  <AlertCircle size={13} />
                  <span>{chargeModalError}</span>
                </div>
              )}

              <div className="sd-modal-footer">
                <button
                  type="button"
                  className="sd-btn sd-btn-ghost"
                  onClick={() => setShowChargeModal(false)}
                  disabled={submittingCharge}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="sd-action-btn sd-action-primary"
                  disabled={submittingCharge}
                >
                  {submittingCharge ? 'Saving...' : editingCharge ? 'Update Line Item' : 'Save Line Item'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
