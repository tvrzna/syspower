package syspower

const version = "0.1.0"

var buildVersion string

// Gets current syspower version
func GetVersion() string {
	if buildVersion == "" {
		return version
	}
	return buildVersion
}
