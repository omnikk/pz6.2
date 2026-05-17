package repository

import "github.com/omnikk/pz6/services/tasks/internal/service"

// TaskRepository вЂ” РёРЅС‚РµСЂС„РµР№СЃ С…СЂР°РЅРёР»РёС‰Р° Р·Р°РґР°С‡.
// РЎРµСЂРІРёСЃРЅС‹Р№ СЃР»РѕР№ СЂР°Р±РѕС‚Р°РµС‚ С‚РѕР»СЊРєРѕ С‡РµСЂРµР· РЅРµРіРѕ, РїРѕСЌС‚РѕРјСѓ СЂРµР°Р»РёР·Р°С†РёСЋ
// (РїР°РјСЏС‚СЊ, Postgres, С‡С‚Рѕ СѓРіРѕРґРЅРѕ) РјРѕР¶РЅРѕ РјРµРЅСЏС‚СЊ Р±РµР· РїСЂР°РІРєРё Р±РёР·РЅРµСЃ-Р»РѕРіРёРєРё.
type TaskRepository interface {
	Create(t service.Task) (service.Task, error)
	List() ([]service.Task, error)
	Get(id string) (service.Task, bool, error)
	Update(id string, title *string, done *bool) (service.Task, bool, error)
	Delete(id string) (bool, error)

	// Search вЂ” Р±РµР·РѕРїР°СЃРЅС‹Р№ РїРѕРёСЃРє С‡РµСЂРµР· РїР°СЂР°РјРµС‚СЂРёР·РѕРІР°РЅРЅС‹Р№ Р·Р°РїСЂРѕСЃ.
	Search(title string) ([]service.Task, error)
	// SearchVulnerable вЂ” РќРђРњР•Р Р•РќРќРћ СѓСЏР·РІРёРјС‹Р№ РјРµС‚РѕРґ РґР»СЏ РґРµРјРѕРЅСЃС‚СЂР°С†РёРё SQL-РёРЅСЉРµРєС†РёРё.
	// Р’ СЂРµР°Р»СЊРЅРѕР№ СЃРёСЃС‚РµРјРµ С‚Р°Рє РґРµР»Р°С‚СЊ РќР•Р›Р¬Р—РЇ.
	SearchVulnerable(title string) ([]service.Task, error)
}
