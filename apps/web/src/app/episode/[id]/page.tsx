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
    <main className="mx-auto max-w-2xl px-4 py-8 sm:px-6 animate-fade-in">
      <Link
        href="/"
        className="inline-flex items-center gap-1.5 text-sm text-gray-400 hover:text-purple-400 no-underline transition-colors duration-150 mb-6"
      >
        <span>←</span>
        <span>一覧に戻る</span>
      </Link>
      <h1 className="mb-6 text-xl font-bold tracking-tight text-gray-100 leading-snug">
        {episode.metadata.title}
      </h1>
      <AudioPlayer audioUrl={episode.audioUrl} />
      <ManuscriptViewer manuscript={episode.manuscript} />
    </main>
  );
}
