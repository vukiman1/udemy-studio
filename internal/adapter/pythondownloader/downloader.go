package pythondownloader

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/platform/processgroup"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	accessTokenEnv     = "UDEMY_ACCESS_TOKEN" //nolint:gosec // reason: the name of the variable, not a credential
	eventsFileArgument = "/dev/fd/3"
	interruptGrace     = 15 * time.Second
	logFileMode        = 0o600
	logDirMode         = 0o750
	logTailBytes       = 2048
	logUnavailable     = "(log unavailable)"
	maxEventLineBytes  = 1 << 20
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type Config struct {
	Python          string
	RepoDir         string
	OutputDir       string
	LogDir          string
	CaptionLanguage string
	KeepAwake       bool
}

type Downloader struct {
	config Config
	logger *slog.Logger
}

func New(config Config, logger *slog.Logger) *Downloader {
	return &Downloader{config: config, logger: logger}
}

type pythonEvent struct {
	Event         string   `json:"event"`
	CourseTitle   string   `json:"course_title"`
	TotalLectures int      `json:"total_lectures"`
	Title         string   `json:"title"`
	Failures      []string `json:"failures"`
}

func (downloader *Downloader) Download(ctx context.Context, job jobs.Job, token secret.Secret, onEvent func(jobs.Event)) error {
	logPath := filepath.Join(downloader.config.LogDir, job.ID+".log")
	logFile, err := openLog(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	eventsReader, eventsWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("create events pipe: %w", err)
	}
	defer eventsReader.Close()

	command := downloader.command(ctx, job, token, logFile, eventsWriter)
	if err := command.Start(); err != nil {
		_ = eventsWriter.Close()
		return fmt.Errorf("start downloader %s: %w", downloader.config.Python, err)
	}
	_ = eventsWriter.Close()
	if downloader.config.KeepAwake {
		downloader.keepAwake(command.Process.Pid)
	}

	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		downloader.forwardEvents(eventsReader, onEvent)
	}()
	waitErr := command.Wait()
	processgroup.Kill(command.Process)
	<-eventsDone

	if ctx.Err() != nil {
		return ctx.Err()
	}
	if waitErr != nil {
		return fmt.Errorf("downloader failed (%w); log %s ends with: %s", waitErr, logPath, lastLogLine(logPath))
	}
	return nil
}

func (downloader *Downloader) command(ctx context.Context, job jobs.Job, token secret.Secret, logFile *os.File, events *os.File) *exec.Cmd {
	arguments := []string{
		"-m", "udemy_downloader", job.Request.CourseURL,
		"--download-captions", "--download-assets",
		"-l", downloader.config.CaptionLanguage,
		"-o", downloader.config.OutputDir,
		"--events-json", eventsFileArgument,
	}
	if job.Request.Chapters != "" {
		arguments = append(arguments, "--chapter", job.Request.Chapters)
	}
	command := exec.CommandContext(ctx, downloader.config.Python, arguments...) //nolint:gosec // reason: the binary is operator configuration and arguments are validated job fields
	command.Dir = downloader.config.RepoDir
	command.Env = append(os.Environ(),
		accessTokenEnv+"="+token.Reveal(),
		"PATH="+filepath.Dir(downloader.config.Python)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	command.Stdout, command.Stderr = logFile, logFile
	command.ExtraFiles = []*os.File{events}
	processgroup.Interruptible(command, interruptGrace)
	return command
}

func (downloader *Downloader) forwardEvents(events io.Reader, onEvent func(jobs.Event)) {
	scanner := bufio.NewScanner(events)
	scanner.Buffer(nil, maxEventLineBytes)
	for scanner.Scan() {
		var decoded pythonEvent
		if err := json.Unmarshal(scanner.Bytes(), &decoded); err != nil {
			downloader.logger.Warn("skipping malformed downloader event", "line", scanner.Text(), "err", err)
			continue
		}
		onEvent(jobs.Event{
			Kind:          jobs.EventKind(decoded.Event),
			CourseTitle:   decoded.CourseTitle,
			TotalLectures: decoded.TotalLectures,
			LectureTitle:  decoded.Title,
			Failures:      decoded.Failures,
		})
	}
	if err := scanner.Err(); err != nil {
		downloader.logger.Warn("reading downloader events stopped", "err", err)
	}
	// Python blocks writing to a pipe nobody reads, so the rest is drained even when it cannot be parsed.
	_, _ = io.Copy(io.Discard, events)
}

func (downloader *Downloader) keepAwake(pid int) {
	//nolint:noctx,gosec // reason: caffeinate exits with the watched process, and its only argument is our child's pid
	caffeinate := exec.Command("caffeinate", "-i", "-w", strconv.Itoa(pid))
	if err := caffeinate.Start(); err != nil {
		downloader.logger.Warn("cannot keep the machine awake during the download", "err", err)
		return
	}
	go func() { _ = caffeinate.Wait() }()
}

func openLog(logPath string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), logDirMode); err != nil {
		return nil, fmt.Errorf("create downloader log dir: %w", err)
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, logFileMode) //nolint:gosec // reason: path is the configured log dir plus a generated job id
	if err != nil {
		return nil, fmt.Errorf("create downloader log %s: %w", logPath, err)
	}
	return logFile, nil
}

func lastLogLine(logPath string) string {
	logFile, err := os.Open(logPath) //nolint:gosec // reason: we wrote this log ourselves
	if err != nil {
		return logUnavailable
	}
	defer logFile.Close()
	info, err := logFile.Stat()
	if err != nil {
		return logUnavailable
	}
	start := max(info.Size()-logTailBytes, 0)
	tail, err := io.ReadAll(io.NewSectionReader(logFile, start, info.Size()-start))
	if err != nil {
		return logUnavailable
	}
	lines := strings.Split(strings.TrimSpace(ansiEscape.ReplaceAllString(string(bytes.ToValidUTF8(tail, nil)), "")), "\n")
	return lines[len(lines)-1]
}
