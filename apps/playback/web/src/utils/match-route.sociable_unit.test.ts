import { describe, expect, it } from "vitest";
import { matchRoute } from "./match-route.ts";

describe("matchRoute", () => {
  it("hash が空文字の時、episode-list になる", () => {
    // Given: 空文字の hash
    // When: matchRoute を実行する
    const route = matchRoute("");

    // Then: episode-list
    expect(route).toEqual({ kind: "episode-list" });
  });

  it("hash が '#/' の時、episode-list になる", () => {
    // Given: ルート直下を指す hash
    // When: matchRoute を実行する
    const route = matchRoute("#/");

    // Then: episode-list
    expect(route).toEqual({ kind: "episode-list" });
  });

  it("hash が '#/episodes/xxx' の時、episodeId を持つ episode-detail になる", () => {
    // Given: episodeId 付きの詳細 hash
    // When: matchRoute を実行する
    const route = matchRoute("#/episodes/xxx");

    // Then: episodeId を持つ episode-detail
    expect(route).toEqual({ kind: "episode-detail", episodeId: "xxx" });
  });

  it("hash が '#/episodes/' の時（episodeId 空）、episode-list になる", () => {
    // Given: episodeId が空の詳細風 hash
    // When: matchRoute を実行する
    const route = matchRoute("#/episodes/");

    // Then: episodeId が確定できないため episode-list
    expect(route).toEqual({ kind: "episode-list" });
  });

  it("hash が一致しない文字列の時、episode-list になる", () => {
    // Given: どの route pattern にも一致しない hash
    // When: matchRoute を実行する
    const route = matchRoute("#/unknown");

    // Then: episode-list
    expect(route).toEqual({ kind: "episode-list" });
  });
});
