/**
 * Playback Worker 内部の runtime config 検証失敗。
 *
 * @require message は secret 値を含めず、欠落した key の診断だけを表す
 * @ensure HTTP の External Error を知らず、cause chain を保持できる
 * @invariant 全ての runtime config 原因はこの Error class に統一する
 */
export class PlaybackRuntimeConfigError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "PlaybackRuntimeConfigError";
  }
}
