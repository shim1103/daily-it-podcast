import { listEpisodesPath } from "../../../contracts/index.ts";

export type MatchedRoute =
  | { kind: "list" }
  | { kind: "get"; episodeId: unknown }
  | { kind: "audio"; episodeId: unknown }
  | { kind: "unmatched" };

function decodePathSegment(segment: string): string {
  try {
    return decodeURIComponent(segment);
  } catch {
    return segment;
  }
}

export function matchPlaybackRoute(method: string, pathname: string): MatchedRoute {
  if (method !== "GET") {
    return { kind: "unmatched" };
  }
  if (pathname === listEpisodesPath) {
    return { kind: "list" };
  }
  const prefix = `${listEpisodesPath}/`;
  if (!pathname.startsWith(prefix)) {
    return { kind: "unmatched" };
  }
  const segments = pathname.slice(prefix.length).split("/");
  if (segments.length === 1) {
    return { kind: "get", episodeId: decodePathSegment(segments[0] ?? "") };
  }
  if (segments.length === 2 && segments[1] === "audio") {
    return { kind: "audio", episodeId: decodePathSegment(segments[0] ?? "") };
  }
  return { kind: "unmatched" };
}
