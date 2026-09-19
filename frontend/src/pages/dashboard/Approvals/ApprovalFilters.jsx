import React from 'react';
import { Search, Filter, Calendar, UserCheck, ShieldAlert, Layers } from 'lucide-react';

export default function ApprovalFilters({
  activeCategory,
  onCategoryChange,
  categoryCounts,
  searchQuery,
  onSearchChange,
  typeFilter,
  onTypeFilterChange,
  statusFilter,
  onStatusFilterChange,
  riskFilter,
  onRiskFilterChange,
  moduleFilter,
  onModuleFilterChange,
  requesterFilter,
  onRequesterFilterChange,
  requesterOptions = [],
  dateFilter,
  onDateFilterChange,
  sortBy,
  onSortChange,
  onClearAll,
}) {
  const categories = [
    { id: 'ALL', label: 'All', count: categoryCounts.ALL || 0 },
    { id: 'PENDING', label: 'Pending', count: categoryCounts.PENDING || 0 },
    { id: 'ASSIGNED_TO_ME', label: 'Assigned to Me', count: categoryCounts.ASSIGNED_TO_ME || 0 },
    { id: 'RETURNED_FOR_CHANGES', label: 'Returned for Changes', count: categoryCounts.RETURNED_FOR_CHANGES || 0 },
    { id: 'DOCUMENTS', label: 'Documents', count: categoryCounts.DOCUMENTS || 0 },
    { id: 'COMMERCIAL', label: 'Commercial', count: categoryCounts.COMMERCIAL || 0 },
    { id: 'OPERATIONS', label: 'Operations', count: categoryCounts.OPERATIONS || 0 },
    { id: 'FINANCE', label: 'Finance', count: categoryCounts.FINANCE || 0 },
  ];

  return (
    <div className="approval-filters-container">
      {/* Top Tab & Action Row */}
      <div className="approval-category-tabs-bar">
        <div className="category-tabs-list">
          {categories.map((cat) => (
            <button
              key={cat.id}
              className={`category-tab-btn ${activeCategory === cat.id ? 'active' : ''}`}
              onClick={() => onCategoryChange(cat.id)}
            >
              <span>{cat.label}</span>
              <span className={`tab-count-badge ${activeCategory === cat.id ? 'active' : ''}`}>
                {cat.count}
              </span>
            </button>
          ))}
        </div>

        <div className="category-right-actions">
          <select
            className="sort-by-select"
            value={sortBy}
            onChange={(e) => onSortChange(e.target.value)}
          >
            <option value="NEWEST">Sort by: Newest</option>
            <option value="DUE_DATE">Sort by: Due Date (Earliest)</option>
            <option value="PRIORITY">Sort by: Priority (Urgent)</option>
            <option value="RISK">Sort by: Risk (High to Low)</option>
            <option value="OLDEST">Sort by: Oldest</option>
          </select>
        </div>
      </div>

      {/* Toolbar Search & Select Filters */}
      <div className="approval-toolbar-card" style={{ flexWrap: 'wrap', gap: 10 }}>
        {/* Search Box */}
        <div className="search-input-box" style={{ flex: '1 1 240px' }}>
          <Search size={15} className="search-icon" />
          <input
            type="text"
            placeholder="Search approvals by title, ID, customer, ref, action..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
          />
        </div>

        {/* Filter Dropdowns */}
        <select
          className="filter-select-dropdown"
          value={statusFilter}
          onChange={(e) => onStatusFilterChange(e.target.value)}
        >
          <option value="ALL">All Statuses</option>
          <option value="Pending">Pending Approval</option>
          <option value="Returned for Changes">Returned for Changes</option>
          <option value="Approved">Approved</option>
          <option value="Executing">Executing</option>
          <option value="Completed">Completed</option>
          <option value="Failed">Failed Execution</option>
          <option value="Rejected">Rejected</option>
          <option value="Cancelled">Cancelled</option>
          <option value="Expired">Expired</option>
          <option value="Overdue">Overdue</option>
        </select>

        <select
          className="filter-select-dropdown"
          value={riskFilter || 'ALL'}
          onChange={(e) => onRiskFilterChange && onRiskFilterChange(e.target.value)}
        >
          <option value="ALL">All Risk Levels</option>
          <option value="CRITICAL">Critical Risk</option>
          <option value="HIGH_RISK">High Risk</option>
          <option value="MEDIUM">Medium Risk</option>
          <option value="LOW">Low Risk</option>
        </select>

        <select
          className="filter-select-dropdown"
          value={moduleFilter || 'ALL'}
          onChange={(e) => onModuleFilterChange && onModuleFilterChange(e.target.value)}
        >
          <option value="ALL">All Modules</option>
          <option value="SHIPMENTS">Shipments</option>
          <option value="INVOICES">Invoices / Billing</option>
          <option value="QUOTATIONS">Quotations / Rates</option>
          <option value="CONTRACTS">Contracts</option>
          <option value="CUSTOMERS">Customers</option>
          <option value="DOCUMENTS">Documents</option>
          <option value="COMPLIANCE">Compliance</option>
        </select>

        <select
          className="filter-select-dropdown"
          value={typeFilter}
          onChange={(e) => onTypeFilterChange(e.target.value)}
        >
          <option value="ALL">All Action Types</option>
          <option value="Document Approval">Document Approval</option>
          <option value="Commercial Approval">Commercial Approval</option>
          <option value="Operations Approval">Operations Approval</option>
          <option value="Finance Approval">Finance Approval</option>
          <option value="Clarification Email Approval">Clarification Email</option>
          <option value="AI Action">AI Autonomous Action</option>
        </select>

        <select
          className="filter-select-dropdown"
          value={requesterFilter}
          onChange={(e) => onRequesterFilterChange(e.target.value)}
        >
          <option value="ALL">All Requesters</option>
          {requesterOptions.map((req) => (
            <option key={req} value={req}>
              {req}
            </option>
          ))}
        </select>

        <div className="date-filter-box">
          <Calendar size={14} className="date-icon" />
          <select
            className="filter-select-dropdown date-select"
            value={dateFilter}
            onChange={(e) => onDateFilterChange(e.target.value)}
          >
            <option value="ANYTIME">Any Date</option>
            <option value="TODAY">Due Today</option>
            <option value="DUE_SOON">Due Soon (24h)</option>
            <option value="OVERDUE">Overdue Only</option>
          </select>
        </div>

        {/* Clear All Link */}
        <button type="button" className="btn-clear-filters" onClick={onClearAll}>
          Clear All
        </button>
      </div>
    </div>
  );
}

