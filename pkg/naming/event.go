package naming

import (
	"fmt"

	"github.com/Open-Digital-Twin/ktwin-operator/pkg/event"
)

func GetEventTypeVirtualGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(event.EVENT_TYPE_VIRTUAL_GENERATED, twinInterfaceName)
}

func GetEventTypeRealGenerated(twinInterfaceName string) string {
	return fmt.Sprintf(event.EVENT_TYPE_REAL_GENERATED, twinInterfaceName)
}

func GetNewCloudEventEventBinding(twinInterfaceName string) string {
	return fmt.Sprintf(event.EVENT_TYPE_REAL_GENERATED, twinInterfaceName)
}

func GetNewMQQTEventBinding(twinInterfaceName string) string {
	return fmt.Sprintf(event.EVENT_TYPE_REAL_GENERATED, twinInterfaceName)
}
