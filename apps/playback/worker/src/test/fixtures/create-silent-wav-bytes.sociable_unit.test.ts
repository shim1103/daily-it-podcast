import { describe, expect, it } from "vitest";
import { createSilentWavBytes } from "./create-silent-wav-bytes.ts";

describe("createSilentWavBytes", () => {
  it("1 秒分の RIFF/WAVE ヘッダ付き bytes を返す", () => {
    // Given: 1 秒
    const got = createSilentWavBytes(1);

    // Then: RIFF....WAVE で始まり、24kHz mono 16-bit 相当の data サイズを持つ
    expect(got[0]).toBe(0x52);
    expect(got[1]).toBe(0x49);
    expect(got[2]).toBe(0x46);
    expect(got[3]).toBe(0x46);
    expect(String.fromCharCode(got[8] ?? 0, got[9] ?? 0, got[10] ?? 0, got[11] ?? 0)).toBe("WAVE");
    const dataSize = new DataView(got.buffer).getUint32(40, true);
    expect(dataSize).toBe(24_000 * 2);
  });
});
