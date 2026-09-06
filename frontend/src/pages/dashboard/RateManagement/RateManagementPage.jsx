import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  Layers, ShieldCheck, Clock, AlertTriangle, Globe,
  Search, Plus, Filter, RotateCcw, ArrowUpDown, ChevronDown,
  Eye, Edit3, Trash2, Calendar, DollarSign, MapPin,
  CheckCircle, XCircle, ArrowRight, TrendingUp, Check,
  SlidersHorizontal, ChevronRight, ChevronLeft, X, Receipt, Tag, ListOrdered, GitBranch, History, FileText, Compass, BarChart3, RefreshCw, Send, Zap, Scale, ShieldAlert, BarChart2
} from 'lucide-react';
import { rateService } from '../../../services/rateService';
import RateContractsWorkspace from './RateContractsWorkspace';
import SpotRateWorkspace from './SpotRateWorkspace';
import RateLifecycleWorkspace from './RateLifecycleWorkspace';
import LiveCarrierRatesWorkspace from './LiveCarrierRatesWorkspace';
import ConfirmModal from '../Settings/ConfirmModal';
import ModuleHeroEmptyState from '../../../components/dashboard/ModuleHeroEmptyState';
import toast from 'react-hot-toast';
import './RateManagementPage.css';

// ── Custom Select Component for LogisticsHQ Design System ───────────────────
function CustomSelect({
  value,
  onChange,
  options = [],
  placeholder = 'Select...',
  disabled = false,
  className = '',
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [dropUp, setDropUp] = useState(false);
  const dropdownRef = useRef(null);

  const normalizedOptions = options.map((opt) => {
    if (typeof opt === 'string' || typeof opt === 'number') {
      return { value: String(opt), label: String(opt).replace(/_/g, ' ') };
    }
    return { ...opt, value: String(opt.value), label: String(opt.label || opt.value) };
  });

  const selectedOption = normalizedOptions.find((opt) => String(opt.value) === String(value));

  const toggleOpen = () => {
    if (disabled) return;
    if (!isOpen && dropdownRef.current) {
      const rect = dropdownRef.current.getBoundingClientRect();
      const spaceBelow = window.innerHeight - rect.bottom;
      const expectedHeight = Math.min(260, normalizedOptions.length * 36 + 10);
      if (spaceBelow < expectedHeight && rect.top > expectedHeight) {
        setDropUp(true);
      } else {
        setDropUp(false);
      }
    }
    setIsOpen((prev) => !prev);
  };

  useEffect(() => {
    if (!isOpen) return;
    const handleClickOutside = (e) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') setIsOpen(false);
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

  return (
    <div
      ref={dropdownRef}
      className={`rm-custom-select-container ${isOpen ? 'is-open' : ''} ${disabled ? 'is-disabled' : ''} ${className}`}
    >
      <button
        type="button"
        className={`rm-custom-select-trigger ${isOpen ? 'active' : ''}`}
        onClick={toggleOpen}
        disabled={disabled}
      >
        <div className="rm-custom-select-trigger-content">
          <span className={`rm-custom-select-trigger-label ${!selectedOption ? 'is-placeholder' : ''}`}>
            {selectedOption ? selectedOption.label : placeholder}
          </span>
        </div>
        <ChevronDown size={14} className={`rm-custom-select-chevron ${isOpen ? 'is-rotated' : ''}`} />
      </button>

      {isOpen && (
        <div className={`rm-custom-select-dropdown ${dropUp ? 'is-dropup' : ''}`}>
          <div className="rm-custom-select-options">
            {normalizedOptions.map((opt) => {
              const isSelected = String(opt.value) === String(value);
              return (
                <div
                  key={opt.value}
                  className={`rm-custom-select-option ${isSelected ? 'is-selected' : ''}`}
                  onClick={() => handleSelect(opt.value)}
                >
                  <span>{opt.label}</span>
                  {isSelected && <Check size={14} className="rm-custom-select-option-check" />}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

const TRANSPORT_MODES = ['ALL', 'Ocean FCL', 'Ocean LCL', 'Air Freight', 'Rail', 'Road'];
const SERVICE_TYPES = ['ALL', 'FCL', 'LCL', 'Port-to-Port', 'Door-to-Door', 'Air Freight'];
const EQUIPMENT_TYPES = ['ALL', '20GP', '40GP', '40HC', '45HC', 'REEFER'];
const RATE_TYPES = ['ALL', 'SPOT', 'CONTRACT', 'TARIFF', 'CUSTOM'];
const STATUS_OPTIONS = ['ALL', 'ACTIVE', 'EXPIRING_SOON', 'EXPIRED', 'DRAFT', 'ARCHIVED'];

const CHARGE_CATEGORIES = [
  'FREIGHT',
  'ORIGIN',
  'DESTINATION',
  'SURCHARGE',
  'DOCUMENTATION',
  'CUSTOMS',
  'INSURANCE',
  'TAX',
  'OTHER'
];

const CALCULATION_BASES = [
  { value: 'FLAT', label: 'Flat (Single Fixed Fee)' },
  { value: 'PER_CONTAINER', label: 'Per Container (Quantity × Rate)' },
  { value: 'PER_SHIPMENT', label: 'Per Shipment (Quantity × Rate)' },
  { value: 'PER_WEIGHT', label: 'Per Weight (Weight × Rate)' },
  { value: 'PER_VOLUME', label: 'Per Volume (Volume CBM × Rate)' },
  { value: 'PER_UNIT', label: 'Per Unit (Units × Rate)' },
  { value: 'PERCENTAGE', label: 'Percentage (% of Base Rate)' },
];

const CARRIERS_LIST = [
  'ALL',
  'Maersk Line',
  'MSC',
  'CMA CGM',
  'Hapag-Lloyd',
  'ONE Line',
  'Evergreen',
  'COSCO',
  'ZIM Line',
  'Yang Ming',
  'HMM'
];

export default function RateManagementPage() {
  // ── Search & Filter State ──
  const [activeTab, setActiveTab] = useState('rate_search'); // 'rate_search' | 'lane_search' | 'rate_comparison'
  const [origin, setOrigin] = useState('');
  const [destination, setDestination] = useState('');
  const [transportMode, setTransportMode] = useState('ALL');
  const [serviceType, setServiceType] = useState('ALL');
  const [equipmentType, setEquipmentType] = useState('ALL');
  const [carrierName, setCarrierName] = useState('ALL');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [rateTypeFilter, setRateTypeFilter] = useState('ALL');
  const [searchQuery, setSearchQuery] = useState('');
  const [validDate, setValidDate] = useState('');
  const [showMoreFilters, setShowMoreFilters] = useState(false);

  // ── Data State ──
  const [rates, setRates] = useState([]);
  const [summary, setSummary] = useState(null);
  const [totalCount, setTotalCount] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // ── Task 19.3: Workspace Switcher State ──
  const [activeWorkspace, setActiveWorkspace] = useState('rates'); // 'rates' | 'contracts'

  // ── Drawer & Tab State ──
  const [selectedRate, setSelectedRate] = useState(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerTab, setDrawerTab] = useState('overview'); // 'overview' | 'pricing' | 'versions'
  const [ratePricing, setRatePricing] = useState(null);
  const [pricingLoading, setPricingLoading] = useState(false);
  const [versionChain, setVersionChain] = useState([]);
  const [versionHistory, setVersionHistory] = useState([]);
  const [versionsLoading, setVersionsLoading] = useState(false);

  // ── Modals State ──
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [chargeModalOpen, setChargeModalOpen] = useState(false);
  const [versionModalOpen, setVersionModalOpen] = useState(false);
  const [editingCharge, setEditingCharge] = useState(null);
  const [formSubmitting, setFormSubmitting] = useState(false);

  // ── New Version Form State ──
  const [versionForm, setVersionForm] = useState({
    base_amount: '',
    currency: 'USD',
    effective_date: '',
    expiry_date: '',
    carrier_reference: '',
    contract_reference: '',
    revision_reason: '',
    notes: '',
  });

  // ── Rate Form State ──
  const [formData, setFormData] = useState({
    carrier_name: '',
    carrier_code: '',
    service_provider: '',
    rate_type: 'SPOT',
    transport_mode: 'Ocean FCL',
    service_type: 'FCL',
    equipment_type: '40GP',
    origin_port: '',
    origin_code: '',
    destination_port: '',
    destination_code: '',
    currency: 'USD',
    base_amount: '',
    effective_date: '',
    expiry_date: '',
    carrier_reference: '',
    contract_reference: '',
    notes: '',
  });

  // ── Charge Form State ──
  const [chargeForm, setChargeForm] = useState({
    charge_category: 'FREIGHT',
    charge_code: '',
    charge_name: '',
    calculation_basis: 'FLAT',
    quantity: 1,
    unit_price: '',
    currency: 'USD',
    minimum_amount: '',
    maximum_amount: '',
    included_in_base_rate: false,
    display_order: 0,
    notes: '',
  });

  // ── Fetch Summary KPIs ──
  const fetchSummary = useCallback(async () => {
    try {
      const res = await rateService.getRateSummary();
      const data = res?.data || res;
      setSummary(data);
    } catch (err) {
      console.error('Failed to load rate summary:', err);
    }
  }, []);

  // ── Fetch Rate List ──
  const fetchRates = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const params = {
        page,
        limit,
        search: searchQuery.trim(),
        origin: origin.trim(),
        destination: destination.trim(),
        transport_mode: transportMode,
        service_type: serviceType,
        equipment_type: equipmentType,
        carrier_name: carrierName,
        status: statusFilter,
        rate_type: rateTypeFilter,
        valid_date: validDate,
      };
      const res = await rateService.listRates(params);
      const data = res?.data || res;
      setRates(data?.rates || []);
      setTotalCount(data?.total_count || 0);
      setTotalPages(data?.total_pages || 1);
    } catch (err) {
      console.error('Failed to load rates:', err);
      setError(err?.message || 'Unable to retrieve rate records.');
      setRates([]);
    } finally {
      setLoading(false);
    }
  }, [page, limit, searchQuery, origin, destination, transportMode, serviceType, equipmentType, carrierName, statusFilter, rateTypeFilter, validDate]);

  // ── Fetch Pricing for Drawer ──
  const fetchPricing = useCallback(async (rateId) => {
    if (!rateId) return;
    setPricingLoading(true);
    try {
      const res = await rateService.getRatePricing(rateId);
      const data = res?.data || res;
      setRatePricing(data?.pricing || null);
    } catch (err) {
      console.error('Failed to load rate pricing:', err);
    } finally {
      setPricingLoading(false);
    }
  }, []);

  // ── Fetch Rate Versions & Audit History (Task 19.3) ──
  const fetchVersions = useCallback(async (rateId) => {
    if (!rateId) return;
    setVersionsLoading(true);
    try {
      const [versionsRes, historyRes] = await Promise.all([
        rateService.getRateVersions(rateId),
        rateService.getRateVersionHistory(rateId),
      ]);
      const vList = versionsRes?.data?.versions || versionsRes?.versions || (Array.isArray(versionsRes) ? versionsRes : []);
      setVersionChain(Array.isArray(vList) ? vList : []);

      const hList = historyRes?.data?.history || historyRes?.history || (Array.isArray(historyRes) ? historyRes : []);
      setVersionHistory(Array.isArray(hList) ? hList : []);
    } catch (err) {
      console.error('Failed to load rate versions:', err);
    } finally {
      setVersionsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSummary();
  }, [fetchSummary]);

  useEffect(() => {
    fetchRates();
  }, [fetchRates]);

  const handleSearchSubmit = (e) => {
    if (e) e.preventDefault();
    setPage(1);
    fetchRates();
  };

  const handleResetFilters = () => {
    setOrigin('');
    setDestination('');
    setTransportMode('ALL');
    setServiceType('ALL');
    setEquipmentType('ALL');
    setCarrierName('ALL');
    setStatusFilter('ALL');
    setRateTypeFilter('ALL');
    setSearchQuery('');
    setValidDate('');
    setPage(1);
  };

  const handleOpenDetail = async (rateItem) => {
    setDrawerTab('overview');
    try {
      const res = await rateService.getRate(rateItem.id);
      const data = res?.data || res || rateItem;
      setSelectedRate(data);
      setDrawerOpen(true);
      fetchPricing(data.id);
      fetchVersions(data.id);
    } catch {
      setSelectedRate(rateItem);
      setDrawerOpen(true);
      fetchPricing(rateItem.id);
      fetchVersions(rateItem.id);
    }
  };

  const handleOpenCreateVersion = (rateItem, e) => {
    if (e) e.stopPropagation();
    setSelectedRate(rateItem);
    setVersionForm({
      base_amount: rateItem.base_amount || '',
      currency: rateItem.currency || 'USD',
      effective_date: rateItem.effective_date || '',
      expiry_date: rateItem.expiry_date || '',
      carrier_reference: rateItem.carrier_reference || '',
      contract_reference: rateItem.contract_reference || '',
      revision_reason: 'Annual contract price adjustment',
      notes: rateItem.notes || '',
    });
    setVersionModalOpen(true);
  };

  const handleCreateVersionSubmit = async (e) => {
    e.preventDefault();
    if (!selectedRate) return;
    setFormSubmitting(true);
    try {
      const payload = {
        base_amount: parseFloat(versionForm.base_amount),
        currency: versionForm.currency,
        effective_date: versionForm.effective_date,
        expiry_date: versionForm.expiry_date,
        carrier_reference: versionForm.carrier_reference || undefined,
        contract_reference: versionForm.contract_reference || undefined,
        revision_reason: versionForm.revision_reason || 'Commercial rate revision',
        notes: versionForm.notes || undefined,
      };
      await rateService.createRateVersion(selectedRate.id, payload);
      toast.success(`Version created successfully for ${selectedRate.rate_reference}`);
      setVersionModalOpen(false);
      fetchRates();
      fetchSummary();
      if (drawerOpen && selectedRate) {
        fetchVersions(selectedRate.id);
        fetchPricing(selectedRate.id);
      }
    } catch (err) {
      toast.error(err?.message || err?.details?.error?.message || 'Failed to create rate version');
    } finally {
      setFormSubmitting(false);
    }
  };

  const handleOpenEdit = (rateItem, e) => {
    if (e) e.stopPropagation();
    setSelectedRate(rateItem);
    setFormData({
      carrier_name: rateItem.carrier_name || '',
      carrier_code: rateItem.carrier_code || '',
      service_provider: rateItem.service_provider || '',
      rate_type: rateItem.rate_type || 'SPOT',
      transport_mode: rateItem.transport_mode || 'Ocean FCL',
      service_type: rateItem.service_type || 'FCL',
      equipment_type: rateItem.equipment_type || '40GP',
      origin_port: rateItem.origin_port || '',
      origin_code: rateItem.origin_code || '',
      destination_port: rateItem.destination_port || '',
      destination_code: rateItem.destination_code || '',
      currency: rateItem.currency || 'USD',
      base_amount: rateItem.base_amount || '',
      effective_date: rateItem.effective_date || '',
      expiry_date: rateItem.expiry_date || '',
      carrier_reference: rateItem.carrier_reference || '',
      contract_reference: rateItem.contract_reference || '',
      notes: rateItem.notes || '',
    });
    setEditModalOpen(true);
  };

  const handleCreateSubmit = async (e) => {
    e.preventDefault();
    setFormSubmitting(true);
    try {
      await rateService.createRate({
        ...formData,
        base_amount: parseFloat(formData.base_amount) || 0,
      });
      setCreateModalOpen(false);
      fetchRates();
      fetchSummary();
      resetForm();
    } catch (err) {
      alert(err?.response?.data?.error?.message || err?.message || 'Failed to create rate.');
    } finally {
      setFormSubmitting(false);
    }
  };

  const handleEditSubmit = async (e) => {
    e.preventDefault();
    if (!selectedRate?.id) return;
    setFormSubmitting(true);
    try {
      await rateService.updateRate(selectedRate.id, {
        ...formData,
        base_amount: parseFloat(formData.base_amount) || 0,
      });
      setEditModalOpen(false);
      fetchRates();
      fetchSummary();
      if (selectedRate?.id) {
        fetchPricing(selectedRate.id);
      }
    } catch (err) {
      alert(err?.response?.data?.error?.message || err?.message || 'Failed to update rate.');
    } finally {
      setFormSubmitting(false);
    }
  };

  // Archive Rate Confirmation Modal State
  const [archiveRateModalOpen, setArchiveRateModalOpen] = useState(false);
  const [rateToArchiveId, setRateToArchiveId] = useState(null);
  const [archiveRateLoading, setArchiveRateLoading] = useState(false);

  const handleArchiveRate = (rateId, e) => {
    if (e) e.stopPropagation();
    setRateToArchiveId(rateId);
    setArchiveRateModalOpen(true);
  };

  const handleConfirmArchiveRate = async () => {
    if (!rateToArchiveId) return;
    setArchiveRateLoading(true);
    try {
      await rateService.archiveRate(rateToArchiveId);
      toast.success('Rate record archived');
      setArchiveRateModalOpen(false);
      if (drawerOpen && selectedRate?.id === rateToArchiveId) {
        setDrawerOpen(false);
      }
      setRateToArchiveId(null);
      fetchRates();
      fetchSummary();
    } catch (err) {
      toast.error(err?.message || 'Failed to archive rate.');
    } finally {
      setArchiveRateLoading(false);
    }
  };

  // ── Rate Charges Actions (Task 19.2) ──

  const handleOpenAddCharge = () => {
    if (selectedRate?.status === 'ARCHIVED') {
      alert('Cannot add charges to an archived rate record.');
      return;
    }
    setEditingCharge(null);
    setChargeForm({
      charge_category: 'FREIGHT',
      charge_code: '',
      charge_name: '',
      calculation_basis: 'FLAT',
      quantity: 1,
      unit_price: '',
      currency: selectedRate?.currency || 'USD',
      minimum_amount: '',
      maximum_amount: '',
      included_in_base_rate: false,
      display_order: (ratePricing?.charges?.length || 0) + 1,
      notes: '',
    });
    setChargeModalOpen(true);
  };

  const handleOpenEditCharge = (charge) => {
    if (selectedRate?.status === 'ARCHIVED') {
      alert('Cannot modify charges on an archived rate record.');
      return;
    }
    setEditingCharge(charge);
    setChargeForm({
      charge_category: charge.charge_category || 'FREIGHT',
      charge_code: charge.charge_code || '',
      charge_name: charge.charge_name || '',
      calculation_basis: charge.calculation_basis || 'FLAT',
      quantity: charge.quantity || 1,
      unit_price: charge.unit_price || '',
      currency: charge.currency || selectedRate?.currency || 'USD',
      minimum_amount: charge.minimum_amount != null ? charge.minimum_amount : '',
      maximum_amount: charge.maximum_amount != null ? charge.maximum_amount : '',
      included_in_base_rate: Boolean(charge.included_in_base_rate),
      display_order: charge.display_order || 0,
      notes: charge.notes || '',
    });
    setChargeModalOpen(true);
  };

  const handleChargeSubmit = async (e) => {
    e.preventDefault();
    if (!selectedRate?.id) return;
    setFormSubmitting(true);
    try {
      const payload = {
        charge_category: chargeForm.charge_category,
        charge_code: chargeForm.charge_code.trim(),
        charge_name: chargeForm.charge_name.trim(),
        calculation_basis: chargeForm.calculation_basis,
        quantity: parseFloat(chargeForm.quantity) || 1,
        unit_price: parseFloat(chargeForm.unit_price) || 0,
        currency: chargeForm.currency.trim().toUpperCase() || selectedRate.currency,
        minimum_amount: chargeForm.minimum_amount !== '' ? parseFloat(chargeForm.minimum_amount) : null,
        maximum_amount: chargeForm.maximum_amount !== '' ? parseFloat(chargeForm.maximum_amount) : null,
        included_in_base_rate: chargeForm.included_in_base_rate,
        display_order: parseInt(chargeForm.display_order, 10) || 0,
        notes: chargeForm.notes.trim() || null,
      };

      if (editingCharge) {
        await rateService.updateRateCharge(selectedRate.id, editingCharge.id, payload);
      } else {
        await rateService.addRateCharge(selectedRate.id, payload);
      }

      setChargeModalOpen(false);
      fetchPricing(selectedRate.id);
    } catch (err) {
      alert(err?.response?.data?.error?.message || err?.message || 'Failed to save charge item.');
    } finally {
      setFormSubmitting(false);
    }
  };

  const handleDeleteCharge = async (chargeId) => {
    if (selectedRate?.status === 'ARCHIVED') {
      alert('Cannot delete charges from an archived rate record.');
      return;
    }
    if (!window.confirm('Are you sure you want to remove this charge item?')) return;
    try {
      await rateService.deleteRateCharge(selectedRate.id, chargeId);
      fetchPricing(selectedRate.id);
    } catch (err) {
      alert(err?.response?.data?.error?.message || err?.message || 'Failed to delete charge.');
    }
  };

  // Live frontend preview calculator for modal
  const previewChargeAmount = () => {
    if (chargeForm.included_in_base_rate) return 0;
    const qty = parseFloat(chargeForm.quantity) || 0;
    const price = parseFloat(chargeForm.unit_price) || 0;
    const base = parseFloat(selectedRate?.base_amount) || 0;

    let amt = 0;
    if (chargeForm.calculation_basis === 'FLAT') {
      amt = price;
    } else if (chargeForm.calculation_basis === 'PERCENTAGE') {
      amt = (price / 100) * base;
    } else {
      amt = qty * price;
    }

    if (chargeForm.minimum_amount !== '' && parseFloat(chargeForm.minimum_amount) > 0) {
      const min = parseFloat(chargeForm.minimum_amount);
      if (amt < min) amt = min;
    }
    if (chargeForm.maximum_amount !== '' && parseFloat(chargeForm.maximum_amount) > 0) {
      const max = parseFloat(chargeForm.maximum_amount);
      if (amt > max) amt = max;
    }
    return Math.max(0, amt);
  };

  const resetForm = () => {
    setFormData({
      carrier_name: '',
      carrier_code: '',
      service_provider: '',
      rate_type: 'SPOT',
      transport_mode: 'Ocean FCL',
      service_type: 'FCL',
      equipment_type: '40GP',
      origin_port: '',
      origin_code: '',
      destination_port: '',
      destination_code: '',
      currency: 'USD',
      base_amount: '',
      effective_date: '',
      expiry_date: '',
      carrier_reference: '',
      contract_reference: '',
      notes: '',
    });
  };

  // Helper for status badge
  const renderStatusBadge = (status) => {
    switch (status) {
      case 'ACTIVE':
        return <span className="rm-badge rm-badge--active"><span className="rm-badge-dot" /> Active</span>;
      case 'EXPIRING_SOON':
        return <span className="rm-badge rm-badge--expiring"><span className="rm-badge-dot" /> Expiring Soon</span>;
      case 'EXPIRED':
        return <span className="rm-badge rm-badge--expired"><span className="rm-badge-dot" /> Expired</span>;
      case 'DRAFT':
        return <span className="rm-badge rm-badge--draft"><span className="rm-badge-dot" /> Draft</span>;
      default:
        return <span className="rm-badge rm-badge--archived"><span className="rm-badge-dot" /> {status || 'Archived'}</span>;
    }
  };

  return (
    <div className="rm-page">
      {/* ── Top Header Hero ── */}
      <div className="rm-header-hero">
        <div className="rm-header-left">
          <div className="rm-header-title-row">
            <h1 className="rm-header-title">Rate Management</h1>
            <span className="rm-header-live-tag">
              <span className="rm-pulse-dot" /> LIVE ENGINE
            </span>
          </div>
          <p className="rm-header-sub">
            Centralized freight procurement, carrier tariffs, and real-time commercial rate calculations.
          </p>

          {/* Task 19.3: Workspace Switcher Tabs */}
          <div style={{ display: 'inline-flex', background: '#F1F5F9', padding: '4px', borderRadius: '10px', marginTop: '6px', gap: '4px' }}>
            <button
              type="button"
              onClick={() => setActiveWorkspace('rates')}
              style={{
                background: activeWorkspace === 'rates' ? '#FFFFFF' : 'transparent',
                color: activeWorkspace === 'rates' ? '#2563EB' : '#64748B',
                boxShadow: activeWorkspace === 'rates' ? '0 1px 3px rgba(0,0,0,0.08)' : 'none',
                fontWeight: activeWorkspace === 'rates' ? 750 : 600,
                border: 'none',
                borderRadius: '7px',
                padding: '6px 14px',
                fontSize: '12.5px',
                cursor: 'pointer',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                transition: 'all 0.15s ease',
              }}
            >
              <Layers size={14} /> Rate Repository & Search
            </button>
            <button
              type="button"
              onClick={() => setActiveWorkspace('contracts')}
              style={{
                background: activeWorkspace === 'contracts' ? '#FFFFFF' : 'transparent',
                color: activeWorkspace === 'contracts' ? '#2563EB' : '#64748B',
                boxShadow: activeWorkspace === 'contracts' ? '0 1px 3px rgba(0,0,0,0.08)' : 'none',
                fontWeight: activeWorkspace === 'contracts' ? 750 : 600,
                border: 'none',
                borderRadius: '7px',
                padding: '6px 14px',
                fontSize: '12.5px',
                cursor: 'pointer',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                transition: 'all 0.15s ease',
              }}
            >
              <FileText size={14} /> Carrier Contracts & Versions
            </button>

            <button
              type="button"
              onClick={() => setActiveWorkspace('spot_requests')}
              style={{
                background: activeWorkspace === 'spot_requests' ? '#FFFFFF' : 'transparent',
                color: activeWorkspace === 'spot_requests' ? '#2563EB' : '#64748B',
                boxShadow: activeWorkspace === 'spot_requests' ? '0 1px 3px rgba(0,0,0,0.08)' : 'none',
                fontWeight: activeWorkspace === 'spot_requests' ? 750 : 600,
                border: 'none',
                borderRadius: '7px',
                padding: '6px 14px',
                fontSize: '12.5px',
                cursor: 'pointer',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                transition: 'all 0.15s ease',
              }}
            >
              <Send size={14} /> Spot Rate Requests & Sourcing
            </button>

            <button
              type="button"
              onClick={() => setActiveWorkspace('lifecycle')}
              style={{
                background: activeWorkspace === 'lifecycle' ? '#FFFFFF' : 'transparent',
                color: activeWorkspace === 'lifecycle' ? '#2563EB' : '#64748B',
                boxShadow: activeWorkspace === 'lifecycle' ? '0 1px 3px rgba(0,0,0,0.08)' : 'none',
                fontWeight: activeWorkspace === 'lifecycle' ? 750 : 600,
                border: 'none',
                borderRadius: '7px',
                padding: '6px 14px',
                fontSize: '12.5px',
                cursor: 'pointer',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                transition: 'all 0.15s ease',
              }}
            >
              <ShieldAlert size={14} /> Rate Lifecycle & Risk Intelligence
            </button>

            <button
              type="button"
              onClick={() => setActiveWorkspace('live_carrier')}
              style={{
                background: activeWorkspace === 'live_carrier' ? '#FFFFFF' : 'transparent',
                color: activeWorkspace === 'live_carrier' ? '#2563EB' : '#64748B',
                boxShadow: activeWorkspace === 'live_carrier' ? '0 1px 3px rgba(0,0,0,0.08)' : 'none',
                fontWeight: activeWorkspace === 'live_carrier' ? 750 : 600,
                border: 'none',
                borderRadius: '7px',
                padding: '6px 14px',
                fontSize: '12.5px',
                cursor: 'pointer',
                display: 'inline-flex',
                alignItems: 'center',
                gap: '6px',
                transition: 'all 0.15s ease',
              }}
            >
              <Zap size={14} /> Live Carrier Rates & APIs
            </button>
          </div>
        </div>

        <div className="rm-header-right">
          <div className="rm-header-search-box">
            <Search size={14} className="rm-header-search-icon" />
            <input
              type="text"
              placeholder="Search rates, lanes, carriers..."
              className="rm-header-search-input"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearchSubmit()}
            />
            <span className="rm-header-search-kbd">⌘K</span>
          </div>

          <div className="rm-date-pill">
            <Calendar size={14} style={{ color: '#2563EB' }} />
            <span>Aug 9 – Aug 15, 2026</span>
          </div>

          <button
            type="button"
            className="rm-btn-create"
            onClick={() => { resetForm(); setCreateModalOpen(true); }}
          >
            <Plus size={16} />
            <span>Create New Rate</span>
          </button>
        </div>
      </div>

      {/* ── Conditional Workspace Rendering (Tasks 19.3, 19.4, 19.6 & Carrier Task 5) ── */}
      {activeWorkspace === 'contracts' ? (
        <RateContractsWorkspace
          CustomSelect={CustomSelect}
          onOpenRateDetail={handleOpenDetail}
        />
      ) : activeWorkspace === 'spot_requests' ? (
        <SpotRateWorkspace />
      ) : activeWorkspace === 'lifecycle' ? (
        <RateLifecycleWorkspace
          onViewRate={(id) => handleOpenDetail(id)}
          onViewContract={() => setActiveWorkspace('contracts')}
        />
      ) : activeWorkspace === 'live_carrier' ? (
        <LiveCarrierRatesWorkspace />
      ) : (
        <>
          {/* ── 5 Metric KPI Cards Strip ── */}
          <div className="rm-kpi-grid">
            
            {/* 1. Total Rates */}
            <div className="rm-kpi-card kpi-blue">
              <div className="rm-kpi-watermark">
                <Layers size={64} />
              </div>
              <div className="rm-kpi-header">
                <div className="rm-kpi-icon-wrap icon-blue">
                  <Layers size={16} />
                </div>
                <span className="rm-kpi-tag tag-blue">All Rates</span>
              </div>
              <div className="rm-kpi-body">
                <span className="rm-kpi-label">Total Rates</span>
                <div className="rm-kpi-number">{summary?.total_rates ?? 0}</div>
                <span className="rm-kpi-sub">Commercial tariff repository</span>
              </div>
              <div className="rm-kpi-footer">
                <span className="rm-kpi-dot dot-blue"></span>
                <span className="rm-kpi-footer-text">
                  {summary?.total_rates ? `${summary.total_rates} rates indexed` : '0 rates indexed'}
                </span>
              </div>
            </div>

            {/* 2. Active Rates */}
            <div className="rm-kpi-card kpi-emerald">
              <div className="rm-kpi-watermark">
                <ShieldCheck size={64} />
              </div>
              <div className="rm-kpi-header">
                <div className="rm-kpi-icon-wrap icon-emerald">
                  <ShieldCheck size={16} />
                </div>
                <span className="rm-kpi-tag tag-emerald">In Effect</span>
              </div>
              <div className="rm-kpi-body">
                <span className="rm-kpi-label">Active Rates</span>
                <div className="rm-kpi-number text-emerald">{summary?.active_rates ?? 0}</div>
                <span className="rm-kpi-sub">Commercially valid terms</span>
              </div>
              <div className="rm-kpi-footer">
                <span className="rm-kpi-dot dot-emerald"></span>
                <span className="rm-kpi-footer-text">
                  {summary?.active_pct !== undefined && summary?.active_pct !== null
                    ? `${summary.active_pct.toFixed(0)}% commercially valid`
                    : '100% commercially valid'}
                </span>
              </div>
            </div>

            {/* 3. Expiring Soon */}
            <div className="rm-kpi-card kpi-amber">
              <div className="rm-kpi-watermark">
                <Clock size={64} />
              </div>
              <div className="rm-kpi-header">
                <div className="rm-kpi-icon-wrap icon-amber">
                  <Clock size={16} />
                </div>
                <span className="rm-kpi-tag tag-amber">Action Required</span>
              </div>
              <div className="rm-kpi-body">
                <span className="rm-kpi-label">Expiring Soon</span>
                <div className="rm-kpi-number text-amber">{summary?.expiring_soon_rates ?? 0}</div>
                <span className="rm-kpi-sub">Expiring within 30 days</span>
              </div>
              <div className="rm-kpi-footer">
                <span className="rm-kpi-dot dot-amber"></span>
                <span className="rm-kpi-footer-text">
                  {summary?.expiring_soon_rates ? `${summary.expiring_soon_rates} within 30 days` : '0 within 30 days'}
                </span>
              </div>
            </div>

            {/* 4. Expired Rates */}
            <div className="rm-kpi-card kpi-rose">
              <div className="rm-kpi-watermark">
                <AlertTriangle size={64} />
              </div>
              <div className="rm-kpi-header">
                <div className="rm-kpi-icon-wrap icon-rose">
                  <AlertTriangle size={16} />
                </div>
                <span className="rm-kpi-tag tag-rose">Renewal Needed</span>
              </div>
              <div className="rm-kpi-body">
                <span className="rm-kpi-label">Expired Rates</span>
                <div className="rm-kpi-number text-rose">{summary?.expired_rates ?? 0}</div>
                <span className="rm-kpi-sub">Terms concluded / stale</span>
              </div>
              <div className="rm-kpi-footer">
                <span className="rm-kpi-dot dot-rose"></span>
                <span className="rm-kpi-footer-text">
                  {summary?.expired_rates && summary.expired_rates > 0
                    ? `${summary.expired_rates} require renewal`
                    : 'No expired rates'}
                </span>
              </div>
            </div>

            {/* 5. Lanes Covered */}
            <div className="rm-kpi-card kpi-purple">
              <div className="rm-kpi-watermark">
                <Globe size={64} />
              </div>
              <div className="rm-kpi-header">
                <div className="rm-kpi-icon-wrap icon-purple">
                  <Globe size={16} />
                </div>
                <span className="rm-kpi-tag tag-purple">Network</span>
              </div>
              <div className="rm-kpi-body">
                <span className="rm-kpi-label">Lanes Covered</span>
                <div className="rm-kpi-number text-purple">{summary?.lanes_covered ?? 0}</div>
                <span className="rm-kpi-sub">Active trade corridors</span>
              </div>
              <div className="rm-kpi-footer">
                <span className="rm-kpi-dot dot-purple"></span>
                <span className="rm-kpi-footer-text">
                  {summary?.lanes_covered
                    ? `${summary.lanes_covered} active trade corridor${summary.lanes_covered === 1 ? '' : 's'}`
                    : '0 active trade corridors'}
                </span>
              </div>
            </div>

          </div>

          {/* ── Main Content Area: Search / Filters + Rates Table + Right Sidebar ── */}
          <div className="rm-content-grid">
            <div className="rm-main-column">
              {/* Filter / Search Card */}
              <div className="rm-search-card">
                <div className="rm-search-tabs">
                  <button
                    type="button"
                    className={`rm-search-tab ${activeTab === 'rate_search' ? 'active' : ''}`}
                    onClick={() => setActiveTab('rate_search')}
                  >
                    <Search size={14} /> Rate Search
                  </button>
                  <button
                    type="button"
                    className={`rm-search-tab ${activeTab === 'lane_search' ? 'active' : ''}`}
                    onClick={() => setActiveTab('lane_search')}
                  >
                    <Compass size={14} /> Lane Search
                  </button>
                  <button
                    type="button"
                    className={`rm-search-tab ${activeTab === 'rate_comparison' ? 'active' : ''}`}
                    onClick={() => setActiveTab('rate_comparison')}
                  >
                    <TrendingUp size={14} /> Rate Comparison
                  </button>
                </div>

                <form className="rm-search-body" onSubmit={handleSearchSubmit}>
                  <div className="rm-filters-grid">
                    <div className="rm-field-group">
                      <label>Origin</label>
                      <input
                        type="text"
                        placeholder="e.g. USNYC, New York"
                        className="rm-input"
                        value={origin}
                        onChange={(e) => setOrigin(e.target.value)}
                      />
                    </div>

                    <div className="rm-field-group">
                      <label>Destination</label>
                      <input
                        type="text"
                        placeholder="e.g. INNSA, Nhava Sheva"
                        className="rm-input"
                        value={destination}
                        onChange={(e) => setDestination(e.target.value)}
                      />
                    </div>

                    <div className="rm-field-group">
                      <label>Transport Mode</label>
                      <CustomSelect
                        value={transportMode}
                        onChange={setTransportMode}
                        options={TRANSPORT_MODES}
                        placeholder="Select Mode"
                      />
                    </div>

                    <div className="rm-field-group">
                      <label>Service Type</label>
                      <CustomSelect
                        value={serviceType}
                        onChange={setServiceType}
                        options={SERVICE_TYPES}
                        placeholder="Select Service"
                      />
                    </div>

                    <div className="rm-field-group">
                      <label>Equipment Type</label>
                      <CustomSelect
                        value={equipmentType}
                        onChange={setEquipmentType}
                        options={EQUIPMENT_TYPES}
                        placeholder="Select Equipment"
                      />
                    </div>

                    <div className="rm-field-group">
                      <label>Carrier / Supplier</label>
                      <CustomSelect
                        value={carrierName}
                        onChange={setCarrierName}
                        options={CARRIERS_LIST}
                        placeholder="Select Carrier"
                      />
                    </div>

                    <div className="rm-field-group">
                      <label>Valid Date</label>
                      <input
                        type="date"
                        className="rm-input"
                        value={validDate}
                        onChange={(e) => setValidDate(e.target.value)}
                      />
                    </div>

                    <div className="rm-field-group">
                      <label>Status</label>
                      <CustomSelect
                        value={statusFilter}
                        onChange={setStatusFilter}
                        options={STATUS_OPTIONS}
                        placeholder="Select Status"
                      />
                    </div>
                  </div>

                  {showMoreFilters && (
                    <div className="rm-filters-grid" style={{ paddingTop: '10px', borderTop: '1px dashed #E2E8F0' }}>
                      <div className="rm-field-group">
                        <label>Rate Type</label>
                        <CustomSelect
                          value={rateTypeFilter}
                          onChange={setRateTypeFilter}
                          options={RATE_TYPES}
                          placeholder="Select Rate Type"
                        />
                      </div>
                    </div>
                  )}

                  <div className="rm-search-actions-bar">
                    <button
                      type="button"
                      className="rm-more-filters-btn"
                      onClick={() => setShowMoreFilters(!showMoreFilters)}
                    >
                      <SlidersHorizontal size={13} />
                      <span>{showMoreFilters ? 'Fewer Filters' : 'More Filters'}</span>
                    </button>

                    <div className="rm-search-btn-group">
                      <button
                        type="button"
                        className="rm-btn-reset"
                        onClick={handleResetFilters}
                      >
                        Reset
                      </button>
                      <button
                        type="submit"
                        className="rm-btn-submit-search"
                      >
                        <Search size={14} /> Search Rates
                      </button>
                    </div>
                  </div>
                </form>
              </div>

              {/* Rates Table Card */}
              <div className="rm-table-card">
                <div className="rm-table-header">
                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    <h2 className="rm-table-title">Recent Rates</h2>
                    <span className="rm-table-count-badge">Showing {rates.length} of {totalCount} rate records</span>
                  </div>
                </div>

                <div className="rm-table-wrap">
                  <table className="rm-table">
                    <thead>
                      <tr>
                        <th>Reference</th>
                        <th>Carrier</th>
                        <th>Origin Port</th>
                        <th>Destination</th>
                        <th>Mode / Eq</th>
                        <th>Base Rate</th>
                        <th>Version</th>
                        <th>Validity</th>
                        <th>Status</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {loading ? (
                        <tr>
                          <td colSpan="10" style={{ textAlign: 'center', padding: '40px', color: '#64748B' }}>
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px' }}>
                              <RefreshCw size={16} className="rm-spin" /> Loading rates repository...
                            </div>
                          </td>
                        </tr>
                      ) : rates.length === 0 ? (
                        <tr>
                          <td colSpan="10" style={{ padding: '0', border: 'none' }}>
                            <ModuleHeroEmptyState
                              icon={<Layers size={28} />}
                              badgeTheme="indigo"
                              title="No Commercial Freight Rates Configured"
                              description="Build your rate engine by filing ocean, air, and overland carrier rate cards. Rates automatically power instant quotation pricing, surcharge breakdowns, and margin controls."
                              primaryAction={{
                                label: 'Create First Rate Record',
                                icon: <Plus size={15} />,
                                onClick: () => { resetForm(); setCreateModalOpen(true); },
                              }}
                              secondaryAction={{
                                label: 'Open Rate Contracts Workspace',
                                icon: <ArrowRight size={15} />,
                                onClick: () => setActiveTab('CONTRACTS'),
                              }}
                              features={[
                                {
                                  icon: <Globe size={18} />,
                                  iconBg: '#eff6ff',
                                  iconColor: '#2563eb',
                                  title: 'Multi-Modal Carrier Rate Cards',
                                  desc: 'File standard container buy rates, bunker adjustment factors (BAF), and local THC charges.',
                                },
                                {
                                  icon: <Zap size={18} />,
                                  iconBg: '#ecfdf5',
                                  iconColor: '#059669',
                                  title: 'Instant Quotation Auto-Pricing',
                                  desc: 'Automatically calculate sell rates from active rate contracts during RFQ response.',
                                },
                                {
                                  icon: <ShieldCheck size={18} />,
                                  iconBg: '#f5f3ff',
                                  iconColor: '#7c3aed',
                                  title: 'Validity & Expiry Guard',
                                  desc: 'Track rate validity periods, GRI updates, and seasonal surcharge adjustments.',
                                },
                              ]}
                            />
                          </td>
                        </tr>
                      ) : (
                        rates.map((r) => (
                          <tr key={r.id} onClick={() => handleOpenDetail(r)} style={{ cursor: 'pointer' }}>
                            <td>
                              <div className="rm-ref-cell">
                                <span className="rm-ref-code">{r.rate_reference}</span>
                                {r.rate_type && (
                                  <span className="rm-type-badge">{r.rate_type}</span>
                                )}
                              </div>
                            </td>
                            <td>
                              <div className="rm-carrier-name-cell">
                                <strong>{r.carrier_name}</strong>
                                {r.carrier_code && <span className="rm-scac-tag">{r.carrier_code}</span>}
                              </div>
                            </td>
                            <td>
                              <span className="rm-port-code">{r.origin_port}</span>
                            </td>
                            <td>
                              <span className="rm-port-code">{r.destination_port}</span>
                            </td>
                            <td>
                              <div className="rm-mode-stack">
                                <span>{r.transport_mode}</span>
                                <span className="rm-eq-pill">{r.equipment_type || 'Std'}</span>
                              </div>
                            </td>
                            <td>
                              <div className="rm-price-cell">
                                <span className="rm-price-cur">{r.currency}</span>
                                <span className="rm-price-val">{Number(r.base_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}</span>
                              </div>
                            </td>
                            <td>
                              <span className="rc-badge-pill rc-badge-draft" style={{ fontSize: '11px' }}>
                                v{r.version_number || 1}
                              </span>
                            </td>
                            <td>
                              <div className="rm-validity-cell">
                                <span>{r.effective_date} – {r.expiry_date}</span>
                                {r.days_until_expiry !== undefined && (
                                  <span className="rm-days-tag">
                                    {r.days_until_expiry <= 0 ? 'Expired' : `${r.days_until_expiry}d left`}
                                  </span>
                                )}
                              </div>
                            </td>
                            <td>{renderStatusBadge(r.status)}</td>
                            <td onClick={(e) => e.stopPropagation()}>
                              <div className="rm-table-actions-group">
                                <button
                                  type="button"
                                  className="rm-btn-view-details"
                                  onClick={() => handleOpenDetail(r)}
                                >
                                  <Eye size={13} />
                                  <span>View Details</span>
                                </button>
                                <div className="rm-table-actions-icons">
                                  <button
                                    type="button"
                                    className="rm-action-icon-btn"
                                    title="Create New Version"
                                    onClick={(e) => handleOpenCreateVersion(r, e)}
                                  >
                                    <GitBranch size={13} />
                                  </button>
                                  <button
                                    type="button"
                                    className="rm-action-icon-btn"
                                    title="Edit Rate"
                                    onClick={(e) => handleOpenEdit(r, e)}
                                  >
                                    <Edit3 size={13} />
                                  </button>
                                  <button
                                    type="button"
                                    className="rm-action-icon-btn rm-action-icon-btn--delete"
                                    title="Archive Rate"
                                    onClick={(e) => handleArchiveRate(r.id, e)}
                                  >
                                    <Trash2 size={13} />
                                  </button>
                                </div>
                              </div>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>

                {/* Pagination */}
                {totalCount > limit && (
                  <div className="rm-pagination-bar">
                    <span className="rm-page-info">
                      Page {page} of {totalPages} ({totalCount} total)
                    </span>
                    <div className="rm-page-btn-group">
                      <button
                        type="button"
                        className="rm-page-btn"
                        disabled={page <= 1}
                        onClick={() => setPage(page - 1)}
                      >
                        <ChevronLeft size={14} /> Previous
                      </button>
                      <button
                        type="button"
                        className="rm-page-btn"
                        disabled={page >= totalPages}
                        onClick={() => setPage(page + 1)}
                      >
                        Next <ChevronRight size={14} />
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* ── Right Sidebar Column ── */}
            <div className="rm-sidebar-column">
              {/* Widget 1: Rate Summary Breakdown */}
              <div className="rm-widget-card">
                <h3 className="rm-widget-title">
                  <span>Rate Portfolio Status</span>
                  <span className="rm-widget-title-right">{summary?.total_rates ?? 0} Total</span>
                </h3>

                <div className="rm-summary-bars">
                  <div className="rm-bar-item">
                    <div className="rm-bar-labels">
                      <span>Active Rates</span>
                      <span style={{ color: '#16A34A', fontWeight: 700 }}>
                        {summary?.active_pct ? `${summary.active_pct.toFixed(0)}%` : '0%'} ({summary?.active_rates ?? 0})
                      </span>
                    </div>
                    <div className="rm-progress-track">
                      <div className="rm-progress-fill active" style={{ width: `${summary?.active_pct ?? 0}%` }} />
                    </div>
                  </div>

                  <div className="rm-bar-item">
                    <div className="rm-bar-labels">
                      <span>Expiring Soon</span>
                      <span style={{ color: '#D97706', fontWeight: 700 }}>
                        {summary?.expiring_soon_pct ? `${summary.expiring_soon_pct.toFixed(0)}%` : '0%'} ({summary?.expiring_soon_rates ?? 0})
                      </span>
                    </div>
                    <div className="rm-progress-track">
                      <div className="rm-progress-fill expiring" style={{ width: `${summary?.expiring_soon_pct ?? 0}%` }} />
                    </div>
                  </div>

                  <div className="rm-bar-item">
                    <div className="rm-bar-labels">
                      <span>Expired</span>
                      <span style={{ color: '#DC2626', fontWeight: 700 }}>
                        {summary?.expired_pct ? `${summary.expired_pct.toFixed(0)}%` : '0%'} ({summary?.expired_rates ?? 0})
                      </span>
                    </div>
                    <div className="rm-progress-track">
                      <div className="rm-progress-fill expired" style={{ width: `${summary?.expired_pct ?? 0}%` }} />
                    </div>
                  </div>
                </div>
              </div>

              {/* Widget 2: Top Carriers by Lane Coverage */}
              <div className="rm-widget-card">
                <h3 className="rm-widget-title">
                  <span>Top Carriers</span>
                  <span className="rm-widget-title-right">Lane Coverage</span>
                </h3>

                {(!summary?.top_carriers || summary.top_carriers.length === 0) ? (
                  <p style={{ fontSize: '12.5px', color: '#94A3B8', margin: 0, fontStyle: 'italic' }}>No carrier statistics filed yet.</p>
                ) : (
                  <div className="rm-carrier-coverage-list">
                    {summary.top_carriers.map((tc, idx) => (
                      <div key={idx} className="rm-carrier-item">
                        <div className="rm-carrier-name-group">
                          <span className="rm-carrier-logo-badge" style={{ width: '24px', height: '24px', fontSize: '10.5px' }}>
                            {tc.carrier_code || tc.carrier_name?.[0]}
                          </span>
                          <span>{tc.carrier_name}</span>
                        </div>
                        <span className="rm-carrier-lane-count">{tc.lane_count} lanes ({tc.rate_count} rates)</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Widget 3: Recent Rate Updates */}
              <div className="rm-widget-card">
                <h3 className="rm-widget-title">
                  <span>Recent Rate Updates</span>
                  <span className="rm-widget-title-right">Live Feed</span>
                </h3>

                {(!summary?.recent_updates || summary.recent_updates.length === 0) ? (
                  <p style={{ fontSize: '12.5px', color: '#94A3B8', margin: 0, fontStyle: 'italic' }}>No recent rate updates.</p>
                ) : (
                  <div className="rm-recent-updates-list">
                    {summary.recent_updates.map((u) => (
                      <div key={u.id} className="rm-update-item">
                        <div className="rm-update-main">
                          <span className="rm-update-carrier">{u.carrier_name}</span>
                          <span className="rm-update-lane">{u.origin_port} → {u.dest_port}</span>
                        </div>
                        <span className="rm-update-price">{u.currency} {Number(u.base_amount).toLocaleString()}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </>
      )}

      {/* ── Slide-Over Detail Drawer (with Task 19.2 & 19.3 Tabs) ── */}
      {drawerOpen && selectedRate && (
        <div className="rm-drawer-overlay" onClick={() => setDrawerOpen(false)}>
          <div className="rm-drawer" onClick={(e) => e.stopPropagation()}>
            <div className="rm-drawer-header">
              <div>
                <h2 className="rm-drawer-title">{selectedRate.rate_reference}</h2>
                <p className="rm-drawer-sub">{selectedRate.carrier_name} • {selectedRate.rate_type} Rate Card</p>
              </div>
              <button type="button" className="rm-drawer-close" onClick={() => setDrawerOpen(false)}>
                <X size={16} />
              </button>
            </div>

            {/* Drawer Navigation Tabs */}
            <div className="rm-drawer-tabs">
              <button
                type="button"
                className={`rm-drawer-tab-btn ${drawerTab === 'overview' ? 'active' : ''}`}
                onClick={() => setDrawerTab('overview')}
              >
                <FileText size={14} /> Overview & Specs
              </button>
              <button
                type="button"
                className={`rm-drawer-tab-btn ${drawerTab === 'pricing' ? 'active' : ''}`}
                onClick={() => setDrawerTab('pricing')}
              >
                <Receipt size={14} /> Pricing & Charges ({ratePricing?.charge_count ?? 0})
              </button>
              <button
                type="button"
                className={`rm-drawer-tab-btn ${drawerTab === 'versions' ? 'active' : ''}`}
                onClick={() => setDrawerTab('versions')}
              >
                <GitBranch size={14} /> Versions & Lineage ({versionChain.length || 1})
              </button>
            </div>

            {/* Tab 1: Overview & Specs */}
            {drawerTab === 'overview' && (
              <div className="rm-drawer-body">
                {/* Status Banner */}
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: '#F8FAFC', padding: '12px 16px', borderRadius: '12px', border: '1px solid #E2E8F0' }}>
                  <span style={{ fontSize: '13px', fontWeight: 650, color: '#475569' }}>Current Validity Status</span>
                  {renderStatusBadge(selectedRate.status)}
                </div>

                {/* Pricing Snapshot */}
                <div className="rm-drawer-section">
                  <span className="rm-drawer-section-title">Commercial Base Price</span>
                  <div className="rm-drawer-card">
                    <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
                      <span style={{ fontSize: '13px', color: '#64748B' }}>Base Freight Amount:</span>
                      <strong style={{ fontSize: '20px', color: '#0F172A', fontWeight: 850 }}>
                        {selectedRate.currency} {Number(selectedRate.base_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                      </strong>
                    </div>
                    <div style={{ fontSize: '12.5px', color: '#64748B' }}>
                      Validity Window: <strong style={{ color: '#0F172A' }}>{selectedRate.validity_text || `${selectedRate.effective_date} to ${selectedRate.expiry_date}`}</strong>
                    </div>
                  </div>
                </div>

                {/* Route & Specifications */}
                <div className="rm-drawer-section">
                  <span className="rm-drawer-section-title">Routing & Equipment</span>
                  <div className="rm-drawer-card">
                    <div className="rm-drawer-grid">
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Origin Port</span>
                        <span className="rm-drawer-field-val">{selectedRate.origin_port} {selectedRate.origin_code ? `(${selectedRate.origin_code})` : ''}</span>
                      </div>
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Destination Port</span>
                        <span className="rm-drawer-field-val">{selectedRate.destination_port} {selectedRate.destination_code ? `(${selectedRate.destination_code})` : ''}</span>
                      </div>
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Transport Mode</span>
                        <span className="rm-drawer-field-val">{selectedRate.transport_mode}</span>
                      </div>
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Service Type</span>
                        <span className="rm-drawer-field-val">{selectedRate.service_type}</span>
                      </div>
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Equipment</span>
                        <span className="rm-drawer-field-val">{selectedRate.equipment_type || 'Standard'}</span>
                      </div>
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Rate Type</span>
                        <span className="rm-drawer-field-val">{selectedRate.rate_type}</span>
                      </div>
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Version</span>
                        <span className="rm-drawer-field-val">v{selectedRate.version_number || 1} ({selectedRate.version_status || 'CURRENT'})</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Carrier & Contract References */}
                <div className="rm-drawer-section">
                  <span className="rm-drawer-section-title">References & Notes</span>
                  <div className="rm-drawer-card">
                    <div className="rm-drawer-grid">
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Carrier Reference</span>
                        <span className="rm-drawer-field-val">{selectedRate.carrier_reference || '—'}</span>
                      </div>
                      <div className="rm-drawer-field">
                        <span className="rm-drawer-field-label">Contract Reference</span>
                        <span className="rm-drawer-field-val">{selectedRate.contract_reference || '—'}</span>
                      </div>
                    </div>
                    {selectedRate.notes && (
                      <div className="rm-drawer-field" style={{ marginTop: '6px' }}>
                        <span className="rm-drawer-field-label">Notes & Remarks</span>
                        <p style={{ margin: '4px 0 0', fontSize: '13px', color: '#475569', lineHeight: '1.45' }}>{selectedRate.notes}</p>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )}

            {/* Tab 2: Pricing & Charges Breakdown */}
            {drawerTab === 'pricing' && (
              <div className="rm-drawer-body">
                {pricingLoading ? (
                  <div style={{ padding: '40px', textAlign: 'center', color: '#64748B' }}>
                    <RefreshCw size={24} className="spin" style={{ margin: '0 auto 8px', display: 'block', color: '#2563EB' }} />
                    <span>Calculating commercial pricing...</span>
                  </div>
                ) : (
                  <>
                    {/* Commercial Pricing Summary Hero */}
                    <div className="rm-pricing-hero-card">
                      <div className="rm-pricing-hero-top">
                        <span className="rm-pricing-hero-label">Commercial Total</span>
                        <span className="rm-pricing-hero-val">
                          {selectedRate.currency} {Number(ratePricing?.commercial_total ?? selectedRate.base_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                        </span>
                      </div>

                      <div className="rm-pricing-hero-grid">
                        <div className="rm-pricing-hero-subfield">
                          <span className="rm-pricing-hero-sublabel">Base Carrier Rate</span>
                          <span className="rm-pricing-hero-subval">
                            {selectedRate.currency} {Number(ratePricing?.base_rate ?? selectedRate.base_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                          </span>
                        </div>
                        <div className="rm-pricing-hero-subfield">
                          <span className="rm-pricing-hero-sublabel">Additional Charges</span>
                          <span className="rm-pricing-hero-subval" style={{ color: '#60A5FA' }}>
                            + {selectedRate.currency} {Number(ratePricing?.additional_charges ?? 0).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* Multi-currency notice if applicable */}
                    {ratePricing?.is_multi_currency && (
                      <div className="rm-multi-currency-alert">
                        <AlertTriangle size={16} style={{ flexShrink: 0 }} />
                        <span>
                          Notice: This rate contains charges in multiple currencies ({ratePricing.currencies.join(', ')}). Totals are maintained per-currency without automated conversion.
                        </span>
                      </div>
                    )}

                    {/* Category Totals Pills */}
                    {ratePricing?.category_totals?.length > 0 && (
                      <div className="rm-drawer-section">
                        <span className="rm-drawer-section-title">Category Breakdown</span>
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                          {ratePricing.category_totals.map((cat, idx) => (
                            <div key={idx} style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '6px 10px', fontSize: '12px' }}>
                              <span style={{ fontWeight: 700, color: '#475569' }}>{cat.Category}: </span>
                              <strong style={{ color: '#0F172A' }}>{cat.Currency} {Number(cat.TotalAmount).toLocaleString(undefined, { minimumFractionDigits: 2 })}</strong>
                              <span style={{ color: '#94A3B8', marginLeft: '4px' }}>({cat.ChargeCount})</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Charges Table Header & Add Button */}
                    <div className="rm-drawer-section">
                      <div className="rm-charges-header-row">
                        <h3 className="rm-charges-title">Itemized Charges ({ratePricing?.charges?.length ?? 0})</h3>
                        {selectedRate.status !== 'ARCHIVED' && (
                          <button
                            type="button"
                            className="rm-btn-add-charge"
                            onClick={handleOpenAddCharge}
                          >
                            <Plus size={13} /> Add Charge Item
                          </button>
                        )}
                      </div>

                      {(!ratePricing?.charges || ratePricing.charges.length === 0) ? (
                        <div style={{ background: '#F8FAFC', border: '1px dashed #CBD5E1', borderRadius: '12px', padding: '24px 16px', textAlign: 'center' }}>
                          <Receipt size={28} style={{ color: '#94A3B8', margin: '0 auto 8px', display: 'block' }} />
                          <p style={{ margin: '0 0 8px', fontSize: '13px', fontWeight: 650, color: '#475569' }}>No additional charges filed</p>
                          <p style={{ margin: '0 0 12px', fontSize: '12px', color: '#64748B', maxWidth: '340px', marginLeft: 'auto', marginRight: 'auto' }}>
                            Add itemized freight charges (THC, Documentation, Surcharges, Customs) to build a complete commercial rate breakdown.
                          </p>
                          {selectedRate.status !== 'ARCHIVED' && (
                            <button
                              type="button"
                              className="rm-btn-add-charge"
                              onClick={handleOpenAddCharge}
                            >
                              <Plus size={13} /> Add First Charge
                            </button>
                          )}
                        </div>
                      ) : (
                        <div className="rm-charges-table-wrap">
                          <table className="rm-charges-table">
                            <thead>
                              <tr>
                                <th>Charge Name</th>
                                <th>Category</th>
                                <th>Basis</th>
                                <th>Rate / Qty</th>
                                <th>Amount</th>
                                <th>Actions</th>
                              </tr>
                            </thead>
                            <tbody>
                              {ratePricing.charges.map((ch) => (
                                <tr key={ch.id}>
                                  <td>
                                    <div style={{ fontWeight: 700, color: '#0F172A' }}>{ch.charge_name}</div>
                                    {ch.charge_code && <span style={{ fontSize: '10.5px', color: '#64748B' }}>{ch.charge_code}</span>}
                                  </td>
                                  <td>
                                    <span className={`rm-charge-cat-tag rm-charge-cat-tag--${ch.charge_category}`}>
                                      {ch.charge_category}
                                    </span>
                                  </td>
                                  <td>
                                    <span className="rm-charge-calc-badge">{ch.calculation_basis}</span>
                                  </td>
                                  <td>
                                    {ch.calculation_basis === 'PERCENTAGE' ? (
                                      <span>{ch.unit_price}%</span>
                                    ) : (
                                      <span>{ch.currency} {Number(ch.unit_price).toLocaleString()} × {ch.quantity}</span>
                                    )}
                                  </td>
                                  <td>
                                    <strong style={{ color: ch.included_in_base_rate ? '#16A34A' : '#0F172A' }}>
                                      {ch.included_in_base_rate ? 'INCLUDED' : `${ch.currency} ${Number(ch.calculated_amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}`}
                                    </strong>
                                  </td>
                                  <td>
                                    {selectedRate.status !== 'ARCHIVED' ? (
                                      <div style={{ display: 'flex', gap: '4px' }}>
                                        <button
                                          type="button"
                                          className="rm-action-btn"
                                          title="Edit Charge"
                                          onClick={() => handleOpenEditCharge(ch)}
                                        >
                                          <Edit size={12} />
                                        </button>
                                        <button
                                          type="button"
                                          className="rm-action-btn"
                                          title="Delete Charge"
                                          style={{ color: '#DC2626' }}
                                          onClick={() => handleDeleteCharge(ch.id)}
                                        >
                                          <Trash2 size={12} />
                                        </button>
                                      </div>
                                    ) : (
                                      <span style={{ fontSize: '11px', color: '#94A3B8' }}>Locked</span>
                                    )}
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      )}
                    </div>
                  </>
                )}
              </div>
            )}

            {/* Tab 3: Versions & Lineage */}
            {drawerTab === 'versions' && (
              <div className="rm-drawer-body">
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <span className="rm-drawer-section-title">Commercial Version Timeline</span>
                  {selectedRate.status !== 'ARCHIVED' && (
                    <button
                      type="button"
                      className="rc-btn-primary"
                      style={{ padding: '6px 12px', fontSize: '12px' }}
                      onClick={(e) => handleOpenCreateVersion(selectedRate, e)}
                    >
                      <Plus size={13} /> Create New Version
                    </button>
                  )}
                </div>

                {versionsLoading ? (
                  <div style={{ textAlign: 'center', padding: '30px', color: '#64748B' }}>Loading version lineage...</div>
                ) : (
                  <div className="rc-version-timeline">
                    {versionChain.length === 0 ? (
                      <div className="rc-version-card is-current">
                        <div className="rc-version-header">
                          <span className="rc-version-number">
                            <GitBranch size={14} style={{ color: '#2563EB' }} /> Version 1
                          </span>
                          <span className="rc-badge-pill rc-badge-active">CURRENT</span>
                        </div>
                        <div className="rc-version-amount">
                          {selectedRate.currency} {Number(selectedRate.base_amount).toLocaleString()}
                        </div>
                        <div className="rc-version-dates">
                          Validity: {selectedRate.effective_date} to {selectedRate.expiry_date}
                        </div>
                      </div>
                    ) : (
                      versionChain.map((ver, idx) => (
                        <React.Fragment key={ver.rate_id}>
                          <div className={`rc-version-card ${ver.version_status === 'CURRENT' ? 'is-current' : 'is-superseded'}`}>
                            <div className="rc-version-header">
                              <span className="rc-version-number">
                                <GitBranch size={14} style={{ color: ver.version_status === 'CURRENT' ? '#2563EB' : '#64748B' }} />
                                Version {ver.version_number}
                              </span>
                              <span className={`rc-badge-pill ${ver.version_status === 'CURRENT' ? 'rc-badge-active' : 'rc-badge-draft'}`}>
                                {ver.version_status}
                              </span>
                            </div>
                            <div className="rc-version-amount">
                              {ver.currency} {Number(ver.base_amount).toLocaleString()}
                            </div>
                            <div className="rc-version-dates">
                              Validity: {ver.effective_date} to {ver.expiry_date}
                            </div>
                          </div>
                          {idx < versionChain.length - 1 && (
                            <div className="rc-version-arrow">↓</div>
                          )}
                        </React.Fragment>
                      ))
                    )}

                    {/* Audit History Panel */}
                    <div className="rm-drawer-section" style={{ marginTop: '16px' }}>
                      <span className="rm-drawer-section-title">Audit Log & Revisions</span>
                      {versionHistory.length === 0 ? (
                        <div style={{ fontSize: '12px', color: '#94A3B8', fontStyle: 'italic', padding: '8px 0' }}>
                          No audit events recorded for this rate card.
                        </div>
                      ) : (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                          {versionHistory.map((h) => (
                            <div key={h.id} style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '10px 12px', fontSize: '12px' }}>
                              <div style={{ fontWeight: 650, color: '#0F172A' }}>{h.description}</div>
                              <div style={{ color: '#64748B', marginTop: '4px', display: 'flex', justifyContent: 'space-between' }}>
                                <span>By {h.performed_by || 'Commercial Team'}</span>
                                <span>{new Date(h.created_at).toLocaleDateString()}</span>
                              </div>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}

            <div className="rm-drawer-footer">
              <button
                type="button"
                className="rc-btn-primary"
                style={{ padding: '8px 14px', fontSize: '12.5px' }}
                onClick={(e) => handleOpenCreateVersion(selectedRate, e)}
              >
                <GitBranch size={14} /> New Version
              </button>
              <button
                type="button"
                className="rm-btn-reset"
                onClick={(e) => handleOpenEdit(selectedRate, e)}
              >
                <Edit3 size={14} /> Edit Rate
              </button>
              <button
                type="button"
                className="rm-btn-reset"
                style={{ color: '#DC2626', borderColor: '#FECACA' }}
                onClick={(e) => handleArchiveRate(selectedRate.id, e)}
              >
                <Trash2 size={14} /> Archive
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Create Rate Modal ── */}
      {createModalOpen && (
        <div className="rm-modal-overlay" onClick={() => setCreateModalOpen(false)}>
          <div className="rm-modal" onClick={(e) => e.stopPropagation()}>
            <div className="rm-modal-header">
              <h2 className="rm-modal-title">Create New Carrier Rate Card</h2>
              <button type="button" className="rm-drawer-close" onClick={() => setCreateModalOpen(false)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleCreateSubmit}>
              <div className="rm-modal-body">
                <div className="rm-modal-form-grid">
                  <div className="rm-field-group">
                    <label>Carrier / Supplier *</label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. Maersk Line"
                      className="rm-input"
                      value={formData.carrier_name}
                      onChange={(e) => setFormData({ ...formData, carrier_name: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Carrier SCAC Code</label>
                    <input
                      type="text"
                      placeholder="e.g. MAEU"
                      className="rm-input"
                      value={formData.carrier_code}
                      onChange={(e) => setFormData({ ...formData, carrier_code: e.target.value.toUpperCase() })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Origin Port / City *</label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. USNYC, New York"
                      className="rm-input"
                      value={formData.origin_port}
                      onChange={(e) => setFormData({ ...formData, origin_port: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Destination Port / City *</label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. INNSA, Nhava Sheva"
                      className="rm-input"
                      value={formData.destination_port}
                      onChange={(e) => setFormData({ ...formData, destination_port: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Transport Mode</label>
                    <CustomSelect
                      value={formData.transport_mode}
                      onChange={(val) => setFormData({ ...formData, transport_mode: val })}
                      options={TRANSPORT_MODES.filter(m => m !== 'ALL')}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Service Type</label>
                    <CustomSelect
                      value={formData.service_type}
                      onChange={(val) => setFormData({ ...formData, service_type: val })}
                      options={SERVICE_TYPES.filter(st => st !== 'ALL')}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Equipment Type</label>
                    <CustomSelect
                      value={formData.equipment_type}
                      onChange={(val) => setFormData({ ...formData, equipment_type: val })}
                      options={EQUIPMENT_TYPES.filter(eq => eq !== 'ALL')}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Rate Type</label>
                    <CustomSelect
                      value={formData.rate_type}
                      onChange={(val) => setFormData({ ...formData, rate_type: val })}
                      options={RATE_TYPES.filter(rt => rt !== 'ALL')}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Base Freight Amount (USD) *</label>
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      required
                      placeholder="e.g. 2850.00"
                      className="rm-input"
                      value={formData.base_amount}
                      onChange={(e) => setFormData({ ...formData, base_amount: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Currency</label>
                    <input
                      type="text"
                      className="rm-input"
                      value={formData.currency}
                      onChange={(e) => setFormData({ ...formData, currency: e.target.value.toUpperCase() })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Effective Date *</label>
                    <input
                      type="date"
                      required
                      className="rm-input"
                      value={formData.effective_date}
                      onChange={(e) => setFormData({ ...formData, effective_date: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Expiry Date *</label>
                    <input
                      type="date"
                      required
                      className="rm-input"
                      value={formData.expiry_date}
                      onChange={(e) => setFormData({ ...formData, expiry_date: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Carrier Reference</label>
                    <input
                      type="text"
                      placeholder="e.g. MAEU-2026-N09"
                      className="rm-input"
                      value={formData.carrier_reference}
                      onChange={(e) => setFormData({ ...formData, carrier_reference: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Contract Reference</label>
                    <input
                      type="text"
                      placeholder="e.g. CTR-2026-448"
                      className="rm-input"
                      value={formData.contract_reference}
                      onChange={(e) => setFormData({ ...formData, contract_reference: e.target.value })}
                    />
                  </div>
                </div>

                <div className="rm-field-group">
                  <label>Notes & Routing Conditions</label>
                  <textarea
                    rows={2}
                    placeholder="Include detention terms, bunker clauses, or validity caveats..."
                    className="rm-input"
                    value={formData.notes}
                    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  />
                </div>
              </div>

              <div className="rm-modal-footer">
                <button type="button" className="rm-btn-reset" onClick={() => setCreateModalOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="rm-btn-create" disabled={formSubmitting}>
                  {formSubmitting ? 'Creating...' : 'Create Rate Card'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Edit Rate Modal ── */}
      {editModalOpen && (
        <div className="rm-modal-overlay" onClick={() => setEditModalOpen(false)}>
          <div className="rm-modal" onClick={(e) => e.stopPropagation()}>
            <div className="rm-modal-header">
              <h2 className="rm-modal-title">Edit Rate Card {selectedRate?.rate_reference}</h2>
              <button type="button" className="rm-drawer-close" onClick={() => setEditModalOpen(false)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleEditSubmit}>
              <div className="rm-modal-body">
                <div className="rm-modal-form-grid">
                  <div className="rm-field-group">
                    <label>Carrier / Supplier</label>
                    <input
                      type="text"
                      className="rm-input"
                      value={formData.carrier_name}
                      onChange={(e) => setFormData({ ...formData, carrier_name: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Carrier Code</label>
                    <input
                      type="text"
                      className="rm-input"
                      value={formData.carrier_code}
                      onChange={(e) => setFormData({ ...formData, carrier_code: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Origin Port / Code</label>
                    <input
                      type="text"
                      className="rm-input"
                      value={formData.origin_port}
                      onChange={(e) => setFormData({ ...formData, origin_port: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Destination Port / Code</label>
                    <input
                      type="text"
                      className="rm-input"
                      value={formData.destination_port}
                      onChange={(e) => setFormData({ ...formData, destination_port: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Base Freight Amount</label>
                    <input
                      type="number"
                      step="0.01"
                      className="rm-input"
                      value={formData.base_amount}
                      onChange={(e) => setFormData({ ...formData, base_amount: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Currency</label>
                    <input
                      type="text"
                      className="rm-input"
                      value={formData.currency}
                      onChange={(e) => setFormData({ ...formData, currency: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Effective Date</label>
                    <input
                      type="date"
                      className="rm-input"
                      value={formData.effective_date}
                      onChange={(e) => setFormData({ ...formData, effective_date: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Expiry Date</label>
                    <input
                      type="date"
                      className="rm-input"
                      value={formData.expiry_date}
                      onChange={(e) => setFormData({ ...formData, expiry_date: e.target.value })}
                    />
                  </div>
                </div>

                <div className="rm-field-group">
                  <label>Notes</label>
                  <textarea
                    rows={2}
                    className="rm-input"
                    value={formData.notes}
                    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  />
                </div>
              </div>

              <div className="rm-modal-footer">
                <button type="button" className="rm-btn-reset" onClick={() => setEditModalOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="rm-btn-create" disabled={formSubmitting}>
                  {formSubmitting ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Add / Edit Rate Charge Modal (Task 19.2) ── */}
      {chargeModalOpen && (
        <div className="rm-modal-overlay" onClick={() => setChargeModalOpen(false)}>
          <div className="rm-modal" onClick={(e) => e.stopPropagation()}>
            <div className="rm-modal-header">
              <h2 className="rm-modal-title">
                {editingCharge ? `Edit Charge: ${editingCharge.charge_name}` : `Add Charge to ${selectedRate?.rate_reference}`}
              </h2>
              <button type="button" className="rm-drawer-close" onClick={() => setChargeModalOpen(false)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleChargeSubmit}>
              <div className="rm-modal-body">
                <div className="rm-modal-form-grid">
                  <div className="rm-field-group">
                    <label>Charge Category *</label>
                    <CustomSelect
                      value={chargeForm.charge_category}
                      onChange={(val) => setChargeForm({ ...chargeForm, charge_category: val })}
                      options={CHARGE_CATEGORIES}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Charge Name *</label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. Origin Terminal Handling Charge"
                      className="rm-input"
                      value={chargeForm.charge_name}
                      onChange={(e) => setChargeForm({ ...chargeForm, charge_name: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Charge Code</label>
                    <input
                      type="text"
                      placeholder="e.g. OTHC, BAF, DOCS"
                      className="rm-input"
                      value={chargeForm.charge_code}
                      onChange={(e) => setChargeForm({ ...chargeForm, charge_code: e.target.value.toUpperCase() })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Calculation Basis *</label>
                    <CustomSelect
                      value={chargeForm.calculation_basis}
                      onChange={(val) => setChargeForm({ ...chargeForm, calculation_basis: val })}
                      options={CALCULATION_BASES}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>{chargeForm.calculation_basis === 'PERCENTAGE' ? 'Percentage (%) *' : 'Unit Rate *'}</label>
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      required
                      placeholder={chargeForm.calculation_basis === 'PERCENTAGE' ? 'e.g. 5.0' : 'e.g. 150.00'}
                      className="rm-input"
                      value={chargeForm.unit_price}
                      onChange={(e) => setChargeForm({ ...chargeForm, unit_price: e.target.value })}
                    />
                  </div>

                  {chargeForm.calculation_basis !== 'PERCENTAGE' && chargeForm.calculation_basis !== 'FLAT' && (
                    <div className="rm-field-group">
                      <label>Quantity</label>
                      <input
                        type="number"
                        step="0.01"
                        min="0"
                        className="rm-input"
                        value={chargeForm.quantity}
                        onChange={(e) => setChargeForm({ ...chargeForm, quantity: e.target.value })}
                      />
                    </div>
                  )}

                  <div className="rm-field-group">
                    <label>Currency</label>
                    <input
                      type="text"
                      className="rm-input"
                      value={chargeForm.currency}
                      onChange={(e) => setChargeForm({ ...chargeForm, currency: e.target.value.toUpperCase() })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Minimum Amount (Optional)</label>
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      placeholder="e.g. 50.00"
                      className="rm-input"
                      value={chargeForm.minimum_amount}
                      onChange={(e) => setChargeForm({ ...chargeForm, minimum_amount: e.target.value })}
                    />
                  </div>

                  <div className="rm-field-group">
                    <label>Maximum Amount (Optional)</label>
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      placeholder="e.g. 500.00"
                      className="rm-input"
                      value={chargeForm.maximum_amount}
                      onChange={(e) => setChargeForm({ ...chargeForm, maximum_amount: e.target.value })}
                    />
                  </div>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '8px 0' }}>
                  <input
                    type="checkbox"
                    id="inclBaseRate"
                    checked={chargeForm.included_in_base_rate}
                    onChange={(e) => setChargeForm({ ...chargeForm, included_in_base_rate: e.target.checked })}
                  />
                  <label htmlFor="inclBaseRate" style={{ fontSize: '12.5px', fontWeight: 600, color: '#334155', cursor: 'pointer' }}>
                    Already Included in Base Freight Amount (does not add to commercial total)
                  </label>
                </div>

                {/* Calculation preview banner */}
                <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '10px', padding: '10px 14px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <span style={{ fontSize: '12.5px', color: '#64748B', fontWeight: 650 }}>Estimated Line Amount:</span>
                  <strong style={{ fontSize: '15px', color: chargeForm.included_in_base_rate ? '#16A34A' : '#2563EB', fontWeight: 800 }}>
                    {chargeForm.included_in_base_rate ? 'INCLUDED IN BASE' : `${chargeForm.currency || 'USD'} ${previewChargeAmount().toLocaleString(undefined, { minimumFractionDigits: 2 })}`}
                  </strong>
                </div>

                <div className="rm-field-group">
                  <label>Notes & Conditions</label>
                  <textarea
                    rows={2}
                    placeholder="Specific tariffs, demurrage rules, or charge conditions..."
                    className="rm-input"
                    value={chargeForm.notes}
                    onChange={(e) => setChargeForm({ ...chargeForm, notes: e.target.value })}
                  />
                </div>
              </div>

              <div className="rm-modal-footer">
                <button type="button" className="rm-btn-reset" onClick={() => setChargeModalOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="rm-btn-create" disabled={formSubmitting}>
                  {formSubmitting ? 'Saving...' : editingCharge ? 'Update Charge' : 'Add Charge Item'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Modal: Create New Rate Version (Task 19.3) ── */}
      {versionModalOpen && selectedRate && (
        <div className="rm-modal-overlay" onClick={() => setVersionModalOpen(false)}>
          <div className="rm-modal" style={{ maxWidth: '580px' }} onClick={(e) => e.stopPropagation()}>
            <div className="rm-modal-header">
              <h2 className="rm-modal-title">Create New Version of Rate</h2>
              <button type="button" className="rm-drawer-close" onClick={() => setVersionModalOpen(false)}>
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleCreateVersionSubmit}>
              <div className="rm-modal-body">
                <div style={{ background: '#EFF6FF', border: '1px solid #BFDBFE', padding: '12px 14px', borderRadius: '10px', fontSize: '12.5px', color: '#1E40AF' }}>
                  <strong>Creating Version {(selectedRate.version_number || 1) + 1}:</strong> Previous version {(selectedRate.version_number || 1)} will remain historically preserved for existing quotations and marked as superseded.
                </div>

                <div className="rm-modal-form-grid">
                  <div className="rm-field-group">
                    <label>Base Freight Amount *</label>
                    <input
                      type="number"
                      step="0.01"
                      required
                      className="rm-input"
                      value={versionForm.base_amount}
                      onChange={(e) => setVersionForm({ ...versionForm, base_amount: e.target.value })}
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Currency</label>
                    <CustomSelect
                      value={versionForm.currency}
                      onChange={(val) => setVersionForm({ ...versionForm, currency: val })}
                      options={['USD', 'EUR', 'GBP', 'INR', 'AED', 'SGD', 'CNY']}
                    />
                  </div>
                </div>

                <div className="rm-modal-form-grid">
                  <div className="rm-field-group">
                    <label>Effective Date *</label>
                    <input
                      type="date"
                      required
                      className="rm-input"
                      value={versionForm.effective_date}
                      onChange={(e) => setVersionForm({ ...versionForm, effective_date: e.target.value })}
                    />
                  </div>
                  <div className="rm-field-group">
                    <label>Expiry Date *</label>
                    <input
                      type="date"
                      required
                      className="rm-input"
                      value={versionForm.expiry_date}
                      onChange={(e) => setVersionForm({ ...versionForm, expiry_date: e.target.value })}
                    />
                  </div>
                </div>

                <div className="rm-field-group">
                  <label>Revision Reason / Audit Note *</label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. Q3 Carrier GRI Adjustment / Bunker revision"
                    className="rm-input"
                    value={versionForm.revision_reason}
                    onChange={(e) => setVersionForm({ ...versionForm, revision_reason: e.target.value })}
                  />
                </div>

                <div className="rm-field-group">
                  <label>Notes</label>
                  <textarea
                    rows={2}
                    placeholder="Additional terms or contract reference..."
                    className="rm-input"
                    value={versionForm.notes}
                    onChange={(e) => setVersionForm({ ...versionForm, notes: e.target.value })}
                  />
                </div>
              </div>

              <div className="rm-modal-footer">
                <button type="button" className="rm-btn-reset" onClick={() => setVersionModalOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="rc-btn-primary" disabled={formSubmitting}>
                  {formSubmitting ? 'Creating...' : `Save as Version ${(selectedRate.version_number || 1) + 1}`}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Modal: Archive Rate Confirmation ── */}
      <ConfirmModal
        isOpen={archiveRateModalOpen}
        title="Archive Freight Rate"
        message="Are you sure you want to archive this freight rate card? This will preserve historical records but mark the rate as archived."
        confirmText="Archive Rate"
        confirmStyle="danger"
        isLoading={archiveRateLoading}
        onConfirm={handleConfirmArchiveRate}
        onCancel={() => {
          setArchiveRateModalOpen(false);
          setRateToArchiveId(null);
        }}
      />
    </div>
  );
}
