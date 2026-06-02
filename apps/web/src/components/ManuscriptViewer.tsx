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
    <div className="mt-4">
      <p className="text-xs text-gray-400 mb-2">
        テキストを選択すると意味を検索できます
      </p>

      <section
        onMouseUp={handleTextSelect}
        className="bg-white rounded-xl p-4 shadow-sm border border-gray-100 cursor-text select-text"
      >
        <p className="mb-3 text-gray-700 leading-relaxed">{manuscript.body.opening}</p>

        {manuscript.body.topics.map((topic, i) => (
          <div key={i} className="mb-4">
            <h3 className="text-sm font-bold mb-1 text-gray-900">{topic.title}</h3>
            <p className="text-gray-600 leading-relaxed">{topic.script}</p>
          </div>
        ))}

        <p className="text-gray-700 leading-relaxed">{manuscript.body.closing}</p>
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
