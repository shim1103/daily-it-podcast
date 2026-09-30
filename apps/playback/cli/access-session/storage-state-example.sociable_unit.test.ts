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

type ExampleCookie = { name: string; value: string; domain: string; expires: number };

type ExampleStorageState = { cookies: ExampleCookie[]; origins: unknown[] };

function readExample(): ExampleStorageState {
  return JSON.parse(readFileSync(EXAMPLE_PATH, "utf8"));
}

describe("storage-state.example.json", () => {
  it("has_an_access_cookie_whose_expiry_the_check_can_read", () => {
    // Given: 人が手で組み立てる時の、JSON の構造の例
    const example = JSON.stringify(readExample());

    // When: 期限確認で読む
    const status = checkAccessSession(example, 0);

    // Then: Access cookie の期限（数値）が読める。期限の placeholder は 0 なので、そのままでは失効になる
    expect(status.expiresAtSec).toBe(0);
    expect(status.remainingSec).toBe(0);
  });

  it("uses_placeholders_for_the_value_and_the_domain", () => {
    // Given: JSON の構造の例
    const [cookie] = readExample().cookies;

    // When / Then: cookie の値と domain は、実値ではなく placeholder（< で始まる）である
    expect(cookie?.name).toBe(ACCESS_COOKIE_NAME);
    expect(cookie?.value.startsWith("<")).toBe(true);
    expect(cookie?.domain.startsWith("<")).toBe(true);
  });
});
