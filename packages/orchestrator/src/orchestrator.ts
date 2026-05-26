import type {
  InfoFetcher,
  ScriptGenerator,
  TtsService,
  DriveService,
  PodcastConfig,
} from '@daily-it-podcast/core';
import { DefaultManuscriptBuilder } from '@daily-it-podcast/manuscript-builder';

export interface OrchestratorDeps {
  infoFetcher: InfoFetcher;
  scriptGenerator: ScriptGenerator;
  ttsService: TtsService;
  driveService: DriveService;
}

export class Orchestrator {
  constructor(
    private readonly config: PodcastConfig,
    private readonly deps: OrchestratorDeps,
  ) {}

  async orchestrate(): Promise<string> {
    console.log('[orchestrator] starting with config:', JSON.stringify(this.config, null, 2));

    const builder = new DefaultManuscriptBuilder(
      this.deps.infoFetcher,
      this.deps.scriptGenerator,
    );

    console.log('[orchestrator] building manuscript...');
    const manuscript = await builder.build();
    console.log('[orchestrator] manuscript built:', manuscript.timestamp);

    console.log('[orchestrator] synthesizing audio...');
    const audioBuffer = await this.deps.ttsService.synthesize(manuscript);
    console.log('[orchestrator] audio synthesized, size:', audioBuffer.length);

    console.log('[orchestrator] saving to drive...');
    const episodeId = await this.deps.driveService.save(audioBuffer, manuscript);
    console.log('[orchestrator] saved, episode id:', episodeId);

    return episodeId;
  }
}
