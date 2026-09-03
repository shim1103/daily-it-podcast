#!/usr/bin/env bash
# name: generator-install-cursor-cli-test
# description: install-cursor-cli.sh が fake 環境で `agent` を PATH 解決させることを確認する素 bash test。
# @require リポジトリ内から呼ぶ。bash 3.2+。
# @ensure INSTALLER_URL を no-op スクリプトへ向け、PATH 上に cursor-agent stub だけを置いた状態で
#         install-cursor-cli.sh を実行すると、その後 `agent` が PATH で解決する。
# @invariant 実ネットワークへ出ない（INSTALLER_URL を file スキームの no-op に差し替える）。
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
script="${root}/scripts/generator/install-cursor-cli.sh"

if [ ! -x "$script" ]; then
  echo "FAIL: ${script} が実行可能でない" >&2
  exit 1
fi

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

# fake HOME（install-cursor-cli.sh が ~/.local/bin を作る先）
fake_home="${workdir}/home"
mkdir -p "${fake_home}/.local/bin"

# PATH 上に cursor-agent stub だけを置く（agent は無い = symlink 作成経路を通す）
fake_path_dir="${workdir}/bin"
mkdir -p "$fake_path_dir"
cat > "${fake_path_dir}/cursor-agent" <<'STUB'
#!/usr/bin/env bash
echo "cursor-agent stub"
STUB
chmod +x "${fake_path_dir}/cursor-agent"

# INSTALLER_URL を no-op スクリプトへ（curl -o で取得される想定の中身）
noop_installer="${workdir}/noop-install.sh"
cat > "$noop_installer" <<'NOOP'
#!/usr/bin/env bash
echo "noop installer: 何もしない"
NOOP

# GITHUB_PATH 相当（echo 追記先）。無害な一時 file。
github_path_file="${workdir}/github_path"
: > "$github_path_file"

set +e
HOME="$fake_home" \
PATH="${fake_path_dir}:/usr/bin:/bin" \
GITHUB_PATH="$github_path_file" \
INSTALLER_URL="file://${noop_installer}" \
  bash "$script"
status=$?
set -e

if [ "$status" -ne 0 ]; then
  echo "FAIL: install-cursor-cli.sh が exit ${status}" >&2
  exit 1
fi

# 実行後、fake HOME/.local/bin/agent が cursor-agent を指しているか
if [ ! -e "${fake_home}/.local/bin/agent" ]; then
  echo "FAIL: ${fake_home}/.local/bin/agent が作られていない" >&2
  exit 1
fi

# その PATH で `agent` が解決するか
if ! HOME="$fake_home" PATH="${fake_home}/.local/bin:${fake_path_dir}:/usr/bin:/bin" command -v agent >/dev/null 2>&1; then
  echo "FAIL: 実行後の PATH で agent が解決しない" >&2
  exit 1
fi

echo "PASS: install-cursor-cli.sh が fake 環境で agent を PATH 解決させた"
