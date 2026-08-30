import { mapHttpStatusToError } from "../../../contracts/index.ts";
import type { PlaybackHttpErrorCode } from "../../../contracts/index.ts";

/** worker との HTTP 契約に現れる失敗を、web 側の語彙で言い直した code。 */
export const contractOriginErrorCodes = [
  "episode_not_found",
  "validation_error",
  "configuration_error",
  "unavailable",
] as const;

export type ContractOriginErrorCode = (typeof contractOriginErrorCodes)[number];

/** web 側だけで起き、worker との HTTP 契約には現れない失敗の code。 */
export const clientOnlyErrorCodes = ["client_error", "network_error", "invalid_response"] as const;

export type ClientOnlyErrorCode = (typeof clientOnlyErrorCodes)[number];

/** web 内部で閉じた失敗の語彙。UI 向け表示文への変換は担わない。 */
export const playbackApiErrorCodes = [
  ...contractOriginErrorCodes,
  ...clientOnlyErrorCodes,
] as const;

export type PlaybackApiErrorCode = (typeof playbackApiErrorCodes)[number];

/**
 * 契約 code を web 側 code へ 1 対 1 で対応付ける表。
 *
 * why: 契約 code をそのまま web の語彙として使うと、契約 enum の改名や意味変更が web 全体へ直接漏れる。
 *   同名でも別 layer の語として宣言し、写像を 1 箇所へ集める
 */
export const contractErrorCodeMapping: {
  readonly [K in PlaybackHttpErrorCode]: ContractOriginErrorCode;
} = {
  episode_not_found: "episode_not_found",
  validation_error: "validation_error",
  configuration_error: "configuration_error",
  unavailable: "unavailable",
};

/**
 * 非成功 HTTP status を web 側で閉じた API error code へ写す。
 *
 * @require status は Response.ok が false の HTTP 応答の status
 * @ensure 契約 code は web 側 code、未知の 4xx は client_error、その他は unavailable を返す
 * @invariant status 番号を契約側で再判定しない
 */
export function mapHttpStatusToApiError(status: number): PlaybackApiErrorCode {
  const contractCode = mapHttpStatusToError(status);
  if (contractCode !== undefined) {
    return contractErrorCodeMapping[contractCode];
  }
  if (status >= 400 && status < 500) {
    return "client_error";
  }
  return "unavailable";
}
