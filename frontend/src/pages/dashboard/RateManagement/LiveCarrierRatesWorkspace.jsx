import React, { useState } from 'react';
import {
  Zap, Search, RotateCcw, TrendingUp, Clock, DollarSign,
  ShieldCheck, AlertCircle, ArrowRight, CheckCircle2, ChevronRight
} from 'lucide-react';
import { rateService } from '../../../services/rateService';
import toast from 'react-hot-toast';
import './LiveCarrierRatesWorkspace.css';

export default function LiveCarrierRatesWorkspace() {
  const [originPort, setOriginPort] = useState('INNSA');
  const [destinationPort, setDestinationPort] = useState('NLRTM');
  const [equipmentType, setEquipmentType] = useState('40HC');
  const [rateType, setRateType] = useState('ALL');
  const [carrierSCAC, setCarrierSCAC] = useState('');

  const [loading, setLoading] = useState(false);
  const [searchResult, setSearchResult] = useState(null);
  const [searched, setSearched] = useState(false);

  const handleSearch = async (e) => {
    if (e) e.preventDefault();
    if (!originPort || !destinationPort) {
      toast.error('Origin and Destination ports are required');
      return;
    }

    try {
      setLoading(true);
      const payload = {
        origin_port: originPort.trim().toUpperCase(),
        destination_port: destinationPort.trim().toUpperCase(),
        equipment_type: equipmentType,
        rate_type: rateType === 'ALL' ? undefined : rateType,
        carrier_scac: carrierSCAC ? carrierSCAC.trim().toUpperCase() : undefined,
      };

      const res = await rateService.searchCarrierLiveRates(payload);
      const data = res?.data?.data || res?.data || res;
      setSearchResult(data);
      setSearched(true);
      if (data?.rates?.length > 0) {
        toast.success(`Found ${data.rates.length} live carrier rates`);
      } else if (data?.message) {
        toast(data.message, { icon: 'ℹ️' });
      }
    } catch (err) {
      console.error('Failed to search carrier live rates:', err);
      toast.error(err.response?.data?.error?.message || 'Failed to search live carrier rates');
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    setOriginPort('INNSA');
    setDestinationPort('NLRTM');
    setEquipmentType('40HC');
    setRateType('ALL');
    setCarrierSCAC('');
    setSearchResult(null);
    setSearched(false);
  };

  const rates = searchResult?.rates || [];

  return (
    <div className="lcr-container">
      {/* Search Header Form */}
      <div className="lcr-search-card">
        <div className="lcr-search-header">
          <div className="lcr-search-title">
            <Zap size={20} color="#2563eb" />
            <span>Direct Carrier API Live Rates</span>
          </div>
          <span className="lcr-search-badge">
            <ShieldCheck size={12} /> DCSA Verified APIs
          </span>
        </div>

        <form className="lcr-form-grid" onSubmit={handleSearch}>
          <div className="lcr-field">
            <label>Origin Port (UN/LOCODE)</label>
            <input
              type="text"
              className="lcr-input"
              placeholder="e.g. INNSA"
              value={originPort}
              onChange={(e) => setOriginPort(e.target.value)}
              required
            />
          </div>

          <div className="lcr-field">
            <label>Destination Port (UN/LOCODE)</label>
            <input
              type="text"
              className="lcr-input"
              placeholder="e.g. NLRTM"
              value={destinationPort}
              onChange={(e) => setDestinationPort(e.target.value)}
              required
            />
          </div>

          <div className="lcr-field">
            <label>Equipment Type</label>
            <select
              className="lcr-input"
              value={equipmentType}
              onChange={(e) => setEquipmentType(e.target.value)}
            >
              <option value="40HC">40' High Cube (40HC)</option>
              <option value="20GP">20' General (20GP)</option>
              <option value="40GP">40' General (40GP)</option>
              <option value="45HC">45' High Cube (45HC)</option>
              <option value="20RF">20' Reefer (20RF)</option>
              <option value="40RF">40' Reefer (40RF)</option>
            </select>
          </div>

          <div className="lcr-field">
            <label>Rate Type</label>
            <select
              className="lcr-input"
              value={rateType}
              onChange={(e) => setRateType(e.target.value)}
            >
              <option value="ALL">All (Spot & Contract)</option>
              <option value="SPOT">Spot Rates Only</option>
              <option value="CONTRACT">Contract Rates Only</option>
            </select>
          </div>

          <div className="lcr-field">
            <label>Specific Carrier SCAC</label>
            <input
              type="text"
              className="lcr-input"
              placeholder="e.g. MAEU (Optional)"
              value={carrierSCAC}
              onChange={(e) => setCarrierSCAC(e.target.value)}
            />
          </div>

          <div style={{ display: 'flex', gap: '8px' }}>
            <button
              type="submit"
              className="lcr-btn-search"
              style={{ flex: 1 }}
              disabled={loading}
            >
              {loading ? <RotateCcw className="animate-spin" size={16} /> : <Search size={16} />}
              <span>{loading ? 'Querying...' : 'Live Search'}</span>
            </button>
            {searched && (
              <button
                type="button"
                className="lcr-btn-search"
                style={{ background: '#f1f5f9', color: '#475569', width: '40px', padding: 0 }}
                onClick={handleReset}
                title="Reset search"
              >
                <RotateCcw size={15} />
              </button>
            )}
          </div>
        </form>
      </div>

      {/* KPI Cards when search results exist */}
      {searched && searchResult && (
        <div className="lcr-kpi-grid">
          <div className="lcr-kpi-card">
            <div className="lcr-kpi-icon green">
              <DollarSign size={22} />
            </div>
            <div className="lcr-kpi-content">
              <span className="lcr-kpi-label">Cheapest Live Rate</span>
              <span className="lcr-kpi-val">
                {searchResult.cheapest_amount != null
                  ? `$${searchResult.cheapest_amount.toLocaleString()} USD`
                  : 'N/A'}
              </span>
              <span className="lcr-kpi-sub">
                {searchResult.cheapest_carrier || 'No rates found'}
              </span>
            </div>
          </div>

          <div className="lcr-kpi-card">
            <div className="lcr-kpi-icon blue">
              <Clock size={22} />
            </div>
            <div className="lcr-kpi-content">
              <span className="lcr-kpi-label">Fastest Transit</span>
              <span className="lcr-kpi-val">
                {searchResult.fastest_transit != null
                  ? `${searchResult.fastest_transit} Days`
                  : 'N/A'}
              </span>
              <span className="lcr-kpi-sub">
                {searchResult.fastest_carrier || 'No transit data'}
              </span>
            </div>
          </div>

          <div className="lcr-kpi-card">
            <div className="lcr-kpi-icon purple">
              <TrendingUp size={22} />
            </div>
            <div className="lcr-kpi-content">
              <span className="lcr-kpi-label">Total Carrier Offers</span>
              <span className="lcr-kpi-val">{searchResult.total_rates_count || 0}</span>
              <span className="lcr-kpi-sub">Normalized & compared</span>
            </div>
          </div>
        </div>
      )}

      {/* Results Table */}
      {searched && (
        <div className="lcr-results-card">
          <div className="lcr-results-header">
            <div style={{ fontWeight: 700, color: '#0f172a', fontSize: '0.95rem' }}>
              Live Carrier Rates ({searchResult?.origin_port} ➔ {searchResult?.destination_port})
            </div>
            <div style={{ fontSize: '0.8rem', color: '#64748b' }}>
              Equipment: <strong>{searchResult?.equipment_type}</strong>
            </div>
          </div>

          {rates.length > 0 ? (
            <div style={{ overflowX: 'auto' }}>
              <table className="lcr-table">
                <thead>
                  <tr>
                    <th>Shipping Line</th>
                    <th>Type</th>
                    <th>Base Ocean</th>
                    <th>Surcharges</th>
                    <th>Total Landed Buy</th>
                    <th>Transit / Free Days</th>
                    <th>Validity Window</th>
                  </tr>
                </thead>
                <tbody>
                  {rates.map((rate, idx) => (
                    <tr key={rate.id || idx}>
                      <td>
                        <div className="lcr-carrier-cell">
                          <span className="lcr-carrier-name">
                            {rate.carrier_name}
                            {rate.is_cheapest && (
                              <span className="lcr-tag cheapest">★ Cheapest</span>
                            )}
                            {rate.is_fastest && (
                              <span className="lcr-tag fastest">⚡ Fastest</span>
                            )}
                            {rate.is_best_value && (
                              <span className="lcr-tag value">👑 Best Value</span>
                            )}
                          </span>
                          <span className="lcr-scac-badge">{rate.carrier_scac}</span>
                        </div>
                      </td>
                      <td>
                        <span className={`lcr-type-badge ${rate.is_contract_rate ? 'contract' : 'spot'}`}>
                          {rate.is_contract_rate ? 'Contract Rate' : 'Live Spot'}
                        </span>
                      </td>
                      <td>
                        <div style={{ fontWeight: 600 }}>
                          {rate.currency} {rate.base_ocean_price?.toLocaleString()}
                        </div>
                      </td>
                      <td>
                        <div className="lcr-breakdown-sub">
                          Origin: {rate.currency} {rate.origin_surcharges?.toLocaleString() || 0}
                          <br />
                          Dest: {rate.currency} {rate.dest_surcharges?.toLocaleString() || 0}
                        </div>
                      </td>
                      <td>
                        <div className="lcr-price-val">
                          {rate.currency} {rate.total_buy_price?.toLocaleString()}
                        </div>
                      </td>
                      <td>
                        <div style={{ fontSize: '0.85rem', fontWeight: 600 }}>
                          {rate.transit_time_days ? `${rate.transit_time_days} days` : 'TBA'}
                        </div>
                        <div className="lcr-breakdown-sub">
                          Free Demurrage: {rate.free_days ? `${rate.free_days}d` : 'Standard'}
                        </div>
                      </td>
                      <td>
                        <div style={{ fontSize: '0.78rem', color: '#475569' }}>
                          {new Date(rate.valid_from).toLocaleDateString()} ➔{' '}
                          {new Date(rate.valid_until).toLocaleDateString()}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="lcr-empty">
              <div className="lcr-empty-icon">🚢</div>
              <div style={{ fontWeight: 700, fontSize: '1rem', color: '#0f172a', marginBottom: '6px' }}>
                No Live Carrier Rates Returned
              </div>
              <p style={{ fontSize: '0.88rem', maxWidth: '450px', margin: '0 auto' }}>
                {searchResult?.message ||
                  'No connected carriers returned active spot or contract tariffs for this port pair. Verify carrier credentials or try another route.'}
              </p>
            </div>
          )}
        </div>
      )}

      {/* Initial State before searching */}
      {!searched && (
        <div className="lcr-results-card" style={{ padding: '40px 24px', textAlign: 'center' }}>
          <div style={{ fontSize: '2.5rem', marginBottom: '12px' }}>⚡</div>
          <h3 style={{ fontSize: '1.1rem', fontWeight: 700, color: '#0f172a', marginBottom: '8px' }}>
            Live Multi-Carrier Rate Search & Benchmarking
          </h3>
          <p style={{ color: '#64748b', fontSize: '0.88rem', maxWidth: '500px', margin: '0 auto 20px' }}>
            Query connected shipping line APIs (Maersk, MSC, Hapag-Lloyd, CMA CGM) in real time. Compare landed rates, transit times, and demurrage allotments.
          </p>
          <button
            type="button"
            className="lcr-btn-search"
            style={{ margin: '0 auto', padding: '0 24px' }}
            onClick={() => handleSearch()}
          >
            <Search size={16} /> Run Benchmark Query (INNSA ➔ NLRTM)
          </button>
        </div>
      )}
    </div>
  );
}
