package admin

import (
	"context"
	"slices"
	"strings"

	"gorm.io/gorm"
)

type Entity struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	ReadOnly bool   `json:"read_only,omitempty"`
}

var entities = []Entity{
	{Name: "users", Label: "Пользователи"}, {Name: "roles", Label: "Роли"},
	{Name: "profiles", Label: "Профили · PostgreSQL"}, {Name: "profile_documents", Label: "Анкеты · MongoDB"},
	{Name: "profile_photos", Label: "Фотографии"}, {Name: "interests", Label: "Интересы"},
	{Name: "user_interests", Label: "Интересы пользователей"}, {Name: "matching_preferences", Label: "Настройки поиска"},
	{Name: "swipes", Label: "Свайпы"}, {Name: "matches", Label: "Мэтчи"}, {Name: "recommendations", Label: "Рекомендации"},
	{Name: "courses", Label: "Предметы"}, {Name: "user_courses", Label: "Выбранные предметы"},
	{Name: "academic_profiles", Label: "Академические профили"}, {Name: "course_choices", Label: "Выбор тиммейтов"},
	{Name: "teams", Label: "Команды"}, {Name: "team_members", Label: "Участники команд"}, {Name: "team_invitations", Label: "Приглашения"},
	{Name: "chats", Label: "Чаты"}, {Name: "messages", Label: "Сообщения"},
	{Name: "notification_types", Label: "Типы уведомлений"}, {Name: "notifications", Label: "Уведомления"},
	{Name: "admin_audit_logs", Label: "Журнал действий", ReadOnly: true},
}

type Field struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Key       bool   `json:"key"`
	Required  bool   `json:"required"`
	Generated bool   `json:"generated"`
	ReadOnly  bool   `json:"read_only"`
}

func schema(ctx context.Context, db *gorm.DB, entity Entity) ([]Field, error) {
	var fields []Field
	err := db.WithContext(ctx).Raw(`SELECT c.column_name AS name, c.data_type AS type,
 EXISTS(SELECT 1 FROM information_schema.table_constraints tc JOIN information_schema.key_column_usage k
 ON k.constraint_schema=tc.constraint_schema AND k.constraint_name=tc.constraint_name AND k.table_name=tc.table_name
 WHERE tc.table_schema=c.table_schema AND tc.table_name=c.table_name AND tc.constraint_type='PRIMARY KEY' AND k.column_name=c.column_name) AS key,
 c.is_nullable='NO' AND c.column_default IS NULL AND c.is_identity='NO' AS required,
 (COALESCE(c.column_default,'') LIKE 'nextval(%' OR c.is_identity='YES' OR c.is_generated<>'NEVER') AS generated
 FROM information_schema.columns c WHERE c.table_schema='public' AND c.table_name=? ORDER BY c.ordinal_position`, entity.Name).Scan(&fields).Error
	if err != nil {
		return nil, err
	}
	fields = slices.DeleteFunc(fields, func(f Field) bool {
		return slices.Contains([]string{"password", "error_message", "deletion_last_error", "deletion_attempts", "deletion_last_attempt_at"}, f.Name)
	})
	for i := range fields {
		f := &fields[i]
		f.ReadOnly = entity.ReadOnly || f.Name == "deletion_requested_at" || f.Name == "consent_notice_version" || f.Name == "consent_accepted_at"
	}
	return fields, nil
}

// Identifiers here come exclusively from the entity catalog and database metadata.
func ident(value string) string { return `"` + strings.ReplaceAll(value, `"`, `""`) + `"` }

func columns(fields []Field) string {
	names := make([]string, 0, len(fields))
	for _, f := range fields {
		names = append(names, "r."+ident(f.Name))
	}
	return strings.Join(names, ",")
}
