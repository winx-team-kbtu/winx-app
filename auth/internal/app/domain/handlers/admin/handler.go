package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	header "auth/internal/app/core/contracts/microservices/header-contract"
	"auth/internal/app/core/helpers/password"
	"auth/internal/app/core/helpers/response"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Handler struct {
	db             *gorm.DB
	deleteAccount  func(context.Context, int64) error
	revokeSessions func(context.Context, int64) error
}

func New(db *gorm.DB, deleteAccount, revokeSessions func(context.Context, int64) error) *Handler {
	return &Handler{db: db, deleteAccount: deleteAccount, revokeSessions: revokeSessions}
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	r.GET("/catalog", h.catalog)
	r.GET("/analytics", h.analytics)
	r.GET("/records", h.list)
	r.POST("/records", h.mutate)
	r.PATCH("/records", h.mutate)
	r.DELETE("/records", h.mutate)
}
func (h *Handler) catalog(c *gin.Context) { ok(c, entities) }
func ok(c *gin.Context, data any)         { c.JSON(200, response.SuccessResponse(data, response.OK)) }
func fail(c *gin.Context, status int, message string) {
	c.JSON(status, response.ErrorResponse(message))
}

func (h *Handler) getEntity(c *gin.Context) (Entity, []Field, bool) {
	name := c.Query("entity")
	i := slices.IndexFunc(entities, func(e Entity) bool { return e.Name == name && name != "profile_documents" })
	if i < 0 {
		fail(c, 404, "Сущность не найдена.")
		return Entity{}, nil, false
	}
	e := entities[i]
	fields, err := schema(c.Request.Context(), h.db, e)
	if err != nil || len(fields) == 0 {
		fail(c, 500, "Схема сущности недоступна.")
		return e, nil, false
	}
	return e, fields, true
}
func page(c *gin.Context) int {
	n, _ := strconv.Atoi(c.Query("page"))
	return min(10000, max(1, n))
}
func (h *Handler) list(c *gin.Context) {
	e, fields, valid := h.getEntity(c)
	if !valid {
		return
	}
	p := page(c)
	q := strings.TrimSpace(c.Query("q"))
	if len(q) > 240 {
		fail(c, 400, "Слишком длинный поиск.")
		return
	}
	cols := columns(fields)
	sortKeys := keyNames(fields)
	if len(sortKeys) == 0 {
		fail(c, 500, "Первичный ключ отсутствует.")
		return
	}
	var args []any
	where := ""
	if q != "" {
		// Search only fields that are visible in the result.
		where = " WHERE strpos(lower(concat_ws(' ', " + cols + ")),lower(?))>0"
		args = append(args, q)
	}
	args = append(args, 26, (p-1)*25)
	order := make([]string, 0, len(sortKeys))
	for _, k := range sortKeys {
		order = append(order, "r."+ident(k)+" DESC")
	}
	statement := "SELECT row_to_json(row_data)::text FROM (SELECT " + cols + " FROM public." + ident(e.Name) + " r" + where +
		" ORDER BY " + strings.Join(order, ",") + " LIMIT ? OFFSET ?) row_data"
	var raw []string
	if err := h.db.WithContext(c.Request.Context()).Raw(statement, args...).Scan(&raw).Error; err != nil {
		fail(c, 500, "Не удалось загрузить записи.")
		return
	}
	rows := make([]json.RawMessage, 0, min(len(raw), 25))
	for _, row := range raw[:min(len(raw), 25)] {
		rows = append(rows, json.RawMessage(row))
	}
	ok(c, gin.H{"fields": fields, "rows": rows, "page": p, "has_more": len(raw) > 25})
}

type input struct {
	Key      map[string]json.RawMessage `json:"key"`
	Values   map[string]json.RawMessage `json:"values"`
	Password string                     `json:"password"`
}

var invalid = errors.New("invalid fields")
var missing = errors.New("missing row")
var protected = errors.New("protected record")
var unauthorized = errors.New("admin access required")

func keyNames(fields []Field) []string {
	var keys []string
	for _, f := range fields {
		if f.Key {
			keys = append(keys, f.Name)
		}
	}
	return keys
}
func (h *Handler) mutate(c *gin.Context) {
	e, fields, valid := h.getEntity(c)
	if !valid {
		return
	}
	if e.ReadOnly {
		fail(c, 405, "Журнал доступен для чтения.")
		return
	}
	actor, err := header.GetAuthUser(c.Request.Context())
	if err != nil {
		fail(c, 401, "Войдите в аккаунт.")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 128<<10)
	var data input
	if c.ShouldBindJSON(&data) != nil {
		fail(c, 400, "Неверный формат записи.")
		return
	}
	keys := keyNames(fields)
	action := map[string]string{"POST": "create", "PATCH": "update", "DELETE": "delete"}[c.Request.Method]
	if len(keys) == 0 {
		fail(c, 409, "Запись без первичного ключа.")
		return
	}
	if action != "create" {
		if len(data.Key) != len(keys) {
			fail(c, 422, "Укажите полный ключ записи.")
			return
		}
		for _, k := range keys {
			if len(data.Key[k]) == 0 || string(data.Key[k]) == "null" {
				fail(c, 422, "Укажите полный ключ записи.")
				return
			}
		}
	}
	if data.Values == nil {
		data.Values = map[string]json.RawMessage{}
	}
	var names []string
	for key := range data.Values {
		i := slices.IndexFunc(fields, func(f Field) bool { return f.Name == key })
		if i < 0 || fields[i].Generated || fields[i].ReadOnly || fields[i].Key && action != "create" {
			fail(c, 422, "Недопустимое поле: "+key)
			return
		}
		names = append(names, key)
	}
	if data.Password != "" {
		if e.Name != "users" || action == "delete" || len(data.Password) < 8 || len(data.Password) > 72 {
			fail(c, 422, "Пароль должен содержать от 8 до 72 байт.")
			return
		}
		hash, err := password.Hash(data.Password)
		if err != nil {
			fail(c, 500, "Не удалось сохранить пароль.")
			return
		}
		data.Values["password"], _ = json.Marshal(hash)
		names = append(names, "password")
	}
	if e.Name == "users" && action == "create" && data.Password == "" {
		fail(c, 422, "Укажите пароль.")
		return
	}
	slices.Sort(names)
	table := "public." + ident(e.Name)
	where := make([]string, 0, len(keys))
	for _, k := range keys {
		where = append(where, "r."+ident(k)+" IS NOT DISTINCT FROM k."+ident(k))
	}
	keyJSON, _ := json.Marshal(data.Key)
	valuesJSON, _ := json.Marshal(data.Values)
	ctx := c.Request.Context()
	var userID int64
	var changed string
	err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(26311908)").Error; err != nil {
			return err
		}
		var allowed bool
		if err := tx.Raw(`SELECT EXISTS(SELECT 1 FROM users u JOIN roles r ON r.id=u.role_id WHERE u.id=? AND u.deletion_requested_at IS NULL AND r.slug='admin')`, actor.ID).Scan(&allowed).Error; err != nil {
			return err
		}
		if !allowed {
			return unauthorized
		}
		keyRecord := "jsonb_populate_record(NULL::" + table + ",CAST(? AS jsonb)) k"
		newRecord := "jsonb_populate_record(NULL::" + table + ",CAST(? AS jsonb)) v"
		if e.Name == "users" && action != "create" {
			if json.Unmarshal(data.Key["id"], &userID) != nil || userID <= 0 {
				return invalid
			}
			if userID == actor.ID && (action == "delete" || data.Values["role_id"] != nil || data.Password != "") {
				return protected
			}
		}
		if e.Name == "roles" && action != "create" {
			var core bool
			if err := tx.Raw("SELECT EXISTS(SELECT 1 FROM "+table+" r, "+keyRecord+" WHERE "+strings.Join(where, " AND ")+" AND r.slug IN ('admin','user'))", string(keyJSON)).Scan(&core).Error; err != nil {
				return err
			}
			if core && (action == "delete" || data.Values["slug"] != nil) {
				return protected
			}
		}
		keyParts := make([]string, 0, len(keys)*2)
		for _, k := range keys {
			keyParts = append(keyParts, "'"+k+"'", "r."+ident(k))
		}
		returning := " RETURNING jsonb_build_object(" + strings.Join(keyParts, ",") + ")::text"
		var query string
		var args []any
		switch action {
		case "create":
			if len(names) == 0 {
				query = "INSERT INTO " + table + " AS r DEFAULT VALUES"
			} else {
				var cols, vals []string
				for _, n := range names {
					cols = append(cols, ident(n))
					vals = append(vals, "v."+ident(n))
				}
				query = "INSERT INTO " + table + " AS r (" + strings.Join(cols, ",") + ") SELECT " + strings.Join(vals, ",") + " FROM " + newRecord
				args = []any{string(valuesJSON)}
			}
		case "update":
			if len(names) == 0 {
				return invalid
			}
			var sets []string
			for _, n := range names {
				sets = append(sets, ident(n)+"=v."+ident(n))
			}
			query = "UPDATE " + table + " r SET " + strings.Join(sets, ",") + " FROM " + newRecord + ", " + keyRecord + " WHERE " + strings.Join(where, " AND ")
			if e.Name == "users" {
				query += " AND r.deletion_requested_at IS NULL"
			}
			args = []any{string(valuesJSON), string(keyJSON)}
		case "delete":
			if e.Name == "users" {
				query = "UPDATE " + table + " r SET deletion_requested_at=COALESCE(r.deletion_requested_at,NOW()) FROM " + keyRecord + " WHERE " + strings.Join(where, " AND ")
			} else {
				query = "DELETE FROM " + table + " r USING " + keyRecord + " WHERE " + strings.Join(where, " AND ")
			}
			args = []any{string(keyJSON)}
		}
		result := tx.Raw(query+returning, args...).Scan(&changed)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return missing
		}
		return tx.Exec("INSERT INTO admin_audit_logs(actor_id,entity,action,record_key) VALUES (?,?,?,CAST(? AS jsonb))", actor.ID, e.Name, action, changed).Error
	})
	if err != nil {
		var pg *pgconn.PgError
		switch {
		case errors.Is(err, missing):
			fail(c, 404, "Запись не найдена.")
		case errors.Is(err, unauthorized):
			fail(c, 403, "Нужна роль администратора.")
		case errors.Is(err, protected):
			fail(c, 409, "Системную роль или собственные права нельзя изменить этим действием.")
		case errors.Is(err, invalid):
			fail(c, 422, "Проверьте значения полей.")
		case errors.As(err, &pg) && strings.HasPrefix(pg.Code, "23"):
			fail(c, 409, "Запись конфликтует с существующими данными или связями.")
		case errors.As(err, &pg) && strings.HasPrefix(pg.Code, "22"):
			fail(c, 422, "Проверьте типы и формат полей.")
		default:
			fail(c, 500, "Не удалось сохранить изменения.")
		}
		return
	}
	if e.Name == "users" && action == "delete" {
		if err := h.deleteAccount(ctx, userID); err != nil {
			c.JSON(202, response.SuccessResponse(json.RawMessage(changed), "Удаление поставлено в очередь."))
			return
		}
	}
	if e.Name == "users" && action == "update" && (data.Password != "" || data.Values["email"] != nil) {
		if err := h.revokeSessions(ctx, userID); err != nil {
			c.JSON(202, response.SuccessResponse(json.RawMessage(changed), "Запись сохранена. Не удалось завершить все сессии."))
			return
		}
	}
	ok(c, json.RawMessage(changed))
}
