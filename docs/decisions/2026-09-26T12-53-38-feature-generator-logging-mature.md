---
name: port.RetryReporterはnil非許容にしComposition Root結線漏れをpanicで顕在化する
date: 2026-09-26T12:53:38
branch: feature/generator-logging-mature
---

## 1. Decision

1. `port.RetryReporter`を必須依存として扱う。DIされる全Adapter（`geminiapi`・`cursorapi`・`speech/gemini`・`r2`・`httpget`+記事source5種）のconstructorで`retry == nil`を検査し、panicする。
2. 各Adapter内部の`if retry != nil { retry.Retry(...) }`という条件分岐は除去し、`retry.Retry(...)`を直接呼ぶ（constructorで非nilを保証済みのため）。
3. `port.NoopRetryReporter{}`（何もしないRetryReporter実装）を`port`パッケージへ新設する。Retry呼び出しの有無・内容を検証しないtestのDummyとして使う。Retryの呼び出しを実際に検証するtestは既存のSpy（`retryReporterSpy`）をそのまま使い、Dummyへ置き換えない。

## 2. Reason

1. 前段（同日付近い時刻に作成した`2026-09-25T18-24-04`のDecision）で`if retry != nil`という暗黙fallbackを実装したが、これは`configuration-boundary.md`§7「暗黙defaultの禁止」に反する。未注入時に内部で黙って動作を変える形は、Composition Rootが結線をサボっても実行時に気づけない経路を残す。
2. `retry`がnilという状態は、実行環境由来の失敗（ネットワーク不通等）ではなく、Composition Rootの結線漏れという**プログラマの契約違反**である。Goの慣習では、環境由来の失敗はerror値で返し呼び出し元に対処させるが、契約違反はpanicで即座に落として気づかせる方が適切（バグは復旧ではなく修正の対象のため）。
3. constructorという「プロセス起動時に1回だけ実行される場所」でのpanicは、起動を継続させない（fail-fast at startup）という望ましい振る舞いであり、`runtime.DisplayLocation`が既に同種の判断（tzdata解決失敗時のpanic）を採用している。
4. `nil`はGoの型システム上、コンパイル時に検出できない（interfaceのゼロ値として正当な値であり、型チェックはこれを弾かない）。実行時チェック（panic）が、この言語の限界を運用でカバーする現実的な手段である。

## 3. Rejected

1. `nil`を許容し内部で`if retry != nil`分岐する案（前段Decisionで一度採用） — Composition Rootの結線漏れが無言で失われる（Retry通知が黙って消えるだけで、誰も気づけない）。
2. `retry`を`(..., error)`の戻り値で扱う案（constructor全体をerror-returningに変更） — Composition Root内の他のconstructor（`newGeminiTextWriterPrimary`等）は単一戻り値のsignatureで統一されており、この1箇所だけerror伝播経路を持たせると設計の非対称が生まれる。「呼び出し契約違反はpanic」という一般的なGo慣習にも合致しないため見送る。
3. `staticcheck`等のlinterでnil渡しを機械検出する案 — 有効な追加の防御層だが（別Decision・タスクで`.golangci.yml`へ`staticcheck`導入を別途検討する）、既存コードのconstructorでのpanicチェックに代替するものではない。testが将来のコード変更まで保証しないのと同様、静的解析も全パターンを保証しない（soundnessが無い）。両方を防御の異なる層として併用する。
