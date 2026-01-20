import { useCallback, useRef, useState, useEffect } from 'react';

interface UseAudioOptions {
  onAudioData?: (data: Float32Array) => void;
  targetSampleRate?: number;
}

interface UseAudioReturn {
  isRecording: boolean;
  isSupported: boolean;
  startRecording: () => Promise<void>;
  stopRecording: () => void;
  audioLevel: number;
}

// Resample audio from source rate to target rate
function resampleAudio(
  inputData: Float32Array,
  sourceSampleRate: number,
  targetSampleRate: number
): Float32Array {
  if (sourceSampleRate === targetSampleRate) {
    return inputData;
  }

  const ratio = sourceSampleRate / targetSampleRate;
  const outputLength = Math.round(inputData.length / ratio);
  const output = new Float32Array(outputLength);

  for (let i = 0; i < outputLength; i++) {
    const sourceIndex = i * ratio;
    const index = Math.floor(sourceIndex);
    const fraction = sourceIndex - index;

    if (index + 1 < inputData.length) {
      // Linear interpolation
      output[i] = inputData[index] * (1 - fraction) + inputData[index + 1] * fraction;
    } else {
      output[i] = inputData[index];
    }
  }

  return output;
}

export function useAudio(options: UseAudioOptions = {}): UseAudioReturn {
  const { onAudioData, targetSampleRate = 16000 } = options;

  const [isRecording, setIsRecording] = useState(false);
  const [isSupported, setIsSupported] = useState(true);
  const [audioLevel, setAudioLevel] = useState(0);

  const mediaStreamRef = useRef<MediaStream | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  const scriptProcessorRef = useRef<ScriptProcessorNode | null>(null);

  useEffect(() => {
    // Check if audio recording is supported
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      setIsSupported(false);
    }

    return () => {
      stopRecording();
    };
  }, []);

  const startRecording = useCallback(async () => {
    if (isRecording) return;

    try {
      // Get microphone access - let browser choose optimal sample rate
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
        },
      });

      mediaStreamRef.current = stream;

      // Create audio context with default sample rate (browser native)
      const audioContext = new AudioContext();
      audioContextRef.current = audioContext;

      const sourceSampleRate = audioContext.sampleRate;
      console.log(`Audio context sample rate: ${sourceSampleRate}Hz, target: ${targetSampleRate}Hz`);

      // Create source from microphone
      const source = audioContext.createMediaStreamSource(stream);

      // Create analyser for audio level visualization
      const analyser = audioContext.createAnalyser();
      analyser.fftSize = 256;
      analyserRef.current = analyser;
      source.connect(analyser);

      // Create script processor for raw audio data
      const bufferSize = 4096;
      const scriptProcessor = audioContext.createScriptProcessor(
        bufferSize,
        1,
        1
      );
      scriptProcessorRef.current = scriptProcessor;

      scriptProcessor.onaudioprocess = (event) => {
        if (onAudioData) {
          const inputData = event.inputBuffer.getChannelData(0);
          // Resample to target sample rate (16kHz for Deepgram)
          const resampledData = resampleAudio(
            new Float32Array(inputData),
            sourceSampleRate,
            targetSampleRate
          );
          onAudioData(resampledData);
        }
      };

      source.connect(scriptProcessor);
      scriptProcessor.connect(audioContext.destination);

      // Start audio level monitoring
      const updateAudioLevel = () => {
        if (!analyserRef.current) return;

        const dataArray = new Uint8Array(analyserRef.current.frequencyBinCount);
        analyserRef.current.getByteFrequencyData(dataArray);

        // Calculate average level
        const average = dataArray.reduce((a, b) => a + b) / dataArray.length;
        setAudioLevel(average / 255);

        animationFrameRef.current = requestAnimationFrame(updateAudioLevel);
      };
      updateAudioLevel();

      setIsRecording(true);
      console.log('Recording started');
    } catch (error) {
      console.error('Failed to start recording:', error);
      throw error;
    }
  }, [isRecording, onAudioData, targetSampleRate]);

  const stopRecording = useCallback(() => {
    // Stop animation frame
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
      animationFrameRef.current = null;
    }

    // Stop media stream
    if (mediaStreamRef.current) {
      mediaStreamRef.current.getTracks().forEach((track: MediaStreamTrack) => track.stop());
      mediaStreamRef.current = null;
    }

    // Close audio context
    if (audioContextRef.current) {
      audioContextRef.current.close();
      audioContextRef.current = null;
    }

    setIsRecording(false);
    setAudioLevel(0);
  }, []);

  return {
    isRecording,
    isSupported,
    startRecording,
    stopRecording,
    audioLevel,
  };
}

// Utility function to convert Float32Array to Int16Array (for sending to server)
export function float32ToInt16(float32Array: Float32Array): Int16Array {
  const int16Array = new Int16Array(float32Array.length);
  for (let i = 0; i < float32Array.length; i++) {
    const s = Math.max(-1, Math.min(1, float32Array[i]));
    int16Array[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
  }
  return int16Array;
}

// Utility function to convert Int16Array to ArrayBuffer
export function int16ToArrayBuffer(int16Array: Int16Array): ArrayBuffer {
  const buffer = new ArrayBuffer(int16Array.byteLength);
  const view = new Int16Array(buffer);
  view.set(int16Array);
  return buffer;
}

export default useAudio;
