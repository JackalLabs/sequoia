package ssh

import (
	"io"
	"io/fs"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
	//"github.com/charmbracelet/wish"
	//"github.com/charmbracelet/wish/bubbletea"
	//"github.com/charmbracelet/wish/logging"
	//gossh "golang.org/x/crypto/ssh"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	//"github.com/charmbracelet/lipgloss"

	"github.com/rs/zerolog/log"
)

const contentsSize = 3000

type model struct {
	isFileAtTop bool
	logFileName string
	logFile     *os.File
	fileInfo    fs.FileInfo
	contents    []byte
	ready       bool
	logViewPort viewport.Model
}

type latestLogMsg []byte

type logContentMsg []byte

type logFileMsg *os.File

type errorMsg struct {
	err error
}

func LogFileCmd(logFile logFileMsg) tea.Cmd {
	return func() tea.Msg {
		return logFile
	}
}

func (m model) Init() tea.Cmd {
	file, err := os.Open(m.logFileName)
	if err != nil {
		log.Error().Err(err).Msg("failed to open log file")
		return tea.Quit
	}
	return LogFileCmd(file)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case logFileMsg:
		m.logFile = msg
		fileInfo, err := m.logFile.Stat()
		if err != nil {
			log.Error().Err(err).Msg("couldn't fetch file info")
		}
		m.fileInfo = fileInfo
		m.contents = make([]byte, contentsSize)
		m.logViewPort.SetContent(string(m.contents))
		m.logViewPort.GotoBottom()
		return m, readLatestLog(m.logFile)
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}

		if k := msg.String(); k == "r" {
			return m, readLatestLog(m.logFile)
		}

		//handle buffered content loading
		if m.logViewPort.AtTop() && !m.isFileAtTop {
			current, err := m.logFile.Seek(0, io.SeekCurrent)
			if err != nil {
				log.Error().Err(err).Msg("log file seek error")
				return m, tea.Quit
			}

			if current == 0 {
				m.isFileAtTop = true
				return m, nil
			}

			offset := current - contentsSize
			if offset < 0 {
				offset = 0
			}

			_, err = m.logFile.Seek(offset, io.SeekStart)
			if err != nil {
				log.Error().Err(err).Msg("log file seek error")
				return m, tea.Quit
			}

			buf := make([]byte, current-offset)
			n, err := m.logFile.Read(buf)
			if err != nil {
				log.Error().Err(err).Msg("failed to load more log to buffer")
				return m, tea.Quit
			}
			_, err = m.logFile.Seek(offset, io.SeekStart)
			if err != nil {
				log.Error().Err(err).Msg("log file seek error")
				return m, tea.Quit
			}

			if int64(n) != (current - offset) {
				log.Error().Msg("log buffer is not loaded correctly")
				return m, tea.Quit
			}
			oldLineCount := m.logViewPort.TotalLineCount()
			m.contents = append(buf, m.contents...)
			m.logViewPort.SetContent(string(m.contents))
			newLineCount := m.logViewPort.TotalLineCount()
			m.logViewPort.SetYOffset(newLineCount - oldLineCount)
		}
	case latestLogMsg:
		m.contents = msg
		m.logViewPort.SetContent(string(m.contents))
		m.logViewPort.GotoBottom()
	case tea.WindowSizeMsg:
		if !m.ready {
			m.logViewPort = viewport.New(msg.Width, msg.Height)

			m.ready = true

		} else {
			m.logViewPort.Width = msg.Width
			m.logViewPort.Height = msg.Height
		}
	case errorMsg:
		log.Error().Err(msg.err).Msg("bubbletea encountered an error")
		return m, tea.Quit
	}

	m.logViewPort, cmd = m.logViewPort.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return m.logViewPort.View()
}

func MakeTeaHandler(logFileName string) bubbletea.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		log.Info().Str("user", s.User()).Str("addr", s.RemoteAddr().String()).Msg("New ssh connection request")
		m := model{
			ready:       false,
			logFileName: logFileName,
		}

		return m, nil

	}
}

func readLatestLog(logFile *os.File) tea.Cmd {
	return func() tea.Msg {
		fileInfo, err := logFile.Stat()
		if err != nil {
			return errorMsg{err}
		}
		fileSize := fileInfo.Size()

		bufferSize := contentsSize
		offset := fileSize - contentsSize
		if offset < 0 {
			offset = 0
			bufferSize = int(fileSize)
		}

		contents := make([]byte, bufferSize)
		n, err := logFile.Seek(offset, io.SeekStart)
		if err != nil {
			return errorMsg{err}
		}

		if n < 0 {
			n, err = logFile.Seek(0, io.SeekStart)
			if err != nil {
				return errorMsg{err}
			}
		}

		_, err = logFile.Read(contents)
		if err != nil {
			return errorMsg{err}
		}

		return latestLogMsg(contents)
	}
}
