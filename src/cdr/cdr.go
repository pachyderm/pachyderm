// Package cdr provides common data references.
//
// TODO: document what common data references are.
package cdr

func IsImmutable(ref *Ref) bool {
	switch body := ref.Body.(type) {
	case *Ref_ContentHash:
		return true
	case nil:
		return true
	case *Ref_Concat:
		var ret bool
		for _, ref := range body.Concat.Refs {
			ret = ret && IsImmutable(ref)
		}
		return ret
	case *Ref_Cipher:
		return IsImmutable(body.Cipher.Inner)
	case *Ref_Compress:
		return IsImmutable(body.Compress.Inner)
	case *Ref_SizeLimits:
		return IsImmutable(body.SizeLimits.Inner)
	case *Ref_Slice:
		return IsImmutable(body.Slice.Inner)
	default:
		return false
	}
}
