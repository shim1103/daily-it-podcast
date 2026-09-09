"""
name: diagrams-runtime
description: daily-it-podcast の runtime 構成図を diagrams で生成する。
when_to_use: architecture 図 PNG を生成・更新する時に使う。
layer: 5
links:
  - "[[icons]]"

生成物 `apps/diagrams/runtime.png` は runtime 構成図の SSoT。README は置かず、参照元は
本 module。PNG は手編集しない。変更は `apps/diagrams/**` へ入れて再生成する。

生成手順:
  cd apps/diagrams && python runtime.py

@require Graphviz の `dot` と `rsvg-convert` が PATH にある。Custom icon は
  `icon_path` と `rasterize` が解決できる。
@ensure `render` は `apps/diagrams/runtime.png` を書き、その Path を返す。
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
    "github-actions",
    "hackernews",
    "lobsters",
    "itmedia-rss",
    "vite",
)

_OUTPUT = Path(__file__).resolve().parent / "runtime"


def render() -> Path:
    """runtime 構成図 PNG を生成して path を返す。

    @require `dot` と `rsvg-convert` が実行できる。ネット未接続でも、icon cache があれば生成できる。
    @ensure 戻り値の file が存在し、拡張子は `.png` である。
    @invariant 出力先は `_OUTPUT.png` に固定する。
    """
    from diagrams import Cluster, Diagram, Edge
    from diagrams.custom import Custom
    from diagrams.onprem.client import Users
    from diagrams.programming.language import Go, TypeScript
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
    gha_icon = str(rasterize(icon_path("github-actions")))
    hn_icon = str(rasterize(icon_path("hackernews")))
    lobsters_icon = str(rasterize(icon_path("lobsters")))
    itmedia_icon = str(rasterize(icon_path("itmedia-rss")))
    vite_icon = str(rasterize(icon_path("vite")))

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
        user = Users("Listener")

        with Cluster("Sources"):
            hn = Custom("Hacker News\n(Firebase API)", hn_icon)
            lobsters = Custom("Lobsters\n(hottest.json)", lobsters_icon)
            itmedia = Custom("ITmedia NEWS\n(RSS 2.0)", itmedia_icon)
            # why: 3 サイトは対等な並列取得元。invisible edge で同一 rank に固定し横並びにする。
            hn - Edge(style="invis") - lobsters - Edge(style="invis") - itmedia

        with Cluster("Generator (Go + GHA cron)"):
            actions = Custom("GitHub Actions\n(cron / manual)", gha_icon)
            go_cli = Go("Go CLI")
            # why: 原稿は Cursor Cloud Agents が第一経路。枯渇時は Gemini へ fallback する
            #   （text_writer fallback）。どちらも「原稿」を作るので両方を原稿経路として描く。
            cursor_api = Custom("Cursor Cloud Agents\n(Script, primary)", cursor_icon)
            gemini_text = Custom("Gemini\n(Script, fallback)", gemini_icon)
            gemini_tts = Custom("Gemini TTS\n(Audio)", gemini_icon)

        drive = Custom("Google Drive\n(Audio + Script)", drive_icon)

        with Cluster("Playback (Cloudflare) — TypeScript"):
            access = Custom("Cloudflare Access\n(OAuth 2.0 access control)", access_icon)
            cdn = Cloudflare("DNS / CDN")
            # why: web / worker / contracts すべて TS。contracts の zod schema と hc<AppType> の
            #   型共有が構成の要なので、generator 側の Go node と対称に言語 node を 1 個置く。
            ts = TypeScript("TypeScript\n(contracts type sharing)")
            with Cluster("Playback UI"):
                vite = Custom("Vite\n(bundle build / serve)", vite_icon)
                react = React("React\n(rendering, playback control)")
                vite >> Edge(label="bundle", style="dashed") >> react
            with Cluster("Workers (Drive proxy BFF)"):
                workers = Custom("Cloudflare Workers\n(execution runtime)", workers_icon)
                # why: Hono の route は 2 本。/episodes は RPC client（hc<AppType>）が叩く JSON、
                #   /episodes/:id/audio は素の HTTP GET（audio/wav・Range 部分応答）で RPC を経由しない。
                hono = Custom("Hono\n(routes: RPC + audio GET)", hono_icon)

        # 生成フロー
        hn >> Edge(label="fetch") >> go_cli
        lobsters >> Edge(label="fetch") >> go_cli
        itmedia >> Edge(label="fetch") >> go_cli
        actions >> Edge(label="cron / manual") >> go_cli
        go_cli >> Edge(label="generate script") >> cursor_api
        go_cli >> Edge(label="fallback on exhaustion", style="dashed") >> gemini_text
        go_cli >> Edge(label="TTS") >> gemini_tts
        go_cli >> Edge(label="save (OAuth 2.0)") >> drive

        # 再生フロー
        user >> Edge(label="access") >> access >> cdn >> vite
        ts >> Edge(label="type", style="dotted") >> react
        ts >> Edge(label="type", style="dotted") >> hono
        react >> Edge(label="list JSON (Hono RPC)") >> hono
        react >> Edge(label="audio WAV / Range (HTTP GET)") >> hono
        workers >> Edge(label="execute", style="dashed") >> hono
        hono >> Edge(label="Drive read (OAuth 2.0)") >> drive

    return Path(str(_OUTPUT) + ".png")


if __name__ == "__main__":
    print(render())