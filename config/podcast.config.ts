import type { PodcastConfig } from '@daily-it-podcast/core';

export const config: PodcastConfig = {
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
