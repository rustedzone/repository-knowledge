package evalharness

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var resultPathSegment = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func Record(options RecordOptions, result GradeResult) (RecordResult, error) {
	var recorded RecordResult
	if strings.TrimSpace(options.ResultsRoot) == "" {
		return recorded, fmt.Errorf("results root is required")
	}
	if _, err := time.Parse("2006-01-02", options.RunDate); err != nil {
		return recorded, fmt.Errorf("run date must use YYYY-MM-DD: %w", err)
	}
	if !resultPathSegment.MatchString(result.CaseID) || !resultPathSegment.MatchString(result.Agent) || !resultPathSegment.MatchString(result.Condition) {
		return recorded, fmt.Errorf("case, agent, and condition must be safe result path segments")
	}
	if result.TrialNumber < 1 {
		return recorded, fmt.Errorf("trial number must be at least 1")
	}
	if result.DurationMillis < 1 {
		return recorded, fmt.Errorf("a positive duration is required for recorded results")
	}
	artifactInfo, err := os.Stat(options.Artifact)
	if err != nil {
		return recorded, fmt.Errorf("inspect preserved artifact: %w", err)
	}
	if !artifactInfo.Mode().IsRegular() {
		return recorded, fmt.Errorf("preserved artifact must be a regular file")
	}

	resultsRoot, err := filepath.Abs(options.ResultsRoot)
	if err != nil {
		return recorded, fmt.Errorf("resolve results root: %w", err)
	}
	directory := filepath.Join(resultsRoot, result.CaseID, result.Agent)
	stem := fmt.Sprintf("%s-%s-%d", options.RunDate, result.Condition, result.TrialNumber)
	resultPath := filepath.Join(directory, stem+".json")
	artifactPath := filepath.Join(directory, stem+"-artifact"+artifactExtension(options.Artifact))
	for _, path := range []string{resultPath, artifactPath} {
		if _, err := os.Lstat(path); err == nil {
			return recorded, fmt.Errorf("immutable evaluation output already exists: %s", path)
		} else if !os.IsNotExist(err) {
			return recorded, fmt.Errorf("inspect evaluation output %s: %w", path, err)
		}
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return recorded, fmt.Errorf("create evaluation result directory: %w", err)
	}
	if err := copyRegularFileExclusive(options.Artifact, artifactPath); err != nil {
		return recorded, fmt.Errorf("preserve evaluation artifact: %w", err)
	}
	relativeArtifact, err := filepath.Rel(resultsRoot, artifactPath)
	if err != nil {
		_ = os.Remove(artifactPath)
		return recorded, err
	}
	result.Target = ""
	familyDirectory := "cases"
	if result.Family == FamilyBenchmark {
		familyDirectory = "benchmarks"
	}
	result.Rubric = filepath.ToSlash(filepath.Join(familyDirectory, result.CaseID, filepath.Base(result.Rubric)))
	result.RunDate = options.RunDate
	result.PreservedArtifact = filepath.ToSlash(relativeArtifact)
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		_ = os.Remove(artifactPath)
		return recorded, err
	}
	data = append(data, '\n')
	if err := writeExclusive(resultPath, data); err != nil {
		_ = os.Remove(artifactPath)
		return recorded, fmt.Errorf("write immutable evaluation result: %w", err)
	}
	return RecordResult{ResultPath: resultPath, ArtifactPath: artifactPath, Result: result}, nil
}

func artifactExtension(path string) string {
	lower := strings.ToLower(path)
	for _, extension := range []string{".tar.gz", ".tar.zst", ".patch", ".diff", ".zip"} {
		if strings.HasSuffix(lower, extension) {
			return extension
		}
	}
	if extension := filepath.Ext(path); extension != "" {
		return extension
	}
	return ".artifact"
}

func copyRegularFileExclusive(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		_ = os.Remove(destination)
		return err
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(destination)
		return err
	}
	return nil
}

func writeExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}
