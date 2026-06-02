import { MockInfoFetcher, HackerNewsInfoFetcher } from '@daily-it-podcast/info-fetcher';
import { MockScriptGenerator, GeminiScriptGenerator } from '@daily-it-podcast/script-generator';
import { MockTtsService, GeminiTtsService } from '@daily-it-podcast/tts';
import { MockDriveService, GoogleDriveService } from '@daily-it-podcast/drive';
import type { PodcastConfig } from '@daily-it-podcast/core';
import type { OrchestratorDeps } from './orchestrator.js';

export function createDeps(config: PodcastConfig): OrchestratorDeps {
  const ttsService =
    config.apiProvider.tts === 'gemini' ? new GeminiTtsService() : new MockTtsService();

  const driveService =
    config.apiProvider.drive === 'google' ? new GoogleDriveService() : new MockDriveService();

  const infoFetcher =
    config.apiProvider.infoFetcher === 'hackernews'
      ? new HackerNewsInfoFetcher({ maxItems: 5 })
      : new MockInfoFetcher();

  const scriptGenerator =
    config.apiProvider.scriptGenerator === 'gemini'
      ? new GeminiScriptGenerator()
      : new MockScriptGenerator();

  return {
    infoFetcher,
    scriptGenerator,
    ttsService,
    driveService,
  };
}
