package metadata

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
)

// EditMetadata transactionally mutates metadata.  All operations are attempted, in order, but if
// any fail, the entire operation fails.
func EditMetadata(ctx context.Context, db *pachsql.DB, req *metadata.EditMetadataRequest) error {
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		var errs error
		for i, edit := range req.GetEdits() {
			if err := editInTx(ctx, tx, edit); err != nil {
				errors.JoinInto(&errs, errors.Wrapf(err, "edit #%d", i))
			}
		}
		if errs != nil {
			return errs
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "apply edits")
	}
	return nil
}

func editMetadata(edit *metadata.Edit, md *map[string]string) error {
	if *md == nil {
		*md = make(map[string]string)
	}
	switch x := edit.GetOp().(type) {
	case *metadata.Edit_AddKey_:
		k, v := x.AddKey.Key, x.AddKey.Value
		if _, ok := (*md)[k]; ok {
			return errors.Errorf("add_key target key %q already exists; use edit_key instead", k)
		}
		(*md)[k] = v
	case *metadata.Edit_EditKey_:
		k, v := x.EditKey.Key, x.EditKey.Value
		(*md)[k] = v
	case *metadata.Edit_DeleteKey_:
		k := x.DeleteKey.Key
		delete(*md, k)
	case *metadata.Edit_Replace_:
		*md = x.Replace.Replacement
	}
	return nil
}

func editInTx(ctx context.Context, tx *pachsql.Tx, edit *metadata.Edit) error {
	switch x := edit.GetTarget().(type) {
	case *metadata.Edit_Project:
		p, err := pfsdb.PickProject(ctx, x.Project, tx)
		if err != nil {
			return errors.Wrap(err, "pick project")
		}
		if err := editMetadata(edit, &p.Metadata); err != nil {
			return errors.Wrapf(err, "edit project %q", p.GetProject().GetName())
		}
		if err := pfsdb.UpdateProject(ctx, tx, p.ID, p.ProjectInfo); err != nil {
			return errors.Wrapf(err, "update project %q", p.GetProject().GetName())
		}
	default:
		return errors.Errorf("unknown target %v", edit.GetTarget())
	}
	return nil
}
