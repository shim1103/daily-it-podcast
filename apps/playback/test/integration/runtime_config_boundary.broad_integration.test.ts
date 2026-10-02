// @vitest-environment node
import { afterAll, describe, expect, it, vi } from "vitest";
import { ErrorResponseSchema, listEpisodesPath } from "../../contracts/index.ts";
import workerEntry from "../../worker/src/worker-entry.ts";

/**
 * scope: Broad Integration
 * real: Worker route, Composition Root, HTTP error mapping
 * double: none
 * precondition: 本番route は episode の mode: "r2"・進捗の progressMode: "d1" 固定であり、Worker env に必須 binding（EPISODES・EPISODE_PROGRESS）の一方を注入しない
 * postcondition: 設定不足は InMemory の空成功ではなく 500 configuration_error になる
 * invariant: HTTP error body は playback contract の schema を満たす
 */
describe("Playback Worker runtime config boundary", () => {
  const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    errorSpy.mockRestore();
  });

  it("R2 binding が無い production相当の Worker は 500 を返す", async () => {
    // Given: 本番route は episode r2 ＋ 進捗 d1 固定であり、Worker env に EPISODES（R2 binding）が無い
    const request = new Request(`https://worker.example${listEpisodesPath}`);
    const env = {};

    // When: 実際の Worker HTTP 入口を呼ぶ
    const response = await workerEntry.fetch(request, env);

    // Then: InMemoryへ暗黙 fallbackせず設定エラー
    expect(response.status).toBe(500);
    const body: unknown = await response.json();
    expect(body).toEqual({ code: "configuration_error" });
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(errorSpy).toHaveBeenCalledTimes(1);
    expect(errorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "ConfigurationError",
        cause: expect.objectContaining({
          name: "PlaybackRuntimeConfigError",
          message: "EPISODES（R2 binding）が未設定です",
        }),
      }),
    );
  });

  it("D1 binding が無い production相当の Worker は 500 を返す", async () => {
    // Given: 本番route は episode r2 ＋ 進捗 d1 固定であり、Worker env に EPISODES はあるが EPISODE_PROGRESS（D1 binding）が無い
    const request = new Request(`https://worker.example${listEpisodesPath}`);
    const env = { EPISODES: { get: async () => null, list: async () => ({ objects: [] }) } };

    // When: 実際の Worker HTTP 入口を呼ぶ
    const response = await workerEntry.fetch(request, env);

    // Then: 進捗を in-memory へ暗黙 fallbackせず設定エラー
    expect(response.status).toBe(500);
    const body: unknown = await response.json();
    expect(body).toEqual({ code: "configuration_error" });
    expect(errorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "ConfigurationError",
        cause: expect.objectContaining({
          name: "PlaybackRuntimeConfigError",
          message: "EPISODE_PROGRESS（D1 binding）が未設定です",
        }),
      }),
    );
  });
});
