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

  it("normalizes_all_time_fields_to_utc_fixed_width_when_progress_has_offsets", () => {
    // Given: 3 つの時刻 field がすべて offset 付き・小数部なしの進捗
    const progress = {
      positionSec: 12,
      firstPlayedAt: "2026-10-01T09:00:00+09:00",
      firstCompletedAt: "2026-10-01T09:05:00+09:00",
      lastPlayedAt: "2026-10-01T09:10:00+09:00",
    };

    // When: 進捗 schema で parse する
    const got = episodeProgressSchema.safeParse(progress);

    // Then: どの field も UTC 固定幅へ正規化される
    expect(got.success).toBe(true);
    expect(got.data).toEqual({
      positionSec: 12,
      firstPlayedAt: "2026-10-01T00:00:00.000Z",
      firstCompletedAt: "2026-10-01T00:05:00.000Z",
      lastPlayedAt: "2026-10-01T00:10:00.000Z",
    });
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

describe("ProgressWriteRequestSchema の clientAt（進捗の時刻契約）", () => {
  function parseClientAt(clientAt: string) {
    return ProgressWriteRequestSchema.safeParse({ positionSec: 12, clientAt });
  }

  it("normalizes_to_utc_fixed_width_when_clientAt_has_offset", () => {
    // Given: +09:00 の offset 付き clientAt
    const clientAt = "2026-10-01T09:00:00+09:00";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: 同じ instant の UTC 固定幅へ正規化される
    expect(got.data?.clientAt).toBe("2026-10-01T00:00:00.000Z");
  });

  it("pads_to_fixed_width_when_clientAt_omits_seconds", () => {
    // Given: 秒なしの clientAt
    const clientAt = "2026-10-01T00:00Z";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: 秒・ミリ秒が 0 で補われる
    expect(got.data?.clientAt).toBe("2026-10-01T00:00:00.000Z");
  });

  it("pads_to_milliseconds_when_clientAt_has_one_fraction_digit", () => {
    // Given: 小数部 1 桁の clientAt
    const clientAt = "2026-10-01T00:00:00.5Z";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: 3 桁へ補われる
    expect(got.data?.clientAt).toBe("2026-10-01T00:00:00.500Z");
  });

  it("truncates_to_milliseconds_when_clientAt_has_six_fraction_digits", () => {
    // Given: 小数部 6 桁で、丸めると桁上がりする clientAt
    const clientAt = "2026-10-01T00:00:00.999999Z";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: 3 桁へ切り捨てられる（丸めない）
    expect(got.data?.clientAt).toBe("2026-10-01T00:00:00.999Z");
  });

  it("accepts_when_clientAt_is_one_millisecond_after_lower_bound", () => {
    // Given: 下限（1970-01-01T00:00:00Z）の 1ms 後
    const clientAt = "1970-01-01T00:00:00.001Z";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: 受理される
    expect(got.data?.clientAt).toBe("1970-01-01T00:00:00.001Z");
  });

  it("rejects_when_clientAt_equals_lower_bound", () => {
    // Given: 下限と同じ instant（下限は含まない）
    const clientAt = "1970-01-01T00:00:00Z";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: 拒否される
    expect(got.success).toBe(false);
  });

  it("rejects_when_clientAt_is_before_lower_bound_after_offset_conversion", () => {
    // Given: 文字列は 1970 年だが、offset 変換後は 1969-12-31T23:59:59Z になる
    const clientAt = "1970-01-01T08:59:59+09:00";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: instant で判定され拒否される
    expect(got.success).toBe(false);
  });

  it("accepts_when_clientAt_is_one_millisecond_before_upper_bound", () => {
    // Given: 上限（2100-01-01T00:00:00Z）の 1ms 前
    const clientAt = "2099-12-31T23:59:59.999Z";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: 受理される
    expect(got.data?.clientAt).toBe("2099-12-31T23:59:59.999Z");
  });

  it("accepts_when_clientAt_looks_past_upper_bound_but_is_before_it_after_offset_conversion", () => {
    // Given: 文字列は 2100 年だが、offset 変換後は 2099-12-31T23:59:59Z になる
    const clientAt = "2100-01-01T08:59:59+09:00";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: instant で判定され受理される
    expect(got.data?.clientAt).toBe("2099-12-31T23:59:59.000Z");
  });

  it("rejects_when_clientAt_equals_upper_bound", () => {
    // Given: 上限と同じ instant（上限は含まない）
    const clientAt = "2100-01-01T00:00:00Z";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: 拒否される
    expect(got.success).toBe(false);
  });

  it("rejects_when_clientAt_is_past_upper_bound_after_offset_conversion", () => {
    // Given: 文字列は 2099 年だが、offset 変換後は 2100-01-01T00:59:59Z になる
    const clientAt = "2099-12-31T23:59:59-01:00";

    // When: parse する
    const got = parseClientAt(clientAt);

    // Then: instant で判定され拒否される
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

  it("normalizes_first_timestamps_to_utc_fixed_width_and_keeps_fixed_width_values", () => {
    // Given: offset 付きの firstPlayedAt と、すでに固定幅 UTC の firstCompletedAt
    const response = {
      firstPlayedAt: "2026-10-01T09:00:00+09:00",
      firstCompletedAt: "2026-10-01T00:05:00.000Z",
    };

    // When: 応答 schema で parse する
    const got = ProgressWriteResponseSchema.safeParse(response);

    // Then: 前者は正規化され、後者は同じ値のまま通る（冪等）
    expect(got.data).toEqual({
      firstPlayedAt: "2026-10-01T00:00:00.000Z",
      firstCompletedAt: "2026-10-01T00:05:00.000Z",
    });
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

  it("normalizes_since_to_utc_fixed_width_when_since_has_offset", () => {
    // Given: +09:00 の offset 付き since
    const query = { since: "2026-10-01T09:00:00+09:00" };

    // When: parse する
    const got = ProgressPullQuerySchema.safeParse(query);

    // Then: 同じ instant の UTC 固定幅へ正規化される
    expect(got.data?.since).toBe("2026-10-01T00:00:00.000Z");
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
