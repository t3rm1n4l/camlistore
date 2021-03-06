/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package google

import (
	"errors"
	"io"
	"time"

	"camlistore.org/pkg/blobref"
	"camlistore.org/pkg/blobserver"
	"camlistore.org/pkg/googlestorage"
	"camlistore.org/pkg/jsonconfig"
)

type Storage struct {
	hub    *blobserver.SimpleBlobHub
	bucket string // the gs bucket containing blobs
	client *googlestorage.Client
}

func newFromConfig(_ blobserver.Loader, config jsonconfig.Obj) (blobserver.Storage, error) {
	auth := config.RequiredObject("auth")
	gs := &Storage{
		&blobserver.SimpleBlobHub{},
		config.RequiredString("bucket"),
		googlestorage.NewClient(MakeOauthTransport(
			auth.RequiredString("client_id"),
			auth.RequiredString("client_secret"),
			auth.RequiredString("refresh_token"),
		)),
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if err := auth.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}

func (gs *Storage) EnumerateBlobs(dest chan<- blobref.SizedBlobRef, after string, limit int, wait time.Duration) error {
	// TODO: Implement stub
	return nil
}

func (gs *Storage) ReceiveBlob(blob *blobref.BlobRef, source io.Reader) (blobref.SizedBlobRef, error) {
	// TODO: Implement stub
	return blobref.SizedBlobRef{}, nil
}

func (gs *Storage) StatBlobs(dest chan<- blobref.SizedBlobRef, blobs []*blobref.BlobRef, wait time.Duration) error {
	// TODO: Implement stub
	return errors.New("TODO: implement StatBlobs")
}

func (gs *Storage) FetchStreaming(blob *blobref.BlobRef) (io.ReadCloser, int64, error) {
	// TODO: Implement stub
	return nil, 0, errors.New("TODO: implement FetchStreaming")
}

func (gs *Storage) RemoveBlobs(blobs []*blobref.BlobRef) error {
	// TODO: Implement stub
	return errors.New("TODO: implement RemoveBlobs")
}

func (gs *Storage) GetBlobHub() blobserver.BlobHub {
	return gs.hub
}
