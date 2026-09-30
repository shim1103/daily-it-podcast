/**
 * 標準入力などの byte 列の stream を最後まで読み、前後の空白を除いた文字列にする。
 * chunk ごとに decode すると、多 byte 文字が chunk の境目で壊れるため、連結してから decode する。
 */
export async function readAllText(stream: AsyncIterable<Buffer>): Promise<string> {
  const chunks: Buffer[] = [];
  for await (const chunk of stream) {
    chunks.push(chunk);
  }
  return Buffer.concat(chunks).toString("utf8").trim();
}
