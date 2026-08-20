import type { HttpStatusClassification, PlaybackHttpErrorCode } from "../../../contracts/index.ts";

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
 * 非成功の HTTP status 分類を、web 側で閉じた API error code へ写す。
 *
 * @require classification は classifyHttpStatus が返した success 以外の分類
 * @ensure 契約 code は contractErrorCodeMapping を通した web 側 code、client_error は client_error を返す。
 *   契約に無い kind では TypeError を throw し、既存 code へ倒さない
 * @invariant status 番号を再判定しない
 */
export function toApiErrorCode(
  classification: Exclude<HttpStatusClassification, { kind: "success" }>,
): PlaybackApiErrorCode {
  switch (classification.kind) {
    case "error":
      return contractErrorCodeMapping[classification.code];
    case "client_error":
      return "client_error";
    default: {
      const exhaustive: never = classification;
      // why: caller は code ごとに retry・再入力・報告を選ぶ。未知の kind を既存 code へ倒すと
      //   起きていない失敗を伝え、誤った操作を選ばせる。前提違反として大域脱出する
      throw new TypeError(`未知の HTTP status 分類: ${JSON.stringify(exhaustive)}`);
    }
  }
}
