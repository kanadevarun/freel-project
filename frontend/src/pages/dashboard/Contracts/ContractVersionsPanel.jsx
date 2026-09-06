import React, { useState, useEffect, useCallback } from 'react';
import { 
  History, GitCommit, ArrowRight, CheckCircle2, ShieldCheck, 
  Sparkles, RefreshCw, Eye, GitCompare, ArrowUpRight, Lock, 
  Clock, AlertCircle, Plus, FileText
} from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import ConfirmModal from '../Settings/ConfirmModal';
import toast from 'react-hot-toast';
import ContractVersionComparisonModal from './ContractVersionComparisonModal';
import './ContractVersionsPanel.css';

export default function ContractVersionsPanel({ contractId, contractStatus, onVersionPromoted }) {
  const [versions, setVersions] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isPromoting, setIsPromoting] = useState(false);
  const [promotingVersion, setPromotingVersion] = useState(null);

  // Compare Modal State
  const [showCompareModal, setShowCompareModal] = useState(false);
  const [baseVersion, setBaseVersion] = useState(null);
  const [targetVersion, setTargetVersion] = useState(null);

  // New Draft Version Modal State
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createForm, setCreateForm] = useState({ version_label: '', change_summary: '' });
  const [isCreating, setIsCreating] = useState(false);

  const fetchVersions = useCallback(async () => {
    if (!contractId) return;
    setIsLoading(true);
    try {
      const res = await contractsService.getContractVersions(contractId);
      setVersions(res?.data || []);
    } catch (err) {
      console.error('Failed to load contract versions', err);
      toast.error('Failed to load version history');
    } finally {
      setIsLoading(false);
    }
  }, [contractId]);

  useEffect(() => {
    fetchVersions();
  }, [fetchVersions]);

  const handleMakeEffective = (version) => {
    setPromotingVersion(version);
  };

  const handleConfirmPromotion = async () => {
    if (!promotingVersion) return;
    setIsPromoting(true);
    try {
      await contractsService.makeVersionEffective(contractId, promotingVersion.id);
      toast.success(`Promoted ${promotingVersion.version_label} to EFFECTIVE`);
      setPromotingVersion(null);
      await fetchVersions();
      if (onVersionPromoted) onVersionPromoted();
    } catch (err) {
      console.error('Promotion error', err);
      toast.error(err?.response?.data?.error?.message || 'Failed to promote version');
    } finally {
      setIsPromoting(false);
    }
  };

  const handleCreateVersion = async (e) => {
    e.preventDefault();
    setIsCreating(true);
    try {
      await contractsService.createContractVersion(contractId, createForm);
      toast.success('Created new draft version snapshot');
      setShowCreateModal(false);
      setCreateForm({ version_label: '', change_summary: '' });
      await fetchVersions();
    } catch (err) {
      console.error('Create version error', err);
      toast.error('Failed to create new draft version');
    } finally {
      setIsCreating(false);
    }
  };

  const handleOpenCompare = (v1, v2) => {
    setBaseVersion(v1);
    setTargetVersion(v2);
    setShowCompareModal(true);
  };

  return (
    <div className="versions-panel-container">
      {/* Top Action Header */}
      <div className="versions-top-bar">
        <div className="versions-title-group">
          <div className="v-icon-badge">
            <History size={16} />
          </div>
          <div>
            <h4>Contract Version History & Immutability</h4>
            <p>Every commercial update creates an immutable snapshot. Effective versions are never silently overwritten.</p>
          </div>
        </div>

        <div className="versions-actions-group">
          <button 
            className="v-refresh-btn" 
            onClick={fetchVersions} 
            title="Refresh Versions"
            disabled={isLoading}
          >
            <RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />
          </button>
          <button 
            className="v-create-btn"
            onClick={() => setShowCreateModal(true)}
          >
            <Plus size={14} /> New Draft Version
          </button>
        </div>
      </div>

      {/* Version History List */}
      {isLoading ? (
        <div className="versions-loading">
          <RefreshCw size={24} className="spin-icon text-blue" />
          <span>Loading immutable versions...</span>
        </div>
      ) : versions.length === 0 ? (
        <div className="versions-empty-state">
          <History size={36} className="text-muted" />
          <h5>No Version History Found</h5>
          <p>No snapshots recorded yet. Create a draft version to start version tracking.</p>
        </div>
      ) : (
        <div className="versions-timeline-wrapper">
          {versions.map((ver, idx) => {
            const isEffective = ver.status === 'EFFECTIVE';
            const isSuperseded = ver.status === 'SUPERSEDED';
            const isDraft = ver.status === 'DRAFT';
            const prevVersion = versions[idx + 1] || null;

            return (
              <div 
                key={ver.id} 
                className={`version-card ${isEffective ? 'effective' : isSuperseded ? 'superseded' : 'draft'}`}
              >
                {/* Status Indicator Marker */}
                <div className="v-timeline-rail">
                  <div className={`v-status-dot ${ver.status.toLowerCase()}`}>
                    {isEffective ? <CheckCircle2 size={13} /> : isSuperseded ? <Lock size={12} /> : <Clock size={12} />}
                  </div>
                  {idx < versions.length - 1 && <div className="v-rail-line"></div>}
                </div>

                {/* Card Content */}
                <div className="v-card-body">
                  <div className="v-card-header">
                    <div className="v-label-group">
                      <span className="v-number">v{ver.version_number}.0</span>
                      <h5 className="v-label-title">{ver.version_label}</h5>
                      <span className={`v-badge ${ver.status.toLowerCase()}`}>
                        {ver.status}
                      </span>
                    </div>

                    <div className="v-card-controls">
                      {prevVersion && (
                        <button 
                          className="v-compare-btn"
                          onClick={() => handleOpenCompare(prevVersion, ver)}
                          title={`Compare with v${prevVersion.version_number}.0`}
                        >
                          <GitCompare size={13} /> Compare
                        </button>
                      )}

                      {!isEffective && !isSuperseded && (
                        <button
                          className="v-promote-btn"
                          onClick={() => handleMakeEffective(ver)}
                          disabled={isPromoting}
                        >
                          <Sparkles size={13} /> Make Effective
                        </button>
                      )}
                    </div>
                  </div>

                  {/* Summary */}
                  <div className="v-card-summary">
                    <p>{ver.change_summary || 'Snapshot of agreement commercial terms.'}</p>
                  </div>

                  {/* Metadata Row */}
                  <div className="v-card-meta">
                    {ver.effective_date && (
                      <span className="meta-item">
                        <strong>Effective:</strong> {ver.effective_date}
                      </span>
                    )}
                    {ver.expiry_date && (
                      <span className="meta-item">
                        <strong>Expiry:</strong> {ver.expiry_date}
                      </span>
                    )}
                    {ver.created_at && (
                      <span className="meta-item">
                        <strong>Created:</strong> {new Date(ver.created_at).toLocaleDateString()}
                      </span>
                    )}
                    {isSuperseded && ver.superseded_at && (
                      <span className="meta-item text-amber">
                        <strong>Superseded:</strong> {new Date(ver.superseded_at).toLocaleDateString()}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Version Comparison Modal */}
      {showCompareModal && baseVersion && targetVersion && (
        <ContractVersionComparisonModal
          contractId={contractId}
          baseVersion={baseVersion}
          targetVersion={targetVersion}
          onClose={() => {
            setShowCompareModal(false);
            setBaseVersion(null);
            setTargetVersion(null);
          }}
        />
      )}

      {/* Create Draft Version Modal */}
      {showCreateModal && (
        <div className="v-modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="v-modal-card" onClick={e => e.stopPropagation()}>
            <div className="v-modal-header">
              <div className="modal-title-left">
                <FileText size={16} className="text-blue" />
                <h5>Create Draft Version Snapshot</h5>
              </div>
              <button className="v-close-modal" onClick={() => setShowCreateModal(false)}>×</button>
            </div>

            <form onSubmit={handleCreateVersion}>
              <div className="v-modal-body">
                <div className="form-group">
                  <label>Version Label</label>
                  <input 
                    type="text" 
                    placeholder="e.g. v2.0 Rate Revision Draft"
                    value={createForm.version_label}
                    onChange={e => setCreateForm({...createForm, version_label: e.target.value})}
                  />
                  <span className="v-field-hint">Leave blank for automatic numbering (e.g. v2.0).</span>
                </div>

                <div className="form-group">
                  <label>Change Summary / Rationale</label>
                  <textarea 
                    rows={3}
                    placeholder="Describe proposed changes or renegotiation points..."
                    value={createForm.change_summary}
                    onChange={e => setCreateForm({...createForm, change_summary: e.target.value})}
                  />
                </div>
              </div>

              <div className="v-modal-footer">
                <button type="button" className="btn-cancel" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-confirm" disabled={isCreating}>
                  {isCreating ? 'Creating Snapshot...' : 'Create Draft Version'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Promotion Confirmation Modal */}
      <ConfirmModal
        isOpen={Boolean(promotingVersion)}
        title="Promote Contract Version to Effective"
        message={`Are you sure you want to promote ${promotingVersion?.version_label} to EFFECTIVE? The previous version will become SUPERSEDED and remain immutable.`}
        confirmText="Promote to Effective"
        confirmStyle="primary"
        isLoading={isPromoting}
        onConfirm={handleConfirmPromotion}
        onCancel={() => setPromotingVersion(null)}
      />
    </div>
  );
}
