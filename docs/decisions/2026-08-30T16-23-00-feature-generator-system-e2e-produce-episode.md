---
name: Generator System の最終 postcondition は test Drive 上の json+wav ペアと cleanup
date: 2026-08-30T16:23:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Generator System Test の成功 postcondition は次に限定する。(a) `cmd/generator` subprocess が exit 0。(b) `TEST_DRIVE_FOLDER_ID` 相当の folder に、実行前後の差分で現れた **1 stem** について `{stem}.json` と `{stem}.wav` が揃う。(c) 両 file が非空。(d) JSON の `episodeId` が stem と一致する。
2. System は成功 1 case だけを持つ。失敗系・HTTP 列・配線・schema 全 field・`no_source_items` 分類は下位 Scope の所有のまま再 assert しない。
3. Fetch 0 件など書込前終了は **precondition 不成立**として赤にする。skip で緑にしない。
4. 観測（list / get）と cleanup（delete）は **test package** が Drive API で行う。production `EpisodeWriter` に test 用 delete 公開や分岐を足さない。
5. case 終了時に本 run が作った Drive file を削除する。本番 folder / 本番 credential は使わない。
6. GHA の System workflow は Cursor CLI 公式 install 後に `agent` が PATH で解決できる状態にしてから `test-system.sh` を呼ぶ（BinaryName=`agent`）。
7. 本 Decision は先行 Decision（`2026-08-30T11-56-01` / `2026-08-26T17-47-00`）のうち「Verification 未定のため System を D に残す」範囲を、上記 Verification に限って埋める。gate 外・GHA 専用・subprocess 入口は維持する。

## 2. Reason

1. Unit / Narrow / Broad では実 credential 付きの入口→Drive 永続を self-validate できない。最終成果物は `contracts/drive-layout.md` のペアリングである。
2. schema 再検証は Application 書込直前の所有。System が全 field を再 assert すると minimization 越境になる。stem 一致と非空は layout 契約の最終観測に足りる。
3. Fetch 0 件を skip 緑にすると GetX / credential 障害も隠す。Pyramid 先端の Repeatable 限界は運用（test 用 GetX 条件）で受ける。
4. Writer へ cleanup API を足すと production Port が test 関心を抱える。観測・掃除は System の test 責務。
5. probe で確定した通り GHA は公式 install なしでは `agent` が無く、System / 本番 produce とも実 TextWriter に到達できない。

## 3. Rejected

1. Composition / `FromEnv` 直呼びを System とする案 — Driving Adapter を飛ばす（先行 Rejected）。
2. Broad の upload 回数 assert を System に残す案 — 下位所有の再 assert。
3. Fetch 0 件を成功扱いや skip 緑にする案 — precondition 不成立を隠蔽する。
4. 本番 Secrets / 本番 Drive folder を System に流用する案 — credential 禁止事項と同型。
5. production `EpisodeWriter` に delete を公開して test から呼ぶ案 — Port が test cleanup を知る。
