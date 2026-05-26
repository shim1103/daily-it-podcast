'use client';

import type { EpisodeMetadata } from '@daily-it-podcast/core';
import Link from 'next/link';

interface Props {
  episodes: EpisodeMetadata[];
}

export function EpisodeList({ episodes }: Props) {
  if (episodes.length === 0) {
    return <p style={{ color: '#666' }}>エピソードがありません</p>;
  }

  return (
    <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
      {episodes.map((ep) => (
        <li
          key={ep.id}
          style={{
            background: '#fff',
            borderRadius: 8,
            padding: '14px 16px',
            marginBottom: 10,
            boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
          }}
        >
          <Link
            href={`/episode/${ep.id}`}
            style={{ color: '#0070f3', textDecoration: 'none', fontWeight: 500 }}
          >
            {ep.title}
          </Link>
          <p style={{ margin: '4px 0 0', fontSize: 12, color: '#999' }}>
            {new Date(ep.timestamp).toLocaleString('ja-JP')}
          </p>
        </li>
      ))}
    </ul>
  );
}
