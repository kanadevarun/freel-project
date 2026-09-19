import React from 'react';
import { Layers, ShieldCheck, ArrowRight, Info } from 'lucide-react';

export function HonestPlaceholder({
  moduleName,
  description,
  features = [],
}) {
  return (
    <div className="p-8 max-w-5xl mx-auto space-y-6">
      {/* Header breadcrumb */}
      <div className="flex items-center space-x-2 text-xs text-slate-500">
        <span>SPortal</span>
        <span>/</span>
        <span className="font-semibold text-slate-900">{moduleName}</span>
      </div>

      {/* Main operational card */}
      <div className="bg-white rounded-xl border border-slate-200 p-8 shadow-xs">
        <div className="flex items-start justify-between">
          <div className="flex items-center space-x-4">
            <div className="w-12 h-12 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center border border-blue-100">
              <Layers className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-slate-900">{moduleName}</h1>
              <p className="text-slate-500 text-xs mt-1">{description}</p>
            </div>
          </div>
          <div className="inline-flex items-center space-x-1.5 px-3 py-1 bg-emerald-50 text-emerald-800 border border-emerald-200 rounded-full text-xs font-semibold">
            <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
            <span>Active Module</span>
          </div>
        </div>

        {features.length > 0 && (
          <div className="mt-8 pt-6 border-t border-slate-100">
            <h4 className="text-xs font-semibold text-slate-800 mb-3 uppercase tracking-wider">Module Capabilities</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {features.map((feat, idx) => (
                <div key={idx} className="flex items-center space-x-2.5 p-3 rounded-lg bg-slate-50 border border-slate-200 text-xs text-slate-700">
                  <ArrowRight className="w-3.5 h-3.5 text-blue-600 flex-shrink-0" />
                  <span>{feat}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
