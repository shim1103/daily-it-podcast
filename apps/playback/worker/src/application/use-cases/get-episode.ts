import { episodeAudioPath, type GetEpisodeResponse } from "../../../../contracts/index.ts";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import { verifyManuscript } from "../manuscript/verify-manuscript.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";

/**
 * 対象 episodeId の原稿を取得し、検証してから `GetEpisodeResponse` を返す。
 *
 * Port は取得したままの生 json を返すだけなので、schema 適合・stem 一致・wav 有無の判定は
 * この use-case が行う（generator `WriteEpisode.Run` の read 方向鏡像）。
 *
 * 検証順（precedence）: json エントリ在否 → hasAudio → schema → stem。複合失敗時は先の判定が勝つ。
 *
 * @ensure json エントリ無し・wav 欠落・schema 不適合・stem 不一致のいずれも単一の
 *   {@link EpisodeContentError} を throw する。失敗理由は message で区別できる。
 */
export async function getEpisode(
  repository: EpisodeRepository,
  episodeId: string,
): Promise<GetEpisodeResponse> {
  const detail = await repository.getManuscript(episodeId);
  if (detail === undefined) {
    throw new EpisodeContentError(`JSON エントリが無い: ${episodeId}`);
  }
  if (!detail.hasAudio) {
    throw new EpisodeContentError(`wav が無い: ${episodeId}`);
  }
  const manuscript = verifyManuscript(detail.json, episodeId);
  return {
    ...manuscript,
    audioRef: episodeAudioPath(episodeId),
  };
}
