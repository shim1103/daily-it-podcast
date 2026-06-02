import type { InfoFetcher, ThemeInfo } from '@daily-it-podcast/core';
import { InfoFetchError } from '@daily-it-podcast/core';

const BASE_URL = 'https://hacker-news.firebaseio.com/v0';

interface HnStory {
  id: number;
  title: string;
  url?: string;
  text?: string;
  score: number;
  time: number;
  type: string;
}

export interface HackerNewsConfig {
  /** 取得する最大件数（デフォルト 5） */
  maxItems?: number;
}

async function fetchJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    throw new InfoFetchError(`HN API 失敗: ${res.status} ${url}`);
  }
  return res.json() as Promise<T>;
}

function storyToThemeInfo(story: HnStory): ThemeInfo {
  // url があればリンク先タイトル文脈、なければ Ask HN / text 本文を使う
  const rawText = story.url
    ? `${story.title} ( ${story.url} ) — score: ${story.score}`
    : `${story.title}\n${story.text ?? ''}`.trim();

  return {
    source: 'hackernews',
    title: story.title,
    rawText,
    fetchedAt: new Date(story.time * 1000).toISOString(),
  };
}

export class HackerNewsInfoFetcher implements InfoFetcher {
  private readonly maxItems: number;

  constructor(config: HackerNewsConfig = {}) {
    this.maxItems = config.maxItems ?? 5;
  }

  async fetch(): Promise<ThemeInfo[]> {
    // top stories の ID 一覧を取得し、先頭 maxItems 件の詳細を並列取得
    const ids = await fetchJson<number[]>(`${BASE_URL}/topstories.json`);
    const topIds = ids.slice(0, this.maxItems);

    const stories = await Promise.all(
      topIds.map((id) => fetchJson<HnStory>(`${BASE_URL}/item/${id}.json`)),
    );

    return stories
      .filter((s) => s.type === 'story')
      .map(storyToThemeInfo);
  }
}
