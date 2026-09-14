import { describe, expect, it } from "vitest";
import {
  createSilentMp3Bytes,
  sampleRate,
  samplesPerFrame,
  silentFrameByteLength,
} from "./create-silent-mp3-bytes.ts";

/** createSilentMp3Bytes 圧縮前の 1 frame golden（libmp3lame 72 byte CBR）。 */
const GOLDEN_SILENT_FRAME_HEX =
  "ffe318c43b00000348000000005555555555555555555555555555555555555555555555555555555555555555555555555555555555554c414d45342e3055555555555555555555";
const GOLDEN_SILENT_FRAME = Uint8Array.from(
  Array.from({ length: GOLDEN_SILENT_FRAME_HEX.length / 2 }, (_, i) =>
    Number.parseInt(GOLDEN_SILENT_FRAME_HEX.slice(i * 2, i * 2 + 2), 16),
  ),
);

describe("createSilentMp3Bytes", () => {
  it("1 秒分の MPEG sync 付き silent mp3 bytes を返す", () => {
    // Given: 1 秒
    const got = createSilentMp3Bytes(1);
    const expectedFrameCount = Math.max(1, Math.round((1 * sampleRate) / samplesPerFrame));
    const expected = new Uint8Array(expectedFrameCount * silentFrameByteLength);
    for (let i = 0; i < expectedFrameCount; i += 1) {
      expected.set(GOLDEN_SILENT_FRAME, i * silentFrameByteLength);
    }

    // Then: 圧縮前 golden と全 byte 一致し、尺は 1 秒に近似
    expect(GOLDEN_SILENT_FRAME.byteLength).toBe(silentFrameByteLength);
    expect(got).toEqual(expected);
    expect((expectedFrameCount * samplesPerFrame) / sampleRate).toBeCloseTo(1, 1);
  });
});
