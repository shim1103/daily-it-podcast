import { episodeAudioContentType } from "../../../contracts/index.ts";

function toArrayBuffer(bytes: Uint8Array): ArrayBuffer {
  const buffer = bytes.buffer;
  if (typeof SharedArrayBuffer !== "undefined" && buffer instanceof SharedArrayBuffer) {
    const copy = new Uint8Array(bytes.byteLength);
    copy.set(bytes);
    return copy.buffer;
  }
  return (buffer as ArrayBuffer).slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength);
}

/** `parseRange` の判別可能戻り値。`none` は Range 無視（全体を返す）、`unsatisfiable` は 416。 */
type ParsedRange =
  | { kind: "satisfiable"; start: number; end: number }
  | { kind: "unsatisfiable" }
  | { kind: "none" };

/**
 * `Range` request header（`bytes=start-end` 形）を解釈する。
 *
 * @require rangeHeader は non-null（呼び出し側が Range header の有無を判定済み）。totalLength は
 *   0 以上の整数。
 * @ensure header が `bytes=N-M` / `bytes=N-`（終端省略時は末尾まで）の形で、開始位置が
 *   `totalLength` 未満なら `{ kind:"satisfiable", start, end }`（end は totalLength-1 でクランプ、
 *   逆順なら start に丸める）。開始位置が `totalLength` 以上なら `{ kind:"unsatisfiable" }`（RFC 7233）。
 *   単位が bytes でない・複数 range・解釈不能な形式なら `{ kind:"none" }`（Range を無視する合図）。
 */
function parseRange(rangeHeader: string, totalLength: number): ParsedRange {
  const match = /^bytes=(\d+)-(\d*)$/.exec(rangeHeader.trim());
  if (match === null) {
    return { kind: "none" };
  }
  const start = Number(match[1]);
  if (start >= totalLength) {
    return { kind: "unsatisfiable" };
  }
  const requestedEnd = match[2] === "" ? totalLength - 1 : Number(match[2]);
  const end = Math.min(Math.max(requestedEnd, start), totalLength - 1);
  return { kind: "satisfiable", start, end };
}

/**
 * 音声 byte 列から HTTP 応答を組む。`Range` request header があれば範囲だけを `206` で返す
 * （`<audio>` の任意位置 seek に必須。Range 非対応サーバーだと `currentTime` 代入がブラウザに
 * 無視され seek 先ではなく先頭のまま扱われる）。
 *
 * @require bytes は完成音声の byte 列。rangeHeader は request の `Range` header（無ければ null）。
 * @ensure rangeHeader が `bytes=N-M` 形で解釈でき開始位置が bytes.length 未満なら、該当範囲だけを
 *   `206` + `Content-Range: bytes N-M/total` + `Content-Length` で返す。開始位置が bytes.length 以上
 *   なら `416` + `Content-Range: bytes *​/total` を返す。rangeHeader が無い・解釈できない形式なら
 *   bytes 全体を `200` で返す。どの応答も契約 Content-Type と `Accept-Ranges: bytes` を持つ
 *   （後者でブラウザに任意位置 seek が効くことを伝える）。
 */
export function createAudioResponse(bytes: Uint8Array, rangeHeader: string | null): Response {
  const headers: HeadersInit = {
    "Content-Type": episodeAudioContentType,
    "Accept-Ranges": "bytes",
  };

  const range =
    rangeHeader === null ? { kind: "none" as const } : parseRange(rangeHeader, bytes.length);

  if (range.kind === "unsatisfiable") {
    return new Response(null, {
      status: 416,
      headers: { ...headers, "Content-Range": `bytes */${bytes.length}` },
    });
  }
  if (range.kind === "satisfiable") {
    const slice = bytes.subarray(range.start, range.end + 1);
    return new Response(toArrayBuffer(slice), {
      status: 206,
      headers: {
        ...headers,
        "Content-Range": `bytes ${range.start}-${range.end}/${bytes.length}`,
        "Content-Length": String(slice.length),
      },
    });
  }

  return new Response(toArrayBuffer(bytes), {
    status: 200,
    headers,
  });
}
