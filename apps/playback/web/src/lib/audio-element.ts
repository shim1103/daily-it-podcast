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
 * `<audio>` の音源 URL を差し替える。`<audio src>` を Component 側の React props で
 * controlled にすると、命令的な `play()` / `seekAudioElement()` が新 `src` の反映前に走って
 * 空要素を再生してしまう。音源指定もこの層へ寄せ、ViewModel が `play` / `seek` の直前に
 * 呼ぶことで順序を保証する。
 *
 * @require src は絶対 URL（呼び出し側が baseUrl と結合済み）
 * @ensure `el.src` を `src` にし、`load()` で読み込ませる。`load()` は再生位置を 0 へ戻すので
 *   別 episode 切替時の頭出しも兼ねる。同一 URL への再セットで無駄な `load()` を避ける差分判定は
 *   呼び出し側（`useEpisodePlayback.moveTo`）の責務で、この関数は毎回 `load()` する
 */
export function setAudioSource(el: HTMLAudioElement, src: string): void {
  el.src = src;
  el.load();
}

/**
 * `stop()`（UI 表示は「停止」）時に `<audio>` をその位置で止める。頭出しも source 読み直しもしない。
 *
 * @ensure `pause()` だけを呼ぶ。`currentTime` は変えない（B Decision §1-1/§1-2「停止中でも位置は残る」）。
 *   `load()` も呼ばない（source 読み直しは別 episode 切替専用の `setAudioSource` の責務）。
 *   同じ episode を再び再生するときは、残った `currentTime` からの再開になる
 */
export function pauseAudioElement(el: HTMLAudioElement): void {
  el.pause();
}

/** `HTMLMediaElement.HAVE_METADATA`。これ以上なら `duration` が確定し `currentTime` 代入が効く。 */
const HAVE_METADATA = 1;

/**
 * `currentTime` 代入後、その位置への実 seek が終わるのを待つ。
 *
 * why: `currentTime` 代入は即座に `seeking` 状態へ入るが、対象位置のデータ取得が終わるまで
 *   `seeked` event は発火しない。`seeked` を待たずに `play()` すると、ブラウザが「今取得できて
 *   いるデータ」（多くは先頭付近）から再生を始めてしまい、seek 先ではなく 0:00 付近から再生される。
 *
 * @ensure `el.currentTime` を `positionSec` にしたあと、`seeked` event の発火で解決する Promise を返す。
 */
function waitForSeekComplete(el: HTMLAudioElement, positionSec: number): Promise<void> {
  el.currentTime = positionSec;
  return new Promise<void>((resolve) => {
    const onSeeked: EventListener = () => {
      el.removeEventListener("seeked", onSeeked);
      resolve();
    };
    el.addEventListener("seeked", onSeeked);
  });
}

/**
 * `<audio>` の再生位置を移動する。`opts.play` のときだけ続けて `play()` する（B Decision §1-3）。
 *
 * why: `setAudioSource`（`load()`）直後は `readyState` が `HAVE_NOTHING` で、その状態の
 *   `currentTime` 代入は browser に無視され 0 のまま残る（topic の sec bar / 位置付き play が
 *   0:00 に飛ぶ原因）。metadata 未取得なら `loadedmetadata` を一度だけ待ってから代入する。
 *
 * @ensure `readyState < HAVE_METADATA` なら `loadedmetadata` の一度きり listener を張り、発火を
 *   待ってから `currentTime` を `positionSec` にする（`readyState >= HAVE_METADATA` なら即座に）。
 *   `opts.play` が false ならそこで解決した Promise を返す。`opts.play` が true なら、その代入が
 *   実際に反映される `seeked` event まで待ってから `el.play()` を呼び、その結果を
 *   `Promise.resolve()` で包んで返す（古い実装が undefined を返しても Promise 化）。
 *   rejection は呼び出し側が握る。`load()` は呼ばない
 *   （source 読み直しは別 episode 切替専用の `setAudioSource` の責務）
 */
export function seekAudioElement(
  el: HTMLAudioElement,
  positionSec: number,
  opts: { play: boolean },
): Promise<void> {
  const applySeek = (): Promise<void> => {
    if (!opts.play) {
      el.currentTime = positionSec;
      return Promise.resolve();
    }
    return waitForSeekComplete(el, positionSec).then(() => Promise.resolve(el.play()));
  };

  if (el.readyState >= HAVE_METADATA) {
    return applySeek();
  }

  return new Promise<void>((resolve, reject) => {
    const onLoadedMetadata: EventListener = () => {
      el.removeEventListener("loadedmetadata", onLoadedMetadata);
      applySeek().then(resolve, reject);
    };
    el.addEventListener("loadedmetadata", onLoadedMetadata);
  });
}
