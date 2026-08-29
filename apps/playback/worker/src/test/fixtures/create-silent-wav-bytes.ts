/**
 * 無音 PCM の再生可能な WAV bytes を組み立てる。
 *
 * @require durationSec は 0 以上
 * @ensure RIFF/WAVE ヘッダ付きの PCM WAV を返す。実際の再生尺は durationSec に一致する
 * @invariant mono / 16-bit PCM
 */
export function createSilentWavBytes(durationSec: number, sampleRate = 24_000): Uint8Array {
  const channels = 1;
  const bitDepth = 16;
  const totalSec = Math.max(0, durationSec);
  const numSamples = Math.floor(totalSec * sampleRate);
  const pcmByteLength = numSamples * channels * (bitDepth / 8);
  const buffer = new ArrayBuffer(44 + pcmByteLength);
  const view = new DataView(buffer);

  writeAscii(view, 0, "RIFF");
  view.setUint32(4, 36 + pcmByteLength, true);
  writeAscii(view, 8, "WAVE");
  writeAscii(view, 12, "fmt ");
  view.setUint32(16, 16, true);
  view.setUint16(20, 1, true);
  view.setUint16(22, channels, true);
  view.setUint32(24, sampleRate, true);
  view.setUint32(28, (sampleRate * channels * bitDepth) / 8, true);
  view.setUint16(32, (channels * bitDepth) / 8, true);
  view.setUint16(34, bitDepth, true);
  writeAscii(view, 36, "data");
  view.setUint32(40, pcmByteLength, true);

  return new Uint8Array(buffer);
}

function writeAscii(view: DataView, offset: number, text: string): void {
  for (let i = 0; i < text.length; i += 1) {
    view.setUint8(offset + i, text.charCodeAt(i));
  }
}
