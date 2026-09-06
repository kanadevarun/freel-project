import { useEffect } from 'react';
import { Download, X } from 'lucide-react';
import './ConfirmModal.css';

export default function InvoiceHistoryModal({ isOpen, onClose, invoices }) {
  // Close on Escape key
  useEffect(() => {
    const handleEsc = (e) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const handleOverlayClick = (e) => {
    if (e.target.classList.contains('confirm-modal-overlay')) {
      onClose();
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    const d = new Date(dateString);
    return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  };

  return (
    <div className="confirm-modal-overlay" onClick={handleOverlayClick}>
      <div className="confirm-modal-container" style={{maxWidth: '800px', width: '90%'}}>
        
        <div className="confirm-modal-body" style={{padding: '24px', textAlign: 'left'}}>
          <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px'}}>
            <h3 className="confirm-modal-title" style={{margin: 0, textAlign: 'left'}}>Billing History</h3>
            <button 
              onClick={onClose} 
              style={{background: 'none', border: 'none', cursor: 'pointer', color: '#64748B'}}
            >
              <X size={20} />
            </button>
          </div>
          
          <div style={{maxHeight: '60vh', overflowY: 'auto'}}>
            <table className="invoices-table">
              <thead style={{position: 'sticky', top: 0, background: '#fff'}}>
                <tr>
                  <th>DATE</th>
                  <th>INVOICE #</th>
                  <th>AMOUNT</th>
                  <th>STATUS</th>
                  <th style={{textAlign: 'right'}}>ACTION</th>
                </tr>
              </thead>
              <tbody>
                {invoices && invoices.length > 0 ? (
                  invoices.map(inv => (
                    <tr key={inv.id}>
                      <td>{formatDate(inv.issued_at)}</td>
                      <td>{inv.number || 'Pending'}</td>
                      <td>${inv.amount_due?.toFixed(2)}</td>
                      <td>
                        <span className={`invoice-status ${inv.status.toLowerCase()}`}>
                          {inv.status}
                        </span>
                      </td>
                      <td style={{textAlign: 'right'}}>
                        {inv.invoice_pdf_url ? (
                          <a 
                            href={inv.invoice_pdf_url} 
                            target="_blank" 
                            rel="noopener noreferrer"
                            style={{color: '#2563EB', textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: '4px', fontSize: '0.875rem'}}
                          >
                            <Download size={14} /> Download
                          </a>
                        ) : (
                          <span style={{color: '#94A3B8', fontSize: '0.875rem'}}>Not available</span>
                        )}
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan="5" style={{textAlign: 'center', padding: '48px 0', color: '#64748B'}}>
                      No billing history available
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
