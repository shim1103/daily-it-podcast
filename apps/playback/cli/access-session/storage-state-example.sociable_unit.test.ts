// @vitest-environment node
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { ACCESS_COOKIE_NAME, checkAccessSession } from "./storage-state.ts";

const EXAMPLE_PATH = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  "storage-state.example.json",
);

const EXPIRES_AT_ISO = "2027-01-15T08:00:00.000Z";
const EXPIRES_AT_SEC = Date.parse(EXPIRES_AT_ISO) / 1000;

type ExampleCookie = {
  name: string;
  value: string;
  domain: string;
  expires: number | string;
};

type ExampleStorageState = { cookies: ExampleCookie[]; origins: unknown[] };

function readExample(): ExampleStorageState {
  return JSON.parse(readFileSync(EXAMPLE_PATH, "utf8"));
}

describe("storage-state.example.json", () => {
  it("is_rejected_by_the_check_until_the_expiry_is_replaced", () => {
    // Given: 何も置き換えていない、JSON の構造の例
    const example = JSON.stringify(readExample());

    // When / Then: expires が数値ではないので、期限を読めず拒否される（そのまま登録しても通らない）
    expect(() => checkAccessSession(example, EXPIRES_AT_SEC)).toThrow(ACCESS_COOKIE_NAME);
  });

  it("is_readable_by_the_check_once_the_placeholders_are_replaced", () => {
    // Given: placeholder を実値へ置き換えた JSON の構造の例（expires は DevTools の日時をそのまま貼る）
    const example = readExample();
    example.cookies = example.cookies.map((cookie) => ({
      ...cookie,
      value: "fake-cookie-value",
      domain: "playback.example.workers.dev",
      expires: EXPIRES_AT_ISO,
    }));

    // When: 期限確認で読む
    const status = checkAccessSession(JSON.stringify(example), EXPIRES_AT_SEC - 100);

    // Then: Access cookie の期限が読める
    expect(status).toEqual({ expiresAtSec: EXPIRES_AT_SEC, remainingSec: 100 });
  });

  it("uses_string_placeholders_for_the_value_the_domain_and_the_expiry", () => {
    // Given: JSON の構造の例
    const [cookie] = readExample().cookies;

    // When / Then: 置き換える 3 項目は、実値ではなく placeholder（< で始まる文字列）である
    expect(cookie?.name).toBe(ACCESS_COOKIE_NAME);
    for (const placeholder of [cookie?.value, cookie?.domain, cookie?.expires]) {
      expect(typeof placeholder).toBe("string");
      expect(String(placeholder).startsWith("<")).toBe(true);
    }
  });
});
