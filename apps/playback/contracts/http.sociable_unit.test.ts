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

const validListItem = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  topics: [{ title: "題1" }, { title: "題2" }],
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

  it("episode が topics を持つ時受け入れる", () => {
    // Given: topics 付きの episode 1件
    const body = { episodes: [validListItem] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 成功する
    expect(got.success).toBe(true);
  });

  it("topics が空配列の episode を受け入れる", () => {
    // Given: topics が 0 件の episode
    const body = { episodes: [{ ...validListItem, topics: [] }] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 成功する（0 件以上を許容する）
    expect(got.success).toBe(true);
  });

  it("episode に topics が無い時拒否する", () => {
    // Given: topics 欠落の episode
    const { topics: _topics, ...withoutTopics } = validListItem;
    const body = { episodes: [withoutTopics] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("topics 要素に title 以外の余剰 field が混ざる時拒否する", () => {
    // Given: topic 要素へ preface を足した episode
    const body = {
      episodes: [{ ...validListItem, topics: [{ title: "題1", preface: "前置き" }] }],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 失敗する（list の topic は title だけの strict object）
    expect(got.success).toBe(false);
  });

  it("topics 要素の title が空文字の時拒否する", () => {
    // Given: title が空文字の topic を持つ episode
    const body = { episodes: [{ ...validListItem, topics: [{ title: "" }] }] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("episode に契約外の field を足す時拒否する", () => {
    // Given: 契約外 field を持つ episode
    const body = { episodes: [{ ...validListItem, extra: 1 }] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 失敗する（strict object を維持する）
    expect(got.success).toBe(false);
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
