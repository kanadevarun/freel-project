import React, { useState, useEffect, useCallback } from 'react';
import { 
  Plus, Search, Filter, FileText, CheckCircle, 
  ShieldAlert, FileEdit, ChevronRight, AlertTriangle, 
  AlertCircle, Plane, Ship, Truck, Layers, Calendar, 
  Clock, DollarSign, ArrowUpRight, Sparkles, ShieldCheck,
  RefreshCw, X, ArrowUpDown, ArrowUp, ArrowDown
} from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import PageHeader from '../../../components/dashboard/PageHeader';
import ContractDrawer from './ContractDrawer';
import ContractForm from './ContractForm';
import ContractAttentionPanel from './ContractAttentionPanel';
import ContractCreationChoiceModal from './ContractCreationChoiceModal';
import ContractImportModal from './ContractImportModal';
import ContractImportReviewModal from './ContractImportReviewModal';
import ModuleHeroEmptyState from '../../../components/dashboard/ModuleHeroEmptyState';
import './ContractsPage.css';

export default function ContractsPage() {
  const [overview, setOverview] = useState({
    total_contracts: 0,
    active_contracts: 0,
    expiring_soon: 0,
    expired_contracts: 0,
    draft_contracts: 0,
    total_value: 0
  });
  const [contracts, setContracts] = useState([]);
  const [attentionItems, setAttentionItems] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isEvaluating, setIsEvaluating] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeTab, setActiveTab] = useState('ALL');
  
  // Sorting State
  const [sortField, setSortField] = useState('created_at');
  const [sortDirection, setSortDirection] = useState('desc');
  
  // Creation & Import Modals State
  const [isChoiceModalOpen, setIsChoiceModalOpen] = useState(false);
  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [isReviewModalOpen, setIsReviewModalOpen] = useState(false);
  const [importData, setImportData] = useState(null);
  
  // Drawer & Form States
  const [selectedContract, setSelectedContract] = useState(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [contractToEdit, setContractToEdit] = useState(null);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [overviewData, contractsList, attentionRes] = await Promise.all([
        contractsService.getOverview().catch(() => null),
        contractsService.listContracts({ 
          search: searchQuery,
          contract_type: activeTab !== 'ALL' ? activeTab : undefined
        }),
        contractsService.getAttentionItems().catch(() => ({ data: [] }))
      ]);

      setOverview(overviewData || {
        total_contracts: 0,
        active_contracts: 0,
        expiring_soon: 0,
        expired_contracts: 0,
        draft_contracts: 0,
        total_value: 0
      });
      setContracts(contractsList?.data?.data || contractsList?.data || contractsList || []);
      setAttentionItems(attentionRes?.data || []);
    } catch (err) {
      console.error('Failed to fetch contracts', err);
    } finally {
      setIsLoading(false);
    }
  }, [searchQuery, activeTab]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleEvaluate = async () => {
    setIsEvaluating(true);
    try {
      await contractsService.evaluateLifecycle();
      await fetchData();
    } catch (err) {
      console.error('Failed to evaluate lifecycle', err);
    } finally {
      setIsEvaluating(false);
    }
  };

  const handleRowClick = (contract) => {
    setSelectedContract(contract);
    setIsDrawerOpen(true);
  };

  const handleSelectContractById = (contractID) => {
    const found = contracts.find((c) => c.id === contractID);
    if (found) {
      setSelectedContract(found);
      setIsDrawerOpen(true);
    } else {
      contractsService.getContract(contractID).then((res) => {
        if (res) {
          setSelectedContract(res.data || res);
          setIsDrawerOpen(true);
        }
      });
    }
  };

  const handleCreateNew = () => {
    setIsChoiceModalOpen(true);
  };

  const handleSelectManual = () => {
    setIsChoiceModalOpen(false);
    setContractToEdit(null);
    setIsFormOpen(true);
  };

  const handleSelectImport = () => {
    setIsChoiceModalOpen(false);
    setIsImportModalOpen(true);
  };

  const handleExtractionComplete = (res) => {
    setIsImportModalOpen(false);
    setImportData(res);
    setIsReviewModalOpen(true);
  };

  const handleImportSuccess = (createdContract) => {
    setIsReviewModalOpen(false);
    fetchData();
    if (createdContract) {
      setSelectedContract(createdContract);
      setIsDrawerOpen(true);
    }
  };

  const handleReupload = () => {
    setIsReviewModalOpen(false);
    setIsImportModalOpen(true);
  };

  const handleEdit = (contract) => {
    setIsDrawerOpen(false);
    setContractToEdit(contract);
    setIsFormOpen(true);
  };

  const onFormSuccess = () => {
    setIsFormOpen(false);
    fetchData();
  };

  const getModeIcon = (mode) => {
    switch ((mode || '').toUpperCase()) {
      case 'AIR': return <Plane size={13} className="mode-ico-air" />;
      case 'OCEAN':
      case 'SEA': return <Ship size={13} className="mode-ico-ocean" />;
      case 'ROAD': return <Truck size={13} className="mode-ico-road" />;
      default: return <Layers size={13} className="mode-ico-multi" />;
    }
  };

  const computeAttention = (contract) => {
    if (contract.status === 'EXPIRED') {
      return (
        <span className="attention-pill expired" title="Contract expired">
          <AlertCircle size={11} /> Expired
        </span>
      );
    }
    if (contract.status === 'ACTIVE' && contract.expiry_date) {
      const today = new Date();
      const expiry = new Date(contract.expiry_date);
      const diffDays = Math.ceil((expiry.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));
      if (diffDays <= 7 && diffDays >= 0) {
        return (
          <span className="attention-pill critical" title={`Critical: Expires in ${diffDays} days`}>
            <AlertCircle size={11} /> {diffDays}d left
          </span>
        );
      }
      if (diffDays <= 30 && diffDays > 7) {
        return (
          <span className="attention-pill warning" title={`Action required: Expires in ${diffDays} days`}>
            <AlertTriangle size={11} /> {diffDays}d left
          </span>
        );
      }
    }
    return <span className="attention-clean">—</span>;
  };

  const renderStatusBadge = (status) => {
    switch (status) {
      case 'ACTIVE':
        return (
          <span className="status-badge active">
            <span className="status-dot-pulse"></span>
            Active
          </span>
        );
      case 'DRAFT':
        return (
          <span className="status-badge draft">
            <span className="status-dot-slate"></span>
            Draft
          </span>
        );
      case 'EXPIRED':
        return (
          <span className="status-badge expired">
            <span className="status-dot-rose"></span>
            Expired
          </span>
        );
      case 'ARCHIVED':
        return (
          <span className="status-badge default">
            <span className="status-dot-gray"></span>
            Archived
          </span>
        );
      default:
        return <span className="status-badge default">{status}</span>;
    }
  };

  const handleSort = (field) => {
    if (sortField === field) {
      setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const getSortIcon = (field) => {
    if (sortField !== field) return <ArrowUpDown size={12} className="th-sort-icon muted" />;
    return sortDirection === 'asc' 
      ? <ArrowUp size={12} className="th-sort-icon active" /> 
      : <ArrowDown size={12} className="th-sort-icon active" />;
  };

  const sortedContracts = [...contracts].sort((a, b) => {
    let aVal = a[sortField];
    let bVal = b[sortField];

    if (sortField === 'attention') {
      const getAttentionRank = (c) => {
        if (c.status === 'EXPIRED') return 1;
        if (c.status === 'ACTIVE' && c.expiry_date) {
          const diff = Math.ceil((new Date(c.expiry_date).getTime() - Date.now()) / 86400000);
          if (diff <= 7 && diff >= 0) return 2;
          if (diff <= 30 && diff > 7) return 3;
        }
        return 4;
      };
      aVal = getAttentionRank(a);
      bVal = getAttentionRank(b);
    } else if (sortField === 'expiry_date') {
      aVal = a.expiry_date ? new Date(a.expiry_date).getTime() : 9999999999999;
      bVal = b.expiry_date ? new Date(b.expiry_date).getTime() : 9999999999999;
    } else if (typeof aVal === 'string') {
      aVal = (aVal || '').toLowerCase();
      bVal = (bVal || '').toLowerCase();
      return sortDirection === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
    }

    if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
    if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
    return 0;
  });

  return (
    <div className="contracts-page">
      
      {/* ── Top Page Header ── */}
      <div className="contracts-page-header">
        <div className="cph-left">
          <div className="cph-badge-row">
            <span className="cph-tag">COMMERCIAL SUITE</span>
            <span className="cph-dot">·</span>
            <span className="cph-status-live">
              <span className="cph-pulse"></span>
              Live Repository
            </span>
          </div>
          <h1 className="cph-title">Commercial Contracts</h1>
          <p className="cph-description">
            Unified commercial agreement repository, party SLAs, rate integration, and lifecycle management.
          </p>
        </div>
        <div className="cph-actions">
          <button className="cph-btn-primary" onClick={handleCreateNew}>
            <Plus size={16} />
            <span>New Contract</span>
          </button>
        </div>
      </div>

      {/* ── Contracts Commercial KPI Strip ── */}
      <div className="contracts-kpi-grid">
        
        {/* 1. Total Repository */}
        <div className="contracts-kpi-card kpi-blue">
          <div className="kpi-watermark-icon">
            <FileText size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper kpi-icon-blue">
              <FileText size={18} />
            </div>
            <span className="kpi-tag kpi-tag-blue">All Agreements</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Total Repository</span>
            <div className="kpi-card-number">{overview.total_contracts}</div>
            <span className="kpi-card-subtitle">Commercial agreements registered</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-blue"></span>
            <span className="kpi-footer-text">
              {overview.total_contracts ? `${overview.total_contracts} total active & historical records` : 'Portfolio baseline'}
            </span>
          </div>
        </div>

        {/* 2. In Effect / Active */}
        <div className="contracts-kpi-card kpi-emerald">
          <div className="kpi-watermark-icon">
            <ShieldCheck size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper kpi-icon-emerald">
              <ShieldCheck size={18} />
            </div>
            <span className="kpi-tag kpi-tag-emerald">In Effect</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Active Agreements</span>
            <div className="kpi-card-number text-emerald">{overview.active_contracts}</div>
            <span className="kpi-card-subtitle">Currently in commercial effect</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-emerald"></span>
            <span className="kpi-footer-text">
              {overview.total_contracts > 0 
                ? `${Math.round((overview.active_contracts / overview.total_contracts) * 100)}% portfolio operational` 
                : '100% operational'}
            </span>
          </div>
        </div>

        {/* 3. Action Required / Expiring */}
        <div className="contracts-kpi-card kpi-amber">
          <div className="kpi-watermark-icon">
            <Clock size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper kpi-icon-amber">
              <Clock size={18} />
            </div>
            <span className="kpi-tag kpi-tag-amber">Action Required</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Expiring Soon</span>
            <div className="kpi-card-number text-amber">{overview.expiring_soon}</div>
            <span className="kpi-card-subtitle">Expiring within 30 days</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-amber"></span>
            <span className="kpi-footer-text">
              {overview.expiring_soon > 0 ? `${overview.expiring_soon} requires renewal review` : 'No immediate renewal actions'}
            </span>
          </div>
        </div>

        {/* 4. Pipeline / Draft */}
        <div className="contracts-kpi-card kpi-purple">
          <div className="kpi-watermark-icon">
            <FileEdit size={64} />
          </div>
          <div className="kpi-card-header">
            <div className="kpi-icon-wrapper kpi-icon-purple">
              <FileEdit size={18} />
            </div>
            <span className="kpi-tag kpi-tag-purple">Pipeline</span>
          </div>
          <div className="kpi-card-body">
            <span className="kpi-card-label">Draft Contracts</span>
            <div className="kpi-card-number text-purple">{overview.draft_contracts}</div>
            <span className="kpi-card-subtitle">Awaiting execution & activation</span>
          </div>
          <div className="kpi-card-footer">
            <span className="kpi-indicator-dot dot-purple"></span>
            <span className="kpi-footer-text">
              {overview.draft_contracts > 0 ? `${overview.draft_contracts} pending counterparties` : 'Execution queue clear'}
            </span>
          </div>
        </div>

      </div>

      {/* ── Contextual Portfolio Footer Strip ── */}
      <div className="contracts-portfolio-strip">
        <div className="portfolio-strip-left">
          <Layers size={16} className="portfolio-strip-icon" />
          <span>
            <strong>Master Agreement Repository</strong> — Cross-module bindings with Rates, Spot Quotations, and Operational Shipments.
          </span>
        </div>
        <div className="portfolio-strip-right">
          <span className="cph-pulse"></span>
          <span>Verified Commercial Records</span>
        </div>
      </div>

      {/* ── Attention & Risk Management Intelligence Banner ── */}
      <ContractAttentionPanel
        items={attentionItems}
        loading={isLoading || isEvaluating}
        onEvaluate={handleEvaluate}
        onSelectContract={handleSelectContractById}
      />

      {/* ── Main Workspace Table Card ── */}
      <div className="contracts-workspace">
        <div className="workspace-toolbar">
          <div className="tabs">
            {[
              { key: 'ALL', label: 'All Contracts' },
              { key: 'CUSTOMER', label: 'Customer SLAs' },
              { key: 'CARRIER', label: 'Carrier Agreements' },
              { key: 'VENDOR', label: 'Vendor Contracts' }
            ].map(tab => (
              <button
                key={tab.key}
                className={`tab-btn ${activeTab === tab.key ? 'active' : ''}`}
                onClick={() => setActiveTab(tab.key)}
              >
                <span>{tab.label}</span>
              </button>
            ))}
          </div>

          <div className="toolbar-actions">
            <div className="search-box">
              <Search size={14} className="search-icon" />
              <input
                type="text"
                placeholder="Search by contract name, ref, or party..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              {searchQuery && (
                <button 
                  type="button" 
                  className="search-clear-btn" 
                  onClick={() => setSearchQuery('')}
                >
                  <X size={13} />
                </button>
              )}
            </div>
            <div className="toolbar-count-badge">
              <span>{contracts.length} {contracts.length === 1 ? 'Record' : 'Records'}</span>
            </div>
          </div>
        </div>

        {/* ── Table ── */}
        <div className="table-container">
          <table className="data-table">
            <thead>
              <tr>
                <th style={{ width: '25%' }} className="sortable-th" onClick={() => handleSort('contract_name')}>
                  <div className="th-content">
                    <span>Agreement Name</span>
                    {getSortIcon('contract_name')}
                  </div>
                </th>
                <th style={{ width: '13%' }} className="sortable-th" onClick={() => handleSort('contract_reference')}>
                  <div className="th-content">
                    <span>Reference</span>
                    {getSortIcon('contract_reference')}
                  </div>
                </th>
                <th style={{ width: '18%' }} className="sortable-th" onClick={() => handleSort('party_name')}>
                  <div className="th-content">
                    <span>Counterparty</span>
                    {getSortIcon('party_name')}
                  </div>
                </th>
                <th style={{ width: '13%' }} className="sortable-th" onClick={() => handleSort('transport_mode')}>
                  <div className="th-content">
                    <span>Mode & Type</span>
                    {getSortIcon('transport_mode')}
                  </div>
                </th>
                <th style={{ width: '11%' }} className="sortable-th" onClick={() => handleSort('status')}>
                  <div className="th-content">
                    <span>Status</span>
                    {getSortIcon('status')}
                  </div>
                </th>
                <th style={{ width: '10%' }} className="sortable-th" onClick={() => handleSort('expiry_date')}>
                  <div className="th-content">
                    <span>Expiry</span>
                    {getSortIcon('expiry_date')}
                  </div>
                </th>
                <th style={{ width: '10%' }} className="sortable-th" onClick={() => handleSort('attention')}>
                  <div className="th-content">
                    <span>Attention</span>
                    {getSortIcon('attention')}
                  </div>
                </th>
                <th style={{ width: '4%' }}></th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan="8" className="table-status-cell">
                    <div className="loading-spinner-wrap">
                      <Clock size={20} className="spin-animate text-muted" />
                      <span>Loading commercial agreements...</span>
                    </div>
                  </td>
                </tr>
              ) : sortedContracts.length === 0 && !searchQuery && activeTab === 'ALL' ? (
                <tr>
                  <td colSpan="8" style={{ padding: '0', border: 'none' }}>
                    <ModuleHeroEmptyState
                      icon={<FileText size={28} />}
                      badgeTheme="emerald"
                      title="No Active Commercial Contracts"
                      description="Govern customer service agreements, carrier master service agreements (MSAs), and SLA commitments with automated expiry tracking and clause management."
                      primaryAction={{
                        label: 'Create First Contract',
                        icon: <Plus size={15} />,
                        onClick: () => setIsChoiceModalOpen(true),
                      }}
                      secondaryAction={{
                        label: 'Import Contracts',
                        icon: <Sparkles size={15} />,
                        onClick: () => setIsImportModalOpen(true),
                      }}
                      features={[
                        {
                          icon: <ShieldCheck size={18} />,
                          iconBg: '#eff6ff',
                          iconColor: '#2563eb',
                          title: 'Carrier & Shipper Agreements',
                          desc: 'Track minimum quantity commitments (MQC), free days, detention terms, and rate inclusions.',
                        },
                        {
                          icon: <Clock size={18} />,
                          iconBg: '#fef3c7',
                          iconColor: '#d97706',
                          title: 'Automated Renewal & Expiry Guard',
                          desc: 'Get proactive alerts before contracts lapse to prevent commercial rate leakage.',
                        },
                        {
                          icon: <FileEdit size={18} />,
                          iconBg: '#f5f3ff',
                          iconColor: '#7c3aed',
                          title: 'Clause & Margin Compliance',
                          desc: 'Ensure quotation pricing strictly conforms to signed client contractual terms and volume tiers.',
                        },
                      ]}
                    />
                  </td>
                </tr>
              ) : sortedContracts.length === 0 ? (
                <tr>
                  <td colSpan="8" className="table-status-cell">
                    <div className="table-empty-wrap">
                      <FileText size={32} className="text-muted mb-2" />
                      <p className="empty-head">No Contracts Found</p>
                      <p className="empty-sub">
                        {searchQuery ? `No contracts matching "${searchQuery}"` : 'No agreements in this category yet. Click "New Contract" to add one.'}
                      </p>
                    </div>
                  </td>
                </tr>
              ) : (
                sortedContracts.map(contract => {
                  const attention = computeAttention(contract);
                  const isPendingParty = !contract.party_name || contract.party_name === 'Pending Resolution' || contract.party_name === 'Pending';
                  return (
                    <tr key={contract.id} onClick={() => handleRowClick(contract)} className="clickable-row">
                      <td>
                        <div className="contract-name-cell">
                          <span className="contract-name-text">{contract.contract_name}</span>
                          {contract.owner && (
                            <span className="contract-owner-sub">
                              <span className="owner-dot"></span>
                              Owner: {contract.owner}
                            </span>
                          )}
                        </div>
                      </td>
                      <td>
                        <span className="contract-ref-badge">{contract.contract_reference}</span>
                      </td>
                      <td>
                        <div className="party-cell-wrapper">
                          <span className={`party-name-cell ${isPendingParty ? 'party-pending' : ''}`}>
                            {contract.party_name || 'Pending Resolution'}
                          </span>
                        </div>
                      </td>
                      <td>
                        <div className="mode-type-wrap">
                          <span className="mode-badge-cell">
                            {getModeIcon(contract.transport_mode)}
                            <span className="mode-name">{contract.transport_mode || 'MULTIMODAL'}</span>
                          </span>
                          <span className="type-subtext">{contract.contract_type?.replace(/_/g, ' ')}</span>
                        </div>
                      </td>
                      <td>
                        {renderStatusBadge(contract.status)}
                      </td>
                      <td>
                        <span className="expiry-cell-text">{contract.expiry_date || 'No Expiry'}</span>
                      </td>
                      <td>
                        {attention}
                      </td>
                      <td className="row-action-cell">
                        <ChevronRight size={15} className="row-chevron" />
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* ── Table Footer Summary ── */}
        <div className="table-footer-bar">
          <div className="table-footer-left">
            <span>Showing <strong>{sortedContracts.length}</strong> commercial {sortedContracts.length === 1 ? 'agreement' : 'agreements'}</span>
          </div>
          <div className="table-footer-right">
            <span>Click any row to open contract intelligence drawer</span>
          </div>
        </div>
      </div>

      {/* ── Detail Drawer ── */}
      {isDrawerOpen && (
        <ContractDrawer 
          contract={selectedContract} 
          onClose={() => setIsDrawerOpen(false)} 
          onEdit={() => handleEdit(selectedContract)}
          onUpdate={fetchData}
        />
      )}

      {/* ── Create / Edit Form ── */}
      {isFormOpen && (
        <ContractForm 
          contract={contractToEdit}
          onClose={() => setIsFormOpen(false)}
          onSuccess={onFormSuccess}
        />
      )}

      {/* ── Contract Creation Choice Modal ── */}
      {isChoiceModalOpen && (
        <ContractCreationChoiceModal
          onClose={() => setIsChoiceModalOpen(false)}
          onSelectManual={handleSelectManual}
          onSelectImport={handleSelectImport}
        />
      )}

      {/* ── AI Contract Document Import Modal ── */}
      {isImportModalOpen && (
        <ContractImportModal
          onClose={() => setIsImportModalOpen(false)}
          onExtractionComplete={handleExtractionComplete}
        />
      )}

      {/* ── AI Contract Review & Confirmation Modal ── */}
      {isReviewModalOpen && (
        <ContractImportReviewModal
          importData={importData}
          onClose={() => setIsReviewModalOpen(false)}
          onReupload={handleReupload}
          onSuccess={handleImportSuccess}
        />
      )}
    </div>
  );
}
