import { describe, expect, it, vi } from "vitest";
import { EpisodeNotFoundError } from "../../entities/errors/episode-not-found-error.ts";
import { DriveError } from "./drive-error.ts";
import { GoogleDriveEpisodeRepository } from "./google-drive-episode-repository.ts";
import { ManuscriptSchema } from "./manuscript-schema.ts";

/**
 * Drive HTTP を Stub 化した `fetch` 相当関数。
 *
 * why: OAuth token 取得と Drive REST 呼び出しはどちらも `fetch` 経由の外部 I/O であり、
 * Adapter 内部のロジック（token 取得 → 一覧 → 内容取得の組み立て）を実物のまま検証しつつ、
 * 実 Drive への通信だけを断つには、この境界を Stub にするのが最小になる。
 */
type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

const dummyOAuthConfig = {
  clientId: "dummy-client-id",
  clientSecret: "dummy-client-secret",
  refreshToken: "dummy-refresh-token",
};
const dummyFolderId = "dummy-folder-id";

const validManuscript = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [
      {
        title: "題",
        preface: "前置き",
        detail: "詳細",
        startSec: 0,
      },
    ],
    closing: "終了",
  },
};

const audioBytes = new Uint8Array([
  0x52, 0x49, 0x46, 0x46, 0x04, 0x00, 0x00, 0x00, 0x57, 0x41, 0x56, 0x45,
]);

type DriveFileEntry = { id: string; name: string };

/**
 * `files.list` に渡された `q` から `name = '...'` 条件を抽出する。
 *
 * why: 絞り込み query が実際に Drive へ送られていることを、
 * Stub 側で名前一致のみへ応答を絞ることで実証する。
 */
function extractNameFilters(query: string): string[] | undefined {
  const matches = [...query.matchAll(/name = '([^']*)'/g)];
  if (matches.length === 0) {
    return undefined;
  }
  return matches.map((match) => match[1] ?? "");
}

/**
 * token取得と files.list を成功させ、files.list の一覧結果だけ差し替え可能な Stub を作る。
 *
 * why: 一覧・1件・音声の各 test で「フォルダ直下に何が置かれているか」だけを変えたいため、
 * token 取得成功という共通前提を test file 内で重複定義しない。
 *
 * files.list への `q` に `name = '...'` 条件がある時は、名前一致分だけへ応答を絞る。
 * 絞り込み query が実際に送られていることと、余分な file が応答に混ざらないことの両方を検証できる。
 */
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

describe("GoogleDriveEpisodeRepository", () => {
  describe("listEpisodes", () => {
    it("schema 適合 JSON は一覧に出る", async () => {
      // Given: フォルダ直下に適合 JSON + 対応 wav がある
      const fetchStub = stubFetch({
        files: [
          { id: "file-json-1", name: "ep-1.json" },
          { id: "file-wav-1", name: "ep-1.wav" },
        ],
        downloads: { "file-json-1": JSON.stringify(validManuscript) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const got = await repository.listEpisodes();

      // Then: 適合分が出る
      expect(got).toHaveLength(1);
      expect(got[0]?.episodeId).toBe("ep-1");
    });

    it("schema 不適合 JSON は一覧に出ない", async () => {
      // Given: フォルダ直下に不適合 JSON がある
      const fetchStub = stubFetch({
        files: [{ id: "file-json-1", name: "bad.json" }],
        downloads: { "file-json-1": JSON.stringify({ episodeId: "bad" }) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const got = await repository.listEpisodes();

      // Then: 出ない
      expect(got).toHaveLength(0);
    });

    it("download した json 自体が不正 JSON の件は一覧に出ない", async () => {
      // Given: フォルダ直下の json download が JSON として解析できない
      const fetchStub = stubFetch({
        files: [{ id: "file-json-1", name: "ep-1.json" }],
        downloads: { "file-json-1": "not json" },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const got = await repository.listEpisodes();

      // Then: 出ない
      expect(got).toHaveLength(0);
    });

    it("stem と JSON 内 episodeId が不一致の件は一覧に出ない", async () => {
      // Given: stem が ep-1、中身の episodeId が別物
      const fetchStub = stubFetch({
        files: [{ id: "file-json-1", name: "ep-1.json" }],
        downloads: {
          "file-json-1": JSON.stringify({ ...validManuscript, episodeId: "ep-other" }),
        },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const got = await repository.listEpisodes();

      // Then: 出ない
      expect(got).toHaveLength(0);
    });

    it("複数 json の download を並行に開始する", async () => {
      // Given: フォルダ直下に json が2件あり、両方の download 応答をまだ返さない
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
                { id: "file-json-1", name: "ep-1.json" },
                { id: "file-json-2", name: "ep-2.json" },
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
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する（download 応答は未解決のまま呼び出しだけ観測する）
      const got = repository.listEpisodes();
      await vi.waitFor(() => {
        expect(startedFileIds).toHaveLength(2);
      });

      // Then: 1件目の応答を待たずに2件目の download が開始されている
      expect(startedFileIds.sort()).toEqual(["file-json-1", "file-json-2"]);

      pendingResolvers.forEach((resolve) => {
        resolve(JSON.stringify({ ...validManuscript, episodeId: "ep-1" }));
      });
      await got;
    });

    it("音声が無い json のみの件も一覧には出る", async () => {
      // Given: 音声の有無を一覧は見ない契約（drive-layout.md）
      const fetchStub = stubFetch({
        files: [{ id: "file-json-1", name: "ep-1.json" }],
        downloads: { "file-json-1": JSON.stringify(validManuscript) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const got = await repository.listEpisodes();

      // Then: 音声不在でも出る
      expect(got).toHaveLength(1);
    });
  });

  describe("getEpisode", () => {
    it("json のみ（wav 無し）は EpisodeNotFoundError になる", async () => {
      // Given: 対応 wav が無い
      const fetchStub = stubFetch({
        files: [{ id: "file-json-1", name: "ep-1.json" }],
        downloads: { "file-json-1": JSON.stringify(validManuscript) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 1件取得する
      const act = repository.getEpisode("ep-1");

      // Then: Domain 不在
      await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
    });

    it("json が無い（stem 該当なし）は EpisodeNotFoundError になる", async () => {
      // Given: フォルダが空
      const fetchStub = stubFetch({ files: [] });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 1件取得する
      const act = repository.getEpisode("missing");

      // Then: Domain 不在
      await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
    });

    it("schema 不適合 JSON は json+wav が揃っていても EpisodeNotFoundError になる", async () => {
      // Given: json は schema 不適合、wav は存在する
      const fetchStub = stubFetch({
        files: [
          { id: "file-json-1", name: "ep-1.json" },
          { id: "file-wav-1", name: "ep-1.wav" },
        ],
        downloads: { "file-json-1": JSON.stringify({ episodeId: "ep-1" }) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 1件取得する
      const act = repository.getEpisode("ep-1");

      // Then: Domain 不在
      await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
    });

    it("stem と JSON 内 episodeId が不一致の件は EpisodeNotFoundError になる", async () => {
      // Given: stem が ep-1、中身の episodeId が別物、wav はある
      const fetchStub = stubFetch({
        files: [
          { id: "file-json-1", name: "ep-1.json" },
          { id: "file-wav-1", name: "ep-1.wav" },
        ],
        downloads: {
          "file-json-1": JSON.stringify({ ...validManuscript, episodeId: "ep-other" }),
        },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 1件取得する
      const act = repository.getEpisode("ep-1");

      // Then: Domain 不在
      await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
    });

    it("json+wav が揃い schema 適合の時、返却原稿が manuscript schema に適合する", async () => {
      // Given: 適合 json + 対応 wav
      const fetchStub = stubFetch({
        files: [
          { id: "file-json-1", name: "ep-1.json" },
          { id: "file-wav-1", name: "ep-1.wav" },
        ],
        downloads: { "file-json-1": JSON.stringify(validManuscript) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 1件取得する
      const got = await repository.getEpisode("ep-1");

      // Then: manuscript schema 適合
      expect(ManuscriptSchema.safeParse(got).success).toBe(true);
    });
  });

  describe("getEpisodeAudio", () => {
    it("json+wav が揃う時、wav byte が取れる", async () => {
      // Given: 適合 json + 対応 wav
      const fetchStub = stubFetch({
        files: [
          { id: "file-json-1", name: "ep-1.json" },
          { id: "file-wav-1", name: "ep-1.wav" },
        ],
        downloads: {
          "file-json-1": JSON.stringify(validManuscript),
          "file-wav-1": audioBytes,
        },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 音声を取得する
      const got = await repository.getEpisodeAudio("ep-1");

      // Then: byte が一致する
      expect(got).toEqual(audioBytes);
    });

    it("wav が無い（json のみ）は EpisodeNotFoundError になる", async () => {
      // Given: 対応 wav が無い
      const fetchStub = stubFetch({
        files: [{ id: "file-json-1", name: "ep-1.json" }],
        downloads: { "file-json-1": JSON.stringify(validManuscript) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 音声を取得する
      const act = repository.getEpisodeAudio("ep-1");

      // Then: Domain 不在
      await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
    });

    it("json 自体が無い（stem 該当なし）は EpisodeNotFoundError になる", async () => {
      // Given: フォルダが空
      const fetchStub = stubFetch({ files: [] });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 音声を取得する
      const act = repository.getEpisodeAudio("missing");

      // Then: Domain 不在
      await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
    });
  });

  describe("getEpisode / getEpisodeAudio の絞り込み取得", () => {
    it("getEpisode はフォルダ内の無関係な file を無視し、対象 stem の json+wav だけを見る", async () => {
      // Given: フォルダに無関係な大量の file と、対象の json+wav が混在する
      const unrelatedFiles = Array.from({ length: 50 }, (_, i) => ({
        id: `unrelated-${i}`,
        name: `other-${i}.json`,
      }));
      const fetchStub = stubFetch({
        files: [
          ...unrelatedFiles,
          { id: "file-json-1", name: "ep-1.json" },
          { id: "file-wav-1", name: "ep-1.wav" },
        ],
        downloads: { "file-json-1": JSON.stringify(validManuscript) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 1件取得する
      const got = await repository.getEpisode("ep-1");

      // Then: 絞り込み経由でも対象が正しく取れる
      expect(got.episodeId).toBe("ep-1");
    });

    it("getEpisode は files.list へ対象 episodeId の json/wav 名を絞り込む q を渡す", async () => {
      // Given: files.list への呼び出しを観測する
      const fetchStub = stubFetch({
        files: [
          { id: "file-json-1", name: "ep-1.json" },
          { id: "file-wav-1", name: "ep-1.wav" },
        ],
        downloads: { "file-json-1": JSON.stringify(validManuscript) },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 1件取得する
      await repository.getEpisode("ep-1");

      // Then: files.list へ渡された q が対象 episodeId の名前だけへ絞られている
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

    it("getEpisodeAudio はフォルダ内の無関係な file を無視し、対象 stem の wav だけを見る", async () => {
      // Given: フォルダに無関係な大量の file と、対象の json+wav が混在する
      const unrelatedFiles = Array.from({ length: 50 }, (_, i) => ({
        id: `unrelated-${i}`,
        name: `other-${i}.wav`,
      }));
      const fetchStub = stubFetch({
        files: [
          ...unrelatedFiles,
          { id: "file-json-1", name: "ep-1.json" },
          { id: "file-wav-1", name: "ep-1.wav" },
        ],
        downloads: {
          "file-json-1": JSON.stringify(validManuscript),
          "file-wav-1": audioBytes,
        },
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 音声を取得する
      const got = await repository.getEpisodeAudio("ep-1");

      // Then: 絞り込み経由でも対象が正しく取れる
      expect(got).toEqual(audioBytes);
    });
  });

  describe("Drive HTTP 自体の失敗", () => {
    it("token 取得が失敗する時、DriveError になる", async () => {
      // Given: token endpoint が非 2xx を返す
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response("invalid_grant", { status: 400 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: Drive HTTP 自体の失敗
      await expect(act).rejects.toBeInstanceOf(DriveError);
    });

    it("network error が起きる時、DriveError になり元 error を cause に持つ", async () => {
      // Given: fetch 自体が reject する
      const networkError = new Error("network down");
      const fetchStub: FetchLike = vi.fn(async () => {
        throw networkError;
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: DriveError へ変換され cause を保持する
      await expect(act).rejects.toBeInstanceOf(DriveError);
      await expect(act).rejects.toHaveProperty("cause", networkError);
    });

    it("token endpoint の応答が不正 JSON の時、DriveError になる", async () => {
      // Given: token endpoint が 2xx だが body が JSON として解析できない
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response("not json", { status: 200 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: 応答解析の失敗として DriveError になる
      await expect(act).rejects.toBeInstanceOf(DriveError);
    });

    it("token endpoint の応答に access_token が無い時、DriveError になる", async () => {
      // Given: token endpoint は 2xx で JSON も解析できるが access_token が無い
      const fetchStub: FetchLike = vi.fn(async (input: string) => {
        if (input === "https://oauth2.googleapis.com/token") {
          return new Response(JSON.stringify({}), { status: 200 });
        }
        throw new Error(`想定外の呼び出し: ${input}`);
      });
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: token 欠落として DriveError になる
      await expect(act).rejects.toBeInstanceOf(DriveError);
    });

    it("files.list が非 2xx を返す時、DriveError になる", async () => {
      // Given: token 取得は成功、一覧取得が失敗
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
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: Drive HTTP 自体の失敗
      await expect(act).rejects.toBeInstanceOf(DriveError);
    });

    it("Drive file id やフォルダ id を DriveError の message に含めない", async () => {
      // Given: フォルダ id を含む状態で files.list が失敗する
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
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: message にフォルダ id が漏れない
      await expect(act).rejects.toSatisfy((error: unknown) => {
        return error instanceof DriveError && !error.message.includes(dummyFolderId);
      });
    });

    it("files.list の応答が不正 JSON の時、DriveError になる", async () => {
      // Given: token 取得は成功、files.list が 2xx だが body が JSON として解析できない
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
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: 応答解析の失敗として DriveError になる
      await expect(act).rejects.toBeInstanceOf(DriveError);
    });

    it("files.list の要素に name が無い時、DriveError になる", async () => {
      // Given: files 配列は返るが要素の name が欠落している
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
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: Drive 応答自体の形式不正
      await expect(act).rejects.toBeInstanceOf(DriveError);
    });

    it("files.list の要素の id が number 型の時、DriveError になる", async () => {
      // Given: files 配列は返るが要素の id が string ではない
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
      const repository = new GoogleDriveEpisodeRepository({
        fetch: fetchStub,
        oauth: dummyOAuthConfig,
        folderId: dummyFolderId,
      });

      // When: 一覧を取得する
      const act = repository.listEpisodes();

      // Then: Drive 応答自体の形式不正
      await expect(act).rejects.toBeInstanceOf(DriveError);
    });
  });
});
