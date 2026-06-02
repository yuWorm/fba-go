package service

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/url"
	"path"
	"strings"

	fbaerrors "github.com/yuWorm/fba-go/core/errors"
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
	if pluginType == "zip" {
		// The Go compatibility host does not accept multipart plugin uploads yet; preserve Python's empty-file guard.
		return pluginBadRequest("ZIP 压缩包不能为空", nil)
	}
	if repoURL == "" {
		return pluginBadRequest("Git 仓库地址不能为空", nil)
	}
	_, err := s.repo.InstallPlugin(ctx, dto.PluginInstallParam{
		Type:    pluginType,
		RepoURL: repoURL,
		Name:    pluginNameFromRepoURL(repoURL),
	})
	return err
}

func (s *PluginService) Uninstall(ctx context.Context, name string) error {
	if err := s.repo.UninstallPlugin(ctx, name); err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return pluginNotFound("插件不存在", err)
		}
		return err
	}
	return nil
}

func (s *PluginService) ToggleStatus(ctx context.Context, name string) error {
	if err := s.repo.TogglePluginStatus(ctx, name); err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return pluginNotFound("插件不存在", err)
		}
		return err
	}
	return nil
}

func (s *PluginService) Download(ctx context.Context, name string) (string, error) {
	item, err := s.repo.GetPlugin(ctx, name)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return "", pluginNotFound("插件不存在", err)
		}
		return "", err
	}
	return "plugin " + item.ID + " package", nil
}

func pluginBadRequest(message string, cause error) error {
	return fbaerrors.New(http.StatusBadRequest, http.StatusBadRequest, message, cause)
}

func pluginNotFound(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
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
