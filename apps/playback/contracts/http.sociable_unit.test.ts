import { describe, expect, it } from "vitest";
import {
  episodeAudioPath,
  episodePath,
  episodeProgressCompletePath,
  episodeProgressPath,
  episodeProgressSchema,
  ListEpisodesResponseSchema,
  listEpisodesPath,
  ProgressPullQuerySchema,
  ProgressPullResponseSchema,
  ProgressWriteRequestSchema,
  ProgressWriteResponseSchema,
  progressPullPath,
  PROGRESS_WRITE_MAX_ATTEMPTS,
  progressWriteRetryableHttpErrorCodes,
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
  progress: null,
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

  it("progress が無い時拒否する", () => {
    const { progress: _progress, ...withoutProgress } = validEpisodeItem;
    const body = { episodes: [withoutProgress] };

    const got = ListEpisodesResponseSchema.safeParse(body);

    expect(got.success).toBe(false);
  });

  it("progress に適合 object を載せる時受理する", () => {
    const body = {
      episodes: [
        {
          ...validEpisodeItem,
          progress: {
            positionSec: 12,
            firstPlayedAt: "2026-09-19T10:00:00.000Z",
            firstCompletedAt: null,
            lastPlayedAt: "2026-09-19T10:05:00.000Z",
          },
        },
      ],
    };

    const got = ListEpisodesResponseSchema.safeParse(body);

    expect(got.success).toBe(true);
  });
});

describe("episodeProgressSchema", () => {
  it("offset 付き ISO と null の firstCompletedAt を受理する", () => {
    const got = episodeProgressSchema.safeParse({
      positionSec: 0,
      firstPlayedAt: "2026-09-19T19:00:00.000+09:00",
      firstCompletedAt: null,
      lastPlayedAt: "2026-09-19T19:00:00.000+09:00",
    });

    expect(got.success).toBe(true);
  });

  it("firstCompletedAt 欠落は拒否する（null 明示が必要）", () => {
    const got = episodeProgressSchema.safeParse({
      positionSec: 0,
      firstPlayedAt: "2026-09-19T10:00:00.000Z",
      lastPlayedAt: "2026-09-19T10:00:00.000Z",
    });

    expect(got.success).toBe(false);
  });
});

describe("episodeProgressPath", () => {
  it("episode path の後に progress 段が 1 つ続く", () => {
    expect(episodeProgressPath("ep-1")).toBe(`${episodePath("ep-1")}/progress`);
  });
});

describe("episodeProgressCompletePath", () => {
  it("progress path の後に complete 段が 1 つ続く", () => {
    expect(episodeProgressCompletePath("ep-1")).toBe(`${episodeProgressPath("ep-1")}/complete`);
  });
});

describe("progressPullPath", () => {
  it("一覧 path とは別の進捗 pull 用 path である", () => {
    expect(progressPullPath).toBe("/progress");
    expect(progressPullPath).not.toBe(listEpisodesPath);
  });
});

describe("ProgressWriteRequestSchema", () => {
  it("positionSec と offset 付き clientAt を受理する", () => {
    const got = ProgressWriteRequestSchema.safeParse({
      positionSec: 12,
      clientAt: "2026-09-19T19:00:00.000+09:00",
    });

    expect(got.success).toBe(true);
  });

  it("clientAt 欠落は拒否する", () => {
    const got = ProgressWriteRequestSchema.safeParse({ positionSec: 12 });

    expect(got.success).toBe(false);
  });

  it("契約外 field を拒否する", () => {
    const got = ProgressWriteRequestSchema.safeParse({
      positionSec: 12,
      clientAt: "2026-09-19T10:00:00.000Z",
      extra: 1,
    });

    expect(got.success).toBe(false);
  });
});

describe("ProgressWriteResponseSchema", () => {
  it("勝ち側 first* だけを受理し positionSec は載せない", () => {
    const got = ProgressWriteResponseSchema.safeParse({
      firstPlayedAt: "2026-09-19T10:00:00.000Z",
      firstCompletedAt: null,
    });

    expect(got.success).toBe(true);
  });

  it("positionSec 付きは拒否する", () => {
    const got = ProgressWriteResponseSchema.safeParse({
      firstPlayedAt: "2026-09-19T10:00:00.000Z",
      firstCompletedAt: null,
      positionSec: 12,
    });

    expect(got.success).toBe(false);
  });
});

describe("ProgressPullQuerySchema", () => {
  it("offset 付き since を受理する", () => {
    const got = ProgressPullQuerySchema.safeParse({
      since: "2026-09-19T19:00:00.000+09:00",
    });

    expect(got.success).toBe(true);
  });

  it("since 欠落は拒否する", () => {
    const got = ProgressPullQuerySchema.safeParse({});

    expect(got.success).toBe(false);
  });
});

describe("ProgressPullResponseSchema", () => {
  it("空配列を受理する", () => {
    const got = ProgressPullResponseSchema.safeParse({ episodes: [] });

    expect(got.success).toBe(true);
  });

  it("episodeId と progress object の組を受理する", () => {
    const got = ProgressPullResponseSchema.safeParse({
      episodes: [
        {
          episodeId: "ep-1",
          progress: {
            positionSec: 12,
            firstPlayedAt: "2026-09-19T10:00:00.000Z",
            firstCompletedAt: null,
            lastPlayedAt: "2026-09-19T10:05:00.000Z",
          },
        },
      ],
    });

    expect(got.success).toBe(true);
  });

  it("progress null は拒否する（pull は更新行だけ）", () => {
    const got = ProgressPullResponseSchema.safeParse({
      episodes: [{ episodeId: "ep-1", progress: null }],
    });

    expect(got.success).toBe(false);
  });
});

describe("progress Write retry 契約", () => {
  it("最大試行は正の有限回数である", () => {
    expect(PROGRESS_WRITE_MAX_ATTEMPTS).toBe(3);
    expect(PROGRESS_WRITE_MAX_ATTEMPTS).toBeGreaterThan(0);
  });

  it("再試行対象の契約 code は unavailable のみである", () => {
    expect([...progressWriteRetryableHttpErrorCodes]).toEqual(["unavailable"]);
  });
});
