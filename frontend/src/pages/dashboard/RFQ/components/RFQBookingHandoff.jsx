import React, { useState } from 'react';
import {
  Ship,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  Clock,
  Circle,
  ArrowRight,
  ExternalLink,
  Plus,
  Send,
  Check,
  XCircle,
  FileText,
  Anchor,
  Calendar,
  Layers,
  Info,
  ShieldCheck,
  TrendingUp,
  MapPin,
  ChevronRight,
  Loader2,
  DollarSign,
  Activity,
  FileCheck2,
  PackageCheck
} from 'lucide-react';
import { rfqService } from '../../../../services/rfqService';
import { toast } from 'react-hot-toast';
import { useNavigate } from 'react-router-dom';

export default function RFQBookingHandoff({
  rfq,
  bookingHandoffData,
  quotesData,
  requirements,
  documentsData,
  shipmentHandoffData,
  onSwitchTab,
  onMutationSuccess,
}) {
  const navigate = useNavigate();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [transitioningId, setTransitioningId] = useState(null);

  // 1. Resolve Eligibility & Data
  const eligibility = bookingHandoffData?.eligibility || {
    is_eligible: false,
    missing_prerequisites: ['Evaluating commercial & operational criteria...'],
    commercial_closure_status: 'PENDING',
  };

  const bookings = bookingHandoffData?.bookings || [];
  const activeBooking = bookingHandoffData?.summary?.active_booking || (bookings.length > 0 ? bookings[0] : null);
  const totalBookings = bookingHandoffData?.summary?.total_bookings || bookings.length;

  const approvedQuote = quotesData?.approved_quote || (eligibility.approved_quote_id && quotesData?.quotes?.find(q => q.id === eligibility.approved_quote_id));
  const recommendedQuote = quotesData?.recommended_quote;
  const quotesCount = quotesData?.quotes?.length ?? quotesData?.summary?.total_quotes ?? 0;

  const opReadiness = requirements?.operational_readiness;
  const blockingCount = opReadiness?.blocking_count ?? 0;
  const missingReqCount = opReadiness?.missing_required_count ?? 0;

  const docSummary = documentsData?.summary;
  const missingDocsCount = docSummary?.missing_documents ?? 0;
  const approvedDocsCount = docSummary?.approved_documents ?? 0;
  const requiredDocsCount = docSummary?.required_documents ?? 0;

  const activeShipment = shipmentHandoffData?.summary?.active_shipment || (shipmentHandoffData?.shipments?.length > 0 ? shipmentHandoffData.shipments[0] : null);

  // Form State for Booking Creation
  const [formData, setFormData] = useState({
    booking_number: '',
    carrier_name: eligibility?.approved_carrier || approvedQuote?.carrier_name || '',
    vessel_name: '',
    voyage_number: '',
    origin_port: rfq?.origin || '',
    destination_port: rfq?.destination || '',
    etd: '',
    eta: '',
    cargo_summary: rfq?.items?.map(i => `${i.quantity || 1}x ${i.description || 'Cargo'}`).join(', ') || '',
    special_instructions: '',
  });

  const handleOpenCreateModal = () => {
    setFormData({
      booking_number: `BK-${new Date().toISOString().slice(0, 10).replace(/-/g, '')}-${Math.floor(1000 + Math.random() * 9000)}`,
      carrier_name: eligibility?.approved_carrier || approvedQuote?.carrier_name || '',
      vessel_name: '',
      voyage_number: '',
      origin_port: rfq?.origin || '',
      destination_port: rfq?.destination || '',
      etd: rfq?.target_date ? new Date(rfq.target_date).toISOString().slice(0, 16) : '',
      eta: '',
      cargo_summary: rfq?.items?.map(i => `${i.quantity || 1}x ${i.description || 'Cargo'}`).join(', ') || '',
      special_instructions: '',
    });
    setIsCreateModalOpen(true);
  };

  const handleCreateBooking = async (e) => {
    e.preventDefault();
    if (!formData.carrier_name && !eligibility?.approved_carrier) {
      toast.error('Carrier name is required');
      return;
    }

    try {
      setIsSubmitting(true);
      await rfqService.createRFQBooking(rfq.id, {
        quote_id: eligibility?.approved_quote_id || approvedQuote?.id || null,
        booking_number: formData.booking_number || undefined,
        carrier_name: formData.carrier_name || eligibility?.approved_carrier,
        vessel_name: formData.vessel_name || undefined,
        voyage_number: formData.voyage_number || undefined,
        origin_port: formData.origin_port || rfq?.origin,
        destination_port: formData.destination_port || rfq?.destination,
        etd: formData.etd ? new Date(formData.etd).toISOString() : undefined,
        eta: formData.eta ? new Date(formData.eta).toISOString() : undefined,
        cargo_summary: formData.cargo_summary || undefined,
        special_instructions: formData.special_instructions || undefined,
      });

      toast.success('Carrier booking handoff created successfully!');
      setIsCreateModalOpen(false);
      if (onMutationSuccess) onMutationSuccess();
    } catch (err) {
      console.error('Failed to create booking:', err);
      toast.error(err?.response?.data?.error?.message || err?.message || 'Failed to create booking');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleStatusTransition = async (bookingId, targetStatus) => {
    try {
      setTransitioningId(bookingId);
      await rfqService.updateRFQBookingStatus(rfq.id, bookingId, {
        status: targetStatus,
      });
      toast.success(`Booking status transitioned to ${targetStatus}`);
      if (onMutationSuccess) onMutationSuccess();
    } catch (err) {
      console.error('Failed to update booking status:', err);
      toast.error(err?.response?.data?.error?.message || err?.message || 'Failed to update booking status');
    } finally {
      setTransitioningId(null);
    }
  };

  // Status Stepper Helper
  const getStepState = (currentStatus, stepName) => {
    const order = ['DRAFT', 'REQUESTED', 'CONFIRMED', 'COMPLETED'];
    const currIdx = order.indexOf(currentStatus);
    const stepIdx = order.indexOf(stepName);

    if (currentStatus === 'CANCELLED') return 'cancelled';
    if (currIdx > stepIdx) return 'completed';
    if (currIdx === stepIdx) return 'active';
    return 'pending';
  };

  const getStatusBadgeClass = (status) => {
    switch (status) {
      case 'CONFIRMED':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800';
      case 'REQUESTED':
        return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-800';
      case 'DRAFT':
        return 'bg-slate-100 text-slate-700 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700';
      case 'COMPLETED':
        return 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-950/40 dark:text-purple-300 dark:border-purple-800';
      case 'CANCELLED':
        return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800';
      default:
        return 'bg-slate-100 text-slate-700 border-slate-200';
    }
  };

  // Checklist items
  const isQuoteApproved = Boolean(eligibility.approved_quote_id || approvedQuote);
  const isCarrierSelected = Boolean(eligibility.approved_carrier || approvedQuote?.carrier_name);
  const isTradeReqsComplete = blockingCount === 0;
  const isDocsComplete = missingDocsCount === 0;
  const isEligible = eligibility.is_eligible;

  if (bookingHandoffData === null) {
    return (
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-12 text-center shadow-xs" data-testid="rfq-booking-loading">
        <div className="w-12 h-12 bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-100 dark:border-indigo-800 rounded-2xl flex items-center justify-center mx-auto mb-3.5 text-indigo-600 dark:text-indigo-400">
          <Loader2 className="w-6 h-6 animate-spin text-indigo-600 dark:text-indigo-400" />
        </div>
        <h3 className="text-sm font-bold text-slate-900 dark:text-white mb-1">Loading Carrier Booking Handoff...</h3>
        <p className="text-xs text-slate-500 dark:text-slate-400 max-w-md mx-auto leading-relaxed">
          Evaluating booking eligibility gate, space reservations, and active carrier allocations from the backend.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      
      {/* ── 1. UNIFIED RFQ EXECUTION LIFECYCLE HEADER ── */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm">
        <div className="flex items-center justify-between gap-2 mb-3">
          <div className="flex items-center gap-2">
            <span className="text-[11px] font-extrabold uppercase tracking-wider text-slate-500 dark:text-slate-400">
              Commercial-to-Execution Lifecycle
            </span>
            <span className="text-[10px] px-2 py-0.5 rounded-full font-bold bg-indigo-50 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-300 border border-indigo-200/60 dark:border-indigo-800">
              Booking Handoff Center
            </span>
          </div>
          <button
            onClick={() => navigate('/dashboard/bookings')}
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 transition-colors"
          >
            <span>Open Bookings Workspace</span>
            <ExternalLink className="w-3.5 h-3.5" />
          </button>
        </div>

        {/* 5-Step Lifecycle Breadcrumb */}
        <div className="grid grid-cols-1 sm:grid-cols-5 gap-2 pt-1 border-t border-slate-100 dark:border-slate-800/80">
          {/* Stage 1: Lead */}
          <div
            onClick={() => rfq?.lead_id && navigate(`/dashboard/leads/${rfq.lead_id}`)}
            className="flex items-center justify-between p-2 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200/70 dark:border-slate-700/60 cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <div className="min-w-0">
              <span className="text-[10px] font-bold text-slate-400 uppercase block">1. Inception</span>
              <span className="text-xs font-bold text-slate-800 dark:text-slate-200 truncate block">
                {rfq?.lead_id ? `Lead #${rfq.lead_id}` : 'Direct RFQ'}
              </span>
            </div>
            <CheckCircle2 className="w-4 h-4 text-emerald-500 shrink-0" />
          </div>

          {/* Stage 2: RFQ & Trade Specs */}
          <div
            onClick={() => onSwitchTab('overview')}
            className="flex items-center justify-between p-2 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200/70 dark:border-slate-700/60 cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <div className="min-w-0">
              <span className="text-[10px] font-bold text-slate-400 uppercase block">2. RFQ Spec</span>
              <span className="text-xs font-bold text-slate-800 dark:text-slate-200 truncate block">
                {rfq?.rfq_number || `#${rfq?.id}`}
              </span>
            </div>
            <CheckCircle2 className="w-4 h-4 text-emerald-500 shrink-0" />
          </div>

          {/* Stage 3: Carrier Quote */}
          <div
            onClick={() => onSwitchTab('quotes')}
            className={`flex items-center justify-between p-2 rounded-lg border cursor-pointer transition-colors ${
              isQuoteApproved
                ? 'bg-emerald-50/70 border-emerald-200 dark:bg-emerald-950/30 dark:border-emerald-800'
                : quotesCount > 0
                ? 'bg-amber-50/70 border-amber-200 dark:bg-amber-950/30 dark:border-amber-800'
                : 'bg-slate-50 dark:bg-slate-800/50 border-slate-200/70 dark:border-slate-700/60'
            }`}
          >
            <div className="min-w-0">
              <span className="text-[10px] font-bold text-slate-400 uppercase block">3. Carrier Quote</span>
              <span className="text-xs font-bold text-slate-800 dark:text-slate-200 truncate block">
                {isQuoteApproved ? `${approvedQuote?.carrier_name || eligibility.approved_carrier}` : quotesCount > 0 ? `${quotesCount} Received` : 'Pending'}
              </span>
            </div>
            {isQuoteApproved ? (
              <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
            ) : quotesCount > 0 ? (
              <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
            ) : (
              <Circle className="w-4 h-4 text-slate-300 shrink-0" />
            )}
          </div>

          {/* Stage 4: Booking Handoff (Current Stage) */}
          <div
            className={`flex items-center justify-between p-2 rounded-lg border ring-2 ring-indigo-500/20 ${
              activeBooking?.status === 'CONFIRMED'
                ? 'bg-emerald-50/80 border-emerald-300 dark:bg-emerald-950/40 dark:border-emerald-700'
                : activeBooking
                ? 'bg-blue-50/80 border-blue-300 dark:bg-blue-950/40 dark:border-blue-700'
                : isEligible
                ? 'bg-indigo-50/80 border-indigo-300 dark:bg-indigo-950/40 dark:border-indigo-700'
                : 'bg-amber-50/70 border-amber-200 dark:bg-amber-950/30 dark:border-amber-800'
            }`}
          >
            <div className="min-w-0">
              <span className="text-[10px] font-extrabold text-indigo-600 dark:text-indigo-400 uppercase block">4. Carrier Booking</span>
              <span className="text-xs font-bold text-slate-900 dark:text-slate-100 truncate block">
                {activeBooking ? `${activeBooking.booking_number} (${activeBooking.status})` : isEligible ? 'Ready to Create' : 'Prerequisites Pending'}
              </span>
            </div>
            {activeBooking?.status === 'CONFIRMED' ? (
              <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
            ) : activeBooking ? (
              <Clock className="w-4 h-4 text-blue-600 shrink-0" />
            ) : isEligible ? (
              <Circle className="w-4 h-4 text-indigo-600 fill-indigo-600 shrink-0" />
            ) : (
              <XCircle className="w-4 h-4 text-amber-600 shrink-0" />
            )}
          </div>

          {/* Stage 5: Shipment Execution */}
          <div
            onClick={() => onSwitchTab('shipment')}
            className={`flex items-center justify-between p-2 rounded-lg border cursor-pointer transition-colors ${
              activeShipment
                ? 'bg-emerald-50/70 border-emerald-200 dark:bg-emerald-950/30 dark:border-emerald-800'
                : activeBooking?.status === 'CONFIRMED'
                ? 'bg-teal-50/70 border-teal-200 dark:bg-teal-950/30 dark:border-teal-800'
                : 'bg-slate-50 dark:bg-slate-800/50 border-slate-200/70 dark:border-slate-700/60'
            }`}
          >
            <div className="min-w-0">
              <span className="text-[10px] font-bold text-slate-400 uppercase block">5. Shipment</span>
              <span className="text-xs font-bold text-slate-800 dark:text-slate-200 truncate block">
                {activeShipment ? `Shipment #${activeShipment.id}` : activeBooking?.status === 'CONFIRMED' ? 'Ready for Dispatch' : 'Waiting on Booking'}
              </span>
            </div>
            {activeShipment ? (
              <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
            ) : activeBooking?.status === 'CONFIRMED' ? (
              <Clock className="w-4 h-4 text-teal-600 shrink-0" />
            ) : (
              <Circle className="w-4 h-4 text-slate-300 shrink-0" />
            )}
          </div>
        </div>
      </div>

      {/* ── 2. STATE A: BOOKING BLOCKED STATE ── */}
      {!isEligible && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Left: Operational Readiness Checklist (7 cols) */}
          <div className="lg:col-span-7 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <ShieldCheck className="w-4 h-4 text-slate-500" />
                <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100">
                  Booking Readiness Checklist
                </h3>
              </div>
              <span className="text-xs font-bold px-2 py-0.5 rounded bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800">
                Gate Blocked ({eligibility.missing_prerequisites?.length || 1} criteria pending)
              </span>
            </div>

            <div className="space-y-2.5">
              {/* Check 1: Carrier Quote Approved */}
              <div className="flex items-center justify-between p-3 rounded-lg border bg-slate-50/60 dark:bg-slate-800/40 border-slate-200/80 dark:border-slate-700/60 text-xs">
                <div className="flex items-center gap-2.5">
                  {isQuoteApproved ? (
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                  ) : (
                    <XCircle className="w-4 h-4 text-rose-500 shrink-0" />
                  )}
                  <div>
                    <strong className="text-slate-800 dark:text-slate-200 block">Commercial Quote Approval</strong>
                    <span className="text-slate-500 dark:text-slate-400 text-[11px]">
                      {isQuoteApproved ? `Approved carrier quote: ${approvedQuote?.carrier_name || eligibility.approved_carrier}` : 'No quote approved yet. Pricing & commercial sign-off required.'}
                    </span>
                  </div>
                </div>
                <span className={`px-2 py-0.5 rounded font-bold text-[10px] ${isQuoteApproved ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/60 dark:text-emerald-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-900/60 dark:text-rose-300'}`}>
                  {isQuoteApproved ? 'SATISFIED' : 'PENDING'}
                </span>
              </div>

              {/* Check 2: Selected Carrier */}
              <div className="flex items-center justify-between p-3 rounded-lg border bg-slate-50/60 dark:bg-slate-800/40 border-slate-200/80 dark:border-slate-700/60 text-xs">
                <div className="flex items-center gap-2.5">
                  {isCarrierSelected ? (
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                  ) : (
                    <XCircle className="w-4 h-4 text-rose-500 shrink-0" />
                  )}
                  <div>
                    <strong className="text-slate-800 dark:text-slate-200 block">Carrier Allocation</strong>
                    <span className="text-slate-500 dark:text-slate-400 text-[11px]">
                      {isCarrierSelected ? `Allocated Carrier: ${eligibility.approved_carrier || approvedQuote?.carrier_name}` : 'Carrier not locked. Select an approved quote.'}
                    </span>
                  </div>
                </div>
                <span className={`px-2 py-0.5 rounded font-bold text-[10px] ${isCarrierSelected ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/60 dark:text-emerald-300' : 'bg-rose-100 text-rose-800 dark:bg-rose-900/60 dark:text-rose-300'}`}>
                  {isCarrierSelected ? 'ALLOCATED' : 'MISSING'}
                </span>
              </div>

              {/* Check 3: Trade Requirements */}
              <div className="flex items-center justify-between p-3 rounded-lg border bg-slate-50/60 dark:bg-slate-800/40 border-slate-200/80 dark:border-slate-700/60 text-xs">
                <div className="flex items-center gap-2.5">
                  {isTradeReqsComplete ? (
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                  ) : (
                    <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0" />
                  )}
                  <div>
                    <strong className="text-slate-800 dark:text-slate-200 block">Trade & Cargo Parameters</strong>
                    <span className="text-slate-500 dark:text-slate-400 text-[11px]">
                      {isTradeReqsComplete ? 'Origin, destination, incoterms, and cargo specifications complete.' : `${blockingCount} blocking parameter${blockingCount === 1 ? '' : 's'} missing.`}
                    </span>
                  </div>
                </div>
                <span className={`px-2 py-0.5 rounded font-bold text-[10px] ${isTradeReqsComplete ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/60 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-900/60 dark:text-amber-300'}`}>
                  {isTradeReqsComplete ? 'COMPLETE' : `${blockingCount} BLOCKING`}
                </span>
              </div>

              {/* Check 4: Mandatory Documents */}
              <div className="flex items-center justify-between p-3 rounded-lg border bg-slate-50/60 dark:bg-slate-800/40 border-slate-200/80 dark:border-slate-700/60 text-xs">
                <div className="flex items-center gap-2.5">
                  {isDocsComplete ? (
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
                  ) : (
                    <AlertCircle className="w-4 h-4 text-amber-500 shrink-0" />
                  )}
                  <div>
                    <strong className="text-slate-800 dark:text-slate-200 block">Mandatory Documentation</strong>
                    <span className="text-slate-500 dark:text-slate-400 text-[11px]">
                      {isDocsComplete ? `All required stage documents satisfied (${approvedDocsCount}/${requiredDocsCount}).` : `${missingDocsCount} required document${missingDocsCount === 1 ? '' : 's'} pending review or upload.`}
                    </span>
                  </div>
                </div>
                <span className={`px-2 py-0.5 rounded font-bold text-[10px] ${isDocsComplete ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/60 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-900/60 dark:text-amber-300'}`}>
                  {isDocsComplete ? 'SATISFIED' : 'PENDING'}
                </span>
              </div>
            </div>
          </div>

          {/* Right: Contextual Next Action Card (5 cols) */}
          <div className="lg:col-span-5 bg-gradient-to-br from-amber-50/90 to-orange-50/60 dark:from-amber-950/30 dark:to-orange-950/20 border border-amber-200 dark:border-amber-800/60 rounded-xl p-5 shadow-sm flex flex-col justify-between">
            <div>
              <div className="flex items-center gap-2 text-xs font-bold text-amber-800 dark:text-amber-300 uppercase tracking-wider mb-2">
                <Info className="w-4 h-4" />
                <span>Recommended Operational Action</span>
              </div>
              <h4 className="text-sm font-bold text-slate-900 dark:text-slate-100 mb-2 leading-snug">
                {!isQuoteApproved
                  ? 'Approve a carrier quotation to unlock booking handoff.'
                  : !isTradeReqsComplete
                  ? 'Complete the remaining operational trade requirements.'
                  : 'Review and approve required trade documentation.'}
              </h4>
              <p className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed mb-4">
                {!isQuoteApproved
                  ? 'To prevent operational misalignment, carrier bookings require an officially approved quote with locked buy/sell rates.'
                  : !isTradeReqsComplete
                  ? 'Ensure all required routing ports, Incoterms, and cargo weights are fully specified on the RFQ.'
                  : 'Upload and approve Commercial Invoice or Packing List before initiating carrier booking EDI.'}
              </p>
            </div>

            <div className="space-y-2 pt-3 border-t border-amber-200/60 dark:border-amber-800/50">
              {!isQuoteApproved && (
                <button
                  onClick={() => onSwitchTab('quotes')}
                  className="w-full py-2.5 px-3.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold shadow-sm transition-all flex items-center justify-center gap-2"
                >
                  <span>Review & Approve Carrier Quotes</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
              )}
              {!isTradeReqsComplete && (
                <button
                  onClick={() => onSwitchTab('requirements')}
                  className="w-full py-2.5 px-3.5 rounded-lg bg-amber-600 hover:bg-amber-700 text-white text-xs font-bold shadow-sm transition-all flex items-center justify-center gap-2"
                >
                  <span>Check Trade Requirements</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
              )}
              {!isDocsComplete && (
                <button
                  onClick={() => onSwitchTab('documents')}
                  className="w-full py-2.5 px-3.5 rounded-lg bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 border border-slate-300 dark:border-slate-700 text-xs font-bold hover:bg-slate-50 transition-colors flex items-center justify-center gap-2"
                >
                  <span>Manage RFQ Documents</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ── 3. STATE B: BOOKING ELIGIBLE BUT NO BOOKING CREATED YET ── */}
      {isEligible && bookings.length === 0 && (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-emerald-200 dark:border-emerald-800/80 overflow-hidden shadow-sm">
          {/* Header Banner */}
          <div className="p-5 bg-gradient-to-r from-emerald-50 via-teal-50 to-emerald-50 dark:from-emerald-950/40 dark:via-teal-950/30 dark:to-emerald-950/40 border-b border-emerald-200/80 dark:border-emerald-800/80 flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-emerald-600 text-white shadow-sm">
                <ShieldCheck className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-slate-900 dark:text-slate-100">
                  Ready to Create Carrier Booking
                </h3>
                <p className="text-xs text-emerald-800 dark:text-emerald-300 font-medium">
                  Commercial quote approved with <strong className="underline">{eligibility.approved_carrier || approvedQuote?.carrier_name}</strong>. All trade and documentation gates satisfied.
                </p>
              </div>
            </div>

            <button
              onClick={handleOpenCreateModal}
              className="px-5 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold shadow-md hover:shadow transition-all inline-flex items-center gap-2"
              data-testid="create-booking-btn"
            >
              <Plus className="w-4 h-4" />
              <span>+ Create Carrier Booking</span>
            </button>
          </div>

          {/* Operational Pre-fill Summary Grid */}
          <div className="p-6">
            <h4 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-4">
              Approved Commercial Handoff Parameters
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {/* Carrier & Quote */}
              <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
                <span className="text-[10px] font-bold text-slate-400 uppercase block mb-1">Selected Carrier</span>
                <p className="text-sm font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
                  <span>{eligibility.approved_carrier || approvedQuote?.carrier_name}</span>
                </p>
                <span className="text-xs text-slate-500 dark:text-slate-400 block mt-1">
                  Quote Ref: <strong className="font-mono text-slate-700 dark:text-slate-300">{approvedQuote?.quote_reference || `#${eligibility.approved_quote_id || 'N/A'}`}</strong>
                </span>
              </div>

              {/* Commercial Pricing */}
              <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
                <span className="text-[10px] font-bold text-slate-400 uppercase block mb-1">Approved Rate & Margin</span>
                <p className="text-sm font-bold text-emerald-600 dark:text-emerald-400">
                  {approvedQuote ? `${approvedQuote.currency} ${approvedQuote.sell_price?.toLocaleString()}` : 'Commercially Locked'}
                </p>
                <span className="text-xs text-slate-500 dark:text-slate-400 block mt-1">
                  Buy: {approvedQuote ? `${approvedQuote.currency} ${approvedQuote.buy_price?.toLocaleString()}` : 'N/A'} (Margin: +{approvedQuote?.margin_percentage?.toFixed(1) || 0}%)
                </span>
              </div>

              {/* Route */}
              <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
                <span className="text-[10px] font-bold text-slate-400 uppercase block mb-1">Ports & Route</span>
                <p className="text-xs font-bold text-slate-800 dark:text-slate-200">
                  {rfq?.origin || 'Origin'} → {rfq?.destination || 'Destination'}
                </p>
                <span className="text-xs text-slate-500 dark:text-slate-400 block mt-1">
                  Incoterms: <strong>{rfq?.incoterms || 'FOB'}</strong> · Ready: {rfq?.target_date ? new Date(rfq.target_date).toLocaleDateString() : 'TBD'}
                </span>
              </div>

              {/* Cargo */}
              <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
                <span className="text-[10px] font-bold text-slate-400 uppercase block mb-1">Cargo Summary</span>
                <p className="text-xs font-medium text-slate-800 dark:text-slate-200 truncate" title={rfq?.items?.[0]?.description}>
                  {rfq?.items?.map(i => `${i.quantity}x ${i.description}`).join(', ') || 'Standard FCL Cargo'}
                </p>
                <span className="text-xs text-slate-500 dark:text-slate-400 block mt-1">
                  Ready for booking dispatch
                </span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── 4. STATE C: REAL BOOKINGS EXIST (COMPACT OPERATIONAL CARDS) ── */}
      {bookings.length > 0 && (
        <div className="space-y-6">
          {/* Active / Primary Booking Card */}
          {activeBooking && (
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
              {/* Header */}
              <div className="p-5 border-b border-slate-100 dark:border-slate-800/80 bg-slate-50/70 dark:bg-slate-900/50 flex flex-wrap items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <div className="p-2.5 rounded-xl bg-indigo-600 text-white shadow-sm">
                    <Ship className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2.5">
                      <h3 className="text-base font-bold text-slate-900 dark:text-slate-100 font-mono">
                        {activeBooking.booking_number}
                      </h3>
                      <span className={`px-2.5 py-0.5 rounded-full text-xs font-bold border ${getStatusBadgeClass(activeBooking.status)}`}>
                        {activeBooking.status}
                      </span>
                    </div>
                    <p className="text-xs text-slate-500 dark:text-slate-400 flex items-center gap-2 mt-0.5">
                      <span>Carrier: <strong className="text-slate-700 dark:text-slate-300">{activeBooking.carrier_name}</strong></span>
                      {activeBooking.carrier_scac && (
                        <span className="font-mono px-1.5 py-0.2 rounded bg-slate-200 dark:bg-slate-700 text-[10px] text-slate-700 dark:text-slate-300 font-bold">
                          {activeBooking.carrier_scac}
                        </span>
                      )}
                      {activeBooking.quote_id && (
                        <span>• Linked Quote: <strong className="font-mono text-slate-700 dark:text-slate-300">#{activeBooking.quote_id}</strong></span>
                      )}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <button
                    onClick={() => navigate(`/dashboard/bookings/${activeBooking.id}`)}
                    className="px-3.5 py-2 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs font-semibold transition-colors inline-flex items-center gap-1.5 shadow-sm"
                  >
                    <span>Open in Bookings Workspace</span>
                    <ExternalLink className="w-3.5 h-3.5" />
                  </button>
                  <button
                    onClick={handleOpenCreateModal}
                    className="px-3 py-2 rounded-lg bg-indigo-50 dark:bg-indigo-950/60 hover:bg-indigo-100 dark:hover:bg-indigo-900 text-indigo-700 dark:text-indigo-300 text-xs font-bold transition-colors inline-flex items-center gap-1 border border-indigo-200/60 dark:border-indigo-800"
                    title="Add another booking"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    <span>New Booking</span>
                  </button>
                </div>
              </div>

              {/* 4-Step Booking Progression Stepper */}
              <div className="p-5 border-b border-slate-100 dark:border-slate-800/80 bg-white dark:bg-slate-900">
                <div className="grid grid-cols-4 gap-2 relative">
                  {[
                    { id: 'DRAFT', label: '1. Draft Created', desc: 'Commercial parameters set' },
                    { id: 'REQUESTED', label: '2. Requested', desc: 'Sent to carrier EDI/Portal' },
                    { id: 'CONFIRMED', label: '3. Confirmed', desc: 'Space & Vessel locked' },
                    { id: 'COMPLETED', label: '4. Handed Off', desc: 'Ready for Shipment Execution' },
                  ].map((step, idx) => {
                    const state = getStepState(activeBooking.status, step.id);
                    return (
                      <div key={step.id} className="flex flex-col items-center text-center relative z-10">
                        <div
                          className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-all mb-1.5 ${
                            state === 'completed'
                              ? 'bg-emerald-600 text-white shadow-sm'
                              : state === 'active'
                              ? 'bg-indigo-600 text-white ring-4 ring-indigo-100 dark:ring-indigo-950/60 shadow-sm'
                              : 'bg-slate-100 dark:bg-slate-800 text-slate-400 border border-slate-200 dark:border-slate-700'
                          }`}
                        >
                          {state === 'completed' ? <Check className="w-4 h-4" /> : idx + 1}
                        </div>
                        <span className={`text-xs font-bold ${state === 'active' ? 'text-indigo-600 dark:text-indigo-400' : state === 'completed' ? 'text-emerald-700 dark:text-emerald-300' : 'text-slate-500 dark:text-slate-400'}`}>
                          {step.label}
                        </span>
                        <span className="text-[10px] text-slate-400 dark:text-slate-500 hidden sm:block">
                          {step.desc}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Compact Key Information Grid */}
              <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Routing */}
                <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
                  <div className="flex items-center gap-2 text-xs font-bold text-slate-500 dark:text-slate-400 mb-3">
                    <MapPin className="w-4 h-4 text-indigo-500" />
                    <span>PORTS & ROUTING</span>
                  </div>
                  <div className="space-y-2">
                    <div>
                      <span className="text-[10px] font-semibold text-slate-400 uppercase">Origin Port</span>
                      <p className="text-sm font-bold text-slate-800 dark:text-slate-100">{activeBooking.origin_port}</p>
                    </div>
                    <div className="border-t border-slate-200 dark:border-slate-700 pt-2">
                      <span className="text-[10px] font-semibold text-slate-400 uppercase">Destination Port</span>
                      <p className="text-sm font-bold text-slate-800 dark:text-slate-100">{activeBooking.destination_port}</p>
                    </div>
                  </div>
                </div>

                {/* Vessel & Voyage */}
                <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
                  <div className="flex items-center gap-2 text-xs font-bold text-slate-500 dark:text-slate-400 mb-3">
                    <Anchor className="w-4 h-4 text-blue-500" />
                    <span>VESSEL & SCHEDULE</span>
                  </div>
                  <div className="space-y-2">
                    <div>
                      <span className="text-[10px] font-semibold text-slate-400 uppercase">Vessel & Voyage</span>
                      <p className="text-sm font-bold text-slate-800 dark:text-slate-100">
                        {activeBooking.vessel_name ? `${activeBooking.vessel_name} (${activeBooking.voyage_number || 'N/A'})` : 'Vessel Pending Assignment'}
                      </p>
                    </div>
                    <div className="border-t border-slate-200 dark:border-slate-700 pt-2 flex items-center justify-between text-xs">
                      <div>
                        <span className="text-[10px] font-semibold text-slate-400 uppercase block">ETD</span>
                        <span className="font-semibold text-slate-700 dark:text-slate-300">
                          {activeBooking.etd ? new Date(activeBooking.etd).toLocaleDateString() : 'TBD'}
                        </span>
                      </div>
                      <div className="text-right">
                        <span className="text-[10px] font-semibold text-slate-400 uppercase block">ETA</span>
                        <span className="font-semibold text-slate-700 dark:text-slate-300">
                          {activeBooking.eta ? new Date(activeBooking.eta).toLocaleDateString() : 'TBD'}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Cargo & Special Notes */}
                <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
                  <div className="flex items-center gap-2 text-xs font-bold text-slate-500 dark:text-slate-400 mb-3">
                    <Layers className="w-4 h-4 text-purple-500" />
                    <span>CARGO & INSTRUCTIONS</span>
                  </div>
                  <div className="space-y-2">
                    <div>
                      <span className="text-[10px] font-semibold text-slate-400 uppercase">Cargo Summary</span>
                      <p className="text-xs font-medium text-slate-800 dark:text-slate-200 truncate">
                        {activeBooking.cargo_summary || 'Standard Containerized Cargo'}
                      </p>
                    </div>
                    <div className="border-t border-slate-200 dark:border-slate-700 pt-2">
                      <span className="text-[10px] font-semibold text-slate-400 uppercase">Special Instructions</span>
                      <p className="text-xs text-slate-600 dark:text-slate-400 truncate">
                        {activeBooking.special_instructions || 'Direct bill of lading requested'}
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Action Toolbar */}
              <div className="p-4 bg-slate-50/80 dark:bg-slate-900/80 border-t border-slate-100 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400">
                  <span>Created by: <strong className="text-slate-700 dark:text-slate-300">{activeBooking.created_by || 'Operations Team'}</strong></span>
                  <span>•</span>
                  <span>Created: {new Date(activeBooking.created_at).toLocaleString()}</span>
                </div>

                <div className="flex items-center gap-2">
                  {activeBooking.status === 'DRAFT' && (
                    <button
                      disabled={transitioningId === activeBooking.id}
                      onClick={() => handleStatusTransition(activeBooking.id, 'REQUESTED')}
                      className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold transition-colors inline-flex items-center gap-1.5 shadow-sm disabled:opacity-50"
                    >
                      {transitioningId === activeBooking.id ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Send className="w-3.5 h-3.5" />}
                      <span>Request Carrier Booking</span>
                    </button>
                  )}

                  {activeBooking.status === 'REQUESTED' && (
                    <button
                      disabled={transitioningId === activeBooking.id}
                      onClick={() => handleStatusTransition(activeBooking.id, 'CONFIRMED')}
                      className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold transition-colors inline-flex items-center gap-1.5 shadow-sm disabled:opacity-50"
                    >
                      {transitioningId === activeBooking.id ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
                      <span>Confirm Carrier Booking</span>
                    </button>
                  )}

                  {activeBooking.status === 'CONFIRMED' && (
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-semibold text-emerald-700 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/50 px-3 py-1.5 rounded-lg border border-emerald-200 dark:border-emerald-800 inline-flex items-center gap-1.5">
                        <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                        <span>Booking Confirmed — Ready for Shipment Execution</span>
                      </span>
                      <button
                        onClick={() => onSwitchTab('shipment')}
                        className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold transition-colors inline-flex items-center gap-1.5 shadow-sm"
                      >
                        <span>View Shipment Execution</span>
                        <ArrowRight className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  )}

                  <button
                    onClick={() => onSwitchTab('activity')}
                    className="px-3 py-2 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 text-xs font-semibold hover:bg-slate-50 transition-colors"
                  >
                    View in Activity Timeline
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Historical Bookings Table (if multiple exist) */}
          {bookings.length > 1 && (
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
              <div className="p-4 border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/40">
                <h4 className="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                  All Linked Carrier Bookings ({bookings.length})
                </h4>
              </div>
              <div className="divide-y divide-slate-100 dark:divide-slate-800">
                {bookings.map(b => (
                  <div key={b.id} className="p-3.5 flex flex-wrap items-center justify-between gap-3 text-xs">
                    <div className="flex items-center gap-3">
                      <span className="font-mono font-bold text-slate-900 dark:text-slate-100">{b.booking_number}</span>
                      <span className="text-slate-600 dark:text-slate-400 font-medium">{b.carrier_name}</span>
                      <span className="text-slate-400">{b.origin_port} → {b.destination_port}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold border ${getStatusBadgeClass(b.status)}`}>
                        {b.status}
                      </span>
                      <button
                        onClick={() => navigate(`/dashboard/bookings/${b.id}`)}
                        className="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400 hover:text-slate-600"
                        title="Open booking"
                      >
                        <ExternalLink className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* ── 5. CREATE BOOKING MODAL ── */}
      {isCreateModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 w-full max-w-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            {/* Modal Header */}
            <div className="p-5 border-b border-slate-100 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-800/50 flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-lg bg-indigo-600 text-white">
                  <Ship className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-slate-900 dark:text-slate-100">
                    Initiate Carrier Booking Handoff
                  </h3>
                  <p className="text-xs text-slate-500 dark:text-slate-400">
                    Link RFQ {rfq?.rfq_number || `#${rfq?.id}`} to downstream carrier booking execution
                  </p>
                </div>
              </div>
              <button
                onClick={() => setIsCreateModalOpen(false)}
                className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800"
              >
                ✕
              </button>
            </div>

            {/* Modal Form */}
            <form onSubmit={handleCreateBooking} className="p-6 space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                    Booking Reference Number *
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.booking_number}
                    onChange={(e) => setFormData({ ...formData, booking_number: e.target.value })}
                    className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 font-mono focus:ring-2 focus:ring-indigo-500 outline-none"
                    placeholder="e.g. BK-20260827-001"
                  />
                </div>

                <div>
                  <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                    Carrier Name *
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.carrier_name}
                    onChange={(e) => setFormData({ ...formData, carrier_name: e.target.value })}
                    className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 font-medium focus:ring-2 focus:ring-indigo-500 outline-none"
                    placeholder="e.g. Maersk, Hapag-Lloyd"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                    Origin Port *
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.origin_port}
                    onChange={(e) => setFormData({ ...formData, origin_port: e.target.value })}
                    className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 focus:ring-2 focus:ring-indigo-500 outline-none"
                    placeholder="e.g. INNSA"
                  />
                </div>

                <div>
                  <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                    Destination Port *
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.destination_port}
                    onChange={(e) => setFormData({ ...formData, destination_port: e.target.value })}
                    className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 focus:ring-2 focus:ring-indigo-500 outline-none"
                    placeholder="e.g. DEHAM"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                    Vessel Name (Optional)
                  </label>
                  <input
                    type="text"
                    value={formData.vessel_name}
                    onChange={(e) => setFormData({ ...formData, vessel_name: e.target.value })}
                    className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-indigo-500 outline-none"
                    placeholder="e.g. Hamburg Express"
                  />
                </div>

                <div>
                  <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                    Voyage Number (Optional)
                  </label>
                  <input
                    type="text"
                    value={formData.voyage_number}
                    onChange={(e) => setFormData({ ...formData, voyage_number: e.target.value })}
                    className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-indigo-500 outline-none"
                    placeholder="e.g. HE-042W"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                    Estimated Departure (ETD)
                  </label>
                  <input
                    type="datetime-local"
                    value={formData.etd}
                    onChange={(e) => setFormData({ ...formData, etd: e.target.value })}
                    className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-indigo-500 outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                    Estimated Arrival (ETA)
                  </label>
                  <input
                    type="datetime-local"
                    value={formData.eta}
                    onChange={(e) => setFormData({ ...formData, eta: e.target.value })}
                    className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-indigo-500 outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                  Cargo Summary
                </label>
                <input
                  type="text"
                  value={formData.cargo_summary}
                  onChange={(e) => setFormData({ ...formData, cargo_summary: e.target.value })}
                  className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-indigo-500 outline-none"
                  placeholder="e.g. 2x 40HC Industrial Machinery (18,500 kg)"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
                  Special Instructions / Customs Notes
                </label>
                <textarea
                  rows="2"
                  value={formData.special_instructions}
                  onChange={(e) => setFormData({ ...formData, special_instructions: e.target.value })}
                  className="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-indigo-500 outline-none"
                  placeholder="e.g. Temperature sensitive, Direct bill of lading required..."
                />
              </div>

              {/* Modal Actions */}
              <div className="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsCreateModalOpen(false)}
                  className="px-4 py-2 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-bold hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold shadow-md transition-all inline-flex items-center gap-2 disabled:opacity-50"
                >
                  {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Check className="w-4 h-4" />}
                  <span>Create Booking Record</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
