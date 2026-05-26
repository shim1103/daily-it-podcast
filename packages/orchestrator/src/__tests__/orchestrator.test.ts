import { describe, it, expect } from 'vitest';
import { Orchestrator } from '../orchestrator.js';
import { MockInfoFetcher } from '@daily-it-podcast/info-fetcher';
import { MockScriptGenerator } from '@daily-it-podcast/script-generator';
import { MockTtsService } from '@daily-it-podcast/tts';
import { MockDriveService } from '@daily-it-podcast/drive';
import type { PodcastConfig } from '@daily-it-podcast/core';

const testConfig: PodcastConfig = {
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

describe('Orchestrator (end-to-end)', () => {
  it('Given 全システムがモック・設定が有効 When orchestrate() 実行 Then episode ID が返る', async () => {
    const orchestrator = new Orchestrator(testConfig, {
      infoFetcher: new MockInfoFetcher(),
      scriptGenerator: new MockScriptGenerator(),
      ttsService: new MockTtsService(),
      driveService: new MockDriveService(),
    });

    const episodeId = await orchestrator.orchestrate();

    expect(typeof episodeId).toBe('string');
    expect(episodeId.length).toBeGreaterThan(0);
  });
});
