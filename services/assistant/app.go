// Package assistant owns release-assistant transport assembly.
package assistant

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/nostosx/agentx-contrib/filestore"
	application "github.com/nostosx/agentx-examples/application/assistant"
	"github.com/nostosx/agentx-examples/application/assistant/command"
	agentinfra "github.com/nostosx/agentx-examples/infrastructure/agent"
	"github.com/nostosx/agentx-examples/infrastructure/config"
	"github.com/nostosx/agentx-examples/infrastructure/logging"
	serverinfra "github.com/nostosx/agentx-examples/infrastructure/server"
	serviceapp "github.com/nostosx/agentx-examples/services/assistant/internal/application"
	httpport "github.com/nostosx/agentx-examples/services/assistant/internal/port/http"
	"github.com/nostosx/agentx/model"
	"github.com/nostosx/agentx/server"
	"github.com/nostosx/servex-contrib/transport/httpx"
	"github.com/nostosx/servex/app"
	"github.com/nostosx/servex/cqrs"
	"github.com/nostosx/servex/health"
)

type Service struct {
	protocols *server.Server
	token     string
}

func New(m model.Model, token, stateDir string, middleware ...cqrs.CommandMiddleware[command.DeployBuildCommand, command.DeployBuildResult]) (*Service, error) {
	if m == nil {
		return nil, errors.New("assistant requires a model")
	}
	cp, err := filestore.Checkpoints(filepath.Join(stateDir, "checkpoints"))
	if err != nil {
		return nil, err
	}
	sessions, err := filestore.Sessions(filepath.Join(stateDir, "sessions"))
	if err != nil {
		return nil, err
	}
	commands := serviceapp.NewCommand(serviceapp.CommandDeps{DeployBuild: command.NewDeployBuildHandler(), Middleware: middleware})
	protocols := server.New(application.New(m, cp, sessions, commands.DeployBuild), server.WithCheckpointer(cp), server.WithCORS("http://localhost:3000"), server.WithA2A(server.A2AConfig{Name: "ops", Description: "Deploys builds after approval."}))
	return &Service{protocols: protocols, token: token}, nil
}
func (s *Service) Register(ctx context.Context, http *httpx.Server) {
	http.HandleHTTP("/", httpport.CancelWithOwner(ctx, httpport.RequireToken(s.token, s.protocols)))
}
func (s *Service) Close() { s.protocols.Close() }

// Run assembles the assistant and owns protocols and transport lifecycle.
func Run(ctx context.Context, cfg config.Config) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	log, closeLog, err := logging.New(cfg.LogLevel, nil, cfg.LogFile)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, closeLog()) }()
	if err := cfg.Validate(); err != nil {
		return err
	}
	m, err := agentinfra.ModelFromEnv()
	if err != nil {
		return err
	}
	service, err := New(m, cfg.Token, ".agentx")
	if err != nil {
		return err
	}
	defer service.Close()
	var application *app.App
	check := health.CheckerFunc(func(ctx context.Context) error { return application.Check(ctx) })
	servers, err := serverinfra.New(ctx, cfg, check, log)
	if err != nil {
		return err
	}
	service.Register(ctx, servers.HTTP)
	application = app.New(app.WithLogger(log), app.WithHooks(servers.HTTP), app.WithStopTimeout(10*time.Second))
	return application.Run(ctx)
}
