import { Suspense } from 'react';
import { driveService } from '@/lib/drive';
import { EpisodeList } from '@/components/EpisodeList';

function EpisodeListSkeleton() {
  return (
    <ul className="list-none m-0 p-0 space-y-3">
      {[...Array(4)].map((_, i) => (
        <li key={i} className="bg-[#111827] border border-white/8 rounded-2xl px-5 py-4 animate-pulse">
          <div className="h-4 bg-white/8 rounded w-3/4 mb-2" />
          <div className="h-3 bg-white/5 rounded w-1/3" />
        </li>
      ))}
    </ul>
  );
}

async function EpisodeListLoader() {
  const episodes = await driveService.listEpisodes();
  return <EpisodeList episodes={episodes} />;
}

export default function HomePage() {
  return (
    <main className="mx-auto max-w-2xl px-4 py-12 sm:px-6 animate-fade-in">
      <div className="mb-10">
        <div className="flex items-center gap-3 mb-2">
          <span className="text-2xl">🎙</span>
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-purple-400 via-pink-400 to-cyan-400 bg-clip-text text-transparent">
            Daily IT Podcast
          </h1>
        </div>
        <p className="text-sm text-gray-500 pl-11">毎日のITニュースをpodcastで</p>
      </div>
      <Suspense fallback={<EpisodeListSkeleton />}>
        <EpisodeListLoader />
      </Suspense>
    </main>
  );
}
