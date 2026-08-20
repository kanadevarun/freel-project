import { useState } from 'react';
import PropTypes from 'prop-types';
import { createCampaign } from '../../../services/outreachService';

/**
 * CampaignBuilder — A 3-step wizard modal for creating a new campaign.
 *
 * Simple meaning: Instead of one giant form, we break campaign creation into
 * 3 manageable steps. At each step the user fills in part of the info and
 * clicks "Next" to proceed. Only when they hit "Create Campaign" on step 3
 * does anything get saved to the backend.
 *
 * Step 1: Name the campaign
 * Step 2: Write the first email sequence step
 * Step 3: Review everything and confirm
 *
 * @param {{ onClose: () => void, onCampaignCreated: () => void }}
 */
export default function CampaignBuilder({ onClose, onCampaignCreated }) {
  // Which step is the user currently on (1, 2, or 3)
  const [currentStep, setCurrentStep] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // ── Form State ────────────────────────────────────────────────────────────
  // Step 1 data
  const [name, setName] = useState('');
  // Step 2 data — the first email sequence step
  const [emailSubject, setEmailSubject] = useState('');
  const [emailBody, setEmailBody] = useState('');

  // ── Navigation ────────────────────────────────────────────────────────────

  function goNext() {
    setError('');
    if (currentStep === 1 && !name.trim()) {
      setError('Campaign name is required.');
      return;
    }
    if (currentStep === 2 && (!emailSubject.trim() || !emailBody.trim())) {
      setError('Both subject and email body are required.');
      return;
    }
    setCurrentStep(prev => prev + 1);
  }

  function goBack() {
    setError('');
    setCurrentStep(prev => prev - 1);
  }

  // ── Submit ────────────────────────────────────────────────────────────────

  async function handleCreate() {
    setLoading(true);
    setError('');
    try {
      // Only create the campaign for now — sequence step support
      // can be added via a separate "Add Steps" API call after creation.
      await createCampaign({ name: name.trim() });
      onCampaignCreated(); // Tell parent to refresh the list
      onClose();
    } catch (err) {
      const msg = typeof err === 'string' ? err : (err?.message || err?.code || 'Failed to create campaign. Please try again.');
      setError(msg);
    } finally {
      setLoading(false);
    }
  }

  // ── Step Indicator ────────────────────────────────────────────────────────

  const STEPS = ['Name', 'Email Step', 'Review'];

  return (
    <div className="outreach-modal-overlay" onClick={onClose}>
      <div className="outreach-modal" onClick={e => e.stopPropagation()}>

        {/* Header */}
        <div className="modal-header">
          <div className="modal-title">📢 New Campaign</div>
          <button className="modal-close-btn" onClick={onClose}>✕</button>
        </div>

        {/* Step Indicator — shows which step the user is on */}
        <div className="step-indicator">
          {STEPS.map((label, i) => {
            const stepNum = i + 1;
            const isDone   = currentStep > stepNum;
            const isActive = currentStep === stepNum;
            return (
              <div key={label} style={{ display: 'contents' }}>
                <div
                  className={`step-dot ${isDone ? 'done' : isActive ? 'active' : ''}`}
                  title={label}
                >
                  {isDone ? '✓' : stepNum}
                </div>
                {i < STEPS.length - 1 && (
                  <div className={`step-line ${isDone ? 'done' : ''}`} />
                )}
              </div>
            );
          })}
        </div>

        {/* Error */}
        {error && <div className="outreach-error" style={{ marginBottom: 16 }}>{typeof error === 'string' ? error : (error?.message || error?.code || 'An error occurred')}</div>}

        {/* ── STEP 1: Name ── */}
        {currentStep === 1 && (
          <div className="outreach-form">
            <div className="form-group">
              <label>Campaign Name *</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder="e.g., Q3 Freight Prospects"
                autoFocus
              />
            </div>
            <div style={{ fontSize: 12.5, color: '#64748B', background: '#F8FAFC', borderRadius: 8, padding: '10px 12px' }}>
              💡 A clear name helps you track which campaign is targeting which audience.
            </div>
          </div>
        )}

        {/* ── STEP 2: Email Step ── */}
        {currentStep === 2 && (
          <div className="outreach-form">
            <div style={{ fontSize: 13, color: '#64748B', marginBottom: 4 }}>
              Define the first email step for this campaign. You can add more steps after creation.
            </div>
            <div className="form-group">
              <label>Subject Line</label>
              <input
                type="text"
                value={emailSubject}
                onChange={e => setEmailSubject(e.target.value)}
                placeholder="e.g., Smarter Freight for [Company Name]"
                autoFocus
              />
            </div>
            <div className="form-group">
              <label>Email Body</label>
              <textarea
                value={emailBody}
                onChange={e => setEmailBody(e.target.value)}
                placeholder="Hi {{first_name}}, I noticed your company..."
                style={{ minHeight: 140 }}
              />
            </div>
          </div>
        )}

        {/* ── STEP 3: Review ── */}
        {currentStep === 3 && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div style={{ fontSize: 13.5, color: '#64748B' }}>
              Review your campaign before creating it. It will start in <strong>DRAFT</strong> status — you can activate it whenever you&apos;re ready.
            </div>
            <div style={{ background: '#F8FAFC', borderRadius: 12, padding: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
              <div><span style={{ fontSize: 11, color: '#94A3B8', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 600 }}>Campaign Name</span>
                <div style={{ fontWeight: 700, fontSize: 15, marginTop: 2 }}>{name}</div>
              </div>
              <div><span style={{ fontSize: 11, color: '#94A3B8', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 600 }}>Step 1 Subject</span>
                <div style={{ fontSize: 13.5, marginTop: 2 }}>{emailSubject || '—'}</div>
              </div>
              <div><span style={{ fontSize: 11, color: '#94A3B8', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 600 }}>Status on Create</span>
                <div style={{ fontSize: 13.5, color: '#64748B', marginTop: 2 }}>📝 Draft (Inactive)</div>
              </div>
            </div>
          </div>
        )}

        {/* Navigation Buttons */}
        <div className="modal-actions" style={{ marginTop: 20 }}>
          {currentStep > 1 ? (
            <button className="outreach-btn outreach-btn-ghost" onClick={goBack}>← Back</button>
          ) : (
            <button className="outreach-btn outreach-btn-ghost" onClick={onClose}>Cancel</button>
          )}

          {currentStep < 3 ? (
            <button className="outreach-btn outreach-btn-primary" onClick={goNext}>
              Next →
            </button>
          ) : (
            <button
              className="outreach-btn outreach-btn-primary"
              onClick={handleCreate}
              disabled={loading}
            >
              {loading ? 'Creating...' : '🚀 Create Campaign'}
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
