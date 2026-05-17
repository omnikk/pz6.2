package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Done        bool   `json:"done"`
}

// Repo вЂ” СѓР·РєРёР№ РёРЅС‚РµСЂС„РµР№СЃ С…СЂР°РЅРёР»РёС‰Р°, РєРѕС‚РѕСЂС‹Р№ РЅСѓР¶РµРЅ СЃРµСЂРІРёСЃРЅРѕРјСѓ СЃР»РѕСЋ.
// РћР±СЉСЏРІР»РµРЅ Р—Р”Р•РЎР¬ (Р° РЅРµ РёРјРїРѕСЂС‚РёСЂРѕРІР°РЅ РёР· repository), С‡С‚РѕР±С‹ РёР·Р±РµР¶Р°С‚СЊ
// С†РёРєР»РёС‡РµСЃРєРѕР№ Р·Р°РІРёСЃРёРјРѕСЃС‚Рё: repository СѓР¶Рµ РёРјРїРѕСЂС‚РёСЂСѓРµС‚ service.Task.
type Repo interface {
	Create(t Task) (Task, error)
	List() ([]Task, error)
	Get(id string) (Task, bool, error)
	Update(id string, title *string, done *bool) (Task, bool, error)
	Delete(id string) (bool, error)
	Search(title string) ([]Task, error)
	SearchVulnerable(title string) ([]Task, error)
}

type TaskService struct {
	repo Repo
}

func New(repo Repo) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) generateID() string {
	// uuid С‡РёС‚Р°РµРјС‹Р№ вЂ” РЅРѕ СѓРєРѕСЂРѕС‚РёРј РґРѕ С„РѕСЂРјР°С‚Р° t_xxxxxxxx_timestamp,
	// С‡С‚РѕР±С‹ РЅРµ РїР»РѕРґРёС‚СЊ РґР»РёРЅРЅС‹С… СЃС‚СЂРѕРє РІ Р»РѕРіР°С…
	return fmt.Sprintf("t_%s_%d", uuid.NewString()[:8], time.Now().Unix())
}

func (s *TaskService) Create(title, description, dueDate string) (Task, error) {
	t := Task{
		ID:          s.generateID(),
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Done:        false,
	}
	return s.repo.Create(t)
}

func (s *TaskService) List() ([]Task, error) {
	return s.repo.List()
}

func (s *TaskService) Get(id string) (Task, bool, error) {
	return s.repo.Get(id)
}

func (s *TaskService) Update(id string, title *string, done *bool) (Task, bool, error) {
	return s.repo.Update(id, title, done)
}

func (s *TaskService) Delete(id string) (bool, error) {
	return s.repo.Delete(id)
}

func (s *TaskService) Search(title string) ([]Task, error) {
	return s.repo.Search(title)
}

func (s *TaskService) SearchVulnerable(title string) ([]Task, error) {
	return s.repo.SearchVulnerable(title)
}
