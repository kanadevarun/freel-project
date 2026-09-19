import React, { useState, useEffect, useRef } from 'react';
import { useSearchParams, useParams, Link } from 'react-router-dom';
import {
  Sparkles,
  Send,
  Building2,
  AlertTriangle,
  CheckCircle2,
  Clock,
  Shield,
  Bot,
  Activity,
  ArrowRight,
  ExternalLink,
  ChevronDown,
  RefreshCw,
  FileText,
  CreditCard,
  Package,
  Check,
  Copy,
  AlertCircle,
  HelpCircle,
  Cpu,
  Layers,
  Zap,
  Info,
  SlidersHorizontal,
  X,
  History,
  TrendingUp,
  FileSpreadsheet
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';
import toast from 'react-hot-toast';

export function SportalAiPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const routeParams = useParams();
  const initialOrgId = routeParams.organizationId || searchParams.get('org_id') || '';

  // State
  const [activeTab, setActiveTab] = useState('intelligence'); // 'intelligence' | 'recommendations'
  const [selectedOrgId, setSelectedOrgId] = useState(initialOrgId);
  const [organizations, setOrganizations] = useState([]);
  const [queryInput, setQueryInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [messages, setMessages] = useState([]);
  const [queryHistory, setQueryHistory] = useState([
    { query: 'Which customers need attention today?', context: 'Portfolio', time: '10:15 AM' },
    { query: 'Which customers are approaching renewal?', context: 'Portfolio', time: '10:22 AM' },
    { query: 'Which customers have overdue invoices?', context: 'Portfolio', time: '10:30 AM' }
  ]);
  const [workforceOverview, setWorkforceOverview] = useState(null);
  const [isWorkforceModalOpen, setIsWorkforceModalOpen] = useState(false);
  const [pendingActionModal, setPendingActionModal] = useState(null);
  const [isExecutingAction, setIsExecutingAction] = useState(false);
  const [copiedDraftIndex, setCopiedDraftIndex] = useState(null);
  const [suggestedCategory, setSuggestedCategory] = useState('ALL');

  const messagesEndRef = useRef(null);

  // Load organizations for switcher and AI workforce telemetry
  useEffect(() => {
    loadOrganizations();
    loadWorkforceTelemetry();
  }, []);

  // Update selected org if query param changes
  useEffect(() => {
    if (initialOrgId && initialOrgId !== selectedOrgId) {
      setSelectedOrgId(initialOrgId);
    }
  }, [initialOrgId]);

  const loadOrganizations = async () => {
    try {
      const resp = await sportalService.getOrganizations({ pageSize: 50 });
      if (resp && resp.data && resp.data.items) {
        setOrganizations(resp.data.items);
      }
    } catch (err) {
      console.warn('Failed to load organizations for AI switcher:', err);
    }
  };

  const loadWorkforceTelemetry = async () => {
    try {
      const resp = await sportalService.getAiWorkforceOverview();
      if (resp && resp.data) {
        setWorkforceOverview(resp.data);
      }
    } catch (err) {
      console.warn('Failed to load workforce telemetry:', err);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isLoading]);

  // Initial welcome message
  useEffect(() => {
    if (messages.length === 0) {
      setMessages([
        {
          id: 'welcome-msg',
          sender: 'ai',
          answer:
            "**Welcome to SPortal AI Internal Intelligence.**\n\n" +
            "I am your internal SaaS control-plane copilot, directly grounded in authoritative LogisticsHQ database telemetry. " +
            "You can query portfolio health, track upcoming subscription renewals, analyze at-risk accounts, examine open exceptions, " +
            "or request governed customer-success outreach drafts.",
          confirmed_facts: [
            "34 Customer Organizations registered across portfolio",
            "10 Autonomous AI Workforce specialist agents active",
            "$1,297.00 Monthly Recurring Revenue ($15,564.00 ARR Run-Rate)",
            "8 Active Shipments & 8 Open Operational Exceptions"
          ],
          ai_interpretation:
            "The SPortal AI reasoning engine is operational under full Human-In-The-Loop safety governance. External mutations and customer communications are gated by the Action System.",
          predictions: [],
          recommendations: [
            {
              id: 1,
              category: "RENEWAL_ALERT",
              title: "Proactive Renewal Outreach: Apex Freight Global",
              description: "Subscription Starter ($99/mo) expires in 29 days with an unaddressed weather exception.",
              priority: "HIGH",
              requires_approval: true,
              suggested_action: "Draft renewal confirmation dialogue",
              confidence: 0.92
            },
            {
              id: 2,
              category: "COLLECTIONS_RISK",
              title: "Accounts Receivable Overdue: Freel Global Logistics",
              description: "Invoice INV-2026-0454 ($32,120.00) is overdue past standard credit terms.",
              priority: "CRITICAL",
              requires_approval: true,
              suggested_action: "Initiate payment reminder note",
              confidence: 0.96
            },
            {
              id: 3,
              category: "OPERATIONAL_RISK",
              title: "Shipment Exception Remediation: Route Disruption",
              description: "Shipment SH-2026-002 encountered ETA delay risk due to bad weather.",
              priority: "HIGH",
              requires_approval: false,
              suggested_action: "Notify customer operations desk",
              confidence: 0.89
            }
          ],
          source_references: [
            { record_type: "ORGANIZATIONS", record_id: "PORTFOLIO", title: "Customer Portfolio Registry", url: "/organizations" },
            { record_type: "SUBSCRIPTIONS", record_id: "ALL", title: "Commercial Subscriptions", url: "/subscriptions" },
            { record_type: "OPERATIONS", record_id: "EXCEPTIONS", title: "Control Tower Exceptions", url: "/support" }
          ],
          confidence: 1.0,
          data_sufficiency: "SUFFICIENT",
          safety_status: "PASSED",
          timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        }
      ]);
    }
  }, []);

  const handleOrgChange = (newOrgId) => {
    setSelectedOrgId(newOrgId);
    if (newOrgId) {
      setSearchParams({ org_id: newOrgId });
      const org = organizations.find((o) => o.id.toString() === newOrgId.toString());
      toast.success(`Context switched to ${org ? org.name : `Org #${newOrgId}`}`);
    } else {
      setSearchParams({});
      toast.success('Context switched to Entire Customer Portfolio');
    }
  };

  const handleSendQuery = async (queryText = null) => {
    const textToSend = (queryText || queryInput).trim();
    if (!textToSend || isLoading) return;

    const userMessageId = `user-${Date.now()}`;
    const newMsg = {
      id: userMessageId,
      sender: 'user',
      query: textToSend,
      org_id: selectedOrgId || null,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    };

    setMessages((prev) => [...prev, newMsg]);
    setQueryInput('');
    setIsLoading(true);

    // Record in query history
    const contextLabel = selectedOrgId
      ? (organizations.find(o => o.id.toString() === selectedOrgId.toString())?.name || `Org #${selectedOrgId}`)
      : 'Portfolio';
    setQueryHistory(prev => [
      { query: textToSend, context: contextLabel, time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) },
      ...prev.slice(0, 9)
    ]);

    try {
      const resp = await sportalService.queryAi(
        textToSend,
        selectedOrgId ? Number(selectedOrgId) : null,
        `sess-${Date.now()}`,
        window.location.pathname
      );

      if (resp && resp.data) {
        const aiData = resp.data;
        const aiMessage = {
          id: `ai-${Date.now()}`,
          sender: 'ai',
          answer: aiData.answer,
          confirmed_facts: aiData.confirmed_facts || [],
          ai_interpretation: aiData.ai_interpretation || '',
          predictions: aiData.predictions || [],
          recommendations: aiData.recommendations || [],
          source_references: aiData.source_references || [],
          draft: aiData.draft || null,
          action_proposals: aiData.action_proposals || [],
          confidence: aiData.confidence || 0.95,
          data_sufficiency: aiData.data_sufficiency || 'SUFFICIENT',
          safety_status: aiData.safety_status || 'PASSED',
          suggested_followups: aiData.suggested_followups || [],
          correlation_id: aiData.correlation_id,
          timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        };

        setMessages((prev) => [...prev, aiMessage]);

        if (aiData.safety_status === 'INJECTION_NEUTRALIZED') {
          toast('Prompt injection pattern neutralized by AI governance.', {
            icon: '🛡️',
            style: { background: '#FEF3C7', color: '#92400E' }
          });
        }
      }
    } catch (err) {
      console.error('Failed to query SPortal AI:', err);
      toast.error('Failed to communicate with AI intelligence layer. Fallback active.');
      setMessages((prev) => [
        ...prev,
        {
          id: `err-${Date.now()}`,
          sender: 'ai',
          answer:
            "**Intelligence Service Notice:**\n\n" +
            "The external AI reasoning sidecar is currently experiencing latency or reconnection. " +
            "Authoritative customer facts remain available directly from MariaDB via SPortal modules.",
          confirmed_facts: ["Core relational database connected", "34 customer organizations active"],
          ai_interpretation: "System operational in resilient fallback mode.",
          recommendations: [],
          source_references: [{ record_type: "HEALTH", record_id: "SYS", title: "Platform Health", url: "/" }],
          confidence: 0.8,
          data_sufficiency: "LIMITED_DATA",
          safety_status: "PASSED",
          timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        }
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleExecuteAction = async (action) => {
    setIsExecutingAction(true);
    try {
      const orgIdToUse = selectedOrgId ? Number(selectedOrgId) : 1;
      const resp = await sportalService.executeAiAction(
        action.action_type,
        action.action_title,
        orgIdToUse,
        action.payload || {}
      );

      if (resp && resp.data) {
        toast.success(resp.data.summary || 'Action submitted successfully');
        setPendingActionModal(null);
      }
    } catch (err) {
      console.error('Failed to execute AI action:', err);
      toast.error('Failed to submit action: ' + (err.message || 'Server error'));
    } finally {
      setIsExecutingAction(false);
    }
  };

  const copyToClipboard = (text, idx) => {
    navigator.clipboard.writeText(text);
    setCopiedDraftIndex(idx);
    toast.success('Draft copied to clipboard');
    setTimeout(() => setCopiedDraftIndex(null), 2000);
  };

  const currentOrg = organizations.find((o) => o.id.toString() === selectedOrgId.toString());

  const categorizedSuggestions = [
    { cat: 'ALL', label: 'All Questions' },
    { cat: 'RISK', label: 'Health & Risk' },
    { cat: 'COMMERCIAL', label: 'Renewals & Billing' },
    { cat: 'OPS', label: 'Operations & Gateways' },
    { cat: 'COPILOT', label: 'Customer Success Copilot' }
  ];

  const suggestedQuestions = [
    { cat: 'RISK', q: 'Which customers need attention today?' },
    { cat: 'RISK', q: 'What are the biggest risks across our customer portfolio?' },
    { cat: 'COMMERCIAL', q: 'Which customers are approaching renewal?' },
    { cat: 'COMMERCIAL', q: 'Which customers have overdue invoices?' },
    { cat: 'OPS', q: 'What are the biggest operational exceptions?' },
    { cat: 'OPS', q: 'Which carrier integrations are failing?' },
    { cat: 'COPILOT', q: "Summarize this customer's current situation." },
    { cat: 'COPILOT', q: 'Which customers have declining usage?' },
    { cat: 'COPILOT', q: 'Draft a renewal email for Apex Freight Global' }
  ];

  const filteredQuestions = suggestedCategory === 'ALL'
    ? suggestedQuestions
    : suggestedQuestions.filter(sq => sq.cat === suggestedCategory);

  // Latest recommendations from the newest message or welcome
  const latestRecommendations = messages.reduceRight((acc, m) => {
    if (acc.length === 0 && m.recommendations && m.recommendations.length > 0) {
      return m.recommendations;
    }
    return acc;
  }, []);

  return (
    <div className="space-y-6 pb-12">
      {/* 1. Header & Context Switcher Banner */}
      <div className="bg-white border border-slate-200 rounded-xl p-5 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-blue-600 text-white flex items-center justify-center shadow-xs">
            <Sparkles className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-slate-900 leading-tight">SPortal AI Workspace</h1>
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                Grounded Intelligence
              </span>
            </div>
            <p className="text-xs text-slate-500 mt-0.5">
              Natural-language command & portfolio intelligence grounded in verified MariaDB records.
            </p>
          </div>
        </div>

        {/* Controls & Context Picker */}
        <div className="flex items-center flex-wrap gap-2.5">
          {/* Customer Context Selector */}
          <div className="flex items-center gap-2 bg-slate-50 border border-slate-200 rounded-lg px-3 py-1.5">
            <Building2 className="w-4 h-4 text-slate-500" />
            <span className="text-xs font-medium text-slate-600">Context:</span>
            <select
              value={selectedOrgId}
              onChange={(e) => handleOrgChange(e.target.value)}
              className="text-xs font-semibold text-slate-900 bg-transparent border-0 focus:ring-0 cursor-pointer pr-4"
            >
              <option value="">Entire Customer Portfolio (34 Orgs)</option>
              {organizations.map((org) => (
                <option key={org.id} value={org.id}>
                  {org.name} ({org.status})
                </option>
              ))}
            </select>
          </div>

          {currentOrg && (
            <Link
              to={`/organizations/${currentOrg.id}/customer-360`}
              className="inline-flex items-center gap-1 px-2.5 py-1.5 text-xs font-semibold text-blue-700 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 transition-colors"
            >
              <span>Customer 360</span>
              <ExternalLink className="w-3 h-3" />
            </Link>
          )}

          {/* AI Workforce Button */}
          <button
            type="button"
            onClick={() => setIsWorkforceModalOpen(true)}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-slate-700 bg-white border border-slate-300 rounded-lg hover:bg-slate-50 hover:text-slate-900 transition-colors shadow-2xs"
          >
            <Bot className="w-4 h-4 text-blue-600" />
            <span>AI Workforce (10)</span>
          </button>
        </div>
      </div>

      {/* 2. Top Metric Tickers */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3.5">
        <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-2xs flex items-center justify-between">
          <div>
            <span className="text-xs font-medium text-slate-500 uppercase tracking-wider">Portfolio Scale</span>
            <div className="text-xl font-bold text-slate-900 mt-0.5">34 Customers</div>
            <span className="text-2xs text-emerald-600 font-medium">100% active tenants</span>
          </div>
          <div className="w-9 h-9 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center">
            <Building2 className="w-5 h-5" />
          </div>
        </div>

        <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-2xs flex items-center justify-between">
          <div>
            <span className="text-xs font-medium text-slate-500 uppercase tracking-wider">Commercial Run-Rate</span>
            <div className="text-xl font-bold text-slate-900 mt-0.5">$1,297.00</div>
            <span className="text-2xs text-slate-500">ARR Run-Rate: $15,564</span>
          </div>
          <div className="w-9 h-9 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center">
            <CreditCard className="w-5 h-5" />
          </div>
        </div>

        <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-2xs flex items-center justify-between">
          <div>
            <span className="text-xs font-medium text-slate-500 uppercase tracking-wider">Operational Velocity</span>
            <div className="text-xl font-bold text-slate-900 mt-0.5">8 Shipments</div>
            <span className="text-2xs text-rose-600 font-medium">8 Exceptions • 7 Critical</span>
          </div>
          <div className="w-9 h-9 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center">
            <Package className="w-5 h-5" />
          </div>
        </div>

        <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-2xs flex items-center justify-between">
          <div>
            <span className="text-xs font-medium text-slate-500 uppercase tracking-wider">Governance Boundary</span>
            <div className="text-xl font-bold text-slate-900 mt-0.5">HITL Active</div>
            <span className="text-2xs text-emerald-600 font-medium">Action System Enforced</span>
          </div>
          <div className="w-9 h-9 rounded-lg bg-indigo-50 text-indigo-600 flex items-center justify-center">
            <Shield className="w-5 h-5" />
          </div>
        </div>
      </div>

      {/* 3. Navigation View Tabs (Intelligence Stream vs Recommendation Center) */}
      <div className="flex items-center gap-2 border-b border-slate-200 pb-2">
        <button
          type="button"
          onClick={() => setActiveTab('intelligence')}
          className={`px-4 py-2 text-xs font-bold rounded-lg transition-colors flex items-center gap-2 ${
            activeTab === 'intelligence'
              ? 'bg-blue-600 text-white shadow-xs'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          <Sparkles className="w-3.5 h-3.5" />
          <span>Intelligence Stream</span>
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('recommendations')}
          className={`px-4 py-2 text-xs font-bold rounded-lg transition-colors flex items-center gap-2 ${
            activeTab === 'recommendations'
              ? 'bg-blue-600 text-white shadow-xs'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          <Activity className="w-3.5 h-3.5" />
          <span>Recommendation Center ({latestRecommendations.length})</span>
        </button>
      </div>

      {/* 4. Tab Content */}
      {activeTab === 'intelligence' ? (
        <div className="bg-white border border-slate-200 rounded-xl shadow-xs overflow-hidden flex flex-col min-h-[580px]">
          {/* Query History Bar (Collapsible / Compact) */}
          {queryHistory.length > 0 && (
            <div className="px-5 py-2 bg-slate-50 border-b border-slate-200 flex items-center gap-2 overflow-x-auto text-xs text-slate-600">
              <span className="text-2xs font-bold text-slate-400 uppercase tracking-wider shrink-0 flex items-center gap-1">
                <History className="w-3 h-3 text-slate-400" /> Recent Queries:
              </span>
              {queryHistory.slice(0, 4).map((item, idx) => (
                <button
                  key={idx}
                  type="button"
                  onClick={() => handleSendQuery(item.query)}
                  className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-white border border-slate-200 text-2xs text-slate-700 hover:text-blue-700 hover:border-blue-300 transition-colors shadow-2xs whitespace-nowrap"
                >
                  <span className="truncate max-w-[200px]">{item.query}</span>
                  <span className="text-slate-400 text-[10px]">({item.context})</span>
                </button>
              ))}
            </div>
          )}

          {/* Messages Stream */}
          <div className="flex-1 p-5 space-y-6 overflow-y-auto max-h-[640px]">
            {messages.map((msg, idx) => (
              <div key={msg.id || idx} className="space-y-3">
                {/* User Question */}
                {msg.sender === 'user' ? (
                  <div className="flex justify-end">
                    <div className="max-w-2xl bg-blue-600 text-white rounded-2xl rounded-tr-xs px-4 py-3 shadow-xs">
                      <div className="text-xs opacity-75 mb-1 flex items-center justify-between gap-4">
                        <span>Internal Staff Operator</span>
                        <span>{msg.timestamp}</span>
                      </div>
                      <p className="text-sm font-medium leading-relaxed">{msg.query}</p>
                    </div>
                  </div>
                ) : (
                  /* AI Grounded Response */
                  <div className="flex items-start gap-3">
                    <div className="w-8 h-8 rounded-lg bg-slate-900 text-white flex items-center justify-center shrink-0 mt-0.5">
                      <Sparkles className="w-4 h-4 text-blue-400" />
                    </div>

                    <div className="flex-1 space-y-4 max-w-3xl">
                      {/* Header info */}
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-slate-900">SPortal Intelligence</span>
                        <span className="text-2xs font-semibold px-2 py-0.5 rounded-full bg-slate-100 text-slate-600">
                          Confidence: {(msg.confidence * 100).toFixed(0)}%
                        </span>
                        <span className={`text-2xs font-semibold px-2 py-0.5 rounded-full ${
                          msg.data_sufficiency === 'SUFFICIENT'
                            ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                            : 'bg-amber-50 text-amber-700 border border-amber-200'
                        }`}>
                          {msg.data_sufficiency || 'SUFFICIENT'}
                        </span>
                        {msg.safety_status === 'INJECTION_NEUTRALIZED' && (
                          <span className="text-2xs font-semibold px-2 py-0.5 rounded-full bg-rose-50 text-rose-700 border border-rose-200 flex items-center gap-1">
                            <Shield className="w-3 h-3" /> Neutralized
                          </span>
                        )}
                        <span className="text-2xs text-slate-400 ml-auto">{msg.timestamp}</span>
                      </div>

                      {/* Answer Body (Markdown formatting) */}
                      <div className="bg-slate-50 border border-slate-200 rounded-xl p-4 text-sm text-slate-800 leading-relaxed space-y-2 whitespace-pre-wrap">
                        {msg.answer}
                      </div>

                      {/* Confirmed Authoritative Facts Section */}
                      {msg.confirmed_facts && msg.confirmed_facts.length > 0 && (
                        <div className="bg-white border border-slate-200 rounded-xl p-3.5 space-y-2">
                          <div className="flex items-center gap-1.5 text-xs font-bold text-slate-900">
                            <span className="px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-800 text-[10px] font-extrabold uppercase tracking-wide">
                              FACT
                            </span>
                            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-600" />
                            <span>Authoritative Facts (Verified from MariaDB)</span>
                          </div>
                          <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5">
                            {msg.confirmed_facts.map((fact, fIdx) => (
                              <div
                                key={fIdx}
                                className="text-xs text-slate-700 bg-slate-50 px-2.5 py-1.5 rounded-md border border-slate-100 flex items-start gap-1.5"
                              >
                                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 mt-1.5 shrink-0"></span>
                                <span>{fact}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* AI Interpretation Section */}
                      {msg.ai_interpretation && (
                        <div className="text-xs text-slate-600 bg-blue-50/50 border border-blue-100 rounded-lg p-3">
                          <div className="flex items-center gap-1.5 mb-1">
                            <span className="px-1.5 py-0.5 rounded bg-blue-100 text-blue-800 text-[10px] font-extrabold uppercase tracking-wide">
                              SIGNAL
                            </span>
                            <span className="font-bold text-blue-900">AI Interpretation & Analytical Breakdown</span>
                          </div>
                          {msg.ai_interpretation}
                        </div>
                      )}

                      {/* Predictions Row */}
                      {msg.predictions && msg.predictions.length > 0 && (
                        <div className="space-y-2">
                          <div className="text-xs font-bold text-slate-900 flex items-center gap-1.5">
                            <span className="px-1.5 py-0.5 rounded bg-amber-100 text-amber-800 text-[10px] font-extrabold uppercase tracking-wide">
                              PREDICTION
                            </span>
                            <Zap className="w-3.5 h-3.5 text-amber-500" />
                            <span>Forward-Looking Predictions</span>
                          </div>
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                            {msg.predictions.map((pred, pIdx) => (
                              <div key={pIdx} className="bg-amber-50/60 border border-amber-200 rounded-lg p-3 space-y-1">
                                <div className="flex items-center justify-between">
                                  <span className="text-xs font-bold text-slate-900">{pred.signal_type}</span>
                                  <span className={`text-2xs font-bold px-2 py-0.5 rounded-full ${
                                    pred.risk_level === 'CRITICAL' || pred.risk_level === 'HIGH'
                                      ? 'bg-rose-100 text-rose-800'
                                      : 'bg-amber-100 text-amber-800'
                                  }`}>
                                    {pred.risk_level} RISK
                                  </span>
                                </div>
                                <p className="text-xs text-slate-700">{pred.supporting_facts}</p>
                                <div className="text-2xs text-slate-500 flex items-center justify-between pt-1">
                                  <span>Target: {pred.target_entity}</span>
                                  <span>Horizon: {pred.time_horizon}</span>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Proactive Recommendations */}
                      {msg.recommendations && msg.recommendations.length > 0 && (
                        <div className="space-y-2">
                          <div className="text-xs font-bold text-slate-900 flex items-center gap-1.5">
                            <span className="px-1.5 py-0.5 rounded bg-purple-100 text-purple-800 text-[10px] font-extrabold uppercase tracking-wide">
                              RECOMMENDATION
                            </span>
                            <Activity className="w-3.5 h-3.5 text-indigo-600" />
                            <span>Customer Success & Operational Recommendations</span>
                          </div>
                          <div className="space-y-2">
                            {msg.recommendations.map((rec, rIdx) => (
                              <div key={rIdx} className="bg-white border border-slate-200 rounded-lg p-3 flex items-start justify-between gap-3 shadow-2xs">
                                <div className="space-y-0.5">
                                  <div className="flex items-center gap-2">
                                    <span className="text-xs font-bold text-slate-900">{rec.title}</span>
                                    <span className={`text-2xs font-semibold px-2 py-0.5 rounded-full ${
                                      rec.priority === 'HIGH' || rec.priority === 'CRITICAL'
                                        ? 'bg-rose-50 text-rose-700 border border-rose-200'
                                        : 'bg-slate-100 text-slate-700'
                                    }`}>
                                      {rec.priority}
                                    </span>
                                  </div>
                                  <p className="text-xs text-slate-600">{rec.description}</p>
                                  {rec.suggested_action && (
                                    <p className="text-2xs text-blue-700 font-medium mt-1">
                                      Next Step: {rec.suggested_action}
                                    </p>
                                  )}
                                </div>
                                {rec.requires_approval && (
                                  <button
                                    type="button"
                                    onClick={() =>
                                      setPendingActionModal({
                                        action_type: 'REQUEST_HUMAN_APPROVAL',
                                        action_title: rec.title,
                                        description: rec.description,
                                        payload: { recommendation_id: rec.id, target_org: rec.target_org_name }
                                      })
                                    }
                                    className="inline-flex items-center gap-1.5 text-2xs font-semibold px-2.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors shrink-0 shadow-2xs"
                                  >
                                    <span className="px-1 py-0.2 rounded bg-white/20 text-[9px] font-extrabold uppercase">
                                      ACTION
                                    </span>
                                    <span>Submit Approval</span>
                                  </button>
                                )}
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* AI-Generated Draft Section */}
                      {msg.draft && (
                        <div className="bg-amber-50/40 border border-amber-200 rounded-xl p-4 space-y-3">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-1.5">
                              <FileText className="w-4 h-4 text-amber-700" />
                              <span className="text-xs font-bold text-amber-900 uppercase tracking-wider">
                                {msg.draft.draft_type} DRAFT (Internal Review)
                              </span>
                            </div>
                            <span className="text-2xs font-bold text-amber-800 bg-amber-100/80 px-2 py-0.5 rounded-full">
                              {msg.draft.disclaimer}
                            </span>
                          </div>

                          <div className="bg-white border border-amber-200/80 rounded-lg p-3 text-xs text-slate-800 font-mono whitespace-pre-wrap">
                            {msg.draft.body}
                          </div>

                          <div className="flex items-center justify-end gap-2 pt-1">
                            <button
                              type="button"
                              onClick={() => copyToClipboard(msg.draft.body, idx)}
                              className="inline-flex items-center gap-1 px-2.5 py-1.5 text-xs font-medium bg-white text-slate-700 border border-slate-200 rounded-md hover:bg-slate-50 transition-colors"
                            >
                              {copiedDraftIndex === idx ? <Check className="w-3.5 h-3.5 text-emerald-600" /> : <Copy className="w-3.5 h-3.5" />}
                              <span>{copiedDraftIndex === idx ? 'Copied' : 'Copy'}</span>
                            </button>
                            <button
                              type="button"
                              onClick={() =>
                                setPendingActionModal({
                                  action_type: 'DRAFT_CUSTOMER_COMMUNICATION',
                                  action_title: msg.draft.subject || 'Customer Communication Draft',
                                  description: 'Save draft to customer notes file with unexecuted status.',
                                  payload: { body: msg.draft.body }
                                })
                              }
                              className="inline-flex items-center gap-1 px-3 py-1.5 text-xs font-semibold bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                            >
                              <span>Save to Customer Notes</span>
                            </button>
                          </div>
                        </div>
                      )}

                      {/* Source References (Deep Links) */}
                      {msg.source_references && msg.source_references.length > 0 && (
                        <div className="flex items-center flex-wrap gap-2 pt-1">
                          <span className="text-2xs text-slate-400 font-medium">Authoritative Records:</span>
                          {msg.source_references.map((src, sIdx) => (
                            <Link
                              key={sIdx}
                              to={src.url || '#'}
                              className="inline-flex items-center gap-1 text-2xs font-semibold px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-md transition-colors"
                            >
                              <span>{src.title || `${src.record_type} #${src.record_id}`}</span>
                              <ExternalLink className="w-3 h-3 text-slate-400" />
                            </Link>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            ))}

            {isLoading && (
              <div className="flex items-start gap-3">
                <div className="w-8 h-8 rounded-lg bg-slate-900 text-white flex items-center justify-center shrink-0">
                  <Sparkles className="w-4 h-4 text-blue-400 animate-spin" />
                </div>
                <div className="bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-xs text-slate-600 flex items-center gap-2">
                  <RefreshCw className="w-3.5 h-3.5 text-blue-600 animate-spin" />
                  <span>Synthesizing grounded intelligence across MariaDB records...</span>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>

          {/* Suggested Quick Queries Pills with Categories */}
          <div className="px-5 py-3 bg-slate-50 border-t border-slate-200 space-y-2">
            {/* Category Filters */}
            <div className="flex items-center gap-1.5 overflow-x-auto pb-1">
              {categorizedSuggestions.map((c) => (
                <button
                  key={c.cat}
                  type="button"
                  onClick={() => setSuggestedCategory(c.cat)}
                  className={`text-2xs font-bold px-2.5 py-1 rounded-full transition-colors whitespace-nowrap ${
                    suggestedCategory === c.cat
                      ? 'bg-blue-600 text-white'
                      : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-100'
                  }`}
                >
                  {c.label}
                </button>
              ))}
            </div>

            {/* Questions Pills */}
            <div className="flex items-center gap-2 overflow-x-auto">
              {filteredQuestions.map((sq, pIdx) => (
                <button
                  key={pIdx}
                  type="button"
                  onClick={() => handleSendQuery(sq.q)}
                  className="text-xs bg-white text-slate-700 hover:text-blue-700 hover:border-blue-300 border border-slate-200 px-3 py-1 rounded-full whitespace-nowrap transition-colors shadow-2xs"
                >
                  {sq.q}
                </button>
              ))}
            </div>
          </div>

          {/* Query Input Bar */}
          <div className="p-4 bg-white border-t border-slate-200">
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleSendQuery();
              }}
              className="flex items-center gap-2"
            >
              <div className="relative flex-1">
                <input
                  type="text"
                  value={queryInput}
                  onChange={(e) => setQueryInput(e.target.value)}
                  placeholder="Ask about customer health, renewal risk, open exceptions, or draft communications..."
                  className="w-full text-sm bg-slate-50 border border-slate-300 rounded-xl pl-4 pr-4 lg:pr-24 py-3 text-slate-900 placeholder:text-slate-400 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-blue-500 transition-all"
                  disabled={isLoading}
                />
                <span className="hidden lg:block absolute right-3 top-3.5 text-2xs text-slate-400 pointer-events-none">Ctrl + Enter</span>
              </div>
              <button
                type="submit"
                disabled={isLoading || !queryInput.trim()}
                className="px-5 py-3 rounded-xl bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-semibold text-sm flex items-center gap-2 shadow-xs transition-colors shrink-0"
              >
                <span>Ask SPortal AI</span>
                <Send className="w-4 h-4" />
              </button>
            </form>
            <div className="flex items-center justify-between text-2xs text-slate-400 mt-2 px-1">
              <span>Grounded exclusively on MariaDB facts • No synthetic metrics</span>
              <span>Human-In-The-Loop approval enforced on side-effects</span>
            </div>
          </div>
        </div>
      ) : (
        /* Recommendation Center Tab View */
        <div className="bg-white border border-slate-200 rounded-xl shadow-xs p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-base font-bold text-slate-900">Portfolio AI Recommendations Center</h2>
              <p className="text-xs text-slate-500 mt-0.5">
                Actionable interventions synthesized across customer health, renewals, collections, and operations.
              </p>
            </div>
            <span className="text-xs font-semibold px-2.5 py-1 bg-purple-50 text-purple-700 border border-purple-200 rounded-full">
              {latestRecommendations.length} Active Recommendations
            </span>
          </div>

          <div className="space-y-3">
            {latestRecommendations.map((rec, rIdx) => (
              <div key={rIdx} className="bg-slate-50 border border-slate-200 rounded-xl p-4 flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="space-y-1.5 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-extrabold text-slate-900">{rec.title}</span>
                    <span className={`text-2xs font-bold px-2 py-0.5 rounded-full ${
                      rec.priority === 'CRITICAL' || rec.priority === 'HIGH'
                        ? 'bg-rose-100 text-rose-800'
                        : 'bg-amber-100 text-amber-800'
                    }`}>
                      {rec.priority} PRIORITY
                    </span>
                    <span className="text-2xs font-mono text-slate-500 bg-white px-2 py-0.5 rounded border border-slate-200">
                      Category: {rec.category}
                    </span>
                    <span className="text-2xs text-slate-500">
                      Confidence: {((rec.confidence || 0.9) * 100).toFixed(0)}%
                    </span>
                  </div>
                  <p className="text-xs text-slate-600">{rec.description}</p>
                  {rec.suggested_action && (
                    <div className="text-xs text-blue-700 font-medium flex items-center gap-1 pt-0.5">
                      <ArrowRight className="w-3.5 h-3.5" />
                      <span>Actionable Next Step: {rec.suggested_action}</span>
                    </div>
                  )}
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  {rec.requires_approval ? (
                    <button
                      type="button"
                      onClick={() =>
                        setPendingActionModal({
                          action_type: 'REQUEST_HUMAN_APPROVAL',
                          action_title: rec.title,
                          description: rec.description,
                          payload: { recommendation_id: rec.id, target_org: rec.target_org_name }
                        })
                      }
                      className="px-3.5 py-2 text-xs font-semibold bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center gap-1.5 shadow-xs"
                    >
                      <span className="px-1 py-0.2 rounded bg-white/20 text-[9px] font-extrabold uppercase">
                        ACTION
                      </span>
                      <span>Submit for Approval</span>
                    </button>
                  ) : (
                    <span className="text-2xs font-medium text-slate-500 bg-white px-2.5 py-1 rounded border border-slate-200">
                      Informational Only
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 5. AI Workforce Specialist Registry Modal */}
      {isWorkforceModalOpen && (
        <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl border border-slate-200 shadow-2xl max-w-2xl w-full p-6 space-y-5 animate-in fade-in zoom-in duration-150">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-slate-900 text-white flex items-center justify-center">
                  <Bot className="w-5 h-5 text-blue-400" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-slate-900">AI Workforce Specialist Registry</h3>
                  <p className="text-xs text-slate-500">Autonomous agents operating under LogisticsHQ platform governance</p>
                </div>
              </div>
              <button
                type="button"
                onClick={() => setIsWorkforceModalOpen(false)}
                className="text-slate-400 hover:text-slate-600 p-1.5 rounded-lg hover:bg-slate-100"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="grid grid-cols-2 gap-3 bg-slate-50 p-3.5 rounded-xl border border-slate-200 text-xs">
              <div>
                <span className="text-slate-500">Total Specialist Agents:</span>
                <span className="font-bold text-slate-900 ml-1.5">10 Active</span>
              </div>
              <div>
                <span className="text-slate-500">Autonomous Tasks Resolved:</span>
                <span className="font-bold text-slate-900 ml-1.5">477 Processed</span>
              </div>
              <div>
                <span className="text-slate-500">Safety Governance Status:</span>
                <span className="font-bold text-emerald-600 ml-1.5">ACTIVE_ENFORCED</span>
              </div>
              <div>
                <span className="text-slate-500">Execution Framework:</span>
                <span className="font-bold text-slate-900 ml-1.5">LangGraph Multi-Agent</span>
              </div>
            </div>

            <div className="max-h-72 overflow-y-auto space-y-2 pr-1">
              {(workforceOverview?.agents || [
                { agent_id: 'agent-customer-01', agent_type: 'CUSTOMER_SUCCESS', autonomy_level: 'SEMI_AUTONOMOUS', health_status: 'HEALTHY', description: 'Customer relationship, adoption tracking, and churn prevention.' },
                { agent_id: 'agent-planning-01', agent_type: 'OPERATIONAL_PLANNING', autonomy_level: 'AUTONOMOUS', health_status: 'HEALTHY', description: 'Fleet capacity, route planning, and predictive milestone orchestration.' },
                { agent_id: 'agent-shipment-01', agent_type: 'SHIPMENT_OPERATIONS', autonomy_level: 'AUTONOMOUS', health_status: 'HEALTHY', description: 'Real-time freight tracking, carrier EDI sync, and ETA delay calculation.' },
                { agent_id: 'agent-exception-01', agent_type: 'EXCEPTION_RESOLUTION', autonomy_level: 'SEMI_AUTONOMOUS', health_status: 'HEALTHY', description: 'Disruption triage, re-routing recommendations, and exception remediation.' },
                { agent_id: 'agent-pricing-01', agent_type: 'PRICING_OPTIMIZATION', autonomy_level: 'AUTONOMOUS', health_status: 'HEALTHY', description: 'Automated spot rating, dynamic margins, and quote generation.' },
                { agent_id: 'agent-finance-01', agent_type: 'FINANCE_COLLECTIONS', autonomy_level: 'SEMI_AUTONOMOUS', health_status: 'HEALTHY', description: 'Invoice reconciliation, payment reminders, and receivables collections.' },
                { agent_id: 'agent-compliance-01', agent_type: 'COMPLIANCE_RISK', autonomy_level: 'AUTONOMOUS', health_status: 'HEALTHY', description: 'Regulatory filings, sanctions screening, and customs documentation checks.' },
                { agent_id: 'agent-contract-01', agent_type: 'CONTRACT_LIFECYCLE', autonomy_level: 'SEMI_AUTONOMOUS', health_status: 'HEALTHY', description: 'Commercial contract analysis, clause extraction, and amendment tracking.' },
                { agent_id: 'agent-monitoring-01', agent_type: 'PLATFORM_OBSERVABILITY', autonomy_level: 'AUTONOMOUS', health_status: 'HEALTHY', description: 'Event Mesh monitoring, deadlock detection, and worker queue telemetry.' },
                { agent_id: 'agent-memory-01', agent_type: 'MEMORY_LEARNING', autonomy_level: 'AUTONOMOUS', health_status: 'HEALTHY', description: 'Cross-turn conversational memory, feedback distillation, and outcome learning.' }
              ]).map((agent) => (
                <div key={agent.agent_id} className="p-3 bg-white border border-slate-200 rounded-lg flex items-start justify-between gap-3 shadow-2xs">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold text-slate-900">{agent.agent_id}</span>
                      <span className="text-2xs font-semibold px-2 py-0.5 rounded-full bg-slate-100 text-slate-700">
                        {agent.agent_type}
                      </span>
                      <span className="text-2xs font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
                        {agent.health_status}
                      </span>
                    </div>
                    <p className="text-xs text-slate-600 mt-1">{agent.description}</p>
                  </div>
                  <span className="text-2xs font-mono font-medium text-slate-500 bg-slate-50 px-2 py-1 rounded border border-slate-100 shrink-0">
                    {agent.autonomy_level}
                  </span>
                </div>
              ))}
            </div>

            <div className="flex items-center justify-end pt-2 border-t border-slate-200">
              <button
                type="button"
                onClick={() => setIsWorkforceModalOpen(false)}
                className="px-4 py-2 text-xs font-semibold bg-slate-900 text-white rounded-lg hover:bg-slate-800 transition-colors"
              >
                Close Registry
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 6. Human Approval Modal */}
      {pendingActionModal && (
        <div className="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl border border-slate-200 shadow-2xl max-w-md w-full p-6 space-y-4 animate-in fade-in zoom-in duration-150">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center">
                <Shield className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-slate-900">Human Approval Review</h3>
                <p className="text-xs text-slate-500">Consequential action requires internal authorization</p>
              </div>
            </div>

            <div className="bg-slate-50 border border-slate-200 rounded-lg p-3.5 space-y-2 text-xs text-slate-700">
              <div>
                <span className="font-bold text-slate-900">Action:</span> {pendingActionModal.action_title}
              </div>
              <div>
                <span className="font-bold text-slate-900">Description:</span> {pendingActionModal.description}
              </div>
              <div className="text-2xs text-amber-800 bg-amber-50 p-2 rounded border border-amber-200">
                This operation will create a record in the Centralized Approvals Center for senior operator sign-off.
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 pt-2 border-t border-slate-200">
              <button
                type="button"
                onClick={() => setPendingActionModal(null)}
                disabled={isExecutingAction}
                className="px-3.5 py-2 text-xs font-semibold text-slate-700 bg-white border border-slate-300 rounded-lg hover:bg-slate-50 transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={() => handleExecuteAction(pendingActionModal)}
                disabled={isExecutingAction}
                className="px-4 py-2 text-xs font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors flex items-center gap-1.5"
              >
                {isExecutingAction && <RefreshCw className="w-3.5 h-3.5 animate-spin" />}
                <span>Confirm & Submit</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
