import React, { useState } from 'react';
import {
  DollarSign,
  TrendingUp,
  Ship,
  Clock,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Plus,
  RefreshCw,
  MoreVertical,
  X,
  ChevronRight,
  Sparkles,
  ShieldCheck,
  Calendar,
  ArrowUpRight,
  Info,
  Percent,
  Award,
  Trash2,
  Eye,
  Edit2,
  FileText,
  Check
} from 'lucide-react';
import { rfqService } from '../../../../services/rfqService';

const STATUS_CONFIG = {
  APPROVED: {
    label: 'Approved',
    badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    iconClass: 'text-emerald-600',
    icon: CheckCircle2,
  },
  RECOMMENDED: {
    label: 'Recommended',
    badgeClass: 'bg-indigo-50 text-indigo-700 border-indigo-200',
    iconClass: 'text-indigo-600',
    icon: Award,
  },
  SELECTED_FOR_CUSTOMER: {
    label: 'Customer Selected',
    badgeClass: 'bg-blue-50 text-blue-700 border-blue-200',
    iconClass: 'text-blue-600',
    icon: Sparkles,
  },
  UNDER_REVIEW: {
    label: 'Under Review',
    badgeClass: 'bg-amber-50 text-amber-700 border-amber-200',
    iconClass: 'text-amber-600',
    icon: Clock,
  },
  RECEIVED: {
    label: 'Received',
    badgeClass: 'bg-slate-100 text-slate-700 border-slate-200',
    iconClass: 'text-slate-600',
    icon: FileText,
  },
  REQUESTED: {
    label: 'Requested',
    badgeClass: 'bg-purple-50 text-purple-700 border-purple-200',
    iconClass: 'text-purple-600',
    icon: Clock,
  },
  DRAFT: {
    label: 'Draft',
    badgeClass: 'bg-slate-50 text-slate-600 border-slate-200',
    iconClass: 'text-slate-500',
    icon: Edit2,
  },
  REJECTED: {
    label: 'Rejected',
    badgeClass: 'bg-rose-50 text-rose-700 border-rose-200',
    iconClass: 'text-rose-600',
    icon: XCircle,
  },
  EXPIRED: {
    label: 'Expired',
    badgeClass: 'bg-rose-50 text-rose-600 border-rose-200',
    iconClass: 'text-rose-500',
    icon: AlertTriangle,
  },
  WITHDRAWN: {
    label: 'Withdrawn',
    badgeClass: 'bg-slate-100 text-slate-500 border-slate-200',
    iconClass: 'text-slate-400',
    icon: XCircle,
  },
};

const CHARGE_TYPES = [
  { value: 'FREIGHT', label: 'Base Ocean Freight' },
  { value: 'ORIGIN_CHARGE', label: 'Origin Terminal Charges (OTHC)' },
  { value: 'DESTINATION_CHARGE', label: 'Destination Terminal Charges (DTHC)' },
  { value: 'FUEL_SURCHARGE', label: 'Bunker Surcharge (BAF)' },
  { value: 'SECURITY_SURCHARGE', label: 'Carrier Security (ISPS)' },
  { value: 'DOCUMENTATION', label: 'Documentation / B/L Fee' },
  { value: 'CUSTOMS', label: 'Customs Clearance' },
  { value: 'INSURANCE', label: 'Cargo Marine Insurance' },
  { value: 'HANDLING', label: 'Handling & Administrative' },
  { value: 'OTHER', label: 'Other Surcharge' },
];

const getCarrierAvatar = (carrierName = '') => {
  const name = (carrierName || '').toLowerCase();
  if (name.includes('maersk')) {
    return { bg: 'bg-sky-50', text: 'text-sky-700', border: 'border-sky-200', code: 'MSK' };
  }
  if (name.includes('hapag')) {
    return { bg: 'bg-orange-50', text: 'text-orange-700', border: 'border-orange-200', code: 'HL' };
  }
  if (name.includes('msc')) {
    return { bg: 'bg-amber-50', text: 'text-amber-700', border: 'border-amber-200', code: 'MSC' };
  }
  if (name.includes('cma')) {
    return { bg: 'bg-blue-50', text: 'text-blue-700', border: 'border-blue-200', code: 'CMA' };
  }
  if (name.includes('evergreen')) {
    return { bg: 'bg-emerald-50', text: 'text-emerald-700', border: 'border-emerald-200', code: 'EMC' };
  }
  if (name.includes('cosco')) {
    return { bg: 'bg-indigo-50', text: 'text-indigo-700', border: 'border-indigo-200', code: 'COS' };
  }
  return { bg: 'bg-slate-100', text: 'text-slate-700', border: 'border-slate-200', code: (carrierName || 'CR').substring(0, 2).toUpperCase() };
};

export default function RFQQuotes({
  rfq,
  quotesData,
  requirements,
  onMutationSuccess,
  onSwitchTab,
  onRefresh,
}) {
  const [activeMenuId, setActiveMenuId] = useState(null);
  const [drawerQuote, setDrawerQuote] = useState(null);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [modalStep, setModalStep] = useState(1);
  const [actionLoading, setActionLoading] = useState(false);
  const [feedbackMessage, setFeedbackMessage] = useState(null);

  // Form State for Add / Edit Quote
  const [formData, setFormData] = useState({
    carrier_name: '',
    carrier_id: '',
    quote_reference: '',
    currency: 'USD',
    buy_price: '',
    sell_price: '',
    ocean_freight: '',
    origin_charges: '',
    destination_charges: '',
    transit_time_days: '',
    free_days: '7',
    valid_from: '',
    valid_until: '',
    etd: '',
    eta: '',
    notes: '',
    charges: [],
  });

  // Current quotes & comparison data
  const quotes = quotesData?.quotes || [];
  const comparison = quotesData?.comparison || [];
  const summary = quotesData?.summary || {
    total_quotes: quotes.length,
    received_quotes: quotes.filter(q => q.status === 'RECEIVED').length,
    primary_currency: 'USD',
  };
  const readiness = quotesData?.rfq_readiness || requirements?.operational_readiness || {
    overall_status: 'READY_FOR_QUOTATION',
    readiness_score: 100,
  };

  // Find comparison metrics map
  const comparisonMap = {};
  comparison.forEach(c => {
    comparisonMap[c.quote_id] = c;
  });

  // Safe Currency Formatter
  const formatMoney = (val, currency = summary.primary_currency || 'USD') => {
    if (val === undefined || val === null || isNaN(val)) return '—';
    return `${currency} ${Number(val).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  const showToast = (msg, type = 'success') => {
    setFeedbackMessage({ text: msg, type });
    setTimeout(() => setFeedbackMessage(null), 4000);
  };

  // Actions
  const handleRecommend = async (quoteId) => {
    try {
      setActionLoading(true);
      await rfqService.recommendQuote(rfq.id, quoteId);
      showToast('Carrier quote marked as recommended operational option.');
      if (onMutationSuccess) onMutationSuccess();
      if (drawerQuote && drawerQuote.id === quoteId) {
        setDrawerQuote(prev => prev ? { ...prev, is_recommended: true, status: 'RECOMMENDED' } : null);
      }
    } catch (err) {
      showToast(err.response?.data?.error?.message || err.message || 'Failed to recommend quote', 'error');
    } finally {
      setActionLoading(false);
      setActiveMenuId(null);
    }
  };

  const handleApprove = async (quoteId) => {
    try {
      setActionLoading(true);
      await rfqService.approveQuote(rfq.id, quoteId, { notes: 'Approved via RFQ Quotes Workspace' });
      showToast('Carrier quote APPROVED! RFQ advanced to Quote Generated stage.');
      if (onMutationSuccess) onMutationSuccess();
      if (drawerQuote && drawerQuote.id === quoteId) {
        setDrawerQuote(prev => prev ? { ...prev, status: 'APPROVED', is_recommended: true } : null);
      }
    } catch (err) {
      showToast(err.response?.data?.error?.message || err.message || 'Failed to approve quote', 'error');
    } finally {
      setActionLoading(false);
      setActiveMenuId(null);
    }
  };

  const handleSelectForCustomer = async (quoteId) => {
    try {
      setActionLoading(true);
      await rfqService.selectQuoteForCustomer(rfq.id, quoteId);
      showToast('Quote selected for customer presentation.');
      if (onMutationSuccess) onMutationSuccess();
      if (drawerQuote && drawerQuote.id === quoteId) {
        setDrawerQuote(prev => prev ? { ...prev, status: 'SELECTED_FOR_CUSTOMER' } : null);
      }
    } catch (err) {
      showToast(err.response?.data?.error?.message || err.message || 'Failed to select quote', 'error');
    } finally {
      setActionLoading(false);
      setActiveMenuId(null);
    }
  };

  const handleStatusChange = async (quoteId, newStatus) => {
    try {
      setActionLoading(true);
      await rfqService.updateQuoteStatus(rfq.id, quoteId, { status: newStatus });
      showToast(`Quote status transitioned to ${newStatus}.`);
      if (onMutationSuccess) onMutationSuccess();
      if (drawerQuote && drawerQuote.id === quoteId) {
        setDrawerQuote(prev => prev ? { ...prev, status: newStatus } : null);
      }
    } catch (err) {
      showToast(err.response?.data?.error?.message || err.message || 'Failed to update quote status', 'error');
    } finally {
      setActionLoading(false);
      setActiveMenuId(null);
    }
  };

  const handleDelete = async (quoteId) => {
    if (!window.confirm('Are you sure you want to withdraw this carrier quote? This preserves audit history.')) {
      return;
    }
    try {
      setActionLoading(true);
      await rfqService.deleteQuote(rfq.id, quoteId);
      showToast('Carrier quote withdrawn.');
      if (onMutationSuccess) onMutationSuccess();
      if (drawerQuote && drawerQuote.id === quoteId) {
        setDrawerQuote(null);
      }
    } catch (err) {
      showToast(err.response?.data?.error?.message || err.message || 'Failed to withdraw quote', 'error');
    } finally {
      setActionLoading(false);
      setActiveMenuId(null);
    }
  };

  const handleCreateQuote = async (e) => {
    e.preventDefault();
    if (!formData.carrier_name.trim()) {
      showToast('Carrier name is required', 'error');
      return;
    }
    const buy = parseFloat(formData.buy_price);
    const sell = parseFloat(formData.sell_price);
    if (isNaN(buy) || buy <= 0 || isNaN(sell) || sell <= 0) {
      showToast('Buy price and Sell price must be numbers greater than 0', 'error');
      return;
    }

    try {
      setActionLoading(true);
      const payload = {
        carrier_name: formData.carrier_name.trim(),
        carrier_id: formData.carrier_id ? formData.carrier_id.trim() : undefined,
        quote_reference: formData.quote_reference ? formData.quote_reference.trim() : undefined,
        currency: formData.currency,
        buy_price: buy,
        sell_price: sell,
        ocean_freight: formData.ocean_freight ? parseFloat(formData.ocean_freight) : undefined,
        origin_charges: formData.origin_charges ? parseFloat(formData.origin_charges) : undefined,
        destination_charges: formData.destination_charges ? parseFloat(formData.destination_charges) : undefined,
        transit_time_days: formData.transit_time_days ? parseInt(formData.transit_time_days, 10) : undefined,
        free_days: formData.free_days ? parseInt(formData.free_days, 10) : undefined,
        valid_from: formData.valid_from ? new Date(formData.valid_from).toISOString() : undefined,
        valid_until: formData.valid_until ? new Date(formData.valid_until).toISOString() : undefined,
        etd: formData.etd ? new Date(formData.etd).toISOString() : undefined,
        eta: formData.eta ? new Date(formData.eta).toISOString() : undefined,
        notes: formData.notes ? formData.notes.trim() : undefined,
        charges: formData.charges.filter(c => c.amount > 0),
      };

      await rfqService.createQuote(rfq.id, payload);
      showToast('Carrier quote added successfully!');
      setIsAddModalOpen(false);
      setModalStep(1);
      setFormData({
        carrier_name: '',
        carrier_id: '',
        quote_reference: '',
        currency: 'USD',
        buy_price: '',
        sell_price: '',
        ocean_freight: '',
        origin_charges: '',
        destination_charges: '',
        transit_time_days: '',
        free_days: '7',
        valid_from: '',
        valid_until: '',
        etd: '',
        eta: '',
        notes: '',
        charges: [],
      });
      if (onMutationSuccess) onMutationSuccess();
    } catch (err) {
      showToast(err.response?.data?.error?.message || err.message || 'Failed to create quote', 'error');
    } finally {
      setActionLoading(false);
    }
  };

  const handleAddChargeItem = () => {
    setFormData(prev => ({
      ...prev,
      charges: [
        ...prev.charges,
        { type: 'OTHER', description: '', amount: 0, currency: prev.currency || 'USD' }
      ]
    }));
  };

  const handleUpdateChargeItem = (idx, field, value) => {
    setFormData(prev => {
      const updated = [...prev.charges];
      updated[idx] = { ...updated[idx], [field]: value };
      return { ...prev, charges: updated };
    });
  };

  const handleRemoveChargeItem = (idx) => {
    setFormData(prev => ({
      ...prev,
      charges: prev.charges.filter((_, i) => i !== idx)
    }));
  };

  // Readiness status badge
  const isReadinessBlocked = readiness.overall_status === 'INFORMATION_REQUIRED' ||
    readiness.overall_status === 'REQUIREMENTS_INCOMPLETE' ||
    readiness.blocking_count > 0;

  if (quotesData === null) {
    return (
      <div className="bg-white border border-slate-200/80 rounded-xl p-12 text-center shadow-xs" data-testid="rfq-quotes-loading">
        <div className="w-12 h-12 bg-indigo-50 border border-indigo-100 rounded-2xl flex items-center justify-center mx-auto mb-3.5 text-indigo-600">
          <RefreshCw className="w-6 h-6 animate-spin text-indigo-600" />
        </div>
        <h3 className="text-sm font-bold text-slate-900 mb-1">Loading Carrier Quotes...</h3>
        <p className="text-xs text-slate-500 max-w-md mx-auto leading-relaxed">
          Synchronizing multi-carrier quotations, live pricing benchmarks, and margin calculations from the operational database.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4" data-testid="rfq-quotes-workspace">

      {/* ── Toast Feedback ── */}
      {feedbackMessage && (
        <div
          className={`p-3 rounded-lg border text-xs font-semibold flex items-center justify-between shadow-xs transition-all ${
            feedbackMessage.type === 'error'
              ? 'bg-rose-50 border-rose-200 text-rose-800'
              : 'bg-emerald-50 border-emerald-200 text-emerald-800'
          }`}
        >
          <div className="flex items-center gap-2">
            {feedbackMessage.type === 'error' ? <AlertTriangle className="w-4 h-4 text-rose-600" /> : <CheckCircle2 className="w-4 h-4 text-emerald-600" />}
            <span>{feedbackMessage.text}</span>
          </div>
          <button onClick={() => setFeedbackMessage(null)} className="text-slate-400 hover:text-slate-600">
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* ── 1. Light, Compact Header ── */}
      <div className="bg-white border border-slate-200/80 rounded-xl px-5 py-4 flex flex-col md:flex-row md:items-center justify-between gap-3 shadow-xs">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <h2 className="text-base font-bold text-slate-900 tracking-tight">
              Carrier Quotes & Commercial Decision
            </h2>
            <span className="px-2 py-0.5 bg-indigo-50 text-indigo-700 border border-indigo-200/70 rounded-full text-[11px] font-semibold">
              {quotes.length} {quotes.length === 1 ? 'Quote' : 'Quotes'} Persisted
            </span>
          </div>
          <p className="text-xs text-slate-500 font-normal">
            Review multi-carrier rate options, evaluate buy/sell margins, and approve the operational booking quote.
          </p>
        </div>

        <div className="flex items-center gap-2 self-start md:self-auto">
          {onRefresh && (
            <button
              onClick={onRefresh}
              className="p-2 border border-slate-200 bg-white hover:bg-slate-50 text-slate-600 rounded-lg text-xs font-semibold transition-colors flex items-center gap-1.5 shadow-2xs"
              title="Refresh Quotes"
            >
              <RefreshCw className="w-3.5 h-3.5 text-slate-500" />
            </button>
          )}

          <button
            onClick={() => {
              setModalStep(1);
              setIsAddModalOpen(true);
            }}
            className="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-bold transition-all shadow-xs flex items-center gap-1.5 cursor-pointer"
            data-testid="add-carrier-quote-btn"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Add Carrier Quote</span>
          </button>
        </div>
      </div>

      {/* ── 2. Authoritative RFQ Operational Readiness Banner ── */}
      {isReadinessBlocked ? (
        <div className="bg-rose-50/70 border border-rose-200 rounded-xl p-3.5 flex items-start justify-between gap-3 shadow-2xs">
          <div className="flex items-start gap-2.5">
            <AlertTriangle className="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-bold text-rose-900">RFQ Requirements Incomplete</span>
                <span className="px-1.5 py-0.5 bg-rose-200/60 text-rose-800 text-[10px] font-bold rounded">
                  Readiness Score: {readiness.readiness_score}%
                </span>
              </div>
              <p className="text-xs text-rose-700 mt-0.5">
                {readiness.next_best_action || 'Complete mandatory trade parameters and compliance documents before final commercial quotation.'}
              </p>
            </div>
          </div>
          {onSwitchTab && (
            <button
              onClick={() => onSwitchTab('requirements')}
              className="px-2.5 py-1.5 bg-rose-100 hover:bg-rose-200 text-rose-800 text-xs font-semibold rounded-lg transition-colors flex items-center gap-1 shrink-0"
            >
              <span>View Requirements</span>
              <ChevronRight className="w-3 h-3" />
            </button>
          )}
        </div>
      ) : (
        <div className="bg-emerald-50/50 border border-emerald-200/70 rounded-xl px-4 py-2.5 flex items-center justify-between text-xs text-emerald-800 shadow-2xs">
          <div className="flex items-center gap-2">
            <ShieldCheck className="w-4 h-4 text-emerald-600" />
            <span className="font-bold">Operational Readiness Complete</span>
            <span className="text-emerald-700">· Trade requirements satisfied for carrier quotation & commercial review.</span>
          </div>
          <span className="text-[11px] font-semibold text-emerald-700">Readiness: 100%</span>
        </div>
      )}

      {/* ── 3. Unified Horizontal KPI Strip (No Visual Clutter) ── */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-2.5">
        
        {/* Total Quotes */}
        <div className="bg-white border border-slate-200/80 rounded-xl p-3 shadow-2xs">
          <div className="flex items-center justify-between text-[11px] font-semibold text-slate-500 mb-1">
            <span>TOTAL QUOTES</span>
            <Ship className="w-3.5 h-3.5 text-slate-400" />
          </div>
          <div className="text-lg font-extrabold text-slate-900 tracking-tight">
            {summary.total_quotes || quotes.length}
          </div>
          <div className="text-[10.5px] text-slate-500 font-medium truncate mt-0.5">
            {summary.received_quotes || 0} Received · {summary.under_review_quotes || 0} Under Review
          </div>
        </div>

        {/* Lowest Buy Cost */}
        <div className="bg-white border border-slate-200/80 rounded-xl p-3 shadow-2xs">
          <div className="flex items-center justify-between text-[11px] font-semibold text-slate-500 mb-1">
            <span>LOWEST BUY</span>
            <DollarSign className="w-3.5 h-3.5 text-blue-500" />
          </div>
          <div className="text-lg font-extrabold text-blue-600 tracking-tight truncate">
            {summary.lowest_buy_amount !== undefined && summary.lowest_buy_amount !== null
              ? formatMoney(summary.lowest_buy_amount)
              : '—'}
          </div>
          <div className="text-[10.5px] text-slate-500 font-medium truncate mt-0.5">
            Best direct carrier rate
          </div>
        </div>

        {/* Best Commercial Margin */}
        <div className="bg-white border border-slate-200/80 rounded-xl p-3 shadow-2xs">
          <div className="flex items-center justify-between text-[11px] font-semibold text-slate-500 mb-1">
            <span>BEST MARGIN</span>
            <TrendingUp className="w-3.5 h-3.5 text-emerald-500" />
          </div>
          <div className="text-lg font-extrabold text-emerald-600 tracking-tight truncate">
            {summary.highest_margin_amount !== undefined && summary.highest_margin_amount !== null
              ? `+${formatMoney(summary.highest_margin_amount)}`
              : '—'}
          </div>
          <div className="text-[10.5px] text-emerald-700 font-semibold truncate mt-0.5">
            {summary.highest_margin_percentage ? `${summary.highest_margin_percentage}% margin` : 'Commercial profit'}
          </div>
        </div>

        {/* Fastest Transit */}
        <div className="bg-white border border-slate-200/80 rounded-xl p-3 shadow-2xs">
          <div className="flex items-center justify-between text-[11px] font-semibold text-slate-500 mb-1">
            <span>FASTEST TRANSIT</span>
            <Clock className="w-3.5 h-3.5 text-amber-500" />
          </div>
          <div className="text-lg font-extrabold text-slate-900 tracking-tight">
            {summary.fastest_transit_days ? `${summary.fastest_transit_days} Days` : '—'}
          </div>
          <div className="text-[10.5px] text-slate-500 font-medium truncate mt-0.5">
            Port to Port transit
          </div>
        </div>

        {/* Expiring Soon */}
        <div className="bg-white border border-slate-200/80 rounded-xl p-3 shadow-2xs">
          <div className="flex items-center justify-between text-[11px] font-semibold text-slate-500 mb-1">
            <span>VALIDITY ALERT</span>
            <AlertTriangle className={`w-3.5 h-3.5 ${summary.quotes_expiring_soon > 0 ? 'text-amber-500' : 'text-slate-400'}`} />
          </div>
          <div className={`text-lg font-extrabold tracking-tight ${summary.quotes_expiring_soon > 0 ? 'text-amber-600' : 'text-slate-900'}`}>
            {summary.quotes_expiring_soon || 0} Expiring
          </div>
          <div className="text-[10.5px] text-slate-500 font-medium truncate mt-0.5">
            {summary.expired_quotes || 0} Expired quotes
          </div>
        </div>

        {/* Decision State */}
        <div className="bg-white border border-slate-200/80 rounded-xl p-3 shadow-2xs">
          <div className="flex items-center justify-between text-[11px] font-semibold text-slate-500 mb-1">
            <span>DECISION STATE</span>
            <Award className="w-3.5 h-3.5 text-indigo-500" />
          </div>
          <div className="text-xs font-bold text-slate-900 tracking-tight truncate mt-1">
            {summary.approved_quote_id ? (
              <span className="text-emerald-700 flex items-center gap-1">
                <CheckCircle2 className="w-3.5 h-3.5 shrink-0" />
                <span>Quote Approved</span>
              </span>
            ) : summary.recommended_quote_id ? (
              <span className="text-indigo-700 flex items-center gap-1">
                <Award className="w-3.5 h-3.5 shrink-0" />
                <span>Recommended</span>
              </span>
            ) : (
              <span className="text-slate-500">Pending Review</span>
            )}
          </div>
          <div className="text-[10.5px] text-slate-500 font-medium truncate mt-0.5">
            {summary.approved_quote_id ? 'Ready for booking' : 'Selection required'}
          </div>
        </div>

      </div>

      {/* Currency warning if mixed currencies present */}
      {summary.has_mixed_currencies && (
        <div className="bg-amber-50/70 border border-amber-200/80 rounded-lg px-3.5 py-2 text-xs text-amber-800 flex items-center gap-2 font-medium">
          <Info className="w-4 h-4 text-amber-600 shrink-0" />
          <span>Quotes contain multiple currencies. Benchmarks are calculated based on dominant currency ({summary.primary_currency}).</span>
        </div>
      )}

      {/* ── 4. Primary Workspace Table ── */}
      {quotes.length === 0 ? (
        <div className="bg-white border border-slate-200/80 rounded-xl p-12 text-center shadow-xs">
          <div className="w-12 h-12 bg-indigo-50 border border-indigo-100 rounded-2xl flex items-center justify-center mx-auto mb-3.5 text-indigo-600">
            <Ship className="w-6 h-6" />
          </div>
          <h3 className="text-sm font-bold text-slate-900 mb-1">No Carrier Quotes Recorded</h3>
          <p className="text-xs text-slate-500 max-w-md mx-auto mb-5 leading-relaxed">
            There are currently no persistent carrier quotations for this RFQ. Request SPOT rates from ocean liners or add carrier quotation details manually.
          </p>
          <button
            onClick={() => {
              setModalStep(1);
              setIsAddModalOpen(true);
            }}
            className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-bold transition-colors inline-flex items-center gap-1.5 shadow-xs cursor-pointer"
            data-testid="empty-add-quote-btn"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Add First Carrier Quote</span>
          </button>
        </div>
      ) : (
        <div className="bg-white border border-slate-200 rounded-xl overflow-hidden shadow-xs">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse" data-testid="rfq-quotes-table">
              <thead>
                <tr className="bg-slate-50/90 border-b border-slate-200 text-[11px] font-bold text-slate-500 uppercase tracking-wider">
                  <th className="py-3.5 px-4 w-[19%]">Carrier & Ref</th>
                  <th className="py-3.5 px-2.5 w-[11%]">Status</th>
                  <th className="py-3.5 px-2.5 text-right w-[9%]">Buy Cost</th>
                  <th className="py-3.5 px-2.5 text-right w-[9%]">Sell Price</th>
                  <th className="py-3.5 px-2.5 text-right w-[10%]">Margin</th>
                  <th className="py-3.5 px-2.5 w-[9%]">Transit & ETD</th>
                  <th className="py-3.5 px-2.5 w-[8%]">Validity</th>
                  <th className="py-3.5 px-2.5 w-[12%]">Intelligence</th>
                  <th className="py-3.5 px-3 text-center w-[13%]">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-xs text-slate-800">
                {quotes.map((quote) => {
                  const comp = comparisonMap[quote.id] || {};
                  const statusInfo = STATUS_CONFIG[quote.status] || STATUS_CONFIG.DRAFT;
                  const StatusIcon = statusInfo.icon;
                  const isApproved = quote.status === 'APPROVED' || quote.status === 'SELECTED_FOR_CUSTOMER';
                  const isRecommended = quote.is_recommended || quote.status === 'RECOMMENDED';
                  const isExpiringSoon = quote.validity_status === 'EXPIRING_SOON';
                  const isExpired = quote.validity_status === 'EXPIRED';
                  const carrierAvatar = getCarrierAvatar(quote.carrier_name);

                  return (
                    <tr
                      key={quote.id}
                      className={`transition-colors border-b border-slate-100 last:border-0 ${
                        isApproved
                          ? 'bg-emerald-50/30 border-l-[3.5px] border-l-emerald-500 hover:bg-emerald-50/50'
                          : isRecommended
                          ? 'bg-indigo-50/30 border-l-[3.5px] border-l-indigo-500 hover:bg-indigo-50/50'
                          : 'border-l-[3.5px] border-l-transparent hover:bg-slate-50/80'
                      }`}
                      data-testid={`quote-row-${quote.id}`}
                    >
                      {/* Carrier & Reference */}
                      <td className="py-3.5 px-4 align-middle">
                        <div className="flex items-center gap-2.5">
                          <div className={`w-8 h-8 rounded-lg border ${carrierAvatar.bg} ${carrierAvatar.border} ${carrierAvatar.text} flex items-center justify-center font-extrabold text-[11px] shrink-0 shadow-2xs`}>
                            {carrierAvatar.code}
                          </div>
                          <div className="min-w-0">
                            <div className="flex items-center gap-1.5 flex-nowrap">
                              <span className="font-bold text-slate-900 text-[13px] whitespace-nowrap">
                                {quote.carrier_name}
                              </span>
                              {quote.carrier_id && (
                                <span className="px-1 py-0.2 bg-slate-100 text-slate-600 text-[9.5px] font-mono font-bold rounded border border-slate-200/60 whitespace-nowrap">
                                  {quote.carrier_id}
                                </span>
                              )}
                            </div>
                            <div className="text-[11px] text-slate-500 font-mono tracking-tight whitespace-nowrap mt-0.5">
                              {quote.quote_reference || 'Ref: N/A'}
                            </div>
                          </div>
                        </div>
                      </td>

                      {/* Status */}
                      <td className="py-3.5 px-2.5 align-middle">
                        <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-bold border whitespace-nowrap ${statusInfo.badgeClass}`}>
                          <StatusIcon className={`w-3 h-3 shrink-0 ${statusInfo.iconClass}`} />
                          <span>{statusInfo.label}</span>
                        </span>
                      </td>

                      {/* Buy Cost */}
                      <td className="py-3.5 px-2.5 text-right align-middle">
                        <div className="font-extrabold text-slate-900 text-[13px] tabular-nums whitespace-nowrap">
                          {formatMoney(quote.buy_price, quote.currency)}
                        </div>
                        {quote.ocean_freight > 0 ? (
                          <div className="text-[10px] text-slate-400 font-medium whitespace-nowrap mt-0.5">
                            Freight: {formatMoney(quote.ocean_freight, quote.currency)}
                          </div>
                        ) : (
                          <div className="text-[10px] text-slate-400 font-medium whitespace-nowrap mt-0.5">
                            Direct Buy
                          </div>
                        )}
                      </td>

                      {/* Sell Price */}
                      <td className="py-3.5 px-2.5 text-right align-middle">
                        <div className="font-extrabold text-slate-900 text-[13px] tabular-nums whitespace-nowrap">
                          {formatMoney(quote.sell_price, quote.currency)}
                        </div>
                        <div className="text-[10px] text-slate-400 font-medium whitespace-nowrap mt-0.5">
                          Target Sell
                        </div>
                      </td>

                      {/* Margin */}
                      <td className="py-3.5 px-2.5 text-right align-middle">
                        <div className={`font-extrabold text-[13px] tabular-nums whitespace-nowrap ${quote.margin_amount >= 0 ? 'text-emerald-600' : 'text-rose-600'}`}>
                          {quote.margin_amount >= 0 ? `+${formatMoney(quote.margin_amount, quote.currency)}` : formatMoney(quote.margin_amount, quote.currency)}
                        </div>
                        <div className="mt-0.5">
                          <span
                            className={`inline-flex items-center px-1.5 py-0.2 rounded text-[10px] font-extrabold whitespace-nowrap ${
                              quote.margin_percentage >= 15
                                ? 'bg-emerald-100 text-emerald-800 border border-emerald-200/60'
                                : quote.margin_percentage > 0
                                ? 'bg-slate-100 text-slate-700 border border-slate-200'
                                : 'bg-rose-100 text-rose-800 border border-rose-200'
                            }`}
                          >
                            {quote.margin_percentage}% Margin
                          </span>
                        </div>
                      </td>

                      {/* Transit & ETD */}
                      <td className="py-3.5 px-2.5 align-middle">
                        <div className="font-bold text-slate-900 text-[12px] flex items-center gap-1 whitespace-nowrap">
                          <Clock className="w-3 h-3 text-slate-400 shrink-0" />
                          <span>{quote.transit_time_days ? `${quote.transit_time_days} Days` : '—'}</span>
                        </div>
                        {quote.etd ? (
                          <div className="text-[10.5px] text-slate-500 font-medium whitespace-nowrap mt-0.5">
                            ETD: {new Date(quote.etd).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}
                          </div>
                        ) : quote.free_days ? (
                          <div className="text-[10px] text-slate-400 font-medium whitespace-nowrap mt-0.5">
                            {quote.free_days} Free Days
                          </div>
                        ) : (
                          <div className="text-[10px] text-slate-400 font-medium whitespace-nowrap mt-0.5">
                            Standard Loop
                          </div>
                        )}
                      </td>

                      {/* Validity */}
                      <td className="py-3.5 px-2.5 align-middle">
                        {quote.valid_until ? (
                          <div>
                            <div className="text-slate-800 font-semibold text-xs whitespace-nowrap">
                              {new Date(quote.valid_until).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}
                            </div>
                            <div className="mt-0.5">
                              {isExpired ? (
                                <span className="inline-flex items-center px-1.5 py-0.2 rounded bg-rose-100 text-rose-700 border border-rose-200 text-[9.5px] font-extrabold whitespace-nowrap">
                                  EXPIRED
                                </span>
                              ) : isExpiringSoon ? (
                                <span className="inline-flex items-center px-1.5 py-0.2 rounded bg-amber-100 text-amber-800 border border-amber-200 text-[9.5px] font-extrabold whitespace-nowrap">
                                  {quote.days_until_expiry}D LEFT
                                </span>
                              ) : (
                                <span className="inline-flex items-center px-1.5 py-0.2 rounded bg-emerald-50 text-emerald-700 border border-emerald-200/60 text-[9.5px] font-bold whitespace-nowrap">
                                  Valid
                                </span>
                              )}
                            </div>
                          </div>
                        ) : (
                          <div>
                            <span className="text-slate-400 text-[11px] font-medium whitespace-nowrap">No Expiry</span>
                            <div className="text-[10px] text-slate-400 mt-0.5 whitespace-nowrap">Standard SPOT</div>
                          </div>
                        )}
                      </td>

                      {/* Intelligence & Decision */}
                      <td className="py-3.5 px-2.5 align-middle">
                        <div className="flex flex-wrap items-center gap-1 mb-0.5">
                          {comp.is_lowest_cost && (
                            <span className="px-1.5 py-0.2 bg-blue-50 text-blue-700 border border-blue-200/70 rounded text-[9.5px] font-extrabold flex items-center gap-0.5 whitespace-nowrap">
                              <span>🏆 Lowest Buy</span>
                            </span>
                          )}
                          {comp.is_highest_margin && (
                            <span className="px-1.5 py-0.2 bg-emerald-50 text-emerald-700 border border-emerald-200/70 rounded text-[9.5px] font-extrabold flex items-center gap-0.5 whitespace-nowrap">
                              <span>📈 Best Margin</span>
                            </span>
                          )}
                          {comp.is_fastest && (
                            <span className="px-1.5 py-0.2 bg-amber-50 text-amber-700 border border-amber-200/70 rounded text-[9.5px] font-extrabold flex items-center gap-0.5 whitespace-nowrap">
                              <span>⚡ Fastest</span>
                            </span>
                          )}
                          {isApproved && (
                            <span className="px-1.5 py-0.2 bg-emerald-100 text-emerald-800 border border-emerald-300 rounded text-[9.5px] font-extrabold flex items-center gap-0.5 whitespace-nowrap">
                              <span>✓ Approved</span>
                            </span>
                          )}
                        </div>
                        <div className="text-[10.5px] text-slate-500 leading-tight line-clamp-1" title={comp.recommendation_reason || quote.ai_reasoning}>
                          {comp.recommendation_reason || quote.ai_reasoning || 'Standard carrier quotation record.'}
                        </div>
                      </td>

                      {/* Actions */}
                      <td className="py-3.5 px-3 text-center align-middle">
                        <div className="flex items-center justify-center gap-1.5 whitespace-nowrap relative">
                          <button
                            onClick={() => setDrawerQuote(quote)}
                            className="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-md text-[11px] font-bold transition-all shadow-2xs cursor-pointer"
                            data-testid={`view-quote-btn-${quote.id}`}
                          >
                            Details
                          </button>

                          {/* Quick Decision Action Buttons */}
                          {!isApproved && !isExpired && (
                            <>
                              {!isRecommended ? (
                                <button
                                  disabled={actionLoading}
                                  onClick={() => handleRecommend(quote.id)}
                                  className="px-2 py-1 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border border-indigo-200 rounded-md text-[11px] font-bold transition-all disabled:opacity-50 cursor-pointer"
                                  data-testid={`recommend-btn-${quote.id}`}
                                >
                                  Recommend
                                </button>
                              ) : (
                                <button
                                  disabled={actionLoading}
                                  onClick={() => handleApprove(quote.id)}
                                  className="px-2.5 py-1 bg-emerald-600 hover:bg-emerald-700 text-white rounded-md text-[11px] font-bold shadow-2xs transition-all disabled:opacity-50 cursor-pointer"
                                  data-testid={`approve-btn-${quote.id}`}
                                >
                                  Approve
                                </button>
                              )}
                            </>
                          )}

                          {/* More Options Dropdown */}
                          <div className="relative">
                            <button
                              onClick={() => setActiveMenuId(activeMenuId === quote.id ? null : quote.id)}
                              className="p-1 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-md transition-colors cursor-pointer"
                              data-testid={`quote-menu-btn-${quote.id}`}
                            >
                              <MoreVertical className="w-3.5 h-3.5" />
                            </button>

                            {activeMenuId === quote.id && (
                              <div
                                className="absolute right-0 top-full mt-1 w-48 bg-white rounded-xl shadow-xl border border-slate-200 py-1.5 z-30 text-left animate-in fade-in zoom-in-95 duration-100"
                                onClick={(e) => e.stopPropagation()}
                              >
                                <button
                                  onClick={() => {
                                    setDrawerQuote(quote);
                                    setActiveMenuId(null);
                                  }}
                                  className="w-full px-3.5 py-2 text-xs text-slate-700 hover:bg-slate-50 flex items-center gap-2 font-medium"
                                >
                                  <Eye className="w-3.5 h-3.5 text-slate-400" />
                                  <span>View Deep Breakdown</span>
                                </button>

                                {!isApproved && (
                                  <>
                                    <button
                                      onClick={() => handleStatusChange(quote.id, 'UNDER_REVIEW')}
                                      className="w-full px-3.5 py-2 text-xs text-slate-700 hover:bg-slate-50 flex items-center gap-2 font-medium"
                                    >
                                      <Clock className="w-3.5 h-3.5 text-amber-500" />
                                      <span>Mark Under Review</span>
                                    </button>
                                    <button
                                      onClick={() => handleApprove(quote.id)}
                                      className="w-full px-3.5 py-2 text-xs text-emerald-700 hover:bg-emerald-50 flex items-center gap-2 font-bold"
                                    >
                                      <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                                      <span>Approve Operational Quote</span>
                                    </button>
                                  </>
                                )}

                                {isApproved && (
                                  <button
                                    onClick={() => handleSelectForCustomer(quote.id)}
                                    className="w-full px-3.5 py-2 text-xs text-blue-700 hover:bg-blue-50 flex items-center gap-2 font-bold"
                                  >
                                    <Sparkles className="w-3.5 h-3.5 text-blue-600" />
                                    <span>Select for Customer</span>
                                  </button>
                                )}

                                <div className="border-t border-slate-100 my-1" />

                                <button
                                  onClick={() => handleStatusChange(quote.id, 'REJECTED')}
                                  className="w-full px-3.5 py-2 text-xs text-rose-600 hover:bg-rose-50 flex items-center gap-2 font-medium"
                                >
                                  <XCircle className="w-3.5 h-3.5" />
                                  <span>Reject Quote</span>
                                </button>

                                <button
                                  onClick={() => handleDelete(quote.id)}
                                  className="w-full px-3.5 py-2 text-xs text-slate-500 hover:bg-slate-50 flex items-center gap-2 font-medium"
                                >
                                  <Trash2 className="w-3.5 h-3.5" />
                                  <span>Withdraw Quote</span>
                                </button>
                              </div>
                            )}
                          </div>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}


      {/* ── 5. Right-Side Deep Information Slide-Over Drawer ── */}
      {drawerQuote && (
        <div className="rfq-drawer-overlay" onClick={() => setDrawerQuote(null)}>
          <div className="rfq-drawer-panel p-6 space-y-5" onClick={(e) => e.stopPropagation()} data-testid="quote-detail-drawer">
            
            {/* Drawer Header */}
            <div className="flex items-start justify-between border-b border-slate-200/80 pb-4">
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="text-base font-bold text-slate-900">
                    {drawerQuote.carrier_name}
                  </h3>
                  <span className={`px-2 py-0.5 rounded-full text-[10.5px] font-bold border ${STATUS_CONFIG[drawerQuote.status]?.badgeClass || 'bg-slate-100'}`}>
                    {STATUS_CONFIG[drawerQuote.status]?.label || drawerQuote.status}
                  </span>
                </div>
                <p className="text-xs text-slate-500 font-mono">
                  Quote Ref: {drawerQuote.quote_reference || 'N/A'} · Currency: {drawerQuote.currency}
                </p>
              </div>
              <button
                onClick={() => setDrawerQuote(null)}
                className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Section 1: Commercial Summary Cards */}
            <div>
              <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2.5 flex items-center gap-1.5">
                <DollarSign className="w-3.5 h-3.5 text-indigo-600" />
                <span>Commercial Margin Breakdown</span>
              </h4>

              <div className="grid grid-cols-3 gap-2.5">
                <div className="bg-slate-50 border border-slate-200 rounded-lg p-2.5 text-center">
                  <div className="text-[10.5px] font-semibold text-slate-500 mb-0.5">BUY COST</div>
                  <div className="text-sm font-extrabold text-slate-900">
                    {formatMoney(drawerQuote.buy_price, drawerQuote.currency)}
                  </div>
                </div>

                <div className="bg-slate-50 border border-slate-200 rounded-lg p-2.5 text-center">
                  <div className="text-[10.5px] font-semibold text-slate-500 mb-0.5">SELL PRICE</div>
                  <div className="text-sm font-extrabold text-slate-900">
                    {formatMoney(drawerQuote.sell_price, drawerQuote.currency)}
                  </div>
                </div>

                <div className="bg-emerald-50 border border-emerald-200 rounded-lg p-2.5 text-center">
                  <div className="text-[10.5px] font-bold text-emerald-700 mb-0.5">NET MARGIN</div>
                  <div className="text-sm font-extrabold text-emerald-700">
                    +{formatMoney(drawerQuote.margin_amount, drawerQuote.currency)}
                  </div>
                  <div className="text-[10px] font-bold text-emerald-600 mt-0.5">
                    {drawerQuote.margin_percentage}% Margin
                  </div>
                </div>
              </div>
            </div>

            {/* Section 2: Itemized Charges */}
            <div>
              <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2 flex items-center justify-between">
                <span>Itemized Surcharges & Fees</span>
                <span className="text-[11px] font-normal text-slate-500">
                  {drawerQuote.charges?.length || 0} Charges
                </span>
              </h4>

              {drawerQuote.charges && drawerQuote.charges.length > 0 ? (
                <div className="border border-slate-200 rounded-lg divide-y divide-slate-100 overflow-hidden text-xs">
                  {drawerQuote.charges.map((charge, idx) => (
                    <div key={idx} className="p-2.5 flex items-center justify-between bg-white hover:bg-slate-50">
                      <div>
                        <div className="font-semibold text-slate-800">{charge.description || charge.type}</div>
                        <div className="text-[10.5px] text-slate-400">{charge.type}</div>
                      </div>
                      <div className="font-bold text-slate-900">
                        {formatMoney(charge.amount, charge.currency || drawerQuote.currency)}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="border border-slate-200 rounded-lg p-3 text-center text-xs text-slate-500 bg-slate-50/50">
                  No individual itemized surcharges listed. Total buy price includes all ocean freight components.
                </div>
              )}
            </div>

            {/* Section 3: Operational Schedule & Free Time */}
            <div>
              <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2.5 flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5 text-indigo-600" />
                <span>Operational Schedule & Validity</span>
              </h4>

              <div className="grid grid-cols-2 gap-2 text-xs">
                <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                  <span className="text-[11px] text-slate-500 block">Port-to-Port Transit</span>
                  <span className="font-bold text-slate-900">
                    {drawerQuote.transit_time_days ? `${drawerQuote.transit_time_days} Days` : 'Standard Transit'}
                  </span>
                </div>

                <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                  <span className="text-[11px] text-slate-500 block">Free Days (Demurrage)</span>
                  <span className="font-bold text-slate-900">
                    {drawerQuote.free_days ? `${drawerQuote.free_days} Free Days` : '7 Free Days'}
                  </span>
                </div>

                <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                  <span className="text-[11px] text-slate-500 block">Target ETD</span>
                  <span className="font-bold text-slate-900">
                    {drawerQuote.etd ? new Date(drawerQuote.etd).toLocaleDateString() : '—'}
                  </span>
                </div>

                <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                  <span className="text-[11px] text-slate-500 block">Target ETA</span>
                  <span className="font-bold text-slate-900">
                    {drawerQuote.eta ? new Date(drawerQuote.eta).toLocaleDateString() : '—'}
                  </span>
                </div>
              </div>
            </div>

            {/* Section 4: Decision Intelligence & Audit Trail */}
            <div>
              <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2 flex items-center gap-1.5">
                <Sparkles className="w-3.5 h-3.5 text-indigo-600" />
                <span>Comparative Intelligence & Decision Notes</span>
              </h4>

              <div className="p-3 bg-indigo-50/50 border border-indigo-100 rounded-lg text-xs text-indigo-900 space-y-1.5">
                <div className="font-semibold flex items-center gap-1.5">
                  <Award className="w-3.5 h-3.5 text-indigo-600" />
                  <span>Recommendation Analysis</span>
                </div>
                <p className="text-indigo-800 leading-relaxed">
                  {comparisonMap[drawerQuote.id]?.recommendation_reason || drawerQuote.ai_reasoning || 'Deterministic commercial quote.'}
                </p>
                {drawerQuote.reliability_score && (
                  <div className="text-[11px] text-indigo-700 font-medium pt-1 border-t border-indigo-200/50 flex items-center justify-between">
                    <span>Carrier Reliability: {drawerQuote.reliability_score}%</span>
                    <span>Historical Success: {drawerQuote.historical_success_rate || 92.5}%</span>
                  </div>
                )}
              </div>

              {drawerQuote.approved_by && (
                <div className="mt-2 p-2.5 bg-emerald-50 border border-emerald-200 rounded-lg text-xs text-emerald-800 flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                  <div>
                    <div className="font-bold">Approved by {drawerQuote.approved_by}</div>
                    <div className="text-[10.5px] text-emerald-700">
                      {drawerQuote.approved_at ? new Date(drawerQuote.approved_at).toLocaleString() : ''}
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Drawer Actions */}
            <div className="pt-4 border-t border-slate-200 flex flex-col gap-2">
              {drawerQuote.status !== 'APPROVED' && (
                <button
                  disabled={actionLoading}
                  onClick={() => handleApprove(drawerQuote.id)}
                  className="w-full py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-bold shadow-xs transition-colors flex items-center justify-center gap-1.5 disabled:opacity-50 cursor-pointer"
                  data-testid="drawer-approve-btn"
                >
                  <CheckCircle2 className="w-4 h-4" />
                  <span>Approve This Carrier Quote</span>
                </button>
              )}

              {!drawerQuote.is_recommended && drawerQuote.status !== 'APPROVED' && (
                <button
                  disabled={actionLoading}
                  onClick={() => handleRecommend(drawerQuote.id)}
                  className="w-full py-2.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border border-indigo-200 rounded-lg text-xs font-bold transition-colors flex items-center justify-center gap-1.5 disabled:opacity-50 cursor-pointer"
                >
                  <Award className="w-4 h-4" />
                  <span>Set as Recommended Option</span>
                </button>
              )}

              {drawerQuote.status === 'APPROVED' && (
                <button
                  disabled={actionLoading}
                  onClick={() => handleSelectForCustomer(drawerQuote.id)}
                  className="w-full py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-bold shadow-xs transition-colors flex items-center justify-center gap-1.5 disabled:opacity-50 cursor-pointer"
                >
                  <Sparkles className="w-4 h-4" />
                  <span>Select for Customer Quotation</span>
                </button>
              )}
            </div>

          </div>
        </div>
      )}

      {/* ── 6. 4-Step Progressive Add Carrier Quote Modal ── */}
      {isAddModalOpen && (
        <div className="rfq-modal-overlay" onClick={() => setIsAddModalOpen(false)}>
          <div className="rfq-modal-card max-w-xl p-6" onClick={(e) => e.stopPropagation()} data-testid="add-quote-modal">
            
            {/* Modal Header */}
            <div className="flex items-center justify-between border-b border-slate-200 pb-3.5 mb-4">
              <div>
                <h3 className="text-base font-bold text-slate-900">Add Carrier Quote</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Step {modalStep} of 4 · {
                    modalStep === 1 ? 'Carrier & Identification' :
                    modalStep === 2 ? 'Commercial Pricing & Surcharges' :
                    modalStep === 3 ? 'Operational Schedule' : 'Notes & Review'
                  }
                </p>
              </div>
              <button
                onClick={() => setIsAddModalOpen(false)}
                className="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Step Progress Bar */}
            <div className="grid grid-cols-4 gap-1.5 mb-5">
              {[1, 2, 3, 4].map(s => (
                <div
                  key={s}
                  className={`h-1.5 rounded-full transition-colors ${
                    s <= modalStep ? 'bg-indigo-600' : 'bg-slate-200'
                  }`}
                />
              ))}
            </div>

            <form onSubmit={handleCreateQuote} className="space-y-4">

              {/* Step 1: Carrier & Reference */}
              {modalStep === 1 && (
                <div className="space-y-3.5">
                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">
                      Carrier Name <span className="text-rose-500">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. Maersk Line, Hapag-Lloyd, MSC, CMA CGM"
                      value={formData.carrier_name}
                      onChange={(e) => setFormData(prev => ({ ...prev, carrier_name: e.target.value }))}
                      className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500"
                      data-testid="input-carrier-name"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Carrier Code / ID (SCAC)
                      </label>
                      <input
                        type="text"
                        placeholder="e.g. MAEU, HLCU, MSCU"
                        value={formData.carrier_id}
                        onChange={(e) => setFormData(prev => ({ ...prev, carrier_id: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs font-mono focus:outline-none focus:border-indigo-500"
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Carrier Quote Reference #
                      </label>
                      <input
                        type="text"
                        placeholder="e.g. QT-2026-0899"
                        value={formData.quote_reference}
                        onChange={(e) => setFormData(prev => ({ ...prev, quote_reference: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs font-mono focus:outline-none focus:border-indigo-500"
                        data-testid="input-quote-reference"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">
                      Quotation Currency
                    </label>
                    <select
                      value={formData.currency}
                      onChange={(e) => setFormData(prev => ({ ...prev, currency: e.target.value }))}
                      className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500 bg-white"
                    >
                      <option value="USD">USD ($)</option>
                      <option value="EUR">EUR (€)</option>
                      <option value="GBP">GBP (£)</option>
                      <option value="SGD">SGD (S$)</option>
                      <option value="INR">INR (₹)</option>
                    </select>
                  </div>
                </div>
              )}

              {/* Step 2: Commercial Pricing & Charges */}
              {modalStep === 2 && (
                <div className="space-y-3.5">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Total Buy Cost ({formData.currency}) <span className="text-rose-500">*</span>
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        required
                        placeholder="0.00"
                        value={formData.buy_price}
                        onChange={(e) => setFormData(prev => ({ ...prev, buy_price: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs font-bold focus:outline-none focus:border-indigo-500"
                        data-testid="input-buy-price"
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Target Sell Price ({formData.currency}) <span className="text-rose-500">*</span>
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        required
                        placeholder="0.00"
                        value={formData.sell_price}
                        onChange={(e) => setFormData(prev => ({ ...prev, sell_price: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs font-bold focus:outline-none focus:border-indigo-500"
                        data-testid="input-sell-price"
                      />
                    </div>
                  </div>

                  {/* Dynamic Margin Preview */}
                  {formData.buy_price > 0 && formData.sell_price > 0 && (
                    <div className="p-2.5 bg-slate-50 border border-slate-200 rounded-lg flex items-center justify-between text-xs">
                      <span className="text-slate-600 font-medium">Estimated Margin:</span>
                      <span className="font-bold text-emerald-700">
                        +{formatMoney(parseFloat(formData.sell_price) - parseFloat(formData.buy_price), formData.currency)} (
                        {(((parseFloat(formData.sell_price) - parseFloat(formData.buy_price)) / parseFloat(formData.sell_price)) * 100).toFixed(1)}%)
                      </span>
                    </div>
                  )}

                  {/* Itemized Charges Sub-section */}
                  <div className="pt-2 border-t border-slate-200">
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-xs font-bold text-slate-700">Itemized Surcharges (Optional)</span>
                      <button
                        type="button"
                        onClick={handleAddChargeItem}
                        className="text-[11px] text-indigo-600 hover:text-indigo-800 font-semibold flex items-center gap-1"
                      >
                        <Plus className="w-3 h-3" />
                        <span>Add Charge</span>
                      </button>
                    </div>

                    {formData.charges.map((charge, idx) => (
                      <div key={idx} className="flex items-center gap-2 mb-2">
                        <select
                          value={charge.type}
                          onChange={(e) => handleUpdateChargeItem(idx, 'type', e.target.value)}
                          className="px-2 py-1.5 border border-slate-300 rounded text-xs w-1/3 bg-white"
                        >
                          {CHARGE_TYPES.map(t => (
                            <option key={t.value} value={t.value}>{t.label}</option>
                          ))}
                        </select>
                        <input
                          type="text"
                          placeholder="Description"
                          value={charge.description}
                          onChange={(e) => handleUpdateChargeItem(idx, 'description', e.target.value)}
                          className="px-2 py-1.5 border border-slate-300 rounded text-xs flex-1"
                        />
                        <input
                          type="number"
                          step="0.01"
                          placeholder="Amount"
                          value={charge.amount || ''}
                          onChange={(e) => handleUpdateChargeItem(idx, 'amount', parseFloat(e.target.value) || 0)}
                          className="px-2 py-1.5 border border-slate-300 rounded text-xs w-24 font-mono text-right"
                        />
                        <button
                          type="button"
                          onClick={() => handleRemoveChargeItem(idx)}
                          className="p-1 text-rose-500 hover:text-rose-700"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Step 3: Operational Schedule & Validity */}
              {modalStep === 3 && (
                <div className="space-y-3.5">
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Transit Time (Days)
                      </label>
                      <input
                        type="number"
                        placeholder="e.g. 21"
                        value={formData.transit_time_days}
                        onChange={(e) => setFormData(prev => ({ ...prev, transit_time_days: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500"
                        data-testid="input-transit-days"
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Free Days (Demurrage)
                      </label>
                      <input
                        type="number"
                        placeholder="7"
                        value={formData.free_days}
                        onChange={(e) => setFormData(prev => ({ ...prev, free_days: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Quote Valid From
                      </label>
                      <input
                        type="date"
                        value={formData.valid_from}
                        onChange={(e) => setFormData(prev => ({ ...prev, valid_from: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500 bg-white"
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Quote Valid Until (Expiry)
                      </label>
                      <input
                        type="date"
                        value={formData.valid_until}
                        onChange={(e) => setFormData(prev => ({ ...prev, valid_until: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500 bg-white"
                        data-testid="input-valid-until"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Estimated Departure (ETD)
                      </label>
                      <input
                        type="date"
                        value={formData.etd}
                        onChange={(e) => setFormData(prev => ({ ...prev, etd: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500 bg-white"
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-semibold text-slate-700 mb-1">
                        Estimated Arrival (ETA)
                      </label>
                      <input
                        type="date"
                        value={formData.eta}
                        onChange={(e) => setFormData(prev => ({ ...prev, eta: e.target.value }))}
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500 bg-white"
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Step 4: Notes & Final Review */}
              {modalStep === 4 && (
                <div className="space-y-3.5">
                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">
                      Commercial & Operational Notes
                    </label>
                    <textarea
                      rows={3}
                      placeholder="Special routing conditions, transshipment hubs, equipment availability, or vessel notes..."
                      value={formData.notes}
                      onChange={(e) => setFormData(prev => ({ ...prev, notes: e.target.value }))}
                      className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs focus:outline-none focus:border-indigo-500"
                    />
                  </div>

                  <div className="p-3 bg-slate-50 rounded-lg border border-slate-200 text-xs space-y-1.5">
                    <div className="font-bold text-slate-900">Quotation Summary Check:</div>
                    <div className="text-slate-600">
                      Carrier: <span className="font-bold text-slate-900">{formData.carrier_name || '—'}</span> ({formData.quote_reference || 'No Ref'})
                    </div>
                    <div className="text-slate-600">
                      Pricing: <span className="font-bold text-slate-900">{formatMoney(parseFloat(formData.buy_price) || 0, formData.currency)}</span> Buy · <span className="font-bold text-slate-900">{formatMoney(parseFloat(formData.sell_price) || 0, formData.currency)}</span> Sell
                    </div>
                    <div className="text-slate-600">
                      Transit: <span className="font-bold text-slate-900">{formData.transit_time_days || 'Standard'} Days</span> · Free Days: <span className="font-bold text-slate-900">{formData.free_days}</span>
                    </div>
                  </div>
                </div>
              )}

              {/* Modal Navigation Buttons */}
              <div className="pt-4 border-t border-slate-200 flex items-center justify-between">
                {modalStep > 1 ? (
                  <button
                    type="button"
                    onClick={() => setModalStep(s => s - 1)}
                    className="px-4 py-2 border border-slate-300 hover:bg-slate-50 text-slate-700 rounded-lg text-xs font-semibold transition-colors"
                  >
                    ← Back
                  </button>
                ) : (
                  <div />
                )}

                {modalStep < 4 ? (
                  <button
                    type="button"
                    onClick={() => {
                      if (modalStep === 1 && !formData.carrier_name.trim()) {
                        showToast('Carrier name is required', 'error');
                        return;
                      }
                      if (modalStep === 2) {
                        const buy = parseFloat(formData.buy_price);
                        const sell = parseFloat(formData.sell_price);
                        if (isNaN(buy) || buy <= 0 || isNaN(sell) || sell <= 0) {
                          showToast('Valid buy and sell prices are required', 'error');
                          return;
                        }
                      }
                      setModalStep(s => s + 1);
                    }}
                    className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-bold transition-colors cursor-pointer"
                    data-testid="modal-next-btn"
                  >
                    Next Step →
                  </button>
                ) : (
                  <button
                    type="submit"
                    disabled={actionLoading}
                    className="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-bold transition-colors shadow-xs disabled:opacity-50 cursor-pointer"
                    data-testid="submit-quote-btn"
                  >
                    {actionLoading ? 'Saving Quote...' : 'Persist Carrier Quote'}
                  </button>
                )}
              </div>

            </form>
          </div>
        </div>
      )}

    </div>
  );
}
