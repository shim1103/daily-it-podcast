export default function EpisodeLoading() {
  return (
    <main className="mx-auto max-w-2xl px-4 py-8 sm:px-6 animate-fade-in">
      <div className="h-4 w-20 bg-white/8 rounded mb-6 animate-pulse" />
      <div className="h-6 bg-white/8 rounded w-2/3 mb-6 animate-pulse" />
      <div className="bg-[#111827] border border-white/8 rounded-2xl px-5 py-4 mb-6 animate-pulse">
        <div className="h-10 bg-white/5 rounded" />
      </div>
      <div className="bg-[#111827] border border-white/8 rounded-2xl p-5 space-y-4 animate-pulse">
        <div className="h-3 bg-white/8 rounded w-full" />
        <div className="h-3 bg-white/8 rounded w-5/6" />
        <div className="h-3 bg-white/5 rounded w-4/6" />
        <div className="border-l-2 border-purple-500/20 pl-4 space-y-2">
          <div className="h-3 bg-purple-500/15 rounded w-1/3" />
          <div className="h-3 bg-white/5 rounded w-full" />
          <div className="h-3 bg-white/5 rounded w-5/6" />
        </div>
        <div className="h-3 bg-white/8 rounded w-full" />
        <div className="h-3 bg-white/8 rounded w-3/4" />
      </div>
    </main>
  );
}
