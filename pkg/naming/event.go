package naming

import (
	"fmt"

	"github.com/Open-Digital-Twin/ktwin-operator/pkg/event"
)

func GetEventTypeVirtualGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(event.EVENT_VIRTUAL_TO_REAL, twinInterfaceName)
}

func GetEventTypeRealGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(event.EVENT_REAL_TO_VIRTUAL, twinInterfaceName)
}

func GetNewCloudEventEventBinding(twinInterfaceName string) string {
	return fmt.Sprintf(event.EVENT_REAL_TO_VIRTUAL, twinInterfaceName)
}

func GetNewMQQTEventBinding(twinInterfaceName string) string {
	return fmt.Sprintf(event.EVENT_REAL_TO_VIRTUAL, twinInterfaceName)
}
