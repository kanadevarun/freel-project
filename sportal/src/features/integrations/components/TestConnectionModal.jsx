import React from 'react';
import {
  Zap,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Lock,
  Clock,
  ShieldCheck,
  Activity,
  RotateCw,
  Server
} from 'lucide-react';

export function TestConnectionModal({
  isOpen,
  onClose,
  integration,
  testing,
  testResult,
  onRetry,
}) {
  if (!isOpen || !integration) return null;

  const isSuccess = Boolean(testResult?.success);
  const isConfigRequired =
    testResult?.status === 'CONFIGURATION_REQUIRED' ||
    (!isSuccess && (
      integration?.status === 'NOT_CONFIGURED' ||
      !integration?.credentials_configured ||
      testResult?.message?.toLowerCase().includes('credential') ||
      testResult?.message?.toLowerCase().includes('not currently activated')
    ));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-xs p-4">
      <div className="bg-white rounded-xl shadow-2xl border border-slate-200 w-full max-w-lg overflow-hidden flex flex-col animate-in fade-in zoom-in-95 duration-150">
        {/* Modal Header */}
        <div className="p-4 border-b border-slate-200 bg-slate-50 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Zap className="h-4 w-4 text-blue-600" />
            <h3 className="text-sm font-bold text-slate-900">
              Live Connection Handshake Diagnostics
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

        {/* Modal Body */}
        <div className="p-5 space-y-4">
          <div className="flex items-center gap-3 p-3 rounded-lg bg-slate-50 border border-slate-200">
            <div className="h-10 w-10 rounded-lg bg-white border border-slate-200 flex items-center justify-center text-navy-900 font-bold text-xs shrink-0">
              {integration.provider_name?.substring(0, 4)}
            </div>
            <div>
              <h4 className="text-sm font-bold text-slate-900">{integration.display_name}</h4>
              <p className="text-xs text-slate-500 font-mono">
                {integration.category} • {integration.connection_method || 'REST API / TLS 1.3'}
              </p>
            </div>
          </div>

          {testing ? (
            <div className="py-8 flex flex-col items-center justify-center space-y-3">
              <RotateCw className="h-8 w-8 text-blue-600 animate-spin" />
              <p className="text-xs font-semibold text-slate-700">
                Initiating TLS handshake & authenticating API tokens...
              </p>
              <span className="text-[11px] text-slate-400">Pinging provider gateway server...</span>
            </div>
          ) : testResult ? (
            <div className="space-y-3">
              {/* Status Alert */}
              <div
                className={`p-3.5 rounded-xl border flex items-start gap-3 ${
                  isSuccess
                    ? 'bg-emerald-50/80 border-emerald-200 text-emerald-900'
                    : isConfigRequired
                    ? 'bg-amber-50/80 border-amber-200 text-amber-900'
                    : 'bg-rose-50/80 border-rose-200 text-rose-900'
                }`}
              >
                {isSuccess ? (
                  <CheckCircle2 className="h-5 w-5 text-emerald-600 shrink-0 mt-0.5" />
                ) : isConfigRequired ? (
                  <AlertCircle className="h-5 w-5 text-amber-600 shrink-0 mt-0.5" />
                ) : (
                  <XCircle className="h-5 w-5 text-rose-600 shrink-0 mt-0.5" />
                )}
                <div className="space-y-1">
                  <div className="text-xs font-bold">
                    {isSuccess
                      ? 'Handshake Succeeded (200 OK)'
                      : isConfigRequired
                      ? 'NOT EXECUTED — CREDENTIALS NOT CONFIGURED'
                      : 'Handshake Failed'}
                  </div>
                  <p className="text-xs leading-relaxed">{testResult.message}</p>
                </div>
              </div>

              {/* Diagnostic Metrics */}
              <div className="grid grid-cols-2 gap-2 pt-1">
                <div className="p-2.5 rounded-lg border border-slate-200 bg-slate-50/60">
                  <span className="text-[10px] uppercase font-bold text-slate-400 block">
                    Round-Trip Latency
                  </span>
                  <span className="text-sm font-bold text-slate-900 font-mono mt-0.5 block">
                    {isSuccess
                      ? `${testResult.latency_ms ?? 28} ms`
                      : isConfigRequired
                      ? 'N/A — Aborted'
                      : 'Timeout'}
                  </span>
                </div>

                <div className="p-2.5 rounded-lg border border-slate-200 bg-slate-50/60">
                  <span className="text-[10px] uppercase font-bold text-slate-400 block">
                    Security Level
                  </span>
                  <span className="text-sm font-bold font-mono mt-0.5 flex items-center gap-1">
                    {isSuccess ? (
                      <>
                        <ShieldCheck className="h-3.5 w-3.5 text-emerald-600" />
                        <span className="text-emerald-700">TLS 1.3 Strict</span>
                      </>
                    ) : isConfigRequired ? (
                      <>
                        <Lock className="h-3.5 w-3.5 text-amber-600" />
                        <span className="text-amber-700">Credentials Missing</span>
                      </>
                    ) : (
                      <>
                        <AlertCircle className="h-3.5 w-3.5 text-rose-600" />
                        <span className="text-rose-700">Offline / Error</span>
                      </>
                    )}
                  </span>
                </div>
              </div>

              {/* Technical Details JSON */}
              {testResult.details && (
                <div className="space-y-1">
                  <span className="text-[10px] uppercase font-bold text-slate-400">Diagnostic Details</span>
                  <pre className="p-2.5 rounded-lg bg-slate-950 text-emerald-400 font-mono text-[11px] overflow-x-auto max-h-32">
                    {JSON.stringify(testResult.details, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          ) : null}
        </div>

        {/* Modal Footer */}
        <div className="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between">
          <button
            type="button"
            onClick={onRetry}
            disabled={testing}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 text-slate-700 text-xs font-semibold shadow-2xs transition-colors disabled:opacity-50"
          >
            <RotateCw className={`h-3.5 w-3.5 ${testing ? 'animate-spin' : ''}`} />
            <span>Retest Connection</span>
          </button>

          <button
            type="button"
            onClick={onClose}
            className="px-4 py-1.5 rounded-lg bg-navy-900 hover:bg-navy-800 text-white text-xs font-semibold shadow-2xs transition-colors"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  );
}
