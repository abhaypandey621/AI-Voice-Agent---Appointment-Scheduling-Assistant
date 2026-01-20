import React from 'react';
import { clsx } from 'clsx';
import {
  Phone,
  Calendar,
  CalendarPlus,
  CalendarX,
  CalendarClock,
  LogOut,
  List,
  Loader2,
  CheckCircle2,
  XCircle,
  Wrench,
} from 'lucide-react';
import type { ToolCallPayload } from '../../types';

interface ToolDisplayProps {
  toolCalls: ToolCallPayload[];
  activeToolCall: ToolCallPayload | null;
  className?: string;
}

export const ToolDisplay: React.FC<ToolDisplayProps> = ({
  toolCalls,
  activeToolCall,
  className,
}) => {
  if (toolCalls.length === 0) {
    return null;
  }

  return (
    <div className={clsx('bg-white rounded-2xl shadow-lg p-4', className)}>
      <div className="flex items-center gap-2 mb-4">
        <Wrench className="w-5 h-5 text-agent-primary" />
        <h3 className="font-semibold text-gray-800">Tool Activity</h3>
      </div>

      <div className="space-y-3">
        {toolCalls.map((toolCall) => (
          <ToolCallCard
            key={toolCall.id}
            toolCall={toolCall}
            isActive={activeToolCall?.id === toolCall.id}
          />
        ))}
      </div>
    </div>
  );
};

interface ToolCallCardProps {
  toolCall: ToolCallPayload;
  isActive: boolean;
}

const ToolCallCard: React.FC<ToolCallCardProps> = ({ toolCall, isActive }) => {
  const toolInfo = getToolInfo(toolCall.name);
  const Icon = toolInfo.icon;

  const getStatusIcon = () => {
    switch (toolCall.status) {
      case 'executing':
        return <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />;
      case 'completed':
        return <CheckCircle2 className="w-4 h-4 text-green-500" />;
      case 'failed':
        return <XCircle className="w-4 h-4 text-red-500" />;
      default:
        return <div className="w-4 h-4 rounded-full border-2 border-gray-300" />;
    }
  };

  return (
    <div
      className={clsx(
        'p-3 rounded-xl border transition-all duration-300',
        isActive
          ? 'border-agent-primary bg-agent-primary/5 shadow-md'
          : toolCall.status === 'completed'
          ? 'border-green-200 bg-green-50'
          : toolCall.status === 'failed'
          ? 'border-red-200 bg-red-50'
          : 'border-gray-200 bg-gray-50'
      )}
    >
      <div className="flex items-center gap-3">
        <div
          className={clsx(
            'w-10 h-10 rounded-xl flex items-center justify-center',
            toolInfo.bgColor
          )}
        >
          <Icon className={clsx('w-5 h-5', toolInfo.iconColor)} />
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h4 className="font-medium text-gray-800 text-sm truncate">
              {toolInfo.label}
            </h4>
            {getStatusIcon()}
          </div>
          <p className="text-xs text-gray-500 mt-0.5 truncate">
            {toolInfo.description}
          </p>
        </div>
      </div>

      {/* Arguments display */}
      {Object.keys(toolCall.arguments).length > 0 && (
        <div className="mt-3 pl-13">
          <div className="bg-white/50 rounded-lg p-2 text-xs">
            {Object.entries(toolCall.arguments).map(([key, value]) => (
              <div key={key} className="flex gap-2">
                <span className="text-gray-500 font-medium">{formatKey(key)}:</span>
                <span className="text-gray-700 truncate">
                  {formatValue(value)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

interface ToolInfo {
  label: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  bgColor: string;
  iconColor: string;
}

function getToolInfo(toolName: string): ToolInfo {
  const tools: Record<string, ToolInfo> = {
    identify_user: {
      label: 'Identify User',
      description: 'Identifying user by phone number',
      icon: Phone,
      bgColor: 'bg-blue-100',
      iconColor: 'text-blue-600',
    },
    fetch_slots: {
      label: 'Fetch Slots',
      description: 'Checking available time slots',
      icon: Calendar,
      bgColor: 'bg-purple-100',
      iconColor: 'text-purple-600',
    },
    book_appointment: {
      label: 'Book Appointment',
      description: 'Creating new appointment',
      icon: CalendarPlus,
      bgColor: 'bg-green-100',
      iconColor: 'text-green-600',
    },
    retrieve_appointments: {
      label: 'Get Appointments',
      description: 'Retrieving user appointments',
      icon: List,
      bgColor: 'bg-indigo-100',
      iconColor: 'text-indigo-600',
    },
    cancel_appointment: {
      label: 'Cancel Appointment',
      description: 'Cancelling appointment',
      icon: CalendarX,
      bgColor: 'bg-red-100',
      iconColor: 'text-red-600',
    },
    modify_appointment: {
      label: 'Modify Appointment',
      description: 'Updating appointment details',
      icon: CalendarClock,
      bgColor: 'bg-orange-100',
      iconColor: 'text-orange-600',
    },
    end_conversation: {
      label: 'End Call',
      description: 'Ending the conversation',
      icon: LogOut,
      bgColor: 'bg-gray-100',
      iconColor: 'text-gray-600',
    },
  };

  return (
    tools[toolName] || {
      label: toolName,
      description: 'Executing action',
      icon: Wrench,
      bgColor: 'bg-gray-100',
      iconColor: 'text-gray-600',
    }
  );
}

function formatKey(key: string): string {
  return key
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (l) => l.toUpperCase());
}

function formatValue(value: unknown): string {
  if (typeof value === 'string') return value;
  if (typeof value === 'number') return value.toString();
  if (typeof value === 'boolean') return value ? 'Yes' : 'No';
  if (value === null || value === undefined) return '-';
  return JSON.stringify(value);
}

export default ToolDisplay;
