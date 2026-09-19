import React, { useState, useEffect, useMemo, useCallback } from 'react';
import {
  Inbox,
  Search,
  RefreshCw,
  Clock,
  User,
  CheckCircle2,
  AlertCircle,
  Eye,
  Mail,
  Phone,
  Building2,
  Globe,
  ArrowRight,
  X,
  UserCheck,
  Check,
  ExternalLink,
  MessageSquare,
  Sparkles,
  Layers,
  Ship,
  Package,
  Calendar,
  Copy,
  TrendingUp,
  SlidersHorizontal,
  ChevronRight,
  ShieldCheck,
  Briefcase,
  Cpu,
  Compass,
  Zap,
  Tag
} from 'lucide-react';
import PageHeader from '../../components/common/PageHeader';
import EmptyState from '../../components/common/EmptyState';
import { sportalService } from '../../services/sportalService';

// Color map for avatar monograms based on initials
const AVATAR_GRADIENTS = [
  'from-blue-600 to-indigo-600',
  'from-sky-500 to-blue-600',
  'from-indigo-500 to-purple-600',
  'from-emerald-500 to-teal-600',
  'from-amber-500 to-orange-600',
  'from-purple-500 to-pink-600',
];

const getAvatarGradient = (name = '') => {
  const code = (name.charCodeAt(0) || 0) + (name.charCodeAt(1) || 0);
  return AVATAR_GRADIENTS[code % AVATAR_GRADIENTS.length];
};

// Helper to parse forwarder metadata packed in message
const parseForwarderMetadata = (rawMsg = '') => {
  if (!rawMsg) return { notes: '', role: '', corridor: '', tms: '' };
  const parts = rawMsg.split(' | ');
  let notes = '';
  let role = '';
  let corridor = '';
  let tms = '';
  parts.forEach(part => {
    const trimmed = part.trim();
    if (trimmed.startsWith('Notes:')) notes = trimmed.replace(/^Notes:\s*/, '');
    else if (/^(Forwarder )?Role:/i.test(trimmed)) role = trimmed.replace(/^(Forwarder )?Role:\s*/i, '');
    else if (/^(Primary )?Corridor:/i.test(trimmed)) corridor = trimmed.replace(/^(Primary )?Corridor:\s*/i, '');
    else if (/^(Current )?TMS:/i.test(trimmed)) tms = trimmed.replace(/^(Current )?TMS:\s*/i, '');
    else if (!notes) notes = trimmed;
  });
  return { notes: notes || rawMsg, role, corridor, tms };
};

export function DemoRequestsPage() {
  const [requests, setRequests] = useState([]);
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('ALL');
  const [page, setPage] = useState(1);
  const [copiedField, setCopiedField] = useState(null);
  const limit = 25;

  // Selected item modal
  const [selectedLead, setSelectedLead] = useState(null);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [feedbackMsg, setFeedbackMsg] = useState('');

  // Editable fields in modal
  const [editStatus, setEditStatus] = useState('');
  const [editNotes, setEditNotes] = useState('');
  const [editAssignedTo, setEditAssignedTo] = useState('');

  // Fetch data
  const fetchDemoRequests = useCallback(async () => {
    setLoading(true);
    try {
      const res = await sportalService.listDemoRequests({
        status: statusFilter,
        search: search.trim(),
        page,
        limit,
      });

      if (res?.items) {
        setRequests(res.items);
        setTotalCount(res.total ?? res.items.length);
      } else if (res?.data?.items) {
        setRequests(res.data.items);
        setTotalCount(res.data.total ?? res.data.items.length);
      } else if (Array.isArray(res)) {
        setRequests(res);
        setTotalCount(res.length);
      } else if (Array.isArray(res?.data)) {
        setRequests(res.data);
        setTotalCount(res.data.length);
      }
    } catch (err) {
      console.error('Failed to load demo requests:', err);
    } finally {
      setLoading(false);
    }
  }, [statusFilter, search, page]);

  useEffect(() => {
    fetchDemoRequests();
  }, [fetchDemoRequests]);

  // Derived KPI metrics
  const kpis = useMemo(() => {
    const total = totalCount || requests.length;
    const newCount = requests.filter(r => r.status === 'NEW').length;
    const contactedCount = requests.filter(r => r.status === 'CONTACTED' || r.status === 'QUALIFIED').length;
    const convertedCount = requests.filter(r => r.status === 'CONVERTED').length;
    return {
      total,
      newCount,
      contactedCount,
      convertedCount,
    };
  }, [requests, totalCount]);

  const handleOpenDetail = (lead) => {
    setSelectedLead(lead);
    setEditStatus(lead.status || 'NEW');
    setEditNotes(lead.notes || '');
    setEditAssignedTo(lead.assigned_to || '');
    setFeedbackMsg('');
    setDetailModalOpen(true);
  };

  const handleSaveLead = async () => {
    if (!selectedLead) return;
    setUpdating(true);
    setFeedbackMsg('');
    try {
      const payload = {
        status: editStatus,
        notes: editNotes,
        assigned_to: editAssignedTo,
      };
      const updated = await sportalService.updateDemoRequest(selectedLead.id, payload);
      setFeedbackMsg('Lead updated successfully!');
      
      const updatedItem = (updated?.id ? updated : updated?.data) || { ...selectedLead, ...payload };
      setSelectedLead(updatedItem);
      setRequests(prev => prev.map(item => item.id === selectedLead.id ? updatedItem : item));
      
      setTimeout(() => {
        setFeedbackMsg('');
      }, 3000);
    } catch (err) {
      console.error('Failed to update demo request:', err);
      setFeedbackMsg('Error: ' + (err.message || 'Failed to update'));
    } finally {
      setUpdating(false);
    }
  };

  const copyToClipboard = (text, fieldName) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopiedField(fieldName);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const formatDate = (isoStr) => {
    if (!isoStr) return '—';
    try {
      const d = new Date(isoStr);
      return d.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return isoStr;
    }
  };

  const formatRelativeTime = (isoStr) => {
    if (!isoStr) return '—';
    try {
      const diffMs = Date.now() - new Date(isoStr).getTime();
      const diffMins = Math.floor(diffMs / 60000);
      if (diffMins < 1) return 'Just now';
      if (diffMins < 60) return `${diffMins}m ago`;
      const diffHours = Math.floor(diffMins / 60);
      if (diffHours < 24) return `${diffHours}h ago`;
      const diffDays = Math.floor(diffHours / 24);
      if (diffDays === 1) return 'Yesterday';
      if (diffDays < 7) return `${diffDays}d ago`;
      return formatDate(isoStr).split(',')[0];
    } catch {
      return formatDate(isoStr);
    }
  };

  const renderStatusBadge = (status) => {
    switch (status) {
      case 'NEW':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-sky-50 text-sky-700 border border-sky-200 shadow-2xs">
            <span className="w-1.5 h-1.5 rounded-full bg-sky-500 animate-pulse" />
            New Lead
          </span>
        );
      case 'CONTACTED':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200 shadow-2xs">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-500" />
            Contacted
          </span>
        );
      case 'QUALIFIED':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-purple-50 text-purple-700 border border-purple-200 shadow-2xs">
            <span className="w-1.5 h-1.5 rounded-full bg-purple-500" />
            Qualified
          </span>
        );
      case 'CONVERTED':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 shadow-2xs">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            Converted
          </span>
        );
      case 'DISMISSED':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-600 border border-slate-200 shadow-2xs">
            <span className="w-1.5 h-1.5 rounded-full bg-slate-400" />
            Dismissed
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-700">
            {status}
          </span>
        );
    }
  };

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      
      {/* ═══ PAGE HEADER ═══ */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-1">
        <div>
          <div className="flex items-center gap-2 mb-1.5">
            <span className="px-2.5 py-0.5 text-[11px] font-bold tracking-wider uppercase rounded-full bg-blue-50 text-blue-700 border border-blue-200/80">
              Freight Forwarder Inbound
            </span>
            <span className="text-xs text-slate-400 font-medium">• Commercial Operations</span>
          </div>
          <h1 className="text-2xl font-black text-slate-900 tracking-tight">
            Demo Requests & Inbound Pipeline
          </h1>
          <p className="text-xs sm:text-sm text-slate-500 mt-1">
            Review freight forwarders, NVOCCs, and logistics brokers requesting automated tariff and quoting walkthroughs.
          </p>
        </div>

        <button
          onClick={() => fetchDemoRequests()}
          disabled={loading}
          className="self-start sm:self-auto inline-flex items-center gap-2 px-4 py-2 text-xs font-bold text-slate-700 bg-white border border-slate-200 rounded-xl hover:bg-slate-50 hover:border-slate-300 transition-all shadow-2xs disabled:opacity-50"
        >
          <RefreshCw className={`w-3.5 h-3.5 text-slate-500 ${loading ? 'animate-spin text-blue-600' : ''}`} />
          <span>Sync Pipeline</span>
        </button>
      </div>

      {/* ═══ STUNNING BALANCED KPI CARDS ═══ */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-5">
        
        {/* Card 1: Total Leads */}
        <div className="relative overflow-hidden rounded-2xl bg-white border border-slate-200/90 p-5 shadow-xs hover:shadow-md transition-all duration-200 group">
          <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-blue-600 to-indigo-600" />
          <div className="flex items-start justify-between">
            <div className="w-10 h-10 rounded-xl bg-blue-50 border border-blue-100 flex items-center justify-center text-blue-600 shadow-2xs group-hover:scale-105 transition-transform">
              <Inbox className="w-5 h-5" />
            </div>
            <span className="text-[10px] font-bold uppercase tracking-wider px-2.5 py-0.5 rounded-full bg-blue-50 text-blue-700 border border-blue-200/70">
              Total Inquiries
            </span>
          </div>
          <div className="mt-4">
            <div className="text-3xl font-extrabold text-slate-900 tracking-tight">
              {kpis.total}
            </div>
            <div className="text-xs text-slate-500 mt-1 font-medium">
              Registered forwarder leads
            </div>
          </div>
          <div className="mt-3.5 pt-3 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-500">
            <span>Website intake</span>
            <span className="font-semibold text-blue-600">100% Inbound</span>
          </div>
        </div>

        {/* Card 2: Awaiting Action */}
        <div className="relative overflow-hidden rounded-2xl bg-white border border-slate-200/90 p-5 shadow-xs hover:shadow-md transition-all duration-200 group">
          <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-amber-400 to-orange-500" />
          <div className="flex items-start justify-between">
            <div className="w-10 h-10 rounded-xl bg-amber-50 border border-amber-100 flex items-center justify-center text-amber-600 shadow-2xs group-hover:scale-105 transition-transform">
              <Clock className="w-5 h-5" />
            </div>
            <span className="inline-flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-wider px-2.5 py-0.5 rounded-full bg-amber-50 text-amber-700 border border-amber-200/70">
              <span className="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse" />
              Requires Action
            </span>
          </div>
          <div className="mt-4">
            <div className="text-3xl font-extrabold text-amber-600 tracking-tight">
              {kpis.newCount}
            </div>
            <div className="text-xs text-slate-500 mt-1 font-medium">
              Awaiting initial outreach
            </div>
          </div>
          <div className="mt-3.5 pt-3 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-500">
            <span>Target Response</span>
            <span className="font-semibold text-amber-600">&lt; 2h SLA</span>
          </div>
        </div>

        {/* Card 3: In Active Qualification */}
        <div className="relative overflow-hidden rounded-2xl bg-white border border-slate-200/90 p-5 shadow-xs hover:shadow-md transition-all duration-200 group">
          <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-purple-500 to-indigo-600" />
          <div className="flex items-start justify-between">
            <div className="w-10 h-10 rounded-xl bg-purple-50 border border-purple-100 flex items-center justify-center text-purple-600 shadow-2xs group-hover:scale-105 transition-transform">
              <UserCheck className="w-5 h-5" />
            </div>
            <span className="text-[10px] font-bold uppercase tracking-wider px-2.5 py-0.5 rounded-full bg-purple-50 text-purple-700 border border-purple-200/70">
              In Qualification
            </span>
          </div>
          <div className="mt-4">
            <div className="text-3xl font-extrabold text-purple-700 tracking-tight">
              {kpis.contactedCount}
            </div>
            <div className="text-xs text-slate-500 mt-1 font-medium">
              Demonstration & sandbox test
            </div>
          </div>
          <div className="mt-3.5 pt-3 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-500">
            <span>Sandbox status</span>
            <span className="font-semibold text-purple-600">Active Pilots</span>
          </div>
        </div>

        {/* Card 4: Converted to Customer */}
        <div className="relative overflow-hidden rounded-2xl bg-white border border-slate-200/90 p-5 shadow-xs hover:shadow-md transition-all duration-200 group">
          <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-emerald-500 to-teal-600" />
          <div className="flex items-start justify-between">
            <div className="w-10 h-10 rounded-xl bg-emerald-50 border border-emerald-100 flex items-center justify-center text-emerald-600 shadow-2xs group-hover:scale-105 transition-transform">
              <CheckCircle2 className="w-5 h-5" />
            </div>
            <span className="text-[10px] font-bold uppercase tracking-wider px-2.5 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200/70">
              Won Forwarders
            </span>
          </div>
          <div className="mt-4">
            <div className="text-3xl font-extrabold text-emerald-600 tracking-tight">
              {kpis.convertedCount}
            </div>
            <div className="text-xs text-slate-500 mt-1 font-medium">
              Signed commercial contracts
            </div>
          </div>
          <div className="mt-3.5 pt-3 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-500">
            <span>Conversion Rate</span>
            <span className="font-semibold text-emerald-600">
              {kpis.total > 0 ? `${Math.round((kpis.convertedCount / kpis.total) * 100)}% Won` : '—'}
            </span>
          </div>
        </div>

      </div>

      {/* ═══ SEARCH & SEGMENTED FILTER TOOLBAR ═══ */}
      <div className="bg-white p-3 sm:p-4 rounded-2xl border border-slate-200 shadow-2xs flex flex-col md:flex-row gap-3 items-stretch md:items-center justify-between">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            placeholder="Search by forwarder name, company, trade lane..."
            className="w-full pl-9 pr-4 py-2 text-xs sm:text-sm bg-slate-50/80 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white text-slate-800 placeholder-slate-400 transition"
          />
          {search && (
            <button
              onClick={() => setSearch('')}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>

        {/* Status Filter Tabs */}
        <div className="flex items-center gap-1 overflow-x-auto p-1 bg-slate-100/90 rounded-xl border border-slate-200/60 scrollbar-none">
          {[
            { id: 'ALL', label: 'All Leads', count: kpis.total },
            { id: 'NEW', label: 'New', count: kpis.newCount },
            { id: 'CONTACTED', label: 'Contacted' },
            { id: 'QUALIFIED', label: 'Qualified' },
            { id: 'CONVERTED', label: 'Converted', count: kpis.convertedCount },
            { id: 'DISMISSED', label: 'Dismissed' },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => {
                setStatusFilter(tab.id);
                setPage(1);
              }}
              className={`px-3 py-1.5 text-xs font-semibold rounded-lg transition-all whitespace-nowrap flex items-center gap-1.5 ${
                statusFilter === tab.id
                  ? 'bg-white text-slate-900 shadow-2xs font-bold'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-slate-200/50'
              }`}
            >
              <span>{tab.label}</span>
              {typeof tab.count === 'number' && (
                <span className={`text-[10px] px-1.5 py-0.2 rounded-full ${
                  statusFilter === tab.id
                    ? 'bg-blue-100 text-blue-700 font-bold'
                    : 'bg-slate-200 text-slate-600'
                }`}>
                  {tab.count}
                </span>
              )}
            </button>
          ))}
        </div>
      </div>

      {/* ═══ STREAMLINED 4-COLUMN LEADS TABLE ═══ */}
      <div className="bg-white rounded-2xl border border-slate-200 shadow-2xs overflow-hidden">
        {loading ? (
          <div className="py-24 text-center text-slate-500 flex flex-col items-center justify-center space-y-3">
            <RefreshCw className="w-8 h-8 animate-spin text-blue-600" />
            <p className="text-sm font-semibold text-slate-700">Loading inbound demo pipeline...</p>
            <p className="text-xs text-slate-400">Fetching latest inquiries and forwarder telemetry</p>
          </div>
        ) : requests.length === 0 ? (
          <div className="py-16 text-center">
            <EmptyState
              icon={Inbox}
              title="No demo requests found"
              description={search || statusFilter !== 'ALL' ? 'No prospects matched your current search filters. Try clearing or relaxing your filter.' : 'Inbound demo requests submitted via the public website will appear here for review.'}
            />
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs sm:text-sm">
              <thead>
                <tr className="bg-slate-50/90 border-b border-slate-200 text-[11px] font-bold text-slate-400 uppercase tracking-wider">
                  <th className="py-3.5 px-6 w-[40%]">Prospect & Company</th>
                  <th className="py-3.5 px-5 w-[25%]">Operational Scale</th>
                  <th className="py-3.5 px-5 w-[20%]">Status & Activity</th>
                  <th className="py-3.5 px-6 w-[15%] text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {requests.map((lead) => {
                  const gradient = getAvatarGradient(lead.full_name || lead.company_name);
                  const initial = lead.full_name ? lead.full_name.charAt(0).toUpperCase() : 'F';
                  const parsed = parseForwarderMetadata(lead.message);

                  return (
                    <tr
                      key={lead.id}
                      onClick={() => handleOpenDetail(lead)}
                      className="hover:bg-slate-50/70 transition-colors cursor-pointer group"
                    >
                      {/* Column 1: Prospect & Forwarder */}
                      <td className="py-4.5 px-6">
                        <div className="flex items-center gap-3.5">
                          {/* Vibrant Monogram Avatar */}
                          <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${gradient} text-white flex items-center justify-center font-bold text-sm shadow-xs flex-shrink-0 group-hover:scale-105 transition-transform`}>
                            {initial}
                          </div>
                          
                          <div className="min-w-0">
                            <div className="font-bold text-slate-900 group-hover:text-blue-600 transition-colors flex items-center gap-2 truncate">
                              <span className="text-sm">{lead.full_name}</span>
                              {lead.status === 'NEW' && (
                                <span className="w-2 h-2 rounded-full bg-sky-500 animate-ping flex-shrink-0" title="New uncontacted lead" />
                              )}
                            </div>
                            
                            <div className="text-xs text-slate-500 flex items-center gap-2 mt-0.5 truncate">
                              <span className="font-medium text-slate-700 flex items-center gap-1 truncate">
                                <Building2 className="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
                                {lead.company_name}
                              </span>
                              {lead.country && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-slate-100 text-slate-600 border border-slate-200/60 font-medium flex-shrink-0">
                                  {lead.country}
                                </span>
                              )}
                            </div>

                            {parsed.role && (
                              <div className="text-[11px] text-slate-400 truncate mt-0.5">
                                {parsed.role}
                              </div>
                            )}
                          </div>
                        </div>
                      </td>

                      {/* Column 2: Operational Scale */}
                      <td className="py-4.5 px-5">
                        <div className="space-y-1">
                          {lead.shipment_volume ? (
                            <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-lg bg-slate-100 text-slate-800 text-xs font-semibold border border-slate-200/70">
                              <Ship className="w-3.5 h-3.5 text-blue-600 flex-shrink-0" />
                              <span>{lead.shipment_volume}</span>
                            </div>
                          ) : (
                            <span className="text-xs text-slate-400">Scale unspec.</span>
                          )}

                          <div className="flex items-center gap-2 text-[11px] text-slate-500">
                            {lead.company_size && (
                              <span>{lead.company_size} team</span>
                            )}
                            {parsed.tms && (
                              <span className="text-slate-400">• TMS: {parsed.tms}</span>
                            )}
                          </div>
                        </div>
                      </td>

                      {/* Column 3: Status & Activity */}
                      <td className="py-4.5 px-5 whitespace-nowrap">
                        <div className="space-y-1">
                          <div>
                            {renderStatusBadge(lead.status)}
                          </div>
                          <div className="text-[11px] text-slate-400 flex items-center gap-1">
                            <Clock className="w-3 h-3 text-slate-300" />
                            <span>{formatRelativeTime(lead.created_at)}</span>
                          </div>
                        </div>
                      </td>

                      {/* Column 4: Review Action Button (Right Most) */}
                      <td className="py-4.5 px-6 text-right whitespace-nowrap">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleOpenDetail(lead);
                          }}
                          className="inline-flex items-center gap-1.5 px-4 py-2 text-xs font-bold text-blue-700 bg-blue-50/90 hover:bg-blue-600 hover:text-white border border-blue-200/80 hover:border-blue-600 rounded-xl transition-all duration-200 shadow-2xs hover:shadow-xs group/btn"
                        >
                          <Eye className="w-3.5 h-3.5" />
                          <span>Review</span>
                          <ChevronRight className="w-3.5 h-3.5 text-blue-400 group-hover/btn:text-white group-hover/btn:translate-x-0.5 transition-all" />
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* ═══ PREMIUM EXECUTIVE REVIEW MODAL ═══ */}
      {detailModalOpen && selectedLead && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-fadeIn">
          <div className="bg-white rounded-3xl shadow-2xl border border-slate-200 w-full max-w-3xl max-h-[92vh] flex flex-col overflow-hidden animate-scaleUp">
            
            {/* Modal Header */}
            <div className="px-6 py-5 border-b border-slate-200 bg-gradient-to-r from-slate-50 via-white to-blue-50/30 flex items-start justify-between">
              <div className="flex items-center gap-4">
                <div className={`w-14 h-14 rounded-2xl bg-gradient-to-br ${getAvatarGradient(selectedLead.full_name)} text-white flex items-center justify-center font-black text-xl shadow-md`}>
                  {selectedLead.full_name ? selectedLead.full_name.charAt(0).toUpperCase() : 'F'}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-lg font-black text-slate-900 leading-tight">
                      {selectedLead.full_name}
                    </h3>
                    {renderStatusBadge(selectedLead.status)}
                  </div>
                  <div className="flex items-center gap-2 text-xs text-slate-500 mt-1">
                    <Building2 className="w-3.5 h-3.5 text-slate-400" />
                    <span className="font-semibold text-slate-800">{selectedLead.company_name}</span>
                    {selectedLead.country && <span>• {selectedLead.country}</span>}
                    <span>• Submitted {formatDate(selectedLead.created_at)}</span>
                  </div>
                </div>
              </div>

              <button
                onClick={() => setDetailModalOpen(false)}
                className="text-slate-400 hover:text-slate-700 p-2 rounded-xl hover:bg-slate-100 transition"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Modal Scrollable Body */}
            <div className="p-6 overflow-y-auto space-y-6">
              
              {feedbackMsg && (
                <div className={`p-3.5 rounded-xl text-xs font-semibold flex items-center gap-2 ${
                  feedbackMsg.startsWith('Error')
                    ? 'bg-rose-50 text-rose-700 border border-rose-200'
                    : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                }`}>
                  <CheckCircle2 className="w-4 h-4 flex-shrink-0" />
                  <span>{feedbackMsg}</span>
                </div>
              )}

              {/* SECTION 1: Direct Contact & Forwarder Profile Matrix */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3 bg-slate-50/80 p-4 rounded-2xl border border-slate-200">
                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Direct Email</span>
                  <div className="flex items-center gap-2">
                    <a
                      href={`mailto:${selectedLead.email}`}
                      className="text-xs font-bold text-blue-600 hover:underline truncate"
                    >
                      {selectedLead.email}
                    </a>
                    <button
                      onClick={() => copyToClipboard(selectedLead.email, 'email')}
                      className="text-slate-400 hover:text-slate-600 p-1"
                      title="Copy Email"
                    >
                      {copiedField === 'email' ? <Check className="w-3 h-3 text-emerald-600" /> : <Copy className="w-3 h-3" />}
                    </button>
                  </div>
                </div>

                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Direct Phone</span>
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-slate-800">
                      {selectedLead.phone || 'Not provided'}
                    </span>
                    {selectedLead.phone && (
                      <button
                        onClick={() => copyToClipboard(selectedLead.phone, 'phone')}
                        className="text-slate-400 hover:text-slate-600 p-1"
                        title="Copy Phone"
                      >
                        {copiedField === 'phone' ? <Check className="w-3 h-3 text-emerald-600" /> : <Copy className="w-3 h-3" />}
                      </button>
                    )}
                  </div>
                </div>

                <div className="space-y-1">
                  <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Operational Scale</span>
                  <span className="text-xs font-bold text-slate-800 block">
                    {selectedLead.company_size ? `${selectedLead.company_size} Emps` : 'Scale unspec.'}
                    {selectedLead.shipment_volume && ` • ${selectedLead.shipment_volume}`}
                  </span>
                </div>
              </div>

              {/* SECTION 2: Forwarder Capabilities Requested */}
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Sparkles className="w-4 h-4 text-blue-600" />
                  <span className="text-xs font-bold text-slate-900 uppercase tracking-wider">
                    Forwarder Platform Modules Requested
                  </span>
                </div>

                <div className="flex flex-wrap gap-2">
                  {selectedLead.services ? (
                    selectedLead.services.split(',').map((srv, idx) => (
                      <span
                        key={idx}
                        className="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl bg-blue-50 text-blue-700 border border-blue-200/80 text-xs font-semibold shadow-2xs"
                      >
                        <Check className="w-3 h-3 text-blue-500" />
                        <span>{srv.trim()}</span>
                      </span>
                    ))
                  ) : (
                    <span className="text-xs text-slate-400 italic">No specific capability tags selected (General Demo)</span>
                  )}
                </div>
              </div>

              {/* SECTION 3: Forwarder Corridor, TMS & Prospect Message */}
              {(() => {
                const meta = parseForwarderMetadata(selectedLead.message);
                return (
                  <div className="space-y-3">
                    {(meta.role || meta.corridor || meta.tms) && (
                      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 p-3.5 rounded-xl bg-slate-50 border border-slate-200/80 text-xs">
                        {meta.role && (
                          <div>
                            <span className="text-[10px] font-bold text-slate-400 uppercase block">Forwarder Role</span>
                            <span className="font-semibold text-slate-800">{meta.role}</span>
                          </div>
                        )}
                        {meta.corridor && (
                          <div>
                            <span className="text-[10px] font-bold text-slate-400 uppercase block">Primary Trade Lane</span>
                            <span className="font-semibold text-slate-800">{meta.corridor}</span>
                          </div>
                        )}
                        {meta.tms && (
                          <div>
                            <span className="text-[10px] font-bold text-slate-400 uppercase block">Current TMS System</span>
                            <span className="font-semibold text-slate-800">{meta.tms}</span>
                          </div>
                        )}
                      </div>
                    )}

                    {meta.notes && (
                      <div className="space-y-1.5">
                        <span className="text-xs font-bold text-slate-900 uppercase tracking-wider block">
                          Prospect Inbound Notes
                        </span>
                        <div className="p-4 rounded-2xl bg-amber-50/60 border border-amber-200/80 text-xs text-slate-700 leading-relaxed font-medium">
                          "{meta.notes}"
                        </div>
                      </div>
                    )}
                  </div>
                );
              })()}

              {/* SECTION 4: Sales Triage & Specialist Assignment */}
              <div className="border-t border-slate-200 pt-5 space-y-4">
                <div className="flex items-center justify-between">
                  <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wider">
                    Lead Qualification & Triage
                  </h4>
                  <span className="text-[11px] text-slate-400">
                    Source: {selectedLead.source || 'Website'} {selectedLead.ip_address && `• IP: ${selectedLead.ip_address}`}
                  </span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {/* Status Dropdown */}
                  <div>
                    <label className="text-xs font-bold text-slate-700 block mb-1.5">
                      Pipeline Stage
                    </label>
                    <select
                      value={editStatus}
                      onChange={(e) => setEditStatus(e.target.value)}
                      className="w-full text-xs font-semibold bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white"
                    >
                      <option value="NEW">NEW (Awaiting review)</option>
                      <option value="CONTACTED">CONTACTED (Follow-up underway)</option>
                      <option value="QUALIFIED">QUALIFIED (Sandbox demo scheduled)</option>
                      <option value="CONVERTED">CONVERTED (Live customer account)</option>
                      <option value="DISMISSED">DISMISSED (Not qualified / Spam)</option>
                    </select>
                  </div>

                  {/* Assigned Specialist */}
                  <div>
                    <label className="text-xs font-bold text-slate-700 block mb-1.5">
                      Assigned Account Executive
                    </label>
                    <input
                      type="text"
                      value={editAssignedTo}
                      onChange={(e) => setEditAssignedTo(e.target.value)}
                      placeholder="e.g. Marcus Vance or Sales Desk"
                      className="w-full text-xs bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white"
                    >
                    </input>
                  </div>
                </div>

                {/* Internal Notes */}
                <div>
                  <label className="text-xs font-bold text-slate-700 block mb-1.5">
                    Internal Account & Call Notes
                  </label>
                  <textarea
                    rows={3}
                    value={editNotes}
                    onChange={(e) => setEditNotes(e.target.value)}
                    placeholder="Log demonstration findings, trade lanes discussed, CargoWise integrations needed, or follow-up date..."
                    className="w-full text-xs bg-slate-50 border border-slate-300 rounded-xl p-3 text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white"
                  />
                </div>
              </div>

            </div>

            {/* Modal Footer */}
            <div className="px-6 py-4 border-t border-slate-200 bg-slate-50/90 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <a
                href={`mailto:${selectedLead.email}?subject=LogisticsHQ Demonstration — ${selectedLead.company_name}&body=Hi ${selectedLead.full_name},%0D%0A%0D%0AThank you for requesting a demonstration of LogisticsHQ for ${selectedLead.company_name}.%0D%0A%0D%0AWe would love to walk you through our autonomous freight forwarding workflows, rate management, and carrier connectivity.`}
                className="inline-flex items-center justify-center gap-2 px-4 py-2 text-xs font-bold text-slate-700 bg-white border border-slate-300 hover:bg-slate-100 rounded-xl transition shadow-2xs"
              >
                <Mail className="w-3.5 h-3.5 text-blue-600" />
                <span>Email {selectedLead.full_name.split(' ')[0]}</span>
              </a>

              <div className="flex items-center justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setDetailModalOpen(false)}
                  className="px-4 py-2 text-xs font-bold text-slate-600 hover:text-slate-900 transition"
                >
                  Close
                </button>

                <button
                  type="button"
                  onClick={handleSaveLead}
                  disabled={updating}
                  className="inline-flex items-center gap-2 px-5 py-2 text-xs font-bold text-white bg-blue-600 hover:bg-blue-700 rounded-xl transition shadow-sm disabled:opacity-50"
                >
                  {updating ? (
                    <>
                      <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                      <span>Saving...</span>
                    </>
                  ) : (
                    <>
                      <Check className="w-3.5 h-3.5" />
                      <span>Save Changes</span>
                    </>
                  )}
                </button>
              </div>
            </div>

          </div>
        </div>
      )}

    </div>
  );
}

export default DemoRequestsPage;
