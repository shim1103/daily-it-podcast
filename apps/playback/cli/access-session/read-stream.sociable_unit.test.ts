import { describe, expect, it } from "vitest";
import { readAllText } from "./read-stream.ts";

async function* chunksOf(...texts: string[]): AsyncGenerator<Buffer> {
  for (const text of texts) {
    yield Buffer.from(text, "utf8");
  }
}

describe("readAllText", () => {
  it("joins_all_chunks_and_trims_the_surrounding_whitespace", async () => {
    // Given: 複数の chunk に分かれ、前後に空白・改行がある入力（日本語の途中で chunk が切れても壊れない）
    const stream = chunksOf("  abc", "def\n");

    // When: 最後まで読む
    const text = await readAllText(stream);

    // Then: chunk が連結され、前後の空白が取り除かれる
    expect(text).toBe("abcdef");
  });

  it("returns_an_empty_string_when_the_stream_has_no_chunks", async () => {
    // Given: 何も流れない入力
    const stream = chunksOf();

    // When: 最後まで読む
    const text = await readAllText(stream);

    // Then: 空文字を返す
    expect(text).toBe("");
  });
});
