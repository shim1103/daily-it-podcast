import { describe, expect, it } from "vitest";
import { listEpisodesPath } from "../../../contracts/index.ts";
import { matchPlaybackRoute } from "./match-playback-route.ts";

describe("matchPlaybackRoute", () => {
  it("GET かつ一覧 path の時、list に一致する", () => {
    // Given: 一覧 path
    // When: 一覧 path へ GET する
    const got = matchPlaybackRoute("GET", listEpisodesPath);

    // Then: list に一致する
    expect(got).toEqual({ kind: "list" });
  });

  it("GET でない時、path に関わらず unmatched になる", () => {
    // Given: 一覧 path への POST
    // When: 一覧 path へ POST する
    const got = matchPlaybackRoute("POST", listEpisodesPath);

    // Then: unmatched になる
    expect(got).toEqual({ kind: "unmatched" });
  });

  it("GET かつ 1件 path の時、episodeId を decode して get に一致する", () => {
    // Given: 1件 path
    // When: 1件 path へ GET する
    const got = matchPlaybackRoute("GET", `${listEpisodesPath}/ep-1`);

    // Then: episodeId を持つ get に一致する
    expect(got).toEqual({ kind: "get", episodeId: "ep-1" });
  });

  it("GET かつ音声 path の時、episodeId を decode して audio に一致する", () => {
    // Given: 音声 path
    // When: 音声 path へ GET する
    const got = matchPlaybackRoute("GET", `${listEpisodesPath}/ep-1/audio`);

    // Then: episodeId を持つ audio に一致する
    expect(got).toEqual({ kind: "audio", episodeId: "ep-1" });
  });

  it("1件 path の episodeId が decode 不可能な時、encode されたまま返す", () => {
    // Given: 単独 % を含む decode 不可能な episodeId
    // When: 単独 % を含む 1件 path へ GET する
    const got = matchPlaybackRoute("GET", `${listEpisodesPath}/100%`);

    // Then: decode を諦めそのまま返す
    expect(got).toEqual({ kind: "get", episodeId: "100%" });
  });

  it("一覧 path の prefix に一致しない時、unmatched になる", () => {
    // Given: 一覧 path と別系統の path
    // When: 別系統の path へ GET する
    const got = matchPlaybackRoute("GET", "/unknown");

    // Then: unmatched になる
    expect(got).toEqual({ kind: "unmatched" });
  });

  it("2 段目が audio でない時、unmatched になる", () => {
    // Given: 2 段目が audio ではない path
    // When: 2 段目が audio でない path へ GET する
    const got = matchPlaybackRoute("GET", `${listEpisodesPath}/ep-1/unknown`);

    // Then: unmatched になる
    expect(got).toEqual({ kind: "unmatched" });
  });

  it("3 段以上の path の時、unmatched になる", () => {
    // Given: 3 段の path
    // When: 3 段の path へ GET する
    const got = matchPlaybackRoute("GET", `${listEpisodesPath}/ep-1/audio/extra`);

    // Then: unmatched になる
    expect(got).toEqual({ kind: "unmatched" });
  });
});
