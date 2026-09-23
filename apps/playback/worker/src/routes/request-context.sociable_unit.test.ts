import { Hono } from "hono";
import { requestId } from "hono/request-id";
import { afterEach, describe, expect, it, vi } from "vitest";
import { requestIdHeaderName, requestLoggingMiddleware } from "./request-context.ts";

const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});

afterEach(() => {
  logSpy.mockClear();
});

function buildApp() {
  return new Hono<{ Variables: { requestId: string } }>()
    .use(requestId())
    .use(requestLoggingMiddleware)
    .get("/episodes", (c) => c.json({ ok: true }))
    .get("/boom", () => {
      throw new Error("boom");
    })
    .onError((_error, c) => {
      return c.json({ requestId: c.get("requestId") }, 500);
    });
}

describe("requestLoggingMiddleware", () => {
  it("正常応答に requestId header を付与する", async () => {
    // Given: middleware を積んだ app
    const app = buildApp();

    // When: 一覧 path へ GET する
    const got = await app.request("http://example.test/episodes");

    // Then: response header に requestId が乗る
    const requestId = got.headers.get(requestIdHeaderName);
    expect(requestId).toBeTruthy();
  });

  it("開始ログと完了ログを同じ requestId で 2 件出す", async () => {
    // Given: middleware を積んだ app
    const app = buildApp();

    // When: 一覧 path へ GET する
    const got = await app.request("http://example.test/episodes");
    const requestId = got.headers.get(requestIdHeaderName);

    // Then: request_start と request_end が同じ requestId で記録される
    expect(logSpy).toHaveBeenCalledTimes(2);
    expect(logSpy).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({
        event: "request_start",
        requestId,
        method: "GET",
        path: "/episodes",
      }),
    );
    expect(logSpy).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({
        event: "request_end",
        requestId,
        method: "GET",
        path: "/episodes",
        status: 200,
      }),
    );
  });

  it("handler が throw した時も、次の middleware 折返しで完了ログを出す", async () => {
    // Given: throw する route
    const app = buildApp();

    // When: 失敗 path へ GET する
    const got = await app.request("http://example.test/boom");
    const body: unknown = await got.json();

    // Then: onError が確定させた status で完了ログが出て、requestId が onError 側にも伝播する
    expect(logSpy).toHaveBeenCalledTimes(2);
    expect(logSpy).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ event: "request_end", status: 500, path: "/boom" }),
    );
    expect(body).toEqual({ requestId: expect.any(String) });
  });

  it("requestごとに異なる requestId を発行する", async () => {
    // Given: middleware を積んだ app
    const app = buildApp();

    // When: 同じ path へ 2 回 GET する
    const first = await app.request("http://example.test/episodes");
    const second = await app.request("http://example.test/episodes");

    // Then: 毎回別の requestId
    expect(first.headers.get(requestIdHeaderName)).not.toBe(
      second.headers.get(requestIdHeaderName),
    );
  });
});
