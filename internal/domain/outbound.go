package domain

// GroupItem represents an outbound item within a selector or group.
type GroupItem struct {
	Tag          string `json:"tag"`
	Type         string `json:"type"`
	URLTestTime  int64  `json:"urlTestTime"`
	URLTestDelay int32  `json:"urlTestDelay"`
}

// OutboundGroup represents a group or selector of outbounds.
type OutboundGroup struct {
	Tag        string      `json:"tag"`
	Type       string      `json:"type"`
	Selectable bool        `json:"selectable"`
	Selected   string      `json:"selected"`
	IsExpand   bool        `json:"isExpand"`
	Items      []GroupItem `json:"items"`
}

// OutboundInfo represents an individual outbound definition.
type OutboundInfo struct {
	Tag          string `json:"tag"`
	Type         string `json:"type"`
	URLTestTime  int64  `json:"urlTestTime"`
	URLTestDelay int32  `json:"urlTestDelay"`
}
