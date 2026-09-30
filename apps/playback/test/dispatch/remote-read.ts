import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { unstable_readConfig as readConfig } from "wrangler";
import {
  EPISODE_PROGRESS_D1_BINDING,
  EPISODE_PROGRESS_TABLE,
  episodeProgressColumns,
} from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";
import { appRoot, smokeConfigPath } from "./smoke-config.ts";

/**
 * binding とは独立な経路（wrangler CLI の `--remote`）で、実 Cloudflare 側の状態を読む。
 * binding が local の模擬に落ちていれば、ここでは読めず失敗する。
 */

const execFileAsync = promisify(execFile);

/** vitest の testTimeout（60 秒）より短くして、CLI の停止を test の失敗として観測する。 */
const WRANGLER_TIMEOUT_MS = 45_000;

const EPISODES_BINDING_NAME = "EPISODES";

type R2BucketConfig = { binding: string; bucket_name?: string };

type D1CountOutput = Array<{ results: Array<{ n: number }> }>;

function isD1CountOutput(value: unknown): value is D1CountOutput {
  return (
    Array.isArray(value) &&
    value.length === 1 &&
    Array.isArray(value[0]?.results) &&
    value[0].results.length === 1 &&
    typeof value[0].results[0]?.n === "number"
  );
}

async function runWrangler(args: readonly string[]): Promise<string> {
  const { stdout } = await execFileAsync("npx", ["wrangler", ...args], {
    cwd: appRoot,
    timeout: WRANGLER_TIMEOUT_MS,
  });
  return stdout;
}

/** 実 TEST D1 に、その episode_id の行が何件あるか。 */
export async function countRemoteProgressRows(episodeId: string): Promise<number> {
  // why: episodeId は `d1-smoke-probe-<数字|manual>` の生成値で、SQL の区切り文字を含まない
  const sql = `SELECT COUNT(*) AS n FROM ${EPISODE_PROGRESS_TABLE} WHERE ${episodeProgressColumns.episodeId} = '${episodeId}'`;
  const stdout = await runWrangler([
    "d1",
    "execute",
    EPISODE_PROGRESS_D1_BINDING,
    "--remote",
    "--config",
    smokeConfigPath,
    "--json",
    "--command",
    sql,
  ]);
  const parsed: unknown = JSON.parse(stdout);
  if (!isD1CountOutput(parsed)) {
    throw new Error("d1 execute の出力が想定の形（results[0].n）ではない");
  }
  return parsed[0].results[0].n;
}

/** 実 TEST bucket の object 本文。無ければ CLI が失敗して throw する。 */
export async function readRemoteEpisodesObject(key: string): Promise<string> {
  // why: wrangler の Config 型は r2_buckets を解決できず any になる。読む 2 field だけを型で固定する
  const buckets: R2BucketConfig[] = readConfig({ config: smokeConfigPath }).r2_buckets;
  const bucketName = buckets.find(
    (bucket) => bucket.binding === EPISODES_BINDING_NAME,
  )?.bucket_name;
  if (bucketName === undefined) {
    throw new Error(`${EPISODES_BINDING_NAME} の bucket_name が config に無い`);
  }
  return runWrangler([
    "r2",
    "object",
    "get",
    `${bucketName}/${key}`,
    "--remote",
    "--pipe",
    "--config",
    smokeConfigPath,
  ]);
}
