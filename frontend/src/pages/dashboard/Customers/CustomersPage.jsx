import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Users, Search, Plus, Building2, Mail, Phone, ExternalLink } from 'lucide-react';
import api from '../../../services/api';
import PageHeader from '../../../components/dashboard/PageHeader';
import './CustomersPage.css';

export default function CustomersPage() {
  const [customers, setCustomers] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchCustomers();
  }, []);

  const fetchCustomers = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.get('/api/v1/companies');
      const data = Array.isArray(res) ? res : (res?.data || []);
      setCustomers(data);
    } catch (err) {
      console.error('Failed to load customers:', err);
      setError('Unable to load customer directory. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const filtered = customers.filter((c) => {
    if (!searchQuery.trim()) return true;
    const q = searchQuery.toLowerCase();
    return (
      (c.name && c.name.toLowerCase().includes(q)) ||
      (c.contact_name && c.contact_name.toLowerCase().includes(q)) ||
      (c.contact_email && c.contact_email.toLowerCase().includes(q)) ||
      (c.industry && c.industry.toLowerCase().includes(q))
    );
  });

  return (
    <div className="customers-page">
      <div className="customers-header-row">
        <PageHeader
          title="Customer & Account Directory"
          subtitle="Manage verified shipper accounts, BCO consignees, and commercial credit profiles"
        />
        <button className="btn-add-customer" onClick={() => navigate('/dashboard/leads')}>
          <Plus size={16} />
          Convert Lead to Customer
        </button>
      </div>

      {/* Search Bar */}
      <div className="customer-search-card">
        <div className="search-field">
          <Search size={15} className="search-icon" />
          <input
            type="text"
            placeholder="Search by company name, contact, email, industry..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {error && (
        <div className="customer-error-banner">
          <span>⚠️ {error}</span>
          <button onClick={fetchCustomers} className="btn-retry">Retry</button>
        </div>
      )}

      {loading ? (
        <div className="customer-skeleton">
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton-customer-row" />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className="customers-empty-state">
          <div className="empty-icon-box">🏢</div>
          <h3>No customers yet</h3>
          <p>
            {searchQuery
              ? 'No customer accounts match your search query.'
              : 'Add your first verified client account or qualify sales prospects in Leads to automatically establish customer accounts.'}
          </p>
          <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', marginTop: '20px' }}>
            <button className="btn-primary-action" onClick={() => navigate('/dashboard/leads')}>
              View Sales Leads
            </button>
          </div>
        </div>
      ) : (
        <div className="customer-table-card">
          <table className="customers-table">
            <thead>
              <tr>
                <th>Company</th>
                <th>Industry / Domain</th>
                <th>Primary Contact</th>
                <th>Email & Phone</th>
                <th>Account Status</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((c) => (
                <tr key={c.id}>
                  <td>
                    <div className="company-cell">
                      <div className="company-avatar">
                        {c.name?.slice(0, 2).toUpperCase() || 'CO'}
                      </div>
                      <div>
                        <strong>{c.name}</strong>
                        {c.domain && <div className="domain-sub">{c.domain}</div>}
                      </div>
                    </div>
                  </td>
                  <td>{c.industry || 'General Freight'}</td>
                  <td>{c.contact_name || '—'}</td>
                  <td>
                    <div className="contact-info-col">
                      {c.contact_email && <span><Mail size={12} /> {c.contact_email}</span>}
                      {c.contact_phone && <span><Phone size={12} /> {c.contact_phone}</span>}
                      {!c.contact_email && !c.contact_phone && '—'}
                    </div>
                  </td>
                  <td>
                    <span className={`status-pill ${c.status?.toLowerCase() || 'active'}`}>
                      {c.status || 'ACTIVE'}
                    </span>
                  </td>
                  <td>{c.created_at ? new Date(c.created_at).toLocaleDateString() : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
