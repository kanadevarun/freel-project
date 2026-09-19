import React, { useState, useEffect, useCallback } from 'react';
import {
  X,
  Brain,
  Sparkles,
  RefreshCw,
  Search,
  Filter,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Edit3,
  Trash2,
  Flag,
  Shield,
  Layers,
  ChevronRight,
  Database,
  TrendingUp,
  Activity,
  Award,
  AlertOctagon,
  Eye,
  Info,
  Check,
  Tag,
  ArrowRight
} from 'lucide-react';
import autonomyService from '../../services/autonomyService';

export default function AgentMemoryLearningDrawer({
  isOpen,
  onClose,
  initialCategory = 'ALL',
  onMemoryUpdated
}) {
  const [activeTab, setActiveTab] = useState('memories'); // 'memories' | 'patterns' | 'outcomes' | 'summary'
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [successMsg, setSuccessMsg] = useState(null);

  // Data states
  const [memories, setMemories] = useState([]);
  const [patterns, setPatterns] = useState([]);
  const [outcomes, setOutcomes] = useState([]);
  const [summary, setSummary] = useState(null);

  // Filters
  const [categoryFilter, setCategoryFilter] = useState(initialCategory);
  const [includeStale, setIncludeStale] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedMemory, setSelectedMemory] = useState(null);

  // Modals / Actions
  const [correctionModalOpen, setCorrectionModalOpen] = useState(false);
  const [correctionText, setCorrectionText] = useState('');
  const [correctionReason, setCorrectionReason] = useState('');

  const [invalidationModalOpen, setInvalidationModalOpen] = useState(false);
  const [invalidationReason, setInvalidationReason] = useState('');

  const loadAllData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [sumRes, memRes, patRes, outRes] = await Promise.allSettled([
        autonomyService.getMemoryLearningSummary(),
        autonomyService.listMemories({
          category: categoryFilter !== 'ALL' ? categoryFilter : undefined,
          include_stale: includeStale,
          limit: 100
        }),
        autonomyService.listPatterns({ limit: 50 }),
        autonomyService.listOutcomes({ limit: 50 })
      ]);

      if (sumRes.status === 'fulfilled') {
        setSummary(sumRes.value?.summary || sumRes.value?.data?.summary || null);
      }
      if (memRes.status === 'fulfilled') {
        const memList = memRes.value?.memories || memRes.value?.data?.memories || [];
        setMemories(memList);
      }
      if (patRes.status === 'fulfilled') {
        const patList = patRes.value?.patterns || patRes.value?.data?.patterns || [];
        setPatterns(patList);
      }
      if (outRes.status === 'fulfilled') {
        const outList = outRes.value?.outcomes || outRes.value?.data?.outcomes || [];
        setOutcomes(outList);
      }
    } catch (err) {
      console.error('Failed loading memory learning data:', err);
      setError('Unable to load memory items or telemetry. Degrading gracefully.');
    } finally {
      setLoading(false);
    }
  }, [categoryFilter, includeStale]);

  useEffect(() => {
    if (isOpen) {
      loadAllData();
    }
  }, [isOpen, loadAllData]);

  // Handle human correction
  const handleCorrectMemory = async () => {
    if (!selectedMemory || !correctionText.trim()) return;
    setActionLoading(true);
    setError(null);
    try {
      await autonomyService.correctMemory(selectedMemory.id, {
        corrected_content: correctionText.trim(),
        reason: correctionReason.trim() || 'Manual operator refinement'
      });
      setSuccessMsg('Memory item corrected and audited successfully.');
      setCorrectionModalOpen(false);
      setCorrectionText('');
      setCorrectionReason('');
      await loadAllData();
      if (onMemoryUpdated) onMemoryUpdated();
    } catch (err) {
      setError(err?.response?.data?.message || 'Failed to submit memory correction.');
    } finally {
      setActionLoading(false);
    }
  };

  // Handle invalidation
  const handleInvalidateMemory = async () => {
    if (!selectedMemory) return;
    setActionLoading(true);
    setError(null);
    try {
      await autonomyService.invalidateMemory(selectedMemory.id, {
        reason: invalidationReason.trim() || 'Invalidated by operator'
      });
      setSuccessMsg('Memory item marked as invalidated.');
      setInvalidationModalOpen(false);
      setInvalidationReason('');
      setSelectedMemory(null);
      await loadAllData();
      if (onMemoryUpdated) onMemoryUpdated();
    } catch (err) {
      setError(err?.response?.data?.message || 'Failed to invalidate memory item.');
    } finally {
      setActionLoading(false);
    }
  };

  // Handle flag unreliable
  const handleFlagUnreliable = async (mem) => {
    setActionLoading(true);
    setError(null);
    try {
      await autonomyService.flagMemoryUnreliable(mem.id, 'Operator marked as unreliable observation');
      setSuccessMsg(`Memory #${mem.id} flagged as unreliable with decayed confidence.`);
      await loadAllData();
      if (onMemoryUpdated) onMemoryUpdated();
    } catch (err) {
      setError(err?.response?.data?.message || 'Failed to flag memory.');
    } finally {
      setActionLoading(false);
    }
  };

  // Handle pattern detection trigger
  const handleTriggerPatternDetection = async () => {
    setActionLoading(true);
    setError(null);
    try {
      const resp = await autonomyService.detectPatterns();
      const count = resp?.patterns?.total_patterns || resp?.data?.patterns?.total_patterns || 0;
      setSuccessMsg(`Pattern detection complete. Identified ${count} recurring pattern(s).`);
      await loadAllData();
    } catch (err) {
      setError(err?.response?.data?.message || 'Pattern detection failed.');
    } finally {
      setActionLoading(false);
    }
  };

  // Filter memories by search
  const filteredMemories = memories.filter((m) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      (m.title && m.title.toLowerCase().includes(q)) ||
      (m.content && m.content.toLowerCase().includes(q)) ||
      (m.category && m.category.toLowerCase().includes(q)) ||
      (m.entity_id && m.entity_id.toLowerCase().includes(q))
    );
  });

  const getProvenanceBadge = (prov) => {
    switch (prov) {
      case 'HUMAN_ENTERED':
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-indigo-50 text-indigo-700 border border-indigo-200">Human Entered</span>;
      case 'SYSTEM_DERIVED':
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200">System Verified</span>;
      case 'AI_DERIVED':
      default:
        return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-purple-50 text-purple-700 border border-purple-200">AI Learned</span>;
    }
  };

  const getConfidenceBadge = (conf) => {
    const num = typeof conf === 'number' ? conf : 0.75;
    if (num >= 0.85) {
      return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">HIGH ({Math.round(num * 100)}%)</span>;
    } else if (num >= 0.6) {
      return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">MEDIUM ({Math.round(num * 100)}%)</span>;
    }
    return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200">LOW ({Math.round(num * 100)}%)</span>;
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-slate-900/40 backdrop-blur-sm flex justify-end transition-opacity animate-fade-in" role="dialog" aria-modal="true">
      <div className="w-full max-w-4xl bg-white shadow-2xl h-full flex flex-col border-l border-slate-200">
        
        {/* Drawer Header */}
        <div className="px-6 py-4 border-b border-slate-200 bg-slate-50/80 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-indigo-600 text-white rounded-xl shadow-sm">
              <Brain className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-bold text-slate-900">Agent Memory & Outcome Learning</h2>
                <span className="px-2 py-0.5 bg-indigo-100 text-indigo-700 text-xs font-semibold rounded-full">Phase 5 Task 5.13</span>
              </div>
              <p className="text-xs text-slate-500">
                Safe, tenant-isolated contextual AI memory derived from verified operational outcomes.
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={loadAllData}
              disabled={loading || actionLoading}
              className="p-2 text-slate-500 hover:text-slate-800 hover:bg-slate-100 rounded-lg transition"
              title="Refresh Data"
              aria-label="Refresh Data"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={onClose}
              className="p-2 text-slate-500 hover:text-slate-800 hover:bg-slate-100 rounded-lg transition"
              title="Close Drawer"
              aria-label="Close Drawer"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Global Alert / Notifications */}
        {error && (
          <div className="px-6 py-2.5 bg-rose-50 border-b border-rose-200 text-rose-700 text-xs flex items-center justify-between">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
            <button onClick={() => setError(null)} className="text-rose-500 hover:text-rose-800"><X className="w-3.5 h-3.5" /></button>
          </div>
        )}
        {successMsg && (
          <div className="px-6 py-2.5 bg-emerald-50 border-b border-emerald-200 text-emerald-700 text-xs flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 flex-shrink-0" />
              <span>{successMsg}</span>
            </div>
            <button onClick={() => setSuccessMsg(null)} className="text-emerald-500 hover:text-emerald-800"><X className="w-3.5 h-3.5" /></button>
          </div>
        )}

        {/* Navigation Tabs */}
        <div className="px-6 border-b border-slate-200 bg-white flex items-center gap-2 overflow-x-auto">
          <button
            onClick={() => setActiveTab('memories')}
            className={`py-3 px-4 text-xs font-semibold border-b-2 flex items-center gap-2 transition ${
              activeTab === 'memories'
                ? 'border-indigo-600 text-indigo-700 bg-indigo-50/30'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Database className="w-4 h-4" />
            <span>Operational Memories</span>
            <span className="ml-1 px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-700 text-[10px]">
              {memories.length}
            </span>
          </button>
          <button
            onClick={() => setActiveTab('patterns')}
            className={`py-3 px-4 text-xs font-semibold border-b-2 flex items-center gap-2 transition ${
              activeTab === 'patterns'
                ? 'border-indigo-600 text-indigo-700 bg-indigo-50/30'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <TrendingUp className="w-4 h-4" />
            <span>Learned Patterns</span>
            <span className="ml-1 px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-700 text-[10px]">
              {patterns.length}
            </span>
          </button>
          <button
            onClick={() => setActiveTab('outcomes')}
            className={`py-3 px-4 text-xs font-semibold border-b-2 flex items-center gap-2 transition ${
              activeTab === 'outcomes'
                ? 'border-indigo-600 text-indigo-700 bg-indigo-50/30'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Award className="w-4 h-4" />
            <span>Verified Outcomes</span>
            <span className="ml-1 px-1.5 py-0.5 rounded-full bg-slate-100 text-slate-700 text-[10px]">
              {outcomes.length}
            </span>
          </button>
          <button
            onClick={() => setActiveTab('summary')}
            className={`py-3 px-4 text-xs font-semibold border-b-2 flex items-center gap-2 transition ${
              activeTab === 'summary'
                ? 'border-indigo-600 text-indigo-700 bg-indigo-50/30'
                : 'border-transparent text-slate-600 hover:text-slate-900'
            }`}
          >
            <Activity className="w-4 h-4" />
            <span>Quality & Telemetry</span>
          </button>
        </div>

        {/* Drawer Body Content */}
        <div className="flex-1 overflow-y-auto p-6 bg-slate-50">
          
          {/* TAB 1: OPERATIONAL MEMORIES */}
          {activeTab === 'memories' && (
            <div className="space-y-4">
              {/* Filter and Search Bar */}
              <div className="flex flex-wrap items-center justify-between gap-3 bg-white p-3.5 rounded-xl border border-slate-200 shadow-sm">
                <div className="flex items-center gap-2 flex-1 min-w-[200px]">
                  <Search className="w-4 h-4 text-slate-400" />
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search memories by keyword, entity ID, or category..."
                    className="w-full text-xs bg-transparent border-0 focus:ring-0 focus:outline-none text-slate-800 placeholder-slate-400"
                  />
                </div>
                <div className="flex items-center gap-3 text-xs">
                  <select
                    value={categoryFilter}
                    onChange={(e) => setCategoryFilter(e.target.value)}
                    className="border border-slate-200 rounded-lg px-2.5 py-1.5 text-xs text-slate-700 bg-slate-50 focus:bg-white focus:outline-none focus:ring-1 focus:ring-indigo-500"
                  >
                    <option value="ALL">All Categories</option>
                    <option value="OPERATIONAL">Operational</option>
                    <option value="CUSTOMER">Customer</option>
                    <option value="CARRIER">Carrier</option>
                    <option value="PRICING">Pricing</option>
                    <option value="FINANCE">Finance</option>
                    <option value="COMPLIANCE">Compliance</option>
                  </select>

                  <label className="flex items-center gap-1.5 cursor-pointer text-slate-600 select-none">
                    <input
                      type="checkbox"
                      checked={includeStale}
                      onChange={(e) => setIncludeStale(e.target.checked)}
                      className="rounded text-indigo-600 focus:ring-indigo-500 w-3.5 h-3.5"
                    />
                    <span>Include Stale/Invalid</span>
                  </label>
                </div>
              </div>

              {/* Memory Cards Grid */}
              {filteredMemories.length === 0 ? (
                <div className="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-500 text-xs">
                  <Brain className="w-8 h-8 mx-auto text-slate-300 mb-2" />
                  <p className="font-semibold text-slate-700">No memories found</p>
                  <p className="text-slate-400 mt-1">
                    No contextual memories recorded matching this filter.
                  </p>
                </div>
              ) : (
                <div className="space-y-3">
                  {filteredMemories.map((mem) => (
                    <div
                      key={mem.id}
                      className={`bg-white rounded-xl border p-4 transition-all shadow-sm ${
                        mem.is_stale || mem.status === 'INVALIDATED'
                          ? 'border-slate-200 bg-slate-50/60 opacity-80'
                          : 'border-slate-200 hover:border-indigo-300'
                      }`}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="space-y-1 flex-1">
                          <div className="flex items-center flex-wrap gap-2">
                            <span className="font-bold text-xs text-slate-900">{mem.title}</span>
                            <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-slate-100 text-slate-700">
                              {mem.category}
                            </span>
                            {getProvenanceBadge(mem.provenance_type)}
                            {getConfidenceBadge(mem.confidence)}
                            {mem.is_stale && (
                              <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-amber-100 text-amber-800">
                                STALE
                              </span>
                            )}
                            {mem.status === 'INVALIDATED' && (
                              <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-rose-100 text-rose-800">
                                INVALIDATED
                              </span>
                            )}
                            {mem.conflict_status === 'CORRECTED' && (
                              <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-blue-100 text-blue-800">
                                CORRECTED
                              </span>
                            )}
                          </div>
                          <p className="text-xs text-slate-700 leading-relaxed mt-1">{mem.content}</p>
                          
                          {/* Original content if corrected */}
                          {mem.original_content && (
                            <div className="mt-2 text-[11px] text-slate-500 bg-slate-50 p-2 rounded border border-slate-200">
                              <span className="font-semibold text-slate-600">Prior content before correction:</span> {mem.original_content}
                            </div>
                          )}

                          {/* Metadata row */}
                          <div className="flex items-center gap-4 text-[11px] text-slate-400 pt-2">
                            {mem.entity_id && (
                              <span>Entity: <strong className="text-slate-600">{mem.entity_type ? `${mem.entity_type}: ` : ''}{mem.entity_id}</strong></span>
                            )}
                            <span>Observed: <strong className="text-slate-600">{mem.times_observed || 1}x</strong></span>
                            <span>Weight: <strong className="text-slate-600">{mem.recency_weight || 1.0}</strong></span>
                            {mem.outcome_id && (
                              <span>Outcome: <code className="text-slate-600">{mem.outcome_id}</code></span>
                            )}
                          </div>
                        </div>

                        {/* Interactive actions for this memory */}
                        <div className="flex items-center gap-1.5 flex-shrink-0">
                          <button
                            onClick={() => {
                              setSelectedMemory(mem);
                              setCorrectionText(mem.content);
                              setCorrectionModalOpen(true);
                            }}
                            className="p-1.5 text-slate-500 hover:text-indigo-600 hover:bg-indigo-50 rounded transition"
                            title="Correct Memory (Human Audit)"
                            aria-label={`Correct memory #${mem.id}`}
                          >
                            <Edit3 className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => handleFlagUnreliable(mem)}
                            className="p-1.5 text-slate-500 hover:text-amber-600 hover:bg-amber-50 rounded transition"
                            title="Flag as Unreliable"
                            aria-label={`Flag memory #${mem.id} as unreliable`}
                          >
                            <Flag className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => {
                              setSelectedMemory(mem);
                              setInvalidationModalOpen(true);
                            }}
                            className="p-1.5 text-slate-500 hover:text-rose-600 hover:bg-rose-50 rounded transition"
                            title="Invalidate Memory"
                            aria-label={`Invalidate memory #${mem.id}`}
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* TAB 2: LEARNED PATTERNS */}
          {activeTab === 'patterns' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between bg-white p-4 rounded-xl border border-slate-200">
                <div>
                  <h3 className="text-xs font-bold text-slate-900">Synthesized Domain Patterns</h3>
                  <p className="text-xs text-slate-500">
                    Recurring behavioral tendencies clustered across verified outcomes (carriers, exceptions, and workflows).
                  </p>
                </div>
                <button
                  onClick={handleTriggerPatternDetection}
                  disabled={actionLoading}
                  className="px-3.5 py-1.5 bg-indigo-600 text-white rounded-lg text-xs font-semibold hover:bg-indigo-700 transition flex items-center gap-1.5 shadow-sm"
                >
                  <Sparkles className="w-3.5 h-3.5" />
                  <span>Run Pattern Detection</span>
                </button>
              </div>

              {patterns.length === 0 ? (
                <div className="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-500 text-xs">
                  <TrendingUp className="w-8 h-8 mx-auto text-slate-300 mb-2" />
                  <p className="font-semibold text-slate-700">No domain patterns detected yet</p>
                  <p className="text-slate-400 mt-1">
                    Patterns require recurring verified outcomes across shipments or carriers.
                  </p>
                </div>
              ) : (
                <div className="space-y-3">
                  {patterns.map((pat) => (
                    <div key={pat.id || pat.pattern_id} className="bg-white rounded-xl border border-slate-200 p-4 shadow-sm">
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <div className="flex items-center gap-2">
                            <h4 className="text-xs font-bold text-slate-900">{pat.title}</h4>
                            <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-indigo-50 text-indigo-700 border border-indigo-200">
                              {pat.pattern_type}
                            </span>
                            <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                              {Math.round((pat.success_rate || 1.0) * 100)}% Success
                            </span>
                          </div>
                          <p className="text-xs text-slate-700 mt-1">{pat.description}</p>
                          {pat.recommended_strategy && (
                            <div className="mt-2 text-xs bg-indigo-50/60 text-indigo-900 p-2 rounded-lg border border-indigo-100 flex items-start gap-2">
                              <Sparkles className="w-4 h-4 text-indigo-600 flex-shrink-0 mt-0.5" />
                              <div>
                                <span className="font-bold">Recommended Strategy:</span> {pat.recommended_strategy}
                              </div>
                            </div>
                          )}
                          <div className="flex items-center gap-4 text-[11px] text-slate-400 mt-3">
                            <span>Entity: <strong className="text-slate-600">{pat.entity_identifier}</strong></span>
                            <span>Observations: <strong className="text-slate-600">{pat.supporting_observations}x</strong></span>
                            <span>Confidence: <strong className="text-slate-600">{pat.confidence}</strong></span>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* TAB 3: VERIFIED OUTCOMES */}
          {activeTab === 'outcomes' && (
            <div className="space-y-4">
              <div className="bg-white p-4 rounded-xl border border-slate-200">
                <h3 className="text-xs font-bold text-slate-900">Recorded Operational Outcomes</h3>
                <p className="text-xs text-slate-500">
                  Authoritative comparison between expected plan targets and actual business results.
                </p>
              </div>

              {outcomes.length === 0 ? (
                <div className="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-500 text-xs">
                  <Award className="w-8 h-8 mx-auto text-slate-300 mb-2" />
                  <p className="font-semibold text-slate-700">No outcomes recorded</p>
                  <p className="text-slate-400 mt-1">Outcomes will appear as plans execute and verify.</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {outcomes.map((out) => (
                    <div key={out.id || out.outcome_id} className="bg-white rounded-xl border border-slate-200 p-4 shadow-sm">
                      <div className="flex items-start justify-between gap-3">
                        <div className="space-y-1 flex-1">
                          <div className="flex items-center gap-2">
                            <span className="font-bold text-xs text-slate-900">{out.outcome_type}</span>
                            <span className={`px-2 py-0.5 rounded text-[10px] font-semibold ${
                              out.status === 'SUCCESS'
                                ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                                : out.status === 'PARTIAL_SUCCESS'
                                ? 'bg-amber-50 text-amber-700 border border-amber-200'
                                : 'bg-rose-50 text-rose-700 border border-rose-200'
                            }`}>
                              {out.status}
                            </span>
                            {out.is_verified ? (
                              <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">
                                Verified
                              </span>
                            ) : (
                              <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-slate-100 text-slate-600">
                                Unverified
                              </span>
                            )}
                          </div>

                          <div className="grid grid-cols-1 md:grid-cols-2 gap-2 mt-2 text-xs">
                            <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                              <span className="font-semibold text-slate-500 block text-[10px] uppercase tracking-wider">Expected:</span>
                              <span className="text-slate-700">{out.expected_result || 'N/A'}</span>
                            </div>
                            <div className="p-2.5 bg-slate-50 rounded-lg border border-slate-200">
                              <span className="font-semibold text-slate-500 block text-[10px] uppercase tracking-wider">Actual Result:</span>
                              <span className="text-slate-700">{out.actual_result || 'N/A'}</span>
                            </div>
                          </div>

                          {out.failure_category && (
                            <div className="text-xs text-rose-700 mt-1">
                              <strong>Failure Category:</strong> {out.failure_category}
                            </div>
                          )}

                          <div className="flex items-center gap-4 text-[11px] text-slate-400 pt-2">
                            <span>Entity: <strong className="text-slate-600">{out.source_entity_type}: {out.source_entity_id}</strong></span>
                            {out.verification_method && <span>Method: <strong className="text-slate-600">{out.verification_method}</strong></span>}
                            <span>Correlation: <code className="text-slate-600">{out.correlation_id}</code></span>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* TAB 4: QUALITY & TELEMETRY */}
          {activeTab === 'summary' && summary && (
            <div className="space-y-6">
              {/* Telemetry KPI Cards */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
                  <span className="text-xs text-slate-500 font-medium">Total Memories</span>
                  <div className="text-xl font-bold text-slate-900 mt-1">{summary.total_memories || 0}</div>
                  <span className="text-[10px] text-emerald-600 font-medium">{summary.active_memories || 0} active</span>
                </div>
                <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
                  <span className="text-xs text-slate-500 font-medium">Outcomes Logged</span>
                  <div className="text-xl font-bold text-slate-900 mt-1">{summary.total_outcomes || 0}</div>
                  <span className="text-[10px] text-blue-600 font-medium">{summary.verified_outcomes || 0} verified</span>
                </div>
                <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
                  <span className="text-xs text-slate-500 font-medium">Outcome Success</span>
                  <div className="text-xl font-bold text-emerald-600 mt-1">
                    {Math.round((summary.overall_success_rate || 0) * 100)}%
                  </div>
                  <span className="text-[10px] text-slate-400">{summary.successful_outcomes || 0} successes</span>
                </div>
                <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-sm">
                  <span className="text-xs text-slate-500 font-medium">Acceptance Rate</span>
                  <div className="text-xl font-bold text-indigo-600 mt-1">
                    {Math.round((summary.recommendation_accept_pct || 1.0) * 100)}%
                  </div>
                  <span className="text-[10px] text-slate-400">Human approved</span>
                </div>
              </div>

              {/* Memory Categories Breakdown */}
              <div className="bg-white p-5 rounded-xl border border-slate-200 shadow-sm space-y-3">
                <h4 className="text-xs font-bold text-slate-900">Memory Distribution by Domain</h4>
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                  {Object.entries(summary.memory_category_counts || {}).map(([cat, count]) => (
                    <div key={cat} className="p-3 bg-slate-50 rounded-lg border border-slate-200 flex items-center justify-between">
                      <span className="text-xs font-semibold text-slate-700">{cat}</span>
                      <span className="px-2 py-0.5 rounded-full text-xs font-bold bg-white text-indigo-700 border border-slate-200">
                        {count}
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Safety & Isolation Invariants */}
              <div className="bg-indigo-50/50 p-4 rounded-xl border border-indigo-100 space-y-2 text-xs text-indigo-900">
                <div className="flex items-center gap-2 font-bold text-indigo-950">
                  <Shield className="w-4 h-4 text-indigo-600" />
                  <span>Authoritative Governance & Safety Invariants</span>
                </div>
                <ul className="list-disc pl-5 space-y-1 text-slate-700 text-[11px]">
                  <li>Tenant isolation strictly enforced in Go database queries (cross-tenant leakage impossible).</li>
                  <li>Memory is contextual evidence only — authoritative state and current customer instructions always take precedence.</li>
                  <li>Prompt injection patterns (e.g. override, escalate privileges) sanitized on both ingestion and retrieval.</li>
                  <li>Learning never autonomously modifies business rules, policies, permissions, or autonomy levels.</li>
                </ul>
              </div>
            </div>
          )}

        </div>

        {/* MODAL: Correct Memory */}
        {correctionModalOpen && selectedMemory && (
          <div className="fixed inset-0 z-60 bg-slate-900/50 flex items-center justify-center p-4">
            <div className="bg-white rounded-2xl shadow-xl max-w-lg w-full p-6 space-y-4 border border-slate-200">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                  <Edit3 className="w-4 h-4 text-indigo-600" />
                  <span>Correct Memory (Human Audit)</span>
                </h3>
                <button onClick={() => setCorrectionModalOpen(false)} className="text-slate-400 hover:text-slate-700">
                  <X className="w-4 h-4" />
                </button>
              </div>
              <p className="text-xs text-slate-500">
                This correction will be permanently audited and recorded under your authorized user identity.
              </p>
              <div className="space-y-3">
                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1">Corrected Content</label>
                  <textarea
                    rows={3}
                    value={correctionText}
                    onChange={(e) => setCorrectionText(e.target.value)}
                    className="w-full text-xs p-3 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                    placeholder="Enter revised operational memory content..."
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1">Reason for Correction</label>
                  <input
                    type="text"
                    value={correctionReason}
                    onChange={(e) => setCorrectionReason(e.target.value)}
                    className="w-full text-xs p-2.5 border border-slate-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                    placeholder="e.g., Customer explicitly changed delivery hours preference"
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setCorrectionModalOpen(false)}
                  className="px-4 py-2 text-xs font-medium text-slate-600 hover:bg-slate-100 rounded-lg transition"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCorrectMemory}
                  disabled={actionLoading || !correctionText.trim()}
                  className="px-4 py-2 text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition disabled:opacity-50"
                >
                  Save Audited Correction
                </button>
              </div>
            </div>
          </div>
        )}

        {/* MODAL: Invalidate Memory */}
        {invalidationModalOpen && selectedMemory && (
          <div className="fixed inset-0 z-60 bg-slate-900/50 flex items-center justify-center p-4">
            <div className="bg-white rounded-2xl shadow-xl max-w-md w-full p-6 space-y-4 border border-slate-200">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold text-rose-700 flex items-center gap-2">
                  <AlertOctagon className="w-4 h-4 text-rose-600" />
                  <span>Invalidate Memory Item</span>
                </h3>
                <button onClick={() => setInvalidationModalOpen(false)} className="text-slate-400 hover:text-slate-700">
                  <X className="w-4 h-4" />
                </button>
              </div>
              <p className="text-xs text-slate-600">
                Are you sure you want to invalidate memory #{selectedMemory.id} (&ldquo;{selectedMemory.title}&rdquo;)? It will be marked stale and will no longer be retrieved during AI context assembly.
              </p>
              <div>
                <label className="text-xs font-semibold text-slate-700 block mb-1">Reason for Invalidation</label>
                <input
                  type="text"
                  value={invalidationReason}
                  onChange={(e) => setInvalidationReason(e.target.value)}
                  className="w-full text-xs p-2.5 border border-slate-200 rounded-lg focus:ring-2 focus:ring-rose-500 focus:outline-none"
                  placeholder="e.g. Carrier discontinued sailing service on this corridor"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setInvalidationModalOpen(false)}
                  className="px-4 py-2 text-xs font-medium text-slate-600 hover:bg-slate-100 rounded-lg transition"
                >
                  Cancel
                </button>
                <button
                  onClick={handleInvalidateMemory}
                  disabled={actionLoading}
                  className="px-4 py-2 text-xs font-semibold text-white bg-rose-600 hover:bg-rose-700 rounded-lg transition disabled:opacity-50"
                >
                  Confirm Invalidation
                </button>
              </div>
            </div>
          </div>
        )}

      </div>
    </div>
  );
}
