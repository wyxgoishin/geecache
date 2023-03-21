package geecache

// A ByteView holds an immutable view of bytes
type ByteView struct {
	content []byte
}

func (b *ByteView) Len() int {
	return len(b.content)
}

// ByteSlice returns a copy of the data as a byte slice
func (b *ByteView) ByteSlice() []byte {
	return cloneBytes(b.content)
}

// String returns the data as a string, making a copy if necessary
func (b *ByteView) String() string {
	return string(b.content)
}

func cloneBytes(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
