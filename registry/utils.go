package registry

import (
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var labelValueSanitizer = strings.NewReplacer(
	"\n", "",
	"\\", "",
	"\"", "",
)

var validLabelNameRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func sanitizelabelPair(labelName, labelValue string) (string, string, bool) {
	if !validLabelNameRE.MatchString(labelName) {
		slog.Warn("Dropping invalid metric label name", "original", labelName)
		return "", "", false
	}

	cleanValue := labelValueSanitizer.Replace(labelValue)

	return labelName, cleanValue, true
}

func MustWriteTextfile(filePath string, fileContent string) {
	tmpDir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(tmpDir, "ethtool_exporter.prom-*")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(fileContent))
	if err != nil {
		// Hard to cover by unit tests, because it would require to brake\overflow the filesystem, or mock the os.File
		//coverage:ignore
		slog.Error("Cannot write formated metric to file", "file", tmpFile.Name(), "error", err)
		panic(err)
	}
	err = tmpFile.Close()
	if err != nil {
		//coverage:ignore
		panic(err)
	}

	// Not sure, we should check how default mod is set
	// if err := os.Chmod(tmpFile.Name(), 0o644); err != nil {
	// 	return err
	// }
	err = os.Rename(tmpFile.Name(), filePath)
	if err != nil {
		panic(err)
	}
}
