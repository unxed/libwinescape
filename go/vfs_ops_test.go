package winescape

import (
	"bytes"
	"path"
	"testing"
)

func TestWriteFile_ReadFile(t *testing.T) {
	tmpFile := "/tmp/winescape_vfs_test.tmp"
	defer Unlink(tmpFile)

	data := []byte("vfs operations test payload 12345")
	if err := WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	readBack, err := ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if !bytes.Equal(readBack, data) {
		t.Errorf("read mismatch: got %q, want %q", string(readBack), string(data))
	}
}

func TestMkdirAll_RemoveAll(t *testing.T) {
	tree := "/tmp/winescape_tree_test/sub1/sub2/sub3"
	defer RemoveAll("/tmp/winescape_tree_test")

	if err := MkdirAll(tree, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	var st Stat_t
	if err := Stat(tree, &st); err != nil || !st.IsDir() {
		t.Fatalf("Stat directory tree failed: %v", err)
	}

	leafFile := path.Join(tree, "leaf.txt")
	if err := WriteFile(leafFile, []byte("leaf"), 0644); err != nil {
		t.Fatalf("WriteFile leaf failed: %v", err)
	}

	if err := RemoveAll("/tmp/winescape_tree_test"); err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	if err := Stat("/tmp/winescape_tree_test", &st); err == nil {
		t.Errorf("expected tree to be removed")
	}
}
