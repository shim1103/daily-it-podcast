import { z } from "zod";
import type {
  EpisodeListItem,
  EpisodeManuscript,
  EpisodeRepository,
} from "../../application/ports/episode-repository.ts";
import { EpisodeNotFoundError } from "../../entities/errors/episode-not-found-error.ts";
import { DriveError } from "./drive-error.ts";
import { DriveFileEntrySchema } from "./drive-file-entry-schema.ts";
import { ManuscriptSchema } from "./manuscript-schema.ts";

const tokenEndpoint = "https://oauth2.googleapis.com/token";
const driveFilesEndpoint = "https://www.googleapis.com/drive/v3/files";

const jsonExtension = ".json";
const wavExtension = ".wav";

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

type DriveOAuthConfig = {
  clientId: string;
  clientSecret: string;
  refreshToken: string;
};

type GoogleDriveEpisodeRepositoryDeps = {
  fetch: FetchLike;
  oauth: DriveOAuthConfig;
  folderId: string;
};

type DriveFileEntry = {
  id: string;
  name: string;
};

function stemOf(name: string, extension: string): string | undefined {
  if (!name.endsWith(extension)) {
    return undefined;
  }
  return name.slice(0, name.length - extension.length);
}

/**
 * Google Drive REST API（v3）で `EpisodeRepository` を満たす本番 Driven Adapter。
 *
 * @require deps.folderId は Drive 上の対象フォルダ id
 * @ensure Drive HTTP 自体の失敗（token 取得・network・非 2xx）は DriveError、
 *   応答は成功したが内容が不完全・不適合（json 欠落・schema 不適合・stem 不一致・wav 欠落）は EpisodeNotFoundError
 */
export class GoogleDriveEpisodeRepository implements EpisodeRepository {
  private readonly fetch: FetchLike;
  private readonly oauth: DriveOAuthConfig;
  private readonly folderId: string;

  constructor(deps: GoogleDriveEpisodeRepositoryDeps) {
    this.fetch = deps.fetch;
    this.oauth = deps.oauth;
    this.folderId = deps.folderId;
  }

  async listEpisodes(): Promise<EpisodeListItem[]> {
    const accessToken = await this.fetchAccessToken();
    const entries = await this.listFolderEntries(accessToken);
    const jsonEntries = entries.filter((entry) => entry.name.endsWith(jsonExtension));

    const manuscripts = await Promise.all(
      jsonEntries.map((entry) => {
        const stem = stemOf(entry.name, jsonExtension);
        /* v8 ignore next 3 -- jsonEntries は直前で同じ jsonExtension の endsWith 判定を通過済みのため、この分岐は実行時に到達しない */
        if (stem === undefined) {
          return undefined;
        }
        return this.tryDownloadManuscript(accessToken, entry.id, stem);
      }),
    );

    const items: EpisodeListItem[] = [];
    for (const manuscript of manuscripts) {
      if (manuscript === undefined) {
        continue;
      }
      const { body, ...rest } = manuscript;
      items.push({ ...rest, topics: body.topics.map((topic) => ({ title: topic.title })) });
    }
    return items;
  }

  async getEpisode(episodeId: string): Promise<EpisodeManuscript> {
    const accessToken = await this.fetchAccessToken();
    const entries = await this.listEntriesByEpisodeId(accessToken, episodeId);

    const jsonEntry = entries.find((entry) => stemOf(entry.name, jsonExtension) === episodeId);
    if (jsonEntry === undefined) {
      throw new EpisodeNotFoundError(`JSON エントリが無い: ${episodeId}`);
    }
    const wavEntry = entries.find((entry) => stemOf(entry.name, wavExtension) === episodeId);
    if (wavEntry === undefined) {
      throw new EpisodeNotFoundError(`wav が無い: ${episodeId}`);
    }

    const manuscript = await this.tryDownloadManuscript(accessToken, jsonEntry.id, episodeId);
    if (manuscript === undefined) {
      throw new EpisodeNotFoundError(
        `原稿 JSON が schema に不適合または stem 不一致: ${episodeId}`,
      );
    }
    return manuscript;
  }

  async getEpisodeAudio(episodeId: string): Promise<Uint8Array> {
    const accessToken = await this.fetchAccessToken();
    const entries = await this.listEntriesByEpisodeId(accessToken, episodeId);

    const wavEntry = entries.find((entry) => stemOf(entry.name, wavExtension) === episodeId);
    if (wavEntry === undefined) {
      throw new EpisodeNotFoundError(`音声エントリが無い: ${episodeId}`);
    }

    return this.downloadBytes(accessToken, wavEntry.id);
  }

  /**
   * json を download し、schema 適合かつ stem 一致の場合だけ原稿を返す。
   *
   * why: List / Get の両方で「json 内容を読んで schema・stem を判定する」手順が共通するため、
   * 呼び出し側の分岐（List は握りつぶす、Get は EpisodeNotFoundError にする）だけを外側に残す。
   */
  private async tryDownloadManuscript(
    accessToken: string,
    fileId: string,
    stem: string,
  ): Promise<EpisodeManuscript | undefined> {
    const bytes = await this.downloadBytes(accessToken, fileId);
    const text = new TextDecoder().decode(bytes);

    let parsedJson: unknown;
    try {
      parsedJson = JSON.parse(text);
    } catch {
      return undefined;
    }

    const parsed = ManuscriptSchema.safeParse(parsedJson);
    if (!parsed.success || parsed.data.episodeId !== stem) {
      return undefined;
    }
    return parsed.data;
  }

  private async fetchAccessToken(): Promise<string> {
    const body = new URLSearchParams({
      client_id: this.oauth.clientId,
      client_secret: this.oauth.clientSecret,
      refresh_token: this.oauth.refreshToken,
      grant_type: "refresh_token",
    });

    const response = await this.request(tokenEndpoint, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: body.toString(),
    });

    let payload: unknown;
    try {
      payload = await response.json();
    } catch (cause) {
      throw new DriveError("Drive OAuth token 応答の解析に失敗", { cause });
    }

    const accessToken = (payload as { access_token?: unknown } | null)?.access_token;
    if (typeof accessToken !== "string" || accessToken.length === 0) {
      throw new DriveError("Drive OAuth token 応答に access_token が無い");
    }
    return accessToken;
  }

  /**
   * フォルダ直下の全 file を取得する。
   *
   * why: `listEpisodes` は全件が必要な唯一の use case。それ以外（`getEpisode` /
   * `getEpisodeAudio`）は特定 episodeId の json/wav だけが要るため `listEntriesByEpisodeId` を使う。
   */
  private async listFolderEntries(accessToken: string): Promise<DriveFileEntry[]> {
    return this.queryFolderEntries(
      accessToken,
      `'${this.folderId}' in parents and trashed = false`,
    );
  }

  /**
   * フォルダ直下から、対象 episodeId の json/wav 名だけへ絞り込んで file を取得する。
   *
   * why: `getEpisode` / `getEpisodeAudio` は1件の episodeId だけを要求されるため、
   * Drive API v3 の `q` へ name 条件を足すことで、フォルダ内 file 数に関わらず応答を定数サイズにする。
   */
  private async listEntriesByEpisodeId(
    accessToken: string,
    episodeId: string,
  ): Promise<DriveFileEntry[]> {
    const jsonName = `${episodeId}${jsonExtension}`;
    const wavName = `${episodeId}${wavExtension}`;
    const query =
      `'${this.folderId}' in parents and trashed = false ` +
      `and (name = '${jsonName}' or name = '${wavName}')`;
    return this.queryFolderEntries(accessToken, query);
  }

  private async queryFolderEntries(accessToken: string, q: string): Promise<DriveFileEntry[]> {
    const query = new URLSearchParams({
      q,
      fields: "files(id,name)",
    });

    const response = await this.request(`${driveFilesEndpoint}?${query.toString()}`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    let payload: unknown;
    try {
      payload = await response.json();
    } catch (cause) {
      throw new DriveError("Drive 一覧応答の解析に失敗", { cause });
    }

    const files = (payload as { files?: unknown } | null)?.files;
    const parsed = z.array(DriveFileEntrySchema).safeParse(files);
    if (!parsed.success) {
      throw new DriveError("Drive 一覧応答の形式が不正");
    }
    return parsed.data;
  }

  private async downloadBytes(accessToken: string, fileId: string): Promise<Uint8Array> {
    const response = await this.request(
      `${driveFilesEndpoint}/${encodeURIComponent(fileId)}?alt=media`,
      { headers: { Authorization: `Bearer ${accessToken}` } },
    );
    const buffer = await response.arrayBuffer();
    return new Uint8Array(buffer);
  }

  /**
   * Drive HTTP 呼び出しの共通経路。network error・非 2xx をここで DriveError へ畳む。
   *
   * @require input・init は fetch と同じ引数
   * @ensure 戻り値は必ず 2xx。Drive file id やフォルダ id を message に含めない
   */
  private async request(input: string, init?: RequestInit): Promise<Response> {
    let response: Response;
    try {
      response = await this.fetch(input, init);
    } catch (cause) {
      throw new DriveError("Drive HTTP 呼び出しに失敗", { cause });
    }
    if (!response.ok) {
      throw new DriveError(`Drive HTTP 呼び出しが非 2xx: ${response.status}`);
    }
    return response;
  }
}
