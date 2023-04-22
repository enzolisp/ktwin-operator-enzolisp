package event

const (
	EVENT_REAL_TO_VIRTUAL        string = "ktwin.real.virtual.generated"
	EVENT_VIRTUAL_TO_REAL        string = "ktwin.virtual.real.generated"
	EVENT_REAL_TO_EVENT_STORE    string = "ktwin.real.store.generated"
	EVENT_VIRTUAL_TO_EVENT_STORE string = "ktwin.real.store.generated"
	EVENT_VIRTUAL_TO_VIRTUAL     string = "ktwin.virtual.virtual.generated"
)

type TwinEvent interface {
	CreateEvent()
}

type twinEvent struct{}

func (*twinEvent) CreateEvent() {}
