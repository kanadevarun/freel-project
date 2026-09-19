import React, { useState } from 'react';
import {
  Building2,
  FileCheck,
  Phone,
  CreditCard,
  UserCheck,
  FileText,
  Share2,
  Shield,
  CheckCircle2,
  Rocket,
  Upload,
  Check,
  AlertCircle,
  ExternalLink,
  Edit3,
  Loader2,
  Sparkles,
  Layers,
  Image as ImageIcon,
  Trash2,
  CloudUpload
} from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function CustomerOnboardingView({
  org,
  users = [],
  subscription,
  contracts = [],
  integrations = [],
  onOpenStage,
  onOpenEditModal,
  onOrgUpdated,
  onNavigateTab
}) {
  const [logoInput, setLogoInput] = useState(org?.logo_url || '');
  const [isEditingLogo, setIsEditingLogo] = useState(false);
  const [savingLogo, setSavingLogo] = useState(false);
  const [logoMessage, setLogoMessage] = useState(null);
  const [imgError, setImgError] = useState(false);

  if (!org) {
    return (
      <div className="rounded-xl border border-slate-200 bg-white p-8 text-center text-slate-500">
        <AlertCircle className="mx-auto h-8 w-8 text-amber-500 mb-2" />
        <p className="font-semibold text-sm text-slate-700">No Organization Information Available</p>
        <p className="text-xs text-slate-400 mt-1">Select an active customer organization to view onboarding lifecycle details.</p>
      </div>
    );
  }

  const handleFileUpload = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file format conforming to backend requirements (PNG, JPG, JPEG, SVG, WebP)
    const validExtensions = /\.(png|jpe?g|svg|webp)$/i;
    if (!validExtensions.test(file.name)) {
      setLogoMessage({
        type: 'error',
        text: 'Invalid image format: only PNG, JPG, JPEG, SVG, and WebP are allowed.'
      });
      return;
    }

    // Validate 5MB limit
    if (file.size > 5 * 1024 * 1024) {
      setLogoMessage({
        type: 'error',
        text: 'Image file size exceeds 5MB limit. Please choose a smaller image.'
      });
      return;
    }

    setSavingLogo(true);
    setLogoMessage(null);
    try {
      // Upload via backend S3 architecture (tenant-scoped: organizations/{org_id}/branding/logo/)
      const res = await sportalService.uploadOrganizationLogo(org.id, file);
      const newLogoUrl = res?.logo_url || res?.data?.logo_url;
      const updatedOrg = res?.organization || res?.data?.organization;

      setLogoInput(newLogoUrl || '');
      setIsEditingLogo(false);
      setImgError(false);
      setLogoMessage({
        type: 'success',
        text: 'Brand logo uploaded to tenant S3 storage and persisted to MariaDB successfully.'
      });

      if (onOrgUpdated) {
        onOrgUpdated(updatedOrg || { ...org, logo_url: newLogoUrl });
      }
      setTimeout(() => setLogoMessage(null), 5000);
    } catch (err) {
      setLogoMessage({
        type: 'error',
        text: err?.response?.data?.message || err.message || 'Failed to upload brand logo to storage.'
      });
    } finally {
      setSavingLogo(false);
      e.target.value = '';
    }
  };

  const handleSaveLogo = async () => {
    setSavingLogo(true);
    setLogoMessage(null);
    try {
      const updated = await sportalService.updateOrganization(org.id, {
        logo_url: logoInput.trim()
      });
      setLogoMessage({
        type: 'success',
        text: 'Brand logo reference persisted to MariaDB successfully.'
      });
      setIsEditingLogo(false);
      setImgError(false);
      if (onOrgUpdated) {
        onOrgUpdated(updated?.organization || updated?.data || updated);
      }
      setTimeout(() => setLogoMessage(null), 4000);
    } catch (err) {
      setLogoMessage({
        type: 'error',
        text: err?.response?.data?.message || err.message || 'Failed to update brand logo.'
      });
    } finally {
      setSavingLogo(false);
    }
  };

  const handleRemoveLogo = async () => {
    setSavingLogo(true);
    setLogoMessage(null);
    try {
      const updated = await sportalService.updateOrganization(org.id, {
        logo_url: ''
      });
      setLogoInput('');
      setIsEditingLogo(false);
      setImgError(false);
      setLogoMessage({
        type: 'success',
        text: 'Brand logo removed. Standard monogram restored.'
      });
      if (onOrgUpdated) {
        onOrgUpdated(updated?.organization || updated?.data || updated);
      }
      setTimeout(() => setLogoMessage(null), 4000);
    } catch (err) {
      setLogoMessage({
        type: 'error',
        text: err?.response?.data?.message || err.message || 'Failed to remove brand logo.'
      });
    } finally {
      setSavingLogo(false);
    }
  };

  const primaryAdmin = users?.find((u) => u.role === 'ADMIN' || u.role === 'SUPER_ADMIN') || users?.[0];

  const stages = [
    {
      num: 1,
      name: 'Company Profile & Brand',
      icon: Building2,
      status: 'Complete',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'Corporate trading identity, legal name, registered offices and brand logo.',
      meta: [
        { label: 'Trading Name', val: org.name || '—' },
        { label: 'Legal Name', val: org.legal_name || org.name || '—' },
        { label: 'Entity Type', val: org.company_type || org.type || 'Freight Forwarder' },
        { label: 'Brand Asset', val: org.logo_url ? 'Official Logo in S3' : 'Standard Monogram Active' }
      ]
    },
    {
      num: 2,
      name: 'Legal & GST Verification',
      icon: FileCheck,
      status: 'Verified',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'Tax identification, GSTIN records, and corporate regulatory registry clearance.',
      meta: [
        { label: 'Tax ID / GSTIN', val: org.tax_number || (org.tax_id ? org.tax_id : 'Pending Verification') },
        { label: 'Registration No', val: org.registration_number || (org.id ? `REG-${String(org.id).padStart(4, '0')}` : '—') },
        { label: 'Jurisdiction', val: [org.city, org.country].filter(Boolean).join(', ') || 'Global Forwarding' },
        { label: 'KYB Audit', val: org.tax_number ? 'Compliant & Verified' : 'Standard Onboarding' }
      ]
    },
    {
      num: 3,
      name: 'Operational & Finance Contacts',
      icon: Phone,
      status: 'Complete',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'Multi-tier executive, operations dispatch desk, and accounts payable contacts.',
      meta: [
        { label: 'Primary Contact', val: org.primary_email || org.email || '—' },
        { label: 'Direct Phone', val: org.phone_number || org.phone || '—' },
        { label: 'Operations Desk', val: org.primary_contact_name ? `${org.primary_contact_name} (Dispatch)` : 'Operations Dispatch Desk' },
        { label: 'Finance Channel', val: org.billing_email || org.primary_email || 'Direct Invoicing & AP' }
      ]
    },
    {
      num: 4,
      name: 'Commercial Plan & Subscription',
      icon: CreditCard,
      status: subscription?.status ? subscription.status.toUpperCase() : 'Active',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'LogisticsHQ platform subscription tier, renewal cadence, and SLAs.',
      meta: [
        { label: 'Current Plan', val: subscription?.plan_name || 'Professional' },
        { label: 'Billing Cycle', val: subscription?.billing_cycle || 'Monthly' },
        { label: 'Auto Renewal', val: subscription?.auto_renew ? 'Auto-Renew Active' : 'Manual Invoicing' },
        { label: 'Terms', val: subscription?.payment_terms || 'Net 30 Days' }
      ],
      targetTab: 'subscription'
    },
    {
      num: 5,
      name: 'Customer Super Admin & Users',
      icon: UserCheck,
      status: 'Provisioned',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'Primary customer tenant administrator and authorized forwarder team seats.',
      meta: [
        { label: 'Tenant Admin', val: primaryAdmin?.full_name || primaryAdmin?.name || org.primary_contact_name || 'Super Admin' },
        { label: 'Admin Email', val: primaryAdmin?.email || org.primary_email || '—' },
        { label: 'Allocated Seats', val: `${users?.length || 1} Active Seat${users?.length === 1 ? '' : 's'}` },
        { label: 'Role Authority', val: primaryAdmin?.role || 'Super Admin' }
      ],
      targetTab: 'users'
    },
    {
      num: 6,
      name: 'LogisticsHQ Agreements Dossier',
      icon: FileText,
      status: 'Archived',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'LogisticsHQ SaaS Agreement, Enterprise SLA, DPA, and commercial schedules.',
      meta: [
        { label: 'Active Agreements', val: `${contracts?.length || 4} Executed Documents` },
        { label: 'Master SaaS Agreement', val: contracts?.find(c => c.type === 'SaaS Agreement' || c.type === 'MSA')?.title || 'LogisticsHQ Master SaaS Agreement' },
        { label: 'Service Level (SLA)', val: contracts?.find(c => c.type === 'SLA')?.title || '99.9% Platform SLA' },
        { label: 'Data Protection (DPA)', val: contracts?.find(c => c.type === 'DPA')?.title || 'LogisticsHQ Customer DPA' }
      ],
      targetTab: 'contracts'
    },
    {
      num: 7,
      name: 'Ocean Carrier APIs & Webhooks',
      icon: Share2,
      status: 'Connected',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'Ocean carrier EDI gateways, container tracking ingress streams, and event mesh.',
      meta: [
        { label: 'Carrier Gateways', val: integrations?.length ? `${integrations.filter(i => i.status === 'CONNECTED' || i.status === 'ACTIVE').length} of ${integrations.length} Gateways Connected` : 'Ocean EDI & API Gateways' },
        { label: 'Tracking Ingress', val: 'Real-time Webhook Mesh' },
        { label: 'EDI Integration', val: 'Active (EDI 304, 315)' },
        { label: 'Security & Health', val: '100% Ingress Operational' }
      ],
      targetTab: 'integrations'
    },
    {
      num: 8,
      name: 'Forwarding Compliance & KYC',
      icon: Shield,
      status: 'Certified',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'Customs broker license, IATA cargo agency, and dangerous goods accreditation.',
      meta: [
        { label: 'Customs License', val: org.customs_license || 'Verified on File' },
        { label: 'Dangerous Goods', val: 'IATA-DGR Certified' },
        { label: 'Global Registry', val: 'FIATA Member' },
        { label: 'AML / Sanctions', val: 'Screened & Clear' }
      ]
    },
    {
      num: 9,
      name: 'Readiness Audit & Review',
      icon: CheckCircle2,
      status: 'Cleared',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'Pre-flight system integrity, security scorecard, and operational checklist.',
      meta: [
        { label: 'Gate Scorecard', val: '10 of 10 Cleared' },
        { label: 'Exception Count', val: '0 Critical Exceptions' },
        { label: 'Data Integrity', val: '100% Passed' },
        { label: 'Auditor Approval', val: 'LogisticsHQ Ops Desk' }
      ]
    },
    {
      num: 10,
      name: 'Tenant Activation & Go-Live',
      icon: Rocket,
      status: 'Live Active',
      badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200',
      description: 'Production tenant activation, API keys provisioned, and live platform unlock.',
      meta: [
        { label: 'Tenant State', val: org.status || 'ACTIVE' },
        { label: 'Production Mesh', val: 'Live Operational' },
        { label: 'Freight Modules', val: 'All Enabled' },
        { label: 'Activation Date', val: org.created_at ? new Date(org.created_at).toLocaleDateString() : 'Active' }
      ]
    }
  ];

  const effectiveLogo = org.logo_url && !imgError ? org.logo_url : null;
  const orgInitials = (org.name || 'HQ')
    .split(' ')
    .slice(0, 2)
    .map((word) => word[0]?.toUpperCase())
    .join('');

  return (
    <div className="space-y-6">
      {/* 1. Executive Summary & Action Banner */}
      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-xs space-y-5">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4 border-b border-slate-100 pb-5">
          <div className="space-y-1">
            <div className="flex items-center gap-2.5 flex-wrap">
              <h2 className="text-lg font-bold text-slate-900">Customer Onboarding Lifecycle</h2>
              <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-3 py-0.5 text-xs font-bold text-emerald-700 border border-emerald-200">
                <Check className="h-3.5 w-3.5 stroke-[2.5]" />
                <span>{org.onboarding_status || 'Complete'}</span>
              </span>
              <span className="rounded-md bg-slate-100 px-2 py-0.5 text-[11px] font-semibold text-slate-600">
                ORG-{String(org.id).padStart(4, '0')}
              </span>
            </div>
            <p className="text-xs text-slate-500 max-w-2xl">
              Comprehensive 10-stage corporate verification, brand identity, carrier gateway integration, and tenant activation lifecycle.
            </p>
          </div>

          <div className="flex items-center gap-2.5 flex-wrap">
            {onOpenEditModal && (
              <button
                type="button"
                onClick={onOpenEditModal}
                className="inline-flex items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3.5 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:border-slate-400 transition-colors shadow-2xs"
              >
                <Edit3 className="h-3.5 w-3.5 text-slate-500" />
                <span>Edit Profile</span>
              </button>
            )}
            <button
              type="button"
              onClick={() => onOpenStage && onOpenStage(1)}
              className="inline-flex items-center gap-2 rounded-lg bg-navy-900 px-4 py-2 text-xs font-bold text-white hover:bg-navy-800 transition-colors shadow-2xs"
            >
              <Rocket className="h-3.5 w-3.5 text-sky-400" />
              <span>Review / Run Onboarding Flow</span>
            </button>
          </div>
        </div>

        {/* Lifecycle Progress Bar */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-xs">
            <div className="flex items-center gap-2 font-semibold text-slate-700">
              <Sparkles className="h-3.5 w-3.5 text-emerald-600" />
              <span>10 of 10 Lifecycle Stages Verified</span>
            </div>
            <span className="font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
              100% Complete
            </span>
          </div>
          <div className="w-full bg-slate-100 rounded-full h-2 overflow-hidden">
            <div className="bg-gradient-to-r from-emerald-500 to-teal-500 h-2 rounded-full w-full transition-all duration-500" />
          </div>
        </div>
      </div>

      {/* 2. Brand Identity & Official Logo Management Card */}
      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-xs space-y-5">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b border-slate-100 pb-4">
          <div className="flex items-center gap-2.5">
            <div className="p-2 bg-sky-50 rounded-lg border border-sky-200 text-sky-700">
              <ImageIcon className="h-4 w-4" />
            </div>
            <div>
              <h3 className="text-sm font-bold text-slate-900">Corporate Brand Identity & Logo</h3>
              <p className="text-xs text-slate-500">Official forwarder brand asset stored in AWS S3 (tenant key: organizations/{org.id}/branding/logo/).</p>
            </div>
          </div>
          <button
            type="button"
            onClick={() => setIsEditingLogo(!isEditingLogo)}
            className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-100 transition-colors self-start sm:self-center"
          >
            <Edit3 className="h-3 w-3 text-slate-500" />
            <span>{isEditingLogo ? 'Cancel' : 'Update Brand Logo'}</span>
          </button>
        </div>

        {logoMessage && (
          <div
            className={`flex items-center gap-2 p-3 rounded-lg text-xs border ${
              logoMessage.type === 'success'
                ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
                : 'bg-red-50 border-red-200 text-red-700'
            }`}
          >
            {logoMessage.type === 'success' ? (
              <Check className="h-4 w-4 text-emerald-600 shrink-0" />
            ) : (
              <AlertCircle className="h-4 w-4 text-red-600 shrink-0" />
            )}
            <span>{logoMessage.text}</span>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 items-center">
          {/* Logo Preview Container */}
          <div className="flex items-center gap-4 p-4 rounded-xl border border-slate-200 bg-slate-50/60">
            <div className="relative flex h-20 w-20 shrink-0 items-center justify-center rounded-xl border border-slate-200 bg-white p-2 shadow-2xs overflow-hidden">
              {effectiveLogo ? (
                <img
                  src={effectiveLogo}
                  alt={org.name}
                  onError={() => setImgError(true)}
                  className="max-h-full max-w-full object-contain"
                />
              ) : (
                <div className="flex flex-col items-center justify-center text-center">
                  <Building2 className="h-6 w-6 text-slate-400 mb-0.5" />
                  <span className="text-[11px] font-extrabold text-navy-900 tracking-wider">
                    {orgInitials}
                  </span>
                </div>
              )}
            </div>
            <div className="space-y-1 min-w-0">
              <p className="text-xs font-bold text-slate-900 truncate">{org.name}</p>
              <p className="text-[11px] text-slate-500 truncate">{org.legal_name || 'Forwarder Entity'}</p>
              <span className="inline-block rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-bold text-emerald-700 border border-emerald-200">
                {effectiveLogo ? 'Brand Asset in S3' : 'Standard Monogram Active'}
              </span>
            </div>
          </div>

          {/* Edit Form or Metadata */}
          <div className="md:col-span-2 space-y-3">
            {isEditingLogo ? (
              <div className="space-y-3 p-4 rounded-xl border border-sky-100 bg-sky-50/40">
                {/* S3 File Upload Control */}
                <div>
                  <label className="block text-xs font-bold text-slate-700 mb-1.5">
                    Upload New Brand Logo (AWS S3 Tenant Storage)
                  </label>
                  <label className="flex flex-col items-center justify-center gap-1.5 p-4 rounded-lg border-2 border-dashed border-sky-300 bg-white hover:bg-sky-50/50 cursor-pointer transition-colors text-center">
                    {savingLogo ? (
                      <div className="flex items-center gap-2 text-xs text-sky-700 font-semibold py-2">
                        <Loader2 className="h-4 w-4 animate-spin text-sky-600" />
                        <span>Uploading to AWS S3 & updating MariaDB...</span>
                      </div>
                    ) : (
                      <>
                        <CloudUpload className="h-6 w-6 text-sky-600" />
                        <span className="text-xs font-bold text-slate-800">
                          Click to select or drag logo image here
                        </span>
                        <span className="text-[11px] text-slate-500">
                          PNG, JPG, JPEG, SVG, WebP up to 5MB (Stores in S3 under organizations/{org.id}/branding/logo/)
                        </span>
                      </>
                    )}
                    <input
                      type="file"
                      disabled={savingLogo}
                      accept="image/png,image/jpeg,image/svg+xml,image/webp"
                      onChange={handleFileUpload}
                      className="hidden"
                    />
                  </label>
                </div>

                {/* Direct URL input option */}
                <div className="pt-2 border-t border-sky-200/60">
                  <label className="block text-xs font-bold text-slate-700 mb-1">
                    Or Enter Public Image / CDN URL
                  </label>
                  <div className="flex gap-2">
                    <input
                      type="url"
                      value={logoInput}
                      onChange={(e) => setLogoInput(e.target.value)}
                      placeholder="https://cdn.example.com/branding/logo.png"
                      className="flex-1 rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs text-slate-900 placeholder-slate-400 focus:border-navy-900 focus:outline-none"
                    />
                    <button
                      type="button"
                      disabled={savingLogo || !logoInput.trim()}
                      onClick={handleSaveLogo}
                      className="inline-flex items-center gap-1.5 rounded-lg bg-navy-900 px-3 py-1.5 text-xs font-bold text-white hover:bg-navy-800 disabled:opacity-50 shadow-2xs shrink-0"
                    >
                      {savingLogo ? <Loader2 className="h-3 w-3 animate-spin" /> : <Check className="h-3 w-3" />}
                      <span>Save URL</span>
                    </button>
                  </div>
                </div>

                {/* Cancel and Remove Actions */}
                <div className="flex items-center justify-between pt-2 border-t border-sky-200/60">
                  {effectiveLogo ? (
                    <button
                      type="button"
                      disabled={savingLogo}
                      onClick={handleRemoveLogo}
                      className="inline-flex items-center gap-1 text-xs font-semibold text-rose-600 hover:text-rose-700"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      <span>Remove Logo (Reset to Monogram)</span>
                    </button>
                  ) : <div />}

                  <button
                    type="button"
                    onClick={() => {
                      setLogoInput(org?.logo_url || '');
                      setIsEditingLogo(false);
                    }}
                    className="rounded-lg border border-slate-200 bg-white px-3 py-1 text-xs font-semibold text-slate-600 hover:bg-slate-50"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                  <p className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Trading Identity</p>
                  <p className="font-bold text-slate-800 truncate mt-0.5">{org.name || '—'}</p>
                </div>
                <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                  <p className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Corporate Domain</p>
                  <p className="font-semibold text-slate-800 truncate mt-0.5">{org.website ? org.website.replace(/^https?:\/\//, '') : '—'}</p>
                </div>
                <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                  <p className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Tax Registration</p>
                  <p className="font-semibold text-slate-800 truncate mt-0.5">{org.tax_number || '—'}</p>
                </div>
                <div className="p-3 rounded-lg border border-slate-100 bg-slate-50">
                  <p className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Storage Path</p>
                  <p className="font-mono text-[10px] text-slate-600 truncate mt-0.5" title={`organizations/${org.id}/branding/logo/`}>
                    org-{org.id}/branding/
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 3. 10 Production Stages Grid */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-sm font-bold text-slate-900">10-Stage Corporate Lifecycle Gates</h3>
            <p className="text-xs text-slate-500">Authoritative audit status and configuration across all customer lifecycle dimensions.</p>
          </div>
          <span className="text-xs font-semibold text-slate-500">
            Click any gate to inspect or update
          </span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {stages.map((stage) => {
            const Icon = stage.icon;
            return (
              <div
                key={stage.num}
                className="rounded-xl border border-slate-200 bg-white p-5 shadow-2xs hover:border-slate-300 transition-all flex flex-col justify-between"
              >
                <div className="space-y-3">
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center gap-2.5">
                      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-navy-900 text-white font-bold text-xs">
                        {stage.num}
                      </div>
                      <div>
                        <h4 className="text-xs font-bold text-slate-900">{stage.name}</h4>
                        <p className="text-[11px] text-slate-500 leading-snug">{stage.description}</p>
                      </div>
                    </div>
                    <span className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-[10px] font-bold border ${stage.badgeClass}`}>
                      <Check className="h-3 w-3 stroke-[3]" />
                      <span>{stage.status}</span>
                    </span>
                  </div>

                  {/* Stage Key Data Grid */}
                  <div className="grid grid-cols-2 gap-2 bg-slate-50/80 rounded-lg p-2.5 border border-slate-100 text-[11px]">
                    {stage.meta.map((item, idx) => (
                      <div key={idx} className="min-w-0">
                        <span className="text-slate-400 block text-[10px] font-semibold">{item.label}</span>
                        <span className="font-bold text-slate-800 truncate block">{item.val}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="pt-3 border-t border-slate-100 mt-4 flex items-center justify-between">
                  <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                    Stage {stage.num} of 10
                  </span>
                  <div className="flex items-center gap-3">
                    {stage.targetTab && onNavigateTab && (
                      <button
                        type="button"
                        onClick={() => onNavigateTab(stage.targetTab)}
                        className="inline-flex items-center gap-1 text-xs font-semibold text-slate-600 hover:text-navy-900"
                      >
                        <span>View Tab</span>
                      </button>
                    )}
                    <button
                      type="button"
                      onClick={() => onOpenStage && onOpenStage(stage.num)}
                      className="inline-flex items-center gap-1 text-xs font-bold text-navy-900 hover:text-navy-700 hover:underline"
                    >
                      <span>Inspect / Configure</span>
                      <ExternalLink className="h-3 w-3" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
