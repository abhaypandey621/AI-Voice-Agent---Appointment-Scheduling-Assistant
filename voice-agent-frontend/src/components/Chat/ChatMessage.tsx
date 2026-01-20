import React from 'react';
import { clsx } from 'clsx';
import { format } from 'date-fns';
import { User, Bot } from 'lucide-react';
import type { ConversationMessage } from '../../types';

interface ChatMessageProps {
  message: ConversationMessage;
}

export const ChatMessage: React.FC<ChatMessageProps> = ({ message }) => {
  const isUser = message.role === 'user';
  const isSystem = message.role === 'system';

  if (isSystem) {
    return (
      <div className="flex justify-center py-2">
        <span className="text-xs text-gray-500 bg-gray-100 px-3 py-1 rounded-full">
          {message.content}
        </span>
      </div>
    );
  }

  return (
    <div
      className={clsx(
        'flex gap-3 p-4 rounded-xl animate-slide-up',
        isUser ? 'flex-row-reverse' : 'flex-row'
      )}
    >
      {/* Avatar */}
      <div
        className={clsx(
          'flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center',
          isUser ? 'bg-blue-500' : 'bg-agent-primary'
        )}
      >
        {isUser ? (
          <User className="w-4 h-4 text-white" />
        ) : (
          <Bot className="w-4 h-4 text-white" />
        )}
      </div>

      {/* Message bubble */}
      <div
        className={clsx(
          'max-w-[80%] rounded-2xl px-4 py-3',
          isUser
            ? 'bg-blue-500 text-white rounded-br-md'
            : 'bg-gray-100 text-gray-800 rounded-bl-md'
        )}
      >
        <p className="text-sm leading-relaxed whitespace-pre-wrap">
          {message.content}
        </p>
        <span
          className={clsx(
            'text-xs mt-1 block',
            isUser ? 'text-blue-100' : 'text-gray-400'
          )}
        >
          {format(new Date(message.timestamp), 'HH:mm')}
        </span>
      </div>
    </div>
  );
};

interface TranscriptIndicatorProps {
  text: string;
  isFinal: boolean;
}

export const TranscriptIndicator: React.FC<TranscriptIndicatorProps> = ({
  text,
  isFinal,
}) => {
  if (!text) return null;

  return (
    <div className="flex gap-3 p-4">
      <div className="flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center bg-blue-500">
        <User className="w-4 h-4 text-white" />
      </div>
      <div className="bg-blue-100 text-blue-800 rounded-2xl rounded-bl-md px-4 py-3 max-w-[80%]">
        <p className="text-sm leading-relaxed">
          {text}
          {!isFinal && (
            <span className="inline-block ml-1 animate-pulse">...</span>
          )}
        </p>
      </div>
    </div>
  );
};

export default ChatMessage;
