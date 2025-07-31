package registry

import (
	"log/slog"
	"os"
	"path/filepath"
)

// TODO: actually implement function, return proper error and test it
func sanitizelabelPair(labelName, labelValue string) (string, string) {
	// slog.Error("Skipping label for metric", "metric", metricRecord.Name, "labelName", labelName, "labelValue", labelValue, "error", err)
	return labelName, labelValue
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
