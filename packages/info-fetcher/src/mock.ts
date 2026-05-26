import type { InfoFetcher, ThemeInfo } from '@daily-it-podcast/core';

export class MockInfoFetcher implements InfoFetcher {
  async fetch(): Promise<ThemeInfo[]> {
    return [
      {
        source: 'manual-text',
        title: 'TypeScript 5.8 リリース',
        rawText:
          'TypeScript 5.8 が正式リリースされた。主な変更点として、return 文内の型絞り込み改善と --erasableSyntaxOnly フラグの追加がある。',
        fetchedAt: new Date().toISOString(),
      },
      {
        source: 'manual-text',
        title: 'Node.js 22 LTS 安定版',
        rawText:
          'Node.js 22 が LTS（長期サポート版）として安定版に昇格した。V8 エンジンの最新化と require(esm) のサポートが注目点。',
        fetchedAt: new Date().toISOString(),
      },
    ];
  }
}
