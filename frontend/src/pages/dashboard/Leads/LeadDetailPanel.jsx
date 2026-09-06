import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import PropTypes from 'prop-types';
import { 
  Mail, 
  Send, 
  Reply, 
  Sparkles, 
  CheckCircle2, 
  XCircle, 
  AlertCircle, 
  Clock, 
  User, 
  Trash2, 
  Edit3, 
  ChevronDown, 
  ChevronUp, 
  RefreshCw, 
  Inbox,
  AlertTriangle,
  ArrowRight,
  SendHorizontal,
  FileText
} from 'lucide-react';
import { updateLead, deleteLead, getLeadTimeline, retryClarificationEmail, getLeadInteractions, replyToInteraction, retryEmailInteraction, getEmailDraft, saveEmailDraft, deleteEmailDraft, getConnectedMailboxes } from '../../../services/leadsService';
import { getLeadOutreachActivity } from '../../../services/outreachService';
import StatusBadge from '../../../components/dashboard/StatusBadge';
import LocationPickerMap from '../../../components/dashboard/LocationPickerMap';
import { LEAD_STATUS } from './LeadsPage';



/** Shows a color-coded AI score badge. */
function AIScoreBadge({ score }) {
  if (!score) return null;
  const cls = score >= 70 ? 'ai-score-high' : score >= 40 ? 'ai-score-medium' : 'ai-score-low';
  return <span className={`ai-score-badge ${cls}`}>⚡ {score}/100</span>;
}

AIScoreBadge.propTypes = {
  score: PropTypes.number,
};

/** EmailContentFormatter handles email body truncation and collapses quoted text history. */
function EmailContentFormatter({ content }) {
  const [showMore, setShowMore] = useState(false);
  const [showQuoted, setShowQuoted] = useState(false);

  if (!content) return null;

  const lines = content.split('\n');
  const regularLines = [];
  const quotedLines = [];
  let isQuotedSection = false;

  for (let line of lines) {
    const trimmed = line.trim();
    const isQuotePrefix = trimmed.startsWith('>') || 
                          trimmed.toLowerCase().startsWith('on ') && trimmed.toLowerCase().includes('wrote:') ||
                          trimmed.startsWith('---') ||
                          trimmed.toLowerCase().startsWith('from:') ||
                          trimmed.toLowerCase().startsWith('sent:');
    
    if (isQuotePrefix) {
      isQuotedSection = true;
    }

    if (isQuotedSection) {
      quotedLines.push(line);
    } else {
      regularLines.push(line);
    }
  }

  const regularText = regularLines.join('\n');
  const quotedText = quotedLines.join('\n');

  const maxChars = 600;
  const isLong = regularText.length > maxChars;
  const displayedText = isLong && !showMore ? regularText.slice(0, maxChars) + '...' : regularText;

  return (
    <div className="email-formatted-content">
      <div className="email-main-body" style={{ whiteSpace: 'pre-line' }}>
        {displayedText}
      </div>

      {isLong && (
        <button 
          className="email-toggle-link"
          onClick={() => setShowMore(!showMore)}
        >
          {showMore ? 'Show less' : 'Show more'}
        </button>
      )}

      {quotedLines.length > 0 && (
        <div className="email-quoted-section">
          <button 
            className="email-toggle-link quote-toggle"
            onClick={() => setShowQuoted(!showQuoted)}
          >
            {showQuoted ? 'Hide quoted text' : 'Show quoted text'}
          </button>
          {showQuoted && (
            <div className="email-quoted-body" style={{ whiteSpace: 'pre-line' }}>
              {quotedText}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

EmailContentFormatter.propTypes = {
  content: PropTypes.string,
};

/** Helper to split timeline descriptions and inject styled inline status pills */
function formatTimelineDescription(desc) {
  if (!desc) return '';
  const statusKeywords = ['NEW', 'IN_PROGRESS', 'QUALIFIED', 'CONVERTED', 'REJECTED', 'ACTIVE'];
  const parts = desc.split(/(NEW|IN_PROGRESS|QUALIFIED|CONVERTED|REJECTED|ACTIVE)/g);
  return parts.map((part, index) => {
    if (statusKeywords.includes(part)) {
      let className = 'timeline-status-pill';
      if (part === 'NEW') className += ' pill-new';
      else if (part === 'IN_PROGRESS') className += ' pill-in-progress';
      else if (part === 'QUALIFIED') className += ' pill-qualified';
      else if (part === 'CONVERTED') className += ' pill-converted';
      else if (part === 'REJECTED') className += ' pill-rejected';
      else if (part === 'ACTIVE') className += ' pill-active';
      const label = part.replace('_', ' ');
      return (
        <span key={index} className={className}>
          {label}
        </span>
      );
    }
    return part;
  });
}

/** Helper to format date and time stamp cleanly */
function formatTimelineTime(timestamp) {
  if (!timestamp) return { dateStr: '', timeStr: '' };
  const date = new Date(timestamp);
  const dateStr = date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  });
  const timeStr = date.toLocaleTimeString(undefined, {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  });
  return { dateStr, timeStr };
}

/** CustomThreadSelector renders a modern interactive thread selection card dropdown */
function CustomThreadSelector({ threadGroups, selectedThreadId, onSelectThread }) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef(null);

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const activeGroup = threadGroups.find(g => g.thread_id === selectedThreadId) || threadGroups[0];
  if (!activeGroup) return null;

  const getCleanSubject = (subject) => {
    if (!subject) return 'General Inquiry';
    return subject.replace(/^re:\s*/i, '');
  };

  const formatLastActive = (timestamp) => {
    if (!timestamp) return '';
    return new Date(timestamp).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit'
    });
  };

  return (
    <div className="custom-thread-selector-wrapper" ref={containerRef} style={{ marginBottom: '20px' }}>
      <label className="thread-selector-label">SELECT CONVERSATION THREAD</label>
      <div className={`thread-selector-trigger ${isOpen ? 'open' : ''}`} onClick={() => setIsOpen(!isOpen)}>
        <div className="thread-trigger-left">
          <span className="thread-trigger-icon">💬</span>
          <div className="thread-trigger-info">
            <span className="thread-trigger-subject">{getCleanSubject(activeGroup.subject)}</span>
            <div className="thread-trigger-meta">
              <span className="thread-meta-pill count-pill">{activeGroup.interactions.length} messages</span>
              <span className="thread-meta-pill time-pill">Last active {formatLastActive(activeGroup.lastActivity)}</span>
            </div>
          </div>
        </div>
        <span className={`thread-trigger-chevron ${isOpen ? 'rotated' : ''}`}>▼</span>
      </div>

      {isOpen && (
        <div className="thread-selector-dropdown-menu">
          {threadGroups.map((group) => {
            const isSelected = group.thread_id === activeGroup.thread_id;
            return (
              <div
                key={group.thread_id}
                className={`thread-dropdown-item ${isSelected ? 'selected' : ''}`}
                onClick={() => {
                  onSelectThread(group.thread_id);
                  setIsOpen(false);
                }}
              >
                <div className="thread-item-content">
                  <div className="thread-item-header">
                    <span className="thread-item-subject">{getCleanSubject(group.subject)}</span>
                    {isSelected && <span className="thread-active-badge">✓ Active</span>}
                  </div>
                  <div className="thread-item-meta">
                    <span className="thread-meta-pill count-pill">{group.interactions.length} messages</span>
                    <span className="thread-meta-pill time-pill">Last active {formatLastActive(group.lastActivity)}</span>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

export default function LeadDetailPanel({ lead, initialTab, onClose, onLeadUpdated, onConvertToRFQ, onConvertToCustomer, onDirtyChange }) {
  const navigate = useNavigate();
  const [formState, setFormState] = useState({
    company_name: lead.company_name || '',
    contact_name: lead.contact_name || '',
    email: lead.email || '',
    phone: lead.phone || '',
    source: lead.source || '',
    notes: lead.notes || '',
    location: lead.location || '',
    status: lead.status || 'NEW',
    tags: lead.tags || [],
  });

  const [deleting, setDeleting] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [timeline, setTimeline] = useState([]);
  const [retryingMap, setRetryingMap] = useState({});
  const [loadingTimeline, setLoadingTimeline] = useState(false);
  const [showScrollHint, setShowScrollHint] = useState(false);
  const timelineScrollRef = useRef(null);

  const [activeTab, setActiveTab] = useState(initialTab || 'overview'); // 'overview', 'emails', 'timeline'

  useEffect(() => {
    if (initialTab) {
      setActiveTab(initialTab);
    }
  }, [initialTab]);
  const [interactions, setInteractions] = useState([]);
  const [outreachActivities, setOutreachActivities] = useState([]);
  const [loadingInteractions, setLoadingInteractions] = useState(false);
  const [selectedThreadId, setSelectedThreadId] = useState('');

  const mandatoryFields = [
    { key: 'origin_port', label: 'Origin' },
    { key: 'destination_port', label: 'Destination' },
    { key: 'cargo_description', label: 'Cargo Description' },
    { key: 'target_date', label: 'Ready Date' },
    { key: 'cargo_weight', label: 'Cargo Weight' },
    { key: 'cargo_volume', label: 'Cargo Volume' },
    { key: 'incoterms', label: 'Incoterms' }
  ];

  const getLatestRFQContext = () => {
    if (!interactions || interactions.length === 0) return {};
    const reversed = [...interactions].reverse();
    const found = reversed.find(i => {
      if (!i.partial_rfq_context) return false;
      if (typeof i.partial_rfq_context === 'string') {
        try {
          const parsed = JSON.parse(i.partial_rfq_context);
          return Object.keys(parsed).length > 0;
        } catch (e) {
          return false;
        }
      }
      return Object.keys(i.partial_rfq_context).length > 0;
    });
    if (!found) return {};
    if (typeof found.partial_rfq_context === 'string') {
      try {
        return JSON.parse(found.partial_rfq_context);
      } catch (e) {
        return {};
      }
    }
    return found.partial_rfq_context;
  };

  const getLatestRFQContextForThread = (threadId) => {
    if (!interactions || interactions.length === 0) return {};
    const threadInteractions = threadId
      ? interactions.filter(i => (i.thread_id || 'general') === threadId)
      : interactions;
    const reversed = [...threadInteractions].reverse();
    const found = reversed.find(i => {
      if (!i.partial_rfq_context) return false;
      if (typeof i.partial_rfq_context === 'string') {
        try {
          const parsed = JSON.parse(i.partial_rfq_context);
          return Object.keys(parsed).length > 0;
        } catch (e) {
          return false;
        }
      }
      return Object.keys(i.partial_rfq_context).length > 0;
    });
    if (!found) return {};
    if (typeof found.partial_rfq_context === 'string') {
      try {
        return JSON.parse(found.partial_rfq_context);
      } catch (e) {
        return {};
      }
    }
    return found.partial_rfq_context;
  };

  const rfqContext = getLatestRFQContext();

  const getMissingFields = (ctx) => {
    const missing = [];
    mandatoryFields.forEach(f => {
      const val = ctx[f.key];
      if (val === undefined || val === null || val === '') {
        missing.push(f.label);
      } else if (typeof val === 'number' && val <= 0) {
        missing.push(f.label);
      }
    });
    return missing;
  };

  const missingFields = getMissingFields(rfqContext);

  const latestWithRFQ = interactions && interactions.find(i => i.linked_rfq_id && i.linked_rfq_id > 0);
  const createdRFQId = lead?.linked_rfq_id || (latestWithRFQ ? latestWithRFQ.linked_rfq_id : null);
  const createdRFQNumber = lead?.linked_rfq_number || null;
  const isConverted = lead?.status === 'CONVERTED' || !!createdRFQId;
  const hasCreatedRFQ = isConverted;

  // Group interactions by thread_id
  const getThreadGroups = (list) => {
    const groups = {};
    list.forEach(item => {
      const tId = item.thread_id || 'general';
      if (!groups[tId]) {
        groups[tId] = {
          thread_id: tId,
          subject: item.subject || 'RFQ Inquiry',
          interactions: [],
          lastActivity: new Date(item.created_at).getTime()
        };
      }
      groups[tId].interactions.push(item);
      const time = new Date(item.created_at).getTime();
      if (time > groups[tId].lastActivity) {
        groups[tId].lastActivity = time;
      }
    });

    Object.values(groups).forEach(g => {
      g.interactions.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
      const firstInbound = g.interactions.find(i => i.direction === 'INBOUND');
      if (firstInbound && firstInbound.subject) {
        g.subject = firstInbound.subject;
      }
    });

    return Object.values(groups).sort((a, b) => b.lastActivity - a.lastActivity);
  };

  const threadGroups = getThreadGroups(interactions);

  useEffect(() => {
    if (interactions && interactions.length > 0) {
      const groups = getThreadGroups(interactions);
      if (groups.length > 0) {
        if (!selectedThreadId || !groups.find(g => g.thread_id === selectedThreadId)) {
          setSelectedThreadId(groups[0].thread_id);
        }
      }
    } else {
      setSelectedThreadId('');
    }
  }, [interactions]);

  const getConversationWorkflowState = (interactionsList, leadObj, threadId) => {
    const threadInteractions = threadId
      ? interactionsList.filter(i => (i.thread_id || 'general') === threadId)
      : interactionsList;

    const latestWithRFQ = interactionsList && interactionsList.find(i => i.linked_rfq_id && i.linked_rfq_id > 0);
    const createdRFQId = latestWithRFQ ? latestWithRFQ.linked_rfq_id : null;
    const hasCreatedRFQ = !!createdRFQId || leadObj?.status === 'CONVERTED';
    if (hasCreatedRFQ) {
      return { state: 'RFQ_CREATED', rfqId: createdRFQId };
    }

    const threadCtx = getLatestRFQContextForThread(threadId);

    const missing = [];
    mandatoryFields.forEach(f => {
      const val = threadCtx[f.key];
      if (val === undefined || val === null || val === '') {
        missing.push(f.label);
      } else if (typeof val === 'number' && val <= 0) {
        missing.push(f.label);
      }
    });

    const sortedInteractions = [...threadInteractions].sort(
      (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    );

    if (missing.length === 0 && sortedInteractions.length > 0) {
      return { state: 'RFQ_READY', missingFields: [] };
    }

    const inboundList = sortedInteractions.filter(i => i.direction === 'INBOUND');
    const outboundList = sortedInteractions.filter(i => i.direction === 'OUTBOUND' && i.status === 'SENT');

    const latestInbound = inboundList[inboundList.length - 1];
    const latestOutbound = outboundList[outboundList.length - 1];

    if (inboundList.length === 0) {
      return { state: 'NEW_INQUIRY', missingFields: missing };
    }

    if (!latestOutbound) {
      const hasAnyContext = Object.keys(threadCtx).length > 0;
      if (hasAnyContext && missing.length > 0) {
        return { state: 'INFORMATION_INCOMPLETE', missingFields: missing };
      }
      return { state: 'NEW_INQUIRY', missingFields: missing };
    }

    const latestInboundTime = latestInbound ? new Date(latestInbound.created_at).getTime() : 0;
    const latestOutboundTime = latestOutbound ? new Date(latestOutbound.created_at).getTime() : 0;

    if (latestOutboundTime > latestInboundTime) {
      return {
        state: 'WAITING_FOR_CUSTOMER',
        missingFields: missing,
        lastRequestTime: latestOutbound.created_at
      };
    } else {
      return {
        state: 'CUSTOMER_RESPONDED',
        missingFields: missing,
        lastCustomerReplyTime: latestInbound.created_at
      };
    }
  };

  const workflowState = getConversationWorkflowState(interactions, lead, selectedThreadId);
  
  // Composer, drafts, mailboxes and suggestion states
  const [availableMailboxes, setAvailableMailboxes] = useState([]);
  const [replyingToId, setReplyingToId] = useState(null); // interaction.id
  const [replyForm, setReplyForm] = useState({ from: '', to: '', cc: '', subject: '', body: '' });
  const [sendingReply, setSendingReply] = useState(false);
  const [sendError, setSendError] = useState(null);
  const [discardedSuggestionIds, setDiscardedSuggestionIds] = useState(new Set());
  const [expandedInterIds, setExpandedInterIds] = useState(new Set());

  // Load available organization connected mailboxes
  useEffect(() => {
    let active = true;
    const fetchMailboxes = async () => {
      try {
        const res = await getConnectedMailboxes();
        const list = res?.data || res || [];
        if (active && Array.isArray(list)) {
          setAvailableMailboxes(list);
          const primary = list.find(mb => mb.is_primary) || list[0];
          if (primary && primary.email) {
            setReplyForm(prev => ({ ...prev, from: prev.from || primary.email }));
          }
        }
      } catch (err) {
        console.error('Failed to load connected mailboxes:', err);
      }
    };
    fetchMailboxes();
    return () => { active = false; };
  }, []);

  const toggleExpandInter = (id) => {
    setExpandedInterIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const [expandedTimelineIdx, setExpandedTimelineIdx] = useState(null);

  const toggleExpandTimeline = (idx) => {
    setExpandedTimelineIdx(prev => (prev === idx ? null : idx));
  };
  
  // Draft persistence states
  const [draftStatus, setDraftStatus] = useState(''); // '', 'saving', 'saved', 'error'
  const autoSaveTimeoutRef = useRef(null);
  const isLoadingDraftRef = useRef(false);

  // Auto-save draft effect
  useEffect(() => {
    if (!replyingToId || !lead?.id || isLoadingDraftRef.current) return;
    
    // Check if the form matches the default empty state
    if (!replyForm.to && !replyForm.cc && !replyForm.subject && !replyForm.body) {
      return;
    }

    if (autoSaveTimeoutRef.current) {
      clearTimeout(autoSaveTimeoutRef.current);
    }

    setDraftStatus('saving');
    autoSaveTimeoutRef.current = setTimeout(async () => {
      try {
        await saveEmailDraft(lead.id, replyingToId, {
          from: replyForm.from,
          to: replyForm.to,
          cc: replyForm.cc,
          subject: replyForm.subject,
          body: replyForm.body
        });
        setDraftStatus('saved');
      } catch (err) {
        console.error('Failed to auto-save draft:', err);
        setDraftStatus('error');
      }
    }, 1500); // 1.5 seconds debounce

    return () => {
      if (autoSaveTimeoutRef.current) {
        clearTimeout(autoSaveTimeoutRef.current);
      }
    };
  }, [replyForm, replyingToId, lead?.id]);

  // Load interactions chronologically on mount / lead id change
  useEffect(() => {
    if (!lead?.id) return;
    let active = true;
    const fetchInteractions = async () => {
      setLoadingInteractions(true);
      try {
        const data = await getLeadInteractions(lead.id, 'asc');
        if (active) {
          setInteractions(data?.data || data || []);
        }
      } catch (err) {
        console.error('Failed to load interactions:', err);
      } finally {
        if (active) setLoadingInteractions(false);
      }
    };
    fetchInteractions();
    return () => {
      active = false;
    };
  }, [lead.id]);

  useEffect(() => {
    if (!lead?.id) return;
    let active = true;
    const fetchOutreach = async () => {
      try {
        const data = await getLeadOutreachActivity(lead.id);
        if (active) {
          setOutreachActivities(data || []);
        }
      } catch (err) {
        console.error('Failed to load lead outreach activities:', err);
      }
    };
    fetchOutreach();
    return () => {
      active = false;
    };
  }, [lead.id]);

  const handleStartNewEmail = async (templateType = 'CUSTOM', customBody = '', customSubject = '') => {
    setReplyingToId(0);
    setSendError(null);
    setDraftStatus('');
    isLoadingDraftRef.current = true;

    const defaultFrom = availableMailboxes.find(mb => mb.is_primary)?.email || availableMailboxes[0]?.email || '';

    try {
      const draftRes = await getEmailDraft(lead.id, 0);
      const draft = draftRes?.data || draftRes;
      if (draft && draft.body) {
        setReplyForm({
          from: draft.from || defaultFrom,
          to: draft.to || lead?.email || '',
          cc: draft.cc || '',
          subject: draft.subject || '',
          body: draft.body || ''
        });
        isLoadingDraftRef.current = false;
        return;
      }
    } catch (err) {
      console.error('No draft for new email:', err);
    }

    let sub = customSubject || `Inquiry regarding freight shipment - ${lead?.company_name || lead?.contact_name || ''}`;
    let body = customBody;

    if (!body) {
      if (templateType === 'INITIAL_INQUIRY') {
        sub = `Inquiry regarding freight shipping - ${lead?.company_name || lead?.contact_name || ''}`;
        body = `Hi ${lead?.contact_name || 'Customer'},\n\nThank you for connecting with us regarding your freight forwarding requirements.\n\nCould you please share the origin, destination, cargo type, and approximate weight for your upcoming shipment so we can prepare a competitive rate quote?\n\nBest regards,\nSales Team`;
      } else if (templateType === 'REQUEST_DETAILS') {
        sub = `Shipment details needed for quote - ${lead?.company_name || lead?.contact_name || ''}`;
        body = `Hi ${lead?.contact_name || 'Customer'},\n\nWe are following up regarding your freight inquiry. To calculate the most accurate rates and transit options, we need the following details:\n• Origin Port/City\n• Destination Port/City\n• Cargo Weight & Volume (CBM/KG)\n• Expected Ready Date\n\nPlease let us know at your convenience.\n\nBest regards,\nSales Team`;
      } else {
        body = `Hi ${lead?.contact_name || 'Customer'},\n\n`;
      }
    }

    setReplyForm({
      from: defaultFrom,
      to: lead?.email || '',
      cc: '',
      subject: sub,
      body: body
    });

    setTimeout(() => {
      isLoadingDraftRef.current = false;
    }, 100);
  };

  const handleOpenComposer = async (interaction, prefillBody = '') => {
    setReplyingToId(interaction.id);
    setSendError(null);
    setDraftStatus('');
    isLoadingDraftRef.current = true;

    const defaultFrom = availableMailboxes.find(mb => mb.is_primary)?.email || availableMailboxes[0]?.email || '';

    try {
      // 1. Check for saved draft on backend
      const draftRes = await getEmailDraft(lead.id, interaction.id);
      const draft = draftRes?.data || draftRes;
      if (draft && draft.body) {
        setReplyForm({
          from: draft.from || defaultFrom,
          to: draft.to || '',
          cc: draft.cc || '',
          subject: draft.subject || '',
          body: draft.body || ''
        });
        isLoadingDraftRef.current = false;
        return;
      }
    } catch (err) {
      console.error('Failed to check email draft:', err);
    }

    // 2. Default prefill if no draft was found
    let sub = interaction.subject || '';
    if (sub && !sub.toLowerCase().startsWith('re:')) {
      sub = `Re: ${sub}`;
    }
    setReplyForm({
      from: defaultFrom,
      to: interaction.sender || lead.email || '',
      cc: '',
      subject: sub,
      body: prefillBody
    });
    
    // Allow auto-save to run on subsequent changes
    setTimeout(() => {
      isLoadingDraftRef.current = false;
    }, 100);
  };

  const handleSendReply = async () => {
    if (!replyForm.to.trim() || !replyForm.subject.trim() || !replyForm.body.trim()) {
      setSendError('Recipient, Subject, and Body are required.');
      return;
    }
    setSendingReply(true);
    setSendError(null);
    try {
      const targetInterId = replyingToId !== null && replyingToId !== undefined ? replyingToId : 0;
      await replyToInteraction(lead.id, targetInterId, replyForm);
      
      // Clean up the draft since we successfully sent
      try {
        await deleteEmailDraft(lead.id, targetInterId);
      } catch (e) {
        console.error('Failed to clear draft after send:', e);
      }

      const data = await getLeadInteractions(lead.id, 'asc');
      setInteractions(data?.data || data || []);
      const freshTimeline = await getLeadTimeline(lead.id);
      setTimeline(freshTimeline?.data || freshTimeline || []);
      
      setReplyingToId(null);
      setReplyForm({ to: '', cc: '', subject: '', body: '' });
      setDraftStatus('');
    } catch (err) {
      console.error('Failed to send email:', err);
      setSendError(err.message || 'Unable to send email. Please try again.');
    } finally {
      setSendingReply(false);
    }
  };

  const handleSendSuggestedReplyDirect = async (interaction, body) => {
    setSendingReply(true);
    setSendError(null);
    
    let sub = interaction.subject || '';
    if (sub && !sub.toLowerCase().startsWith('re:')) {
      sub = `Re: ${sub}`;
    }

    const payload = {
      to: interaction.sender || lead.email || '',
      cc: '',
      subject: sub,
      body: body
    };

    try {
      await replyToInteraction(lead.id, interaction.id, payload);
      
      // Clean up any draft for this parent interaction
      try {
        await deleteEmailDraft(lead.id, interaction.id);
      } catch (e) {
        // ignore
      }

      const data = await getLeadInteractions(lead.id, 'asc');
      setInteractions(data?.data || data || []);
      const freshTimeline = await getLeadTimeline(lead.id);
      setTimeline(freshTimeline?.data || freshTimeline || []);
    } catch (err) {
      console.error('Failed to send suggested reply:', err);
      alert(err.message || 'Unable to send suggested reply. Please edit and try again.');
    } finally {
      setSendingReply(false);
    }
  };

  const handleDiscardDraft = async () => {
    if (!replyingToId) return;

    setDraftStatus('');
    try {
      await deleteEmailDraft(lead.id, replyingToId);
    } catch (err) {
      console.error('Failed to delete draft:', err);
    }

    setReplyingToId(null);
    setSendError(null);
    setReplyForm({ to: '', cc: '', subject: '', body: '' });
  };

  const [retryingInteractions, setRetryingInteractions] = useState({});

  const handleRetrySend = async (interaction) => {
    if (retryingInteractions[interaction.id]) return;
    
    setRetryingInteractions(prev => ({ ...prev, [interaction.id]: true }));
    setSendError(null);

    try {
      const res = await retryEmailInteraction(lead.id, interaction.id);
      const updated = res?.data || res;
      
      // Update interactions list and timeline
      const freshData = await getLeadInteractions(lead.id, 'asc');
      setInteractions(freshData?.data || freshData || []);
      const freshTimeline = await getLeadTimeline(lead.id);
      setTimeline(freshTimeline?.data || freshTimeline || []);
      
      if (updated.status === 'FAILED') {
        alert(updated.last_error || 'Retry failed to send.');
      }
    } catch (err) {
      console.error('Failed to retry email:', err);
      alert(err.message || 'Unable to retry sending. Please try again.');
    } finally {
      setRetryingInteractions(prev => ({ ...prev, [interaction.id]: false }));
    }
  };

  const handleEditAndRetry = (interaction) => {
    // Open the composer prefilled with the failed interaction details.
    setReplyingToId(interaction.id);
    setReplyForm({
      to: interaction.recipients || '',
      cc: interaction.cc_recipients || '',
      subject: interaction.subject || '',
      body: interaction.content || ''
    });
    setSendError(null);
    setDraftStatus('');
  };

  useEffect(() => {
    setFormState({
      company_name: lead.company_name || '',
      contact_name: lead.contact_name || '',
      email: lead.email || '',
      phone: lead.phone || '',
      source: lead.source || '',
      notes: lead.notes || '',
      location: lead.location || '',
      status: lead.status || 'NEW',
      tags: lead.tags || [],
    });
  }, [lead]);

  // Load activities timeline on mount / lead id change
  useEffect(() => {
    if (!lead?.id) return;
    let active = true;
    const fetchTimeline = async () => {
      setLoadingTimeline(true);
      try {
        const events = await getLeadTimeline(lead.id);
        if (active) {
          setTimeline(events?.data || events || []);
        }
      } catch (err) {
        console.error('Failed to load lead timeline:', err);
      } finally {
        if (active) setLoadingTimeline(false);
      }
    };
    fetchTimeline();
    return () => {
      active = false;
    };
  }, [lead.id]);

  const handleTimelineScroll = () => {
    const el = timelineScrollRef.current;
    if (!el) return;
    const isAtBottom = el.scrollHeight - el.scrollTop <= el.clientHeight + 15;
    setShowScrollHint(!isAtBottom);
  };

  const handleRetry = async (interactionId) => {
    if (retryingMap[interactionId]) return;
    setRetryingMap(prev => ({ ...prev, [interactionId]: true }));
    try {
      await retryClarificationEmail(lead.id, interactionId);
      // Reload timeline on success
      const freshTimeline = await getLeadTimeline(lead.id);
      setTimeline(freshTimeline?.data || freshTimeline || []);
      alert('Clarification email sent successfully!');
    } catch (err) {
      console.error('Retry failed:', err);
      alert('Failed to resend email: ' + (err.message || 'Unknown error'));
    } finally {
      setRetryingMap(prev => ({ ...prev, [interactionId]: false }));
    }
  };

  useEffect(() => {
    if (timeline.length > 3) {
      setShowScrollHint(true);
    } else {
      setShowScrollHint(false);
    }
  }, [timeline]);

  if (!lead) return null;

  // Derive a 2-letter avatar from company name
  const avatar = formState.company_name
    ? formState.company_name.slice(0, 2).toUpperCase()
    : '??';

  async function handleDelete() {
    if (!confirm(`Delete Lead? Are you sure you want to permanently delete ${formState.company_name || lead.company_name}?`)) return;
    setDeleting(true);
    try {
      await deleteLead(lead.id);
      onLeadUpdated(null); // Deselect after delete
    } catch (e) {
      alert(e.message || 'Failed to delete lead');
    } finally {
      setDeleting(false);
    }
  }

  function handleChange(e) {
    const { name, value } = e.target;
    setFormState(prev => ({ ...prev, [name]: value }));
  }

  function handleFieldChange(name, value) {
    setFormState(prev => ({ ...prev, [name]: value }));
  }

  function handleAddTag(newTag) {
    const trimmed = newTag.trim();
    if (!trimmed) return;
    const exists = formState.tags.some(t => t.toLowerCase() === trimmed.toLowerCase());
    if (exists) return;
    setFormState(prev => ({
      ...prev,
      tags: [...prev.tags, trimmed]
    }));
  }

  function handleRemoveTag(tagToRemove) {
    setFormState(prev => ({
      ...prev,
      tags: prev.tags.filter(t => t.toLowerCase() !== tagToRemove.toLowerCase())
    }));
  }

  const isDirty =
    formState.company_name !== (lead.company_name || '') ||
    formState.contact_name !== (lead.contact_name || '') ||
    formState.email !== (lead.email || '') ||
    formState.phone !== (lead.phone || '') ||
    formState.source !== (lead.source || '') ||
    formState.notes !== (lead.notes || '') ||
    formState.location !== (lead.location || '') ||
    formState.status !== (lead.status || 'NEW') ||
    formState.tags.length !== (lead.tags || []).length ||
    formState.tags.some((t, idx) => t !== (lead.tags || [])[idx]);

  useEffect(() => {
    if (onDirtyChange) {
      onDirtyChange(isDirty);
    }
  }, [isDirty, onDirtyChange]);

  async function handleSave() {
    if (!formState.company_name.trim()) {
      alert('Company name is required.');
      return;
    }
    setUpdating(true);
    try {
      const payload = {
        ...formState,
        tags: formState.tags,
      };
      const updated = await updateLead(lead.id, payload);
      onLeadUpdated(updated); // update parent state immediately

      // Refresh timeline after change
      const freshTimeline = await getLeadTimeline(lead.id);
      setTimeline(freshTimeline?.data || freshTimeline || []);
    } catch (e) {
      alert(e.message || 'Failed to save changes');
    } finally {
      setUpdating(false);
    }
  }

  const statusMap = {
    [LEAD_STATUS.NEW]:         { label: 'New',         type: 'info' },
    [LEAD_STATUS.QUALIFIED]:   { label: 'Qualified',   type: 'success' },
    [LEAD_STATUS.IN_PROGRESS]: { label: 'In Progress', type: 'warning' },
    [LEAD_STATUS.REJECTED]:    { label: 'Rejected',    type: 'danger' },
    [LEAD_STATUS.CONVERTED]:   { label: 'Converted',   type: 'converted' },
    'ACTIVE':                  { label: 'Active',      type: 'active' },
  };
  const statusConfig = statusMap[formState.status] || { label: formState.status, type: 'neutral' };

  // Determine if there is real AI data (non-zero score or existing report)
  const hasAIData = (lead.ai_score > 0) || (lead.ai_research_report && lead.ai_research_report !== "Failed to parse AI response.");

  return (
    <div className="leads-detail-panel">
      {/* Panel Sticky Header */}
      <div className="panel-header">
        <div className="panel-header-left">
          <div className="panel-avatar">{avatar}</div>
          <div className="panel-header-info">
            <h2 className="panel-title">{formState.company_name || lead.company_name}</h2>
            {formState.email && (
              <a href={`mailto:${formState.email}`} className="panel-header-email">
                {formState.email}
              </a>
            )}
            <div className="panel-header-badges">
              <StatusBadge status={formState.status || 'NEW'} customLabel={statusConfig.label} customType={statusConfig.type} />
              {isConverted && (
                <button
                  type="button"
                  onClick={() => {
                    if (createdRFQId) {
                      navigate(`/dashboard/rfqs/${createdRFQId}`);
                    } else {
                      navigate('/dashboard/rfqs');
                    }
                  }}
                  style={{
                    background: '#ECFDF5',
                    color: '#047857',
                    border: '1px solid #A7F3D0',
                    borderRadius: '6px',
                    padding: '2px 8px',
                    fontSize: '11px',
                    fontWeight: '700',
                    cursor: 'pointer',
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '4px',
                    transition: 'all 0.15s ease',
                  }}
                  title="Open RFQ Detail Workspace"
                >
                  <span>✓ {createdRFQNumber || (createdRFQId ? `RFQ #${createdRFQId}` : 'RFQ Converted')}</span>
                  <span>→</span>
                </button>
              )}
              {formState.tags && formState.tags.map((t, idx) => (
                <span key={idx} className="lead-tag-pill removeable">
                  {t}
                  <span className="tag-remove-x" onClick={() => handleRemoveTag(t)}>✕</span>
                </span>
              ))}
            </div>

          </div>
        </div>
        <button className="panel-close-btn" onClick={onClose} aria-label="Close panel">✕</button>
      </div>

      {/* Panel Body */}
      <div className="panel-body">
        <div className="conversation-tabs">
          <button 
            className={`conversation-tab-btn ${activeTab === 'overview' ? 'active' : ''}`}
            onClick={() => setActiveTab('overview')}
          >
            <FileText size={15} /> Overview
          </button>
          <button 
            className={`conversation-tab-btn ${activeTab === 'emails' ? 'active' : ''}`}
            onClick={() => setActiveTab('emails')}
          >
            <Mail size={15} /> Email Conversation ({interactions.length})
          </button>
          <button 
            className={`conversation-tab-btn ${activeTab === 'timeline' ? 'active' : ''}`}
            onClick={() => setActiveTab('timeline')}
          >
            <Clock size={15} /> Timeline & Activities ({timeline.length})
          </button>
          <button 
            className={`conversation-tab-btn ${activeTab === 'outreach' ? 'active' : ''}`}
            onClick={() => setActiveTab('outreach')}
          >
            <Sparkles size={15} /> Outreach History ({outreachActivities.length})
          </button>
        </div>

        {activeTab === 'overview' && (
          <div className="overview-tab-content">
            {/* 1. Basic Lead Details */}
            <div className="panel-card">
              <div className="card-header">
                <span className="card-header-icon">🏢</span>
                <h3 className="card-title">Company & Pipeline</h3>
              </div>
              <div className="panel-grid-2col">
                <div className="panel-grid-item span-2">
                  <span className="panel-grid-label">Company Name *</span>
                  <input
                    type="text"
                    name="company_name"
                    value={formState.company_name}
                    onChange={handleChange}
                    placeholder="e.g. Acme Corp"
                    className="styled-form-input"
                  />
                </div>

                <div className="panel-grid-item">
                  <span className="panel-grid-label">Pipeline Status</span>
                  <div className="panel-grid-value">
                    <CustomSelect
                      name="status"
                      value={formState.status}
                      onChange={(e) => handleFieldChange('status', e.target.value)}
                      options={[
                        { value: LEAD_STATUS.NEW, label: 'New' },
                        { value: LEAD_STATUS.IN_PROGRESS, label: 'In Progress' },
                        { value: LEAD_STATUS.QUALIFIED, label: 'Qualified' },
                        { value: LEAD_STATUS.CONVERTED, label: 'Converted' },
                        { value: LEAD_STATUS.REJECTED, label: 'Rejected' }
                      ]}
                      disabled={updating}
                    />
                  </div>
                </div>

                <div className="panel-grid-item">
                  <span className="panel-grid-label">Lead Source</span>
                  <div className="panel-grid-value">
                    <CustomSelect
                      name="source"
                      value={formState.source}
                      onChange={(e) => handleFieldChange('source', e.target.value)}
                      options={[
                        { value: '', label: 'Select source...' },
                        { value: 'Trade Show', label: 'Trade Show' },
                        { value: 'Referral', label: 'Referral' },
                        { value: 'LinkedIn', label: 'LinkedIn' },
                        { value: 'Cold Outreach', label: 'Cold Outreach' },
                        { value: 'Website', label: 'Website' },
                        { value: 'ImportYeti', label: 'ImportYeti' },
                        { value: 'Other', label: 'Other' }
                      ]}
                      disabled={updating}
                      placeholder="Select source..."
                    />
                  </div>
                </div>

                {lead.campaign_id && (
                  <div className="panel-grid-item span-2" style={{ marginTop: 8, padding: '10px 14px', background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: 8 }}>
                    <div style={{ fontSize: 11, fontWeight: 600, color: '#64748B', textTransform: 'uppercase' }}>Campaign Origin</div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 4 }}>
                      <span style={{ fontSize: 13.5, fontWeight: 700, color: '#1E293B' }}>📢 {lead.campaign_name || `Campaign #${lead.campaign_id}`}</span>
                      {lead.converted_from_outreach_at && (
                        <span style={{ fontSize: 11.5, color: '#64748B' }}>
                          (Attributed on {new Date(lead.converted_from_outreach_at).toLocaleDateString()})
                        </span>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* 2. Contact Details */}
            <div className="panel-card">
              <div className="card-header">
                <span className="card-header-icon">👤</span>
                <h3 className="card-title">Contact Information</h3>
              </div>
              <div className="panel-grid-2col">
                <div className="panel-grid-item">
                  <span className="panel-grid-label">Contact Name</span>
                  <input
                    type="text"
                    name="contact_name"
                    value={formState.contact_name}
                    onChange={handleChange}
                    placeholder="e.g. Jane Doe"
                    className="styled-form-input"
                  />
                </div>
                <div className="panel-grid-item">
                  <span className="panel-grid-label">Phone Number</span>
                  <input
                    type="text"
                    name="phone"
                    value={formState.phone}
                    onChange={handleChange}
                    placeholder="e.g. +1 (555) 0199"
                    className="styled-form-input"
                  />
                </div>
                <div className="panel-grid-item span-2">
                  <span className="panel-grid-label">Email Address</span>
                  <div className="email-input-with-action">
                    <input
                      type="email"
                      name="email"
                      value={formState.email}
                      onChange={handleChange}
                      placeholder="e.g. jane@company.com"
                      className="styled-form-input"
                    />
                    {formState.email && (
                      <a href={`mailto:${formState.email}`} className="quick-email-btn" title="Send Email">
                        ✉️ Email
                      </a>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* 3. Geographic Location & Google Map */}
            <div className="panel-card">
              <div className="card-header">
                <span className="card-header-icon">📍</span>
                <h3 className="card-title">Geographic Location</h3>
              </div>
              <LocationPickerMap
                value={formState.location}
                onChange={(val) => handleFieldChange('location', val)}
                disabled={updating}
              />
            </div>

            {/* 4. Tags & Notes */}
            <div className="panel-card">
              <div className="card-header">
                <span className="card-header-icon">🏷️</span>
                <h3 className="card-title">Tags & Notes</h3>
              </div>
              <div className="panel-grid-2col">
                <div className="panel-grid-item span-2">
                  <span className="panel-grid-label">Add Tag (Press Enter or click away)</span>
                  <div className="panel-tags-editor">
                    <input
                      type="text"
                      className="panel-add-tag-input"
                      placeholder="+ Add tag..."
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          const val = e.target.value.trim();
                          if (val) {
                            handleAddTag(val);
                            e.target.value = '';
                          }
                        }
                      }}
                      onBlur={(e) => {
                        const val = e.target.value.trim();
                        if (val) {
                          handleAddTag(val);
                          e.target.value = '';
                        }
                      }}
                    />
                  </div>
                </div>

                <div className="panel-grid-item span-2">
                  <span className="panel-grid-label">Notes & Inquiry Details</span>
                  <textarea
                    name="notes"
                    value={formState.notes}
                    onChange={handleChange}
                    placeholder="Any quick notes about this lead..."
                    rows={3}
                    className="styled-form-textarea"
                  />
                </div>
              </div>
            </div>



            {/* Section: RFQ Status & Shipment Information */}
            <div className="panel-section rfq-progress-section" id="rfq-progress-card">
              <h3 className="panel-section-title">RFQ Status</h3>
              {(() => {
                const headerData = (() => {
                  switch (workflowState.state) {
                    case 'NEW_INQUIRY':
                      return {
                        icon: '🔵',
                        label: 'New customer inquiry',
                        description: 'A customer has contacted you with a freight request.',
                        badgeColor: '#EFF6FF',
                        textColor: '#1E40AF',
                        borderColor: '#BFDBFE'
                      };
                    case 'INFORMATION_INCOMPLETE':
                      return {
                        icon: '🟡',
                        label: 'More information needed',
                        description: 'Some shipment details are still required before an RFQ can be created.',
                        badgeColor: '#FEF3C7',
                        textColor: '#92400E',
                        borderColor: '#FDE68A'
                      };
                    case 'WAITING_FOR_CUSTOMER':
                      const timeDiffStr = workflowState.lastRequestTime ? (() => {
                        const diffMs = Date.now() - new Date(workflowState.lastRequestTime).getTime();
                        const diffHrs = Math.floor(diffMs / (1000 * 60 * 60));
                        if (diffHrs === 0) {
                          const diffMins = Math.floor(diffMs / (1000 * 60));
                          return `Last request sent ${diffMins} min${diffMins !== 1 ? 's' : ''} ago`;
                        }
                        return `Last request sent ${diffHrs} hour${diffHrs !== 1 ? 's' : ''} ago`;
                      })() : 'Clarification request sent';
                      return {
                        icon: '🟠',
                        label: 'Waiting for customer',
                        description: 'A request for additional shipment details has been sent.',
                        secondaryText: timeDiffStr,
                        badgeColor: '#FFEDD5',
                        textColor: '#C2410C',
                        borderColor: '#FED7AA'
                      };
                    case 'CUSTOMER_RESPONDED':
                      return {
                        icon: '🔵',
                        label: 'Customer replied',
                        description: 'New shipment information was received and added to this request.',
                        badgeColor: '#EFF6FF',
                        textColor: '#1E40AF',
                        borderColor: '#BFDBFE'
                      };
                    case 'RFQ_READY':
                      return {
                        icon: '🔵',
                        label: 'Ready to create RFQ',
                        description: 'All required shipment information has been collected.',
                        badgeColor: '#DBEAFE',
                        textColor: '#1E40AF',
                        borderColor: '#BFDBFE'
                      };
                    case 'RFQ_CREATED':
                      return {
                        icon: '🟢',
                        label: 'RFQ created',
                        description: 'This customer inquiry has been converted into an RFQ.',
                        badgeColor: '#D1FAE5',
                        textColor: '#065F46',
                        borderColor: '#A7F3D0'
                      };
                    default:
                      return {
                        icon: '🔵',
                        label: 'Inbound Inquiry',
                        description: 'Customer conversation is active.',
                        badgeColor: '#F1F5F9',
                        textColor: '#475569',
                        borderColor: '#E2E8F0'
                      };
                  }
                })();

                return (
                  <div style={{
                    padding: '16px',
                    backgroundColor: '#F8FAFC',
                    border: '1px solid #E2E8F0',
                    borderRadius: '8px',
                    marginBottom: '20px'
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                      <span style={{ fontSize: '16px' }}>{headerData.icon}</span>
                      <span style={{
                        fontWeight: '600',
                        fontSize: '14px',
                        color: headerData.textColor,
                        backgroundColor: headerData.badgeColor,
                        padding: '3px 8px',
                        borderRadius: '4px',
                        border: `1px solid ${headerData.borderColor}`
                      }}>
                        {headerData.label}
                      </span>
                      {headerData.secondaryText && (
                        <span style={{ fontSize: '11px', color: '#64748B', marginLeft: 'auto' }}>
                          {headerData.secondaryText}
                        </span>
                      )}
                    </div>
                    <p style={{ margin: 0, fontSize: '13px', color: '#475569' }}>
                      {headerData.description}
                    </p>
                  </div>
                );
              })()}

              <h4 style={{ margin: '0 0 10px 0', fontSize: '13px', fontWeight: '600', color: '#475569' }}>Shipment Information</h4>
              <div style={{
                padding: '16px',
                backgroundColor: '#F8FAFC',
                border: '1px solid #E2E8F0',
                borderRadius: '8px'
              }}>
                <div style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr',
                  gap: '12px'
                }}>
                  {mandatoryFields.map(f => {
                    const val = rfqContext[f.key];
                    const isPresent = val !== undefined && val !== null && val !== '' && (typeof val !== 'number' || val > 0);
                    return (
                      <div key={f.key} style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        fontSize: '13px',
                        borderBottom: '1px dashed #E2E8F0',
                        paddingBottom: '8px'
                      }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                          <span style={{
                            color: isPresent ? '#10B981' : '#EF4444',
                            fontWeight: 'bold',
                            fontSize: '14px'
                          }}>
                            {isPresent ? '✓' : '✗'}
                          </span>
                          <span style={{ color: '#475569', fontWeight: '500' }}>{f.label}</span>
                        </div>
                        <span style={{
                          fontWeight: '600',
                          color: isPresent ? '#0F172A' : '#EF4444'
                        }}>
                          {isPresent ? String(val) : 'Required'}
                        </span>
                      </div>
                    );
                  })}
                </div>

                {hasCreatedRFQ && createdRFQId && (
                  <div style={{
                    marginTop: '16px',
                    paddingTop: '12px',
                    borderTop: '1px solid #E2E8F0',
                    textAlign: 'right'
                  }}>
                    <a
                      href={`/dashboard/rfqs/${createdRFQId}`}
                      className="view-rfq-link"
                      style={{
                        color: '#4F46E5',
                        fontWeight: '600',
                        fontSize: '13px',
                        textDecoration: 'none',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '4px'
                      }}
                    >
                      View RFQ →
                    </a>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'emails' && (
          <div className="panel-section" id="rfq-emails-tab">
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '14px', flexWrap: 'wrap', gap: '10px' }}>
              <h3 className="panel-section-title" style={{ margin: 0 }}>Email Conversations</h3>
              {lead?.email && replyingToId === null && (
                <button
                  className="btn btn-primary btn-sm"
                  onClick={() => handleStartNewEmail('CUSTOM')}
                  style={{
                    fontSize: '12.5px',
                    padding: '6px 14px',
                    borderRadius: '8px',
                    background: 'linear-gradient(135deg, #6366F1 0%, #4F46E5 100%)',
                    color: '#ffffff',
                    border: 'none',
                    cursor: 'pointer',
                    fontWeight: '600',
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '6px',
                    boxShadow: '0 2px 8px rgba(79, 70, 229, 0.25)'
                  }}
                >
                  <Send size={14} /> Compose Email to Lead
                </button>
              )}
            </div>
            
            {/* Part 2 — Lead Workflow Header */}
            {(() => {
              const headerData = (() => {
                switch (workflowState.state) {
                  case 'NEW_INQUIRY':
                    return {
                      icon: '🔵',
                      label: 'New customer inquiry',
                      description: 'A customer has contacted you with a freight request.',
                      badgeColor: '#EFF6FF',
                      textColor: '#1E40AF',
                      borderColor: '#BFDBFE'
                    };
                  case 'INFORMATION_INCOMPLETE':
                    return {
                      icon: '🟡',
                      label: 'More information needed',
                      description: 'Some shipment details are still required before an RFQ can be created.',
                      badgeColor: '#FEF3C7',
                      textColor: '#92400E',
                      borderColor: '#FDE68A'
                    };
                  case 'WAITING_FOR_CUSTOMER':
                    const timeDiffStr = workflowState.lastRequestTime ? (() => {
                      const diffMs = Date.now() - new Date(workflowState.lastRequestTime).getTime();
                      const diffHrs = Math.floor(diffMs / (1000 * 60 * 60));
                      if (diffHrs === 0) {
                        const diffMins = Math.floor(diffMs / (1000 * 60));
                        return `Last request sent ${diffMins} min${diffMins !== 1 ? 's' : ''} ago`;
                      }
                      return `Last request sent ${diffHrs} hour${diffHrs !== 1 ? 's' : ''} ago`;
                    })() : 'Clarification request sent';
                    return {
                      icon: '🟠',
                      label: 'Waiting for customer',
                      description: 'A request for additional shipment details has been sent.',
                      secondaryText: timeDiffStr,
                      badgeColor: '#FFEDD5',
                      textColor: '#C2410C',
                      borderColor: '#FED7AA'
                    };
                  case 'CUSTOMER_RESPONDED':
                    return {
                      icon: '🔵',
                      label: 'Customer replied',
                      description: 'New shipment information was received and added to this request.',
                      badgeColor: '#EFF6FF',
                      textColor: '#1E40AF',
                      borderColor: '#BFDBFE'
                    };
                  case 'RFQ_READY':
                    return {
                      icon: '🔵',
                      label: 'Ready to create RFQ',
                      description: 'All required shipment information has been collected.',
                      badgeColor: '#DBEAFE',
                      textColor: '#1E40AF',
                      borderColor: '#BFDBFE'
                    };
                  case 'RFQ_CREATED':
                    return {
                      icon: '🟢',
                      label: 'RFQ created',
                      description: 'This customer inquiry has been converted into an RFQ.',
                      badgeColor: '#D1FAE5',
                      textColor: '#065F46',
                      borderColor: '#A7F3D0'
                    };
                  default:
                    return {
                      icon: '🔵',
                      label: 'Inbound Inquiry',
                      description: 'Customer conversation is active.',
                      badgeColor: '#F1F5F9',
                      textColor: '#475569',
                      borderColor: '#E2E8F0'
                    };
                }
              })();

              return (
                <div style={{
                  padding: '16px',
                  backgroundColor: '#F8FAFC',
                  border: '1px solid #E2E8F0',
                  borderRadius: '8px',
                  marginBottom: '16px'
                }} className="rfq-conversation-header">
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                    <span style={{ fontSize: '16px' }}>{headerData.icon}</span>
                    <span style={{
                      fontWeight: '600',
                      fontSize: '14px',
                      color: headerData.textColor,
                      backgroundColor: headerData.badgeColor,
                      padding: '3px 8px',
                      borderRadius: '4px',
                      border: `1px solid ${headerData.borderColor}`
                    }}>
                      {headerData.label}
                    </span>
                    {headerData.secondaryText && (
                      <span style={{ fontSize: '11px', color: '#64748B', marginLeft: 'auto' }}>
                        {headerData.secondaryText}
                      </span>
                    )}
                  </div>
                  <p style={{ margin: 0, fontSize: '13px', color: '#475569', marginBottom: workflowState.missingFields?.length > 0 ? '10px' : '0' }}>
                    {headerData.description}
                  </p>
                  {workflowState.missingFields?.length > 0 && (
                    <div style={{ fontSize: '12px', color: '#D97706' }}>
                      <strong>Still needed:</strong>
                      <ul style={{ margin: '4px 0 0 16px', padding: 0, listStyleType: 'disc' }}>
                        {workflowState.missingFields.map(field => (
                          <li key={field} style={{ margin: '2px 0' }}>{field}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              );
            })()}

            {/* Part 8 — Custom Thread Grouping Selection card */}
            {threadGroups.length > 1 && (
              <CustomThreadSelector
                threadGroups={threadGroups}
                selectedThreadId={selectedThreadId}
                onSelectThread={(id) => setSelectedThreadId(id)}
              />
            )}

            <div className="email-conversation-container">
              {/* Part 5 — Conversation Timeline Lifecycle Marker for Converted Lead */}
              {(lead.status === 'CONVERTED' || lead.linked_rfq_id) && (
                <div className="rfq-lifecycle-banner" style={{
                  padding: '12px 16px',
                  backgroundColor: '#f0fdf4',
                  border: '1px solid #bbf7d0',
                  borderRadius: '10px',
                  marginBottom: '16px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between'
                }}>
                  <div>
                    <div style={{ fontSize: '13px', fontWeight: '700', color: '#166534', display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <span>✓ RFQ Created</span>
                    </div>
                    <div style={{ fontSize: '12px', color: '#15803d', marginTop: '2px' }}>
                      {lead.linked_rfq_id ? `RFQ-${lead.linked_rfq_id}` : 'RFQ record'} was created from this conversation.
                    </div>
                  </div>
                  {lead.linked_rfq_id && (
                    <button
                      type="button"
                      onClick={() => navigate(`/dashboard/rfqs/${lead.linked_rfq_id}`)}
                      style={{
                        fontSize: '12px',
                        fontWeight: '600',
                        padding: '6px 12px',
                        borderRadius: '6px',
                        backgroundColor: '#166534',
                        color: '#ffffff',
                        border: 'none',
                        cursor: 'pointer'
                      }}
                    >
                      View RFQ →
                    </button>
                  )}
                </div>
              )}

              {loadingInteractions ? (
                <div className="timeline-loading">Loading emails...</div>
              ) : interactions.length === 0 ? (
                <div>
                  {!lead?.email ? (
                    <div style={{
                      padding: '24px',
                      backgroundColor: '#FFFBEB',
                      border: '1.5px dashed #F59E0B',
                      borderRadius: '12px',
                      textAlign: 'center',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      gap: '12px'
                    }}>
                      <AlertTriangle size={32} color="#D97706" />
                      <div>
                        <h4 style={{ margin: '0 0 4px 0', fontSize: '15px', color: '#92400E', fontWeight: '700' }}>No Email Address Found</h4>
                        <p style={{ margin: 0, fontSize: '13px', color: '#B45309' }}>
                          This lead does not have an email address saved. Add an email address to start sending freight inquiries or rate quotes.
                        </p>
                      </div>
                      <button
                        className="btn btn-secondary btn-sm"
                        onClick={() => setActiveTab('overview')}
                        style={{
                          fontSize: '12.5px',
                          padding: '6px 16px',
                          borderRadius: '8px',
                          backgroundColor: '#FFFFFF',
                          color: '#D97706',
                          border: '1px solid #FCD34D',
                          fontWeight: '600',
                          cursor: 'pointer',
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: '6px'
                        }}
                      >
                        <Edit3 size={14} /> Edit Lead & Add Email
                      </button>
                    </div>
                  ) : (
                    <div style={{
                      padding: '28px 20px',
                      backgroundColor: '#FFFFFF',
                      border: '1px solid #E2E8F0',
                      borderRadius: '14px',
                      textAlign: 'center',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      gap: '14px',
                      boxShadow: '0 4px 14px rgba(15,23,42,0.03)'
                    }}>
                      <div style={{
                        width: '52px',
                        height: '52px',
                        borderRadius: '50%',
                        backgroundColor: '#EEF2FF',
                        color: '#4F46E5',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center'
                      }}>
                        <Mail size={24} />
                      </div>
                      <div>
                        <h4 style={{ margin: '0 0 6px 0', fontSize: '15.5px', color: '#0F172A', fontWeight: '700' }}>
                          Start Conversation with {lead.company_name || lead.contact_name || 'Lead'}
                        </h4>
                        <p style={{ margin: 0, fontSize: '13px', color: '#64748B', maxWidth: '440px', lineHeight: '1.5' }}>
                          This lead was added manually. You can initiate the conversation by choosing an email template below or composing a custom message.
                        </p>
                      </div>

                      {/* Quick Email Templates Grid */}
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '10px', justifyContent: 'center', marginTop: '6px' }}>
                        <button
                          type="button"
                          onClick={() => handleStartNewEmail('INITIAL_INQUIRY')}
                          style={{
                            fontSize: '12px',
                            fontWeight: '600',
                            padding: '8px 14px',
                            borderRadius: '8px',
                            backgroundColor: '#EEF2FF',
                            color: '#4338CA',
                            border: '1px solid #C7D2FE',
                            cursor: 'pointer',
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '6px'
                          }}
                        >
                          📩 Initial Freight Inquiry
                        </button>
                        <button
                          type="button"
                          onClick={() => handleStartNewEmail('REQUEST_DETAILS')}
                          style={{
                            fontSize: '12px',
                            fontWeight: '600',
                            padding: '8px 14px',
                            borderRadius: '8px',
                            backgroundColor: '#ECFDF5',
                            color: '#047857',
                            border: '1px solid #A7F3D0',
                            cursor: 'pointer',
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '6px'
                          }}
                        >
                          📦 Request Shipment Details
                        </button>
                        <button
                          type="button"
                          onClick={() => handleStartNewEmail('CUSTOM')}
                          style={{
                            fontSize: '12px',
                            fontWeight: '600',
                            padding: '8px 14px',
                            borderRadius: '8px',
                            backgroundColor: '#F8FAFC',
                            color: '#334155',
                            border: '1px solid #CBD5E1',
                            cursor: 'pointer',
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '6px'
                          }}
                        >
                          ✍️ Custom Email
                        </button>
                      </div>
                    </div>
                  )}

                  {/* Inline Composer for New Email when replyingToId === 0 */}
                  {replyingToId === 0 && (
                    <div className="reply-composer" style={{ marginTop: '20px' }}>
                      <div className="composer-header">
                        <div className="composer-header-title">
                          <Send size={16} /> New Email to {lead.contact_name || lead.email}
                        </div>
                        {draftStatus && (
                          <span className={`draft-status-label ${draftStatus}`}>
                            {draftStatus === 'saving' && '✍️ Saving draft...'}
                            {draftStatus === 'saved' && '💾 Draft saved'}
                            {draftStatus === 'error' && '❌ Failed to save draft'}
                          </span>
                        )}
                      </div>
                      <div className="composer-body">
                        <div className="composer-field">
                          <label><User size={13} /> From (Sender Account)</label>
                          {availableMailboxes.length > 0 ? (
                            <select 
                              value={replyForm.from} 
                              disabled={sendingReply}
                              onChange={(e) => setReplyForm(prev => ({ ...prev, from: e.target.value }))}
                              style={{
                                width: '100%',
                                padding: '8px 12px',
                                borderRadius: '6px',
                                border: '1px solid #CBD5E1',
                                fontSize: '13px',
                                backgroundColor: '#FFFFFF',
                                color: '#0F172A',
                                fontWeight: '500'
                              }}
                            >
                              {availableMailboxes.map(mb => (
                                <option key={mb.id || mb.email} value={mb.email}>
                                  {mb.email} {mb.is_primary ? '(Primary)' : ''}
                                </option>
                              ))}
                            </select>
                          ) : (
                            <input 
                              type="text" 
                              value={replyForm.from} 
                              disabled={sendingReply}
                              onChange={(e) => setReplyForm(prev => ({ ...prev, from: e.target.value }))}
                              placeholder="sender@domain.com"
                            />
                          )}
                        </div>
                        <div className="composer-field">
                          <label><User size={13} /> To</label>
                          <input 
                            type="text" 
                            value={replyForm.to} 
                            disabled={sendingReply}
                            onChange={(e) => setReplyForm(prev => ({ ...prev, to: e.target.value }))}
                            placeholder="recipient@example.com"
                          />
                        </div>
                        <div className="composer-field">
                          <label><Mail size={13} /> CC (Optional)</label>
                          <input 
                            type="text" 
                            value={replyForm.cc} 
                            disabled={sendingReply}
                            onChange={(e) => setReplyForm(prev => ({ ...prev, cc: e.target.value }))}
                            placeholder="optional.cc@example.com"
                          />
                        </div>
                        <div className="composer-field">
                          <label><FileText size={13} /> Subject Line</label>
                          <input 
                            type="text" 
                            value={replyForm.subject} 
                            disabled={sendingReply}
                            onChange={(e) => setReplyForm(prev => ({ ...prev, subject: e.target.value }))}
                            placeholder="Inquiry regarding freight shipment..."
                          />
                        </div>
                        <div className="composer-field">
                          <label><Edit3 size={13} /> Message Body</label>
                          <textarea 
                            rows={7}
                            value={replyForm.body} 
                            disabled={sendingReply}
                            onChange={(e) => setReplyForm(prev => ({ ...prev, body: e.target.value }))}
                            placeholder="Type your email message here..."
                          />
                        </div>
                      </div>
                      {sendError && (
                        <div className="composer-error">
                          <AlertTriangle size={15} /> {sendError}
                        </div>
                      )}
                      <div className="composer-actions">
                        <button
                          className="btn btn-secondary btn-sm"
                          onClick={handleDiscardDraft}
                          disabled={sendingReply}
                          style={{
                            fontSize: '12px',
                            padding: '6px 14px',
                            borderRadius: '6px',
                            backgroundColor: '#FEF2F2',
                            color: '#DC2626',
                            border: '1px solid #FCA5A5',
                            cursor: 'pointer',
                            fontWeight: '600',
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '5px'
                          }}
                        >
                          <Trash2 size={13} /> Discard Draft
                        </button>
                        
                        <div style={{ display: 'flex', gap: '8px' }}>
                          <button
                            className="btn btn-secondary btn-sm"
                            onClick={() => setReplyingToId(null)}
                            disabled={sendingReply}
                            style={{
                              fontSize: '12px',
                              padding: '6px 14px',
                              borderRadius: '6px',
                              backgroundColor: '#E2E8F0',
                              color: '#475569',
                              border: 'none',
                              cursor: 'pointer',
                              fontWeight: '600'
                            }}
                          >
                            Cancel
                          </button>
                          <button
                            className="btn btn-primary btn-sm"
                            onClick={handleSendReply}
                            disabled={sendingReply}
                            style={{
                              fontSize: '12.5px',
                              padding: '6px 18px',
                              borderRadius: '6px',
                              background: 'linear-gradient(135deg, #6366F1 0%, #4F46E5 100%)',
                              color: '#fff',
                              border: 'none',
                              cursor: 'pointer',
                              fontWeight: '600',
                              display: 'inline-flex',
                              alignItems: 'center',
                              gap: '6px',
                              boxShadow: '0 2px 8px rgba(79, 70, 229, 0.3)'
                            }}
                          >
                            <Send size={14} />
                            {sendingReply ? 'Sending Email...' : 'Send Email'}
                          </button>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <>
                  {(() => {
                    const activeGroup = threadGroups.find(g => g.thread_id === selectedThreadId) || threadGroups[0];
                    if (!activeGroup) return null;
                    const otherGroups = threadGroups.filter(g => g.thread_id !== activeGroup.thread_id);

                    const sortedActiveList = [...activeGroup.interactions].sort(
                      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
                    );

                    return (
                      <div>
                        {/* Inline Composer for Fresh Email when replyingToId === 0 */}
                        {replyingToId === 0 && (
                          <div className="reply-composer" style={{ marginBottom: '20px' }}>
                            <div className="composer-header">
                              <div className="composer-header-title">
                                <Send size={16} /> New Email to {lead.contact_name || lead.email}
                              </div>
                              {draftStatus && (
                                <span className={`draft-status-label ${draftStatus}`}>
                                  {draftStatus === 'saving' && '✍️ Saving draft...'}
                                  {draftStatus === 'saved' && '💾 Draft saved'}
                                  {draftStatus === 'error' && '❌ Failed to save draft'}
                                </span>
                              )}
                            </div>
                            <div className="composer-body">
                              <div className="composer-field">
                                <label><User size={13} /> From (Sender Account)</label>
                                {availableMailboxes.length > 0 ? (
                                  <select 
                                    value={replyForm.from} 
                                    disabled={sendingReply}
                                    onChange={(e) => setReplyForm(prev => ({ ...prev, from: e.target.value }))}
                                    style={{
                                      width: '100%',
                                      padding: '8px 12px',
                                      borderRadius: '6px',
                                      border: '1px solid #CBD5E1',
                                      fontSize: '13px',
                                      backgroundColor: '#FFFFFF',
                                      color: '#0F172A',
                                      fontWeight: '500'
                                    }}
                                  >
                                    {availableMailboxes.map(mb => (
                                      <option key={mb.id || mb.email} value={mb.email}>
                                        {mb.email} {mb.is_primary ? '(Primary)' : ''}
                                      </option>
                                    ))}
                                  </select>
                                ) : (
                                  <input 
                                    type="text" 
                                    value={replyForm.from} 
                                    disabled={sendingReply}
                                    onChange={(e) => setReplyForm(prev => ({ ...prev, from: e.target.value }))}
                                    placeholder="sender@domain.com"
                                  />
                                )}
                              </div>
                              <div className="composer-field">
                                <label><User size={13} /> To</label>
                                <input 
                                  type="text" 
                                  value={replyForm.to} 
                                  disabled={sendingReply}
                                  onChange={(e) => setReplyForm(prev => ({ ...prev, to: e.target.value }))}
                                  placeholder="recipient@example.com"
                                />
                              </div>
                              <div className="composer-field">
                                <label><Mail size={13} /> CC (Optional)</label>
                                <input 
                                  type="text" 
                                  value={replyForm.cc} 
                                  disabled={sendingReply}
                                  onChange={(e) => setReplyForm(prev => ({ ...prev, cc: e.target.value }))}
                                  placeholder="optional.cc@example.com"
                                />
                              </div>
                              <div className="composer-field">
                                <label><FileText size={13} /> Subject Line</label>
                                <input 
                                  type="text" 
                                  value={replyForm.subject} 
                                  disabled={sendingReply}
                                  onChange={(e) => setReplyForm(prev => ({ ...prev, subject: e.target.value }))}
                                  placeholder="Inquiry regarding freight shipment..."
                                />
                              </div>
                              <div className="composer-field">
                                <label><Edit3 size={13} /> Message Body</label>
                                <textarea 
                                  rows={7}
                                  value={replyForm.body} 
                                  disabled={sendingReply}
                                  onChange={(e) => setReplyForm(prev => ({ ...prev, body: e.target.value }))}
                                  placeholder="Type your email message here..."
                                />
                              </div>
                            </div>
                            {sendError && (
                              <div className="composer-error">
                                <AlertTriangle size={15} /> {sendError}
                              </div>
                            )}
                            <div className="composer-actions">
                              <button
                                className="btn btn-secondary btn-sm"
                                onClick={handleDiscardDraft}
                                disabled={sendingReply}
                                style={{
                                  fontSize: '12px',
                                  padding: '6px 14px',
                                  borderRadius: '6px',
                                  backgroundColor: '#FEF2F2',
                                  color: '#DC2626',
                                  border: '1px solid #FCA5A5',
                                  cursor: 'pointer',
                                  fontWeight: '600',
                                  display: 'inline-flex',
                                  alignItems: 'center',
                                  gap: '5px'
                                }}
                              >
                                <Trash2 size={13} /> Discard Draft
                              </button>
                              
                              <div style={{ display: 'flex', gap: '8px' }}>
                                <button
                                  className="btn btn-secondary btn-sm"
                                  onClick={() => setReplyingToId(null)}
                                  disabled={sendingReply}
                                  style={{
                                    fontSize: '12px',
                                    padding: '6px 14px',
                                    borderRadius: '6px',
                                    backgroundColor: '#E2E8F0',
                                    color: '#475569',
                                    border: 'none',
                                    cursor: 'pointer',
                                    fontWeight: '600'
                                  }}
                                >
                                  Cancel
                                </button>
                                <button
                                  className="btn btn-primary btn-sm"
                                  onClick={handleSendReply}
                                  disabled={sendingReply}
                                  style={{
                                    fontSize: '12.5px',
                                    padding: '6px 18px',
                                    borderRadius: '6px',
                                    background: 'linear-gradient(135deg, #6366F1 0%, #4F46E5 100%)',
                                    color: '#fff',
                                    border: 'none',
                                    cursor: 'pointer',
                                    fontWeight: '600',
                                    display: 'inline-flex',
                                    alignItems: 'center',
                                    gap: '6px',
                                    boxShadow: '0 2px 8px rgba(79, 70, 229, 0.3)'
                                  }}
                                >
                                  <Send size={14} />
                                  {sendingReply ? 'Sending Email...' : 'Send Email'}
                                </button>
                              </div>
                            </div>
                          </div>
                        )}

                        {/* Section Header: Current Conversation */}
                        <div style={{
                          fontSize: '13px',
                          fontWeight: '600',
                          color: '#475569',
                          marginBottom: '12px',
                          borderBottom: '1px solid #E2E8F0',
                          paddingBottom: '6px'
                        }}>
                          Current RFQ Conversation ({sortedActiveList.length} messages)
                        </div>

                        {sortedActiveList.map((inter, idx) => {
                          const isDiscarded = discardedSuggestionIds.has(inter.id);
                          const formattedDate = new Date(inter.created_at).toLocaleString(undefined, {
                            month: 'short',
                            day: 'numeric',
                            year: 'numeric',
                            hour: 'numeric',
                            minute: '2-digit',
                            hour12: true
                          });

                          const isRetrying = retryingInteractions[inter.id];
                          const isExpanded = idx === 0 || expandedInterIds.has(inter.id);

                          // Helper function for dynamic activity markers
                          const renderActivityMarker = () => {
                            if (idx === 0) return null;
                            const prevInter = sortedActiveList[idx - 1];

                            const parseCtx = (iObj) => {
                              if (!iObj || !iObj.partial_rfq_context) return {};
                              if (typeof iObj.partial_rfq_context === 'string') {
                                try {
                                  return JSON.parse(iObj.partial_rfq_context);
                                } catch (e) {
                                  return {};
                                }
                              }
                              return iObj.partial_rfq_context;
                            };

                            const currCtx = parseCtx(inter);
                            const prevCtx = parseCtx(prevInter);

                            // Case 1: Outbound message sent => Waiting for customer
                            if (inter.direction === 'OUTBOUND' && inter.status === 'SENT') {
                              return (
                                <div style={{
                                  margin: '16px 0',
                                  display: 'flex',
                                  flexDirection: 'column',
                                  alignItems: 'center',
                                  textAlign: 'center'
                                }} className="activity-marker">
                                  <div style={{ width: '100%', height: '1px', backgroundColor: '#E2E8F0', marginBottom: '8px' }} />
                                  <span style={{
                                    fontSize: '12px',
                                    fontWeight: '600',
                                    color: '#C2410C',
                                    backgroundColor: '#FFEDD5',
                                    padding: '2px 8px',
                                    borderRadius: '4px'
                                  }}>
                                    Waiting for customer
                                  </span>
                                  <span style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
                                    Additional shipment details were requested.
                                  </span>
                                  <div style={{ width: '100%', height: '1px', backgroundColor: '#E2E8F0', marginTop: '8px' }} />
                                </div>
                              );
                            }

                            // Case 2: Inbound message => check if fields added
                            if (inter.direction === 'INBOUND') {
                              let prevInbound = null;
                              for (let i = idx - 1; i >= 0; i--) {
                                if (sortedActiveList[i].direction === 'INBOUND') {
                                  prevInbound = sortedActiveList[i];
                                  break;
                                }
                              }

                              const prevInboundCtx = parseCtx(prevInbound);
                              const addedFields = [];
                              mandatoryFields.forEach(f => {
                                const currVal = currCtx[f.key];
                                const prevVal = prevInboundCtx[f.key];
                                const isCurrPresent = currVal !== undefined && currVal !== null && currVal !== '' && (typeof currVal !== 'number' || currVal > 0);
                                const isPrevPresent = prevVal !== undefined && prevVal !== null && prevVal !== '' && (typeof prevVal !== 'number' || prevVal > 0);
                                if (isCurrPresent && !isPrevPresent) {
                                  addedFields.push(f.label);
                                }
                              });

                              const isCurrReady = mandatoryFields.every(f => {
                                const val = currCtx[f.key];
                                return val !== undefined && val !== null && val !== '' && (typeof val !== 'number' || val > 0);
                              });
                              const isPrevReady = prevInbound ? mandatoryFields.every(f => {
                                const val = prevInboundCtx[f.key];
                                return val !== undefined && val !== null && val !== '' && (typeof val !== 'number' || val > 0);
                              }) : false;

                              const rfqJustBecameReady = isCurrReady && !isPrevReady;

                              if (rfqJustBecameReady) {
                                return (
                                  <div style={{
                                    margin: '16px 0',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    alignItems: 'center',
                                    textAlign: 'center'
                                  }} className="activity-marker">
                                    <div style={{ width: '100%', height: '1px', backgroundColor: '#E2E8F0', marginBottom: '8px' }} />
                                    <span style={{
                                      fontSize: '12px',
                                      fontWeight: '600',
                                      color: '#1E40AF',
                                      backgroundColor: '#DBEAFE',
                                      padding: '2px 8px',
                                      borderRadius: '4px'
                                    }}>
                                      RFQ ready
                                    </span>
                                    <span style={{ fontSize: '11px', color: '#64748B', marginTop: '2px' }}>
                                      All required shipment details have been collected.
                                    </span>
                                    <div style={{ width: '100%', height: '1px', backgroundColor: '#E2E8F0', marginTop: '8px' }} />
                                  </div>
                                );
                              }

                              if (addedFields.length > 0) {
                                return (
                                  <div style={{
                                    margin: '16px 0',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    alignItems: 'center',
                                    textAlign: 'center'
                                  }} className="activity-marker">
                                    <div style={{ width: '100%', height: '1px', backgroundColor: '#E2E8F0', marginBottom: '8px' }} />
                                    <span style={{
                                      fontSize: '12px',
                                      fontWeight: '600',
                                      color: '#1E40AF',
                                      backgroundColor: '#EFF6FF',
                                      padding: '2px 8px',
                                      borderRadius: '4px'
                                    }}>
                                      Information received
                                    </span>
                                    <div style={{ display: 'flex', gap: '8px', marginTop: '2px', fontSize: '11px', color: '#64748B' }}>
                                      {addedFields.map(field => (
                                        <span key={field}>{field} added</span>
                                      ))}
                                    </div>
                                    <div style={{ width: '100%', height: '1px', backgroundColor: '#E2E8F0', marginTop: '8px' }} />
                                  </div>
                                );
                              }
                            }

                            return null;
                          };

                          if (!isExpanded) {
                            return (
                              <div key={inter.id || idx}>
                                {renderActivityMarker()}
                                <div
                                  className="collapsed-email-card"
                                  onClick={() => toggleExpandInter(inter.id)}
                                  style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'space-between',
                                    padding: '12px 16px',
                                    backgroundColor: inter.direction === 'INBOUND' ? '#ffffff' : '#f8fafc',
                                    border: '1px solid #e2e8f0',
                                    borderLeft: inter.direction === 'INBOUND' ? '4px solid #10b981' : '4px solid #6366f1',
                                    borderRadius: '10px',
                                    marginBottom: '10px',
                                    cursor: 'pointer',
                                    boxShadow: '0 2px 4px rgba(15,23,42,0.03)',
                                    transition: 'all 0.15s ease'
                                  }}
                                >
                                  <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flex: 1, minWidth: 0 }}>
                                    <div style={{
                                      width: '32px',
                                      height: '32px',
                                      borderRadius: '50%',
                                      backgroundColor: inter.direction === 'INBOUND' ? '#ECFDF5' : '#EEF2FF',
                                      color: inter.direction === 'INBOUND' ? '#047857' : '#4936D0',
                                      display: 'flex',
                                      alignItems: 'center',
                                      justifyContent: 'center',
                                      flexShrink: 0
                                    }}>
                                      {inter.direction === 'INBOUND' ? <Inbox size={16} /> : <SendHorizontal size={16} />}
                                    </div>
                                    <div style={{ display: 'flex', flexDirection: 'column', gap: '3px', flex: 1, minWidth: 0 }}>
                                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                        <strong style={{ fontSize: '13.5px', color: '#0f172a' }}>
                                          {inter.direction === 'INBOUND' ? `Customer: ${inter.sender || lead.email}` : `Sent to ${inter.recipients || lead.email}`}
                                        </strong>
                                      </div>
                                      <span style={{ fontSize: '12.5px', color: '#64748b', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                        <strong style={{ color: '#334155' }}>{inter.subject || '(No Subject)'}</strong> — {inter.content ? inter.content.slice(0, 70) : ''}...
                                      </span>
                                    </div>
                                  </div>

                                  <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flexShrink: 0 }}>
                                    <span className={`email-direction-badge ${inter.direction === 'INBOUND' ? 'inbound' : inter.status === 'SENT' ? 'outbound' : 'failed'}`}>
                                      {inter.direction === 'INBOUND' ? 'Received' : inter.status === 'SENT' ? 'Sent' : 'Failed'}
                                    </span>
                                    <span style={{ fontSize: '11.5px', color: '#94a3b8', display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                                      <Clock size={12} /> {formattedDate}
                                    </span>
                                    <span style={{ fontSize: '11px', color: '#4F46E5', fontWeight: '600', display: 'inline-flex', alignItems: 'center', gap: '2px' }}>
                                      View <ChevronDown size={14} />
                                    </span>
                                  </div>
                                </div>
                              </div>
                            );
                          }

                          return (
                            <div key={inter.id || idx}>
                              {/* Render Activity Marker if calculated */}
                              {renderActivityMarker()}

                              <div className={`email-card ${inter.status === 'FAILED' ? 'card-failed' : ''}`} style={{
                                borderLeft: inter.direction === 'INBOUND' ? '4px solid #10b981' : '4px solid #6366f1',
                                marginBottom: '18px'
                              }}>
                                <div className="email-card-header" onClick={() => toggleExpandInter(inter.id)} style={{ cursor: 'pointer' }}>
                                  <div style={{
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    alignItems: 'center',
                                    fontSize: '13px',
                                    color: '#475569'
                                  }}>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
                                      <span className={`email-direction-badge ${inter.direction === 'INBOUND' ? 'inbound' : inter.status === 'SENT' ? 'outbound' : 'failed'}`}>
                                        {inter.direction === 'INBOUND' ? <Inbox size={12} /> : <SendHorizontal size={12} />}
                                        {inter.direction === 'INBOUND' ? 'Received from Customer' : 'Sent to Customer'}
                                      </span>
                                      <div className="email-meta-field">
                                        <User size={13} />
                                        <strong>{inter.direction === 'INBOUND' ? (inter.sender || lead.email) : (inter.recipients || 'Customer')}</strong>
                                      </div>
                                      <span style={{ color: '#CBD5E1' }}>·</span>
                                      <div className="email-meta-date">
                                        <Clock size={12} /> {formattedDate}
                                      </div>
                                    </div>
                                    <button type="button" style={{
                                      background: 'none',
                                      border: 'none',
                                      color: '#64748b',
                                      fontSize: '12px',
                                      fontWeight: '600',
                                      cursor: 'pointer',
                                      display: 'inline-flex',
                                      alignItems: 'center',
                                      gap: '4px'
                                    }} onClick={(e) => { e.stopPropagation(); toggleExpandInter(inter.id); }}>
                                      Collapse <ChevronUp size={14} />
                                    </button>
                                  </div>

                                  {/* Executive Email Subject Banner */}
                                  <div className="email-subject-banner">
                                    <div>
                                      <span className="email-subject-label">Subject Title</span>
                                      <h4 className="email-subject-text">{inter.subject || '(No Subject)'}</h4>
                                    </div>
                                  </div>
                                </div>
                                
                                <div className="email-card-body">
                                  <EmailContentFormatter content={inter.content} />
                                  
                                  {/* Compact status indicator for RFQ details extracted */}
                                  {(() => {
                                    let interContext = {};
                                    if (inter.partial_rfq_context) {
                                      if (typeof inter.partial_rfq_context === 'string') {
                                        try {
                                          interContext = JSON.parse(inter.partial_rfq_context);
                                        } catch (e) {
                                          interContext = {};
                                        }
                                      } else {
                                        interContext = inter.partial_rfq_context;
                                      }
                                    }
                                    if (Object.keys(interContext).length === 0) return null;

                                    return (
                                      <div style={{
                                        marginTop: '14px',
                                        padding: '10px 14px',
                                        backgroundColor: '#F8FAFC',
                                        border: '1px solid #E2E8F0',
                                        borderRadius: '8px',
                                        fontSize: '12px',
                                        color: '#475569'
                                      }} className="interaction-rfq-info">
                                        <strong style={{ display: 'block', marginBottom: '6px', color: '#0F172A', fontSize: '12.5px' }}>
                                          Shipment Parameters Extracted:
                                        </strong>
                                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px 14px' }}>
                                          {mandatoryFields.map(f => {
                                            const val = interContext[f.key];
                                            const isPresent = val !== undefined && val !== null && val !== '' && (typeof val !== 'number' || val > 0);
                                            return (
                                              <span key={f.key} style={{
                                                display: 'inline-flex',
                                                alignItems: 'center',
                                                gap: '5px',
                                                padding: '2px 8px',
                                                borderRadius: '4px',
                                                backgroundColor: isPresent ? '#ECFDF5' : '#FEF2F2',
                                                color: isPresent ? '#047857' : '#991B1B',
                                                fontSize: '11.5px',
                                                fontWeight: '600'
                                              }}>
                                                {isPresent ? <CheckCircle2 size={13} color="#10B981" /> : <XCircle size={13} color="#EF4444" />}
                                                {f.label}: {isPresent ? String(val) : 'Missing'}
                                              </span>
                                            );
                                          })}
                                        </div>
                                      </div>
                                    );
                                  })()}
                                </div>

                                {inter.status === 'FAILED' && (
                                  <div className="email-error-banner">
                                    <AlertTriangle size={18} className="error-icon" color="#DC2626" />
                                    <span className="error-text">Could not deliver this email to {inter.recipients || lead.email}.</span>
                                    <div className="error-actions">
                                      <button 
                                        className="btn btn-primary btn-xs"
                                        onClick={() => handleRetrySend(inter)}
                                        disabled={isRetrying}
                                        style={{
                                          fontSize: '11.5px',
                                          padding: '4px 10px',
                                          borderRadius: '6px',
                                          backgroundColor: '#DC2626',
                                          color: '#fff',
                                          border: 'none',
                                          cursor: 'pointer',
                                          fontWeight: '600',
                                          display: 'inline-flex',
                                          alignItems: 'center',
                                          gap: '4px'
                                        }}
                                      >
                                        <RefreshCw size={12} className={isRetrying ? 'spin' : ''} />
                                        {isRetrying ? 'Retrying...' : 'Retry Send'}
                                      </button>
                                      <button 
                                        className="btn btn-secondary btn-xs"
                                        onClick={() => handleEditAndRetry(inter)}
                                        disabled={isRetrying}
                                        style={{
                                          fontSize: '11.5px',
                                          padding: '4px 10px',
                                          borderRadius: '6px',
                                          backgroundColor: '#FFFFFF',
                                          color: '#374151',
                                          border: '1px solid #D1D5DB',
                                          cursor: 'pointer',
                                          fontWeight: '600',
                                          display: 'inline-flex',
                                          alignItems: 'center',
                                          gap: '4px'
                                        }}
                                      >
                                        <Edit3 size={12} /> Edit & Retry
                                      </button>
                                    </div>
                                  </div>
                                )}

                                <div className="email-card-actions" style={{ marginTop: '12px', display: 'flex', gap: '8px' }}>
                                  {replyingToId !== inter.id && (
                                    <button 
                                      className="btn btn-primary btn-sm"
                                      onClick={() => handleOpenComposer(inter)}
                                      style={{
                                        fontSize: '12.5px',
                                        padding: '7px 16px',
                                        borderRadius: '8px',
                                        background: 'linear-gradient(135deg, #6366F1 0%, #4F46E5 100%)',
                                        color: '#ffffff',
                                        border: 'none',
                                        cursor: 'pointer',
                                        fontWeight: '600',
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        gap: '6px',
                                        boxShadow: '0 2px 8px rgba(79, 70, 229, 0.25)'
                                      }}
                                    >
                                      <Reply size={15} /> Reply to this Message
                                    </button>
                                  )}
                                </div>

                                {/* AI Suggested Reply Box inside inbound email card */}
                                {inter.direction === 'INBOUND' && inter.drafted_reply && !isDiscarded && replyingToId !== inter.id && (
                                  <div className="suggested-reply-box">
                                    <div className="suggested-reply-header">
                                      <span className="suggested-reply-badge">
                                        <Sparkles size={14} /> AI Suggested Reply
                                      </span>
                                      <span style={{ fontSize: '11.5px', color: '#64748B', fontWeight: '500' }}>
                                        Auto-generated based on customer email
                                      </span>
                                    </div>

                                    {/* Incomplete RFQ warning banner */}
                                    {workflowState.missingFields?.length > 0 && (
                                      <div style={{
                                        margin: '0 0 12px 0',
                                        padding: '10px 14px',
                                        backgroundColor: '#FFFBEB',
                                        borderLeft: '4px solid #F59E0B',
                                        borderRadius: '6px',
                                        fontSize: '12px',
                                        color: '#D97706'
                                      }} className="rfq-still-need-banner">
                                        <strong style={{ display: 'flex', alignItems: 'center', gap: '4px', marginBottom: '4px' }}>
                                          <AlertCircle size={14} /> Shipment info still required:
                                        </strong>
                                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px 12px' }}>
                                          {workflowState.missingFields.map(field => (
                                            <span key={field} style={{
                                              backgroundColor: '#FEF3C7',
                                              padding: '2px 8px',
                                              borderRadius: '4px',
                                              fontSize: '11px',
                                              fontWeight: '600'
                                            }}>
                                              • {field}
                                            </span>
                                          ))}
                                        </div>
                                      </div>
                                    )}

                                    <div className="suggested-reply-content">
                                      {inter.drafted_reply}
                                    </div>

                                    <div className="suggested-reply-actions">
                                      <button
                                        className="btn btn-sm"
                                        onClick={() => handleOpenComposer(inter, inter.drafted_reply)}
                                        style={{
                                          fontSize: '12px',
                                          padding: '6px 14px',
                                          borderRadius: '6px',
                                          backgroundColor: '#4F46E5',
                                          color: '#fff',
                                          border: 'none',
                                          cursor: 'pointer',
                                          fontWeight: '600',
                                          display: 'inline-flex',
                                          alignItems: 'center',
                                          gap: '5px'
                                        }}
                                      >
                                        <Edit3 size={13} /> Edit in Composer
                                      </button>
                                      <button
                                        className="btn btn-sm"
                                        onClick={() => handleSendSuggestedReplyDirect(inter, inter.drafted_reply)}
                                        disabled={sendingReply}
                                        style={{
                                          fontSize: '12px',
                                          padding: '6px 14px',
                                          borderRadius: '6px',
                                          backgroundColor: '#059669',
                                          color: '#fff',
                                          border: 'none',
                                          cursor: 'pointer',
                                          fontWeight: '600',
                                          display: 'inline-flex',
                                          alignItems: 'center',
                                          gap: '5px'
                                        }}
                                      >
                                        <Send size={13} />
                                        {sendingReply && replyingToId === inter.id ? 'Sending...' : 'Send Direct Reply'}
                                      </button>
                                      <button
                                        className="btn btn-sm"
                                        onClick={() => {
                                          setDiscardedSuggestionIds(prev => {
                                            const next = new Set(prev);
                                            next.add(inter.id);
                                            return next;
                                          });
                                        }}
                                        style={{
                                          fontSize: '12px',
                                          padding: '6px 14px',
                                          borderRadius: '6px',
                                          backgroundColor: '#E2E8F0',
                                          color: '#475569',
                                          border: 'none',
                                          cursor: 'pointer',
                                          fontWeight: '600',
                                          display: 'inline-flex',
                                          alignItems: 'center',
                                          gap: '5px'
                                        }}
                                      >
                                        <XCircle size={13} /> Discard
                                      </button>
                                    </div>
                                  </div>
                                )}

                                {/* Inline Reply Composer */}
                                {replyingToId === inter.id && (
                                  <div className="reply-composer" style={{ margin: '14px 20px 20px 20px' }}>
                                    <div className="composer-header">
                                      <div className="composer-header-title">
                                        <Reply size={16} /> Reply to {inter.sender || lead.email}
                                      </div>
                                      {draftStatus && (
                                        <span className={`draft-status-label ${draftStatus}`}>
                                          {draftStatus === 'saving' && '✍️ Saving draft...'}
                                          {draftStatus === 'saved' && '💾 Draft saved'}
                                          {draftStatus === 'error' && '❌ Failed to save draft'}
                                        </span>
                                      )}
                                    </div>
                                    <div className="composer-body">
                                      <div className="composer-field">
                                        <label><User size={13} /> From (Sender Account)</label>
                                        {availableMailboxes.length > 0 ? (
                                          <select 
                                            value={replyForm.from} 
                                            disabled={sendingReply}
                                            onChange={(e) => setReplyForm(prev => ({ ...prev, from: e.target.value }))}
                                            style={{
                                              width: '100%',
                                              padding: '8px 12px',
                                              borderRadius: '6px',
                                              border: '1px solid #CBD5E1',
                                              fontSize: '13px',
                                              backgroundColor: '#FFFFFF',
                                              color: '#0F172A',
                                              fontWeight: '500'
                                            }}
                                          >
                                            {availableMailboxes.map(mb => (
                                              <option key={mb.id || mb.email} value={mb.email}>
                                                {mb.email} {mb.is_primary ? '(Primary)' : ''}
                                              </option>
                                            ))}
                                          </select>
                                        ) : (
                                          <input 
                                            type="text" 
                                            value={replyForm.from} 
                                            disabled={sendingReply}
                                            onChange={(e) => setReplyForm(prev => ({ ...prev, from: e.target.value }))}
                                            placeholder="sender@domain.com"
                                          />
                                        )}
                                      </div>
                                      <div className="composer-field">
                                        <label><User size={13} /> To</label>
                                        <input 
                                          type="text" 
                                          value={replyForm.to} 
                                          disabled={sendingReply}
                                          onChange={(e) => setReplyForm(prev => ({ ...prev, to: e.target.value }))}
                                          placeholder="recipient@example.com"
                                        />
                                      </div>
                                      <div className="composer-field">
                                        <label><Mail size={13} /> CC (Optional)</label>
                                        <input 
                                          type="text" 
                                          value={replyForm.cc} 
                                          disabled={sendingReply}
                                          onChange={(e) => setReplyForm(prev => ({ ...prev, cc: e.target.value }))}
                                          placeholder="optional.cc@example.com"
                                        />
                                      </div>
                                      <div className="composer-field">
                                        <label><FileText size={13} /> Subject Line</label>
                                        <input 
                                          type="text" 
                                          value={replyForm.subject} 
                                          disabled={sendingReply}
                                          onChange={(e) => setReplyForm(prev => ({ ...prev, subject: e.target.value }))}
                                          placeholder="Re: Inquiry regarding freight shipment..."
                                        />
                                      </div>
                                      <div className="composer-field">
                                        <label><Edit3 size={13} /> Message Body</label>
                                        <textarea 
                                          rows={7}
                                          value={replyForm.body} 
                                          disabled={sendingReply}
                                          onChange={(e) => setReplyForm(prev => ({ ...prev, body: e.target.value }))}
                                          placeholder="Type your reply message here..."
                                        />
                                      </div>
                                    </div>
                                    {sendError && (
                                      <div className="composer-error">
                                        <AlertTriangle size={15} /> {sendError}
                                      </div>
                                    )}
                                    <div className="composer-actions">
                                      <button
                                        className="btn btn-secondary btn-sm"
                                        onClick={handleDiscardDraft}
                                        disabled={sendingReply}
                                        style={{
                                          fontSize: '12px',
                                          padding: '6px 14px',
                                          borderRadius: '6px',
                                          backgroundColor: '#FEF2F2',
                                          color: '#DC2626',
                                          border: '1px solid #FCA5A5',
                                          cursor: 'pointer',
                                          fontWeight: '600',
                                          display: 'inline-flex',
                                          alignItems: 'center',
                                          gap: '5px'
                                        }}
                                      >
                                        <Trash2 size={13} /> Discard Draft
                                      </button>
                                      
                                      <div style={{ display: 'flex', gap: '8px' }}>
                                        <button
                                          className="btn btn-secondary btn-sm"
                                          onClick={() => setReplyingToId(null)}
                                          disabled={sendingReply}
                                          style={{
                                            fontSize: '12px',
                                            padding: '6px 14px',
                                            borderRadius: '6px',
                                            backgroundColor: '#E2E8F0',
                                            color: '#475569',
                                            border: 'none',
                                            cursor: 'pointer',
                                            fontWeight: '600'
                                          }}
                                        >
                                          Cancel
                                        </button>
                                        <button
                                          className="btn btn-primary btn-sm"
                                          onClick={handleSendReply}
                                          disabled={sendingReply}
                                          style={{
                                            fontSize: '12.5px',
                                            padding: '6px 18px',
                                            borderRadius: '6px',
                                            background: 'linear-gradient(135deg, #6366F1 0%, #4F46E5 100%)',
                                            color: '#fff',
                                            border: 'none',
                                            cursor: 'pointer',
                                            fontWeight: '600',
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            gap: '6px',
                                            boxShadow: '0 2px 8px rgba(79, 70, 229, 0.3)'
                                          }}
                                        >
                                          <Send size={14} />
                                          {sendingReply ? 'Sending Email...' : 'Send Email'}
                                        </button>
                                      </div>
                                    </div>
                                  </div>
                                )}
                              </div>
                            </div>
                          );
                        })}

                        {/* Persistent Quick Reply Bar for active thread */}
                        {replyingToId === null && (
                          <div style={{
                            marginTop: '16px',
                            padding: '16px 20px',
                            backgroundColor: '#F8FAFC',
                            border: '1.5px dashed #CBD5E1',
                            borderRadius: '12px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'space-between',
                            flexWrap: 'wrap',
                            gap: '12px'
                          }} className="thread-quick-reply-bar">
                            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                              <div style={{
                                width: '36px',
                                height: '36px',
                                borderRadius: '50%',
                                backgroundColor: '#EEF2FF',
                                color: '#4F46E5',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center'
                              }}>
                                <Reply size={18} />
                              </div>
                              <div>
                                <div style={{ fontSize: '13.5px', fontWeight: '700', color: '#0F172A' }}>
                                  Reply to this Email Thread
                                </div>
                                <div style={{ fontSize: '12px', color: '#64748B' }}>
                                  Send a message directly in this thread ({activeGroup.subject || 'RFQ Conversation'})
                                </div>
                              </div>
                            </div>
                            <button
                              type="button"
                              onClick={() => {
                                const latestInter = sortedActiveList[0];
                                if (latestInter) {
                                  handleOpenComposer(latestInter);
                                } else {
                                  handleStartNewEmail('CUSTOM');
                                }
                              }}
                              style={{
                                fontSize: '12.5px',
                                fontWeight: '600',
                                padding: '8px 18px',
                                borderRadius: '8px',
                                background: 'linear-gradient(135deg, #6366F1 0%, #4F46E5 100%)',
                                color: '#ffffff',
                                border: 'none',
                                cursor: 'pointer',
                                display: 'inline-flex',
                                alignItems: 'center',
                                gap: '6px',
                                boxShadow: '0 2px 8px rgba(79, 70, 229, 0.25)'
                              }}
                            >
                              <Reply size={14} /> Reply to Thread
                            </button>
                          </div>
                        )}

                        {/* Part 8 — Older Conversations */}
                        {otherGroups.length > 0 && (
                          <div style={{ marginTop: '30px' }} className="older-threads-section">
                            <div style={{
                              fontSize: '13px',
                              fontWeight: '600',
                              color: '#64748B',
                              marginBottom: '12px',
                              borderBottom: '1px solid #E2E8F0',
                              paddingBottom: '6px'
                            }}>
                              Older Conversations
                            </div>
                            {otherGroups.map(group => {
                              const lastActDate = new Date(group.lastActivity).toLocaleDateString(undefined, {
                                month: 'short',
                                day: 'numeric',
                                hour: 'numeric',
                                minute: '2-digit'
                              });
                              const cleanSubject = group.subject ? group.subject.replace(/^re:\s*/i, '') : 'Inquiry';
                              return (
                                <div
                                  key={group.thread_id}
                                  onClick={() => setSelectedThreadId(group.thread_id)}
                                  style={{
                                    padding: '12px 16px',
                                    backgroundColor: '#F8FAFC',
                                    border: '1px solid #E2E8F0',
                                    borderRadius: '6px',
                                    cursor: 'pointer',
                                    marginBottom: '8px',
                                    transition: 'all 0.2s',
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    alignItems: 'center'
                                  }}
                                  className="older-thread-card"
                                >
                                  <div>
                                    <strong style={{ fontSize: '13px', color: '#334155', display: 'block' }}>
                                      RFQ: {cleanSubject}
                                    </strong>
                                    <span style={{ fontSize: '11px', color: '#94A3B8' }}>
                                      Last active: {lastActDate}
                                    </span>
                                  </div>
                                  <span style={{
                                    fontSize: '12px',
                                    color: '#475569',
                                    backgroundColor: '#E2E8F0',
                                    padding: '2px 8px',
                                    borderRadius: '12px',
                                    fontWeight: '500'
                                  }}>
                                    {group.interactions.length} msg{group.interactions.length !== 1 ? 's' : ''}
                                  </span>
                                </div>
                              );
                            })}
                          </div>
                        )}
                      </div>
                    );
                  })()}
                </>
              )}
            </div>
          </div>
        )}

        {activeTab === 'timeline' && (
          <div className="panel-section">
            <h3 className="panel-section-title" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span>Major Work Milestones</span>
              <span style={{ fontSize: '11px', fontWeight: '500', color: '#64748b' }}>Filtered major activities</span>
            </h3>

            {loadingTimeline ? (
              <div className="timeline-loading">Loading activity timeline...</div>
            ) : (() => {
              // 1. Filter out minor noise (tag edits, assigned owner changes, uninformative duplicates)
              const majorEvents = timeline.filter(event => {
                if (!event.action) return false;
                const act = event.action;
                const desc = (event.description || '').toLowerCase();
                
                // Exclude minor field/owner/tag noise
                if (act === 'OWNER_CHANGED') return false;
                if (act === 'TAGS_CHANGED') return false;
                if (desc.includes('tags updated to')) return false;
                if (desc.includes('assigned to user')) return false;

                // Keep major milestones
                if (act === 'CREATED' || act === 'CONVERTED' || act === 'STATUS_CHANGED') return true;
                if (act.startsWith('EMAIL') || desc.includes('email') || desc.includes('inquiry')) return true;
                if (desc.includes('rfq') || desc.includes('rate') || desc.includes('created')) return true;

                return true;
              });

              if (majorEvents.length === 0) {
                return <div className="timeline-empty">No major milestones logged yet.</div>;
              }

              return (
                <div className="timeline-scroll-container">
                  <div 
                    className="timeline-list" 
                    ref={timelineScrollRef} 
                    onScroll={handleTimelineScroll}
                  >
                    {majorEvents.map((event, idx) => {
                      const isExpanded = expandedTimelineIdx === idx;
                      const { dateStr, timeStr } = formatTimelineTime(event.timestamp);

                      // Categorize icon, title, and AI sentiment analysis
                      let icon = '📝';
                      let milestoneType = 'SYSTEM_EVENT';
                      let sentimentBadge = { label: '🟢 POSITIVE', type: 'positive' };
                      let aiTakeaway = '';

                      const descLower = (event.description || '').toLowerCase();

                      if (event.action === 'CREATED' || descLower.includes('lead details were created')) {
                        icon = '🚀';
                        milestoneType = 'LEAD_CREATED';
                        aiTakeaway = 'Inbound shipping lead initiated into LogisticsHQ workspace.';
                        sentimentBadge = { label: '🟢 POSITIVE — New Lead Inquiry', type: 'positive' };
                      } else if (event.action === 'CONVERTED' || descLower.includes('converted')) {
                        icon = '📦';
                        milestoneType = 'RFQ_CONVERTED';
                        aiTakeaway = 'Inquiry successfully verified and converted into formal RFQ shipment record.';
                        sentimentBadge = { label: '🟢 POSITIVE — RFQ Created', type: 'positive' };
                      } else if (event.action === 'STATUS_CHANGED') {
                        icon = '⚡';
                        milestoneType = 'STATUS_CHANGED';
                        aiTakeaway = `Pipeline progression: ${event.description}`;
                        sentimentBadge = { label: '⚡ PIPELINE UPDATE', type: 'info' };
                      } else if (event.action.startsWith('EMAIL_INBOUND') || descLower.includes('inbound') || descLower.includes('received')) {
                        icon = '📥';
                        milestoneType = 'EMAIL_INBOUND';
                        aiTakeaway = 'Customer sent email containing freight shipment specifications.';
                        sentimentBadge = { label: '🟢 POSITIVE — Customer Response', type: 'positive' };
                      } else if (event.action.startsWith('EMAIL_OUTBOUND_FAILED') || descLower.includes('failed')) {
                        icon = '⚠️';
                        milestoneType = 'EMAIL_FAILED';
                        aiTakeaway = 'Automated outbound email failed delivery. Retry required.';
                        sentimentBadge = { label: '🔴 URGENT — Delivery Attention Required', type: 'danger' };
                      } else if (event.action.startsWith('EMAIL_OUTBOUND') || descLower.includes('sent')) {
                        icon = '📤';
                        milestoneType = 'EMAIL_OUTBOUND';
                        aiTakeaway = 'Automated/manual email reply sent to customer.';
                        sentimentBadge = { label: '🔵 SENT — Outbound Response', type: 'info' };
                      } else {
                        icon = '📋';
                        milestoneType = 'MILESTONE';
                        aiTakeaway = event.description;
                        sentimentBadge = { label: '🟢 LOGGED WORK', type: 'positive' };
                      }

                      let interactionId = null;
                      if (event.action === 'EMAIL_OUTBOUND_FAILED') {
                        const match = event.description.match(/failed to send clarification email for interaction (\d+)/i);
                        if (match) {
                          interactionId = parseInt(match[1]);
                        }
                      }

                      return (
                        <div 
                          key={idx} 
                          className="major-milestone-wrapper"
                          style={{ display: 'flex', gap: '12px', marginBottom: '14px', position: 'relative' }}
                        >
                          {/* Left Stem Line & Avatar Circle */}
                          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', minWidth: '28px' }}>
                            <div style={{
                              width: '28px',
                              height: '28px',
                              borderRadius: '50%',
                              backgroundColor: milestoneType === 'EMAIL_INBOUND' ? '#d1fae5' : milestoneType === 'EMAIL_OUTBOUND' ? '#dbeafe' : '#e0e7ff',
                              color: milestoneType === 'EMAIL_INBOUND' ? '#047857' : milestoneType === 'EMAIL_OUTBOUND' ? '#1d4ed8' : '#4338ca',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              fontSize: '14px',
                              boxShadow: '0 1px 3px rgba(15,23,42,0.08)'
                            }}>
                              {icon}
                            </div>
                            {idx !== majorEvents.length - 1 && (
                              <div style={{ flex: 1, width: '2px', backgroundColor: '#e2e8f0', margin: '4px 0 0 0' }} />
                            )}
                          </div>

                          {/* Right Milestone Card */}
                          <div
                            onClick={() => toggleExpandTimeline(idx)}
                            style={{
                              flex: 1,
                              minWidth: 0,
                              border: isExpanded ? '1.5px solid #3b82f6' : '1px solid #e2e8f0',
                              borderRadius: '10px',
                              padding: '12px 14px',
                              backgroundColor: isExpanded ? '#f0f9ff' : '#ffffff',
                              boxShadow: isExpanded ? '0 4px 14px rgba(59, 130, 246, 0.12)' : '0 1px 3px rgba(15,23,42,0.03)',
                              cursor: 'pointer',
                              transition: 'all 0.2s ease'
                            }}
                          >
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '8px' }}>
                              <span style={{ fontWeight: '700', fontSize: '13px', color: '#0f172a', lineHeight: '1.4' }}>
                                {formatTimelineDescription(event.description)}
                              </span>
                              <span style={{ fontSize: '11px', color: '#2563eb', fontWeight: '600', flexShrink: 0 }}>
                                {isExpanded ? '▲ Hide Details' : '▼ View Details & AI Analysis'}
                              </span>
                            </div>

                            {/* Collapsed summary pill */}
                            {!isExpanded && (
                              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginTop: '6px' }}>
                                <span style={{
                                  fontSize: '11px',
                                  fontWeight: '600',
                                  padding: '2px 8px',
                                  borderRadius: '10px',
                                  backgroundColor: sentimentBadge.type === 'positive' ? '#d1fae5' : sentimentBadge.type === 'danger' ? '#fee2e2' : '#dbeafe',
                                  color: sentimentBadge.type === 'positive' ? '#047857' : sentimentBadge.type === 'danger' ? '#b91c1c' : '#1d4ed8'
                                }}>
                                  {sentimentBadge.label}
                                </span>
                                <span style={{ fontSize: '11px', color: '#94a3b8' }}>
                                  {dateStr}, {timeStr}
                                </span>
                              </div>
                            )}

                            {/* Expanded details with AI Sentiment Analysis & Takeaways */}
                            {isExpanded && (
                              <div style={{ marginTop: '12px', paddingTop: '10px', borderTop: '1px solid #bae6fd' }}>
                                <div style={{
                                  padding: '10px 12px',
                                  backgroundColor: '#ffffff',
                                  border: '1px solid #93c5fd',
                                  borderRadius: '8px',
                                  marginBottom: '8px',
                                  boxShadow: '0 1px 3px rgba(0,0,0,0.02)'
                                }}>
                                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '6px' }}>
                                    <span style={{ fontSize: '11px', fontWeight: '700', color: '#0369a1', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                                      🤖 AI Sentiment & Intent Analysis
                                    </span>
                                    <span style={{
                                      fontSize: '10.5px',
                                      fontWeight: '700',
                                      padding: '1px 6px',
                                      borderRadius: '4px',
                                      backgroundColor: sentimentBadge.type === 'positive' ? '#d1fae5' : '#dbeafe',
                                      color: sentimentBadge.type === 'positive' ? '#047857' : '#1d4ed8'
                                    }}>
                                      {sentimentBadge.label}
                                    </span>
                                  </div>
                                  <p style={{ margin: 0, fontSize: '12.5px', color: '#334155', lineHeight: '1.45' }}>
                                    {aiTakeaway}
                                  </p>
                                </div>

                                {interactionId && (
                                  <div style={{ marginTop: '8px' }}>
                                    <button
                                      className="btn btn-secondary btn-sm"
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        handleRetry(interactionId);
                                      }}
                                      disabled={retryingMap[interactionId]}
                                      style={{
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        gap: '6px',
                                        fontSize: '12px',
                                        padding: '4px 10px',
                                        borderRadius: '4px',
                                        backgroundColor: '#2563eb',
                                        color: '#fff',
                                        border: 'none',
                                        cursor: 'pointer'
                                      }}
                                    >
                                      {retryingMap[interactionId] ? 'Resending...' : '🔄 Retry Sending Reply'}
                                    </button>
                                  </div>
                                )}

                                <div style={{ marginTop: '6px', fontSize: '11px', color: '#64748b', display: 'flex', alignItems: 'center', gap: '6px' }}>
                                  <span>Log Actor: <strong>{event.actor || 'System'}</strong></span>
                                  <span>·</span>
                                  <span>{dateStr}, {timeStr}</span>
                                </div>
                              </div>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              );
            })()}
          </div>
        )}

        {activeTab === 'outreach' && (
          <div className="panel-section">
            <h3 className="panel-section-title">Outreach Activity Timeline</h3>
            {outreachActivities.length === 0 ? (
              <div className="timeline-empty">No outreach activities logged for this lead.</div>
            ) : (
              <div className="timeline-scroll-container">
                <div className="timeline-list" style={{ paddingLeft: 10 }}>
                  {outreachActivities.map((act) => (
                    <div key={act.id} className="timeline-item-container" style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
                      <div style={{ fontSize: 16, marginTop: 2 }}>
                        {act.activity_type === 'EMAIL' ? '✉️' : act.activity_type === 'CALL' ? '📞' : act.activity_type === 'MEETING' ? '👥' : '🔔'}
                      </div>
                      <div style={{ flex: 1, background: '#F8FAFC', padding: 12, borderRadius: 8, border: '1px solid #E2E8F0' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <span style={{ fontWeight: 700, color: '#0F172A', fontSize: 13 }}>{act.subject}</span>
                          <span style={{ fontSize: 11, color: '#64748B' }}>
                            {new Date(act.created_at).toLocaleDateString()}
                          </span>
                        </div>
                        {act.description && (
                          <div style={{ fontSize: 12.5, color: '#475569', marginTop: 4, background: '#FFFFFF', padding: 6, borderRadius: 6, border: '1px solid #F1F5F9' }}>
                            {act.description}
                          </div>
                        )}
                        <div style={{ display: 'flex', gap: 10, marginTop: 6, alignItems: 'center' }}>
                          <span className={`activity-status-pill-grad ${act.status.toLowerCase()}`} style={{ fontSize: 10, padding: '2px 6px' }}>
                            {act.status}
                          </span>
                          <span style={{ fontSize: 11, color: '#64748B' }}>Creator: <strong>{act.creator_name || 'System'}</strong></span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Sticky Panel Footer Actions */}
      <div className="panel-footer">
        {isConverted ? (
          <button
            className="panel-btn panel-btn-success"
            onClick={() => {
              if (createdRFQId) {
                navigate(`/dashboard/rfqs/${createdRFQId}`);
              } else {
                navigate('/dashboard/rfqs');
              }
            }}
            style={{ backgroundColor: '#10b981', color: '#ffffff', fontWeight: '700', fontSize: '13px' }}
          >
            ✓ Converted to RFQ {createdRFQNumber ? `(${createdRFQNumber})` : (createdRFQId ? `(#${createdRFQId})` : '')} — View RFQ →
          </button>
        ) : (
          <div style={{ display: 'flex', gap: '8px', width: '100%' }}>
            <button
              className="panel-btn panel-btn-primary"
              style={{ flex: 1 }}
              onClick={() => onConvertToRFQ(lead)}
            >
              📋 Convert to RFQ
            </button>
            <button
              className="panel-btn"
              style={{ flex: 1, backgroundColor: '#1e293b', color: '#ffffff', fontWeight: '600' }}
              onClick={() => onConvertToCustomer && onConvertToCustomer(lead)}
            >
              🏢 Convert to Customer
            </button>
          </div>
        )}

        <div className="panel-footer-row">
          <button
            className={`panel-btn panel-btn-save`}
            onClick={handleSave}
            disabled={!isDirty || updating}
          >
            💾 {updating ? 'Saving...' : 'Save Changes'}
          </button>
          <button
            className="panel-btn panel-btn-danger"
            onClick={handleDelete}
            disabled={deleting}
          >
            {deleting ? 'Deleting...' : '🗑️ Delete'}
          </button>
        </div>
      </div>
    </div>
  );
}

LeadDetailPanel.propTypes = {
  lead: PropTypes.object,
  onClose: PropTypes.func.isRequired,
  onLeadUpdated: PropTypes.func.isRequired,
  onConvertToRFQ: PropTypes.func.isRequired,
  onDirtyChange: PropTypes.func,
  users: PropTypes.array,
};

// ── Custom Dropdown Component ──
function CustomSelect({ name, value, onChange, options, disabled, placeholder }) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef(null);

  useEffect(() => {
    const handleOutsideClick = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    if (isOpen) document.addEventListener('mousedown', handleOutsideClick);
    return () => document.removeEventListener('mousedown', handleOutsideClick);
  }, [isOpen]);

  const selectedOption = options.find(o => o.value === value);

  return (
    <div className={`leads-custom-dropdown ${disabled ? 'disabled' : ''} ${isOpen ? 'open' : ''}`} ref={containerRef}>
      <div 
        className="leads-custom-dropdown-selected" 
        onClick={() => !disabled && setIsOpen(!isOpen)}
      >
        {selectedOption ? selectedOption.label : <span className="placeholder">{placeholder || 'Select...'}</span>}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </div>
      {isOpen && !disabled && (
        <ul className="leads-custom-dropdown-list">
          {options.map(opt => (
            <li
              key={opt.value}
              className={value === opt.value ? 'selected' : ''}
              onClick={(e) => {
                e.stopPropagation();
                onChange({ target: { name, value: opt.value } });
                setIsOpen(false);
              }}
            >
              {opt.label}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

