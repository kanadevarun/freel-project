import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Users,
  Briefcase,
  FileText,
  DollarSign,
  AlertTriangle,
  Search,
  Filter,
  Plus,
  Download,
  MoreVertical,
  ChevronLeft,
  ChevronRight,
  Building2,
  ExternalLink,
  Edit2,
  Archive,
  RefreshCw,
  ShieldAlert,
  Zap,
  TrendingUp,
  TrendingDown,
  Activity,
  ArrowRight,
  CheckCircle2,
  AlertCircle,
  HelpCircle,
  MessageSquare,
} from 'lucide-react';
import customerService from '../../../services/customerService';
import CustomerModal from './CustomerModal';
import CustomerFollowupDrawer from '../../../components/autonomy/CustomerFollowupDrawer';
import ModuleHeroEmptyState from '../../../components/dashboard/ModuleHeroEmptyState';
import './CustomersPage.css';

const COUNTRY_FLAGS = {
  India: '🇮🇳',
  UAE: '🇦🇪',
  'United Arab Emirates': '🇦🇪',
  Sweden: '🇸🇪',
  Singapore: '🇸🇬',
  Greece: '🇬🇷',
  USA: '🇺🇸',
  'United States': '🇺🇸',
  Germany: '🇩🇪',
  UK: '🇬🇧',
  'United Kingdom': '🇬🇧',
  China: '🇨🇳',
  Japan: '🇯🇵',
  Netherlands: '🇳🇱',
};

export default function CustomersPage() {
  const navigate = useNavigate();
  const [customers, setCustomers] = useState([]);
  const [kpis, setKPIs] = useState({});
  const [intelSummary, setIntelSummary] = useState({});
  const [attentionItems, setAttentionItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Filters & Pagination State
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [typeFilter, setTypeFilter] = useState('ALL');
  const [countryFilter, setCountryFilter] = useState('ALL');
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);
  const [total, setTotal] = useState(0);

  const availableCountries = useMemo(() => {
    const defaults = ['India', 'UAE', 'Sweden', 'Singapore', 'USA', 'Germany', 'United Kingdom', 'Netherlands'];
    const set = new Set(defaults);
    customers.forEach((c) => {
      if (c.country && c.country.trim()) {
        set.add(c.country.trim());
      }
    });
    return Array.from(set).sort();
  }, [customers]);

  // Attention Panel Active Tab
  const [attentionTab, setAttentionTab] = useState('WARNING');

  // Modals
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState(null);
  const [activeKebabId, setActiveKebabId] = useState(null);
  const [kebabPos, setKebabPos] = useState({ top: 0, right: 0 });

  // Phase 5 Task 5.4: Autonomous Customer Follow-Up Drawer State
  const [isFollowupDrawerOpen, setIsFollowupDrawerOpen] = useState(false);
  const [followupCustomer, setFollowupCustomer] = useState(null);

  useEffect(() => {
    const handleOutsideClick = () => setActiveKebabId(null);
    document.addEventListener('click', handleOutsideClick);
    return () => document.removeEventListener('click', handleOutsideClick);
  }, []);

  useEffect(() => {
    fetchData();
  }, [search, statusFilter, typeFilter, countryFilter, page, limit]);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [listRes, kpiRes, summaryRes, attentionRes] = await Promise.all([
        customerService.getCustomers({
          search,
          status: statusFilter,
          customer_type: typeFilter,
          country: countryFilter,
          page,
          limit,
        }),
        customerService.getCustomerKPIs().catch((kpiErr) => {
          console.warn('Customer KPIs non-critical fetch warning:', kpiErr);
          return {};
        }),
        customerService.getIntelligenceSummary().catch((sumErr) => {
          console.warn('Customer Intelligence summary non-critical fetch warning:', sumErr);
          return {};
        }),
        customerService.getAttentionItems().catch((attErr) => {
          console.warn('Customer Attention items non-critical fetch warning:', attErr);
          return [];
        }),
      ]);

      const listData = Array.isArray(listRes?.customers)
        ? listRes.customers
        : (Array.isArray(listRes) ? listRes : []);
      const totalCount = listRes?.total !== undefined ? listRes.total : listData.length;

      setCustomers(listData);
      setTotal(totalCount);
      setKPIs(kpiRes || {});
      setIntelSummary(summaryRes || {});
      setAttentionItems(Array.isArray(attentionRes) ? attentionRes : []);
    } catch (err) {
      console.error('Failed to load customer directory:', err);
      setError('Unable to load customer directory. Please check network connection.');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateNew = () => {
    setSelectedCustomer(null);
    setIsModalOpen(true);
  };

  const handleEdit = (customer) => {
    setSelectedCustomer(customer);
    setIsModalOpen(true);
    setActiveKebabId(null);
  };

  const handleArchive = async (id) => {
    try {
      await customerService.archiveCustomer(id);
      fetchData();
    } catch (err) {
      console.error('Failed to archive customer:', err);
    } finally {
      setActiveKebabId(null);
    }
  };

  const renderHealthIndicator = (status, score) => {
    const s = status || 'HEALTHY';
    const sc = score || 85;
    switch (s) {
      case 'HEALTHY':
        return <span className="intel-pill health-healthy">🟢 Healthy ({sc})</span>;
      case 'WATCH':
        return <span className="intel-pill health-watch">🟡 Watch ({sc})</span>;
      case 'AT_RISK':
        return <span className="intel-pill health-atrisk">🟠 At Risk ({sc})</span>;
      case 'CRITICAL':
        return <span className="intel-pill health-critical">🔴 Critical ({sc})</span>;
      default:
        return <span className="intel-pill health-healthy">🟢 Healthy ({sc})</span>;
    }
  };

  const renderTrendIndicator = (trend) => {
    const t = trend || 'STABLE';
    switch (t) {
      case 'INCREASING':
        return <span className="trend-tag trend-increasing">↑ Increasing</span>;
      case 'STABLE':
        return <span className="trend-tag trend-stable">→ Stable</span>;
      case 'DECLINING':
        return <span className="trend-tag trend-declining">↓ Declining</span>;
      case 'INACTIVE':
        return <span className="trend-tag trend-inactive">⏸ Inactive</span>;
      default:
        return <span className="trend-tag trend-stable">→ Stable</span>;
    }
  };

  const renderCreditBadge = (status) => {
    const st = status || 'GOOD_STANDING';
    switch (st) {
      case 'GOOD_STANDING':
        return <span className="credit-badge status-good">Good Standing</span>;
      case 'REVIEW_REQUIRED':
        return <span className="credit-badge status-review">Review Required</span>;
      case 'ON_HOLD':
        return <span className="credit-badge status-hold">Credit On Hold</span>;
      case 'NO_CREDIT_LIMIT':
        return <span className="credit-badge status-nolimit">No Credit Limit</span>;
      default:
        return <span className="credit-badge status-good">Good Standing</span>;
    }
  };

  const getAvatarColor = (name) => {
    const colors = [
      'avatar-blue',
      'avatar-purple',
      'avatar-emerald',
      'avatar-amber',
      'avatar-indigo',
      'avatar-rose',
      'avatar-teal',
    ];
    let hash = 0;
    for (let i = 0; i < (name || '').length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    return colors[Math.abs(hash) % colors.length];
  };

  const filteredAttentionItems = attentionItems.filter((item) => {
    if (attentionTab === 'CRITICAL') return item.severity === 'CRITICAL';
    if (attentionTab === 'WARNING') return item.severity === 'WARNING';
    if (attentionTab === 'ATTENTION') return item.severity === 'ATTENTION';
    return item.severity === 'INFO';
  });

  const totalPages = Math.ceil(total / limit) || 1;

  return (
    <div className="cust-dir-page">
      {/* Header Banner */}
      <div className="cust-header-bar">
        <div className="header-titles">
          <h1 className="cust-page-title">Customer Accounts & Directory</h1>
          <p className="cust-page-subtitle">
            Comprehensive directory of verified shippers, commercial contracts, contacts, addresses, and finance terms.
          </p>
        </div>
        <div className="cust-header-actions">
          <button className="btn-secondary-action" onClick={fetchData}>
            <RefreshCw size={15} /> Refresh Directory
          </button>
          <button className="btn-primary-action" onClick={handleCreateNew}>
            <Plus size={16} /> New Customer
          </button>
        </div>
      </div>

      {/* Business KPI Cards Grid */}
      <div className="cust-kpi-grid">
        <div className="cust-kpi-card kpi-card-healthy">
          <div className="kpi-icon-badge bg-blue-badge" style={{ backgroundColor: '#eff6ff', color: '#2563eb' }}>
            <Building2 size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">Total Accounts</span>
            <div className="kpi-value-row">
              <span className="kpi-value">{total || customers.length || 0}</span>
              <span className="kpi-sub-tag text-emerald">Active Customer Directory</span>
            </div>
          </div>
        </div>

        <div className="cust-kpi-card kpi-card-watch">
          <div className="kpi-icon-badge bg-indigo-badge" style={{ backgroundColor: '#eef2ff', color: '#4f46e5' }}>
            <Users size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">Commercial Shippers</span>
            <div className="kpi-value-row">
              <span className="kpi-value">{customers.filter(c => (c.customer_type || 'SHIPPER') === 'SHIPPER').length || customers.length}</span>
              <span className="kpi-sub-tag text-indigo" style={{ color: '#4f46e5' }}>Direct Cargo Owners</span>
            </div>
          </div>
        </div>

        <div className="cust-kpi-card kpi-card-atrisk">
          <div className="kpi-icon-badge bg-green-badge" style={{ backgroundColor: '#ecfdf5', color: '#059669' }}>
            <CheckCircle2 size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">Credit Approved</span>
            <div className="kpi-value-row">
              <span className="kpi-value">{customers.filter(c => (c.credit_status || 'GOOD_STANDING') === 'GOOD_STANDING').length || customers.length}</span>
              <span className="kpi-sub-tag text-emerald">Good Financial Standing</span>
            </div>
          </div>
        </div>

        <div className="cust-kpi-card kpi-card-opps">
          <div className="kpi-icon-badge bg-teal-badge" style={{ backgroundColor: '#f0fdfa', color: '#0d9488' }}>
            <DollarSign size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">Payment Terms</span>
            <div className="kpi-value-row">
              <span className="kpi-value">NET15 / NET30</span>
              <span className="kpi-sub-tag text-teal" style={{ color: '#0d9488' }}>Standard Terms Enforced</span>
            </div>
          </div>
        </div>
      </div>

      {/* Search & Filter Toolbar */}
      <div className="cust-toolbar-card">
        <div className="cust-search-box">
          <Search size={16} className="search-icon" />
          <input
            type="text"
            placeholder="Search customers by company name, customer code, contact person, or email..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </div>

        <div className="cust-filters-group">
          <div className="cust-filter-select">
            <label>Status:</label>
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setPage(1);
              }}
            >
              <option value="ALL">All Statuses</option>
              <option value="ACTIVE">Active</option>
              <option value="INACTIVE">Inactive</option>
              <option value="PROSPECT">Prospect</option>
            </select>
          </div>

          <div className="cust-filter-select">
            <label>Type:</label>
            <select
              value={typeFilter}
              onChange={(e) => {
                setTypeFilter(e.target.value);
                setPage(1);
              }}
            >
              <option value="ALL">All Types</option>
              <option value="SHIPPER">Shipper / Exporter</option>
              <option value="CONSIGNEE">Consignee / Importer</option>
              <option value="FORWARDER">Freight Forwarder</option>
              <option value="3PL">3PL Provider</option>
            </select>
          </div>

          <div className="cust-filter-select">
            <label>Country:</label>
            <select
              value={countryFilter}
              onChange={(e) => {
                setCountryFilter(e.target.value);
                setPage(1);
              }}
            >
              <option value="ALL">All Countries</option>
              {availableCountries.map((c) => (
                <option key={c} value={c}>
                  {c} {COUNTRY_FLAGS[c] || ''}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Customer Directory Table */}
      <div className="cust-table-card">
        {loading ? (
          <div className="cust-loading-state">
            <div className="cust-spinner" />
            <span>Evaluating Customer Intelligence & Activity Signals...</span>
          </div>
        ) : error ? (
          <div className="cust-error-state">
            <AlertTriangle size={24} />
            <span>{error}</span>
            <button className="btn-secondary-action" onClick={fetchData} style={{ marginTop: '0.75rem' }}>
              <RefreshCw size={14} /> Retry Connection
            </button>
          </div>
        ) : customers.length === 0 && !search && statusFilter === 'ALL' && typeFilter === 'ALL' && countryFilter === 'ALL' ? (
          <ModuleHeroEmptyState
            icon={<Building2 size={28} />}
            badgeTheme="blue"
            title="No Registered Customer Accounts Yet"
            description="Manage your verified shippers, consignees, and 3PL partners with automated health scores, credit limit controls, and 360° operational visibility."
            primaryAction={{
              label: 'Create First Customer',
              icon: <Plus size={15} />,
              onClick: handleCreateNew,
            }}
            secondaryAction={{
              label: 'Import from Leads',
              icon: <Users size={15} />,
              onClick: () => navigate('/dashboard/leads'),
            }}
            features={[
              {
                icon: <Activity size={18} />,
                iconBg: '#eff6ff',
                iconColor: '#2563eb',
                title: '360° Account Health & Retention',
                desc: 'Real-time telemetry on quote-to-booking conversion, invoice aging, and order volume patterns.',
              },
              {
                icon: <DollarSign size={18} />,
                iconBg: '#ecfdf5',
                iconColor: '#059669',
                title: 'Credit Standing & Margin Guard',
                desc: 'Configure flexible payment terms, credit limits, and automatic alerts for overdue balances.',
              },
              {
                icon: <FileText size={18} />,
                iconBg: '#f5f3ff',
                iconColor: '#7c3aed',
                title: 'Centralized Master Documents',
                desc: 'Store company KYC, tax registrations, commercial contracts, and key operational contacts.',
              },
            ]}
          />
        ) : customers.length === 0 ? (
          <div className="cust-empty-state">
            <div className="empty-state-icon-circle">
              <Building2 size={36} />
            </div>
            <h3>No Customer Accounts Found</h3>
            <p>There are no customer profiles matching your active search query or filter criteria.</p>
            <div className="empty-state-actions">
              <button
                className="btn-secondary-action"
                onClick={() => {
                  setSearch('');
                  setStatusFilter('ALL');
                  setTypeFilter('ALL');
                  setCountryFilter('ALL');
                  setPage(1);
                }}
              >
                Clear Filters
              </button>
              <button className="btn-primary-action" onClick={handleCreateNew}>
                <Plus size={16} /> Create First Customer
              </button>
            </div>
          </div>
        ) : (
          <table className="cust-table">
            <thead>
              <tr>
                <th>Customer Name</th>
                <th>Health Status</th>
                <th>Activity</th>
                <th>Credit Standing</th>
                <th>Contact & Owner</th>
                <th>Country</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {customers.map((c) => {
                const avatarColor = getAvatarColor(c.name);
                const flag = COUNTRY_FLAGS[c.country] || '🌐';

                return (
                  <tr key={c.id}>
                    <td>
                      <div className="cust-name-cell">
                        <div className={`cust-avatar ${avatarColor}`}>
                          {(c.name || 'CU').slice(0, 2).toUpperCase()}
                        </div>
                        <div className="cust-name-info">
                          <button
                            className="cust-name-btn"
                            onClick={() => navigate(`/dashboard/customers/${c.id}`)}
                          >
                            {c.name}
                          </button>
                          <div className="cust-code-chip">{c.customer_code}</div>
                        </div>
                      </div>
                    </td>
                    <td>{renderHealthIndicator(c.health_status, c.health_score)}</td>
                    <td>{renderTrendIndicator(c.activity_trend)}</td>
                    <td>{renderCreditBadge(c.credit_status)}</td>
                    <td>
                      <div className="contact-cell">
                        <span className="contact-person-name">{c.contact_name || '—'}</span>
                        <span className="contact-person-email">{c.contact_email || '—'}</span>
                        {c.account_owner_name?.trim() && (
                          <span className="owner-tag-inline">👤 {c.account_owner_name.trim()}</span>
                        )}
                      </div>
                    </td>
                    <td>
                      <span className="country-chip">
                        {flag} {c.country || 'India'}
                      </span>
                    </td>
                    <td>
                      <div className="table-actions-cell">
                        <button
                          className="btn-view-360"
                          onClick={() => navigate(`/dashboard/customers/${c.id}`)}
                        >
                          View 360°
                        </button>
                        <button
                          className="btn-followup-action"
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '4px',
                            padding: '4px 8px',
                            fontSize: '12px',
                            fontWeight: '600',
                            borderRadius: '6px',
                            border: '1px solid #bfdbfe',
                            backgroundColor: '#eff6ff',
                            color: '#1d4ed8',
                            cursor: 'pointer',
                            whiteSpace: 'nowrap'
                          }}
                          onClick={() => {
                            setFollowupCustomer(c);
                            setIsFollowupDrawerOpen(true);
                          }}
                        >
                          <MessageSquare size={13} /> Follow-Up
                        </button>
                        <div className="kebab-wrapper">
                          <button
                            className="btn-kebab"
                            onClick={(e) => {
                              e.stopPropagation();
                              if (activeKebabId === c.id) {
                                setActiveKebabId(null);
                              } else {
                                const rect = e.currentTarget.getBoundingClientRect();
                                setKebabPos({
                                  top: rect.bottom + 4,
                                  right: window.innerWidth - rect.right,
                                });
                                setActiveKebabId(c.id);
                              }
                            }}
                          >
                            <MoreVertical size={16} />
                          </button>

                          {activeKebabId === c.id && (
                            <div
                              className="kebab-menu"
                              style={{ top: kebabPos.top, right: kebabPos.right }}
                              onClick={(e) => e.stopPropagation()}
                            >
                              <button onClick={() => { navigate(`/dashboard/customers/${c.id}`); setActiveKebabId(null); }}>
                                <ExternalLink size={14} /> View 360° Profile
                              </button>
                              <button onClick={() => { handleEdit(c); setActiveKebabId(null); }}>
                                <Edit2 size={14} /> Edit Customer
                              </button>
                              <button onClick={() => handleArchive(c.id)} className="text-danger">
                                <Archive size={14} /> Archive Customer
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
        )}

        {/* Pagination Footer */}
        <div className="cust-pagination-footer">
          <span className="pagination-info">
            Showing {(page - 1) * limit + 1} to {Math.min(page * limit, total)} of {total} accounts
          </span>
          <div className="pagination-controls">
            <button
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
              className="pagination-btn"
            >
              <ChevronLeft size={16} /> Prev
            </button>
            <span className="page-indicator">
              Page {page} of {totalPages}
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() => setPage(page + 1)}
              className="pagination-btn"
            >
              Next <ChevronRight size={16} />
            </button>
          </div>
        </div>
      </div>

      {/* Customer Create/Edit Modal */}
      <CustomerModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={fetchData}
        initialData={selectedCustomer}
      />

      {/* Phase 5 Task 5.4: Autonomous Customer Follow-Up Drawer */}
      <CustomerFollowupDrawer
        isOpen={isFollowupDrawerOpen}
        onClose={() => setIsFollowupDrawerOpen(false)}
        customerId={followupCustomer?.id}
        customerData={followupCustomer}
        onFollowupUpdated={fetchData}
      />
    </div>
  );
}
