import { ConfigurationError } from "../../../contracts/index.ts";
import { PlaybackRuntimeConfigError } from "../composition/root.ts";

/**
 * Composition Root の内部 runtime config Error を HTTP boundary 向け External Error へ写す。
 *
 * @require なし
 * @ensure `PlaybackRuntimeConfigError` は `ConfigurationError` へ変換し、元 Error を cause へ残す。
 *   それ以外はそのまま返す
 */
export function mapRuntimeConfigErrorToExternal(error: unknown): unknown {
  if (error instanceof PlaybackRuntimeConfigError) {
    return new ConfigurationError("設定を確認できません", { cause: error });
  }
  return error;
}
