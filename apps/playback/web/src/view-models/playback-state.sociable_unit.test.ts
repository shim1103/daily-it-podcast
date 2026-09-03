import { describe, expect, it } from "vitest";
import {
  deriveEpisodeRows,
  derivePageStatus,
  derivePlayedEpisode,
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
    phase: PlaybackPhase;
    positionSec: number;
    durationSec: number | null;
  }> = {},
): PlaybackState {
  return {
    kind: "active",
    episodeId: overrides.episodeId ?? "ep-1",
    phase: overrides.phase ?? { phase: "loading" },
    positionSec: overrides.positionSec ?? 0,
    durationSec: overrides.durationSec ?? null,
  };
}

describe("derivePlayedEpisode", () => {
  it("playback が kind=idle の時、null を返す", () => {
    // Given: 再生なし
    // When: 再生中 episode を導出する
    const got = derivePlayedEpisode(episodes, idlePlayback);

    // Then: null
    expect(got).toBeNull();
  });

  it("playback が kind=active で episodeId が一覧に無い時、null を返す", () => {
    // Given: 存在しない再生 id
    // When: 再生中 episode を導出する
    const got = derivePlayedEpisode(episodes, activePlayback({ episodeId: "missing" }));

    // Then: null
    expect(got).toBeNull();
  });

  it("playback が kind=active で episodeId が一覧に存在する時、その episode を返す", () => {
    // Given: ep-2 を再生
    // When: 再生中 episode を導出する
    const got = derivePlayedEpisode(
      episodes,
      activePlayback({ episodeId: "ep-2", phase: { phase: "playing" } }),
    );

    // Then: ep-2
    expect(got?.episodeId).toBe("ep-2");
  });
});

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

  it("選択中でも再生中でもない row は isSelected=false・isPlaying=false になる", () => {
    // Given: 選択なし・再生なし
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, { selection: noSelection, playback: idlePlayback });

    // Then: 全 row が false
    expect(got).toEqual([
      { episodeId: "ep-1", isSelected: false, isPlaying: false },
      { episodeId: "ep-2", isSelected: false, isPlaying: false },
    ]);
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
      { episodeId: "ep-1", isSelected: true, isPlaying: false },
      { episodeId: "ep-2", isSelected: false, isPlaying: false },
    ]);
  });

  it("kind=active かつ phase=playing で再生中の episode の row だけ isPlaying=true になる", () => {
    // Given: ep-2 を playing
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, {
      selection: noSelection,
      playback: activePlayback({ episodeId: "ep-2", phase: { phase: "playing" } }),
    });

    // Then: ep-2 の row のみ isPlaying=true
    expect(got).toEqual([
      { episodeId: "ep-1", isSelected: false, isPlaying: false },
      { episodeId: "ep-2", isSelected: false, isPlaying: true },
    ]);
  });

  it("kind=active でも phase が playing 以外なら isPlaying=false になる", () => {
    // Given: ep-2 を paused（まだ playing ではない）
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, {
      selection: noSelection,
      playback: activePlayback({ episodeId: "ep-2", phase: { phase: "paused" } }),
    });

    // Then: どの row も isPlaying=false
    expect(got.every((row) => row.isPlaying === false)).toBe(true);
  });

  it("kind=idle なら全 row が isPlaying=false になる", () => {
    // Given: 再生なし
    // When: row を投影する
    const got = deriveEpisodeRows(episodes, { selection: noSelection, playback: idlePlayback });

    // Then: 全 row が isPlaying=false
    expect(got.every((row) => row.isPlaying === false)).toBe(true);
  });
});
