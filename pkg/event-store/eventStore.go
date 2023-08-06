package eventStore

import (
	dtdv0 "ktwin/operator/api/dtd/v0"
)

const (
	EVENT_STORE_SERVICE string = "event-store"
)

func NewTwinEventStore() TwinEventStore {
	return &twinEventStore{}
}

type TwinEventStore interface {
	CreateTwinInterface(twinInterface *dtdv0.TwinInstance) error
	DeleteTwinInterface(twinInterface *dtdv0.TwinInstance) error
	CreateTwinInstance(twinInstance *dtdv0.TwinInstance) error
	DeleteTwinInstance(twinInstance *dtdv0.TwinInstance) error
}

type twinEventStore struct{}

func (t *twinEventStore) CreateTwinInterface(twinInterface *dtdv0.TwinInstance) error {
	// TwinInstance

	// Interface
	return nil
}

func (t *twinEventStore) DeleteTwinInterface(twinInterface *dtdv0.TwinInstance) error {
	// TwinInstance

	return nil
}

func (t *twinEventStore) CreateTwinInstance(twinInstance *dtdv0.TwinInstance) error {
	//

	return nil
}

func (t *twinEventStore) DeleteTwinInstance(twinInstance *dtdv0.TwinInstance) error {
	//

	return nil
}
