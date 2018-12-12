package sourcetype

import (
	"fmt"

	"github.com/charter-se/structured/errors"
)

const (
	Unset   = iota
	Unknown = iota
	Missing = iota
	Git     = iota
	File    = iota
	Dir     = iota
)

type SourceType int

func Parse(s string) (SourceType, error) {
	switch s {
	case "":
		return SourceType(Missing), nil
	case "git":
		return SourceType(Git), nil
	case "file":
		return SourceType(File), nil
	case "dir":
		return SourceType(Dir), nil
	default:
		return SourceType(Unknown), nil
	}
	return SourceType(Unknown), errors.WithFields(errors.Fields{"Type": s}).New("Failed to parse sourcetype")
}

func Print(st SourceType) string {
	switch st {
	case Unset:
		return "unset"
	case Unknown:
		return "unknown"
	case Git:
		return "git"
	case File:
		return "file"
	case Dir:
		return "dir"
	default:
		return fmt.Sprintf("Unknown type: %v", st)
	}
}
