import { beforeEach, describe, expect, it, vi } from "vitest";
import { EPISODE_PROGRESS_D1_BINDING } from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";
import { createLocalD1Binding } from "./create-local-d1-binding.ts";

/**
 * scope: Sociable Unit
 * real: createLocalD1Binding の binding 欠落分岐
 * double: wrangler の getPlatformProxy（実 proxy を起動しない）
 *
 * 実 proxy の起動と実 binding の読み書きは `test/integration/local_d1_binding.narrow_integration.test.ts` が所有する。
 */

const { getPlatformProxy } = vi.hoisted(() => ({ getPlatformProxy: vi.fn() }));

vi.mock("wrangler", () => ({ getPlatformProxy }));

describe("createLocalD1Binding", () => {
  beforeEach(() => {
    getPlatformProxy.mockReset();
  });

  it("disposes_proxy_and_rejects_when_binding_is_missing_in_env", async () => {
    // Given: EPISODE_PROGRESS binding を持たない env を返す proxy
    const dispose = vi.fn(async () => {});
    getPlatformProxy.mockResolvedValue({ env: {}, dispose });

    // When: local D1 binding を用意する
    const got = createLocalD1Binding();

    // Then: 黙って Fake に落とさず binding 名つきで reject し、起動済みの proxy を解放する
    await expect(got).rejects.toThrow(EPISODE_PROGRESS_D1_BINDING);
    expect(dispose).toHaveBeenCalledTimes(1);
  });
});
