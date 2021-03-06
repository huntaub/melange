package framework

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
)

// File View Will Render a File to the Http Response
type FileView struct {
	File io.ReadCloser
	Name string
	Size int
}

func (f *FileView) Write(w io.Writer) {
	_, err := io.Copy(w, f.File)
	if err != nil {
		fmt.Println("Got an error copying the file", err)
	}

	err = f.File.Close()
	if err != nil {
		fmt.Println("Got an error closing file", err)
	}
}

func (b *FileView) ContentLength() int {
	return b.Size
}

func (b *FileView) ContentType() string {
	return mime.TypeByExtension(filepath.Ext(b.Name))
}

func (b *FileView) Code() int {
	return 200
}

func (h *FileView) Headers() Headers { return nil }

var (
	errNoFile = errors.New("Couldn't find that file.")
)

// Blatantly stolen from robfig's Revel Framework
func ServeFile(prefix string, request string) View {
	view, err := GetFile(prefix, request)

	if err == errNoFile {
		return Error404
	} else if err != nil {
		fmt.Println("Unexpected Error Serving File", err)
		return Error500
	}

	return view
}
