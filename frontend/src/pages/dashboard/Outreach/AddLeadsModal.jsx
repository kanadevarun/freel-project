import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { listLeads } from '../../../services/leadsService';
import { addCampaignAudience } from '../../../services/outreachService';

export default function AddLeadsModal({ campaignId, onClose, onLeadsAdded }) {
    const [leads, setLeads] = useState([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedLeadIds, setSelectedLeadIds] = useState(new Set());
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState('');

    const fetchLeads = useCallback(async () => {
        setLoading(true);
        try {
            const data = await listLeads({ limit: 200 });
            setLeads(data?.leads || data || []);
        } catch (err) {
            console.error('Failed to fetch org leads:', err);
            setError('Failed to load organization leads.');
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchLeads();
    }, [fetchLeads]);

    const filteredLeads = leads.filter(lead => {
        const q = searchQuery.toLowerCase();
        const company = (lead.company_name || lead.company || '').toLowerCase();
        const name = (lead.contact_name || lead.name || '').toLowerCase();
        const email = (lead.email || '').toLowerCase();
        return company.includes(q) || name.includes(q) || email.includes(q);
    });

    function toggleSelectAll() {
        if (selectedLeadIds.size === filteredLeads.length) {
            setSelectedLeadIds(new Set());
        } else {
            setSelectedLeadIds(new Set(filteredLeads.map(l => l.id || l.lead_id)));
        }
    }

    function toggleSelectLead(id) {
        const next = new Set(selectedLeadIds);
        if (next.has(id)) next.delete(id);
        else next.add(id);
        setSelectedLeadIds(next);
    }

    async function handleAddSelected() {
        if (selectedLeadIds.size === 0) return;
        setSubmitting(true);
        setError('');

        try {
            await addCampaignAudience(campaignId, Array.from(selectedLeadIds));
            onLeadsAdded();
            onClose();
        } catch (err) {
            setError(err.message || 'Failed to add leads to campaign.');
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <div className="outreach-modal-overlay" onClick={onClose}>
            <div className="outreach-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: 700 }}>
                <div className="modal-header">
                    <div>
                        <div className="modal-title">Add Leads to Campaign</div>
                        <div style={{ fontSize: 12, color: '#64748B', marginTop: 2 }}>
                            Select existing leads from your organization to target in this outreach campaign.
                        </div>
                    </div>
                    <button className="modal-close-btn" onClick={onClose} disabled={submitting}>✕</button>
                </div>

                {error && <div className="outreach-error" style={{ marginBottom: 16 }}>{error}</div>}

                <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <input
                            type="text"
                            placeholder="Search leads by company, name, or email..."
                            value={searchQuery}
                            onChange={e => setSearchQuery(e.target.value)}
                            style={{ flex: 1, padding: '9px 12px', border: '1px solid #E2E8F0', borderRadius: 8, fontSize: 13, background: '#F8FAFC' }}
                        />
                        <div style={{ fontSize: 13, fontWeight: 600, color: '#00BFA5', marginLeft: 16, whiteSpace: 'nowrap' }}>
                            {selectedLeadIds.size} Leads Selected
                        </div>
                    </div>

                    <div style={{ border: '1px solid #E2E8F0', borderRadius: 10, maxHeight: 320, overflowY: 'auto' }}>
                        {loading ? (
                            <div style={{ padding: 32, textAlign: 'center', color: '#94A3B8' }}>Loading available leads...</div>
                        ) : filteredLeads.length === 0 ? (
                            <div style={{ padding: 32, textAlign: 'center', color: '#94A3B8' }}>No leads found matching your search.</div>
                        ) : (
                            <table className="outreach-table" style={{ margin: 0 }}>
                                <thead>
                                    <tr>
                                        <th style={{ width: 40, textAlign: 'center' }}>
                                            <input
                                                type="checkbox"
                                                checked={filteredLeads.length > 0 && selectedLeadIds.size === filteredLeads.length}
                                                onChange={toggleSelectAll}
                                            />
                                        </th>
                                        <th>Company / Contact</th>
                                        <th>Email</th>
                                        <th>Status</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {filteredLeads.map(lead => {
                                        const id = lead.id || lead.lead_id;
                                        const isSelected = selectedLeadIds.has(id);
                                        return (
                                            <tr key={id} onClick={() => toggleSelectLead(id)} style={{ cursor: 'pointer' }}>
                                                <td style={{ textAlign: 'center' }} onClick={e => e.stopPropagation()}>
                                                    <input
                                                        type="checkbox"
                                                        checked={isSelected}
                                                        onChange={() => toggleSelectLead(id)}
                                                    />
                                                </td>
                                                <td>
                                                    <div className="campaign-name-cell">
                                                        <span className="campaign-name">{lead.company_name || lead.company || '—'}</span>
                                                        <span className="campaign-date">{lead.contact_name || lead.name || '—'}</span>
                                                    </div>
                                                </td>
                                                <td style={{ color: '#475569', fontSize: 13 }}>{lead.email || '—'}</td>
                                                <td>
                                                    <span style={{ fontSize: 12, fontWeight: 500, color: '#475569' }}>
                                                        {lead.status || 'NEW'}
                                                    </span>
                                                </td>
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        )}
                    </div>
                </div>

                <div className="modal-actions" style={{ marginTop: 24 }}>
                    <button type="button" className="outreach-btn outreach-btn-ghost" onClick={onClose} disabled={submitting}>
                        Cancel
                    </button>
                    <button
                        type="button"
                        className="outreach-btn outreach-btn-primary"
                        onClick={handleAddSelected}
                        disabled={submitting || selectedLeadIds.size === 0}
                    >
                        {submitting ? 'Adding...' : `Add ${selectedLeadIds.size} Leads`}
                    </button>
                </div>
            </div>
        </div>
    );
}

AddLeadsModal.propTypes = {
    campaignId: PropTypes.string.isRequired,
    onClose: PropTypes.func.isRequired,
    onLeadsAdded: PropTypes.func.isRequired,
};