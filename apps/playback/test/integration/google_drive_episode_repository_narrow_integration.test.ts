// @vitest-environment node
import http from "node:http";
import type { AddressInfo } from "node:net";
import { afterEach, describe, expect, it } from "vitest";
import { DriveError } from "../../worker/src/infrastructure/drive/drive-error.ts";
import { GoogleDriveEpisodeRepository } from "../../worker/src/infrastructure/drive/google-drive-episode-repository.ts";
import { validAudioBytes } from "../../worker/src/test/fixtures/audio-bytes.ts";

/**
 * scope: Narrow Integration
 * 実物境界: GoogleDriveEpisodeRepository が送信する外向き HTTP（token / files.list / download）
 * Double: 本番 Google Drive / OAuth は使わない。custom fetch が URL host を local listener へ書き換え、実 fetch で届ける。
 * @require dummy OAuth / Folder ID を Adapter へ直接渡す。upstream は controllable な test server。
 * @ensure list / download 相当の成功と代表失敗（非 2xx・形式不正）が DriveError 契約へ写る。
 * @invariant error message・assertion 失敗文言に dummy secret / folder id / file id の実値を含めない。
 */

const narrowClientId = "gdrive-narrow-client-id-real-value";
const narrowClientSecret = "gdrive-narrow-client-secret-real-value";
const narrowRefreshToken = "gdrive-narrow-refresh-token-real-value";
const narrowFolderId = "gdrive-narrow-folder-id-real-value";
const narrowJsonFileId = "gdrive-narrow-json-file-id-real-value";
const narrowWavFileId = "gdrive-narrow-wav-file-id-real-value";
const narrowAccessToken = "ya29.gdrive-narrow-access-token-real-value";

const productionHosts = new Set(["oauth2.googleapis.com", "www.googleapis.com"]);

const manuscriptJson = {
  episodeId: "narrow-ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: { text: "開始", startSec: 0 },
    topics: [{ title: "題", preface: "前置き", detail: "詳細", startSec: 0 }],
    closing: { summary: "終了", startSec: 55 },
  },
};

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

type UpstreamCall = {
  method: string;
  pathname: string;
  search: string;
  authorization: string | null;
};

const sensitiveValues = [
  narrowClientId,
  narrowClientSecret,
  narrowRefreshToken,
  narrowFolderId,
  narrowJsonFileId,
  narrowWavFileId,
  narrowAccessToken,
];

function isDriveErrorWithoutSensitiveValues(error: unknown): boolean {
  if (!(error instanceof DriveError)) {
    return false;
  }
  // why: expect(...).not.toContain(secret) は失敗時に secret 実値を assertion 文言へ出すため、predicate 内で判定する
  return !sensitiveValues.some((value) => error.message.includes(value));
}

function createHostRedirectFetch(localOrigin: string): FetchLike {
  return async (input, init) => {
    const original = new URL(input);
    if (!productionHosts.has(original.hostname)) {
      throw new Error(`想定外の host への呼び出し`);
    }
    const redirected = new URL(`${original.pathname}${original.search}`, localOrigin);
    return fetch(redirected, init);
  };
}

async function listen(
  handler: http.RequestListener,
): Promise<{ origin: string; server: http.Server; calls: UpstreamCall[] }> {
  const calls: UpstreamCall[] = [];
  const server = http.createServer((req, res) => {
    const url = new URL(req.url ?? "/", "http://127.0.0.1");
    calls.push({
      method: req.method ?? "",
      pathname: url.pathname,
      search: url.search,
      authorization: req.headers.authorization ?? null,
    });
    handler(req, res);
  });
  await new Promise<void>((resolve) => {
    server.listen(0, "127.0.0.1", () => resolve());
  });
  const address = server.address() as AddressInfo;
  return { origin: `http://127.0.0.1:${address.port}`, server, calls };
}

function createRepository(fetchImpl: FetchLike): GoogleDriveEpisodeRepository {
  return new GoogleDriveEpisodeRepository({
    fetch: fetchImpl,
    oauth: {
      clientId: narrowClientId,
      clientSecret: narrowClientSecret,
      refreshToken: narrowRefreshToken,
    },
    folderId: narrowFolderId,
  });
}

function writeJson(res: http.ServerResponse, status: number, body: unknown): void {
  res.writeHead(status, { "Content-Type": "application/json" });
  res.end(JSON.stringify(body));
}

function writeText(res: http.ServerResponse, status: number, body: string): void {
  res.writeHead(status, { "Content-Type": "text/plain" });
  res.end(body);
}

function oauthSuccessHandler(
  req: http.IncomingMessage,
  res: http.ServerResponse,
  next: http.RequestListener,
): void {
  const url = new URL(req.url ?? "/", "http://127.0.0.1");
  if (req.method === "POST" && url.pathname === "/token") {
    writeJson(res, 200, { access_token: narrowAccessToken });
    return;
  }
  next(req, res);
}

describe("GoogleDriveEpisodeRepository Narrow Integration", () => {
  let activeServer: http.Server | undefined;

  afterEach(async () => {
    if (activeServer === undefined) {
      return;
    }
    const server = activeServer;
    activeServer = undefined;
    await new Promise<void>((resolve, reject) => {
      server.close((error) => {
        if (error) {
          reject(error);
          return;
        }
        resolve();
      });
    });
  });

  it("listManuscripts returns stem and raw json when token list and download succeed over real HTTP", async () => {
    // Given: token / list / download が 2xx を返す local upstream
    const { origin, server, calls } = await listen((req, res) => {
      oauthSuccessHandler(req, res, (innerReq, innerRes) => {
        const url = new URL(innerReq.url ?? "/", "http://127.0.0.1");
        if (innerReq.method === "GET" && url.pathname === "/drive/v3/files") {
          writeJson(innerRes, 200, {
            files: [{ id: narrowJsonFileId, name: "narrow-ep-1.json" }],
          });
          return;
        }
        if (
          innerReq.method === "GET" &&
          url.pathname === `/drive/v3/files/${narrowJsonFileId}` &&
          url.searchParams.get("alt") === "media"
        ) {
          writeJson(innerRes, 200, manuscriptJson);
          return;
        }
        writeText(innerRes, 404, "unexpected");
      });
    });
    activeServer = server;
    const repository = createRepository(createHostRedirectFetch(origin));

    // When: 一覧を取得する
    const got = await repository.listManuscripts();

    // Then: list + download が実 HTTP 経由で成功し、生 payload が返る
    expect(got).toEqual([{ stem: "narrow-ep-1", json: manuscriptJson }]);
    expect(calls.some((call) => call.pathname === "/token")).toBe(true);
    expect(
      calls.some((call) => call.pathname === "/drive/v3/files" && call.search.includes("q=")),
    ).toBe(true);
    const downloaded = calls.some(
      (call) =>
        call.pathname.startsWith("/drive/v3/files/") &&
        call.pathname !== "/drive/v3/files" &&
        call.search.includes("alt=media"),
    );
    expect(downloaded).toBe(true);
    // why: expect 失敗文言へ access token 実値を出さないため boolean で観測する
    const bearerMatchesToken = calls.some(
      (call) =>
        call.pathname.startsWith("/drive/v3/files") &&
        call.authorization === `Bearer ${narrowAccessToken}`,
    );
    expect(bearerMatchesToken).toBe(true);
  });

  it("getAudio returns wav bytes when download succeeds over real HTTP", async () => {
    // Given: token / 絞り込み list / wav download が 2xx を返す local upstream
    const { origin, server } = await listen((req, res) => {
      oauthSuccessHandler(req, res, (innerReq, innerRes) => {
        const url = new URL(innerReq.url ?? "/", "http://127.0.0.1");
        if (innerReq.method === "GET" && url.pathname === "/drive/v3/files") {
          writeJson(innerRes, 200, {
            files: [
              { id: narrowJsonFileId, name: "narrow-ep-1.json" },
              { id: narrowWavFileId, name: "narrow-ep-1.wav" },
            ],
          });
          return;
        }
        if (
          innerReq.method === "GET" &&
          url.pathname === `/drive/v3/files/${narrowWavFileId}` &&
          url.searchParams.get("alt") === "media"
        ) {
          innerRes.writeHead(200, { "Content-Type": "audio/wav" });
          innerRes.end(Buffer.from(validAudioBytes));
          return;
        }
        writeText(innerRes, 404, "unexpected");
      });
    });
    activeServer = server;
    const repository = createRepository(createHostRedirectFetch(origin));

    // When: 音声を取得する
    const got = await repository.getAudio("narrow-ep-1");

    // Then: download 相当が実 HTTP 経由で成功する
    expect(got).toEqual(validAudioBytes);
  });

  it("throws DriveError without sensitive values when files.list returns non-2xx", async () => {
    // Given: files.list が 500 を返す local upstream（request() 非 2xx の代表1本）
    const { origin, server } = await listen((req, res) => {
      oauthSuccessHandler(req, res, (innerReq, innerRes) => {
        const url = new URL(innerReq.url ?? "/", "http://127.0.0.1");
        if (innerReq.method === "GET" && url.pathname === "/drive/v3/files") {
          writeText(innerRes, 500, "server error");
          return;
        }
        writeText(innerRes, 404, "unexpected");
      });
    });
    activeServer = server;
    const repository = createRepository(createHostRedirectFetch(origin));

    // When / Then: DriveError になり、message に dummy 実値を含めない
    await expect(repository.listManuscripts()).rejects.toSatisfy(
      isDriveErrorWithoutSensitiveValues,
    );
  });

  it("throws DriveError when files.list payload shape is invalid", async () => {
    // Given: files.list が形式不正（id が number）を返す local upstream
    const { origin, server } = await listen((req, res) => {
      oauthSuccessHandler(req, res, (innerReq, innerRes) => {
        const url = new URL(innerReq.url ?? "/", "http://127.0.0.1");
        if (innerReq.method === "GET" && url.pathname === "/drive/v3/files") {
          writeJson(innerRes, 200, { files: [{ id: 1, name: "narrow-ep-1.json" }] });
          return;
        }
        writeText(innerRes, 404, "unexpected");
      });
    });
    activeServer = server;
    const repository = createRepository(createHostRedirectFetch(origin));

    // When / Then: 形式不正は DriveError
    await expect(repository.listManuscripts()).rejects.toSatisfy(
      isDriveErrorWithoutSensitiveValues,
    );
  });

  it("throws DriveError when files.list body is not JSON", async () => {
    // Given: files.list が 200 だが JSON でない local upstream
    const { origin, server } = await listen((req, res) => {
      oauthSuccessHandler(req, res, (innerReq, innerRes) => {
        const url = new URL(innerReq.url ?? "/", "http://127.0.0.1");
        if (innerReq.method === "GET" && url.pathname === "/drive/v3/files") {
          writeText(innerRes, 200, "not json");
          return;
        }
        writeText(innerRes, 404, "unexpected");
      });
    });
    activeServer = server;
    const repository = createRepository(createHostRedirectFetch(origin));

    await expect(repository.listManuscripts()).rejects.toSatisfy(
      isDriveErrorWithoutSensitiveValues,
    );
  });

  it("throws DriveError when token response omits access_token", async () => {
    // Given: token 応答に access_token が無い local upstream
    const { origin, server } = await listen((_req, res) => {
      writeJson(res, 200, {});
    });
    activeServer = server;
    const repository = createRepository(createHostRedirectFetch(origin));

    await expect(repository.listManuscripts()).rejects.toSatisfy(
      isDriveErrorWithoutSensitiveValues,
    );
  });

  it("throws DriveError when token body is not JSON", async () => {
    // Given: token 応答が 200 だが JSON でない local upstream
    const { origin, server } = await listen((_req, res) => {
      writeText(res, 200, "not json");
    });
    activeServer = server;
    const repository = createRepository(createHostRedirectFetch(origin));

    await expect(repository.listManuscripts()).rejects.toSatisfy(
      isDriveErrorWithoutSensitiveValues,
    );
  });

  it("throws DriveError with cause when real HTTP transport fails", async () => {
    // Given: 接続不能な local origin（listener 無し）
    const repository = createRepository(createHostRedirectFetch("http://127.0.0.1:1"));

    // When / Then: network 失敗が DriveError へ写り、cause を持つ
    await expect(repository.listManuscripts()).rejects.toSatisfy(
      (error: unknown) =>
        isDriveErrorWithoutSensitiveValues(error) &&
        error instanceof DriveError &&
        error.cause instanceof Error,
    );
  });
});
