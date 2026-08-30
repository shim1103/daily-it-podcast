import type { PlaybackHttpErrorCode } from "./http.ts";

const knownHttpStatus = {
  400: "validation_error",
  404: "episode_not_found",
  500: "configuration_error",
  503: "unavailable",
} as const satisfies Record<number, PlaybackHttpErrorCode>;

type KnownHttpStatus = keyof typeof knownHttpStatus;

function isKnownHttpStatus(status: number): status is KnownHttpStatus {
  return Object.hasOwn(knownHttpStatus, status);
}

/**
 * HTTP status を web↔worker 契約の error code へ写す。
 *
 * @require status は HTTP 応答の status
 * @ensure 宣言表にある番号はその code。未知の status は undefined
 * @invariant 既知の 404 を 400 へ畳まない
 */
export function mapHttpStatusToError(status: number): PlaybackHttpErrorCode | undefined {
  if (isKnownHttpStatus(status)) {
    return knownHttpStatus[status];
  }
  return undefined;
}
