import { useState } from 'react';
import { generateEmail } from '../../../services/outreachService';

/**
 * EmailComposer — An AI-powered email writing tool for outreach campaigns.
 *
 * Simple meaning: The sales team can type in the name of a company they want
 * to email, optionally add their industry and the goal of the email, then click
 * "✨ Generate with AI". The AI will write a professional, personalized email
 * for them. They can edit it and use it in their campaign.
 *
 * How it works technically:
 *   1. User fills in the 3 input fields (company, industry, goal).
 *   2. On "Generate", we call POST /api/v1/outreach/generate-email.
 *   3. The backend builds a prompt using the "generate_email" template and sends it to OpenAI.
 *   4. The AI returns a JSON response with "subject" and "body".
 *   5. We populate the text fields so the user can edit and copy them.
 */
export default function EmailComposer() {
  // ── AI Inputs ─────────────────────────────────────────────────────────────
  const [companyName, setCompanyName] = useState('');
  const [industry, setIndustry]       = useState('');
  const [goal, setGoal]               = useState('');

  // ── AI Output ─────────────────────────────────────────────────────────────
  const [subject, setSubject] = useState('');
  const [body, setBody]       = useState('');

  // ── UI State ──────────────────────────────────────────────────────────────
  const [generating, setGenerating] = useState(false);
  const [error, setError]           = useState('');
  const [copied, setCopied]         = useState(false);

  // ── Generate Email ────────────────────────────────────────────────────────

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

      // Populate the editable output fields
      setSubject(result.subject || '');
      setBody(result.body || '');
    } catch (err) {
      setError(err.message || 'AI generation failed. Please try again.');
    } finally {
      setGenerating(false);
    }
  }

  // ── Copy to Clipboard ─────────────────────────────────────────────────────

  async function handleCopy() {
    const fullEmail = `Subject: ${subject}\n\n${body}`;
    await navigator.clipboard.writeText(fullEmail);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  // ── Render ────────────────────────────────────────────────────────────────
  return (
    <div className="email-composer-wrap">

      {/* Header Banner */}
      <div className="email-composer-header">
        <h2>✨ AI Email Composer</h2>
        <p>Generate personalized cold outreach emails in seconds. Edit before sending.</p>
      </div>

      <div className="email-composer-body">

        {/* ── AI Context Inputs ── */}
        <div className="ai-generate-bar">
          <div className="form-group">
            <label>Target Company *</label>
            <input
              type="text"
              value={companyName}
              onChange={e => setCompanyName(e.target.value)}
              placeholder="e.g., Techlogix India"
            />
          </div>
          <div className="form-group">
            <label>Industry</label>
            <input
              type="text"
              value={industry}
              onChange={e => setIndustry(e.target.value)}
              placeholder="e.g., Electronics, Textiles"
            />
          </div>
          <div className="form-group">
            <label>Email Goal</label>
            <input
              type="text"
              value={goal}
              onChange={e => setGoal(e.target.value)}
              placeholder="e.g., schedule a demo"
            />
          </div>
          <button
            className="ai-generate-btn"
            onClick={handleGenerate}
            disabled={generating}
          >
            {generating ? '⏳ Generating...' : '✨ Generate'}
          </button>
        </div>

        {/* Error */}
        {error && <div className="outreach-error">{error}</div>}

        {/* ── Output — shows after generation ── */}
        {subject || body ? (
          <div className="composer-output">
            <div>
              <label>Subject Line</label>
              {/* User can directly edit the generated subject */}
              <input
                className="composer-subject-input"
                type="text"
                value={subject}
                onChange={e => setSubject(e.target.value)}
              />
            </div>
            <div>
              <label>Email Body</label>
              {/* User can directly edit the generated body */}
              <textarea
                className="composer-body-textarea"
                value={body}
                onChange={e => setBody(e.target.value)}
              />
            </div>
            <div className="composer-actions">
              <button
                className="outreach-btn outreach-btn-ghost"
                onClick={handleGenerate}
                disabled={generating}
              >
                🔄 Regenerate
              </button>
              <button
                className="outreach-btn outreach-btn-primary"
                onClick={handleCopy}
              >
                {copied ? '✅ Copied!' : '📋 Copy to Clipboard'}
              </button>
            </div>
          </div>
        ) : (
          /* Placeholder when nothing has been generated yet */
          <div className="composer-placeholder">
            <div className="composer-placeholder-icon">🤖</div>
            <div className="composer-placeholder-title">Your AI-generated email will appear here</div>
            <div className="composer-placeholder-sub">
              Fill in the company name above and click &quot;✨ Generate&quot; to get started
            </div>
          </div>
        )}

      </div>
    </div>
  );
}
