import React, { useState, useEffect } from 'react';
import { 
  FileSpreadsheet, Plus, Star, DollarSign, Clock, ShieldAlert, 
  Trash2, Edit3, CheckCircle2, AlertCircle, RefreshCw, Layers,
  X, ChevronDown, Bookmark, Sparkles, FileText
} from 'lucide-react';
import contractsService from '../../../services/contractsService';
import ConfirmModal from '../Settings/ConfirmModal';
import toast from 'react-hot-toast';
import './ContractTermsPanel.css';

export default function ContractTermsPanel({ contract, versions = [] }) {
  const [terms, setTerms] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showModal, setShowModal] = useState(false);
  const [editingTerm, setEditingTerm] = useState(null);
  const [deletingTermId, setDeletingTermId] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // Form State
  const [formData, setFormData] = useState({
    term_category: 'COMMERCIAL',
    term_title: '',
    term_value: '',
    value_type: 'STRING',
    currency: 'USD',
    effective_date: '',
    expiry_date: '',
    display_order: 0,
    is_critical: false,
  });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (contract?.id) {
      fetchTerms();
    }
  }, [contract?.id]);

  const fetchTerms = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await contractsService.getContractTerms(contract.id);
      setTerms(res.data?.data || []);
    } catch (err) {
      console.error('Failed to load contract terms:', err);
      setError('Unable to load commercial terms.');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenAdd = () => {
    setEditingTerm(null);
    setFormData({
      term_category: 'COMMERCIAL',
      term_title: '',
      term_value: '',
      value_type: 'STRING',
      currency: contract?.currency || 'USD',
      effective_date: contract?.effective_date || '',
      expiry_date: contract?.expiry_date || '',
      display_order: terms.length + 1,
      is_critical: false,
    });
    setShowModal(true);
  };

  const handleOpenEdit = (term) => {
    setEditingTerm(term);
    setFormData({
      term_category: term.term_category,
      term_title: term.term_title,
      term_value: term.term_value,
      value_type: term.value_type || 'STRING',
      currency: term.currency || 'USD',
      effective_date: term.effective_date || '',
      expiry_date: term.expiry_date || '',
      display_order: term.display_order || 0,
      is_critical: term.is_critical || false,
    });
    setShowModal(true);
  };

  const handleSaveTerm = async (e) => {
    e.preventDefault();
    if (!formData.term_title || !formData.term_value) return;

    try {
      setSaving(true);
      if (editingTerm) {
        await contractsService.updateContractTerm(contract.id, editingTerm.id, formData);
      } else {
        await contractsService.createContractTerm(contract.id, formData);
        toast.success('Contract term added successfully');
      }
      setShowModal(false);
      await fetchTerms();
    } catch (err) {
      console.error('Failed to save term:', err);
      toast.error(err.response?.data?.error?.message || 'Failed to save term. Please verify input data.');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteTerm = (termId) => {
    setDeletingTermId(termId);
  };

  const handleConfirmDeleteTerm = async () => {
    if (!deletingTermId) return;
    setIsDeleting(true);
    try {
      await contractsService.deleteContractTerm(contract.id, deletingTermId);
      toast.success('Contract term deleted');
      setDeletingTermId(null);
      await fetchTerms();
    } catch (err) {
      console.error('Failed to delete term:', err);
      toast.error(err.response?.data?.error?.message || 'Failed to delete term.');
    } finally {
      setIsDeleting(false);
    }
  };

  // Group terms by category
  const categories = [
    { key: 'COMMERCIAL', label: 'Commercial & Financial' },
    { key: 'PAYMENT', label: 'Payment Terms & Credit' },
    { key: 'PRICING', label: 'Pricing & Surcharges' },
    { key: 'SERVICE_LEVEL', label: 'Service Level Agreements (SLA)' },
    { key: 'LIABILITY', label: 'Liability & Claims' },
    { key: 'TERMINATION', label: 'Termination & Cancellation' },
    { key: 'RENEWAL', label: 'Renewal & Extension' },
    { key: 'COMPLIANCE', label: 'Regulatory & Compliance' },
    { key: 'OPERATIONAL', label: 'Operational & Lane Rules' },
    { key: 'OTHER', label: 'General / Other Terms' },
  ];

  const criticalTerms = terms.filter(t => t.is_critical);

  return (
    <div className="contract-terms-panel">
      {/* Header Bar */}
      <div className="panel-header-row">
        <div>
          <h3 className="panel-title">Important Terms & Commercial Clauses</h3>
          <p className="panel-subtitle">
            Structured commercial agreements, payment schedules, free days, liability limits, and SLA commitments.
          </p>
        </div>
        <button className="btn-primary-action" onClick={handleOpenAdd}>
          <Plus size={16} />
          <span>Add Key Term</span>
        </button>
      </div>

      {loading ? (
        <div className="panel-loading-state">
          <RefreshCw className="spin-icon" size={24} />
          <span>Loading structured terms...</span>
        </div>
      ) : error ? (
        <div className="panel-error-banner">
          <AlertCircle size={18} />
          <span>{error}</span>
          <button className="btn-retry" onClick={fetchTerms}>Retry</button>
        </div>
      ) : terms.length === 0 ? (
        <div className="panel-empty-state">
          <div className="empty-icon-circle">
            <FileSpreadsheet size={32} />
          </div>
          <h4>No Key Terms Defined</h4>
          <p>Add key structured terms such as Credit Days, Free Detention Days, Penalty Clauses, and Minimum Commitments.</p>
          <button className="btn-primary-action" onClick={handleOpenAdd}>
            <Plus size={16} />
            <span>Add First Commercial Term</span>
          </button>
        </div>
      ) : (
        <div className="terms-container">
          {/* Critical Terms Spotlight */}
          {criticalTerms.length > 0 && (
            <div className="critical-terms-spotlight">
              <div className="spotlight-header">
                <Star size={16} className="text-amber-500 fill-amber-500" />
                <h4>Critical Governance & Liability Terms ({criticalTerms.length})</h4>
              </div>
              <div className="critical-terms-grid">
                {criticalTerms.map(term => (
                  <div key={term.id} className="critical-term-card">
                    <div className="term-card-header">
                      <span className="term-category-badge">{term.term_category}</span>
                      <div className="term-actions">
                        <button className="btn-icon" onClick={() => handleOpenEdit(term)} title="Edit term"><Edit3 size={14} /></button>
                        <button className="btn-icon delete" onClick={() => handleDeleteTerm(term.id)} title="Delete term"><Trash2 size={14} /></button>
                      </div>
                    </div>
                    <h5 className="term-title">{term.term_title}</h5>
                    <div className="term-value-display highlight">
                      {term.term_value} {term.currency && <span className="term-curr">({term.currency})</span>}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Grouped Terms by Category */}
          <div className="terms-categories-list">
            {categories.map(cat => {
              const catTerms = terms.filter(t => t.term_category === cat.key);
              if (catTerms.length === 0) return null;

              return (
                <div key={cat.key} className="category-group-block">
                  <h4 className="category-title">{cat.label} ({catTerms.length})</h4>
                  <div className="category-terms-grid">
                    {catTerms.map(term => (
                      <div key={term.id} className={`term-card ${term.is_critical ? 'is-critical-border' : ''}`}>
                        <div className="term-card-header">
                          <div className="term-title-wrapper">
                            {term.is_critical && <Star size={13} className="text-amber-500 fill-amber-500 mr-1" />}
                            <h5 className="term-title">{term.term_title}</h5>
                          </div>
                          <div className="term-actions">
                            <button className="btn-icon" onClick={() => handleOpenEdit(term)} title="Edit term"><Edit3 size={13} /></button>
                            <button className="btn-icon delete" onClick={() => handleDeleteTerm(term.id)} title="Delete term"><Trash2 size={13} /></button>
                          </div>
                        </div>

                        <div className="term-value-display">
                          {term.term_value}
                        </div>

                        {(term.effective_date || term.expiry_date) && (
                          <div className="term-dates-footer">
                            <Clock size={12} />
                            <span>
                              {term.effective_date ? `From ${term.effective_date}` : ''}
                              {term.expiry_date ? ` until ${term.expiry_date}` : ''}
                            </span>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Add / Edit Modal */}
      {showModal && (
        <div className="modal-backdrop" onClick={() => setShowModal(false)}>
          <div className="term-modal-content" onClick={e => e.stopPropagation()}>
            <div className="term-modal-header">
              <div className="term-modal-header-left">
                <div className="term-header-badge">
                  <Bookmark size={18} />
                </div>
                <div>
                  <h4>{editingTerm ? 'Edit Key Term' : 'Add Important Contract Term'}</h4>
                  <p className="term-modal-subtitle">Define structured commercial clauses, operational covenants, and governing terms</p>
                </div>
              </div>
              <button className="btn-close-modal" onClick={() => setShowModal(false)} title="Close">
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleSaveTerm} className="term-form">
              <div className="term-form-group">
                <label className="term-label">
                  Category <span className="term-req">*</span>
                </label>
                <div className="term-select-wrapper">
                  <select 
                    className="term-select"
                    value={formData.term_category}
                    onChange={(e) => setFormData({ ...formData, term_category: e.target.value })}
                    required
                  >
                    {categories.map(c => (
                      <option key={c.key} value={c.key}>{c.label}</option>
                    ))}
                  </select>
                  <ChevronDown size={14} className="term-select-chevron" />
                </div>
              </div>

              <div className="term-form-group">
                <label className="term-label">
                  Term Title <span className="term-req">*</span>
                </label>
                <input 
                  type="text"
                  className="term-input"
                  placeholder="e.g. Free Detention Days (Origin & Destination)"
                  value={formData.term_title}
                  onChange={(e) => setFormData({ ...formData, term_title: e.target.value })}
                  required
                />
              </div>

              <div className="term-form-group">
                <label className="term-label">
                  Term Value / Clause Specification <span className="term-req">*</span>
                </label>
                <textarea 
                  rows={3}
                  className="term-textarea"
                  placeholder="e.g. 14 Days free time at discharge port for 40' HC containers; standard demurrage thereafter."
                  value={formData.term_value}
                  onChange={(e) => setFormData({ ...formData, term_value: e.target.value })}
                  required
                />
              </div>

              <div className="term-form-row-2">
                <div className="term-form-group">
                  <label className="term-label">Value Type</label>
                  <div className="term-select-wrapper">
                    <select 
                      className="term-select"
                      value={formData.value_type}
                      onChange={(e) => setFormData({ ...formData, value_type: e.target.value })}
                    >
                      <option value="STRING">Text Clause</option>
                      <option value="NUMBER">Numeric Value</option>
                      <option value="CURRENCY">Monetary / Currency</option>
                      <option value="PERCENTAGE">Percentage (%)</option>
                      <option value="DATE">Calendar Date</option>
                    </select>
                    <ChevronDown size={14} className="term-select-chevron" />
                  </div>
                </div>

                <div className="term-form-group">
                  <label className="term-label">Currency (if applicable)</label>
                  <input 
                    type="text"
                    className="term-input"
                    placeholder="e.g. USD"
                    value={formData.currency}
                    onChange={(e) => setFormData({ ...formData, currency: e.target.value })}
                  />
                </div>
              </div>

              <div className="term-form-row-2">
                <div className="term-form-group">
                  <label className="term-label">Effective Date</label>
                  <input 
                    type="date"
                    className="term-input term-input-date"
                    value={formData.effective_date}
                    onChange={(e) => setFormData({ ...formData, effective_date: e.target.value })}
                  />
                </div>
                <div className="term-form-group">
                  <label className="term-label">Expiry Date</label>
                  <input 
                    type="date"
                    className="term-input term-input-date"
                    value={formData.expiry_date}
                    onChange={(e) => setFormData({ ...formData, expiry_date: e.target.value })}
                  />
                </div>
              </div>

              <div className="term-checkbox-card">
                <label className="term-checkbox-label">
                  <input 
                    type="checkbox"
                    checked={formData.is_critical}
                    onChange={(e) => setFormData({ ...formData, is_critical: e.target.checked })}
                  />
                  <div>
                    <span className="checkbox-title">Mark as Critical Governance / High-Risk Term</span>
                    <span className="checkbox-sub">Highlights term with priority gold badge in executive contract snapshots</span>
                  </div>
                </label>
              </div>

              <div className="modal-actions-bar">
                <button type="button" className="btn-cancel" onClick={() => setShowModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-submit" disabled={saving}>
                  <CheckCircle2 size={14} />
                  <span>{saving ? 'Saving...' : editingTerm ? 'Update Term' : 'Save Term'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Term Confirmation Modal */}
      <ConfirmModal
        isOpen={Boolean(deletingTermId)}
        title="Delete Commercial Contract Term"
        message="Are you sure you want to delete this contract term? This clause will be removed from future contract versions."
        confirmText="Delete Term"
        confirmStyle="danger"
        isLoading={isDeleting}
        onConfirm={handleConfirmDeleteTerm}
        onCancel={() => setDeletingTermId(null)}
      />
    </div>
  );
}
