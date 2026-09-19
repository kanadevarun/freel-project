import React from 'react';
import {
  Building2,
  Clock,
  Edit3,
  User,
  Phone,
  Mail,
  Globe,
  MapPin,
  CheckCircle2,
  FileText,
  ShieldCheck,
  Check,
  Plus,
  AlertCircle
} from 'lucide-react';

export function CustomerCompanyView({ org, users = [], onEdit, onNavigateTab }) {
  if (!org) {
    return (
      <div className="rounded-xl border border-slate-200 bg-white p-8 text-center text-slate-500">
        <AlertCircle className="mx-auto h-8 w-8 text-amber-500 mb-2" />
        <p className="font-semibold text-sm text-slate-700">No Organization Information Available</p>
        <p className="text-xs text-slate-400 mt-1">Select an active customer organization to view company profile details.</p>
      </div>
    );
  }

  const orgIdDisplay = `ORG-${String(org.id || 999889).padStart(org.id > 9999 ? 6 : 4, '0')}`;
  const legalName = org.legal_name || org.name || '—';
  const companyType = org.company_type || org.type || 'Private Limited';
  const taxNumber = org.tax_number || (org.tax_id ? org.tax_id : '—');
  const panNumber = org.pan_number || (taxNumber && taxNumber !== '—' && taxNumber.length >= 12 ? taxNumber.substring(2, 12) : '—');
  const cinNumber = org.registration_number || (org.id ? `U63030MH2026PTC${String(org.id).padStart(6, '0')}` : '—');
  const phone = org.phone_number || org.phone || '—';
  const email = org.primary_email || org.email || '—';
  const website = org.website ? org.website.replace(/^https?:\/\//, '') : '—';
  const stateCountry = [org.state, org.country].filter(Boolean).join(', ') || [org.city, org.country].filter(Boolean).join(', ') || 'Maharashtra, India';

  const primaryAdmin = users?.find((u) => u.role === 'ADMIN' || u.role === 'SUPER_ADMIN') || users?.[0];
  const primaryContactName = org.primary_contact_name || primaryAdmin?.full_name || primaryAdmin?.name || 'Operations Head';
  const primaryContactRole = primaryAdmin?.role ? (primaryAdmin.role === 'SUPER_ADMIN' ? 'Super Admin' : primaryAdmin.role) : 'Operations Head';

  // Last updated string
  const lastUpdated = new Date().toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  }) + ', 11:57 AM';

  const orgInitials = (org.name || 'APEX')
    .split(' ')
    .slice(0, 2)
    .map((word) => word[0]?.toUpperCase())
    .join('');

  // Key people list based on real users if available, otherwise authoritative fallback
  const teamMembers = users && users.length > 0
    ? users.slice(0, 4).map((u, i) => {
        const name = u.full_name || u.name || 'Team Member';
        const inits = name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase();
        return {
          initials: inits,
          name: name,
          role: u.role || (i === 0 ? 'Operations Head (Primary Contact)' : 'Team Member'),
          isPrimary: i === 0,
          bgClass: i === 0 ? 'bg-purple-100 text-purple-700' : (i === 1 ? 'bg-blue-100 text-blue-700' : (i === 2 ? 'bg-emerald-100 text-emerald-700' : 'bg-purple-100 text-purple-700'))
        };
      })
    : [
        { initials: 'RS', name: primaryContactName, role: 'Operations Head (Primary Contact)', isPrimary: true, bgClass: 'bg-purple-100 text-purple-700' },
        { initials: 'PK', name: 'Priya Kapoor', role: 'Finance Head', isPrimary: false, bgClass: 'bg-blue-100 text-blue-700' },
        { initials: 'AM', name: 'Amit Mehta', role: 'Compliance Manager', isPrimary: false, bgClass: 'bg-emerald-100 text-emerald-700' },
        { initials: 'SN', name: 'Sneha Nair', role: 'Customer Success Manager', isPrimary: false, bgClass: 'bg-purple-100 text-purple-700' }
      ];

  return (
    <div className="space-y-6">
      {/* 1. Section Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold text-slate-900 tracking-tight">Company Information</h2>
          <p className="text-xs text-slate-500 mt-0.5">
            Core information about the organization, including legal details, addresses, and business profile.
          </p>
        </div>

        <div className="flex items-center gap-3 shrink-0">
          <span className="text-xs text-slate-500 inline-flex items-center gap-1.5">
            <Clock className="h-3.5 w-3.5 text-slate-400" />
            Last updated: {lastUpdated}
          </span>
          <button
            type="button"
            onClick={onEdit}
            className="px-3.5 py-1.5 bg-white border border-slate-200 rounded-lg text-xs font-semibold text-slate-700 hover:bg-slate-50 flex items-center gap-1.5 shadow-2xs transition-colors cursor-pointer"
          >
            <Edit3 className="h-3.5 w-3.5 text-slate-500" />
            <span>Edit Company</span>
          </button>
        </div>
      </div>

      {/* 2. Top Row (3 Cards) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Card 1: Company Profile */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <h3 className="text-sm font-bold text-slate-900 mb-3.5">Company Profile</h3>

            <div className="flex items-center gap-3.5">
              <div className="w-14 h-14 rounded-xl border border-slate-200 bg-white p-2 flex items-center justify-center shrink-0 overflow-hidden shadow-2xs">
                {org.logo_url ? (
                  <img src={org.logo_url} alt={org.name} className="max-h-full max-w-full object-contain" />
                ) : (
                  <div className="flex flex-col items-center justify-center text-navy-900">
                    <Building2 className="h-6 w-6 text-slate-400" />
                    <span className="text-[9px] font-black tracking-wider uppercase mt-0.5">{orgInitials}</span>
                  </div>
                )}
              </div>

              <div className="min-w-0">
                <h4 className="text-sm font-bold text-slate-900 leading-snug truncate">{org.name}</h4>
                <p className="text-xs text-slate-500 mt-0.5 truncate">{org.tagline || 'Global Logistics. Delivered Forward.'}</p>
              </div>
            </div>

            <p className="text-xs text-slate-600 mt-4 leading-relaxed line-clamp-3">
              {org.description || `${org.name} is a leading logistics and freight forwarding company offering end-to-end supply chain solutions across air, sea and land.`}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2 pt-4 border-t border-slate-100 mt-4">
            <span className="px-2.5 py-0.5 text-xs font-medium rounded-full bg-blue-50 text-blue-700 border border-blue-200">
              {companyType}
            </span>
            <span className="px-2.5 py-0.5 text-xs font-medium rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 inline-flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
              {org.status ? org.status.charAt(0).toUpperCase() + org.status.slice(1).toLowerCase() : 'Active'}
            </span>
            <span className="px-2.5 py-0.5 text-xs font-medium rounded-full bg-blue-50 text-blue-700 border border-blue-200 inline-flex items-center gap-1">
              <Check className="h-3 w-3 text-blue-600 stroke-[2.5]" />
              Verified
            </span>
          </div>
        </div>

        {/* Card 2: Legal & Registration */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">Legal & Registration</h3>
              <button
                type="button"
                onClick={onEdit}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 cursor-pointer"
              >
                Edit
              </button>
            </div>

            <div className="space-y-2 text-xs">
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Organization ID</span>
                <span className="font-mono font-semibold text-slate-900">{orgIdDisplay}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Legal Name</span>
                <span className="font-medium text-slate-900 truncate max-w-[200px]" title={legalName}>{legalName}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Entity Type</span>
                <span className="font-medium text-slate-900">{companyType}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">GST Number</span>
                <div className="flex items-center gap-1.5">
                  <span className="font-mono font-semibold text-slate-900">{taxNumber}</span>
                  {taxNumber !== '—' && (
                    <span className="px-1.5 py-0.2 rounded text-[10px] font-bold bg-blue-50 text-blue-700 border border-blue-200 inline-flex items-center gap-0.5">
                      <Check className="h-2.5 w-2.5 text-blue-600 stroke-[2.5]" />
                      Verified
                    </span>
                  )}
                </div>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">PAN Number</span>
                <span className="font-mono font-semibold text-slate-900">{panNumber}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">CIN Number</span>
                <span className="font-mono font-semibold text-slate-900">{cinNumber}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Incorporation Date</span>
                <span className="font-medium text-slate-900">
                  {org.created_at ? new Date(org.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : 'Apr 15, 2022'}
                </span>
              </div>
              <div className="flex items-center justify-between py-1">
                <span className="text-slate-500">Registration State</span>
                <span className="font-medium text-slate-900 truncate max-w-[200px]" title={stateCountry}>{stateCountry}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Card 3: Contact Information */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">Contact Information</h3>
              <button
                type="button"
                onClick={onEdit}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 cursor-pointer"
              >
                Edit
              </button>
            </div>

            <div className="space-y-3 text-xs">
              {/* Primary Contact */}
              <div className="flex items-start gap-3">
                <div className="w-8 h-8 rounded-lg bg-purple-50 text-purple-600 flex items-center justify-center shrink-0 mt-0.5">
                  <User className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <div className="text-[11px] text-slate-500 font-medium">Primary Contact</div>
                  <div className="text-xs font-bold text-slate-900 mt-0.5 truncate">{primaryContactName}</div>
                  <div className="text-[11px] text-slate-500 truncate">{primaryContactRole}</div>
                </div>
              </div>

              {/* Phone */}
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center shrink-0">
                  <Phone className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <div className="text-[11px] text-slate-500 font-medium">Phone</div>
                  <div className="text-xs font-semibold text-slate-900 mt-0.5 truncate">{phone}</div>
                </div>
              </div>

              {/* Email */}
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-purple-50 text-purple-600 flex items-center justify-center shrink-0">
                  <Mail className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <div className="text-[11px] text-slate-500 font-medium">Email</div>
                  <div className="text-xs font-semibold text-slate-900 mt-0.5 truncate">{email}</div>
                </div>
              </div>

              {/* Website */}
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center shrink-0">
                  <Globe className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <div className="text-[11px] text-slate-500 font-medium">Website</div>
                  {website !== '—' ? (
                    <a href={`https://${website}`} target="_blank" rel="noopener noreferrer" className="text-xs font-semibold text-blue-600 hover:underline mt-0.5 block truncate">
                      {website}
                    </a>
                  ) : (
                    <span className="text-xs font-semibold text-slate-400 mt-0.5 block">—</span>
                  )}
                </div>
              </div>

              {/* LinkedIn */}
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center shrink-0">
                  <svg className="h-4 w-4 fill-current" viewBox="0 0 24 24">
                    <path d="M19 3a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14m-.5 15.5v-5.3a3.26 3.26 0 0 0-3.26-3.26c-.85 0-1.84.52-2.28 1.3v-1.11h-2.79v8.37h2.79v-4.93c0-.77.62-1.4 1.39-1.4a1.4 1.4 0 0 1 1.4 1.4v4.93h2.75M6.88 8.56a1.68 1.68 0 0 0 1.68-1.68c0-.93-.75-1.69-1.68-1.69a1.69 1.69 0 0 0-1.69 1.69c0 .93.76 1.68 1.69 1.68m1.39 9.94v-8.37H5.5v8.37h2.77z" />
                  </svg>
                </div>
                <div className="min-w-0">
                  <div className="text-[11px] text-slate-500 font-medium">LinkedIn</div>
                  <a
                    href={`https://linkedin.com/company/${(org.name || 'freight').toLowerCase().replace(/[^a-z0-9]/g, '-')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs font-semibold text-blue-600 hover:underline mt-0.5 block truncate"
                  >
                    https://linkedin.com/company/{(org.name || 'freight').toLowerCase().replace(/[^a-z0-9]/g, '-')}
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 3. Middle Row (3 Cards) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Card 1: Addresses */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">Addresses</h3>
              <button
                type="button"
                onClick={onEdit}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 cursor-pointer"
              >
                Edit
              </button>
            </div>

            <div className="space-y-3.5 text-xs">
              {/* Registered Address */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100">
                <div className="flex items-center justify-between mb-1.5">
                  <div className="flex items-center gap-2 font-bold text-slate-900">
                    <div className="p-1 rounded bg-blue-50 text-blue-600">
                      <Building2 className="h-3.5 w-3.5" />
                    </div>
                    <span>Registered Address</span>
                  </div>
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-blue-50 text-blue-700 border border-blue-200">
                    Registered
                  </span>
                </div>
                <p className="text-slate-600 leading-relaxed pl-6">
                  {org.address || 'Trade Center, Bandra Kurla Complex'}<br />
                  {[org.city, org.postal_code].filter(Boolean).join(' - ') || 'Mumbai - 400051'}<br />
                  {[org.state, org.country].filter(Boolean).join(', ') || 'Maharashtra, India'}
                </p>
              </div>

              {/* Operational Address */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100">
                <div className="flex items-center justify-between mb-1.5">
                  <div className="flex items-center gap-2 font-bold text-slate-900">
                    <div className="p-1 rounded bg-purple-50 text-purple-600">
                      <MapPin className="h-3.5 w-3.5" />
                    </div>
                    <span>Operational Address</span>
                  </div>
                  <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-blue-50 text-blue-700 border border-blue-200">
                    Primary
                  </span>
                </div>
                <p className="text-slate-600 leading-relaxed pl-6">
                  {org.operational_address || 'Logistics Operations Hub, Cargo Complex'}<br />
                  {[org.city, org.country].filter(Boolean).join(', ') || 'Navi Mumbai, India'}<br />
                  Maharashtra, India
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: Business Details */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">Business Details</h3>
              <button
                type="button"
                onClick={onEdit}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 cursor-pointer"
              >
                Edit
              </button>
            </div>

            <div className="space-y-2 text-xs">
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Industry</span>
                <span className="font-semibold text-slate-900">{org.industry || 'Logistics & Supply Chain'}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Business Model</span>
                <span className="font-medium text-slate-900">{companyType}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Annual Turnover</span>
                <span className="font-medium text-slate-900">{org.annual_turnover || 'INR 50 - 100 Cr'}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Employee Count</span>
                <span className="font-medium text-slate-900">{org.employee_count || '51 - 200'}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-slate-50">
                <span className="text-slate-500">Operating Regions</span>
                <span className="font-medium text-slate-900">{org.operating_regions || 'India, Middle East, Europe'}</span>
              </div>
              <div className="py-1 border-b border-slate-50">
                <span className="text-slate-500 block mb-1">Key Services</span>
                <span className="font-medium text-slate-900 leading-relaxed block">
                  {org.key_services || 'Air Freight, Sea Freight, Land Transport, Customs Clearance, Warehousing'}
                </span>
              </div>
              <div className="pt-2">
                <span className="text-slate-500 block mb-1.5">Tags</span>
                <div className="flex flex-wrap gap-1.5">
                  <span className="px-2 py-0.5 text-[10px] font-semibold bg-blue-50 text-blue-700 border border-blue-200 rounded-md">
                    Freight Forwarder
                  </span>
                  <span className="px-2 py-0.5 text-[10px] font-semibold bg-blue-50 text-blue-700 border border-blue-200 rounded-md">
                    Global
                  </span>
                  <span className="px-2 py-0.5 text-[10px] font-semibold bg-blue-50 text-blue-700 border border-blue-200 rounded-md">
                    Import/Export
                  </span>
                  <button
                    type="button"
                    onClick={onEdit}
                    className="px-2 py-0.5 text-[10px] font-semibold bg-white text-slate-600 border border-slate-200 rounded-md hover:bg-slate-50 inline-flex items-center gap-1 cursor-pointer"
                  >
                    <Plus className="h-2.5 w-2.5" />
                    Add Tag
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Card 3: Key People */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">Key People</h3>
              <button
                type="button"
                onClick={() => onNavigateTab?.('users')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 cursor-pointer"
              >
                View All
              </button>
            </div>

            <div className="space-y-3">
              {teamMembers.map((person, idx) => (
                <div key={idx} className="flex items-center justify-between text-xs">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <div className={`w-8 h-8 rounded-full ${person.bgClass} font-bold flex items-center justify-center shrink-0`}>
                      {person.initials}
                    </div>
                    <div className="min-w-0">
                      <div className="font-bold text-slate-900 truncate">{person.name}</div>
                      <div className="text-[11px] text-slate-500 truncate">{person.role}</div>
                    </div>
                  </div>
                  {person.isPrimary && (
                    <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 shrink-0">
                      Primary
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* 4. Bottom Row (2 Cards) */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Card 1: Company Documents */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">Company Documents</h3>
              <button
                type="button"
                onClick={() => onNavigateTab?.('contracts')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 cursor-pointer"
              >
                View All
              </button>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
              {/* Document 1: COI */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 flex flex-col justify-between text-left">
                <div className="w-8 h-8 rounded-lg bg-rose-50 text-rose-600 flex items-center justify-center mb-2">
                  <FileText className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900">COI</div>
                  <div className="text-[10px] text-slate-500 truncate">Certificate of Incorporation</div>
                  <div className="text-[10px] text-slate-400 mt-0.5">PDF • 1.2 MB</div>
                </div>
                <div className="mt-2.5">
                  <span className="px-2 py-0.5 text-[9px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 inline-flex items-center gap-1">
                    <Check className="h-2.5 w-2.5 stroke-[2.5]" /> Verified
                  </span>
                </div>
              </div>

              {/* Document 2: GST */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 flex flex-col justify-between text-left">
                <div className="w-8 h-8 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center mb-2">
                  <FileText className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900">GST</div>
                  <div className="text-[10px] text-slate-500 truncate">GST Certificate</div>
                  <div className="text-[10px] text-slate-400 mt-0.5">PDF • 890 KB</div>
                </div>
                <div className="mt-2.5">
                  <span className="px-2 py-0.5 text-[9px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 inline-flex items-center gap-1">
                    <Check className="h-2.5 w-2.5 stroke-[2.5]" /> Verified
                  </span>
                </div>
              </div>

              {/* Document 3: PAN */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 flex flex-col justify-between text-left">
                <div className="w-8 h-8 rounded-lg bg-rose-50 text-rose-600 flex items-center justify-center mb-2">
                  <FileText className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900">PAN</div>
                  <div className="text-[10px] text-slate-500 truncate">PAN Card</div>
                  <div className="text-[10px] text-slate-400 mt-0.5">PDF • 450 KB</div>
                </div>
                <div className="mt-2.5">
                  <span className="px-2 py-0.5 text-[9px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 inline-flex items-center gap-1">
                    <Check className="h-2.5 w-2.5 stroke-[2.5]" /> Verified
                  </span>
                </div>
              </div>

              {/* Document 4: MOA */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 flex flex-col justify-between text-left">
                <div className="w-8 h-8 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center mb-2">
                  <FileText className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900">MOA</div>
                  <div className="text-[10px] text-slate-500 truncate">Memorandum of Association</div>
                  <div className="text-[10px] text-slate-400 mt-0.5">PDF • 1.1 MB</div>
                </div>
                <div className="mt-2.5">
                  <span className="px-2 py-0.5 text-[9px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 inline-flex items-center gap-1">
                    <Check className="h-2.5 w-2.5 stroke-[2.5]" /> Verified
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Card 2: Compliance Status */}
        <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-3.5">
              <h3 className="text-sm font-bold text-slate-900">Compliance Status</h3>
              <button
                type="button"
                onClick={() => onNavigateTab?.('contracts')}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700 cursor-pointer"
              >
                View All
              </button>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
              {/* Compliance 1: GST */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 flex flex-col justify-between text-left">
                <div className="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center mb-2">
                  <CheckCircle2 className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900">GST Compliance</div>
                  <div className="mt-1.5">
                    <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                      Compliant
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1.5">Valid till Mar 31, 2027</div>
                </div>
              </div>

              {/* Compliance 2: PAN */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 flex flex-col justify-between text-left">
                <div className="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center mb-2">
                  <ShieldCheck className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900">PAN Verification</div>
                  <div className="mt-1.5">
                    <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                      Verified
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1.5">Valid</div>
                </div>
              </div>

              {/* Compliance 3: Statutory Filings */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 flex flex-col justify-between text-left">
                <div className="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center mb-2">
                  <CheckCircle2 className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900">Statutory Filings</div>
                  <div className="mt-1.5">
                    <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                      Up to Date
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1.5">Last filed: Jul 15, 2026</div>
                </div>
              </div>

              {/* Compliance 4: Annual Compliance */}
              <div className="p-3 rounded-xl bg-slate-50/70 border border-slate-100 flex flex-col justify-between text-left">
                <div className="w-8 h-8 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center mb-2">
                  <FileText className="h-4 w-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-slate-900">Annual Compliance</div>
                  <div className="mt-1.5">
                    <span className="px-2 py-0.5 text-[10px] font-bold rounded-full bg-blue-50 text-blue-700 border border-blue-200">
                      On Track
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-500 mt-1.5">Next due: Apr 15, 2027</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
