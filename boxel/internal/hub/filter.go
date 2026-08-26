package hub

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

const maxChat = 60

var (
	urlRe   = regexp.MustCompile(`(?i)(https?://|www\.|discord\.gg|discord\.com/invite|t\.me/|\.ru/|\.onion)`)
	phoneRe = regexp.MustCompile(`\+?\d[\d\s-]{7,}\d`)
	banned  = []string{
		"kurwa", "chuj", "cipa", "fuck", "shit", "nigger", "hitler",
		"pedo", "nazi", "zginiesz", "zabije", "zabiję",
	}
	friendCodeRe = regexp.MustCompile(`^BX-[A-Z0-9]{6}$`)
	palette      = []string{
		"#ff6b3d", "#5ec8ff", "#7cff6b", "#ffd24a",
		"#ff4fa3", "#c45c4a", "#9fe7ff", "#c48a3a",
	}
)

func sanitizeName(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	name := strings.TrimSpace(b.String())
	runes := []rune(name)
	if len(runes) > 16 {
		name = string(runes[:16])
	}
	if len([]rune(name)) < 2 {
		return "Gracz"
	}
	return name
}

func sanitizeChat(raw string) (string, bool) {
	text := strings.Join(strings.Fields(raw), " ")
	if text == "" {
		return "", false
	}
	runes := []rune(text)
	if len(runes) > maxChat {
		text = string(runes[:maxChat])
	}
	if urlRe.MatchString(text) || phoneRe.MatchString(text) {
		return "", false
	}
	lower := strings.ToLower(text)
	for _, w := range banned {
		if strings.Contains(lower, w) {
			return "", false
		}
	}
	return text, true
}

func makeFriendCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	out := []byte("BX-")
	for i := 0; i < 6; i++ {
		out = append(out, alphabet[int(buf[i])%len(alphabet)])
	}
	return string(out)
}

func makeRoomCode() string {
	buf := make([]byte, 2)
	_, _ = rand.Read(buf)
	n := 1000 + (int(buf[0])*256+int(buf[1]))%9000
	return itoa4(n)
}

func itoa4(n int) string {
	b := []byte{'0', '0', '0', '0'}
	for i := 3; i >= 0; i-- {
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b)
}

func makeID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func validFriendCode(code string) bool {
	return friendCodeRe.MatchString(code)
}

func colorForID(id string) string {
	h := uint32(0)
	for i := 0; i < len(id); i++ {
		h = h*31 + uint32(id[i])
	}
	return palette[int(h)%len(palette)]
}

func SanitizeName(raw string) string { return sanitizeName(raw) }

func MakeFriendCode() string { return makeFriendCode() }

var ChatPresets = []string{"Cześć!", "Chodź tu!", "Pomocy!", "Super!", "Czekaj!", "Gramy!"}
