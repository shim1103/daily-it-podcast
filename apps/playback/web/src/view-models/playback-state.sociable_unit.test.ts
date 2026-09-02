import { describe, expect, it } from "vitest";
import {
  deriveBlockingError,
  deriveIsPlayed,
  deriveIsPlaying,
  deriveIsSelected,
  derivePlayedEpisode,
  deriveSelectedEpisode,
  type EpisodeData,
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

const episodes = [episode, { ...episode, episodeId: "ep-2", title: "題2" }];

describe("deriveIsPlaying", () => {
  it("playbackPhase が playing の時、true を返す", () => {
    // Given: playbackPhase=playing
    // When: 再生中か導出する
    const got = deriveIsPlaying("playing");

    // Then: true
    expect(got).toBe(true);
  });

  it("playbackPhase が playing 以外の時、false を返す", () => {
    // Given: playbackPhase=idle
    // When: 再生中か導出する
    const got = deriveIsPlaying("idle");

    // Then: false
    expect(got).toBe(false);
  });
});

describe("deriveSelectedEpisode", () => {
  it("selectedEpisodeId が null の時、null を返す", () => {
    // Given: 選択 id なし
    // When: 選択中 episode を導出する
    const got = deriveSelectedEpisode(episodes, null);

    // Then: null
    expect(got).toBeNull();
  });

  it("selectedEpisodeId が一覧に存在する時、その episode を返す", () => {
    // Given: ep-1 を選択
    // When: 選択中 episode を導出する
    const got = deriveSelectedEpisode(episodes, "ep-1");

    // Then: ep-1
    expect(got).toEqual(episode);
  });

  it("selectedEpisodeId が一覧に無い時、null を返す", () => {
    // Given: 存在しない id を選択
    // When: 選択中 episode を導出する
    const got = deriveSelectedEpisode(episodes, "missing");

    // Then: null
    expect(got).toBeNull();
  });
});

describe("derivePlayedEpisode", () => {
  it("playedEpisodeId が null の時、null を返す", () => {
    // Given: 再生 id なし
    // When: 再生中 episode を導出する
    const got = derivePlayedEpisode(episodes, null);

    // Then: null
    expect(got).toBeNull();
  });

  it("playedEpisodeId が一覧に存在する時、その episode を返す", () => {
    // Given: ep-2 を再生
    // When: 再生中 episode を導出する
    const got = derivePlayedEpisode(episodes, "ep-2");

    // Then: ep-2
    expect(got?.episodeId).toBe("ep-2");
  });

  it("playedEpisodeId が一覧に無い時、null を返す", () => {
    // Given: 存在しない再生 id
    // When: 再生中 episode を導出する
    const got = derivePlayedEpisode(episodes, "missing");

    // Then: null
    expect(got).toBeNull();
  });
});

describe("deriveBlockingError", () => {
  it("catalogStatus が error の時、catalog-load-failed を返す", () => {
    // Given: catalog load 失敗
    // When: blocking error を導出する
    const got = deriveBlockingError({
      catalogStatus: { status: "error" },
      episodes: [],
      selectedEpisodeId: null,
      playedEpisodeId: null,
      playbackPhase: "idle",
    });

    // Then: catalog-load-failed
    expect(got).toEqual({ kind: "catalog-load-failed" });
  });

  it("catalog loading 中は blocking error を返さない", () => {
    // Given: catalog loading
    // When: blocking error を導出する
    const got = deriveBlockingError({
      catalogStatus: { status: "loading" },
      episodes: [],
      selectedEpisodeId: "ep-1",
      playedEpisodeId: null,
      playbackPhase: "idle",
    });

    // Then: null
    expect(got).toBeNull();
  });

  it("一覧に無い selectedEpisodeId の時、invalid-selected-episode を返す", () => {
    // Given: 未知の選択 id
    // When: blocking error を導出する
    const got = deriveBlockingError({
      catalogStatus: { status: "success" },
      episodes,
      selectedEpisodeId: "missing",
      playedEpisodeId: null,
      playbackPhase: "idle",
    });

    // Then: invalid-selected-episode
    expect(got).toEqual({ kind: "invalid-selected-episode", episodeId: "missing" });
  });

  it("playbackPhase が error の時、audio-load-failed を返す", () => {
    // Given: audio load 失敗
    // When: blocking error を導出する
    const got = deriveBlockingError({
      catalogStatus: { status: "success" },
      episodes,
      selectedEpisodeId: null,
      playedEpisodeId: "ep-1",
      playbackPhase: "error",
    });

    // Then: audio-load-failed
    expect(got).toEqual({ kind: "audio-load-failed", episodeId: "ep-1" });
  });

  it("blocking 条件が無い時、null を返す", () => {
    // Given: catalog success・valid 選択・再生 idle
    // When: blocking error を導出する
    const got = deriveBlockingError({
      catalogStatus: { status: "success" },
      episodes,
      selectedEpisodeId: "ep-1",
      playedEpisodeId: null,
      playbackPhase: "idle",
    });

    // Then: null
    expect(got).toBeNull();
  });
});

describe("deriveIsSelected", () => {
  it("episodeId が selectedEpisodeId と一致する時、true を返す", () => {
    // Given: ep-1 を選択
    // When: 選択中か導出する
    const got = deriveIsSelected("ep-1", "ep-1");

    // Then: true
    expect(got).toBe(true);
  });
});

describe("deriveIsPlayed", () => {
  it("episodeId が playedEpisodeId と一致する時、true を返す", () => {
    // Given: ep-1 を再生
    // When: 再生対象か導出する
    const got = deriveIsPlayed("ep-1", "ep-1");

    // Then: true
    expect(got).toBe(true);
  });
});
