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
    <div style={{ marginTop: 16 }}>
      <p style={{ fontSize: 13, color: '#888', marginBottom: 8 }}>
        テキストを選択すると意味を検索できます
      </p>

      <section
        onMouseUp={handleTextSelect}
        style={{
          background: '#fff',
          borderRadius: 8,
          padding: '16px',
          boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
          cursor: 'text',
          userSelect: 'text',
        }}
      >
        <p style={{ margin: '0 0 12px', color: '#333' }}>{manuscript.body.opening}</p>

        {manuscript.body.topics.map((topic, i) => (
          <div key={i} style={{ marginBottom: 16 }}>
            <h3 style={{ fontSize: 15, fontWeight: 'bold', margin: '0 0 6px' }}>{topic.title}</h3>
            <p style={{ margin: 0, color: '#444', lineHeight: 1.7 }}>{topic.script}</p>
          </div>
        ))}

        <p style={{ margin: 0, color: '#333' }}>{manuscript.body.closing}</p>
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
