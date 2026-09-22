/**
 * 進捗行が存在しないのに update（PATCH）しようとした時の Domain Error。
 *
 * External へは NotFoundError（HTTP 404 / episode_not_found）へ写す。
 * create（POST）が既に行ありの場合は本 Error にしない（冪等成功は Use Case の契約）。
 *
 * @require message は診断用。secret / storage 固有 id を含めない
 * @ensure name は ProgressNotFoundError。cause で元の失敗を保持できる
 * @invariant 独自 property を持たない（文脈は cause chain で保持）
 */
export class ProgressNotFoundError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "ProgressNotFoundError";
  }
}
