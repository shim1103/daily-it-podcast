import { describe, expect, it } from "vitest";
import { playbackHttpErrorCodes } from "../../../contracts/index.ts";
import {
  clientOnlyErrorCodes,
  contractErrorCodeMapping,
  playbackApiErrorCodes,
  toApiErrorCode,
} from "./playback-api-error.ts";

describe("toApiErrorCode", () => {
  it("kind が error の時、契約の code をそのまま返す", () => {
    // Given: 契約 code を持つ分類
    const classification = { kind: "error", code: "episode_not_found" } as const;

    // When: API error code へ写す
    const got = toApiErrorCode(classification);

    // Then: 契約 code がそのまま出る
    expect(got).toBe("episode_not_found");
  });

  it("kind が client_error の時、client_error を返す", () => {
    // Given: 契約 code を持たない client 側分類
    const classification = { kind: "client_error" } as const;

    // When: API error code へ写す
    const got = toApiErrorCode(classification);

    // Then: client 専用 code
    expect(got).toBe("client_error");
  });

  it("契約 code が unavailable の時も、そのまま返す", () => {
    // Given: 契約 code のうち unavailable を持つ分類
    const classification = { kind: "error", code: "unavailable" } as const;

    // When: API error code へ写す
    const got = toApiErrorCode(classification);

    // Then: 契約 code がそのまま出る
    expect(got).toBe("unavailable");
  });

  it("契約 code が validation_error の時も、そのまま返す", () => {
    // Given: 契約 code のうち validation_error を持つ分類
    const classification = { kind: "error", code: "validation_error" } as const;

    // When: API error code へ写す
    const got = toApiErrorCode(classification);

    // Then: 契約 code がそのまま出る
    expect(got).toBe("validation_error");
  });

  it("契約に未知の kind が来た時、既存 code へ倒さず throw する", () => {
    // Given: 契約 enum の拡張で増え、型検査を通さずに届いた未知の分類
    const unknown = { kind: "server_error" } as unknown as Parameters<typeof toApiErrorCode>[0];

    // Then: caller の分岐を誤らせる code を返さない
    expect(() => toApiErrorCode(unknown)).toThrow(TypeError);
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
