import { act, renderHook } from "@testing-library/react";
import { useRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { createFakeHashSelectionAdapter } from "../lib/hash-selection-adapter.fake.ts";
import type { EpisodeData, SelectionState } from "./playback-state.ts";
import { useHashSelectionSync } from "./use-hash-selection-sync.ts";

const episodeOne: EpisodeData = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [{ title: "小題", preface: "前", detail: "詳", startSec: 0 }],
    closing: "終了",
  },
  audioRef: "/episodes/ep-1/audio",
};

const episodeTwo: EpisodeData = { ...episodeOne, episodeId: "ep-2", title: "題2" };

const noSelection: SelectionState = { selected: false };
const selectOne: SelectionState = { selected: true, episode: episodeOne };
const selectTwo: SelectionState = { selected: true, episode: episodeTwo };

describe("useHashSelectionSync", () => {
  it("catalog 未完了の間は選択があっても hash を書き換えない（同期保留）", () => {
    // Given: 既に ep-1 を持つ Fake adapter と catalogReady=false
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter("ep-1");
    const setEpisodeId = vi.spyOn(adapter, "setEpisodeId");

    // When: catalog 未完了で選択 ep-2 を渡して render する
    renderHook(() =>
      useHashSelectionSync(
        { catalogReady: false, selection: selectTwo },
        onHashEpisodeIdChange,
        adapter,
      ),
    );

    // Then: hash 書き込みなし・元の値のまま
    expect(setEpisodeId).not.toHaveBeenCalled();
    expect(adapter.getEpisodeId()).toBe("ep-1");
  });

  it("catalog 未完了の間は外部 hash 変化を onHashEpisodeIdChange へ流さない", () => {
    // Given: 空の Fake adapter と catalogReady=false
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();
    renderHook(() =>
      useHashSelectionSync(
        { catalogReady: false, selection: noSelection },
        onHashEpisodeIdChange,
        adapter,
      ),
    );

    // When: 外部で hash を ep-1 へ変える
    act(() => {
      adapter.externalChange("ep-1");
    });

    // Then: 保留中なので流さない
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();
  });

  it("catalog 未完了 → 完了へ変わると、保留中に起きた外部 hash 変化を解除後に流す", () => {
    // Given: catalogReady=false・選択なしで mount された hook
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();
    const { rerender } = renderHook(
      ({ catalogReady }) =>
        useHashSelectionSync(
          { catalogReady, selection: noSelection },
          onHashEpisodeIdChange,
          adapter,
        ),
      { initialProps: { catalogReady: false } },
    );

    // When: 保留中に外部で hash を ep-1 へ変え、その後 catalog 完了へ遷移する
    act(() => {
      adapter.externalChange("ep-1");
    });
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();
    rerender({ catalogReady: true });

    // Then: 解除後に 1 回だけ ep-1 を流す
    expect(onHashEpisodeIdChange).toHaveBeenCalledExactlyOnceWith("ep-1");
  });

  it("catalog 完了後、選択なしで render しても onHashEpisodeIdChange を呼ばない", () => {
    // Given: onHashEpisodeIdChange の spy と Fake adapter
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();

    // When: catalogReady=true・選択なしで render する
    renderHook(() =>
      useHashSelectionSync(
        { catalogReady: true, selection: noSelection },
        onHashEpisodeIdChange,
        adapter,
      ),
    );

    // Then: 発火なし
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();
  });

  it("catalog 完了後、選択中の episodeId を adapter へ書き込む", () => {
    // Given: 空の Fake adapter
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();

    // When: catalogReady=true・ep-1 選択で render する
    renderHook(() =>
      useHashSelectionSync(
        { catalogReady: true, selection: selectOne },
        onHashEpisodeIdChange,
        adapter,
      ),
    );

    // Then: adapter が ep-1 になる
    expect(adapter.getEpisodeId()).toBe("ep-1");
  });

  it("catalog 完了後、選択が別 episode へ変わると adapter を追従させる", () => {
    // Given: ep-1 を同期済みの hook
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();
    const { rerender } = renderHook(
      ({ selection }) =>
        useHashSelectionSync({ catalogReady: true, selection }, onHashEpisodeIdChange, adapter),
      { initialProps: { selection: selectOne as SelectionState } },
    );

    // When: ep-2 へ変える
    rerender({ selection: selectTwo });

    // Then: adapter が ep-2 になる
    expect(adapter.getEpisodeId()).toBe("ep-2");
  });

  it("catalog 完了後、選択中 → 選択なしへ変わると hash をクリアする", () => {
    // Given: ep-1 を同期済みの hook
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();
    const { rerender } = renderHook(
      ({ selection }) =>
        useHashSelectionSync({ catalogReady: true, selection }, onHashEpisodeIdChange, adapter),
      { initialProps: { selection: selectOne as SelectionState } },
    );

    // When: 選択なしへ変える
    rerender({ selection: noSelection });

    // Then: adapter が null になる
    expect(adapter.getEpisodeId()).toBeNull();
  });

  it("catalog 未完了 → 完了へ変わると、既存 hash（非空）を onHashEpisodeIdChange へ 1 回流す（deep-link 復元）", () => {
    // Given: 既に ep-1 を持つ Fake adapter・catalogReady=false で mount
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter("ep-1");
    const { rerender } = renderHook(
      ({ catalogReady }) =>
        useHashSelectionSync(
          { catalogReady, selection: noSelection },
          onHashEpisodeIdChange,
          adapter,
        ),
      { initialProps: { catalogReady: false } },
    );
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();

    // When: catalog 完了へ遷移する
    rerender({ catalogReady: true });

    // Then: 既存 hash ep-1 が 1 回だけ流れる
    expect(onHashEpisodeIdChange).toHaveBeenCalledExactlyOnceWith("ep-1");
  });

  it("catalog 完了時の既存 hash が空なら onHashEpisodeIdChange は呼ばれない", () => {
    // Given: 空の Fake adapter・catalogReady=false で mount
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();
    const { rerender } = renderHook(
      ({ catalogReady }) =>
        useHashSelectionSync(
          { catalogReady, selection: noSelection },
          onHashEpisodeIdChange,
          adapter,
        ),
      { initialProps: { catalogReady: false } },
    );

    // When: catalog 完了へ遷移する
    rerender({ catalogReady: true });

    // Then: 空 hash は流さない
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();
  });

  it("deep-link 復元後、caller が selection をその episode へ更新し hash へ書き戻しても無限ループしない", () => {
    // Given: 既存 hash ep-1・catalog 未完了で mount。caller は流れてきた episodeId を selection へ反映する
    let currentSelection: SelectionState = noSelection;
    const onHashEpisodeIdChange = vi.fn((episodeId: string | null) => {
      currentSelection = episodeId === "ep-1" ? selectOne : noSelection;
    });
    const adapter = createFakeHashSelectionAdapter("ep-1");
    const { rerender } = renderHook(
      ({ catalogReady, selection }) =>
        useHashSelectionSync({ catalogReady, selection }, onHashEpisodeIdChange, adapter),
      { initialProps: { catalogReady: false, selection: currentSelection } },
    );

    // When: catalog 完了へ遷移（deep-link 復元で ep-1 が流れ caller が selection を更新）→
    //   その selection を渡して再 render → もう一度発火機会
    rerender({ catalogReady: true, selection: currentSelection });
    rerender({ catalogReady: true, selection: currentSelection });
    act(() => {
      adapter.setEpisodeId(adapter.getEpisodeId());
    });

    // Then: onHashEpisodeIdChange は 1 回だけ・hash は ep-1 のまま（無限書き戻しなし）
    expect(onHashEpisodeIdChange).toHaveBeenCalledExactlyOnceWith("ep-1");
    expect(adapter.getEpisodeId()).toBe("ep-1");
  });

  it("catalog 完了後、外部 hash が非空へ変わると、その episodeId で onHashEpisodeIdChange を呼ぶ", () => {
    // Given: catalogReady=true・選択なしで同期済みの hook
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();
    renderHook(() =>
      useHashSelectionSync(
        { catalogReady: true, selection: noSelection },
        onHashEpisodeIdChange,
        adapter,
      ),
    );

    // When: 外部で hash を ep-1 へ変える
    act(() => {
      adapter.externalChange("ep-1");
    });

    // Then: ep-1 で呼ばれる
    expect(onHashEpisodeIdChange).toHaveBeenCalledWith("ep-1");
  });

  it("catalog 完了後、外部 hash が空へ変わると、null で onHashEpisodeIdChange を呼ぶ", () => {
    // Given: ep-1 を同期済みの hook
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();
    const { rerender } = renderHook(
      ({ selection }) =>
        useHashSelectionSync({ catalogReady: true, selection }, onHashEpisodeIdChange, adapter),
      { initialProps: { selection: selectOne as SelectionState } },
    );
    rerender({ selection: selectOne });

    // When: 外部で hash を空へ変える
    act(() => {
      adapter.externalChange(null);
    });

    // Then: null で呼ばれる
    expect(onHashEpisodeIdChange).toHaveBeenCalledWith(null);
  });

  it("自分の書き込みが誘発した通知を onHashEpisodeIdChange へ流さない", () => {
    // Given: 空の Fake adapter
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();

    // When: catalogReady=true・ep-1 選択で render し、書き込みが listener を発火させる
    renderHook(() =>
      useHashSelectionSync(
        { catalogReady: true, selection: selectOne },
        onHashEpisodeIdChange,
        adapter,
      ),
    );

    // Then: echo は流れない
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();
  });

  it("外部 hash 変化由来で caller が同じ episode を選択し hash へ書き戻しても無限ループしない", () => {
    // Given: 外部変化を受けたら selection をその episode へ更新する caller を模す
    let currentSelection: SelectionState = noSelection;
    const onHashEpisodeIdChange = vi.fn((episodeId: string | null) => {
      currentSelection =
        episodeId === "ep-1" ? selectOne : episodeId === "ep-2" ? selectTwo : noSelection;
    });
    const adapter = createFakeHashSelectionAdapter();
    const { rerender } = renderHook(
      ({ selection }) =>
        useHashSelectionSync({ catalogReady: true, selection }, onHashEpisodeIdChange, adapter),
      { initialProps: { selection: currentSelection } },
    );

    // When: 外部変化 → onHashEpisodeIdChange → caller 再 render → もう一度発火機会
    act(() => {
      adapter.externalChange("ep-1");
    });
    rerender({ selection: currentSelection });
    act(() => {
      adapter.setEpisodeId(adapter.getEpisodeId());
    });

    // Then: onHashEpisodeIdChange は 1 回だけ・hash は ep-1 のまま
    expect(onHashEpisodeIdChange).toHaveBeenCalledTimes(1);
    expect(adapter.getEpisodeId()).toBe("ep-1");
  });

  it("deep-link 復元の 1 回のあと、外部 hash 不変のまま再 render しても余分な再 render・発火が起きない", () => {
    // Given: ep-1 を持つ Fake adapter と render 回数 probe（catalog は最初から完了）
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter("ep-1");
    const rerenderCount = 3;
    let renderCount = 0;
    const useProbe = (): void => {
      const seen = useRef(0);
      seen.current += 1;
      renderCount = seen.current;
      useHashSelectionSync(
        { catalogReady: true, selection: selectOne },
        onHashEpisodeIdChange,
        adapter,
      );
    };
    const { rerender } = renderHook(() => useProbe());
    // 初回に deep-link 復元で 1 回だけ流れる
    expect(onHashEpisodeIdChange).toHaveBeenCalledExactlyOnceWith("ep-1");

    // When: 外部 hash を変えずに rerenderCount 回 rerender する
    for (let i = 0; i < rerenderCount; i += 1) {
      rerender();
    }

    // Then: render 回数は 1 + rerenderCount・以後の追加発火はない（1 回のまま）
    expect(renderCount).toBe(1 + rerenderCount);
    expect(onHashEpisodeIdChange).toHaveBeenCalledTimes(1);
  });

  it("unmount すると subscription が解除され、以後 onHashEpisodeIdChange を呼ばない", () => {
    // Given: 同期済みの hook
    const onHashEpisodeIdChange = vi.fn();
    const adapter = createFakeHashSelectionAdapter();
    const { unmount } = renderHook(() =>
      useHashSelectionSync(
        { catalogReady: true, selection: noSelection },
        onHashEpisodeIdChange,
        adapter,
      ),
    );

    // When: unmount してから外部変化を起こす
    unmount();
    act(() => {
      adapter.externalChange("ep-1");
    });

    // Then: 発火なし・listener も残らない
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();
    expect(adapter.listenerCount()).toBe(0);
  });

  it("adapter 未指定でも例外を投げずに render できる", () => {
    // Given: onHashEpisodeIdChange の spy（adapter は既定を使う）
    const onHashEpisodeIdChange = vi.fn();

    // When: adapter なしで render する
    // Then: 例外なし
    expect(() =>
      renderHook(() =>
        useHashSelectionSync(
          { catalogReady: false, selection: noSelection },
          onHashEpisodeIdChange,
        ),
      ),
    ).not.toThrow();
  });
});
