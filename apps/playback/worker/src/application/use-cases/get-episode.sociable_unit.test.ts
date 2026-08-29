import { describe, expect, it } from "vitest";
import { GetEpisodeResponseSchema, episodeAudioPath } from "../../../../contracts/index.ts";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import { getEpisode } from "./get-episode.ts";

/**
 * scope: Sociable Unit
 * real: getEpisode use-case, verify-manuscript の純関数
 * double: EpisodeRepository を「生 payload を返す」Fake Port に差し替え
 *
 * why: Port は取得したままの json を返すだけ。schema 適合・stem 一致・wav 欠落の判定は
 * この use-case が `verifyManuscript` を使って行う。Port 生 payload → 検証 → Error/成功 の
 * 結線をここで固定する。純関数単体の入出力契約は `verify-manuscript.sociable_unit.test.ts`。
 */
const validManuscriptJson = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [{ title: "題", preface: "前置き", detail: "詳細", startSec: 0 }],
    closing: "終了",
  },
};

function createFakeRepository(overrides: Partial<EpisodeRepository> = {}): EpisodeRepository {
  return {
    listManuscripts: async () => {
      throw new Error("not used");
    },
    getManuscript: async () => ({ json: validManuscriptJson, hasAudio: true }),
    getEpisodeAudio: async () => {
      throw new Error("not used");
    },
    ...overrides,
  };
}

describe("getEpisode", () => {
  it("Port が返した生 json を検証し、audioRef を付けて GetEpisodeResponse を返す", async () => {
    // Given: 適合 json + wav ありを返す Fake Port
    const repository = createFakeRepository();

    // When: 1件取得 UseCase を実行する
    const got = await getEpisode(repository, "ep-1");

    // Then: 検証済み + audioRef
    expect(GetEpisodeResponseSchema.safeParse(got).success).toBe(true);
    expect(got.audioRef).toBe(episodeAudioPath("ep-1"));
    expect(got.body).toEqual(validManuscriptJson.body);
  });

  it("Port が undefined（json エントリ無し）を返す時、EpisodeContentError（JSON エントリが無い）", async () => {
    // Given: 取得対象が存在しない
    const repository = createFakeRepository({ getManuscript: async () => undefined });

    // When: 1件取得する
    const act = getEpisode(repository, "missing");

    // Then: Domain の実体不備、message は JSON エントリ欠落
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/JSON エントリが無い/);
  });

  it("hasAudio が false の時、EpisodeContentError（wav が無い）", async () => {
    // Given: json はあるが wav 無し
    const repository = createFakeRepository({
      getManuscript: async () => ({ json: validManuscriptJson, hasAudio: false }),
    });

    // When: 1件取得する
    const act = getEpisode(repository, "ep-1");

    // Then: message は wav 欠落
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/wav が無い/);
  });

  it("schema 不適合の json の時、EpisodeContentError（schema に不適合）", async () => {
    // Given: wav はあるが json が schema 不適合
    const repository = createFakeRepository({
      getManuscript: async () => ({ json: { episodeId: "ep-1" }, hasAudio: true }),
    });

    // When: 1件取得する
    const act = getEpisode(repository, "ep-1");

    // Then: message は schema 不適合
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/schema に不適合/);
  });

  it("stem 不一致の json の時、EpisodeContentError（stem と JSON の episodeId が不一致）", async () => {
    // Given: wav はあるが json 内 episodeId が要求と不一致
    const repository = createFakeRepository({
      getManuscript: async () => ({
        json: { ...validManuscriptJson, episodeId: "ep-other" },
        hasAudio: true,
      }),
    });

    // When: 1件取得する
    const act = getEpisode(repository, "ep-1");

    // Then: message は stem 不一致
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/stem と JSON の episodeId が不一致/);
  });

  it("4 失敗ケースは同一 Error 型で message だけが異なる", async () => {
    // Given: json エントリ無し / wav 無し / schema 不適合 / stem 不一致
    const cases: Array<{ over: Partial<EpisodeRepository>; expected: RegExp }> = [
      { over: { getManuscript: async () => undefined }, expected: /JSON エントリが無い/ },
      {
        over: {
          getManuscript: async () => ({ json: validManuscriptJson, hasAudio: false }),
        },
        expected: /wav が無い/,
      },
      {
        over: {
          getManuscript: async () => ({ json: { episodeId: "ep-1" }, hasAudio: true }),
        },
        expected: /schema に不適合/,
      },
      {
        over: {
          getManuscript: async () => ({
            json: { ...validManuscriptJson, episodeId: "ep-other" },
            hasAudio: true,
          }),
        },
        expected: /stem と JSON の episodeId が不一致/,
      },
    ];

    // When / Then: 型は同一、message は個別
    const messages = new Set<string>();
    for (const { over, expected } of cases) {
      const repository = createFakeRepository(over);
      const error = await getEpisode(repository, "ep-1").catch((caught: unknown) => caught);
      expect(error).toBeInstanceOf(EpisodeContentError);
      expect((error as Error).message).toMatch(expected);
      messages.add((error as Error).message);
    }
    expect(messages.size).toBe(4);
  });

  it("複合失敗（wav 欠落 かつ schema 不適合）は precedence どおり wav 欠落を先に報告する", async () => {
    // Given: schema 不適合 json かつ wav 無し
    const repository = createFakeRepository({
      getManuscript: async () => ({ json: { episodeId: "ep-1" }, hasAudio: false }),
    });

    // When: 1件取得する
    const act = getEpisode(repository, "ep-1");

    // Then: hasAudio → schema → stem の順。wav 欠落が先
    await expect(act).rejects.toThrow(/wav が無い/);
  });
});
