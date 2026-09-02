import { getLocationHash, onLocationHashChange, setLocationHash } from "./location-hash.ts";

export type HashSelectionAdapter = {
  getEpisodeId(): string | null;
  setEpisodeId(id: string | null): void;
  subscribe(cb: () => void): () => void;
};

/**
 * `location.hash` を episode 選択 id の Driven Adapter として包む。
 *
 * @ensure getEpisodeId は空 hash を null とする。setEpisodeId(null) は hash を消す。
 *   subscribe は hashchange を購読し、解除関数を返す
 */
export function createHashSelectionAdapter(): HashSelectionAdapter {
  return {
    getEpisodeId(): string | null {
      const hash = getLocationHash();
      return hash === "" ? null : hash;
    },
    setEpisodeId(id: string | null): void {
      setLocationHash(id ?? "");
    },
    subscribe(cb: () => void): () => void {
      return onLocationHashChange(cb);
    },
  };
}
