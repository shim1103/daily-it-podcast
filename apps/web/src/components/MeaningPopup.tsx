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
      className="fixed z-50 max-w-xs w-60 bg-[#1f2937] border border-purple-500/30 rounded-2xl px-4 py-3 shadow-[0_8px_32px_rgba(0,0,0,0.6)] text-sm animate-fade-in"
    >
      <button
        onClick={onClose}
        aria-label="閉じる"
        className="absolute top-2 right-2.5 bg-transparent border-none cursor-pointer text-base text-gray-500 hover:text-gray-300 transition-colors"
      >
        ×
      </button>
      <p className="mb-1 font-semibold text-purple-300">{word}</p>
      <p className="text-gray-400 leading-relaxed">
        {loading ? (
          <span className="inline-flex items-center gap-1.5 text-gray-500">
            <span className="inline-block w-1.5 h-1.5 bg-purple-500 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
            <span className="inline-block w-1.5 h-1.5 bg-purple-500 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
            <span className="inline-block w-1.5 h-1.5 bg-purple-500 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
          </span>
        ) : meaning}
      </p>
    </div>
  );
}
