import { describe, expect, it } from "vitest";
import { episodeAudioPath, ListEpisodesResponseSchema } from "../../../../contracts/index.ts";
import { StubProgressRepository } from "../ports/progress-repository.ts";
import type { EpisodeRepository, RawManuscriptEntry } from "../ports/episode-repository.ts";
import { listEpisodes } from "./list-episodes.ts";

/**
 * scope: Sociable Unit
 * real: listEpisodes use-case, verify-manuscript の純関数
 * double: EpisodeRepository を「生 payload の配列を返す」Fake Port に差し替え
 */
const validManuscriptJson = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: { text: "開始", startSec: 0 },
    topics: [
      { title: "第一", preface: "前1", detail: "詳1", startSec: 0 },
      { title: "第二", preface: "前2", detail: "詳2", startSec: 30 },
    ],
    ending: { text: "終了", startSec: 55 },
  },
};

function createFakeRepository(entries: RawManuscriptEntry[]): EpisodeRepository {
  return {
    listManuscripts: async () => entries,
    getAudio: async () => {
      throw new Error("not used");
    },
  };
}

describe("listEpisodes", () => {
  it("Port が返した生 json 配列を検証し、原稿全文付き ListEpisodesResponse に包む", async () => {
    // Given: 適合 json 1件を返す Fake Port
    const repository = createFakeRepository([{ stem: "ep-1", json: validManuscriptJson }]);

    // When: 一覧 UseCase を実行する
    const got = await listEpisodes(repository, new StubProgressRepository());

    // Then: 契約 schema を満たし、body 全文と audioRef がある
    expect(ListEpisodesResponseSchema.safeParse(got).success).toBe(true);
    expect(got.episodes).toEqual([
      {
        episodeId: "ep-1",
        date: "2026-08-17",
        title: "題",
        durationSec: 60,
        body: validManuscriptJson.body,
        audioRef: episodeAudioPath("ep-1"),
        progress: null,
      },
    ]);
  });

  it("schema 不適合の entry は throw せず一覧から除外する", async () => {
    // Given: 不適合 json だけ
    const repository = createFakeRepository([{ stem: "bad", json: { episodeId: "bad" } }]);

    // When: 一覧を取得する
    const got = await listEpisodes(repository, new StubProgressRepository());

    // Then: 除外され空一覧
    expect(got.episodes).toEqual([]);
  });

  it("stem と json 内 episodeId が不一致の entry は throw せず除外する", async () => {
    // Given: stem と episodeId がズレた適合 json
    const repository = createFakeRepository([
      { stem: "ep-1", json: { ...validManuscriptJson, episodeId: "ep-other" } },
    ]);

    // When: 一覧を取得する
    const got = await listEpisodes(repository, new StubProgressRepository());

    // Then: 行に出ない
    expect(got.episodes).toEqual([]);
  });

  it("不正 JSON 由来の非 object entry は throw せず除外する", async () => {
    // Given: decode 失敗で string が入った entry
    const repository = createFakeRepository([{ stem: "ep-1", json: "not json" }]);

    // When: 一覧を取得する
    const got = await listEpisodes(repository, new StubProgressRepository());

    // Then: 除外される
    expect(got.episodes).toEqual([]);
  });

  it("適合分と不適合分が混在する時、適合分だけの部分一覧を返す", async () => {
    // Given: 適合 1 + 不適合 1
    const repository = createFakeRepository([
      { stem: "ep-1", json: validManuscriptJson },
      { stem: "bad", json: { episodeId: "bad" } },
    ]);

    // When: 一覧を取得する
    const got = await listEpisodes(repository, new StubProgressRepository());

    // Then: 適合分だけ
    expect(got.episodes.map((item) => item.episodeId)).toEqual(["ep-1"]);
  });

  it("Port が空配列（該当なし）を返す時、空一覧を返す", async () => {
    // Given: 該当なし
    const repository = createFakeRepository([]);

    // When: 一覧を取得する
    const got = await listEpisodes(repository, new StubProgressRepository());

    // Then: 空
    expect(got.episodes).toEqual([]);
  });

  it("Port が返す順序に依らず、date降順（新しい日付が先頭）で返す", async () => {
    // Given: Port が date の昇順で返す（R2 List の key 辞書順を想定した非日付順）
    const older = { ...validManuscriptJson, episodeId: "ep-old", date: "2026-08-01" };
    const newer = { ...validManuscriptJson, episodeId: "ep-new", date: "2026-09-10" };
    const middle = { ...validManuscriptJson, episodeId: "ep-mid", date: "2026-08-20" };
    const repository = createFakeRepository([
      { stem: "ep-old", json: older },
      { stem: "ep-new", json: newer },
      { stem: "ep-mid", json: middle },
    ]);

    // When: 一覧を取得する
    const got = await listEpisodes(repository, new StubProgressRepository());

    // Then: date降順に並び替わる
    expect(got.episodes.map((item) => item.episodeId)).toEqual(["ep-new", "ep-mid", "ep-old"]);
  });

  it("date が同値の entry は例外を投げず両方とも一覧に残す", async () => {
    // Given: 同一 date の entry 2件
    const first = { ...validManuscriptJson, episodeId: "ep-a", date: "2026-08-17" };
    const second = { ...validManuscriptJson, episodeId: "ep-b", date: "2026-08-17" };
    const repository = createFakeRepository([
      { stem: "ep-a", json: first },
      { stem: "ep-b", json: second },
    ]);

    // When: 一覧を取得する
    const got = await listEpisodes(repository, new StubProgressRepository());

    // Then: 両方とも残る（順序は未規定）
    expect(got.episodes.map((item) => item.episodeId).sort()).toEqual(["ep-a", "ep-b"]);
  });
});
