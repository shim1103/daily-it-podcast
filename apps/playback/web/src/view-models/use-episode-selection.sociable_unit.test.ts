import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { EpisodeData } from "./playback-state.ts";
import { useEpisodeSelection } from "./use-episode-selection.ts";

const episodeOne: EpisodeData = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [{ title: "小題", preface: "前", detail: "詳", startSec: 0 }],
    ending: "終了",
  },
  audioRef: "/episodes/ep-1/audio",
};

const episodeTwo: EpisodeData = { ...episodeOne, episodeId: "ep-2", title: "題2" };

const episodes = [episodeOne, episodeTwo];

describe("useEpisodeSelection", () => {
  it("組み立て直後は selection.selected=false を返す", () => {
    // Given: なし
    // When: hook を render する
    const { result } = renderHook(() => useEpisodeSelection(episodes));

    // Then: 選択なし
    expect(result.current.selection).toEqual({ selected: false });
  });

  it("一覧に無い id を select しても selection は変わらない（no-op）", () => {
    // Given: hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));

    // When: 一覧に無い id を select する
    act(() => {
      result.current.select("ghost");
    });

    // Then: 選択なしのまま
    expect(result.current.selection).toEqual({ selected: false });
  });

  it("一覧に無い id を select しても、既存の選択は維持される（no-op）", () => {
    // Given: ep-1 を選択済みの hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));
    act(() => {
      result.current.select("ep-1");
    });

    // When: 一覧に無い id を select する
    act(() => {
      result.current.select("ghost");
    });

    // Then: ep-1 の選択が維持される
    expect(result.current.selection).toEqual({ selected: true, episode: episodeOne });
  });

  it("一覧に無い id を toggle しても selection は変わらない（no-op）", () => {
    // Given: 未選択の hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));

    // When: 一覧に無い id を toggle する
    act(() => {
      result.current.toggle("ghost");
    });

    // Then: 選択なしのまま
    expect(result.current.selection).toEqual({ selected: false });
  });

  it("一覧に有る id を select すると、その episode 実体で selected=true になる", () => {
    // Given: hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));

    // When: ep-1 を select する
    act(() => {
      result.current.select("ep-1");
    });

    // Then: ep-1 の実体が選択中
    expect(result.current.selection).toEqual({ selected: true, episode: episodeOne });
  });

  it("別の実在 id を select すると selection の episode が上書きされる", () => {
    // Given: ep-1 を選択済みの hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));
    act(() => {
      result.current.select("ep-1");
    });

    // When: ep-2 を select する
    act(() => {
      result.current.select("ep-2");
    });

    // Then: ep-2 の実体が選択中
    expect(result.current.selection).toEqual({ selected: true, episode: episodeTwo });
  });

  it("deselect() すると selection が selected=false になる", () => {
    // Given: ep-1 を選択済みの hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));
    act(() => {
      result.current.select("ep-1");
    });

    // When: deselect する
    act(() => {
      result.current.deselect();
    });

    // Then: 選択なし
    expect(result.current.selection).toEqual({ selected: false });
  });

  it("未選択の実在 id を toggle すると選択される", () => {
    // Given: 未選択の hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));

    // When: ep-1 を toggle する
    act(() => {
      result.current.toggle("ep-1");
    });

    // Then: ep-1 の実体が選択中
    expect(result.current.selection).toEqual({ selected: true, episode: episodeOne });
  });

  it("選択中の id を toggle すると選択解除される", () => {
    // Given: ep-1 を選択済みの hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));
    act(() => {
      result.current.select("ep-1");
    });

    // When: 同じ ep-1 を toggle する
    act(() => {
      result.current.toggle("ep-1");
    });

    // Then: 選択なし
    expect(result.current.selection).toEqual({ selected: false });
  });

  it("別の実在 id を toggle すると選択が切り替わる", () => {
    // Given: ep-1 を選択済みの hook
    const { result } = renderHook(() => useEpisodeSelection(episodes));
    act(() => {
      result.current.select("ep-1");
    });

    // When: ep-2 を toggle する
    act(() => {
      result.current.toggle("ep-2");
    });

    // Then: ep-2 の実体が選択中
    expect(result.current.selection).toEqual({ selected: true, episode: episodeTwo });
  });

  it("select / deselect / toggle は再 render をまたいで同一参照を保つ", () => {
    // Given: hook
    const { result, rerender } = renderHook(() => useEpisodeSelection(episodes));
    const first = {
      select: result.current.select,
      deselect: result.current.deselect,
      toggle: result.current.toggle,
    };

    // When: 状態を変えて再 render する
    act(() => {
      result.current.select("ep-1");
    });
    rerender();

    // Then: callback 参照は不変
    expect(result.current.select).toBe(first.select);
    expect(result.current.deselect).toBe(first.deselect);
    expect(result.current.toggle).toBe(first.toggle);
  });
});
