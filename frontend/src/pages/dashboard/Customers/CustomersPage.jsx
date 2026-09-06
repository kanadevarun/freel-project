import React, { useState, useEffect } from 'react';
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
} from 'lucide-react';
import customerService from '../../../services/customerService';
import CustomerModal from './CustomerModal';
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

  // Attention Panel Active Tab
  const [attentionTab, setAttentionTab] = useState('WARNING');

  // Modals
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState(null);
  const [activeKebabId, setActiveKebabId] = useState(null);
  const [kebabPos, setKebabPos] = useState({ top: 0, right: 0 });

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
        customerService.getCustomerKPIs(),
        customerService.getIntelligenceSummary(),
        customerService.getAttentionItems(),
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
          <h1 className="cust-page-title">Customers Commercial Intelligence Workspace</h1>
          <p className="cust-page-subtitle">
            Proactive account health, risk detection, relationship governance, and customer directory.
          </p>
        </div>
        <div className="cust-header-actions">
          <button className="btn-secondary-action" onClick={fetchData}>
            <RefreshCw size={15} /> Refresh Intelligence
          </button>
          <button className="btn-primary-action" onClick={handleCreateNew}>
            <Plus size={16} /> New Customer
          </button>
        </div>
      </div>

      {/* Intelligence KPI Cards Grid */}
      <div className="cust-kpi-grid">
        <div className="cust-kpi-card kpi-card-healthy">
          <div className="kpi-icon-badge bg-green-badge">
            <CheckCircle2 size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">Healthy Accounts</span>
            <div className="kpi-value-row">
              <span className="kpi-value">{intelSummary.healthy_count || total || 0}</span>
              <span className="kpi-sub-tag text-emerald">80–100 Health Score</span>
            </div>
          </div>
        </div>

        <div className="cust-kpi-card kpi-card-watch">
          <div className="kpi-icon-badge bg-amber-badge">
            <Activity size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">Under Watch</span>
            <div className="kpi-value-row">
              <span className="kpi-value">{intelSummary.watch_count || 0}</span>
              <span className="kpi-sub-tag text-amber">60–79 Health Score</span>
            </div>
          </div>
        </div>

        <div className="cust-kpi-card kpi-card-atrisk">
          <div className="kpi-icon-badge bg-orange-badge">
            <AlertTriangle size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">At Risk Accounts</span>
            <div className="kpi-value-row">
              <span className="kpi-value">{intelSummary.at_risk_count ?? 0}</span>
              <span className="kpi-sub-tag text-orange">30–59 Health Score</span>
            </div>
          </div>
        </div>

        <div className="cust-kpi-card kpi-card-critical">
          <div className="kpi-icon-badge bg-red-badge">
            <ShieldAlert size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">Critical Accounts</span>
            <div className="kpi-value-row">
              <span className="kpi-value">{intelSummary.critical_count ?? 0}</span>
              <span className="kpi-sub-tag text-red">0–29 Health Score</span>
            </div>
          </div>
        </div>

        <div className="cust-kpi-card kpi-card-opps">
          <div className="kpi-icon-badge bg-purple-badge">
            <Zap size={20} />
          </div>
          <div className="kpi-info">
            <span className="kpi-label">Detected Opportunities</span>
            <div className="kpi-value-row">
              <span className="kpi-value">{intelSummary.total_opportunities ?? 0}</span>
              <span className="kpi-sub-tag text-purple">Renewals & Follow-ups</span>
            </div>
          </div>
        </div>
      </div>

      {/* Proactive Customer Attention Required Panel */}
      <div className="cust-attention-panel">
        <div className="attention-panel-header">
          <div className="attention-title">
            <div className="att-title-icon">
              <ShieldAlert size={18} />
            </div>
            <div>
              <h3>Attention Required Panel</h3>
              <p className="att-subtitle">Proactive account alerts requiring commercial manager review</p>
            </div>
          </div>

          <div className="attention-tabs">
            <button
              className={`att-tab tab-critical ${attentionTab === 'CRITICAL' ? 'active' : ''}`}
              onClick={() => setAttentionTab('CRITICAL')}
            >
              🔴 Critical ({attentionItems.filter((i) => i.severity === 'CRITICAL').length})
            </button>

            <button
              className={`att-tab tab-warning ${attentionTab === 'WARNING' ? 'active' : ''}`}
              onClick={() => setAttentionTab('WARNING')}
            >
              🟠 Warnings ({attentionItems.filter((i) => i.severity === 'WARNING').length})
            </button>

            <button
              className={`att-tab tab-attention ${attentionTab === 'ATTENTION' ? 'active' : ''}`}
              onClick={() => setAttentionTab('ATTENTION')}
            >
              🟡 Attention ({attentionItems.filter((i) => i.severity === 'ATTENTION').length})
            </button>
          </div>
        </div>

        <div className="attention-items-feed">
          {filteredAttentionItems.length === 0 ? (
            <div className="attention-empty-state">
              <CheckCircle2 size={20} className="text-emerald" />
              <span>No active critical alerts under this classification. All monitored accounts are in good standing.</span>
            </div>
          ) : (
            filteredAttentionItems.map((item, idx) => (
              <div key={idx} className="attention-card-row">
                <div className={`att-severity-bar bar-${item.severity.toLowerCase()}`} />
                <div className="att-content">
                  <div className="att-customer-line">
                    <strong className="att-cust-name">
                      {item.customer_name || 'Unnamed Customer'}
                    </strong>

                    {item.customer_code && (
                      <span className="att-cust-code-badge">
                        {item.customer_code}
                      </span>
                    )}

                    <span className="att-owner-tag">
                      Owner: {item.account_owner_name || 'Unassigned'}
                    </span>
                  </div>
                  <div className="att-issue-row">
                    <span className="att-issue-title">
                      {item.title || 'Customer attention required'}
                    </span>

                    {item.reason && (
                      <span className="att-reason-short">
                        — {item.reason.split('.')[0]}
                      </span>
                    )}
                  </div>
                </div>
                <div className="att-actions">
                  <button
                    className="btn-investigate"
                    onClick={() => navigate(`/dashboard/customers/${item.customer_id}`)}
                  >
                    Investigate 360° <ArrowRight size={13} />
                  </button>
                </div>
              </div>
            ))
          )}
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
              <option value="India">India 🇮🇳</option>
              <option value="UAE">UAE 🇦🇪</option>
              <option value="Sweden">Sweden 🇸🇪</option>
              <option value="Singapore">Singapore 🇸🇬</option>
              <option value="USA">USA 🇺🇸</option>
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
    </div>
  );
}
