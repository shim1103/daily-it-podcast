/**
 * 対象 episode の実体（原稿 JSON / wav）が揃わない・内容が契約に適合しない時の Domain Error。
 *
 * 失敗理由（JSON エントリ欠落 / wav 欠落 / schema 不適合 / stem 不一致）は message で分類する。
 * 種別ごとのクラス細分はしない。
 *
 * @require message は診断用。secret / Drive file id を含めない
 * @ensure name は EpisodeContentError。cause で元の失敗を保持できる
 * @invariant 独自 property を持たない（文脈は cause chain で保持）
 */
export class EpisodeContentError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "EpisodeContentError";
  }
}
