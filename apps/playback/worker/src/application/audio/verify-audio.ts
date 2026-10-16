import type { EpisodeId } from "../../../../domain/index.ts";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import type { EpisodeAudio } from "../../entities/models/episode-audio.ts";

const id3TagSignature = [0x49, 0x44, 0x33] as const;
const frameSyncMask = 0xe0;

function startsWithId3Tag(bytes: Uint8Array): boolean {
  return id3TagSignature.every((expected, index) => bytes[index] === expected);
}

function startsWithFrameSync(bytes: Uint8Array): boolean {
  const second = bytes[1];
  return bytes[0] === 0xff && second !== undefined && (second & frameSyncMask) === frameSyncMask;
}

/**
 * Driven Adapter が取得した生 byte を、空でなく mp3 の署名を持つ検証済み音声へ変換する。
 *
 * R2 / in-memory どちらの Adapter の戻りにも使える純関数で、副作用も Port 依存も持たない。
 * 署名は ID3v2 tag（先頭 3 byte が "ID3"）か MPEG audio の frame sync（先頭 byte が 0xFF で
 * 2 byte 目の上位 3 bit が立つ）のどちらか。署名の判定は先頭 byte のみ。再生可能性は保証しない。
 *
 * 複合失敗時の判定順: 空を先に判定し、通過後に署名を判定する。
 *
 * @require episodeId は取得元の episodeId（= error message の識別用）
 * @ensure 適合時は引数の `bytes` をそのまま（変換・コピーなしで）返す。空・署名なしは
 *   {@link EpisodeContentError} を throw し、理由は message で区別できる
 */
export function verifyAudio(bytes: Uint8Array, episodeId: EpisodeId): EpisodeAudio {
  if (bytes.byteLength === 0) {
    throw new EpisodeContentError(`音声が空: ${episodeId}`);
  }
  if (!startsWithId3Tag(bytes) && !startsWithFrameSync(bytes)) {
    throw new EpisodeContentError(`音声が mp3 でない: ${episodeId}`);
  }
  return bytes;
}
