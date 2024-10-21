package fileset

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"google.golang.org/protobuf/proto"
)

type Pin int64

func (s *Storage) PinTx(ctx context.Context, tx *sqlx.Tx, handle *Handle) (Pin, error) {
	handle, err := s.CloneTx(tx, handle, track.ExpireNow)
	if err != nil {
		return 0, err
	}
	token := handle.token
	md, err := s.store.GetTx(tx, token)
	if err != nil {
		return 0, err
	}
	id, err := computeId(tx, s.store, md)
	if err != nil {
		return 0, err
	}
	rootPb, err := proto.Marshal(md)
	if err != nil {
		return 0, errors.EnsureStack(err)
	}
	var pin Pin
	if err := tx.GetContext(ctx, &pin, `
		INSERT INTO storage.fileset_pins(fileset_id, root_pb)
		VALUES($1, $2)
		RETURNING id
	`, id[:], rootPb); err != nil {
		return 0, errors.Wrapf(err, "insert pin")
	}
	if err := s.tracker.CreateTx(tx, pinTrackerID(pin), []string{token.TrackerID()}, track.NoTTL); err != nil {
		return 0, errors.Wrap(err, "create tracker object and references")
	}
	return pin, nil
}

const pinTrackerPrefix = "pin/"

func pinTrackerID(pin Pin) string {
	return pinTrackerPrefix + strconv.FormatInt(int64(pin), 10)
}

func (s *Storage) DeletePinTx(ctx context.Context, tx *pachsql.Tx, pin Pin) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM storage.fileset_pins WHERE id = $1", pin); err != nil {
		return errors.EnsureStack(err)
	}
	return s.tracker.DeleteTx(tx, pinTrackerID(pin))
}

func (s *Storage) GetPinHandleTx(ctx context.Context, tx *pachsql.Tx, pin Pin, ttl time.Duration) (*Handle, error) {
	var rootPb []byte
	if err := tx.GetContext(ctx, &rootPb, `
		SELECT root_pb
		FROM storage.fileset_pins
		WHERE id = $1
	`, pin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pacherr.NewNotExist("pin", strconv.FormatInt(int64(pin), 10))
		}
		return nil, errors.EnsureStack(err)
	}
	md := &Metadata{}
	if err := proto.Unmarshal(rootPb, md); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return s.newHandle(tx, md, ttl)
}

func CreatePinsTable(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE storage.fileset_pins (
			id BIGSERIAL NOT NULL PRIMARY KEY,
			fileset_id BYTEA NOT NULL,
			root_pb BYTEA NOT NULL
		)
	`); err != nil {
		return errors.Wrap(err, "create storage.fileset_pins table")
	}
	return nil
}
