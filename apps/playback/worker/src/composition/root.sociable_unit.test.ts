import { describe, expect, it } from "vitest";
import { GoogleDriveEpisodeRepository } from "../infrastructure/drive/google-drive-episode-repository.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/drive/in-memory-episode-repository.ts";
import { createEpisodeRepository, createPlaybackControllers } from "./root.ts";

describe("createEpisodeRepository", () => {
  it("Drive の env が全て揃う時、GoogleDriveEpisodeRepository を選ぶ", () => {
    // Given: OAuth と DRIVE_FOLDER_ID が揃った env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token",
      DRIVE_FOLDER_ID: "folder-id",
    };

    // When: repository を組み立てる
    const got = createEpisodeRepository(env);

    // Then: 本番 Adapter が選ばれる
    expect(got.kind).toBe("drive");
    if (got.kind === "drive") {
      expect(got.repository).toBeInstanceOf(GoogleDriveEpisodeRepository);
    }
  });

  it("Drive の env が空の時、InMemoryEpisodeRepository を選ぶ（意図的な Fake 利用）", () => {
    // Given: 空 env（ローカル開発・test相当）

    // When: repository を組み立てる
    const got = createEpisodeRepository({});

    // Then: Fake が選ばれる
    expect(got.kind).toBe("in-memory");
    if (got.kind === "in-memory") {
      expect(got.repository).toBeInstanceOf(InMemoryEpisodeRepository);
    }
  });

  it("Drive の env が1つでも欠ける時、misconfigured として欠落 key を返す（無言 Fake fallback をしない）", () => {
    // Given: DRIVE_FOLDER_ID だけが無い env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token",
    };

    // When: repository を組み立てる
    const got = createEpisodeRepository(env);

    // Then: 判定結果として欠落を表現する。Fake へ黙って落ちない
    expect(got.kind).toBe("misconfigured");
    if (got.kind === "misconfigured") {
      expect(got.missing).toEqual(["DRIVE_FOLDER_ID"]);
    }
  });

  it("Drive の env が1個だけ揃う時も misconfigured になる", () => {
    // Given: GOOGLE_OAUTH_CLIENT_ID だけがある env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
    };

    // When: repository を組み立てる
    const got = createEpisodeRepository(env);

    // Then: 中途半端な設定は misconfigured
    expect(got.kind).toBe("misconfigured");
    if (got.kind === "misconfigured") {
      expect(got.missing).toEqual([
        "GOOGLE_OAUTH_CLIENT_SECRET",
        "GOOGLE_OAUTH_REFRESH_TOKEN",
        "DRIVE_FOLDER_ID",
      ]);
    }
  });
});

describe("createPlaybackControllers", () => {
  it("Drive の env が空の時、その env に基づく Controller 一式を組み立てる", async () => {
    // Given: 空 env（意図的な Fake 利用）
    const env = {};

    // When: Controller 一式を組み立てて一覧を呼ぶ
    const got = createPlaybackControllers(env);

    // Then: ready として Controller 一式が返り、InMemoryEpisodeRepository（空）由来の空一覧が返る
    expect(got.kind).toBe("ready");
    if (got.kind === "ready") {
      const list = await got.controllers.listEpisodesController({});
      expect(list).toEqual({ episodes: [] });
    }
  });

  it("Drive の env が1つでも欠ける時、misconfigured を返し Controller を組み立てない", () => {
    // Given: 一部だけ揃った env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
    };

    // When: Controller 一式を組み立てる
    const got = createPlaybackControllers(env);

    // Then: misconfigured として欠落を伝える。無言で Fake へは逃げない
    expect(got.kind).toBe("misconfigured");
    if (got.kind === "misconfigured") {
      expect(got.missing).toEqual([
        "GOOGLE_OAUTH_CLIENT_SECRET",
        "GOOGLE_OAUTH_REFRESH_TOKEN",
        "DRIVE_FOLDER_ID",
      ]);
    }
  });
});
