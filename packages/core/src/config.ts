export type SpeakerMode = 'single' | 'dialogue';
export type ApiProviderKey = 'mock' | 'claude' | 'gemini';

export interface InfoSourceConfig {
  enabled: boolean;
}

export interface PodcastConfig {
  duration: { min: number; max: number; target: number };
  speakerMode: SpeakerMode;
  infoSources: {
    manualText: InfoSourceConfig;
    twitter: InfoSourceConfig;
    mastodon: InfoSourceConfig;
    newsFeed: InfoSourceConfig;
  };
  templateKey: string;
  apiProvider: {
    scriptGenerator: ApiProviderKey;
    tts: ApiProviderKey;
    drive?: 'mock' | 'google';
  };
  drive: {
    folderId: string;
  };
  cron: { enabled: boolean };
}
