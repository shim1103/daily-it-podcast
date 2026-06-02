import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { PodcastConfig } from '@daily-it-podcast/core';

vi.mock('@daily-it-podcast/tts', () => ({
  MockTtsService: vi.fn().mockImplementation(function MockTtsService() {
    return { synthesize: vi.fn(), _name: 'MockTtsService' };
  }),
  GeminiTtsService: vi.fn().mockImplementation(function GeminiTtsService() {
    return { synthesize: vi.fn(), _name: 'GeminiTtsService' };
  }),
}));

vi.mock('@daily-it-podcast/drive', () => ({
  MockDriveService: vi.fn().mockImplementation(function MockDriveService() {
    return { save: vi.fn(), listEpisodes: vi.fn(), getEpisode: vi.fn(), _name: 'MockDriveService' };
  }),
  GoogleDriveService: vi.fn().mockImplementation(function GoogleDriveService() {
    return { save: vi.fn(), listEpisodes: vi.fn(), getEpisode: vi.fn(), _name: 'GoogleDriveService' };
  }),
}));

vi.mock('@daily-it-podcast/info-fetcher', () => ({
  MockInfoFetcher: vi.fn().mockImplementation(() => ({ fetch: vi.fn() })),
}));

vi.mock('@daily-it-podcast/script-generator', () => ({
  MockScriptGenerator: vi.fn().mockImplementation(() => ({ generate: vi.fn() })),
}));

const baseConfig: PodcastConfig = {
  duration: { min: 5, max: 30, target: 5 },
  speakerMode: 'single',
  infoSources: {
    manualText: { enabled: true },
    twitter: { enabled: false },
    mastodon: { enabled: false },
    newsFeed: { enabled: false },
  },
  templateKey: 'default',
  apiProvider: { scriptGenerator: 'mock', tts: 'mock' },
  drive: { folderId: 'test-folder' },
  cron: { enabled: false },
};

describe('createDeps', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('Given tts=mock When createDeps() 実行 Then MockTtsService が選択される', async () => {
    const { createDeps } = await import('../di.js');
    const { MockTtsService } = await import('@daily-it-podcast/tts');

    createDeps(baseConfig);

    expect(MockTtsService).toHaveBeenCalledTimes(1);
  });

  it('Given tts=gemini When createDeps() 実行 Then GeminiTtsService が選択される', async () => {
    const { createDeps } = await import('../di.js');
    const { GeminiTtsService } = await import('@daily-it-podcast/tts');

    createDeps({
      ...baseConfig,
      apiProvider: { scriptGenerator: 'mock', tts: 'gemini' },
    });

    expect(GeminiTtsService).toHaveBeenCalledTimes(1);
  });

  it('Given drive=mock (default) When createDeps() 実行 Then MockDriveService が選択される', async () => {
    const { createDeps } = await import('../di.js');
    const { MockDriveService } = await import('@daily-it-podcast/drive');

    createDeps(baseConfig);

    expect(MockDriveService).toHaveBeenCalledTimes(1);
  });

  it('Given drive=google When createDeps() 実行 Then GoogleDriveService が選択される', async () => {
    const { createDeps } = await import('../di.js');
    const { GoogleDriveService } = await import('@daily-it-podcast/drive');

    createDeps({
      ...baseConfig,
      apiProvider: { scriptGenerator: 'mock', tts: 'mock', drive: 'google' },
    });

    expect(GoogleDriveService).toHaveBeenCalledTimes(1);
  });
});
