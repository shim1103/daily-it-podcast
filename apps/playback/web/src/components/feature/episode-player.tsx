import { forwardRef, type ReactElement } from "react";
import { buildRequestUrl } from "../../utils/build-request-url.ts";
import "./episode-player.css";

export type EpisodePlayerProps = {
  baseUrl: string;
  audioRef: string;
};

/**
 * baseUrl と audioRef から `<audio controls src>` を組み立てる。
 *
 * @require baseUrl は playback worker の origin相当、audioRef は GetEpisodeResponse["audioRef"]
 * @ensure `<audio controls src={buildRequestUrl(baseUrl, audioRef)}>` を返す。ref は audio 要素へ転送する
 * @invariant fetchAudio()は呼ばない。EpisodeList の SelectedEpisodeDetail により
 *   EpisodeManuscript との対描画と topic シークを前提とする
 */
export const EpisodePlayer = forwardRef<HTMLAudioElement, EpisodePlayerProps>(
  function EpisodePlayer({ baseUrl, audioRef }, ref): ReactElement {
    return (
      <div className="episode-player">
        {/* biome-ignore lint/a11y/useMediaCaption: <track>形式の同期字幕は持たないが、EpisodeList の contract により EpisodePlayer は常に EpisodeManuscript（原稿全文）と対で描画され、音声内容のテキストアクセス手段は原稿全文が担うため a11y 要件は満たす。EpisodePlayer が単独利用される設計に変わった場合はこの前提が崩れる */}
        <audio ref={ref} controls src={buildRequestUrl(baseUrl, audioRef)} />
      </div>
    );
  },
);
