package metric

// metric tag name
const (
	TagChannel = "channel"
	TagType    = "type"
)

// metric tag value
const (
	ChannelWeb    = "WebConsole"
	ChannelEms    = "Email"
	ChannelWeWork = "WeWork"

	TypeAccessToken     = "assessToken"
	TypeWeWorkAPPDetail = "appDetail"
)

// metric tag map
var (
	WebTags             = map[string]string{TagChannel: ChannelWeb}
	EmsTags             = map[string]string{TagChannel: ChannelEms}
	WeWorkTags          = map[string]string{TagChannel: ChannelWeWork}
	AccessTokenTags     = map[string]string{TagChannel: ChannelWeWork, TagType: TypeAccessToken}
	WeWorkAppDetailTags = map[string]string{TagChannel: ChannelWeWork, TagType: TypeWeWorkAPPDetail}
)
