import { NotFoundError, UnavailableError } from "../../../contracts/index.ts";
import { EpisodeNotFoundError } from "../entities/errors/episode-not-found-error.ts";

type InternalErrorKind = "domain_not_found" | "other";

function classifyInternalError(error: unknown): InternalErrorKind {
  if (error instanceof EpisodeNotFoundError) {
    return "domain_not_found";
  }
  return "other";
}

export function mapInternalErrorToExternal(error: unknown): NotFoundError | UnavailableError {
  const kind = classifyInternalError(error);
  switch (kind) {
    case "domain_not_found":
      return new NotFoundError("エピソードが無い", { cause: error });
    case "other":
      return new UnavailableError("利用できない", { cause: error });
    default: {
      const exhaustive: never = kind;
      return new UnavailableError("利用できない", { cause: exhaustive });
    }
  }
}
