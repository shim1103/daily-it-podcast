import { describe, expect, it } from "vitest";
import { episodeAudioContentType } from "../../../contracts/index.ts";
import { createAudioResponse } from "./audio-response.ts";

describe("createAudioResponse", () => {
  it("subarray の可視範囲だけを契約 Content-Type の byte Response にする", async () => {
    // Given: prefix と suffix を持つ byte array の一部
    const source = new Uint8Array([0x00, 0x52, 0x49, 0x46, 0x46, 0xff]);
    const visible = source.subarray(1, 5);

    // When: 音声 Response を作る
    const response = createAudioResponse(visible);

    // Then: 可視範囲だけを返す
    expect(response.status).toBe(200);
    expect(response.headers.get("Content-Type")).toBe(episodeAudioContentType);
    expect(new Uint8Array(await response.arrayBuffer())).toEqual(
      new Uint8Array([0x52, 0x49, 0x46, 0x46]),
    );
  });

  it("SharedArrayBuffer の byte array を copy して Response にする", async () => {
    // Given: SharedArrayBuffer 上の可視範囲
    const buffer = new SharedArrayBuffer(6);
    const source = new Uint8Array(buffer);
    source.set([0x00, 0x57, 0x41, 0x56, 0x45, 0xff]);
    const visible = source.subarray(1, 5);

    // When: 音声 Response を作る
    const response = createAudioResponse(visible);

    // Then: SharedArrayBuffer の全体ではなく可視範囲だけを返す
    expect(new Uint8Array(await response.arrayBuffer())).toEqual(
      new Uint8Array([0x57, 0x41, 0x56, 0x45]),
    );
  });
});
