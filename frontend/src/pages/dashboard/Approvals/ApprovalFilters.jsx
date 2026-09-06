import React from 'react';
import { Search, Filter, Calendar, UserCheck } from 'lucide-react';

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
  requesterFilter,
  onRequesterFilterChange,
  dateFilter,
  onDateFilterChange,
  sortBy,
  onSortChange,
  onClearAll,
}) {
  const categories = [
    { id: 'ALL', label: 'All', count: categoryCounts.ALL || 0 },
    { id: 'ASSIGNED_TO_ME', label: 'Assigned to Me', count: categoryCounts.ASSIGNED_TO_ME || 0 },
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
            <option value="OLDEST">Sort by: Oldest</option>
          </select>
        </div>
      </div>

      {/* Toolbar Search & Select Filters */}
      <div className="approval-toolbar-card">
        {/* Search Box */}
        <div className="search-input-box">
          <Search size={15} className="search-icon" />
          <input
            type="text"
            placeholder="Search approvals by title, ID, customer, shipment..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
          />
        </div>

        {/* Filter Dropdowns */}
        <select
          className="filter-select-dropdown"
          value={typeFilter}
          onChange={(e) => onTypeFilterChange(e.target.value)}
        >
          <option value="ALL">All Types</option>
          <option value="Document Approval">Document Approval</option>
          <option value="Commercial Approval">Commercial Approval</option>
          <option value="Operations Approval">Operations Approval</option>
          <option value="Finance Approval">Finance Approval</option>
        </select>

        <select
          className="filter-select-dropdown"
          value={statusFilter}
          onChange={(e) => onStatusFilterChange(e.target.value)}
        >
          <option value="ALL">All Status</option>
          <option value="Pending">Pending</option>
          <option value="Due Soon">Due Soon</option>
          <option value="Overdue">Overdue</option>
          <option value="Approved">Approved</option>
          <option value="Rejected">Rejected</option>
        </select>

        <select
          className="filter-select-dropdown"
          value={requesterFilter}
          onChange={(e) => onRequesterFilterChange(e.target.value)}
        >
          <option value="ALL">All Requesters</option>
          <option value="Varun Kanade">Varun Kanade</option>
          <option value="Arjun Singh">Arjun Singh</option>
          <option value="Neha Kapoor">Neha Kapoor</option>
          <option value="Rohit Mehta">Rohit Mehta</option>
          <option value="Pooja Shah">Pooja Shah</option>
          <option value="Vikram Kumar">Vikram Kumar</option>
        </select>

        <div className="date-filter-box">
          <Calendar size={14} className="date-icon" />
          <select
            className="filter-select-dropdown date-select"
            value={dateFilter}
            onChange={(e) => onDateFilterChange(e.target.value)}
          >
            <option value="ANYTIME">Anytime</option>
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
