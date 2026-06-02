'use client';

import type { EpisodeMetadata } from '@daily-it-podcast/core';
import Link from 'next/link';

interface Props {
  episodes: EpisodeMetadata[];
}

export function EpisodeList({ episodes }: Props) {
  if (episodes.length === 0) {
    return (
      <div className="text-center py-16">
        <p className="text-4xl mb-3">📭</p>
        <p className="text-gray-500 text-sm">エピソードがありません</p>
      </div>
    );
  }

  return (
    <ul className="list-none m-0 p-0 space-y-3">
      {episodes.map((ep, i) => (
        <li
          key={ep.id}
          className="group animate-fade-in"
          style={{ animationDelay: `${i * 60}ms` }}
        >
          <Link
            href={`/episode/${ep.id}`}
            className="block bg-[#111827] hover:bg-[#1f2937] border border-white/8 hover:border-purple-500/40 rounded-2xl px-5 py-4 no-underline transition-all duration-200 hover:shadow-[0_0_20px_rgba(168,85,247,0.15)] hover:-translate-y-0.5"
          >
            <div className="flex items-start justify-between gap-3">
              <div className="flex-1 min-w-0">
                <p className="font-semibold text-gray-100 group-hover:text-purple-300 transition-colors duration-150 truncate">
                  {ep.title}
                </p>
                <p className="mt-1 text-xs text-gray-500">
                  {new Date(ep.timestamp).toLocaleString('ja-JP', {
                    year: 'numeric', month: 'long', day: 'numeric',
                    hour: '2-digit', minute: '2-digit',
                  })}
                </p>
              </div>
              <span className="flex-shrink-0 text-gray-600 group-hover:text-purple-400 transition-colors duration-150 mt-0.5">
                ▶
              </span>
            </div>
          </Link>
        </li>
      ))}
    </ul>
  );
}
