package gateway

import (
	"context"
	"github.com/sjmshsh/HopeIM"
	"github.com/sjmshsh/HopeIM/container"
	"github.com/sjmshsh/HopeIM/logger"
	"github.com/sjmshsh/HopeIM/naming"
	"github.com/sjmshsh/HopeIM/naming/consul"
	"github.com/sjmshsh/HopeIM/services/gateway/conf"
	"github.com/sjmshsh/HopeIM/services/gateway/serv"
	"github.com/sjmshsh/HopeIM/websocket"
	"github.com/sjmshsh/HopeIM/wire"
	"github.com/spf13/cobra"
	"time"
)

type ServerStartOptions struct {
	config   string
	protocol string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Start a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./gateway/conf.yaml", "Config file")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol of ws or tcp")
	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}
	_ = logger.Init(logger.Settings{
		Level: "trace",
	})

	handler := &serv.Handler{
		ServiceID: config.ServiceID,
	}

	var srv HopeIM.Server
	service := &naming.DefaultService{
		Id:       config.ServiceID,
		Name:     config.ServiceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: opts.protocol,
		Tags:     config.Tags,
	}
	if opts.protocol == "ws" {
		srv = websocket.NewServer(config.Listen, service)
	}

	srv.SetReadWait(time.Minute * 2)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	_ = container.Init(srv, wire.SNChat, wire.SNLogin)

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)

	// set a dialer
	container.SetDialer(serv.NewDialer(config.ServiceID))

	return container.Start()
}
