import {
  GetEpisodeRequestSchema,
  ValidationError,
  type GetEpisodeRequest,
} from "../../../contracts/index.ts";

export function parseGetEpisodeRequest(body: unknown): GetEpisodeRequest {
  try {
    return GetEpisodeRequestSchema.parse(body);
  } catch (error) {
    throw new ValidationError("入力が契約に不適合", { cause: error });
  }
}
