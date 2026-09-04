import { describe, expect, it } from "vitest";
import {
  deriveEpisodeRows,
  deriveNowPlaying,
  derivePageStatus,
  type EpisodeData,
  type PlaybackPhase,
  type PlaybackState,
  type SelectionState,
} from "./playback-state.ts";

const episode: EpisodeData = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [
      { title: "小題A", preface: "前A", detail: "詳A", startSec: 0 },
      { title: "小題B", preface: "前B", detail: "詳B", startSec: 30 },
    ],
    closing: "終了",
  },
  audioRef: "/episodes/ep-1/audio",
};

const episodeTwo: EpisodeData = { ...episode, episodeId: "ep-2", title: "題2" };

const episodes = [episode, episodeTwo];

const noSelection: SelectionState = { selected: false };
const idlePlayback: PlaybackState = { kind: "idle" };

function activePlayback(
  overrides: Partial<{
    episodeId: string;
    audioRef: string;
    phase: PlaybackPhase;
    positionSec: number;
    durationSec: number | null;
  }> = {},
): PlaybackState {
  return {
    kind: "active",
    episodeId: overrides.episodeId ?? "ep-1",
    audioRef: overrides.audioRef ?? "/episodes/ep-1/audio",
    phase: overrides.phase ?? { phase: "loading" },
    positionSec: overrides.positionSec ?? 0,
    durationSec: overrides.durationSec ?? null,
  };
}

describe("derivePageStatus", () => {
  it("catalogStatus が error の時、kind=unavailable reason=catalog-load-failed を返す", () => {
    // Given: catalog load 失敗
    // When: page status を導出する
    const got = derivePageStatus({ status: "error" });

    // Then: unavailable / catalog-load-failed
    expect(got).toEqual({ kind: "unavailable", reason: "catalog-load-failed" });
  });

  it("catalogStatus が loading の時、kind=loading を返す", () => {
    // Given: catalog loading
    // When: page status を導出する
    const got = derivePageStatus({ status: "loading" });

    // Then: loading
    expect(got).toEqual({ kind: "loading" });
  });

  it("catalogStatus が success の時、kind=ready を返す", () => {
    // Given: catalog success
    // When: page status を導出する
    const got = derivePageStatus({ status: "success" });

    // Then: ready
    expect(got).toEqual({ kind: "ready" });
  });
});

describe("deriveEpisodeRows", () => {
  it("空配列を渡すと空配列を返す", () => {
    // Given: episode が無い
    // When: row を投影する
    const got = deriveEpisodeRows([], { selection: noSelection, playback: idlePlayback });

    // Then: 空
    expect(got).toEqual([]);
  });

  it("選択中でも再生対象でもない row は isSelected=false・isActivePlayback=false・isPlaying=false になる", () => {
    // Given: 選択なし・再生なし
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, { selection: noSelection, playback: idlePlayback });

    // Then: 全 row が false
    expect(got).toEqual([
      { episode, episodeId: "ep-1", isSelected: false, isActivePlayback: false, isPlaying: false },
      {
        episode: episodeTwo,
        episodeId: "ep-2",
        isSelected: false,
        isActivePlayback: false,
        isPlaying: false,
      },
    ]);
  });

  it("各 row が入力 episode の実体を order 通りに同一参照で持つ", () => {
    // Given: 2 件の episode
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, { selection: noSelection, playback: idlePlayback });

    // Then: episode 実体が入力順で同一参照
    expect(got[0]?.episode).toBe(episode);
    expect(got[1]?.episode).toBe(episodeTwo);
  });

  it("選択中の episode の row だけ isSelected=true になる", () => {
    // Given: ep-1 を選択
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, {
      selection: { selected: true, episode },
      playback: idlePlayback,
    });

    // Then: ep-1 の row のみ isSelected=true
    expect(got).toEqual([
      { episode, episodeId: "ep-1", isSelected: true, isActivePlayback: false, isPlaying: false },
      {
        episode: episodeTwo,
        episodeId: "ep-2",
        isSelected: false,
        isActivePlayback: false,
        isPlaying: false,
      },
    ]);
  });

  it("kind=active かつ phase=playing で再生中の episode の row だけ isActivePlayback=true・isPlaying=true になる", () => {
    // Given: ep-2 を playing
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, {
      selection: noSelection,
      playback: activePlayback({ episodeId: "ep-2", phase: { phase: "playing" } }),
    });

    // Then: ep-2 の row のみ isActivePlayback=true かつ isPlaying=true
    expect(got).toEqual([
      { episode, episodeId: "ep-1", isSelected: false, isActivePlayback: false, isPlaying: false },
      {
        episode: episodeTwo,
        episodeId: "ep-2",
        isSelected: false,
        isActivePlayback: true,
        isPlaying: true,
      },
    ]);
  });

  it("kind=active で phase=paused なら isActivePlayback=false（停止中はボタン「再生」に戻る）", () => {
    // Given: ep-2 を paused（ユーザーが停止。位置は残るが再生進行中ではない）
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, {
      selection: noSelection,
      playback: activePlayback({ episodeId: "ep-2", phase: { phase: "paused" } }),
    });

    // Then: ep-2 も isActivePlayback=false・isPlaying=false（「再生」表示に戻り、押すと続きから）
    expect(got[1]).toMatchObject({ episodeId: "ep-2", isActivePlayback: false, isPlaying: false });
    expect(got.every((row) => row.isActivePlayback === false)).toBe(true);
  });

  it("kind=active で phase=ended / error でも isActivePlayback=false", () => {
    // Given: ep-2 が ended
    const endedRows = deriveEpisodeRows(episodes, {
      selection: noSelection,
      playback: activePlayback({ episodeId: "ep-2", phase: { phase: "ended" } }),
    });
    // Given: ep-2 が error
    const errorRows = deriveEpisodeRows(episodes, {
      selection: noSelection,
      playback: activePlayback({
        episodeId: "ep-2",
        phase: { phase: "error", reason: "audio-load-failed" },
      }),
    });

    // Then: どちらも再生進行中ではない
    expect(endedRows[1]).toMatchObject({ episodeId: "ep-2", isActivePlayback: false });
    expect(errorRows[1]).toMatchObject({ episodeId: "ep-2", isActivePlayback: false });
  });

  it("kind=active で phase=loading の episode は isActivePlayback=true になる（loading 中も停止できる）", () => {
    // Given: ep-1 を loading
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, {
      selection: noSelection,
      playback: activePlayback({ episodeId: "ep-1", phase: { phase: "loading" } }),
    });

    // Then: ep-1 は再生進行中（停止ボタン表示）
    expect(got[0]).toMatchObject({ episodeId: "ep-1", isActivePlayback: true, isPlaying: false });
  });

  it("kind=idle なら全 row が isPlaying=false になる", () => {
    // Given: 再生なし
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, { selection: noSelection, playback: idlePlayback });

    // Then: 全 row が isPlaying=false
    expect(got.every((row) => row.isPlaying === false)).toBe(true);
  });
});

describe("deriveNowPlaying", () => {
  it("kind=idle なら null を返す（mini-player に見出しを出さない）", () => {
    // Given: 再生対象なし
    // When: now-playing を投影する
    const got = deriveNowPlaying(episodes, idlePlayback);

    // Then: null
    expect(got).toBeNull();
  });

  it("kind=active なら再生中 episode の日付と通し番号付き title を返す", () => {
    // Given: ep-2（一覧の 2 件目・古い方）を再生中
    // When: now-playing を投影する
    const got = deriveNowPlaying(episodes, activePlayback({ episodeId: "ep-2" }));

    // Then: 表示用の日付（YYYY/MM/DD）と通し番号付き title（最古が 1）
    expect(got).toEqual({ date: "2026/08/17", numberedTitle: "1.　題2" });
  });

  it("先頭（新しい方）の episode ほど大きい通し番号になる", () => {
    // Given: ep-1（一覧の先頭・新しい方）を再生中
    // When: now-playing を投影する
    const got = deriveNowPlaying(episodes, activePlayback({ episodeId: "ep-1" }));

    // Then: 件数 2 - index 0 = 2
    expect(got).toEqual({ date: "2026/08/17", numberedTitle: "2.　題1" });
  });

  it("再生中 episodeId が一覧に無ければ null を返す", () => {
    // Given: 一覧に無い episodeId を active が指す
    // When: now-playing を投影する
    const got = deriveNowPlaying(episodes, activePlayback({ episodeId: "ep-missing" }));

    // Then: null（引き当て不能）
    expect(got).toBeNull();
  });
});
