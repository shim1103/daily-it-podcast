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
    "cloudflare-r2",
    "cursor",
    "gemini",
    "hono",
    "github-actions",
    "hackernews",
    "lobsters",
    "techcrunch",
    "publickey",
    "cloudwatch",
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
    from diagrams.onprem.client import Client
    from diagrams.programming.language import Go, TypeScript
    from diagrams.programming.framework import React

    from icons import icon_path, rasterize

    _OUTPUT.parent.mkdir(parents=True, exist_ok=True)

    font = "Hiragino Sans"
    graph_attr = {
        "fontname": font,
        "fontsize": "16",
        "pad": "0.6",
        "nodesep": "0.55",
        "ranksep": "0.9",
        "bgcolor": "white",
        "splines": "spline",
    }
    node_attr = {"fontname": font, "fontsize": "11"}
    edge_attr = {"fontname": font, "fontsize": "10"}

    workers_icon = str(rasterize(icon_path("cloudflare-workers")))
    access_icon = str(rasterize(icon_path("cloudflare-access")))
    r2_icon = str(rasterize(icon_path("cloudflare-r2")))
    cursor_icon = str(rasterize(icon_path("cursor")))
    gemini_icon = str(rasterize(icon_path("gemini")))
    hono_icon = str(rasterize(icon_path("hono")))
    gha_icon = str(rasterize(icon_path("github-actions")))
    hn_icon = str(rasterize(icon_path("hackernews")))
    lobsters_icon = str(rasterize(icon_path("lobsters")))
    publickey_icon = str(rasterize(icon_path("publickey")))
    techcrunch_icon = str(rasterize(icon_path("techcrunch")))
    cloudwatch_icon = str(rasterize(icon_path("cloudwatch")))
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
        # why: single-user app。複数人 icon（Users）ではなく 1 人の Client node にする。
        user = Client("User (Browser)")

        # why: SourceItem を返す 5 Adapter（DESIGN / Decision 2026-09-13T15-08-55）。
        #   ITmedia 単源ではない。JSON API と feed（Atom/RSS/RDF）を並べる。図では
        #   Go CLI への矢印を 1 本へ集約し、内訳は node 名だけで示す。
        with Cluster("Sources"):
            hn = Custom("Hacker News", hn_icon)
            lobsters = Custom("Lobsters", lobsters_icon)
            publickey = Custom("Publickey", publickey_icon)
            techcrunch = Custom("TechCrunch", techcrunch_icon)
            cloudwatch = Custom("Cloud Watch", cloudwatch_icon)
            (
                hn
                - Edge(style="invis")
                - lobsters
                - Edge(style="invis")
                - publickey
                - Edge(style="invis")
                - techcrunch
                - Edge(style="invis")
                - cloudwatch
            )

        with Cluster("Generator (Go + GHA cron)"):
            actions = Custom("GitHub Actions", gha_icon)
            go_cli = Go("Go CLI\n(ProduceEpisode)")
            # why: 原稿は Gemini free 枠が第一経路。枯渇時は Cursor Cloud Agents へ、
            #   さらに枯渇したら Gemini paid 枠（SPARE_GEMINI_API_KEY）へ fallback する
            #   （Decision 2026-09-16T00-39-21）。TTS も同じ Gemini free→paid の 2 段。
            #   Gemini node は primary（free）と Spare（paid, final fallback）で 2 role を持つため
            #   1 node に統合し、fallback 方向を dashed edge で示す。
            gemini_text = Custom("Gemini\n(free, primary)", gemini_icon)
            cursor_api = Custom("Cursor Cloud Agents\n(fallback)", cursor_icon)
            gemini_spare = Custom("Gemini\n(paid, final fallback)", gemini_icon)
            gemini_tts = Custom("Gemini TTS\n(free, primary)", gemini_icon)
            ffmpeg = Go("ffmpeg Adapter")

        # why: R2 は Workers isolate の外（Cloudflare 側が持つ独立した storage service）。
        #   generator からは S3 互換 API（Internet 経由）、playback からは Worker binding
        #   （Cloudflare 内部、Internet を経由しない）という異なる経路で同じ bucket へ到達する
        #   ため、isolate を持つ Playback Cluster の外に独立 node として置く。
        with Cluster("Cloudflare"):
            r2 = Custom("R2\n({id}.json + {id}.mp3)", r2_icon)

            with Cluster("Playback — TypeScript"):
                # why: Access が TLS 終端・認証を行い、assets 配信か Worker 起動かを Edge 側で
                #   振り分ける。この Edge routing 自体は runtime 図の主題ではないため 1 node に畳む。
                access = Custom("Cloudflare Edge\n(TLS, Access auth, routing)", access_icon)

                # why: contracts の zod schema と hc<AppType> の型共有が React ↔ Hono の構成の要。
                #   generator 側の Go node と対称に言語 node を 1 個置き、両側へ type edge を引く。
                ts = TypeScript("TypeScript\n(contracts type sharing)")

                # why: Vite は build time のみ動く。実行時 runtime には存在しないため、
                #   React（source）→ Vite（build）→ static assets の順で生成物と分けて描く。
                react = React("React")
                assets = Custom("Static assets\n(built by Vite)", vite_icon)
                react >> Edge(label="build", style="dashed") >> assets

                with Cluster("Worker isolate"):
                    # why: Workers は isolate を提供する runtime、Hono はその isolate 内で動く
                    #   1つの Worker app。親子関係を Cluster 内包含で示す。
                    workers = Custom("Cloudflare Workers\n(V8 isolate)", workers_icon)
                    hono = Custom("Hono (Worker app)", hono_icon)
                    workers - Edge(style="invis") - hono

        # generator flow
        # why: 5 source はいずれも Go CLI が List で読む対等な Adapter。矢印は 1 本に集約し、
        #   内訳は Sources Cluster の node 名で示す。
        hn >> Edge(label="fetch (5 sources)") >> go_cli
        actions >> Edge(label="cron / manual") >> go_cli
        # why: 原稿 fallback は Gemini free → Cursor → Gemini paid の 3 段（Decision 2026-09-16T00-39-21）。
        go_cli >> Edge(label="draft") >> gemini_text
        gemini_text >> Edge(label="fallback on exhaustion", style="dashed") >> cursor_api
        cursor_api >> Edge(label="fallback on exhaustion", style="dashed") >> gemini_spare
        # why: TTS fallback は Gemini free → Gemini paid の 2 段。TextWriter 側と同じ
        #   SPARE_GEMINI_API_KEY を使うため fallback 先の node（gemini_spare）を共有する。
        go_cli >> Edge(label="TTS") >> gemini_tts
        gemini_tts >> Edge(label="fallback on exhaustion", style="dashed") >> gemini_spare
        go_cli >> Edge(label="encode") >> ffmpeg
        go_cli >> Edge(label="save (S3 API, over Internet)") >> r2

        # playback flow
        user >> Edge(label="HTTPS") >> access
        access >> Edge(label="static") >> assets
        access >> Edge(label="dynamic") >> workers
        ts >> Edge(label="type", style="dotted") >> react
        ts >> Edge(label="type", style="dotted") >> hono
        assets >> Edge(label="RPC / audio (HTTP)") >> hono
        # why: Worker binding は Cloudflare 内部の直結 API。fetch() の Internet 呼び出しではない
        #   ため、generator の save edge（実線・Internet 経由）と区別して太線で描く。
        hono >> Edge(label="R2 binding (no Internet hop)", penwidth="2") >> r2

    return Path(str(_OUTPUT) + ".png")


if __name__ == "__main__":
    print(render())
