import type { ThemeInfo, ThemeScript, Manuscript, EpisodeMetadata, Episode } from './types.js';

export interface InfoFetcher {
  fetch(): Promise<ThemeInfo[]>;
}

export interface ScriptGenerator {
  generate(info: ThemeInfo): Promise<ThemeScript>;
}

export interface ManuscriptBuilder {
  build(): Promise<Manuscript>;
}

export interface TtsService {
  synthesize(manuscript: Manuscript): Promise<Buffer>;
}

export interface DriveService {
  save(audioBuffer: Buffer, manuscript: Manuscript): Promise<string>;
  listEpisodes(): Promise<EpisodeMetadata[]>;
  getEpisode(id: string): Promise<Episode>;
}
