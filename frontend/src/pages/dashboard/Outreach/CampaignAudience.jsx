import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import { getCampaignAudience, removeCampaignAudience } from '../../../services/outreachService';
import AddLeadsModal from './AddLeadsModal';

export default function CampaignAudience({ campaignId, onAudienceChanged }) {
    const [audience, setAudience] = useState([]);
    const [loading, setLoading] = useState(true);
    const [showAddModal, setShowAddModal] = useState(false);
    const [error, setError] = useState('');

    const fetchAudience = useCallback(async () => {
        setLoading(true);
        try {
            const data = await getCampaignAudience(campaignId);
            setAudience(data?.leads || data || []);
        } catch (err) {
            console.error('Failed to load audience:', err);
            setError('Failed to load campaign audience.');
        } finally {
            setLoading(false);
        }
    }, [campaignId]);

    useEffect(() => {
        fetchAudience();
    }, [fetchAudience]);

    async function handleRemove(leadId, companyName) {
        if (!window.confirm(`Remove ${companyName || 'this lead'} from the campaign?`)) return;
        try {
            await removeCampaignAudience(campaignId, leadId);
            await fetchAudience();
            if (onAudienceChanged) onAudienceChanged();
        } catch (err) {
            window.alert(err.message || 'Failed to remove lead from audience.');
        }
    }

    const totalLeads = audience.length;
    const validEmails = audience.filter(l => l.email && l.email.trim() !== '').length;
    const missingEmails = totalLeads - validEmails;

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
            {/* Audience Stats Bar */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
                <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 10, padding: 16 }}>
                    <div style={{ fontSize: 11, fontWeight: 600, color: '#94A3B8', textTransform: 'uppercase' }}>Total Leads</div>
                    <div style={{ fontSize: 24, fontWeight: 700, color: '#1E293B', marginTop: 4 }}>{totalLeads}</div>
                </div>
                <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 10, padding: 16 }}>
                    <div style={{ fontSize: 11, fontWeight: 600, color: '#22C55E', textTransform: 'uppercase' }}>Valid Emails</div>
                    <div style={{ fontSize: 24, fontWeight: 700, color: '#1E293B', marginTop: 4 }}>{validEmails}</div>
                </div>
                <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 10, padding: 16 }}>
                    <div style={{ fontSize: 11, fontWeight: 600, color: '#F59E0B', textTransform: 'uppercase' }}>Missing Email</div>
                    <div style={{ fontSize: 24, fontWeight: 700, color: '#1E293B', marginTop: 4 }}>{missingEmails}</div>
                </div>
            </div>

            {/* Action Header */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: '#1E293B' }}>Campaign Target Audience</div>
                <button
                    className="outreach-btn outreach-btn-primary"
                    onClick={() => setShowAddModal(true)}
                >
                    + Add Leads
                </button>
            </div>

            {error && <div className="outreach-error">{error}</div>}

            {/* Table */}
            <div className="outreach-table-wrap">
                {loading ? (
                    <div style={{ padding: 32, textAlign: 'center', color: '#94A3B8' }}>Loading audience...</div>
                ) : audience.length === 0 ? (
                    <div className="outreach-empty">
                        <div className="outreach-empty-title">No leads in this campaign</div>
                        <div className="outreach-empty-sub">
                            Add existing leads from your organization to start targeting prospects in this campaign.
                        </div>
                        <button className="outreach-btn outreach-btn-primary" onClick={() => setShowAddModal(true)}>
                            + Add Leads
                        </button>
                    </div>
                ) : (
                    <table className="outreach-table">
                        <thead>
                            <tr>
                                <th>Company / Contact</th>
                                <th>Email</th>
                                <th>Lead Status</th>
                                <th>Added On</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {audience.map(lead => (
                                <tr key={lead.id || lead.lead_id}>
                                    <td>
                                        <div className="campaign-name-cell">
                                            <span className="campaign-name">{lead.company_name || lead.company || '—'}</span>
                                            <span className="campaign-date">{lead.contact_name || lead.name || '—'}</span>
                                        </div>
                                    </td>
                                    <td>
                                        {lead.email && lead.email.trim() !== '' ? (
                                            <span style={{ color: '#334155', fontSize: 13 }}>{lead.email}</span>
                                        ) : (
                                            <span style={{ color: '#DC2626', fontSize: 12, fontWeight: 600, background: 'rgba(239,68,68,0.08)', padding: '2px 8px', borderRadius: 99 }}>
                                                Missing Email
                                            </span>
                                        )}
                                    </td>
                                    <td>
                                        <span style={{ fontSize: 12, fontWeight: 500, color: '#475569' }}>
                                            {lead.status || 'NEW'}
                                        </span>
                                    </td>
                                    <td style={{ color: '#64748B', fontSize: 13 }}>
                                        {lead.added_at ? new Date(lead.added_at).toLocaleDateString() : '—'}
                                    </td>
                                    <td>
                                        <button
                                            className="outreach-btn outreach-btn-danger"
                                            style={{ padding: '4px 10px', fontSize: 12 }}
                                            onClick={() => handleRemove(lead.id || lead.lead_id, lead.company_name || lead.company)}
                                        >
                                            Remove
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>

            {showAddModal && (
                <AddLeadsModal
                    campaignId={campaignId}
                    onClose={() => setShowAddModal(false)}
                    onLeadsAdded={() => {
                        fetchAudience();
                        if (onAudienceChanged) onAudienceChanged();
                    }}
                />
            )}
        </div>
    );
}

CampaignAudience.propTypes = {
    campaignId: PropTypes.string.isRequired,
    onAudienceChanged: PropTypes.func,
};