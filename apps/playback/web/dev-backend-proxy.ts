import type { Connect } from "vite";

/**
 * Hono app（Worker route一式）をVite dev serverのmiddlewareへ橋渡しする共通処理。
 * `/episodes*` だけをHono appへ透過的に中継する。Range/HEADはHono app自身
 * （routes/audio-response.ts）が処理するため、この層では何も解釈しない。
 *
 * @require handle は Request を受けて Response を返す fetch-like 関数
 * @ensure `/episodes*` 以外は次の middleware へ委譲する。request header・response header・
 *   status・body を素通しする（本番相当のHono appの応答をそのまま反映する）
 */
export function createDevBackendMiddleware(
  handle: (req: Request) => Promise<Response>,
): Connect.NextHandleFunction {
  return async (req, res, next) => {
    if (!req.url?.startsWith("/episodes")) {
      next();
      return;
    }
    const response = await handle(
      new Request(new URL(req.url, "http://localhost"), {
        method: req.method,
        headers: new Headers(req.headers as Record<string, string>),
      }),
    );
    res.statusCode = response.status;
    response.headers.forEach((value, key) => {
      res.setHeader(key, value);
    });
    res.end(Buffer.from(await response.arrayBuffer()));
  };
}
