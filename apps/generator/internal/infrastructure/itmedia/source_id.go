package itmedia

// SourceID は ITmedia NEWS を表す情報源識別子。
//
// why: これは ITmedia NEWS の adapter。1 outlet 1 adapter で、hackernews/ lobsters/ と対称にする。
// 他の outlet（Publickey、InfoQ …）は共有 feed list ではなく、それぞれ infrastructure/<outlet>/ の
// 固有 adapter を持つ。generic な "rss" SourceID は作らない。
const SourceID = "itmedia"
