import { Link } from 'react-router-dom';
import { AlertCircle, CheckCircle2 } from 'lucide-react';
import './AuthMessage.css';

export default function AuthMessage({ type = 'error', message, actionLink, actionText }) {
  if (!message) return null;

  const isSuccess = type === 'success';

  return (
    <div className={`auth-message ${isSuccess ? 'auth-message-success' : 'auth-message-error'}`}>
      <div className="auth-message-icon">
        {isSuccess ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
      </div>
      <div className="auth-message-content">
        <span>{message}</span>
        {actionLink && actionText && (
          <Link to={actionLink} className="auth-message-action">
            {actionText}
          </Link>
        )}
      </div>
    </div>
  );
}
