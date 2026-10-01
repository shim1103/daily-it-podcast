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

  it("returns_injected_responses_when_responses_are_given", async () => {
    // Given: first / all / run の応答を差し込んだ Fake の database
    const { database } = await createFakeLocalD1Binding({
      first: async () => ({ n: 1 }),
      all: async () => [{ n: 2 }],
      run: async () => ({ success: true, meta: { changes: 3 } }),
    });

    // When: それぞれの操作を実行する
    const first = await database.prepare("SELECT 1 AS n").first();
    const all = await database.prepare("SELECT 2 AS n").all();
    const written = await database.prepare("INSERT INTO t VALUES (1)").run();

    // Then: 差し込んだ応答がそのまま返る
    expect(first).toEqual({ n: 1 });
    expect(all.results).toEqual([{ n: 2 }]);
    expect(written).toEqual({ success: true, meta: { changes: 3 } });
  });

  it("records_sql_and_bound_values_in_call_order_when_preparing_statements", async () => {
    // Given: 応答を差し込まない Fake
    const { database, calls } = await createFakeLocalD1Binding();

    // When: bind ありと bind なしの statement を順に prepare する
    await database.prepare("SELECT ?").bind("a", 1).first();
    await database.prepare("SELECT 1").first();

    // Then: SQL と bind 値が呼び出し順に記録される（bind しない呼び出しは値が空）
    expect(calls).toEqual([
      { sql: "SELECT ?", values: ["a", 1] },
      { sql: "SELECT 1", values: [] },
    ]);
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
