import { describe, expect, it } from "vitest";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";
import { verifyAudio } from "./verify-audio.ts";

describe("verifyAudio", () => {
  it("空の byte 列の時、EpisodeContentError を throw し message は音声が空を示す", () => {
    // Given: 0 byte の音声
    const bytes = new Uint8Array([]);

    // When / Then: 空は内容不適合
    expect(() => verifyAudio(bytes, "ep-1")).toThrow(EpisodeContentError);
    expect(() => verifyAudio(bytes, "ep-1")).toThrow("音声が空: ep-1");
  });

  it("ID3v2 tag で始まる時、同じ参照をそのまま返す", () => {
    // Given: 先頭 3 byte が "ID3"
    const bytes = new Uint8Array([0x49, 0x44, 0x33, 0x04, 0x00]);

    // When: 検証する
    const got = verifyAudio(bytes, "ep-1");

    // Then: 変換もコピーもされない
    expect(got).toBe(bytes);
  });

  it("MPEG audio の frame sync で始まる時、同じ参照をそのまま返す", () => {
    // Given: 実 fixture の無音 mp3（先頭 0xFF 0xE3）
    // When: 検証する
    const got = verifyAudio(validAudioBytes, "ep-1");

    // Then: 変換もコピーもされない
    expect(got).toBe(validAudioBytes);
  });

  it("署名が無い時、EpisodeContentError を throw し message は mp3 でないを示す", () => {
    // Given: ID3 でも frame sync でもない byte 列
    const bytes = new Uint8Array([0x00, 0x01, 0x02]);

    // When / Then
    expect(() => verifyAudio(bytes, "ep-1")).toThrow(EpisodeContentError);
    expect(() => verifyAudio(bytes, "ep-1")).toThrow("音声が mp3 でない: ep-1");
  });

  it("先頭が 0xFF だが 2 byte 目が frame sync でない時、EpisodeContentError を throw する", () => {
    // Given: 2 byte 目の上位 3 bit が立たない
    const bytes = new Uint8Array([0xff, 0x00]);

    // When / Then
    expect(() => verifyAudio(bytes, "ep-1")).toThrow("音声が mp3 でない: ep-1");
  });

  it("先頭が ID だけで 3 byte 目が違う時、EpisodeContentError を throw する", () => {
    // Given: "ID" の後が "3" でない
    const bytes = new Uint8Array([0x49, 0x44, 0x00]);

    // When / Then
    expect(() => verifyAudio(bytes, "ep-1")).toThrow("音声が mp3 でない: ep-1");
  });

  it("1 byte だけの時、EpisodeContentError を throw する", () => {
    // Given: 0xFF 1 byte のみ（2 byte 目が無い）
    const bytes = new Uint8Array([0xff]);

    // When / Then
    expect(() => verifyAudio(bytes, "ep-1")).toThrow("音声が mp3 でない: ep-1");
  });
});
