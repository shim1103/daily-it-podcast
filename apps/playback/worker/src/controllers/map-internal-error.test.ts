import { describe, expect, it } from "vitest";
import { NotFoundError, UnavailableError } from "../../../contracts/index.ts";
import { EpisodeNotFoundError } from "../entities/errors/episode-not-found-error.ts";
import { DriveError } from "../infrastructure/drive/drive-error.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

describe("mapInternalErrorToExternal", () => {
  it("EpisodeNotFoundError の時、NotFoundError に cause を付ける", () => {
    // Given: Domain 不在
    const internal = new EpisodeNotFoundError("JSON エントリが無い: ep-1");

    // When: Internal を External へ写す
    const got = mapInternalErrorToExternal(internal);

    // Then: NotFoundError が元 Error を cause に持つ
    expect(got).toBeInstanceOf(NotFoundError);
    expect(got.cause).toBe(internal);
  });

  it("DriveError の時、UnavailableError に cause を付ける", () => {
    // Given: Infrastructure 失敗
    const internal = new DriveError("Drive 読取に失敗");

    // When: Internal を External へ写す
    const got = mapInternalErrorToExternal(internal);

    // Then: UnavailableError が元 Error を cause に持つ
    expect(got).toBeInstanceOf(UnavailableError);
    expect(got.cause).toBe(internal);
  });

  it("未知の Internal Error の時、UnavailableError に cause を付ける", () => {
    // Given: Domain 不在でも Infrastructure でもない Error
    const internal = new Error("想定外");

    // When: Internal を External へ写す
    const got = mapInternalErrorToExternal(internal);

    // Then: 契約内の UnavailableError に畳む
    expect(got).toBeInstanceOf(UnavailableError);
    expect(got.cause).toBe(internal);
  });
});
