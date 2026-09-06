import React, { useState, useEffect } from 'react';
import { 
  GitCompare, ArrowRight, RefreshCw, AlertCircle, 
  CheckCircle2, X, Plus, Minus, Edit3
} from 'lucide-react';
import { contractsService } from '../../../services/contractsService';
import './ContractVersionComparisonModal.css';

export default function ContractVersionComparisonModal({ contractId, baseVersion, targetVersion, onClose }) {
  const [comparison, setComparison] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function fetchDiff() {
      if (!contractId || !baseVersion || !targetVersion) return;
      setIsLoading(true);
      try {
        const res = await contractsService.compareContractVersions(contractId, baseVersion.id, targetVersion.id);
        setComparison(res?.data || null);
      } catch (err) {
        console.error('Failed to compare versions', err);
      } finally {
        setIsLoading(false);
      }
    }
    fetchDiff();
  }, [contractId, baseVersion, targetVersion]);

  const formatFieldName = (name) => {
    return name
      .replace(/_/g, ' ')
      .replace(/\b\w/g, c => c.toUpperCase());
  };

  return (
    <div className="diff-modal-overlay" onClick={onClose}>
      <div className="diff-modal-window" onClick={e => e.stopPropagation()}>
        {/* Modal Header */}
        <div className="diff-modal-header">
          <div className="diff-title-group">
            <div className="diff-icon-badge">
              <GitCompare size={18} />
            </div>
            <div>
              <h3>Version Comparison Delta</h3>
              <p>
                Comparing <strong>{baseVersion.version_label} (v{baseVersion.version_number}.0)</strong> with{' '}
                <strong>{targetVersion.version_label} (v{targetVersion.version_number}.0)</strong>
              </p>
            </div>
          </div>
          <button className="diff-close-btn" onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        {/* Modal Body */}
        <div className="diff-modal-body">
          {isLoading ? (
            <div className="diff-loading">
              <RefreshCw size={24} className="spin-icon text-blue" />
              <span>Analyzing parameter differences...</span>
            </div>
          ) : !comparison || comparison.changes.length === 0 ? (
            <div className="diff-empty-state">
              <CheckCircle2 size={32} className="text-emerald" />
              <h4>Identical Commercial Parameters</h4>
              <p>No parameter differences detected between these two contract versions.</p>
            </div>
          ) : (
            <div className="diff-changes-container">
              <div className="diff-table-header">
                <span className="col-field">Field / Term</span>
                <span className="col-val prev">Base Version (v{comparison.base_version_num}.0)</span>
                <span className="col-val prop">Target Version (v{comparison.target_version_num}.0)</span>
                <span className="col-type">Change</span>
              </div>

              <div className="diff-table-body">
                {comparison.changes.map((chg, idx) => (
                  <div key={idx} className={`diff-row ${chg.change_type.toLowerCase()}`}>
                    <span className="col-field">
                      <strong>{formatFieldName(chg.field_name)}</strong>
                      <code>{chg.field_name}</code>
                    </span>

                    <span className="col-val prev">
                      {chg.previous_value ? (
                        <span className="val-text">{chg.previous_value}</span>
                      ) : (
                        <span className="val-empty">— Empty —</span>
                      )}
                    </span>

                    <span className="col-val prop">
                      {chg.proposed_value ? (
                        <span className="val-text highlight">{chg.proposed_value}</span>
                      ) : (
                        <span className="val-empty">— Removed —</span>
                      )}
                    </span>

                    <span className="col-type">
                      <span className={`diff-tag ${chg.change_type.toLowerCase()}`}>
                        {chg.change_type === 'MODIFY' && <Edit3 size={11} />}
                        {chg.change_type === 'ADD' && <Plus size={11} />}
                        {chg.change_type === 'REMOVE' && <Minus size={11} />}
                        {chg.change_type}
                      </span>
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Modal Footer */}
        <div className="diff-modal-footer">
          <div className="diff-footer-info">
            <span>Historical versions remain immutable. Promoting a target version updates active parameters deterministically.</span>
          </div>
          <button className="diff-btn-close" onClick={onClose}>
            Close Comparison
          </button>
        </div>
      </div>
    </div>
  );
}
