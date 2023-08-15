package naming

import "fmt"

const (
	EVENT_REAL_TO_VIRTUAL string = "ktwin.real.%s"
	EVENT_VIRTUAL_TO_REAL string = "ktwin.virtual.%s"
	EVENT_TO_EVENT_STORE  string = "ktwin.event.store"
)

func GetEventTypeVirtualGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(EVENT_VIRTUAL_TO_REAL, twinInterfaceName)
}

func GetEventTypeRealGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(EVENT_REAL_TO_VIRTUAL, twinInterfaceName)
}

func GetNewCloudEventEventBinding(twinInterfaceName string) string {
	return fmt.Sprintf(EVENT_REAL_TO_VIRTUAL, twinInterfaceName)
}

func GetNewMQQTEventBinding(twinInterfaceName string) string {
	return fmt.Sprintf(EVENT_REAL_TO_VIRTUAL, twinInterfaceName)
}
