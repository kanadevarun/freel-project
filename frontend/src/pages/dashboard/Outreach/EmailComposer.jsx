import { useState } from 'react';
import { generateEmail } from '../../../services/outreachService';
import { Sparkles, Copy, RefreshCw, AlertCircle } from 'lucide-react';
import './OutreachPage.css';

export default function EmailComposer() {
  const [companyName, setCompanyName] = useState('');
  const [industry, setIndustry] = useState('');
  const [goal, setGoal] = useState('');

  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');

  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  async function handleGenerate() {
    if (!companyName.trim()) {
      setError('Company name is required to generate an email.');
      return;
    }

    setGenerating(true);
    setError('');

    try {
      const result = await generateEmail({
        company_name: companyName.trim(),
        industry: industry.trim(),
        goal: goal.trim() || 'introduce LogisticsHQ as a freight forwarding partner',
      });

      setSubject(result.subject || '');
      setBody(result.body || '');
    } catch (err) {
      setError(err.message || 'AI generation failed. Please try again.');
    } finally {
      setGenerating(false);
    }
  }

  async function handleCopy() {
    const fullEmail = `Subject: ${subject}\n\n${body}`;
    await navigator.clipboard.writeText(fullEmail);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="email-composer-wrap" style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 16, overflow: 'hidden', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.05)' }}>
      {/* ─── Split Screen Layout ─── */}
      <div style={{ display: 'grid', gridTemplateColumns: '380px 1fr', minHeight: 480 }}>
        
        {/* Left Side: Input Form Context */}
        <div style={{ borderRight: '1px solid #E2E8F0', padding: 24, background: '#F8FAFC' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16 }}>
            <Sparkles size={18} style={{ color: '#2563EB' }} />
            <h3 style={{ margin: 0, fontSize: 15, fontWeight: 750, color: '#0F172A' }}>AI Assistant Context</h3>
          </div>
          <p style={{ fontSize: 12.5, color: '#64748B', lineHeight: 1.5, marginBottom: 20 }}>
            Configure target company parameters and campaign pitch goal. AI will construct cold outreach templates.
          </p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div className="form-group-premium">
              <label className="input-label-premium">Target Company *</label>
              <input
                type="text"
                value={companyName}
                onChange={e => setCompanyName(e.target.value)}
                placeholder="e.g., Techlogix India"
                className="modal-input-premium"
                style={{ background: '#FFFFFF' }}
              />
            </div>

            <div className="form-group-premium">
              <label className="input-label-premium">Industry Segment</label>
              <input
                type="text"
                value={industry}
                onChange={e => setIndustry(e.target.value)}
                placeholder="e.g., Electronics, Textiles"
                className="modal-input-premium"
                style={{ background: '#FFFFFF' }}
              />
            </div>

            <div className="form-group-premium">
              <label className="input-label-premium">Campaign Pitch Goal</label>
              <input
                type="text"
                value={goal}
                onChange={e => setGoal(e.target.value)}
                placeholder="e.g., schedule a demo"
                className="modal-input-premium"
                style={{ background: '#FFFFFF' }}
              />
            </div>

            {error && (
              <div style={{ display: 'flex', gap: 6, alignItems: 'flex-start', background: '#FEF2F2', border: '1px solid #FCA5A5', color: '#991B1B', padding: 10, borderRadius: 8, fontSize: 12 }}>
                <AlertCircle size={15} style={{ flexShrink: 0, marginTop: 1 }} />
                <span>{error}</span>
              </div>
            )}

            <button
              className="outreach-btn outreach-btn-primary"
              onClick={handleGenerate}
              disabled={generating}
              style={{ width: '100%', padding: '10px 14px', fontSize: 13, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 8, marginTop: 10 }}
            >
              {generating ? (
                <>
                  <div className="outreach-spinner" style={{ width: 14, height: 14, border: '2px solid #FFFFFF', borderTopColor: 'transparent', margin: 0 }} />
                  Generating...
                </>
              ) : (
                <>
                  <Sparkles size={14} />
                  Generate Email Template
                </>
              )}
            </button>
          </div>
        </div>

        {/* Right Side: Generated Content Workspace */}
        <div style={{ padding: 24, display: 'flex', flexDirection: 'column', background: '#FFFFFF' }}>
          {subject || body ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 16, height: '100%' }}>
              <div>
                <label className="input-label-premium" style={{ display: 'block', marginBottom: 6 }}>Subject Line</label>
                <input
                  type="text"
                  value={subject}
                  onChange={e => setSubject(e.target.value)}
                  className="modal-input-premium"
                  style={{ width: '100%', fontSize: 14, fontWeight: 600, color: '#0F172A', boxSizing: 'border-box' }}
                />
              </div>

              <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
                <label className="input-label-premium" style={{ display: 'block', marginBottom: 6 }}>Email Template Body</label>
                <textarea
                  value={body}
                  onChange={e => setBody(e.target.value)}
                  className="modal-input-premium"
                  style={{ width: '100%', flex: 1, minHeight: 240, fontSize: 13.5, fontFamily: 'monospace', color: '#334155', boxSizing: 'border-box', lineHeight: 1.6 }}
                />
              </div>

              <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end', borderTop: '1px solid #E2E8F0', paddingTop: 16 }}>
                <button
                  className="activity-btn-outline"
                  onClick={handleGenerate}
                  disabled={generating}
                  style={{ padding: '8px 14px', fontSize: 12.5, display: 'inline-flex', alignItems: 'center', gap: 6 }}
                >
                  <RefreshCw size={14} className={generating ? 'animate-spin' : ''} />
                  Regenerate
                </button>
                <button
                  className="outreach-btn outreach-btn-primary"
                  onClick={handleCopy}
                  style={{ padding: '8px 16px', fontSize: 12.5, display: 'inline-flex', alignItems: 'center', gap: 6 }}
                >
                  <Copy size={14} />
                  {copied ? 'Copied!' : 'Copy to Clipboard'}
                </button>
              </div>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', flex: 1, padding: 40, border: '1.5px dashed #E2E8F0', borderRadius: 12, textAlign: 'center' }}>
              <div style={{ fontSize: 32, marginBottom: 12 }}>🪄</div>
              <h4 style={{ margin: '0 0 6px 0', fontSize: 14.5, fontWeight: 700, color: '#334155' }}>AI Generation Workspace</h4>
              <p style={{ margin: 0, fontSize: 12.5, color: '#64748B', maxWidth: 420, lineHeight: 1.5 }}>
                Fill out the context attributes in the side panel and hit generate. Your AI generated email copy will appear here for review and customization.
              </p>
            </div>
          )}
        </div>

      </div>
    </div>
  );
}