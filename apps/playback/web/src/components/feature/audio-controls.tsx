import type { ReactElement, RefObject } from "react";

export type AudioControlsProps = {
  audioRef: RefObject<HTMLAudioElement | null>;
  audioSrc: string;
};

/**
 * audioRef と audioSrc から再生 controls を組み立てる（契約 stub）。
 *
 * @require audioSrc は再生対象 episode の URL
 * @ensure `<audio controls>` を placeholder として描画する
 */
export function AudioControls({ audioRef, audioSrc }: AudioControlsProps): ReactElement {
  return (
    <div className="audio-controls">
      {/* biome-ignore lint/a11y/useMediaCaption: 原稿全文は EpisodeManuscript が担う前提。単独利用時は契約を見直す */}
      <audio ref={audioRef} controls src={audioSrc} />
    </div>
  );
}
