import type { ProgressRepository } from "./progress-repository.ts";

/**
 * 呼び出し回数を記録する ProgressRepository Spy（Application の test 用）。
 * 本番 adapter は置かない。呼ばれなかったことの検証に使う。
 */
export function createSpyProgressRepository(): {
  repository: ProgressRepository;
  callCount: () => number;
} {
  let calls = 0;
  const mark = <T>(fn: () => T): T => {
    calls += 1;
    return fn();
  };
  return {
    callCount: () => calls,
    repository: {
      getByEpisodeIds: async () => mark(() => new Map()),
      upsertProgress: async () => {
        mark(() => undefined);
      },
      listUpdatedSince: async () => mark(() => []),
      getProgressWithVersion: async () => mark(() => null),
      insertProgressIfAbsent: async () => mark(() => "written" as const),
      replaceProgressIfVersion: async () => mark(() => "written" as const),
      listChangedAfter: async () => mark(() => ({ cursor: 0, entries: [] })),
    },
  };
}
