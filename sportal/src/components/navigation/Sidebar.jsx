import React from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Building2,
  TrendingUp,
  CreditCard,
  Receipt,
  Users,
  Compass,
  BarChart2,
  Heart,
  Share2,
  FileText,
  Headphones,
  Sparkles,
  Settings,
  ShieldCheck,
  Inbox,
} from 'lucide-react';
import { ROUTES } from '../../app/config/env';

const NAV_ITEMS = [
  { name: 'Dashboard', path: ROUTES.DASHBOARD, icon: LayoutDashboard },
  { name: 'Organizations', path: ROUTES.ORGANIZATIONS, icon: Building2 },
  { name: 'Demo Requests', path: ROUTES.DEMO_REQUESTS, icon: Inbox },
  { name: 'Onboarding', path: ROUTES.ONBOARDING, icon: TrendingUp },
  { name: 'Subscriptions', path: ROUTES.SUBSCRIPTIONS, icon: CreditCard },
  { name: 'Billing & Finance', path: ROUTES.BILLING, icon: Receipt },
  { name: 'Users & Roles', path: ROUTES.USERS, icon: Users },
  { name: 'Customer 360', path: '/customer-360', icon: Compass },
  { name: 'Usage & Analytics', path: ROUTES.USAGE, icon: BarChart2 },
  { name: 'Customer Health', path: ROUTES.CUSTOMER_HEALTH, icon: Heart },
  { name: 'Integrations', path: ROUTES.INTEGRATIONS, icon: Share2 },
  { name: 'Documents & Compliance', path: ROUTES.DOCUMENTS, icon: FileText },
  { name: 'Support & Activity', path: ROUTES.SUPPORT, icon: Headphones },
  { name: 'SPortal AI', path: ROUTES.AI, icon: Sparkles },
  { name: 'Settings', path: ROUTES.SETTINGS, icon: Settings },
];

export function Sidebar() {
  return (
    <aside
      aria-label="Sidebar navigation"
      className="w-64 bg-[#0A1628] text-slate-300 flex flex-col flex-shrink-0 min-h-screen border-r border-[#17263C] select-none"
    >
      {/* Brand Header */}
      <div className="h-16 flex items-center px-5 border-b border-[#17263C] space-x-3">
        <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white shadow-md shadow-blue-600/30 flex-shrink-0">
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2">
            <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </div>
        <div className="flex flex-col min-w-0">
          <span className="text-white font-bold text-base tracking-tight leading-tight">LogisticsHQ</span>
          <span className="text-[8px] text-blue-400 font-semibold tracking-wider uppercase truncate">SMART LOGISTICS. STRONGER TOGETHER.</span>
        </div>
      </div>

      {/* Navigation List matching sportalDashboard.png */}
      <div className="flex-1 overflow-y-auto py-3 px-3 space-y-1 custom-scrollbar">
        {NAV_ITEMS.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.name}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) =>
                `group relative flex items-center px-3.5 py-2.5 rounded-lg text-[13px] font-medium transition-colors duration-150 ${
                  isActive
                    ? 'bg-[#2563EB] text-white font-semibold shadow-sm'
                    : 'text-slate-400 hover:text-slate-100 hover:bg-[#132238]'
                }`
              }
            >
              {({ isActive }) => (
                <div className="flex items-center space-x-3 min-w-0 w-full">
                  <Icon
                    className={`w-4 h-4 flex-shrink-0 transition-colors ${
                      isActive ? 'text-white' : 'text-slate-400 group-hover:text-slate-200'
                    }`}
                  />
                  <span className="truncate">{item.name}</span>
                </div>
              )}
            </NavLink>
          );
        })}
      </div>

      {/* Bottom Brand Card */}
      <div className="p-3 border-t border-[#17263C] bg-[#07101E]">
        <div className="bg-[#0D1C33] rounded-xl p-3 border border-[#1E3252] flex items-center space-x-3">
          <div className="w-8 h-8 rounded-lg bg-blue-600/30 border border-blue-500/40 flex items-center justify-center text-blue-400 flex-shrink-0">
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
            </svg>
          </div>
          <div className="flex flex-col min-w-0 leading-tight">
            <span className="text-white text-xs font-semibold">LogisticsHQ</span>
            <span className="text-[10px] text-slate-400">Smarter Logistics.</span>
            <span className="text-[10px] text-slate-400">Stronger Business.</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
