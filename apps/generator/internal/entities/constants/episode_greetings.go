package constants

// OpeningGreeting / ClosingFarewell は ProduceEpisode が TTS 前に組み立てる定型。
// ClosingFarewell 最終文言は確定済み。Decision: docs/decisions/2026-08-29T14-13-00-docs-produce-episode-run-spec-date-greeting-injection.md

const (
	// OpeningGreetingTemplate の %s には JST 暦日を読み上げ用に整形した文字列を Builder が渡す。
	OpeningGreetingTemplate = "おはようございます。%sのITニュースをお伝えします。"
	ClosingFarewell         = "以上、%sのITニュースでした。"
)
