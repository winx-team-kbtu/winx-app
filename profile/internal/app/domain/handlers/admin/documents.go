package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	header "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/domain/repositories/profile"
	"winx-profile/internal/app/domain/repositories/profilemeta"
	"winx-profile/internal/app/models/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/gorm"
)

func ProfileDocuments(db *gorm.DB, documents *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		viewer, err := header.GetAuthUser(ctx)
		if err != nil {
			fail(c, 401, "Войдите в аккаунт.")
			return
		}
		var allowed bool
		if err := db.WithContext(ctx).Raw(`SELECT EXISTS(SELECT 1 FROM users u JOIN roles r ON r.id=u.role_id WHERE u.id=? AND r.slug='admin' AND u.deletion_requested_at IS NULL)`, viewer.ID).Scan(&allowed).Error; err != nil {
			fail(c, 500, "Не удалось проверить права.")
			return
		}
		if !allowed {
			fail(c, 403, "Нужна роль администратора.")
			return
		}
		if c.Query("entity") != "profile_documents" {
			fail(c, 404, "Сущность не найдена.")
			return
		}
		fields, writable := documentFields()
		if c.Request.Method == http.MethodGet {
			list(c, documents.Collection("profiles"), fields)
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 128<<10)
		var input struct {
			Key struct {
				UserID int64 `json:"user_id"`
			} `json:"key"`
			Values map[string]json.RawMessage `json:"values"`
		}
		if c.ShouldBindJSON(&input) != nil {
			fail(c, 400, "Неверный формат записи.")
			return
		}
		action := map[string]string{"POST": "create", "PATCH": "update", "DELETE": "delete"}[c.Request.Method]
		id := input.Key.UserID
		if action == "create" {
			_ = json.Unmarshal(input.Values["user_id"], &id)
			delete(input.Values, "user_id")
		}
		if id <= 0 {
			fail(c, 422, "Укажите ID пользователя.")
			return
		}
		for name := range input.Values {
			if !writable[name] {
				fail(c, 422, "Недопустимое поле: "+name)
				return
			}
		}
		repo := profile.NewRepository(documents)
		var previous models.Profile
		var changed bool
		err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var email string
			row := tx.Raw("SELECT email FROM users WHERE id=? AND deletion_requested_at IS NULL FOR UPDATE", id).Scan(&email)
			if row.Error != nil {
				return row.Error
			}
			if row.RowsAffected == 0 {
				return profile.ErrNotFound
			}
			existing, err := repo.GetByUserID(ctx, id)
			if err != nil && !errors.Is(err, profile.ErrNotFound) {
				return err
			}
			previous = existing
			if action == "create" && err == nil {
				return duplicate
			}
			if action != "create" && err != nil {
				return err
			}
			if action == "delete" {
				if err := repo.DeleteByUserID(ctx, id); err != nil {
					return err
				}
				changed = true
				for _, table := range []string{"user_interests", "matching_preferences", "profiles"} {
					if err := tx.Exec("DELETE FROM "+table+" WHERE user_id=?", id).Error; err != nil {
						return err
					}
				}
			} else {
				patch, _ := json.Marshal(input.Values)
				next := existing
				if err := json.Unmarshal(patch, &next); err != nil {
					return invalid
				}
				next.UserID, next.Email, next.UpdatedAt = id, email, time.Now().UTC()
				if action == "create" {
					next.CreatedAt = next.UpdatedAt
					_, err = repo.Create(ctx, next)
				} else {
					_, err = repo.Update(ctx, next)
				}
				if err != nil {
					return err
				}
				changed = true
				if err := profilemeta.NewRepository(tx).Sync(ctx, next); err != nil {
					return err
				}
			}
			return tx.Exec("INSERT INTO admin_audit_logs(actor_id,entity,action,record_key) VALUES (?,'profile_documents',?,jsonb_build_object('user_id',?::bigint))", viewer.ID, action, id).Error
		})
		if err != nil {
			if changed {
				cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
				defer cancel()
				switch action {
				case "create":
					_ = repo.DeleteByUserID(cleanup, id)
				case "update":
					_, _ = repo.Update(cleanup, previous)
				case "delete":
					_, _ = repo.Create(cleanup, previous)
				}
			}
			switch {
			case errors.Is(err, profile.ErrNotFound):
				fail(c, 404, "Пользователь или анкета не найдены.")
			case errors.Is(err, duplicate), mongo.IsDuplicateKeyError(err):
				fail(c, 409, "Анкета уже существует.")
			case errors.Is(err, invalid), errors.Is(err, profilemeta.ErrInvalidInterestIDs):
				fail(c, 422, "Проверьте значения полей.")
			default:
				fail(c, 500, "Не удалось сохранить анкету.")
			}
			return
		}
		c.JSON(200, response.SuccessResponse(gin.H{"user_id": id}, response.OK))
	}
}

var invalid = errors.New("invalid profile")
var duplicate = errors.New("duplicate profile")

func fail(c *gin.Context, status int, message string) {
	c.JSON(status, response.ErrorResponse(message))
}
func documentFields() ([]gin.H, map[string]bool) {
	var fields []gin.H
	writable := map[string]bool{}
	t := reflect.TypeOf(models.Profile{})
	for n := 0; n < t.NumField(); n++ {
		f := t.Field(n)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		key := name == "user_id"
		readOnly := name == "id" || name == "email" || strings.HasSuffix(name, "_at")
		kind := f.Type.Kind()
		if kind == reflect.Pointer {
			kind = f.Type.Elem().Kind()
		}
		fieldType := "jsonb"
		switch kind {
		case reflect.String:
			fieldType = "text"
		case reflect.Int64:
			fieldType = "bigint"
		case reflect.Bool:
			fieldType = "boolean"
		}
		if strings.HasSuffix(name, "_at") {
			fieldType = "timestamp"
		}
		fields = append(fields, gin.H{"name": name, "type": fieldType, "key": key, "read_only": readOnly, "required": key, "generated": false})
		if !readOnly && !key {
			writable[name] = true
		}
	}
	return fields, writable
}

func list(c *gin.Context, col *mongo.Collection, fields []gin.H) {
	q := strings.TrimSpace(c.Query("q"))
	if len(q) > 240 {
		fail(c, 400, "Слишком длинный поиск.")
		return
	}
	p, _ := strconv.Atoi(c.Query("page"))
	p = min(10000, max(1, p))
	filter := bson.M{}
	if q != "" {
		regex := bson.Regex{Pattern: regexp.QuoteMeta(q), Options: "i"}
		options := bson.A{bson.M{"email": regex}, bson.M{"first_name": regex}}
		if id, err := strconv.ParseInt(q, 10, 64); err == nil {
			options = append(options, bson.M{"user_id": id})
		}
		filter["$or"] = options
	}
	cursor, err := col.Find(c.Request.Context(), filter, options.Find().SetSort(bson.D{{Key: "user_id", Value: -1}}).SetSkip(int64((p-1)*25)).SetLimit(26))
	if err != nil {
		fail(c, 500, "Не удалось загрузить анкеты.")
		return
	}
	rows := make([]models.Profile, 0)
	if err := cursor.All(c.Request.Context(), &rows); err != nil {
		fail(c, 500, "Не удалось загрузить анкеты.")
		return
	}
	c.JSON(200, response.SuccessResponse(gin.H{"fields": fields, "rows": rows[:min(len(rows), 25)], "page": p, "has_more": len(rows) > 25}, response.OK))
}
