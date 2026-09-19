import React, { useState, useEffect } from 'react';
import { X, Building2, AlertCircle, CheckCircle2, Loader2 } from 'lucide-react';
import { sportalService } from '../../services/sportalService';

export function EditOrganizationModal({ isOpen, onClose, organization, onSuccess }) {
  const [formData, setFormData] = useState({
    name: '',
    legal_name: '',
    registration_number: '',
    tax_number: '',
    primary_email: '',
    phone_number: '',
    website: '',
    address: '',
    city: '',
    state: '',
    country: '',
    postal_code: '',
    industry: '',
    company_type: '',
    logo_url: '',
  });

  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');

  useEffect(() => {
    if (organization) {
      setFormData({
        name: organization.name || '',
        legal_name: organization.legal_name || '',
        registration_number: organization.registration_number || '',
        tax_number: organization.tax_number || '',
        primary_email: organization.primary_email || '',
        phone_number: organization.phone_number || '',
        website: organization.website || '',
        address: organization.address || '',
        city: organization.city || '',
        state: organization.state || '',
        country: organization.country || '',
        postal_code: organization.postal_code || '',
        industry: organization.industry || '',
        company_type: organization.company_type || '',
        logo_url: organization.logo_url || '',
      });
      setErrorMessage('');
    }
  }, [organization]);

  if (!isOpen || !organization) return null;

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (errorMessage) setErrorMessage('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.name.trim()) {
      setErrorMessage('Company trading name cannot be empty');
      return;
    }

    setLoading(true);
    setErrorMessage('');

    try {
      const updated = await sportalService.updateOrganization(organization.id, formData);
      onSuccess?.(updated);
      onClose();
    } catch (err) {
      setErrorMessage(err.message || 'Failed to update organization profile.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-xs overflow-y-auto">
      <div className="relative w-full max-w-2xl rounded-xl bg-white shadow-2xl border border-slate-200 my-8 overflow-hidden">
        {/* Modal Header */}
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4 bg-slate-50/70">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-navy-900 text-white shadow-xs">
              <Building2 className="h-5 w-5 text-sky-400" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-slate-900 tracking-tight">Edit Organization Profile</h3>
              <p className="text-xs text-slate-500">Update company credentials, contact info, and tax identity for ID #{organization.id}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            disabled={loading}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Error Alert */}
        {errorMessage && (
          <div className="mx-6 mt-4 flex items-start gap-3 rounded-lg border border-red-200 bg-red-50/80 p-3 text-sm text-red-700">
            <AlertCircle className="h-5 w-5 shrink-0 text-red-500 mt-0.5" />
            <div>
              <p className="font-semibold">Update Conflict / Validation Notice</p>
              <p className="text-xs mt-0.5 text-red-600">{errorMessage}</p>
            </div>
          </div>
        )}

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                Company Trading Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                name="name"
                required
                value={formData.name}
                onChange={handleChange}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                Legal Company Name
              </label>
              <input
                type="text"
                name="legal_name"
                value={formData.legal_name}
                onChange={handleChange}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                Primary Contact Email
              </label>
              <input
                type="email"
                name="primary_email"
                value={formData.primary_email}
                onChange={handleChange}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                Phone Number
              </label>
              <input
                type="text"
                name="phone_number"
                value={formData.phone_number}
                onChange={handleChange}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                GSTIN / Tax ID Number
              </label>
              <input
                type="text"
                name="tax_number"
                value={formData.tax_number}
                onChange={handleChange}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all uppercase"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                Registration / CIN Number
              </label>
              <input
                type="text"
                name="registration_number"
                value={formData.registration_number}
                onChange={handleChange}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                Website
              </label>
              <input
                type="url"
                name="website"
                value={formData.website}
                onChange={handleChange}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                Industry
              </label>
              <input
                type="text"
                name="industry"
                value={formData.industry}
                onChange={handleChange}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
                Company Logo URL / S3 Asset Key
              </label>
              <input
                type="text"
                name="logo_url"
                value={formData.logo_url}
                onChange={handleChange}
                placeholder="https://... or s3://..."
                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 focus:ring-1 focus:ring-navy-900 outline-none transition-all"
              />
            </div>
          </div>

          {/* Address Details */}
          <div className="border-t border-slate-100 pt-3">
            <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">Registered Address</h4>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div>
                <input
                  type="text"
                  name="city"
                  value={formData.city}
                  onChange={handleChange}
                  placeholder="City"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 outline-none"
                />
              </div>
              <div>
                <input
                  type="text"
                  name="state"
                  value={formData.state}
                  onChange={handleChange}
                  placeholder="State"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 outline-none"
                />
              </div>
              <div>
                <input
                  type="text"
                  name="country"
                  value={formData.country}
                  onChange={handleChange}
                  placeholder="Country"
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 outline-none"
                />
              </div>
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1">
              Company Logo URL / Asset Reference
            </label>
            <input
              type="text"
              name="logo_url"
              value={formData.logo_url}
              onChange={handleChange}
              placeholder="e.g. https://... or organizations/2/branding/logo/..."
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:border-navy-900 outline-none"
            />
          </div>

          {/* Modal Footer */}
          <div className="flex items-center justify-end gap-3 border-t border-slate-200 pt-4 mt-6">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex items-center gap-2 rounded-lg bg-navy-900 px-5 py-2 text-sm font-medium text-white hover:bg-navy-800 shadow-xs transition-colors disabled:opacity-50"
            >
              {loading ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin text-sky-400" />
                  <span>Saving Updates...</span>
                </>
              ) : (
                <>
                  <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                  <span>Save Changes</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
