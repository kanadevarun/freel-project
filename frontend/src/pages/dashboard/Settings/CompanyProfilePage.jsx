import React, { useState, useEffect, useRef, useMemo } from 'react';
import { useRBAC } from '../../../context/RBACContext';
import api from '../../../services/api';
import toast from 'react-hot-toast';
import { COUNTRIES, getFlagEmoji } from '../../../utils/countries';
import './CompanyProfilePage.css';

// ── Helpers ──────────────────────────────────────────────────────────────────
const toStr = (v) => (v == null ? '' : String(v));
const nullify = (v) => (typeof v === 'string' && v.trim() === '' ? null : typeof v === 'string' ? v.trim() : v);

function mapOrg(org) {
  return {
    name:                toStr(org.name),
    legal_name:          toStr(org.legal_name),
    registration_number: toStr(org.registration_number),
    tax_number:          toStr(org.tax_number),
    website:             toStr(org.website),
    primary_email:       toStr(org.primary_email),
    phone_number:        toStr(org.phone_number),
    support_email:       toStr(org.support_email),
    address:             toStr(org.address),
    city:                toStr(org.city),
    state:               toStr(org.state),
    country:             toStr(org.country),
    postal_code:         toStr(org.postal_code),
    industry:            toStr(org.industry),
    company_type:        toStr(org.company_type),
    default_currency:    toStr(org.default_currency) || 'USD',
    default_timezone:    toStr(org.default_timezone) || 'UTC',
    date_format:         toStr(org.date_format) || 'MM/DD/YYYY',
    logo_url:            toStr(org.logo_url),
  };
}

// ── Component ─────────────────────────────────────────────────────────────────
export default function CompanyProfilePage() {
  const { can } = useRBAC();
  const canEdit = can('SETTINGS', 'UPDATE');

  const [loading,   setLoading]   = useState(true);
  const [saving,    setSaving]    = useState(false);
  const [uploading, setUploading] = useState(false);

  const savedRef    = useRef(null);
  const fileInputRef = useRef(null);
  const [form, setForm] = useState({
    name: '', legal_name: '', registration_number: '', tax_number: '',
    website: '', primary_email: '', phone_number: '', support_email: '',
    address: '', city: '', state: '', country: '', postal_code: '',
    industry: '', company_type: '',
    default_currency: 'USD', default_timezone: 'UTC', date_format: 'MM/DD/YYYY', logo_url: '',
  });

  // Dirty = form differs from last-saved DB snapshot
  const isDirty = useMemo(() => {
    if (!savedRef.current) return false;
    return Object.keys(savedRef.current).some(k => form[k] !== savedRef.current[k]);
  }, [form]);

  // ── Load ──────────────────────────────────────────────────────────────────
  useEffect(() => {
    (async () => {
      try {
        const res   = await api.get('/api/v1/organizations/profile');
        const mapped = mapOrg(res ?? {});
        setForm(mapped);
        savedRef.current = mapped;
      } catch {
        toast.error('Failed to load company profile');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  // ── Handlers ──────────────────────────────────────────────────────────────
  const onChange = (e) => {
    const { name, value } = e.target;
    setForm(p => ({ ...p, [name]: value }));
  };

  const onCancel = () => {
    if (savedRef.current) setForm(savedRef.current);
  };

  const onSave = async () => {
    if (!canEdit || saving || !isDirty) return;
    if (!form.name.trim()) { toast.error('Company Name is required'); return; }

    const payload = {
      name: form.name.trim(),
      legal_name: nullify(form.legal_name),
      registration_number: nullify(form.registration_number),
      tax_number: nullify(form.tax_number),
      website: nullify(form.website),
      primary_email: nullify(form.primary_email),
      phone_number: nullify(form.phone_number),
      support_email: nullify(form.support_email),
      address: nullify(form.address),
      city: nullify(form.city),
      state: nullify(form.state),
      country: nullify(form.country),
      postal_code: nullify(form.postal_code),
      industry: nullify(form.industry),
      company_type: nullify(form.company_type),
      default_currency: form.default_currency || 'USD',
      default_timezone: form.default_timezone || 'UTC',
      date_format: form.date_format || 'MM/DD/YYYY',
    };

    try {
      setSaving(true);
      const res    = await api.put('/api/v1/organizations/profile', payload);
      const mapped = mapOrg(res ?? {});
      setForm(mapped);
      savedRef.current = mapped;
      toast.success('Company profile saved');
    } catch (err) {
      toast.error(err?.message || 'Failed to save changes');
    } finally {
      setSaving(false);
    }
  };

  const onLogoFile = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    e.target.value = '';
    if (file.size > 2 * 1024 * 1024) { toast.error('Logo must be under 2 MB'); return; }
    if (!['image/jpeg','image/png','image/svg+xml'].includes(file.type)) {
      toast.error('Only JPG, PNG or SVG allowed'); return;
    }
    const fd = new FormData();
    fd.append('logo', file);
    try {
      setUploading(true);
      const tid = toast.loading('Uploading logo…');
      const res  = await api.post('/api/v1/organizations/profile/logo', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      const url = toStr((res ?? {}).logo_url);
      setForm(p => ({ ...p, logo_url: url }));
      savedRef.current = { ...savedRef.current, logo_url: url };
      toast.success('Logo updated', { id: tid });
    } catch {
      toast.error('Failed to upload logo');
    } finally {
      setUploading(false);
    }
  };

  const initials = form.name
    ? form.name.trim().split(/\s+/).slice(0, 2).map(w => w[0]).join('').toUpperCase()
    : '?';

  // ── Skeleton ──────────────────────────────────────────────────────────────
  if (loading) {
    return (
      <div className="cpg">
        <div className="cpg-skeleton-hdr" />
        <div className="cpg-card">
          <div className="cpg-skeleton-row" />
          <div className="cpg-skeleton-row short" />
          <div className="cpg-skeleton-row" />
        </div>
        <div className="cpg-card">
          <div className="cpg-skeleton-row short" />
          <div className="cpg-skeleton-row" />
        </div>
      </div>
    );
  }

  // ── Render ────────────────────────────────────────────────────────────────
  return (
    <div className="cpg">

      {/* ── Page Header ─────────────────────────────────────────────────── */}
      <div className="cpg-header">
        <div className="cpg-header-left">
          <h1 className="cpg-title">Company Profile</h1>
          <p className="cpg-subtitle">Manage your organization's information and preferences.</p>
        </div>

        {canEdit && (
          <div className="cpg-header-actions">
            <button
              className={`cpg-btn cpg-btn-cancel ${!isDirty ? 'cpg-btn-disabled' : ''}`}
              onClick={onCancel}
              disabled={!isDirty || saving}
              type="button"
            >
              Cancel
            </button>
            <button
              className={`cpg-btn cpg-btn-save ${!isDirty ? 'cpg-btn-save-idle' : ''}`}
              onClick={onSave}
              disabled={!isDirty || saving}
              type="button"
              aria-busy={saving}
            >
              {saving && <span className="cpg-spin" />}
              {saving ? 'Saving…' : 'Save Changes'}
            </button>
          </div>
        )}
      </div>

      {/* Read-only notice */}
      {!canEdit && (
        <div className="cpg-notice">
          <svg width="15" height="15" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          You have read-only access. Contact your Super Admin to make edits.
        </div>
      )}

      {/* ══ Company Information ══════════════════════════════════════════ */}
      <div className="cpg-card">
        <div className="cpg-card-header">
          <div className="cpg-card-icon">
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.2" viewBox="0 0 24 24">
              <path d="M3 21h18M3 7h18M3 14h18M6 21V7m6 14V7m6 14V7"/>
            </svg>
          </div>
          <h2 className="cpg-card-title">Company Information</h2>
        </div>

        {/* Logo */}
        <div className="cpg-logo-row">
          <div className="cpg-logo-box">
            {form.logo_url
              ? <img src={form.logo_url} alt="logo" className="cpg-logo-img" />
              : <span className="cpg-logo-letters">{initials}</span>
            }
          </div>
          <div>
            {canEdit && (
              <button
                className="cpg-logo-btn"
                onClick={() => fileInputRef.current?.click()}
                disabled={uploading}
                type="button"
              >
                {uploading ? 'Uploading…' : 'Change Logo'}
              </button>
            )}
            <p className="cpg-logo-hint">JPG, PNG or SVG · Max 2 MB</p>
            <input ref={fileInputRef} type="file" accept=".jpg,.jpeg,.png,.svg"
              onChange={onLogoFile} style={{ display:'none' }} />
          </div>
        </div>

        <div className="cpg-fields">
          <div className="cpg-row-2">
            <F label="Company Name" required>
              <input name="name" value={form.name} onChange={onChange} disabled={!canEdit} placeholder="Your company name" />
            </F>
            <F label="Legal / Business Name">
              <input name="legal_name" value={form.legal_name} onChange={onChange} disabled={!canEdit} placeholder="Registered legal name" />
            </F>
          </div>
          <div className="cpg-row-2">
            <F label="Registration Number">
              <input name="registration_number" value={form.registration_number} onChange={onChange} disabled={!canEdit} placeholder="e.g. U63090MH2020PTC…" />
            </F>
            <F label="Tax / VAT / GST Number">
              <input name="tax_number" value={form.tax_number} onChange={onChange} disabled={!canEdit} placeholder="e.g. 27AAACB1234C1Z5" />
            </F>
          </div>
          <div className="cpg-row-1">
            <F label="Website">
              <input name="website" type="url" value={form.website} onChange={onChange} disabled={!canEdit} placeholder="https://yourcompany.com" />
            </F>
          </div>
        </div>
      </div>

      {/* ══ Contact Information ══════════════════════════════════════════ */}
      <div className="cpg-card">
        <div className="cpg-card-header">
          <div className="cpg-card-icon">
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.2" viewBox="0 0 24 24">
              <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
              <polyline points="22,6 12,13 2,6"/>
            </svg>
          </div>
          <h2 className="cpg-card-title">Contact Information</h2>
        </div>

        <div className="cpg-fields">
          <div className="cpg-row-2">
            <F label="Primary Email">
              <input name="primary_email" type="email" value={form.primary_email} onChange={onChange} disabled={!canEdit} placeholder="info@yourcompany.com" />
            </F>
            <F label="Phone Number">
              <input name="phone_number" type="tel" value={form.phone_number} onChange={onChange} disabled={!canEdit} placeholder="+91 98765 43210" />
            </F>
          </div>
          <div className="cpg-row-1">
            <F label="Support Email">
              <input name="support_email" type="email" value={form.support_email} onChange={onChange} disabled={!canEdit} placeholder="support@yourcompany.com" />
            </F>
          </div>
          <div className="cpg-row-1">
            <F label="Address">
              <input name="address" value={form.address} onChange={onChange} disabled={!canEdit} placeholder="Street address, floor, building" />
            </F>
          </div>
          <div className="cpg-row-4">
            <F label="City">
              <input name="city" value={form.city} onChange={onChange} disabled={!canEdit} placeholder="City" />
            </F>
            <F label="State / Province">
              <input name="state" value={form.state} onChange={onChange} disabled={!canEdit} placeholder="State" />
            </F>
            <F label="Country">
              <CustomSelect 
                name="country" 
                value={form.country} 
                onChange={onChange} 
                disabled={!canEdit} 
                placeholder="Select country"
                options={COUNTRIES.map(c => ({
                  value: c.label,
                  label: `${getFlagEmoji(c.iso)} ${c.label}`
                }))}
              />
            </F>
            <F label="Postal / ZIP Code">
              <input name="postal_code" value={form.postal_code} onChange={onChange} disabled={!canEdit} placeholder="PIN / ZIP" />
            </F>
          </div>
        </div>
      </div>

      {/* ══ Business Details ═════════════════════════════════════════════ */}
      <div className="cpg-card" style={{ marginBottom: 0 }}>
        <div className="cpg-card-header">
          <div className="cpg-card-icon">
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.2" viewBox="0 0 24 24">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
          </div>
          <h2 className="cpg-card-title">Business Details</h2>
        </div>

        <div className="cpg-fields">
          <div className="cpg-row-2">
            <F label="Industry">
              <CustomSelect 
                name="industry" 
                value={form.industry} 
                onChange={onChange} 
                disabled={!canEdit} 
                placeholder="Select industry"
                options={[
                  { value: "Freight Forwarding", label: "Freight Forwarding" },
                  { value: "NVOCC", label: "NVOCC" },
                  { value: "Customs Brokerage", label: "Customs Brokerage" },
                  { value: "Logistics Provider", label: "Logistics Provider (3PL/4PL)" },
                  { value: "Carrier", label: "Carrier (Ocean/Air)" },
                  { value: "Other", label: "Other" }
                ]}
              />
            </F>
            <F label="Company Type">
              <CustomSelect 
                name="company_type" 
                value={form.company_type} 
                onChange={onChange} 
                disabled={!canEdit}
                placeholder="Select type"
                options={[
                  { value: "Private Limited", label: "Private Limited" },
                  { value: "Public Limited", label: "Public Limited" },
                  { value: "LLC", label: "LLC" },
                  { value: "Partnership", label: "Partnership" },
                  { value: "Sole Proprietorship", label: "Sole Proprietorship" }
                ]}
              />
            </F>
          </div>
          <div className="cpg-row-2">
            <F label="Default Timezone">
              <CustomSelect 
                name="default_timezone" 
                value={form.default_timezone} 
                onChange={onChange} 
                disabled={!canEdit}
                options={[
                  { value: "UTC", label: "UTC" },
                  { value: "Asia/Kolkata", label: "Asia/Kolkata (IST +5:30)" },
                  { value: "Asia/Dubai", label: "Asia/Dubai (GST +4:00)" },
                  { value: "Asia/Singapore", label: "Asia/Singapore (SGT +8:00)" },
                  { value: "Europe/London", label: "Europe/London (GMT/BST)" },
                  { value: "America/New_York", label: "America/New_York (EST/EDT)" },
                  { value: "America/Los_Angeles", label: "America/Los_Angeles (PST/PDT)" }
                ]}
              />
            </F>
            <F label="Default Currency">
              <CustomSelect 
                name="default_currency" 
                value={form.default_currency} 
                onChange={onChange} 
                disabled={!canEdit}
                options={[
                  { value: "USD", label: "USD – US Dollar" },
                  { value: "EUR", label: "EUR – Euro" },
                  { value: "GBP", label: "GBP – British Pound" },
                  { value: "INR", label: "INR – Indian Rupee" },
                  { value: "SGD", label: "SGD – Singapore Dollar" },
                  { value: "AED", label: "AED – UAE Dirham" },
                  { value: "CNY", label: "CNY – Chinese Yuan" }
                ]}
              />
            </F>
          </div>
        </div>
      </div>

    </div>
  );
}

// ── Custom Select Component ───────────────────────────────────────────────────
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
    <div className={`cpg-custom-dropdown ${disabled ? 'disabled' : ''}`} ref={containerRef}>
      <div 
        className="cpg-custom-dropdown-selected" 
        onClick={() => !disabled && setIsOpen(!isOpen)}
      >
        {selectedOption ? selectedOption.label : <span className="placeholder">{placeholder || 'Select...'}</span>}
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </div>
      {isOpen && !disabled && (
        <ul className="cpg-custom-dropdown-list">
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

// ── Field wrapper ─────────────────────────────────────────────────────────────
function F({ label, required, children }) {
  return (
    <div className="cpg-field">
      <label className="cpg-label">
        {label}{required && <span className="cpg-req"> *</span>}
      </label>
      {children}
    </div>
  );
}
