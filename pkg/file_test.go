package pkg

import (
	"os"
	"path"
	"testing"
)

func Test_file_write(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Error(err, "should succeed")
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	filename := path.Join(dir, "file_write.txt")

	f := &file{
		filename: filename,
	}
	n, err := f.Write([]byte("hello\nman"))
	if err != nil {
		t.Error(err, "should succeed")
	}
	if n != 9 {
		t.Errorf("write should be %d", n)
	}
}

func Test_file_readlines(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Error(err, "should succeed")
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	filename := path.Join(dir, "file_write.txt")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Error(err, "should succeed")
		t.FailNow()
	}
	_, err = f.Write([]byte("line1\nline2\nline3\nline4"))
	if err != nil {
		t.Error(err, "should succeed")
		t.FailNow()
	}
	c := &file{
		filename: filename,
	}
	str, err := c.ReadLines(1, 1)
	if err != nil {
		t.Error(err, "should succeed")
		t.FailNow()
	}
	if len(str) != 1 || str[0] != "line2" {
		t.Error("should return line2, value:", str[0])
	}

}
