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
@invariant catalog の URL はこの module が単一の定義源である。svg cache は
  `{name}.svg`、png cache は同じ stem の `{name}.png` に置く。
"""

from __future__ import annotations

from pathlib import Path
import subprocess
from typing import Callable
from urllib.request import Request, urlopen

Retrieve = Callable[[str, str], tuple[str, object]]
Convert = Callable[[Path, Path], None]

# why: diagrams 組み込みに公式 icon が無い製品だけをここに置く。
# why: CDN は空 User-Agent の urlretrieve を 403 にするので GitHub raw を正とする。
ICON_CATALOG: dict[str, str] = {
    "cloudflare-workers": (
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cloudflareworkers.svg"
    ),
    "cloudflare-access": (
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cloudflare.svg"
    ),
    "cursor": (
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/cursor.svg"
    ),
    "google-drive": (
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/googledrive.svg"
    ),
    "gemini": (
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/googlegemini.svg"
    ),
    "hono": (
        "https://raw.githubusercontent.com/simple-icons/simple-icons/develop/icons/hono.svg"
    ),
}

_DEFAULT_CACHE_DIR = Path(__file__).resolve().parent / "icons"
_USER_AGENT = "daily-it-podcast-diagrams/1.0"


def _retrieve(url: str, filename: str) -> tuple[str, None]:
    """URL から file を取得して `filename` へ書く。

    @require `url` は http(s) で取得できる。
    @ensure `filename` に本文が書き出され、第 1 戻り値はその path である。
    @invariant User-Agent 無しの取得は使わない。
    """
    # why: Simple Icons CDN は空 UA の urlretrieve を 403 にする。
    request = Request(url, headers={"User-Agent": _USER_AGENT})
    with urlopen(request, timeout=30) as response:
        Path(filename).write_bytes(response.read())
    return filename, None


def icon_path(
    name: str,
    cache_dir: Path | None = None,
    retrieve: Retrieve | None = None,
) -> Path:
    """catalog 名に対応する cache file の path を返す。

    @require `name` は `ICON_CATALOG` の key である。
    @ensure 戻り値の path にサイズ 0 より大きい file が存在する。cache miss の時だけ
      retrieve が 1 回呼ばれる。
    @invariant 同じ `cache_dir` と `name` なら常に同じ path を返す。
    """
    url = ICON_CATALOG.get(name)
    if url is None:
        raise ValueError(f"icon catalog に無い name: {name}")

    directory = cache_dir if cache_dir is not None else _DEFAULT_CACHE_DIR
    directory.mkdir(parents=True, exist_ok=True)
    path = directory / f"{name}.svg"
    if path.is_file() and path.stat().st_size > 0:
        return path

    fetch = retrieve if retrieve is not None else _retrieve
    fetch(url, str(path))
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
