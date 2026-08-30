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

export function createAudioResponse(bytes: Uint8Array): Response {
  return new Response(toArrayBuffer(bytes), {
    status: 200,
    headers: { "Content-Type": episodeAudioContentType },
  });
}
