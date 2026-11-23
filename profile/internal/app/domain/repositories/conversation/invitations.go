package conversation

import (
	"crypto/rand"
	"errors"
	"slices"
	"strings"
	"unicode/utf8"

	"winx-profile/internal/app/domain/repositories/team"

	"gorm.io/gorm"
)

var (
	ErrInviteRecipient = errors.New("Ответить на приглашение может только получатель.")
	ErrInviteMissing   = errors.New("Приглашение больше недоступно.")
	ErrInvitePending   = errors.New("Ответьте на полученное приглашение.")
	ErrAlreadyMember   = errors.New("Вы уже состоите в другой команде по этому предмету.")
)

type invitation struct {
	ID, TeamID, InviterID int64
	Status                string
}

func loadInvitation(tx *gorm.DB, chatID int64) (invitation, error) {
	var invite invitation
	err := tx.Raw("SELECT id,team_id,inviter_id,status FROM team_invitations WHERE chat_id=?", chatID).Scan(&invite).Error
	return invite, err
}

func sendInvitation(tx *gorm.DB, room lockedRoom, user int64, name string) error {
	invite, err := loadInvitation(tx, room.ID)
	if err != nil {
		return err
	}
	if invite.Status == "pending" {
		if invite.InviterID == user {
			return nil
		}
		return ErrInvitePending
	}
	if err := lockEnrollment(tx, room); err != nil {
		return err
	}
	peer := room.peer(user)
	var teamID int64
	if err := tx.Raw("SELECT team_id FROM team_members WHERE user_id=? AND course_id=?", user, room.CourseID).Scan(&teamID).Error; err != nil {
		return err
	}
	if teamID == 0 {
		name = strings.TrimSpace(name)
		length := utf8.RuneCountInString(name)
		if length < 2 || length > 50 {
			return ErrName
		}
		code, err := newCode()
		if err != nil {
			return err
		}
		created, err := team.NewRepository(tx).Create(tx.Statement.Context, user, room.CourseID, name, code)
		if errors.Is(err, team.ErrAlreadyInTeam) || errors.Is(err, team.ErrCodeTaken) {
			return ErrConflict
		}
		if err != nil {
			return err
		}
		teamID = created.ID
	}
	joined, err := checkTeamSeat(tx, room.CourseID, teamID, user, peer, ErrOtherTeam)
	if err != nil {
		return err
	}
	if joined {
		if err := tx.Exec("DELETE FROM team_invitations WHERE chat_id=?", room.ID).Error; err != nil {
			return err
		}
		return markJoined(tx, room.ID, teamID)
	}
	return tx.Exec(`INSERT INTO team_invitations(chat_id,team_id,inviter_id) VALUES(?,?,?)
 ON CONFLICT (chat_id) DO UPDATE SET id=EXCLUDED.id,team_id=EXCLUDED.team_id,
 inviter_id=EXCLUDED.inviter_id,status='pending'`, room.ID, teamID, user).Error
}

func recipientInvitation(tx *gorm.DB, room lockedRoom, user, invitationID int64) (invitation, error) {
	invite, err := loadInvitation(tx, room.ID)
	if err != nil {
		return invitation{}, err
	}
	if invite.ID == 0 || invite.ID != invitationID {
		return invitation{}, ErrInviteMissing
	}
	if invite.InviterID != room.peer(user) {
		return invitation{}, ErrInviteRecipient
	}
	return invite, nil
}

func confirmAcceptedInvitation(tx *gorm.DB, room lockedRoom, user, invitationID int64) error {
	invite, err := recipientInvitation(tx, room, user, invitationID)
	if err == nil && invite.Status != "accepted" {
		return ErrInviteMissing
	}
	return err
}

func answerInvitation(tx *gorm.DB, room lockedRoom, user, invitationID int64, accept bool) error {
	invite, err := recipientInvitation(tx, room, user, invitationID)
	if err != nil {
		return err
	}
	if !accept {
		return tx.Exec("UPDATE team_invitations SET status='declined' WHERE id=?", invite.ID).Error
	}
	if invite.Status != "pending" {
		return ErrInviteMissing
	}
	if err := lockEnrollment(tx, room); err != nil {
		return err
	}
	joined, err := checkTeamSeat(tx, room.CourseID, invite.TeamID, invite.InviterID, user, ErrAlreadyMember)
	if err != nil {
		return err
	}
	if !joined {
		if err := tx.Exec(`INSERT INTO team_members(team_id,user_id,course_id) VALUES(?,?,?)`, invite.TeamID, user, room.CourseID).Error; err != nil {
			return err
		}
	}
	if err := tx.Exec("UPDATE team_invitations SET status='accepted' WHERE id=?", invite.ID).Error; err != nil {
		return err
	}
	return markJoined(tx, room.ID, invite.TeamID)
}

func lockEnrollment(tx *gorm.DB, room lockedRoom) error {
	var eligible []int64
	if err := tx.Raw(`SELECT uc.user_id FROM user_courses uc JOIN users u ON u.id=uc.user_id
 WHERE uc.user_id IN (?,?) AND uc.course_id=? AND u.deletion_requested_at IS NULL
 ORDER BY uc.user_id FOR SHARE OF uc,u`, room.UserOneID, room.UserTwoID, room.CourseID).Scan(&eligible).Error; err != nil {
		return err
	}
	if len(eligible) != 2 {
		return ErrCourse
	}
	return nil
}

func checkTeamSeat(tx *gorm.DB, courseID, teamID, inviter, invitee int64, otherTeam error) (bool, error) {
	// Joining by code and removing a member take the same team lock.
	var locked int64
	if err := tx.Raw("SELECT id FROM teams WHERE id=? AND course_id=? FOR UPDATE", teamID, courseID).Scan(&locked).Error; err != nil {
		return false, err
	}
	if locked == 0 {
		return false, ErrInviteMissing
	}
	var members []int64
	if err := tx.Raw("SELECT user_id FROM team_members WHERE team_id=?", teamID).Scan(&members).Error; err != nil {
		return false, err
	}
	if !slices.Contains(members, inviter) {
		return false, ErrInviteMissing
	}
	if slices.Contains(members, invitee) {
		return true, nil
	}
	var inviteeTeam int64
	if err := tx.Raw("SELECT team_id FROM team_members WHERE user_id=? AND course_id=?", invitee, courseID).Scan(&inviteeTeam).Error; err != nil {
		return false, err
	}
	if inviteeTeam != 0 {
		return false, otherTeam
	}
	if len(members) >= team.MaxMembers {
		return false, ErrFull
	}
	return false, nil
}

func markJoined(tx *gorm.DB, chatID, teamID int64) error {
	return tx.Exec("UPDATE chats SET status='joined',team_id=?,updated_at=NOW() WHERE id=?", teamID, chatID).Error
}

func newCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var buf [6]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	for i, v := range buf {
		buf[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(buf[:]), nil
}
