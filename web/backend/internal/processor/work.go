package processor

import (
	"backend/constant"
	"backend/internal/context"
	"backend/model"
	"errors"
	"hash/fnv"
	"math"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"
)

// workEntryRule lists the optional WorkEntry fields a target requires.
type workEntryRule struct {
	category bool
	hours    bool
}

var (
	workRecordRule = workEntryRule{category: true, hours: true}
	todoRule       = workEntryRule{category: false, hours: false}
)

// workOptionStore adapts the category and project storage to shared logic.
type workOptionStore struct {
	name string
	// categories carry a color, projects do not
	colored  bool
	list     func() ([]*model.WorkOption, error)
	get      func(id string) (*model.WorkOption, error)
	create   func(option *model.WorkOption) error
	update   func(option *model.WorkOption) error
	delete   func(id string) (int, error)
	notFound error
}

func (p *Processor) categoryStore() *workOptionStore {
	return &workOptionStore{
		name:     "Category",
		colored:  true,
		list:     p.ListCategories,
		get:      p.GetCategory,
		create:   p.CreateCategory,
		update:   p.UpdateCategory,
		delete:   p.DeleteCategory,
		notFound: context.ErrCategoryNotFound,
	}
}

func (p *Processor) projectStore() *workOptionStore {
	return &workOptionStore{
		name:     "Project",
		list:     p.ListProjects,
		get:      p.GetProject,
		create:   p.CreateProject,
		update:   p.UpdateProject,
		delete:   p.DeleteProject,
		notFound: context.ErrProjectNotFound,
	}
}

func (p *Processor) GetWorkOptions() (*model.ResponseWorkOptions, *model.ErrorDetail) {
	categories, err := p.ListCategories()
	if err != nil {
		p.ProcLog.Errorf("Failed to list categories: %v", err)
		return nil, errInternal("Failed to list categories")
	}
	projects, err := p.ListProjects()
	if err != nil {
		p.ProcLog.Errorf("Failed to list projects: %v", err)
		return nil, errInternal("Failed to list projects")
	}
	setting, err := p.GetWorkSetting()
	if err != nil {
		p.ProcLog.Errorf("Failed to get work setting: %v", err)
		return nil, errInternal("Failed to get work setting")
	}

	for _, category := range categories {
		category.Color = categoryColor(category)
	}

	return &model.ResponseWorkOptions{
		Message:      "Get work options successful",
		Categories:   derefAll(categories),
		Projects:     derefAll(projects),
		AllowViewAll: setting.AllowViewAll,
	}, nil
}

func (p *Processor) CreateWorkCategory(req *model.RequestCreateWorkOption) (*model.ResponseWorkOption, *model.ErrorDetail) {
	return p.createWorkOption(p.categoryStore(), req)
}

func (p *Processor) UpdateWorkCategory(id string, req *model.RequestUpdateWorkOption) (*model.ResponseWorkOption, *model.ErrorDetail) {
	return p.updateWorkOption(p.categoryStore(), id, req)
}

func (p *Processor) CreateWorkProject(req *model.RequestCreateWorkOption) (*model.ResponseWorkOption, *model.ErrorDetail) {
	return p.createWorkOption(p.projectStore(), req)
}

func (p *Processor) UpdateWorkProject(id string, req *model.RequestUpdateWorkOption) (*model.ResponseWorkOption, *model.ErrorDetail) {
	return p.updateWorkOption(p.projectStore(), id, req)
}

// DeleteWorkCategory deletes the category; entries using it become
// uncategorized.
func (p *Processor) DeleteWorkCategory(id string) (*model.ResponseDeleteWorkOption, *model.ErrorDetail) {
	return p.deleteWorkOption(p.categoryStore(), id)
}

// DeleteWorkProject deletes the project; entries using it lose their project.
func (p *Processor) DeleteWorkProject(id string) (*model.ResponseDeleteWorkOption, *model.ErrorDetail) {
	return p.deleteWorkOption(p.projectStore(), id)
}

func (p *Processor) createWorkOption(store *workOptionStore, req *model.RequestCreateWorkOption) (*model.ResponseWorkOption, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing create %s: %s", store.name, req.Name)

	name, errDetail := p.checkWorkOptionName(store, "", req.Name)
	if errDetail != nil {
		return nil, errDetail
	}

	option := &model.WorkOption{
		Name:   name,
		Active: true,
	}
	if store.colored {
		options, err := store.list()
		if err != nil {
			p.ProcLog.Errorf("Failed to list %s: %v", store.name, err)
			return nil, errInternal("Failed to list " + strings.ToLower(store.name))
		}
		option.Color = leastUsedColor(options)
	}
	if err := store.create(option); err != nil {
		p.ProcLog.Errorf("Failed to create %s %s: %v", store.name, name, err)
		return nil, errInternal("Failed to create " + strings.ToLower(store.name))
	}

	return &model.ResponseWorkOption{
		Message: "Create " + strings.ToLower(store.name) + " successful",
		Option:  option,
	}, nil
}

func (p *Processor) updateWorkOption(store *workOptionStore, id string, req *model.RequestUpdateWorkOption) (*model.ResponseWorkOption, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing update %s: %s", store.name, id)

	option, err := store.get(id)
	if errors.Is(err, store.notFound) {
		return nil, errNotFound(store.name + " not found")
	}
	if err != nil {
		p.ProcLog.Errorf("Failed to get %s %s: %v", store.name, id, err)
		return nil, errInternal("Failed to get " + strings.ToLower(store.name))
	}

	if req.Name != nil {
		name, errDetail := p.checkWorkOptionName(store, id, *req.Name)
		if errDetail != nil {
			return nil, errDetail
		}
		option.Name = name
	}
	if req.Active != nil {
		option.Active = *req.Active
	}
	if req.Color != nil {
		if !store.colored {
			return nil, errBadRequest(store.name + " has no color")
		}
		if !slices.Contains(constant.WORK_CATEGORY_COLORS, *req.Color) {
			return nil, errBadRequest("Unknown color")
		}
		option.Color = *req.Color
	} else if store.colored {
		option.Color = categoryColor(option)
	}

	if err := store.update(option); err != nil {
		if errors.Is(err, store.notFound) {
			return nil, errNotFound(store.name + " not found")
		}
		p.ProcLog.Errorf("Failed to update %s %s: %v", store.name, id, err)
		return nil, errInternal("Failed to update " + strings.ToLower(store.name))
	}

	return &model.ResponseWorkOption{
		Message: "Update " + strings.ToLower(store.name) + " successful",
		Option:  option,
	}, nil
}

func (p *Processor) deleteWorkOption(store *workOptionStore, id string) (*model.ResponseDeleteWorkOption, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing delete %s: %s", store.name, id)

	cleared, err := store.delete(id)
	if errors.Is(err, store.notFound) {
		return nil, errNotFound(store.name + " not found")
	}
	if err != nil {
		p.ProcLog.Errorf("Failed to delete %s %s: %v", store.name, id, err)
		return nil, errInternal("Failed to delete " + strings.ToLower(store.name))
	}
	p.ProcLog.Infof("%s %s deleted, cleared from %d entries", store.name, id, cleared)

	return &model.ResponseDeleteWorkOption{
		Message: "Delete " + strings.ToLower(store.name) + " successful",
		Cleared: cleared,
	}, nil
}

// checkWorkOptionName trims name and rejects blanks and case-insensitive
// duplicates among the other options (selfID is skipped).
func (p *Processor) checkWorkOptionName(store *workOptionStore, selfID, name string) (string, *model.ErrorDetail) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errBadRequest("Name must not be blank")
	}

	options, err := store.list()
	if err != nil {
		p.ProcLog.Errorf("Failed to list %s: %v", store.name, err)
		return "", errInternal("Failed to list " + strings.ToLower(store.name))
	}
	for _, option := range options {
		if option.ID != selfID && strings.EqualFold(option.Name, name) {
			return "", &model.ErrorDetail{
				HttpStatus: http.StatusConflict,
				Detail:     store.name + " already exists",
			}
		}
	}
	return name, nil
}

func (p *Processor) SaveWorkSetting(req *model.RequestUpdateWorkSetting) (*model.ResponseWorkSetting, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing update work setting: allowViewAll=%t", *req.AllowViewAll)

	setting := &model.WorkSetting{
		AllowViewAll: *req.AllowViewAll,
	}
	if err := p.UpdateWorkSetting(setting); err != nil {
		p.ProcLog.Errorf("Failed to update work setting: %v", err)
		return nil, errInternal("Failed to update work setting")
	}

	return &model.ResponseWorkSetting{
		Message:      "Update work setting successful",
		AllowViewAll: setting.AllowViewAll,
	}, nil
}

// my work records

func (p *Processor) ListMyWorkRecords(acc *model.Account, req *model.RequestListMyWorkRecords) (*model.ResponseWorkRecordList, *model.ErrorDetail) {
	if errDetail := checkWorkRange(req.From, req.To); errDetail != nil {
		return nil, errDetail
	}

	return p.listWorkRecords(&model.WorkRecordFilter{
		Account: acc.Account,
		From:    req.From,
		To:      req.To,
	})
}

func (p *Processor) CreateMyWorkRecord(acc *model.Account, req *model.RequestSaveWorkEntry) (*model.ResponseWorkRecord, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing create work record for %s", acc.Account)

	entry, errDetail := p.checkWorkEntry(req, nil, workRecordRule)
	if errDetail != nil {
		return nil, errDetail
	}

	record := &model.WorkRecord{
		Account:   acc.Account,
		WorkEntry: *entry,
		CreatedAt: time.Now(),
	}
	if err := p.CreateWorkRecord(record); err != nil {
		p.ProcLog.Errorf("Failed to create work record for %s: %v", acc.Account, err)
		return nil, errInternal("Failed to create work record")
	}

	return &model.ResponseWorkRecord{
		Message: "Create work record successful",
		Record:  record,
	}, nil
}

func (p *Processor) UpdateMyWorkRecord(acc *model.Account, id string, req *model.RequestSaveWorkEntry) (*model.ResponseWorkRecord, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing update work record %s for %s", id, acc.Account)

	record, errDetail := p.getMyWorkRecord(acc, id)
	if errDetail != nil {
		return nil, errDetail
	}

	entry, errDetail := p.checkWorkEntry(req, &record.WorkEntry, workRecordRule)
	if errDetail != nil {
		return nil, errDetail
	}
	record.WorkEntry = *entry

	if err := p.UpdateWorkRecord(record); err != nil {
		if errors.Is(err, context.ErrWorkRecordNotFound) {
			return nil, errNotFound("Work record not found")
		}
		p.ProcLog.Errorf("Failed to update work record %s: %v", id, err)
		return nil, errInternal("Failed to update work record")
	}

	return &model.ResponseWorkRecord{
		Message: "Update work record successful",
		Record:  record,
	}, nil
}

func (p *Processor) DeleteMyWorkRecord(acc *model.Account, id string) (*model.ResponseDeleteWorkRecord, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing delete work record %s for %s", id, acc.Account)

	if _, errDetail := p.getMyWorkRecord(acc, id); errDetail != nil {
		return nil, errDetail
	}

	if err := p.DeleteWorkRecord(id); err != nil {
		if errors.Is(err, context.ErrWorkRecordNotFound) {
			return nil, errNotFound("Work record not found")
		}
		p.ProcLog.Errorf("Failed to delete work record %s: %v", id, err)
		return nil, errInternal("Failed to delete work record")
	}

	return &model.ResponseDeleteWorkRecord{
		Message: "Delete work record successful",
	}, nil
}

// getMyWorkRecord answers 404 for other people's records too, so their IDs
// are not revealed.
func (p *Processor) getMyWorkRecord(acc *model.Account, id string) (*model.WorkRecord, *model.ErrorDetail) {
	record, err := p.GetWorkRecord(id)
	if errors.Is(err, context.ErrWorkRecordNotFound) || (err == nil && record.Account != acc.Account) {
		return nil, errNotFound("Work record not found")
	}
	if err != nil {
		p.ProcLog.Errorf("Failed to get work record %s: %v", id, err)
		return nil, errInternal("Failed to get work record")
	}
	return record, nil
}

// todos

func (p *Processor) ListMyTodos(acc *model.Account) (*model.ResponseTodos, *model.ErrorDetail) {
	todos, err := p.ListTodos(acc.Account)
	if err != nil {
		p.ProcLog.Errorf("Failed to list todos for %s: %v", acc.Account, err)
		return nil, errInternal("Failed to list todos")
	}

	sort.SliceStable(todos, func(i, j int) bool {
		if todos[i].Date != todos[j].Date {
			return todos[i].Date < todos[j].Date
		}
		return todos[i].ID < todos[j].ID
	})

	return &model.ResponseTodos{
		Message: "List todos successful",
		Todos:   derefAll(todos),
	}, nil
}

func (p *Processor) CreateMyTodo(acc *model.Account, req *model.RequestSaveWorkEntry) (*model.ResponseTodo, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing create todo for %s", acc.Account)

	entry, errDetail := p.checkWorkEntry(req, nil, todoRule)
	if errDetail != nil {
		return nil, errDetail
	}

	todo := &model.Todo{
		Account:   acc.Account,
		WorkEntry: *entry,
		CreatedAt: time.Now(),
	}
	if err := p.CreateTodo(todo); err != nil {
		p.ProcLog.Errorf("Failed to create todo for %s: %v", acc.Account, err)
		return nil, errInternal("Failed to create todo")
	}

	return &model.ResponseTodo{
		Message: "Create todo successful",
		Todo:    todo,
	}, nil
}

func (p *Processor) UpdateMyTodo(acc *model.Account, id string, req *model.RequestSaveWorkEntry) (*model.ResponseTodo, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing update todo %s for %s", id, acc.Account)

	todo, errDetail := p.getMyTodo(acc, id)
	if errDetail != nil {
		return nil, errDetail
	}

	entry, errDetail := p.checkWorkEntry(req, &todo.WorkEntry, todoRule)
	if errDetail != nil {
		return nil, errDetail
	}
	todo.WorkEntry = *entry

	if err := p.UpdateTodo(todo); err != nil {
		if errors.Is(err, context.ErrTodoNotFound) {
			return nil, errNotFound("Todo not found")
		}
		p.ProcLog.Errorf("Failed to update todo %s: %v", id, err)
		return nil, errInternal("Failed to update todo")
	}

	return &model.ResponseTodo{
		Message: "Update todo successful",
		Todo:    todo,
	}, nil
}

func (p *Processor) DeleteMyTodo(acc *model.Account, id string) (*model.ResponseDeleteTodo, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing delete todo %s for %s", id, acc.Account)

	if _, errDetail := p.getMyTodo(acc, id); errDetail != nil {
		return nil, errDetail
	}

	if err := p.DeleteTodo(id); err != nil {
		if errors.Is(err, context.ErrTodoNotFound) {
			return nil, errNotFound("Todo not found")
		}
		p.ProcLog.Errorf("Failed to delete todo %s: %v", id, err)
		return nil, errInternal("Failed to delete todo")
	}

	return &model.ResponseDeleteTodo{
		Message: "Delete todo successful",
	}, nil
}

// CompleteMyTodo turns the todo into a work record dated req.Date (the
// client's today, so the user's time zone decides the day). Category, hours
// and project given in req fill or override the todo's; the result must pass
// the work record rules.
func (p *Processor) CompleteMyTodo(acc *model.Account, id string, req *model.RequestCompleteTodo) (*model.ResponseWorkRecord, *model.ErrorDetail) {
	p.ProcLog.Debugf("Processing complete todo %s for %s", id, acc.Account)

	if !isWorkDate(req.Date) {
		return nil, errBadRequest("Date must be YYYY-MM-DD")
	}

	todo, errDetail := p.getMyTodo(acc, id)
	if errDetail != nil {
		return nil, errDetail
	}

	fill := &model.RequestSaveWorkEntry{
		Date:        req.Date,
		CategoryID:  todo.CategoryID,
		Description: todo.Description,
		Hours:       todo.Hours,
		ProjectID:   todo.ProjectID,
	}
	if req.CategoryID != "" {
		fill.CategoryID = req.CategoryID
	}
	if req.Hours != 0 {
		fill.Hours = req.Hours
	}
	if req.ProjectID != "" {
		fill.ProjectID = req.ProjectID
	}
	entry, errDetail := p.checkWorkEntry(fill, &todo.WorkEntry, workRecordRule)
	if errDetail != nil {
		return nil, errDetail
	}

	record := &model.WorkRecord{
		Account:   acc.Account,
		WorkEntry: *entry,
		CreatedAt: time.Now(),
	}
	if err := p.CompleteTodo(id, record); err != nil {
		if errors.Is(err, context.ErrTodoNotFound) {
			return nil, errNotFound("Todo not found")
		}
		p.ProcLog.Errorf("Failed to complete todo %s: %v", id, err)
		return nil, errInternal("Failed to complete todo")
	}

	return &model.ResponseWorkRecord{
		Message: "Complete todo successful",
		Record:  record,
	}, nil
}

func (p *Processor) getMyTodo(acc *model.Account, id string) (*model.Todo, *model.ErrorDetail) {
	todo, err := p.GetTodo(id)
	if errors.Is(err, context.ErrTodoNotFound) || (err == nil && todo.Account != acc.Account) {
		return nil, errNotFound("Todo not found")
	}
	if err != nil {
		p.ProcLog.Errorf("Failed to get todo %s: %v", id, err)
		return nil, errInternal("Failed to get todo")
	}
	return todo, nil
}

// everyone's records

func (p *Processor) ListAllWorkRecords(acc *model.Account, req *model.RequestListWorkRecords) (*model.ResponseWorkRecordList, *model.ErrorDetail) {
	if errDetail := p.checkViewAll(acc); errDetail != nil {
		return nil, errDetail
	}
	if errDetail := checkWorkRange(req.From, req.To); errDetail != nil {
		return nil, errDetail
	}

	account := ""
	if req.Account != "" {
		account = normalizeAccount(req.Account)
	}
	return p.listWorkRecords(&model.WorkRecordFilter{
		Account:    account,
		From:       req.From,
		To:         req.To,
		CategoryID: req.CategoryID,
		ProjectID:  req.ProjectID,
	})
}

func (p *Processor) ListWorkMembers(acc *model.Account) (*model.ResponseWorkMembers, *model.ErrorDetail) {
	if errDetail := p.checkViewAll(acc); errDetail != nil {
		return nil, errDetail
	}

	accounts, err := p.ListAccounts()
	if err != nil {
		p.ProcLog.Errorf("Failed to list accounts: %v", err)
		return nil, errInternal("Failed to list members")
	}

	members := make([]model.WorkMember, 0, len(accounts))
	for _, a := range accounts {
		members = append(members, model.WorkMember{
			Account: a.Account,
			Name:    a.Name,
		})
	}

	return &model.ResponseWorkMembers{
		Message: "List members successful",
		Members: members,
	}, nil
}

// checkViewAll allows admins always, and everyone else only when the admin
// turned on AllowViewAll.
func (p *Processor) checkViewAll(acc *model.Account) *model.ErrorDetail {
	if acc.Role == constant.ROLE_ADMIN {
		return nil
	}

	setting, err := p.GetWorkSetting()
	if err != nil {
		p.ProcLog.Errorf("Failed to get work setting: %v", err)
		return errInternal("Failed to get work setting")
	}
	if !setting.AllowViewAll {
		return errForbidden("Viewing everyone's work records is not allowed")
	}
	return nil
}

// listWorkRecords returns every matching record, newest date first and,
// within a day, the latest created first.
func (p *Processor) listWorkRecords(filter *model.WorkRecordFilter) (*model.ResponseWorkRecordList, *model.ErrorDetail) {
	records, err := p.ListWorkRecords(filter)
	if err != nil {
		p.ProcLog.Errorf("Failed to list work records: %v", err)
		return nil, errInternal("Failed to list work records")
	}

	sort.SliceStable(records, func(i, j int) bool {
		a, b := records[i], records[j]
		if a.Date != b.Date {
			return a.Date > b.Date
		}
		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.After(b.CreatedAt)
		}
		return a.ID > b.ID
	})

	totalHours := 0.0
	for _, record := range records {
		totalHours += record.Hours
	}

	return &model.ResponseWorkRecordList{
		Message:    "List work records successful",
		Records:    derefAll(records),
		Total:      len(records),
		TotalHours: math.Round(totalHours*100) / 100,
	}, nil
}

// checkWorkRange requires an inclusive from..to range of YYYY-MM-DD dates.
func checkWorkRange(from, to string) *model.ErrorDetail {
	if !isWorkDate(from) || !isWorkDate(to) {
		return errBadRequest("From and to must be YYYY-MM-DD dates")
	}
	if from > to {
		return errBadRequest("From must not be after to")
	}
	return nil
}

// checkWorkEntry validates and normalizes an entry against rule. prev is the
// stored entry when updating: an option that was deactivated after it was
// chosen is still accepted as long as it is not changed.
func (p *Processor) checkWorkEntry(req *model.RequestSaveWorkEntry, prev *model.WorkEntry, rule workEntryRule) (*model.WorkEntry, *model.ErrorDetail) {
	if !isWorkDate(req.Date) {
		return nil, errBadRequest("Date must be YYYY-MM-DD")
	}
	if req.Hours < 0 || req.Hours > constant.WORK_MAX_HOURS || (rule.hours && req.Hours == 0) {
		return nil, errBadRequest("Hours must be greater than 0 and at most 24")
	}
	if steps := req.Hours / constant.WORK_HOURS_STEP; steps != math.Trunc(steps) {
		return nil, errBadRequest("Hours must be a multiple of 0.5")
	}
	if rule.category && req.CategoryID == "" {
		return nil, errBadRequest("Category is required")
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		return nil, errBadRequest("Description must not be blank")
	}

	var prevCategory, prevProject string
	if prev != nil {
		prevCategory, prevProject = prev.CategoryID, prev.ProjectID
	}
	if req.CategoryID != "" {
		if errDetail := p.checkWorkOption(p.categoryStore(), req.CategoryID, prevCategory); errDetail != nil {
			return nil, errDetail
		}
	}
	if req.ProjectID != "" {
		if errDetail := p.checkWorkOption(p.projectStore(), req.ProjectID, prevProject); errDetail != nil {
			return nil, errDetail
		}
	}

	return &model.WorkEntry{
		Date:        req.Date,
		CategoryID:  req.CategoryID,
		Description: description,
		Hours:       req.Hours,
		ProjectID:   req.ProjectID,
	}, nil
}

func (p *Processor) checkWorkOption(store *workOptionStore, id, prevID string) *model.ErrorDetail {
	option, err := store.get(id)
	if errors.Is(err, store.notFound) {
		return errBadRequest("Unknown " + strings.ToLower(store.name))
	}
	if err != nil {
		p.ProcLog.Errorf("Failed to get %s %s: %v", store.name, id, err)
		return errInternal("Failed to get " + strings.ToLower(store.name))
	}
	if !option.Active && id != prevID {
		return errBadRequest(store.name + " is inactive")
	}
	return nil
}

// categoryColor is the category's color, or a stable one derived from its ID
// for categories stored before categories had colors.
func categoryColor(category *model.WorkOption) string {
	if category.Color != "" {
		return category.Color
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(category.ID))
	return constant.WORK_CATEGORY_COLORS[hash.Sum32()%uint32(len(constant.WORK_CATEGORY_COLORS))]
}

// leastUsedColor picks the palette color used by the fewest categories,
// earliest in the palette on ties.
func leastUsedColor(categories []*model.WorkOption) string {
	used := map[string]int{}
	for _, category := range categories {
		used[categoryColor(category)]++
	}

	best := constant.WORK_CATEGORY_COLORS[0]
	for _, color := range constant.WORK_CATEGORY_COLORS {
		if used[color] < used[best] {
			best = color
		}
	}
	return best
}

func isWorkDate(date string) bool {
	_, err := time.Parse(constant.WORK_DATE_LAYOUT, date)
	return err == nil
}

func derefAll[T any](values []*T) []T {
	result := make([]T, 0, len(values))
	for _, value := range values {
		result = append(result, *value)
	}
	return result
}
