import React, { useState } from 'react';
import { ChevronDown, ChevronRight, Database } from 'lucide-react';
import './SharedAI.css';

export default function AIEvidenceList({ evidence = [], defaultExpanded = true, title = 'Verifiable Database Facts' }) {
  const [expanded, setExpanded] = useState(defaultExpanded);

  if (!evidence || evidence.length === 0) return null;

  return (
    <div className="sai-evidence-container">
      <div 
        className="sai-evidence-header" 
        onClick={() => setExpanded(!expanded)}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && setExpanded(!expanded)}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <Database size={12} />
          <span>{title} ({evidence.length})</span>
        </div>
        {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
      </div>
      {expanded && (
        <table className="sai-evidence-table">
          <thead>
            <tr>
              <th>Module / Ref</th>
              <th>Field</th>
              <th>Observed Value</th>
              <th>Context</th>
            </tr>
          </thead>
          <tbody>
            {evidence.map((item, idx) => (
              <tr key={idx}>
                <td>
                  <span className="sai-source-ref">
                    {item.source_module ? `${item.source_module.toUpperCase()}: ` : ''}
                    {item.source_ref || item.field_name || `Record #${item.source_entity_id || idx + 1}`}
                  </span>
                </td>
                <td style={{ color: '#64748b' }}>{item.field_name}</td>
                <td className="sai-evidence-val">{String(item.observed_value ?? '—')}</td>
                <td style={{ color: '#475569' }}>{item.description || 'Verified persistent record'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
