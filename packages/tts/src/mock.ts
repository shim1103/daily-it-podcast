import type { TtsService, Manuscript } from '@daily-it-podcast/core';

export class MockTtsService implements TtsService {
  async synthesize(_manuscript: Manuscript): Promise<Buffer> {
    return Buffer.alloc(0);
  }
}
