import type { DriveService, Manuscript, EpisodeMetadata, Episode } from '@daily-it-podcast/core';

const MOCK_EPISODES: Episode[] = [
  {
    metadata: {
      id: 'mock-episode-001',
      timestamp: '2026-01-01T09:00:00.000Z',
      title: '2026年1月1日のITニュース',
    },
    manuscript: {
      timestamp: '2026-01-01T09:00:00.000Z',
      body: {
        opening: '本日のITニュースをお届けします。',
        topics: [
          {
            title: 'TypeScript 5.8 リリース',
            script: 'TypeScript 5.8 が正式リリースされました。',
            durationEstimateSec: 60,
          },
        ],
        closing: '以上、本日のITニュースでした。',
      },
    },
    audioUrl: '/mock-audio/episode-001.mp3',
  },
];

export class MockDriveService implements DriveService {
  private episodes: Episode[] = [...MOCK_EPISODES];

  async save(_audioBuffer: Buffer, manuscript: Manuscript): Promise<string> {
    const id = `mock-episode-${Date.now()}`;
    this.episodes.push({
      metadata: {
        id,
        timestamp: manuscript.timestamp,
        title: `${manuscript.timestamp} のITニュース`,
      },
      manuscript,
      audioUrl: `/mock-audio/${id}.mp3`,
    });
    return id;
  }

  async listEpisodes(): Promise<EpisodeMetadata[]> {
    return this.episodes.map((e) => e.metadata);
  }

  async getEpisode(id: string): Promise<Episode> {
    const episode = this.episodes.find((e) => e.metadata.id === id);
    if (!episode) {
      const { DriveError } = await import('@daily-it-podcast/core');
      throw new DriveError(`episode not found: ${id}`);
    }
    return episode;
  }
}
