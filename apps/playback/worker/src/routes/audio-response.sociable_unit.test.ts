import { describe, expect, it } from "vitest";
import { episodeAudioContentType } from "../../../contracts/index.ts";
import { createAudioResponse } from "./audio-response.ts";

describe("createAudioResponse", () => {
  it("Range 指定が無い時、subarray の可視範囲だけを契約 Content-Type の byte Response にする", async () => {
    // Given: prefix と suffix を持つ byte array の一部
    const source = new Uint8Array([0x00, 0x52, 0x49, 0x46, 0x46, 0xff]);
    const visible = source.subarray(1, 5);

    // When: Range ヘッダ無しで音声 Response を作る
    const response = createAudioResponse(visible, null);

    // Then: 可視範囲だけを 200 で返す。Range を伝える Accept-Ranges を持つ
    expect(response.status).toBe(200);
    expect(response.headers.get("Content-Type")).toBe(episodeAudioContentType);
    expect(response.headers.get("Accept-Ranges")).toBe("bytes");
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
    const response = createAudioResponse(visible, null);

    // Then: SharedArrayBuffer の全体ではなく可視範囲だけを返す
    expect(new Uint8Array(await response.arrayBuffer())).toEqual(
      new Uint8Array([0x57, 0x41, 0x56, 0x45]),
    );
  });

  it("Range: bytes=2-3 を指定すると、その範囲だけを 206 Partial Content で返す", async () => {
    // Given: 6 byte の音声（index 0-5）
    const bytes = new Uint8Array([0x10, 0x11, 0x12, 0x13, 0x14, 0x15]);

    // When: 2-3 byte 目だけを要求する
    const response = createAudioResponse(bytes, "bytes=2-3");

    // Then: 206・Content-Range・Content-Length・該当 2 byte を返す
    expect(response.status).toBe(206);
    expect(response.headers.get("Content-Range")).toBe("bytes 2-3/6");
    expect(response.headers.get("Content-Length")).toBe("2");
    expect(response.headers.get("Accept-Ranges")).toBe("bytes");
    expect(new Uint8Array(await response.arrayBuffer())).toEqual(new Uint8Array([0x12, 0x13]));
  });

  it("Range: bytes=4- のように終端省略なら、開始位置から末尾までを返す", async () => {
    // Given: 6 byte の音声
    const bytes = new Uint8Array([0x10, 0x11, 0x12, 0x13, 0x14, 0x15]);

    // When: 4 byte 目以降を要求する（終端省略）
    const response = createAudioResponse(bytes, "bytes=4-");

    // Then: 206・4-5/6・末尾までの 2 byte
    expect(response.status).toBe(206);
    expect(response.headers.get("Content-Range")).toBe("bytes 4-5/6");
    expect(new Uint8Array(await response.arrayBuffer())).toEqual(new Uint8Array([0x14, 0x15]));
  });

  it("Range の開始位置が総 byte 数以上なら 416 を返す", async () => {
    // Given: 6 byte の音声
    const bytes = new Uint8Array([0x10, 0x11, 0x12, 0x13, 0x14, 0x15]);

    // When: 総 byte 数と同じ開始位置を要求する（範囲外）
    const response = createAudioResponse(bytes, "bytes=6-");

    // Then: 416・Content-Range で総サイズを伝える
    expect(response.status).toBe(416);
    expect(response.headers.get("Content-Range")).toBe("bytes */6");
  });

  it("Range の終端が開始位置より小さい（逆順）時は、開始位置 1 byte だけを返す", async () => {
    // Given: 6 byte の音声
    const bytes = new Uint8Array([0x10, 0x11, 0x12, 0x13, 0x14, 0x15]);

    // When: 終端 2 が開始位置 4 より小さい Range を要求する
    const response = createAudioResponse(bytes, "bytes=4-2");

    // Then: end は start に丸められ、開始位置 1 byte だけを返す
    expect(response.status).toBe(206);
    expect(response.headers.get("Content-Range")).toBe("bytes 4-4/6");
    expect(new Uint8Array(await response.arrayBuffer())).toEqual(new Uint8Array([0x14]));
  });

  it("解釈できない Range ヘッダは無視し、全体を 200 で返す", async () => {
    // Given: 6 byte の音声
    const bytes = new Uint8Array([0x10, 0x11, 0x12, 0x13, 0x14, 0x15]);

    // When: 解釈不能な Range を渡す
    const response = createAudioResponse(bytes, "not-a-range");

    // Then: Range を無視して 200 全体
    expect(response.status).toBe(200);
    expect(new Uint8Array(await response.arrayBuffer())).toEqual(bytes);
  });
});
