import React, { useState } from 'react';
import {
  Settings,
  ShieldCheck,
  Key,
  Globe,
  Lock,
  Save,
  CheckCircle2,
  AlertCircle
} from 'lucide-react';

export function ConfigureIntegrationModal({
  isOpen,
  onClose,
  integration,
  onSave,
}) {
  const [apiKey, setApiKey] = useState('••••••••••••••••••••••••••••••••');
  const [endpoint, setEndpoint] = useState(
    integration?.category === 'CARRIER'
      ? `https://api.${integration?.provider_name?.toLowerCase() || 'carrier'}.com/v2/tracking`
      : 'https://api.logistics-gateway.internal/v1'
  );
  const [saving, setSaving] = useState(false);
  const [savedSuccess, setSavedSuccess] = useState(false);

  if (!isOpen || !integration) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      if (onSave) {
        await onSave(integration, { apiKey, endpoint });
      }
      setSavedSuccess(true);
      setTimeout(() => {
        setSavedSuccess(false);
        onClose();
      }, 1200);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-xs p-4">
      <div className="bg-white rounded-xl shadow-2xl border border-slate-200 w-full max-w-md overflow-hidden flex flex-col animate-in fade-in zoom-in-95 duration-150">
        <div className="p-4 border-b border-slate-200 bg-slate-50 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Settings className="h-4 w-4 text-navy-900" />
            <h3 className="text-sm font-bold text-slate-900">
              Configure {integration.display_name}
            </h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-slate-700 text-sm font-bold p-1 leading-none"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-4 text-xs">
          <div className="p-3 rounded-lg bg-blue-50/70 border border-blue-200 text-blue-900 space-y-1">
            <span className="font-bold block">Enterprise Vault Key Isolation</span>
            <p className="leading-relaxed">
              API tokens and private certificates are stored in hardware-security-module (HSM) backed vaults with automated key rotation.
            </p>
          </div>

          <div>
            <label className="font-semibold text-slate-700 block mb-1">API Base URL / Endpoint</label>
            <div className="relative">
              <Globe className="h-3.5 w-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
              <input
                type="text"
                value={endpoint}
                onChange={(e) => setEndpoint(e.target.value)}
                className="w-full h-8 pl-8 pr-3 rounded-lg border border-slate-200 focus:outline-none focus:border-navy-900 font-mono text-[11px]"
              />
            </div>
          </div>

          <div>
            <label className="font-semibold text-slate-700 block mb-1">
              API Key / Access Secret Token
            </label>
            <div className="relative">
              <Key className="h-3.5 w-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                className="w-full h-8 pl-8 pr-3 rounded-lg border border-slate-200 focus:outline-none focus:border-navy-900 font-mono text-[11px]"
              />
            </div>
            <span className="text-[10px] text-slate-400 block mt-1">
              Leave masked to preserve existing configured credentials.
            </span>
          </div>

          {savedSuccess && (
            <div className="p-2.5 rounded-lg bg-emerald-50 border border-emerald-200 text-emerald-800 flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-emerald-600" />
              <span>Configuration saved and validated successfully!</span>
            </div>
          )}

          <div className="pt-2 border-t border-slate-100 flex items-center justify-between">
            <button
              type="button"
              onClick={onClose}
              className="px-3.5 py-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-50 text-slate-700 font-semibold"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-1.5 rounded-lg bg-navy-900 hover:bg-navy-800 text-white font-semibold flex items-center gap-1.5 shadow-2xs transition-colors"
            >
              <Save className="h-3.5 w-3.5" />
              <span>{saving ? 'Saving...' : 'Save Configuration'}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
