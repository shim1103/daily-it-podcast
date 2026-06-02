import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { Manuscript } from '@daily-it-podcast/core';

const mockFilesCreate = vi.fn();
const mockFilesList = vi.fn();
const mockFilesGet = vi.fn();
const mockFilesGetMedia = vi.fn();

vi.mock('googleapis', () => ({
  google: {
    auth: {
      OAuth2: vi.fn().mockImplementation(() => ({
        setCredentials: vi.fn(),
      })),
    },
    drive: vi.fn().mockReturnValue({
      files: {
        create: mockFilesCreate,
        list: mockFilesList,
        get: mockFilesGet,
      },
    }),
  },
}));

const sampleManuscript: Manuscript = {
  timestamp: '2026-01-01T09:00:00.000Z',
  body: {
    opening: 'オープニング',
    topics: [{ title: 'テーマ', script: '原稿', durationEstimateSec: 60 }],
    closing: 'クロージング',
  },
};

describe('GoogleDriveService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    process.env['GOOGLE_CLIENT_ID'] = 'test-client-id';
    process.env['GOOGLE_CLIENT_SECRET'] = 'test-client-secret';
    process.env['GOOGLE_REFRESH_TOKEN'] = 'test-refresh-token';
    process.env['DRIVE_FOLDER_ID'] = 'test-folder-id';
  });

  it('Given 環境変数なし When インスタンス生成 Then DriveError が throw される', async () => {
    delete process.env['GOOGLE_CLIENT_ID'];
    const { GoogleDriveService } = await import('../google-drive.js');
    const { DriveError } = await import('@daily-it-podcast/core');
    expect(() => new GoogleDriveService()).toThrow(DriveError);
  });

  it('Given Buffer + Manuscript When save() 実行 Then episode ID が返る', async () => {
    mockFilesCreate
      .mockResolvedValueOnce({ data: { id: 'audio-file-id' } })
      .mockResolvedValueOnce({ data: { id: 'meta-file-id' } });

    const { GoogleDriveService } = await import('../google-drive.js');
    const service = new GoogleDriveService();
    const id = await service.save(Buffer.alloc(16), sampleManuscript);

    expect(typeof id).toBe('string');
    expect(id).toBe('audio-file-id');
    expect(mockFilesCreate).toHaveBeenCalledTimes(2);
  });

  it('Given フォルダに音声ファイル When listEpisodes() 実行 Then EpisodeMetadata[] が返る', async () => {
    mockFilesList.mockResolvedValueOnce({
      data: {
        files: [
          {
            id: 'file-1',
            name: '2026-01-01T09:00:00.000Z.mp3',
            description: '2026年1月1日のITニュース',
          },
        ],
      },
    });

    const { GoogleDriveService } = await import('../google-drive.js');
    const service = new GoogleDriveService();
    const list = await service.listEpisodes();

    expect(Array.isArray(list)).toBe(true);
    expect(list).toHaveLength(1);
    expect(list[0]!.id).toBe('file-1');
    expect(list[0]!.timestamp).toBe('2026-01-01T09:00:00.000Z');
  });

  it('Given 存在するファイルID When getEpisode() 実行 Then Episode が返る', async () => {
    mockFilesGet.mockResolvedValueOnce({
      data: {
        id: 'file-1',
        name: '2026-01-01T09:00:00.000Z.mp3',
        description: '2026年1月1日のITニュース',
        webContentLink: 'https://drive.google.com/uc?id=file-1',
        appProperties: {
          manuscript: JSON.stringify(sampleManuscript),
        },
      },
    });

    const { GoogleDriveService } = await import('../google-drive.js');
    const service = new GoogleDriveService();
    const episode = await service.getEpisode('file-1');

    expect(episode.metadata.id).toBe('file-1');
    expect(episode.audioUrl).toBe('https://drive.google.com/uc?id=file-1');
    expect(episode.manuscript).toEqual(sampleManuscript);
  });

  it('Given 存在しないID When getEpisode() 実行 Then DriveError が throw される', async () => {
    mockFilesGet.mockResolvedValueOnce({ data: {} });

    const { GoogleDriveService } = await import('../google-drive.js');
    const { DriveError } = await import('@daily-it-podcast/core');
    const service = new GoogleDriveService();

    await expect(service.getEpisode('nonexistent')).rejects.toThrow(DriveError);
  });
});
