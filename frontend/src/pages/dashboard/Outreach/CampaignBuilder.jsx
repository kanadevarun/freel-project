import { useState } from 'react';
import PropTypes from 'prop-types';
import { createCampaign } from '../../../services/outreachService';
import { AlertCircle } from 'lucide-react';

export default function CampaignBuilder({ onClose, onCampaignCreated }) {
  const [currentStep, setCurrentStep] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const [name, setName] = useState('');
  const [emailSubject, setEmailSubject] = useState('');
  const [emailBody, setEmailBody] = useState('');

  const STEPS = ['Campaign', 'Email', 'Review'];

  function goNext() {
    setError('');
    if (currentStep === 1 && !name.trim()) {
      setError('Campaign name is required.');
      return;
    }
    if (currentStep === 2) {
      if (!emailSubject.trim()) {
        setError('Email subject is required.');
        return;
      }
      if (!emailBody.trim()) {
        setError('Email body is required.');
        return;
      }
    }
    setCurrentStep(prev => prev + 1);
  }

  function goBack() {
    setError('');
    setCurrentStep(prev => prev - 1);
  }

  async function handleCreate() {
    setLoading(true);
    setError('');

    try {
      const payload = {
        name: name.trim(),
        subject: emailSubject.trim(),
        body: emailBody.trim(),
        channel: 'EMAIL',
        status: 'DRAFT',
      };

      await createCampaign(payload);
      await onCampaignCreated();
      onClose();
    } catch (err) {
      const message =
        typeof err === 'string'
          ? err
          : err?.response?.data?.message || err?.message || 'Failed to create campaign. Please try again.';
      setError(message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="outreach-modal-overlay" onClick={onClose}>
      <div 
        className="outreach-modal" 
        onClick={e => e.stopPropagation()}
        style={{ 
          display: 'flex', 
          flexDirection: 'column', 
          maxHeight: '85vh', 
          width: '560px',
          padding: 0, 
          overflow: 'hidden',
          borderRadius: 16,
          background: '#FFFFFF',
          boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
        }}
      >
        {/* Header Section */}
        <div style={{ padding: '24px 28px 16px', borderBottom: '1px solid #F1F5F9', flexShrink: 0 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <div style={{ fontSize: 18, fontWeight: 800, color: '#0F172A', lineHeight: 1.2 }}>New Campaign</div>
              <div style={{ fontSize: 13, color: '#64748B', marginTop: 2 }}>
                Create and prepare your outreach campaign.
              </div>
            </div>
            <button 
              className="modal-close-btn" 
              onClick={onClose} 
              disabled={loading} 
              type="button"
              style={{ background: 'none', border: 'none', fontSize: 18, color: '#64748B', cursor: 'pointer', padding: 4 }}
            >
              ✕
            </button>
          </div>

          {/* Stepper progress indicator with perfectly aligned vertical columns */}
          <div style={{ marginTop: 20, position: 'relative', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 20px' }}>
            {/* Background line connecting all dots */}
            <div 
              style={{ 
                position: 'absolute', 
                top: '14px', 
                left: '60px', 
                right: '60px', 
                height: '2px', 
                background: '#E2E8F0', 
                zIndex: 1
              }} 
            />
            
            {/* Progress line highlighting completed steps */}
            <div 
              style={{ 
                position: 'absolute', 
                top: '14px', 
                left: '60px', 
                right: '60px', 
                width: currentStep === 1 ? '0%' : currentStep === 2 ? '50%' : '100%',
                height: '2px', 
                background: '#00BFA5', 
                zIndex: 2,
                transition: 'width 250ms ease'
              }} 
            />

            {STEPS.map((label, index) => {
              const stepNumber = index + 1;
              const isDone = currentStep > stepNumber;
              const isActive = currentStep === stepNumber;

              return (
                <div 
                  key={label} 
                  style={{ 
                    display: 'flex', 
                    flexDirection: 'column', 
                    alignItems: 'center', 
                    zIndex: 3, 
                    position: 'relative',
                    width: '80px'
                  }}
                >
                  <div 
                    className={`step-dot ${isDone ? 'done' : isActive ? 'active' : ''}`} 
                    style={{ margin: 0 }}
                  >
                    {isDone ? '✓' : stepNumber}
                  </div>
                  <span
                    style={{
                      fontSize: '11.5px',
                      fontWeight: currentStep === index + 1 ? 800 : 500,
                      color: currentStep === index + 1 ? '#00BFA5' : '#94A3B8',
                      marginTop: '8px',
                      whiteSpace: 'nowrap'
                    }}
                  >
                    {label}
                  </span>
                </div>
              );
            })}
          </div>
        </div>

        {/* Scrollable Content Body */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 16 }}>
          {error && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '10px 14px', background: '#FEF2F2', border: '1px solid #FCA5A5', borderRadius: 8, color: '#991B1B', fontSize: 13, fontWeight: 600, flexShrink: 0 }}>
              <AlertCircle size={15} />
              <span>{error}</span>
            </div>
          )}

          {/* Step 1: Details */}
          {currentStep === 1 && (
            <div className="outreach-form" style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div>
                <h3 style={{ margin: 0, fontSize: 15, fontWeight: 750, color: '#1E293B' }}>Campaign Details</h3>
                <p style={{ margin: '4px 0 0', fontSize: 12.5, color: '#64748B', lineHeight: 1.4 }}>
                  Use a clear name that helps your team identify the campaign and its target audience.
                </p>
              </div>
              <div className="form-group" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                <label style={{ fontSize: 12, fontWeight: 700, color: '#475569' }}>Campaign Name *</label>
                <input
                  type="text"
                  value={name}
                  onChange={e => setName(e.target.value)}
                  placeholder="e.g., Q4 Mumbai Importers Outreach"
                  autoFocus
                  disabled={loading}
                  style={{ padding: '10px 12px', fontSize: 13.5, border: '1px solid #CBD5E1', borderRadius: 8, width: '100%', boxSizing: 'border-box' }}
                />
              </div>
              <div style={{ fontSize: 12, color: '#64748B', background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: 8, padding: '10px 12px', lineHeight: 1.4 }}>
                Tip: Include region or customer segment for clarity.
              </div>
            </div>
          )}

          {/* Step 2: Write Email */}
          {currentStep === 2 && (
            <div className="outreach-form" style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div>
                <h3 style={{ margin: 0, fontSize: 15, fontWeight: 750, color: '#1E293B' }}>Initial Outreach Email</h3>
                <p style={{ margin: '4px 0 0', fontSize: 12.5, color: '#64748B', lineHeight: 1.4 }}>
                  Write the initial email message for this campaign.
                </p>
              </div>
              <div className="form-group" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                <label style={{ fontSize: 12, fontWeight: 700, color: '#475569' }}>Subject Line *</label>
                <input
                  type="text"
                  value={emailSubject}
                  onChange={e => setEmailSubject(e.target.value)}
                  placeholder="e.g., Reduce your freight costs with smarter logistics"
                  autoFocus
                  disabled={loading}
                  style={{ padding: '10px 12px', fontSize: 13.5, border: '1px solid #CBD5E1', borderRadius: 8, width: '100%', boxSizing: 'border-box' }}
                />
              </div>
              <div className="form-group" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                <label style={{ fontSize: 12, fontWeight: 700, color: '#475569' }}>Email Body *</label>
                <textarea
                  value={emailBody}
                  onChange={e => setEmailBody(e.target.value)}
                  placeholder={`Hi {{first_name}},\n\nI wanted to reach out regarding...`}
                  style={{ minHeight: 120, maxHeight: 150, padding: '10px 12px', fontSize: 13.5, border: '1px solid #CBD5E1', borderRadius: 8, width: '100%', boxSizing: 'border-box', fontFamily: 'inherit', resize: 'vertical' }}
                  disabled={loading}
                />
              </div>
              <div style={{ fontSize: 11.5, color: '#64748B', background: '#F8FAFC', borderRadius: 8, padding: '10px 12px', border: '1px solid #E2E8F0', lineHeight: 1.4 }}>
                You can use tokens such as <strong>{'{{first_name}}'}</strong> and <strong>{'{{company_name}}'}</strong>.
              </div>
            </div>
          )}

          {/* Step 3: Review */}
          {currentStep === 3 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div>
                <h3 style={{ margin: 0, fontSize: 15, fontWeight: 750, color: '#1E293B' }}>Review Campaign</h3>
                <p style={{ margin: '4px 0 0', fontSize: 12.5, color: '#64748B', lineHeight: 1.4 }}>
                  Review details before creating your campaign.
                </p>
              </div>
              <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: 10, padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
                <div>
                  <div style={{ fontSize: 10.5, color: '#94A3B8', textTransform: 'uppercase', fontWeight: 700 }}>Campaign Name</div>
                  <div style={{ fontSize: 14.5, fontWeight: 750, marginTop: 2, color: '#1E293B' }}>{name}</div>
                </div>
                <div>
                  <div style={{ fontSize: 10.5, color: '#94A3B8', textTransform: 'uppercase', fontWeight: 700 }}>Subject</div>
                  <div style={{ fontSize: 13.5, fontWeight: 600, marginTop: 2, color: '#334155' }}>{emailSubject}</div>
                </div>
                <div>
                  <div style={{ fontSize: 10.5, color: '#94A3B8', textTransform: 'uppercase', fontWeight: 700 }}>Email Preview</div>
                  <div style={{ marginTop: 4, fontSize: 13, color: '#475569', lineHeight: 1.5, whiteSpace: 'pre-wrap', maxHeight: 95, overflowY: 'auto', background: '#FFFFFF', padding: '8px 10px', borderRadius: 6, border: '1px solid #E2E8F0' }}>
                    {emailBody}
                  </div>
                </div>
                <div style={{ borderTop: '1px solid #E2E8F0', paddingTop: 10, fontSize: 12, color: '#64748B' }}>
                  Status upon creation: <strong>Draft</strong>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div style={{ padding: '16px 28px 20px', borderTop: '1px solid #F1F5F9', flexShrink: 0, display: 'flex', justifyContent: 'flex-end', gap: 12, background: '#FFFFFF', borderBottomLeftRadius: 16, borderBottomRightRadius: 16 }}>
          {currentStep > 1 ? (
            <button 
              type="button" 
              className="outreach-btn outreach-btn-ghost" 
              style={{ padding: '9px 18px', fontSize: 13, borderRadius: 8, fontWeight: 600, cursor: 'pointer' }}
              onClick={goBack} 
              disabled={loading}
            >
              Back
            </button>
          ) : (
            <button 
              type="button" 
              className="outreach-btn outreach-btn-ghost" 
              style={{ padding: '9px 18px', fontSize: 13, borderRadius: 8, fontWeight: 600, cursor: 'pointer' }}
              onClick={onClose} 
              disabled={loading}
            >
              Cancel
            </button>
          )}

          {currentStep < 3 ? (
            <button 
              type="button" 
              className="outreach-btn outreach-btn-primary" 
              style={{ padding: '9px 20px', fontSize: 13, borderRadius: 8, fontWeight: 600, cursor: 'pointer' }}
              onClick={goNext} 
              disabled={loading}
            >
              Continue
            </button>
          ) : (
            <button 
              type="button" 
              className="outreach-btn outreach-btn-primary" 
              style={{ padding: '9px 20px', fontSize: 13, borderRadius: 8, fontWeight: 600, cursor: 'pointer' }}
              onClick={handleCreate} 
              disabled={loading}
            >
              {loading ? 'Creating...' : 'Create Campaign'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

CampaignBuilder.propTypes = {
  onClose: PropTypes.func.isRequired,
  onCampaignCreated: PropTypes.func.isRequired,
};