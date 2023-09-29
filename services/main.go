package services

import (
	"context"
	"flag"
	"fmt"
	"github.com/sjmshsh/HopeIM/logger"

	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "kim",
		Version: version,
		Short:   "King IM Cloud",
	}
	ctx := context.Background()

	fmt.Println(ctx)
	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
