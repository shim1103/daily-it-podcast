import { afterAll, describe, expect, it, vi } from "vitest";
import { ErrorResponseSchema, listEpisodesPath } from "../contracts/index.ts";
import workerEntry from "../worker/src/worker-entry.ts";

/**
 * scope: Broad Integration
 * real: Worker route, Composition Root, HTTP error mapping
 * double: none
 * precondition: Worker env に Drive secret を注入しない
 * postcondition: 設定不足は InMemory の空成功ではなく 500 configuration_error になる
 * invariant: HTTP error body は playback contract の schema を満たす
 */
describe("Playback Worker runtime config boundary", () => {
  const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    errorSpy.mockRestore();
  });

  it("Drive config が無い production相当の Worker は 500 を返す", async () => {
    // Given: Worker に一部の secret binding だけがあり、folder id が無い
    const request = new Request(`https://worker.example${listEpisodesPath}`);
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "dummy-client-id-regression",
      GOOGLE_OAUTH_CLIENT_SECRET: "dummy-client-secret-regression",
      GOOGLE_OAUTH_REFRESH_TOKEN: "dummy-refresh-token-regression",
    };

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
          message: "Playback runtime config が不正です: DRIVE_FOLDER_ID が未設定です",
        }),
      }),
    );
    const logged = JSON.stringify(errorSpy.mock.calls[0]?.[0]);
    for (const secret of Object.values(env)) {
      expect(logged).not.toContain(secret);
      expect(JSON.stringify(body)).not.toContain(secret);
    }
  });
});
