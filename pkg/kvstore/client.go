package kvstore

import (
	"context"
	"fmt"

	"github.com/ccfish2/infra/pkg/option"
)

var (
	defaultClient    BackendOperations
	defaultClientset = make(chan struct{})
)

func Client() BackendOperations {
	<-defaultClientset
	return defaultClient
}

func initClient(ctx context.Context, module backendModule, opts *ExtraOptions) error {
	c, errChan := module.newClient(ctx, opts)
	if c == nil {
		err := <-errChan
		return err
	}
	defaultClient = c
	select {
	case <-defaultClientset:
	default:
		close(defaultClientset)
	}
	go func() {
		err, isErr := <-errChan
		if isErr && err != nil {
			fmt.Println("unable to connect to kvstore")
		}
		if !option.Config.JoinCluster {
			// delete legacy prefixes
		}
	}()
	return nil
}
