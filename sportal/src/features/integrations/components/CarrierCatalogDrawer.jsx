import React, { useState } from 'react';
import {
  Ship,
  Search,
  CheckCircle2,
  ExternalLink,
  Plus,
  ShieldCheck,
  Zap,
  Globe,
  Radio,
  FileText
} from 'lucide-react';

export function CarrierCatalogDrawer({
  isOpen,
  onClose,
  catalog = [],
  configuredIntegrations = [],
  onSelectCarrier,
}) {
  const [search, setSearch] = useState('');
  const [modeFilter, setModeFilter] = useState('ALL');

  if (!isOpen) return null;

  const configuredCodes = new Set(
    configuredIntegrations.map((i) => i.provider_name?.toUpperCase())
  );

  const filtered = catalog.filter((c) => {
    const name = c.name || c.display_name || '';
    const code = c.code || c.provider_name || c.scac || '';
    const desc = c.description || '';
    const term = search.toLowerCase();
    const matchesSearch =
      !term ||
      name.toLowerCase().includes(term) ||
      code.toLowerCase().includes(term) ||
      desc.toLowerCase().includes(term);
    return matchesSearch;
  });

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-slate-900/40 backdrop-blur-xs flex justify-end">
      <div className="w-full max-w-2xl bg-white h-full shadow-2xl flex flex-col animate-in slide-in-from-right duration-200">
        {/* Header */}
        <div className="p-5 border-b border-slate-200 bg-slate-50 flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <Ship className="h-5 w-5 text-navy-900" />
              <h3 className="text-base font-bold text-slate-900">
                Global Ocean Carrier & EDI Gateway Catalog
              </h3>
            </div>
            <p className="text-xs text-slate-500 mt-1">
              Connect direct liner shipping APIs, EDI 214/304 gateways, or INTTRA aggregators
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-slate-700 font-bold p-1 text-lg leading-none"
          >
            ✕
          </button>
        </div>

        {/* Search and Filters */}
        <div className="p-4 border-b border-slate-100 bg-white flex items-center gap-3">
          <div className="relative flex-1">
            <Search className="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="Search carrier name, SCAC, or capabilities..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full h-9 pl-9 pr-3 text-xs rounded-lg border border-slate-200 focus:outline-none focus:border-navy-900"
            />
          </div>
        </div>

        {/* List of Carriers */}
        <div className="flex-1 overflow-y-auto p-5 space-y-3 divide-y divide-slate-100 custom-scrollbar">
          {filtered.length === 0 ? (
            <div className="py-12 text-center text-slate-400">
              <Ship className="h-8 w-8 mx-auto text-slate-300 mb-2" />
              <p className="text-sm font-semibold text-slate-600">No carrier gateways found</p>
              <p className="text-xs text-slate-400 mt-1">Try adjusting your search query</p>
            </div>
          ) : (
            filtered.map((carrier, idx) => {
              const code = carrier.code || carrier.provider_name || carrier.scac || `CARRIER-${idx + 1}`;
              const name = carrier.name || carrier.display_name || code;
              const description =
                carrier.description ||
                `${name} direct ocean liner shipping API and EDI 214 telemetry gateway connection.`;
              const capabilities =
                Array.isArray(carrier.supported_capabilities) && carrier.supported_capabilities.length > 0
                  ? carrier.supported_capabilities
                  : Array.isArray(carrier.supported_protocols) && carrier.supported_protocols.length > 0
                  ? carrier.supported_protocols
                  : ['Container Tracking', 'REST API', 'EDI 214'];
              const isConfigured = configuredCodes.has(code.toUpperCase());

              return (
                <div key={code} className="pt-3.5 first:pt-0 flex items-start justify-between gap-4">
                  <div className="space-y-1.5 flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <h4 className="text-sm font-bold text-slate-900">{name}</h4>
                      <span className="font-mono text-[11px] font-bold bg-blue-50 text-blue-700 px-2 py-0.5 rounded border border-blue-200">
                        {code}
                      </span>
                      {isConfigured && (
                        <span className="inline-flex items-center gap-1 text-[10px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 px-2 py-0.5 rounded-full">
                          <CheckCircle2 className="h-3 w-3" /> Configured
                        </span>
                      )}
                    </div>

                    <p className="text-xs text-slate-600 leading-relaxed">{description}</p>

                    <div className="flex flex-wrap gap-1.5 pt-1">
                      {capabilities.map((cap) => (
                        <span
                          key={cap}
                          className="text-[10px] font-medium bg-slate-100 text-slate-600 px-1.5 py-0.5 rounded"
                        >
                          {cap}
                        </span>
                      ))}
                    </div>
                  </div>

                  <button
                    type="button"
                    onClick={() =>
                      onSelectCarrier({
                        ...carrier,
                        code,
                        name,
                        provider_name: code,
                        display_name: name,
                        description,
                      })
                    }
                    className={`px-3 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1 shrink-0 transition-colors shadow-2xs ${
                      isConfigured
                        ? 'border border-slate-200 bg-white text-slate-700 hover:bg-slate-50'
                        : 'bg-navy-900 text-white hover:bg-navy-800'
                    }`}
                  >
                    {isConfigured ? (
                      <span>Manage</span>
                    ) : (
                      <>
                        <Plus className="h-3.5 w-3.5" />
                        <span>Connect</span>
                      </>
                    )}
                  </button>
                </div>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between text-xs text-slate-500">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-emerald-600" />
            <span>Encrypted OAuth2 / TLS 1.3 Key Vault</span>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 text-slate-700 text-xs font-medium"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
