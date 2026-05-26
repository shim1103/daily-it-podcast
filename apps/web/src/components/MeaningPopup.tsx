'use client';

interface Props {
  word: string;
  position: { x: number; y: number };
  onClose: () => void;
}

export function MeaningPopup({ word, position, onClose }: Props) {
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
      {/* MVP: モック意味表示。将来は辞書APIまたはLLMに差し替える。 */}
      <p style={{ margin: 0, color: '#555' }}>「{word}」の意味（モック）</p>
    </div>
  );
}
