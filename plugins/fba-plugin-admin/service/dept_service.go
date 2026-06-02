package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/yuWorm/fba-go/core/rbac"
	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/model"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type DeptService struct {
	repo repo.Repository
}

func NewDeptService(repository repo.Repository) *DeptService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &DeptService{repo: repository}
}

func (s *DeptService) Get(ctx context.Context, id int) (dto.DeptDetail, error) {
	item, err := s.repo.GetDept(ctx, id)
	if err != nil {
		return dto.DeptDetail{}, err
	}
	return dto.DeptFromModel(item), nil
}

func (s *DeptService) Tree(ctx context.Context, filter repo.DeptFilter) ([]dto.DeptDetail, error) {
	return s.TreeForUser(ctx, filter, nil)
}

func (s *DeptService) TreeForUser(ctx context.Context, filter repo.DeptFilter, user *rbac.CurrentUser) ([]dto.DeptDetail, error) {
	items, err := s.repo.ListDepts(ctx, filter)
	if err != nil {
		return nil, err
	}
	items, err = s.applyDataPermission(ctx, items, user)
	if err != nil {
		return nil, err
	}
	return buildDeptTree(items), nil
}

func (s *DeptService) Create(ctx context.Context, param dto.DeptParam) error {
	return s.repo.CreateDept(ctx, param)
}

func (s *DeptService) Update(ctx context.Context, id int, param dto.DeptParam) error {
	return s.repo.UpdateDept(ctx, id, param)
}

func (s *DeptService) Delete(ctx context.Context, id int) error {
	return s.repo.DeleteDept(ctx, id)
}

func (s *DeptService) applyDataPermission(ctx context.Context, items []model.Dept, user *rbac.CurrentUser) ([]model.Dept, error) {
	if user == nil || user.IsSuperAdmin {
		return items, nil
	}

	var rules []model.DataRule
	hasFilteringRole := false
	for _, role := range user.Roles {
		if !role.Enabled {
			continue
		}
		// Python short-circuits data permission when any enabled role opts out of scope filtering.
		if !role.IsFilterScopes {
			return items, nil
		}
		hasFilteringRole = true
		scopes, err := s.repo.RoleScopes(ctx, int(role.ID))
		if err != nil {
			return nil, err
		}
		for _, scope := range scopes {
			if scope.Status != 1 {
				continue
			}
			_, scopeRules, err := s.repo.DataScopeRules(ctx, scope.ID)
			if err != nil {
				return nil, err
			}
			rules = append(rules, scopeRules...)
		}
	}
	if !hasFilteringRole {
		return []model.Dept{}, nil
	}
	if len(rules) == 0 {
		return []model.Dept{}, nil
	}

	applicable := applicableDeptRules(rules)
	// Python skips rules whose model/column cannot apply; if nothing produces a predicate, it falls back to allow-all.
	if len(applicable) == 0 {
		return items, nil
	}
	filtered := make([]model.Dept, 0, len(items))
	for _, item := range items {
		if matchDeptRules(item, applicable, user) {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func buildDeptTree(items []model.Dept) []dto.DeptDetail {
	byID := make(map[int]model.Dept, len(items))
	childrenByParent := make(map[int][]model.Dept, len(items))
	for _, item := range items {
		byID[item.ID] = item
		if item.ParentID != nil {
			childrenByParent[*item.ParentID] = append(childrenByParent[*item.ParentID], item)
		}
	}

	var buildNode func(model.Dept, map[int]bool) dto.DeptDetail
	buildNode = func(item model.Dept, visiting map[int]bool) dto.DeptDetail {
		detail := dto.DeptFromModel(item)
		children := childrenByParent[item.ID]
		if len(children) == 0 {
			return detail
		}

		detail.Children = make([]dto.DeptDetail, 0, len(children))
		visiting[item.ID] = true
		defer delete(visiting, item.ID)
		for _, child := range children {
			// Build values from the leaves upward so grandchildren are not lost through stale value copies.
			if visiting[child.ID] {
				continue
			}
			detail.Children = append(detail.Children, buildNode(child, visiting))
		}
		return detail
	}

	roots := make([]dto.DeptDetail, 0, len(items))
	for _, item := range items {
		// Keep filtered or orphaned children visible as roots, matching the Python tree helper.
		parentExists := false
		if item.ParentID != nil {
			_, parentExists = byID[*item.ParentID]
		}
		if item.ParentID == nil || !parentExists {
			roots = append(roots, buildNode(item, map[int]bool{}))
		}
	}
	return roots
}

func applicableDeptRules(rules []model.DataRule) []model.DataRule {
	filtered := make([]model.DataRule, 0, len(rules))
	for _, rule := range rules {
		if rule.Model != "dept" && rule.Model != "__ALL__" {
			continue
		}
		if deptRuleColumn(rule.Column) == "" {
			continue
		}
		filtered = append(filtered, rule)
	}
	return filtered
}

func matchDeptRules(item model.Dept, rules []model.DataRule, user *rbac.CurrentUser) bool {
	andRules := make([]model.DataRule, 0)
	orRules := make([]model.DataRule, 0)
	for _, rule := range rules {
		switch rule.Operator {
		case 1:
			orRules = append(orRules, rule)
		default:
			andRules = append(andRules, rule)
		}
	}

	// Python combines all AND rules as one branch, then ORs that branch with any OR rules.
	if len(andRules) > 0 {
		all := true
		for _, rule := range andRules {
			if !matchDeptRule(item, rule, user) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	for _, rule := range orRules {
		if matchDeptRule(item, rule, user) {
			return true
		}
	}
	return false
}

func matchDeptRule(item model.Dept, rule model.DataRule, user *rbac.CurrentUser) bool {
	left, ok := deptRuleValue(item, deptRuleColumn(rule.Column))
	if !ok {
		return false
	}
	values := strings.Split(rule.Value, ",")
	switch rule.Expression {
	case 0:
		return compareDeptRuleValues(left, resolveDeptRuleValue(strings.TrimSpace(rule.Value), user)) == 0
	case 1:
		return compareDeptRuleValues(left, resolveDeptRuleValue(strings.TrimSpace(rule.Value), user)) != 0
	case 2:
		return compareDeptRuleValues(left, resolveDeptRuleValue(strings.TrimSpace(rule.Value), user)) > 0
	case 3:
		return compareDeptRuleValues(left, resolveDeptRuleValue(strings.TrimSpace(rule.Value), user)) >= 0
	case 4:
		return compareDeptRuleValues(left, resolveDeptRuleValue(strings.TrimSpace(rule.Value), user)) < 0
	case 5:
		return compareDeptRuleValues(left, resolveDeptRuleValue(strings.TrimSpace(rule.Value), user)) <= 0
	case 6:
		for _, value := range values {
			if compareDeptRuleValues(left, resolveDeptRuleValue(strings.TrimSpace(value), user)) == 0 {
				return true
			}
		}
		return false
	case 7:
		for _, value := range values {
			if compareDeptRuleValues(left, resolveDeptRuleValue(strings.TrimSpace(value), user)) == 0 {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func deptRuleColumn(column string) string {
	switch column {
	case "__dept_id__":
		column = "dept_id"
	case "__created_by__":
		column = "created_by"
	}
	switch column {
	case "name", "parent_id", "leader", "phone", "email", "status":
		return column
	default:
		return ""
	}
}

func deptRuleValue(item model.Dept, column string) (any, bool) {
	switch column {
	case "name":
		return item.Name, true
	case "parent_id":
		if item.ParentID == nil {
			return nil, true
		}
		return *item.ParentID, true
	case "leader":
		if item.Leader == nil {
			return nil, true
		}
		return *item.Leader, true
	case "phone":
		if item.Phone == nil {
			return nil, true
		}
		return *item.Phone, true
	case "email":
		if item.Email == nil {
			return nil, true
		}
		return *item.Email, true
	case "status":
		return item.Status, true
	default:
		return nil, false
	}
}

func resolveDeptRuleValue(value string, user *rbac.CurrentUser) any {
	switch value {
	case "${user_id}", "{{ user_id }}":
		return int(user.ID)
	case "${dept_id}", "{{ dept_id }}":
		if user.DeptID == nil {
			return nil
		}
		return int(*user.DeptID)
	default:
		return value
	}
}

func compareDeptRuleValues(left any, right any) int {
	leftNumber, leftOK := numericDeptRuleValue(left)
	rightNumber, rightOK := numericDeptRuleValue(right)
	if leftOK && rightOK {
		switch {
		case leftNumber < rightNumber:
			return -1
		case leftNumber > rightNumber:
			return 1
		default:
			return 0
		}
	}
	leftText := stringDeptRuleValue(left)
	rightText := stringDeptRuleValue(right)
	switch {
	case leftText < rightText:
		return -1
	case leftText > rightText:
		return 1
	default:
		return 0
	}
}

func numericDeptRuleValue(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if v == "" {
			return 0, false
		}
		n, err := strconv.ParseFloat(v, 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func stringDeptRuleValue(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return strconv.FormatFloat(mustNumericOrZero(v), 'f', -1, 64)
	}
}

func mustNumericOrZero(value any) float64 {
	n, ok := numericDeptRuleValue(value)
	if !ok {
		return 0
	}
	return n
}
