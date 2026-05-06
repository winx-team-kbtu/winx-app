// cmd/seeder seeds Postgres and MongoDB with test users for local recommendation testing.
// All users are placed at realistic coordinates within Almaty, Kazakhstan.
//
// Usage (from recommendation/):
//
//	go run ./cmd/seeder [-n 100] [-password secret]
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
	{43.2890, 76.9678, "Bostandyk"},
	{43.2045, 76.8723, "Auezov"},
	{43.3234, 76.8934, "Zhetisu"},
	{43.2678, 76.8345, "Almaly"},
	{43.1678, 76.9123, "Gorky Park"},
	{43.2456, 76.7923, "Sary-Arka"},
	{43.2134, 76.9834, "Nauryz Park"},
	{43.3456, 76.9012, "Turksib"},
	{43.1234, 76.8567, "Karasay"},
	{43.2923, 76.7645, "Zhambyl"},
}

type seedUser struct {
	name          string
	email         string
	gender        string
	birthDate     string
	lookingFor    string
	aboutMe       string
	interestedIn  []string
	interestNames []string
	districtIdx   int
}

var testUsers = []seedUser{
	// ── Men ──────────────────────────────────────────────────────────────────
	{
		name: "Arman Bekov", email: "arman.bekov@seed.local",
		gender: "man", birthDate: "1998-03-15", lookingFor: "relationship",
		aboutMe:       "Love mountains and good coffee. Working in tech.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Gym", "Hiking", "Coffee", "Tech"},
		districtIdx:   0,
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
		name: "Yerlan Ospanov", email: "yerlan.ospanov@seed.local",
		gender: "man", birthDate: "1993-08-19", lookingFor: "relationship",
		aboutMe:       "Software engineer. Hiking on weekends.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Tech", "Hiking", "Camping", "Cycling"},
		districtIdx:   4,
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
		name: "Nurzhan Akhmetov", email: "nurzhan.akhmetov@seed.local",
		gender: "man", birthDate: "1994-12-03", lookingFor: "relationship",
		aboutMe:       "Startup founder. Love board games and tech talks.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Startups", "Tech", "Board Games", "Coffee"},
		districtIdx:   8,
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
		name: "Miras Dzhaksybekov", email: "miras.dzhaks@seed.local",
		gender: "man", birthDate: "1997-07-31", lookingFor: "dating",
		aboutMe:       "Pilates instructor. Meditation keeps me sane.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Pilates", "Yoga", "Meditation", "Gym"},
		districtIdx:   12,
	},
	{
		name: "Talgat Abenov", email: "talgat.abenov@seed.local",
		gender: "man", birthDate: "1992-11-11", lookingFor: "friends",
		aboutMe:       "Surfer (Kapchagai lake). Tennis player.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Tennis", "Swimming", "Running", "Cycling"},
		districtIdx:   14,
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
		name: "Amir Baizhanov", email: "amir.baizhanov@seed.local",
		gender: "man", birthDate: "2001-09-09", lookingFor: "dating",
		aboutMe:       "Cyclist and runner. Training for a marathon.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Cycling", "Running", "Gym", "Camping"},
		districtIdx:   18,
	},
	{
		name: "Aidyn Seitenov", email: "aidyn.seitenov@seed.local",
		gender: "man", birthDate: "1999-04-12", lookingFor: "dating",
		aboutMe:       "Basketball coach and gym rat. Looking for someone active.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Basketball", "Gym", "Running", "Cooking"},
		districtIdx:   20,
	},
	{
		name: "Bauyrzhan Kenzhebaev", email: "bauyrzhan.kenzh@seed.local",
		gender: "man", birthDate: "1997-08-23", lookingFor: "relationship",
		aboutMe:       "Architect by profession, artist at heart. I design spaces and feelings.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Painting", "Photography", "Coffee", "Travel"},
		districtIdx:   21,
	},
	{
		name: "Temirlan Dzhunusov", email: "temirlan.dzhun@seed.local",
		gender: "man", birthDate: "2000-01-17", lookingFor: "friends",
		aboutMe:       "University student. Esports player and gym beginner.",
		interestedIn:  []string{"woman", "man"},
		interestNames: []string{"Gaming", "Gym", "Movies", "Tech"},
		districtIdx:   22,
	},
	{
		name: "Zhangir Bekmuratov", email: "zhangir.bekmuratov@seed.local",
		gender: "man", birthDate: "1990-06-05", lookingFor: "relationship",
		aboutMe:       "Entrepreneur and coffee addict. Building the next big thing in Almaty.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Startups", "Coffee", "Board Games", "Wine"},
		districtIdx:   23,
	},
	{
		name: "Alibek Smagambetov", email: "alibek.smag@seed.local",
		gender: "man", birthDate: "1995-03-28", lookingFor: "dating",
		aboutMe:       "Mountaineer. Summited Khan Tengri twice. Looking for a partner in crime.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Hiking", "Camping", "Climbing", "Photography"},
		districtIdx:   24,
	},
	{
		name: "Sanzhar Rakhimov", email: "sanzhar.rakhimov@seed.local",
		gender: "man", birthDate: "1998-10-14", lookingFor: "relationship",
		aboutMe:       "Jazz musician and coffee shop regular. Life is better with good music.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Live Music", "Guitar", "Coffee", "Reading"},
		districtIdx:   25,
	},
	{
		name: "Dias Akhanov", email: "dias.akhanov@seed.local",
		gender: "man", birthDate: "2002-07-03", lookingFor: "friends",
		aboutMe:       "Film student. Making short films about Almaty street culture.",
		interestedIn:  []string{"woman", "man"},
		interestNames: []string{"Movies", "Photography", "Writing", "Coffee"},
		districtIdx:   26,
	},
	{
		name: "Azamat Nurpeisov", email: "azamat.nurp@seed.local",
		gender: "man", birthDate: "1993-12-19", lookingFor: "relationship",
		aboutMe:       "Doctor by day, marathon runner by weekend. Work hard, run harder.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Running", "Gym", "Hiking", "Cooking"},
		districtIdx:   27,
	},
	{
		name: "Dauren Bekturov", email: "dauren.bekturov@seed.local",
		gender: "man", birthDate: "1996-09-08", lookingFor: "dating",
		aboutMe:       "DJ and music producer. Spinning tracks on weekends at local clubs.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Concerts", "Dancing", "Live Music", "Travel"},
		districtIdx:   28,
	},
	{
		name: "Ruslan Satiev", email: "ruslan.satiev@seed.local",
		gender: "man", birthDate: "1991-05-22", lookingFor: "relationship",
		aboutMe:       "Wine importer and foodie. Always chasing the best meal in the city.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Wine", "Cooking", "Travel", "Brunch"},
		districtIdx:   29,
	},
	{
		name: "Erlan Dzhaksybekov", email: "erlan.dzhaksy@seed.local",
		gender: "man", birthDate: "1999-02-14", lookingFor: "dating",
		aboutMe:       "Personal trainer. Obsessed with nutrition science and good sleep.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Gym", "Running", "Cooking", "Meditation"},
		districtIdx:   1,
	},
	{
		name: "Kanat Ospanov", email: "kanat.ospanov@seed.local",
		gender: "man", birthDate: "1997-11-30", lookingFor: "relationship",
		aboutMe:       "Veterinarian. If you have a pet, we already have something to talk about.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Dogs", "Cats", "Hiking", "Coffee"},
		districtIdx:   3,
	},
	{
		name: "Serik Beisenov", email: "serik.beisenov@seed.local",
		gender: "man", birthDate: "1994-07-07", lookingFor: "friends",
		aboutMe:       "Poker player and chess enthusiast. Strategic thinker.",
		interestedIn:  []string{"woman", "man"},
		interestNames: []string{"Board Games", "Coffee", "Beer", "Cycling"},
		districtIdx:   5,
	},
	{
		name: "Timur Khasanov", email: "timur.khasanov@seed.local",
		gender: "man", birthDate: "1989-04-01", lookingFor: "relationship",
		aboutMe:       "Lawyer turned startup investor. Looking for genuine connections.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Startups", "Wine", "Travel", "Reading"},
		districtIdx:   7,
	},
	{
		name: "Adil Zhumabayev", email: "adil.zhumabayev@seed.local",
		gender: "man", birthDate: "2000-08-15", lookingFor: "dating",
		aboutMe:       "Graphic designer. Coffee shop hopper and plant dad.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Coffee", "Photography", "Painting", "Movies"},
		districtIdx:   9,
	},
	{
		name: "Zhandos Bekzhanov", email: "zhandos.bekzhanov@seed.local",
		gender: "man", birthDate: "1998-06-20", lookingFor: "relationship",
		aboutMe:       "Ice hockey player. Winter is my season.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Skiing", "Snowboarding", "Gym", "Beer"},
		districtIdx:   11,
	},
	{
		name: "Sultan Mahmudov", email: "sultan.mahmudov@seed.local",
		gender: "man", birthDate: "1992-03-11", lookingFor: "relationship",
		aboutMe:       "Yoga teacher and meditation guide. Slow living advocate.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Yoga", "Meditation", "Reading", "Hiking"},
		districtIdx:   13,
	},
	{
		name: "Bekzat Abenov", email: "bekzat.abenov@seed.local",
		gender: "man", birthDate: "2003-01-25", lookingFor: "friends",
		aboutMe:       "Just graduated. Figuring out life one day at a time.",
		interestedIn:  []string{"woman", "man"},
		interestNames: []string{"Gaming", "Movies", "Beer", "Football"},
		districtIdx:   15,
	},
	{
		name: "Yerkin Musaev", email: "yerkin.musaev@seed.local",
		gender: "man", birthDate: "1996-12-08", lookingFor: "dating",
		aboutMe:       "Travel blogger. 47 countries down, the rest to go.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Travel", "Photography", "Hiking", "Coffee"},
		districtIdx:   17,
	},
	{
		name: "Nurbek Kenzhebayev", email: "nurbek.kenzhebayev@seed.local",
		gender: "man", birthDate: "1990-09-14", lookingFor: "relationship",
		aboutMe:       "Surgeon. Precision in everything, including my relationships.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Running", "Tennis", "Wine", "Reading"},
		districtIdx:   19,
	},
	{
		name: "Mukhtar Dzhaksanov", email: "mukhtar.dzhaks@seed.local",
		gender: "man", birthDate: "1995-05-03", lookingFor: "dating",
		aboutMe:       "Barista champion 2024. Coffee is my love language.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Coffee", "Cooking", "Baking", "Concerts"},
		districtIdx:   0,
	},
	{
		name: "Darkhan Seitenov", email: "darkhan.seit@seed.local",
		gender: "man", birthDate: "1997-10-30", lookingFor: "relationship",
		aboutMe:       "Cyclist who commutes 30km daily. No car, no problem.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Cycling", "Running", "Hiking", "Camping"},
		districtIdx:   2,
	},
	{
		name: "Ilyas Nurbekov", email: "ilyas.nurbekov@seed.local",
		gender: "man", birthDate: "2001-04-19", lookingFor: "friends",
		aboutMe:       "Music producer making beats in my bedroom studio.",
		interestedIn:  []string{"woman", "man"},
		interestNames: []string{"Live Music", "Guitar", "Concerts", "Gaming"},
		districtIdx:   4,
	},
	{
		name: "Arsen Bekmukhambetov", email: "arsen.bekmukhambetov@seed.local",
		gender: "man", birthDate: "1993-07-22", lookingFor: "relationship",
		aboutMe:       "Business analyst. Weekend warrior on skis.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Skiing", "Snowboarding", "Wine", "Travel"},
		districtIdx:   6,
	},
	{
		name: "Alexandr Petrov", email: "alexandr.petrov@seed.local",
		gender: "man", birthDate: "1988-02-28", lookingFor: "relationship",
		aboutMe:       "Engineer. Fixing things at work and at home.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Tech", "Camping", "Beer", "Football"},
		districtIdx:   8,
	},
	{
		name: "Maxim Volkov", email: "maxim.volkov@seed.local",
		gender: "man", birthDate: "1994-11-05", lookingFor: "dating",
		aboutMe:       "CrossFit coach. I live in the gym but I like restaurants too.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Gym", "Running", "Cooking", "Beer"},
		districtIdx:   10,
	},
	{
		name: "Dmitri Sokolov", email: "dmitri.sokolov@seed.local",
		gender: "man", birthDate: "1986-08-17", lookingFor: "relationship",
		aboutMe:       "Chef and restaurant owner. Food is how I show love.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Cooking", "Wine", "Travel", "Brunch"},
		districtIdx:   12,
	},
	{
		name: "Amir Dzhaksybekov", email: "amir.dzhaksy@seed.local",
		gender: "man", birthDate: "2000-05-11", lookingFor: "dating",
		aboutMe:       "Crypto trader and tech nerd. Always watching screens.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Tech", "Gaming", "Board Games", "Coffee"},
		districtIdx:   14,
	},
	{
		name: "Nursultan Beisenov", email: "nursultan.beis@seed.local",
		gender: "man", birthDate: "1999-12-01", lookingFor: "relationship",
		aboutMe:       "Rock climber. If you're afraid of heights, I'll teach you not to be.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Climbing", "Hiking", "Camping", "Photography"},
		districtIdx:   16,
	},
	{
		name: "Baurzhan Suleimenov", email: "baurzhan.suleim@seed.local",
		gender: "man", birthDate: "1991-09-27", lookingFor: "relationship",
		aboutMe:       "Economist. I can explain inflation but not why I'm still single.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Reading", "Coffee", "Wine", "Board Games"},
		districtIdx:   18,
	},
	{
		name: "Timur Rakhimov", email: "timur.rakhimov@seed.local",
		gender: "man", birthDate: "1996-06-13", lookingFor: "dating",
		aboutMe:       "Volleyball player and beach volleyball enthusiast.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Volleyball", "Swimming", "Beach", "Beer"},
		districtIdx:   20,
	},
	{
		name: "Ruslan Bekzhanov", email: "ruslan.bekzhanov@seed.local",
		gender: "man", birthDate: "1987-03-06", lookingFor: "relationship",
		aboutMe:       "Divorced dad of one. Looking for something real.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Cooking", "Hiking", "Movies", "Dogs"},
		districtIdx:   22,
	},
	{
		name: "Zhandos Akhmetov", email: "zhandos.akhmetov@seed.local",
		gender: "man", birthDate: "2002-10-22", lookingFor: "friends",
		aboutMe:       "Breakdancer. Hit me up if you want to dance battle.",
		interestedIn:  []string{"woman", "man"},
		interestNames: []string{"Dancing", "Concerts", "Live Music", "Gym"},
		districtIdx:   24,
	},
	{
		name: "Aidar Nurpeisov", email: "aidar.nurpeisov@seed.local",
		gender: "man", birthDate: "1995-01-18", lookingFor: "relationship",
		aboutMe:       "Marine biologist who ended up in landlocked Kazakhstan. Still loves the sea.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Travel", "Swimming", "Reading", "Photography"},
		districtIdx:   26,
	},
	{
		name: "Kairan Bekuov", email: "kairan.bekuov@seed.local",
		gender: "man", birthDate: "1998-07-09", lookingFor: "dating",
		aboutMe:       "MMA fighter training for regional championship.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Boxing", "Gym", "Running", "Cooking"},
		districtIdx:   28,
	},
	{
		name: "Daniil Kozlov", email: "daniil.kozlov@seed.local",
		gender: "man", birthDate: "1993-04-30", lookingFor: "relationship",
		aboutMe:       "Urban farmer growing vegetables on my balcony. Sustainable living.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Cooking", "Baking", "Dogs", "Volunteering"},
		districtIdx:   1,
	},
	{
		name: "Marko Djuricic", email: "marko.djuricic@seed.local",
		gender: "man", birthDate: "1990-11-14", lookingFor: "dating",
		aboutMe:       "Serbian expat living in Almaty for 5 years. Still discovering this amazing city.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Football", "Travel", "Coffee", "Concerts"},
		districtIdx:   3,
	},
	{
		name: "Arslan Khassenov", email: "arslan.khassenov@seed.local",
		gender: "man", birthDate: "1997-02-08", lookingFor: "relationship",
		aboutMe:       "Stand-up comedian. I'll make you laugh, guaranteed.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Concerts", "Live Music", "Beer", "Movies"},
		districtIdx:   5,
	},
	{
		name: "Farukh Tashmatov", email: "farukh.tashmatov@seed.local",
		gender: "man", birthDate: "1994-08-21", lookingFor: "relationship",
		aboutMe:       "Uzbek expat, software developer. Love plov and programming equally.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Tech", "Cooking", "Travel", "Football"},
		districtIdx:   7,
	},
	{
		name: "Yerzhan Seitbekov", email: "yerzhan.seitbekov@seed.local",
		gender: "man", birthDate: "1986-06-03", lookingFor: "relationship",
		aboutMe:       "Single father, financial advisor. Stability and adventure in equal measure.",
		interestedIn:  []string{"woman"},
		interestNames: []string{"Hiking", "Cooking", "Reading", "Running"},
		districtIdx:   9,
	},

	// ── Women ─────────────────────────────────────────────────────────────────
	{
		name: "Aizat Nurova", email: "aizat.nurova@seed.local",
		gender: "woman", birthDate: "2000-07-22", lookingFor: "dating",
		aboutMe:       "Yoga instructor. Always up for brunch and travel.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Yoga", "Travel", "Brunch", "Photography"},
		districtIdx:   1,
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
		name: "Ainur Satova", email: "ainur.satova@seed.local",
		gender: "woman", birthDate: "2001-01-05", lookingFor: "dating",
		aboutMe:       "Artist and dreamer. Always painting something.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Painting", "Photography", "Dancing", "Live Music"},
		districtIdx:   5,
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
		name: "Kamila Rakhimova", email: "kamila.rakhimova@seed.local",
		gender: "woman", birthDate: "2002-06-18", lookingFor: "dating",
		aboutMe:       "Medeu regular. Love skiing and snowboarding.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Skiing", "Snowboarding", "Hiking", "Photography"},
		districtIdx:   9,
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
		name: "Samal Aitzhanova", email: "samal.aitzhanova@seed.local",
		gender: "woman", birthDate: "2000-03-22", lookingFor: "relationship",
		aboutMe:       "Singer at local jazz bar. Writing my first album.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Singing", "Live Music", "Writing", "Dancing"},
		districtIdx:   13,
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
		name: "Nazgul Bekmukhambetova", email: "nazgul.bekmukhambetova@seed.local",
		gender: "woman", birthDate: "1995-02-14", lookingFor: "relationship",
		aboutMe:       "Wine sommelier. Weekend brunch is sacred.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Wine", "Brunch", "Cooking", "Travel"},
		districtIdx:   17,
	},
	{
		name: "Aliya Seitenova", email: "aliya.seitenova@seed.local",
		gender: "woman", birthDate: "1993-06-30", lookingFor: "relationship",
		aboutMe:       "Bakery owner. My croissants will change your life.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Baking", "Coffee", "Reading", "Meditation"},
		districtIdx:   19,
	},
	{
		name: "Asel Nurbekova", email: "asel.nurbekova@seed.local",
		gender: "woman", birthDate: "1998-04-18", lookingFor: "relationship",
		aboutMe:       "Pediatrician. I care for kids at work and rescue cats at home.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Cats", "Yoga", "Reading", "Volunteering"},
		districtIdx:   21,
	},
	{
		name: "Meruert Beisenova", email: "meruert.beisenova@seed.local",
		gender: "woman", birthDate: "2001-11-29", lookingFor: "dating",
		aboutMe:       "Architecture student. I see beauty in everything around me.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Photography", "Coffee", "Travel", "Painting"},
		districtIdx:   23,
	},
	{
		name: "Bibigul Akhmetova", email: "bibigul.akhmetova@seed.local",
		gender: "woman", birthDate: "1996-08-05", lookingFor: "relationship",
		aboutMe:       "Lawyer. Arguing is my job. Dating is my hobby.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Wine", "Reading", "Running", "Travel"},
		districtIdx:   25,
	},
	{
		name: "Sandugash Rakhimova", email: "sandugash.rakhimova@seed.local",
		gender: "woman", birthDate: "1994-03-14", lookingFor: "friends",
		aboutMe:       "Dance instructor. Salsa and cumbia are my therapy.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Dancing", "Live Music", "Concerts", "Yoga"},
		districtIdx:   27,
	},
	{
		name: "Zarina Bekova", email: "zarina.bekova@seed.local",
		gender: "woman", birthDate: "2003-05-20", lookingFor: "friends",
		aboutMe:       "Marketing student. Social media manager for a local brand.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Photography", "Coffee", "Fashion", "Brunch"},
		districtIdx:   29,
	},
	{
		name: "Dana Seitkali", email: "dana.seitkali@seed.local",
		gender: "woman", birthDate: "1997-01-31", lookingFor: "relationship",
		aboutMe:       "Nutritionist. I'll judge your food choices, but nicely.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Cooking", "Running", "Yoga", "Hiking"},
		districtIdx:   0,
	},
	{
		name: "Aigul Smagulova", email: "aigul.smagulova@seed.local",
		gender: "woman", birthDate: "1992-09-08", lookingFor: "relationship",
		aboutMe:       "Primary school teacher. Patience is my superpower.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Reading", "Baking", "Cats", "Movies"},
		districtIdx:   2,
	},
	{
		name: "Symbat Dzhaksybekov", email: "symbat.dzhaksy@seed.local",
		gender: "woman", birthDate: "2000-12-15", lookingFor: "dating",
		aboutMe:       "Swimmer and freediver. I'm more comfortable underwater.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Swimming", "Yoga", "Travel", "Photography"},
		districtIdx:   4,
	},
	{
		name: "Altynai Bekzhanova", email: "altynai.bekzhanova@seed.local",
		gender: "woman", birthDate: "1995-06-27", lookingFor: "relationship",
		aboutMe:       "Journalist. I ask good questions. You'll see.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Writing", "Coffee", "Travel", "Reading"},
		districtIdx:   6,
	},
	{
		name: "Akmaral Nurpeisova", email: "akmaral.nurp@seed.local",
		gender: "woman", birthDate: "1999-03-03", lookingFor: "dating",
		aboutMe:       "Pastry chef. I speak fluent croissant and tiramisu.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Baking", "Cooking", "Coffee", "Wine"},
		districtIdx:   8,
	},
	{
		name: "Moldir Akhanova", email: "moldir.akhanova@seed.local",
		gender: "woman", birthDate: "2001-07-14", lookingFor: "friends",
		aboutMe:       "Environmental activist. Planet first, everything else second.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Hiking", "Volunteering", "Running", "Camping"},
		districtIdx:   10,
	},
	{
		name: "Anel Bekuova", email: "anel.bekuova@seed.local",
		gender: "woman", birthDate: "1996-10-25", lookingFor: "relationship",
		aboutMe:       "Interior designer. My home looks like a Pinterest board.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Coffee", "Painting", "Travel", "Brunch"},
		districtIdx:   12,
	},
	{
		name: "Botakoz Seitenova", email: "botakoz.seiten@seed.local",
		gender: "woman", birthDate: "1993-01-19", lookingFor: "relationship",
		aboutMe:       "Gynaecologist. I know more about life than most.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Yoga", "Running", "Reading", "Wine"},
		districtIdx:   14,
	},
	{
		name: "Madina Rakhimova", email: "madina.rakhimova@seed.local",
		gender: "woman", birthDate: "2002-04-07", lookingFor: "dating",
		aboutMe:       "Aspiring singer. Open mics on Thursdays, DMs always open.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Singing", "Live Music", "Dancing", "Concerts"},
		districtIdx:   16,
	},
	{
		name: "Sholpan Abenova", email: "sholpan.abenova@seed.local",
		gender: "woman", birthDate: "1997-08-12", lookingFor: "relationship",
		aboutMe:       "Equestrian. I ride horses and I'm great at it.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Hiking", "Camping", "Photography", "Travel"},
		districtIdx:   18,
	},
	{
		name: "Aigerim Kenzhebaeva", email: "aigerim.kenzh@seed.local",
		gender: "woman", birthDate: "1999-11-22", lookingFor: "dating",
		aboutMe:       "HR manager. I screen CVs at work and personalities in life.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Coffee", "Brunch", "Movies", "Reading"},
		districtIdx:   20,
	},
	{
		name: "Kamshat Bekmuratova", email: "kamshat.bekmuratova@seed.local",
		gender: "woman", birthDate: "1991-05-16", lookingFor: "relationship",
		aboutMe:       "Entrepreneur. Running three businesses and somehow still dating.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Startups", "Coffee", "Wine", "Travel"},
		districtIdx:   22,
	},
	{
		name: "Nurgul Smagambetova", email: "nurgul.smag@seed.local",
		gender: "woman", birthDate: "1994-07-04", lookingFor: "friends",
		aboutMe:       "Pilates instructor and morning person. 5am club, anyone?",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Pilates", "Running", "Yoga", "Brunch"},
		districtIdx:   24,
	},
	{
		name: "Dina Ospanova", email: "dina.ospanova@seed.local",
		gender: "woman", birthDate: "1998-02-28", lookingFor: "relationship",
		aboutMe:       "Translator. Fluent in Kazakh, Russian, English and sarcasm.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Reading", "Travel", "Coffee", "Movies"},
		districtIdx:   26,
	},
	{
		name: "Gaukhar Bekzhanova", email: "gaukhar.bekzhanova@seed.local",
		gender: "woman", birthDate: "1990-09-01", lookingFor: "relationship",
		aboutMe:       "Bank director. Serious at work, spontaneous everywhere else.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Wine", "Travel", "Yoga", "Reading"},
		districtIdx:   28,
	},
	{
		name: "Raushan Akhmetova", email: "raushan.akhmetova@seed.local",
		gender: "woman", birthDate: "2000-06-10", lookingFor: "dating",
		aboutMe:       "Makeup artist. I make people look their best, including myself.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Fashion", "Dancing", "Photography", "Brunch"},
		districtIdx:   1,
	},
	{
		name: "Aizhan Satova", email: "aizhan.satova@seed.local",
		gender: "woman", birthDate: "1996-12-21", lookingFor: "relationship",
		aboutMe:       "Mountain rescue volunteer. I've saved lives, I can save the date too.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Hiking", "Climbing", "Camping", "Running"},
		districtIdx:   3,
	},
	{
		name: "Zhansaya Musaeva", email: "zhansaya.musaeva@seed.local",
		gender: "woman", birthDate: "2003-08-30", lookingFor: "friends",
		aboutMe:       "First year uni student. Figuring everything out one coffee at a time.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Coffee", "Movies", "Gaming", "Reading"},
		districtIdx:   5,
	},
	{
		name: "Olga Ivanova", email: "olga.ivanova@seed.local",
		gender: "woman", birthDate: "1988-04-14", lookingFor: "relationship",
		aboutMe:       "Russian literature professor. Dostoyevsky is my comfort read.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Reading", "Wine", "Coffee", "Concerts"},
		districtIdx:   7,
	},
	{
		name: "Anastasia Volkova", email: "anastasia.volkova@seed.local",
		gender: "woman", birthDate: "1995-10-09", lookingFor: "dating",
		aboutMe:       "Ballet dancer. Grace is not just for the stage.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Dancing", "Yoga", "Live Music", "Photography"},
		districtIdx:   9,
	},
	{
		name: "Sofiya Petrova", email: "sofiya.petrova@seed.local",
		gender: "woman", birthDate: "2001-02-17", lookingFor: "friends",
		aboutMe:       "Pharmacy student and part-time barista. Dispensing good vibes.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Coffee", "Baking", "Yoga", "Movies"},
		districtIdx:   11,
	},
	{
		name: "Malika Seitbekov", email: "malika.seitbekov@seed.local",
		gender: "woman", birthDate: "1997-05-27", lookingFor: "relationship",
		aboutMe:       "Diplomat. I've lived on four continents. Almaty is still my favourite.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Travel", "Wine", "Reading", "Photography"},
		districtIdx:   13,
	},
	{
		name: "Farida Tashmatova", email: "farida.tashmatova@seed.local",
		gender: "woman", birthDate: "1993-11-04", lookingFor: "relationship",
		aboutMe:       "Uzbek expat. I cook the best lagman in Kazakhstan. Come find out.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Cooking", "Travel", "Cats", "Coffee"},
		districtIdx:   15,
	},
	{
		name: "Zulfiya Khasanova", email: "zulfiya.khasanova@seed.local",
		gender: "woman", birthDate: "1999-07-19", lookingFor: "dating",
		aboutMe:       "Graphic novelist. Drawing stories about Almaty one page at a time.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Painting", "Writing", "Coffee", "Movies"},
		districtIdx:   17,
	},
	{
		name: "Leila Dzhaksybekova", email: "leila.dzhaksy@seed.local",
		gender: "woman", birthDate: "1990-03-23", lookingFor: "relationship",
		aboutMe:       "Cardiologist. I listen to hearts professionally.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Running", "Yoga", "Wine", "Reading"},
		districtIdx:   19,
	},
	{
		name: "Aknur Beisenova", email: "aknur.beisenova@seed.local",
		gender: "woman", birthDate: "2002-01-13", lookingFor: "friends",
		aboutMe:       "Streamer and content creator. My cat has more followers than me.",
		interestedIn:  []string{"man", "woman"},
		interestNames: []string{"Gaming", "Cats", "Movies", "Coffee"},
		districtIdx:   21,
	},
	{
		name: "Aiaru Bekuova", email: "aiaru.bekuova@seed.local",
		gender: "woman", birthDate: "1994-09-07", lookingFor: "relationship",
		aboutMe:       "Mountaineer. I do what scares me. Dating included.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Hiking", "Climbing", "Camping", "Photography"},
		districtIdx:   23,
	},
	{
		name: "Tolkyn Seitenova", email: "tolkyn.seitenova@seed.local",
		gender: "woman", birthDate: "1998-06-16", lookingFor: "relationship",
		aboutMe:       "Music therapist. Sound heals everything.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Live Music", "Yoga", "Meditation", "Coffee"},
		districtIdx:   25,
	},
	{
		name: "Aruzhan Bekova", email: "aruzhan.bekova@seed.local",
		gender: "woman", birthDate: "2000-10-02", lookingFor: "dating",
		aboutMe:       "Physical therapist. I fix people, not broken hearts.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Gym", "Pilates", "Running", "Yoga"},
		districtIdx:   27,
	},
	{
		name: "Gulsim Akhmetova", email: "gulsim.akhmetova@seed.local",
		gender: "woman", birthDate: "1991-08-25", lookingFor: "relationship",
		aboutMe:       "Founder of women's hiking club. 200+ members, still growing.",
		interestedIn:  []string{"man"},
		interestNames: []string{"Hiking", "Travel", "Camping", "Photography"},
		districtIdx:   29,
	},
}

func main() {
	n := flag.Int("n", len(testUsers), "number of users to seed (max 100)")
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
			userID, 18, 45, 50,
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
