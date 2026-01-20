import React from 'react';
import { clsx } from 'clsx';
import { format } from 'date-fns';
import {
  FileText,
  Calendar,
  Heart,
  Tag,
  Clock,
  DollarSign,
  X,
} from 'lucide-react';
import type { CallSummary as CallSummaryType, CostBreakdown } from '../../types';

interface CallSummaryProps {
  summary: CallSummaryType;
  costBreakdown?: CostBreakdown | null;
  onClose?: () => void;
  className?: string;
}

export const CallSummary: React.FC<CallSummaryProps> = ({
  summary,
  costBreakdown,
  onClose,
  className,
}) => {
  return (
    <div
      className={clsx(
        'bg-white rounded-2xl shadow-xl overflow-hidden animate-fade-in',
        className
      )}
    >
      {/* Header */}
      <div className="bg-gradient-to-r from-agent-primary to-agent-secondary p-6 text-white">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-xl bg-white/20 flex items-center justify-center">
              <FileText className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-xl font-bold">Call Summary</h2>
              <p className="text-white/80 text-sm">
                {format(new Date(summary.created_at), 'MMMM d, yyyy • h:mm a')}
              </p>
            </div>
          </div>
          {onClose && (
            <button
              onClick={onClose}
              className="p-2 rounded-lg hover:bg-white/20 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          )}
        </div>

        {/* Duration */}
        <div className="flex items-center gap-2 mt-4 text-white/90">
          <Clock className="w-4 h-4" />
          <span className="text-sm">
            Call duration: {formatDuration(summary.duration)}
          </span>
        </div>
      </div>

      {/* Content */}
      <div className="p-6 space-y-6">
        {/* Summary text */}
        <section>
          <h3 className="font-semibold text-gray-800 mb-2 flex items-center gap-2">
            <FileText className="w-4 h-4 text-agent-primary" />
            Summary
          </h3>
          <p className="text-gray-600 leading-relaxed">{summary.summary}</p>
        </section>

        {/* Appointments booked */}
        {summary.appointments_booked && summary.appointments_booked.length > 0 && (
          <section>
            <h3 className="font-semibold text-gray-800 mb-3 flex items-center gap-2">
              <Calendar className="w-4 h-4 text-green-500" />
              Appointments Booked
            </h3>
            <div className="space-y-2">
              {summary.appointments_booked.map((apt, index) => (
                <div
                  key={apt.id || index}
                  className="bg-green-50 border border-green-100 rounded-xl p-3"
                >
                  <div className="font-medium text-green-800">
                    {format(new Date(apt.date_time), 'EEEE, MMMM d, yyyy')}
                  </div>
                  <div className="text-sm text-green-600">
                    {format(new Date(apt.date_time), 'h:mm a')} • {apt.duration} minutes
                  </div>
                  {apt.purpose && (
                    <div className="text-sm text-green-700 mt-1">{apt.purpose}</div>
                  )}
                </div>
              ))}
            </div>
          </section>
        )}

        {/* User preferences */}
        {summary.user_preferences && summary.user_preferences.length > 0 && (
          <section>
            <h3 className="font-semibold text-gray-800 mb-3 flex items-center gap-2">
              <Heart className="w-4 h-4 text-pink-500" />
              User Preferences
            </h3>
            <div className="flex flex-wrap gap-2">
              {summary.user_preferences.map((pref, index) => (
                <span
                  key={index}
                  className="px-3 py-1 bg-pink-50 text-pink-700 rounded-full text-sm"
                >
                  {pref}
                </span>
              ))}
            </div>
          </section>
        )}

        {/* Key topics */}
        {summary.key_topics && summary.key_topics.length > 0 && (
          <section>
            <h3 className="font-semibold text-gray-800 mb-3 flex items-center gap-2">
              <Tag className="w-4 h-4 text-blue-500" />
              Key Topics
            </h3>
            <div className="flex flex-wrap gap-2">
              {summary.key_topics.map((topic, index) => (
                <span
                  key={index}
                  className="px-3 py-1 bg-blue-50 text-blue-700 rounded-full text-sm"
                >
                  {topic}
                </span>
              ))}
            </div>
          </section>
        )}

        {/* Cost breakdown */}
        {costBreakdown && (
          <section className="border-t border-gray-100 pt-6">
            <h3 className="font-semibold text-gray-800 mb-3 flex items-center gap-2">
              <DollarSign className="w-4 h-4 text-yellow-500" />
              Cost Breakdown
            </h3>
            <CostBreakdownDisplay cost={costBreakdown} />
          </section>
        )}
      </div>
    </div>
  );
};

interface CostBreakdownDisplayProps {
  cost: CostBreakdown;
}

const CostBreakdownDisplay: React.FC<CostBreakdownDisplayProps> = ({ cost }) => {
  const items = [
    {
      label: 'Speech to Text (Deepgram)',
      value: cost.stt_cost,
      detail: `${cost.stt_minutes.toFixed(2)} minutes`,
    },
    {
      label: 'Text to Speech (Cartesia)',
      value: cost.tts_cost,
      detail: `${cost.tts_characters.toLocaleString()} characters`,
    },
    {
      label: 'LLM Processing',
      value: cost.llm_cost,
      detail: `${cost.llm_tokens.toLocaleString()} tokens`,
    },
  ];

  if (cost.avatar_cost > 0) {
    items.push({
      label: 'Avatar Streaming',
      value: cost.avatar_cost,
      detail: '',
    });
  }

  return (
    <div className="bg-gray-50 rounded-xl p-4">
      <div className="space-y-3">
        {items.map((item, index) => (
          <div key={index} className="flex items-center justify-between">
            <div>
              <span className="text-sm text-gray-700">{item.label}</span>
              {item.detail && (
                <span className="text-xs text-gray-400 ml-2">({item.detail})</span>
              )}
            </div>
            <span className="font-mono text-sm text-gray-800">
              ${item.value.toFixed(4)}
            </span>
          </div>
        ))}
      </div>

      <div className="border-t border-gray-200 mt-3 pt-3 flex items-center justify-between">
        <span className="font-semibold text-gray-800">Total Cost</span>
        <span className="font-mono font-bold text-lg text-agent-primary">
          ${cost.total_cost.toFixed(4)}
        </span>
      </div>
    </div>
  );
};

function formatDuration(seconds: number): string {
  if (seconds < 60) {
    return `${seconds} seconds`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  if (remainingSeconds === 0) {
    return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
  }
  return `${minutes}m ${remainingSeconds}s`;
}

export default CallSummary;
