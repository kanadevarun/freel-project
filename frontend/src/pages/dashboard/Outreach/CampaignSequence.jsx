import { useState, useEffect, useCallback } from 'react';
import PropTypes from 'prop-types';
import {
    getCampaignSequence,
    addCampaignSequenceStep,
    updateCampaignSequenceStep,
    deleteCampaignSequenceStep,
    reorderCampaignSequence,
} from '../../../services/outreachService';

export default function CampaignSequence({ campaignId, onSequenceChanged }) {
    const [sequence, setSequence] = useState([]);
    const [loading, setLoading] = useState(true);
    const [showStepModal, setShowStepModal] = useState(false);
    const [editingStep, setEditingStep] = useState(null);
    const [error, setError] = useState('');

    const [stepName, setStepName] = useState('');
    const [subject, setSubject] = useState('');
    const [body, setBody] = useState('');
    const [delayDays, setDelayDays] = useState(3);
    const [submitting, setSubmitting] = useState(false);

    const fetchSequence = useCallback(async () => {
        setLoading(true);
        try {
            const data = await getCampaignSequence(campaignId);
            setSequence(data?.steps || data || []);
        } catch (err) {
            console.error('Failed to load sequence:', err);
            setError('Failed to load campaign sequence.');
        } finally {
            setLoading(false);
        }
    }, [campaignId]);

    useEffect(() => {
        fetchSequence();
    }, [fetchSequence]);

    function openAddModal() {
        setEditingStep(null);
        setStepName(sequence.length === 0 ? 'Initial Outreach' : `Follow-up ${sequence.length}`);
        setSubject('');
        setBody('');
        setDelayDays(sequence.length === 0 ? 0 : 3);
        setShowStepModal(true);
    }

    function openEditModal(step) {
        setEditingStep(step);
        setStepName(step.name || step.step_name || '');
        setSubject(step.subject || '');
        setBody(step.body || '');
        setDelayDays(step.delay_days !== undefined ? step.delay_days : step.delay || 3);
        setShowStepModal(true);
    }

    async function handleSaveStep(e) {
        e.preventDefault();
        if (!subject.trim() || !body.trim()) {
            setError('Subject and body are required for sequence steps.');
            return;
        }

        setSubmitting(true);
        setError('');

        try {
            const payload = {
                name: stepName.trim() || `Step ${sequence.length + 1}`,
                subject: subject.trim(),
                body: body.trim(),
                delay_days: sequence.length === 0 ? 0 : Number(delayDays) || 0,
                channel: 'EMAIL',
            };

            if (editingStep) {
                const stepId = editingStep.id || editingStep.step_id;
                await updateCampaignSequenceStep(campaignId, stepId, payload);
            } else {
                await addCampaignSequenceStep(campaignId, payload);
            }

            setShowStepModal(false);
            await fetchSequence();
            if (onSequenceChanged) onSequenceChanged();
        } catch (err) {
            setError(err.message || 'Failed to save sequence step.');
        } finally {
            setSubmitting(false);
        }
    }

    async function handleDeleteStep(stepId) {
        if (!window.confirm('Delete this sequence step?')) return;
        try {
            await deleteCampaignSequenceStep(campaignId, stepId);
            await fetchSequence();
            if (onSequenceChanged) onSequenceChanged();
        } catch (err) {
            window.alert(err.message || 'Failed to delete sequence step.');
        }
    }

    async function handleMove(index, direction) {
        const newSeq = [...sequence];
        const targetIndex = direction === 'up' ? index - 1 : index + 1;
        if (targetIndex < 0 || targetIndex >= newSeq.length) return;

        const temp = newSeq[index];
        newSeq[index] = newSeq[targetIndex];
        newSeq[targetIndex] = temp;

        const stepIds = newSeq.map(s => s.id || s.step_id);
        try {
            await reorderCampaignSequence(campaignId, stepIds);
            setSequence(newSeq);
            if (onSequenceChanged) onSequenceChanged();
        } catch (err) {
            window.alert(err.message || 'Failed to reorder sequence.');
        }
    }

    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <div style={{ fontSize: 15, fontWeight: 600, color: '#1E293B' }}>Outreach Sequence Steps</div>
                    <div style={{ fontSize: 12.5, color: '#64748B', marginTop: 2 }}>
                        Configure the automated sequence of emails sent to campaign leads over time.
                    </div>
                </div>
                <button className="outreach-btn outreach-btn-primary" onClick={openAddModal}>
                    + Add Sequence Step
                </button>
            </div>

            {error && <div className="outreach-error">{error}</div>}

            {loading ? (
                <div style={{ padding: 32, textAlign: 'center', color: '#94A3B8' }}>Loading sequence...</div>
            ) : sequence.length === 0 ? (
                <div className="outreach-table-wrap">
                    <div className="outreach-empty">
                        <div className="outreach-empty-title">No sequence steps defined</div>
                        <div className="outreach-empty-sub">
                            Add at least one email step so your campaign can deliver messages to prospects.
                        </div>
                        <button className="outreach-btn outreach-btn-primary" onClick={openAddModal}>
                            + Add Sequence Step
                        </button>
                    </div>
                </div>
            ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 12, maxWidth: 800 }}>
                    {sequence.map((step, index) => {
                        const stepId = step.id || step.step_id;
                        const isFirst = index === 0;
                        const delay = step.delay_days !== undefined ? step.delay_days : step.delay || 0;

                        return (
                            <div key={stepId} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                                {!isFirst && (
                                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8, padding: '4px 0' }}>
                                        <div style={{ height: 2, width: 20, background: '#CBD5E1' }} />
                                        <span style={{ fontSize: 11.5, fontWeight: 600, color: '#64748B', background: '#F1F5F9', padding: '3px 10px', borderRadius: 99, border: '1px solid #E2E8F0' }}>
                                            Wait {delay} {delay === 1 ? 'Day' : 'Days'}
                                        </span>
                                        <div style={{ height: 2, width: 20, background: '#CBD5E1' }} />
                                    </div>
                                )}

                                <div style={{ background: '#FFFFFF', border: '1px solid #E2E8F0', borderRadius: 12, padding: 20, display: 'flex', flexDirection: 'column', gap: 12, boxShadow: '0 1px 3px rgba(0,0,0,0.04)' }}>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                                            <span style={{ width: 24, height: 24, borderRadius: '50%', background: 'rgba(0,191,165,0.1)', color: '#00BFA5', fontSize: 12, fontWeight: 700, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                                {index + 1}
                                            </span>
                                            <div>
                                                <div style={{ fontSize: 14, fontWeight: 700, color: '#1E293B' }}>{step.name || step.step_name || `Step ${index + 1}`}</div>
                                                <div style={{ fontSize: 12, color: '#64748B', marginTop: 1 }}>
                                                    {isFirst ? 'Send Immediately' : `Send ${delay} days after previous step`}
                                                </div>
                                            </div>
                                        </div>

                                        <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                                            <button className="outreach-btn outreach-btn-ghost" style={{ padding: '4px 8px', fontSize: 11.5 }} onClick={() => handleMove(index, 'up')} disabled={index === 0}>
                                                ↑
                                            </button>
                                            <button className="outreach-btn outreach-btn-ghost" style={{ padding: '4px 8px', fontSize: 11.5 }} onClick={() => handleMove(index, 'down')} disabled={index === sequence.length - 1}>
                                                ↓
                                            </button>
                                            <button className="outreach-btn outreach-btn-ghost" style={{ padding: '4px 10px', fontSize: 12 }} onClick={() => openEditModal(step)}>
                                                Edit
                                            </button>
                                            <button className="outreach-btn outreach-btn-danger" style={{ padding: '4px 10px', fontSize: 12 }} onClick={() => handleDeleteStep(stepId)}>
                                                Delete
                                            </button>
                                        </div>
                                    </div>

                                    <div style={{ background: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: 8, padding: 14, display: 'flex', flexDirection: 'column', gap: 6 }}>
                                        <div style={{ fontSize: 12, fontWeight: 600, color: '#334155' }}>
                                            Subject: <span style={{ fontWeight: 400, color: '#475569' }}>{step.subject}</span>
                                        </div>
                                        <div style={{ fontSize: 12, color: '#475569', lineHeight: 1.5, whiteSpace: 'pre-wrap', maxHeight: 80, overflowY: 'auto' }}>
                                            {step.body}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}

            {showStepModal && (
                <div className="outreach-modal-overlay" onClick={() => setShowStepModal(false)}>
                    <div className="outreach-modal" onClick={e => e.stopPropagation()}>
                        <div className="modal-header">
                            <div className="modal-title">{editingStep ? 'Edit Sequence Step' : 'Add Sequence Step'}</div>
                            <button className="modal-close-btn" onClick={() => setShowStepModal(false)} disabled={submitting}>✕</button>
                        </div>

                        {error && <div className="outreach-error" style={{ marginBottom: 16 }}>{error}</div>}

                        <form onSubmit={handleSaveStep} className="outreach-form">
                            <div className="form-group">
                                <label>Step Name</label>
                                <input
                                    type="text"
                                    value={stepName}
                                    onChange={e => setStepName(e.target.value)}
                                    placeholder="e.g., Initial Outreach / Follow-up 1"
                                    required
                                />
                            </div>

                            {sequence.length > 0 && !isFirstStep(editingStep, sequence) && (
                                <div className="form-group">
                                    <label>Delay (Days after previous step)</label>
                                    <input
                                        type="number"
                                        min="1"
                                        value={delayDays}
                                        onChange={e => setDelayDays(e.target.value)}
                                        required
                                    />
                                </div>
                            )}

                            <div className="form-group">
                                <label>Email Subject *</label>
                                <input
                                    type="text"
                                    value={subject}
                                    onChange={e => setSubject(e.target.value)}
                                    placeholder="e.g., Follow up on freight rate inquiry"
                                    required
                                />
                            </div>

                            <div className="form-group">
                                <label>Email Body *</label>
                                <textarea
                                    value={body}
                                    onChange={e => setBody(e.target.value)}
                                    placeholder="Hi {{first_name}}, ..."
                                    style={{ minHeight: 140 }}
                                    required
                                />
                            </div>

                            <div className="modal-actions" style={{ marginTop: 20 }}>
                                <button type="button" className="outreach-btn outreach-btn-ghost" onClick={() => setShowStepModal(false)} disabled={submitting}>
                                    Cancel
                                </button>
                                <button type="submit" className="outreach-btn outreach-btn-primary" disabled={submitting}>
                                    {submitting ? 'Saving...' : 'Save Step'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
}

function isFirstStep(editingStep, sequence) {
    if (!editingStep) return sequence.length === 0;
    return sequence[0]?.id === editingStep.id || sequence[0]?.step_id === editingStep.step_id;
}

CampaignSequence.propTypes = {
    campaignId: PropTypes.string.isRequired,
    onSequenceChanged: PropTypes.func,
};