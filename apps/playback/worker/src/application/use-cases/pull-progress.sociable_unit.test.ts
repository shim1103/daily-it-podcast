import { describe, expect, it } from "vitest";
import { ProgressPullResponseSchema } from "../../../../contracts/index.ts";
import { StubProgressRepository } from "../ports/progress-repository.ts";
import { pullProgress } from "./pull-progress.ts";

/**
 * scope: Sociable Unit
 * real: pullProgress use-case（A 足場）
 * double: StubProgressRepository
 */
describe("pullProgress", () => {
  it("Port 経由で空の ProgressPullResponse を返す", async () => {
    // Given: Stub Port
    const repository = new StubProgressRepository();

    // When: pull を実行する
    const got = await pullProgress(repository, "2026-09-22T10:00:00.000Z");

    // Then: 契約 schema を満たし、差分なしは空配列
    expect(ProgressPullResponseSchema.safeParse(got).success).toBe(true);
    expect(got).toEqual({ episodes: [] });
  });
});
