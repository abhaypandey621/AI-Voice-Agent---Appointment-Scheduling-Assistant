import React, { useRef, useEffect, useState } from 'react';
import { clsx } from 'clsx';
import { Send, MessageSquare } from 'lucide-react';
import { ChatMessage, TranscriptIndicator } from './ChatMessage';
import type { ConversationMessage } from '../../types';

interface ChatPanelProps {
  messages: ConversationMessage[];
  currentTranscript: string;
  isTranscriptFinal: boolean;
  onSendMessage: (text: string) => void;
  isConnected: boolean;
  className?: string;
}

export const ChatPanel: React.FC<ChatPanelProps> = ({
  messages,
  currentTranscript,
  isTranscriptFinal,
  onSendMessage,
  isConnected,
  className,
}) => {
  const [inputText, setInputText] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, currentTranscript]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputText.trim() && isConnected) {
      onSendMessage(inputText.trim());
      setInputText('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <div className={clsx('flex flex-col bg-white rounded-2xl shadow-lg', className)}>
      {/* Header */}
      <div className="flex items-center gap-3 px-4 py-3 border-b border-gray-100">
        <div className="w-8 h-8 rounded-full bg-agent-primary/10 flex items-center justify-center">
          <MessageSquare className="w-4 h-4 text-agent-primary" />
        </div>
        <div>
          <h3 className="font-semibold text-gray-800">Conversation</h3>
          <p className="text-xs text-gray-500">
            {messages.length} message{messages.length !== 1 ? 's' : ''}
          </p>
        </div>
      </div>

      {/* Messages area */}
      <div className="flex-1 overflow-y-auto min-h-[300px] max-h-[500px]">
        {messages.length === 0 && !currentTranscript && (
          <div className="flex flex-col items-center justify-center h-full text-gray-400 py-8">
            <MessageSquare className="w-12 h-12 mb-3 opacity-50" />
            <p className="text-sm">Start a conversation</p>
            <p className="text-xs mt-1">Speak or type your message</p>
          </div>
        )}

        {messages.map((message, index) => (
          <ChatMessage key={`${message.timestamp}-${index}`} message={message} />
        ))}

        {/* Show current transcript */}
        <TranscriptIndicator
          text={currentTranscript}
          isFinal={isTranscriptFinal}
        />

        <div ref={messagesEndRef} />
      </div>

      {/* Input area */}
      <form onSubmit={handleSubmit} className="p-4 border-t border-gray-100">
        <div className="flex gap-2">
          <input
            ref={inputRef}
            type="text"
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={isConnected ? 'Type a message...' : 'Connect to start chatting'}
            disabled={!isConnected}
            className={clsx(
              'flex-1 px-4 py-2 rounded-xl border border-gray-200',
              'focus:outline-none focus:ring-2 focus:ring-agent-primary/50 focus:border-agent-primary',
              'text-sm transition-all duration-200',
              'disabled:bg-gray-50 disabled:cursor-not-allowed'
            )}
          />
          <button
            type="submit"
            disabled={!isConnected || !inputText.trim()}
            className={clsx(
              'px-4 py-2 rounded-xl bg-agent-primary text-white',
              'hover:bg-agent-secondary transition-colors duration-200',
              'disabled:opacity-50 disabled:cursor-not-allowed',
              'flex items-center justify-center'
            )}
          >
            <Send className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};

export default ChatPanel;
