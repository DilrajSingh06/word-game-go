package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"wordgame/game"
)

var (
	pool *pgxpool.Pool
	rdb  *redis.Client
	ctx  = context.Background()
)

func today() string { return time.Now().UTC().Format("2006-01-02") }

// secondsUntilMidnight is how long today's word and any in-progress game
// stay valid in Redis. Everything resets by itself at midnight.
func secondsUntilMidnight() time.Duration {
	now := time.Now().UTC()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return midnight.Sub(now)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// wordOfTheDay picks a stable word for today, caching it in Redis so every
// player gets the same one.
func wordOfTheDay() (string, error) {
	key := "word:" + today()
	if v, err := rdb.Get(ctx, key).Result(); err == nil && v != "" {
		return v, nil
	}
	var word string
	// Deterministic for a given date, so a restart does not change the answer.
	err := pool.QueryRow(ctx,
		`SELECT word FROM words ORDER BY md5(word || $1) LIMIT 1`, today(),
	).Scan(&word)
	if err != nil {
		return "", err
	}
	rdb.Set(ctx, key, word, secondsUntilMidnight())
	return word, nil
}

func allowedWords() (map[string]bool, error) {
	rows, err := pool.Query(ctx, `SELECT word FROM words`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var w string
		if err := rows.Scan(&w); err != nil {
			return nil, err
		}
		out[strings.ToUpper(w)] = true
	}
	return out, rows.Err()
}

type guessState struct {
	Guesses []string      `json:"guesses"`
	Marks   [][]game.Mark `json:"marks"`
	Done    bool          `json:"done"`
	Won     bool          `json:"won"`
}

func stateKey(player string) string { return "game:" + today() + ":" + player }

func loadState(player string) guessState {
	var st guessState
	if raw, err := rdb.Get(ctx, stateKey(player)).Result(); err == nil {
		_ = json.Unmarshal([]byte(raw), &st)
	}
	return st
}

func saveState(player string, st guessState) {
	b, _ := json.Marshal(st)
	rdb.Set(ctx, stateKey(player), b, secondsUntilMidnight())
}

func main() {
	var err error
	pool, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("redis url: %v", err)
	}
	rdb = redis.NewClient(opt)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		out := map[string]any{"status": "ok", "postgres": false, "redis": false}
		if err := pool.Ping(ctx); err == nil {
			out["postgres"] = true
		}
		if err := rdb.Ping(ctx).Err(); err == nil {
			out["redis"] = true
		}
		code := http.StatusOK
		if out["postgres"] != true || out["redis"] != true {
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, out)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, 200, map[string]any{
			"service":   "word-game",
			"today":     today(),
			"endpoints": []string{"GET /health", "GET /game?player=NAME", "POST /guess"},
		})
	})

	// Current state of today's game for a player.
	mux.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		player := r.URL.Query().Get("player")
		if player == "" {
			writeJSON(w, 400, map[string]string{"error": "player is required"})
			return
		}
		st := loadState(player)
		writeJSON(w, 200, map[string]any{
			"today": today(), "player": player,
			"guesses": st.Guesses, "marks": st.Marks,
			"done": st.Done, "won": st.Won,
			"remaining": game.MaxGuesses - len(st.Guesses),
		})
	})

	mux.HandleFunc("/guess", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "use POST"})
			return
		}
		var body struct {
			Player string `json:"player"`
			Guess  string `json:"guess"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Player == "" {
			writeJSON(w, 400, map[string]string{"error": "player and guess are required"})
			return
		}

		allowed, err := allowedWords()
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		if err := game.ValidateGuess(body.Guess, allowed); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}

		st := loadState(body.Player)
		if st.Done {
			writeJSON(w, 409, map[string]string{"error": game.ErrGameOver.Error()})
			return
		}
		if len(st.Guesses) >= game.MaxGuesses {
			writeJSON(w, 409, map[string]string{"error": game.ErrOutOfTries.Error()})
			return
		}

		answer, err := wordOfTheDay()
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}

		g := strings.ToUpper(strings.TrimSpace(body.Guess))
		marks := game.Score(g, answer)
		st.Guesses = append(st.Guesses, g)
		st.Marks = append(st.Marks, marks)
		st.Won = game.IsWin(marks)
		st.Done = st.Won || len(st.Guesses) >= game.MaxGuesses
		saveState(body.Player, st)

		resp := map[string]any{
			"guess": g, "marks": marks, "won": st.Won, "done": st.Done,
			"remaining": game.MaxGuesses - len(st.Guesses),
		}
		if st.Done {
			resp["answer"] = answer
			_, _ = pool.Exec(ctx,
				`INSERT INTO results (day, player, guesses, won) VALUES ($1,$2,$3,$4)
				 ON CONFLICT (day, player) DO NOTHING`,
				today(), body.Player, len(st.Guesses), st.Won)
		}
		writeJSON(w, 200, resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf("0.0.0.0:%s", port) // 0.0.0.0 so the container is reachable
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
