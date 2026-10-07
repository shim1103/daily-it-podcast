import { describe, expect, it } from "vitest";
import {
  episodeAudioPath,
  ListEpisodesResponseSchema,
  UnavailableError,
} from "../../../contracts/index.ts";
import type { ListEpisodesUseCaseOutput } from "../application/use-cases/list-episodes.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { createListEpisodesController } from "./list-episodes-controller.ts";

/**
 * scope: Sociable Unit
 * real: createListEpisodesController, map-internal-error
 * double: ListEpisodesUseCase を test 内の Stub に差し替え
 */
const topics = [
  { title: "第一", preface: "前1", detail: "詳1", startSec: 0 },
  { title: "第二", preface: "前2", detail: "詳2", startSec: 30 },
] as const;

const useCaseOutput: ListEpisodesUseCaseOutput = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題1",
      durationSec: 60,
      body: {
        opening: { text: "開始", startSec: 0 },
        topics,
        ending: { text: "終了", startSec: 55 },
      },
      progress: {
        positionSec: 12,
        firstPlayedAt: "2026-09-22T10:00:00.000Z",
        firstCompletedAt: "2026-09-22T10:30:00.000Z",
        lastPlayedAt: "2026-09-22T11:00:00.000Z",
      },
    },
    {
      episodeId: "ep-2",
      date: "2026-08-10",
      title: "題2",
      durationSec: 90,
      body: {
        opening: { text: "開始2", startSec: 1 },
        topics: [{ title: "唯一", preface: "前", detail: "詳", startSec: 2 }],
        ending: { text: "終了2", startSec: 80 },
      },
      progress: null,
    },
  ],
};

describe("createListEpisodesController", () => {
  it("UseCase が成功する時、ListEpisodesResponse schema を満たす", async () => {
    // Given: 進捗あり・なしの 2 件を返す Stub UseCase
    const controller = createListEpisodesController(async () => useCaseOutput);

    // When: 一覧を取得する
    const got = await controller();

    // Then: 契約 schema を満たす
    expect(ListEpisodesResponseSchema.safeParse(got).success).toBe(true);
  });

  it("各 episode の audioRef は episodeAudioPath(episodeId) で足される", async () => {
    // Given: audioRef を持たない UseCaseOutput
    const controller = createListEpisodesController(async () => useCaseOutput);

    // When: 一覧を取得する
    const got = await controller();

    // Then: episode ごとの音声 path が付く
    expect(got.episodes.map((episode) => episode.audioRef)).toEqual([
      episodeAudioPath("ep-1"),
      episodeAudioPath("ep-2"),
    ]);
  });

  it("原稿の field を 1 つずつ Response へ写す", async () => {
    // Given: 原稿 field を持つ UseCaseOutput
    const controller = createListEpisodesController(async () => useCaseOutput);

    // When: 一覧を取得する
    const got = await controller();

    // Then: 先頭 episode の field が一致し、順序も保たれる
    const first = got.episodes[0];
    expect(got.episodes.map((episode) => episode.episodeId)).toEqual(["ep-1", "ep-2"]);
    expect(first?.date).toBe("2026-08-17");
    expect(first?.title).toBe("題1");
    expect(first?.durationSec).toBe(60);
    expect(first?.body.opening).toEqual({ text: "開始", startSec: 0 });
    expect(first?.body.topics).toEqual(topics);
    expect(first?.body.ending).toEqual({ text: "終了", startSec: 55 });
  });

  it("progress が値の時、4 field を 1 つずつ写す", async () => {
    // Given: 進捗ありの episode を含む UseCaseOutput
    const controller = createListEpisodesController(async () => useCaseOutput);

    // When: 一覧を取得する
    const got = await controller();

    // Then: 進捗の各 field が一致する
    expect(got.episodes[0]?.progress).toEqual({
      positionSec: 12,
      firstPlayedAt: "2026-09-22T10:00:00.000Z",
      firstCompletedAt: "2026-09-22T10:30:00.000Z",
      lastPlayedAt: "2026-09-22T11:00:00.000Z",
    });
  });

  it("progress が null の時、null のまま写す", async () => {
    // Given: 進捗なしの episode を含む UseCaseOutput
    const controller = createListEpisodesController(async () => useCaseOutput);

    // When: 一覧を取得する
    const got = await controller();

    // Then: undefined でなく null
    expect(got.episodes[1]?.progress).toBeNull();
  });

  it("body の topics は UseCaseOutput と別の配列としてコピーされる", async () => {
    // Given: 配列参照を持つ UseCaseOutput
    const controller = createListEpisodesController(async () => useCaseOutput);

    // When: 一覧を取得する
    const got = await controller();

    // Then: 同じ中身で、参照は共有しない
    const sourceTopics = useCaseOutput.episodes[0]?.body.topics;
    const gotTopics = got.episodes[0]?.body.topics;
    expect(gotTopics).toEqual(sourceTopics);
    expect(gotTopics).not.toBe(sourceTopics);
    expect(Array.isArray(gotTopics)).toBe(true);
  });

  it("UseCase が空の一覧を返す時、空の episodes を返す", async () => {
    // Given: 該当なしを返す Stub UseCase
    const controller = createListEpisodesController(async () => ({ episodes: [] }));

    // When: 一覧を取得する
    const got = await controller();

    // Then: 空配列
    expect(got).toEqual({ episodes: [] });
  });

  it("UseCase が R2Error を throw する時、UnavailableError に cause 付きで変換する", async () => {
    // Given: Infrastructure 失敗を throw する Stub UseCase
    const r2Error = new R2Error("R2 読取に失敗");
    const controller = createListEpisodesController(async () => {
      throw r2Error;
    });

    // When: 一覧を取得する
    const act = controller();

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === r2Error,
    );
  });
});
