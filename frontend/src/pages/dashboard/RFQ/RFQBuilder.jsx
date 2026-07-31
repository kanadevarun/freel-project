import React, { useState } from 'react';
import toast from 'react-hot-toast';
import { rfqService } from '../../../services/rfqService';
import './RFQPage.css';

/**
 * RFQBuilder is a modal component for Smart AI Extraction.
 */
export default function RFQBuilder({ onClose, onSuccess }) {
  const [rawText, setRawText] = useState('');
  const [isExtracting, setIsExtracting] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // Extracted Data
  const [extractedData, setExtractedData] = useState(null);
  
  // Form State (initially populated from AI)
  const [formData, setFormData] = useState({
    customer_id: '',
    origin: '',
    destination: '',
    incoterms: '',
  });

  const handleExtract = async () => {
    if (!rawText.trim()) {
      toast.error('Please paste a shipment request first');
      return;
    }

    setIsExtracting(true);
    try {
      const response = await rfqService.parseShipmentRequest(rawText);
      const { data, confidence_score, missing_fields } = response.data;
      
      setExtractedData({
        confidence_score,
        missing_fields: missing_fields || [],
      });

      // Pre-fill form with extracted data
      setFormData(prev => ({
        ...prev,
        origin: data?.origin || '',
        destination: data?.destination || '',
        incoterms: data?.incoterms || '',
      }));
      
      toast.success('Successfully extracted details');
    } catch (error) {
      console.error('Extraction failed:', error);
      toast.error('Failed to extract details from text');
    } finally {
      setIsExtracting(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.customer_id) {
      toast.error('Customer ID is required');
      return;
    }

    setIsSubmitting(true);
    try {
      const payload = {
        customer_id: parseInt(formData.customer_id, 10),
        origin: formData.origin,
        destination: formData.destination,
        incoterms: formData.incoterms,
        items: [] // For MVP, we skip detailed line items in the builder
      };
      
      await rfqService.createRFQ(payload);
      toast.success('RFQ Created Successfully!');
      onSuccess();
    } catch (error) {
      console.error('Failed to create RFQ:', error);
      toast.error('Failed to create RFQ');
    } finally {
      setIsSubmitting(false);
    }
  };

  const getConfidenceClass = (score) => {
    if (score >= 80) return 'high';
    if (score >= 50) return 'medium';
    return 'low';
  };

  return (
    <div className="modal-backdrop">
      <div className="modal-container rfq-builder-modal">
        <div className="modal-header">
          <h2>Create RFQ</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>
        
        <div className="modal-body">
          {/* AI Smart Extraction Section */}
          <div className="ai-extract-section">
            <div className="ai-extract-header">
              <div className="ai-extract-title">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                  <polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline>
                  <line x1="12" y1="22.08" x2="12" y2="12"></line>
                </svg>
                Smart Extraction
              </div>
            </div>
            
            <textarea
              className="ai-textarea"
              placeholder="Paste customer email, WhatsApp message, or raw text here..."
              value={rawText}
              onChange={(e) => setRawText(e.target.value)}
            />
            
            <div className="ai-extract-actions">
              <button 
                type="button"
                className="btn-ai" 
                onClick={handleExtract}
                disabled={isExtracting || !rawText}
              >
                {isExtracting ? (
                  <>
                    <span className="spinner-small"></span> Parsing...
                  </>
                ) : (
                  <>✨ AI Parse</>
                )}
              </button>
            </div>

            {/* AI Results Banner */}
            {extractedData && (
              <div className="ai-results-banner">
                <div className={`ai-confidence ${getConfidenceClass(extractedData.confidence_score)}`}>
                  Confidence Score: {extractedData.confidence_score}%
                </div>
                
                {extractedData.missing_fields.length > 0 && (
                  <div className="ai-missing-fields">
                    <div className="ai-missing-title">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                        <line x1="12" y1="9" x2="12" y2="13"></line>
                        <line x1="12" y1="17" x2="12.01" y2="17"></line>
                      </svg>
                      Missing Fields
                    </div>
                    <div className="ai-missing-tags">
                      {extractedData.missing_fields.map((field, idx) => (
                        <span key={idx} className="ai-missing-tag">
                          {field}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Form Section */}
          <form onSubmit={handleSubmit} className="standard-form">
            <div className="form-group">
              <label>Customer ID *</label>
              <input 
                type="number" 
                className="form-control"
                value={formData.customer_id}
                onChange={e => setFormData({...formData, customer_id: e.target.value})}
                required
              />
            </div>
            
            <div className="form-row">
              <div className="form-group">
                <label>Origin</label>
                <input 
                  type="text" 
                  className="form-control"
                  value={formData.origin}
                  onChange={e => setFormData({...formData, origin: e.target.value})}
                />
              </div>
              <div className="form-group">
                <label>Destination</label>
                <input 
                  type="text" 
                  className="form-control"
                  value={formData.destination}
                  onChange={e => setFormData({...formData, destination: e.target.value})}
                />
              </div>
            </div>

            <div className="form-group">
              <label>Incoterms</label>
              <input 
                type="text" 
                className="form-control"
                value={formData.incoterms}
                onChange={e => setFormData({...formData, incoterms: e.target.value})}
                placeholder="e.g. FOB, CIF, EXW"
              />
            </div>
            
            <div className="modal-footer" style={{ marginTop: '24px' }}>
              <button type="button" className="btn-secondary" onClick={onClose}>Cancel</button>
              <button type="submit" className="btn-primary" disabled={isSubmitting}>
                {isSubmitting ? 'Saving...' : 'Create RFQ'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
