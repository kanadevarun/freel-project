import React from 'react';
import { Sparkles, HelpCircle, CheckCircle2 } from 'lucide-react';
import AIConfidenceIndicator from './AIConfidenceIndicator';
import './SharedAI.css';

export default function AIRecommendationCard({
  title = 'AI Context Synthesis & Recommendations',
  executiveSummary,
  tradeoffs,
  questions = [],
  confidence = 'HIGH',
  disclaimer = 'Decision-support intelligence only. Does not constitute legal, credit, or financial advice.',
}) {
  if (!executiveSummary && (!questions || questions.length === 0)) return null;

  return (
    <div className="sai-card" style={{ background: '#f8fafc', border: '1px solid #e2e8f0' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8, paddingBottom: 6, borderBottom: '1px solid #e2e8f0' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <Sparkles size={14} className="text-teal-600" />
          <h4 style={{ margin: 0, fontSize: '0.82rem', fontWeight: 750, color: '#0f172a' }}>{title}</h4>
        </div>
        <AIConfidenceIndicator confidence={confidence} />
      </div>

      {executiveSummary && (
        <p style={{ margin: '0 0 8px 0', fontSize: '0.78rem', color: '#334155', lineHeight: 1.5 }}>
          {executiveSummary}
        </p>
      )}

      {tradeoffs && (
        <div style={{ marginBottom: 8, fontSize: '0.74rem', color: '#475569', background: '#ffffff', border: '1px solid #e2e8f0', borderRadius: 6, padding: '7px 10px' }}>
          <strong style={{ color: '#0f172a', display: 'block', marginBottom: 2 }}>Tradeoffs & Operational Balance:</strong>
          {tradeoffs}
        </div>
      )}

      {questions && questions.length > 0 && (
        <div style={{ marginTop: 6 }}>
          <span style={{ fontSize: '0.72rem', fontWeight: 750, color: '#0f172a', display: 'flex', alignItems: 'center', gap: 4 }}>
            <HelpCircle size={12} className="text-blue-600" /> Suggested Operator Inquiries:
          </span>
          <ul className="sai-rec-questions">
            {questions.map((q, idx) => (
              <li key={idx}>{q}</li>
            ))}
          </ul>
        </div>
      )}

      {disclaimer && (
        <div style={{ marginTop: 8, paddingTop: 6, borderTop: '1px solid #f1f5f9', fontSize: '0.66rem', color: '#64748b', fontStyle: 'italic' }}>
          {disclaimer}
        </div>
      )}
    </div>
  );
}
