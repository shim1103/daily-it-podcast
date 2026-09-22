import { describe, expect, it } from "vitest";
import {
  EPISODE_PROGRESS_TABLE,
  PROGRESS_COMPLETE_ZONE_SEC,
  PROGRESS_D1_ADAPTER_MAX_ATTEMPTS,
  episodeProgressColumns,
} from "./progress.ts";

describe("progress constants", () => {
  it("完走ゾーンは 3 秒である", () => {
    expect(PROGRESS_COMPLETE_ZONE_SEC).toBe(3);
  });

  it("D1 adapter は再試行せず 1 試行だけである", () => {
    expect(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS).toBe(1);
  });

  it("表名と列名は migration の識別子と一致する", () => {
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
