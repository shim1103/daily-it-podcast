package config

// r2ConfigField は R2Config の field 順（AccessKeyID → SecretAccessKey → AccountID → Bucket）。
type r2ConfigField int

const (
	fieldR2AccessKeyID r2ConfigField = iota
	fieldR2SecretAccessKey
	fieldR2AccountID
	fieldR2Bucket
	r2ConfigFieldCount
)

var r2ConfigFieldKeys = [r2ConfigFieldCount]string{
	fieldR2AccessKeyID:     R2AccessKeyIDEnv,
	fieldR2SecretAccessKey: R2SecretAccessKeyEnv,
	fieldR2AccountID:       R2AccountIDEnv,
	fieldR2Bucket:          R2BucketEnv,
}

// LoadR2 は R2_* process environment だけを読み、検証済み R2Config を構築する。
// 現行本番 Load とは独立。cutover 前に Composition 結線口が使えるようにする。
//
// @require lookupは注入されたenvironment参照だけを入力源にし、dotenv fileをloadしない。
// @ensure 全 R2 field が有効な時だけ R2Config を返す。
// @ensure 1 fieldでも違反があれば field 順で全違反を束ねた *Errors を返し、R2Config は zero value。
// @invariant raw runtime値をerrorへ含めない。Load（Drive 正本）の必須集合を変えない。
func LoadR2(lookup LookupEnv) (R2Config, error) {
	var values [r2ConfigFieldCount]string
	var violations []*Error
	for field, key := range r2ConfigFieldKeys {
		value, kind := validateEnvValue(lookup, key)
		if kind != "" {
			violations = append(violations, configErr(key, kind))
			continue
		}
		values[field] = value
	}

	if len(violations) > 0 {
		return R2Config{}, &Errors{Violations: violations}
	}

	return R2Config{
		AccessKeyID:     newSecret(values[fieldR2AccessKeyID]),
		SecretAccessKey: newSecret(values[fieldR2SecretAccessKey]),
		AccountID:       values[fieldR2AccountID],
		Bucket:          values[fieldR2Bucket],
	}, nil
}
