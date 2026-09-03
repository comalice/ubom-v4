package main

import (
	"errors"

	"ubom-v4/app"
	"ubom-v4/artifact"
	"ubom-v4/store"
)

type Config struct {
	HTTP      HTTPConfig
	Database  DatabaseConfig
	Artifacts ArtifactConfig
}

type HTTPConfig struct {
	ListenAddress string
}

type DatabaseConfig struct {
	Path string
	DSN  string
}

type ArtifactConfig struct {
	SnapshotRoot string
	LocalSource  LocalSourceConfig
}

type LocalSourceConfig struct {
	ID   string
	Root string
}

func (c ArtifactConfig) snapshotService(metadata store.Store) (*app.SnapshotService, error) {
	if c.SnapshotRoot == "" && c.LocalSource.Root == "" {
		return nil, nil
	}
	if c.SnapshotRoot == "" || c.LocalSource.ID == "" || c.LocalSource.Root == "" {
		return nil, errors.New("artifact snapshot root, local source ID, and local source root must be configured together")
	}
	source, err := metadata.GetSource(artifact.SourceID(c.LocalSource.ID))
	if err != nil {
		return nil, err
	}
	adapter, err := app.NewLocalFileAdapter(app.LocalFileAdapterConfig{Source: source, Root: c.LocalSource.Root})
	if err != nil {
		return nil, err
	}
	content, err := artifact.NewFileSnapshotStore(artifact.FileSnapshotStoreConfig{Root: c.SnapshotRoot})
	if err != nil {
		return nil, err
	}
	return app.NewSnapshotService(metadata, content, adapter), nil
}
