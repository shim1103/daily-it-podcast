import { createSilentMp3Bytes } from "./create-silent-mp3-bytes.ts";

/** テスト用の再生可能な無音 mp3（1 秒・8kHz）。 */
export const validAudioBytes = createSilentMp3Bytes(1);

/**
 * fake episode 用の再生可能な無音 mp3 bytes を返す。尺は durationSec に近似一致する。
 */
export function createFakeEpisodeAudioBytes(durationSec: number): Uint8Array {
  return createSilentMp3Bytes(Math.max(1, durationSec));
}
