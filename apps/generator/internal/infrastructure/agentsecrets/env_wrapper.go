package agentsecrets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// what: CLI 境界の秘密供給 wrapper の確定値。`agentsecrets env -- <command> [args...]` が
// keychain から秘密を解決して子 process の env へ注入する（decision 2026-08-22T11-55-22 §1）。
const (
	// why: Go process が秘密値を保持しないため、exec する program は呼びたい CLI 自身ではなく wrapper になる。
	EnvBinary     = "agentsecrets"
	EnvSubcommand = "env"

	// what: wrapper 自身の flag と、wrapper が起動する command の argv を分ける区切り。
	ArgSeparator = "--"
)

// EnvWrapper は CLI 境界の秘密供給を 1 回分だけ記述する。
//
// why: `agentsecrets env` は解決する秘密を絞る flag を持たず、常に active project の
// 全 secret を子 process へ注入する。active project は git root ではなく実行時の
// current directory 直下の設定 file で決まる（`agentsecrets` skill `project-binding.md` §1）。
// したがって渡る秘密の範囲を絞る唯一の手段は、子 process の working directory を
// 目的別 project の設定 dir へ向けることである。ProjectDir がその指定を担う。
//
// why not: `agentsecrets exec` は秘密値を stdout から呼び出し元へ返すため採用しない。
// Go process が値を見た時点で zero-knowledge が崩れる（decision 2026-08-22T11-55-22 §3-1）。
type EnvWrapper struct {
	// ProjectDir は wrapper を起動する working directory。この dir 直下の設定 file が
	// 指す project の secret だけが子 process へ渡る。
	ProjectDir string

	// SecretKeys はこの呼び出しが依存する秘密キー名の宣言。
	//
	// warn: ⚠️ `agentsecrets env` は key 名を受け取らないため、この値は実行時に wrapper へ
	// 渡らない。実際に注入されるのは ProjectDir が指す project の全 secret であり、
	// 宣言と実注入は乖離しうる。ここは HTTP 境界の Inject と同じく「どの秘密に依存するか」を
	// code へ明示する契約 documentation として機能し、ProjectDir の分離が正しいかを
	// 人が検証するときの基準になる。将来 wrapper が key 指定機構を持てばここが受け口になる。
	SecretKeys []string
}

// Command は子 command の argv を wrapper の argv で包み、exec する program 名と argv を返す。
//
// @require childArgv の要素に秘密値を含めない。
// @ensure 戻りの program 名は EnvBinary。argv は EnvSubcommand・ArgSeparator の順で始まり、childArgv がそのまま続く。
// @ensure SecretKeys は argv へ載せない（wrapper に key 指定 flag が無いため）。
func (w EnvWrapper) Command(childArgv []string) (string, []string) {
	args := make([]string, 0, 2+len(childArgv))
	args = append(args, EnvSubcommand, ArgSeparator)
	args = append(args, childArgv...)
	return EnvBinary, args
}

// Validate は wrapper が秘密の範囲を絞れる状態かを確認する。
//
// @ensure ProjectDir が空白のみ、または絶対 path でないとき error を返す。それ以外は nil を返す。
// @ensure dir の実在は検査しない。存在しない dir は wrapper の起動失敗として実行時に現れる。
func (w EnvWrapper) Validate() error {
	// why: ProjectDir が空だと exec は親 process の cwd で wrapper を起動する。その cwd が
	// 指す project の全 secret が子 process へ渡るため、絞り込みが黙って外れる。
	// 空を「絞らない」の意味で通さず、設定漏れとして落とす。
	if strings.TrimSpace(w.ProjectDir) == "" {
		return fmt.Errorf("agentsecrets: EnvWrapper.ProjectDir is empty")
	}
	// why: 相対 path は親 process の cwd を基準に解決される。起動位置が変われば active project
	// も変わり、渡る秘密の範囲が code から読めなくなる。絶対 path を要求して解決先を固定する。
	if !filepath.IsAbs(w.ProjectDir) {
		return fmt.Errorf("agentsecrets: EnvWrapper.ProjectDir must be an absolute path")
	}
	// why not: dir の実在と設定 file の有無をここで検査しない。既存 adapter は proxy 未起動を
	// 接続 error として実行時に畳んでおり、事前検査の前例が無い。一貫性（philosophy §5-1）に
	// 従い、環境側の不備は wrapper の起動失敗として同じ経路へ落とす。
	return nil
}

// ProjectsRootName は目的別 AgentSecrets project の設定 dir を束ねる root の名前。
//
// what: この root の下に project ごとの dir を置き、各 dir 直下へ `agentsecrets use-project`
// が `.agentsecrets/project.json` を生成する。
//
// why not: repo 内へ置かない。repo root の設定 file は git 追跡され、全 vendor の secret を
// 持つ単一 project を指す。その隣へ目的別 project を足すと、秘密境界の構成が repo の
// clone 状態と結びつき、clone していない環境で絞り込みが黙って外れる。
const ProjectsRootName = ".agentsecrets-projects"

// ProjectDir は home 配下の目的別 project 設定 dir の path を組み立てる。
//
// why: dir の配置規約は AgentSecrets の知識であり、これを呼び出し側が組み立てると
// project 追加のたびに同じ規約が写経される。ここが唯一の組み立て地点になる。
//
// @require name は project 名。home は絶対 path。
// @ensure home が絶対 path のとき、戻りも絶対 path。
// @ensure home が空のとき相対 path を返す。呼び出し側は Validate で実行時 error へ落とせる。
func ProjectDir(home string, name string) string {
	return filepath.Join(home, ProjectsRootName, name)
}

// DefaultProjectDir は目的別 project 設定 dir の既定 path を実行環境から解決する。
//
// why: 接続先の位置情報は隠すべき具体値であり、呼び出し側が home を読んで組み立てると
// 解決規則が呼び出し元の数だけ散る。Client が DefaultProxyURL を自分で解決するのと同じく、
// 位置の解決を agentsecrets 側へ寄せる。
//
// @ensure HOME が絶対 path のとき、戻りも絶対 path。
// @ensure HOME が未設定のとき Validate が弾く相対 path を返す。環境不備は実行時に現れる。
func DefaultProjectDir(name string) string {
	return ProjectDir(os.Getenv("HOME"), name)
}
