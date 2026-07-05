'use client';

import { useState } from 'react';
import type { Manuscript } from '@daily-it-podcast/core';
import { MeaningPopup } from './MeaningPopup';

interface Props {
  manuscript: Manuscript;
}

interface PopupState {
  word: string;
  position: { x: number; y: number };
}

export function ManuscriptViewer({ manuscript }: Props) {
  const [popup, setPopup] = useState<PopupState | null>(null);

  function handleTextSelect(e: React.MouseEvent) {
    const selection = window.getSelection();
    const word = selection?.toString().trim();
    if (!word) return;

    setPopup({
      word,
      position: { x: e.clientX, y: e.clientY + 12 },
    });
  }

  return (
    <div className="mt-2">
      <div className="flex items-center gap-2 mb-3">
        <span className="text-purple-400 text-sm font-medium">📄 原稿</span>
        <span className="text-xs text-gray-600 ml-auto">テキストを選択すると意味を検索</span>
      </div>

      <section
        onMouseUp={handleTextSelect}
        className="bg-[#111827] border border-white/8 rounded-2xl p-5 cursor-text select-text space-y-5"
      >
        <p className="text-gray-300 leading-relaxed text-sm">{manuscript.body.opening}</p>

        {manuscript.body.topics.map((topic, i) => (
          <div key={i} className="border-l-2 border-purple-500/40 pl-4">
            <h3 className="text-sm font-semibold mb-1.5 text-purple-300">{topic.title}</h3>
            <p className="text-gray-400 leading-relaxed text-sm">{topic.script}</p>
          </div>
        ))}

        <p className="text-gray-300 leading-relaxed text-sm">{manuscript.body.closing}</p>
      </section>

      {popup && (
        <MeaningPopup
          word={popup.word}
          position={popup.position}
          onClose={() => setPopup(null)}
        />
      )}
    </div>
  );
}
