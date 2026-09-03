"""
name: diagrams-runtime
description: daily-it-podcast の runtime 構成図を diagrams で生成する。
when_to_use: architecture 図 PNG を生成・更新する時に使う。
layer: 5
links:
  - "[[icons]]"

@require Graphviz の `dot` と `rsvg-convert` が PATH にある。Custom icon は
  `icon_path` と `rasterize` が解決できる。
@ensure `render` は `docs/architecture/runtime.png` を書き、その Path を返す。
@invariant Custom node 名は `CUSTOM_NODE_NAMES` が単一の定義源であり、catalog 外の名前を
  図へ直接書かない。
"""

from __future__ import annotations

from pathlib import Path

CUSTOM_NODE_NAMES: tuple[str, ...] = (
    "cloudflare-workers",
    "cloudflare-access",
    "cursor",
    "google-drive",
    "gemini",
    "hono",
)

_REPO_ROOT = Path(__file__).resolve().parents[2]
_OUTPUT = _REPO_ROOT / "docs" / "architecture" / "runtime"


def render() -> Path:
    """runtime 構成図 PNG を生成して path を返す。

    @require `dot` と `rsvg-convert` が実行できる。ネット未接続でも、icon cache があれば生成できる。
    @ensure 戻り値の file が存在し、拡張子は `.png` である。
    @invariant 出力先は `_OUTPUT.png` に固定する。
    """
    from diagrams import Cluster, Diagram, Edge
    from diagrams.custom import Custom
    from diagrams.onprem.ci import GithubActions
    from diagrams.onprem.client import Users
    from diagrams.programming.language import Go
    from diagrams.programming.framework import React
    from diagrams.saas.cdn import Cloudflare

    from icons import icon_path, rasterize

    _OUTPUT.parent.mkdir(parents=True, exist_ok=True)

    font = "Hiragino Sans"
    graph_attr = {
        "fontname": font,
        "fontsize": "16",
        "pad": "0.6",
        "nodesep": "0.6",
        "ranksep": "0.9",
        "bgcolor": "white",
        "splines": "spline",
    }
    node_attr = {"fontname": font, "fontsize": "11"}
    edge_attr = {"fontname": font, "fontsize": "10"}

    workers_icon = str(rasterize(icon_path("cloudflare-workers")))
    access_icon = str(rasterize(icon_path("cloudflare-access")))
    cursor_icon = str(rasterize(icon_path("cursor")))
    drive_icon = str(rasterize(icon_path("google-drive")))
    gemini_icon = str(rasterize(icon_path("gemini")))
    hono_icon = str(rasterize(icon_path("hono")))

    with Diagram(
        "daily-it-podcast runtime",
        filename=str(_OUTPUT),
        show=False,
        direction="LR",
        graph_attr=graph_attr,
        node_attr=node_attr,
        edge_attr=edge_attr,
        outformat="png",
    ):
        user = Users("リスナー")

        with Cluster("情報源"):
            sources = GithubActions("HackerNews / Lobsters\nITmedia NEWS")

        with Cluster("Generator (Go + GHA cron)"):
            actions = GithubActions("GitHub Actions")
            go_cli = Go("Go CLI")
            cursor_api = Custom("Cursor Cloud Agents\n(原稿)", cursor_icon)
            gemini = Custom("Gemini TTS\n(音声)", gemini_icon)

        drive = Custom("Google Drive\n(音声 + 原稿)", drive_icon)

        with Cluster("Playback (Cloudflare)"):
            access = Custom("Cloudflare Access\n(入場制御)", access_icon)
            cdn = Cloudflare("DNS / CDN")
            workers = Custom("Workers + Hono\n(Drive 代理 BFF)", workers_icon)
            react = React("Vite + React\n(再生 UI)")

        # 生成フロー
        sources >> Edge(label="取得") >> go_cli
        actions >> Edge(label="cron / 手動") >> go_cli
        go_cli >> Edge(label="原稿生成") >> cursor_api
        go_cli >> Edge(label="TTS") >> gemini
        go_cli >> Edge(label="保存") >> drive

        # 再生フロー
        user >> Edge(label="アクセス") >> access >> cdn >> react
        react >> Edge(label="HTTP") >> workers
        workers >> Edge(label="Drive 読取") >> drive

    return Path(str(_OUTPUT) + ".png")


if __name__ == "__main__":
    print(render())
