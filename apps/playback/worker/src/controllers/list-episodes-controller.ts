import type { ListEpisodesResponse } from "../../../contracts/index.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

export type ListEpisodesUseCase = () => Promise<ListEpisodesResponse>;

export type ListEpisodesController = (body: unknown) => Promise<ListEpisodesResponse>;

/**
 * 一覧 JSON を返す Controller を組み立てる。
 *
 * @require useCase は一覧の ListEpisodesResponse を返す
 * @ensure 戻り関数は unknown を受け、契約の ListEpisodesResponse を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createListEpisodesController(useCase: ListEpisodesUseCase): ListEpisodesController {
  return async function listEpisodesController(_body: unknown): Promise<ListEpisodesResponse> {
    try {
      return await useCase();
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
