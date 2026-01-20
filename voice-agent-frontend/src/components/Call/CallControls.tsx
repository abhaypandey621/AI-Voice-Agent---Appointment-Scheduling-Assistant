import React from 'react';
import { clsx } from 'clsx';
import { Mic, MicOff, PhoneOff, Phone } from 'lucide-react';
import type { CallState } from '../../types';

interface CallControlsProps {
  callState: CallState;
  isRecording: boolean;
  isConnected: boolean;
  audioLevel: number;
  onConnect: () => void;
  onDisconnect: () => void;
  onStartRecording: () => void;
  onStopRecording: () => void;
  onEndCall: () => void;
  className?: string;
}

export const CallControls: React.FC<CallControlsProps> = ({
  callState,
  isRecording,
  isConnected,
  audioLevel,
  onConnect,
  onDisconnect: _onDisconnect,
  onStartRecording,
  onStopRecording,
  onEndCall,
  className,
}) => {
  const showEndCall = callState !== 'idle' && callState !== 'ended';
  const showConnect = callState === 'idle';
  const showMicControls = isConnected && callState !== 'ending' && callState !== 'ended';

  return (
    <div className={clsx('flex items-center justify-center gap-4', className)}>
      {/* Connect button */}
      {showConnect && (
        <button
          onClick={onConnect}
          className={clsx(
            'flex items-center gap-3 px-8 py-4 rounded-2xl',
            'bg-gradient-to-r from-agent-primary to-agent-secondary',
            'text-white font-semibold text-lg',
            'hover:shadow-lg hover:scale-105 transition-all duration-300',
            'focus:outline-none focus:ring-4 focus:ring-agent-primary/30'
          )}
        >
          <Phone className="w-6 h-6" />
          Start Call
        </button>
      )}

      {/* Mic control */}
      {showMicControls && (
        <div className="relative">
          {/* Audio level indicator */}
          <div
            className={clsx(
              'absolute inset-0 rounded-full transition-all duration-100',
              isRecording && 'animate-pulse'
            )}
            style={{
              transform: `scale(${1 + audioLevel * 0.3})`,
              backgroundColor: isRecording
                ? `rgba(239, 68, 68, ${audioLevel * 0.3})`
                : 'transparent',
            }}
          />

          <button
            onClick={isRecording ? onStopRecording : onStartRecording}
            className={clsx(
              'relative z-10 w-16 h-16 rounded-full flex items-center justify-center',
              'transition-all duration-300',
              'focus:outline-none focus:ring-4',
              isRecording
                ? 'bg-red-500 text-white hover:bg-red-600 focus:ring-red-500/30'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200 focus:ring-gray-500/30'
            )}
          >
            {isRecording ? (
              <MicOff className="w-7 h-7" />
            ) : (
              <Mic className="w-7 h-7" />
            )}
          </button>
        </div>
      )}

      {/* End call button */}
      {showEndCall && (
        <button
          onClick={onEndCall}
          disabled={callState === 'ending'}
          className={clsx(
            'flex items-center gap-2 px-6 py-3 rounded-xl',
            'bg-red-500 text-white font-medium',
            'hover:bg-red-600 transition-colors duration-200',
            'focus:outline-none focus:ring-4 focus:ring-red-500/30',
            'disabled:opacity-50 disabled:cursor-not-allowed'
          )}
        >
          <PhoneOff className="w-5 h-5" />
          End Call
        </button>
      )}
    </div>
  );
};

interface CallStatusProps {
  callState: CallState;
  error?: string | null;
  className?: string;
}

export const CallStatus: React.FC<CallStatusProps> = ({
  callState,
  error,
  className,
}) => {
  const getStatusInfo = () => {
    switch (callState) {
      case 'connecting':
        return { text: 'Connecting...', color: 'text-yellow-600', pulse: true };
      case 'connected':
        return { text: 'Connected', color: 'text-green-600', pulse: false };
      case 'speaking':
        return { text: 'Agent Speaking', color: 'text-blue-600', pulse: true };
      case 'listening':
        return { text: 'Listening...', color: 'text-green-600', pulse: true };
      case 'processing':
        return { text: 'Processing...', color: 'text-purple-600', pulse: true };
      case 'ending':
        return { text: 'Ending call...', color: 'text-orange-600', pulse: true };
      case 'ended':
        return { text: 'Call Ended', color: 'text-gray-600', pulse: false };
      default:
        return { text: 'Ready to Connect', color: 'text-gray-500', pulse: false };
    }
  };

  const status = getStatusInfo();

  return (
    <div className={clsx('text-center', className)}>
      <div className="flex items-center justify-center gap-2">
        {status.pulse && (
          <div
            className={clsx(
              'w-2 h-2 rounded-full animate-pulse',
              callState === 'listening'
                ? 'bg-green-500'
                : callState === 'speaking'
                  ? 'bg-blue-500'
                  : 'bg-yellow-500'
            )}
          />
        )}
        <span className={clsx('font-medium', status.color)}>{status.text}</span>
      </div>

      {error && (
        <p className="text-red-500 text-sm mt-2 bg-red-50 px-3 py-1 rounded-lg">
          {error}
        </p>
      )}
    </div>
  );
};

export default CallControls;
