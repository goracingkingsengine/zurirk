// generated by stringer -type HashKind; DO NOT EDIT

package engine

import "fmt"

const _hashKind_name = "noEntryexactfailedLowfailedHigh"

var _hashKind_index = [...]uint8{7, 12, 21, 31}

func (i hashKind) String() string {
	if i >= hashKind(len(_hashKind_index)) {
		return fmt.Sprintf("hashKind(%d)", i)
	}
	hi := _hashKind_index[i]
	lo := uint8(0)
	if i > 0 {
		lo = _hashKind_index[i-1]
	}
	return _hashKind_name[lo:hi]
}
