import type { PlaybackHttpErrorCode } from "../../../contracts/index.ts";

type ExternalErrorName =
  | "ValidationError"
  | "NotFoundError"
  | "ConfigurationError"
  | "UnavailableError";

type HttpErrorMapping = {
  readonly status: number;
  readonly code: PlaybackHttpErrorCode;
};

const externalHttpErrorMapping: {
  readonly [K in ExternalErrorName]: HttpErrorMapping;
} = {
  ValidationError: { status: 400, code: "validation_error" },
  NotFoundError: { status: 404, code: "episode_not_found" },
  ConfigurationError: { status: 500, code: "configuration_error" },
  UnavailableError: { status: 503, code: "unavailable" },
};

type CauseLog = {
  name: string;
  message: string;
  cause?: CauseLog;
};

type ErrorLogPayload = {
  name: string;
  message: string;
  stack: string | undefined;
  cause: CauseLog | undefined;
  requestId: string;
};

function isMappedExternalErrorName(name: string): name is ExternalErrorName {
  return Object.hasOwn(externalHttpErrorMapping, name);
}

function toCauseLog(cause: unknown): CauseLog | undefined {
  if (!(cause instanceof Error)) {
    return undefined;
  }
  const nested = toCauseLog(cause.cause);
  if (nested === undefined) {
    return { name: cause.name, message: cause.message };
  }
  return { name: cause.name, message: cause.message, cause: nested };
}

function toErrorLogPayload(error: Error, requestId: string): ErrorLogPayload {
  return {
    name: error.name,
    message: error.message,
    stack: error.stack,
    cause: toCauseLog(error.cause),
    requestId,
  };
}

function logUnmappedError(error: unknown, requestId: string): void {
  if (error instanceof Error) {
    console.error({ ...toErrorLogPayload(error, requestId), name: "UnmappedError" });
    return;
  }
  console.error({
    name: "UnmappedError",
    message: String(error),
    requestId,
  });
}

export function createHttpErrorResponse(error: unknown, requestId: string): Response {
  if (error instanceof Error && isMappedExternalErrorName(error.name)) {
    const mapped = externalHttpErrorMapping[error.name];
    console.error(toErrorLogPayload(error, requestId));
    return Response.json({ code: mapped.code }, { status: mapped.status });
  }

  logUnmappedError(error, requestId);
  // why: 未知 Error は契約外のため、unavailable へ誤分類せず契約 enum を捏造しない。
  return new Response(null, { status: 500 });
}
