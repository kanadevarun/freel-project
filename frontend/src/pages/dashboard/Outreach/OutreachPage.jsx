import { useState, useEffect, useCallback } from 'react';
import PageHeader from '../../../components/dashboard/PageHeader';
import { listCampaigns, getOutreachAnalytics } from '../../../services/outreachService';
import { CAMPAIGN_STATUS } from './constants';
import CampaignList from './CampaignList';
import CampaignBuilder from './CampaignBuilder';
import EmailComposer from './EmailComposer';
import CampaignDetail from './CampaignDetail';
import OutreachDashboard from './OutreachDashboard';
import ActivityModal from './ActivityModal';
import ProspectList from './ProspectList';
import ProspectDetail from './ProspectDetail';
import FollowUpQueue from './FollowUpQueue';
import { BarChart2, Folder, Users, Clock, Sparkles, AlertTriangle, CheckCircle2, Mail, Play, Pause, FileText } from 'lucide-react';
import './OutreachPage.css';

export default function OutreachPage() {
  const [allCampaigns, setAllCampaigns] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showBuilder, setShowBuilder] = useState(false);
  const [selectedCampaignId, setSelectedCampaignId] = useState(null);
  const [selectedProspectId, setSelectedProspectId] = useState(null);
  const [activeTab, setActiveTab] = useState('OVERVIEW'); // 'OVERVIEW', 'CAMPAIGNS', 'PROSPECTS', 'FOLLOW_UPS', 'COMPOSER'

  // Prospect enrollment modal states
  const [showAddProspect, setShowAddProspect] = useState(false);
  const [selectedCampaignForProspect, setSelectedCampaignForProspect] = useState('');
  const [showAddLeadsModalForCampaign, setShowAddLeadsModalForCampaign] = useState(null);

  // Activity modal states
  const [showActivityModal, setShowActivityModal] = useState(false);
  const [editingActivity, setEditingActivity] = useState(null);
  const [refreshCount, setRefreshCount] = useState(0);

  const [analytics, setAnalytics] = useState(null);
  const [analyticsLoading, setAnalyticsLoading] = useState(true);

  const fetchAnalytics = useCallback(async () => {
    try {
      const stats = await getOutreachAnalytics();
      setAnalytics(stats);
    } catch (err) {
      console.error('Failed to load outreach metrics:', err);
    } finally {
      setAnalyticsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAnalytics();
  }, [fetchAnalytics, refreshCount]);

  const fetchCampaigns = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listCampaigns({ limit: 100 });
      setAllCampaigns(data?.data || []);
    } catch (err) {
      console.error('Failed to load campaigns:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCampaigns();
  }, [fetchCampaigns]);

  const counts = {
    total: allCampaigns.length,
    active: allCampaigns.filter(c => c.status === CAMPAIGN_STATUS.ACTIVE).length,
    draft: allCampaigns.filter(c => c.status === CAMPAIGN_STATUS.DRAFT).length,
    completed: allCampaigns.filter(c => c.status === CAMPAIGN_STATUS.COMPLETED).length,
  };

  // Render Campaign Detail view
  if (selectedCampaignId) {
    return (
      <CampaignDetail
        campaignId={selectedCampaignId}
        onBack={() => setSelectedCampaignId(null)}
        onCampaignUpdated={fetchCampaigns}
        onSelectProspect={(leadId) => {
          setSelectedCampaignId(null);
          setSelectedProspectId(leadId);
        }}
      />
    );
  }

  // Render Prospect Detail view
  if (selectedProspectId) {
    return (
      <ProspectDetail
        leadId={selectedProspectId}
        onBack={() => setSelectedProspectId(null)}
        onScheduleFollowUp={(e, p) => {
          e.stopPropagation();
          setEditingActivity({
            lead_id: p.lead_id,
            lead_company_name: p.company_name,
            activity_type: 'FOLLOW_UP',
          });
          setShowActivityModal(true);
        }}
        onSelectCampaign={(campaignId) => {
          setSelectedProspectId(null);
          setSelectedCampaignId(campaignId);
        }}
      />
    );
  }

  // Render headers dynamically based on the active tab
  const getHeaderDetails = () => {
    switch (activeTab) {
      case 'CAMPAIGNS':
        return {
          title: 'Campaigns',
          subtitle: 'Create and manage outreach sequences for your prospects.',
          actions: (
            <button
              className="outreach-btn outreach-btn-primary"
              style={{ padding: '9px 18px', fontSize: 13, borderRadius: 8, fontWeight: 600, marginTop: 22 }}
              onClick={() => setShowBuilder(true)}
            >
              + New Campaign
            </button>
          )
        };
      case 'PROSPECTS':
        return {
          title: 'Prospects',
          subtitle: 'Manage and track contacts currently being engaged through outreach campaigns.',
          actions: (
            <button
              className="outreach-btn outreach-btn-primary"
              style={{ padding: '9px 18px', fontSize: 13, borderRadius: 8, fontWeight: 600, marginTop: 22 }}
              onClick={() => setShowAddProspect(true)}
            >
              + Add Prospect
            </button>
          )
        };
      case 'FOLLOW_UPS':
        return {
          title: 'Follow-ups',
          subtitle: 'Stay on top of upcoming outreach actions and keep prospects moving through the pipeline.',
          actions: (
            <button
              className="outreach-btn outreach-btn-primary"
              style={{ padding: '9px 18px', fontSize: 13, borderRadius: 8, fontWeight: 600, marginTop: 22 }}
              onClick={() => {
                setEditingActivity(null);
                setShowActivityModal(true);
              }}
            >
              + New Follow-up
            </button>
          )
        };
      case 'COMPOSER':
        return {
          title: 'AI Email Composer',
          subtitle: 'Draft high-converting outreach email sequences with generative AI.',
          actions: null
        };
      case 'OVERVIEW':
      default:
        return {
          title: 'Outreach',
          subtitle: 'Manage outreach campaigns and follow-ups with prospects and customers.',
          actions: (
            <div style={{ display: 'flex', gap: 10, marginTop: 22 }}>
              <button
                className="activity-btn-outline"
                style={{ padding: '8px 16px', fontSize: 13, borderColor: '#CBD5E1', color: '#1E293B', fontWeight: 600 }}
                onClick={() => setShowBuilder(true)}
              >
                + New Campaign
              </button>
              <button
                className="outreach-btn outreach-btn-primary"
                style={{ padding: '9px 18px', fontSize: 13, borderRadius: 8, fontWeight: 600 }}
                onClick={() => {
                  setEditingActivity(null);
                  setShowActivityModal(true);
                }}
              >
                + New Outreach
              </button>
            </div>
          )
        };
    }
  };

  const header = getHeaderDetails();

  return (
    <div className="outreach-page">
      {/* ─── Page Header ─── */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 20 }}>
        <PageHeader
          title={header.title}
          subtitle={header.subtitle}
        />
        {header.actions}
      </div>

      {/* ─── Sticky KPI Stat Cards Bar ─── */}
      <div style={{ position: 'sticky', top: 0, zIndex: 10, background: '#F8FAFC', paddingBottom: 16, paddingTop: 4 }}>
        <div className="outreach-dashboard-stats-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', gap: 12 }}>
          {/* Total Campaigns */}
          <div className="dashboard-stat-card border-blue">
            <div className="dashboard-stat-icon-wrap bg-blue-tint text-blue">
              <Folder size={18} />
            </div>
            <div className="dashboard-stat-info">
              <span className="dashboard-stat-label">Total Campaigns</span>
              <div className="dashboard-stat-value-row">
                <span className="dashboard-stat-num">{counts.total}</span>
                <span className="trend-badge positive">Campaigns</span>
              </div>
            </div>
          </div>

          {/* Active Campaigns */}
          <div className="dashboard-stat-card border-green">
            <div className="dashboard-stat-icon-wrap bg-green-tint text-green">
              <Play size={18} />
            </div>
            <div className="dashboard-stat-info">
              <span className="dashboard-stat-label">Active Campaigns</span>
              <div className="dashboard-stat-value-row">
                <span className="dashboard-stat-num">{counts.active}</span>
                <span className="trend-badge positive">Running</span>
              </div>
            </div>
          </div>

          {/* Draft Campaigns */}
          <div className="dashboard-stat-card border-orange">
            <div className="dashboard-stat-icon-wrap bg-orange-tint text-orange">
              <FileText size={18} />
            </div>
            <div className="dashboard-stat-info">
              <span className="dashboard-stat-label">Draft Campaigns</span>
              <div className="dashboard-stat-value-row">
                <span className="dashboard-stat-num">{counts.draft}</span>
                <span className="trend-badge positive" style={{ background: '#FFFBEB', color: '#B45309' }}>Ready</span>
              </div>
            </div>
          </div>

          {/* Paused Campaigns */}
          <div className="dashboard-stat-card border-purple">
            <div className="dashboard-stat-icon-wrap bg-purple-tint text-purple">
              <Pause size={18} />
            </div>
            <div className="dashboard-stat-info">
              <span className="dashboard-stat-label">Paused Campaigns</span>
              <div className="dashboard-stat-value-row">
                <span className="dashboard-stat-num">{counts.paused}</span>
                <span className="trend-badge negative" style={{ background: '#F5F3FF', color: '#6D28D9' }}>On hold</span>
              </div>
            </div>
          </div>

          {/* Completed Campaigns */}
          <div className="dashboard-stat-card border-teal">
            <div className="dashboard-stat-icon-wrap" style={{ background: '#F0FDFA', color: '#14B8A6', width: 36, height: 36, borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <CheckCircle2 size={18} />
            </div>
            <div className="dashboard-stat-info">
              <span className="dashboard-stat-label">Completed</span>
              <div className="dashboard-stat-value-row">
                <span className="dashboard-stat-num">{counts.completed}</span>
                <span className="trend-badge positive" style={{ background: '#F0FDFA', color: '#0F766E' }}>Done</span>
              </div>
            </div>
          </div>

          {/* Total Prospects */}
          <div className="dashboard-stat-card border-cyan">
            <div className="dashboard-stat-icon-wrap" style={{ background: '#ECFEFF', color: '#06B6D4', width: 36, height: 36, borderRadius: 8, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Users size={18} />
            </div>
            <div className="dashboard-stat-info">
              <span className="dashboard-stat-label">Total Prospects</span>
              <div className="dashboard-stat-value-row">
                <span className="dashboard-stat-num">{analytics?.active_prospects ?? 0}</span>
                <span className="trend-badge positive" style={{ background: '#ECFEFF', color: '#0E7490' }}>Contacts</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', borderBottom: '1px solid #E2E8F0', marginBottom: 24, gap: 4, width: '100%', flexShrink: 0 }}>
        <button
          className={`status-tab-btn ${activeTab === 'OVERVIEW' ? 'active' : ''}`}
          style={{ 
            padding: '12px 18px', 
            fontSize: '13.5px', 
            fontWeight: 700,
            display: 'inline-flex', 
            alignItems: 'center',
            border: 'none',
            background: 'none',
            borderBottom: activeTab === 'OVERVIEW' ? '2px solid #2563EB' : '2px solid transparent',
            color: activeTab === 'OVERVIEW' ? '#2563EB' : '#64748B',
            cursor: 'pointer',
            transition: 'all 200ms ease',
            borderRadius: 0,
            marginBottom: '-1px'
          }}
          onClick={() => setActiveTab('OVERVIEW')}
        >
          <BarChart2 size={15} style={{ marginRight: 8 }} /> Overview
        </button>
        
        <button
          className={`status-tab-btn ${activeTab === 'CAMPAIGNS' ? 'active' : ''}`}
          style={{ 
            padding: '12px 18px', 
            fontSize: '13.5px', 
            fontWeight: 700,
            display: 'inline-flex', 
            alignItems: 'center',
            border: 'none',
            background: 'none',
            borderBottom: activeTab === 'CAMPAIGNS' ? '2px solid #2563EB' : '2px solid transparent',
            color: activeTab === 'CAMPAIGNS' ? '#2563EB' : '#64748B',
            cursor: 'pointer',
            transition: 'all 200ms ease',
            borderRadius: 0,
            marginBottom: '-1px'
          }}
          onClick={() => setActiveTab('CAMPAIGNS')}
        >
          <Folder size={15} style={{ marginRight: 8 }} /> Campaigns
          <span 
            className="tab-count-pill" 
            style={{ 
              marginLeft: 8, 
              background: activeTab === 'CAMPAIGNS' ? '#EFF6FF' : '#F1F5F9', 
              color: activeTab === 'CAMPAIGNS' ? '#2563EB' : '#64748B', 
              padding: '2px 8px', 
              borderRadius: 12, 
              fontSize: '11px', 
              fontWeight: 800 
            }}
          >
            {allCampaigns.length}
          </span>
        </button>

        <button
          className={`status-tab-btn ${activeTab === 'PROSPECTS' ? 'active' : ''}`}
          style={{ 
            padding: '12px 18px', 
            fontSize: '13.5px', 
            fontWeight: 700,
            display: 'inline-flex', 
            alignItems: 'center',
            border: 'none',
            background: 'none',
            borderBottom: activeTab === 'PROSPECTS' ? '2px solid #2563EB' : '2px solid transparent',
            color: activeTab === 'PROSPECTS' ? '#2563EB' : '#64748B',
            cursor: 'pointer',
            transition: 'all 200ms ease',
            borderRadius: 0,
            marginBottom: '-1px'
          }}
          onClick={() => setActiveTab('PROSPECTS')}
        >
          <Users size={15} style={{ marginRight: 8 }} /> Prospects
        </button>

        <button
          className={`status-tab-btn ${activeTab === 'FOLLOW_UPS' ? 'active' : ''}`}
          style={{ 
            padding: '12px 18px', 
            fontSize: '13.5px', 
            fontWeight: 700,
            display: 'inline-flex', 
            alignItems: 'center',
            border: 'none',
            background: 'none',
            borderBottom: activeTab === 'FOLLOW_UPS' ? '2px solid #2563EB' : '2px solid transparent',
            color: activeTab === 'FOLLOW_UPS' ? '#2563EB' : '#64748B',
            cursor: 'pointer',
            transition: 'all 200ms ease',
            borderRadius: 0,
            marginBottom: '-1px'
          }}
          onClick={() => setActiveTab('FOLLOW_UPS')}
        >
          <Clock size={15} style={{ marginRight: 8 }} /> Follow-ups
        </button>
      </div>

      {/* ─── Tab Content Workspace ─── */}
      {activeTab === 'OVERVIEW' && (
        <OutreachDashboard
          analytics={analytics}
          onSelectCampaign={(c) => setSelectedCampaignId(c.id)}
          onNewCampaignClick={() => setShowBuilder(true)}
          onViewCampaignsClick={() => setActiveTab('CAMPAIGNS')}
          onTabChange={setActiveTab}
          campaignsCount={counts}
          campaigns={allCampaigns}
          refreshCount={refreshCount}
          onEditActivity={(act) => {
            setEditingActivity(act);
            setShowActivityModal(true);
          }}
          onActivityChanged={() => setRefreshCount(prev => prev + 1)}
        />
      )}

      {activeTab === 'CAMPAIGNS' && (
        <CampaignList
          campaigns={allCampaigns}
          loading={loading}
          onCampaignsChanged={fetchCampaigns}
          onNewCampaignClick={() => setShowBuilder(true)}
          onSelectCampaign={(c) => setSelectedCampaignId(c.id)}
        />
      )}

      {activeTab === 'PROSPECTS' && (
        <ProspectList
          onSelectProspect={setSelectedProspectId}
          onScheduleFollowUp={(e, p) => {
            e.stopPropagation();
            setEditingActivity({
              lead_id: p.lead_id,
              lead_company_name: p.company_name,
              activity_type: 'FOLLOW_UP',
            });
            setShowActivityModal(true);
          }}
          refreshCount={refreshCount}
          onAddProspectClick={() => setShowAddProspect(true)}
          campaigns={allCampaigns}
        />
      )}

      {activeTab === 'FOLLOW_UPS' && (
        <FollowUpQueue
          onOpenProspect={setSelectedProspectId}
          onNewFollowUpClick={() => {
            setEditingActivity(null);
            setShowActivityModal(true);
          }}
        />
      )}


      {/* ─── Campaign Builder Modal ─── */}
      {showBuilder && (
        <CampaignBuilder
          onClose={() => setShowBuilder(false)}
          onCampaignCreated={(newId) => {
            fetchCampaigns();
            if (newId) setSelectedCampaignId(newId);
          }}
        />
      )}

      {/* ─── Activity Form Modal (Create/Edit) ─── */}
      {showActivityModal && (
        <ActivityModal
          activity={editingActivity}
          onClose={() => {
            setShowActivityModal(false);
            setEditingActivity(null);
          }}
          onSaveSuccess={() => {
            setShowActivityModal(false);
            setEditingActivity(null);
            setRefreshCount(prev => prev + 1);
          }}
        />
      )}

      {/* ─── Add Prospect (Campaign Selector overlay) ─── */}
      {showAddProspect && (
        <div className="outreach-modal-overlay" onClick={() => setShowAddProspect(false)}>
          <div className="outreach-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: 440 }}>
            <div className="modal-header">
              <div className="modal-title">Select Outreach Campaign</div>
              <button className="modal-close-btn" onClick={() => setShowAddProspect(false)}>✕</button>
            </div>
            <div style={{ padding: '10px 0 20px 0' }}>
              <p style={{ margin: '0 0 16px 0', fontSize: 13, color: '#64748B', lineHeight: 1.5 }}>
                Select the outreach campaign you want to enroll new prospects into.
              </p>
              
              {allCampaigns.filter(c => c.status === 'ACTIVE' || c.status === 'DRAFT').length === 0 ? (
                <div style={{ padding: 12, background: '#FEF2F2', border: '1px solid #FCA5A5', borderRadius: 8, fontSize: 13, color: '#991B1B', fontWeight: 600 }}>
                  No active or draft campaigns found. Please create a campaign first before adding prospects.
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                  <div>
                    <label style={{ display: 'block', fontSize: 12, fontWeight: 700, color: '#334155', marginBottom: 6 }}>Outreach Campaign</label>
                    <select
                      value={selectedCampaignForProspect}
                      onChange={e => setSelectedCampaignForProspect(e.target.value)}
                      style={{ width: '100%', padding: '9px 12px', border: '1px solid #CBD5E1', borderRadius: 8, fontSize: 13 }}
                    >
                      <option value="">-- Choose Campaign --</option>
                      {allCampaigns.filter(c => c.status === 'ACTIVE' || c.status === 'DRAFT').map(c => (
                        <option key={c.id} value={c.id}>{c.name} ({c.status})</option>
                      ))}
                    </select>
                  </div>
                  
                  <button
                    className="outreach-btn outreach-btn-primary"
                    disabled={!selectedCampaignForProspect}
                    onClick={() => {
                      setShowAddLeadsModalForCampaign(parseInt(selectedCampaignForProspect));
                      setShowAddProspect(false);
                      setSelectedCampaignForProspect('');
                    }}
                    style={{ width: '100%', padding: '10px', fontSize: 13, borderRadius: 8, fontWeight: 600, display: 'flex', justifyContent: 'center' }}
                  >
                    Select Campaign & Add Prospects
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ─── Add Leads Modal (CRM Directory check) ─── */}
      {showAddLeadsModalForCampaign && (
        <AddLeadsModal
          campaignId={showAddLeadsModalForCampaign}
          onClose={() => setShowAddLeadsModalForCampaign(null)}
          onLeadsAdded={() => {
            setShowAddLeadsModalForCampaign(null);
            setRefreshCount(prev => prev + 1);
          }}
        />
      )}
    </div>
  );
}