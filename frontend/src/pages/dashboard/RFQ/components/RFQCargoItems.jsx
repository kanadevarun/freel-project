import React, { useState } from 'react';
import {
  Package,
  Ship,
  Plane,
  Truck,
  ArrowRight,
  Shield,
  CheckCircle2,
  AlertCircle,
  Clock,
  Info,
  ChevronDown,
  ChevronUp,
  FileText,
  Thermometer,
  Flame,
  Calendar,
  MapPin
} from 'lucide-react';

const INCOTERMS_EXPLANATIONS = {
  FOB: {
    title: 'Free On Board (FOB)',
    summary: 'Buyer bears all carriage costs, marine insurance, and risk once goods are loaded on the vessel.',
    seller: 'Clears goods for export and loads them onto the nominated vessel at origin port.',
    buyer: 'Pays main ocean freight, insurance, destination terminal fees, and import customs.',
  },
  CIF: {
    title: 'Cost, Insurance and Freight (CIF)',
    summary: 'Seller pays for ocean freight and marine cargo insurance to destination port.',
    seller: 'Arranges and pays for carriage and minimum marine insurance cover to the destination port.',
    buyer: 'Assumes risk once cargo is loaded and handles destination port clearance & inland haulage.',
  },
  EXW: {
    title: 'Ex Works (EXW)',
    summary: 'Buyer assumes maximum obligation, arranging collection directly from seller premises.',
    seller: 'Makes packaged goods available at facility/factory.',
    buyer: 'Handles export haulage, export clearance, ocean freight, and destination delivery.',
  },
  DDP: {
    title: 'Delivered Duty Paid (DDP)',
    summary: 'Seller bears all risks and transportation costs including destination import duty and taxes.',
    seller: 'Full door-to-door responsibility, all freight legs, and destination customs duty.',
    buyer: 'Receives cleared goods at destination delivery address.',
  },
  DAP: {
    title: 'Delivered At Place (DAP)',
    summary: 'Seller pays for carriage to named destination, ready for unloading by the buyer.',
    seller: 'Pays freight carriage to named destination terminal or premises.',
    buyer: 'Handles import clearance formalities, tariffs, and cargo unloading.',
  },
  FCA: {
    title: 'Free Carrier (FCA)',
    summary: 'Seller delivers goods, cleared for export, to carrier nominated by buyer at named place.',
    seller: 'Delivers cargo to carrier/depot and completes export clearance.',
    buyer: 'Assumes carriage costs and risk from handover point.',
  },
  CFR: {
    title: 'Cost and Freight (CFR)',
    summary: 'Seller pays freight to destination port, but risk transfers upon loading on vessel.',
    seller: 'Arranges and pays for freight carriage to named destination port.',
    buyer: 'Procures marine cargo insurance and manages destination clearance.',
  },
  CPT: {
    title: 'Carriage Paid To (CPT)',
    summary: 'Seller pays carriage to named place; risk transfers upon handover to first carrier.',
    seller: 'Arranges freight carriage to named destination.',
    buyer: 'Assumes risk upon first carrier pickup and procures cargo insurance.',
  },
  CIP: {
    title: 'Carriage and Insurance Paid To (CIP)',
    summary: 'Seller pays carriage and comprehensive all-risks insurance to named destination.',
    seller: 'Arranges freight carriage and maximum insurance coverage.',
    buyer: 'Handles destination import formalities and final unloading.',
  }
};

export default function RFQCargoItems({ rfq, completeness, requirements, onSwitchTab }) {
  const [expandedRow, setExpandedRow] = useState(null);

  const items = Array.isArray(rfq?.items) ? rfq.items : [];
  const incotermKey = (rfq?.incoterms || 'FOB').toUpperCase();
  const incotermInfo = INCOTERMS_EXPLANATIONS[incotermKey] || {
    title: `${rfq?.incoterms || 'Standard'} Incoterm`,
    summary: 'Commercial carriage responsibility mapped per trade contract agreement.',
    seller: 'Standard export obligations and delivery terms apply.',
    buyer: 'Commercial carriage responsibility per agreed logistics contract.',
  };

  // Derive mode hint
  const originStr = (rfq?.origin || '').toUpperCase();
  const destStr = (rfq?.destination || '').toUpperCase();
  const isAir = originStr.includes('AIRPORT') || destStr.includes('AIRPORT') || originStr.startsWith('AIR');
  const ModeIcon = isAir ? Plane : Ship;
  const modeLabel = isAir ? 'Air Freight' : 'Ocean Freight';

  // Map real readiness checklist from Task 10 requirements engine
  const reqGroups = requirements?.groups || [];
  const shipmentGroup = reqGroups.find(g => g.category === 'SHIPMENT_INFO');
  const cargoGroup = reqGroups.find(g => g.category === 'CARGO_OPERATIONAL');
  const complianceGroup = reqGroups.find(g => g.category === 'CONDITIONAL_COMPLIANCE');

  const allReqs = [
    ...(shipmentGroup?.requirements || []),
    ...(cargoGroup?.requirements || []),
    ...(complianceGroup?.requirements || [])
  ];

  // Helper to extract port name vs code
  const formatPort = (val) => {
    if (!val) return { name: 'Not yet specified', code: null, isMissing: true };
    const parts = val.split('(');
    if (parts.length > 1) {
      return {
        name: parts[0].trim(),
        code: parts[1].replace(')', '').trim(),
        isMissing: false
      };
    }
    return { name: val.trim(), code: null, isMissing: false };
  };

  const originPort = formatPort(rfq?.origin);
  const destPort = formatPort(rfq?.destination);

  const toggleRow = (idx) => {
    setExpandedRow(expandedRow === idx ? null : idx);
  };

  return (
    <div className="space-y-5 animate-fadeIn" data-testid="rfq-cargo-workspace">
      
      {/* ── 1. Shipment Profile & Route Journey ── */}
      <div className="bg-white border border-slate-200 rounded-xl p-6 shadow-xs">
        <div className="flex items-center justify-between pb-4 border-b border-slate-100 mb-6">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-indigo-50 border border-indigo-100 flex items-center justify-center text-indigo-600">
              <ModeIcon className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-base font-bold text-slate-900">
                Shipment Profile & Routing
              </h3>
              <p className="text-xs text-slate-500">
                Port of Loading (POL) to Port of Discharge (POD) routing and commercial terms.
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <span className="px-2.5 py-1 bg-slate-100 text-slate-700 font-bold text-xs rounded-lg border border-slate-200">
              {modeLabel}
            </span>
            <span className="px-2.5 py-1 bg-indigo-50 text-indigo-700 font-bold text-xs rounded-lg border border-indigo-200">
              Incoterm: {rfq?.incoterms || 'FOB'}
            </span>
          </div>
        </div>

        {/* Horizontal Route Journey Visualization */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-center">
          
          {/* Origin Node */}
          <div className="lg:col-span-4 bg-slate-50 border border-slate-200 rounded-xl p-4 flex items-center gap-3.5">
            <div className="w-10 h-10 rounded-xl bg-blue-100/70 border border-blue-200 flex items-center justify-center text-blue-700 flex-shrink-0">
              <MapPin className="w-5 h-5" />
            </div>
            <div className="min-w-0">
              <div className="text-[11px] font-bold text-blue-700 uppercase tracking-wider">
                Origin (POL)
              </div>
              <div className="text-sm font-bold text-slate-900 truncate mt-0.5">
                {originPort.name}
              </div>
              {originPort.code ? (
                <div className="text-[11px] font-semibold text-slate-500 mt-0.5 inline-block bg-white px-1.5 py-0.2 rounded border border-slate-200">
                  Code: {originPort.code}
                </div>
              ) : originPort.isMissing ? (
                <div className="text-[11px] font-semibold text-amber-600 mt-0.5">
                  Pending customer confirmation
                </div>
              ) : null}
            </div>
          </div>

          {/* Center Transit Graphic */}
          <div className="lg:col-span-4 flex flex-col items-center justify-center py-2">
            <div className="flex items-center justify-center gap-2 w-full">
              <div className="h-0.5 flex-1 bg-slate-200"></div>
              <div className="px-3 py-1 bg-indigo-50 border border-indigo-200 rounded-full text-indigo-700 flex items-center gap-1.5 text-xs font-bold shadow-2xs">
                <ModeIcon className="w-3.5 h-3.5" />
                <span>{modeLabel}</span>
              </div>
              <div className="h-0.5 flex-1 bg-slate-200"></div>
              <ArrowRight className="w-4 h-4 text-slate-400 -ml-1.5 flex-shrink-0" />
            </div>

            {rfq?.target_date && (
              <div className="flex items-center gap-1 text-[11px] text-slate-500 font-medium mt-2">
                <Calendar className="w-3.5 h-3.5 text-slate-400" />
                <span>Target Departure: {new Date(rfq.target_date).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}</span>
              </div>
            )}
          </div>

          {/* Destination Node */}
          <div className="lg:col-span-4 bg-slate-50 border border-slate-200 rounded-xl p-4 flex items-center gap-3.5">
            <div className="w-10 h-10 rounded-xl bg-emerald-100/70 border border-emerald-200 flex items-center justify-center text-emerald-700 flex-shrink-0">
              <MapPin className="w-5 h-5" />
            </div>
            <div className="min-w-0">
              <div className="text-[11px] font-bold text-emerald-700 uppercase tracking-wider">
                Destination (POD)
              </div>
              <div className="text-sm font-bold text-slate-900 truncate mt-0.5">
                {destPort.name}
              </div>
              {destPort.code ? (
                <div className="text-[11px] font-semibold text-slate-500 mt-0.5 inline-block bg-white px-1.5 py-0.2 rounded border border-slate-200">
                  Code: {destPort.code}
                </div>
              ) : destPort.isMissing ? (
                <div className="text-[11px] font-semibold text-amber-600 mt-0.5">
                  Pending customer confirmation
                </div>
              ) : null}
            </div>
          </div>

        </div>

        {/* Commercial Incoterms Breakdown */}
        <div className="mt-5 p-4 bg-slate-50/70 border border-slate-200 rounded-xl">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-2">
            <div className="flex items-center gap-2">
              <span className="text-xs font-bold text-slate-800 uppercase tracking-wider">
                Commercial Terms: {incotermInfo.title}
              </span>
            </div>
            <span className="text-[11px] text-slate-500 font-medium">
              Incoterms® 2020 Rules
            </span>
          </div>

          <p className="text-xs text-slate-600 mb-3 leading-relaxed">
            {incotermInfo.summary}
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
            <div className="p-2.5 bg-white rounded-lg border border-slate-200">
              <span className="font-bold text-slate-800 block mb-0.5">Seller Responsibility:</span>
              <span className="text-slate-600 leading-normal">{incotermInfo.seller}</span>
            </div>
            <div className="p-2.5 bg-white rounded-lg border border-slate-200">
              <span className="font-bold text-slate-800 block mb-0.5">Buyer Responsibility:</span>
              <span className="text-slate-600 leading-normal">{incotermInfo.buyer}</span>
            </div>
          </div>
        </div>
      </div>

      {/* ── 2. Cargo Manifest & Operational Breakdown Table ── */}
      <div className="bg-white border border-slate-200 rounded-xl overflow-hidden shadow-xs">
        <div className="px-6 py-4 bg-slate-50/80 border-b border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <div className="flex items-center gap-2">
              <Package className="w-4 h-4 text-indigo-600" />
              <h3 className="text-xs font-bold text-slate-900 uppercase tracking-wider">
                Cargo Manifest ({items.length} {items.length === 1 ? 'Item' : 'Items'})
              </h3>
            </div>
            <p className="text-xs text-slate-500 mt-0.5">
              Registered cargo lines, unit specifications, and volumetric measurements.
            </p>
          </div>

          {/* Integrated Manifest Metrics */}
          <div className="flex items-center gap-3">
            <div className="px-3 py-1.5 bg-white border border-slate-200 rounded-lg text-xs font-semibold text-slate-700">
              Total Weight: <span className="font-bold text-slate-900">{(completeness?.totalWeight || 0).toLocaleString()} KG</span>
            </div>
            <div className="px-3 py-1.5 bg-white border border-slate-200 rounded-lg text-xs font-semibold text-slate-700">
              Total Volume: <span className="font-bold text-slate-900">{(completeness?.totalVolume || 0).toLocaleString()} CBM</span>
            </div>
          </div>
        </div>

        {items.length === 0 ? (
          <div className="p-12 text-center bg-slate-50/50">
            <Package className="w-8 h-8 text-slate-400 mx-auto mb-2" />
            <div className="text-xs font-bold text-slate-700">No cargo items defined</div>
            <p className="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
              Cargo lines have not been added to this RFQ yet. Items can be extracted from customer emails or specified during quote intake.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="rfq-table">
              <thead>
                <tr>
                  <th style={{ width: '40px' }}>#</th>
                  <th>Commodity / Description</th>
                  <th>Quantity</th>
                  <th>Gross Weight</th>
                  <th>Volume</th>
                  <th style={{ textAlign: 'right', width: '90px' }}>Details</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, idx) => {
                  const isExpanded = expandedRow === idx;
                  const descLower = (item.description || '').toLowerCase();
                  const isDG = descLower.includes('dangerous') || descLower.includes('hazardous') || descLower.includes('flammable') || descLower.includes('chemical') || descLower.includes('solvent');
                  const isReefer = descLower.includes('temperature') || descLower.includes('reefer') || descLower.includes('frozen') || descLower.includes('perishable') || descLower.includes('chilled');

                  return (
                    <React.Fragment key={item.id || idx}>
                      <tr
                        onClick={() => toggleRow(idx)}
                        className="cursor-pointer hover:bg-slate-50/80 transition-colors"
                      >
                        <td className="font-bold text-slate-400 text-xs">
                          {idx + 1}
                        </td>
                        <td>
                          <div className="font-bold text-slate-900 text-xs flex items-center gap-2">
                            <span>{item.description || 'General Cargo'}</span>
                            {isDG && (
                              <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-rose-50 text-rose-700 border border-rose-200 rounded text-[10px] font-bold">
                                <Flame className="w-2.5 h-2.5 text-rose-600" /> DG
                              </span>
                            )}
                            {isReefer && (
                              <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-blue-50 text-blue-700 border border-blue-200 rounded text-[10px] font-bold">
                                <Thermometer className="w-2.5 h-2.5 text-blue-600" /> Temp-Controlled
                              </span>
                            )}
                          </div>
                          <div className="text-[11px] text-slate-400 mt-0.5">
                            Line item #{item.id || idx + 1}
                          </div>
                        </td>
                        <td>
                          <span className="font-semibold text-slate-800 text-xs">
                            {item.quantity ? `${item.quantity} Units` : '1 Unit'}
                          </span>
                        </td>
                        <td>
                          {item.weight_kg ? (
                            <span className="font-semibold text-slate-800 text-xs">
                              {Number(item.weight_kg).toLocaleString()} KG
                            </span>
                          ) : (
                            <span className="text-xs font-semibold text-rose-600 bg-rose-50 px-2 py-0.5 rounded border border-rose-200">
                              Weight Required
                            </span>
                          )}
                        </td>
                        <td>
                          {item.volume_cbm ? (
                            <span className="font-semibold text-slate-800 text-xs">
                              {Number(item.volume_cbm).toLocaleString()} CBM
                            </span>
                          ) : (
                            <span className="text-xs font-semibold text-rose-600 bg-rose-50 px-2 py-0.5 rounded border border-rose-200">
                              Volume Required
                            </span>
                          )}
                        </td>
                        <td style={{ textAlign: 'right' }}>
                          <button
                            type="button"
                            className="text-xs font-bold text-indigo-600 hover:text-indigo-800 inline-flex items-center gap-1"
                          >
                            {isExpanded ? 'Less' : 'Inspect'}
                            {isExpanded ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                          </button>
                        </td>
                      </tr>

                      {/* Expandable Details Drawer inside row */}
                      {isExpanded && (
                        <tr className="bg-indigo-50/30">
                          <td colSpan={6} className="p-4 border-b border-indigo-100">
                            <div className="bg-white border border-indigo-100 rounded-xl p-4 shadow-xs space-y-3">
                              <div className="flex items-center justify-between pb-2 border-b border-slate-100">
                                <div className="text-xs font-bold text-slate-800">
                                  Line #{idx + 1} Detailed Specifications
                                </div>
                                <span className="text-[11px] text-slate-400">
                                  Commodity ID: {item.id || 'N/A'}
                                </span>
                              </div>

                              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
                                <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                                  <span className="text-slate-500 block mb-0.5">Commodity Description</span>
                                  <span className="font-bold text-slate-900">{item.description || 'Not specified'}</span>
                                </div>
                                <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                                  <span className="text-slate-500 block mb-0.5">Unit Weight</span>
                                  <span className="font-bold text-slate-900">
                                    {item.weight_kg ? `${(Number(item.weight_kg) / (item.quantity || 1)).toFixed(1)} KG / unit` : 'Not specified'}
                                  </span>
                                </div>
                                <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                                  <span className="text-slate-500 block mb-0.5">Unit Volume</span>
                                  <span className="font-bold text-slate-900">
                                    {item.volume_cbm ? `${(Number(item.volume_cbm) / (item.quantity || 1)).toFixed(2)} CBM / unit` : 'Not specified'}
                                  </span>
                                </div>
                              </div>

                              {/* Special Handling / Compliance tags */}
                              <div className="flex items-center gap-2 pt-1">
                                <span className="text-xs text-slate-500 font-semibold">Special Handling:</span>
                                {isDG ? (
                                  <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-rose-50 text-rose-700 text-xs font-bold rounded border border-rose-200">
                                    <Flame className="w-3 h-3 text-rose-600" /> Dangerous Goods Declaration Required
                                  </span>
                                ) : isReefer ? (
                                  <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-blue-50 text-blue-700 text-xs font-bold rounded border border-blue-200">
                                    <Thermometer className="w-3 h-3 text-blue-600" /> Temperature-Controlled Cargo
                                  </span>
                                ) : (
                                  <span className="text-xs text-slate-400 font-medium">Standard General Cargo (Ambient)</span>
                                )}
                              </div>
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* ── 3. Shipment Readiness & Intelligence (100% Real Task 10 Data) ── */}
      {requirements && (
        <div className="bg-white border border-slate-200 rounded-xl p-5 shadow-xs">
          <div className="flex items-center justify-between pb-3.5 border-b border-slate-100 mb-4">
            <div className="flex items-center gap-2">
              <Shield className="w-4 h-4 text-emerald-600" />
              <h3 className="text-xs font-bold text-slate-900 uppercase tracking-wider">
                Shipment Operational Readiness
              </h3>
            </div>

            {onSwitchTab && (
              <button
                onClick={() => onSwitchTab('requirements')}
                className="text-xs font-bold text-indigo-600 hover:text-indigo-700"
              >
                View Full Requirements →
              </button>
            )}
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-xs">
            {allReqs.map((req) => {
              const isSatisfied = req.status === 'SATISFIED';
              const isUnderReview = req.status === 'UNDER_REVIEW';
              const isMissing = req.status === 'MISSING';
              const isNA = req.status === 'NOT_APPLICABLE';

              return (
                <div
                  key={req.id}
                  className={`p-3 rounded-xl border flex items-start gap-2.5 transition-colors ${
                    isSatisfied
                      ? 'bg-emerald-50/50 border-emerald-200 text-emerald-950'
                      : isMissing
                      ? 'bg-rose-50/50 border-rose-200 text-rose-950'
                      : isUnderReview
                      ? 'bg-amber-50/50 border-amber-200 text-amber-950'
                      : 'bg-slate-50 border-slate-200 text-slate-700'
                  }`}
                >
                  {isSatisfied ? (
                    <CheckCircle2 className="w-4 h-4 text-emerald-600 flex-shrink-0 mt-0.5" />
                  ) : isMissing ? (
                    <AlertCircle className="w-4 h-4 text-rose-600 flex-shrink-0 mt-0.5" />
                  ) : isUnderReview ? (
                    <Clock className="w-4 h-4 text-amber-600 flex-shrink-0 mt-0.5" />
                  ) : (
                    <Info className="w-4 h-4 text-slate-400 flex-shrink-0 mt-0.5" />
                  )}

                  <div className="min-w-0">
                    <div className="font-bold text-xs truncate">
                      {req.title}
                    </div>
                    <div className="text-[11px] text-slate-500 mt-0.5 line-clamp-1">
                      {req.value || (isSatisfied ? 'Confirmed' : isNA ? 'Not applicable' : 'Pending input')}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

    </div>
  );
}
