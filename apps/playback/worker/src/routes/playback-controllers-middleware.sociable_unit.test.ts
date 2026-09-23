import { describe, expect, it, vi } from "vitest";
import { Hono } from "hono";
import type { PlaybackControllers } from "../composition/root.ts";
import {
  createPlaybackControllersMiddleware,
  type PlaybackControllersVariables,
} from "./playback-controllers-middleware.ts";

const listEpisodesController = vi.fn(async () => ({ episodes: [] }));

vi.mock("../composition/root.ts", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../composition/root.ts")>();
  return {
    ...actual,
    createPlaybackControllers: vi.fn(() => ({
      listEpisodesController,
      getAudioController: vi.fn(),
      createProgressController: vi.fn(),
      updateProgressController: vi.fn(),
      completeProgressController: vi.fn(),
      pullProgressController: vi.fn(),
    })),
  };
});

import { createPlaybackControllers } from "../composition/root.ts";

describe("createPlaybackControllersMiddleware", () => {
  it("1 request で Composition Root を 1 回だけ呼び、route は context の controllers を使う", async () => {
    // Given: middleware 付きの最小 Hono
    const app = new Hono<{ Variables: PlaybackControllersVariables }>()
      .use(createPlaybackControllersMiddleware())
      .get("/probe", async (c) => {
        const { listEpisodesController: list } = c.get("controllers");
        return c.json(await list());
      });

    // When: 同じ app へ 1 回 request する
    const got = await app.request("http://example.test/probe", {}, {});

    // Then: CR は 1 回・応答は controller 経由
    expect(got.status).toBe(200);
    expect(createPlaybackControllers).toHaveBeenCalledTimes(1);
    expect(createPlaybackControllers).toHaveBeenCalledWith({}, { mode: "r2" }, undefined);
    expect(listEpisodesController).toHaveBeenCalledTimes(1);
  });

  it("useCaseOverrides を middleware へ渡す時、CR へそのまま渡す", async () => {
    // Given: override 付き middleware
    const overrides = {
      useCases: {
        listEpisodes: vi.fn(),
        getAudio: vi.fn(),
        createProgress: vi.fn(),
        updateProgress: vi.fn(),
        completeProgress: vi.fn(),
        pullProgress: vi.fn(),
      },
    };
    const app = new Hono<{ Variables: PlaybackControllersVariables }>()
      .use(createPlaybackControllersMiddleware(overrides))
      .get("/probe", (c) => {
        const controllers: PlaybackControllers = c.get("controllers");
        expect(controllers.listEpisodesController).toBeDefined();
        return c.body(null, 204);
      });

    // When
    await app.request("http://example.test/probe", {}, { EPISODES: {} });

    // Then
    expect(createPlaybackControllers).toHaveBeenCalledWith(
      { EPISODES: {} },
      { mode: "r2" },
      overrides,
    );
  });
});
