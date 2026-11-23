package conversation

import (
	"context"
	"encoding/json"
	"errors"

	"winx-profile/internal/app/models/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("Диалог не найден.")
	ErrClosed    = errors.New("Общение уже завершено.")
	ErrCourse    = errors.New("У обоих участников должен быть выбран этот предмет.")
	ErrFull      = errors.New("В команде уже 5 участников.")
	ErrOtherTeam = errors.New("Собеседник уже состоит в другой команде по этому предмету.")
	ErrName      = errors.New("Введите название команды: от 2 до 50 символов.")
	ErrConflict  = errors.New("Состав команды изменился. Попробуйте ещё раз.")
)

type Repository struct{ db *gorm.DB }

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

const roomQuery = `SELECT c.id,c.course_id,cr.code AS course_code,cr.title AS course_title,
 CASE WHEN c.user_one_id=@user THEN c.user_two_id ELSE c.user_one_id END AS peer_id,
 COALESCE(p.name,'Студент') AS peer_name,COALESCE(ap.preferred_role,'') AS preferred_role,
 COALESCE(ap.skills,'[]'::jsonb) AS skills,c.status,c.initiated_by,c.closed_by,c.team_id,c.created_at,
 mt.id AS my_team_id,COALESCE(mt.name,'') AS my_team_name,
 (SELECT COUNT(*) FROM team_members WHERE team_id=mt.id) AS my_team_size,pm.team_id AS peer_team_id,
 i.id AS invitation_id,i.inviter_id,i.status AS invitation_status,it.name AS invitation_team_name,
 COALESCE(lm.content,'') AS last_message,lm.created_at AS last_message_at,
 (c.status<>'declined' AND EXISTS(SELECT 1 FROM user_courses WHERE user_id=c.user_one_id AND course_id=c.course_id)
 AND EXISTS(SELECT 1 FROM user_courses WHERE user_id=c.user_two_id AND course_id=c.course_id)) AS can_write
 FROM chats c JOIN courses cr ON cr.id=c.course_id
 JOIN users ua ON ua.id=c.user_one_id AND ua.deletion_requested_at IS NULL
 JOIN users ub ON ub.id=c.user_two_id AND ub.deletion_requested_at IS NULL
 LEFT JOIN profiles p ON p.user_id=CASE WHEN c.user_one_id=@user THEN c.user_two_id ELSE c.user_one_id END
 LEFT JOIN academic_profiles ap ON ap.user_id=CASE WHEN c.user_one_id=@user THEN c.user_two_id ELSE c.user_one_id END
 LEFT JOIN team_members me ON me.user_id=@user AND me.course_id=c.course_id
 LEFT JOIN teams mt ON mt.id=me.team_id
 LEFT JOIN team_members pm ON pm.user_id=CASE WHEN c.user_one_id=@user THEN c.user_two_id ELSE c.user_one_id END AND pm.course_id=c.course_id
 LEFT JOIN team_invitations i ON i.chat_id=c.id
 LEFT JOIN teams it ON it.id=i.team_id
 LEFT JOIN LATERAL(SELECT content,created_at FROM messages WHERE chat_id=c.id ORDER BY id DESC LIMIT 1) lm ON TRUE
 WHERE (c.user_one_id=@user OR c.user_two_id=@user) AND c.course_id IS NOT NULL `

func decode(room *models.Conversation) error {
	room.Skills = []string{}
	if len(room.SkillsJSON) > 0 {
		return json.Unmarshal(room.SkillsJSON, &room.Skills)
	}
	return nil
}
func (r *Repository) List(ctx context.Context, user, course int64, limit, offset int) ([]models.Conversation, error) {
	rooms := []models.Conversation{}
	err := r.db.WithContext(ctx).Raw(roomQuery+` AND c.course_id=@course AND c.status<>'declined'
	ORDER BY COALESCE(i.status='pending' AND i.inviter_id<>@user,false) DESC,
 COALESCE(lm.created_at,c.created_at) DESC,c.id DESC LIMIT @limit OFFSET @offset`,
		map[string]interface{}{"user": user, "course": course, "limit": limit, "offset": offset}).Scan(&rooms).Error
	if err != nil {
		return nil, err
	}
	for i := range rooms {
		if err := decode(&rooms[i]); err != nil {
			return nil, err
		}
	}
	return rooms, nil
}
func (r *Repository) Get(ctx context.Context, user, id int64) (models.Conversation, error) {
	var room models.Conversation
	result := r.db.WithContext(ctx).Raw(roomQuery+` AND c.id=@id`, map[string]interface{}{"user": user, "id": id}).Scan(&room)
	if result.Error != nil {
		return room, result.Error
	}
	if result.RowsAffected == 0 {
		return room, ErrNotFound
	}
	return room, decode(&room)
}
func (r *Repository) Messages(ctx context.Context, user, id, before int64) ([]models.ConversationMessage, error) {
	if _, err := r.Get(ctx, user, id); err != nil {
		return nil, err
	}
	rows := []models.ConversationMessage{}
	err := r.db.WithContext(ctx).Raw(`SELECT * FROM(SELECT id,chat_id,sender_id,content,created_at FROM messages
 WHERE chat_id=? AND (?::bigint=0 OR id<?) ORDER BY id DESC LIMIT 50) m ORDER BY id`, id, before, before).Scan(&rows).Error
	return rows, err
}

// OpenForTeammate reuses the course conversation, including its closure state.
func (r *Repository) OpenForTeammate(ctx context.Context, user, peer, course int64) (int64, error) {
	var chatID int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var teamID int64
		if err := tx.Raw(`SELECT t.id FROM teams t JOIN team_members me ON me.team_id=t.id
 WHERE t.course_id=? AND me.user_id=? FOR SHARE OF t`, course, user).Scan(&teamID).Error; err != nil {
			return err
		}
		if teamID == 0 || user == peer {
			return ErrNotFound
		}
		// Take a fresh snapshot after locking the team, in case a member just left.
		var teammates int64
		if err := tx.Raw(`SELECT COUNT(*) FROM team_members m JOIN users u ON u.id=m.user_id
 WHERE m.team_id=? AND m.user_id IN (?,?) AND u.deletion_requested_at IS NULL`, teamID, user, peer).Scan(&teammates).Error; err != nil {
			return err
		}
		if teammates != 2 {
			return ErrNotFound
		}
		first, second := min(user, peer), max(user, peer)
		if err := tx.Exec(`INSERT INTO chats(course_id,user_one_id,user_two_id,initiated_by,status,team_id)
 VALUES(?,?,?,?,'joined',?) ON CONFLICT(course_id,user_one_id,user_two_id) DO NOTHING`, course, first, second, user, teamID).Error; err != nil {
			return err
		}
		return tx.Raw(`SELECT id FROM chats WHERE course_id=? AND user_one_id=? AND user_two_id=?`, course, first, second).Scan(&chatID).Error
	})
	return chatID, err
}

type lockedRoom struct {
	ID, CourseID, UserOneID, UserTwoID int64
	Status                             string
}

func (room lockedRoom) peer(user int64) int64 {
	if room.UserOneID == user {
		return room.UserTwoID
	}
	return room.UserOneID
}

func (r *Repository) Decide(ctx context.Context, user, id int64, decision, name string, invitationID int64) (models.Conversation, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var room lockedRoom
		result := tx.Raw(`SELECT id,course_id,user_one_id,user_two_id,status FROM chats
 WHERE id=? AND course_id IS NOT NULL AND (user_one_id=? OR user_two_id=?) FOR UPDATE`, id, user, user).Scan(&room)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		switch room.Status {
		case "declined":
			if decision == "decline" {
				return nil
			}
			return ErrClosed
		case "joined":
			if decision == "accept" {
				return confirmAcceptedInvitation(tx, room, user, invitationID)
			}
			if decision == "invite" || decision == "add" {
				return nil
			}
			return ErrClosed
		}

		switch decision {
		case "decline":
			return endConversation(tx, room, user)
		case "add", "invite":
			return sendInvitation(tx, room, user, name)
		case "accept", "reject":
			return answerInvitation(tx, room, user, invitationID, decision == "accept")
		default:
			return ErrConflict
		}
	})
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && (pg.Code == "23505" || pg.Code == "40P01") {
			return models.Conversation{}, ErrConflict
		}
		return models.Conversation{}, err
	}
	return r.Get(ctx, user, id)
}

func endConversation(tx *gorm.DB, room lockedRoom, user int64) error {
	if err := tx.Exec("DELETE FROM team_invitations WHERE chat_id=?", room.ID).Error; err != nil {
		return err
	}
	if err := tx.Exec(`UPDATE chats SET status='declined',closed_by=?,updated_at=NOW() WHERE id=?`, user, room.ID).Error; err != nil {
		return err
	}
	peer := room.peer(user)
	return tx.Exec(`INSERT INTO course_choices(user_id,candidate_id,course_id,choice)
 VALUES(?,?,?,'pass'),(?,?,?,'pass') ON CONFLICT(user_id,candidate_id,course_id)
 DO UPDATE SET choice='pass',updated_at=NOW()`, user, peer, room.CourseID, peer, user, room.CourseID).Error
}
