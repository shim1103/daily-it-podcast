// @vitest-environment node
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ErrorResponseSchema,
  GetEpisodeResponseSchema,
  ListEpisodesResponseSchema,
  episodeAudioContentType,
  episodeAudioPath,
  episodePath,
  listEpisodesPath,
} from "../contracts/index.ts";
import workerEntry from "../worker/src/worker-entry.ts";
import { validAudioBytes } from "../worker/src/test/fixtures/audio-bytes.ts";

/**
 * scope: Broad Integration
 * real: Worker entry, Composition Root, Controllers, UseCases, GoogleDriveEpisodeRepository
 * double: Drive HTTP（global fetch）。真の外部だけを差し替える
 * precondition: Drive secret は dummy。設定は 4 key 充足（production 相当の Drive 選択）
 * postcondition: list / get / audio 成功が入口から観測でき、代表 infra 失敗は 503 unavailable
 * invariant: Narrow の実 HTTP I/O や SU の分岐表を再 assert しない。既存 config BI と役割分担する
 */

const sufficientEnv = {
  GOOGLE_OAUTH_CLIENT_ID: "bi-drive-client-id-real-value",
  GOOGLE_OAUTH_CLIENT_SECRET: "bi-drive-client-secret-real-value",
  GOOGLE_OAUTH_REFRESH_TOKEN: "bi-drive-refresh-token-real-value",
  DRIVE_FOLDER_ID: "bi-drive-folder-id-real-value",
};

const manuscriptJson = {
  episodeId: "ep-bi-1",
  date: "2026-08-17",
  title: "Broad 配線の題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [
      { title: "第一", preface: "前1", detail: "詳1", startSec: 0 },
      { title: "第二", preface: "前2", detail: "詳2", startSec: 30 },
    ],
    closing: "終了",
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

function createDriveDoubleFetch(options: {
  files: DriveFileEntry[];
  downloads?: Record<string, string | Uint8Array>;
  listStatus?: number;
}): typeof fetch {
  const downloads = options.downloads ?? {};
  const listStatus = options.listStatus ?? 200;
  return (async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url === "https://oauth2.googleapis.com/token") {
      return new Response(JSON.stringify({ access_token: "bi-access-token" }), { status: 200 });
    }
    if (url.startsWith("https://www.googleapis.com/drive/v3/files?")) {
      if (listStatus !== 200) {
        return new Response("drive list failed", { status: listStatus });
      }
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
    throw new Error(`Drive double 未対応: ${url}`);
  }) as typeof fetch;
}

const successDriveFiles: DriveFileEntry[] = [
  { id: "bi-json-1", name: "ep-bi-1.json" },
  { id: "bi-wav-1", name: "ep-bi-1.wav" },
];

const successDownloads: Record<string, string | Uint8Array> = {
  "bi-json-1": JSON.stringify(manuscriptJson),
  "bi-wav-1": validAudioBytes,
};

describe("Playback Worker composition with Drive double", () => {
  const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

  afterEach(() => {
    vi.unstubAllGlobals();
    errorSpy.mockClear();
  });

  it("returns list episodes from Worker entry when Drive double succeeds", async () => {
    // Given: 設定十分な env + Drive double が適合原稿を返す
    vi.stubGlobal(
      "fetch",
      createDriveDoubleFetch({ files: successDriveFiles, downloads: successDownloads }),
    );
    const request = new Request(`https://worker.example${listEpisodesPath}`);

    // When: Worker HTTP 入口を呼ぶ
    const response = await workerEntry.fetch(request, sufficientEnv);

    // Then: Composition 経由で一覧成功が入口から観測できる
    expect(response.status).toBe(200);
    const body: unknown = await response.json();
    expect(ListEpisodesResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toEqual({
      episodes: [
        {
          episodeId: "ep-bi-1",
          date: "2026-08-17",
          title: "Broad 配線の題",
          durationSec: 60,
          topics: [{ title: "第一" }, { title: "第二" }],
          audioRef: episodeAudioPath("ep-bi-1"),
        },
      ],
    });
  });

  it("returns episode detail from Worker entry when Drive double succeeds", async () => {
    // Given: 設定十分な env + Drive double が json+wav を返す
    vi.stubGlobal(
      "fetch",
      createDriveDoubleFetch({ files: successDriveFiles, downloads: successDownloads }),
    );
    const request = new Request(`https://worker.example${episodePath("ep-bi-1")}`);

    // When: Worker HTTP 入口を呼ぶ
    const response = await workerEntry.fetch(request, sufficientEnv);

    // Then: 詳細成功が入口から観測できる
    expect(response.status).toBe(200);
    const body: unknown = await response.json();
    expect(GetEpisodeResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toMatchObject({
      episodeId: "ep-bi-1",
      title: "Broad 配線の題",
      audioRef: episodeAudioPath("ep-bi-1"),
    });
  });

  it("returns episode audio from Worker entry when Drive double succeeds", async () => {
    // Given: 設定十分な env + Drive double が wav byte を返す
    vi.stubGlobal(
      "fetch",
      createDriveDoubleFetch({ files: successDriveFiles, downloads: successDownloads }),
    );
    const request = new Request(`https://worker.example${episodeAudioPath("ep-bi-1")}`);

    // When: Worker HTTP 入口を呼ぶ
    const response = await workerEntry.fetch(request, sufficientEnv);

    // Then: 音声成功が入口から観測できる
    expect(response.status).toBe(200);
    expect(response.headers.get("Content-Type")).toBe(episodeAudioContentType);
    const bytes = new Uint8Array(await response.arrayBuffer());
    expect(bytes).toEqual(validAudioBytes);
  });

  it("returns 503 unavailable when Drive double fails with infra error", async () => {
    // Given: 設定十分な env + Drive double が files.list で infra 失敗する
    vi.stubGlobal(
      "fetch",
      createDriveDoubleFetch({
        files: successDriveFiles,
        downloads: successDownloads,
        listStatus: 500,
      }),
    );
    const request = new Request(`https://worker.example${listEpisodesPath}`);

    // When: Worker HTTP 入口を呼ぶ
    const response = await workerEntry.fetch(request, sufficientEnv);

    // Then: 合成経路の error 伝播で境界の公開形（503 unavailable）になる
    expect(response.status).toBe(503);
    const body: unknown = await response.json();
    expect(body).toEqual({ code: "unavailable" });
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(errorSpy).toHaveBeenCalled();
    const logged = JSON.stringify(errorSpy.mock.calls[0]?.[0]);
    for (const secret of Object.values(sufficientEnv)) {
      expect(logged).not.toContain(secret);
      expect(JSON.stringify(body)).not.toContain(secret);
    }
  });
});
