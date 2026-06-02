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
    <main className="mx-auto max-w-2xl px-4 py-8 sm:px-6">
      <Link
        href="/"
        className="text-sm text-indigo-600 hover:text-indigo-800 no-underline"
      >
        ← 一覧に戻る
      </Link>
      <h1 className="mt-4 mb-4 text-xl font-bold tracking-tight text-gray-900">
        {episode.metadata.title}
      </h1>
      <AudioPlayer audioUrl={episode.audioUrl} />
      <ManuscriptViewer manuscript={episode.manuscript} />
    </main>
  );
}
