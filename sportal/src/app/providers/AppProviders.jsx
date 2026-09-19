import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from '../../contexts/AuthContext';
import { ErrorBoundary } from '../../components/common/ErrorBoundary';

export function AppProviders({ children }) {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <AuthProvider>
          {children}
          <Toaster
            position="top-right"
            toastOptions={{
              duration: 4000,
              style: {
                background: '#0B192C',
                color: '#FFFFFF',
                fontSize: '13px',
                borderRadius: '8px',
              },
            }}
          />
        </AuthProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}
