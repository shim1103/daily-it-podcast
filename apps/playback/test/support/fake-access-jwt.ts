function encodeSegment(value: unknown): string {
  return Buffer.from(JSON.stringify(value)).toString("base64url");
}

/** 署名を検証しない、`exp` だけを持つ偽の Access JWT。test 専用。 */
export function createFakeAccessJwt(expiresAtSec: number): string {
  return `${encodeSegment({ alg: "RS256" })}.${encodeSegment({ exp: expiresAtSec })}.signature`;
}

/** `exp` を持たない payload の偽 JWT。期限が読めない入力の test 用。 */
export function createFakeAccessJwtWithoutExpiry(): string {
  return `${encodeSegment({ alg: "RS256" })}.${encodeSegment({ sub: "someone" })}.signature`;
}
