package main

type ReqClient struct {
	Client Client `json:"client"`
}

type Client struct {
	Hl            string `json:"hl"`
	Gl            string `json:"gl"`
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
}

type ReqBody struct {
	Context ReqClient `json:"context"`
	Params  string    `json:"params"`
}

type ResponseContext struct {
	VisitorData                     string                          `json:"visitorData"`
	ServiceTrackingParams           ServiceTrackingParams           `json:"serviceTrackingParams"`
	MainAppWebResponseContext       MainAppWebResponseContext       `json:"mainAppWebResponseContext"`
	WebResponseContextExtensionData WebResponseContextExtensionData `json:"webResponseContextExtensionData"`
}

type ServiceTrackingParams []struct {
	Service string `json:"service"`
	Params  Params `json:"params"`
}

type Params []struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MainAppWebResponseContext struct {
	LoggedOut bool `json:"loggedOut"`
}

type WebResponseContextExtensionData struct {
	HasDecorated bool `json:"hasDecorated"`
}

type Action struct {
	ClickTrackingParams         string                      `json:"clickTrackingParams"`
	UpdateEngagementPanelAction UpdateEngagementPanelAction `json:"updateEngagementPanelAction"`
}

type Actions []Action

type UpdateEngagementPanelAction struct {
	TargetId string  `json:"targetId"`
	Content  Content `json:"content"`
}

type Content struct {
	TranscriptRenderer TranscriptRenderer `json:"transcriptRenderer"`
}

type Body struct {
	TranscriptBodyRenderer TranscriptBodyRenderer `json:"transcriptBodyRenderer"`
}

type TranscriptBodyRenderer struct {
	CueGroups CueGroups `json:"cueGroups"`
}

type CueGroups []struct {
	TranscriptCueGroupRenderer TranscriptCueGroupRenderer `json:"transcriptCueGroupRenderer"`
}

type TranscriptCueGroupRenderer struct {
	FormattedStartOffset FormattedStartOffset `json:"formattedStartOffset"`
	Cues                 Cues                 `json:"cues"`
}

type FormattedStartOffset struct {
	SimpleText string `json:"simpleText"`
}

type Cues []struct {
	TranscriptCueRenderer TranscriptCueRenderer `json:"transcriptCueRenderer"`
}

type TranscriptCueRenderer struct {
	Cue           Cue    `json:"cue"`
	StartOffsetMs string `json:"startOffsetMs"`
	DurationMs    string `json:"durationMs"`
}

type Cue struct {
	SimpleText string `json:"simpleText"`
}

type Footer struct {
	TranscriptFooterRenderer `json:"transcriptFooterRenderer"`
}

type TranscriptFooterRenderer struct {
	LanguageMenu LanguageMenu `json:"languageMenu"`
}

type LanguageMenu struct {
	SortFilterSubMenuRenderer SortFilterSubMenuRenderer `json:"sortFilterSubMenuRenderer"`
}

type SortFilterSubMenuRenderer struct {
	SubMenuItems   SubMenuItems `json:"subMenuItems"`
	TrackingParams string       `json:"trackingParams"`
}

type SubMenuItems []struct {
	Title        string       `json:"title"`
	Selected     bool         `json:"selected"`
	Continuation Continuation `json:"continuation"`
}

type Continuation struct {
	ReloadContinuationData ReloadContinuationData `json:"reloadContinuationData"`
}

type ReloadContinuationData struct {
	Continuation        string `json:"continuation"`
	ClickTrackingParams string `json:"clickTrackingParams"`
}

type TranscriptRenderer struct {
	Body           Body   `json:"body"`
	Footer         Footer `json:"footer"`
	TrackingParams string `json:"trackingParams"`
}

type ResTranscriptAPI struct {
	ResponseContext ResponseContext `json:"responseContext"`
	Actions         Actions         `json:"actions"`
	TrackingParams  string          `json:"trackingParams"`
}

// Youtube DATA API

type VideoListResponse struct {
	Kind     string      `json:"kind"`
	Etag     string      `json:"etag"`
	Items    []VideoItem `json:"items"`
	PageInfo PageInfo    `json:"pageInfo"`
}

type PageInfo struct {
	TotalResults   int `json:"totalResults"`
	ResultsPerPage int `json:"resultsPerPage"`
}

type VideoItem struct {
	Kind           string          `json:"kind"`
	Etag           string          `json:"etag"`
	Id             string          `json:"id"`
	ContentDetails VideoItemDetail `json:"contentDetails"`
}

type VideoItemDetail struct {
	Duration        string      `json:"duration"`
	Dimension       string      `json:"dimension"`
	Definition      string      `json:"definition"`
	Caption         string      `json:"caption"`
	LicensedContent bool        `json:"licensedContent"`
	ContentRating   interface{} `json:"contentRating"`
	Projection      string      `json:"projection"`
}