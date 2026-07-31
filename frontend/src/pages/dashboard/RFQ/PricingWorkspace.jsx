import React, { useState } from 'react';
import toast from 'react-hot-toast';
import { rfqService } from '../../../services/rfqService';

import AgentStatusBadge from '../../../components/agent/AgentStatusBadge';

/**
 * PricingWorkspace is the dedicated screen for the Pricing team to evaluate and submit quotes.
 */
export default function PricingWorkspace({ rfq, quotes, onQuoteSubmitted }) {
  const [carrierName, setCarrierName] = useState('');
  const [buyPrice, setBuyPrice] = useState('');
  const [sellPrice, setSellPrice] = useState('');
  const [transitTime, setTransitTime] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!carrierName || !buyPrice || !sellPrice) {
      toast.error('Please fill in all required fields');
      return;
    }

    setIsSubmitting(true);
    try {
      const payload = {
        carrier_name: carrierName,
        buy_price: parseFloat(buyPrice),
        sell_price: parseFloat(sellPrice),
        transit_time_days: transitTime ? parseInt(transitTime, 10) : null,
      };
      
      await rfqService.addQuote(rfq.id, payload);
      toast.success('Quote submitted successfully');
      setCarrierName('');
      setBuyPrice('');
      setSellPrice('');
      setTransitTime('');
      onQuoteSubmitted();
    } catch (error) {
      console.error('Failed to submit quote:', error);
      toast.error('Failed to submit quote');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="pricing-workspace">
      <div className="pricing-header">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h3>Carrier Comparison Engine</h3>
            <p>Evaluate routing options and submit pricing quotes for this shipment.</p>
          </div>
          <AgentStatusBadge status={rfq?.agent_status} />
        </div>
      </div>

      {/* Submitted Quotes / Comparisons */}
      <div className="carrier-options-list">
        {quotes && quotes.length > 0 ? (
          quotes.map(quote => (
            <div key={quote.id} className={`carrier-card ${quote.is_recommended ? 'recommended' : ''}`}>
              <div className="carrier-card-header">
                <div className="carrier-name">{quote.carrier_name}</div>
                {quote.is_recommended && <div className="badge-ai">✨ AI Drafted</div>}
              </div>
              
              {quote.ai_reasoning && (
                <div className="ai-reasoning-box">
                  <strong>AI Note:</strong> {quote.ai_reasoning}
                </div>
              )}

              <div className="carrier-metrics">
                <div className="metric">
                  <span className="label">Transit</span>
                  <span className="value">{quote.transit_time_days || '-'} Days</span>
                </div>
                <div className="metric">
                  <span className="label">Buy Rate</span>
                  <span className="value">${quote.buy_price.toFixed(2)}</span>
                </div>
                <div className="metric highlight">
                  <span className="label">Sell Rate</span>
                  <span className="value">${quote.sell_price.toFixed(2)}</span>
                </div>
                <div className="metric">
                  <span className="label">Margin</span>
                  <span className="value text-green">
                    {(((quote.sell_price - quote.buy_price) / quote.sell_price) * 100).toFixed(1)}%
                  </span>
                </div>
                <div className="metric">
                  <span className="label">Reliability</span>
                  <span className="value">{quote.reliability_score || 0}/100</span>
                </div>
                <div className="metric">
                  <span className="label">Success Rate</span>
                  <span className="value">{quote.historical_success_rate ? `${quote.historical_success_rate}%` : 'N/A'}</span>
                </div>
              </div>
            </div>
          ))
        ) : (
          <div className="empty-quotes">
            No quotes submitted yet. Be the first to price this route.
          </div>
        )}
      </div>

      <hr className="divider" />

      {/* Add New Quote Form */}
      <form onSubmit={handleSubmit} className="pricing-form">
        <h4>Add Carrier Option</h4>
        <div className="form-row">
          <div className="form-group">
            <label>Carrier Name</label>
            <input 
              type="text" 
              className="form-control" 
              placeholder="e.g. MAERSK, MSC"
              value={carrierName}
              onChange={e => setCarrierName(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label>Transit Time (Days)</label>
            <input 
              type="number" 
              className="form-control" 
              placeholder="e.g. 14"
              value={transitTime}
              onChange={e => setTransitTime(e.target.value)}
            />
          </div>
        </div>
        
        <div className="form-row">
          <div className="form-group">
            <label>Buy Price ($)</label>
            <input 
              type="number" 
              step="0.01"
              className="form-control" 
              value={buyPrice}
              onChange={e => setBuyPrice(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label>Sell Price ($)</label>
            <input 
              type="number" 
              step="0.01"
              className="form-control" 
              value={sellPrice}
              onChange={e => setSellPrice(e.target.value)}
            />
          </div>
        </div>

        {/* AI Margin Suggestion (Mock for MVP) */}
        {buyPrice && (
          <div className="ai-suggestion-inline">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
            </svg>
            AI Suggestion: Target an 18% margin. Recommended sell price: ${(parseFloat(buyPrice) / 0.82).toFixed(2)}
          </div>
        )}

        <button type="submit" className="btn-primary" disabled={isSubmitting} style={{ marginTop: '16px' }}>
          {isSubmitting ? 'Submitting...' : 'Submit Quote Option'}
        </button>
      </form>
    </div>
  );
}
