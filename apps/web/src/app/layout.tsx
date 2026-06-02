import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Daily IT Podcast',
  description: '毎日のITニュースをpodcastで',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body className="min-h-screen bg-gray-50 font-sans text-gray-900 antialiased">
        {children}
      </body>
    </html>
  );
}
