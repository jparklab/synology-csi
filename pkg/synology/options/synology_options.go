package options

// SynologyOptions contains options to access Synology NAS web api
type SynologyOptions struct {
	Host            string `yaml:"host"  url:"-"`
	Port            int    `yaml:"port"  url:"-"`
	SslVerify       bool   `yaml:"sslVerify" url:",omitempty"`
	LoginApiVersion int    `yaml:"loginApiVersion" url:"version"`
	LoginHttpMethod string `yaml:"loginHttpMethod" url:"-"`
	API             string `yaml:"-" url:"api"`
	Method          string `yaml:"-" url:"method"`

	// === Version 1 and later, DSM 3.2 ===
	// Required.
	// Login account name
	Username string `yaml:"username" url:"account"`
	// Required.
	// Login account password
	Password string `yaml:"password" url:"passwd"`
	// Optional.
	// Application session name.
	// User can assign “SurveillanceStation” to this parameter to login SurveilllanceStation.
	// If not specified, default session is DSM, and SurveillanceStation is also available.
	SessionName string `yaml:"sessionName" url:"session"`

	// === Version 2 and later, DSM  4.1 ===
	// Optional.
	// If format is “cookie”, session ID is included in both response header and response
	// json data. If format is “sid”, se ssion ID is not included in response header, but
	// response json data only. User can append this session ID manually to get access to
	// any other Web API without interrupting other logins.
	// If not specified, default login format is “cookie.”
	Format string `yaml:"-" url:"format"`

	// === Version 3, DSM 4.2 ===
	// Optional.
	// Return synotoken if value is "yes".
	EnableSynoToken *string `yaml:"enableSynoToken" url:"enable_syno_token,omitempty"`
	// Optional.
	// 6-digit OTP code.
	OptCode *string `yaml:"-" url:"otp_code,omitempty"`

	// === Version 4, DSM 5.2 ===

	// === Version 5, DSM 6.0, beta 1 ===

	// === Version 6, DSM 6.0 beta 2 ===
	// Optional.
	// yes or no, default to no.
	EnableDeviceToken *string `yaml:"enableDeviceToken" url:"enable_device_token,omitempty"`
	// Optional.
	// Device id (max: 255).
	DeviceId *string `yaml:"deviceId" url:"device_id,omitempty"`
	// Optional.
	// Device name (max: 255).
	DeviceName *string `yaml:"deviceName" url:"device_name,omitempty"`
}

func NewSynologyOptions() SynologyOptions {
	return SynologyOptions{
		API:             "SYNO.API.Auth",
		Method:          "login",
		LoginHttpMethod: "auto",
		Format:          "sid",
	}
}
