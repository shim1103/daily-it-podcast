import { driveService } from '@/lib/drive';
import { AudioPlayer } from '@/components/AudioPlayer';
import { ManuscriptViewer } from '@/components/ManuscriptViewer';
import Link from 'next/link';

interface Props {
  params: Promise<{ id: string }>;
}

export default async function EpisodePage({ params }: Props) {
  const { id } = await params;
  const episode = await driveService.getEpisode(id);

  return (
    <main style={{ maxWidth: 600, margin: '0 auto', padding: '24px 16px' }}>
      <Link href="/" style={{ color: '#0070f3', textDecoration: 'none', fontSize: 14 }}>
        ← 一覧に戻る
      </Link>
      <h1 style={{ fontSize: 18, fontWeight: 'bold', margin: '16px 0' }}>
        {episode.metadata.title}
      </h1>
      <AudioPlayer audioUrl={episode.audioUrl} />
      <ManuscriptViewer manuscript={episode.manuscript} />
    </main>
  );
}
