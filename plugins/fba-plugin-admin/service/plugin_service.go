package service

import (
	"context"
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/yuWorm/fba-go/core/config"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type PluginService struct {
	repo        repo.Repository
	environment string
}

func NewPluginService(repository repo.Repository) *PluginService {
	return NewPluginServiceWithConfig(repository, config.Options{})
}

func NewPluginServiceWithConfig(repository repo.Repository, opts config.Options) *PluginService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	opts = opts.WithDefaults()
	return &PluginService{repo: repository, environment: opts.App.Environment}
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

func (s *PluginService) Install(context.Context, string, string) error {
	// Go module plugins are compiled into the host binary. Keep the Python API
	// route for frontend compatibility, but make the unsupported runtime behavior explicit.
	return pluginBadRequest("Golang 不支持动态插件安装", nil)
}

func (s *PluginService) Uninstall(ctx context.Context, name string) error {
	if !s.isDevelopment() {
		return pluginBadRequest("禁止在非开发环境下卸载插件", nil)
	}
	item, err := s.repo.GetPlugin(ctx, name)
	if err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return pluginNotFound("插件不存在", err)
		}
		return err
	}
	if item.BuiltIn {
		// Python get_required_plugins blocks uninstall before touching plugin files;
		// Go marks those always-present module plugins as BuiltIn in the compatibility store.
		return pluginBadRequest(fmt.Sprintf("插件 %s 为必需插件，禁止卸载", name), nil)
	}
	if err := s.repo.UninstallPlugin(ctx, name); err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return pluginNotFound("插件不存在", err)
		}
		return err
	}
	return nil
}

func (s *PluginService) isDevelopment() bool {
	return strings.EqualFold(s.environment, "dev")
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
	if _, err := s.repo.GetPlugin(ctx, name); err != nil {
		if stderrors.Is(err, repo.ErrNotFound) {
			return "", pluginNotFound("插件不存在", err)
		}
		return "", err
	}
	// Python can build a zip from plugin source on disk. Go module plugins are
	// compiled into the host binary, so keep the route but do not fake a package.
	return "", pluginBadRequest("Golang 不支持动态插件打包下载", nil)
}

func pluginBadRequest(message string, cause error) error {
	return fbaerrors.New(http.StatusBadRequest, http.StatusBadRequest, message, cause)
}

func pluginNotFound(message string, cause error) error {
	return fbaerrors.New(http.StatusNotFound, http.StatusNotFound, message, cause)
}
