import type { ReactElement, RefObject } from "react";
import type { NowPlayingViewModel } from "../../view-models/playback-state.ts";
import "./audio-controls.css";

export type AudioControlsProps = {
  audioRef: RefObject<HTMLAudioElement | null>;
  nowPlaying: NowPlayingViewModel | null;
};

/**
 * 画面下部に固定される mini-player。`<audio controls>` を常に mount し続ける。
 *
 * @require audioRef は `useEpisodePlayback` が握る ref。音源 URL（`src`）はこの Component が
 *   props で受け取らず、`useEpisodePlayback` が `play` / `seek` の直前に ref 経由で
 *   命令的に張る（`<audio src>` を React で controlled にすると命令的な再生とタイミングが
 *   競合するため）。nowPlaying は再生中 episode の日付・通し番号付き title（idle なら null）
 * @ensure `<audio ref={audioRef} controls>` を画面下部固定で描画する。`src` を持たない初期状態は
 *   何も再生しない空 player（HTML 仕様：`src` も `<source>` も無ければ読み込み対象なし）。
 *   nowPlaying があれば audio（sequence bar）の上に日付と通し番号付き title を出し、
 *   null なら見出しを出さない。scroll しても再生中は画面下に貼り付く（配置・見た目は CSS 側の責務）
 */
export function AudioControls({ audioRef, nowPlaying }: AudioControlsProps): ReactElement {
  return (
    <div className="audio-controls" data-audio-controls>
      {nowPlaying !== null && (
        <p className="audio-controls__now-playing" data-now-playing>
          <span className="audio-controls__now-playing-date" data-now-playing-date>
            {nowPlaying.date}
          </span>
          <span className="audio-controls__now-playing-title" data-now-playing-title>
            {nowPlaying.numberedTitle}
          </span>
        </p>
      )}
      {/* biome-ignore lint/a11y/useMediaCaption: 原稿全文は EpisodeManuscript が担う前提。単独利用時は契約を見直す */}
      <audio ref={audioRef} controls />
    </div>
  );
}
