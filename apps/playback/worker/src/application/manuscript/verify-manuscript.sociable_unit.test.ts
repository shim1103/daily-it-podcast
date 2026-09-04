import { describe, expect, it } from "vitest";
import { episodeAudioPath } from "../../../../contracts/index.ts";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import { selectValidListItem, verifyManuscript } from "./verify-manuscript.ts";

const validManuscript = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [
      { title: "第一", preface: "前1", detail: "詳1", startSec: 0 },
      { title: "第二", preface: "前2", detail: "詳2", startSec: 30 },
    ],
    ending: "終了",
  },
};

describe("verifyManuscript", () => {
  it("schema 適合かつ stem 一致の時、検証済み原稿を返す", () => {
    // Given: 適合 json と一致する stem
    // When: 検証する
    const got = verifyManuscript(validManuscript, "ep-1");

    // Then: 入力と同形の検証済み値
    expect(got.episodeId).toBe("ep-1");
    expect(got.body.topics.map((topic) => topic.title)).toEqual(["第一", "第二"]);
  });

  it("schema 不適合の時、EpisodeContentError を throw し message は schema 不適合を示す", () => {
    // Given: 必須 field 欠落の json
    // When / Then
    expect(() => verifyManuscript({ episodeId: "ep-1" }, "ep-1")).toThrow(EpisodeContentError);
    expect(() => verifyManuscript({ episodeId: "ep-1" }, "ep-1")).toThrow(/schema に不適合/);
  });

  it("不正 JSON 由来の非 object の時、EpisodeContentError を throw し message は schema 不適合を示す", () => {
    // Given: JSON.parse に失敗した生文字列相当
    // When / Then: 判定は schema 不適合へ畳む（不正 JSON 専用種別は作らない）
    expect(() => verifyManuscript("not json", "ep-1")).toThrow(EpisodeContentError);
    expect(() => verifyManuscript("not json", "ep-1")).toThrow(/schema に不適合/);
  });

  it("schema 適合だが stem 不一致の時、EpisodeContentError を throw し message は stem 不一致を示す", () => {
    // Given: 中身の episodeId が stem と別物
    // When / Then
    expect(() => verifyManuscript({ ...validManuscript, episodeId: "ep-other" }, "ep-1")).toThrow(
      /stem と JSON の episodeId が不一致/,
    );
  });

  it("schema 不適合と stem 不一致が同時に起きうる時、schema 不適合を優先して報告する", () => {
    // Given: episodeId は stem と不一致、かつ body 欠落で schema 不適合
    const bothBroken = { episodeId: "ep-other", date: "2026-08-17" };

    // When: 検証する
    const error = (() => {
      try {
        verifyManuscript(bothBroken, "ep-1");
        return undefined;
      } catch (caught) {
        return caught as Error;
      }
    })();

    // Then: precedence どおり schema 不適合が出る（stem 判定は schema 通過後にのみ行う）
    expect(error).toBeInstanceOf(EpisodeContentError);
    expect(error?.message).toMatch(/schema に不適合/);
  });

  it("4 分類のうち schema / stem の 2 分類は同一 Error 型で message だけが異なる", () => {
    // Given: schema 不適合 / stem 不一致
    const messages = new Set<string>();
    for (const [json, stem] of [
      [{ episodeId: "ep-1" }, "ep-1"],
      [{ ...validManuscript, episodeId: "ep-other" }, "ep-1"],
    ] as const) {
      try {
        verifyManuscript(json, stem);
      } catch (caught) {
        expect(caught).toBeInstanceOf(EpisodeContentError);
        messages.add((caught as Error).message);
      }
    }
    expect(messages.size).toBe(2);
  });
});

describe("selectValidListItem", () => {
  it("schema 適合かつ stem 一致の時、原稿全文付き EpisodeListItem を返す", () => {
    // Given: 適合 json
    // When: 一覧用に選ぶ
    const got = selectValidListItem(validManuscript, "ep-1");

    // Then: body 全文と audioRef がある
    expect(got).toEqual({
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題",
      durationSec: 60,
      body: validManuscript.body,
      audioRef: episodeAudioPath("ep-1"),
    });
  });

  it("schema 不適合の時、throw せず undefined を返す（listEpisodes の除外用）", () => {
    // Given: 不適合 json
    // When / Then
    expect(selectValidListItem({ episodeId: "bad" }, "bad")).toBeUndefined();
  });

  it("不正 JSON 由来の non object の時、throw せず undefined を返す", () => {
    expect(selectValidListItem("not json", "ep-1")).toBeUndefined();
  });

  it("stem 不一致の時、throw せず undefined を返す", () => {
    // Given: episodeId が stem と別物
    // When / Then
    expect(
      selectValidListItem({ ...validManuscript, episodeId: "ep-other" }, "ep-1"),
    ).toBeUndefined();
  });
});
