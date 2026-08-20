import type { PlaybackHttpErrorCode } from "./http.ts";

export type HttpStatusClassification =
  | { kind: "success" }
  | { kind: "error"; code: PlaybackHttpErrorCode }
  | { kind: "client_error" };

const knownHttpStatus = {
  200: { kind: "success" },
  400: { kind: "error", code: "validation_error" },
  404: { kind: "error", code: "episode_not_found" },
  500: { kind: "error", code: "configuration_error" },
  503: { kind: "error", code: "unavailable" },
} as const satisfies Record<number, HttpStatusClassification>;

type KnownHttpStatus = keyof typeof knownHttpStatus;

function isKnownHttpStatus(status: number): status is KnownHttpStatus {
  return Object.hasOwn(knownHttpStatus, status);
}

/**
 * 応答の HTTP status を web↔worker 契約の分類へ写す。
 *
 * @require status は HTTP 応答の status
 * @ensure 宣言表にある番号はその行。無い番号は floor(status / 100) の級
 * @invariant 既知の 404 を 400 へ畳まない
 */
export function classifyHttpStatus(status: number): HttpStatusClassification {
  if (isKnownHttpStatus(status)) {
    return knownHttpStatus[status];
  }

  const httpClass = Math.floor(status / 100);
  if (httpClass === 2) {
    return { kind: "success" };
  }
  if (httpClass === 4) {
    return { kind: "client_error" };
  }
  return { kind: "error", code: "unavailable" };
}
