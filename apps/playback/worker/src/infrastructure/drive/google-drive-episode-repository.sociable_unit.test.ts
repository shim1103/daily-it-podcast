import { describe, expect, it, vi } from "vitest";
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";
import { DriveError } from "./drive-error.ts";
import { GoogleDriveEpisodeRepository } from "./google-drive-episode-repository.ts";

/**
 * scope: Sociable Unit
 * real: GoogleDriveEpisodeRepository
 * double: Drive HTTP を Stub 化した `fetch`（境界 provider ではない。実 I/O 契約は Narrow）
 *
 * why: Adapter 内分岐だけをここへ残す —— files.list の `q` 絞り込み、生 payload の decode 渡し、
 * Stub 応答からの DriveError 変換、file id / folder id を message に出さないこと、複数 download
 * の並行開始、「json が無い＝undefined」「wav が無い＝undefined」の戻り値表現。実 TCP/HTTP の
 * list / download 成功と代表失敗は `test/google_drive_episode_repository.narrow_integration.test.ts`
 * が所有する。schema 不適合除外・stem 不一致は use-case SU の所有。
 */
type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

const dummyOAuthConfig = {
  clientId: "dummy-client-id",
  clientSecret: "dummy-client-secret",
  refreshToken: "dummy-refresh-token",
};
const dummyFolderId = "dummy-folder-id";

const manuscriptJson = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [{ title: "題", preface: "前置き", detail: "詳細", startSec: 0 }],
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

function stubFetch(options: {
  files: DriveFileEntry[];
  downloads?: Record<string, string | Uint8Array>;
}): FetchLike {
  const downloads = options.downloads ?? {};
  return vi.fn(async (input: string, init?: RequestInit) => {
    if (input === "https://oauth2.googleapis.com/token") {
      return new Response(JSON.stringify({ access_token: "dummy-access-token" }), {
        status: 200,
      });
    }
    if (input.startsWith("https://www.googleapis.com/drive/v3/files?")) {
      const url = new URL(input);
      const query = url.searchParams.get("q") ?? "";
      const nameFilters = extractNameFilters(query);
      const files =
        nameFilters === undefined
          ? options.files
          : options.files.filter((file) => nameFilters.includes(file.name));
      return new Response(JSON.stringify({ files }), { status: 200 });
    }
    const downloadMatch =
      /^https:\/\/www\.googleapis\.com\/drive\/v3\/files\/([^?]+)\?alt=media$/.exec(input);
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
    void init;
    throw new Error(`Stub 未対応の呼び出し: ${input}`);
  });
}

function createRepository(fetchStub: FetchLike): GoogleDriveEpisodeRepository {
  return new GoogleDriveEpisodeRepository({
    fetch: fetchStub,
    oauth: dummyOAuthConfig,
    folderId: dummyFolderId,
  });
}

describe("GoogleDriveEpisodeRepository", () => {
  describe("生 payload を検証せず返す", () => {
    it("listManuscripts は download した json を stem 付きの生 payload として返す（検証しない）", async () => {
      // Given: 適合 json 1 + schema 不適合 json 1
      const repository = createRepository(
        stubFetch({
          files: [
            { id: "j1", name: "ep-1.json" },
            { id: "j2", name: "bad.json" },
          ],
          downloads: {
            j1: JSON.stringify(manuscriptJson),
            j2: JSON.stringify({ episodeId: "bad" }),
          },
        }),
      );

      // When: 一覧を取得する
      const got = await repository.listManuscripts();

      // Then: 不適合分も除外せず、生のまま両方返す
      expect(got).toEqual([
        { stem: "ep-1", json: manuscriptJson },
        { stem: "bad", json: { episodeId: "bad" } },
      ]);
    });

    it("listManuscripts は json エントリ 0 件でも空配列を返す（throw しない）", async () => {
      // Given: json が無いフォルダ
      const repository = createRepository(stubFetch({ files: [{ id: "w", name: "ep-1.wav" }] }));

      // When / Then
      expect(await repository.listManuscripts()).toEqual([]);
    });

    it("download した bytes が不正 JSON の件は、decode 生文字列をそのまま返す（分類しない）", async () => {
      // Given: JSON として壊れた download
      const repository = createRepository(
        stubFetch({
          files: [{ id: "j1", name: "ep-1.json" }],
          downloads: { j1: "not json" },
        }),
      );

      // When: 一覧を取得する
      const got = await repository.listManuscripts();

      // Then: string のまま渡す
      expect(got).toEqual([{ stem: "ep-1", json: "not json" }]);
    });

    it("getManuscript は download した json と wav 有無をそのまま返す（検証しない）", async () => {
      // Given: schema 不適合 json + wav あり
      const repository = createRepository(
        stubFetch({
          files: [
            { id: "j1", name: "ep-1.json" },
            { id: "w1", name: "ep-1.wav" },
          ],
          downloads: { j1: JSON.stringify({ episodeId: "ep-1" }) },
        }),
      );

      // When: 1件取得する
      const got = await repository.getManuscript("ep-1");

      // Then: 検証せず生 payload + hasAudio
      expect(got).toEqual({ json: { episodeId: "ep-1" }, hasAudio: true });
    });

    it("getManuscript は wav が無くても hasAudio: false で json を返す（throw しない）", async () => {
      // Given: json のみ
      const repository = createRepository(
        stubFetch({
          files: [{ id: "j1", name: "ep-1.json" }],
          downloads: { j1: JSON.stringify(manuscriptJson) },
        }),
      );

      // When: 1件取得する
      const got = await repository.getManuscript("ep-1");

      // Then: hasAudio: false
      expect(got).toEqual({ json: manuscriptJson, hasAudio: false });
    });
  });

  describe("取得対象の不在は undefined で表現する", () => {
    it("json エントリが Drive に無い時、getManuscript は undefined", async () => {
      // Given: フォルダが空
      const repository = createRepository(stubFetch({ files: [] }));

      // When / Then: throw せず undefined
      expect(await repository.getManuscript("missing")).toBeUndefined();
    });

    it("wav エントリが Drive に無い時、getEpisodeAudio は undefined", async () => {
      // Given: json のみ
      const repository = createRepository(
        stubFetch({
          files: [{ id: "j1", name: "ep-1.json" }],
          downloads: { j1: JSON.stringify(manuscriptJson) },
        }),
      );

      // When / Then
      expect(await repository.getEpisodeAudio("ep-1")).toBeUndefined();
    });

    it("wav エントリがある時、getEpisodeAudio は wav byte を返す", async () => {
      // Given: json + wav
      const repository = createRepository(
        stubFetch({
          files: [
            { id: "j1", name: "ep-1.json" },
            { id: "w1", name: "ep-1.wav" },
          ],
          downloads: { j1: JSON.stringify(manuscriptJson), w1: validAudioBytes },
        }),
      );

      // When: 音声を取得する
      const got = await repository.getEpisodeAudio("ep-1");

      // Then: byte 一致
      expect(got).toEqual(validAudioBytes);
    });
  });

  describe("files.list の絞り込み q", () => {
    it("getManuscript は files.list へ対象 episodeId の json/wav 名を絞り込む q を渡す", async () => {
      const fetchStub = stubFetch({
        files: [
          { id: "j1", name: "ep-1.json" },
          { id: "w1", name: "ep-1.wav" },
        ],
        downloads: { j1: JSON.stringify(manuscriptJson) },
      });
      const repository = createRepository(fetchStub);

      await repository.getManuscript("ep-1");

      const listCall = vi
        .mocked(fetchStub)
        .mock.calls.find(([input]) =>
          input.startsWith("https://www.googleapis.com/drive/v3/files?"),
        );
      expect(listCall).toBeDefined();
      const query = new URL(listCall?.[0] ?? "").searchParams.get("q") ?? "";
      expect(query).toContain("name = 'ep-1.json'");
      expect(query).toContain("name = 'ep-1.wav'");
    });

    it("getManuscript はフォルダ内の無関係な大量 file を無視し、対象 stem の json+wav だけを見る", async () => {
      const unrelatedFiles = Array.from({ length: 50 }, (_, i) => ({
        id: `unrelated-${i}`,
        name: `other-${i}.json`,
      }));
      const repository = createRepository(
        stubFetch({
          files: [
            ...unrelatedFiles,
            { id: "j1", name: "ep-1.json" },
            { id: "w1", name: "ep-1.wav" },
          ],
          downloads: { j1: JSON.stringify(manuscriptJson) },
        }),
      );

      const got = await repository.getManuscript("ep-1");
      expect(got?.json).toEqual(manuscriptJson);
      expect(got?.hasAudio).toBe(true);
    });

    it("getEpisodeAudio はフォルダ内の無関係な大量 file を無視し、対象 stem の wav だけを見る", async () => {
      const unrelatedFiles = Array.from({ length: 50 }, (_, i) => ({
        id: `unrelated-${i}`,
        name: `other-${i}.wav`,
      }));
      const repository = createRepository(
        stubFetch({
          files: [
            ...unrelatedFiles,
            { id: "j1", name: "ep-1.json" },
            { id: "w1", name: "ep-1.wav" },
          ],
          downloads: { j1: JSON.stringify(manuscriptJson), w1: validAudioBytes },
        }),
      );

      const got = await repository.getEpisodeAudio("ep-1");
      expect(got).toEqual(validAudioBytes);
    });
  });

  describe("複数 json download の並行開始", () => {
    it("1件目の download 応答を待たずに2件目の download を開始する", async () => {
      const pendingResolvers: Array<(body: string) => void> = [];
      const startedFileIds: string[] = [];
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response(JSON.stringify({ access_token: "dummy-access-token" }), {
            status: 200,
          });
        }
        if (input.startsWith("https://www.googleapis.com/drive/v3/files?")) {
          return new Response(
            JSON.stringify({
              files: [
                { id: "j1", name: "ep-1.json" },
                { id: "j2", name: "ep-2.json" },
              ],
            }),
            { status: 200 },
          );
        }
        const downloadMatch =
          /^https:\/\/www\.googleapis\.com\/drive\/v3\/files\/([^?]+)\?alt=media$/.exec(input);
        if (downloadMatch) {
          const fileId = downloadMatch[1] ?? "";
          startedFileIds.push(fileId);
          return new Promise<Response>((resolve) => {
            pendingResolvers.push((body) => resolve(new Response(body, { status: 200 })));
          });
        }
        throw new Error(`Stub 未対応の呼び出し: ${input}`);
      });
      const repository = createRepository(fetchStub);

      const got = repository.listManuscripts();
      await vi.waitFor(() => {
        expect(startedFileIds).toHaveLength(2);
      });
      expect(startedFileIds.sort()).toEqual(["j1", "j2"]);

      pendingResolvers.forEach((resolve) => {
        resolve(JSON.stringify(manuscriptJson));
      });
      await got;
    });
  });

  describe("Drive HTTP 自体の失敗 → DriveError", () => {
    it("token 取得が非 2xx の時、DriveError", async () => {
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response("invalid_grant", { status: 400 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      await expect(createRepository(fetchStub).listManuscripts()).rejects.toBeInstanceOf(
        DriveError,
      );
    });

    it("network error の時、DriveError になり元 error を cause に持つ", async () => {
      const networkError = new Error("network down");
      const fetchStub: FetchLike = vi.fn(async () => {
        throw networkError;
      });
      const act = createRepository(fetchStub).listManuscripts();
      await expect(act).rejects.toBeInstanceOf(DriveError);
      await expect(act).rejects.toHaveProperty("cause", networkError);
    });

    it("token endpoint の応答が不正 JSON の時、DriveError", async () => {
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response("not json", { status: 200 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      await expect(createRepository(fetchStub).listManuscripts()).rejects.toBeInstanceOf(
        DriveError,
      );
    });

    it("token endpoint の応答に access_token が無い時、DriveError", async () => {
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response(JSON.stringify({}), { status: 200 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      await expect(createRepository(fetchStub).listManuscripts()).rejects.toBeInstanceOf(
        DriveError,
      );
    });

    it("files.list が非 2xx の時、DriveError", async () => {
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response(JSON.stringify({ access_token: "dummy-access-token" }), {
            status: 200,
          });
        }
        if (input.startsWith("https://www.googleapis.com/drive/v3/files?")) {
          return new Response("server error", { status: 500 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      await expect(createRepository(fetchStub).listManuscripts()).rejects.toBeInstanceOf(
        DriveError,
      );
    });

    it("files.list 失敗時、Drive file id やフォルダ id を DriveError の message に含めない", async () => {
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response(JSON.stringify({ access_token: "dummy-access-token" }), {
            status: 200,
          });
        }
        if (input.startsWith("https://www.googleapis.com/drive/v3/files?")) {
          return new Response("server error", { status: 500 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      await expect(createRepository(fetchStub).listManuscripts()).rejects.toSatisfy(
        (error: unknown) => error instanceof DriveError && !error.message.includes(dummyFolderId),
      );
    });

    it("files.list の応答が不正 JSON の時、DriveError", async () => {
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response(JSON.stringify({ access_token: "dummy-access-token" }), {
            status: 200,
          });
        }
        if (input.startsWith("https://www.googleapis.com/drive/v3/files?")) {
          return new Response("not json", { status: 200 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      await expect(createRepository(fetchStub).listManuscripts()).rejects.toBeInstanceOf(
        DriveError,
      );
    });

    it("files.list の要素に name が無い時、DriveError", async () => {
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response(JSON.stringify({ access_token: "dummy-access-token" }), {
            status: 200,
          });
        }
        if (input.startsWith("https://www.googleapis.com/drive/v3/files?")) {
          return new Response(JSON.stringify({ files: [{ id: "file-1" }] }), { status: 200 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      await expect(createRepository(fetchStub).listManuscripts()).rejects.toBeInstanceOf(
        DriveError,
      );
    });

    it("files.list の要素の id が number 型の時、DriveError", async () => {
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response(JSON.stringify({ access_token: "dummy-access-token" }), {
            status: 200,
          });
        }
        if (input.startsWith("https://www.googleapis.com/drive/v3/files?")) {
          return new Response(JSON.stringify({ files: [{ id: 1, name: "ep-1.json" }] }), {
            status: 200,
          });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      await expect(createRepository(fetchStub).listManuscripts()).rejects.toBeInstanceOf(
        DriveError,
      );
    });
  });
});
