import { NotFoundError, UnavailableError } from "../../../contracts/index.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";

type InternalErrorKind = "domain_content" | "other";

function classifyInternalError(error: unknown): InternalErrorKind {
  if (error instanceof EpisodeContentError) {
    return "domain_content";
  }
  return "other";
}

export function mapInternalErrorToExternal(error: unknown): NotFoundError | UnavailableError {
  const kind = classifyInternalError(error);
  switch (kind) {
    case "domain_content":
      return new NotFoundError("エピソードが無い", { cause: error });
    case "other":
      return new UnavailableError("利用できない", { cause: error });
    /* v8 ignore next 4 -- InternalErrorKind は2値の union 型で、型検査上この分岐へ実行が到達しない。将来値が増えた時に tsc が検知するための exhaustiveness check */
    default: {
      const exhaustive: never = kind;
      return new UnavailableError("利用できない", { cause: exhaustive });
    }
  }
}
