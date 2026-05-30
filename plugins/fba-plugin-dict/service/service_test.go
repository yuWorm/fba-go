package service_test

import (
	"context"
	"testing"

	"github.com/yuWorm/fba-plugin-dict/dto"
	"github.com/yuWorm/fba-plugin-dict/model"
	"github.com/yuWorm/fba-plugin-dict/repo"
	"github.com/yuWorm/fba-plugin-dict/service"
)

func TestServiceInvalidatesDictCacheOnWrites(t *testing.T) {
	invalidator := &fakeInvalidator{}
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), invalidator)

	ctx := context.Background()
	if err := svc.CreateType(ctx, dto.DictTypeParam{Name: "Fixture", Code: "fixture"}); err != nil {
		t.Fatalf("CreateType() error = %v", err)
	}
	if err := svc.UpdateType(ctx, 1, dto.DictTypeParam{Name: "通用状态", Code: "sys_status"}); err != nil {
		t.Fatalf("UpdateType() error = %v", err)
	}
	if err := svc.CreateData(ctx, dto.DictDataParam{TypeID: 1, Label: "Fixture", Value: "fixture", Sort: 1, Status: 1}); err != nil {
		t.Fatalf("CreateData() error = %v", err)
	}
	if err := svc.UpdateData(ctx, 1, dto.DictDataParam{TypeID: 1, Label: "停用", Value: "0", Sort: 1, Status: 1}); err != nil {
		t.Fatalf("UpdateData() error = %v", err)
	}
	if err := svc.DeleteData(ctx, []int{1}); err != nil {
		t.Fatalf("DeleteData() error = %v", err)
	}
	if err := svc.DeleteTypes(ctx, []int{1}); err != nil {
		t.Fatalf("DeleteTypes() error = %v", err)
	}

	if invalidator.calls != 6 {
		t.Fatalf("InvalidateDict() calls = %d, want 6", invalidator.calls)
	}
}

func TestServiceFiltersDictDataByTypeCode(t *testing.T) {
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), service.NoopInvalidator{})

	items, err := svc.GetDataByTypeCode(context.Background(), "sys_status")
	if err != nil {
		t.Fatalf("GetDataByTypeCode() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	for _, item := range items {
		if item.TypeCode != "sys_status" {
			t.Fatalf("item.TypeCode = %q, want sys_status", item.TypeCode)
		}
	}
}

func TestServiceReturnsPageData(t *testing.T) {
	svc := service.New(repo.NewMemoryRepository(repo.SeedData()), service.NoopInvalidator{})

	page, err := svc.ListTypes(context.Background(), repo.DictTypeFilter{}, 1, 20, "/api/v1/dict-types")
	if err != nil {
		t.Fatalf("ListTypes() error = %v", err)
	}
	if page.Total != int64(len(model.SeedDictTypes())) {
		t.Fatalf("Total = %d, want %d", page.Total, len(model.SeedDictTypes()))
	}
	if page.Links.Self != "/api/v1/dict-types?page=1&size=20" {
		t.Fatalf("Self link = %q", page.Links.Self)
	}
}

type fakeInvalidator struct {
	calls int
}

func (f *fakeInvalidator) InvalidateDict(context.Context) error {
	f.calls++
	return nil
}
