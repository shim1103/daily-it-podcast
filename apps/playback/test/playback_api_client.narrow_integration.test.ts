// @vitest-environment node
import http from "node:http";
import { afterEach, describe, expect, it } from "vitest";
import { episodePath, listEpisodesPath } from "../contracts/index.ts";
import { createPlaybackApiClient } from "../web/src/api/playback-api-client.ts";

/**
 * scope: Narrow Integration
 * real: Node HTTP listener, globalThis.fetch, PlaybackApiClient
 * double: none（HTTP 応答は local listener が返す。Stub fetch ではない）
 * precondition: secret なし。外部固定 port に依存しない
 * postcondition: 実 TCP/HTTP 到達で list/get の Result と代表 status 写像が観測できる
 * invariant: schema 細部の網羅は SU に残し、本 suite は到達と代表写像に限る
 */

type ListeningServer = {
  baseUrl: string;
  close: () => Promise<void>;
};

const validListBody = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-20",
      title: "今日の IT",
      durationSec: 60,
      topics: [{ title: "題" }],
      audioRef: "/episodes/ep-1/audio",
    },
  ],
};

function closeServer(server: http.Server): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    server.close((error) => (error ? reject(error) : resolve()));
  });
}

async function startJsonServer(
  handler: (req: http.IncomingMessage, res: http.ServerResponse) => void,
): Promise<ListeningServer> {
  const server = http.createServer(handler);

  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", () => resolve());
  });

  const address = server.address();
  if (address === null || typeof address === "string") {
    await closeServer(server);
    throw new Error("HTTP listener の address を取得できない");
  }

  return {
    baseUrl: `http://127.0.0.1:${address.port}`,
    close: () => closeServer(server),
  };
}

describe("PlaybackApiClient narrow HTTP boundary", () => {
  let server: ListeningServer | undefined;

  afterEach(async () => {
    if (server !== undefined) {
      await server.close();
      server = undefined;
    }
  });

  it("returns ok Result for listEpisodes when real HTTP succeeds", async () => {
    // Given: port 0 の local HTTP listener が契約 path へ一覧 JSON を返す
    server = await startJsonServer((req, res) => {
      if (req.method === "GET" && req.url === listEpisodesPath) {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(validListBody));
        return;
      }
      res.writeHead(500);
      res.end();
    });
    const client = createPlaybackApiClient({
      baseUrl: server.baseUrl,
      fetch: globalThis.fetch,
    });

    // When: 実 fetch で一覧を呼ぶ
    const got = await client.listEpisodes();

    // Then: 実 HTTP 成功経路が ok Result になる（到達観測。schema 全量は SU）
    expect(got).toMatchObject({ ok: true, data: { episodes: [{ episodeId: "ep-1" }] } });
  });

  it("maps 404 from real HTTP to episode_not_found", async () => {
    // Given: port 0 の local HTTP listener が 1 件 path へ 404 を返す
    const targetPath = episodePath("ep-1");
    server = await startJsonServer((req, res) => {
      if (req.method === "GET" && req.url === targetPath) {
        res.writeHead(404);
        res.end();
        return;
      }
      res.writeHead(500);
      res.end();
    });
    const client = createPlaybackApiClient({
      baseUrl: server.baseUrl,
      fetch: globalThis.fetch,
    });

    // When: 実 fetch で 1 件を呼ぶ
    const got = await client.getEpisode("ep-1");

    // Then: 代表非成功 status が既存 API error 語彙へ写る
    expect(got).toEqual({ ok: false, error: "episode_not_found" });
  });
});
