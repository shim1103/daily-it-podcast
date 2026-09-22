import { NotFoundError, UnavailableError, ValidationError } from "../../../contracts/index.ts";
import { EpisodeContentError } from "../entities/errors/episode-content-error.ts";
import { ProgressNotFoundError } from "../entities/errors/progress-not-found-error.ts";
import { ProgressRuleError } from "../entities/errors/progress-rule-error.ts";

type InternalErrorKind =
  | "domain_content"
  | "domain_progress_rule"
  | "domain_progress_not_found"
  | "other";

function classifyInternalError(error: unknown): InternalErrorKind {
  if (error instanceof EpisodeContentError) {
    return "domain_content";
  }
  if (error instanceof ProgressRuleError) {
    return "domain_progress_rule";
  }
  if (error instanceof ProgressNotFoundError) {
    return "domain_progress_not_found";
  }
  return "other";
}

/**
 * Internal Error（Domain / Infrastructure）を External Error へ写す。
 *
 * @ensure 進捗ルール違反は ValidationError、進捗行不在・原稿不在は NotFoundError、それ以外は UnavailableError
 * @invariant HTTP code 語彙は増やさず、既存 External 4 種へ畳む
 */
export function mapInternalErrorToExternal(
  error: unknown,
): ValidationError | NotFoundError | UnavailableError {
  const kind = classifyInternalError(error);
  switch (kind) {
    case "domain_content":
      return new NotFoundError("エピソードが無い", { cause: error });
    case "domain_progress_rule":
      return new ValidationError("進捗の更新を受理できない", { cause: error });
    case "domain_progress_not_found":
      return new NotFoundError("進捗が無い", { cause: error });
    case "other":
      return new UnavailableError("利用できない", { cause: error });
    /* v8 ignore next 4 -- InternalErrorKind は閉じており、型検査上この分岐へ実行が到達しない。将来値が増えた時に tsc が検知するための exhaustiveness check */
    default: {
      const exhaustive: never = kind;
      return new UnavailableError("利用できない", { cause: exhaustive });
    }
  }
}
