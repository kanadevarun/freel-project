import React from 'react';
import { useNavigate } from 'react-router-dom';
import { FileCode2, FileSpreadsheet, ShieldAlert, Sparkles } from 'lucide-react';
import PageHeader from '../../../components/dashboard/PageHeader';
import './TemplatesPage.css';

const TEMPLATES = [
  {
    id: 'hbl',
    name: 'Standard Ocean House Bill of Lading (HBL)',
    category: 'Ocean Freight',
    desc: 'FIATA-compliant negotiable and non-negotiable House Bill of Lading with standard carrier clauses.',
    status: 'Standard Template Available',
  },
  {
    id: 'vgm',
    name: 'SOLAS Verified Gross Mass (VGM) Certificate',
    category: 'Maritime Compliance',
    desc: 'Standardized shipper VGM declaration format for container terminal dispatch.',
    status: 'Standard Template Available',
  },
  {
    id: 'packing_list',
    name: 'Commercial Packing List & HS Code Breakdown',
    category: 'Customs & Border',
    desc: 'Automated packing list template with gross weight, net weight, volume CBM, and tariff headings.',
    status: 'Standard Template Available',
  },
  {
    id: 'air_awb',
    name: 'IATA Neutral Air Waybill (e-AWB)',
    category: 'Air Freight',
    desc: 'Standard 8-digit IATA air waybill format with electronic transmission hooks.',
    status: 'Coming Soon',
  },
];

export default function TemplatesPage() {
  const navigate = useNavigate();

  return (
    <div className="templates-page">
      <PageHeader
        title="Documentation Templates Studio"
        subtitle="Standardized logistics document templates, bill of lading clauses, and customs declarations"
      />

      {/* Notice Banner explaining status */}
      <div className="template-notice-card">
        <div className="notice-icon-badge">
          <Sparkles size={18} />
        </div>
        <div className="notice-text">
          <strong>Template Studio Integration Note:</strong> Standard freight templates are available for export in active shipment workspaces. Custom visual template editing and organization-specific clause builders are scheduled for the next platform phase.
        </div>
      </div>

      <div className="templates-grid">
        {TEMPLATES.map((t) => (
          <div key={t.id} className="template-card">
            <div className="template-card-header">
              <span className="template-cat-tag">{t.category}</span>
              <span className={`template-status-pill ${t.status.includes('Available') ? 'available' : 'soon'}`}>
                {t.status}
              </span>
            </div>
            <h3 className="template-name">{t.name}</h3>
            <p className="template-desc">{t.desc}</p>
            <div className="template-footer">
              <button className="btn-use-template" onClick={() => navigate('/dashboard/documents')}>
                <span>View Document Center</span>
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
