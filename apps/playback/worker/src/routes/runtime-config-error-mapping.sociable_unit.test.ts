import { describe, expect, it } from "vitest";
import { ConfigurationError } from "../../../contracts/index.ts";
import { PlaybackRuntimeConfigError } from "../composition/root.ts";
import { mapRuntimeConfigErrorToExternal } from "./runtime-config-error-mapping.ts";

describe("mapRuntimeConfigErrorToExternal", () => {
  it("PlaybackRuntimeConfigError の時、ConfigurationError へ変換し診断を cause へ残す", () => {
    // Given: runtime config が不正だった内部 Error
    const internal = new PlaybackRuntimeConfigError("DRIVE_FOLDER_ID が未設定です");

    // When: 外部向け Error へ変換する
    const got = mapRuntimeConfigErrorToExternal(internal);

    // Then: ConfigurationError へ変換し、内部 Error を cause へ残す
    expect(got).toBeInstanceOf(ConfigurationError);
    expect((got as ConfigurationError).cause).toBe(internal);
  });

  it("PlaybackRuntimeConfigError でない時、そのまま返す", () => {
    // Given: 契約 Error とは無関係な Error
    const other = new Error("予期しない失敗");

    // When: 外部向け Error へ変換する
    const got = mapRuntimeConfigErrorToExternal(other);

    // Then: 変換せずそのまま返す
    expect(got).toBe(other);
  });
});
