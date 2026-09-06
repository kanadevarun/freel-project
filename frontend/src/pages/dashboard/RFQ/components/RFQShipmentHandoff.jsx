import React from 'react';
import {
  Navigation,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  Clock,
  Circle,
  ArrowRight,
  ExternalLink,
  Ship,
  Anchor,
  Box,
  MapPin,
  Calendar,
  Layers,
  ChevronRight,
  Info,
  ShieldCheck,
  Package,
  Activity,
  Check
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export default function RFQShipmentHandoff({
  rfq,
  shipmentHandoffData,
  bookingHandoffData,
  quotesData,
  requirements,
  documentsData,
  onSwitchTab,
}) {
  const navigate = useNavigate();

  // Resolve Real Data
  const sourceBooking = shipmentHandoffData?.source_booking || bookingHandoffData?.summary?.active_booking || (bookingHandoffData?.bookings?.length > 0 ? bookingHandoffData.bookings[0] : null);
  const shipments = shipmentHandoffData?.shipments || [];
  const activeShipment = shipmentHandoffData?.summary?.active_shipment || (shipments.length > 0 ? shipments[0] : null);
  const totalShipments = shipmentHandoffData?.summary?.total_shipments || shipments.length;

  const eligibility = bookingHandoffData?.eligibility;
  const approvedQuote = quotesData?.approved_quote || (eligibility?.approved_quote_id && quotesData?.quotes?.find(q => q.id === eligibility.approved_quote_id));
  const quotesCount = quotesData?.quotes?.length ?? quotesData?.summary?.total_quotes ?? 0;

  const isQuoteApproved = Boolean(eligibility?.approved_quote_id || approvedQuote);
  const isBookingCreated = Boolean(sourceBooking);
  const isBookingConfirmed = sourceBooking?.status === 'CONFIRMED' || sourceBooking?.status === 'COMPLETED';

  // Status Badge Helper
  const getShipmentStatusBadgeClass = (status) => {
    switch (status) {
      case 'DELIVERED':
      case 'COMPLETED':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800';
      case 'IN_TRANSIT':
      case 'DEPARTED':
        return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-800';
      case 'ARRIVED':
        return 'bg-teal-50 text-teal-700 border-teal-200 dark:bg-teal-950/40 dark:text-teal-300 dark:border-teal-800';
      case 'BOOKED':
        return 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-950/40 dark:text-indigo-300 dark:border-indigo-800';
      case 'EXCEPTION':
      case 'DELAYED':
        return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800';
      default:
        return 'bg-slate-100 text-slate-700 border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700';
    }
  };

  // Milestone Progression Helper
  const getMilestoneState = (currentStatus, stepName) => {
    const order = ['BOOKED', 'DEPARTED', 'IN_TRANSIT', 'ARRIVED', 'DELIVERED'];
    const currIdx = order.indexOf(currentStatus);
    const stepIdx = order.indexOf(stepName);

    if (currIdx > stepIdx) return 'completed';
    if (currIdx === stepIdx) return 'active';
    return 'pending';
  };

  if (shipmentHandoffData === null) {
    return (
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-12 text-center shadow-xs" data-testid="rfq-shipment-loading">
        <div className="w-12 h-12 bg-teal-50 dark:bg-teal-950/40 border border-teal-100 dark:border-teal-800 rounded-2xl flex items-center justify-center mx-auto mb-3.5 text-teal-600 dark:text-teal-400">
          <Loader2 className="w-6 h-6 animate-spin text-teal-600 dark:text-teal-400" />
        </div>
        <h3 className="text-sm font-bold text-slate-900 dark:text-white mb-1">Loading Shipment Execution Handoff...</h3>
        <p className="text-xs text-slate-500 dark:text-slate-400 max-w-md mx-auto leading-relaxed">
          Connecting to execution tracking engine, container manifest, and vessel milestones from the backend.
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
            <span className="text-[10px] px-2 py-0.5 rounded-full font-bold bg-teal-50 dark:bg-teal-950/60 text-teal-700 dark:text-teal-300 border border-teal-200/60 dark:border-teal-800">
              Live Shipment Execution
            </span>
          </div>
          <button
            onClick={() => navigate('/dashboard/shipments')}
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-teal-600 dark:text-teal-400 hover:text-teal-800 dark:hover:text-teal-300 transition-colors"
          >
            <span>Open Shipments Workspace</span>
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

          {/* Stage 2: RFQ */}
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
                {isQuoteApproved ? `${approvedQuote?.carrier_name || eligibility?.approved_carrier}` : quotesCount > 0 ? `${quotesCount} Received` : 'Pending'}
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

          {/* Stage 4: Booking */}
          <div
            onClick={() => onSwitchTab('booking')}
            className={`flex items-center justify-between p-2 rounded-lg border cursor-pointer transition-colors ${
              isBookingConfirmed
                ? 'bg-emerald-50/70 border-emerald-200 dark:bg-emerald-950/30 dark:border-emerald-800'
                : isBookingCreated
                ? 'bg-blue-50/70 border-blue-200 dark:bg-blue-950/30 dark:border-blue-800'
                : 'bg-slate-50 dark:bg-slate-800/50 border-slate-200/70 dark:border-slate-700/60'
            }`}
          >
            <div className="min-w-0">
              <span className="text-[10px] font-bold text-slate-400 uppercase block">4. Carrier Booking</span>
              <span className="text-xs font-bold text-slate-800 dark:text-slate-200 truncate block">
                {sourceBooking ? `${sourceBooking.booking_number} (${sourceBooking.status})` : 'Not Initiated'}
              </span>
            </div>
            {isBookingConfirmed ? (
              <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
            ) : isBookingCreated ? (
              <Clock className="w-4 h-4 text-blue-600 shrink-0" />
            ) : (
              <Circle className="w-4 h-4 text-slate-300 shrink-0" />
            )}
          </div>

          {/* Stage 5: Shipment Execution (Current Stage) */}
          <div
            className={`flex items-center justify-between p-2 rounded-lg border ring-2 ring-teal-500/20 ${
              activeShipment
                ? 'bg-emerald-50/80 border-emerald-300 dark:bg-emerald-950/40 dark:border-emerald-700'
                : isBookingConfirmed
                ? 'bg-teal-50/80 border-teal-300 dark:bg-teal-950/40 dark:border-teal-700'
                : 'bg-amber-50/70 border-amber-200 dark:bg-amber-950/30 dark:border-amber-800'
            }`}
          >
            <div className="min-w-0">
              <span className="text-[10px] font-extrabold text-teal-600 dark:text-teal-400 uppercase block">5. Live Shipment</span>
              <span className="text-xs font-bold text-slate-900 dark:text-slate-100 truncate block">
                {activeShipment ? `Shipment #${activeShipment.id} (${activeShipment.status})` : isBookingConfirmed ? 'Ready for Dispatch' : 'Waiting on Booking'}
              </span>
            </div>
            {activeShipment ? (
              <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0" />
            ) : isBookingConfirmed ? (
              <Clock className="w-4 h-4 text-teal-600 shrink-0" />
            ) : (
              <Circle className="w-4 h-4 text-slate-300 shrink-0" />
            )}
          </div>
        </div>
      </div>

      {/* ── 2. STATE A: NO SHIPMENT CREATED YET (DEPENDENCY & HANDOFF VIEW) ── */}
      {!activeShipment && (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
          {/* Header Banner */}
          <div className="p-5 border-b border-slate-100 dark:border-slate-800/80 bg-slate-50/70 dark:bg-slate-900/50 flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className={`p-2.5 rounded-xl text-white shadow-sm ${isBookingConfirmed ? 'bg-teal-600' : 'bg-slate-500'}`}>
                <Navigation className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-slate-900 dark:text-slate-100">
                  Shipment Execution Dependency Flow
                </h3>
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  {isBookingConfirmed
                    ? 'Carrier booking is confirmed. Vessel allocation secured — ready for container dispatch and tracking in Shipments.'
                    : 'Shipment execution requires an active, confirmed carrier booking before dispatch.'}
                </p>
              </div>
            </div>

            {/* Dynamic CTA */}
            <div>
              {!isQuoteApproved && (
                <button
                  onClick={() => onSwitchTab('quotes')}
                  className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold transition-all inline-flex items-center gap-2 shadow-sm"
                >
                  <span>Review Carrier Quotes</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
              )}
              {isQuoteApproved && !isBookingCreated && (
                <button
                  onClick={() => onSwitchTab('booking')}
                  className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold transition-all inline-flex items-center gap-2 shadow-sm"
                >
                  <span>Initiate Booking Handoff</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
              )}
              {isBookingCreated && !isBookingConfirmed && (
                <button
                  onClick={() => onSwitchTab('booking')}
                  className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold transition-all inline-flex items-center gap-2 shadow-sm"
                >
                  <span>Confirm Booking in Booking Tab</span>
                  <ArrowRight className="w-4 h-4" />
                </button>
              )}
              {isBookingConfirmed && (
                <button
                  onClick={() => navigate(`/dashboard/shipments`)}
                  className="px-5 py-2.5 rounded-lg bg-teal-600 hover:bg-teal-700 text-white text-xs font-bold shadow-md hover:shadow transition-all inline-flex items-center gap-2"
                >
                  <span>Open Shipments Workspace</span>
                  <ExternalLink className="w-4 h-4" />
                </button>
              )}
            </div>
          </div>

          {/* Upstream Dependency Hierarchy Cards */}
          <div className="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* Stage 1: Commercial Quote */}
            <div className={`p-4 rounded-xl border ${isQuoteApproved ? 'bg-emerald-50/50 border-emerald-200 dark:bg-emerald-950/20 dark:border-emerald-800' : 'bg-slate-50 border-slate-200 dark:bg-slate-800/40 dark:border-slate-700'}`}>
              <div className="flex items-center justify-between mb-2">
                <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Step 1 • Commercial</span>
                <span className={`text-[10px] font-bold px-2 py-0.5 rounded ${isQuoteApproved ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/60 dark:text-emerald-300' : 'bg-amber-100 text-amber-800 dark:bg-amber-900/60 dark:text-amber-300'}`}>
                  {isQuoteApproved ? 'APPROVED' : 'PENDING'}
                </span>
              </div>
              <h4 className="text-xs font-bold text-slate-900 dark:text-slate-100 mb-1">
                {isQuoteApproved ? approvedQuote?.carrier_name || eligibility?.approved_carrier : 'Quote Approval Pending'}
              </h4>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                {isQuoteApproved
                  ? `Locked Sell: ${approvedQuote?.currency || 'USD'} ${approvedQuote?.sell_price?.toLocaleString() || ''} (${approvedQuote?.margin_percentage?.toFixed(1) || 0}% margin)`
                  : 'Customer quotation and buy/sell rate locking required.'}
              </p>
            </div>

            {/* Stage 2: Carrier Booking */}
            <div className={`p-4 rounded-xl border ${isBookingConfirmed ? 'bg-emerald-50/50 border-emerald-200 dark:bg-emerald-950/20 dark:border-emerald-800' : isBookingCreated ? 'bg-blue-50/50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-800' : 'bg-slate-50 border-slate-200 dark:bg-slate-800/40 dark:border-slate-700'}`}>
              <div className="flex items-center justify-between mb-2">
                <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Step 2 • Booking</span>
                <span className={`text-[10px] font-bold px-2 py-0.5 rounded ${isBookingConfirmed ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/60 dark:text-emerald-300' : isBookingCreated ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/60 dark:text-blue-300' : 'bg-slate-200 text-slate-700 dark:bg-slate-700 dark:text-slate-300'}`}>
                  {sourceBooking ? sourceBooking.status : 'NOT INITIATED'}
                </span>
              </div>
              <h4 className="text-xs font-bold text-slate-900 dark:text-slate-100 mb-1 font-mono">
                {sourceBooking ? sourceBooking.booking_number : 'Booking Not Created'}
              </h4>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                {sourceBooking
                  ? `Carrier: ${sourceBooking.carrier_name} • Route: ${sourceBooking.origin_port} → ${sourceBooking.destination_port}`
                  : 'Carrier EDI booking request must be confirmed.'}
              </p>
            </div>

            {/* Stage 3: Shipment Execution */}
            <div className={`p-4 rounded-xl border ${isBookingConfirmed ? 'bg-teal-50/70 border-teal-200 dark:bg-teal-950/30 dark:border-teal-800' : 'bg-slate-50 border-slate-200 dark:bg-slate-800/40 dark:border-slate-700'}`}>
              <div className="flex items-center justify-between mb-2">
                <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Step 3 • Execution</span>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded bg-slate-200 text-slate-700 dark:bg-slate-700 dark:text-slate-300">
                  {isBookingConfirmed ? 'READY TO DISPATCH' : 'WAITING'}
                </span>
              </div>
              <h4 className="text-xs font-bold text-slate-900 dark:text-slate-100 mb-1">
                {isBookingConfirmed ? 'Ready for Container Dispatch' : 'Awaiting Booking Confirmation'}
              </h4>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                {isBookingConfirmed
                  ? 'Milestone tracking, bill of lading, and container events can now be initiated.'
                  : 'Execution record is created once booking status is confirmed.'}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* ── 3. STATE B: REAL SHIPMENT RECORD LINKED ── */}
      {activeShipment && (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
          {/* Header */}
          <div className="p-5 border-b border-slate-100 dark:border-slate-800/80 bg-slate-50/70 dark:bg-slate-900/50 flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-teal-600 text-white shadow-sm">
                <Navigation className="w-5 h-5" />
              </div>
              <div>
                <div className="flex items-center gap-2.5">
                  <h3 className="text-base font-bold text-slate-900 dark:text-slate-100">
                    Shipment #{activeShipment.id}
                  </h3>
                  <span className={`px-2.5 py-0.5 rounded-full text-xs font-bold border ${getShipmentStatusBadgeClass(activeShipment.status)}`}>
                    {activeShipment.status}
                  </span>
                  {activeShipment.closure_status && activeShipment.closure_status !== 'ACTIVE' && (
                    <span className={`px-2.5 py-0.5 rounded-full text-xs font-bold border ${
                      activeShipment.closure_status === 'CLOSED' ? 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800' : 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-950/40 dark:text-indigo-300 dark:border-indigo-800'
                    }`}>
                      {activeShipment.closure_status === 'CLOSED' ? 'CLOSED' : 'READY FOR CLOSURE'}
                    </span>
                  )}
                  {activeShipment.active_exceptions_count > 0 && (
                    <span className="px-2 py-0.5 rounded bg-red-100 dark:bg-red-950 text-red-700 dark:text-red-300 text-xs font-bold border border-red-200 dark:border-red-800 flex items-center gap-1" title={`${activeShipment.active_exceptions_count} active exception(s)`}>
                      <AlertTriangle className="w-3.5 h-3.5 text-red-600" /> {activeShipment.active_exceptions_count} Active Exceptions
                    </span>
                  )}
                </div>
                <p className="text-xs text-slate-500 dark:text-slate-400 flex items-center gap-2 mt-0.5">
                  <span>Carrier: <strong className="text-slate-700 dark:text-slate-300">{activeShipment.carrier_name || activeShipment.carrier_scac}</strong></span>
                  {activeShipment.carrier_scac && (
                    <span className="font-mono px-1.5 py-0.2 rounded bg-slate-200 dark:bg-slate-700 text-[10px] text-slate-700 dark:text-slate-300 font-bold">
                      {activeShipment.carrier_scac}
                    </span>
                  )}
                  {activeShipment.booking_number && (
                    <span>• Source Booking: <strong className="font-mono text-slate-700 dark:text-slate-300">{activeShipment.booking_number}</strong></span>
                  )}
                </p>
              </div>
            </div>

            <button
              onClick={() => navigate(`/dashboard/shipments`)}
              className="px-4 py-2 rounded-lg bg-teal-600 hover:bg-teal-700 text-white text-xs font-bold transition-colors inline-flex items-center gap-1.5 shadow-sm"
            >
              <span>Open Full Shipment Workspace</span>
              <ExternalLink className="w-3.5 h-3.5" />
            </button>
          </div>

          {/* 5-Step Milestone Progress Stepper */}
          <div className="p-5 border-b border-slate-100 dark:border-slate-800/80 bg-white dark:bg-slate-900">
            <div className="grid grid-cols-5 gap-2 relative">
              {[
                { id: 'BOOKED', label: '1. Booked', desc: 'Space Confirmed' },
                { id: 'DEPARTED', label: '2. Departed', desc: 'Vessel Dispatched' },
                { id: 'IN_TRANSIT', label: '3. In Transit', desc: 'Ocean Voyage' },
                { id: 'ARRIVED', label: '4. Arrived', desc: 'Port Discharge' },
                { id: 'DELIVERED', label: '5. Delivered', desc: 'Gate Out / Cleared' },
              ].map((step, idx) => {
                const state = getMilestoneState(activeShipment.status, step.id);
                return (
                  <div key={step.id} className="flex flex-col items-center text-center relative z-10">
                    <div
                      className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-all mb-1.5 ${
                        state === 'completed'
                          ? 'bg-emerald-600 text-white shadow-sm'
                          : state === 'active'
                          ? 'bg-teal-600 text-white ring-4 ring-teal-100 dark:ring-teal-950/60 shadow-sm'
                          : 'bg-slate-100 dark:bg-slate-800 text-slate-400 border border-slate-200 dark:border-slate-700'
                      }`}
                    >
                      {state === 'completed' ? <Check className="w-4 h-4" /> : idx + 1}
                    </div>
                    <span className={`text-xs font-bold ${state === 'active' ? 'text-teal-600 dark:text-teal-400' : state === 'completed' ? 'text-emerald-700 dark:text-emerald-300' : 'text-slate-500 dark:text-slate-400'}`}>
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
                <MapPin className="w-4 h-4 text-teal-500" />
                <span>ROUTING & PORTS</span>
              </div>
              <div className="space-y-2">
                <div>
                  <span className="text-[10px] font-semibold text-slate-400 uppercase">Origin Port</span>
                  <p className="text-sm font-bold text-slate-800 dark:text-slate-100">{activeShipment.origin_port}</p>
                </div>
                <div className="border-t border-slate-200 dark:border-slate-700 pt-2">
                  <span className="text-[10px] font-semibold text-slate-400 uppercase">Destination Port</span>
                  <p className="text-sm font-bold text-slate-800 dark:text-slate-100">{activeShipment.destination_port}</p>
                </div>
              </div>
            </div>

            {/* Schedule & Vessel */}
            <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
              <div className="flex items-center gap-2 text-xs font-bold text-slate-500 dark:text-slate-400 mb-3">
                <Anchor className="w-4 h-4 text-blue-500" />
                <span>VESSEL & TIMING</span>
              </div>
              <div className="space-y-2">
                <div>
                  <span className="text-[10px] font-semibold text-slate-400 uppercase">Vessel / Voyage</span>
                  <p className="text-sm font-bold text-slate-800 dark:text-slate-100">
                    {activeShipment.vessel_name ? `${activeShipment.vessel_name} (${activeShipment.voyage_number || 'N/A'})` : 'Vessel In Transit'}
                  </p>
                </div>
                <div className="border-t border-slate-200 dark:border-slate-700 pt-2 flex items-center justify-between text-xs">
                  <div>
                    <span className="text-[10px] font-semibold text-slate-400 uppercase block">ETD</span>
                    <span className="font-semibold text-slate-700 dark:text-slate-300">
                      {activeShipment.etd ? new Date(activeShipment.etd).toLocaleDateString() : 'TBD'}
                    </span>
                  </div>
                  <div className="text-right">
                    <span className="text-[10px] font-semibold text-slate-400 uppercase block">ETA</span>
                    <span className="font-semibold text-slate-700 dark:text-slate-300">
                      {activeShipment.eta ? new Date(activeShipment.eta).toLocaleDateString() : 'TBD'}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Container Assets */}
            <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200/70 dark:border-slate-700/60">
              <div className="flex items-center gap-2 text-xs font-bold text-slate-500 dark:text-slate-400 mb-3">
                <Box className="w-4 h-4 text-purple-500" />
                <span>CONTAINERS & BILL OF LADING</span>
              </div>
              <div className="space-y-2">
                <div>
                  <span className="text-[10px] font-semibold text-slate-400 uppercase">Container Numbers</span>
                  <div className="flex flex-wrap gap-1.5 mt-1">
                    {activeShipment.container_numbers && activeShipment.container_numbers.length > 0 ? (
                      activeShipment.container_numbers.map((c, i) => (
                        <span key={i} className="font-mono text-xs px-2 py-0.5 rounded bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-200 font-bold">
                          {c}
                        </span>
                      ))
                    ) : (
                      <span className="text-xs text-slate-500">Containers assigned at depot</span>
                    )}
                  </div>
                </div>
                <div className="border-t border-slate-200 dark:border-slate-700 pt-2 flex items-center justify-between text-xs">
                  <div>
                    <span className="text-[10px] font-semibold text-slate-400 uppercase block">MBL Number</span>
                    <span className="font-mono text-slate-700 dark:text-slate-300">{activeShipment.mbl_number || 'Pending'}</span>
                  </div>
                  <div className="text-right">
                    <span className="text-[10px] font-semibold text-slate-400 uppercase block">HBL Number</span>
                    <span className="font-mono text-slate-700 dark:text-slate-300">{activeShipment.hbl_number || 'Pending'}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── 4. HISTORICAL SHIPMENTS (IF MULTIPLE) ── */}
      {shipments.length > 1 && (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
          <div className="p-4 border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/40">
            <h4 className="text-xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
              All Linked Shipments ({shipments.length})
            </h4>
          </div>
          <div className="divide-y divide-slate-100 dark:divide-slate-800">
            {shipments.map(s => (
              <div key={s.id} className="p-3.5 flex flex-wrap items-center justify-between gap-3 text-xs">
                <div className="flex items-center gap-3">
                  <span className="font-bold text-slate-900 dark:text-slate-100">Shipment #{s.id}</span>
                  <span className="text-slate-600 dark:text-slate-400 font-medium">{s.carrier_name || s.carrier_scac}</span>
                  <span className="text-slate-400">{s.origin_port} → {s.destination_port}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold border ${getShipmentStatusBadgeClass(s.status)}`}>
                    {s.status}
                  </span>
                  {s.active_exceptions_count > 0 && (
                    <span className="px-1.5 py-0.5 rounded bg-red-100 dark:bg-red-950 text-red-700 dark:text-red-300 text-[10px] font-bold border border-red-200 dark:border-red-800 flex items-center gap-1" title={`${s.active_exceptions_count} active exception(s)`}>
                      <AlertTriangle className="w-3 h-3 text-red-600" /> {s.active_exceptions_count}
                    </span>
                  )}
                  <button
                    onClick={() => navigate(`/dashboard/shipments`)}
                    className="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400 hover:text-slate-600"
                    title="Open shipment"
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
  );
}
