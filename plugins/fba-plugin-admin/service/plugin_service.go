package service

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type PluginService struct {
	repo repo.Repository
}

func NewPluginService(repository repo.Repository) *PluginService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &PluginService{repo: repository}
}

func (s *PluginService) All(ctx context.Context) ([]dto.PluginConfigDetail, error) {
	items, err := s.repo.AllPlugins(ctx)
	if err != nil {
		return nil, err
	}
	return dto.PluginsFromModel(items), nil
}

func (s *PluginService) Changed(ctx context.Context) (bool, error) {
	return s.repo.PluginsChanged(ctx)
}

func (s *PluginService) Install(ctx context.Context, pluginType string, repoURL string) error {
	_, err := s.repo.InstallPlugin(ctx, dto.PluginInstallParam{
		Type:    pluginType,
		RepoURL: repoURL,
		Name:    pluginNameFromRepoURL(repoURL),
	})
	return err
}

func (s *PluginService) Uninstall(ctx context.Context, name string) error {
	return s.repo.UninstallPlugin(ctx, name)
}

func (s *PluginService) ToggleStatus(ctx context.Context, name string) error {
	return s.repo.TogglePluginStatus(ctx, name)
}

func (s *PluginService) Download(ctx context.Context, name string) (string, error) {
	item, err := s.repo.GetPlugin(ctx, name)
	if err != nil {
		return "", err
	}
	return "plugin " + item.ID + " package", nil
}

func pluginNameFromRepoURL(raw string) string {
	if raw == "" {
		return "plugin"
	}
	parsed, err := url.Parse(raw)
	source := raw
	if err == nil && parsed.Path != "" {
		source = parsed.Path
	}
	name := strings.TrimSuffix(path.Base(strings.TrimRight(source, "/")), ".git")
	if name == "." || name == "/" || name == "" {
		return "plugin"
	}
	return name
}
