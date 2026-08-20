import { describe, expect, it } from "vitest";
import { GoogleDriveEpisodeRepository } from "../infrastructure/drive/google-drive-episode-repository.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/drive/in-memory-episode-repository.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";
import {
  createEpisodeRepository,
  createPlaybackControllers,
  type PlaybackRepositoryMode,
} from "./root.ts";

const localMode: PlaybackRepositoryMode = "in-memory";

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

  it("Drive の env が空でも、明示的な local mode の時だけ InMemoryEpisodeRepository を選ぶ", () => {
    // Given: 空 env と明示的な local / unit test mode

    // When: repository を組み立てる
    const got = createEpisodeRepository({}, { mode: localMode });

    // Then: Fake が選ばれる
    expect(got.kind).toBe("in-memory");
    if (got.kind === "in-memory") {
      expect(got.repository).toBeInstanceOf(InMemoryEpisodeRepository);
    }
  });

  it("Drive の env が空で mode が無い時、runtime config error を throw する", () => {
    // Given: production相当の暗黙設定なし
    // When / Then: Composition Root は設定不足を返さず throw する
    expect(() => createEpisodeRepository({})).toThrow(PlaybackRuntimeConfigError);
  });

  it("一部設定がある時は local mode を指定しても InMemoryへ切り替えない", () => {
    // Given: 1 key だけ注入された不完全設定と local mode
    const env = { GOOGLE_OAUTH_CLIENT_ID: "client-id" };

    // When / Then: 不完全設定を runtime config error にする
    expect(() => createEpisodeRepository(env, { mode: localMode })).toThrow(
      PlaybackRuntimeConfigError,
    );
  });

  it("空文字設定は local mode を指定しても設定不足のまま扱う", () => {
    // Given: 4 key が空文字の不完全設定と local mode
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "",
      GOOGLE_OAUTH_CLIENT_SECRET: "",
      GOOGLE_OAUTH_REFRESH_TOKEN: "",
      DRIVE_FOLDER_ID: "",
    };

    // When / Then: 空文字を明示的 local 経路へ読み替えず throw する
    expect(() => createEpisodeRepository(env, { mode: localMode })).toThrow(
      PlaybackRuntimeConfigError,
    );
  });

  it.each([
    ["GOOGLE_OAUTH_CLIENT_ID", ""],
    ["GOOGLE_OAUTH_CLIENT_SECRET", "   "],
    ["GOOGLE_OAUTH_REFRESH_TOKEN", ""],
    ["DRIVE_FOLDER_ID", "  "],
  ] as const)("Drive env の %s が空文字/空白なら設定不足になる", (emptyKey, emptyValue) => {
    // Given: 4 key は存在するが、1 key が空文字または空白
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token",
      DRIVE_FOLDER_ID: "folder-id",
      [emptyKey]: emptyValue,
    };

    // When / Then: 空文字を設定済みと扱わない
    expect(() => createEpisodeRepository(env)).toThrow(`${emptyKey} が未設定です`);
  });

  it("Drive の env が1つでも欠ける時、throw する（無言 Fake fallback をしない）", () => {
    // Given: DRIVE_FOLDER_ID だけが無い env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token",
    };

    // When / Then: 判定結果を返さず、Fake へ黙って落ちない
    expect(() => createEpisodeRepository(env)).toThrow("DRIVE_FOLDER_ID が未設定です");
  });

  it("Drive の env が1個だけ揃う時も throw する", () => {
    // Given: GOOGLE_OAUTH_CLIENT_ID だけがある env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
    };

    // When / Then: 中途半端な設定は runtime config error
    expect(() => createEpisodeRepository(env)).toThrow(
      "GOOGLE_OAUTH_CLIENT_SECRET が未設定です; GOOGLE_OAUTH_REFRESH_TOKEN が未設定です; DRIVE_FOLDER_ID が未設定です",
    );
  });
});

describe("createPlaybackControllers", () => {
  it("Drive の env が空の時、その env に基づく Controller 一式を組み立てる", () => {
    // Given: 空 env（意図的な Fake 利用）
    const env = {};

    // When: Controller 一式を組み立てる
    const got = createPlaybackControllers(env, { mode: localMode });

    // Then: ready として Controller 一式が返る
    expect(got.listEpisodesController).toBeDefined();
  });

  it("Drive の env が1つでも欠ける時、throw して Controller を組み立てない", () => {
    // Given: 一部だけ揃った env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
    };

    // When / Then: 設定不足を返さず、Controller も組み立てない
    expect(() => createPlaybackControllers(env)).toThrow(PlaybackRuntimeConfigError);
  });
});
