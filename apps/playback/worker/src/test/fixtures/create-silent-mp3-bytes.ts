export const samplesPerFrame = 576;

export const sampleRate = 8_000;

export const silentFrameByteLength = 72;

// why: libmp3lame の 8 kHz / 8 kbps / mono 無音 1 frame（72 byte CBR）を再現する。frame 連結だけで尺が立つ。
const silentFrame = (() => {
  const frame = new Uint8Array(silentFrameByteLength).fill(0x55);
  frame.set([0xff, 0xe3, 0x18, 0xc4, 0x3b, 0x00, 0x00, 0x03, 0x48, 0x00, 0x00, 0x00, 0x00], 0);
  frame.set([0x4c, 0x41, 0x4d, 0x45, 0x34, 0x2e, 0x30], 55);
  return frame;
})();

/**
 * 無音の再生可能な mp3 bytes を組み立てる。
 *
 * @require なし（負の durationSec は 0 扱い。throw しない）
 * @ensure MPEG sync 付き CBR silent frame の連結を返す。0 でも最低 1 frame。再生尺は max(0, durationSec) に近似一致する
 * @invariant MPEG2.5 Layer III / 8 kHz / 8 kbps / mono
 */
export function createSilentMp3Bytes(durationSec: number): Uint8Array {
  const totalSec = Math.max(0, durationSec);
  const frameCount = Math.max(1, Math.round((totalSec * sampleRate) / samplesPerFrame));
  const bytes = new Uint8Array(frameCount * silentFrame.byteLength);
  for (let i = 0; i < frameCount; i += 1) {
    bytes.set(silentFrame, i * silentFrame.byteLength);
  }
  return bytes;
}
