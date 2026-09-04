import {
  EpisodeIdRequestSchema,
  ValidationError,
  type EpisodeIdRequest,
} from "../../../contracts/index.ts";

export function parseEpisodeIdRequest(body: unknown): EpisodeIdRequest {
  try {
    return EpisodeIdRequestSchema.parse(body);
  } catch (error) {
    throw new ValidationError("入力が契約に不適合", { cause: error });
  }
}
