import { describe, expect, it } from "vitest";
import {
  EPISODE_PROGRESS_D1_BINDING,
  EPISODE_PROGRESS_TABLE,
  PROGRESS_D1_ADAPTER_MAX_ATTEMPTS,
  episodeProgressColumns,
} from "./progress-d1-constants.ts";

describe("progress D1 constants", () => {
  it("D1 adapter は再試行せず 1 試行だけである", () => {
    expect(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS).toBe(1);
  });

  it("D1 binding 名は EPISODE_PROGRESS である", () => {
    expect(EPISODE_PROGRESS_D1_BINDING).toBe("EPISODE_PROGRESS");
  });

  it("表・列名が migration と一致する契約値である", () => {
    expect(EPISODE_PROGRESS_TABLE).toBe("episode_progress");
    expect(episodeProgressColumns).toEqual({
      episodeId: "episode_id",
      positionSec: "position_sec",
      firstPlayedAt: "first_played_at",
      firstCompletedAt: "first_completed_at",
      lastPlayedAt: "last_played_at",
    });
  });
});
