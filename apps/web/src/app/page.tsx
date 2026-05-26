import { driveService } from '@/lib/drive';
import { EpisodeList } from '@/components/EpisodeList';

export default async function HomePage() {
  const episodes = await driveService.listEpisodes();
  return (
    <main style={{ maxWidth: 600, margin: '0 auto', padding: '24px 16px' }}>
      <h1 style={{ fontSize: 22, fontWeight: 'bold', marginBottom: 16 }}>Daily IT Podcast</h1>
      <EpisodeList episodes={episodes} />
    </main>
  );
}
