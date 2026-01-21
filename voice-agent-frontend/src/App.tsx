import React from 'react';
import { Bot, Github, Info } from 'lucide-react';
import { useVoiceAgent } from './hooks/useVoiceAgent';
import { useCallStore } from './store/callStore';
import { Avatar } from './components/Avatar';
import { ChatPanel } from './components/Chat';
import { CallControls, CallStatus } from './components/Call';
import { ToolDisplay } from './components/ToolDisplay';
import { CallSummary } from './components/Summary';
import { LanguageSelector } from './components/UI/LanguageSelector';
import type { AvatarState } from './types';

const App: React.FC = () => {
  const {
    isConnected,
    isRecording,
    callState,
    avatarState,
    error,
    messages,
    currentTranscript,
    toolCalls,
    activeToolCall,
    callSummary,
    costBreakdown,
    audioLevel,
    connect,
    disconnect,
    startRecording,
    stopRecording,
    sendText,
    endCall,
  } = useVoiceAgent();

  const store = useCallStore();
  const [showInfo, setShowInfo] = React.useState(false);

  // Debug log for summary display
  React.useEffect(() => {
    console.log('[App] callState:', callState, 'callSummary:', callSummary);
    if (callSummary) {
      console.log('[App] Summary data:', JSON.stringify(callSummary, null, 2));
    }
  }, [callState, callSummary]);

  // Reset to start new call after viewing summary
  const handleCloseSummary = () => {
    disconnect();
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-100">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-agent-primary to-agent-secondary flex items-center justify-center">
                <Bot className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="font-bold text-gray-800">AI Voice Agent</h1>
                <p className="text-xs text-gray-500">Appointment Assistant</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <LanguageSelector />
              <button
                onClick={() => setShowInfo(!showInfo)}
                className="p-2 rounded-lg hover:bg-gray-100 transition-colors"
                aria-label="Info"
              >
                <Info className="w-5 h-5 text-gray-500" />
              </button>
              <a
                href="https://github.com"
                target="_blank"
                rel="noopener noreferrer"
                className="p-2 rounded-lg hover:bg-gray-100 transition-colors"
                aria-label="GitHub"
              >
                <Github className="w-5 h-5 text-gray-500" />
              </a>
            </div>
          </div>
        </div>
      </header>

      {/* Info banner */}
      {showInfo && (
        <div className="bg-blue-50 border-b border-blue-100 p-4">
          <div className="max-w-7xl mx-auto">
            <div className="flex items-start gap-3">
              <Info className="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-blue-800">
                <p className="font-medium mb-1">How to use:</p>
                <ol className="list-decimal list-inside space-y-1 text-blue-700">
                  <li>Click "Start Call" to connect with the AI assistant</li>
                  <li>Click the microphone button to start speaking</li>
                  <li>You can book, check, modify, or cancel appointments</li>
                  <li>Say "goodbye" or click "End Call" when you're done</li>
                </ol>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Main content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Show summary modal when call ends */}
        {callSummary && callState === 'ended' && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <div className="max-w-lg w-full max-h-[90vh] overflow-y-auto">
              <CallSummary
                summary={callSummary}
                costBreakdown={costBreakdown}
                onClose={handleCloseSummary}
              />
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left column - Avatar and Controls */}
          <div className="lg:col-span-2 space-y-6">
            {/* Avatar */}
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <Avatar
                state={avatarState as AvatarState}
                iframeUrl={store.avatarUrl || undefined}
                className="w-full aspect-video"
              />

              {/* Call status */}
              <CallStatus
                callState={callState}
                error={error}
                className="mt-4"
              />

              {/* Controls */}
              <CallControls
                callState={callState}
                isRecording={isRecording}
                isConnected={isConnected}
                audioLevel={audioLevel}
                onConnect={connect}
                onDisconnect={disconnect}
                onStartRecording={startRecording}
                onStopRecording={stopRecording}
                onEndCall={endCall}
                className="mt-6"
              />
            </div>

            {/* Tool Activity (shown when there are tool calls) */}
            {toolCalls.length > 0 && (
              <ToolDisplay
                toolCalls={toolCalls}
                activeToolCall={activeToolCall}
              />
            )}
          </div>

          {/* Right column - Chat */}
          <div className="lg:col-span-1">
            <ChatPanel
              messages={messages}
              currentTranscript={currentTranscript}
              isTranscriptFinal={false}
              onSendMessage={sendText}
              isConnected={isConnected}
              className="h-full"
            />
          </div>
        </div>

        {/* Instructions for first-time users */}
        {callState === 'idle' && messages.length === 0 && (
          <div className="mt-8">
            <div className="bg-gradient-to-r from-agent-primary/10 to-agent-secondary/10 rounded-2xl p-6">
              <h2 className="text-lg font-semibold text-gray-800 mb-4">
                Welcome to AI Voice Agent
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <FeatureCard
                  title="Book Appointments"
                  description="Schedule new appointments by speaking naturally"
                  icon="📅"
                />
                <FeatureCard
                  title="Manage Schedule"
                  description="View, modify, or cancel existing appointments"
                  icon="🔄"
                />
                <FeatureCard
                  title="Get Summary"
                  description="Receive a detailed summary of your call"
                  icon="📝"
                />
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-200 bg-white mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-500">
            Built with LiveKit, Deepgram, Cartesia, and OpenAI
          </p>
        </div>
      </footer>
    </div>
  );
};

interface FeatureCardProps {
  title: string;
  description: string;
  icon: string;
}

const FeatureCard: React.FC<FeatureCardProps> = ({ title, description, icon }) => (
  <div className="bg-white rounded-xl p-4 shadow-sm">
    <div className="text-2xl mb-2">{icon}</div>
    <h3 className="font-medium text-gray-800 mb-1">{title}</h3>
    <p className="text-sm text-gray-500">{description}</p>
  </div>
);

export default App;
