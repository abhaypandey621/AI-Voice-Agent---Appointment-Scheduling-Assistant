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
  Clock,
  User,
  AlertCircle,
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
  const [expanded, setExpanded] = React.useState(false);

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
      <div
        className="flex items-center gap-3 cursor-pointer"
        onClick={() => setExpanded(!expanded)}
      >
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

        {(toolCall.result || toolCall.error) && (
          <button className="text-xs text-agent-primary hover:underline">
            {expanded ? 'Hide' : 'Details'}
          </button>
        )}
      </div>

      {/* Arguments display */}
      <>{renderArguments(toolCall.arguments)}</>

      {/* Result display - shown when expanded or for important results */}
      {(expanded || toolCall.status === 'completed') && toolCall.result ? (
        <ToolResultDisplay
          toolName={toolCall.name}
          result={toolCall.result}
          expanded={expanded}
        />
      ) : null}

      {/* Error display */}
      {toolCall.error && (
        <div className="mt-3 p-2 bg-red-100 rounded-lg border border-red-200">
          <div className="flex items-center gap-2 text-red-700 text-xs">
            <AlertCircle className="w-4 h-4" />
            <span>{toolCall.error}</span>
          </div>
        </div>
      )}
    </div>
  );
};

interface ToolResultDisplayProps {
  toolName: string;
  result: unknown;
  expanded: boolean;
}

const ToolResultDisplay: React.FC<ToolResultDisplayProps> = ({ toolName, result, expanded }) => {
  const resultData = result as Record<string, unknown>;

  if (!resultData || typeof resultData !== 'object') {
    return null;
  }

  // Handle different tool results
  switch (toolName) {
    case 'identify_user':
      return (
        <div className="mt-3 p-3 bg-blue-50 rounded-lg border border-blue-100">
          <div className="flex items-center gap-2 mb-2">
            <User className="w-4 h-4 text-blue-600" />
            <span className="font-medium text-blue-800 text-sm">User Identified</span>
          </div>
          <div className="space-y-1 text-xs">
            {resultData.name ? (
              <div className="flex gap-2">
                <span className="text-gray-500">Name:</span>
                <span className="text-gray-700 font-medium">{String(resultData.name)}</span>
              </div>
            ) : null}
            {resultData.phone_number ? (
              <div className="flex gap-2">
                <span className="text-gray-500">Phone:</span>
                <span className="text-gray-700">{String(resultData.phone_number)}</span>
              </div>
            ) : null}
            {resultData.email ? (
              <div className="flex gap-2">
                <span className="text-gray-500">Email:</span>
                <span className="text-gray-700">{String(resultData.email)}</span>
              </div>
            ) : null}
            {resultData.is_new_user !== undefined ? (
              <div className="mt-1 text-xs text-blue-600">
                {resultData.is_new_user ? '✨ New user registered' : '👋 Returning user'}
              </div>
            ) : null}
          </div>
        </div>
      );

    case 'fetch_slots':
      const slots = resultData.slots as Array<Record<string, unknown>> || [];
      const availableSlots = slots.filter(s => s.available);
      return (
        <div className="mt-3 p-3 bg-purple-50 rounded-lg border border-purple-100">
          <div className="flex items-center gap-2 mb-2">
            <Clock className="w-4 h-4 text-purple-600" />
            <span className="font-medium text-purple-800 text-sm">
              Available Slots ({availableSlots.length})
            </span>
          </div>
          {expanded && availableSlots.length > 0 && (
            <div className="grid grid-cols-3 gap-1 mt-2">
              {availableSlots.slice(0, 9).map((slot, idx) => (
                <div
                  key={idx}
                  className="text-xs p-1.5 bg-white rounded text-center text-purple-700 border border-purple-200"
                >
                  {String(slot.time)}
                </div>
              ))}
              {availableSlots.length > 9 && (
                <div className="text-xs p-1.5 text-purple-500 text-center">
                  +{availableSlots.length - 9} more
                </div>
              )}
            </div>
          )}
          {!expanded && (
            <div className="text-xs text-purple-600 mt-1">
              Click to see available times
            </div>
          )}
        </div>
      );

    case 'book_appointment':
      return (
        <div className="mt-3 p-3 bg-green-50 rounded-lg border border-green-100">
          <div className="flex items-center gap-2 mb-2">
            <CalendarPlus className="w-4 h-4 text-green-600" />
            <span className="font-medium text-green-800 text-sm">
              {resultData.success ? 'Appointment Booked!' : 'Booking Failed'}
            </span>
          </div>
          {resultData.success ? (
            <div className="space-y-1 text-xs">
              <div className="flex gap-2">
                <span className="text-gray-500">Date & Time:</span>
                <span className="text-gray-700 font-medium">{String(resultData.date_time)}</span>
              </div>
              <div className="flex gap-2">
                <span className="text-gray-500">Duration:</span>
                <span className="text-gray-700">{String(resultData.duration)} minutes</span>
              </div>
              {resultData.purpose ? (
                <div className="flex gap-2">
                  <span className="text-gray-500">Purpose:</span>
                  <span className="text-gray-700">{String(resultData.purpose)}</span>
                </div>
              ) : null}
            </div>
          ) : (
            <div className="text-xs text-red-600">{String(resultData.error || 'Unknown error')}</div>
          )}
        </div>
      );

    case 'retrieve_appointments':
      const appointments = resultData.appointments as Array<Record<string, unknown>> || [];
      const count = resultData.count as number || appointments.length;
      return (
        <div className="mt-3 p-3 bg-indigo-50 rounded-lg border border-indigo-100">
          <div className="flex items-center gap-2 mb-2">
            <List className="w-4 h-4 text-indigo-600" />
            <span className="font-medium text-indigo-800 text-sm">
              {count} Appointment{count !== 1 ? 's' : ''} Found
            </span>
          </div>
          {appointments.length > 0 ? (
            <div className="space-y-2 mt-2">
              {appointments.map((apt, idx) => (
                <div
                  key={idx}
                  className="p-2 bg-white rounded border border-indigo-200 text-xs"
                >
                  <div className="font-medium text-indigo-800">{String(apt.date_time)}</div>
                  <div className="text-gray-600 mt-0.5">
                    <span>{String(apt.duration)} min</span>
                    {apt.purpose ? <span> • {String(apt.purpose)}</span> : null}
                  </div>
                  <div className={clsx(
                    'mt-1 text-xs inline-block px-1.5 py-0.5 rounded',
                    apt.status === 'booked' ? 'bg-green-100 text-green-700' :
                    apt.status === 'cancelled' ? 'bg-red-100 text-red-700' :
                    'bg-gray-100 text-gray-700'
                  )}>
                    {String(apt.status)}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-xs text-indigo-600">No appointments found</div>
          )}
        </div>
      );

    case 'cancel_appointment':
      return (
        <div className="mt-3 p-3 bg-red-50 rounded-lg border border-red-100">
          <div className="flex items-center gap-2 mb-2">
            <CalendarX className="w-4 h-4 text-red-600" />
            <span className="font-medium text-red-800 text-sm">
              {resultData.success ? 'Appointment Cancelled' : 'Cancellation Failed'}
            </span>
          </div>
          {resultData.success ? (
            <div className="text-xs text-gray-600">
              Appointment on {String(resultData.date_time)} has been cancelled
            </div>
          ) : (
            <div className="text-xs text-red-600">{String(resultData.error || 'Unknown error')}</div>
          )}
        </div>
      );

    case 'modify_appointment':
      const changes = resultData.changes as string[] || [];
      return (
        <div className="mt-3 p-3 bg-orange-50 rounded-lg border border-orange-100">
          <div className="flex items-center gap-2 mb-2">
            <CalendarClock className="w-4 h-4 text-orange-600" />
            <span className="font-medium text-orange-800 text-sm">
              {resultData.success ? 'Appointment Modified' : 'Modification Failed'}
            </span>
          </div>
          {resultData.success ? (
            <div className="text-xs text-gray-600">
              {changes.length > 0 ? (
                <ul className="list-disc list-inside">
                  {changes.map((change, idx) => (
                    <li key={idx}>{change}</li>
                  ))}
                </ul>
              ) : (
                <span>Changes applied successfully</span>
              )}
            </div>
          ) : (
            <div className="text-xs text-red-600">{String(resultData.error || 'Unknown error')}</div>
          )}
        </div>
      );

    case 'end_conversation':
      return (
        <div className="mt-3 p-3 bg-gray-50 rounded-lg border border-gray-200">
          <div className="flex items-center gap-2">
            <LogOut className="w-4 h-4 text-gray-600" />
            <span className="font-medium text-gray-800 text-sm">Call Ending</span>
          </div>
          <div className="text-xs text-gray-500 mt-1">
            Generating summary...
          </div>
        </div>
      );

    default:
      // Generic result display
      if (expanded && resultData.message) {
        return (
          <div className="mt-3 p-2 bg-gray-50 rounded-lg text-xs text-gray-600">
            {String(resultData.message)}
          </div>
        );
      }
      return null;
  }
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

function renderArguments(args: Record<string, unknown>): React.ReactNode {
  const keys = Object.keys(args);
  if (keys.length === 0) {
    return null;
  }
  return (
    <div className="mt-3 pl-13">
      <div className="bg-white/50 rounded-lg p-2 text-xs">
        {keys.map((key) => (
          <div key={key} className="flex gap-2">
            <span className="text-gray-500 font-medium">{formatKey(key)}:</span>
            <span className="text-gray-700 truncate">
              {formatValue(args[key])}
            </span>
          </div>
        ))}
      </div>
    </div>
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
