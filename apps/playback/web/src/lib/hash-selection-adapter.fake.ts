import type { HashSelectionAdapter } from "./hash-selection-adapter.ts";

/** `createFakeHashSelectionAdapter` が返す Fake。実 hash の代わりに in-memory で episodeId を保持する。 */
export type FakeHashSelectionAdapter = HashSelectionAdapter & {
  /** subscribe した listener に通知しつつ episodeId を外部変更に見立てて書き換える。 */
  externalChange(id: string | null): void;
  /** 現在購読中の listener 数。unmount で subscription が解除されたことの検証に使う。 */
  listenerCount(): number;
};

/**
 * `HashSelectionAdapter` の in-memory Fake を作る。
 *
 * @ensure `getEpisodeId` は保持値を返す。`setEpisodeId` と `externalChange` はどちらも
 *   保持値を書き換え、subscribe 済み listener をすべて呼ぶ。`subscribe` は解除関数を返す
 */
export function createFakeHashSelectionAdapter(
  initialId: string | null = null,
): FakeHashSelectionAdapter {
  let episodeId: string | null = initialId;
  const listeners = new Set<() => void>();

  const notify = (): void => {
    for (const listener of listeners) {
      listener();
    }
  };

  return {
    getEpisodeId(): string | null {
      return episodeId;
    },
    setEpisodeId(id: string | null): void {
      episodeId = id;
      notify();
    },
    subscribe(cb: () => void): () => void {
      listeners.add(cb);
      return () => {
        listeners.delete(cb);
      };
    },
    externalChange(id: string | null): void {
      episodeId = id;
      notify();
    },
    listenerCount(): number {
      return listeners.size;
    },
  };
}
