package json

type Events []Event
type Event struct {
	//ID   string `json:"id"`
	Name string `json:"name"`
}
type EditEventsRequest struct {
	Add []string `json:"add"`
}
