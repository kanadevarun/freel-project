import React, { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import RFQStatusBadge from './RFQStatusBadge';
import { rfqService } from '../../../../services/rfqService';
import {
  FileText,
  Download,
  MoreVertical,
  ChevronDown,
  Building2,
  MapPin,
  ArrowRight,
  Ship,
  Plane,
  Tag,
  Calendar,
  Sparkles,
  ExternalLink,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Copy,
  Check,
  Share2,
  FolderUp,
  History,
  Shield,
  FileSpreadsheet,
  Plus,
  RefreshCw,
  Printer,
  ChevronRight
} from 'lucide-react';

export default function RFQHeader({ rfq, completeness, requirements, onTabChange, onRefresh }) {
  const navigate = useNavigate();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isUpdatingStage, setIsUpdatingStage] = useState(false);
  const [copiedField, setCopiedField] = useState(null);
  const menuRef = useRef(null);

  const isHealthy = completeness?.health === 'HEALTHY';
  const leadId = rfq?.lead_id;
  const customerId = rfq?.customer_id;

  // Close dropdown on click outside
  useEffect(() => {
    function handleClickOutside(event) {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setIsMenuOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Keyboard navigation: Escape key closes menu
  useEffect(() => {
    function handleKeyDown(event) {
      if (event.key === 'Escape' && isMenuOpen) {
        setIsMenuOpen(false);
      }
    }
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isMenuOpen]);

  const handleCopy = (text, fieldName) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopiedField(fieldName);
    toast.success(`Copied ${fieldName}: ${text}`);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const handleStageUpdate = async (newStage) => {
    if (!rfq?.id || isUpdatingStage) return;
    setIsUpdatingStage(true);
    const toastId = toast.loading(`Updating stage to ${newStage.replace(/_/g, ' ')}...`);
    try {
      await rfqService.updateStage(rfq.id, newStage);
      toast.success(`RFQ stage updated to ${newStage.replace(/_/g, ' ')}`, { id: toastId });
      setIsMenuOpen(false);
      if (onRefresh) {
        onRefresh();
      }
    } catch (err) {
      console.error('Failed to update stage:', err);
      toast.error(err.response?.data?.message || err.message || 'Failed to update RFQ stage', { id: toastId });
    } finally {
      setIsUpdatingStage(false);
    }
  };

  const handlePrintPDF = () => {
    toast.success('Opening print dialog for RFQ Summary...');
    window.print();
  };

  const handleNavigateLead = (e) => {
    e?.preventDefault();
    if (leadId) {
      navigate(`/dashboard/leads?leadId=${leadId}&tab=emails`);
    }
  };

  const handleNavigateCustomer = () => {
    if (customerId) {
      navigate(`/dashboard/customers/${customerId}`);
    } else {
      navigate('/dashboard/customers');
    }
  };

  const handleNavigateAudit = () => {
    navigate(`/dashboard/settings/audit-logs?search=${encodeURIComponent(rfq?.rfq_number || '')}`);
  };

  const createdDate = rfq?.created_at ? new Date(rfq.created_at) : null;
  const createdDateStr = createdDate
    ? createdDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    : 'Aug 27, 2026';
  const createdTimeStr = createdDate
    ? createdDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    : '';

  const updatedDate = rfq?.updated_at ? new Date(rfq.updated_at) : createdDate;
  const updatedDateStr = updatedDate
    ? `${updatedDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })} at ${updatedDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    : 'Aug 27, 2026';

  const targetDateStr = rfq?.target_date
    ? new Date(rfq.target_date).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    : 'Sep 20, 2026';

  // Format Origin & Destination
  const formatPort = (val) => {
    if (!val) return { name: 'Not Specified', code: null };
    const parts = val.split('(');
    if (parts.length > 1) {
      return {
        name: parts[0].trim(),
        code: parts[1].replace(')', '').trim(),
      };
    }
    return { name: val.trim(), code: null };
  };

  const originPort = formatPort(rfq?.origin);
  const destPort = formatPort(rfq?.destination);
  const customerName = rfq?.customer_name || (rfq?.customer_id ? `Customer #${rfq.customer_id}` : 'Inbound Customer');

  // Derive mode
  const originUpper = (rfq?.origin || '').toUpperCase();
  const destUpper = (rfq?.destination || '').toUpperCase();
  const isAir = originUpper.includes('AIRPORT') || destUpper.includes('AIRPORT') || originUpper.startsWith('AIR');
  const ModeIcon = isAir ? Plane : Ship;
  const modeText = isAir ? 'Air Freight' : 'Ocean Freight';

  return (
    <div className="bg-white border-b border-slate-200 shadow-2xs">
      <div className="max-w-[1360px] mx-auto px-8 py-5">
        
        {/* ── Breadcrumb & Global Action Controls ─────────────────────── */}
        <div className="flex flex-wrap items-center justify-between gap-3 pb-3.5 border-b border-slate-100 mb-4">
          <div className="flex items-center gap-2 text-xs font-semibold text-slate-500">
            <button
              onClick={() => navigate('/dashboard/rfqs')}
              className="text-indigo-600 hover:text-indigo-800 font-bold transition-colors"
            >
              RFQs
            </button>
            <span className="text-slate-300">/</span>
            <span className="font-mono text-slate-700 font-bold bg-slate-100 px-2 py-0.5 rounded border border-slate-200 text-[11px]">
              {rfq?.rfq_number || 'RFQ Details'}
            </span>
          </div>

          <div className="flex items-center gap-2.5">
            <button
              type="button"
              onClick={handlePrintPDF}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-white hover:bg-slate-50 text-slate-700 text-xs font-bold rounded-lg border border-slate-200 shadow-2xs transition-all active:scale-98 cursor-pointer"
              title="Print or export RFQ summary"
            >
              <Download className="w-3.5 h-3.5 text-slate-500" />
              <span>Download PDF</span>
            </button>

            {/* ── More Actions Dropdown ── */}
            <div className="relative" ref={menuRef}>
              <button
                type="button"
                id="rfq-more-actions-btn"
                onClick={() => setIsMenuOpen(!isMenuOpen)}
                className={`inline-flex items-center gap-1.5 px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold rounded-lg shadow-xs transition-all active:scale-98 cursor-pointer ${
                  isMenuOpen ? 'ring-2 ring-indigo-400 ring-offset-1 bg-indigo-700' : ''
                }`}
                aria-expanded={isMenuOpen}
                aria-haspopup="true"
              >
                <span>More Actions</span>
                <ChevronDown className={`w-3.5 h-3.5 transition-transform duration-150 ${isMenuOpen ? 'rotate-180' : ''}`} />
              </button>

              {isMenuOpen && (
                <div
                  className="absolute right-0 top-full mt-1.5 w-72 bg-white rounded-xl shadow-xl border border-slate-200 py-1.5 z-50 animate-in fade-in slide-in-from-top-1 duration-150"
                  role="menu"
                  aria-orientation="vertical"
                >
                  {/* Menu Header with RFQ ID & Quick Copy */}
                  <div className="px-3.5 py-2 border-b border-slate-100 flex items-center justify-between bg-slate-50/70 rounded-t-lg">
                    <div className="flex flex-col">
                      <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">RFQ Reference</span>
                      <span className="font-mono text-xs font-bold text-slate-800">{rfq?.rfq_number || 'RFQ'}</span>
                    </div>
                    <button
                      type="button"
                      onClick={() => handleCopy(rfq?.rfq_number, 'RFQ Number')}
                      className="p-1 text-slate-400 hover:text-indigo-600 rounded hover:bg-white transition-colors"
                      title="Copy Reference"
                    >
                      {copiedField === 'RFQ Number' ? <Check className="w-3.5 h-3.5 text-emerald-600" /> : <Copy className="w-3.5 h-3.5" />}
                    </button>
                  </div>

                  {/* Section 1: Quoting & Commercial Actions */}
                  <div className="py-1">
                    <div className="px-3.5 py-1 text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                      Commercial & Quoting
                    </div>
                    
                    <button
                      type="button"
                      className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center gap-2.5 transition-colors cursor-pointer"
                      onClick={() => {
                        setIsMenuOpen(false);
                        if (onTabChange) onTabChange('quotes');
                      }}
                    >
                      <Plus className="w-3.5 h-3.5 text-indigo-500" />
                      <span>Add / Manage Carrier Quotes</span>
                    </button>

                    <button
                      type="button"
                      className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center gap-2.5 transition-colors cursor-pointer"
                      onClick={() => {
                        setIsMenuOpen(false);
                        if (onTabChange) onTabChange('requirements');
                      }}
                    >
                      <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                      <span>Check Operational Readiness</span>
                    </button>

                    <button
                      type="button"
                      className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center gap-2.5 transition-colors cursor-pointer"
                      onClick={() => {
                        setIsMenuOpen(false);
                        if (onTabChange) onTabChange('documents');
                      }}
                    >
                      <FolderUp className="w-3.5 h-3.5 text-blue-500" />
                      <span>Attach / Review Documents</span>
                    </button>
                  </div>

                  <div className="h-px bg-slate-100 my-1" />

                  {/* Section 2: Related Entities */}
                  <div className="py-1">
                    <div className="px-3.5 py-1 text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                      Navigation & Relations
                    </div>

                    <button
                      type="button"
                      className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center justify-between transition-colors cursor-pointer"
                      onClick={() => {
                        setIsMenuOpen(false);
                        handleNavigateCustomer();
                      }}
                    >
                      <span className="flex items-center gap-2.5">
                        <Building2 className="w-3.5 h-3.5 text-slate-400" />
                        <span>View Customer Profile</span>
                      </span>
                      <ExternalLink className="w-3 h-3 text-slate-400" />
                    </button>

                    {leadId && (
                      <button
                        type="button"
                        className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center justify-between transition-colors cursor-pointer"
                        onClick={() => {
                          setIsMenuOpen(false);
                          handleNavigateLead();
                        }}
                      >
                        <span className="flex items-center gap-2.5">
                          <Sparkles className="w-3.5 h-3.5 text-purple-500" />
                          <span>View Inbound Lead #{leadId}</span>
                        </span>
                        <ExternalLink className="w-3 h-3 text-slate-400" />
                      </button>
                    )}

                    <button
                      type="button"
                      className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center gap-2.5 transition-colors cursor-pointer"
                      onClick={() => {
                        setIsMenuOpen(false);
                        if (onTabChange) onTabChange('activity');
                      }}
                    >
                      <History className="w-3.5 h-3.5 text-slate-400" />
                      <span>View Activity History</span>
                    </button>

                    <button
                      type="button"
                      className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center justify-between transition-colors cursor-pointer"
                      onClick={() => {
                        setIsMenuOpen(false);
                        handleNavigateAudit();
                      }}
                    >
                      <span className="flex items-center gap-2.5">
                        <Shield className="w-3.5 h-3.5 text-blue-500" />
                        <span>Universal Audit Trail</span>
                      </span>
                      <ExternalLink className="w-3 h-3 text-slate-400" />
                    </button>
                  </div>

                  <div className="h-px bg-slate-100 my-1" />

                  {/* Section 3: Stage Management */}
                  <div className="py-1">
                    <div className="px-3.5 py-1 text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                      Stage Workflow
                    </div>

                    <div className="grid grid-cols-2 gap-1 px-3 py-1">
                      {['DRAFT', 'IN_REVIEW', 'QUOTE_GENERATED', 'CLOSED'].map((stageKey) => {
                        const isCurrent = (rfq?.stage || '').toUpperCase() === stageKey;
                        const label = stageKey === 'QUOTE_GENERATED' ? 'Quoted' : stageKey.replace(/_/g, ' ');
                        return (
                          <button
                            key={stageKey}
                            type="button"
                            disabled={isCurrent || isUpdatingStage}
                            onClick={() => handleStageUpdate(stageKey)}
                            className={`px-2 py-1 text-[11px] font-bold rounded capitalize border transition-all text-center ${
                              isCurrent
                                ? 'bg-indigo-50 text-indigo-700 border-indigo-300 cursor-default'
                                : 'bg-white hover:bg-slate-50 text-slate-600 border-slate-200 cursor-pointer'
                            }`}
                          >
                            {label}
                          </button>
                        );
                      })}
                    </div>
                  </div>

                  <div className="h-px bg-slate-100 my-1" />

                  {/* Section 4: Utilities */}
                  <div className="py-1">
                    <button
                      type="button"
                      className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center justify-between transition-colors cursor-pointer"
                      onClick={() => handleCopy(window.location.href, 'RFQ Link')}
                    >
                      <span className="flex items-center gap-2.5">
                        <Share2 className="w-3.5 h-3.5 text-slate-400" />
                        <span>Copy Shareable Link</span>
                      </span>
                      {copiedField === 'RFQ Link' ? <Check className="w-3.5 h-3.5 text-emerald-600" /> : null}
                    </button>

                    <button
                      type="button"
                      className="w-full px-3.5 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-indigo-600 flex items-center gap-2.5 transition-colors cursor-pointer"
                      onClick={() => {
                        setIsMenuOpen(false);
                        handlePrintPDF();
                      }}
                    >
                      <Printer className="w-3.5 h-3.5 text-slate-400" />
                      <span>Print Summary</span>
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* ── Main RFQ Identity & Structured Metadata Layout ────────── */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          
          {/* Left Column: RFQ Identity & Route Metadata (7 cols) */}
          <div className="lg:col-span-7 space-y-2.5">
            
            {/* Title Row with RFQ Pill & Status */}
            <div className="flex items-center gap-2.5 flex-wrap">
              <span className="font-mono text-base font-extrabold text-indigo-700 bg-indigo-50 border border-indigo-200 px-2.5 py-0.5 rounded-lg tracking-tight">
                {rfq?.rfq_number}
              </span>
              <RFQStatusBadge label={completeness?.statusLabel} color={completeness?.statusColor} size="small" />
            </div>

            {/* Customer Name */}
            <div className="flex items-center gap-2">
              <Building2 className="w-4 h-4 text-slate-400 flex-shrink-0" />
              <h1 className="text-xl font-extrabold text-slate-900 tracking-tight truncate">
                {customerName}
              </h1>
            </div>

            {/* Route Journey Bar */}
            <div className="flex items-center gap-2 flex-wrap text-xs text-slate-700 font-bold bg-slate-50 border border-slate-200/80 rounded-lg px-3 py-1.5 w-fit">
              <MapPin className="w-3.5 h-3.5 text-slate-400" />
              <span className="text-slate-900">{originPort.name}</span>
              {originPort.code && (
                <span className="text-[10px] font-bold text-slate-500 bg-white px-1.5 py-0.2 rounded border border-slate-200">
                  {originPort.code}
                </span>
              )}
              
              <ArrowRight className="w-3.5 h-3.5 text-slate-400 mx-0.5" />

              <span className="text-slate-900">{destPort.name}</span>
              {destPort.code && (
                <span className="text-[10px] font-bold text-slate-500 bg-white px-1.5 py-0.2 rounded border border-slate-200">
                  {destPort.code}
                </span>
              )}
            </div>

            {/* Trade Attributes & Target Date */}
            <div className="flex items-center gap-2.5 flex-wrap pt-0.5 text-xs">
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-blue-50 text-blue-700 font-bold rounded-md border border-blue-200 text-[11px]">
                <ModeIcon className="w-3 h-3 text-blue-600" />
                <span>{modeText}</span>
              </span>

              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-amber-50 text-amber-700 font-extrabold rounded-md border border-amber-200 text-[11px]">
                <Tag className="w-3 h-3 text-amber-600" />
                <span>{rfq?.incoterms || 'FOB'}</span>
              </span>

              <span className="inline-flex items-center gap-1.5 text-slate-500 font-semibold text-[11.5px]">
                <Calendar className="w-3.5 h-3.5 text-slate-400" />
                <span>Target Departure: <strong className="text-slate-800">{targetDateStr}</strong></span>
              </span>
            </div>
          </div>

          {/* Right Column: Structured Operational Metadata Card (5 cols) */}
          <div className="lg:col-span-5 bg-slate-50/80 border border-slate-200 rounded-xl p-4 shadow-2xs space-y-3">
            
            <div className="grid grid-cols-2 gap-3 text-xs pb-3 border-b border-slate-200/70">
              <div>
                <span className="text-[10.5px] font-bold text-slate-400 uppercase tracking-wider block">Created</span>
                <span className="font-bold text-slate-800 text-xs block mt-0.5">{createdDateStr}</span>
                {createdTimeStr && <span className="text-[10.5px] text-slate-400 block">{createdTimeStr}</span>}
              </div>

              <div>
                <span className="text-[10.5px] font-bold text-slate-400 uppercase tracking-wider block">Last Updated</span>
                <span className="font-bold text-slate-800 text-xs block mt-0.5">{updatedDateStr}</span>
              </div>
            </div>

            {/* Health & Source Line */}
            <div className="grid grid-cols-2 gap-3 text-xs items-center">
              <div>
                <span className="text-[10.5px] font-bold text-slate-400 uppercase tracking-wider block mb-1">RFQ Health</span>
                <div className="flex items-center gap-1.5">
                  {isHealthy ? (
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-emerald-50 text-emerald-700 text-[11px] font-bold rounded-md border border-emerald-200">
                      <CheckCircle2 className="w-3 h-3 text-emerald-600" /> Healthy
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-amber-50 text-amber-700 text-[11px] font-bold rounded-md border border-amber-200">
                      <AlertTriangle className="w-3 h-3 text-amber-600" /> Action Required
                    </span>
                  )}
                </div>
              </div>

              <div>
                <span className="text-[10.5px] font-bold text-slate-400 uppercase tracking-wider block mb-1">Source Lineage</span>
                {leadId ? (
                  <button
                    onClick={handleNavigateLead}
                    className="inline-flex items-center gap-1 text-xs font-bold text-indigo-600 hover:text-indigo-800 hover:underline"
                  >
                    <span>Lead #{leadId}</span>
                    <ExternalLink className="w-3 h-3" />
                  </button>
                ) : (
                  <span className="text-slate-500 font-medium text-xs">Direct Manual Entry</span>
                )}
              </div>
            </div>

            {/* AI Summary Tag */}
            {leadId && (
              <div className="pt-2 border-t border-slate-200/70 flex items-center justify-between">
                <span className="inline-flex items-center gap-1 text-[11px] font-bold text-purple-700 bg-purple-50 px-2 py-0.5 rounded-md border border-purple-200">
                  <Sparkles className="w-3 h-3 text-purple-600" /> AI Parsed Request
                </span>
                <span className="text-[10.5px] text-slate-400 font-medium">Inbound Email Connected</span>
              </div>
            )}

          </div>

        </div>

      </div>
    </div>
  );
}
