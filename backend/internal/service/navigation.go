package service

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/navi-dock/navi-dock/internal/model"
	"github.com/navi-dock/navi-dock/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrGroupNotFound  = errors.New("group not found")
	ErrGroupNameEmpty = errors.New("group name is required")
	ErrItemNotFound   = errors.New("item not found")
	ErrItemNameEmpty  = errors.New("item name is required")
	ErrItemURLEmpty   = errors.New("at least one url is required")
)

type NavigationService struct {
	repo *repository.NavigationRepo
}

func NewNavigationService(repo *repository.NavigationRepo) *NavigationService {
	return &NavigationService{repo: repo}
}

func (s *NavigationService) ListNavigation() ([]model.NavGroup, error) {
	return s.repo.ListGroupsWithItems()
}

func (s *NavigationService) ListGroups() ([]model.NavGroup, error) {
	groups, err := s.repo.ListGroups()
	if err != nil {
		return nil, err
	}
	for i := range groups {
		count, err := s.repo.CountItems(groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].ItemCount = count
	}
	return groups, nil
}

func (s *NavigationService) CreateGroup(in model.CreateGroupInput) (*model.NavGroup, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrGroupNameEmpty
	}
	return s.repo.CreateGroup(name, strings.TrimSpace(in.Icon))
}

func (s *NavigationService) UpdateGroup(id string, in model.UpdateGroupInput) (*model.NavGroup, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrGroupNameEmpty
	}
	group, err := s.repo.UpdateGroup(id, name, strings.TrimSpace(in.Icon))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGroupNotFound
	}
	return group, err
}

func (s *NavigationService) DeleteGroup(id string) error {
	err := s.repo.DeleteGroup(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrGroupNotFound
	}
	return err
}

func (s *NavigationService) SortGroups(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.repo.SortGroups(ids)
}

func (s *NavigationService) SortItems(groupID string, ids []string) error {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return ErrGroupNotFound
	}
	ok, err := s.repo.GroupExists(groupID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrGroupNotFound
	}
	if len(ids) == 0 {
		return nil
	}
	return s.repo.SortItems(groupID, ids)
}

func (s *NavigationService) CreateItem(in model.UpsertItemInput) (*model.NavItem, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrItemNameEmpty
	}
	external := strings.TrimSpace(in.ExternalURL)
	internal := strings.TrimSpace(in.InternalURL)
	if external == "" && internal == "" {
		return nil, ErrItemURLEmpty
	}
	groupID := strings.TrimSpace(in.GroupID)
	if groupID == "" {
		return nil, ErrGroupNotFound
	}
	ok, err := s.repo.GroupExists(groupID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrGroupNotFound
	}

	iconType := strings.TrimSpace(in.IconType)
	if iconType == "" {
		iconType = "iconify"
	}
	if iconType != "iconify" && iconType != "text" && iconType != "image" {
		iconType = "iconify"
	}

	openType := strings.TrimSpace(in.OpenType)
	if openType != "_self" {
		openType = "_blank"
	}

	item := &model.NavItem{
		GroupID:     groupID,
		Name:        name,
		Description: strings.TrimSpace(in.Description),
		IconType:    iconType,
		Icon:        strings.TrimSpace(in.Icon),
		ExternalURL: external,
		InternalURL: internal,
		OpenType:    openType,
		Status:      "online",
	}
	if err := s.repo.CreateItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *NavigationService) UpdateItem(id string, in model.UpsertItemInput) (*model.NavItem, error) {
	item, err := s.repo.GetItem(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrItemNameEmpty
	}
	external := strings.TrimSpace(in.ExternalURL)
	internal := strings.TrimSpace(in.InternalURL)
	if external == "" && internal == "" {
		return nil, ErrItemURLEmpty
	}

	groupID := strings.TrimSpace(in.GroupID)
	if groupID == "" {
		groupID = item.GroupID
	}
	if groupID != item.GroupID {
		ok, err := s.repo.GroupExists(groupID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrGroupNotFound
		}
		item.GroupID = groupID
	}

	iconType := strings.TrimSpace(in.IconType)
	if iconType == "" {
		iconType = "iconify"
	}
	if iconType != "iconify" && iconType != "text" && iconType != "image" {
		iconType = "iconify"
	}

	openType := strings.TrimSpace(in.OpenType)
	if openType != "_self" {
		openType = "_blank"
	}

	item.Name = name
	item.Description = strings.TrimSpace(in.Description)
	item.IconType = iconType
	item.Icon = strings.TrimSpace(in.Icon)
	item.ExternalURL = external
	item.InternalURL = internal
	item.OpenType = openType

	if err := s.repo.UpdateItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *NavigationService) DeleteItem(id string) error {
	err := s.repo.DeleteItem(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrItemNotFound
	}
	return err
}

func (s *NavigationService) ExportNavigation() (*model.NaviDockExport, error) {
	groups, err := s.repo.ListGroupsWithItems()
	if err != nil {
		return nil, err
	}
	out := &model.NaviDockExport{
		App:        "naviDock",
		Version:    1,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Groups:     make([]model.NaviDockExportGroup, 0, len(groups)),
	}
	for _, g := range groups {
		eg := model.NaviDockExportGroup{
			ID:    g.ID,
			Name:  g.Name,
			Icon:  g.Icon,
			Sort:  g.Sort,
			Items: make([]model.NaviDockExportItem, 0, len(g.Items)),
		}
		for _, it := range g.Items {
			eg.Items = append(eg.Items, model.NaviDockExportItem{
				ID:          it.ID,
				Name:        it.Name,
				Description: it.Description,
				IconType:    it.IconType,
				Icon:        it.Icon,
				ExternalURL: it.ExternalURL,
				InternalURL: it.InternalURL,
				OpenType:    it.OpenType,
				Sort:        it.Sort,
				Status:      it.Status,
			})
		}
		out.Groups = append(out.Groups, eg)
	}
	return out, nil
}

func (s *NavigationService) ImportNaviDock(payload model.NaviDockExport) (groupCount, itemCount int, err error) {
	now := time.Now()
	groups := make([]model.NavGroup, 0, len(payload.Groups))
	items := make([]model.NavItem, 0)

	for _, g := range payload.Groups {
		name := strings.TrimSpace(g.Name)
		if name == "" {
			continue
		}
		groupID := strings.TrimSpace(g.ID)
		if groupID == "" {
			groupID = repository.NewID()
		}
		groups = append(groups, model.NavGroup{
			ID:        groupID,
			Name:      name,
			Icon:      strings.TrimSpace(g.Icon),
			Sort:      g.Sort,
			CreatedAt: now,
			UpdatedAt: now,
		})

		for _, c := range g.Items {
			title := strings.TrimSpace(c.Name)
			if title == "" {
				continue
			}
			external := strings.TrimSpace(c.ExternalURL)
			internal := strings.TrimSpace(c.InternalURL)
			if external == "" && internal == "" {
				continue
			}
			itemID := strings.TrimSpace(c.ID)
			if itemID == "" {
				itemID = repository.NewID()
			}
			iconType := strings.TrimSpace(c.IconType)
			if iconType != "iconify" && iconType != "text" && iconType != "image" {
				iconType = "iconify"
			}
			openType := strings.TrimSpace(c.OpenType)
			if openType != "_self" {
				openType = "_blank"
			}
			status := strings.TrimSpace(c.Status)
			if status == "" {
				status = "online"
			}
			items = append(items, model.NavItem{
				ID:          itemID,
				GroupID:     groupID,
				Name:        title,
				Description: strings.TrimSpace(c.Description),
				IconType:    iconType,
				Icon:        strings.TrimSpace(c.Icon),
				ExternalURL: external,
				InternalURL: internal,
				OpenType:    openType,
				Sort:        c.Sort,
				Status:      status,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}
	}

	if len(groups) == 0 {
		return 0, 0, errors.New("no groups in export")
	}
	if err := s.repo.ReplaceAll(groups, items); err != nil {
		return 0, 0, err
	}
	return len(groups), len(items), nil
}

func (s *NavigationService) ImportSunPanel(payload model.SunPanelExport) (groupCount, itemCount int, err error) {
	groupsIn := append([]model.SunPanelGroup(nil), payload.Icons...)
	sort.SliceStable(groupsIn, func(i, j int) bool {
		if groupsIn[i].Sort == groupsIn[j].Sort {
			return groupsIn[i].Title < groupsIn[j].Title
		}
		return groupsIn[i].Sort < groupsIn[j].Sort
	})

	now := time.Now()
	groups := make([]model.NavGroup, 0, len(groupsIn))
	items := make([]model.NavItem, 0)

	for gi, g := range groupsIn {
		name := strings.TrimSpace(g.Title)
		if name == "" {
			continue
		}
		groupID := repository.NewID()
		groups = append(groups, model.NavGroup{
			ID:        groupID,
			Name:      name,
			Sort:      gi,
			CreatedAt: now,
			UpdatedAt: now,
		})

		children := append([]model.SunPanelChild(nil), g.Children...)
		sort.SliceStable(children, func(i, j int) bool {
			if children[i].Sort == children[j].Sort {
				return children[i].Title < children[j].Title
			}
			return children[i].Sort < children[j].Sort
		})

		for ii, c := range children {
			title := strings.TrimSpace(c.Title)
			if title == "" {
				continue
			}
			external := strings.TrimSpace(c.URL)
			internal := strings.TrimSpace(c.LanURL)
			if external == "" && internal == "" {
				continue
			}
			icon := strings.TrimSpace(c.Icon.Text)
			if icon == "" {
				icon = strings.TrimSpace(c.Icon.Src)
			}
			openType := "_blank"
			if c.OpenMethod == 1 {
				openType = "_self"
			}
			items = append(items, model.NavItem{
				ID:          repository.NewID(),
				GroupID:     groupID,
				Name:        title,
				Description: strings.TrimSpace(c.Description),
				IconType:    "iconify",
				Icon:        icon,
				ExternalURL: external,
				InternalURL: internal,
				OpenType:    openType,
				Sort:        ii,
				Status:      "online",
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}
	}

	if err := s.repo.ReplaceAll(groups, items); err != nil {
		return 0, 0, err
	}
	return len(groups), len(items), nil
}
