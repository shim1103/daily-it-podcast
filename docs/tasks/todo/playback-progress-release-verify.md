# feat(playback): 進捗featureの常設検証とintegration PR

## 1. Summary

このIssueでは、常設 `playback-smoke` にD1経路を載せ、Integration／E2Eに必要なprogressケースを追加実装して緑化し、feature-integrationからintegrationへPRする。一過性dispatchの再発明はしない。

## 2. Context

1. 事実: 一過性疎通は `playback-d1-dispatch-smoke` が所有しPASS後削除する。
2. 事実: 常設smokeは既にTEST R2 remote bindingでlist／再生を見ている。D1を同じ束に載せる。
3. 事実: contribution-lifecycle上、feature完成後にfeature-integration→integrationへPRする。

## 3. Canonical Sources

1. 既存: `.github/workflows/playback-smoke.yml`・`web/vite.smoke.config.ts`・`test/smoke/`
2. B: `docs/decisions/2026-09-16T13-46-41-feature-r2-smoke-remove-and-migrate.md`（一過性≠常設）
3. B: branch: `1:terms/contribution-lifecycle/branch-strategy`
4. 先行達成契約: `docs/tasks/todo/playback-*.md`（本feature分）
5. test方針: `testing-strategy`

## 4. Scope

### In Scope

1. 常設smokeへ `EPISODE_PROGRESS`（TEST）経路を載せる追加実装
2. Integration／E2Eにprogress検証が足りなければケース追加実装
3. 上記が緑であること
4. feature-integration → integration へのPR

### Out of Scope

1. 一過性`workflow_dispatch`の再作成・常設化
2. merge意味の再設計、UI意味の変更
3. release（master）直PR

## 5. Contract

既存smokeのtest／prod隔離（TEST資源のみ）を壊さない。

## 6. Constraints

1. 一過性疎通artifactを常設gateに残さない。
2. 未完成stubをintegrationに載せない（branch-strategy）。

## 7. Acceptance Criteria

1. [ ] 常設smokeがTEST D1経路を含み緑
2. [ ] progressに必要なIntegration／E2E追加実装があり緑
3. [ ] feature-integration → integration のPRが作成されている（または同等の載せ完了が観測できる）
4. [ ] 一過性dispatch専用実装を常設として復活させていない

## 8. Verification

1. `npm run test:smoke`（playback）および該当Integration／E2E入口
2. PR URL／branch状態

## 9. Dependencies

1. blocked by: `playback-d1-dispatch-smoke`（一過性の役割分担が終わっていること）
2. blocked by: `playback-d1-progress-infra`・`playback-progress-application`・`playback-web-progress-sync`・`playback-web-progress-ui`（featureとして検証可能な完成形）

## 10. Risks

1. smokeにprogress assertを厚くしすぎる → 疎通／一覧成功＋必要最小のprogress観測に留める

## 11. Notes

数値・hook所有権の未決はlaneのまま。本Issueで決めない。
