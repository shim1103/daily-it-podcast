package constants

import "time"

// FetchWindow は ItemSource.List へ渡す since の遡及幅。
// OccurredAt（情報の発生時刻）そのものではない。
const FetchWindow = 24 * time.Hour
