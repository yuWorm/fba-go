package service

import (
	"context"
	"path"
	"strings"

	"github.com/yuWorm/fba-go/core/pagination"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type LogService struct {
	repo repo.Repository
}

func NewLogService(repository repo.Repository) *LogService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &LogService{repo: repository}
}

func (s *LogService) ListLogin(ctx context.Context, filter repo.LogFilter, page int, size int, basePath string) (pagination.PageData[dto.LoginLogDetail], error) {
	items, total, err := s.repo.ListLoginLogs(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.LoginLogDetail]{}, err
	}
	return pagination.NewPageData(dto.LoginLogsFromModel(items), total, page, size, basePath), nil
}

func (s *LogService) DeleteLogin(ctx context.Context, ids []int) error {
	return s.repo.DeleteLoginLogs(ctx, ids)
}

func (s *LogService) ClearLogin(ctx context.Context) error {
	return s.repo.DeleteAllLoginLogs(ctx)
}

func (s *LogService) ListOpera(ctx context.Context, filter repo.LogFilter, page int, size int, basePath string) (pagination.PageData[dto.OperaLogDetail], error) {
	items, total, err := s.repo.ListOperaLogs(ctx, filter, page, size)
	if err != nil {
		return pagination.PageData[dto.OperaLogDetail]{}, err
	}
	return pagination.NewPageData(dto.OperaLogsFromModel(items), total, page, size, basePath), nil
}

func (s *LogService) DeleteOpera(ctx context.Context, ids []int) error {
	return s.repo.DeleteOperaLogs(ctx, ids)
}

func (s *LogService) ClearOpera(ctx context.Context) error {
	return s.repo.DeleteAllOperaLogs(ctx)
}

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) Upload(_ context.Context, filename string) (dto.UploadURL, error) {
	name := sanitizeUploadFilename(filename)
	return dto.UploadURL{URL: "/static/upload/" + name}, nil
}

func sanitizeUploadFilename(filename string) string {
	name := path.Base(strings.TrimSpace(strings.ReplaceAll(filename, "\\", "/")))
	if name == "." || name == "/" || name == "" {
		return "upload.bin"
	}
	return name
}
