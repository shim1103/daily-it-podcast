// @vitest-environment node
import http from "node:http";
import { afterEach, describe, expect, it } from "vitest";
import { DriveError } from "../worker/src/infrastructure/drive/drive-error.ts";
import { GoogleDriveEpisodeRepository } from "../worker/src/infrastructure/drive/google-drive-episode-repository.ts";
import { validAudioBytes } from "../worker/src/test/fixtures/audio-bytes.ts";

/**
 * scope: Narrow Integration
 * real: Node HTTP listener, globalThis.fetch 経由の TCP/HTTP, GoogleDriveEpisodeRepository
 * double: OAuth / folder は dummy。本番 Google host へは出ない（local listener へ rewrite）
 * precondition: secret 実値なし。外部固定 port に依存しない
 * postcondition: 実 HTTP で list / download 成功と代表失敗が DriveError 契約へ写る
 * invariant: error / 失敗文言に dummy secret・folder id の実値を含めない
 */

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

type ListeningServer = {
  baseUrl: string;
  close: () => Promise<void>;
};

const dummyOAuth = {
  clientId: "ni-drive-client-id-real-value",
  clientSecret: "ni-drive-client-secret-real-value",
  refreshToken: "ni-drive-refresh-token-real-value",
};
const dummyFolderId = "ni-drive-folder-id-real-value";

const manuscriptJson = {
  episodeId: "ep-ni-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [{ title: "題", preface: "前置き", detail: "詳細", startSec: 0 }],
    closing: "終了",
  },
};

function closeServer(server: http.Server): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    server.close((error) => (error ? reject(error) : resolve()));
  });
}

async function startDriveUpstream(
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

/**
 * 本番 host 宛 URL を local listener へ rewrite し、実 fetch で TCP/HTTP する。
 * why: Adapter は token / files endpoint を定数に持つ。generator の DialTLS redirect と同型で
 *   接続先だけを差し替える。
 */
function createProxiedFetch(baseUrl: string): FetchLike {
  return async (input, init) => {
    const remote = new URL(input);
    if (remote.hostname !== "oauth2.googleapis.com" && remote.hostname !== "www.googleapis.com") {
      throw new Error(`想定外 host: ${remote.hostname}`);
    }
    const local = new URL(`${remote.pathname}${remote.search}`, baseUrl);
    return globalThis.fetch(local.toString(), init);
  };
}

function createRepository(fetchImpl: FetchLike): GoogleDriveEpisodeRepository {
  return new GoogleDriveEpisodeRepository({
    fetch: fetchImpl,
    oauth: dummyOAuth,
    folderId: dummyFolderId,
  });
}

function assertNoSecretLeak(text: string): void {
  expect(text).not.toContain(dummyOAuth.clientId);
  expect(text).not.toContain(dummyOAuth.clientSecret);
  expect(text).not.toContain(dummyOAuth.refreshToken);
  expect(text).not.toContain(dummyFolderId);
}

describe("GoogleDriveEpisodeRepository narrow HTTP boundary", () => {
  let server: ListeningServer | undefined;

  afterEach(async () => {
    if (server !== undefined) {
      await server.close();
      server = undefined;
    }
  });

  it("returns manuscript stems when real HTTP list and download succeed", async () => {
    // Given: token / files.list / media download が実 HTTP で成功する local upstream
    server = await startDriveUpstream((req, res) => {
      const url = req.url ?? "";
      if (req.method === "POST" && url === "/token") {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ access_token: "ni-access-token" }));
        return;
      }
      if (req.method === "GET" && url.startsWith("/drive/v3/files?")) {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ files: [{ id: "file-json-1", name: "ep-ni-1.json" }] }));
        return;
      }
      if (req.method === "GET" && url.startsWith("/drive/v3/files/file-json-1?")) {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(manuscriptJson));
        return;
      }
      res.writeHead(500);
      res.end("unexpected");
    });
    const repository = createRepository(createProxiedFetch(server.baseUrl));

    // When: 実 HTTP 経由で一覧を取る
    const got = await repository.listManuscripts();

    // Then: download した生 payload が stem 付きで返る
    expect(got).toEqual([{ stem: "ep-ni-1", json: manuscriptJson }]);
  });

  it("returns wav bytes when real HTTP media download succeeds", async () => {
    // Given: token / name 絞り込み list / wav download が実 HTTP で成功する local upstream
    server = await startDriveUpstream((req, res) => {
      const url = req.url ?? "";
      if (req.method === "POST" && url === "/token") {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ access_token: "ni-access-token" }));
        return;
      }
      if (req.method === "GET" && url.startsWith("/drive/v3/files?")) {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(
          JSON.stringify({
            files: [
              { id: "file-json-1", name: "ep-ni-1.json" },
              { id: "file-wav-1", name: "ep-ni-1.wav" },
            ],
          }),
        );
        return;
      }
      if (req.method === "GET" && url.startsWith("/drive/v3/files/file-wav-1?")) {
        res.writeHead(200, { "Content-Type": "audio/wav" });
        res.end(Buffer.from(validAudioBytes));
        return;
      }
      res.writeHead(500);
      res.end("unexpected");
    });
    const repository = createRepository(createProxiedFetch(server.baseUrl));

    // When: 実 HTTP 経由で音声を取る
    const got = await repository.getEpisodeAudio("ep-ni-1");

    // Then: wav byte が一致する
    expect(got).toEqual(validAudioBytes);
  });

  it("throws DriveError without dummy secrets when real HTTP files.list returns 500", async () => {
    // Given: token は成功し files.list だけ 500 を返す local upstream
    server = await startDriveUpstream((req, res) => {
      const url = req.url ?? "";
      if (req.method === "POST" && url === "/token") {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ access_token: "ni-access-token" }));
        return;
      }
      if (req.method === "GET" && url.startsWith("/drive/v3/files?")) {
        res.writeHead(500);
        res.end("upstream failure");
        return;
      }
      res.writeHead(500);
      res.end("unexpected");
    });
    const repository = createRepository(createProxiedFetch(server.baseUrl));

    // When / Then: 実 HTTP 失敗が DriveError へ写り、dummy secret が漏れない
    let caught: unknown;
    try {
      await repository.listManuscripts();
    } catch (error) {
      caught = error;
    }
    expect(caught).toBeInstanceOf(DriveError);
    assertNoSecretLeak(caught instanceof Error ? caught.message : String(caught));
  });

  it("throws DriveError without dummy secrets when real HTTP list body is malformed", async () => {
    // Given: files.list が 200 だが files 配列形式が不正な local upstream
    server = await startDriveUpstream((req, res) => {
      const url = req.url ?? "";
      if (req.method === "POST" && url === "/token") {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ access_token: "ni-access-token" }));
        return;
      }
      if (req.method === "GET" && url.startsWith("/drive/v3/files?")) {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ files: [{ id: 1, name: "ep-ni-1.json" }] }));
        return;
      }
      res.writeHead(500);
      res.end("unexpected");
    });
    const repository = createRepository(createProxiedFetch(server.baseUrl));

    // When / Then: 形式不正が DriveError へ写り、dummy secret が漏れない
    let caught: unknown;
    try {
      await repository.listManuscripts();
    } catch (error) {
      caught = error;
    }
    expect(caught).toBeInstanceOf(DriveError);
    assertNoSecretLeak(caught instanceof Error ? caught.message : String(caught));
  });
});
