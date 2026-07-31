import { useState, useEffect, useCallback } from 'react';
import PageHeader from '../../../components/dashboard/PageHeader';
import { listCampaigns } from '../../../services/outreachService';
import { CAMPAIGN_STATUS, OUTREACH_TABS } from './constants';
import CampaignList from './CampaignList';
import CampaignBuilder from './CampaignBuilder';
import EmailComposer from './EmailComposer';
import './OutreachPage.css';

/**
 * OutreachPage — The main Outreach module page.
 *
 * Simple meaning: This is the "Outreach" screen that the sales team uses to:
 *   1. See and manage all email campaigns (Campaigns tab)
 *   2. Write AI-generated cold emails (Email Composer tab)
 *
 * At the top there are 4 stat cards (Total, Active, Draft, Completed).
 * Below that are two tabs to switch between the Campaign table and the AI Composer.
 */
export default function OutreachPage() {
  const [allCampaigns, setAllCampaigns] = useState([]);
  const [loading, setLoading]           = useState(true);
  const [activeTab, setActiveTab]       = useState(OUTREACH_TABS.CAMPAIGNS); // Default to Campaigns tab
  const [showBuilder, setShowBuilder]   = useState(false); // Controls CampaignBuilder modal

  // ── Data Fetching ─────────────────────────────────────────────────────────

  const fetchCampaigns = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listCampaigns({ limit: 100 });
      // Backend returns { campaigns: [...], total: N }
      setAllCampaigns(data?.campaigns || []);
    } catch (err) {
      console.error('Failed to load campaigns:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCampaigns();
  }, [fetchCampaigns]);

  // ── Derived Stats ─────────────────────────────────────────────────────────

  // Count campaigns per status for the stat cards
  const counts = {
    total:     allCampaigns.length,
    active:    allCampaigns.filter(c => c.status === CAMPAIGN_STATUS.ACTIVE).length,
    draft:     allCampaigns.filter(c => c.status === CAMPAIGN_STATUS.DRAFT).length,
    completed: allCampaigns.filter(c => c.status === CAMPAIGN_STATUS.COMPLETED).length,
  };

  // ── Tab Definitions ───────────────────────────────────────────────────────

  const TABS = [
    { id: OUTREACH_TABS.CAMPAIGNS, label: '📢 Campaigns', count: counts.total },
    { id: OUTREACH_TABS.COMPOSER,  label: '✨ AI Email Composer' },
  ];

  // ── Render ────────────────────────────────────────────────────────────────
  return (
    <div className="outreach-page">

      {/* Page Header */}
      <PageHeader
        title="Outreach"
        subtitle="Build campaigns and write AI-powered cold emails to convert your leads"
      />

      {/* ─── Stat Cards ─── */}
      <div className="outreach-stats-row">
        <div className="outreach-stat-card">
          <div className="outreach-stat-icon teal">📢</div>
          <div>
            <div className="outreach-stat-value">{counts.total}</div>
            <div className="outreach-stat-label">Total Campaigns</div>
          </div>
        </div>
        <div className="outreach-stat-card">
          <div className="outreach-stat-icon green">🟢</div>
          <div>
            <div className="outreach-stat-value">{counts.active}</div>
            <div className="outreach-stat-label">Active</div>
          </div>
        </div>
        <div className="outreach-stat-card">
          <div className="outreach-stat-icon indigo">📝</div>
          <div>
            <div className="outreach-stat-value">{counts.draft}</div>
            <div className="outreach-stat-label">Draft</div>
          </div>
        </div>
        <div className="outreach-stat-card">
          <div className="outreach-stat-icon amber">✅</div>
          <div>
            <div className="outreach-stat-value">{counts.completed}</div>
            <div className="outreach-stat-label">Completed</div>
          </div>
        </div>
      </div>

      {/* ─── Tab Bar ─── */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
        <div className="outreach-tabs">
          {TABS.map(tab => (
            <button
              key={tab.id}
              className={`outreach-tab ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
              {tab.count !== undefined && (
                <span style={{
                  background: activeTab === tab.id ? 'rgba(0,191,165,0.15)' : '#E2E8F0',
                  color: activeTab === tab.id ? '#00BFA5' : '#475569',
                  fontSize: 11, fontWeight: 600, padding: '1px 7px',
                  borderRadius: 99,
                }}>
                  {tab.count}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* "New Campaign" button — only shown in Campaigns tab */}
        {activeTab === OUTREACH_TABS.CAMPAIGNS && (
          <>
            <div className="outreach-bar-spacer" />
            <button
              className="outreach-btn outreach-btn-primary"
              onClick={() => setShowBuilder(true)}
            >
              + New Campaign
            </button>
          </>
        )}
      </div>

      {/* ─── Tab Content ─── */}
      {activeTab === OUTREACH_TABS.CAMPAIGNS && (
        <CampaignList
          campaigns={allCampaigns}
          loading={loading}
          onCampaignsChanged={fetchCampaigns}
        />
      )}

      {activeTab === OUTREACH_TABS.COMPOSER && (
        <EmailComposer />
      )}

      {/* ─── Campaign Builder Modal ─── */}
      {showBuilder && (
        <CampaignBuilder
          onClose={() => setShowBuilder(false)}
          onCampaignCreated={fetchCampaigns}
        />
      )}

    </div>
  );
}
