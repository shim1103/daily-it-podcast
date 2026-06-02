import { driveService } from '@/lib/drive';
import { EpisodeList } from '@/components/EpisodeList';

export default async function HomePage() {
  const episodes = await driveService.listEpisodes();
  return (
    <main className="mx-auto max-w-2xl px-4 py-8 sm:px-6">
      <h1 className="mb-6 text-2xl font-bold tracking-tight text-gray-900">
        Daily IT Podcast
      </h1>
      <EpisodeList episodes={episodes} />
    </main>
  );
}
