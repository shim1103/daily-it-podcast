/**
 * 入力が HTTP 契約の schema に不適合な時の External Error。
 *
 * @require message は診断用であり応答 JSON には載せない
 * @ensure name は ValidationError。cause で元の失敗を保持できる
 * @invariant 独自 field を持たない
 */
export class ValidationError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "ValidationError";
  }
}

/**
 * 対象 episode が外から見て存在しない時の External Error。
 *
 * @require message は診断用であり応答 JSON には載せない
 * @ensure name は NotFoundError。cause で元の失敗を保持できる
 * @invariant 独自 field を持たない
 */
export class NotFoundError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "NotFoundError";
  }
}

/**
 * 依存先が一時的に使えない時の External Error。
 *
 * @require message は診断用であり応答 JSON には載せない
 * @ensure name は UnavailableError。cause で元の失敗を保持できる
 * @invariant 独自 field を持たない
 */
export class UnavailableError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "UnavailableError";
  }
}
