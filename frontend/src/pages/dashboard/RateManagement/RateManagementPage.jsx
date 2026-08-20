import React, { useState } from 'react';
import { Search, RefreshCw, ArrowRight, ShieldCheck, Clock, DollarSign, Layers } from 'lucide-react';
import api from '../../../services/api';
import PageHeader from '../../../components/dashboard/PageHeader';
import './RateManagementPage.css';

const POPULAR_LANES = [
  { origin: 'INNSA', originName: 'Nhava Sheva, India', dest: 'DEHAM', destName: 'Hamburg, Germany' },
  { origin: 'INNSA', originName: 'Nhava Sheva, India', dest: 'USLAX', destName: 'Los Angeles, USA' },
  { origin: 'CNSHA', originName: 'Shanghai, China', dest: 'INNSA', destName: 'Nhava Sheva, India' },
  { origin: 'SGSIN', originName: 'Singapore', dest: 'NLRTM', destName: 'Rotterdam, Netherlands' },
];

export default function RateManagementPage() {
  const [origin, setOrigin] = useState('INNSA');
  const [destination, setDestination] = useState('DEHAM');
  const [equipment, setEquipment] = useState('40GP');
  const [incoterms, setIncoterms] = useState('FOB');
  const [rates, setRates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [error, setError] = useState(null);

  const handleSearch = async (e) => {
    if (e) e.preventDefault();
    if (!origin.trim() || !destination.trim()) return;

    setLoading(true);
    setError(null);
    setSearched(true);
    try {
      const q = new URLSearchParams({
        origin: origin.trim().toUpperCase(),
        destination: destination.trim().toUpperCase(),
        equipment,
        incoterms,
      });
      const res = await api.get(`/api/v1/rates/search?${q.toString()}`);
      const data = res?.rates || res?.data?.rates || (Array.isArray(res) ? res : []);
      setRates(data);
    } catch (err) {
      console.error('Rate search error:', err);
      setError(err?.message || 'Unable to retrieve rate intelligence. Please try again.');
      setRates([]);
    } finally {
      setLoading(false);
    }
  };

  const setLane = (orig, dest) => {
    setOrigin(orig);
    setDestination(dest);
  };

  return (
    <div className="rate-mgmt-page">
      <PageHeader
        title="Rate Management & Intelligence"
        subtitle="Search live carrier spot benchmarks, contract rate cards, and trade lane tariffs"
      />

      {/* ── Search Bar Card ── */}
      <div className="rate-search-card">
        <form className="rate-search-form" onSubmit={handleSearch}>
          <div className="rate-input-group">
            <label>Origin Port / Code</label>
            <input
              type="text"
              placeholder="e.g. INNSA, Nhava Sheva"
              value={origin}
              onChange={(e) => setOrigin(e.target.value)}
              required
            />
          </div>

          <div className="rate-input-group">
            <label>Destination Port / Code</label>
            <input
              type="text"
              placeholder="e.g. DEHAM, Hamburg"
              value={destination}
              onChange={(e) => setDestination(e.target.value)}
              required
            />
          </div>

          <div className="rate-input-group narrow">
            <label>Equipment</label>
            <select value={equipment} onChange={(e) => setEquipment(e.target.value)}>
              <option value="20GP">20' Standard (20GP)</option>
              <option value="40GP">40' Standard (40GP)</option>
              <option value="40HC">40' High Cube (40HC)</option>
              <option value="45HC">45' High Cube (45HC)</option>
              <option value="REEFER">Reefer Container</option>
            </select>
          </div>

          <div className="rate-input-group narrow">
            <label>Incoterms</label>
            <select value={incoterms} onChange={(e) => setIncoterms(e.target.value)}>
              <option value="FOB">FOB (Free On Board)</option>
              <option value="CIF">CIF (Cost, Insurance & Freight)</option>
              <option value="EXW">EXW (Ex Works)</option>
              <option value="DDP">DDP (Delivered Duty Paid)</option>
            </select>
          </div>

          <button type="submit" className="btn-rate-search" disabled={loading}>
            {loading ? <RefreshCw className="spin" size={16} /> : <Search size={16} />}
            Search Rates
          </button>
        </form>

        {/* Quick Lane Chips */}
        <div className="quick-lanes-row">
          <span className="quick-lanes-label">Popular Lanes:</span>
          {POPULAR_LANES.map((lane, idx) => (
            <button
              key={idx}
              type="button"
              className="quick-lane-btn"
              onClick={() => setLane(lane.origin, lane.dest)}
            >
              {lane.origin} → {lane.dest}
            </button>
          ))}
        </div>
      </div>

      {/* ── Error Banner ── */}
      {error && (
        <div className="rate-error-banner">
          <span>⚠️ {error}</span>
          <button type="button" onClick={handleSearch} className="btn-retry">
            Retry
          </button>
        </div>
      )}

      {/* ── Results Area ── */}
      {loading ? (
        <div className="rate-loading-skeleton">
          <div className="skeleton-header" />
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton-row-card" />
          ))}
        </div>
      ) : searched && rates.length > 0 ? (
        <div className="rate-results-wrapper">
          <div className="results-count-bar">
            <strong>{rates.length}</strong> Rate option(s) found for <strong>{origin} → {destination}</strong> ({equipment})
          </div>

          <div className="rate-cards-grid">
            {rates.map((r, idx) => (
              <div key={r.id || idx} className="rate-result-card">
                <div className="rate-card-header">
                  <div className="carrier-badge-group">
                    <span className="carrier-scac-tag">{r.carrier_scac || r.carrier_name || 'CARRIER'}</span>
                    <span className={`rate-source-tag ${r.source?.toLowerCase()}`}>
                      {r.source === 'CONTRACT_PDF' ? 'Verified Contract' : 'Live Spot Index'}
                    </span>
                  </div>
                  <div className="rate-price-block">
                    <span className="price-currency">{r.currency || 'USD'}</span>
                    <span className="price-amount">${Number(r.total_rate || r.ocean_freight || 0).toLocaleString()}</span>
                  </div>
                </div>

                <div className="rate-card-lane">
                  <span>{r.origin_port || origin}</span>
                  <ArrowRight size={14} className="lane-arrow" />
                  <span>{r.destination_port || destination}</span>
                </div>

                <div className="rate-meta-row">
                  <div className="meta-item">
                    <Clock size={13} />
                    <span>{r.transit_days ? `${r.transit_days} days` : 'Est. 18-24 days'}</span>
                  </div>
                  <div className="meta-item">
                    <ShieldCheck size={13} />
                    <span>Valid until: {r.valid_to ? new Date(r.valid_to).toLocaleDateString() : 'End of Month'}</span>
                  </div>
                  <div className="meta-item">
                    <Layers size={13} />
                    <span>{r.equipment_type || equipment}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : searched ? (
        <div className="rate-empty-state">
          <div className="empty-icon-box">📊</div>
          <h3>No matching rates found for {origin} → {destination}</h3>
          <p>
            No contract or spot tariff is currently filed for this lane. You can upload carrier contract rate sheets in Contracts or request a spot quote.
          </p>
        </div>
      ) : (
        <div className="rate-empty-state initial">
          <div className="empty-icon-box">🔍</div>
          <h3>Search Freight Rate Intelligence</h3>
          <p>
            Enter origin and destination UN/LOCODEs above to query your organization's contracted rate sheets and real-time spot indices.
          </p>
        </div>
      )}
    </div>
  );
}
