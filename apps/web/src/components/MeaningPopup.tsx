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

  // position は動的値のため inline style で座標のみ指定する
  const popupStyle = {
    top: position.y,
    left: Math.min(position.x, window.innerWidth - 260),
  } as React.CSSProperties;

  return (
    <div
      role="dialog"
      aria-label={`${word} の意味`}
      style={popupStyle}
      className="fixed z-50 max-w-xs w-60 bg-white border border-gray-200 rounded-xl px-4 py-3 shadow-lg text-sm"
    >
      <button
        onClick={onClose}
        aria-label="閉じる"
        className="absolute top-1.5 right-2 bg-transparent border-none cursor-pointer text-base text-gray-400 hover:text-gray-600"
      >
        ×
      </button>
      <p className="mb-1 font-bold text-gray-900">{word}</p>
      <p className="text-gray-500">
        {loading ? '検索中...' : meaning}
      </p>
    </div>
  );
}
