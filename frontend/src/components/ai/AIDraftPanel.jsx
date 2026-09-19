import React, { useState } from 'react';
import { FileEdit, ShieldAlert, CheckCircle, Send, X } from 'lucide-react';
import './SharedAI.css';

export default function AIDraftPanel({
  title = 'AI-Generated Draft',
  type = 'Draft',
  initialContent = '',
  recipient,
  onSave,
  onSendForApproval,
  onDiscard,
  isApprovalRequired = true,
  disabled = false
}) {
  const [content, setContent] = useState(initialContent);
  const [isEditing, setIsEditing] = useState(false);

  return (
    <div className="sai-draft-panel">
      <div className="sai-draft-header">
        <div className="flex items-center gap-2">
          <span className="sai-draft-pill">
            <FileEdit className="w-3 h-3 text-amber-600" />
            {type.toUpperCase()} &mdash; NOT SENT
          </span>
          <span className="text-xs font-semibold text-slate-800">{title}</span>
        </div>
        {recipient && (
          <span className="text-xs text-slate-500 font-mono">To: {recipient}</span>
        )}
      </div>

      <div className="sai-draft-body">
        {isEditing ? (
          <textarea
            className="sai-draft-textarea"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={5}
            disabled={disabled}
            aria-label="Edit draft message"
          />
        ) : (
          <div className="sai-draft-preview whitespace-pre-wrap">
            {content}
          </div>
        )}
      </div>

      <div className="sai-draft-footer">
        <div className="flex items-center gap-1.5 text-[11px] text-amber-700">
          <ShieldAlert className="w-3.5 h-3.5 text-amber-500 shrink-0" />
          <span>Requires review. Never sent automatically without explicit authorization.</span>
        </div>
        <div className="flex items-center gap-2">
          {onDiscard && (
            <button
              type="button"
              className="sai-btn sai-btn--secondary"
              onClick={onDiscard}
              disabled={disabled}
            >
              <X className="w-3 h-3" /> Discard
            </button>
          )}
          <button
            type="button"
            className="sai-btn sai-btn--secondary"
            onClick={() => {
              if (isEditing && onSave) onSave(content);
              setIsEditing(!isEditing);
            }}
            disabled={disabled}
          >
            {isEditing ? <CheckCircle className="w-3 h-3 text-emerald-600" /> : <FileEdit className="w-3 h-3" />}
            {isEditing ? 'Save Edits' : 'Edit Draft'}
          </button>
          {onSendForApproval && (
            <button
              type="button"
              className="sai-btn sai-btn--primary"
              onClick={() => onSendForApproval(content)}
              disabled={disabled}
            >
              <Send className="w-3 h-3" />
              {isApprovalRequired ? 'Submit for Approval' : 'Send'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
