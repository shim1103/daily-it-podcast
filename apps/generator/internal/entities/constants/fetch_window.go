package constants

import "time"

// FetchWindow は PostSource 取得の遡及幅。
const FetchWindow = 24 * time.Hour
