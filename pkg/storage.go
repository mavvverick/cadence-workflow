package pkg

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
)

type gcsUtil struct{}

func (*gcsUtil) DownloadObjectToLocal(bucket, object, localDirectory string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	readContext, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return err
	}

	defer readContext.Close()

	data, err := ioutil.ReadAll(readContext)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(localDirectory, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
