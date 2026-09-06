import React, { useEffect, useMemo, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  ChevronLeft,
  Edit2,
  Plus,
  Building2,
  Users,
  FileText,
  FileSpreadsheet,
  ScrollText,
  Ship,
  DollarSign,
  CreditCard,
  MapPin,
  Calendar,
  Clock,
  Mail,
  Phone,
  Globe,
  ExternalLink,
  ArrowRight,
  TrendingUp,
  AlertTriangle,
  CheckCircle2,
  RefreshCw,
  ShieldAlert,
  Zap,
  FileCheck,
  UserPlus,
  Package,
  MoreHorizontal,
  BriefcaseBusiness,
  PackageOpen,
  Truck,
  Activity,
  ShieldCheck,
  Folder,
  XCircle,
  HelpCircle,
  AlertCircle
} from 'lucide-react';

import customerService from '../../../services/customerService';
import api from '../../../services/api';
import CustomerModal from './CustomerModal';
import DocumentUploadModal from '../Documents/DocumentUploadModal';
import DocumentDetailsModal from '../Documents/DocumentDetailsModal';
import './CustomerDetailsPage.css';

const safeArray = (value) => (Array.isArray(value) ? value : []);

const getInitials = (name = '') => {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || 'CU';
};

const formatDate = (value) => {
  if (!value) return '—';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
};

const formatCurrency = (value, currency = 'USD') => {
  const amount = Number(value || 0);
  try {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD',
      maximumFractionDigits: 0,
    }).format(amount);
  } catch {
    return `$${amount.toLocaleString()}`;
  }
};

const normalizeStatus = (value = '') =>
  String(value)
    .replace(/_/g, ' ')
    .toLowerCase();

export default function CustomerDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [refreshingIntel, setRefreshingIntel] = useState(false);

  const [customer, setCustomer] = useState(null);
  const [kpis, setKPIs] = useState({});
  const [rfqs, setRfqs] = useState([]);
  const [quotations, setQuotations] = useState([]);
  const [bookings, setBookings] = useState([]);
  const [shipments, setShipments] = useState([]);
  const [contracts, setContracts] = useState([]);
  const [timeline, setTimeline] = useState([]);
  const [contacts, setContacts] = useState([]);
  const [addresses, setAddresses] = useState([]);

  const [financialProfile, setFinancialProfile] = useState({});
  const [commercialMetrics, setCommercialMetrics] = useState({});
  const [accountOwnership, setAccountOwnership] = useState({});
  const [ownershipHistory, setOwnershipHistory] = useState([]);

  const [intelProfile, setIntelProfile] = useState({});
  const [openRisks, setOpenRisks] = useState([]);
  const [opportunities, setOpportunities] = useState([]);

  const [activeTab, setActiveTab] = useState('overview');
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  const [customerDocs, setCustomerDocs] = useState([]);
  const [isDocUploadOpen, setIsDocUploadOpen] = useState(false);
  const [selectedDocForDetails, setSelectedDocForDetails] = useState(null);

  const fetchCustomerDocuments = async () => {
    try {
      const res = await api.get(`/api/v1/documents?customer_id=${id}`);
      const list = Array.isArray(res) ? res : (res?.data || []);
      setCustomerDocs(list);
    } catch (err) {
      console.warn('Could not load customer documents:', err);
    }
  };

  useEffect(() => {
    if (id) {
      fetchCustomerDocuments();
    }
  }, [id]);

  const fetchCustomer360 = async () => {
    try {
      setLoading(true);
      setError(null);

      const results = await Promise.allSettled([
        customerService.getCustomerDashboard(id),
        customerService.getCustomerRFQs(id),
        customerService.getCustomerQuotations(id),
        customerService.getCustomerBookings(id),
        customerService.getCustomerShipments(id),
        customerService.getCustomerContracts(id),
        customerService.getCustomerTimeline(id),
        customerService.getFinancialProfile(id),
        customerService.getCommercialMetrics(id),
        customerService.getAccountOwnership(id),
        customerService.getOwnershipHistory(id),
        customerService.evaluateCustomerIntelligence(id),
        customerService.getCustomerRisks(id),
        customerService.getCustomerOpportunities(id),
      ]);

      const getResult = (index, fallback) => {
        return results[index]?.status === 'fulfilled'
          ? results[index].value
          : fallback;
      };

      const dash = getResult(0, null);

      if (!dash || !dash.customer) {
        throw new Error('Customer dashboard could not be loaded.');
      }

      const customerData = dash.customer;
      const rfqData = getResult(1, dash.recent_rfqs || []);
      const quoteData = getResult(2, dash.recent_quotations || []);
      const bookingData = getResult(3, dash.recent_bookings || []);
      const shipmentData = getResult(4, dash.recent_shipments || []);
      const contractData = getResult(5, dash.recent_contracts || []);
      const timelineData = getResult(6, dash.timeline || []);

      const financialData = getResult(7, {});
      const commercialData = getResult(8, {});
      const ownershipData = getResult(9, {});
      const ownershipHistoryData = getResult(10, {});
      const intelligenceData = getResult(11, {});
      const riskData = getResult(12, []);
      const opportunityData = getResult(13, []);

      setCustomer(customerData);
      setKPIs(dash.kpis || {});

      setRfqs(safeArray(rfqData));
      setQuotations(safeArray(quoteData));
      setBookings(safeArray(bookingData));
      setShipments(safeArray(shipmentData));
      setContracts(safeArray(contractData));
      setTimeline(safeArray(timelineData));

      setContacts(safeArray(customerData.contacts));
      setAddresses(safeArray(customerData.addresses));

      setFinancialProfile(financialData || {});
      setCommercialMetrics(commercialData || {});
      setAccountOwnership(ownershipData || {});
      setOwnershipHistory(safeArray(ownershipHistoryData));

      setIntelProfile(intelligenceData || {});
      setOpenRisks(
        safeArray(riskData).length
          ? safeArray(riskData)
          : safeArray(intelligenceData?.open_risks)
      );
      setOpportunities(
        safeArray(opportunityData).length
          ? safeArray(opportunityData)
          : safeArray(intelligenceData?.detected_opportunities)
      );
    } catch (err) {
      console.error('Failed to load Customer 360:', err);
      setError('Unable to load this customer account.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCustomer360();
  }, [id]);

  const currency =
    financialProfile.currency ||
    customer?.currency ||
    'USD';

  const health = intelProfile?.health || {};
  const healthScore = health.health_score;
  const healthStatus = health.health_status;

  const activeContractsCount =
    kpis.active_contracts ??
    contracts.filter(
      (contract) =>
        String(contract.status || '').toUpperCase() === 'ACTIVE'
    ).length;

  const activeShipmentsCount =
    kpis.active_shipments ??
    shipments.filter((shipment) =>
      ['ACTIVE', 'IN_TRANSIT', 'IN_PROGRESS'].includes(
        String(shipment.status || '').toUpperCase()
      )
    ).length;

  const totalQuotedValue =
    commercialMetrics.total_quotation_value ??
    commercialMetrics.quotation_pipeline ??
    0;

  const acceptedValue =
    commercialMetrics.accepted_quotation_value ??
    commercialMetrics.accepted_value ??
    0;

  const activeContractValue =
    commercialMetrics.active_contract_value ??
    commercialMetrics.total_active_contract_value ??
    0;

  const tabs = [
    { id: 'overview', label: 'Overview' },
    { id: 'contacts', label: `Contacts (${contacts.length})` },
    { id: 'addresses', label: `Addresses (${addresses.length})` },
    { id: 'commercial', label: 'Commercial' },
    { id: 'operational', label: 'Operational' },
    { id: 'financial', label: 'Financial' },
    { id: 'documents', label: `Documents (${customerDocs.length})` },
    { id: 'timeline', label: 'Timeline' },
    {
      id: 'intelligence',
      label: openRisks.length > 0 ? `Intelligence (${openRisks.length})` : 'Intelligence',
    },
  ];

  const renderCreditStatus = () => {
    const status =
      financialProfile.credit_status ||
      customer?.credit_status ||
      'GOOD_STANDING';

    const className = normalizeStatus(status).replace(/\s/g, '-');

    return (
      <span className={`customer-credit-badge credit-${className}`}>
        {status.replace(/_/g, ' ')}
      </span>
    );
  };

  const renderHealthStatus = () => {
    if (!healthStatus || healthScore === undefined) {
      return (
        <span className="customer-health-badge health-neutral">
          Insufficient Data (50)
        </span>
      );
    }

    const statusClass = normalizeStatus(healthStatus).replace(/\s/g, '-');
    return (
      <span className={`customer-health-badge health-${statusClass}`}>
        {healthStatus.replace(/_/g, ' ')} ({healthScore})
      </span>
    );
  };

  const handleRefreshIntel = async () => {
    setRefreshingIntel(true);
    try {
      await customerService.evaluateCustomerIntelligence(id);
      await fetchCustomer360();
    } catch (err) {
      console.error(err);
    } finally {
      setRefreshingIntel(false);
    }
  };

  if (loading) {
    return (
      <div className="cust-details-loading">
        <div className="cust-spinner" />
        <span>Loading Customer 360° Profile...</span>
      </div>
    );
  }

  if (error || !customer) {
    return (
      <div className="cust-details-error">
        <AlertTriangle size={32} className="text-amber-500" />
        <h3>Customer Profile Not Found</h3>
        <p>{error}</p>
        <button
          className="btn-secondary-action"
          onClick={() => navigate('/dashboard/customers')}
        >
          <ChevronLeft size={16} />
          Back to Customers
        </button>
      </div>
    );
  }

  return (
    <div className="customer-360-page">
      
      {/* ── BREADCRUMB ── */}
      <div className="customer-360-breadcrumb">
        <Link to="/dashboard/customers" className="breadcrumb-link">
          <ChevronLeft size={15} />
          Customers
        </Link>
        <span className="breadcrumb-sep">/</span>
        <span className="breadcrumb-current">{customer.name}</span>
        <span className="breadcrumb-sep">/</span>
        <strong>Customer 360°</strong>
      </div>

      {/* ── HERO HEADER CARD ── */}
      <section className="customer-360-header">
        <div className="customer-header-main">
          
          <div className="customer-avatar-box">
            <Building2 size={24} className="text-blue-600" />
          </div>

          <div className="customer-title-block">
            <div className="customer-title-line">
              <h1>{customer.name}</h1>
              <span className={`customer-status-badge ${customer.status === 'ACTIVE' ? 'active-customer' : 'inactive-customer'}`}>
                {customer.status === 'ACTIVE' ? 'Active Customer' : customer.status || 'Active'}
              </span>
            </div>

            <div className="customer-header-meta">
              <span className="customer-code-pill">{customer.customer_code || customer.code || `CUST-2026-${id}`}</span>
              <span className="meta-separator">•</span>
              <span className="customer-type-pill">{customer.customer_type || 'SHIPPER'}</span>
              <span className="meta-separator">•</span>
              <span className="meta-with-icon">
                <Calendar size={13} className="text-slate-400" />
                Since {formatDate(customer.created_at)}
              </span>
              {(customer.city || customer.country) && (
                <>
                  <span className="meta-separator">•</span>
                  <span className="meta-with-icon">
                    <MapPin size={13} className="text-slate-400" />
                    {[customer.city, customer.country].filter(Boolean).join(', ')}
                  </span>
                </>
              )}
            </div>
          </div>

        </div>

        <div className="customer-header-actions">
          <button
            type="button"
            className="btn-secondary-action"
            onClick={() => setIsEditModalOpen(true)}
          >
            <Edit2 size={15} />
            Edit Customer
          </button>

          <button
            type="button"
            className="btn-secondary-action"
            onClick={() => setActiveTab('financial')}
          >
            <DollarSign size={15} />
            Financial
          </button>

          <button
            type="button"
            className="btn-primary-action"
            onClick={() => navigate('/dashboard/rfqs')}
          >
            <Plus size={16} />
            New RFQ
          </button>
        </div>
      </section>

      {/* ── TOP KPI SUMMARY GRID ── */}
      <section className="customer-kpi-grid">
        
        {/* Total RFQs */}
        <div className="customer-kpi-card" onClick={() => setActiveTab('commercial')}>
          <div className="customer-kpi-icon icon-purple">
            <FileText size={19} />
          </div>
          <div className="customer-kpi-content">
            <span className="kpi-label">Total RFQs</span>
            <strong className="kpi-value">{commercialMetrics.total_rfqs ?? kpis.total_rfqs ?? rfqs.length ?? 0}</strong>
            <small className="kpi-sublabel">Customer opportunities</small>
          </div>
        </div>

        {/* Quotations */}
        <div className="customer-kpi-card" onClick={() => setActiveTab('commercial')}>
          <div className="customer-kpi-icon icon-blue">
            <FileSpreadsheet size={19} />
          </div>
          <div className="customer-kpi-content">
            <span className="kpi-label">Quotations</span>
            <strong className="kpi-value">{commercialMetrics.total_quotations ?? quotations.length ?? 0}</strong>
            <small className="kpi-sublabel">{formatCurrency(totalQuotedValue, currency)} quoted value</small>
          </div>
        </div>

        {/* Active Contracts */}
        <div className="customer-kpi-card" onClick={() => setActiveTab('commercial')}>
          <div className="customer-kpi-icon icon-green">
            <ScrollText size={19} />
          </div>
          <div className="customer-kpi-content">
            <span className="kpi-label">Active Contracts</span>
            <strong className="kpi-value">{activeContractsCount}</strong>
            <small className="kpi-sublabel">
              {activeContractValue ? formatCurrency(activeContractValue, currency) : 'No contract value available'}
            </small>
          </div>
        </div>

        {/* Shipments */}
        <div className="customer-kpi-card" onClick={() => setActiveTab('operational')}>
          <div className="customer-kpi-icon icon-orange">
            <Ship size={19} />
          </div>
          <div className="customer-kpi-content">
            <span className="kpi-label">Shipments</span>
            <strong className="kpi-value">{shipments.length}</strong>
            <small className="kpi-sublabel">{activeShipmentsCount} active movements</small>
          </div>
        </div>

        {/* Accepted Value */}
        <div className="customer-kpi-card" onClick={() => setActiveTab('commercial')}>
          <div className="customer-kpi-icon icon-teal">
            <TrendingUp size={19} />
          </div>
          <div className="customer-kpi-content">
            <span className="kpi-label">Accepted Value</span>
            <strong className="kpi-value">{formatCurrency(acceptedValue, currency)}</strong>
            <small className="kpi-sublabel">Accepted commercial business</small>
          </div>
        </div>

        {/* Credit & Terms */}
        <div className="customer-kpi-card" onClick={() => setActiveTab('financial')}>
          <div className="customer-kpi-icon icon-red">
            <CreditCard size={19} />
          </div>
          <div className="customer-kpi-content">
            <span className="kpi-label">Credit & Terms</span>
            <strong className="kpi-value">{financialProfile.payment_terms || customer.payment_terms || 'NET30'}</strong>
            <small className="kpi-sublabel">Limit: {formatCurrency(financialProfile.credit_limit ?? customer.credit_limit ?? 100000, currency)}</small>
          </div>
        </div>

      </section>

      {/* ── TABS NAVIGATION BAR ── */}
      <div className="customer-tabs-shell">
        <div className="customer-tabs-scroll">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              className={`customer-tab ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* ========================================================= */}
      {/* 1. OVERVIEW TAB */}
      {/* ========================================================= */}
      {activeTab === 'overview' && (
        <section className="customer-overview-layout">
          
          {/* Main 3-Column Grid */}
          <div className="customer-overview-main">

            {/* COMPANY INFORMATION CARD */}
            <div className="customer-panel">
              <div className="customer-panel-header">
                <div className="customer-panel-title">
                  <Building2 size={17} className="text-blue-600" />
                  Company Information
                </div>
                <button
                  type="button"
                  className="customer-icon-button"
                  onClick={() => setIsEditModalOpen(true)}
                  title="Edit Company Info"
                >
                  <Edit2 size={14} />
                </button>
              </div>

              <div className="customer-info-list">
                <div className="customer-info-row">
                  <span className="info-key">Legal Name</span>
                  <strong className="info-val">{customer.name}</strong>
                </div>

                {customer.trading_name && (
                  <div className="customer-info-row">
                    <span className="info-key">Trading Name</span>
                    <strong className="info-val">{customer.trading_name}</strong>
                  </div>
                )}

                <div className="customer-info-row">
                  <span className="info-key">Customer Type</span>
                  <strong className="info-val">{customer.customer_type || 'SHIPPER'}</strong>
                </div>

                <div className="customer-info-row">
                  <span className="info-key">Industry</span>
                  <strong className="info-val">{customer.industry || '—'}</strong>
                </div>

                <div className="customer-info-row">
                  <span className="info-key">Tax ID / GSTIN</span>
                  <strong className="info-val font-mono">{customer.tax_id || '—'}</strong>
                </div>

                <div className="customer-info-row">
                  <span className="info-key">Website</span>
                  {customer.website ? (
                    <a
                      href={customer.website.startsWith('http') ? customer.website : `https://${customer.website}`}
                      target="_blank"
                      rel="noreferrer"
                      className="info-link"
                    >
                      <Globe size={13} />
                      {customer.website.replace(/^https?:\/\//, '')}
                    </a>
                  ) : (
                    <span className="info-val text-slate-400">—</span>
                  )}
                </div>
              </div>
            </div>

            {/* KEY CONTACTS CARD */}
            <div className="customer-panel">
              <div className="customer-panel-header">
                <div className="customer-panel-title">
                  <Users size={17} className="text-indigo-600" />
                  Key Contacts
                </div>
                <button
                  type="button"
                  className="customer-link-button"
                  onClick={() => setActiveTab('contacts')}
                >
                  View all
                </button>
              </div>

              <div className="customer-contacts-list">
                {contacts.length === 0 ? (
                  <div className="customer-panel-empty-box">
                    <Users size={28} className="text-slate-300 mb-2" />
                    <strong>No key contacts recorded</strong>
                    <span>Add commercial or operational contacts.</span>
                  </div>
                ) : (
                  contacts.slice(0, 3).map((contact, index) => {
                    const fullName =
                      [contact.first_name, contact.last_name].filter(Boolean).join(' ') ||
                      contact.name ||
                      'Unnamed Contact';

                    return (
                      <div className="customer-contact-item" key={contact.id || index}>
                        <div className="customer-contact-avatar">
                          {getInitials(fullName)}
                        </div>

                        <div className="customer-contact-main">
                          <div className="customer-contact-name-row">
                            <strong>{fullName}</strong>
                            {contact.is_primary && (
                              <span className="contact-role-badge primary">Primary</span>
                            )}
                            {contact.contact_role && (
                              <span className="contact-role-badge">{contact.contact_role.replace(/_/g, ' ')}</span>
                            )}
                          </div>

                          <span className="contact-title">{contact.job_title || contact.department || 'Customer Contact'}</span>

                          <div className="contact-meta-links">
                            {contact.email && (
                              <a href={`mailto:${contact.email}`} className="contact-meta-item">
                                <Mail size={12} />
                                {contact.email}
                              </a>
                            )}
                            {(contact.phone || contact.mobile) && (
                              <a href={`tel:${contact.phone || contact.mobile}`} className="contact-meta-item">
                                <Phone size={12} />
                                {contact.phone || contact.mobile}
                              </a>
                            )}
                          </div>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </div>

            {/* ACTIVE CONTRACTS CARD */}
            <div className="customer-panel">
              <div className="customer-panel-header">
                <div className="customer-panel-title">
                  <ScrollText size={17} className="text-emerald-600" />
                  Active Contracts
                </div>
                <button
                  type="button"
                  className="customer-link-button"
                  onClick={() => setActiveTab('commercial')}
                >
                  View all
                </button>
              </div>

              <div className="customer-contract-list">
                {contracts.length === 0 ? (
                  <div className="customer-panel-empty-box">
                    <ScrollText size={28} className="text-slate-300 mb-2" />
                    <strong>No active contracts available</strong>
                    <span>Contracts created in Rate Management will appear here.</span>
                  </div>
                ) : (
                  contracts.slice(0, 3).map((contract, index) => (
                    <div className="customer-contract-item" key={contract.id || index}>
                      <div>
                        <strong>{contract.contract_number || contract.reference || contract.code || 'Contract'}</strong>
                        <span>{contract.name || contract.contract_name || contract.type || 'Commercial Agreement'}</span>
                      </div>

                      <div className="customer-contract-meta">
                        <span className={`contract-status ${normalizeStatus(contract.status || 'active')}`}>
                          {contract.status || 'Active'}
                        </span>
                        {(contract.value || contract.contract_value) && (
                          <strong>{formatCurrency(contract.value || contract.contract_value, contract.currency || currency)}</strong>
                        )}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* COMMERCIAL SNAPSHOT CARD */}
            <div className="customer-panel customer-commercial-panel">
              <div className="customer-panel-header">
                <div className="customer-panel-title">
                  <TrendingUp size={17} className="text-teal-600" />
                  Commercial Snapshot
                </div>
                <button
                  type="button"
                  className="customer-link-button"
                  onClick={() => setActiveTab('commercial')}
                >
                  View details
                </button>
              </div>

              <div className="customer-commercial-summary-grid">
                <div className="commercial-summary-box">
                  <span className="summary-label">Quoted Pipeline</span>
                  <strong className="summary-val">{formatCurrency(totalQuotedValue, currency)}</strong>
                </div>

                <div className="commercial-summary-box">
                  <span className="summary-label">Accepted Business</span>
                  <strong className="summary-val text-emerald-600">{formatCurrency(acceptedValue, currency)}</strong>
                </div>

                <div className="commercial-summary-box">
                  <span className="summary-label">Contract Value</span>
                  <strong className="summary-val">{formatCurrency(activeContractValue, currency)}</strong>
                </div>
              </div>
            </div>

            {/* RECENT ACTIVITY CARD */}
            <div className="customer-panel">
              <div className="customer-panel-header">
                <div className="customer-panel-title">
                  <Clock size={17} className="text-orange-600" />
                  Recent Activity
                </div>
                <button
                  type="button"
                  className="customer-link-button"
                  onClick={() => setActiveTab('timeline')}
                >
                  View all
                </button>
              </div>

              <div className="customer-recent-list">
                {timeline.length === 0 ? (
                  <div className="customer-panel-empty-box">
                    <Clock size={28} className="text-slate-300 mb-2" />
                    <strong>No recent customer activity</strong>
                    <span>Account creations, quotes, and events will track here.</span>
                  </div>
                ) : (
                  timeline.slice(0, 4).map((event, index) => (
                    <div className="customer-recent-item" key={event.id || index}>
                      <span className="recent-dot" />
                      <div className="recent-main">
                        <strong>{event.title || event.event_type || 'Customer Activity'}</strong>
                        <span>{event.description || 'Customer activity recorded'}</span>
                      </div>
                      <time className="recent-time">{formatDate(event.created_at || event.event_at || event.timestamp)}</time>
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* FINANCIAL STANDING CARD */}
            <div className="customer-panel">
              <div className="customer-panel-header">
                <div className="customer-panel-title">
                  <CreditCard size={17} className="text-rose-600" />
                  Financial Standing
                </div>
                <button
                  type="button"
                  className="customer-link-button"
                  onClick={() => setActiveTab('financial')}
                >
                  View details
                </button>
              </div>

              <div className="customer-financial-summary-grid">
                <div className="financial-summary-box">
                  <span className="summary-label">Payment Terms</span>
                  <strong className="summary-val">{financialProfile.payment_terms || customer.payment_terms || 'NET60'}</strong>
                </div>

                <div className="financial-summary-box">
                  <span className="summary-label">Credit Limit</span>
                  <strong className="summary-val">{formatCurrency(financialProfile.credit_limit ?? customer.credit_limit ?? 100000, currency)}</strong>
                </div>

                <div className="financial-summary-box">
                  <span className="summary-label">Credit Status</span>
                  <div className="mt-1">{renderCreditStatus()}</div>
                </div>
              </div>
            </div>

          </div>

          {/* Right Activity Timeline Widget */}
          <aside className="customer-activity-sidebar">
            <div className="customer-timeline-panel">
              <div className="customer-panel-header">
                <div className="customer-panel-title">
                  <Clock size={17} className="text-slate-700" />
                  Activity Timeline
                </div>
                <button
                  type="button"
                  className="customer-icon-button"
                  onClick={() => setActiveTab('timeline')}
                  title="Full Timeline"
                >
                  <MoreHorizontal size={18} />
                </button>
              </div>

              <div className="customer-timeline-feed">
                {timeline.length === 0 ? (
                  <div className="customer-panel-empty-box">
                    <Clock size={28} className="text-slate-300 mb-2" />
                    <strong>No activity history available</strong>
                    <span>Historical events will appear in this feed.</span>
                  </div>
                ) : (
                  timeline.slice(0, 6).map((event, index) => (
                    <div className="customer-timeline-event" key={event.id || index}>
                      <div className="timeline-marker">
                        <span />
                      </div>
                      <div className="timeline-event-body">
                        <time>{formatDate(event.created_at || event.event_at || event.timestamp)}</time>
                        <strong>{event.title || event.event_type || 'Customer Activity'}</strong>
                        <p>{event.description || 'Customer account activity recorded.'}</p>
                      </div>
                    </div>
                  ))
                )}
              </div>

              <button
                type="button"
                className="customer-full-timeline-button"
                onClick={() => setActiveTab('timeline')}
              >
                View full timeline
                <ArrowRight size={15} />
              </button>
            </div>
          </aside>

        </section>
      )}

      {/* ========================================================= */}
      {/* 2. CONTACTS TAB */}
      {/* ========================================================= */}
      {activeTab === 'contacts' && (
        <section className="customer-tab-panel">
          <div className="customer-tab-header">
            <div>
              <h2>Customer Contacts</h2>
              <p>Commercial, operational and financial contacts associated with this account.</p>
            </div>
            <button
              type="button"
              className="btn-primary-action"
              onClick={() => setIsEditModalOpen(true)}
            >
              <UserPlus size={15} /> Add Contact
            </button>
          </div>

          <div className="customer-contact-grid">
            {contacts.length === 0 ? (
              <div className="customer-empty-state-large">
                <Users size={36} className="text-blue-500 mb-2" />
                <h3>No contacts added yet</h3>
                <p>Add key personnel, billing managers, and operational contacts to streamline communication.</p>
                <button
                  type="button"
                  className="btn-primary-action mt-3"
                  onClick={() => setIsEditModalOpen(true)}
                >
                  <Plus size={15} /> Add First Contact
                </button>
              </div>
            ) : (
              contacts.map((contact, index) => {
                const fullName =
                  [contact.first_name, contact.last_name].filter(Boolean).join(' ') ||
                  contact.name ||
                  'Unnamed Contact';

                return (
                  <div className="customer-contact-card" key={contact.id || index}>
                    <div className="contact-card-top">
                      <div className="customer-contact-avatar large">
                        {getInitials(fullName)}
                      </div>
                      <div className="contact-card-header-info">
                        <h3>{fullName}</h3>
                        <span>{contact.job_title || contact.department || 'Customer Contact'}</span>
                      </div>
                    </div>

                    <div className="contact-card-details">
                      {contact.email && (
                        <a href={`mailto:${contact.email}`} className="contact-detail-row">
                          <Mail size={14} className="text-slate-400" />
                          <span>{contact.email}</span>
                        </a>
                      )}
                      {(contact.phone || contact.mobile) && (
                        <a href={`tel:${contact.phone || contact.mobile}`} className="contact-detail-row">
                          <Phone size={14} className="text-slate-400" />
                          <span>{contact.phone || contact.mobile}</span>
                        </a>
                      )}
                    </div>

                    <div className="contact-card-footer">
                      {contact.contact_role && (
                        <span className="contact-role-badge">
                          {contact.contact_role.replace(/_/g, ' ')}
                        </span>
                      )}
                      {contact.is_primary && (
                        <span className="contact-role-badge primary">
                          Primary
                        </span>
                      )}
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </section>
      )}

      {/* ========================================================= */}
      {/* 3. ADDRESSES TAB */}
      {/* ========================================================= */}
      {activeTab === 'addresses' && (
        <section className="customer-tab-panel">
          <div className="customer-tab-header">
            <div>
              <h2>Customer Addresses</h2>
              <p>Registered, billing and operational locations associated with this customer account.</p>
            </div>
            <button
              type="button"
              className="btn-primary-action"
              onClick={() => setIsEditModalOpen(true)}
            >
              <Plus size={15} /> Add Address
            </button>
          </div>

          {addresses.length === 0 ? (
            <div className="customer-empty-state-large">
              <MapPin size={36} className="text-blue-500 mb-2" />
              <h3>No customer addresses yet</h3>
              <p>No registered, billing or operational locations have been added to this customer account.</p>
              <button
                type="button"
                className="btn-primary-action mt-3"
                onClick={() => setIsEditModalOpen(true)}
              >
                <Plus size={15} /> Add First Address
              </button>
            </div>
          ) : (
            <div className="customer-address-grid">
              {addresses.map((address, index) => {
                const addressType = address.address_type || address.type || 'Registered Address';
                return (
                  <div className="customer-address-card" key={address.id || index}>
                    <div className="address-card-header">
                      <span className="address-type-pill">{addressType.replace(/_/g, ' ')}</span>
                      {address.is_primary && <span className="address-primary-pill">Primary</span>}
                    </div>

                    <div className="address-card-body">
                      <MapPin size={16} className="address-icon text-slate-400" />
                      <div>
                        {address.street_line1 && <div className="address-line">{address.street_line1}</div>}
                        {address.street_line2 && <div className="address-line">{address.street_line2}</div>}
                        <div className="address-city-state">
                          {[address.city, address.state, address.postal_code].filter(Boolean).join(', ')}
                        </div>
                        {address.country && <div className="address-country">{address.country}</div>}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>
      )}

      {/* ========================================================= */}
      {/* 4. COMMERCIAL TAB */}
      {/* ========================================================= */}
      {activeTab === 'commercial' && (
        <section className="customer-tab-panel customer-commercial-panel">
          <div className="customer-tab-header">
            <div>
              <h2>Commercial Overview</h2>
              <p>Quoted pipeline, accepted business and active contract value for this customer account.</p>
            </div>
            <button
              type="button"
              className="btn-primary-action"
              onClick={() => navigate('/dashboard/rfqs')}
            >
              <Plus size={15} /> Create RFQ
            </button>
          </div>

          <div className="customer-commercial-kpis">
            <div className="customer-commercial-kpi">
              <div className="commercial-kpi-icon rfq">
                <FileText size={18} />
              </div>
              <div className="commercial-kpi-content">
                <span>RFQs</span>
                <strong>{customer.rfq_count || rfqs.length || 0}</strong>
                <small>Customer opportunities</small>
              </div>
            </div>

            <div className="customer-commercial-kpi">
              <div className="commercial-kpi-icon quotation">
                <FileCheck size={18} />
              </div>
              <div className="commercial-kpi-content">
                <span>Quoted Pipeline</span>
                <strong>{formatCurrency(commercialMetrics?.total_quotation_value || totalQuotedValue, currency)}</strong>
                <small>{quotations.length} total quotations</small>
              </div>
            </div>

            <div className="customer-commercial-kpi">
              <div className="commercial-kpi-icon accepted">
                <TrendingUp size={18} />
              </div>
              <div className="commercial-kpi-content">
                <span>Accepted Value</span>
                <strong className="text-emerald-600">{formatCurrency(commercialMetrics?.accepted_quotation_value || acceptedValue, currency)}</strong>
                <small>Accepted commercial business</small>
              </div>
            </div>

            <div className="customer-commercial-kpi">
              <div className="commercial-kpi-icon contract">
                <BriefcaseBusiness size={18} />
              </div>
              <div className="commercial-kpi-content">
                <span>Active Contract Value</span>
                <strong>{formatCurrency(commercialMetrics?.active_contract_value || activeContractValue, currency)}</strong>
                <small>{activeContractsCount} active contracts</small>
              </div>
            </div>
          </div>

          {quotations && quotations.length > 0 ? (
            <div className="customer-commercial-content mt-4">
              <div className="customer-commercial-section-header">
                <div>
                  <h3>Quotation Pipeline</h3>
                  <p>Recent commercial quotations associated with this customer.</p>
                </div>
                <span className="customer-commercial-record-count">{quotations.length} Quotations</span>
              </div>

              <div className="customer-commercial-list">
                {quotations.map((quotation, index) => (
                  <div className="customer-commercial-row" key={quotation.id || index}>
                    <div className="commercial-row-icon">
                      <FileText size={17} />
                    </div>
                    <div className="commercial-row-main">
                      <strong>{quotation.quotation_number || quotation.reference_number || `Quotation ${index + 1}`}</strong>
                      <span>{quotation.status ? quotation.status.replace(/_/g, ' ') : 'Commercial Quotation'}</span>
                    </div>
                    <div className="commercial-row-value">
                      <strong>{formatCurrency(quotation.total_amount || quotation.total_value || quotation.amount || 0, currency)}</strong>
                      <span>Quoted Value</span>
                    </div>
                    <span className={`commercial-status-badge ${String(quotation.status || 'DRAFT').toLowerCase().replace(/_/g, '-')}`}>
                      {quotation.status ? quotation.status.replace(/_/g, ' ') : 'Draft'}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="customer-commercial-empty">
              <div className="customer-commercial-empty-icon">
                <BriefcaseBusiness size={32} />
              </div>
              <h3>No commercial pipeline yet</h3>
              <p>This customer does not currently have any active quotations recorded. Issue an RFQ or generate a quotation to start commercial engagement.</p>
              
              <div className="customer-commercial-empty-stages">
                <div>
                  <span className="commercial-stage-dot rfq-dot" />
                  <span>RFQ received</span>
                </div>
                <span className="commercial-stage-arrow">→</span>
                <div>
                  <span className="commercial-stage-dot quote-dot" />
                  <span>Quotation issued</span>
                </div>
                <span className="commercial-stage-arrow">→</span>
                <div>
                  <span className="commercial-stage-dot accepted-dot" />
                  <span>Business accepted</span>
                </div>
              </div>
            </div>
          )}
        </section>
      )}

      {/* ========================================================= */}
      {/* 5. OPERATIONAL TAB */}
      {/* ========================================================= */}
      {activeTab === 'operational' && (
        <section className="customer-tab-panel customer-operational-panel">
          <div className="customer-tab-header">
            <div>
              <h2>Operational Activity</h2>
              <p>Monitor bookings, shipments and active logistics operations associated with this customer.</p>
            </div>
            <button
              type="button"
              className="btn-primary-action"
              onClick={() => navigate('/dashboard/bookings')}
            >
              <Plus size={15} /> Create Booking
            </button>
          </div>

          <div className="customer-operational-stats">
            <div className="customer-operational-stat">
              <div className="operational-stat-icon booking">
                <Package size={18} />
              </div>
              <div className="operational-stat-content">
                <span>Bookings</span>
                <strong>{bookings?.length || 0}</strong>
                <small>{bookings?.length === 1 ? 'Booking recorded' : 'Bookings recorded'}</small>
              </div>
            </div>

            <div className="customer-operational-stat">
              <div className="operational-stat-icon shipment">
                <Truck size={18} />
              </div>
              <div className="operational-stat-content">
                <span>Shipments</span>
                <strong>{shipments?.length || 0}</strong>
                <small>{shipments?.length === 1 ? 'Shipment recorded' : 'Shipments recorded'}</small>
              </div>
            </div>

            <div className="customer-operational-stat">
              <div className="operational-stat-icon active">
                <Activity size={18} />
              </div>
              <div className="operational-stat-content">
                <span>Active Operations</span>
                <strong>{(shipments?.length || 0) + (bookings?.length || 0)}</strong>
                <small>Current operational activity</small>
              </div>
            </div>
          </div>

          {(!bookings || bookings.length === 0) && (!shipments || shipments.length === 0) ? (
            <div className="customer-operational-empty">
              <div className="operational-empty-icon">
                <PackageOpen size={32} />
              </div>
              <h3>No operational activity yet</h3>
              <p>This customer does not currently have any active bookings or shipments associated with their account.</p>
              <div className="operational-empty-actions">
                <button
                  type="button"
                  className="operational-primary-action"
                  onClick={() => navigate('/dashboard/bookings')}
                >
                  <Package size={16} />
                  Create Booking
                </button>
                <button
                  type="button"
                  className="operational-secondary-action"
                  onClick={() => navigate('/dashboard/shipments')}
                >
                  <Truck size={16} />
                  View Shipments
                </button>
              </div>
            </div>
          ) : (
            <div className="customer-operational-content">
              {bookings?.length > 0 && (
                <div className="operational-data-card">
                  <div className="operational-data-header">
                    <div>
                      <h3>Recent Bookings</h3>
                      <span>Latest bookings associated with this customer</span>
                    </div>
                    <button
                      type="button"
                      className="operational-view-all"
                      onClick={() => navigate('/dashboard/bookings')}
                    >
                      View All
                    </button>
                  </div>

                  <div className="operational-list">
                    {bookings.slice(0, 5).map((booking, index) => (
                      <div className="operational-list-item" key={booking.id || index}>
                        <div className="operational-list-icon booking">
                          <Package size={16} />
                        </div>
                        <div className="operational-list-main">
                          <strong>{booking.booking_number || booking.reference || `Booking ${index + 1}`}</strong>
                          <span>{booking.status || 'Booking created'}</span>
                        </div>
                        <div className="operational-list-status">{booking.status || 'Active'}</div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {shipments?.length > 0 && (
                <div className="operational-data-card mt-4">
                  <div className="operational-data-header">
                    <div>
                      <h3>Recent Shipments</h3>
                      <span>Latest shipment movements for this customer</span>
                    </div>
                    <button
                      type="button"
                      className="operational-view-all"
                      onClick={() => navigate('/dashboard/shipments')}
                    >
                      View All
                    </button>
                  </div>

                  <div className="operational-list">
                    {shipments.slice(0, 5).map((shipment, index) => (
                      <div className="operational-list-item" key={shipment.id || index}>
                        <div className="operational-list-icon shipment">
                          <Truck size={16} />
                        </div>
                        <div className="operational-list-main">
                          <strong>{shipment.shipment_number || shipment.reference || `Shipment ${index + 1}`}</strong>
                          <span>{shipment.status || 'Shipment created'}</span>
                        </div>
                        <div className="operational-list-status">{shipment.status || 'Active'}</div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </section>
      )}

      {/* ========================================================= */}
      {/* 6. FINANCIAL TAB */}
      {/* ========================================================= */}
      {activeTab === 'financial' && (
        <section className="customer-tab-panel">
          <div className="customer-tab-header">
            <div>
              <h2>Financial & Credit Profile</h2>
              <p>Commercial terms, credit standing, and pipeline valuation for this customer account.</p>
            </div>
            <button
              type="button"
              className="btn-secondary-action"
              onClick={() => setIsEditModalOpen(true)}
            >
              <Edit2 size={14} /> Update Terms
            </button>
          </div>

          {/* 4-Card Financial Metric Grid */}
          <div className="customer-financial-grid">
            <div className="financial-detail-card">
              <div className="financial-card-header">
                <Globe size={18} className="text-blue-500" />
                <span className="financial-card-label">Default Currency</span>
              </div>
              <strong className="financial-card-value font-mono">{currency}</strong>
              <small className="financial-card-sub">Invoicing and payment currency</small>
            </div>

            <div className="financial-detail-card">
              <div className="financial-card-header">
                <Clock size={18} className="text-indigo-500" />
                <span className="financial-card-label">Payment Terms</span>
              </div>
              <strong className="financial-card-value">{financialProfile.payment_terms || customer.payment_terms || 'NET60'}</strong>
              <small className="financial-card-sub">Invoice settlement window</small>
            </div>

            <div className="financial-detail-card">
              <div className="financial-card-header">
                <CreditCard size={18} className="text-emerald-500" />
                <span className="financial-card-label">Credit Limit</span>
              </div>
              <strong className="financial-card-value">
                {formatCurrency(financialProfile.credit_limit ?? customer.credit_limit ?? 100000, currency)}
              </strong>
              <small className="financial-card-sub">Authorized credit exposure</small>
            </div>

            <div className="financial-detail-card">
              <div className="financial-card-header">
                <ShieldCheck size={18} className="text-teal-500" />
                <span className="financial-card-label">Credit Status</span>
              </div>
              <div className="mt-1">{renderCreditStatus()}</div>
              <small className="financial-card-sub">Automated account standing</small>
            </div>
          </div>

          {/* Value Breakdown */}
          <div className="financial-commercial-breakdown">
            <div className="breakdown-card">
              <span className="breakdown-label">Quotation Pipeline</span>
              <strong className="breakdown-value">{formatCurrency(totalQuotedValue, currency)}</strong>
              <small className="breakdown-sub">Active open quotations</small>
            </div>

            <div className="breakdown-card">
              <span className="breakdown-label">Accepted Commercial Value</span>
              <strong className="breakdown-value text-emerald-600">{formatCurrency(acceptedValue, currency)}</strong>
              <small className="breakdown-sub">Confirmed customer business</small>
            </div>

            <div className="breakdown-card">
              <span className="breakdown-label">Active Contract Value</span>
              <strong className="breakdown-value">{formatCurrency(activeContractValue, currency)}</strong>
              <small className="breakdown-sub">Long-term rate commitments</small>
            </div>
          </div>

          {/* Commercial Notes Box */}
          <div className="customer-commercial-notes-card">
            <div className="notes-card-header">
              <FileText size={16} className="text-amber-600" />
              <strong>Commercial Notes</strong>
            </div>
            <p className="notes-card-text">
              {financialProfile.commercial_notes || customer.notes || 'VIP Enterprise Shipper account with verified corporate billing profile and standard NET terms.'}
            </p>
          </div>
        </section>
      )}

      {/* ========================================================= */}
      {/* 7. DOCUMENTS TAB */}
      {/* ========================================================= */}
      {activeTab === 'documents' && (
        <section className="customer-tab-panel">
          <div className="customer-tab-header">
            <div>
              <h2>Customer Documents ({customerDocs.length})</h2>
              <p>Bills of lading, commercial invoices, carrier contracts, and compliance records linked to this customer.</p>
            </div>
            <button
              type="button"
              className="btn-primary-action"
              onClick={() => setIsDocUploadOpen(true)}
            >
              <Plus size={15} /> Upload Customer Document
            </button>
          </div>

          {customerDocs.length === 0 ? (
            <div className="customer-empty-state-large">
              <Folder size={36} className="text-blue-500 mb-2" />
              <h3>No customer documents uploaded yet</h3>
              <p>Upload contracts, bills of lading, customs declarations, or commercial invoices for {customer.name}.</p>
              <button
                type="button"
                className="btn-primary-action mt-3"
                onClick={() => setIsDocUploadOpen(true)}
              >
                <Plus size={15} /> Upload First Document
              </button>
            </div>
          ) : (
            <div className="customer-documents-table-wrapper">
              <table className="documents-table">
                <thead>
                  <tr>
                    <th>Document</th>
                    <th>Type</th>
                    <th>Linked Context</th>
                    <th>Status</th>
                    <th>Uploaded</th>
                    <th style={{ textAlign: 'right' }}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {customerDocs.map((d) => (
                    <tr key={d.id} onClick={() => setSelectedDocForDetails(d)}>
                      <td>
                        <div className="doc-name-cell">
                          <FileText size={16} className="doc-icon text-blue-600" />
                          <div>
                            <strong className="doc-filename">{d.original_file_name || d.file_name || `DOC-${d.id}`}</strong>
                            <span className="doc-meta">{d.file_type || 'PDF'} {d.file_size ? `· ${(d.file_size / 1024).toFixed(1)} KB` : ''}</span>
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className="doc-type-badge">{d.doc_type || 'GENERAL'}</span>
                      </td>
                      <td>
                        <div className="doc-context-text">
                          {d.shipment_ref ? `🚢 Shipment: ${d.shipment_ref}` : d.booking_ref ? `📋 Booking: ${d.booking_ref}` : 'Customer Document'}
                        </div>
                      </td>
                      <td>
                        <span className={`compliance-pill ${(d.status || 'VERIFIED').toLowerCase()}`}>
                          <CheckCircle2 size={12} />
                          {d.status || 'VERIFIED'}
                        </span>
                      </td>
                      <td>{d.created_at ? formatDate(d.created_at) : '—'}</td>
                      <td style={{ textAlign: 'right' }} onClick={(e) => e.stopPropagation()}>
                        <div className="doc-action-group">
                          <button
                            type="button"
                            className="doc-action-btn"
                            title="View Details"
                            onClick={() => setSelectedDocForDetails(d)}
                          >
                            <FileText size={14} />
                          </button>
                          <a
                            href={d.file_path && d.file_path.startsWith('http') ? d.file_path : `http://localhost:8080/uploads/${d.s3_key || d.file_name}`}
                            target="_blank"
                            rel="noreferrer"
                            className="doc-action-btn"
                            title="Download File"
                          >
                            <ExternalLink size={14} />
                          </a>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          <DocumentUploadModal
            isOpen={isDocUploadOpen}
            onClose={() => setIsDocUploadOpen(false)}
            onUploadSuccess={() => fetchCustomerDocuments()}
            initialCustomerId={id}
          />

          <DocumentDetailsModal
            isOpen={!!selectedDocForDetails}
            onClose={() => setSelectedDocForDetails(null)}
            document={selectedDocForDetails}
            onDeleteSuccess={() => fetchCustomerDocuments()}
          />
        </section>
      )}

      {/* ========================================================= */}
      {/* 8. TIMELINE TAB */}
      {/* ========================================================= */}
      {activeTab === 'timeline' && (
        <section className="customer-tab-panel">
          <div className="customer-tab-header">
            <div>
              <h2>Customer Activity Timeline</h2>
              <p>Complete commercial, relationship and operational audit history.</p>
            </div>
          </div>

          <div className="full-timeline-wrapper">
            {timeline.length === 0 ? (
              <div className="customer-empty-state-large">
                <Clock size={36} className="text-slate-400 mb-2" />
                <h3>No activity timeline events</h3>
                <p>New orders, contracts, quotations, and account modifications will appear here chronologically.</p>
              </div>
            ) : (
              <div className="full-timeline-tree">
                {timeline.map((event, index) => (
                  <div className="full-timeline-event-card" key={event.id || index}>
                    <div className="timeline-node-column">
                      <div className="timeline-node-circle" />
                      {index < timeline.length - 1 && <div className="timeline-node-line" />}
                    </div>

                    <div className="timeline-event-content-box">
                      <div className="timeline-event-top">
                        <span className="timeline-event-tag">{event.event_type || 'ACTIVITY'}</span>
                        <time className="timeline-event-date">{formatDate(event.created_at || event.event_at || event.timestamp)}</time>
                      </div>
                      <h3 className="timeline-event-heading">{event.title || event.event_type || 'Customer Account Activity'}</h3>
                      <p className="timeline-event-desc">{event.description || 'Customer account milestone recorded.'}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </section>
      )}

      {/* ========================================================= */}
      {/* 9. INTELLIGENCE TAB */}
      {/* ========================================================= */}
      {activeTab === 'intelligence' && (
        <section className="customer-tab-panel">
          <div className="customer-tab-header">
            <div>
              <h2>Customer Intelligence & Risk</h2>
              <p>Automated account health assessment, detected risks, and commercial expansion opportunities.</p>
            </div>

            <button
              type="button"
              className="btn-secondary-action"
              disabled={refreshingIntel}
              onClick={handleRefreshIntel}
            >
              <RefreshCw size={14} className={refreshingIntel ? 'animate-spin' : ''} />
              {refreshingIntel ? 'Evaluating...' : 'Refresh Intelligence'}
            </button>
          </div>

          {/* 3 Intelligence Summary Cards */}
          <div className="customer-intelligence-grid">
            <div className="customer-intelligence-card">
              <div className="intelligence-card-header">
                <TrendingUp size={18} className="text-blue-600" />
                <span>Account Health</span>
              </div>
              <div className="health-score-display">
                <strong>{healthScore ?? 50}</strong>
                <span>/ 100</span>
              </div>
              <div className="mt-2">{renderHealthStatus()}</div>
            </div>

            <div className="customer-intelligence-card">
              <div className="intelligence-card-header">
                <ShieldAlert size={18} className="text-amber-600" />
                <span>Open Risks</span>
              </div>
              <strong className="intelligence-count text-amber-600">{openRisks.length}</strong>
              <span className="intelligence-card-sub">Active account risks requiring attention</span>
            </div>

            <div className="customer-intelligence-card">
              <div className="intelligence-card-header">
                <Zap size={18} className="text-indigo-600" />
                <span>Opportunities</span>
              </div>
              <strong className="intelligence-count text-indigo-600">{opportunities.length}</strong>
              <span className="intelligence-card-sub">Detected commercial expansion opportunities</span>
            </div>
          </div>

          {/* Risks Section */}
          <div className="mt-4">
            <div className="intelligence-section-title">
              <ShieldAlert size={16} className="text-amber-600" />
              <h3>Detected Account Risks</h3>
            </div>

            {openRisks.length > 0 ? (
              <div className="customer-risk-list">
                {openRisks.map((risk, index) => (
                  <div className="customer-risk-card" key={risk.id || index}>
                    <div className="risk-icon-box">
                      <AlertTriangle size={18} className="text-amber-600" />
                    </div>
                    <div className="risk-body">
                      <strong className="risk-title">{risk.title || 'Customer Risk'}</strong>
                      <p className="risk-desc">{risk.description || risk.reason || 'Risk requires commercial review.'}</p>
                    </div>
                    <span className={`risk-severity-badge severity-${(risk.severity || 'ATTENTION').toLowerCase()}`}>
                      {risk.severity || 'ATTENTION'}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="customer-empty-state-green">
                <CheckCircle2 size={24} className="text-emerald-600" />
                <div>
                  <strong>No open risks detected</strong>
                  <span>This customer account is in good standing with zero active operational or financial alerts.</span>
                </div>
              </div>
            )}
          </div>

          {/* Opportunities Section */}
          <div className="mt-4">
            <div className="intelligence-section-title">
              <Zap size={16} className="text-indigo-600" />
              <h3>Commercial Opportunities</h3>
            </div>

            {opportunities.length > 0 ? (
              <div className="customer-opp-list">
                {opportunities.map((opp, index) => (
                  <div className="customer-opp-card" key={opp.id || index}>
                    <div className="opp-icon-box">
                      <Zap size={18} className="text-indigo-600" />
                    </div>
                    <div className="opp-body">
                      <strong className="opp-title">{opp.title || 'Commercial Opportunity'}</strong>
                      <p className="opp-desc">{opp.description || opp.reason || 'Expansion opportunity detected.'}</p>
                    </div>
                    <button
                      type="button"
                      className="btn-secondary-action"
                      onClick={() => navigate('/dashboard/rfqs')}
                    >
                      Explore RFQ
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="customer-empty-state-neutral">
                <HelpCircle size={20} className="text-slate-400" />
                <span>No new commercial opportunities identified at this time. Increase RFQ and quotation engagement to trigger smart suggestions.</span>
              </div>
            )}
          </div>

        </section>
      )}

      {/* ── EDIT CUSTOMER MODAL ── */}
      <CustomerModal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        onSuccess={fetchCustomer360}
        initialData={customer}
      />

    </div>
  );
}

function ActivityIcon() {
  return <Clock size={17} />;
}