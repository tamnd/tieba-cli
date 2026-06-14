package tieba

// Topic is the record emitted for the hot command.
type Topic struct {
	Rank        int    `json:"rank"`
	Name        string `json:"name"`
	Discussions int    `json:"discussions"`
	URL         string `json:"url"`
}

// ─── API wire types ──────────────────────────────────────────────────────────

type apiResponse struct {
	Data struct {
		BangTopic struct {
			TopicList []apiTopic `json:"topic_list"`
		} `json:"bang_topic"`
	} `json:"data"`
}

type apiTopic struct {
	TopicName  string `json:"topic_name"`
	DiscussNum int    `json:"discuss_num"`
	TopicURL   string `json:"topic_url"`
}
