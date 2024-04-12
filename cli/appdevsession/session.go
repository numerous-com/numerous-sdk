package appdevsession

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"numerous/cli/appdev"
	"numerous/cli/assets"
	"numerous/cli/server"
)

type devSession struct {
	executor             commandExecutor
	fileWatcherFactory   fileWatcherFactory
	clock                clock
	appSessions          appdev.AppSessionRepository
	appSessionService    appdev.AppSessionService
	port                 string
	pythonInterpeterPath string
	appModulePath        string
	appClassName         string
	exit                 chan struct{}
	server               *http.Server
	minUpdateInterval    time.Duration
	output               appdev.Output
}

func CreateAndRunDevSession(pythonInterpreterPath string, modulePath string, className string, port string) {
	appSessions := appdev.InMemoryAppSessionRepository{}
	output := appdev.NewLipglossOutput(modulePath, className)
	session := devSession{
		executor:             &execCommandExecutor{},
		fileWatcherFactory:   &FSNotifyFileWatcherFactory{},
		clock:                &timeclock{},
		appSessions:          &appSessions,
		appSessionService:    appdev.NewAppSessionService(&appSessions),
		port:                 port,
		pythonInterpeterPath: pythonInterpreterPath,
		appModulePath:        modulePath,
		appClassName:         className,
		exit:                 make(chan struct{}, 1),
		minUpdateInterval:    time.Second,
		output:               &output,
	}
	session.run()
}

func (d *devSession) run() {
	slog.Debug("running session", slog.Any("session", d))
	d.output.StartingApp(d.port)
	go d.startServer()
	d.setupSystemSignal()
	if err := d.validateAppExists(); err != nil {
		d.output.AppModuleNotFound(err)
		os.Exit(1)
	}

	appChanges, err := WatchAppChanges(d.appModulePath, d.fileWatcherFactory, d.clock, d.minUpdateInterval)
	if err != nil {
		d.output.ErrorWatchingAppFiles(err)
		os.Exit(1)
	}

	go d.handleAppChanges(appChanges)

	d.awaitExit()
}

func (d *devSession) handleAppChanges(appChanges chan string) {
	run := appRunner{
		executor:             d.executor,
		appSessions:          d.appSessions,
		appSessionService:    d.appSessionService,
		port:                 d.port,
		pythonInterpeterPath: d.pythonInterpeterPath,
		appModulePath:        d.appModulePath,
		appClassName:         d.appClassName,
		exit:                 d.exit,
		output:               d.output,
	}
	run.Run()

	for {
		if _, ok := <-appChanges; !ok {
			break
		}

		d.output.FileUpdatedRestartingApp()
		run.Run()
	}
}

func (d *devSession) signalExit() {
	d.exit <- struct{}{}
}

func (d *devSession) startServer() {
	if d.server != nil {
		panic("can only run server once per session")
	}
	registers := []server.HandlerRegister{assets.SPAMRegister}
	d.server = server.CreateServer(server.ServerOptions{
		HTTPPort:          d.port,
		AppSessions:       d.appSessions,
		AppSessionService: d.appSessionService,
		Registers:         registers,
		GQLPath:           "/query",
		PlaygroundPath:    "/playground",
	})

	d.server.ListenAndServe() //nolint:errcheck
}

func (d *devSession) validateAppExists() error {
	_, err := os.Stat(d.appModulePath)
	return err
}

func (d *devSession) setupSystemSignal() {
	systemExitSignal := make(chan os.Signal, 1)
	go func() {
		signal.Notify(systemExitSignal, syscall.SIGINT, syscall.SIGTERM)
		<-systemExitSignal
		d.signalExit()
	}()
}

func (d *devSession) awaitExit() {
	<-d.exit
	d.output.Stopping()
	if d.server != nil {
		if err := d.server.Close(); err != nil {
			slog.Debug("error closing server", slog.String("error", err.Error()))
		}
		d.server = nil
	}
}
