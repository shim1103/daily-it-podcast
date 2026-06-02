'use client';

interface Props {
  audioUrl: string;
}

export function AudioPlayer({ audioUrl }: Props) {
  return (
    <div className="my-4">
      {/* MVPではモックURLのためaudioは再生不可。将来は実音声URLに差し替える。 */}
      <audio
        controls
        src={audioUrl}
        className="w-full rounded-xl"
      >
        お使いのブラウザはaudio要素をサポートしていません。
      </audio>
    </div>
  );
}
