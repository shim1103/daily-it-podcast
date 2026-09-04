import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ErrorResponseSchema,
  episodeAudioContentType,
  episodeAudioPath,
  ListEpisodesResponseSchema,
  listEpisodesPath,
} from "../../contracts/index.ts";
import { validAudioBytes } from "../../worker/src/test/fixtures/audio-bytes.ts";
import workerEntry from "../../worker/src/worker-entry.ts";

/**
 * scope: Broad Integration
 * real: Worker entry・route・Composition Root・Controller・UseCase・GoogleDriveEpisodeRepository
 * double: Drive HTTP（global `fetch` Stub）。真 Google へは行かない
 * precondition: Drive env が揃い production Composition が Drive repository を選ぶ
 * postcondition: list / get audio の成功応答が入口から見える。代表の Drive 失敗は 503 unavailable
 * invariant: PlaybackUseCaseOverrides で use case 直差ししない。secret 実値を assert 失敗文言へ出さない
 */

const driveEnv = {
  GOOGLE_OAUTH_CLIENT_ID: "bi-comp-client-id-dummy",
  GOOGLE_OAUTH_CLIENT_SECRET: "bi-comp-client-secret-dummy",
  GOOGLE_OAUTH_REFRESH_TOKEN: "bi-comp-refresh-token-dummy",
  DRIVE_FOLDER_ID: "bi-comp-folder-id-dummy",
};

const sensitiveValues = Object.values(driveEnv);

const episodeId = "bi-ep-1";
const jsonFileId = "bi-comp-json-file-id";
const wavFileId = "bi-comp-wav-file-id";

const manuscriptJson = {
  episodeId,
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: { text: "開始", startSec: 0 },
    topics: [{ title: "題", preface: "前置き", detail: "詳細", startSec: 0 }],
    closing: { summary: "終了", startSec: 55 },
  },
};

type DriveFileEntry = { id: string; name: string };

function extractNameFilters(query: string): string[] | undefined {
  const matches = [...query.matchAll(/name = '([^']*)'/g)];
  if (matches.length === 0) {
    return undefined;
  }
  return matches.map((match) => match[1] ?? "");
}

function createDriveFetchStub(options: {
  files: DriveFileEntry[];
  downloads?: Record<string, string | Uint8Array>;
  tokenStatus?: number;
}): typeof fetch {
  const downloads = options.downloads ?? {};
  const tokenStatus = options.tokenStatus ?? 200;
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = typeof input === "string" ? input : input instanceof URL ? input.href : input.url;
    if (url === "https://oauth2.googleapis.com/token") {
      if (tokenStatus !== 200) {
        return new Response(null, { status: tokenStatus });
      }
      return new Response(JSON.stringify({ access_token: "bi-comp-access-token-dummy" }), {
        status: 200,
      });
    }
    if (url.startsWith("https://www.googleapis.com/drive/v3/files?")) {
      const parsed = new URL(url);
      const query = parsed.searchParams.get("q") ?? "";
      const nameFilters = extractNameFilters(query);
      const files =
        nameFilters === undefined
          ? options.files
          : options.files.filter((file) => nameFilters.includes(file.name));
      return new Response(JSON.stringify({ files }), { status: 200 });
    }
    const downloadMatch =
      /^https:\/\/www\.googleapis\.com\/drive\/v3\/files\/([^?]+)\?alt=media$/.exec(url);
    if (downloadMatch) {
      const fileId = downloadMatch[1] ?? "";
      const body = downloads[fileId];
      if (body === undefined) {
        return new Response(null, { status: 404 });
      }
      if (typeof body === "string") {
        return new Response(body, { status: 200 });
      }
      const copy = new Uint8Array(body);
      return new Response(copy.buffer, { status: 200 });
    }
    throw new Error("Stub 未対応の呼び出し");
  }) as unknown as typeof fetch;
}

function installDriveFetchStub(options: Parameters<typeof createDriveFetchStub>[0]): void {
  vi.stubGlobal("fetch", createDriveFetchStub(options));
}

function installHappyDriveFetchStub(): void {
  installDriveFetchStub({
    files: [
      { id: jsonFileId, name: `${episodeId}.json` },
      { id: wavFileId, name: `${episodeId}.wav` },
    ],
    downloads: {
      [jsonFileId]: JSON.stringify(manuscriptJson),
      [wavFileId]: validAudioBytes,
    },
  });
}

function textOmitsSensitiveValues(text: string): boolean {
  return !sensitiveValues.some((value) => text.includes(value));
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("Playback Worker composition happy path", () => {
  it("returns 200 list episodes through Worker entry when Drive env is complete", async () => {
    // Given: Drive env が揃い、Drive HTTP は Stub が成功応答を返す
    installHappyDriveFetchStub();
    const request = new Request(`https://worker.example${listEpisodesPath}`);

    // When: Worker HTTP 入口へ一覧 GET
    const response = await workerEntry.fetch(request, driveEnv);

    // Then: 入口から list 成功が見える（下位 mapping の全 field 一致はしない）
    expect(response.status).toBe(200);
    const body: unknown = await response.json();
    const parsed = ListEpisodesResponseSchema.safeParse(body);
    expect(parsed.success).toBe(true);
    if (!parsed.success) {
      return;
    }
    expect(parsed.data.episodes).toHaveLength(1);
    expect(parsed.data.episodes[0]?.body.opening).toEqual({ text: "開始", startSec: 0 });
    expect(textOmitsSensitiveValues(JSON.stringify(body))).toBe(true);
  });

  it("returns 200 audio bytes through Worker entry when Drive env is complete", async () => {
    // Given: Drive env が揃い、対象 wav が Stub にある
    installHappyDriveFetchStub();
    const request = new Request(`https://worker.example${episodeAudioPath(episodeId)}`);

    // When: Worker HTTP 入口へ音声 GET
    const response = await workerEntry.fetch(request, driveEnv);

    // Then: 入口から audio 成功が見える（bytes 完全一致はしない）
    expect(response.status).toBe(200);
    expect(response.headers.get("Content-Type")).toBe(episodeAudioContentType);
    const bytes = new Uint8Array(await response.arrayBuffer());
    expect(bytes.byteLength).toBeGreaterThan(0);
  });

  it("returns 503 unavailable when Drive token endpoint fails through composition", async () => {
    // Given: Drive env は揃うが token 取得が非 2xx（合成で初めて見える error 伝播の代表）
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    installDriveFetchStub({
      files: [],
      tokenStatus: 500,
    });
    const request = new Request(`https://worker.example${listEpisodesPath}`);

    // When: Worker HTTP 入口へ一覧 GET
    const response = await workerEntry.fetch(request, driveEnv);

    // Then: 入口から unavailable が見える（config 不足 BI と重複しない）
    expect(response.status).toBe(503);
    const body: unknown = await response.json();
    expect(body).toEqual({ code: "unavailable" });
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(textOmitsSensitiveValues(JSON.stringify(body))).toBe(true);
    const logged = JSON.stringify(errorSpy.mock.calls[0]?.[0] ?? null);
    expect(textOmitsSensitiveValues(logged)).toBe(true);
  });
});
