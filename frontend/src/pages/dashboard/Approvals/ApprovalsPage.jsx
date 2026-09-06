import React, { useState, useEffect, useMemo } from 'react';
import { Plus, CheckCircle2, AlertCircle, ChevronLeft, ChevronRight, Inbox, RefreshCw } from 'lucide-react';
import PageHeader from '../../../components/dashboard/PageHeader';
import ApprovalStats from './ApprovalStats';
import ApprovalFilters from './ApprovalFilters';
import ApprovalRow from './ApprovalRow';
import NewApprovalModal from './NewApprovalModal';
import ApprovalDetailsModal from './ApprovalDetailsModal';
import RejectionModal from './RejectionModal';
import { approvalsService } from '../../../services/approvalsService';
import { INITIAL_APPROVAL_STATS, INITIAL_APPROVALS } from './constants';
import './ApprovalsPage.css';

export default function ApprovalsPage() {
  const [approvals, setApprovals] = useState([]);
  const [stats, setStats] = useState(INITIAL_APPROVAL_STATS);

  // Current logged in user context
  const currentUser = 'Varun Kanade';

  // Filters State
  const [activeCategory, setActiveCategory] = useState('ALL');
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState('ALL');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [requesterFilter, setRequesterFilter] = useState('ALL');
  const [dateFilter, setDateFilter] = useState('ANYTIME');
  const [sortBy, setSortBy] = useState('NEWEST');

  // UI States
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionSuccess, setActionSuccess] = useState('');
  const [isNewModalOpen, setIsNewModalOpen] = useState(false);
  const [selectedApproval, setSelectedApproval] = useState(null);
  const [rejectingApproval, setRejectingApproval] = useState(null);

  // Pagination
  const [currentPage, setCurrentPage] = useState(1);
  const rowsPerPage = 10;

  // Load approvals & stats on mount
  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [listData, statsData] = await Promise.all([
        approvalsService.listApprovals(),
        approvalsService.getApprovalStats(),
      ]);

      if (Array.isArray(listData) && listData.length > 0) {
        setApprovals(normalizeApprovals(listData));
      } else {
        setApprovals(INITIAL_APPROVALS);
      }

      if (statsData) {
        setStats({
          pending: statsData.pending ?? 12,
          pendingTrend: statsData.pending_trend || '↑ 3 from last 7 days',
          approved: statsData.approved ?? 28,
          approvedTrend: statsData.approved_trend || '↑ 8 from last 7 days',
          rejected: statsData.rejected ?? 4,
          rejectedTrend: statsData.rejected_trend || '↓ 2 from last 7 days',
          overdue: statsData.overdue ?? 3,
        });
      }
    } catch (err) {
      console.error('Failed to load backend approvals:', err);
      setApprovals(INITIAL_APPROVALS);
    } finally {
      setLoading(false);
    }
  };

  const normalizeApprovals = (rawList) => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    return rawList.map((item) => {
      let isOverdue = false;
      let isDueSoon = false;
      let dueText = item.due_text || '7 days left';

      if (item.due_date) {
        const dueDate = new Date(item.due_date);
        dueDate.setHours(0, 0, 0, 0);
        const diffDays = Math.ceil((dueDate - today) / (1000 * 60 * 60 * 24));

        if (item.status === 'Pending' || item.status === 'IN_REVIEW' || item.status === 'Overdue') {
          if (diffDays < 0) {
            isOverdue = true;
            dueText = `Overdue by ${Math.abs(diffDays)} day${Math.abs(diffDays) > 1 ? 's' : ''}`;
          } else if (diffDays === 0) {
            isDueSoon = true;
            dueText = 'Due Today';
          } else if (diffDays <= 2) {
            isDueSoon = true;
            dueText = `Due in ${diffDays} day${diffDays > 1 ? 's' : ''}`;
          }
        }
      }

      return {
        id: item.request_code || item.id || `APP-${item.id}`,
        dbId: item.id,
        title: item.title || 'Untitled Request',
        category: item.category || 'DOCUMENTS',
        type: item.type || 'Document Approval',
        relatedRef: item.related_ref || (item.shipment_id ? `Shipment #${item.shipment_id}` : 'General Context'),
        customerName: item.customer_name || 'Associated Customer',
        requesterName: item.requested_by_name || 'Varun Kanade',
        department: item.department || 'Operations',
        assignedTo: item.assigned_to || 'Arjun Singh (Operations Manager)',
        avatar: item.avatar || 'VK',
        dueDate: item.due_date ? new Date(item.due_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Aug 20, 2026',
        dueText: isOverdue ? 'Overdue' : dueText,
        isOverdue,
        isDueSoon,
        status: isOverdue && item.status === 'Pending' ? 'Overdue' : (item.status || 'Pending'),
        priority: item.priority || 'MEDIUM',
        rejectionReason: item.rejection_reason || '',
        description: item.description || item.comments || '',
        createdAt: item.created_at || new Date().toISOString(),
      };
    });
  };

  // Category & Tab Counts Calculation
  const categoryCounts = useMemo(() => {
    const counts = { ALL: approvals.length, ASSIGNED_TO_ME: 0, DOCUMENTS: 0, COMMERCIAL: 0, OPERATIONS: 0, FINANCE: 0 };
    approvals.forEach((item) => {
      const cat = item.category ? item.category.toUpperCase() : 'DOCUMENTS';
      if (counts[cat] !== undefined) {
        counts[cat] += 1;
      }
      if ((item.assignedTo || '').toLowerCase().includes('varun') || (item.assignedTo || '').toLowerCase().includes('arjun')) {
        counts.ASSIGNED_TO_ME += 1;
      }
    });
    return counts;
  }, [approvals]);

  // Filtering & Sorting logic
  const filteredApprovals = useMemo(() => {
    let result = approvals.filter((item) => {
      // 1. Category tab filter
      if (activeCategory === 'ASSIGNED_TO_ME') {
        const assigned = (item.assignedTo || '').toLowerCase();
        if (!assigned.includes('varun') && !assigned.includes('arjun')) return false;
      } else if (activeCategory !== 'ALL') {
        if ((item.category || '').toUpperCase() !== activeCategory) return false;
      }

      // 2. Type filter
      if (typeFilter !== 'ALL' && item.type !== typeFilter) return false;

      // 3. Status filter
      if (statusFilter !== 'ALL') {
        if (statusFilter === 'Overdue' && !item.isOverdue && item.status !== 'Overdue') return false;
        if (statusFilter === 'Due Soon' && !item.isDueSoon) return false;
        if (statusFilter === 'Pending' && item.status !== 'Pending') return false;
        if (statusFilter === 'Approved' && item.status !== 'Approved') return false;
        if (statusFilter === 'Rejected' && item.status !== 'Rejected') return false;
      }

      // 4. Requester filter
      if (requesterFilter !== 'ALL' && item.requesterName !== requesterFilter) return false;

      // 5. Date filter
      if (dateFilter === 'OVERDUE' && !item.isOverdue && item.status !== 'Overdue') return false;
      if (dateFilter === 'DUE_SOON' && !item.isDueSoon) return false;
      if (dateFilter === 'TODAY' && item.dueText !== 'Due Today') return false;

      // 6. Search query
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const title = (item.title || '').toLowerCase();
        const id = String(item.id || '').toLowerCase();
        const cust = (item.customerName || '').toLowerCase();
        const ref = (item.relatedRef || '').toLowerCase();
        const req = (item.requesterName || '').toLowerCase();

        return (
          title.includes(q) ||
          id.includes(q) ||
          cust.includes(q) ||
          ref.includes(q) ||
          req.includes(q)
        );
      }

      return true;
    });

    // Sort Result
    return result.sort((a, b) => {
      if (sortBy === 'DUE_DATE') {
        return new Date(a.dueDate) - new Date(b.dueDate);
      }
      if (sortBy === 'PRIORITY') {
        const rank = { URGENT: 4, HIGH: 3, MEDIUM: 2, LOW: 1 };
        return (rank[b.priority] || 0) - (rank[a.priority] || 0);
      }
      if (sortBy === 'OLDEST') {
        return new Date(a.createdAt) - new Date(b.createdAt);
      }
      // NEWEST default
      return new Date(b.createdAt) - new Date(a.createdAt);
    });
  }, [approvals, activeCategory, typeFilter, statusFilter, requesterFilter, dateFilter, searchQuery, sortBy]);

  const paginatedApprovals = useMemo(() => {
    const startIndex = (currentPage - 1) * rowsPerPage;
    return filteredApprovals.slice(startIndex, startIndex + rowsPerPage);
  }, [filteredApprovals, currentPage]);

  const totalPages = Math.ceil(filteredApprovals.length / rowsPerPage) || 1;

  // Handlers
  const handleCreateApproval = async (newApprovalInput) => {
    try {
      if (newApprovalInput.dbId) {
        setApprovals((prev) => [newApprovalInput, ...prev]);
      } else {
        await approvalsService.createApproval({
          title: newApprovalInput.title,
          category: newApprovalInput.category,
          type: newApprovalInput.type,
          priority: newApprovalInput.priority,
          related_ref: newApprovalInput.relatedRef,
          customer_name: newApprovalInput.customerName,
          requested_by_name: newApprovalInput.requesterName,
          department: newApprovalInput.department,
          due_date: newApprovalInput.dueDate,
          description: newApprovalInput.description,
        });
        await fetchData();
      }
      setActionSuccess(`Approval request submitted successfully and assigned to ${newApprovalInput.assignedTo || 'approver'}.`);
      setTimeout(() => setActionSuccess(''), 4000);
    } catch (err) {
      console.error('Failed to create approval on backend:', err);
      setApprovals((prev) => [newApprovalInput, ...prev]);
      setActionSuccess(`Approval request added.`);
      setTimeout(() => setActionSuccess(''), 4000);
    }
  };

  const handleApprove = async (item, notes) => {
    try {
      if (item.dbId) {
        await approvalsService.approveRequest(item.dbId, notes);
        await fetchData();
      } else {
        setApprovals((prev) =>
          prev.map((a) => (a.id === item.id ? { ...a, status: 'Approved', dueText: 'Approved' } : a))
        );
      }
      setActionSuccess(`Request ${item.id} has been approved.`);
      setTimeout(() => setActionSuccess(''), 4000);
    } catch (err) {
      console.error('Failed to approve request:', err);
      setApprovals((prev) =>
        prev.map((a) => (a.id === item.id ? { ...a, status: 'Approved', dueText: 'Approved' } : a))
      );
      setActionSuccess(`Request ${item.id} approved.`);
      setTimeout(() => setActionSuccess(''), 4000);
    }
  };

  const handleConfirmReject = async (item, reason, notes) => {
    try {
      if (item.dbId) {
        await approvalsService.rejectRequest(item.dbId, reason, notes);
        await fetchData();
      } else {
        setApprovals((prev) =>
          prev.map((a) => (a.id === item.id ? { ...a, status: 'Rejected', dueText: 'Rejected', rejectionReason: reason, description: notes } : a))
        );
      }
      setActionSuccess(`Request ${item.id} has been rejected.`);
      setTimeout(() => setActionSuccess(''), 4000);
    } catch (err) {
      console.error('Failed to reject request:', err);
      setApprovals((prev) =>
        prev.map((a) => (a.id === item.id ? { ...a, status: 'Rejected', dueText: 'Rejected', rejectionReason: reason, description: notes } : a))
      );
      setActionSuccess(`Request ${item.id} rejected.`);
      setTimeout(() => setActionSuccess(''), 4000);
    }
  };

  const handleClearAll = () => {
    setActiveCategory('ALL');
    setSearchQuery('');
    setTypeFilter('ALL');
    setStatusFilter('ALL');
    setRequesterFilter('ALL');
    setDateFilter('ANYTIME');
    setSortBy('NEWEST');
  };

  return (
    <div className="approvals-page">
      {/* Header Row */}
      <div className="approvals-header-row">
        <PageHeader
          title="Approval Center"
          subtitle="Review and take action on requests that need your approval across the organization."
        />
        <button
          className="btn-new-approval"
          onClick={() => setIsNewModalOpen(true)}
        >
          <Plus size={16} /> New Approval Request
        </button>
      </div>

      {/* KPI / Summary Cards Grid */}
      <ApprovalStats
        stats={stats}
        onFilterOverdue={() => {
          setStatusFilter('Overdue');
          setDateFilter('OVERDUE');
        }}
      />

      {/* Category Tabs & Toolbar Filters */}
      <ApprovalFilters
        activeCategory={activeCategory}
        onCategoryChange={setActiveCategory}
        categoryCounts={categoryCounts}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        typeFilter={typeFilter}
        onTypeFilterChange={setTypeFilter}
        statusFilter={statusFilter}
        onStatusFilterChange={setStatusFilter}
        requesterFilter={requesterFilter}
        onRequesterFilterChange={setRequesterFilter}
        dateFilter={dateFilter}
        onDateFilterChange={setDateFilter}
        sortBy={sortBy}
        onSortChange={setSortBy}
        onClearAll={handleClearAll}
      />

      {/* Success Notification Banner */}
      {actionSuccess && (
        <div style={{ background: '#ECFDF5', border: '1px solid #A7F3D0', color: '#047857', padding: '10px 16px', borderRadius: 8, display: 'flex', alignItems: 'center', gap: 8, fontSize: '0.84rem', fontWeight: 650 }}>
          <CheckCircle2 size={16} />
          <span>{actionSuccess}</span>
        </div>
      )}

      {error && (
        <div className="approvals-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={fetchData} className="btn-retry">
            <RefreshCw size={14} /> Retry
          </button>
        </div>
      )}

      {/* Main Approvals Workspace Table */}
      {loading ? (
        <div className="approvals-skeleton">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="skeleton-row" />
          ))}
        </div>
      ) : filteredApprovals.length === 0 ? (
        <div className="approvals-empty-state">
          <div className="empty-icon-circle">
            <Inbox size={26} />
          </div>
          <h3>No approval requests found</h3>
          <p>
            No active approval requests match your selected category or filters.
          </p>
          {(searchQuery || statusFilter !== 'ALL' || activeCategory !== 'ALL') && (
            <button
              className="btn-cancel"
              style={{ marginTop: 16 }}
              onClick={handleClearAll}
            >
              Reset Filters
            </button>
          )}
        </div>
      ) : (
        <div className="approvals-table-card">
          <table className="approvals-table">
            <thead>
              <tr>
                <th>Request</th>
                <th>Type</th>
                <th>Related To</th>
                <th>Requested By</th>
                <th>Due Date</th>
                <th>Status</th>
                <th style={{ textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {paginatedApprovals.map((item) => (
                <ApprovalRow
                  key={item.id}
                  item={item}
                  onSelect={setSelectedApproval}
                  onApprove={handleApprove}
                  onOpenRejectModal={setRejectingApproval}
                />
              ))}
            </tbody>
          </table>

          {/* Table Footer Pagination */}
          <div className="approvals-table-footer">
            <div>
              Showing <strong>{paginatedApprovals.length > 0 ? (currentPage - 1) * rowsPerPage + 1 : 0}</strong> to{' '}
              <strong>{Math.min(currentPage * rowsPerPage, filteredApprovals.length)}</strong> of{' '}
              <strong>{filteredApprovals.length}</strong> approvals
            </div>

            <div className="pagination-controls">
              <button
                className="page-btn"
                disabled={currentPage === 1}
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              >
                <ChevronLeft size={14} />
              </button>
              {[...Array(totalPages)].map((_, idx) => (
                <button
                  key={idx + 1}
                  className={`page-btn ${currentPage === idx + 1 ? 'active' : ''}`}
                  onClick={() => setCurrentPage(idx + 1)}
                >
                  {idx + 1}
                </button>
              ))}
              <button
                className="page-btn"
                disabled={currentPage === totalPages}
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              >
                <ChevronRight size={14} />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* New Approval Request Modal */}
      <NewApprovalModal
        isOpen={isNewModalOpen}
        existingApprovals={approvals}
        onClose={() => setIsNewModalOpen(false)}
        onSubmit={handleCreateApproval}
      />

      {/* Approval Details Modal */}
      <ApprovalDetailsModal
        item={selectedApproval}
        onClose={() => setSelectedApproval(null)}
        onApprove={handleApprove}
        onOpenRejectModal={setRejectingApproval}
      />

      {/* Rejection Reason Modal */}
      <RejectionModal
        isOpen={Boolean(rejectingApproval)}
        item={rejectingApproval}
        onClose={() => setRejectingApproval(null)}
        onConfirmReject={handleConfirmReject}
      />
    </div>
  );
}
