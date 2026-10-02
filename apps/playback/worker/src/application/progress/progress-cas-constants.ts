/**
 * 進捗の条件付き書込（CAS）の再試行方針の定数。
 * 業務規則ではなく UseCase の運用方針なので、Domain の `entities/constants/progress.ts` ではなく Application に置く。
 */

// what: 競合時の read → merge → 条件付き書込 の試行回数。初回を含む。
// why: 一時的な競合を吸収しつつ、待たせすぎない回数にするため。値の正本はこの定数（Decision 2026-10-01T18-54-16）。
export const PROGRESS_CAS_MAX_ATTEMPTS = 3 as const;
