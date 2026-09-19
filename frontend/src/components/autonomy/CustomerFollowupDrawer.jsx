import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  MessageSquare,
  Send,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Shield,
  Layers,
  FileText,
  RotateCw,
  ThumbsUp,
  Ban,
  HelpCircle,
  Play,
  Mail,
  Sliders,
  ChevronRight
} from 'lucide-react';
import autonomyService from '../../services/autonomyService';

/**
 * CustomerFollowupDrawer.jsx — Phase 5 Task 5.4 Autonomous Customer Follow-Up Drawer
 *
 * Provides:
 * - Authoritative contact details and communication preferences (channel, opt-out, frequency limits)
 * - Grounded draft inspection separating [ACTUAL FACT], [PREDICTION], and [RECOMMENDATION]
 * - Active Autonomous Plan lifecycle with approval gates and Action System boundary execution
 * - Customer Response ingestion simulation with live AI classification and replanning/stopping
 * - Comprehensive communication and audit event history
 */
export default function CustomerFollowupDrawer({
  isOpen,
  onClose,
  customerId,
  customerData,
  onFollowupUpdated
}) {
  const [loading, setLoading] = useState(false);
  const [stateData, setStateData] = useState(null);
  const [error, setError] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);

  // Simulation tab states
  const [activeTab, setActiveTab] = useState('DRAFT'); // 'DRAFT', 'PREFERENCES', 'PLAN', 'HISTORY'
  const [replyText, setReplyText] = useState('');
  const [replyLoading, setReplyLoading] = useState(false);
  const [replyResult, setReplyResult] = useState(null);
  const [triggerLoading, setTriggerLoading] = useState(false);

  // Preference form states
  const [channel, setChannel] = useState('EMAIL');
  const [optOut, setOptOut] = useState(false);
  const [optOutReason, setOptOutReason] = useState('');
  const [businessHours, setBusinessHours] = useState(true);
  const [maxFollowups, setMaxFollowups] = useState(3);
  const [prefSaving, setPrefSaving] = useState(false);

  const fetchFollowupState = useCallback(async () => {
    if (!customerId) return;
    setLoading(true);
    setError(null);
    try {
      const resp = await autonomyService.getCustomerFollowupState(customerId);
      const data = resp?.state || resp?.data?.state || resp?.data || resp;
      if (data) {
        setStateData(data);
        const prefs = data.preferences || {};
        setChannel(prefs.preferred_channel || 'EMAIL');
        setOptOut(!!prefs.opt_out);
        setOptOutReason(prefs.opt_out_reason || '');
        setBusinessHours(prefs.business_hours_only !== false);
        setMaxFollowups(prefs.max_followups_per_incident || 3);
      }
    } catch (err) {
      console.error('Failed loading customer followup state:', err);
      setError(err?.response?.data?.message || err.message || 'Failed loading follow-up state');
    } finally {
      setLoading(false);
    }
  }, [customerId]);

  useEffect(() => {
    if (isOpen && customerId) {
      fetchFollowupState();
    }
  }, [isOpen, customerId, fetchFollowupState]);

  if (!isOpen) return null;

  const handleSavePreferences = async (e) => {
    e.preventDefault();
    setPrefSaving(true);
    try {
      await autonomyService.updateCustomerPreferences(customerId, {
        preferred_channel: channel,
        opt_out: optOut,
        opt_out_reason: optOut ? optOutReason : null,
        business_hours_only: businessHours,
        max_followups_per_incident: parseInt(maxFollowups, 10),
        min_followup_interval_hours: 24,
      });
      await fetchFollowupState();
      if (onFollowupUpdated) onFollowupUpdated();
    } catch (err) {
      alert('Error updating preferences: ' + (err?.response?.data?.message || err.message));
    } finally {
      setPrefSaving(false);
    }
  };

  const handleTriggerDelayFollowup = async () => {
    setTriggerLoading(true);
    try {
      await autonomyService.ingestCustomerFollowupEvent(customerId, {
        event_id: `evt-manual-delay-${Date.now()}`,
        event_type: 'SHIPMENT_DELAY',
        deduplication_key: `dedup-cust-${customerId}-delay-${Date.now()}`,
        payload: {
          tracking_number: 'TRK-GLOBAL-7709',
          delay_hours: 28.0,
          reason: 'Severe weather advisory and vessel berthing delays',
          predicted_eta: new Date(Date.now() + 86400000 * 3).toISOString(),
          last_location: 'Singapore Transit Terminal'
        }
      });
      await fetchFollowupState();
      if (onFollowupUpdated) onFollowupUpdated();
    } catch (err) {
      alert('Failed triggering follow-up event: ' + (err?.response?.data?.message || err.message));
    } finally {
      setTriggerLoading(false);
    }
  };

  const handleSendFollowup = async (recordId) => {
    setActionLoading(true);
    try {
      await autonomyService.sendCustomerFollowup(customerId, recordId);
      await fetchFollowupState();
      if (onFollowupUpdated) onFollowupUpdated();
    } catch (err) {
      alert('Action System send rejected: ' + (err?.response?.data?.message || err.message));
    } finally {
      setActionLoading(false);
    }
  };

  const handleApprovePlan = async (planId) => {
    setActionLoading(true);
    try {
      await autonomyService.approvePlan(planId, 'Approved by operator via Customer Follow-Up UI');
      await fetchFollowupState();
      if (onFollowupUpdated) onFollowupUpdated();
    } catch (err) {
      alert('Approval failed: ' + (err?.response?.data?.message || err.message));
    } finally {
      setActionLoading(false);
    }
  };

  const handleIngestReply = async (recordId) => {
    if (!replyText.trim()) return;
    setReplyLoading(true);
    setReplyResult(null);
    try {
      const resp = await autonomyService.ingestCustomerResponse(customerId, recordId, {
        response_text: replyText
      });
      if (resp && resp.result) {
        setReplyResult(resp.result);
      }
      setReplyText('');
      await fetchFollowupState();
      if (onFollowupUpdated) onFollowupUpdated();
    } catch (err) {
      alert('Failed ingesting reply: ' + (err?.response?.data?.message || err.message));
    } finally {
      setReplyLoading(false);
    }
  };

  const recentRecords = stateData?.recent_records || [];
  const activeRecord = recentRecords.length > 0 ? recentRecords[0] : null;
  const activePlan = stateData?.active_plan;
  const primaryContact = stateData?.primary_contact;

  const parseJSON = (str) => {
    if (!str) return [];
    try {
      return typeof str === 'string' ? JSON.parse(str) : str;
    } catch {
      return [];
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-slate-900/40 backdrop-blur-xs flex justify-end animate-fade-in">
      <div className="w-full max-w-3xl bg-white h-full shadow-2xl flex flex-col border-l border-slate-200">
        
        {/* Top Header */}
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600">
              <MessageSquare className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-semibold text-slate-800">
                  {stateData?.customer_name || customerData?.name || `Customer #${customerId}`}
                </h2>
                <span className={`px-2 py-0.5 text-xs font-semibold rounded-full border ${
                  stateData?.followup_status === 'HEALTHY' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' :
                  stateData?.followup_status === 'AWAITING_RESPONSE' ? 'bg-amber-50 text-amber-700 border-amber-200' :
                  stateData?.followup_status === 'OPTED_OUT' ? 'bg-slate-100 text-slate-700 border-slate-300' :
                  'bg-blue-50 text-blue-700 border-blue-200'
                }`}>
                  {stateData?.followup_status || 'HEALTHY'}
                </span>
                <span className="px-2 py-0.5 text-xs font-medium rounded bg-slate-200/70 text-slate-700">
                  {stateData?.account_tier || 'STANDARD'} TIER
                </span>
              </div>
              <p className="text-xs text-slate-500 mt-0.5">
                Authoritative Contact: {primaryContact ? `${primaryContact.first_name} ${primaryContact.last_name} (${primaryContact.email})` : 'System Verified Contact'}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={fetchFollowupState}
              disabled={loading}
              className="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 transition-colors"
              title="Refresh State"
            >
              <RotateCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={onClose}
              className="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="px-6 border-b border-slate-200 bg-white flex gap-6">
          <button
            data-testid="tab-draft"
            onClick={() => setActiveTab('DRAFT')}
            className={`py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeTab === 'DRAFT'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            }`}
          >
            <FileText className="w-4 h-4" />
            Follow-Up & Grounded Draft
          </button>
          <button
            data-testid="tab-plan"
            onClick={() => setActiveTab('PLAN')}
            className={`py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeTab === 'PLAN'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            }`}
          >
            <Layers className="w-4 h-4" />
            Active Plan & Steps
            {activePlan && (
              <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
            )}
          </button>
          <button
            data-testid="tab-preferences"
            onClick={() => setActiveTab('PREFERENCES')}
            className={`py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeTab === 'PREFERENCES'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            }`}
          >
            <Sliders className="w-4 h-4" />
            Policy Preferences
          </button>
          <button
            onClick={() => setActiveTab('HISTORY')}
            className={`py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeTab === 'HISTORY'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            }`}
          >
            <Clock className="w-4 h-4" />
            History ({recentRecords.length})
          </button>
        </div>

        {/* Body Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {error && (
            <div className="p-4 rounded-lg bg-rose-50 border border-rose-200 flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
              <div className="text-sm text-rose-800">{error}</div>
            </div>
          )}

          {/* TAB 1: DRAFT & INTERACTIVE FOLLOW-UP */}
          {activeTab === 'DRAFT' && (
            <div className="space-y-6">
              {/* Quick Trigger Bar if no active draft */}
              <div className="p-4 rounded-lg bg-slate-50 border border-slate-200 flex items-center justify-between">
                <div>
                  <h4 className="text-sm font-semibold text-slate-800">Operational Scenario Simulation</h4>
                  <p className="text-xs text-slate-500 mt-0.5">
                    Evaluate operational shipment disruption and generate grounded communication.
                  </p>
                </div>
                <button
                  onClick={handleTriggerDelayFollowup}
                  disabled={triggerLoading || stateData?.preferences?.opt_out}
                  className="px-3 py-1.5 bg-blue-600 text-white rounded-md text-xs font-medium hover:bg-blue-700 transition-colors flex items-center gap-1.5 disabled:opacity-50"
                >
                  <Play className="w-3.5 h-3.5" />
                  {triggerLoading ? 'Evaluating...' : 'Simulate Delay Event'}
                </button>
              </div>

              {activeRecord ? (
                <div className="border border-slate-200 rounded-xl overflow-hidden bg-white shadow-xs">
                  {/* Status & Metadata header */}
                  <div className="p-4 bg-slate-50/80 border-b border-slate-200 flex items-center justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
                          {activeRecord.event_type}
                        </span>
                        <span className="text-xs text-slate-400">•</span>
                        <span className="text-xs font-medium text-slate-600">
                          Channel: {activeRecord.channel}
                        </span>
                        <span className="text-xs text-slate-400">•</span>
                        <span className="text-xs font-medium text-slate-600">
                          Version {activeRecord.version}
                        </span>
                      </div>
                      <h3 className="text-sm font-bold text-slate-900 mt-1">
                        {activeRecord.subject}
                      </h3>
                    </div>

                    <div className="flex items-center gap-2">
                      <span className={`px-2.5 py-1 text-xs font-semibold rounded-full border ${
                        activeRecord.status === 'SENT' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' :
                        activeRecord.status === 'REQUIRES_APPROVAL' ? 'bg-amber-50 text-amber-700 border-amber-200' :
                        activeRecord.status === 'APPROVED' ? 'bg-blue-50 text-blue-700 border-blue-200' :
                        activeRecord.status === 'STOPPED' ? 'bg-rose-50 text-rose-700 border-rose-200' :
                        'bg-slate-100 text-slate-700 border-slate-300'
                      }`}>
                        {activeRecord.status}
                      </span>
                    </div>
                  </div>

                  {/* Fact / Prediction / Recommendation Separation Pill Bar */}
                  <div className="p-4 border-b border-slate-100 bg-slate-50/40 space-y-3">
                    <div className="text-xs font-semibold text-slate-700 uppercase tracking-wider">
                      Context Verification Breakdown
                    </div>

                    {/* Actual Facts */}
                    <div>
                      <div className="flex items-center gap-1.5 text-xs font-medium text-emerald-700 mb-1">
                        <CheckCircle2 className="w-3.5 h-3.5" />
                        [ACTUAL FACT] — Authoritative Business System Grounding
                      </div>
                      <div className="flex flex-wrap gap-1.5">
                        {parseJSON(activeRecord.actual_facts).map((f, idx) => (
                          <span key={idx} className="px-2 py-1 bg-emerald-50 border border-emerald-200 text-emerald-800 rounded text-xs">
                            {f}
                          </span>
                        ))}
                      </div>
                    </div>

                    {/* Predictions */}
                    <div>
                      <div className="flex items-center gap-1.5 text-xs font-medium text-blue-700 mb-1">
                        <Clock className="w-3.5 h-3.5" />
                        [PREDICTION] — Machine Learning Forecast (Non-Guaranteed)
                      </div>
                      <div className="flex flex-wrap gap-1.5">
                        {parseJSON(activeRecord.predictions).map((p, idx) => (
                          <span key={idx} className="px-2 py-1 bg-blue-50 border border-blue-200 text-blue-800 rounded text-xs">
                            {p}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>

                  {/* Message Composer / Body */}
                  <div className="p-4 space-y-2">
                    <div className="flex items-center justify-between text-xs text-slate-500">
                      <span>Recipient: <strong>{activeRecord.recipient_name}</strong> ({activeRecord.recipient_email})</span>
                      <span>Execution Boundary: <strong>Go Action System</strong></span>
                    </div>
                    <div className="p-3 bg-slate-50 rounded-lg border border-slate-200 font-mono text-xs text-slate-800 whitespace-pre-wrap leading-relaxed">
                      {activeRecord.full_body}
                    </div>
                  </div>

                  {/* Approval / Action System Send Controls */}
                  <div className="p-4 bg-slate-50 border-t border-slate-200 flex items-center justify-between">
                    <div className="text-xs text-slate-500">
                      {activeRecord.status === 'REQUIRES_APPROVAL' ? (
                        <span className="flex items-center gap-1 text-amber-700 font-medium">
                          <AlertTriangle className="w-3.5 h-3.5" />
                          Human approval required before message dispatch
                        </span>
                      ) : activeRecord.status === 'SENT' ? (
                        <span className="flex items-center gap-1 text-emerald-700 font-medium">
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          Dispatched via Action System. Awaiting customer reply.
                        </span>
                      ) : (
                        <span className="text-slate-600">
                          Approved for controlled execution.
                        </span>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      {activeRecord.status === 'REQUIRES_APPROVAL' && activePlan && (
                        <button
                          onClick={() => handleApprovePlan(activePlan.plan_id)}
                          disabled={actionLoading}
                          className="px-3 py-1.5 bg-emerald-600 text-white rounded-md text-xs font-medium hover:bg-emerald-700 transition-colors flex items-center gap-1.5"
                        >
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          Approve Plan & Draft
                        </button>
                      )}

                      {activeRecord.status === 'APPROVED' && (
                        <button
                          onClick={() => handleSendFollowup(activeRecord.id)}
                          disabled={actionLoading}
                          className="px-3 py-1.5 bg-blue-600 text-white rounded-md text-xs font-medium hover:bg-blue-700 transition-colors flex items-center gap-1.5"
                        >
                          <Send className="w-3.5 h-3.5" />
                          {actionLoading ? 'Dispatching...' : 'Send Communication'}
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="p-8 text-center border-2 border-dashed border-slate-200 rounded-xl">
                  <Mail className="w-8 h-8 text-slate-300 mx-auto mb-2" />
                  <p className="text-sm font-medium text-slate-600">No active follow-up communication</p>
                  <p className="text-xs text-slate-400 mt-1">Customer is currently in healthy standing. Simulate an operational event above to initiate controlled follow-up.</p>
                </div>
              )}

              {/* Customer Response Ingestion Simulator */}
              {activeRecord && activeRecord.status === 'SENT' && (
                <div className="border border-slate-200 rounded-xl p-4 bg-white space-y-3">
                  <div className="flex items-center justify-between">
                    <h4 className="text-sm font-semibold text-slate-800 flex items-center gap-2">
                      <MessageSquare className="w-4 h-4 text-blue-600" />
                      Inbound Customer Reply Simulation
                    </h4>
                    <span className="text-xs text-slate-400">Classified by Python AI Sidecar</span>
                  </div>

                  <p className="text-xs text-slate-500">
                    Simulate customer inbound reply (e.g. approval, address change request, formal complaint, or opt-out request):
                  </p>

                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={replyText}
                      onChange={(e) => setReplyText(e.target.value)}
                      placeholder="e.g. 'Looks acceptable, please proceed' or 'Please update delivery address to Building 3'"
                      className="flex-1 px-3 py-2 border border-slate-300 rounded-lg text-xs focus:ring-2 focus:ring-blue-500 outline-hidden"
                    />
                    <button
                      onClick={() => handleIngestReply(activeRecord.id)}
                      disabled={replyLoading || !replyText.trim()}
                      className="px-4 py-2 bg-slate-800 text-white rounded-lg text-xs font-medium hover:bg-slate-900 transition-colors disabled:opacity-50"
                    >
                      {replyLoading ? 'Classifying...' : 'Ingest Reply'}
                    </button>
                  </div>

                  {/* Pre-fill Quick Chips */}
                  <div className="flex flex-wrap gap-2 pt-1">
                    <button
                      onClick={() => setReplyText('Revised timeline looks acceptable. Please proceed.')}
                      className="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs rounded"
                    >
                      + Confirmation Reply
                    </button>
                    <button
                      onClick={() => setReplyText('Please update the delivery address to Warehouse 4B.')}
                      className="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs rounded"
                    >
                      + Action Request (Replanning)
                    </button>
                    <button
                      onClick={() => setReplyText('Unacceptable delay, please cancel and refund this shipment.')}
                      className="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs rounded"
                    >
                      + Complaint / Escalation
                    </button>
                    <button
                      onClick={() => setReplyText('Unsubscribe me. Do not contact me again.')}
                      className="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs rounded"
                    >
                      + Opt-Out Hard Stop
                    </button>
                  </div>

                  {replyResult && (
                    <div className="p-3 bg-blue-50/60 border border-blue-200 rounded-lg space-y-1.5 text-xs animate-fade-in">
                      <div className="flex items-center justify-between font-semibold text-blue-900">
                        <span>Classification: {replyResult.classification}</span>
                        <span>Sentiment: {replyResult.sentiment}</span>
                      </div>
                      <div className="text-slate-600">
                        Next Recommended Step: <strong>{replyResult.recommended_next_step}</strong> | Updated Plan Status: <strong>{replyResult.updated_plan_status}</strong>
                      </div>
                      {replyResult.action_requested && (
                        <div className="text-amber-800 font-medium">
                          Detected Action Requested: {replyResult.action_requested}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )}
            </div>
          )}

          {/* TAB 2: ACTIVE PLAN & STEPS */}
          {activeTab === 'PLAN' && (
            <div className="space-y-6">
              {activePlan ? (
                <div className="space-y-4">
                  <div className="p-4 rounded-xl bg-slate-50 border border-slate-200 space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-mono text-slate-500">{activePlan.plan_id}</span>
                      <span className={`px-2 py-0.5 text-xs font-bold rounded-full ${
                        activePlan.status === 'COMPLETED' ? 'bg-emerald-100 text-emerald-800' :
                        activePlan.status === 'REQUIRES_APPROVAL' ? 'bg-amber-100 text-amber-800' :
                        activePlan.status === 'REPLANNING' ? 'bg-purple-100 text-purple-800' :
                        'bg-blue-100 text-blue-800'
                      }`}>
                        {activePlan.status}
                      </span>
                    </div>
                    <h3 className="text-sm font-semibold text-slate-800">{activePlan.goal}</h3>
                    <p className="text-xs text-slate-600 leading-relaxed">{activePlan.current_state_summary}</p>
                    <div className="flex items-center gap-4 text-xs text-slate-500 pt-1">
                      <span>Confidence: <strong>{(activePlan.confidence_score * 100).toFixed(0)}%</strong></span>
                      <span>Autonomy Level: <strong>{activePlan.autonomy_level}</strong></span>
                      <span>Waiting State: <strong>{activePlan.waiting_state || 'NONE'}</strong></span>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                      Ordered Plan Execution Steps
                    </h4>
                    {(stateData?.active_plan_steps || []).map((step, idx) => (
                      <div key={step.step_id || idx} className="p-3 bg-white border border-slate-200 rounded-lg flex items-center justify-between text-xs">
                        <div className="flex items-center gap-3">
                          <span className="w-5 h-5 rounded-full bg-slate-100 text-slate-700 font-bold flex items-center justify-center shrink-0">
                            {step.step_number || idx + 1}
                          </span>
                          <div>
                            <div className="font-semibold text-slate-800">{step.title || step.action_type}</div>
                            <div className="text-slate-500 text-[11px]">{step.description}</div>
                          </div>
                        </div>
                        <span className={`px-2 py-0.5 font-medium rounded ${
                          step.status === 'COMPLETED' ? 'bg-emerald-50 text-emerald-700' :
                          step.status === 'AWAITING_APPROVAL' ? 'bg-amber-50 text-amber-700' :
                          'bg-slate-100 text-slate-600'
                        }`}>
                          {step.status}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="p-8 text-center border-2 border-dashed border-slate-200 rounded-xl">
                  <Layers className="w-8 h-8 text-slate-300 mx-auto mb-2" />
                  <p className="text-sm font-medium text-slate-600">No active autonomous plan</p>
                  <p className="text-xs text-slate-400 mt-1">Plans are generated automatically when operational disruptions occur.</p>
                </div>
              )}
            </div>
          )}

          {/* TAB 3: POLICY & COMMUNICATION PREFERENCES */}
          {activeTab === 'PREFERENCES' && (
            <form onSubmit={handleSavePreferences} className="space-y-6">
              <div className="p-4 rounded-xl bg-slate-50 border border-slate-200 space-y-4">
                <div>
                  <h4 className="text-sm font-semibold text-slate-800">Authoritative Communication Policy</h4>
                  <p className="text-xs text-slate-500 mt-0.5">
                    Governs contact safety, frequency limits, channel selection, and opt-out hard stops.
                  </p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">Preferred Channel</label>
                    <select
                      value={channel}
                      onChange={(e) => setChannel(e.target.value)}
                      className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs bg-white focus:ring-2 focus:ring-blue-500 outline-hidden"
                    >
                      <option value="EMAIL">Email (Verified Contact)</option>
                      <option value="IN_APP">In-App Notification</option>
                      <option value="SMS">SMS / Messaging (Policy Controlled)</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-700 mb-1">Max Automated Follow-Ups</label>
                    <input
                      type="number"
                      min="1"
                      max="10"
                      value={maxFollowups}
                      onChange={(e) => setMaxFollowups(e.target.value)}
                      className="w-full px-3 py-2 border border-slate-300 rounded-lg text-xs bg-white focus:ring-2 focus:ring-blue-500 outline-hidden"
                    />
                  </div>
                </div>

                <div className="space-y-3 pt-2">
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={businessHours}
                      onChange={(e) => setBusinessHours(e.target.checked)}
                      className="rounded text-blue-600 focus:ring-blue-500"
                    />
                    <span className="text-xs font-medium text-slate-700">Enforce Business Hours Delivery Only</span>
                  </label>

                  <div className="p-3 bg-amber-50/60 border border-amber-200 rounded-lg space-y-2">
                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={optOut}
                        onChange={(e) => setOptOut(e.target.checked)}
                        className="rounded text-rose-600 focus:ring-rose-500"
                      />
                      <span className="text-xs font-bold text-rose-800 flex items-center gap-1">
                        <Ban className="w-3.5 h-3.5" />
                        Customer Opt-Out (Hard Stop Policy)
                      </span>
                    </label>
                    <p className="text-[11px] text-slate-600">
                      When enabled, the AI and Go enforcement boundary will reject all proactive communications with decision <code>STOP</code>.
                    </p>
                    {optOut && (
                      <input
                        type="text"
                        value={optOutReason}
                        onChange={(e) => setOptOutReason(e.target.value)}
                        placeholder="Reason for opt-out (e.g. client unsubscribed)"
                        className="w-full px-3 py-1.5 border border-amber-300 rounded text-xs bg-white outline-hidden"
                      />
                    )}
                  </div>
                </div>

                <div className="pt-2 flex justify-end">
                  <button
                    type="submit"
                    disabled={prefSaving}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg text-xs font-semibold hover:bg-blue-700 transition-colors disabled:opacity-50"
                  >
                    {prefSaving ? 'Saving...' : 'Save Communication Policy'}
                  </button>
                </div>
              </div>
            </form>
          )}

          {/* TAB 4: COMMUNICATION HISTORY */}
          {activeTab === 'HISTORY' && (
            <div className="space-y-3">
              <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                Audited Communication Records
              </h4>
              {recentRecords.length > 0 ? (
                recentRecords.map((rec) => (
                  <div key={rec.id} className="p-4 border border-slate-200 rounded-xl bg-white space-y-2 text-xs">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-slate-800">Record #{rec.id}</span>
                        <span className="text-slate-400">•</span>
                        <span className="font-semibold text-slate-600">{rec.event_type}</span>
                        <span className="text-slate-400">•</span>
                        <span className="text-slate-500">{rec.channel}</span>
                      </div>
                      <span className={`px-2 py-0.5 font-bold rounded ${
                        rec.status === 'SENT' ? 'bg-emerald-50 text-emerald-700' :
                        rec.status === 'REQUIRES_APPROVAL' ? 'bg-amber-50 text-amber-700' :
                        rec.status === 'STOPPED' ? 'bg-rose-50 text-rose-700' :
                        'bg-slate-100 text-slate-600'
                      }`}>
                        {rec.status}
                      </span>
                    </div>
                    <div className="font-medium text-slate-700">{rec.subject}</div>
                    <div className="text-slate-500 flex items-center justify-between pt-1 border-t border-slate-100">
                      <span>Recipient: {rec.recipient_name} ({rec.recipient_email})</span>
                      <span>Created: {new Date(rec.created_at).toLocaleString()}</span>
                    </div>
                    {rec.customer_response && (
                      <div className="p-2 bg-slate-50 rounded border border-slate-200 text-slate-700">
                        <strong>Reply:</strong> "{rec.customer_response}" ({rec.response_classification || 'UNCLASSIFIED'})
                      </div>
                    )}
                  </div>
                ))
              ) : (
                <div className="p-8 text-center text-slate-400 text-xs">
                  No communication records recorded yet.
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
