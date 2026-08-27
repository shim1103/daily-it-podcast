package config

// LookupEnv はprocess environmentのkeyを検索する契約である。
type LookupEnv func(string) (string, bool)

// Loader はprocess environmentを一度だけ読み、Configを検証する契約である。
//
// @require lookupはprocess environmentだけを参照し、dotenv fileをloadしない。
// @ensure invalidな全fieldをConfigのfield順で集約したErrorを返す。
// @invariant raw valueをerrorへ含めない。
type Loader func(LookupEnv) (Config, error)
