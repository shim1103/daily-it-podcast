---
name: playback worker の HTTP 入口を Hono app の fetch へ切替える
date: 2026-08-26T18:00:53
session_id: none
branch: feature/playback-worker-hono-entry-cutover
prev: なし
---

## 1. Summary

`docs/tasks/todo/playback-worker-hono-entry-cutover.md` を `issue-manager` skillのflowで完了した。`worker-entry.ts`のdefault exportを`app.ts`のHono instance（`app.fetch`）へ委譲する形に変更し、旧自作router（`fetch.ts`、`match-playback-route.ts`）と対応testを削除した。

## 2. Changes

1. 事前調査で、Issue本文に列挙されていない`routes/fetch.ts`への依存元2件（`apps/playback/test/runtime_config_boundary.broad_integration.test.ts`と、executor実装中に発見された`apps/playback/web/vite-config.sociable_unit.test.ts`）を発見し、削除前にimport先を`worker-entry.ts`/`app.ts`経由へ更新した。
2. `mapRuntimeConfigErrorToExternal`（Issue Risksが明記したerror mapping移行対象）を`routes/runtime-config-error-mapping.ts`へ独立させ、`app.ts`・`fetch.ts`の相互依存を解消してから`fetch.ts`を削除した。
3. `web/vite.config.ts`のdev-only middlewareがfake use caseの`overrides`を注入する必要があったため、`app.ts`に`createApp(useCaseOverrides?)`ファクトリを追加し、production用`app`はその呼び出し結果とした。production route定義・error写像の複製は無い。
4. `reviewer` agentの査読で、`vite.config.ts`のdocstring不正確表現（route未一致の実際の応答が「Honoの素の404」ではなく「`app.ts`のnotFound→400 validation_error」）と、削除された`fetch.sociable_unit.test.ts`が持っていた`not.toBeInstanceOf(Error)`の明示assertion欠落を指摘され、両方対応した。
5. AC-3（`npm run dev`実機確認）はsandbox制約（vite listenのEPERM）で`dangerouslyDisableSandbox`が必要だった。curlで一覧・詳細・audio取得の200応答を確認した。
6. commitを4本（worker側切替本体、web側dev-only middleware追従、broad integration test参照先変更、DESIGN.md更新）に分割してpush。pre-commit hookの`mktemp`失敗、git push時のssh認証failもsandbox制約起因で、いずれも`dangerouslyDisableSandbox`で解消した。

### Commits

1. `5427dc4`
2. `521ad40`
3. `57cef5e`
4. `eb2d52a`
