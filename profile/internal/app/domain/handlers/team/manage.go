package team

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	header "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/response"
)

func ManageMember(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		u, err := header.GetAuthUser(ctx.Request.Context())
		if err != nil {
			ctx.JSON(401, response.ErrorResponse("Войдите в аккаунт."))
			return
		}
		var input struct {
			CourseID int64 `json:"course_id"`
			MemberID int64 `json:"member_id"`
			TeamID   int64 `json:"team_id"`
		}
		if ctx.ShouldBindJSON(&input) != nil || input.CourseID <= 0 || input.MemberID <= 0 {
			ctx.JSON(422, response.ErrorResponse("Укажите предмет и участника."))
			return
		}
		forbidden := errors.New("Удалять участников может только создатель команды.")
		missing := errors.New("Участник не найден.")
		err = db.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
			var team struct {
				ID        int64
				CreatedBy *int64
			}
			r := tx.Raw(`SELECT id,created_by FROM teams WHERE course_id=? AND (?::bigint=0 OR id=?) AND id IN(SELECT team_id FROM team_members WHERE user_id=?) FOR UPDATE`, input.CourseID, input.TeamID, input.TeamID, u.ID).Scan(&team)
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected == 0 {
				return missing
			}
			if input.MemberID != u.ID && (team.CreatedBy == nil || *team.CreatedBy != u.ID) {
				return forbidden
			}
			removed := tx.Exec("DELETE FROM team_members WHERE team_id=? AND user_id=?", team.ID, input.MemberID)
			if removed.Error != nil {
				return removed.Error
			}
			if removed.RowsAffected == 0 {
				return missing
			}
			if err := tx.Exec(`WITH reopened AS (
 UPDATE chats SET status='open',team_id=NULL,updated_at=NOW()
 WHERE team_id=? AND status='joined' AND (user_one_id=? OR user_two_id=?) RETURNING id
) DELETE FROM team_invitations WHERE chat_id IN (SELECT id FROM reopened)`, team.ID, input.MemberID, input.MemberID).Error; err != nil {
				return err
			}
			var next int64
			if err := tx.Raw("SELECT user_id FROM team_members WHERE team_id=? ORDER BY joined_at,user_id LIMIT 1", team.ID).Scan(&next).Error; err != nil {
				return err
			}
			if next == 0 {
				return tx.Exec("DELETE FROM teams WHERE id=?", team.ID).Error
			}
			if team.CreatedBy == nil || *team.CreatedBy == input.MemberID {
				return tx.Exec("UPDATE teams SET created_by=? WHERE id=?", next, team.ID).Error
			}
			return nil
		})
		if err != nil {
			status := 500
			if errors.Is(err, forbidden) {
				status = 403
			}
			if errors.Is(err, missing) {
				status = 404
			}
			message := err.Error()
			if status == 500 {
				message = "Не удалось обновить состав."
			}
			ctx.JSON(status, response.ErrorResponse(message))
			return
		}
		ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.Updated))
	}
}
