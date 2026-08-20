import React, { useState } from 'react';
import toast from 'react-hot-toast';
import { contractsService } from '../../../services/contractsService';

export default function ReviewModal({ item, onClose, onSuccess }) {
  let initialData = {};
  try {
    initialData = typeof item.extracted_data === 'string'
      ? JSON.parse(item.extracted_data)
      : item.extracted_data || {};
  } catch (e) {
    initialData = {};
  }

  // Editable Form fields
  const [origin, setOrigin] = useState(initialData.origin_port || '');
  const [destination, setDestination] = useState(initialData.destination_port || '');
  const [viaPort, setViaPort] = useState(initialData.via_port || '');
  const [serviceCode, setServiceCode] = useState(initialData.service_code || '');
  const [equipmentType, setEquipmentType] = useState(initialData.equipment_type || '40GP');
  const [oceanFreight, setOceanFreight] = useState(initialData.ocean_freight || 0);
  const [originCharges, setOriginCharges] = useState(initialData.origin_charges || 0);
  const [destinationCharges, setDestinationCharges] = useState(initialData.destination_charges || 0);
  const [validFrom, setValidFrom] = useState((initialData.valid_from || '').split('T')[0]);
  const [validUntil, setValidUntil] = useState((initialData.valid_until || '').split('T')[0]);
  
  // Custom surcharges editor
  const [surcharges, setSurcharges] = useState(initialData.surcharges || []);

  const [notes, setNotes] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleAddSurcharge = () => {
    setSurcharges([...surcharges, { code: '', description: '', amount: 0, unit: 'PER_TEU', included: false }]);
  };

  const handleRemoveSurcharge = (idx) => {
    setSurcharges(surcharges.filter((_, i) => i !== idx));
  };

  const handleSurchargeChange = (idx, field, value) => {
    const updated = surcharges.map((s, i) => {
      if (i === idx) {
        return { ...s, [field]: value };
      }
      return s;
    });
    setSurcharges(updated);
  };

  const handleApprove = async () => {
    setIsSubmitting(true);
    const apiToast = toast.loading('Ingesting rate into database...');
    try {
      // Re-compile corrected payload
      const correctedPayload = {
        ...initialData,
        origin_port: origin,
        destination_port: destination,
        via_port: viaPort || null,
        service_code: serviceCode || null,
        equipment_type: equipmentType,
        ocean_freight: parseFloat(oceanFreight),
        origin_charges: parseFloat(originCharges),
        destination_charges: parseFloat(destinationCharges),
        valid_from: validFrom ? `${validFrom}T00:00:00Z` : initialData.valid_from,
        valid_until: validUntil ? `${validUntil}T23:59:59Z` : initialData.valid_until,
        surcharges: surcharges.map(s => ({
          ...s,
          amount: parseFloat(s.amount)
        })),
        // Recalculate buy price
        total_buy_price: parseFloat(oceanFreight) + parseFloat(originCharges) + parseFloat(destinationCharges) + 
          surcharges.reduce((acc, curr) => acc + (curr.included ? 0 : parseFloat(curr.amount)), 0),
      };

      await contractsService.approveReviewItem(item.id, correctedPayload, notes);
      toast.success('Rate successfully corrected & ingested into live catalog.', { id: apiToast });
      onSuccess();
    } catch (err) {
      console.error(err);
      toast.error('Failed to approve/correct rate.', { id: apiToast });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReject = async () => {
    if (!notes.trim()) {
      toast.error('Please add notes explaining why this rate was rejected.');
      return;
    }
    setIsSubmitting(true);
    const apiToast = toast.loading('Marking rate as rejected...');
    try {
      await contractsService.rejectReviewItem(item.id, notes);
      toast.success('Rate successfully rejected.', { id: apiToast });
      onSuccess();
    } catch (err) {
      console.error(err);
      toast.error('Failed to reject rate.', { id: apiToast });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="review-modal-backdrop">
      <div className="review-modal-card">
        <div className="modal-header">
          <div>
            <h3>Verify Flagged Extraction</h3>
            <p className="modal-subtitle">Review anomalies flagged by LLM validator</p>
          </div>
          <button className="btn-modal-close" onClick={onClose}>✕</button>
        </div>

        <div className="modal-body-layout">
          {/* Left Panel: Original Extraction & AI Diagnostics */}
          <div className="modal-diagnostic-panel">
            <div className="anomaly-header">
              <span className="anomaly-tag">⚠️ AI Review Flags</span>
              <div className="flags-list">
                {(item.review_flags || []).map(f => (
                  <span key={f} className="flag-item">{f}</span>
                ))}
              </div>
            </div>

            <div className="diagnostic-section">
              <h5>AI reasoning details</h5>
              <p className="ai-reasoning-desc">{item.ai_reasoning}</p>
            </div>

            <div className="diagnostic-section">
              <h5>Source Text Extract (Page {item.source_page})</h5>
              <blockquote className="source-blockquote">
                "{item.source_text}"
              </blockquote>
            </div>

            <div className="diagnostic-section">
              <h5>Document Reference</h5>
              <div className="source-doc-ref">
                <span>📄 PDF Attachment</span>
                {item.source_image_url && (
                  <a href={item.source_image_url} target="_blank" rel="noreferrer" className="btn-link">
                    View Context Page →
                  </a>
                )}
              </div>
            </div>
          </div>

          {/* Right Panel: Inline Corrections Form */}
          <div className="modal-form-panel">
            <div className="form-section-title">Manual Corrections Editor</div>

            <div className="form-grid">
              <div className="form-group">
                <label>Origin Port (LOCODE)</label>
                <input
                  type="text"
                  value={origin}
                  onChange={(e) => setOrigin(e.target.value.toUpperCase())}
                  className="form-input"
                />
              </div>

              <div className="form-group">
                <label>Destination Port (LOCODE)</label>
                <input
                  type="text"
                  value={destination}
                  onChange={(e) => setDestination(e.target.value.toUpperCase())}
                  className="form-input"
                />
              </div>

              <div className="form-group">
                <label>Via Port</label>
                <input
                  type="text"
                  value={viaPort}
                  onChange={(e) => setViaPort(e.target.value.toUpperCase())}
                  className="form-input"
                  placeholder="e.g. SGSIN"
                />
              </div>

              <div className="form-group">
                <label>Service Code</label>
                <input
                  type="text"
                  value={serviceCode}
                  onChange={(e) => setServiceCode(e.target.value.toUpperCase())}
                  className="form-input"
                  placeholder="e.g. AE-1"
                />
              </div>

              <div className="form-group">
                <label>Equipment Type</label>
                <input
                  type="text"
                  value={equipmentType}
                  onChange={(e) => setEquipmentType(e.target.value)}
                  className="form-input"
                />
              </div>

              <div className="form-group">
                <label>Ocean Freight ($ USD)</label>
                <input
                  type="number"
                  value={oceanFreight}
                  onChange={(e) => setOceanFreight(e.target.value)}
                  className="form-input"
                />
              </div>

              <div className="form-group">
                <label>Origin Charges ($ USD)</label>
                <input
                  type="number"
                  value={originCharges}
                  onChange={(e) => setOriginCharges(e.target.value)}
                  className="form-input"
                />
              </div>

              <div className="form-group">
                <label>Destination Charges ($ USD)</label>
                <input
                  type="number"
                  value={destinationCharges}
                  onChange={(e) => setDestinationCharges(e.target.value)}
                  className="form-input"
                />
              </div>

              <div className="form-group">
                <label>Valid From</label>
                <input
                  type="date"
                  value={validFrom}
                  onChange={(e) => setValidFrom(e.target.value)}
                  className="form-input"
                />
              </div>

              <div className="form-group">
                <label>Valid Until</label>
                <input
                  type="date"
                  value={validUntil}
                  onChange={(e) => setValidUntil(e.target.value)}
                  className="form-input"
                />
              </div>
            </div>

            {/* Surcharges List Editor */}
            <div className="surcharges-editor-section">
              <div className="surcharges-header">
                <h5>Extracted Surcharges</h5>
                <button type="button" onClick={handleAddSurcharge} className="btn-add-surcharge">
                  + Add Surcharge
                </button>
              </div>

              <div className="surcharges-list">
                {surcharges.map((s, idx) => (
                  <div key={idx} className="surcharge-editor-row">
                    <input
                      type="text"
                      placeholder="Code"
                      value={s.code}
                      onChange={(e) => handleSurchargeChange(idx, 'code', e.target.value.toUpperCase())}
                      className="surcharge-input-code"
                    />
                    <input
                      type="text"
                      placeholder="Description"
                      value={s.description}
                      onChange={(e) => handleSurchargeChange(idx, 'description', e.target.value)}
                      className="surcharge-input-desc"
                    />
                    <input
                      type="number"
                      placeholder="Amt"
                      value={s.amount}
                      onChange={(e) => handleSurchargeChange(idx, 'amount', e.target.value)}
                      className="surcharge-input-amt"
                    />
                    <select
                      value={s.unit}
                      onChange={(e) => handleSurchargeChange(idx, 'unit', e.target.value)}
                      className="surcharge-input-unit"
                    >
                      <option value="PER_TEU">PER_TEU</option>
                      <option value="PER_CONTAINER">PER_CONTAINER</option>
                      <option value="PER_SHIPMENT">PER_SHIPMENT</option>
                    </select>
                    <label className="surcharge-input-checkbox">
                      <input
                        type="checkbox"
                        checked={s.included}
                        onChange={(e) => handleSurchargeChange(idx, 'included', e.target.checked)}
                      />
                      <span>Incl.</span>
                    </label>
                    <button
                      type="button"
                      onClick={() => handleRemoveSurcharge(idx)}
                      className="btn-remove-surcharge"
                    >
                      ✕
                    </button>
                  </div>
                ))}
              </div>
            </div>

            {/* Review Notes */}
            <div className="form-group full-width" style={{ marginTop: '16px' }}>
              <label>Review Notes (Required if rejecting, optional for approval)</label>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="State the corrections made or reason for rejecting this extraction..."
                className="form-textarea"
                rows={3}
              />
            </div>
          </div>
        </div>

        {/* Modal Footer Actions */}
        <div className="modal-footer">
          <div className="footer-left">
            <button
              onClick={handleReject}
              disabled={isSubmitting}
              className="btn-danger-outline"
            >
              Reject Rate Entry
            </button>
          </div>
          <div className="footer-right">
            <button onClick={onClose} className="btn-secondary">
              Cancel
            </button>
            <button
              onClick={handleApprove}
              disabled={isSubmitting}
              className="btn-primary"
            >
              Approve & Ingest Rate
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
