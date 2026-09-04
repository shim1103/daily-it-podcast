"""
name: diagrams-icons
description: diagrams catalog 外の公式 icon を URL から cache し、再利用可能な path を返す。
when_to_use: architecture 図の Custom node へ icon path が必要な時に使う。
layer: 5
links:
  - "[[runtime]]"

@require 呼び出し側は `ICON_CATALOG` に存在する name を渡す。既定の retrieve は
  ネット接続を必要とする。test は retrieve を差し替えてネットへ出ない。
@ensure `icon_path` は cache 上の既存 svg の path を返す。無ければ retrieve で
  書き出してから同じ path を返す。未知名は ValueError を送出する。`rasterize`
  は svg に対応する png path を返す。
@invariant catalog の URL・公式ブランドカラーはこの module が単一の定義源である。svg cache は
  `{name}.svg`、png cache は同じ stem の `{name}.png` に置く。cache へ書き出す svg は
  `ICON_CATALOG` の color（Simple Icons の既定 fill 値）を root `<svg>` へ注入済みの状態。
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import re
import subprocess
from typing import Callable
from urllib.request import Request, urlopen

Retrieve = Callable[[str, str, str], tuple[str, object]]
Convert = Callable[[Path, Path], None]


@dataclass(frozen=True)
class IconSource:
    """catalog 1 件の取得元 URL と公式ブランドカラー（Simple Icons 既定色）。"""

    url: str
    color: str


# why: diagrams 組み込みに公式 icon が無い製品だけをここに置く。
# why: CDN は空 User-Agent の urlretrieve を 403 にするので GitHub raw を正とする。
# why: color は https://cdn.simpleicons.org/{slug} が返す既定 fill 値（Simple Icons の
#   ブランドカラー定義）をそのまま使う。GitHub raw の svg は fill 無し（currentColor 継承）
#   で raster すると黒一色になるため、取得後にこの色を注入する。
ICON_CATALOG: dict[str, IconSource] = {
    "cloudflare-workers": IconSource(
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cloudflareworkers.svg",
        "#F38020",
    ),
    "cloudflare-access": IconSource(
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cloudflare.svg",
        "#F38020",
    ),
    "cursor": IconSource(
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cursor.svg",
        "#000000",
    ),
    "google-drive": IconSource(
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/googledrive.svg",
        "#4285F4",
    ),
    "gemini": IconSource(
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/googlegemini.svg",
        "#8E75B2",
    ),
    "hono": IconSource(
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/hono.svg",
        "#E36002",
    ),
}

_DEFAULT_CACHE_DIR = Path(__file__).resolve().parent / "icons"
_USER_AGENT = "daily-it-podcast-diagrams/1.0"


def _inject_fill(svg_bytes: bytes, color: str) -> bytes:
    """svg の root `<svg ...>` タグへ `fill` 属性を挿入し、既存の `fill` は上書きする。

    @require `svg_bytes` は 1 個以上の `<svg` 開始タグを含む UTF-8 テキスト。
    @ensure 戻り値の root タグは `fill="{color}"` を持つ。他の属性・path 本体は変えない。
    """
    text = svg_bytes.decode("utf-8")
    head, sep, rest = text.partition("<svg")
    if not sep:
        return svg_bytes
    tag_end = rest.index(">")
    tag_attrs, tag_rest = rest[:tag_end], rest[tag_end:]

    if "fill=" in tag_attrs:
        tag_attrs = re.sub(r'fill="[^"]*"', f'fill="{color}"', tag_attrs, count=1)
    else:
        tag_attrs = f' fill="{color}"' + tag_attrs
    return (head + "<svg" + tag_attrs + tag_rest).encode("utf-8")


def _retrieve(url: str, filename: str, color: str) -> tuple[str, None]:
    """URL から svg を取得し、公式ブランドカラーを注入して `filename` へ書く。

    @require `url` は http(s) で取得できる。
    @ensure `filename` に `fill="{color}"` を持つ svg が書き出され、第 1 戻り値はその path である。
    @invariant User-Agent 無しの取得は使わない。
    """
    # why: Simple Icons CDN は空 UA の urlretrieve を 403 にする。
    request = Request(url, headers={"User-Agent": _USER_AGENT})
    with urlopen(request, timeout=30) as response:
        body = response.read()
    Path(filename).write_bytes(_inject_fill(body, color))
    return filename, None


def icon_path(
    name: str,
    cache_dir: Path | None = None,
    retrieve: Retrieve | None = None,
) -> Path:
    """catalog 名に対応する cache file の path を返す。

    @require `name` は `ICON_CATALOG` の key である。
    @ensure 戻り値の path にサイズ 0 より大きい file が存在する。cache miss の時だけ
      retrieve が 1 回呼ばれる。取得した svg は catalog の公式ブランドカラーで塗られる。
    @invariant 同じ `cache_dir` と `name` なら常に同じ path を返す。
    """
    source = ICON_CATALOG.get(name)
    if source is None:
        raise ValueError(f"icon catalog に無い name: {name}")

    directory = cache_dir if cache_dir is not None else _DEFAULT_CACHE_DIR
    directory.mkdir(parents=True, exist_ok=True)
    path = directory / f"{name}.svg"
    if path.is_file() and path.stat().st_size > 0:
        return path

    fetch = retrieve if retrieve is not None else _retrieve
    fetch(source.url, str(path), source.color)
    return path


def _rsvg_convert(svg_path: Path, png_path: Path) -> None:
    """svg を png へ変換する。

    @require `rsvg-convert` が PATH にある。`svg_path` が存在する。
    @ensure `png_path` に png が書き出される。
    @invariant 入力 svg は変更しない。
    """
    subprocess.run(
        [
            "rsvg-convert",
            "-w",
            "256",
            "-h",
            "256",
            "-o",
            str(png_path),
            str(svg_path),
        ],
        check=True,
    )


def rasterize(svg_path: Path, convert: Convert | None = None) -> Path:
    """svg を png にして、その path を返す。

    @require `svg_path` が存在する。既定 convert は `rsvg-convert` を必要とする。
    @ensure 戻り値は `svg_path` と同じ stem の `.png` で、サイズ 0 より大きい。
    @invariant png が既にあれば convert を呼ばない。
    """
    png_path = svg_path.with_suffix(".png")
    if png_path.is_file() and png_path.stat().st_size > 0:
        return png_path

    run = convert if convert is not None else _rsvg_convert
    run(svg_path, png_path)
    return png_path
