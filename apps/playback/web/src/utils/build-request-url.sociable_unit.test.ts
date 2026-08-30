import { describe, expect, it } from "vitest";
import { listEpisodesPath } from "../../../contracts/index.ts";
import { buildRequestUrl } from "./build-request-url.ts";

describe("buildRequestUrl", () => {
  it("baseUrl の末尾に / が無い時、そのまま path を続ける", () => {
    // Given: 末尾 / の無い baseUrl
    const baseUrl = "https://example.test/api";

    // When: 契約 path を繋ぐ
    const got = buildRequestUrl(baseUrl, listEpisodesPath);

    // Then: 段の区切りは 1 つ
    expect(got).toBe("https://example.test/api/episodes");
  });

  it("baseUrl の末尾に / が 1 つある時、重ねずに繋ぐ", () => {
    // Given: 末尾 / が 1 つの baseUrl
    const baseUrl = "https://example.test/api/";

    // When: 契約 path を繋ぐ
    const got = buildRequestUrl(baseUrl, listEpisodesPath);

    // Then: 段の区切りは 1 つ
    expect(got).toBe("https://example.test/api/episodes");
  });

  it("baseUrl の末尾に / が複数ある時、全て落として繋ぐ", () => {
    // Given: 末尾 / が複数の baseUrl
    const baseUrl = "https://example.test/api///";

    // When: 契約 path を繋ぐ
    const got = buildRequestUrl(baseUrl, listEpisodesPath);

    // Then: 段の区切りは 1 つ
    expect(got).toBe("https://example.test/api/episodes");
  });

  it("baseUrl が空文字の時、path だけを返す", () => {
    // Given: 同一 origin を指す空の baseUrl
    const baseUrl = "";

    // When: 契約 path を繋ぐ
    const got = buildRequestUrl(baseUrl, listEpisodesPath);

    // Then: 契約 path がそのまま残る
    expect(got).toBe("/episodes");
  });
});
