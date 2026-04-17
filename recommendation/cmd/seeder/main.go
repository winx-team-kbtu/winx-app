// cmd/seeder seeds Postgres and MongoDB with test users for local recommendation testing.
// All users are placed at realistic coordinates within Almaty, Kazakhstan.
//
// Usage (from recommendation/):
//
//	go run ./cmd/seeder [-n 20] [-password secret]
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"
	"winx-recommendation/configs"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	configs.InitConfig()
}

// almatyLocations are real neighbourhood coordinates within Almaty, Kazakhstan.
// Format: {lat, lng, neighbourhood name, city label}.
var almatyLocations = []struct {
	lat, lng float64
	district string
}{
	{43.2220, 76.8512, "City Centre"},
	{43.1543, 76.9417, "Medeu"},
	{43.2354, 76.9453, "Almaty-2 Station"},
	{43.2225, 76.9462, "Dostyk"},
	{43.1997, 76.8948, "MEGA Alma-Ata"},
	{43.2143, 76.9564, "Kok-Tobe"},
	{43.2585, 76.9478, "Panfilov Park"},
	{43.2715, 76.9156, "Sayakhat"},
	{43.3012, 76.8234, "Alatau"},
	{43.2789, 76.8156, "Shymkentskiy"},
	{43.1456, 76.8234, "Baganashyl"},
	{43.2312, 76.8956, "Almaty Plaza"},
	{43.2478, 76.8634, "Raimbek"},
	{43.2634, 76.9234, "Zhibek Zholy"},
	{43.2156, 76.8256, "Sairan"},
	{43.1856, 76.8512, "Orbita"},
	{43.3156, 76.7856, "Alatau District"},
	{43.2534, 77.0234, "Nauryzbay"},
	{43.1756, 77.1234, "Altyn Emel"},
	{43.1956, 76.6234, "Kaskelen"},
}

type seedUser struct {
	name          string
	email         string
	gender        string
	birthDate     string // YYYY-MM-DD
	lookingFor    string
	aboutMe       string
	interestedIn  []string
	interestNames []string // subset of the interests seeded in DB
	districtIdx   int
}

var testUsers = []seedUser{
	{
		name: "Arman Bekov", email: "arman.bekov@seed.local",
		gender: "man", birthDate: "1998-03-15", lookingFor: "relationship",
		aboutMe:       "Love mountains and good coffee. Working in tech.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Gym", "Hiking", "Coffee", "Tech"},
		districtIdx:   0,
	},
	{
		name: "Aizat Nurova", email: "aizat.nurova@seed.local",
		gender: "woman", birthDate: "2000-07-22", lookingFor: "dating",
		aboutMe:       "Yoga instructor. Always up for brunch and travel.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Yoga", "Travel", "Brunch", "Photography"},
		districtIdx:   1,
	},
	{
		name: "Daniyar Seitkali", email: "daniyar.seitkali@seed.local",
		gender: "man", birthDate: "1995-11-30", lookingFor: "friends",
		aboutMe:       "Football fan and basketball player. Let's play.",
		interestedIn:  []string{"woman", "man"},
		interestNames: []string{"Football", "Basketball", "Gaming", "Beer"},
		districtIdx:   2,
	},
	{
		name: "Malika Dzhaksybekova", email: "malika.dzhaks@seed.local",
		gender: "woman", birthDate: "1997-05-10", lookingFor: "relationship",
		aboutMe:       "Bookworm and barista. Let's read together.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Reading", "Coffee", "Cooking", "Movies"},
		districtIdx:   3,
	},
	{
		name: "Yerlan Ospanov", email: "yerlan.ospanov@seed.local",
		gender: "man", birthDate: "1993-08-19", lookingFor: "relationship",
		aboutMe:       "Software engineer. Hiking on weekends.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Tech", "Hiking", "Camping", "Cycling"},
		districtIdx:   4,
	},
	{
		name: "Ainur Satova", email: "ainur.satova@seed.local",
		gender: "woman", birthDate: "2001-01-05", lookingFor: "dating",
		aboutMe:       "Artist and dreamer. Always painting something.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Painting", "Photography", "Dancing", "Live Music"},
		districtIdx:   5,
	},
	{
		name: "Berik Suleimenov", email: "berik.suleimenov@seed.local",
		gender: "man", birthDate: "1996-04-27", lookingFor: "dating",
		aboutMe:       "Chef by day. Guitar player by night.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Cooking", "Guitar", "Wine", "Concerts"},
		districtIdx:   6,
	},
	{
		name: "Gulnara Bekova", email: "gulnara.bekova@seed.local",
		gender: "woman", birthDate: "1999-09-14", lookingFor: "friends",
		aboutMe:       "Volunteer at animal shelter. Dogs are life.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Dogs", "Volunteering", "Running", "Yoga"},
		districtIdx:   7,
	},
	{
		name: "Nurzhan Akhmetov", email: "nurzhan.akhmetov@seed.local",
		gender: "man", birthDate: "1994-12-03", lookingFor: "relationship",
		aboutMe:       "Startup founder. Love board games and tech talks.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Startups", "Tech", "Board Games", "Coffee"},
		districtIdx:   8,
	},
	{
		name: "Kamila Rakhimova", email: "kamila.rakhimova@seed.local",
		gender: "woman", birthDate: "2002-06-18", lookingFor: "dating",
		aboutMe:       "Medeu regular. Love skiing and snowboarding.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Skiing", "Snowboarding", "Hiking", "Photography"},
		districtIdx:   9,
	},
	{
		name: "Askar Nurmagambetov", email: "askar.nurmagambetov@seed.local",
		gender: "man", birthDate: "1991-02-25", lookingFor: "friends",
		aboutMe:       "Road tripper. I've driven every highway in Kazakhstan.",
		interestedIn:  []string{"woman", "man"},
		interestNames: []string{"Road Trips", "Travel", "Camping", "Cycling"},
		districtIdx:   10,
	},
	{
		name: "Zhuldyz Smagulova", email: "zhuldyz.smagulova@seed.local",
		gender: "woman", birthDate: "1998-10-07", lookingFor: "relationship",
		aboutMe:       "Fashion designer. Cats over everything.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Fashion", "Cats", "Baking", "Movies"},
		districtIdx:   11,
	},
	{
		name: "Miras Dzhaksybekov", email: "miras.dzhaks@seed.local",
		gender: "man", birthDate: "1997-07-31", lookingFor: "dating",
		aboutMe:       "Pilates instructor. Meditation keeps me sane.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Pilates", "Yoga", "Meditation", "Gym"},
		districtIdx:   12,
	},
	{
		name: "Samal Aitzhanova", email: "samal.aitzhanova@seed.local",
		gender: "woman", birthDate: "2000-03-22", lookingFor: "relationship",
		aboutMe:       "Singer at local jazz bar. Writing my first album.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Singing", "Live Music", "Writing", "Dancing"},
		districtIdx:   13,
	},
	{
		name: "Talgat Abenov", email: "talgat.abenov@seed.local",
		gender: "man", birthDate: "1992-11-11", lookingFor: "friends",
		aboutMe:       "Surfer (Kapchagai lake). Tennis player.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Surfing", "Tennis", "Swimming", "Running"},
		districtIdx:   14,
	},
	{
		name: "Diana Kozhanova", email: "diana.kozhanova@seed.local",
		gender: "woman", birthDate: "1999-08-04", lookingFor: "dating",
		aboutMe:       "Gamer and board game enthusiast. Let's play.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Gaming", "Board Games", "Movies", "Tech"},
		districtIdx:   15,
	},
	{
		name: "Eldar Zhaksylykov", email: "eldar.zhaksylykov@seed.local",
		gender: "man", birthDate: "1996-05-16", lookingFor: "relationship",
		aboutMe:       "Photographer and traveller. Always looking for the perfect shot.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Photography", "Travel", "Hiking", "Coffee"},
		districtIdx:   16,
	},
	{
		name: "Nazgul Bekmukhambetova", email: "nazgul.bekmukhambetova@seed.local",
		gender: "woman", birthDate: "1995-02-14", lookingFor: "relationship",
		aboutMe:       "Wine sommelier. Weekend brunch is sacred.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Wine", "Brunch", "Cooking", "Travel"},
		districtIdx:   17,
	},
	{
		name: "Amir Baizhanov", email: "amir.baizhanov@seed.local",
		gender: "man", birthDate: "2001-09-09", lookingFor: "dating",
		aboutMe:       "Cyclist and runner. Training for a marathon.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Cycling", "Running", "Gym", "Camping"},
		districtIdx:   18,
	},
	{
		name: "Aliya Seitenova", email: "aliya.seitenova@seed.local",
		gender: "woman", birthDate: "1993-06-30", lookingFor: "relationship",
		aboutMe:       "Bakery owner. My croissants will change your life.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Baking", "Coffee", "Reading", "Meditation"},
		districtIdx:   19,
	},
}

func main() {
	n := flag.Int("n", len(testUsers), "number of users to seed (max 20)")
	password := flag.String("password", "Password123!", "plain-text password for all seeded users")
	flag.Parse()

	if *n > len(testUsers) {
		*n = len(testUsers)
	}

	ctx := context.Background()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// ── Postgres ─────────────────────────────────────────────────────────────
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		configs.Config.DB.Postgres.Username,
		configs.Config.DB.Postgres.Password,
		configs.Config.DB.Postgres.Host,
		configs.Config.DB.Postgres.Port,
		configs.Config.DB.Postgres.Database,
		configs.Config.DB.Postgres.SSLMode,
	)
	db, err := sql.Open("postgres", dsn)
	must(err, "open postgres")
	defer db.Close()
	must(db.PingContext(ctx), "ping postgres")
	log.Println("connected to postgres")

	// ── MongoDB ───────────────────────────────────────────────────────────────
	mongoClient, err := mongodriver.Connect(
		options.Client().ApplyURI(configs.Config.DB.Mongo.URI),
	)
	must(err, "connect mongodb")
	defer mongoClient.Disconnect(ctx)
	must(mongoClient.Ping(ctx, nil), "ping mongodb")
	log.Println("connected to mongodb")
	mongoDB := mongoClient.Database(configs.Config.DB.Mongo.Database)
	profilesColl := mongoDB.Collection("profiles")

	// ── bcrypt password ───────────────────────────────────────────────────────
	hashed, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	must(err, "bcrypt")
	hashedStr := string(hashed)

	// ── Load interest name→id map from DB ────────────────────────────────────
	interestIDMap := loadInterestIDs(ctx, db)

	seeded := 0
	for i := 0; i < *n; i++ {
		u := testUsers[i]
		loc := almatyLocations[u.districtIdx]

		// 1. Insert user ───────────────────────────────────────────────────────
		var userID int64
		err := db.QueryRowContext(ctx, `
			INSERT INTO users (email, password)
			VALUES ($1, $2)
			ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
			RETURNING id`,
			u.email, hashedStr,
		).Scan(&userID)
		if err != nil {
			log.Printf("  SKIP user %s: %v", u.email, err)
			continue
		}
		log.Printf("[%d/%d] user id=%d  %s (%s)", i+1, *n, userID, u.name, u.email)

		// 2. Insert profile (Postgres) ─────────────────────────────────────────
		// Add a small jitter (±0.005°, ~500 m) so users don't stack exactly.
		jLat := loc.lat + (rng.Float64()-0.5)*0.01
		jLng := loc.lng + (rng.Float64()-0.5)*0.01
		_, err = db.ExecContext(ctx, `
			INSERT INTO profiles (user_id, email, name, city, country, current_location)
			VALUES ($1, $2, $3, $4, $5, ST_SetSRID(ST_MakePoint($6, $7), 4326)::geography)
			ON CONFLICT (user_id) DO UPDATE
			    SET name             = EXCLUDED.name,
			        city             = EXCLUDED.city,
			        country          = EXCLUDED.country,
			        current_location = EXCLUDED.current_location,
			        updated_at       = NOW()`,
			userID, u.email, u.name, "Almaty", "Kazakhstan", jLng, jLat,
		)
		if err != nil {
			log.Printf("  WARN profile %d: %v", userID, err)
		}

		// 3. Insert profile photo (placeholder) ───────────────────────────────
		photoURL := fmt.Sprintf("https://randomuser.me/api/portraits/%s/%d.jpg",
			genderDir(u.gender), (userID%70)+1)
		_, err = db.ExecContext(ctx, `
			INSERT INTO profile_photos
			    (user_id, email, provider, bucket, object_key, url, mime_type, size_bytes, width, height)
			VALUES ($1,$2,'placeholder','seed-bucket',$3,$4,'image/jpeg',102400,400,400)
			ON CONFLICT (user_id) DO UPDATE
			    SET url        = EXCLUDED.url,
			        updated_at = NOW()`,
			userID, u.email,
			fmt.Sprintf("seed/%d/avatar.jpg", userID),
			photoURL,
		)
		if err != nil {
			log.Printf("  WARN photo %d: %v", userID, err)
		}

		// 4. Insert matching preferences ──────────────────────────────────────
		_, err = db.ExecContext(ctx, `
			INSERT INTO matching_preferences
			    (user_id, min_age, max_age, max_distance_km, interested_in, show_me_global, only_show_verified)
			VALUES ($1, $2, $3, $4, $5::varchar(50)[], $6, $7)
			ON CONFLICT (user_id) DO UPDATE
			    SET min_age        = EXCLUDED.min_age,
			        max_age        = EXCLUDED.max_age,
			        max_distance_km= EXCLUDED.max_distance_km,
			        interested_in  = EXCLUDED.interested_in,
			        show_me_global = EXCLUDED.show_me_global,
			        updated_at     = NOW()`,
			userID, 20, 40, 30,
			"{"+joinStrings(u.interestedIn)+"}",
			false, false,
		)
		if err != nil {
			log.Printf("  WARN prefs %d: %v", userID, err)
		}

		// 5. Insert user interests ─────────────────────────────────────────────
		for _, name := range u.interestNames {
			id, ok := interestIDMap[name]
			if !ok {
				continue
			}
			_, _ = db.ExecContext(ctx, `
				INSERT INTO user_interests (user_id, interest_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING`,
				userID, id,
			)
		}

		// 6. Insert profile meta in MongoDB ───────────────────────────────────
		filter := bson.M{"user_id": userID}
		update := bson.M{"$set": bson.M{
			"user_id":       userID,
			"gender":        u.gender,
			"birth_date":    u.birthDate,
			"interested_in": u.interestedIn,
			"looking_for":   u.lookingFor,
			"about_me":      u.aboutMe,
			"updated_at":    time.Now(),
		}}
		_, err = profilesColl.UpdateOne(ctx, filter, update,
			options.UpdateOne().SetUpsert(true))
		if err != nil {
			log.Printf("  WARN mongo %d: %v", userID, err)
		}

		seeded++
	}

	fmt.Printf("\n✓ Seeded %d users. Password for all accounts: %s\n", seeded, *password)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func must(err error, label string) {
	if err != nil {
		log.Fatalf("FATAL %s: %v", label, err)
	}
}

func genderDir(gender string) string {
	if gender == "woman" {
		return "women"
	}
	return "men"
}

func joinStrings(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}

func loadInterestIDs(ctx context.Context, db *sql.DB) map[string]int64 {
	rows, err := db.QueryContext(ctx, `SELECT id, name FROM interests`)
	if err != nil {
		log.Printf("WARN: could not load interests: %v", err)
		return nil
	}
	defer rows.Close()

	m := make(map[string]int64)
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err == nil {
			m[name] = id
		}
	}
	log.Printf("loaded %d interests from DB", len(m))
	return m
}
