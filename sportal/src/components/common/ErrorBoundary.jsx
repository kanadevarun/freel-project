import React from 'react';
import { AlertTriangle, RefreshCw, Home, ChevronDown, ChevronUp } from 'lucide-react';

export class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      showDetails: false,
    };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    this.setState({ errorInfo });
    // Log to console for observability / telemetry capture
    console.error('[SPortal ErrorBoundary Caught]', error, errorInfo);
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      showDetails: false,
    });
    if (this.props.onReset) {
      this.props.onReset();
    }
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="p-8 max-w-4xl mx-auto my-8">
          <div className="bg-white border border-slate-200 rounded-xl shadow-sm p-6 sm:p-8">
            <div className="flex items-start gap-4">
              <div className="w-12 h-12 rounded-xl bg-amber-50 border border-amber-200 flex items-center justify-center shrink-0 text-amber-600">
                <AlertTriangle className="w-6 h-6" />
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold px-2 py-0.5 rounded bg-amber-100 text-amber-800 uppercase tracking-wider">
                    Component Isolation
                  </span>
                </div>
                <h2 className="text-xl font-bold text-slate-900 mt-2">
                  Unable to load this module view
                </h2>
                <p className="text-sm text-slate-600 mt-1">
                  An unexpected rendering condition occurred. Independent modules in SPortal remain fully operational. You can attempt to refresh this section or return to the main dashboard.
                </p>

                {this.state.error && (
                  <div className="mt-4 border border-slate-200 rounded-lg overflow-hidden bg-slate-50">
                    <button
                      onClick={() => this.setState((prev) => ({ showDetails: !prev.showDetails }))}
                      className="w-full px-4 py-2.5 flex items-center justify-between text-xs font-medium text-slate-700 hover:bg-slate-100 transition"
                    >
                      <span>Technical Details ({this.state.error.name || 'Error'})</span>
                      {this.state.showDetails ? (
                        <ChevronUp className="w-4 h-4 text-slate-500" />
                      ) : (
                        <ChevronDown className="w-4 h-4 text-slate-500" />
                      )}
                    </button>
                    {this.state.showDetails && (
                      <div className="p-4 border-t border-slate-200 text-xs font-mono text-slate-800 bg-white max-h-48 overflow-y-auto space-y-2">
                        <div className="font-semibold text-rose-700">
                          {this.state.error.message || 'Unknown error'}
                        </div>
                        {this.state.errorInfo?.componentStack && (
                          <pre className="text-[11px] text-slate-500 whitespace-pre-wrap">
                            {this.state.errorInfo.componentStack}
                          </pre>
                        )}
                      </div>
                    )}
                  </div>
                )}

                <div className="mt-6 flex flex-wrap items-center gap-3">
                  <button
                    onClick={this.handleReset}
                    className="inline-flex items-center gap-2 px-4 py-2 bg-[#0B192C] hover:bg-[#1E3E62] text-white text-xs font-medium rounded-lg shadow-sm transition"
                  >
                    <RefreshCw className="w-3.5 h-3.5" />
                    Retry Component
                  </button>
                  <a
                    href="/"
                    className="inline-flex items-center gap-2 px-4 py-2 bg-white border border-slate-300 hover:bg-slate-50 text-slate-700 text-xs font-medium rounded-lg shadow-sm transition"
                  >
                    <Home className="w-3.5 h-3.5" />
                    Return to Dashboard
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
