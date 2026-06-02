package service

import (
	"context"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	fbaerrors "github.com/yuWorm/fba-go/core/errors"
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

func (s *LogService) DeleteLogin(ctx context.Context, ids []int) (int, error) {
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

func (s *LogService) DeleteOpera(ctx context.Context, ids []int) (int, error) {
	return s.repo.DeleteOperaLogs(ctx, ids)
}

func (s *LogService) ClearOpera(ctx context.Context) error {
	return s.repo.DeleteAllOperaLogs(ctx)
}

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) Upload(_ context.Context, filename string, size int64) (dto.UploadURL, error) {
	name, err := buildUploadFilename(filename, size, time.Now())
	if err != nil {
		return dto.UploadURL{}, err
	}
	return dto.UploadURL{URL: "/static/upload/" + name}, nil
}

func sanitizeUploadFilename(filename string) string {
	name := path.Base(strings.TrimSpace(strings.ReplaceAll(filename, "\\", "/")))
	if name == "." || name == "/" || name == "" {
		return "upload.bin"
	}
	return name
}

func buildUploadFilename(filename string, size int64, now time.Time) (string, error) {
	name := sanitizeUploadFilename(filename)
	ext := strings.TrimPrefix(strings.ToLower(path.Ext(name)), ".")
	if ext == "" {
		return "", fileBadRequest("未知的文件类型")
	}
	if err := verifyUploadFile(ext, size); err != nil {
		return "", err
	}
	base := strings.TrimSuffix(name, path.Ext(name))
	if base == "" {
		base = "upload"
	}
	return base + "_" + strconv.FormatInt(now.Unix(), 10) + "." + ext, nil
}

func verifyUploadFile(ext string, size int64) error {
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp":
		if size > 5*1024*1024 {
			return fileBadRequest("图片超出最大限制，请重新选择")
		}
	case "mp4", "mov", "avi", "flv":
		if size > 20*1024*1024 {
			return fileBadRequest("视频超出最大限制，请重新选择")
		}
	default:
		return fileBadRequest("此文件格式 " + ext + " 暂不支持")
	}
	return nil
}

func fileBadRequest(message string) error {
	return fbaerrors.New(http.StatusBadRequest, http.StatusBadRequest, message, nil)
}
