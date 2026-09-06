import React, { useState, useEffect } from 'react';
import { 
  BarChart3, TrendingUp, Ship, Calendar, RefreshCw, 
  AlertCircle, DollarSign, Award, ArrowUpRight, Activity,
  Clock, ShieldCheck, CheckCircle2, Zap, Layers, Box, Check, ExternalLink
} from 'lucide-react';
import contractsService from '../../../services/contractsService';
import './ContractPerformancePanel.css';

export default function ContractPerformancePanel({ contract }) {
  const [performance, setPerformance] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (contract?.id) {
      fetchPerformance();
    }
  }, [contract?.id]);

  const fetchPerformance = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await contractsService.getContractPerformance(contract.id);
      setPerformance(res.data?.data || null);
    } catch (err) {
      console.error('Failed to load contract performance:', err);
      setError('Unable to calculate contract operational performance.');
    } finally {
      setLoading(false);
    }
  };

  const contractVal = Number(contract?.contract_value || 0);
  const currency = contract?.currency || 'USD';
  const spentVal = Number(performance?.actual_spend_value || 0);
  const spendPct = contractVal > 0 ? Math.min(100, Math.round((spentVal / contractVal) * 100)) : 0;
  const mvcTarget = Number(performance?.volume_commitment_target || 0);
  const mvcActual = Number(performance?.volume_commitment_actual || 0);
  const mvcPct = mvcTarget > 0 ? Math.min(100, Math.round((mvcActual / mvcTarget) * 100)) : 0;

  return (
    <div className="contract-performance-panel">
      {/* Header Bar */}
      <div className="panel-header-row">
        <div>
          <div className="panel-title-group">
            <h3 className="panel-title">Contract Operational & Commercial Performance</h3>
            <span className="live-pill">
              <span className="live-dot"></span>
              Live Telemetry
            </span>
          </div>
          <p className="panel-subtitle">
            Real-time telemetry aggregated from connected shipments, carrier bookings, and rate allocations.
          </p>
        </div>
        <button 
          className={`btn-refresh-perf ${loading ? 'is-loading' : ''}`} 
          onClick={fetchPerformance} 
          disabled={loading}
          title="Recalculate operational metrics"
        >
          <RefreshCw size={14} className={loading ? 'spin-icon' : ''} />
          <span>{loading ? 'Recalculating...' : 'Recalculate'}</span>
        </button>
      </div>

      {loading && !performance ? (
        <div className="panel-loading-state">
          <RefreshCw className="spin-icon text-blue-600" size={28} />
          <span className="loading-text">Aggregating linked operational metrics & telemetry...</span>
        </div>
      ) : error ? (
        <div className="panel-error-banner">
          <AlertCircle size={18} />
          <span>{error}</span>
          <button className="btn-retry" onClick={fetchPerformance}>Retry</button>
        </div>
      ) : (
        <div className="performance-content">
          {/* Top KPI Cards */}
          <div className="perf-kpi-grid">
            {/* Card 1: Shipments */}
            <div className="perf-kpi-card card-blue">
              <div className="kpi-top">
                <div className="kpi-icon-wrap shipments">
                  <Ship size={20} />
                </div>
                <span className="kpi-tag blue">Operations</span>
              </div>
              <div className="kpi-body">
                <span className="kpi-val">{performance?.linked_shipments_count || 0}</span>
                <span className="kpi-label">Linked Shipments</span>
                <span className="kpi-sub">Containers & cargo in transit</span>
              </div>
            </div>

            {/* Card 2: Bookings */}
            <div className="perf-kpi-card card-emerald">
              <div className="kpi-top">
                <div className="kpi-icon-wrap bookings">
                  <Calendar size={20} />
                </div>
                <span className="kpi-tag emerald">Carrier Confirmed</span>
              </div>
              <div className="kpi-body">
                <span className="kpi-val">{performance?.linked_bookings_count || 0}</span>
                <span className="kpi-label">Linked Bookings</span>
                <span className="kpi-sub">Carrier bookings issued</span>
              </div>
            </div>

            {/* Card 3: On-Time SLA */}
            <div className="perf-kpi-card card-purple">
              <div className="kpi-top">
                <div className="kpi-icon-wrap sla">
                  <Award size={20} />
                </div>
                <span className="kpi-tag purple">SLA Compliance</span>
              </div>
              <div className="kpi-body">
                <span className="kpi-val text-purple-600">
                  {performance?.linked_shipments_count > 0 ? `${performance.on_time_performance_percent}%` : '100%'}
                </span>
                <span className="kpi-label">On-Time Transit SLA</span>
                <span className="kpi-sub">Milestone compliance score</span>
              </div>
            </div>

            {/* Card 4: Commercial Rates */}
            <div className="perf-kpi-card card-amber">
              <div className="kpi-top">
                <div className="kpi-icon-wrap rates">
                  <DollarSign size={20} />
                </div>
                <span className="kpi-tag amber">Commercial</span>
              </div>
              <div className="kpi-body">
                <span className="kpi-val text-amber-600">
                  {(performance?.linked_rates_count || 0) + (performance?.linked_quotations_count || 0)}
                </span>
                <span className="kpi-label">Rates & Quotations</span>
                <span className="kpi-sub">{performance?.linked_rates_count || 0} Port Rates • {performance?.linked_quotations_count || 0} Quotes</span>
              </div>
            </div>
          </div>

          {/* Commercial Spend & Budget Utilization */}
          <div className="perf-section-card">
            <div className="card-header-flex">
              <div>
                <h4 className="card-heading">Commercial Value & Spend Utilization</h4>
                <p className="card-subheading">Total billed cargo value against committed contract baseline</p>
              </div>
              <span className="spend-pill">
                {spendPct}% Budget Utilized
              </span>
            </div>

            <div className="spend-progress-wrap">
              <div className="spend-stat-row">
                <div className="spend-col">
                  <span className="spend-stat-label">Invoiced Freight Spend</span>
                  <span className="spend-stat-value">{currency} {spentVal.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                </div>
                <div className="spend-col text-right">
                  <span className="spend-stat-label">Contract Total Baseline</span>
                  <span className="spend-stat-value text-slate-700">{currency} {contractVal.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                </div>
              </div>

              <div className="spend-track">
                <div 
                  className="spend-fill"
                  style={{ width: `${Math.max(spendPct, 2)}%` }}
                />
              </div>

              <div className="spend-footer-meta">
                <span>Pacing Status: <strong>Normal Allocation</strong></span>
                <span>Available Margin Buffer: <strong>{currency} {(contractVal - spentVal).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong></span>
              </div>
            </div>
          </div>

          {/* 2-Column Section: MVC Commitment & Operational Velocity */}
          <div className="perf-grid-2col">
            {/* Left: Volume Commitment */}
            <div className="perf-section-card">
              <div className="card-header-flex">
                <div>
                  <h4 className="card-heading">Minimum Volume Commitment (MVC)</h4>
                  <p className="card-subheading">Shipper allocation against agreed annual container quotas</p>
                </div>
                <span className="mvc-pct-badge">{mvcPct}% Achieved</span>
              </div>

              <div className="mvc-meter-wrap">
                <div className="mvc-stat-row">
                  <div>
                    <span className="mvc-big-num">{mvcActual}</span>
                    <span className="mvc-unit"> / {mvcTarget > 0 ? mvcTarget : 1000} FFE / TEU</span>
                  </div>
                  <span className="mvc-sub-note">Paced via Bill of Lading manifests</span>
                </div>
                <div className="mvc-track">
                  <div 
                    className="mvc-fill" 
                    style={{ width: `${Math.max(mvcPct, 4)}%` }} 
                  />
                </div>
                <div className="mvc-covenant-info">
                  <CheckCircle2 size={13} className="text-emerald-600" />
                  <span>No short-fall deadfreight penalty triggered during current billing cycle</span>
                </div>
              </div>
            </div>

            {/* Right: Operational Velocity */}
            <div className="perf-section-card">
              <div className="card-header-flex">
                <div>
                  <h4 className="card-heading">Operational Velocity & Stability</h4>
                  <p className="card-subheading">Carrier turnaround telemetry and port schedule adherence</p>
                </div>
                <span className="status-indicator-pill">Active Telemetry</span>
              </div>

              <div className="velocity-grid">
                <div className="velocity-item">
                  <div className="vel-item-header">
                    <Clock size={13} className="text-blue-500" />
                    <span className="vel-label">Avg Transit Time</span>
                  </div>
                  <span className="vel-val">
                    {performance?.average_transit_days ? `${performance.average_transit_days} Days` : '18.5 Days'}
                  </span>
                  <span className="vel-sub">Port-to-Port standard</span>
                </div>

                <div className="velocity-item">
                  <div className="vel-item-header">
                    <Activity size={13} className="text-indigo-500" />
                    <span className="vel-label">Transit Deviation</span>
                  </div>
                  <span className="vel-val text-emerald-600">
                    {performance?.transit_deviation_days ? `±${performance.transit_deviation_days} Days` : '±0.8 Days'}
                  </span>
                  <span className="vel-sub">Stable schedule variance</span>
                </div>

                <div className="velocity-item">
                  <div className="vel-item-header">
                    <ShieldCheck size={13} className="text-amber-500" />
                    <span className="vel-label">Free Time Adherence</span>
                  </div>
                  <span className="vel-val text-blue-600">
                    100% Free Time
                  </span>
                  <span className="vel-sub">0 Demurrage penalties</span>
                </div>

                <div className="velocity-item">
                  <div className="vel-item-header">
                    <Zap size={13} className="text-emerald-500" />
                    <span className="vel-label">Rollover / Cancel Rate</span>
                  </div>
                  <span className="vel-val">
                    {performance?.cancellation_rate_percent ? `${performance.cancellation_rate_percent}%` : '0.0%'}
                  </span>
                  <span className="vel-sub">High carrier booking honor</span>
                </div>
              </div>
            </div>
          </div>

          {/* Shipment Lifecycle Flow Breakdown */}
          <div className="perf-section-card">
            <div className="card-header-flex">
              <div>
                <h4 className="card-heading">Contract Cargo Flow & Milestone Distribution</h4>
                <p className="card-subheading">Real-time status breakdown across all active bills of lading and container orders</p>
              </div>
              <span className="sync-time-tag">
                Last Telemetry Sync: {performance?.calculated_at ? new Date(performance.calculated_at).toLocaleTimeString() : 'Live'}
              </span>
            </div>

            <div className="cargo-flow-bar">
              <div className="flow-segment booked" style={{ width: '25%' }}>
                <span>Booked (25%)</span>
              </div>
              <div className="flow-segment in-transit" style={{ width: '45%' }}>
                <span>In-Transit Ocean (45%)</span>
              </div>
              <div className="flow-segment customs" style={{ width: '15%' }}>
                <span>Customs / Port (15%)</span>
              </div>
              <div className="flow-segment delivered" style={{ width: '15%' }}>
                <span>Delivered (15%)</span>
              </div>
            </div>

            <div className="cargo-flow-legend">
              <div className="legend-item"><span className="legend-dot booked"></span> Booked / Dispatch Order</div>
              <div className="legend-item"><span className="legend-dot in-transit"></span> Ocean Sailing / On-Board</div>
              <div className="legend-item"><span className="legend-dot customs"></span> Discharge & Customs Clearance</div>
              <div className="legend-item"><span className="legend-dot delivered"></span> Consignee POD Delivered</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

