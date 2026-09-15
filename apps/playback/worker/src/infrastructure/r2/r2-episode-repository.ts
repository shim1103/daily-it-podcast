import type {
  EpisodeRepository,
  RawManuscriptEntry,
} from "../../application/ports/episode-repository.ts";
import { audioExtension, jsonExtension, stemOf } from "../episode-layout.ts";
import { R2Error } from "./r2-error.ts";

/**
 * R2 object の本文取得結果。Cloudflare Workers 本番の `R2ObjectBody` の読み出し面の最小部分集合。
 */
export type R2ObjectBodyLike = {
  arrayBuffer(): Promise<ArrayBuffer>;
};

/**
 * Workers の R2 binding 最小面。本番型は wrangler 生成の `R2Bucket`（`apps/playback/worker-configuration.d.ts`）
 * の読み出し面に合わせる。put は generator 側の書込 Adapter が使うため、ここでは持たない。
 *
 * why: 配置契約（`contracts/episode-layout.md`）は object 空間を flat に保つため、`list` は
 * `prefix` を受け取らない。将来階層化する時に絞り込みが要る（YAGNI）。
 */
export type R2BucketBinding = {
  get(key: string): Promise<R2ObjectBodyLike | null>;
  list(): Promise<{ objects: Array<{ key: string }> }>;
};

function isR2ObjectBodyLike(value: unknown): value is R2ObjectBodyLike {
  return (
    value !== null &&
    typeof value === "object" &&
    typeof (value as { arrayBuffer?: unknown }).arrayBuffer === "function"
  );
}

/**
 * R2 binding（`R2BucketBinding`）で `EpisodeRepository` を満たす本番 Driven Adapter。
 *
 * 真の外部境界の I/O（list・get・bytes 取り出し）だけを担い、取得した原稿 json は decode したまま
 * 返す。schema 適合・stem 一致・不正 JSON・mp3 欠落の判定はしない（use-case が行う）。
 *
 * why: R2 binding は HTTP ではなく Workers runtime の JS 関数呼び出し契約であり、controllable な
 * local listener へ実送信する Narrow Integration の手法（`GoogleDriveEpisodeRepository` 参照）が
 * 使えない。実 binding runtime（Miniflare）を挟む案は wrangler の transitive dependency への
 * phantom import と内部限定 API 依存を伴い costに見合わないため採らず、Sociable Unit Test
 * （`r2-episode-repository.sociable_unit.test.ts`）1本へ寄せる。
 *
 * @require deps.bucket は Worker の R2 binding（配置契約は `contracts/episode-layout.md`）
 * @ensure R2 I/O 自体の失敗（list・get の例外・非 null だが不正形状の応答）は R2Error を throw する
 * @invariant R2 object key の実値を Error message に含めない
 */
export class R2EpisodeRepository implements EpisodeRepository {
  private readonly bucket: R2BucketBinding;

  constructor(deps: { bucket: R2BucketBinding }) {
    this.bucket = deps.bucket;
  }

  async listManuscripts(): Promise<RawManuscriptEntry[]> {
    const keys = await this.listJsonKeys();

    const entries = await Promise.all(
      keys.map(async (key) => {
        const stem = stemOf(key, jsonExtension);
        /* v8 ignore next 3 -- keys は直前で同じ jsonExtension の endsWith 判定を通過済みのため、この分岐は実行時に到達しない */
        if (stem === undefined) {
          return undefined;
        }
        const json = await this.downloadJson(key);
        return { stem, json } satisfies RawManuscriptEntry;
      }),
    );

    return entries.filter((entry): entry is RawManuscriptEntry => entry !== undefined);
  }

  async getAudio(episodeId: string): Promise<Uint8Array | undefined> {
    const key = `${episodeId}${audioExtension}`;
    const body = await this.get(key);
    if (body === null) {
      return undefined;
    }
    return this.readBytes(body);
  }

  /**
   * json object を取得し、use-case へ渡す生 payload へ best-effort で decode する。
   *
   * why: bytes → 値の decode までが wire format の I/O。「JSON として妥当か」「原稿として適合か」の
   * 判定はこの Adapter の責務ではないため、JSON.parse に失敗しても分類せず decode した生文字列を
   * そのまま返す（use-case の schema 判定が非 object として弾く）。
   */
  private async downloadJson(key: string): Promise<unknown> {
    const body = await this.get(key);
    if (body === null) {
      throw new R2Error("R2 一覧に含まれる object が取得時に見当たらない");
    }
    const bytes = await this.readBytes(body);
    const text = new TextDecoder().decode(bytes);
    try {
      return JSON.parse(text);
    } catch {
      return text;
    }
  }

  private async listJsonKeys(): Promise<string[]> {
    let result: { objects: Array<{ key: string }> };
    try {
      result = await this.bucket.list();
    } catch (cause) {
      throw new R2Error("R2 一覧取得に失敗", { cause });
    }
    return result.objects.map((object) => object.key).filter((key) => key.endsWith(jsonExtension));
  }

  private async get(key: string): Promise<R2ObjectBodyLike | null> {
    let body: R2ObjectBodyLike | null;
    try {
      body = await this.bucket.get(key);
    } catch (cause) {
      throw new R2Error("R2 object 取得に失敗", { cause });
    }
    if (body === null) {
      return null;
    }
    if (!isR2ObjectBodyLike(body)) {
      throw new R2Error("R2 object の応答形式が不正");
    }
    return body;
  }

  private async readBytes(body: R2ObjectBodyLike): Promise<Uint8Array> {
    try {
      const buffer = await body.arrayBuffer();
      return new Uint8Array(buffer);
    } catch (cause) {
      throw new R2Error("R2 object の読み出しに失敗", { cause });
    }
  }
}
