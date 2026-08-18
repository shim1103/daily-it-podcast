import { GetEpisodeResponseSchema } from "../../../../contracts/index.ts";

export const ManuscriptSchema = GetEpisodeResponseSchema.omit({ audioRef: true });
