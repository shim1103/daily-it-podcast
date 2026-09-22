/**
 * 再生進捗の意味ルール違反（skew 超過・merge 後 invariant 破壊など）の Domain Error。
 *
 * 失敗理由は message で分類する。種別ごとのクラス細分はしない。
 * External へは ValidationError（HTTP 400 / validation_error）へ写す。
 *
 * @require message は診断用。secret / storage 固有 id を含めない
 * @ensure name は ProgressRuleError。cause で元の失敗を保持できる
 * @invariant 独自 property を持たない（文脈は cause chain で保持）
 */
export class ProgressRuleError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "ProgressRuleError";
  }
}
