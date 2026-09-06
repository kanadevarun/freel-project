import React, { useState } from 'react';
import {
  AlertTriangle,
  Clock,
  RefreshCw,
  ChevronRight,
  ShieldAlert,
  ArrowRight,
  Layers,
  Sparkles,
  ChevronDown,
  ChevronUp,
  CheckCircle2,
  Building2,
  ArrowUpDown,
} from 'lucide-react';
import './ContractAttentionPanel.css';

export default function ContractAttentionPanel({
  items = [],
  loading = false,
  onEvaluate,
  onSelectContract,
}) {
  const [collapsed, setCollapsed] = useState(false);
  const [filterSeverity, setFilterSeverity] = useState('ALL'); // 'ALL' | 'CRITICAL' | 'WARNING'
  const [sortBy, setSortBy] = useState('CRITICALITY'); // 'CRITICALITY' | 'DAYS_REMAINING' | 'REFERENCE' | 'PARTY' | 'IMPACT'

  if (loading) {
    return (
      <div className="contract-attention-panel attention-loading">
        <div className="attention-loading-shimmer">
          <div className="shimmer-line" style={{ width: '30%' }} />
          <div className="shimmer-line" style={{ width: '70%' }} />
        </div>
      </div>
    );
  }

  // If no items require attention, show clean minimal state
  if (!items || items.length === 0) {
    return (
      <div className="contract-attention-panel attention-healthy">
        <div className="attention-healthy-content">
          <div className="attention-healthy-badge">
            <CheckCircle2 size={15} />
            <span>Commercial Portfolio Healthy</span>
          </div>
          <p className="attention-healthy-text">
            All active commercial agreements, SLAs, and attached rates are currently within valid terms with zero active risks.
          </p>
        </div>
        <button
          className="attention-evaluate-btn outline"
          onClick={onEvaluate}
          title="Run deterministic contract health evaluation"
        >
          <RefreshCw size={13} />
          <span>Re-evaluate</span>
        </button>
      </div>
    );
  }

  const criticalCount = items.filter((i) => i.severity === 'CRITICAL').length;
  const warningCount = items.filter((i) => i.severity === 'WARNING').length;

  const sortedAndFilteredItems = items
    .filter((item) => {
      if (filterSeverity === 'CRITICAL') return item.severity === 'CRITICAL';
      if (filterSeverity === 'WARNING') return item.severity === 'WARNING';
      return true;
    })
    .sort((a, b) => {
      if (sortBy === 'CRITICALITY') {
        const sevOrder = { CRITICAL: 1, WARNING: 2, NORMAL: 3 };
        const sevDiff = (sevOrder[a.severity] || 3) - (sevOrder[b.severity] || 3);
        if (sevDiff !== 0) return sevDiff;
        return (a.days_remaining || 0) - (b.days_remaining || 0);
      }
      if (sortBy === 'DAYS_REMAINING') {
        return (a.days_remaining || 0) - (b.days_remaining || 0);
      }
      if (sortBy === 'REFERENCE') {
        return (a.contract_reference || '').localeCompare(b.contract_reference || '');
      }
      if (sortBy === 'PARTY') {
        return (a.party_name || '').localeCompare(b.party_name || '');
      }
      if (sortBy === 'IMPACT') {
        return (b.linked_records_count || 0) - (a.linked_records_count || 0);
      }
      return 0;
    });

  return (
    <div className={`contract-attention-panel has-alerts ${criticalCount > 0 ? 'border-critical' : 'border-warning'}`}>
      <div className="attention-panel-header">
        <div className="attention-header-left">
          <div className={`attention-header-icon ${criticalCount > 0 ? 'icon-critical' : 'icon-warning'}`}>
            <ShieldAlert size={18} />
          </div>
          <div>
            <div className="attention-header-title-row">
              <h4 className="attention-header-title">
                Commercial Contracts Requiring Attention
              </h4>
              <button
                className={`attention-pill-btn count ${filterSeverity === 'ALL' ? 'active' : ''}`}
                onClick={() => setFilterSeverity('ALL')}
                title="Show all attention contracts"
              >
                {items.length} {items.length === 1 ? 'Agreement' : 'Agreements'}
              </button>
              {criticalCount > 0 && (
                <button
                  className={`attention-pill-btn critical ${filterSeverity === 'CRITICAL' ? 'active' : ''}`}
                  onClick={() => setFilterSeverity(filterSeverity === 'CRITICAL' ? 'ALL' : 'CRITICAL')}
                  title="Filter to critical contracts"
                >
                  {criticalCount} Critical
                </button>
              )}
              {warningCount > 0 && (
                <button
                  className={`attention-pill-btn warning ${filterSeverity === 'WARNING' ? 'active' : ''}`}
                  onClick={() => setFilterSeverity(filterSeverity === 'WARNING' ? 'ALL' : 'WARNING')}
                  title="Filter to expiring contracts"
                >
                  {warningCount} Expiries
                </button>
              )}
            </div>
            <p className="attention-header-sub">
              Deterministic monitoring of contracts with impending expiry, active rate dependencies, or renewal requirements.
            </p>
          </div>
        </div>

        <div className="attention-header-actions">
          <div className="attention-sort-wrap">
            <ArrowUpDown size={12} className="attention-sort-icon" />
            <select
              className="attention-sort-select"
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              title="Sort attention contracts"
            >
              <option value="CRITICALITY">Sort: Criticality (Urgency)</option>
              <option value="DAYS_REMAINING">Sort: Days Left (Ascending)</option>
              <option value="REFERENCE">Sort: Reference (A-Z)</option>
              <option value="PARTY">Sort: Counterparty (A-Z)</option>
              <option value="IMPACT">Sort: Linked Records (High-Low)</option>
            </select>
          </div>

          <button
            className="attention-evaluate-btn"
            onClick={onEvaluate}
            title="Re-run lifecycle intelligence evaluation"
          >
            <RefreshCw size={13} />
            <span>Re-evaluate</span>
          </button>
          <button
            className="attention-collapse-btn"
            onClick={() => setCollapsed(!collapsed)}
            title={collapsed ? 'Expand Attention Panel' : 'Collapse Attention Panel'}
          >
            {collapsed ? <ChevronDown size={16} /> : <ChevronUp size={16} />}
          </button>
        </div>
      </div>

      {!collapsed && (
        <div className="attention-items-grid">
          {sortedAndFilteredItems.map((item) => {
            const isCritical = item.severity === 'CRITICAL';
            const isExpired = item.condition === 'EXPIRED';
            const isRenewalInProgress = item.renewal_status === 'IN_PROGRESS';

            return (
              <div
                key={item.contract_id}
                className={`attention-card ${isCritical ? 'card-critical' : 'card-warning'}`}
                onClick={() => onSelectContract && onSelectContract(item.contract_id)}
              >
                <div className="attention-card-top">
                  <span className="attention-ref-tag">{item.contract_reference}</span>
                  <div className="attention-card-timing">
                    {isExpired ? (
                      <span className="timing-pill expired">
                        <Clock size={11} /> Expired ({Math.abs(item.days_remaining)}d ago)
                      </span>
                    ) : isRenewalInProgress ? (
                      <span className="timing-pill renewal">
                        <Sparkles size={11} /> Renewal Active
                      </span>
                    ) : (
                      <span className={`timing-pill ${item.days_remaining <= 7 ? 'critical' : 'warning'}`}>
                        <Clock size={11} /> {item.days_remaining}d remaining
                      </span>
                    )}
                  </div>
                </div>

                <div className="attention-card-body">
                  <div className="attention-party-row">
                    <Building2 size={12} className="attention-party-icon" />
                    <span className="attention-party-name">{item.party_name || 'Unassigned Counterparty'}</span>
                  </div>
                  <h5 className="attention-card-name">{item.contract_name}</h5>
                  <p className="attention-card-desc">{item.risk_description}</p>
                </div>

                <div className="attention-card-footer">
                  <div className="attention-footer-impact">
                    <Layers size={13} />
                    <span>{item.linked_records_count} linked records</span>
                  </div>
                  <span className="attention-inspect-link">
                    <span>Inspect</span>
                    <ArrowRight size={13} />
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
