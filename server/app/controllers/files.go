package controllers

import (
	"errors"
	"os"

	"airdispat.ch/identity"

	"getmelange.com/app/models"
	gdb "github.com/huntaub/go-db"
)

type FileUploader struct {
	total     int64
	read      int64
	timesRead int

	notifier chan<- float64

	*os.File
}

func (r *FileUploader) Read(p []byte) (int, error) {
	n, err := r.File.Read(p)
	r.read += int64(n)

	// fmt.Println("Just read", n, "total read", r.read, "total total", r.total, "times read", r.timesRead)

	if r.timesRead > 1 && r.notifier != nil {
		// fmt.Println("Notifying", (float64(r.read) / float64(r.total)))
		select {
		case r.notifier <- (float64(r.read) / float64(r.total)):
		default:
		}
	}

	if r.read >= r.total {
		r.read = 0
		r.timesRead++

		if r.timesRead == 3 {
			close(r.notifier)
			r.notifier = nil
		}
	}

	return n, err
}

type UploadController struct {
	Store  *models.Store
	Tables map[string]gdb.Table
}

func (m *UploadController) HandleWSRequest(data map[string]interface{}, ws chan<- interface{}) error {
	fTest, ok := data["filename"]
	if !ok {
		return errors.New("Data must include filename.")
	}

	filename, ok := fTest.(string)
	if !ok {
		return errors.New("Filename must be string.")
	}

	toTest, ok := data["to"]
	if !ok {
		return errors.New("Data must include to.")
	}

	toAddrs, ok := toTest.([]interface{})
	if !ok {
		return errors.New("To must be an array of strings.")
	}

	to := make([]*identity.Address, len(toAddrs))
	for i, v := range toAddrs {
		to[i] = identity.CreateAddressFromString(v.(string))
	}

	tTest, ok := data["type"]
	if !ok {
		return errors.New("Data must include type.")
	}

	typ, ok := tTest.(string)
	if !ok {
		return errors.New("Type of data must be string.")
	}

	nTest, ok := data["name"]
	if !ok {
		return errors.New("Data must include type.")
	}

	name, ok := nTest.(string)
	if !ok {
		return errors.New("Name of data must be string.")
	}

	return m.UploadFile(filename, to, typ, name, ws)
}

func (m *UploadController) UploadFile(filename string, to []*identity.Address, typ string, name string, ws chan<- interface{}) error {
	// Current User Identity
	id, frameErr := CurrentIdentityOrError(m.Store, m.Tables["identity"])
	if frameErr != nil {
		return errors.New("Couldn't get current Identity.")
	}

	// DAP Client
	client, err := DAPClientFromID(id, m.Store)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	info, err := file.Stat()
	if err != nil {
		return err
	}

	n := make(chan float64)

	// Get upload status and send it to web socket
	go func() {
		for {
			data, ok := <-n

			// Exit on Close
			if !ok {
				ws <- map[string]interface{}{
					"type": "uploadedFile",
					"data": nil,
				}

				return
			}

			ws <- map[string]interface{}{
				"type": "uploadProgress",
				"data": data,
			}
		}
	}()

	// Construct the uploader
	uploader := &FileUploader{
		total:    info.Size(),
		notifier: n,
		File:     file,
	}

	// Upload the message
	return client.PublishDataMessage(uploader, to, typ, name)
}