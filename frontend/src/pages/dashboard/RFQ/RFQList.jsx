import React from 'react';
import { RFQ_STAGES, STAGE_CONFIG } from './constants';
import StatusBadge from '../../../components/dashboard/StatusBadge';

/**
 * RFQList displays a table of RFQs.
 */
export default function RFQList({ rfqs, isLoading, onRowClick }) {
  if (isLoading) {
    return (
      <div className="table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>RFQ Number</th>
              <th>Customer ID</th>
              <th>Origin → Destination</th>
              <th>Stage</th>
              <th>Target Date</th>
            </tr>
          </thead>
          <tbody>
            {[1, 2, 3].map((i) => (
              <tr key={i} className="skeleton-row">
                <td><div className="skeleton-box" style={{ width: '80px' }}></div></td>
                <td><div className="skeleton-box" style={{ width: '120px' }}></div></td>
                <td><div className="skeleton-box" style={{ width: '150px' }}></div></td>
                <td><div className="skeleton-box" style={{ width: '100px', borderRadius: '12px' }}></div></td>
                <td><div className="skeleton-box" style={{ width: '90px' }}></div></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  if (!rfqs || rfqs.length === 0) {
    return (
      <div className="empty-state">
        <div className="empty-state-icon">🚢</div>
        <h3>No RFQs found</h3>
        <p>Get started by extracting a new shipment request from a customer.</p>
      </div>
    );
  }

  return (
    <div className="table-container">
      <table className="data-table">
        <thead>
          <tr>
            <th>RFQ Number</th>
            <th>Customer ID</th>
            <th>Origin → Destination</th>
            <th>Stage</th>
            <th>Target Date</th>
          </tr>
        </thead>
        <tbody>
          {rfqs.map((rfq) => {
            const stageConfig = STAGE_CONFIG[rfq.stage] || { label: rfq.stage, color: 'gray' };
            const route = rfq.origin && rfq.destination 
              ? `${rfq.origin} → ${rfq.destination}`
              : 'Not specified';
            
            return (
              <tr 
                key={rfq.id} 
                onClick={() => onRowClick(rfq)}
                style={{ cursor: 'pointer' }}
              >
                <td style={{ fontWeight: 500 }}>{rfq.rfq_number}</td>
                <td>{rfq.customer_id}</td>
                <td>{route}</td>
                <td>
                  <StatusBadge status={stageConfig.label} color={stageConfig.color} />
                </td>
                <td>{rfq.target_date ? new Date(rfq.target_date).toLocaleDateString() : '-'}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
