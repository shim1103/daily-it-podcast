import { forwardRef, type ReactElement } from "react";
import "./episode-audio.css";

export type EpisodeAudioProps = {
  src?: string;
};

/**
 * 再生中 EpisodeListEntry 用の宣言的 `<audio>` のみ。再生 state は持たない。
 *
 * @require src は Page が useEpisodePlayback から解決した URL、または未再生時は undefined
 * @ensure hidden `<audio>` を 1 つ返し ref を要素へ転送する
 * @invariant play / seek / timeupdate を持たない。Page / List root には置かない
 */
export const EpisodeAudio = forwardRef<HTMLAudioElement, EpisodeAudioProps>(
  function EpisodeAudio({ src }, ref): ReactElement {
    return (
      <div className="episode-audio">
        {/* biome-ignore lint/a11y/useMediaCaption: <track>形式の同期字幕は持たないが、EpisodeList の contract により EpisodeAudio は常に EpisodeManuscript（原稿全文）と対で描画され、音声内容のテキストアクセス手段は原稿全文が担うため a11y 要件は満たす。EpisodeAudio が単独利用される設計に変わった場合はこの前提が崩れる */}
        <audio ref={ref} className="episode-audio__element" src={src} />
      </div>
    );
  },
);
