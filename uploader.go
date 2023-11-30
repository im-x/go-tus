package tus

import (
	"bytes"
	"errors"
)

type Uploader struct {
	client   *Client
	url      string
	upload   *Upload
	offset   int64
	aborted  bool
	fileInfo string
}

// Abort aborts the upload process.
// It doens't abort the current chunck, only the remaining.
func (u *Uploader) Abort() {
	u.aborted = true
}

// IsAborted returns true if the upload was aborted.
func (u *Uploader) IsAborted() bool {
	return u.aborted
}

// Url returns the upload url.
func (u *Uploader) Url() string {
	return u.url
}

// Offset returns the current offset uploaded.
func (u *Uploader) Offset() int64 {
	return u.offset
}

// Upload uploads the entire body to the server.
func (u *Uploader) Upload() (string, error) {
	for u.offset < u.upload.size && !u.aborted {
		err := u.UploadChunck()
		// 有fileinfo,直接上传完成
		if u.fileInfo != "" {
			return u.fileInfo, nil
		}

		if err != nil {
			return "", err
		}
	}

	return "", errors.New("server not return file info")
}

// UploadChunck uploads a single chunck.
func (u *Uploader) UploadChunck() error {
	data := make([]byte, u.client.Config.ChunkSize)

	_, err := u.upload.stream.Seek(u.offset, 0)

	if err != nil {
		return err
	}

	size, err := u.upload.stream.Read(data)

	if err != nil {
		return err
	}

	body := bytes.NewBuffer(data[:size])

	newOffset, fileInfo, err := u.client.uploadChunck(u.url, body, int64(size), u.offset)
	u.fileInfo = fileInfo
	// 有fileinfo则表示上传成功
	if len(u.fileInfo) != 0 {
		return nil
	}
	if err != nil {
		return err
	}

	u.offset = newOffset

	return nil
}

// NewUploader creates a new Uploader.
func NewUploader(client *Client, url string, upload *Upload, offset int64) *Uploader {
	uploader := &Uploader{
		client,
		url,
		upload,
		offset,
		false,
		"",
	}
	return uploader
}
