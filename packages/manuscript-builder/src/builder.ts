import type {
  InfoFetcher,
  ScriptGenerator,
  ManuscriptBuilder,
  Manuscript,
} from '@daily-it-podcast/core';

export class DefaultManuscriptBuilder implements ManuscriptBuilder {
  constructor(
    private readonly fetcher: InfoFetcher,
    private readonly generator: ScriptGenerator,
  ) {}

  async build(): Promise<Manuscript> {
    const themes = await this.fetcher.fetch();
    const topics = await Promise.all(themes.map((t) => this.generator.generate(t)));

    return {
      timestamp: new Date().toISOString(),
      body: {
        opening: '本日のITニュースをお届けします。',
        topics,
        closing: '以上、本日のITニュースでした。',
      },
    };
  }
}
