import { describe, it, expect, beforeEach } from 'vitest';
import { MockDriveService } from '../mock.js';
import type { Manuscript } from '@daily-it-podcast/core';

const sampleManuscript: Manuscript = {
  timestamp: '2026-05-26T09:00:00.000Z',
  body: {
    opening: 'オープニング',
    topics: [{ title: 'テーマ', script: '原稿', durationEstimateSec: 60 }],
    closing: 'クロージング',
  },
};

describe('MockDriveService', () => {
  let service: MockDriveService;

  beforeEach(() => {
    service = new MockDriveService();
  });

  it('Given Buffer + Manuscript When save() 実行 Then episode ID が返る', async () => {
    const id = await service.save(Buffer.alloc(0), sampleManuscript);
    expect(typeof id).toBe('string');
    expect(id.length).toBeGreaterThan(0);
  });

  it('Given 保存済みエピソード When listEpisodes() 実行 Then EpisodeMetadata[] が返る', async () => {
    await service.save(Buffer.alloc(0), sampleManuscript);
    const list = await service.listEpisodes();

    expect(Array.isArray(list)).toBe(true);
    expect(list.length).toBeGreaterThan(0);

    for (const meta of list) {
      expect(typeof meta.id).toBe('string');
      expect(typeof meta.timestamp).toBe('string');
      expect(typeof meta.title).toBe('string');
    }
  });

  it('Given 存在するID When getEpisode() 実行 Then Episode が返る', async () => {
    const id = await service.save(Buffer.alloc(0), sampleManuscript);
    const episode = await service.getEpisode(id);

    expect(episode.metadata.id).toBe(id);
    expect(episode.manuscript).toEqual(sampleManuscript);
  });

  it('Given 存在しないID When getEpisode() 実行 Then DriveError が throw される', async () => {
    await expect(service.getEpisode('nonexistent-id')).rejects.toThrow('episode not found');
  });
});
