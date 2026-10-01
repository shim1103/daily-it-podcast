import { beforeEach, describe, expect, it, vi } from "vitest";
import { createFakeLocalD1Binding } from "./create-fake-local-d1-binding.ts";

const { getPlatformProxy } = vi.hoisted(() => ({ getPlatformProxy: vi.fn() }));

vi.mock("wrangler", () => ({ getPlatformProxy }));

describe("createFakeLocalD1Binding", () => {
  beforeEach(() => {
    getPlatformProxy.mockClear();
  });

  it("returns_empty_first_and_all_when_reading_any_query", async () => {
    // Given: SU 用 Fake の database
    const { database } = await createFakeLocalD1Binding();

    // When: 任意の SELECT を first / all で読む
    const first = await database.prepare("SELECT 1 AS n").first();
    const all = await database.prepare("SELECT 1 AS n").all();

    // Then: 読取は常に空
    expect(first).toBeNull();
    expect(all.results).toEqual([]);
  });

  it("returns_success_with_zero_changes_when_running_a_write", async () => {
    // Given: SU 用 Fake の database
    const { database } = await createFakeLocalD1Binding();

    // When: 任意の INSERT を run する
    const written = await database.prepare("INSERT INTO t VALUES (1)").run();

    // Then: 書込は success だが何も変更しない
    expect(written).toEqual({ success: true, meta: { changes: 0 } });
  });

  it("returns_the_same_statement_when_binding_values", async () => {
    // Given: prepare 済みの statement
    const { database } = await createFakeLocalD1Binding();
    const statement = database.prepare("SELECT ? AS n");

    // When: bind する
    const bound = statement.bind(1, "a");

    // Then: statement 自身を返し、bind を連ねても読取は空のまま
    expect(bound).toBe(statement);
    expect(await bound.first()).toBeNull();
  });

  it("resolves_dispose_without_error_when_disposing", async () => {
    // Given: SU 用 Fake の handle
    const handle = await createFakeLocalD1Binding();

    // When / Then: dispose は何も解放せず正常に完了する
    await expect(handle.dispose()).resolves.toBeUndefined();
  });

  it("does_not_start_proxy_when_creating_and_disposing", async () => {
    // Given: getPlatformProxy を観測できる状態（wrangler は spy へ差し替え済み）
    // When: 生成して dispose する
    const handle = await createFakeLocalD1Binding();
    await handle.dispose();

    // Then: 実 proxy を起動していない
    expect(getPlatformProxy).not.toHaveBeenCalled();
  });
});
