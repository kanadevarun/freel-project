import React, { useState, useEffect } from 'react';
import { X, Layers, Plus, Trash2, AlertCircle, CheckCircle2, DollarSign } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function PlanEditorModal({ isOpen, onClose, plan, onSuccess }) {
  const isEditing = Boolean(plan && plan.id);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [priceMonthly, setPriceMonthly] = useState('');
  const [priceAnnual, setPriceAnnual] = useState('');
  const [features, setFeatures] = useState([]);
  const [newFeatureText, setNewFeatureText] = useState('');
  const [limits, setLimits] = useState({
    team_members: 5,
    ai_email_processing: 500,
    rfqs: 100,
    shipments: 100,
    carrier_connections: 1,
    storage_gb: 10,
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (plan) {
      setName(plan.name || '');
      setDescription(plan.description || '');
      setPriceMonthly(String(plan.price_monthly ?? ''));
      setPriceAnnual(String(plan.price_annual ?? ''));
      setFeatures(Array.isArray(plan.features) ? [...plan.features] : []);
      setLimits({
        team_members: plan.limits?.team_members ?? 5,
        ai_email_processing: plan.limits?.ai_email_processing ?? 500,
        rfqs: plan.limits?.rfqs ?? 100,
        shipments: plan.limits?.shipments ?? 100,
        carrier_connections: plan.limits?.carrier_connections ?? 1,
        storage_gb: plan.limits?.storage_gb ?? 10,
      });
    } else {
      setName('');
      setDescription('');
      setPriceMonthly('99');
      setPriceAnnual('950');
      setFeatures(['Up to 5 team members', 'AI email processing', 'Standard support']);
      setLimits({
        team_members: 5,
        ai_email_processing: 500,
        rfqs: 100,
        shipments: 100,
        carrier_connections: 1,
        storage_gb: 10,
      });
    }
    setError('');
  }, [plan, isOpen]);

  if (!isOpen) return null;

  const handleAddFeature = () => {
    if (newFeatureText.trim()) {
      setFeatures([...features, newFeatureText.trim()]);
      setNewFeatureText('');
    }
  };

  const handleRemoveFeature = (idx) => {
    setFeatures(features.filter((_, i) => i !== idx));
  };

  const handleLimitChange = (key, val) => {
    const num = parseInt(val, 10);
    setLimits((prev) => ({
      ...prev,
      [key]: isNaN(num) ? 0 : num,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Plan name is required.');
      return;
    }

    try {
      setSubmitting(true);
      setError('');
      const payload = {
        name: name.trim(),
        description: description.trim(),
        price_monthly: parseFloat(priceMonthly) || 0,
        price_annual: parseFloat(priceAnnual) || 0,
        features,
        limits,
      };

      let res;
      if (isEditing) {
        res = await sportalService.updateSubscriptionPlan(plan.id, payload);
      } else {
        res = await sportalService.createSubscriptionPlan(payload);
      }

      if (onSuccess) {
        onSuccess(res?.data || res);
      }
      onClose();
    } catch (err) {
      setError(err?.response?.data?.message || err.message || 'Failed to save plan.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 backdrop-blur-xs">
      <div className="w-full max-w-2xl max-h-[90vh] flex flex-col rounded-2xl bg-white shadow-2xl border border-slate-100 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-100 px-6 py-4 bg-slate-50/50 shrink-0">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-50 text-blue-700">
              <Layers className="h-5 w-5" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-900">
                {isEditing ? `Configure Plan: ${plan.name}` : 'Create Commercial Plan Tier'}
              </h2>
              <p className="text-xs text-slate-500">
                Manage commercial pricing, operational quotas, and capability flags
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="overflow-y-auto p-6 space-y-5">
          {error && (
            <div className="flex items-start gap-2.5 rounded-xl border border-rose-200 bg-rose-50/60 p-3.5 text-xs text-rose-700">
              <AlertCircle className="h-4 w-4 shrink-0 text-rose-600 mt-0.5" />
              <div>
                <span className="font-semibold">Validation Error:</span> {error}
              </div>
            </div>
          )}

          {/* Basic Info */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
                Plan Name *
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Starter, Growth, Enterprise"
                required
                className="w-full rounded-xl border border-slate-200 p-2.5 text-xs text-slate-800 focus:border-blue-600 focus:outline-hidden"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
                Description
              </label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Brief summary of target customer segment..."
                className="w-full rounded-xl border border-slate-200 p-2.5 text-xs text-slate-800 focus:border-blue-600 focus:outline-hidden"
              />
            </div>
          </div>

          {/* Pricing */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
                Monthly Price (USD $) *
              </label>
              <div className="relative">
                <span className="absolute left-3 top-2.5 text-xs font-bold text-slate-400">$</span>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  value={priceMonthly}
                  onChange={(e) => setPriceMonthly(e.target.value)}
                  placeholder="99.00"
                  required
                  className="w-full rounded-xl border border-slate-200 pl-7 pr-3 py-2.5 text-xs text-slate-800 focus:border-blue-600 focus:outline-hidden font-mono"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
                Annual Price (USD $) *
              </label>
              <div className="relative">
                <span className="absolute left-3 top-2.5 text-xs font-bold text-slate-400">$</span>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  value={priceAnnual}
                  onChange={(e) => setPriceAnnual(e.target.value)}
                  placeholder="950.00"
                  required
                  className="w-full rounded-xl border border-slate-200 pl-7 pr-3 py-2.5 text-xs text-slate-800 focus:border-blue-600 focus:outline-hidden font-mono"
                />
              </div>
            </div>
          </div>

          {/* Operational Limits / Quotas */}
          <div className="space-y-2 pt-2 border-t border-slate-100">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Operational Quotas & Limits (-1 indicates Unlimited)
            </label>
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
              <div>
                <span className="text-[11px] font-medium text-slate-500 block mb-1">Team Members</span>
                <input
                  type="number"
                  value={limits.team_members}
                  onChange={(e) => handleLimitChange('team_members', e.target.value)}
                  className="w-full rounded-lg border border-slate-200 p-2 text-xs font-mono text-slate-800"
                />
              </div>
              <div>
                <span className="text-[11px] font-medium text-slate-500 block mb-1">AI Email Processing / Mo</span>
                <input
                  type="number"
                  value={limits.ai_email_processing}
                  onChange={(e) => handleLimitChange('ai_email_processing', e.target.value)}
                  className="w-full rounded-lg border border-slate-200 p-2 text-xs font-mono text-slate-800"
                />
              </div>
              <div>
                <span className="text-[11px] font-medium text-slate-500 block mb-1">Monthly RFQ Limit</span>
                <input
                  type="number"
                  value={limits.rfqs}
                  onChange={(e) => handleLimitChange('rfqs', e.target.value)}
                  className="w-full rounded-lg border border-slate-200 p-2 text-xs font-mono text-slate-800"
                />
              </div>
              <div>
                <span className="text-[11px] font-medium text-slate-500 block mb-1">Active Shipments Limit</span>
                <input
                  type="number"
                  value={limits.shipments}
                  onChange={(e) => handleLimitChange('shipments', e.target.value)}
                  className="w-full rounded-lg border border-slate-200 p-2 text-xs font-mono text-slate-800"
                />
              </div>
              <div>
                <span className="text-[11px] font-medium text-slate-500 block mb-1">Carrier Connections</span>
                <input
                  type="number"
                  value={limits.carrier_connections}
                  onChange={(e) => handleLimitChange('carrier_connections', e.target.value)}
                  className="w-full rounded-lg border border-slate-200 p-2 text-xs font-mono text-slate-800"
                />
              </div>
              <div>
                <span className="text-[11px] font-medium text-slate-500 block mb-1">Document Storage (GB)</span>
                <input
                  type="number"
                  value={limits.storage_gb}
                  onChange={(e) => handleLimitChange('storage_gb', e.target.value)}
                  className="w-full rounded-lg border border-slate-200 p-2 text-xs font-mono text-slate-800"
                />
              </div>
            </div>
          </div>

          {/* Included Features List */}
          <div className="space-y-2 pt-2 border-t border-slate-100">
            <label className="text-xs font-bold text-slate-700 uppercase tracking-wider block">
              Included Commercial Features
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                value={newFeatureText}
                onChange={(e) => setNewFeatureText(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    handleAddFeature();
                  }
                }}
                placeholder="e.g. Priority 24/7 Phone Support, Custom API Webhooks..."
                className="flex-1 rounded-xl border border-slate-200 p-2 text-xs text-slate-800 focus:border-blue-600 focus:outline-hidden"
              />
              <button
                type="button"
                onClick={handleAddFeature}
                className="rounded-xl bg-slate-100 px-3 py-2 text-xs font-bold text-slate-700 hover:bg-slate-200 transition-colors"
              >
                <Plus className="h-4 w-4" />
              </button>
            </div>

            <div className="space-y-1.5 max-h-36 overflow-y-auto pr-1">
              {features.map((feat, idx) => (
                <div
                  key={idx}
                  className="flex items-center justify-between rounded-lg bg-slate-50 border border-slate-200/80 px-3 py-1.5 text-xs text-slate-800"
                >
                  <span className="truncate">{feat}</span>
                  <button
                    type="button"
                    onClick={() => handleRemoveFeature(idx)}
                    className="text-slate-400 hover:text-rose-600 p-0.5 ml-2 transition-colors"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              ))}
              {features.length === 0 && (
                <p className="text-[11px] text-slate-400 italic">No features listed yet.</p>
              )}
            </div>
          </div>

          {/* Footer Actions */}
          <div className="flex items-center justify-end gap-2.5 pt-3 border-t border-slate-100">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="rounded-xl border border-slate-200 px-4 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="inline-flex items-center gap-1.5 rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white hover:bg-blue-700 disabled:opacity-50 transition-colors shadow-xs"
            >
              {submitting ? 'Saving Plan...' : isEditing ? 'Update Plan' : 'Create Plan'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
