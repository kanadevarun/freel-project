import React, { useState, useEffect, useRef } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import {
  Sparkles,
  X,
  Send,
  CheckCircle2,
  AlertTriangle,
  FileText,
  Ship,
  CreditCard,
  FolderOpen,
  Users,
  ShieldCheck,
  Clock,
  ExternalLink,
  Copy,
  Check,
  Maximize2,
  Minimize2,
  RefreshCw,
  Layers,
  ArrowRight,
  MessageSquare,
  HelpCircle
} from 'lucide-react';
import { copilotService } from '../../services/copilotService';
import './AICopilotDrawer.css';

// Contextual prompt suggestions based on current active module
const MODULE_SUGGESTIONS = {
  DASHBOARD: [
    "Summarize today's cross-module operational bottlenecks",
    "Which pending approvals require immediate attention?",
    "Explain our current shipment delivery timeline risks"
  ],
  SHIPMENTS: [
    "Explain operational status of visible shipments",
    "Which shipments have milestone delays or detention risks?",
    "Draft a carrier check-in inquiry for delayed voyages"
  ],
  TRACKING: [
    "Summarize live container milestone updates",
    "Identify any vessels operating behind schedule"
  ],
  EXCEPTIONS: [
    "Explain operational risk and root causes of open exceptions",
    "Recommend resolution actions for high-severity issues"
  ],
  INVOICES: [
    "Summarize overdue invoices and total outstanding amount",
    "Which customers have aging receivables exceeding 30 days?",
    "Draft a professional customer payment reminder message"
  ],
  RFQS: [
    "Analyze open RFQs and spot inquiry volume",
    "Which RFQs are missing critical route or cargo details?"
  ],
  QUOTATIONS: [
    "Explain expiring quotations and recommended follow-ups",
    "Which quotations have narrow profit margin risks?"
  ],
  CONTRACTS: [
    "Summarize active freight agreements and upcoming expirations",
    "Which contracts have flagged compliance terms or liability gaps?"
  ],
  COMPLIANCE: [
    "Summarize recent contract compliance audit reviews",
    "Which carriers have missing certification documents?"
  ],
  APPROVALS: [
    "Summarize pending high-risk approval requests",
    "Explain the business justification for pending operational actions"
  ],
  NOTIFICATIONS: [
    "Summarize critical escalated alerts across operations",
    "Which alerts need immediate manager acknowledgement?"
  ],
  CUSTOMERS: [
    "Summarize customer account health and active freight volume",
    "Which accounts have pending inquiries awaiting responses?"
  ],
  LEADS: [
    "Identify leads with the highest AI qualification scores",
    "Draft an initial freight inquiry outreach message"
  ]
};

export default function AICopilotDrawer({ isOpen, onClose }) {
  const location = useLocation();
  const navigate = useNavigate();

  // Active drawer state
  const [activeTab, setActiveTab] = useState('chat'); // 'chat' | 'history'
  const [isWide, setIsWide] = useState(false);
  const [sessionId, setSessionId] = useState('');
  const [sessions, setSessions] = useState([]);
  const [messages, setMessages] = useState([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [copiedIndex, setCopiedIndex] = useState(null);
  const [actionHistory, setActionHistory] = useState([]);
  const [executingActionId, setExecutingActionId] = useState(null);
  const [actionFeedback, setActionFeedback] = useState(null);

  const messagesEndRef = useRef(null);

  // Derive active module from current URL
  const getActiveModule = () => {
    const path = location.pathname.toLowerCase();
    if (path.includes('/shipments')) return 'SHIPMENTS';
    if (path.includes('/tracking')) return 'TRACKING';
    if (path.includes('/exceptions')) return 'EXCEPTIONS';
    if (path.includes('/invoices')) return 'INVOICES';
    if (path.includes('/rfqs') || path.includes('/rfq')) return 'RFQS';
    if (path.includes('/quotations') || path.includes('/quote')) return 'QUOTATIONS';
    if (path.includes('/contracts')) return 'CONTRACTS';
    if (path.includes('/compliance')) return 'COMPLIANCE';
    if (path.includes('/approvals')) return 'APPROVALS';
    if (path.includes('/notifications')) return 'NOTIFICATIONS';
    if (path.includes('/customers')) return 'CUSTOMERS';
    if (path.includes('/leads')) return 'LEADS';
    return 'DASHBOARD';
  };

  const currentModule = getActiveModule();

  // Scroll to bottom of message list
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isLoading]);

  // Load sessions when drawer opens
  useEffect(() => {
    if (isOpen) {
      loadSessions();
      loadActionHistory();
    }
  }, [isOpen]);

  // Close drawer on Escape key
  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  const loadSessions = async () => {
    try {
      const res = await copilotService.listSessions(10);
      if (res && res.sessions) {
        setSessions(res.sessions);
      }
    } catch (err) {
      console.warn('Failed to load copilot sessions:', err);
    }
  };

  const loadActionHistory = async () => {
    try {
      const res = await copilotService.listActions(sessionId, 20);
      if (res && res.actions) {
        setActionHistory(res.actions);
      }
    } catch (err) {
      console.warn('Failed to load action history:', err);
    }
  };

  const handleStartNewSession = () => {
    setSessionId('');
    setMessages([]);
    setError(null);
  };

  const handleSelectSession = async (sId) => {
    try {
      setIsLoading(true);
      setSessionId(sId);
      const res = await copilotService.listMessages(sId);
      if (res && res.messages) {
        // Map backend message format to UI state
        const mapped = res.messages.map((m) => {
          let facts = [];
          let sources = [];
          let proposals = [];
          try {
            if (m.confirmed_facts) facts = JSON.parse(m.confirmed_facts);
            if (m.source_references) sources = JSON.parse(m.source_references);
            if (m.action_proposals) proposals = JSON.parse(m.action_proposals);
          } catch (e) {}

          return {
            id: m.id,
            role: m.role,
            content: m.content,
            confirmedFacts: facts,
            sourceReferences: sources,
            actionProposals: proposals,
            draftContent: m.draft_content,
            draftType: m.draft_type,
            confidence: m.confidence_score,
            createdAt: m.created_at,
          };
        });
        setMessages(mapped);
      }
    } catch (err) {
      setError('Failed to load conversation history.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSendMessage = async (queryText) => {
    const text = queryText || inputValue.trim();
    if (!text || isLoading) return;

    setError(null);
    setInputValue('');

    // Add user message optimistically
    const userMsg = {
      role: 'user',
      content: text,
      createdAt: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, userMsg]);
    setIsLoading(true);

    try {
      const response = await copilotService.chat({
        sessionId: sessionId || undefined,
        currentModule,
        currentRoute: location.pathname,
        query: text,
      });

      // Update sessionId if new
      if (response.correlation_id && !sessionId) {
        // Refresh sessions
        loadSessions();
      }

      const aiMsg = {
        role: 'assistant',
        content: response.answer,
        confirmedFacts: response.confirmed_facts || [],
        sourceReferences: response.source_references || [],
        signals: response.signals || [],
        aiInterpretation: response.ai_interpretation || '',
        recommendations: response.recommendations || [],
        suggestedFollowups: response.suggested_followups || [],
        draftContent: response.draft_content || null,
        draftType: response.draft_type || null,
        actionProposals: response.action_proposals || [],
        confidence: response.confidence || 0.95,
        requiresApproval: response.requires_approval || false,
        missingInformation: response.missing_information || [],
        createdAt: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, aiMsg]);
    } catch (err) {
      console.error('Copilot Chat Error:', err);
      setError(err?.message || 'AI Copilot encountered an unexpected error. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopyDraft = (text, idx) => {
    navigator.clipboard.writeText(text);
    setCopiedIndex(idx);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  const handleExecuteProposal = async (proposal) => {
    setExecutingActionId(proposal.action_title);
    setActionFeedback(null);
    try {
      const res = await copilotService.executeAction({
        actionType: proposal.action_type,
        actionTitle: proposal.action_title,
        actionPayload: proposal.payload || {},
        sessionId,
        reason: 'User executed Copilot action recommendation',
      });

      setActionFeedback({
        type: res.status === 'PENDING_APPROVAL' ? 'warning' : 'success',
        message: res.result_summary || `Action submitted: ${res.status}`,
        approvalId: res.approval_id,
      });

      loadActionHistory();
    } catch (err) {
      setActionFeedback({
        type: 'error',
        message: err?.message || 'Failed to execute proposed action.',
      });
    } finally {
      setExecutingActionId(null);
    }
  };

  const handleNavigateSource = (source) => {
    if (source.url) {
      navigate(source.url);
      return;
    }
    const type = (source.record_type || '').toUpperCase();
    const id = source.record_id;
    switch (type) {
      case 'SHIPMENT':
        navigate(`/dashboard/shipments`);
        break;
      case 'INVOICE':
        navigate(`/dashboard/invoices`);
        break;
      case 'CONTRACT':
        navigate(`/dashboard/contracts`);
        break;
      case 'RFQ':
        navigate(`/dashboard/rfqs`);
        break;
      case 'QUOTATION':
        navigate(`/dashboard/quotations`);
        break;
      case 'CUSTOMER':
        navigate(`/dashboard/customers`);
        break;
      case 'APPROVAL':
        navigate(`/dashboard/approvals`);
        break;
      default:
        break;
    }
  };

  if (!isOpen) return null;

  const currentSuggestions = MODULE_SUGGESTIONS[currentModule] || MODULE_SUGGESTIONS.DASHBOARD;

  return (
    <div className="copilot-drawer-backdrop" onClick={onClose}>
      <div
        className={`copilot-drawer ${isWide ? 'drawer-wide' : ''}`}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-label="LogisticsHQ AI Copilot"
      >
        {/* ── Header ── */}
        <div className="copilot-header">
          <div className="copilot-header-top">
            <div className="copilot-title-group">
              <div className="copilot-icon-badge">
                <Sparkles size={18} className="copilot-sparkle-icon" />
              </div>
              <div>
                <h3 className="copilot-title">LogisticsHQ Copilot</h3>
                <span className="copilot-subtitle">Context-Aware Freight Assistant</span>
              </div>
            </div>

            <div className="copilot-header-actions">
              <button
                className="copilot-ctrl-btn"
                title={isWide ? 'Standard view' : 'Wide view'}
                onClick={() => setIsWide(!isWide)}
              >
                {isWide ? <Minimize2 size={16} /> : <Maximize2 size={16} />}
              </button>
              <button
                className="copilot-ctrl-btn copilot-close-btn"
                title="Close Copilot"
                aria-label="Close Copilot"
                onClick={onClose}
              >
                <X size={18} />
              </button>
            </div>
          </div>

          {/* Module Context Indicator */}
          <div className="copilot-context-bar">
            <div className="copilot-context-pill">
              <span className="context-dot" />
              <span className="context-label">Active Context:</span>
              <strong className="context-module">{currentModule}</strong>
              <span className="context-path">({location.pathname})</span>
            </div>

            {sessions.length > 0 && (
              <div className="copilot-session-picker">
                <select
                  value={sessionId}
                  onChange={(e) => (e.target.value ? handleSelectSession(e.target.value) : handleStartNewSession())}
                  className="copilot-session-select"
                >
                  <option value="">+ New Conversation</option>
                  {sessions.map((s) => (
                    <option key={s.session_id} value={s.session_id}>
                      {s.title}
                    </option>
                  ))}
                </select>
              </div>
            )}
          </div>

          {/* Navigation Tabs */}
          <div className="copilot-tabs">
            <button
              className={`copilot-tab-btn ${activeTab === 'chat' ? 'active' : ''}`}
              onClick={() => setActiveTab('chat')}
            >
              <MessageSquare size={14} />
              <span>Copilot Chat</span>
            </button>
            <button
              className={`copilot-tab-btn ${activeTab === 'history' ? 'active' : ''}`}
              onClick={() => {
                setActiveTab('history');
                loadActionHistory();
              }}
            >
              <Layers size={14} />
              <span>Action Audit ({actionHistory.length})</span>
            </button>
          </div>
        </div>

        {/* ── Content Body ── */}
        {activeTab === 'chat' ? (
          <div className="copilot-body">
            {/* Action Feedback Banner */}
            {actionFeedback && (
              <div className={`copilot-banner banner-${actionFeedback.type}`}>
                <div className="banner-content">
                  {actionFeedback.type === 'warning' && <AlertTriangle size={16} />}
                  {actionFeedback.type === 'success' && <CheckCircle2 size={16} />}
                  <span>{actionFeedback.message}</span>
                </div>
                {actionFeedback.approvalId && (
                  <button
                    className="banner-link-btn"
                    onClick={() => {
                      onClose();
                      navigate('/dashboard/approvals');
                    }}
                  >
                    View in Approvals <ArrowRight size={12} />
                  </button>
                )}
              </div>
            )}

            {/* Messages Area */}
            <div className="copilot-messages-container">
              {messages.length === 0 ? (
                <div className="copilot-empty-state">
                  <div className="empty-sparkle-box">
                    <Sparkles size={32} className="text-blue-600" />
                  </div>
                  <h4 className="empty-title">How can I assist you in {currentModule}?</h4>
                  <p className="empty-desc">
                    I am grounded in your authorized organization data. I can summarize visible records,
                    explain operational risk, detect missing details, prepare communication drafts, or submit actions to the Action Center.
                  </p>

                  <div className="empty-suggestions-section">
                    <span className="suggestions-header">Suggested Questions for {currentModule}:</span>
                    <div className="empty-suggestions-list">
                      {currentSuggestions.map((q, idx) => (
                        <button
                          key={idx}
                          className="suggestion-chip-btn"
                          onClick={() => handleSendMessage(q)}
                        >
                          <HelpCircle size={13} className="text-blue-500" />
                          <span>{q}</span>
                        </button>
                      ))}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="copilot-messages-list">
                  {messages.map((msg, idx) => (
                    <div key={idx} className={`copilot-message-row msg-${msg.role}`}>
                      {msg.role === 'user' ? (
                        <div className="user-message-bubble">
                          <p>{msg.content}</p>
                        </div>
                      ) : (
                        <div className="assistant-message-card">
                          {/* Grounding Header */}
                          <div className="assistant-meta-header">
                            <span className="assistant-badge">
                              <Sparkles size={12} /> LogisticsHQ Copilot
                            </span>
                            {msg.confidence && (
                              <span className="confidence-pill" title="Verification confidence against MariaDB state">
                                <CheckCircle2 size={11} className="text-emerald-600" />
                                {Math.round(msg.confidence * 100)}% Grounded Facts
                              </span>
                            )}
                          </div>

                          {/* Confirmed Backend Facts Box */}
                          {msg.confirmedFacts && msg.confirmedFacts.length > 0 && (
                            <div className="confirmed-facts-box">
                              <div className="facts-header">
                                <ShieldCheck size={14} className="text-emerald-600" />
                                <span>Verified Backend Facts:</span>
                              </div>
                              <ul className="facts-list">
                                {msg.confirmedFacts.map((fact, fIdx) => (
                                  <li key={fIdx}>{fact}</li>
                                ))}
                              </ul>
                            </div>
                          )}

                          {/* Source Record References */}
                          {msg.sourceReferences && msg.sourceReferences.length > 0 && (
                            <div className="source-refs-bar">
                              <span className="source-refs-label">Sources:</span>
                              <div className="source-chips">
                                {msg.sourceReferences.map((ref, rIdx) => (
                                  <button
                                    key={rIdx}
                                    className="source-chip"
                                    onClick={() => handleNavigateSource(ref)}
                                    title={`View ${ref.record_type} #${ref.record_id}`}
                                  >
                                    <ExternalLink size={11} />
                                    <span>{ref.title || `${ref.record_type} #${ref.record_id}`}</span>
                                  </button>
                                ))}
                              </div>
                            </div>
                          )}

                          {/* Main Response Content */}
                          <div className="assistant-text-content">
                            <p>{msg.content}</p>
                          </div>

                          {/* Operational Signals / Interpretation */}
                          {msg.aiInterpretation && (
                            <div className="ai-interpretation-box">
                              <span className="interpretation-label">Analysis:</span>
                              <p>{msg.aiInterpretation}</p>
                            </div>
                          )}

                          {/* Generated Draft Box (Copyable) */}
                          {msg.draftContent && (
                            <div className="copilot-draft-card">
                              <div className="draft-header">
                                <div className="draft-tag">
                                  <FileText size={14} className="text-blue-600" />
                                  <span>{msg.draftType || 'Operational Draft'}</span>
                                </div>
                                <button
                                  className="draft-copy-btn"
                                  onClick={() => handleCopyDraft(msg.draftContent, idx)}
                                >
                                  {copiedIndex === idx ? (
                                    <>
                                      <Check size={13} className="text-emerald-600" /> Copied!
                                    </>
                                  ) : (
                                    <>
                                      <Copy size={13} /> Copy Draft
                                    </>
                                  )}
                                </button>
                              </div>
                              <div className="draft-body">
                                <pre>{msg.draftContent}</pre>
                              </div>
                            </div>
                          )}

                          {/* Controlled Action Proposals */}
                          {msg.actionProposals && msg.actionProposals.length > 0 && (
                            <div className="action-proposals-section">
                              <span className="proposals-title">Recommended Controlled Actions:</span>
                              <div className="proposals-list">
                                {msg.actionProposals.map((prop, pIdx) => {
                                  const isExecuting = executingActionId === prop.action_title;
                                  return (
                                    <div key={pIdx} className="proposal-item">
                                      <div className="proposal-info">
                                        <strong className="proposal-name">{prop.action_title}</strong>
                                        <p className="proposal-desc">{prop.description}</p>
                                        <div className="proposal-meta">
                                          {prop.requires_approval ? (
                                            <span className="meta-badge badge-approval">
                                              <AlertTriangle size={11} /> Requires Human Approval Gate
                                            </span>
                                          ) : (
                                            <span className="meta-badge badge-safe">
                                              <CheckCircle2 size={11} /> Safe Action System Execution
                                            </span>
                                          )}
                                        </div>
                                      </div>
                                      <button
                                        className={`proposal-act-btn ${prop.requires_approval ? 'btn-approval' : 'btn-direct'}`}
                                        disabled={isExecuting}
                                        onClick={() => handleExecuteProposal(prop)}
                                      >
                                        {isExecuting ? (
                                          <RefreshCw size={13} className="animate-spin" />
                                        ) : prop.requires_approval ? (
                                          'Submit for Approval'
                                        ) : (
                                          'Execute Action'
                                        )}
                                      </button>
                                    </div>
                                  );
                                })}
                              </div>
                            </div>
                          )}

                          {/* Suggested Followups */}
                          {msg.suggestedFollowups && msg.suggestedFollowups.length > 0 && (
                            <div className="followups-row">
                              <span className="followups-label">Suggested next:</span>
                              {msg.suggestedFollowups.map((sug, sIdx) => (
                                <button
                                  key={sIdx}
                                  className="followup-pill-btn"
                                  onClick={() => handleSendMessage(sug)}
                                >
                                  {sug}
                                </button>
                              ))}
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  ))}

                  {isLoading && (
                    <div className="copilot-loading-row">
                      <div className="copilot-loading-bubble">
                        <Sparkles size={16} className="copilot-spinner" />
                        <span>Copilot analyzing authorized {currentModule} data...</span>
                      </div>
                    </div>
                  )}

                  {error && (
                    <div className="copilot-error-box">
                      <AlertTriangle size={16} className="text-red-500" />
                      <span>{error}</span>
                    </div>
                  )}

                  <div ref={messagesEndRef} />
                </div>
              )}
            </div>

            {/* ── Input Bar ── */}
            <div className="copilot-input-container">
              <form
                className="copilot-input-form"
                onSubmit={(e) => {
                  e.preventDefault();
                  handleSendMessage();
                }}
              >
                <input
                  type="text"
                  className="copilot-input-field"
                  placeholder={`Ask anything about ${currentModule.toLowerCase()}...`}
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  disabled={isLoading}
                />
                <button
                  type="submit"
                  className="copilot-send-btn"
                  disabled={!inputValue.trim() || isLoading}
                  title="Send message"
                >
                  <Send size={16} />
                </button>
              </form>
              <div className="copilot-input-footer">
                <span>Grounded strictly in your authorized organization records. Consequential actions require approval.</span>
              </div>
            </div>
          </div>
        ) : (
          /* ── Action Audit History Tab ── */
          <div className="copilot-history-body">
            <div className="history-header">
              <h4>Controlled Action History</h4>
              <p>Audit trail of all actions requested or executed via AI Copilot</p>
            </div>

            {actionHistory.length === 0 ? (
              <div className="history-empty">
                <Layers size={24} className="text-slate-400" />
                <p>No actions have been executed in this session yet.</p>
              </div>
            ) : (
              <div className="history-list">
                {actionHistory.map((act) => {
                  const isPending = act.status === 'PENDING_APPROVAL';
                  const isExecuted = act.status === 'EXECUTED';
                  return (
                    <div key={act.id} className="history-card">
                      <div className="history-card-top">
                        <strong className="history-action-title">{act.action_title}</strong>
                        <span className={`status-pill pill-${act.status.toLowerCase()}`}>
                          {act.status}
                        </span>
                      </div>
                      <p className="history-summary">{act.result_summary}</p>
                      <div className="history-card-footer">
                        <span className="history-date">
                          {act.created_at ? new Date(act.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}
                        </span>
                        {act.approval_id && (
                          <button
                            className="history-view-link"
                            onClick={() => {
                              onClose();
                              navigate('/dashboard/approvals');
                            }}
                          >
                            Approval #{act.approval_id} <ArrowRight size={11} />
                          </button>
                        )}
                        {act.recommendation_id && (
                          <button
                            className="history-view-link"
                            onClick={() => {
                              onClose();
                              navigate('/dashboard/recommendations');
                            }}
                          >
                            Recommendation #{act.recommendation_id} <ArrowRight size={11} />
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
