#!/usr/bin/env bash
# name: install-hooks
# description: 管理下の git hook を .git/hooks へ symlink する。
# @require リポジトリ内で実行する。git が利用可能。
# @ensure pre-commit と pre-push が scripts/git-hooks を指す symlink になる。
# @invariant 他の hook を消さない。force で上書きする対象は pre-commit / pre-push のみ。
# why not $root/.git/hooks: worktree では .git が directory ではなく file のため。
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
hooks_dir="$(git -C "$root" rev-parse --git-path hooks)"
src_dir="$root/scripts/git-hooks"

mkdir -p "$hooks_dir"

for name in pre-commit pre-push; do
  ln -sfn "$src_dir/$name" "$hooks_dir/$name"
  chmod +x "$src_dir/$name"
  echo "linked  $name -> scripts/git-hooks/$name"
done
