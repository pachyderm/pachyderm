package log

import (
	uuidlib "github.com/satori/go.uuid"
)

type requestKind byte

const (
	interactiveTrace       requestKind = 0x1
	backgroundTrace        requestKind = 0x2
	consoleTrace           requestKind = 0x3
	envoyTrace             requestKind = 0x4
	syntheticTrace         requestKind = 0x5
	envoyServerForcedTrace requestKind = 0xa
	envoyClientForcedTrace requestKind = 0xb
)

func uuid(kind requestKind) uuidlib.UUID {
	id, err := uuidlib.NewV4()
	if err != nil {
		// This is exceedingly unlikely to return an error.  internal/uuid cannot be
		// used because it introduces an import cycle between log -> uuid -> backoff
		// -> log.
		id = uuidlib.UUID{}
	}
	id[6] = (id[6] & 0x0f) | (byte(kind) << 4)
	return id
}
