import { afterAll, describe, expect, it, vi } from "vitest";
import { ErrorResponseSchema, listEpisodesPath } from "../contracts/index.ts";
import { fetch as workerFetch } from "../worker/src/routes/fetch.ts";

/**
 * scope: Broad Integration
 * real: Worker route, Composition Root, HTTP error mapping
 * double: none
 * precondition: Worker env に Drive secret を注入しない
 * postcondition: 設定不足は InMemory の空成功ではなく 503 unavailable になる
 * invariant: HTTP error body は playback contract の schema を満たす
 */
describe("Playback Worker runtime config boundary", () => {
  const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

  afterAll(() => {
    errorSpy.mockRestore();
  });

  it("Drive config が無い production相当の Worker は 503 を返す", async () => {
    // Given: Worker に一部の secret binding だけがあり、folder id が無い
    const request = new Request(`https://worker.example${listEpisodesPath}`);
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id-secret",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret-value",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token-value",
    };

    // When: 実際の Worker HTTP 入口を呼ぶ
    const response = await workerFetch(request, env);

    // Then: InMemoryへ暗黙 fallbackせず設定エラー
    expect(response.status).toBe(503);
    const body: unknown = await response.json();
    expect(body).toEqual({ code: "unavailable" });
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(errorSpy).toHaveBeenCalledTimes(1);
    expect(errorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "UnavailableError",
        cause: expect.objectContaining({
          name: "PlaybackRuntimeConfigError",
          message: "Playback runtime config が不正です: DRIVE_FOLDER_ID が未設定です",
        }),
      }),
    );
    expect(JSON.stringify(errorSpy.mock.calls[0]?.[0])).not.toContain("client-id-secret");
    expect(JSON.stringify(errorSpy.mock.calls[0]?.[0])).not.toContain("client-secret-value");
    expect(JSON.stringify(errorSpy.mock.calls[0]?.[0])).not.toContain("refresh-token-value");
  });
});
