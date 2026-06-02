import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// drive パッケージのモック（実 API 呼び出しを回避）
vi.mock('@daily-it-podcast/drive', () => {
  class MockDriveService {
    async save() { return 'mock-id'; }
    async listEpisodes() { return []; }
    async getEpisode() { throw new Error('not found'); }
  }
  class GoogleDriveService {
    async save() { return 'google-id'; }
    async listEpisodes() { return []; }
    async getEpisode() { throw new Error('not found'); }
  }
  return { MockDriveService, GoogleDriveService };
});

const REQUIRED_ENV_VARS = {
  GOOGLE_CLIENT_ID: 'test-client-id',
  GOOGLE_CLIENT_SECRET: 'test-client-secret',
  GOOGLE_REFRESH_TOKEN: 'test-refresh-token',
  DRIVE_FOLDER_ID: 'test-folder-id',
} as const;

function setAllEnvVars() {
  Object.assign(process.env, REQUIRED_ENV_VARS);
}

function clearAllEnvVars() {
  for (const key of Object.keys(REQUIRED_ENV_VARS)) {
    delete process.env[key];
  }
}

describe('createDriveService', () => {
  beforeEach(() => {
    clearAllEnvVars();
    vi.resetModules();
  });

  afterEach(() => {
    clearAllEnvVars();
  });

  it('Given 環境変数がすべてセットされている When createDriveService() Then GoogleDriveService が返る', async () => {
    setAllEnvVars();

    const { createDriveService } = await import('../lib/drive.js');
    const { GoogleDriveService } = await import('@daily-it-podcast/drive');

    const service = createDriveService();
    expect(service).toBeInstanceOf(GoogleDriveService);
  });

  it('Given 環境変数がない When createDriveService() Then MockDriveService が返る', async () => {
    const { createDriveService } = await import('../lib/drive.js');
    const { MockDriveService } = await import('@daily-it-podcast/drive');

    const service = createDriveService();
    expect(service).toBeInstanceOf(MockDriveService);
  });

  it('Given 環境変数が一部欠けている（CLIENT_ID のみ）When createDriveService() Then MockDriveService が返る', async () => {
    process.env['GOOGLE_CLIENT_ID'] = 'test-client-id';

    const { createDriveService } = await import('../lib/drive.js');
    const { MockDriveService } = await import('@daily-it-podcast/drive');

    const service = createDriveService();
    expect(service).toBeInstanceOf(MockDriveService);
  });

  it('Given DRIVE_FOLDER_ID のみ欠けている When createDriveService() Then MockDriveService が返る', async () => {
    process.env['GOOGLE_CLIENT_ID'] = REQUIRED_ENV_VARS.GOOGLE_CLIENT_ID;
    process.env['GOOGLE_CLIENT_SECRET'] = REQUIRED_ENV_VARS.GOOGLE_CLIENT_SECRET;
    process.env['GOOGLE_REFRESH_TOKEN'] = REQUIRED_ENV_VARS.GOOGLE_REFRESH_TOKEN;
    // DRIVE_FOLDER_ID はセットしない

    const { createDriveService } = await import('../lib/drive.js');
    const { MockDriveService } = await import('@daily-it-podcast/drive');

    const service = createDriveService();
    expect(service).toBeInstanceOf(MockDriveService);
  });
});
