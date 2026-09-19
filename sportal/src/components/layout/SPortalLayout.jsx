import React, { Suspense } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from '../navigation/Sidebar';
import { Header } from '../navigation/Header';
import { ErrorBoundary } from '../common/ErrorBoundary';
import { LoadingSpinner } from '../common/LoadingSpinner';

export function SPortalLayout() {
  return (
    <div className="flex h-screen bg-slate-50 overflow-hidden">
      {/* Dark Navy Sidebar */}
      <Sidebar />

      {/* Main Content Viewport */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Top Header */}
        <Header />

        {/* Scrollable Main Area */}
        <main className="flex-1 overflow-y-auto overflow-x-hidden">
          <div className="min-h-[calc(100vh-8rem)]">
            <ErrorBoundary>
              <Suspense
                fallback={
                  <div className="flex items-center justify-center p-16">
                    <LoadingSpinner size="lg" message="Loading administrative view..." />
                  </div>
                }
              >
                <Outlet />
              </Suspense>
            </ErrorBoundary>
          </div>

          {/* SPortal Footer */}
          <footer className="h-14 border-t border-slate-200 bg-white px-8 flex items-center justify-between text-xs text-slate-500">
            <div>
              <span className="font-semibold text-slate-700">LogisticsHQ</span> © {new Date().getFullYear()}. All rights reserved.
            </div>
            <div className="flex items-center space-x-4">
              <a href="#privacy" className="hover:text-blue-600 transition">Privacy</a>
              <span>|</span>
              <a href="#terms" className="hover:text-blue-600 transition">Terms</a>
              <span>|</span>
              <a href="#support" className="hover:text-blue-600 transition">Support</a>
            </div>
          </footer>
        </main>
      </div>
    </div>
  );
}
