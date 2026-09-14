import type {
  EpisodeRepository,
  RawManuscriptEntry,
} from "../../application/ports/episode-repository.ts";

/**
 * Workers の R2 binding 最小面。本番型は wrangler 生成の `R2Bucket` に合わせる（C）。
 * A stub は binding 未結線でも compile できるよう、ここの契約だけに依存する。
 */
export type R2BucketBinding = {
  get(key: string): Promise<unknown>;
  put(key: string, value: ReadableStream | ArrayBuffer | string | null): Promise<unknown>;
  list(options?: { prefix?: string }): Promise<{ objects: Array<{ key: string }> }>;
};

/**
 * R2 binding で `EpisodeRepository` を満たす本番 Adapter の stub。
 * list / get の挙動本実装は C（Issue）側。
 *
 * @require deps.bucket は Worker の R2 binding（C で結線）
 * @ensure 現 stub は空一覧と undefined 音声だけを返す
 */
export class R2EpisodeRepository implements EpisodeRepository {
  private readonly bucket: R2BucketBinding;

  constructor(deps: { bucket: R2BucketBinding }) {
    this.bucket = deps.bucket;
  }

  async listManuscripts(): Promise<RawManuscriptEntry[]> {
    void this.bucket;
    return [];
  }

  async getAudio(_episodeId: string): Promise<Uint8Array | undefined> {
    void this.bucket;
    return undefined;
  }
}
