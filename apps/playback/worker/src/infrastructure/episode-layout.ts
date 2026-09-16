// why: 値の正本は repo 根 `contracts/episode-layout.md`。HTTP contracts は Infra から import 禁止（dependency-cruiser）
export const jsonExtension = ".json";
export const audioExtension = ".mp3";

/**
 * 配置契約（`contracts/episode-layout.md`）が定める命名から episodeId（stem）を取り出す。
 *
 * @require extension は `jsonExtension` または `audioExtension`
 * @ensure name が extension で終わる時は拡張子を除いた stem、そうでない時は undefined を返す
 */
export function stemOf(name: string, extension: string): string | undefined {
  /* v8 ignore next 3 -- 呼び出し元は endsWith 判定を通過済みの name だけを渡すため、この分岐は実行時に到達しない */
  if (!name.endsWith(extension)) {
    return undefined;
  }
  return name.slice(0, name.length - extension.length);
}
