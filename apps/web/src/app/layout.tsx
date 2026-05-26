import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Daily IT Podcast',
  description: '毎日のITニュースをpodcastで',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body style={{ margin: 0, fontFamily: 'sans-serif', background: '#f5f5f5' }}>
        {children}
      </body>
    </html>
  );
}
