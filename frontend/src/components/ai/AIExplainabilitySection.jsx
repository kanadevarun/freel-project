import React, { useState } from 'react';
import { ChevronDown, ChevronRight, HelpCircle, ShieldCheck, Database, Cpu } from 'lucide-react';
import './SharedAI.css';

export default function AIExplainabilitySection({
  reasoning,
  groundingStatus = 'GROUNDED',
  confidence = 'HIGH',
  modelUsed,
  factors = []
}) {
  const [isOpen, setIsOpen] = useState(false);

  if (!reasoning && (!factors || factors.length === 0)) return null;

  return (
    <div className="sai-explainability">
      <button
        type="button"
        className="sai-explainability-toggle"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
      >
        <span className="flex items-center gap-1.5 font-medium text-slate-700 text-xs">
          <HelpCircle className="w-3.5 h-3.5 text-indigo-500" />
          Why did AI recommend this?
        </span>
        <span className="flex items-center gap-2">
          <span className={`sai-grounding-pill sai-grounding-pill--${String(groundingStatus).toLowerCase()}`}>
            <Database className="w-2.5 h-2.5" />
            {groundingStatus}
          </span>
          {isOpen ? <ChevronDown className="w-3.5 h-3.5 text-slate-400" /> : <ChevronRight className="w-3.5 h-3.5 text-slate-400" />}
        </span>
      </button>

      {isOpen && (
        <div className="sai-explainability-content">
          {reasoning && (
            <p className="text-xs text-slate-600 mb-2 leading-relaxed">
              {reasoning}
            </p>
          )}
          {factors && factors.length > 0 && (
            <div className="sai-explainability-factors">
              <span className="text-[10px] font-bold uppercase tracking-wider text-slate-500">Contributing Evidence:</span>
              <ul className="mt-1 space-y-1">
                {factors.map((f, idx) => (
                  <li key={idx} className="flex items-start gap-1.5 text-xs text-slate-600">
                    <ShieldCheck className="w-3 h-3 text-emerald-500 mt-0.5 shrink-0" />
                    <span>{typeof f === 'string' ? f : f.label || JSON.stringify(f)}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
          {modelUsed && (
            <div className="mt-2 pt-2 border-t border-slate-100 flex items-center justify-between text-[10px] text-slate-400">
              <span className="flex items-center gap-1">
                <Cpu className="w-2.5 h-2.5" /> Evaluated by: {modelUsed}
              </span>
              <span>Confidence: {confidence}</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
