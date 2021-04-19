package alog

var (
	ConfigAppVersion      = ""
	ConfigSentryUrl       = ""
	ConfigSentryPublicKey = ""
)

func InitAlog(
	version string,
	sentryUrl, sentryPublicKey string,
) {
	ConfigAppVersion = version
	ConfigSentryUrl = sentryUrl
	ConfigSentryPublicKey = sentryPublicKey
}
