/**
 * `<audio>` 要素の命令的操作をまとめる Driven Adapter。
 *
 * `<audio>` の markup（`<audio src ref>`）は Component 側に置く。この層は既に mount 済みの
 * 要素に対する `play()` / `pause()` / `currentTime` 代入 / lifecycle・状態 event 購読だけを閉じ込める。
 * ViewModel 層の `PlaybackState["phase"]` は import しない（`lib/` → `view-models/` は依存方向違反、
 * `external-dependencies.md` §5）。
 */

/**
 * `<audio>` の lifecycle event から**直接写せる** phase の語彙。
 *
 * 再生中に `<audio>` が発火する event（`playing` / `pause` / `ended` / `error`）に 1:1 対応する
 * 4 値のみを持つ。ViewModel の `PlaybackState["phase"]` はこれに `loading`（`play()` 呼び出し後
 * まだ event なし）を足した 5 値だが、`loading` は audio event 起点を持たず ViewModel が自前で
 * 管理する。この型はその 1 値を意図的に含めない。
 * `AudioLifecyclePhase ⊂ PlaybackState["phase"]`（構造的サブセット）なので、呼び出し側は
 * `AudioLifecyclePhase` を受ける callback をそのまま `onPhaseChange` へ渡せる。
 */
export type AudioLifecyclePhase = "playing" | "paused" | "ended" | "error";

/**
 * `subscribeAudioState` が `<audio>` の 3 系統の状態変化を写す先。
 *
 * why: phase（lifecycle event）・位置（`timeupdate`）・長さ（`loadedmetadata`）は
 *   発火 event も更新頻度も違うが、いずれも「1 要素の状態を state へ写す」同一責務なので
 *   1 つの購読関数へまとめ、必要な handler だけ呼び出し側が実装する。
 */
export type AudioStateHandlers = {
  onPhaseChange: (phase: AudioLifecyclePhase) => void;
  onPositionChange: (positionSec: number) => void;
  onDurationChange: (durationSec: number) => void;
};

/**
 * 購読対象の `<audio>` lifecycle event 名と、写す先の phase の対。
 *
 * why: event 名 "pause" に対し phase 語彙は過去分詞の "paused"。この 1 対だけ綴りが非対称。
 */
const audioEventPhasePairs: readonly (readonly [string, AudioLifecyclePhase])[] = [
  ["playing", "playing"],
  ["pause", "paused"],
  ["ended", "ended"],
  ["error", "error"],
];

/**
 * `<audio>` の状態変化を購読する。lifecycle event（`playing` / `pause` / `ended` / `error`）は
 * `AudioLifecyclePhase` へ写して `onPhaseChange` へ、`timeupdate` は現在位置を `onPositionChange` へ、
 * `loadedmetadata` は長さを `onDurationChange` へ渡す。
 *
 * @ensure lifecycle event 発火のたび対応 phase で `onPhaseChange` を呼ぶ。
 *   `timeupdate` のたび `el.currentTime` で `onPositionChange` を呼ぶ。
 *   `loadedmetadata` で `el.duration` が有限値のときだけ `onDurationChange` を呼ぶ
 *   （`NaN` / `Infinity` は「長さ未確定」を意味し state へ写さない。呼び出し側の null ガードは不要）。
 *   戻り値の関数を呼ぶと登録した listener をすべて解除する
 */
export function subscribeAudioState(
  el: HTMLAudioElement,
  handlers: AudioStateHandlers,
): () => void {
  const phaseListeners = audioEventPhasePairs.map(([type, phase]): [string, EventListener] => {
    const listener: EventListener = () => {
      handlers.onPhaseChange(phase);
    };
    el.addEventListener(type, listener);
    return [type, listener];
  });
  const onTimeUpdate: EventListener = () => {
    handlers.onPositionChange(el.currentTime);
  };
  const onLoadedMetadata: EventListener = () => {
    // why: metadata 未取得や無限長（ライブ配信）では `duration` が NaN / Infinity になる。
    //   有限でない長さは seek bar が描けず state に持つ意味がないため写さない
    if (Number.isFinite(el.duration)) {
      handlers.onDurationChange(el.duration);
    }
  };
  el.addEventListener("timeupdate", onTimeUpdate);
  el.addEventListener("loadedmetadata", onLoadedMetadata);
  return () => {
    for (const [type, listener] of phaseListeners) {
      el.removeEventListener(type, listener);
    }
    el.removeEventListener("timeupdate", onTimeUpdate);
    el.removeEventListener("loadedmetadata", onLoadedMetadata);
  };
}

/**
 * 別 episode へ切り替える時に `<audio>` を初期状態へ戻す（B Decision §1-3、seek にも適用）。
 *
 * @ensure `pause()` を呼び、`currentTime` を 0 にし、`load()` で source を読み直す
 */
export function resetAudioElement(el: HTMLAudioElement): void {
  el.pause();
  el.currentTime = 0;
  el.load();
}

/**
 * `stop()` 時に `<audio>` を止めて頭出しする。source は読み直さない。
 *
 * @ensure `pause()` を呼び、`currentTime` を 0 にする。`load()` は呼ばない
 *   （source 読み直しは別 episode 切替専用の `resetAudioElement` の責務）
 */
export function pauseAudioElement(el: HTMLAudioElement): void {
  el.pause();
  el.currentTime = 0;
}

/**
 * `<audio>` の再生位置を移動する。`opts.play` のときだけ続けて `play()` する（B Decision §1-3）。
 *
 * @ensure `currentTime` を `positionSec` にする。`opts.play` が true なら `el.play()` を呼び、
 *   その結果を `Promise.resolve()` で包んで返す（古い実装が undefined を返しても Promise 化する）。
 *   `opts.play` が false なら `play()` を呼ばず解決済み Promise を返す。
 *   rejection は呼び出し側が握る。`load()` は呼ばない
 *   （source 読み直しは別 episode 切替専用の `resetAudioElement` の責務）
 */
export function seekAudioElement(
  el: HTMLAudioElement,
  positionSec: number,
  opts: { play: boolean },
): Promise<void> {
  el.currentTime = positionSec;
  return opts.play ? Promise.resolve(el.play()) : Promise.resolve();
}
