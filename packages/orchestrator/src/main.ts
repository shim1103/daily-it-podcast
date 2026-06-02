import { Orchestrator } from './orchestrator.js';
import { createDeps } from './di.js';
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
    infoFetcher: 'hackernews',
    scriptGenerator: 'gemini',
    tts: 'gemini',
    drive: 'google',
  },
  drive: {
    folderId: process.env['DRIVE_FOLDER_ID'] ?? '',
  },
  cron: { enabled: false },
};

const orchestrator = new Orchestrator(config, createDeps(config));

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
