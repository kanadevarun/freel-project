import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { 
  Mail, Info, Plus, Users, User, MoreHorizontal, Settings, 
  CheckCircle2, Inbox, RefreshCcw, BrainCircuit, Workflow, Filter, ExternalLink, Activity,
  ChevronLeft, ChevronRight, Clock, FileText
} from 'lucide-react';
import toast from 'react-hot-toast';
import api from '../../../services/api';
import './EmailSettingsPage.css';
import ConfirmModal from './ConfirmModal';

export default function EmailSettingsPage() {
  const [mailboxes, setMailboxes] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 7;
  
  const [isConnectModalOpen, setIsConnectModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [openDropdownId, setOpenDropdownId] = useState(null);
  const [isConfigureModalOpen, setIsConfigureModalOpen] = useState(false);
  const [editingMailbox, setEditingMailbox] = useState(null);
  const [isConfirmDeleteOpen, setIsConfirmDeleteOpen] = useState(false);
  const [deletingMailboxId, setDeletingMailboxId] = useState(null);
  const [isActionLoading, setIsActionLoading] = useState(false);

  const [newMailbox, setNewMailbox] = useState({
    email: '',
    owner_name: '',
    mailbox_type: 'Shared / Team'
  });

  useEffect(() => {
    fetchData();
  }, []);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const oauthStatus = params.get('oauth');
    const message = params.get('message');
    if (oauthStatus === 'success') {
      toast.success('Gmail mailbox connected successfully!');
      // Clean query params
      window.history.replaceState({}, document.title, window.location.pathname);
      fetchData();
    } else if (oauthStatus === 'error') {
      toast.error(`Gmail connection failed: ${message || 'Unknown error'}`);
      window.history.replaceState({}, document.title, window.location.pathname);
      fetchData();
    }
  }, []);

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const mailboxesRes = await api.get('/api/v1/organizations/mailboxes');
      setMailboxes(mailboxesRes || []);
    } catch (error) {
      console.error(error);
      toast.error('Failed to load email settings.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleConnectGmail = async () => {
    try {
      const res = await api.get('/api/v1/organizations/mailboxes/connect/gmail');
      if (res && res.url) {
        window.location.href = res.url;
      } else {
        toast.error('Failed to get Gmail redirect link');
      }
    } catch (error) {
      toast.error('Failed to start Gmail connection flow');
    }
  };

  const handleDisconnect = async (id) => {
    setIsActionLoading(true);
    try {
      await api.post(`/api/v1/organizations/mailboxes/${id}/disconnect`);
      toast.success('Mailbox disconnected successfully');
      fetchData();
    } catch (e) {
      toast.error('Failed to disconnect mailbox');
    } finally {
      setIsActionLoading(false);
      setOpenDropdownId(null);
    }
  };

  const handleSync = async (id) => {
    const loadingToast = toast.loading('Synchronizing mailbox...');
    try {
      await api.post(`/api/v1/organizations/mailboxes/${id}/sync`);
      toast.success('Mailbox synchronization complete', { id: loadingToast });
      fetchData();
    } catch (e) {
      const errMsg = e.response?.data?.message || 'Sync failed';
      toast.error(`Failed to sync: ${errMsg}`, { id: loadingToast });
    }
    setOpenDropdownId(null);
  };

  const handleToggleProcessing = async (id, currentStatus, currentEnabled) => {
    try {
      const pause = currentEnabled;
      await api.post(`/api/v1/organizations/mailboxes/${id}/toggle-processing`, { pause });
      toast.success(pause ? 'Processing paused' : 'Processing resumed');
      fetchData();
    } catch (e) { toast.error('Failed to toggle processing'); }
    setOpenDropdownId(null);
  };

  const handleRemoveClick = (id) => {
    setDeletingMailboxId(id);
    setIsConfirmDeleteOpen(true);
    setOpenDropdownId(null);
  };

  const confirmRemove = async () => {
    if (!deletingMailboxId) return;
    setIsActionLoading(true);
    try {
      await api.delete(`/api/v1/organizations/mailboxes/${deletingMailboxId}`);
      toast.success('Mailbox removed');
      fetchData();
      setIsConfirmDeleteOpen(false);
    } catch (e) { toast.error('Failed to remove'); }
    setIsActionLoading(false);
    setDeletingMailboxId(null);
  };

  const handleConfigureClick = (mb) => {
    setEditingMailbox({
      ...mb,
      sync_frequency: mb.sync_frequency || 'Real-time',
      processing_enabled: mb.processing_enabled ?? true
    });
    setIsConfigureModalOpen(true);
    setOpenDropdownId(null);
  };

  const handleConfigureSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await api.put(`/api/v1/organizations/mailboxes/${editingMailbox.id}`, editingMailbox);
      toast.success('Mailbox updated successfully!');
      setIsConfigureModalOpen(false);
      fetchData();
    } catch (error) {
      toast.error('Failed to update mailbox.');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = () => setOpenDropdownId(null);
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, []);

  const handleConnectSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await api.post('/api/v1/organizations/mailboxes', newMailbox);
      toast.success('Mailbox connected successfully!');
      setIsConnectModalOpen(false);
      setNewMailbox({ email: '', owner_name: '', mailbox_type: 'Shared / Team' });
      fetchData();
    } catch (error) {
      console.error(error);
      toast.error('Failed to connect mailbox.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatRelative = (dateString) => {
    const diff = Math.floor((new Date() - new Date(dateString)) / 60000);
    if (diff < 1) return 'Just now';
    if (diff < 60) return `${diff} mins ago`;
    if (diff < 1440) return `${Math.floor(diff / 60)} hours ago`;
    return `${Math.floor(diff / 1440)} days ago`;
  };

  const formatAbsolute = (dateString) => {
    return new Date(dateString).toLocaleString('en-US', {
      month: 'short', day: 'numeric', year: 'numeric',
      hour: 'numeric', minute: '2-digit', hour12: true
    });
  };

  const getAvatarClass = (index) => {
    const classes = ['green', 'cyan', 'purple'];
    return classes[index % classes.length];
  };

  if (isLoading) {
    return (
      <div className="ws-page">
        <div style={{height: 100, background: '#f1f5f9', borderRadius: 8, animation: 'pulse 1.5s infinite'}}></div>
      </div>
    );
  }

  return (
    <div className="ws-page">
      <div className="ws-header">
        <h1 className="ws-title">
          <div className="ws-title-icon">
            <Mail size={24} />
          </div>
          Email Settings
        </h1>
        <p className="ws-subtitle">
          Connect and manage your email accounts. Our AI will automatically process customer logistics inquiries and track entire email conversations.
        </p>
      </div>

      {/* Summary Cards */}
      <div className="ws-summary-grid">
        <div className="ws-summary-card">
          <div className="ws-summary-icon-wrapper blue">
            <Mail size={24} />
          </div>
          <div className="ws-summary-content">
            <span className="ws-summary-label">Connected Mailboxes</span>
            <span className="ws-summary-value">{mailboxes.length}</span>
            <span className="ws-summary-desc">Total mailboxes connected</span>
          </div>
        </div>
        
        <div className="ws-summary-card">
          <div className="ws-summary-icon-wrapper green">
            <CheckCircle2 size={24} />
          </div>
          <div className="ws-summary-content">
            <span className="ws-summary-label">Processing Rules</span>
            <span className="ws-summary-value">Active</span>
            <span className="ws-summary-desc">Keywords & intent detection</span>
          </div>
        </div>

        <div className="ws-summary-card">
          <div className="ws-summary-icon-wrapper orange">
            <Workflow size={24} />
          </div>
          <div className="ws-summary-content">
            <span className="ws-summary-label">Thread Tracking</span>
            <span className="ws-summary-value">Enabled</span>
            <span className="ws-summary-desc">Conversation history sync</span>
          </div>
        </div>

        <div className="ws-summary-card">
          <div className="ws-summary-icon-wrapper purple">
            <RefreshCcw size={24} />
          </div>
          <div className="ws-summary-content">
            <span className="ws-summary-label">Email Sync</span>
            <span className="ws-summary-value">Operational</span>
            <span className="ws-summary-desc">Last checked: 2 mins ago</span>
          </div>
        </div>
      </div>

      <div className="ws-grid" style={{ display: 'block' }}>
        <div className="ws-main-col">
          <div className="ws-card">
            <div className="ws-card-header">
              <div className="ws-card-title-row">
                <h2 className="ws-card-title">Connected Mailboxes</h2>
                <div style={{ display: 'flex', gap: 12 }}>
                  <button className="ws-btn ws-btn-primary" style={{ background: '#db4437', borderColor: '#db4437', display: 'flex', alignItems: 'center', gap: 6 }} onClick={handleConnectGmail}>
                    <ExternalLink size={16} /> Connect Gmail
                  </button>
                  <button className="ws-btn ws-btn-outline" style={{ display: 'flex', alignItems: 'center', gap: 6 }} onClick={() => setIsConnectModalOpen(true)}>
                    <Plus size={16} /> Connect Custom Mailbox
                  </button>
                </div>
              </div>
              <p className="ws-card-desc">These mailboxes are monitored for customer inquiries and logistics-related conversations.</p>
            </div>
            
            <div className="ws-table-responsive">
              <table className="ws-mailbox-table">
                <thead>
                  <tr>
                    <th>Mailbox</th>
                    <th>Type</th>
                    <th>Status</th>
                    <th>Last Synced</th>
                    <th>Sync Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
              <tbody>
                {mailboxes.length === 0 ? (
                  <tr>
                    <td colSpan="5" style={{padding: '60px 24px', textAlign: 'center'}}>
                      <div style={{display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 16}}>
                        <div style={{
                          width: 64, height: 64, 
                          background: 'linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%)', 
                          borderRadius: '16px', display: 'flex', alignItems: 'center', 
                          justifyContent: 'center', color: '#2563eb',
                          boxShadow: '0 4px 14px rgba(37, 99, 235, 0.1)',
                          transform: 'rotate(-5deg)',
                          marginBottom: '8px'
                        }}>
                          <Mail size={32} />
                        </div>
                        <div style={{fontWeight: 600, color: '#0f172a', fontSize: '1.125rem'}}>No connected mailboxes</div>
                        <div style={{color: '#64748b', fontSize: '0.875rem', maxWidth: 360, margin: '0 auto', lineHeight: 1.6}}>
                          Connect your team or individual email accounts to let AI process logistics inquiries automatically and track entire threads.
                        </div>
                        <button className="ws-btn ws-btn-primary" style={{marginTop: 12, padding: '10px 20px', borderRadius: '8px'}} onClick={() => setIsConnectModalOpen(true)}>
                          <Plus size={16} /> Connect Mailbox
                        </button>
                      </div>
                    </td>
                  </tr>
                ) : (
                  mailboxes.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage).map((mb, idx) => {
                    const isGmail = mb.provider === 'GMAIL';
                    let dotColor = '#0891B2'; // Teal/Cyan active
                    if (mb.status === 'SYNCING') dotColor = '#2563eb';
                    if (mb.status === 'ERROR') dotColor = '#ef4444';
                    if (mb.status === 'DISCONNECTED') dotColor = '#64748b';

                    return (
                      <tr key={mb.id}>
                        <td>
                          <div className="ws-mailbox-cell">
                            <div className={`ws-avatar ${getAvatarClass(idx)}`}>
                              {mb.email.charAt(0).toUpperCase()}
                            </div>
                            <div>
                              <div className="ws-mailbox-email">
                                {mb.email} {mb.is_primary && <span className="ws-badge primary">Primary</span>}
                              </div>
                              <div className="ws-mailbox-owner">{mb.owner_name}</div>
                            </div>
                          </div>
                        </td>
                        <td>
                          <div className="ws-type-cell">
                            {mb.mailbox_type.includes('Shared') ? <Users size={14} /> : <User size={14} />}
                            {mb.mailbox_type}
                          </div>
                        </td>
                        <td>
                          <div className="ws-status" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                            <div className="ws-status-dot" style={{ backgroundColor: dotColor, width: 8, height: 8, borderRadius: '50%' }}></div>
                            <span style={{ fontWeight: 500 }}>{mb.status}</span>
                          </div>
                          <div style={{ fontSize: '0.75rem', color: '#64748b', marginTop: 2, textTransform: 'uppercase', fontWeight: 600 }}>{mb.provider || 'IMAP'}</div>
                        </td>
                        <td>
                          <div style={{ fontSize: '0.8125rem', color: '#334155' }}>
                            {mb.last_sync_success_at ? formatAbsolute(mb.last_sync_success_at) : 'Never synced'}
                          </div>
                        </td>
                        <td>
                          {mb.status === 'CONNECTED' && (
                            <div className="ws-status" style={{ color: '#0891B2', display: 'flex', alignItems: 'center', gap: 4 }}>
                              <CheckCircle2 size={14} />
                              <span>Up to date</span>
                            </div>
                          )}
                          {mb.status === 'SYNCING' && (
                            <div className="ws-status" style={{ color: '#2563eb', display: 'flex', alignItems: 'center', gap: 4 }}>
                              <RefreshCcw size={14} style={{ animation: 'spin 2s linear infinite' }} />
                              <span>Syncing...</span>
                            </div>
                          )}
                          {mb.status === 'DISCONNECTED' && (
                            <div className="ws-status" style={{ color: '#64748b', display: 'flex', alignItems: 'center', gap: 4 }}>
                              <Info size={14} />
                              <span>Disconnected</span>
                            </div>
                          )}
                          {mb.status === 'ERROR' && (
                            <div>
                              <div className="ws-status" style={{ color: '#ef4444', display: 'flex', alignItems: 'center', gap: 4, marginBottom: 2 }}>
                                <Info size={14} />
                                <span>Failed</span>
                              </div>
                              {mb.last_sync_error && (
                                <div style={{ fontSize: '0.75rem', color: '#ef4444', maxWidth: 180, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={mb.last_sync_error}>
                                  {mb.last_sync_error}
                                </div>
                              )}
                            </div>
                          )}
                        </td>
                        <td>
                          <div className="ws-actions-cell" style={{ position: 'relative' }}>
                            <button className="ws-btn ws-btn-outline" onClick={(e) => { e.stopPropagation(); handleConfigureClick(mb); }}>Configure</button>
                            <button className="ws-btn-icon" onClick={(e) => { e.stopPropagation(); setOpenDropdownId(openDropdownId === mb.id ? null : mb.id); }}>
                              <MoreHorizontal size={16} />
                            </button>
                            {openDropdownId === mb.id && (
                              <div style={{ position: 'absolute', top: '100%', right: 0, background: '#fff', border: '1px solid #e2e8f0', borderRadius: '8px', boxShadow: '0 4px 12px rgba(0,0,0,0.1)', zIndex: 50, minWidth: '180px', padding: '4px' }}>
                                {mb.status !== 'DISCONNECTED' && (
                                  <button onClick={() => handleSync(mb.id)} style={{ width: '100%', textAlign: 'left', padding: '8px 12px', background: 'transparent', border: 'none', cursor: 'pointer', fontSize: '0.875rem', color: '#334155', borderRadius: '4px', display: 'flex', alignItems: 'center', gap: 8 }}>
                                    <RefreshCcw size={14} /> Sync Now
                                  </button>
                                )}
                                {mb.status !== 'DISCONNECTED' && (
                                  <button onClick={() => handleToggleProcessing(mb.id, mb.status, mb.processing_enabled)} style={{ width: '100%', textAlign: 'left', padding: '8px 12px', background: 'transparent', border: 'none', cursor: 'pointer', fontSize: '0.875rem', color: '#334155', borderRadius: '4px', display: 'flex', alignItems: 'center', gap: 8 }}>
                                    <Activity size={14} /> {mb.processing_enabled ? 'Pause Processing' : 'Resume Processing'}
                                  </button>
                                )}
                                {isGmail && (mb.status === 'DISCONNECTED' || mb.status === 'ERROR') ? (
                                  <button onClick={handleConnectGmail} style={{ width: '100%', textAlign: 'left', padding: '8px 12px', background: 'transparent', border: 'none', cursor: 'pointer', fontSize: '0.875rem', color: '#0891B2', borderRadius: '4px', display: 'flex', alignItems: 'center', gap: 8, fontWeight: 500 }}>
                                    <ExternalLink size={14} /> Reconnect Mailbox
                                  </button>
                                ) : (
                                  mb.status !== 'DISCONNECTED' && (
                                    <button onClick={() => handleDisconnect(mb.id)} style={{ width: '100%', textAlign: 'left', padding: '8px 12px', background: 'transparent', border: 'none', cursor: 'pointer', fontSize: '0.875rem', color: '#e11d48', borderRadius: '4px', display: 'flex', alignItems: 'center', gap: 8 }}>
                                      <Info size={14} /> Disconnect Mailbox
                                    </button>
                                  )
                                )}
                                <div style={{ height: 1, background: '#e2e8f0', margin: '4px 0' }}></div>
                                <button onClick={() => handleRemoveClick(mb.id)} style={{ width: '100%', textAlign: 'left', padding: '8px 12px', background: 'transparent', border: 'none', cursor: 'pointer', fontSize: '0.875rem', color: '#ef4444', borderRadius: '4px', display: 'flex', alignItems: 'center', gap: 8 }}>
                                  Remove Mailbox
                                </button>
                              </div>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
            </div>

            {mailboxes.length > 0 && (
              <div className="ws-pagination">
                <div className="ws-pagination-text">
                  Showing {(currentPage - 1) * itemsPerPage + 1} to {Math.min(currentPage * itemsPerPage, mailboxes.length)} of {mailboxes.length} mailboxes
                </div>
                <div className="ws-pagination-controls">
                  <button 
                    className="ws-page-btn" 
                    disabled={currentPage === 1}
                    onClick={() => setCurrentPage(p => p - 1)}
                  >
                    <ChevronLeft size={16} />
                  </button>
                  <div className="ws-page-numbers">
                    {[...Array(Math.ceil(mailboxes.length / itemsPerPage))].map((_, i) => (
                      <div 
                        key={i} 
                        className={`ws-page-number ${currentPage === i + 1 ? 'active' : ''}`}
                        onClick={() => setCurrentPage(i + 1)}
                      >
                        {i + 1}
                      </div>
                    ))}
                  </div>
                  <button 
                    className="ws-page-btn" 
                    disabled={currentPage * itemsPerPage >= mailboxes.length}
                    onClick={() => setCurrentPage(p => p + 1)}
                  >
                    <ChevronRight size={16} />
                  </button>
                  <select className="ws-per-page-select">
                    <option>7 per page</option>
                  </select>
                </div>
              </div>
            )}
            
            <div style={{padding: '12px 24px', background: '#f8fafc', borderTop: '1px solid #e2e8f0', fontSize: '0.875rem', color: '#475569', display: 'flex', alignItems: 'center', gap: 8, borderBottomLeftRadius: 8, borderBottomRightRadius: 8}}>
              <Info size={16} color="#3b82f6" />
              AI will monitor all connected mailboxes and process only logistics-related conversations.
            </div>
          </div>

          {/* AI Email Processing Card */}
          <div className="ws-card" style={{marginTop: 24}}>
            <div className="ws-card-header borderless">
              <h2 className="ws-card-title">AI Email Processing</h2>
              <p className="ws-card-desc" style={{marginTop: 4}}>Configure how our AI identifies and processes email conversations.</p>
            </div>
            
            <div className="ws-toggle-row">
              <div className="ws-toggle-left">
                <div className="ws-toggle-icon purple">
                  <FileText size={20} />
                </div>
                <div className="ws-toggle-text">
                  <div className="ws-toggle-title">Process customer logistics inquiries</div>
                  <div className="ws-toggle-desc">AI will analyze incoming emails and process only logistics-related conversations.</div>
                </div>
              </div>
              <div className="ws-toggle-actions">
                <label className="ws-toggle-switch">
                  <input type="checkbox" defaultChecked />
                  <span className="ws-toggle-slider"></span>
                </label>
                <button className="ws-btn ws-btn-outline">Configure</button>
              </div>
            </div>

            <div className="ws-toggle-row">
              <div className="ws-toggle-left">
                <div className="ws-toggle-icon orange">
                  <Workflow size={20} />
                </div>
                <div className="ws-toggle-text">
                  <div className="ws-toggle-title">Track entire email threads</div>
                  <div className="ws-toggle-desc">Automatically track all replies, participants (To/CC), and the full conversation history.</div>
                </div>
              </div>
              <div className="ws-toggle-actions">
                <label className="ws-toggle-switch">
                  <input type="checkbox" defaultChecked />
                  <span className="ws-toggle-slider"></span>
                </label>
                <button className="ws-btn ws-btn-outline">Configure</button>
              </div>
            </div>

            <div className="ws-toggle-row">
              <div className="ws-toggle-left">
                <div className="ws-toggle-icon teal">
                  <Filter size={20} />
                </div>
                <div className="ws-toggle-text">
                  <div className="ws-toggle-title">Smart filtering</div>
                  <div className="ws-toggle-desc">AI filters out non-logistics emails like spam, newsletters, and internal discussions.</div>
                </div>
              </div>
              <div className="ws-toggle-actions">
                <label className="ws-toggle-switch">
                  <input type="checkbox" defaultChecked />
                  <span className="ws-toggle-slider"></span>
                </label>
                <button className="ws-btn ws-btn-outline">Configure</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      {isConnectModalOpen && createPortal(
        <div className="ws-modal-overlay">
          <div className="ws-modal">
            <div className="ws-modal-header">
              <h2 className="ws-modal-title">Connect Mailbox</h2>
              <button className="ws-modal-close" onClick={() => setIsConnectModalOpen(false)}>
                &times;
              </button>
            </div>
            <form onSubmit={handleConnectSubmit}>
              <div className="ws-modal-body">
                <div className="ws-form-group">
                  <label>Email Address</label>
                  <input 
                    type="email" 
                    required 
                    className="ws-input" 
                    placeholder="e.g. support@acmecorp.com"
                    value={newMailbox.email}
                    onChange={e => setNewMailbox({...newMailbox, email: e.target.value})}
                  />
                </div>
                <div className="ws-form-group">
                  <label>Owner Name</label>
                  <input 
                    type="text" 
                    required 
                    className="ws-input" 
                    placeholder="e.g. Sales Team"
                    value={newMailbox.owner_name}
                    onChange={e => setNewMailbox({...newMailbox, owner_name: e.target.value})}
                  />
                </div>
                <div className="ws-form-group">
                  <label>Mailbox Type</label>
                  <select 
                    className="ws-select"
                    value={newMailbox.mailbox_type}
                    onChange={e => setNewMailbox({...newMailbox, mailbox_type: e.target.value})}
                  >
                    <option value="Shared / Team">Shared / Team</option>
                    <option value="Individual">Individual</option>
                  </select>
                </div>
              </div>
              <div className="ws-modal-footer">
                <button type="button" className="ws-btn ws-btn-outline" onClick={() => setIsConnectModalOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="ws-btn ws-btn-primary" disabled={isSubmitting}>
                  {isSubmitting ? 'Connecting...' : 'Connect Mailbox'}
                </button>
              </div>
            </form>
          </div>
        </div>,
        document.body
      )}


      {isConfigureModalOpen && editingMailbox && createPortal(
        <div className="ws-modal-overlay">
          <div className="ws-modal">
            <div className="ws-modal-header">
              <h2 className="ws-modal-title">Configure Mailbox</h2>
              <button className="ws-modal-close" onClick={() => setIsConfigureModalOpen(false)}>
                &times;
              </button>
            </div>
            <form onSubmit={handleConfigureSubmit}>
              <div className="ws-modal-body">
                <div className="ws-form-group">
                  <label>Email Address</label>
                  <input 
                    type="email" 
                    className="ws-input" 
                    value={editingMailbox.email}
                    disabled
                    style={{backgroundColor: '#f1f5f9', cursor: 'not-allowed'}}
                  />
                  <div style={{fontSize: '0.75rem', color: '#64748b', marginTop: 4}}>Email cannot be changed.</div>
                </div>
                <div className="ws-form-group">
                  <label>Owner Name</label>
                  <input 
                    type="text" 
                    required 
                    className="ws-input" 
                    value={editingMailbox.owner_name}
                    onChange={e => setEditingMailbox({...editingMailbox, owner_name: e.target.value})}
                  />
                </div>
                <div className="ws-form-group">
                  <label>Mailbox Type</label>
                  <select 
                    className="ws-select"
                    value={editingMailbox.mailbox_type}
                    onChange={e => setEditingMailbox({...editingMailbox, mailbox_type: e.target.value})}
                  >
                    <option value="Shared / Team">Shared / Team</option>
                    <option value="Individual">Individual</option>
                  </select>
                </div>
                <div className="ws-form-group">
                  <label>Sync Frequency</label>
                  <select 
                    className="ws-select"
                    value={editingMailbox.sync_frequency}
                    onChange={e => setEditingMailbox({...editingMailbox, sync_frequency: e.target.value})}
                  >
                    <option value="Real-time">Real-time</option>
                    <option value="Hourly">Hourly</option>
                    <option value="Daily">Daily</option>
                  </select>
                </div>
                <div className="ws-form-group" style={{flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginTop: 8}}>
                  <div>
                    <label style={{marginBottom: 2, display: 'block'}}>Enable AI Processing</label>
                    <div style={{fontSize: '0.75rem', color: '#64748b'}}>Process inquiries and track threads</div>
                  </div>
                  <label className="ws-toggle-switch">
                    <input 
                      type="checkbox" 
                      checked={editingMailbox.processing_enabled}
                      onChange={e => setEditingMailbox({...editingMailbox, processing_enabled: e.target.checked})}
                    />
                    <span className="ws-toggle-slider"></span>
                  </label>
                </div>
              </div>
              <div className="ws-modal-footer">
                <button type="button" className="ws-btn ws-btn-outline" onClick={() => setIsConfigureModalOpen(false)}>
                  Cancel
                </button>
                <button type="submit" className="ws-btn ws-btn-primary" disabled={isSubmitting}>
                  {isSubmitting ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>
        </div>,
        document.body
      )}

      <ConfirmModal 
        isOpen={isConfirmDeleteOpen}
        title="Remove Mailbox"
        message="Are you sure you want to remove this mailbox? This action cannot be undone and AI processing will stop immediately."
        confirmText="Remove"
        confirmStyle="danger"
        isLoading={isActionLoading}
        onConfirm={confirmRemove}
        onCancel={() => setIsConfirmDeleteOpen(false)}
      />

    </div>
  );
}
