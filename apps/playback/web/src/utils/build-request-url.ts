/**
 * baseUrl と契約 path を 1 本の request URL へ繋ぐ。
 *
 * @require path は契約由来であり `/` から始まる
 * @ensure baseUrl と path の間の `/` は 1 つだけになる
 * @invariant path 側は書き換えない
 */
export function buildRequestUrl(baseUrl: string, path: string): string {
  // why: path が必ず `/` 始まりなので、baseUrl 末尾の `/` を落とす 1 規則で足りる。URL class は
  //   baseUrl 側の path 段を捨てるため使わない
  return `${baseUrl.replace(/\/+$/, "")}${path}`;
}
