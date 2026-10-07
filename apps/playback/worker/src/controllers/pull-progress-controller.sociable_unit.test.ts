import { describe, expect, it } from "vitest";
import { UnavailableError } from "../../../contracts/index.ts";
import type {
  PullProgressUseCaseInput,
  PullProgressUseCaseOutput,
} from "../application/use-cases/pull-progress.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { createPullProgressController } from "./pull-progress-controller.ts";

const SINCE = "2026-09-22T10:00:00.000Z";

const output: PullProgressUseCaseOutput = {
  episodes: [
    {
      episodeId: "ep-completed",
      progress: {
        positionSec: 58.5,
        firstPlayedAt: "2026-09-20T09:00:00.000Z",
        firstCompletedAt: "2026-09-20T09:30:00.000Z",
        lastPlayedAt: "2026-09-22T11:00:00.000Z",
      },
    },
    {
      episodeId: "ep-playing",
      progress: {
        positionSec: 7,
        firstPlayedAt: "2026-09-21T09:00:00.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-09-22T12:00:00.000Z",
      },
    },
  ],
};

describe("createPullProgressController", () => {
  it("passes_since_to_use_case_input_when_called", async () => {
    // Given: 受け取った Input を記録する use case の stub
    const received: PullProgressUseCaseInput[] = [];
    const controller = createPullProgressController(async (input) => {
      received.push(input);
      return { episodes: [] };
    });

    // When: 検証済みの query を渡す
    await controller({ since: SINCE });

    // Then: use case が since だけを持つ Input を 1 回受け取る
    expect(received).toStrictEqual([{ since: SINCE }]);
  });

  it("returns_every_entry_field_in_response_when_use_case_output_has_entries", async () => {
    // Given: 完走済みと未完走の 2 件を返す use case の stub
    const controller = createPullProgressController(async () => output);

    // When: 呼ぶ
    const got = await controller({ since: SINCE });

    // Then: 各 entry の episodeId と progress の 4 field が、順序を保って Response へ写る
    expect(got).toStrictEqual({
      episodes: [
        {
          episodeId: "ep-completed",
          progress: {
            positionSec: 58.5,
            firstPlayedAt: "2026-09-20T09:00:00.000Z",
            firstCompletedAt: "2026-09-20T09:30:00.000Z",
            lastPlayedAt: "2026-09-22T11:00:00.000Z",
          },
        },
        {
          episodeId: "ep-playing",
          progress: {
            positionSec: 7,
            firstPlayedAt: "2026-09-21T09:00:00.000Z",
            firstCompletedAt: null,
            lastPlayedAt: "2026-09-22T12:00:00.000Z",
          },
        },
      ],
    });
  });

  it("returns_empty_episodes_when_use_case_output_has_no_entries", async () => {
    // Given: 該当なしを返す use case の stub
    const controller = createPullProgressController(async () => ({ episodes: [] }));

    // When: 呼ぶ
    const got = await controller({ since: SINCE });

    // Then: 空配列
    expect(got).toStrictEqual({ episodes: [] });
  });

  it("returns_array_and_entries_copied_from_use_case_output_when_called", async () => {
    // Given: 同じ Output を返す use case の stub
    const controller = createPullProgressController(async () => output);

    // When: 呼ぶ
    const got = await controller({ since: SINCE });

    // Then: Response の配列も entry も、Output と別の参照（readonly の Output を mutable な Response へ写す）
    expect(got.episodes).not.toBe(output.episodes);
    expect(got.episodes[0]).not.toBe(output.episodes[0]);
    expect(got.episodes[0]?.progress).not.toBe(output.episodes[0]?.progress);
  });

  it("omits_fields_outside_response_contract_when_use_case_output_entry_has_extra_fields", async () => {
    // Given: 契約に無い field（seq）を余分に持つ entry を返す use case の stub
    const entryWithExtra = {
      episodeId: "ep-1",
      progress: {
        positionSec: 1,
        firstPlayedAt: "2026-09-21T09:00:00.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-09-22T12:00:00.000Z",
        seq: 5,
      },
      seq: 5,
    };
    const controller = createPullProgressController(async () => ({ episodes: [entryWithExtra] }));

    // When: 呼ぶ
    const got = await controller({ since: SINCE });

    // Then: entry と progress は契約の field だけを持つ
    expect(got).toStrictEqual({
      episodes: [
        {
          episodeId: "ep-1",
          progress: {
            positionSec: 1,
            firstPlayedAt: "2026-09-21T09:00:00.000Z",
            firstCompletedAt: null,
            lastPlayedAt: "2026-09-22T12:00:00.000Z",
          },
        },
      ],
    });
  });

  it("throws_unavailable_error_with_cause_when_use_case_throws_infrastructure_error", async () => {
    // Given: Infrastructure 失敗を throw する use case の stub
    const infraError = new R2Error("storage 読取に失敗");
    const controller = createPullProgressController(async () => {
      throw infraError;
    });

    // When: 呼ぶ
    const act = controller({ since: SINCE });

    // Then: External UnavailableError が Infrastructure を cause に持つ
    await expect(act).rejects.toSatisfy(
      (error: unknown) => error instanceof UnavailableError && error.cause === infraError,
    );
  });
});
