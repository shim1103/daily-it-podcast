import { describe, expect, it } from "vitest";
import {
  GetEpisodeResponseSchema,
  ListEpisodesResponseSchema,
  episodeAudioPath,
  episodePath,
  listEpisodesPath,
} from "./http.ts";

const validTopic = {
  title: "題",
  preface: "前置き",
  detail: "詳細",
  startSec: 0,
};

const validGetEpisode = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [validTopic],
    closing: "終了",
  },
  audioRef: episodeAudioPath("ep-1"),
};

describe("episodePath", () => {
  it("episodeId に / が含まれる時、追加の path 段は 1 つのままにする", () => {
    // Given: / を含む episodeId
    const episodeId = "ep/1";

    // When: JSON 1件の path を組む
    const got = episodePath(episodeId);

    // Then: 区切りが増えない
    expect(got.split("/").length).toBe(listEpisodesPath.split("/").length + 1);
  });
});

describe("ListEpisodesResponseSchema", () => {
  it("空配列を受け入れる", () => {
    // Given: 0件の一覧
    const body = { episodes: [] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 成功する
    expect(got.success).toBe(true);
  });
});

describe("GetEpisodeResponseSchema", () => {
  it("原稿 field と audioRef がある時受け入れる", () => {
    // Given: 契約どおりの 1件
    const body = validGetEpisode;

    // When: parse する
    const got = GetEpisodeResponseSchema.safeParse(body);

    // Then: 成功する
    expect(got.success).toBe(true);
  });

  it("audioRef が無い時拒否する", () => {
    // Given: 原稿のみ
    const { audioRef: _audioRef, ...withoutAudio } = validGetEpisode;

    // When: parse する
    const got = GetEpisodeResponseSchema.safeParse(withoutAudio);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });
});
