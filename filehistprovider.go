package ic0bra

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type FileHistoryProvider struct {
	histDir string
}

func NewFileHistoryProvider(appName string) (*FileHistoryProvider, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("error while looking for local user config dir")
	}
	configDir := filepath.Join(userConfigDir, appName, "history")
	if _, err := os.ReadDir(configDir); err != nil {
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			return nil, fmt.Errorf("error while creating config dir to store the history")
		}
	}
	return &FileHistoryProvider{
		histDir: configDir,
	}, nil // TODO
}

func (p *FileHistoryProvider) GetHistContent(flagName string) ([]string, error) {
	ret := make([]string, 0)
	histFileName := p.GetHistFileName(flagName)
	file, err := os.Open(histFileName)
	if err != nil {
		return []string{}, fmt.Errorf("error while saving history: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}
	return ret, nil
}

func (p *FileHistoryProvider) HistDir() string {
	return p.histDir
}

func (p *FileHistoryProvider) InputFromHist(flagName, txt string) (string, error) {
	histContent, err := p.GetHistContent(flagName)
	if err != nil {
		return "", err
	}
	return selectionFactory(fmt.Sprintf("Select the previous input for '--%s': ", flagName), histContent)
}

func (p *FileHistoryProvider) GetHistFileName(flagName string) string {
	histFileName := flagName + ".hist"
	return filepath.Join(p.histDir, histFileName)
}

func (p *FileHistoryProvider) HasHist(flagName string) bool {
	if _, err := os.Stat(p.GetHistFileName(flagName)); err != nil {
		return false
	}
	return true
}

func (p *FileHistoryProvider) SaveHist(flagName, value string) error {
	histFileName := p.GetHistFileName(flagName)
	if file, err := os.Open(histFileName); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if value == scanner.Text() {
				return nil // already stored
			}
		}
	}
	f, err := os.OpenFile(histFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("error while opening hist file for append: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(value + "\n"); err != nil {
		return fmt.Errorf("error while writing history: %v", err)
	}
	return nil
}
