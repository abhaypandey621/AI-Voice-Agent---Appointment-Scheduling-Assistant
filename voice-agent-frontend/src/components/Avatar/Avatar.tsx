import React from 'react';
import { clsx } from 'clsx';
import type { AvatarState } from '../../types';
import { User, Volume2, Mic, Brain } from 'lucide-react';

interface AvatarProps {
  state: AvatarState;
  imageUrl?: string;
  iframeUrl?: string;
  className?: string;
}

export const Avatar: React.FC<AvatarProps> = ({
  state,
  imageUrl,
  iframeUrl,
  className,
}) => {
  // Render iframe if Tavus conversation URL is available
  if (iframeUrl) {
    return (
      <div className={clsx('relative rounded-2xl overflow-hidden bg-gray-900', className)}>
        <iframe
          src={iframeUrl}
          className="w-full h-full border-0"
          allow="camera; microphone; autoplay"
          title="AI Avatar"
        />
        <AvatarStateIndicator state={state} />
      </div>
    );
  }

  // Fallback to animated avatar
  return (
    <div
      className={clsx(
        'relative flex items-center justify-center rounded-2xl overflow-hidden',
        'bg-gradient-to-br from-agent-primary to-agent-secondary',
        className
      )}
    >
      <div className="absolute inset-0 bg-black/20" />

      {/* Animated background */}
      <div
        className={clsx(
          'absolute inset-0 opacity-30',
          state === 'speaking' && 'animate-pulse-slow',
          state === 'thinking' && 'animate-pulse'
        )}
      >
        <div className="absolute inset-0 bg-gradient-to-t from-agent-accent/50 to-transparent" />
      </div>

      {/* Avatar circle */}
      <div
        className={clsx(
          'relative z-10 w-32 h-32 md:w-48 md:h-48 rounded-full flex items-center justify-center',
          'bg-white/10 backdrop-blur-sm border-2 border-white/20',
          'transition-transform duration-300',
          state === 'speaking' && 'scale-105',
          state === 'thinking' && 'animate-pulse'
        )}
      >
        {imageUrl ? (
          <img
            src={imageUrl}
            alt="AI Assistant"
            className="w-full h-full rounded-full object-cover"
          />
        ) : (
          <User className="w-16 h-16 md:w-24 md:h-24 text-white/80" />
        )}
      </div>

      {/* Sound waves for speaking state */}
      {state === 'speaking' && (
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-end gap-1">
          {[...Array(5)].map((_, i) => (
            <div
              key={i}
              className="w-1 bg-white/60 rounded-full animate-wave"
              style={{
                height: `${Math.random() * 20 + 10}px`,
                animationDelay: `${i * 0.1}s`,
              }}
            />
          ))}
        </div>
      )}

      <AvatarStateIndicator state={state} />
    </div>
  );
};

interface AvatarStateIndicatorProps {
  state: AvatarState;
}

const AvatarStateIndicator: React.FC<AvatarStateIndicatorProps> = ({ state }) => {
  const getStateInfo = () => {
    switch (state) {
      case 'speaking':
        return {
          icon: Volume2,
          text: 'Speaking',
          color: 'bg-green-500',
        };
      case 'listening':
        return {
          icon: Mic,
          text: 'Listening',
          color: 'bg-blue-500',
        };
      case 'thinking':
        return {
          icon: Brain,
          text: 'Thinking',
          color: 'bg-yellow-500',
        };
      default:
        return null;
    }
  };

  const stateInfo = getStateInfo();
  if (!stateInfo) return null;

  const Icon = stateInfo.icon;

  return (
    <div className="absolute bottom-4 right-4 flex items-center gap-2 px-3 py-1.5 rounded-full bg-black/50 backdrop-blur-sm">
      <div className={clsx('w-2 h-2 rounded-full animate-pulse', stateInfo.color)} />
      <Icon className="w-4 h-4 text-white" />
      <span className="text-sm text-white font-medium">{stateInfo.text}</span>
    </div>
  );
};

export default Avatar;
