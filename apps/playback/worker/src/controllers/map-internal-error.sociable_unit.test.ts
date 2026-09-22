import { describe, expect, it } from "vitest";
import { NotFoundError, UnavailableError, ValidationError } from "../../../contracts/index.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import { ProgressNotFoundError } from "../entities/errors/progress-not-found-error.ts";
import { ProgressRuleError } from "../entities/errors/progress-rule-error.ts";
import { R2Error } from "../infrastructure/r2/r2-error.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

describe("mapInternalErrorToExternal", () => {
  it("EpisodeContentError の時、NotFoundError に cause を付ける", () => {
    // Given: Domain の実体不備
    const internal = new EpisodeContentError("JSON エントリが無い: ep-1");

    // When: Internal を External へ写す
    const got = mapInternalErrorToExternal(internal);

    // Then: NotFoundError が元 Error を cause に持つ
    expect(got).toBeInstanceOf(NotFoundError);
    expect(got.cause).toBe(internal);
  });

  it("ProgressRuleError の時、ValidationError に cause を付ける", () => {
    // Given: 進捗の意味ルール違反（skew 等）
    const internal = new ProgressRuleError("clientAt が許容 skew を超える");

    // When: Internal を External へ写す
    const got = mapInternalErrorToExternal(internal);

    // Then: ValidationError（400 validation_error）へ畳む。HTTP code は増やさない
    expect(got).toBeInstanceOf(ValidationError);
    expect(got.cause).toBe(internal);
  });

  it("ProgressNotFoundError の時、NotFoundError に cause を付ける", () => {
    // Given: update 対象の進捗行が無い
    const internal = new ProgressNotFoundError("進捗行が無い: ep-1");

    // When: Internal を External へ写す
    const got = mapInternalErrorToExternal(internal);

    // Then: NotFoundError（404 episode_not_found）へ畳む
    expect(got).toBeInstanceOf(NotFoundError);
    expect(got.cause).toBe(internal);
  });

  it("R2Error の時、UnavailableError に cause を付ける", () => {
    // Given: Infrastructure 失敗
    const internal = new R2Error("R2 読取に失敗");

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
