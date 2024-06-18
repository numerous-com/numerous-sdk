package appdevsession

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"

	"numerous.com/cli/internal/appdev"
)

type appRunner struct {
	executor             commandExecutor
	appSessions          appdev.AppSessionRepository
	appSessionService    appdev.AppSessionService
	port                 string
	pythonInterpeterPath string
	appModulePath        string
	appClassName         string
	exit                 chan struct{}
	output               appdev.Output
	cmd                  command
}

// Runs the app, reading the definition from the app file, killing app process
// if it exists, updating the session (if it exists) with the new definition,
// and launching a python process running the app.
func (r *appRunner) Run() {
	defer r.output.AwaitingAppChanges()

	if r.cmd != nil {
		if err := r.cmd.Kill(); err != nil {
			slog.Info("error killing python app process", slog.String("error", err.Error()))
		}
	}

	appdef := r.readApp()
	if appdef == nil {
		r.cmd = nil
		return
	}

	session, err := r.updateExistingSession(appdef)
	if err != nil {
		if session, err = r.appSessions.Create(*appdef); err != nil {
			r.output.ErrorCreatingAppSession(err)
			r.signalExit()
			r.cmd = nil

			return
		}
		slog.Debug("created session", slog.String("name", session.Name), slog.Any("id", session.ID), slog.Int("elements", len(session.Elements)))
	}

	r.cmd = r.runApp(*session)
}

func (r *appRunner) readApp() *appdev.AppDefinition {
	cmd := r.executor.Create(r.pythonInterpeterPath, "-m", "numerous", "read", r.appModulePath, r.appClassName)
	output, err := cmd.Output()
	if err != nil {
		slog.Debug("Error reading app definition", slog.String("error", err.Error()))
		r.output.ErrorReadingApp(string(output), err)

		return nil
	}

	result, err := appdev.ParseAppDefinition(output)
	if err != nil {
		slog.Warn("Error parsing app definition", slog.String("error", err.Error()))
		r.output.ErrorParsingApp(err)

		return nil
	}

	if result.App != nil {
		slog.Debug("Read app definition", slog.String("name", result.App.Name))
		return result.App
	} else if result.Error != nil {
		r.output.ErrorLoadingApp(result.Error)
	}

	return nil
}

func (r *appRunner) updateExistingSession(def *appdev.AppDefinition) (*appdev.AppSession, error) {
	existingSession, err := r.appSessions.Read(0)
	if err != nil {
		return nil, err
	}

	diff := appdev.GetAppSessionDifference(*existingSession, *def)
	slog.Debug("got app session difference", slog.Int("added", len(diff.Added)), slog.Int("removed", len(diff.Removed)))
	for _, added := range diff.Added {
		if addedSession, err := r.appSessionService.AddElement("server", added); err == nil {
			slog.Debug("added element", slog.Any("elementID", added.ID), slog.String("name", added.Name))
			existingSession = addedSession
		} else {
			slog.Debug("error adding element after update", slog.String("error", err.Error()))
			r.output.ErrorUpdateAddingElement()
		}
	}

	for _, removed := range diff.Removed {
		slog.Debug("handling removed element", slog.Any("element", removed))
		if removedSession, err := r.appSessionService.RemoveElement("server", removed); err == nil {
			existingSession = removedSession
		} else {
			slog.Info("error removing element after update", slog.String("error", err.Error()))
			r.output.ErrorUpdateRemovingElement()
		}
	}

	for _, updated := range diff.Updated {
		slog.Debug("handling updated element", slog.Any("element", updated))
		if updatedSession, err := r.appSessionService.UpdateElementLabel("server", updated); err == nil {
			existingSession = updatedSession
		} else {
			slog.Info("error updating element after update", slog.String("error", err.Error()))
			r.output.ErrorUpdateUpdatingElement()
		}
	}

	return existingSession, nil
}

func (r *appRunner) runApp(session appdev.AppSession) command {
	cmd := r.executor.Create(
		r.pythonInterpeterPath,
		"-m",
		"numerous",
		"run",
		"--graphql-url",
		fmt.Sprintf("http://localhost:%s/query", r.port),
		"--graphql-ws-url",
		fmt.Sprintf("ws://localhost:%s/query", r.port),
		r.appModulePath,
		r.appClassName,
		strconv.FormatUint(uint64(session.ID), 10),
	)
	cmd.SetEnv("PYTHONUNBUFFERED", "1")
	cmd.SetEnv("PATH", os.Getenv("PATH"))
	r.wrapAppOutput(cmd.StdoutPipe, "stdout")
	r.wrapAppOutput(cmd.StderrPipe, "stderr")

	if err := cmd.Start(); err != nil {
		slog.Debug("could not start app", slog.String("error", err.Error()))
		r.output.ErrorStartingApp(err)

		return nil
	}
	r.output.StartedApp()

	return cmd
}

func (r *appRunner) wrapAppOutput(pipeFunc func() (io.ReadCloser, error), streamName string) {
	if reader, err := pipeFunc(); err != nil {
		r.output.ErrorGettingAppOutputStream(streamName)
	} else {
		scanner := bufio.NewScanner(reader)
		go func() {
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				r.output.PrintAppLogLine(scanner.Text())
			}
		}()
	}
}

func (r *appRunner) signalExit() {
	r.exit <- struct{}{}
}
