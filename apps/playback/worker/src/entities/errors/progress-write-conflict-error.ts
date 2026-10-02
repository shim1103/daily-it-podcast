/**
 * 進捗の条件付き書込（CAS）の競合が、最大試行回数まで続いた時の Domain Error。
 *
 * External へは `mapInternalErrorToExternal` の既定経路（"other"）で UnavailableError（HTTP 503）へ写す。
 * 競合は一時的なので、browser の再送（retry）で解消しうる。
 *
 * @require message は診断用。secret / storage 固有 id を含めない
 * @ensure name は ProgressWriteConflictError。cause で元の失敗を保持できる
 * @invariant 独自 property を持たない（文脈は cause chain で保持）
 */
export class ProgressWriteConflictError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "ProgressWriteConflictError";
  }
}
