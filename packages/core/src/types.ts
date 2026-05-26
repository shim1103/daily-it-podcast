export interface ThemeInfo {
  source: string;
  title: string;
  rawText: string;
  fetchedAt: string;
}

export interface ThemeScript {
  title: string;
  script: string;
  durationEstimateSec: number;
}

export interface Manuscript {
  timestamp: string;
  body: {
    opening: string;
    topics: ThemeScript[];
    closing: string;
  };
}

export interface EpisodeMetadata {
  id: string;
  timestamp: string;
  title: string;
}

export interface Episode {
  metadata: EpisodeMetadata;
  manuscript: Manuscript;
  audioUrl: string;
}
