import { Orchestrator } from './orchestrator.js';
import { MockInfoFetcher } from '@daily-it-podcast/info-fetcher';
import { MockScriptGenerator } from '@daily-it-podcast/script-generator';
import { MockTtsService } from '@daily-it-podcast/tts';
import { MockDriveService } from '@daily-it-podcast/drive';
import type { PodcastConfig } from '@daily-it-podcast/core';

const config: PodcastConfig = {
  duration: { min: 5, max: 30, target: 5 },
  speakerMode: 'single',
  infoSources: {
    manualText: { enabled: true },
    twitter: { enabled: false },
    mastodon: { enabled: false },
    newsFeed: { enabled: false },
  },
  templateKey: 'default',
  apiProvider: {
    scriptGenerator: 'mock',
    tts: 'mock',
  },
  drive: {
    folderId: process.env['DRIVE_FOLDER_ID'] ?? '',
  },
  cron: { enabled: false },
};

const orchestrator = new Orchestrator(config, {
  infoFetcher: new MockInfoFetcher(),
  scriptGenerator: new MockScriptGenerator(),
  ttsService: new MockTtsService(),
  driveService: new MockDriveService(),
});

orchestrator
  .orchestrate()
  .then((id) => {
    console.log('[done] episode id:', id);
    process.exit(0);
  })
  .catch((err: unknown) => {
    console.error('[error]', err);
    process.exit(1);
  });
