## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`  
deploy・Access・GHA 運用の正: `DEPLOY.md`  
React + Hono 方針の正: `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md`

未完了の達成契約は `docs/tasks/todo/playback-*.md` が正。本 lane は進捗 index のみ。GitHub Issue 化しない運用。

### 済み（要約）

1. web↔worker HTTP / Drive adapter / UI / 層検知 / React+Hono
2. deploy 前準備・初回 `wrangler deploy`・Access OTP（許可/拒否）・同一 origin の一覧・原稿・再生
3. 認証済み browser E2E（週次 `playback-e2e` + `workflow_dispatch`。安定 fixture は本番 `DRIVE_FOLDER_ID`）
4. Narrow / Broad Integration（gate 内）と `test/integration` + `test/e2e` 配置
5. playback web 直交 Decision（B）と A 契約（型・stub・SU test）
6. 運用後続 docs（Decision 2026-09-04T02-04-00〜02）

### 未完了

1. [ ] `docs/tasks/todo/playback-web-view-models.md`
2. [ ] `docs/tasks/todo/playback-web-ui-rewrite.md`
3. [ ] `docs/tasks/todo/playback-web-legacy-cleanup.md`

### 方針 index（Decision）

```text
gate = secret なし Narrow + Broad（Decision 2026-08-30T16-20-00）
E2E = gate 外・週次月曜 07:00 JST + dispatch（同 Decision）
coverage 分母 = SU + secret なし NI（Decision 2026-08-30T16-20-01）
OTP 手動・週次 storageState・Drive=Worker（Decision 2026-08-30T16-20-03）
test dir = integration/ + e2e/（Decision 2026-08-31T00-12-00）
E2E 正常系 = 本番 Drive 安定 fixture ≥1（Decision 2026-08-31T00-12-01 / 2026-08-31T00-22-00）
selection⊥playback = Decision 2026-09-02T15-00-00-feature-playback-list-episodes-audio-ref-playback-web-selection-playback-orthogonality
rollback = wrangler rollback 全量（Decision 2026-09-04T02-04-00）
observability = 常時 ON・契約は wrangler.jsonc（Decision 2026-09-04T02-04-01）
運用後続完了 = docs/契約まで・再 deploy 非 scope（Decision 2026-09-04T02-04-02）
```

音声は wav。generator 書込とは共有しない。
