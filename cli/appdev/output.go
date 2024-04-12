package appdev

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Output interface {
	StartingApp(port string)
	StartedApp()
	AppModuleNotFound(err error)
	AwaitingAppChanges()
	ErrorReadingApp(readOutput string, err error)
	ErrorParsingApp(err error)
	ErrorStartingApp(err error)
	ErrorWatchingAppFiles(err error)
	FileUpdatedRestartingApp()
	ErrorUpdateAddingElement()
	ErrorUpdateRemovingElement()
	ErrorUpdateUpdatingElement()
	ErrorCreatingAppSession(err error)
	ErrorGettingAppOutputStream(streamName string)
	PrintAppLogLine(line string)
	Stopping()
	ErrorLoadingApp(appError *ParseAppDefinitionError)
}

type FmtOutput struct {
	appModulePath string
	appClassName  string
}

func (o *FmtOutput) StartingApp(port string) {
	fmt.Printf("Starting %s:%s at http://localhost:%s\n", o.appModulePath, o.appClassName, port)
}

func (o *FmtOutput) StartedApp() {
	fmt.Printf("Started %s:%s\n", o.appModulePath, o.appClassName)
}

func (o *FmtOutput) AppModuleNotFound(err error) {
	fmt.Printf("Could not find app module %s: %s\n", o.appModulePath, err)
}

func (o *FmtOutput) AwaitingAppChanges() {
	fmt.Printf("Waiting for changes to '%s'...\n", o.appModulePath)
}

func (o *FmtOutput) ErrorReadingApp(readOutput string, err error) {
	fmt.Printf("An error occurred while reading the app definition of %s:%s!\n", o.appModulePath, o.appClassName)
}

func (o *FmtOutput) ErrorParsingApp(err error) {
	fmt.Println("The app definition was not valid.")
}

func (o *FmtOutput) ErrorStartingApp(err error) {
	fmt.Printf("Could not start %s:%s", o.appModulePath, o.appClassName)
	fmt.Printf("Error: %s", err)
}

func (o *FmtOutput) ErrorWatchingAppFiles(err error) {
	fmt.Printf("file watcher for %s did not start: %s\n", o.appModulePath, err.Error())
}

func (o *FmtOutput) FileUpdatedRestartingApp() {
	fmt.Printf("watcher: %s updated, restarting app.\n", o.appModulePath)
}

func (o *FmtOutput) ErrorUpdateAddingElement() {
	fmt.Println("Could not update app, error adding elements.")
}

func (o *FmtOutput) ErrorUpdateRemovingElement() {
	fmt.Println("Could not update app, error removing elements.")
}

func (o *FmtOutput) ErrorUpdateUpdatingElement() {
	fmt.Println("Could not update app, error updating elements.")
}

func (o *FmtOutput) ErrorCreatingAppSession(err error) {
	fmt.Printf("Could not create app session: %s\n", err)
}

func (o *FmtOutput) ErrorGettingAppOutputStream(streamName string) {
	fmt.Printf("%s:%s> Could not access stream %s in app\n", o.appModulePath, o.appClassName, streamName)
}

func (o *FmtOutput) PrintAppLogLine(line string) {
	fmt.Printf("%s:%s> %s", o.appModulePath, o.appClassName, line)
}

func (o *FmtOutput) Stopping() {
	fmt.Println("Stopping development server")
}

func (o *FmtOutput) ErrorLoadingApp(appError *ParseAppDefinitionError) {
	switch {
	case appError.ModuleNotFound != nil:
		fmt.Printf("The module '%s' imported in your python code was not found\n", appError.ModuleNotFound.Module)
		fmt.Println("Common reasons for this include:")
		fmt.Println(" * Some of your external dependencies have not been installed")
		fmt.Println(" * You have not activated the correct virtual environment")
		fmt.Println(" * Or there might be an error in your import")
	case appError.AppNotFound != nil:
		fmt.Printf("The app '%s' was not found in the specified module '%s'\n", appError.AppNotFound.App, o.appModulePath)
		if len(appError.AppNotFound.FoundApps) > 0 {
			fmt.Println("The following apps were found in the module")
			for _, app := range appError.AppNotFound.FoundApps {
				fmt.Println(" * ", app)
			}
		} else {
			fmt.Println("We found no defined apps in that module, are you sure it is the correct one?")
		}
	case appError.Syntax != nil:
		fmt.Printf("A syntax error occurred, loading your app module '%s'\n", o.appModulePath)
		fmt.Println("Syntex error message:", appError.Syntax.Msg)
		fmt.Printf("The error occurred at line %d, column %d:\n", appError.Syntax.Pos.Line, appError.Syntax.Pos.Offset)
		fmt.Println(appError.Syntax.Context)
	case appError.Unknown != nil:
		fmt.Printf("An unhandled exception of type '%s' was raised loading '%s'\n", appError.Unknown.Typename, o.appModulePath)
		fmt.Println(appError.Unknown.Traceback)
	}
}

var (
	ColorOK        = lipgloss.Color("#23DD65")
	ColorLifecycle = lipgloss.Color("#2365DD")
	ColorError     = lipgloss.Color("#FF2323")
	ColorNotice    = lipgloss.Color("#DDAA22")
)

type LipglossOutput struct {
	appModulePath     string
	appClassName      string
	okStyle           lipgloss.Style
	lifecycleStyle    lipgloss.Style
	logNameStyle      lipgloss.Style
	noticeStyle       lipgloss.Style
	errorHeaderStyle  lipgloss.Style
	errorBodyStyle    lipgloss.Style
	errorContextStyle lipgloss.Style
}

func NewLipglossOutput(appModulePath string, appClassName string) LipglossOutput {
	return LipglossOutput{
		appModulePath:    appModulePath,
		appClassName:     appClassName,
		okStyle:          lipgloss.NewStyle().Bold(true).Foreground(ColorOK),
		lifecycleStyle:   lipgloss.NewStyle().Bold(true).Foreground(ColorLifecycle),
		logNameStyle:     lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true, false, false).Faint(true),
		noticeStyle:      lipgloss.NewStyle().Foreground(ColorNotice),
		errorHeaderStyle: lipgloss.NewStyle().Bold(true).Foreground(ColorError),
		errorBodyStyle:   lipgloss.NewStyle().Foreground(ColorError),
		errorContextStyle: lipgloss.NewStyle().
			Foreground(ColorError).
			BorderForeground(ColorError).
			BorderLeft(true).
			Border(lipgloss.NormalBorder(), false, false, false, true),
	}
}

func (o *LipglossOutput) printErrorHeader(format string, args ...any) {
	o.printStyle(o.errorHeaderStyle, format, args...)
}

func (o *LipglossOutput) printErrorBody(format string, args ...any) {
	o.printStyle(o.errorBodyStyle, format, args...)
}

func (o *LipglossOutput) printErrorContext(format string, args ...any) {
	o.printStyle(o.errorContextStyle, format, args...)
}

func (o *LipglossOutput) printNotice(format string, args ...any) {
	o.printStyle(o.noticeStyle, format, args...)
}

func (o *LipglossOutput) printOK(format string, args ...any) {
	o.printStyle(o.okStyle, format, args...)
}

func (o *LipglossOutput) printLifecycle(format string, args ...any) {
	o.printStyle(o.lifecycleStyle, format, args...)
}

func (o *LipglossOutput) printStyle(style lipgloss.Style, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(style.Render(msg))
}

func (o *LipglossOutput) StartingApp(port string) {
	o.printLifecycle("Starting %s:%s app at http://localhost:%s", o.appModulePath, o.appClassName, port)
}

func (o *LipglossOutput) StartedApp() {
	o.printOK("Started %s:%s", o.appModulePath, o.appClassName)
}

func (o *LipglossOutput) AppModuleNotFound(err error) {
	o.printErrorHeader("Could not find app module %s: %s", o.appModulePath, err)
}

func (o *LipglossOutput) AwaitingAppChanges() {
	o.printLifecycle("Waiting for changes to '%s'...", o.appModulePath)
}

func (o *LipglossOutput) ErrorReadingApp(readOutput string, err error) {
	o.printErrorHeader("An error occurred while reading the app definition of %s:%s: %s", o.appModulePath, o.appClassName, err)
	o.printErrorBody("This could be due to a bug in the numerous app engine.")
	o.printErrorBody("The following error message was produced, while reading the app:")
	for _, line := range strings.Split(readOutput, "\n") {
		o.printErrorContext(line)
	}
}

func (o *LipglossOutput) ErrorParsingApp(err error) {
	o.printErrorHeader("An error occurred while reading the app definition of %s:%s: %s", o.appModulePath, o.appClassName, err)
}

func (o *LipglossOutput) ErrorStartingApp(err error) {
	o.printErrorHeader("Could not start %s:%s: %s", o.appModulePath, o.appClassName, err.Error())
}

func (o *LipglossOutput) ErrorWatchingAppFiles(err error) {
	o.printErrorHeader("Error watching '%s' for changes: %s", o.appModulePath, err.Error())
}

func (o *LipglossOutput) FileUpdatedRestartingApp() {
	o.printNotice("File '%s' changed, restarting app...", o.appModulePath)
}

func (o *LipglossOutput) ErrorUpdateAddingElement() {
	o.printErrorHeader("Could not update app, error adding elements.")
}

func (o *LipglossOutput) ErrorUpdateRemovingElement() {
	o.printErrorHeader("Could not update app, error removing elements.")
}

func (o *LipglossOutput) ErrorUpdateUpdatingElement() {
	o.printErrorHeader("Could not update app, error updating elements.")
}

func (o *LipglossOutput) ErrorCreatingAppSession(err error) {
	o.printErrorHeader("Could not create app session:" + err.Error())
}

func (o *LipglossOutput) ErrorGettingAppOutputStream(streamName string) {
	o.printErrorHeader("Could not access stream %s running %s:%s\n", streamName, o.appModulePath, o.appClassName)
}

func (o *LipglossOutput) PrintAppLogLine(line string) {
	fmt.Print(o.logNameStyle.Render(fmt.Sprintf("%s:%s", o.appModulePath, o.appClassName)))
	fmt.Println(line)
}

func (o *LipglossOutput) Stopping() {
	o.printNotice("Stopping development server")
}

func (o *LipglossOutput) ErrorLoadingApp(appError *ParseAppDefinitionError) {
	switch {
	case appError.ModuleNotFound != nil:
		o.printErrorHeader("The module '%s' imported in your python code was not found", appError.ModuleNotFound.Module)
		o.printNotice("Common reasons for this include:")
		o.printNotice(" * Some of your external dependencies have not been installed")
		o.printNotice(" * You have not activated the correct virtual environment")
		o.printNotice(" * Or there might be an error in your import")
	case appError.AppNotFound != nil:
		o.printErrorHeader("The app '%s' was not found in the specified module '%s'", appError.AppNotFound.App, o.appModulePath)
		if len(appError.AppNotFound.FoundApps) > 0 {
			o.printNotice("The following apps were found in the module")
			for _, app := range appError.AppNotFound.FoundApps {
				o.printNotice(" * %s", app)
			}
		} else {
			o.printNotice("We found no defined apps in that module, are you sure it is the correct one?")
		}
	case appError.Syntax != nil:
		o.printErrorHeader("A syntax error occurred, loading your app module '%s': %s", o.appModulePath, appError.Syntax.Msg)
		o.printErrorBody("The error occurred in \"%s\", line %d, column %d:", o.appModulePath, appError.Syntax.Pos.Line, appError.Syntax.Pos.Offset)
		o.printErrorContext(appError.Syntax.Context)
	case appError.Unknown != nil:
		o.printErrorHeader("An unhandled exception of type '%s' was raised loading '%s'", appError.Unknown.Typename, o.appModulePath)
		o.printErrorContext(appError.Unknown.Traceback)
	}
}
