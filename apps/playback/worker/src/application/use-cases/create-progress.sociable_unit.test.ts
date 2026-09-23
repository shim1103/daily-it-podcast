import { describe, expect, it } from "vitest";
import { ProgressWriteResponseSchema } from "../../../../contracts/index.ts";
import { StubProgressRepository, type ProgressWriteCommand } from "../ports/progress-repository.ts";
import { createProgress } from "./create-progress.ts";

/**
 * scope: Sociable Unit
 * real: createProgress use-case（A 足場）
 * double: StubProgressRepository
 */
const command: ProgressWriteCommand = {
  episodeId: "ep-1",
  positionSec: 12,
  clientAt: "2026-09-22T10:00:00.000Z",
};

describe("createProgress", () => {
  it("Port 経由で ProgressWriteResponse 形の zero value を返す", async () => {
    // Given: Stub Port
    const repository = new StubProgressRepository();

    // When: create を実行する
    const got = await createProgress(repository, command);

    // Then: 契約 schema を満たす（値の意味は C）
    expect(ProgressWriteResponseSchema.safeParse(got).success).toBe(true);
    expect(got).toEqual({ firstPlayedAt: "1970-01-01T00:00:00.000Z", firstCompletedAt: null });
  });
});
