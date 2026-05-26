'use client';

interface Props {
  audioUrl: string;
}

export function AudioPlayer({ audioUrl }: Props) {
  return (
    <div style={{ margin: '16px 0' }}>
      {/* MVPではモックURLのためaudioは再生不可。将来は実音声URLに差し替える。 */}
      <audio
        controls
        src={audioUrl}
        style={{ width: '100%' }}
      >
        お使いのブラウザはaudio要素をサポートしていません。
      </audio>
    </div>
  );
}
