package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var metrics = []string{"registrations", "social_likes", "matches", "social_messages", "team_chats", "team_messages", "teams", "memberships"}

func dayBucket(t time.Time) string { return t.Format("2006-01-02") }

func (h *Handler) analytics(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days != 7 && days != 30 && days != 90 {
		fail(c, 422, "Выберите период 7, 30 или 90 дней.")
		return
	}
	tz, err := time.LoadLocation("Asia/Almaty")
	if err != nil {
		fail(c, 500, "Не удалось определить часовой пояс.")
		return
	}
	now := time.Now().In(tz)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)
	start := today.AddDate(0, 0, 1-days)
	previous := start.AddDate(0, 0, -days)
	lastEnd := now.AddDate(0, 0, -days)
	type countRow struct {
		Day, Metric string
		Total       int64
	}
	var counts []countRow
	err = h.db.WithContext(c.Request.Context()).Raw(`
WITH windows AS (SELECT ?::timestamptz AS earliest, ?::timestamptz AS prior_end, ?::timestamptz AS current_start, ?::timestamptz AS latest),
 events AS (
 SELECT 'registrations' AS metric,created_at AT TIME ZONE 'UTC' AS happened_at,id AS actor FROM users
 UNION ALL SELECT 'social_likes',created_at,swiper_id FROM swipes WHERE direction='right'
 UNION ALL SELECT 'matches',created_at,user_one_id FROM matches
 UNION ALL SELECT CASE WHEN c.course_id IS NULL THEN 'social_messages' ELSE 'team_messages' END,m.created_at,m.sender_id FROM messages m JOIN chats c ON c.id=m.chat_id
 UNION ALL SELECT 'team_chats',created_at,initiated_by FROM chats WHERE course_id IS NOT NULL
 UNION ALL SELECT 'teams',created_at,created_by FROM teams
 UNION ALL SELECT 'memberships',joined_at,user_id FROM team_members
 ),
 eligible AS (
 SELECT metric,happened_at,actor,current_start
 FROM events CROSS JOIN windows
 WHERE (happened_at>=earliest AND happened_at<prior_end) OR (happened_at>=current_start AND happened_at<latest)
 )
 SELECT to_char(happened_at AT TIME ZONE 'Asia/Almaty','YYYY-MM-DD') AS day,metric,count(*) AS total
 FROM eligible GROUP BY 1,2
 UNION ALL
 SELECT (CASE WHEN happened_at>=current_start THEN 'current' ELSE 'previous' END),'active_users',count(DISTINCT actor)
 FROM eligible WHERE metric<>'registrations' GROUP BY 1`, previous, lastEnd, start, now).Scan(&counts).Error
	if err != nil {
		fail(c, 500, "Не удалось рассчитать статистику.")
		return
	}
	series := make([]map[string]any, 0, days)
	daily := make(map[string]map[string]any, days)
	current := map[string]int64{"active_users": 0}
	prior := map[string]int64{"active_users": 0}
	for i := 0; i < days; i++ {
		key := dayBucket(start.AddDate(0, 0, i))
		bucket := map[string]any{"date": key}
		for _, m := range metrics {
			bucket[m] = int64(0)
			current[m] = 0
			prior[m] = 0
		}
		series = append(series, bucket)
		daily[key] = bucket
	}
	for _, row := range counts {
		if row.Metric == "active_users" {
			if row.Day == "current" {
				current[row.Metric] = row.Total
			} else {
				prior[row.Metric] = row.Total
			}
		} else if bucket := daily[row.Day]; bucket != nil {
			bucket[row.Metric] = row.Total
			current[row.Metric] += row.Total
		} else {
			prior[row.Metric] += row.Total
		}
	}
	ok(c, gin.H{"days": days, "from": dayBucket(start), "to": dayBucket(today), "timezone": "Asia/Almaty", "series": series, "current": current, "previous": prior})
}
