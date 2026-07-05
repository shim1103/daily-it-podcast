'use client';

interface Props {
  audioUrl: string;
}

export function AudioPlayer({ audioUrl }: Props) {
  return (
    <div className="mb-6 bg-[#111827] border border-white/8 rounded-2xl px-5 py-4">
      <div className="flex items-center gap-2 mb-3">
        <span className="text-purple-400 text-sm font-medium">🎧 音声</span>
      </div>
      {/* MVPではモックURLのためaudioは再生不可。将来は実音声URLに差し替える。 */}
      <audio
        controls
        src={audioUrl}
        className="w-full rounded-lg"
        style={{ accentColor: '#a855f7', colorScheme: 'dark' }}
      >
        お使いのブラウザはaudio要素をサポートしていません。
      </audio>
    </div>
  );
}
