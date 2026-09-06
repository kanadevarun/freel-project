import React, { useState, useEffect, useRef } from 'react';
import { 
  X, Save, FileText, Building2, Calendar, 
  DollarSign, Truck, Plane, Ship, Layers, ShieldCheck,
  ChevronDown, CheckCircle, Hash, Users, Sparkles, AlertCircle
} from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import toast from 'react-hot-toast';
import './ContractForm.css';

const PRESET_PARTIES = [
  { id: 1, name: 'Maersk Line', type: 'Carrier', icon: <Truck size={14} className="text-blue-600" />, badge: 'CARRIER' },
  { id: 2, name: 'Acme Corp Industries', type: 'Customer', icon: <Users size={14} className="text-indigo-600" />, badge: 'CUSTOMER' },
  { id: 3, name: 'Apex Drayage & Intermodal', type: 'Vendor', icon: <Building2 size={14} className="text-purple-600" />, badge: 'VENDOR' },
  { id: 4, name: 'Cargolux Airlines', type: 'Carrier', icon: <Plane size={14} className="text-amber-600" />, badge: 'CARRIER' },
  { id: 'custom', name: 'Other / Custom Party ID', type: 'Custom', icon: <Hash size={14} className="text-slate-600" />, badge: 'CUSTOM' }
];

const CONTRACT_TYPES = [
  { value: 'CARRIER_AGREEMENT', label: 'Carrier Agreement (Supply)', sub: 'Ocean / Air rate agreement & committed allocations', badge: 'SUPPLY' },
  { value: 'CUSTOMER_SLA', label: 'Customer SLA (Client Master)', sub: 'Customer pricing commitments & guaranteed transit SLAs', badge: 'CLIENT' },
  { value: 'VENDOR_CONTRACT', label: 'Vendor Agreement', sub: 'Drayage, warehouse storage, customs brokerage terms', badge: 'VENDOR' },
  { value: 'FORWARDER_PARTNERSHIP', label: 'Forwarder Partnership', sub: 'Inter-forwarder agency and co-loading agreement', badge: 'PARTNER' }
];

const TRANSPORT_MODES = [
  { value: 'OCEAN', label: 'Ocean Freight (FCL / LCL)', icon: <Ship size={14} className="text-blue-600" /> },
  { value: 'AIR', label: 'Air Cargo Express', icon: <Plane size={14} className="text-amber-600" /> },
  { value: 'ROAD', label: 'Road Transport / Drayage', icon: <Truck size={14} className="text-emerald-600" /> },
  { value: 'RAIL', label: 'Rail Intermodal', icon: <Layers size={14} className="text-purple-600" /> },
  { value: 'MULTIMODAL', label: 'Multimodal / Combined', icon: <Sparkles size={14} className="text-indigo-600" /> }
];

const CURRENCIES = [
  { value: 'USD', label: 'USD ($)', symbol: '$' },
  { value: 'EUR', label: 'EUR (€)', symbol: '€' },
  { value: 'GBP', label: 'GBP (£)', symbol: '£' },
  { value: 'SGD', label: 'SGD (S$)', symbol: 'S$' }
];

export default function ContractForm({ contract, onClose, onSuccess }) {
  const isEdit = !!contract;
  
  const [formData, setFormData] = useState({
    contract_reference: '',
    contract_name: '',
    contract_type: 'CARRIER_AGREEMENT',
    party_id: 1,
    transport_mode: 'OCEAN',
    status: 'DRAFT',
    currency: 'USD',
    contract_value: '',
    effective_date: '',
    expiry_date: '',
    owner: 'Varun Kanade',
    description: '',
    notes: ''
  });

  const [selectedPreset, setSelectedPreset] = useState(1);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Custom Dropdown Open States
  const [openDropdown, setOpenDropdown] = useState(null); // 'PARTY' | 'TYPE' | 'MODE' | 'CURRENCY' | null

  const formRef = useRef(null);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (formRef.current && !formRef.current.contains(event.target)) {
        setOpenDropdown(null);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  useEffect(() => {
    if (contract) {
      setFormData({
        contract_reference: contract.contract_reference || '',
        contract_name: contract.contract_name || '',
        contract_type: contract.contract_type || 'CARRIER_AGREEMENT',
        party_id: contract.party_id || 1,
        transport_mode: contract.transport_mode || 'OCEAN',
        status: contract.status || 'DRAFT',
        currency: contract.currency || 'USD',
        contract_value: contract.contract_value || '',
        effective_date: contract.effective_date || '',
        expiry_date: contract.expiry_date || '',
        owner: contract.owner || 'Varun Kanade',
        description: contract.description || '',
        notes: contract.notes || ''
      });

      const matchedPreset = PRESET_PARTIES.find(p => p.id === contract.party_id);
      setSelectedPreset(matchedPreset ? contract.party_id : 'custom');
    }
  }, [contract]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: name === 'contract_value' || name === 'party_id' ? Number(value) || '' : value
    }));
  };

  const handleSelectParty = (party) => {
    setSelectedPreset(party.id);
    if (party.id !== 'custom') {
      setFormData(prev => ({ ...prev, party_id: Number(party.id) }));
    }
    setOpenDropdown(null);
  };

  const handleSelectContractType = (t) => {
    setFormData(prev => ({ ...prev, contract_type: t.value }));
    setOpenDropdown(null);
  };

  const handleSelectMode = (m) => {
    setFormData(prev => ({ ...prev, transport_mode: m.value }));
    setOpenDropdown(null);
  };

  const handleSelectCurrency = (c) => {
    setFormData(prev => ({ ...prev, currency: c.value }));
    setOpenDropdown(null);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.contract_reference || !formData.contract_name || !formData.party_id) {
      toast.error('Please fill in all required fields (Contract Name, Reference, Party)');
      return;
    }

    setIsSubmitting(true);
    try {
      if (isEdit) {
        await contractsService.updateContract(contract.id, formData);
        toast.success('Contract updated successfully');
      } else {
        await contractsService.createContract(formData);
        toast.success('Contract created successfully');
      }
      onSuccess();
    } catch (err) {
      console.error(err);
      toast.error(err?.response?.data?.error?.message || err.message || 'Failed to save contract');
    } finally {
      setIsSubmitting(false);
    }
  };

  const currentParty = PRESET_PARTIES.find(p => p.id === selectedPreset) || PRESET_PARTIES[0];
  const currentType = CONTRACT_TYPES.find(t => t.value === formData.contract_type) || CONTRACT_TYPES[0];
  const currentMode = TRANSPORT_MODES.find(m => m.value === formData.transport_mode) || TRANSPORT_MODES[0];
  const currentCurrency = CURRENCIES.find(c => c.value === formData.currency) || CURRENCIES[0];

  return (
    <div className="contract-form-overlay" onClick={onClose}>
      <div className="contract-form-modal premium-form-window" ref={formRef} onClick={e => e.stopPropagation()}>
        
        {/* Modal Header */}
        <div className="cf-modal-header">
          <div className="cf-header-left">
            <div className="cf-header-icon-box">
              <FileText size={20} />
            </div>
            <div>
              <div className="cf-header-tag">COMMERCIAL GOVERNANCE</div>
              <h2 className="cf-modal-title">{isEdit ? 'Edit Commercial Contract' : 'Create New Contract'}</h2>
              <p className="cf-modal-sub">
                {isEdit ? 'Modify agreement parameters, counterparty, and financial clauses.' : 'Define commercial terms, validity windows, and counterparty bindings.'}
              </p>
            </div>
          </div>
          <button className="cf-close-btn" onClick={onClose} title="Close">
            <X size={18} />
          </button>
        </div>
        
        {/* Modal Form */}
        <form onSubmit={handleSubmit} className="cf-form">
          <div className="cf-form-body">
            
            {/* Section 1: Core Details */}
            <div className="cf-section-title">
              <ShieldCheck size={14} className="text-blue" />
              <span>Core Agreement Details</span>
            </div>

            <div className="cf-grid-2">
              <div className="cf-field">
                <label>Contract Name <span className="cf-req">*</span></label>
                <input
                  type="text"
                  name="contract_name"
                  className="cf-input"
                  value={formData.contract_name}
                  onChange={handleChange}
                  placeholder="e.g. Asia-Pacific Master Carrier Agreement 2026"
                  required
                />
              </div>

              <div className="cf-field">
                <label>Reference Code <span className="cf-req">*</span></label>
                <input
                  type="text"
                  name="contract_reference"
                  className="cf-input"
                  value={formData.contract_reference}
                  onChange={handleChange}
                  placeholder="e.g. CTR-APAC-2026-01"
                  disabled={isEdit}
                  required
                />
              </div>
            </div>

            <div className="cf-grid-2">
              
              {/* Custom Counterparty Selector */}
              <div className="cf-field relative">
                <label>Counterparty / Party <span className="cf-req">*</span></label>
                <div 
                  className={`cf-custom-select-trigger ${openDropdown === 'PARTY' ? 'active' : ''}`}
                  onClick={() => setOpenDropdown(openDropdown === 'PARTY' ? null : 'PARTY')}
                >
                  <div className="select-left">
                    <span className="party-icon-wrap">{currentParty.icon}</span>
                    <span className="select-main-text">{currentParty.name}</span>
                  </div>
                  <div className="select-right">
                    <span className="party-badge-tag">{currentParty.badge}</span>
                    <ChevronDown size={15} className={`chevron-icon ${openDropdown === 'PARTY' ? 'rotate' : ''}`} />
                  </div>
                </div>

                {openDropdown === 'PARTY' && (
                  <div className="cf-custom-menu-popover">
                    <div className="popover-header">Select Counterparty Entity</div>
                    <div className="popover-list">
                      {PRESET_PARTIES.map(p => (
                        <div
                          key={p.id}
                          className={`popover-item ${selectedPreset === p.id ? 'selected' : ''}`}
                          onClick={() => handleSelectParty(p)}
                        >
                          <div className="item-left">
                            <span className="party-icon-wrap">{p.icon}</span>
                            <div className="item-text">
                              <span className="item-title">{p.name}</span>
                              <span className="item-sub">{p.type} Counterparty</span>
                            </div>
                          </div>
                          <div className="item-right">
                            <span className="party-badge-tag">{p.badge}</span>
                            {selectedPreset === p.id && <CheckCircle size={14} className="text-blue" />}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {selectedPreset === 'custom' && (
                  <input
                    type="number"
                    name="party_id"
                    className="cf-input"
                    value={formData.party_id}
                    onChange={handleChange}
                    placeholder="Enter numeric Party ID (e.g. 5)"
                    style={{ marginTop: '8px' }}
                    required
                  />
                )}
              </div>

              {/* Custom Contract Type Selector */}
              <div className="cf-field relative">
                <label>Contract Type</label>
                <div 
                  className={`cf-custom-select-trigger ${openDropdown === 'TYPE' ? 'active' : ''}`}
                  onClick={() => setOpenDropdown(openDropdown === 'TYPE' ? null : 'TYPE')}
                >
                  <div className="select-left">
                    <span className="select-main-text">{currentType.label}</span>
                  </div>
                  <div className="select-right">
                    <span className="type-badge-tag">{currentType.badge}</span>
                    <ChevronDown size={15} className={`chevron-icon ${openDropdown === 'TYPE' ? 'rotate' : ''}`} />
                  </div>
                </div>

                {openDropdown === 'TYPE' && (
                  <div className="cf-custom-menu-popover">
                    <div className="popover-header">Agreement Domain</div>
                    <div className="popover-list">
                      {CONTRACT_TYPES.map(t => (
                        <div
                          key={t.value}
                          className={`popover-item ${formData.contract_type === t.value ? 'selected' : ''}`}
                          onClick={() => handleSelectContractType(t)}
                        >
                          <div className="item-left">
                            <div className="item-text">
                              <span className="item-title">{t.label}</span>
                              <span className="item-sub">{t.sub}</span>
                            </div>
                          </div>
                          <div className="item-right">
                            <span className="type-badge-tag">{t.badge}</span>
                            {formData.contract_type === t.value && <CheckCircle size={14} className="text-blue" />}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Section 2: Operational & Financial */}
            <div className="cf-section-title mt-2">
              <DollarSign size={14} className="text-emerald" />
              <span>Commercial & Validity Terms</span>
            </div>

            <div className="cf-grid-2">
              
              {/* Custom Transport Mode Selector */}
              <div className="cf-field relative">
                <label>Transport Mode</label>
                <div 
                  className={`cf-custom-select-trigger ${openDropdown === 'MODE' ? 'active' : ''}`}
                  onClick={() => setOpenDropdown(openDropdown === 'MODE' ? null : 'MODE')}
                >
                  <div className="select-left">
                    <span className="mode-icon-wrap">{currentMode.icon}</span>
                    <span className="select-main-text">{currentMode.label}</span>
                  </div>
                  <div className="select-right">
                    <ChevronDown size={15} className={`chevron-icon ${openDropdown === 'MODE' ? 'rotate' : ''}`} />
                  </div>
                </div>

                {openDropdown === 'MODE' && (
                  <div className="cf-custom-menu-popover">
                    <div className="popover-header">Freight Mode</div>
                    <div className="popover-list">
                      {TRANSPORT_MODES.map(m => (
                        <div
                          key={m.value}
                          className={`popover-item ${formData.transport_mode === m.value ? 'selected' : ''}`}
                          onClick={() => handleSelectMode(m)}
                        >
                          <div className="item-left">
                            <span className="mode-icon-wrap">{m.icon}</span>
                            <span className="item-title">{m.label}</span>
                          </div>
                          {formData.transport_mode === m.value && <CheckCircle size={14} className="text-blue" />}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              <div className="cf-field">
                <label>Agreement Owner</label>
                <input
                  type="text"
                  name="owner"
                  className="cf-input"
                  value={formData.owner}
                  onChange={handleChange}
                  placeholder="e.g. Varun Kanade"
                />
              </div>
            </div>

            <div className="cf-grid-2">
              <div className="cf-field">
                <label>Effective Date</label>
                <input
                  type="date"
                  name="effective_date"
                  className="cf-input"
                  value={formData.effective_date}
                  onChange={handleChange}
                />
              </div>

              <div className="cf-field">
                <label>Expiry Date</label>
                <input
                  type="date"
                  name="expiry_date"
                  className="cf-input"
                  value={formData.expiry_date}
                  onChange={handleChange}
                />
              </div>
            </div>

            <div className="cf-grid-2">
              
              {/* Custom Currency Selector */}
              <div className="cf-field relative">
                <label>Currency</label>
                <div 
                  className={`cf-custom-select-trigger ${openDropdown === 'CURRENCY' ? 'active' : ''}`}
                  onClick={() => setOpenDropdown(openDropdown === 'CURRENCY' ? null : 'CURRENCY')}
                >
                  <div className="select-left">
                    <span className="currency-pill">{currentCurrency.symbol}</span>
                    <span className="select-main-text">{currentCurrency.label}</span>
                  </div>
                  <div className="select-right">
                    <ChevronDown size={15} className={`chevron-icon ${openDropdown === 'CURRENCY' ? 'rotate' : ''}`} />
                  </div>
                </div>

                {openDropdown === 'CURRENCY' && (
                  <div className="cf-custom-menu-popover">
                    <div className="popover-list">
                      {CURRENCIES.map(c => (
                        <div
                          key={c.value}
                          className={`popover-item ${formData.currency === c.value ? 'selected' : ''}`}
                          onClick={() => handleSelectCurrency(c)}
                        >
                          <div className="item-left">
                            <span className="currency-pill">{c.symbol}</span>
                            <span className="item-title">{c.label}</span>
                          </div>
                          {formData.currency === c.value && <CheckCircle size={14} className="text-blue" />}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              <div className="cf-field">
                <label>Contract Value</label>
                <input
                  type="number"
                  name="contract_value"
                  className="cf-input"
                  value={formData.contract_value}
                  onChange={handleChange}
                  placeholder="e.g. 500000.00"
                  step="0.01"
                />
              </div>
            </div>

            {/* Section 3: Notes & Scope */}
            <div className="cf-section-title mt-2">
              <FileText size={14} className="text-slate" />
              <span>Scope & Internal Notes</span>
            </div>

            <div className="cf-field">
              <label>Commercial Scope & Description</label>
              <textarea
                name="description"
                className="cf-textarea"
                value={formData.description}
                onChange={handleChange}
                rows={3}
                placeholder="Details on committed TEU volumes, lane coverage, fuel adjustment terms, or service level KPIs..."
              />
            </div>

            <div className="cf-field">
              <label>Internal Commercial Notes</label>
              <textarea
                name="notes"
                className="cf-textarea"
                value={formData.notes}
                onChange={handleChange}
                rows={2}
                placeholder="Confidential remarks, renewal checkpoints, or procurement notes for internal team..."
              />
            </div>

          </div>

          {/* Modal Footer */}
          <div className="cf-modal-footer">
            <button type="button" className="cf-btn-cancel" onClick={onClose} disabled={isSubmitting}>
              Cancel
            </button>
            <button type="submit" className="cf-btn-submit" disabled={isSubmitting}>
              <Save size={15} />
              <span>{isSubmitting ? 'Saving Agreement...' : isEdit ? 'Update Contract' : 'Create Contract'}</span>
            </button>
          </div>
        </form>

      </div>
    </div>
  );
}

