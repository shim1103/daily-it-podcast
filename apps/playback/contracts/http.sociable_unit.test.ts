import { describe, expect, it } from "vitest";
import {
  episodeAudioPath,
  episodePath,
  ListEpisodesResponseSchema,
  listEpisodesPath,
} from "./http.ts";

const validTopic = {
  title: "題",
  preface: "前置き",
  detail: "詳細",
  startSec: 0,
};

const validOpening = {
  text: "開始",
  startSec: 0,
};

const validEnding = {
  text: "終了",
  startSec: 30,
};

const validEpisodeItem = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: validOpening,
    topics: [validTopic],
    ending: validEnding,
  },
  audioRef: episodeAudioPath("ep-1"),
};

describe("episodePath", () => {
  it("episodeId に / が含まれる時、追加の path 段は 1 つのままにする", () => {
    // Given: / を含む episodeId
    const episodeId = "ep/1";

    // When: episode path を組む
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

  it("episode が body 全文と audioRef を持つ時受け入れる", () => {
    // Given: 原稿全文付きの episode 1件
    const body = { episodes: [validEpisodeItem] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 成功する
    expect(got.success).toBe(true);
  });

  it("episode に body が無い時拒否する", () => {
    // Given: body 欠落の episode
    const { body: _body, ...withoutBody } = validEpisodeItem;
    const payload = { episodes: [withoutBody] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("topics だけの slim item を拒否する", () => {
    // Given: 旧 slim list 形（topics のみ）
    const payload = {
      episodes: [
        {
          episodeId: "ep-1",
          date: "2026-08-17",
          title: "題",
          durationSec: 60,
          topics: [{ title: "題1" }],
          audioRef: episodeAudioPath("ep-1"),
        },
      ],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("episode に契約外の field を足す時拒否する", () => {
    // Given: 契約外 field を持つ episode
    const body = { episodes: [{ ...validEpisodeItem, extra: 1 }] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 失敗する（strict object を維持する）
    expect(got.success).toBe(false);
  });

  it("opening が { text, startSec } 形の時受け入れ、startSec を保つ", () => {
    // Given: opening が { text, startSec } 形の episode 1件
    const body = { episodes: [validEpisodeItem] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 成功し、opening を保つ
    expect(got.success).toBe(true);
    if (got.success) {
      expect(got.data.episodes[0]?.body.opening).toEqual({ text: "開始", startSec: 0 });
    }
  });

  it("opening が旧来の文字列の時拒否する", () => {
    // Given: opening が文字列（拡張前の形）の episode
    const payload = {
      episodes: [{ ...validEpisodeItem, body: { ...validEpisodeItem.body, opening: "開始" } }],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("opening に startSec が無い時拒否する", () => {
    // Given: opening.startSec 欠落の episode
    const payload = {
      episodes: [
        { ...validEpisodeItem, body: { ...validEpisodeItem.body, opening: { text: "開始" } } },
      ],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("opening.startSec が負の時拒否する", () => {
    // Given: opening.startSec が負
    const payload = {
      episodes: [
        {
          ...validEpisodeItem,
          body: { ...validEpisodeItem.body, opening: { text: "開始", startSec: -1 } },
        },
      ],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("opening に契約外の field を足す時拒否する", () => {
    // Given: 契約外 field を持つ opening
    const payload = {
      episodes: [
        {
          ...validEpisodeItem,
          body: { ...validEpisodeItem.body, opening: { text: "開始", startSec: 0, extra: 1 } },
        },
      ],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する（strict object を維持する）
    expect(got.success).toBe(false);
  });

  it("ending が startSec を持つ時受け入れる", () => {
    // Given: ending が { text, startSec } 形の episode 1件
    const body = { episodes: [validEpisodeItem] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 成功し、ending.startSec を保つ
    expect(got.success).toBe(true);
    if (got.success) {
      expect(got.data.episodes[0]?.body.ending).toEqual({ text: "終了", startSec: 30 });
    }
  });

  it("ending が旧来の文字列の時拒否する", () => {
    // Given: ending が文字列（拡張前の形）の episode
    const payload = {
      episodes: [{ ...validEpisodeItem, body: { ...validEpisodeItem.body, ending: "終了" } }],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("ending に startSec が無い時拒否する", () => {
    // Given: ending.startSec 欠落の episode
    const payload = {
      episodes: [
        { ...validEpisodeItem, body: { ...validEpisodeItem.body, ending: { text: "終了" } } },
      ],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("ending.startSec が負の時拒否する", () => {
    // Given: ending.startSec が負
    const payload = {
      episodes: [
        {
          ...validEpisodeItem,
          body: { ...validEpisodeItem.body, ending: { text: "終了", startSec: -1 } },
        },
      ],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });

  it("ending に契約外の field を足す時拒否する", () => {
    // Given: 契約外 field を持つ ending
    const payload = {
      episodes: [
        {
          ...validEpisodeItem,
          body: { ...validEpisodeItem.body, ending: { text: "終了", startSec: 30, extra: 1 } },
        },
      ],
    };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(payload);

    // Then: 失敗する（strict object を維持する）
    expect(got.success).toBe(false);
  });

  it("audioRef が無い時拒否する", () => {
    // Given: audioRef 欠落の episode
    const { audioRef: _audioRef, ...withoutAudioRef } = validEpisodeItem;
    const body = { episodes: [withoutAudioRef] };

    // When: parse する
    const got = ListEpisodesResponseSchema.safeParse(body);

    // Then: 失敗する
    expect(got.success).toBe(false);
  });
});
