'use client';

import type { EpisodeMetadata } from '@daily-it-podcast/core';
import Link from 'next/link';

interface Props {
  episodes: EpisodeMetadata[];
}

export function EpisodeList({ episodes }: Props) {
  if (episodes.length === 0) {
    return <p className="text-gray-500">エピソードがありません</p>;
  }

  return (
    <ul className="list-none m-0 p-0 space-y-3">
      {episodes.map((ep) => (
        <li
          key={ep.id}
          className="bg-white rounded-xl px-4 py-3 shadow-sm border border-gray-100"
        >
          <Link
            href={`/episode/${ep.id}`}
            className="font-medium text-indigo-600 hover:text-indigo-800 no-underline"
          >
            {ep.title}
          </Link>
          <p className="mt-1 text-xs text-gray-400">
            {new Date(ep.timestamp).toLocaleString('ja-JP')}
          </p>
        </li>
      ))}
    </ul>
  );
}
