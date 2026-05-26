import type { ScriptGenerator, ThemeInfo, ThemeScript } from '@daily-it-podcast/core';

export class MockScriptGenerator implements ScriptGenerator {
  async generate(info: ThemeInfo): Promise<ThemeScript> {
    return {
      title: info.title,
      script: `${info.title}についてお伝えします。${info.rawText} 以上、${info.title}についてでした。`,
      durationEstimateSec: 60,
    };
  }
}
