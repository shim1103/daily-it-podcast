'use client';

import { useState, useEffect } from 'react';

interface Props {
  word: string;
  position: { x: number; y: number };
  onClose: () => void;
}

export function MeaningPopup({ word, position, onClose }: Props) {
  const [meaning, setMeaning] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setMeaning(null);

    fetch('/api/meaning', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ word }),
    })
      .then((res) => res.json())
      .then((data: { meaning?: string }) => {
        if (!cancelled) setMeaning(data.meaning ?? '説明を取得できませんでした。');
      })
      .catch(() => {
        if (!cancelled) setMeaning('説明を取得できませんでした。');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    // word が変わった場合に前のリクエスト結果を破棄する
    return () => {
      cancelled = true;
    };
  }, [word]);

  return (
    <div
      role="dialog"
      aria-label={`${word} の意味`}
      style={{
        position: 'fixed',
        top: position.y,
        left: Math.min(position.x, window.innerWidth - 260),
        background: '#fff',
        border: '1px solid #ddd',
        borderRadius: 8,
        padding: '12px 16px',
        boxShadow: '0 4px 16px rgba(0,0,0,0.15)',
        zIndex: 1000,
        maxWidth: 240,
        fontSize: 14,
      }}
    >
      <button
        onClick={onClose}
        aria-label="閉じる"
        style={{
          position: 'absolute',
          top: 6,
          right: 8,
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          fontSize: 16,
          color: '#999',
        }}
      >
        ×
      </button>
      <p style={{ margin: '0 0 4px', fontWeight: 'bold' }}>{word}</p>
      <p style={{ margin: 0, color: '#555' }}>
        {loading ? '検索中...' : meaning}
      </p>
    </div>
  );
}
