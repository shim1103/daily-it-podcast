import { buildRequestUrl } from "../../utils/build-request-url.ts";

/**
 * baseUrl と audioRef から `<audio controls src>` を組み立てるだけの要素を作る（Contract Freeze）。
 *
 * @require baseUrl は playback worker の origin相当、audioRef は GetEpisodeResponse["audioRef"]
 * @ensure `<audio controls src={buildRequestUrl(baseUrl, audioRef)}>` を返す
 * @invariant 分岐を持たない。fetchAudio()は呼ばない
 */
export function createEpisodePlayer(baseUrl: string, audioRef: string): HTMLAudioElement {
  const audio = document.createElement("audio");
  audio.controls = true;
  audio.src = buildRequestUrl(baseUrl, audioRef);
  return audio;
}
