import { google } from 'googleapis';
import { Readable } from 'node:stream';
import type { DriveService, Manuscript, EpisodeMetadata, Episode } from '@daily-it-podcast/core';
import { DriveError } from '@daily-it-podcast/core';

function buildTitle(timestamp: string): string {
  return `${timestamp} のITニュース`;
}

export class GoogleDriveService implements DriveService {
  private readonly drive;
  private readonly folderId: string;

  constructor() {
    const clientId = process.env['GOOGLE_CLIENT_ID'];
    const clientSecret = process.env['GOOGLE_CLIENT_SECRET'];
    const refreshToken = process.env['GOOGLE_REFRESH_TOKEN'];
    const folderId = process.env['DRIVE_FOLDER_ID'];

    if (!clientId || !clientSecret || !refreshToken || !folderId) {
      throw new DriveError(
        'Drive 認証に必要な環境変数が不足しています (GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REFRESH_TOKEN, DRIVE_FOLDER_ID)',
      );
    }

    const auth = new google.auth.OAuth2(clientId, clientSecret);
    auth.setCredentials({ refresh_token: refreshToken });

    this.drive = google.drive({ version: 'v3', auth });
    this.folderId = folderId;
  }

  async save(audioBuffer: Buffer, manuscript: Manuscript): Promise<string> {
    const fileName = `${manuscript.timestamp}.mp3`;
    const title = buildTitle(manuscript.timestamp);

    let audioRes;
    try {
      audioRes = await this.drive.files.create({
        requestBody: {
          name: fileName,
          description: title,
          parents: [this.folderId],
          mimeType: 'audio/mpeg',
        },
        media: {
          mimeType: 'audio/mpeg',
          body: Readable.from(audioBuffer),
        },
        fields: 'id',
      });
    } catch (err) {
      throw new DriveError('音声ファイルのアップロードに失敗しました', err);
    }

    const fileId = audioRes.data.id;
    if (!fileId) {
      throw new DriveError('Drive からファイル ID が返りませんでした');
    }

    let jsonRes;
    try {
      jsonRes = await this.drive.files.create({
        requestBody: {
          name: `${manuscript.timestamp}.json`,
          description: title,
          parents: [this.folderId],
          mimeType: 'application/json',
        },
        media: {
          mimeType: 'application/json',
          body: Readable.from(Buffer.from(JSON.stringify(manuscript))),
        },
        fields: 'id',
      });
    } catch (err) {
      throw new DriveError('原稿ファイルのアップロードに失敗しました', err);
    }

    const jsonFileId = jsonRes.data.id;
    if (!jsonFileId) {
      throw new DriveError('原稿ファイルの ID が取得できませんでした');
    }

    try {
      await this.drive.files.update({
        fileId,
        requestBody: {
          appProperties: { jsonFileId },
        },
      });
    } catch (err) {
      throw new DriveError('音声ファイルへの原稿 ID 紐付けに失敗しました', err);
    }

    return fileId;
  }

  async listEpisodes(): Promise<EpisodeMetadata[]> {
    let res;
    try {
      res = await this.drive.files.list({
        q: `'${this.folderId}' in parents and mimeType='audio/mpeg' and trashed=false`,
        fields: 'files(id, name, description)',
        orderBy: 'createdTime desc',
      });
    } catch (err) {
      throw new DriveError('エピソード一覧の取得に失敗しました', err);
    }

    const files = res.data.files ?? [];
    return files.map((f) => ({
      id: f.id ?? '',
      timestamp: (f.name ?? '').replace('.mp3', ''),
      title: f.description ?? buildTitle((f.name ?? '').replace('.mp3', '')),
    }));
  }

  async getEpisode(id: string): Promise<Episode> {
    let audioRes;
    try {
      audioRes = await this.drive.files.get({
        fileId: id,
        fields: 'id, name, description, webContentLink, appProperties',
      });
    } catch (err) {
      throw new DriveError(`エピソードの取得に失敗しました: ${id}`, err);
    }

    const file = audioRes.data;
    if (!file.id || !file.webContentLink) {
      throw new DriveError(`episode not found: ${id}`);
    }

    const jsonFileId = file.appProperties?.jsonFileId;
    if (!jsonFileId) {
      throw new DriveError(`episode ${id} の原稿ファイル ID が見つかりません`);
    }

    let jsonRes;
    try {
      jsonRes = await this.drive.files.get({
        fileId: jsonFileId,
        alt: 'media',
      });
    } catch (err) {
      throw new DriveError(`原稿ファイルの取得に失敗しました: ${jsonFileId}`, err);
    }

    let manuscript: Manuscript;
    try {
      const raw =
        typeof jsonRes.data === 'string' ? jsonRes.data : JSON.stringify(jsonRes.data);
      manuscript = JSON.parse(raw) as Manuscript;
    } catch {
      throw new DriveError(`episode ${id} の manuscript データが不正です`);
    }

    return {
      metadata: {
        id: file.id,
        timestamp: (file.name ?? '').replace('.mp3', ''),
        title: file.description ?? buildTitle((file.name ?? '').replace('.mp3', '')),
      },
      manuscript,
      audioUrl: file.webContentLink,
    };
  }
}
