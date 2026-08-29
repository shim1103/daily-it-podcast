import { createSilentWavBytes } from "./create-silent-wav-bytes.ts";

/** テスト用の再生可能な無音 WAV（1 秒・24kHz）。 */
export const validAudioBytes = createSilentWavBytes(1);

/**
 * dev fake 音声の sample rate。HTML audio が再生できる 8kHz（長尺でも尺は durationSec と一致）。
 */
export const fakeAudioSampleRate = 8_000;

/**
 * fake episode 用の再生可能な無音 WAV bytes を返す。尺は durationSec と一致する。
 */
export function createFakeEpisodeAudioBytes(durationSec: number): Uint8Array {
  return createSilentWavBytes(Math.max(1, durationSec), fakeAudioSampleRate);
}
