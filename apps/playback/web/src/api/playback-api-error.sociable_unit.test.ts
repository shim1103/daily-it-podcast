import { describe, expect, it } from "vitest";
import { playbackHttpErrorCodes } from "../../../contracts/index.ts";
import {
  clientOnlyErrorCodes,
  contractErrorCodeMapping,
  mapHttpStatusToApiError,
  playbackApiErrorCodes,
} from "./playback-api-error.ts";

describe("mapHttpStatusToApiError", () => {
  it.each([
    [404, "episode_not_found"],
    [400, "validation_error"],
    [500, "configuration_error"],
    [503, "unavailable"],
  ] as const)("契約 status %i の時、%s を返す", (status, expected) => {
    // Given: 契約に定義された status
    // When: web 側 error へ写す
    const got = mapHttpStatusToApiError(status);

    // Then: 契約 code に対応する web 側 code
    expect(got).toBe(expected);
  });

  it("表に無い 4xx の時、client_error を返す", () => {
    // Given: 契約に定義されていない 4xx
    // When: web 側 error へ写す
    const got = mapHttpStatusToApiError(418);

    // Then: web 専用 error
    expect(got).toBe("client_error");
  });

  it.each([101, 301, 502, 600])("%i の時、unavailable を返す", (status) => {
    // Given: 契約に定義されていない非 2xx status
    // When: web 側 error へ写す
    const got = mapHttpStatusToApiError(status);

    // Then: web 側の unavailable
    expect(got).toBe("unavailable");
  });
});

describe("contractErrorCodeMapping", () => {
  it("契約 code の全件に web 側 code を割り当てる", () => {
    // Given: 契約が公開する code の集合

    // When: 写像表が持つ key を集める
    const mappedKeys = Object.keys(contractErrorCodeMapping).sort();

    // Then: 契約 code と過不足なく一致する
    expect(mappedKeys).toEqual([...playbackHttpErrorCodes].sort());
  });

  it("割り当て先が web 側 code の集合に収まる", () => {
    // Given: 写像表が返す値の集合
    const mappedValues = Object.values(contractErrorCodeMapping);

    // When: web 側 code に無い値を集める
    const unknown = mappedValues.filter((code) => !playbackApiErrorCodes.includes(code));

    // Then: 未定義の code は無い
    expect(unknown).toEqual([]);
  });

  it("異なる契約 code を同じ web 側 code へ畳まない", () => {
    // Given: 写像表が返す値の集合
    const mappedValues = Object.values(contractErrorCodeMapping);

    // When: 重複を除く
    const deduped = new Set(mappedValues);

    // Then: 件数が変わらない
    expect(deduped.size).toBe(mappedValues.length);
  });
});

describe("clientOnlyErrorCodes", () => {
  it("契約 code と重ならない", () => {
    // Given: 契約 code の集合
    const contractCodes: readonly string[] = playbackHttpErrorCodes;

    // When: client 専用 code のうち契約側にもある物を集める
    const overlapped = clientOnlyErrorCodes.filter((code) => contractCodes.includes(code));

    // Then: 重複は無い
    expect(overlapped).toEqual([]);
  });
});
