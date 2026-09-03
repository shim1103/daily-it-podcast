#!/usr/bin/env bash
# name: generator-install-cursor-cli
# description: Cursor CLI（agent）を install し、`agent` を PATH（~/.local/bin）で解決可能にする。
# @require リポジトリ内から呼ぶ。curl が使える。$HOME が書ける。
# @ensure install 後、`agent` が PATH（~/.local/bin）で解決する。GHA では ~/.local/bin を GITHUB_PATH へ追記する。
# @ensure INSTALLER_URL env で installer 取得元を上書きできる（既定 https://cursor.com/install。test stub 用）。
# @invariant secret 値を扱わない・log へ出さない。argv に秘密を載せない。
set -euo pipefail

INSTALLER_URL="${INSTALLER_URL:-https://cursor.com/install}"

# install_cursor_cli は installer を取得・実行し、`agent` の PATH 解決を保証する。
install_cursor_cli() {
  local local_bin="${HOME}/.local/bin"
  mkdir -p "${local_bin}"

  local installer
  installer="$(mktemp)"
  curl -fsS "${INSTALLER_URL}" -o "${installer}"
  bash "${installer}"
  rm -f "${installer}"

  export PATH="${local_bin}:${PATH}"

  # GHA では後続 step へ PATH を引き継ぐ。
  if [ -n "${GITHUB_PATH:-}" ]; then
    echo "${local_bin}" >> "${GITHUB_PATH}"
  fi

  # installer が `cursor-agent` だけを置く版へ備え、`agent` への symlink を張る。
  if ! command -v agent >/dev/null 2>&1; then
    if command -v cursor-agent >/dev/null 2>&1; then
      ln -sfn "$(command -v cursor-agent)" "${local_bin}/agent"
    else
      echo "Cursor CLI binary 未解決（agent / cursor-agent とも PATH に無い）" >&2
      return 1
    fi
  fi

  command -v agent
  command -v cursor-agent || true
}

install_cursor_cli
