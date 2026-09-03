import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { ApiSuccessData } from "../api/api-result.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { createFakeHashSelectionAdapter } from "../lib/hash-selection-adapter.fake.ts";
import { useEpisodeCatalog } from "./use-episode-catalog.ts";
import { useEpisodeListPage } from "./use-episode-list-page.ts";

vi.mock("./use-episode-catalog.ts", () => ({
  useEpisodeCatalog: vi.fn(),
}));

// why: production（playback-state.ts）と同じく API Client の型から episode 型を導出し、
//   境界共有型（contracts）を test から直接 import しない
type EpisodeData = ApiSuccessData<PlaybackApiClient["listEpisodes"]>["episodes"][number];

const episodeBody = {
  opening: "開始",
  topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 0 }],
  closing: "終了",
};

const episodeOne: EpisodeData = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  body: episodeBody,
  audioRef: "/episodes/ep-1/audio",
};

const episodeTwo: EpisodeData = {
  episodeId: "ep-2",
  date: "2026-08-18",
  title: "題2",
  durationSec: 90,
  body: episodeBody,
  audioRef: "/episodes/ep-2/audio",
};

/**
 * `<audio>` を模した Fake。listener 登録と event 発火だけを持つ最小版。
 */
function createFakeAudioElement(): HTMLAudioElement & { emit(type: string): void } {
  const listeners = new Map<string, Set<EventListener>>();
  const fake = {
    currentTime: 0,
    pause(): void {},
    load(): void {},
    play(): Promise<void> {
      return Promise.resolve();
    },
    addEventListener(type: string, listener: EventListener): void {
      const set = listeners.get(type) ?? new Set<EventListener>();
      set.add(listener);
      listeners.set(type, set);
    },
    removeEventListener(type: string, listener: EventListener): void {
      listeners.get(type)?.delete(listener);
    },
    emit(type: string): void {
      for (const listener of listeners.get(type) ?? []) {
        listener(new Event(type));
      }
    },
  };
  return fake as unknown as HTMLAudioElement & { emit(type: string): void };
}

function createStubApiClient(): PlaybackApiClient {
  return { listEpisodes: vi.fn() };
}

function mockCatalog(overrides: Partial<ReturnType<typeof useEpisodeCatalog>> = {}): void {
  vi.mocked(useEpisodeCatalog).mockReturnValue({
    catalogStatus: { status: "loading" },
    episodes: [],
    load: vi.fn(async (): Promise<void> => {}),
    ...overrides,
  });
}

describe("useEpisodeListPage", () => {
  it("catalog loading 時は pageStatus が loading の compose ViewModel を返す", () => {
    // Given: loading catalog stub
    mockCatalog({ catalogStatus: { status: "loading" }, episodes: [] });
    const apiClient = createStubApiClient();

    // When: hook を render する
    const { result } = renderHook(() => useEpisodeListPage(apiClient));

    // Then: loading・選択なし
    expect(result.current.catalogStatus).toEqual({ status: "loading" });
    expect(result.current.selection).toEqual({ selected: false });
    expect(result.current.pageStatus).toEqual({ kind: "loading" });
  });

  it("catalog success 時は pageStatus が ready の compose ViewModel を返す", () => {
    // Given: success catalog stub（episodes 空）
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [] });
    const apiClient = createStubApiClient();

    // When: hook を render する
    const { result } = renderHook(() => useEpisodeListPage(apiClient));

    // Then: success・選択なし・再生なし・ready・row 空
    expect(result.current.catalogStatus).toEqual({ status: "success" });
    expect(result.current.selection).toEqual({ selected: false });
    expect(result.current.selectedEpisode).toBeNull();
    expect(result.current.playback).toEqual({ kind: "idle" });
    expect(result.current.playedEpisode).toBeNull();
    expect(result.current.pageStatus).toEqual({ kind: "ready" });
    expect(result.current.episodes).toEqual([]);
    expect(result.current.rows).toEqual([]);
  });

  it("catalog error 時は pageStatus が unavailable reason=catalog-load-failed になる", () => {
    // Given: error catalog stub
    mockCatalog({ catalogStatus: { status: "error" }, episodes: [] });
    const apiClient = createStubApiClient();

    // When: hook を render する
    const { result } = renderHook(() => useEpisodeListPage(apiClient));

    // Then: pageStatus は unavailable / catalog-load-failed
    expect(result.current.pageStatus).toEqual({
      kind: "unavailable",
      reason: "catalog-load-failed",
    });
  });

  it("catalog success + episodes に無い id を select しても selection は入らず pageStatus は ready のまま", () => {
    // Given: success catalog stub（episodes 空）
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [] });
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient));

    // When: 一覧に無い id を select する
    act(() => {
      result.current.select("ghost");
    });

    // Then: selection は selected=false のまま・pageStatus は ready
    expect(result.current.selection).toEqual({ selected: false });
    expect(result.current.selectedEpisode).toBeNull();
    expect(result.current.pageStatus).toEqual({ kind: "ready" });
  });

  it("catalog success + 一覧にある id を select すると selectedEpisode が解決し pageStatus は ready", () => {
    // Given: ep-1 を持つ success catalog stub
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient));

    // When: ep-1 を select する
    act(() => {
      result.current.select("ep-1");
    });

    // Then: selectedEpisode が解決・ready・row の isSelected が true
    expect(result.current.selectedEpisode).toEqual(episodeOne);
    expect(result.current.selection).toEqual({ selected: true, episode: episodeOne });
    expect(result.current.pageStatus).toEqual({ kind: "ready" });
    expect(result.current.rows).toEqual([
      { episodeId: "ep-1", isSelected: true, isPlaying: false },
    ]);
  });

  it("catalog loading 中は hash 同期を開始しない（adapter へ書き込まない）", () => {
    // Given: loading catalog stub と Fake adapter
    mockCatalog({ catalogStatus: { status: "loading" }, episodes: [] });
    const adapter = createFakeHashSelectionAdapter();
    const setEpisodeId = vi.spyOn(adapter, "setEpisodeId");
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient, adapter));

    // When: loading 中に select する（一覧に居ないので no-op でもある）
    act(() => {
      result.current.select("ep-1");
    });

    // Then: hash へは書き込まれない
    expect(setEpisodeId).not.toHaveBeenCalled();
  });

  it("catalog success 後は select が hash へ反映される", () => {
    // Given: ep-1 を持つ success catalog stub と Fake adapter
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const adapter = createFakeHashSelectionAdapter();
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient, adapter));

    // When: ep-1 を select する
    act(() => {
      result.current.select("ep-1");
    });

    // Then: adapter が ep-1 になる
    expect(adapter.getEpisodeId()).toBe("ep-1");
  });

  it("catalog success 後、hash が外部で実在 id へ変わると select として解釈する", () => {
    // Given: ep-1 を持つ success catalog stub と Fake adapter
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const adapter = createFakeHashSelectionAdapter();
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient, adapter));

    // When: 外部で hash を ep-1 へ変える
    act(() => {
      adapter.externalChange("ep-1");
    });

    // Then: selection が ep-1 の実体になる
    expect(result.current.selection).toEqual({ selected: true, episode: episodeOne });
  });

  it("catalog success 後、hash が外部で一覧に無い id へ変わっても selection に入らず pageStatus は ready のまま", () => {
    // Given: ep-1 を持つ success catalog stub と Fake adapter
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const adapter = createFakeHashSelectionAdapter();
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient, adapter));

    // When: 外部で hash を一覧に無い id へ変える
    act(() => {
      adapter.externalChange("ghost");
    });

    // Then: selection は selected=false のまま・pageStatus は ready
    expect(result.current.selection).toEqual({ selected: false });
    expect(result.current.pageStatus).toEqual({ kind: "ready" });
  });

  it("catalog success 後、hash が外部で空へ変わると deselect として解釈する", () => {
    // Given: ep-1 を選択済みの hook
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const adapter = createFakeHashSelectionAdapter();
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient, adapter));
    act(() => {
      result.current.select("ep-1");
    });

    // When: 外部で hash を空へ変える
    act(() => {
      adapter.externalChange(null);
    });

    // Then: selection が selected=false になる
    expect(result.current.selection).toEqual({ selected: false });
  });

  it("Deselect しても playback は維持される（selection と playback の直交）", () => {
    // Given: ep-1 を再生中かつ選択中の hook（audioElementRef に何も張らず state だけ確認）
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient));
    act(() => {
      result.current.play("ep-1");
      result.current.select("ep-1");
    });

    // When: deselect する
    act(() => {
      result.current.deselect();
    });

    // Then: selection は selected=false だが playback の episodeId は ep-1 のまま
    expect(result.current.selection).toEqual({ selected: false });
    expect(result.current.playback).toMatchObject({
      kind: "active",
      episodeId: "ep-1",
      phase: { phase: "loading" },
    });
  });

  it("実在する別 episode を Play しても selection は変わらない（playback と selection の直交）", () => {
    // Given: ep-1・ep-2 を持つ success catalog で ep-1 を選択中の hook
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne, episodeTwo] });
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient));
    act(() => {
      result.current.select("ep-1");
    });

    // When: 実在する ep-2 を play する
    act(() => {
      result.current.play("ep-2");
    });

    // Then: selection は ep-1 のまま・playback のみ ep-2・ready
    expect(result.current.selection).toEqual({ selected: true, episode: episodeOne });
    expect(result.current.playback).toMatchObject({
      kind: "active",
      episodeId: "ep-2",
      phase: { phase: "loading" },
    });
    expect(result.current.pageStatus).toEqual({ kind: "ready" });
  });

  it("Play した audio が error event を出しても pageStatus は ready（unavailable にならない）で playback.phase が error になる", () => {
    // Given: ep-1 を持つ success catalog で fake audio を張り ep-1 を play 済みの hook
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const apiClient = createStubApiClient();
    const audio = createFakeAudioElement();
    const { result } = renderHook(() => useEpisodeListPage(apiClient));
    act(() => {
      result.current.audioElementRef.current = audio;
      result.current.play("ep-1");
    });

    // When: audio が error event を発火する
    act(() => {
      audio.emit("error");
    });

    // Then: pageStatus は ready のまま・playback.phase が error/audio-load-failed
    expect(result.current.pageStatus).toEqual({ kind: "ready" });
    expect(result.current.playback).toMatchObject({
      kind: "active",
      episodeId: "ep-1",
      phase: { phase: "error", reason: "audio-load-failed" },
    });
  });

  it("seek(episodeId, positionSec) を playback から compose して公開し、呼んでも selection は変わらない（直交）", () => {
    // Given: ep-1 を持つ success catalog で ep-1 を選択中の hook
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListPage(apiClient));
    act(() => {
      result.current.select("ep-1");
    });

    // When: seek する（audioElementRef を張っていないので state だけ倒るが compose の配線を確認）
    act(() => {
      result.current.seek("ep-1", 120);
    });

    // Then: seek は関数として公開され、selection は ep-1 のまま・playback は ep-1 の active
    expect(typeof result.current.seek).toBe("function");
    expect(result.current.selection).toEqual({ selected: true, episode: episodeOne });
    expect(result.current.playback).toMatchObject({
      kind: "active",
      episodeId: "ep-1",
      positionSec: 120,
    });
  });

  it("再生中に row の isPlaying が phase=playing でのみ true になる", () => {
    // Given: ep-1 を持つ success catalog で fake audio を張り ep-1 を play 済みの hook
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne] });
    const apiClient = createStubApiClient();
    const audio = createFakeAudioElement();
    const { result } = renderHook(() => useEpisodeListPage(apiClient));
    act(() => {
      result.current.audioElementRef.current = audio;
      result.current.play("ep-1");
    });

    // When: audio が playing event を発火する
    act(() => {
      audio.emit("playing");
    });

    // Then: row の isPlaying は true
    expect(result.current.rows).toEqual([
      { episodeId: "ep-1", isSelected: false, isPlaying: true },
    ]);
  });

  it("load / select / deselect / toggleSelection / play / stop / seek を compose して公開する", () => {
    // Given: success catalog stub
    const load = vi.fn(async (): Promise<void> => {});
    mockCatalog({ catalogStatus: { status: "success" }, episodes: [episodeOne], load });
    const apiClient = createStubApiClient();

    // When: hook を render する
    const { result } = renderHook(() => useEpisodeListPage(apiClient));

    // Then: catalog.load がそのまま公開される
    expect(result.current.load).toBe(load);
    expect(typeof result.current.select).toBe("function");
    expect(typeof result.current.deselect).toBe("function");
    expect(typeof result.current.toggleSelection).toBe("function");
    expect(typeof result.current.play).toBe("function");
    expect(typeof result.current.stop).toBe("function");
    expect(typeof result.current.seek).toBe("function");
    expect(result.current.audioElementRef.current).toBeNull();
  });
});
