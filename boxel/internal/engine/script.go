package engine

import "strings"

type ToolboxItem struct {
	Kind  CodeKind
	Label string
	Hat   bool
}

var Toolbox = []ToolboxItem{
	{HatStart, "gdy gra się zaczyna", true},
	{HatCoin, "gdy zbierze monetę", true},
	{HatTouch, "gdy dotknie klocka", true},
	{HatGoal, "gdy dotrze do mety", true},
	{HatDeath, "gdy straci życie", true},
	{HatAlways, "zawsze", true},
	{AddScore, "dodaj punkty", false},
	{SetMessage, "ustaw wiadomość", false},
	{Sound, "zagraj dźwięk", false},
	{Teleport, "teleportuj na start", false},
	{Win, "wygraj", false},
	{Lose, "przegrana", false},
	{Wait, "poczekaj sekundy", false},
	{IfScore, "jeśli punkty ≥", false},
	{Repeat, "powtórz razy", false},
}

var SoundCycle = []SoundName{SndCoin, SndJump, SndWin, SndBump, SndLose, SndPop}

func SoundLabel(s SoundName) string {
	switch s {
	case SndCoin:
		return "moneta"
	case SndJump:
		return "skok"
	case SndWin:
		return "fanfara"
	case SndLose:
		return "buuu"
	case SndBump:
		return "buch"
	default:
		return "pop"
	}
}

func IsHat(k CodeKind) bool {
	switch k {
	case HatStart, HatCoin, HatTouch, HatGoal, HatDeath, HatAlways:
		return true
	}
	return false
}

func CanAdjust(k CodeKind) bool {
	switch k {
	case AddScore, IfScore, Repeat, Wait, HatTouch, Sound:
		return true
	}
	return false
}

func NodeCategory(kind CodeKind) string {
	switch kind {
	case HatStart, HatCoin, HatTouch, HatGoal, HatDeath, HatAlways:
		return "event"
	case SetMessage:
		return "looks"
	case Sound:
		return "sound"
	case Teleport:
		return "motion"
	case AddScore:
		return "sensing"
	default:
		return "control"
	}
}

func NodeLabel(n CodeNode) string {
	switch n.Kind {
	case HatStart:
		return "gdy gra się zaczyna"
	case HatCoin:
		return "gdy zbierze monetę"
	case HatTouch:
		name := Def(n.Block).Name
		if n.Block == 0 {
			name = Def(int(Lava)).Name
		}
		return "gdy dotknie: " + name
	case HatGoal:
		return "gdy dotrze do mety"
	case HatDeath:
		return "gdy straci życie"
	case HatAlways:
		return "zawsze"
	case AddScore:
		a := n.Amount
		if a == 0 {
			a = 10
		}
		return "dodaj " + itoa(a) + " pkt"
	case SetMessage:
		t := n.Text
		if t == "" {
			t = "…"
		}
		return `wiadomość: "` + t + `"`
	case Sound:
		s := n.Sound
		if s == "" {
			s = SndPop
		}
		return "zagraj " + SoundLabel(s)
	case Teleport:
		return "teleportuj na start"
	case Win:
		return "wygraj"
	case Lose:
		return "przegrana"
	case Wait:
		sec := n.Seconds
		if sec == 0 {
			sec = 1
		}
		return "poczekaj " + ftoa(sec) + " s"
	case IfScore:
		return "jeśli punkty ≥ " + itoa(n.Gte)
	case Repeat:
		t := n.Times
		if t == 0 {
			t = 3
		}
		return "powtórz " + itoa(t) + " razy"
	default:
		return string(n.Kind)
	}
}

func NewFromToolbox(kind CodeKind) CodeNode {
	n := MakeNode(kind, nil)
	switch kind {
	case HatTouch:
		n.Block = int(Lava)
	case AddScore:
		n.Amount = 10
	case SetMessage:
		n.Text = "Cześć!"
	case Sound:
		n.Sound = SndCoin
	case Wait:
		n.Seconds = 1
	case IfScore:
		n.Gte = 30
	case Repeat:
		n.Times = 3
	}
	return n
}

func AdjustNode(n *CodeNode, dir int) {
	if dir == 0 {
		return
	}
	switch n.Kind {
	case AddScore:
		if n.Amount == 0 {
			n.Amount = 10
		}
		n.Amount += dir
		if n.Amount < 1 {
			n.Amount = 1
		}
		if n.Amount > 100 {
			n.Amount = 100
		}
	case IfScore:
		n.Gte += dir * 5
		if n.Gte < 0 {
			n.Gte = 0
		}
		if n.Gte > 999 {
			n.Gte = 999
		}
	case Repeat:
		if n.Times == 0 {
			n.Times = 3
		}
		n.Times += dir
		if n.Times < 1 {
			n.Times = 1
		}
		if n.Times > 20 {
			n.Times = 20
		}
	case Wait:
		if n.Seconds == 0 {
			n.Seconds = 1
		}
		n.Seconds += float64(dir) * 0.5
		if n.Seconds < 0.5 {
			n.Seconds = 0.5
		}
		if n.Seconds > 5 {
			n.Seconds = 5
		}
	case HatTouch:
		n.Block += dir
		if n.Block < 1 {
			n.Block = 15
		}
		if n.Block > 15 {
			n.Block = 1
		}
	case Sound:
		idx := 0
		for i, s := range SoundCycle {
			if s == n.Sound {
				idx = i
				break
			}
		}
		idx = (idx + dir + len(SoundCycle)) % len(SoundCycle)
		n.Sound = SoundCycle[idx]
	}
}

func SanitizeScripts(in []ScriptStack) []ScriptStack {
	if len(in) > 24 {
		in = in[:24]
	}
	out := make([]ScriptStack, 0, len(in))
	for _, st := range in {
		blocks := sanitizeNodes(st.Blocks, true, 0)
		if len(blocks) == 0 {
			continue
		}
		id := st.ID
		if id == "" {
			id = NewID()
		}
		out = append(out, ScriptStack{ID: id, Blocks: blocks})
	}
	return out
}

func sanitizeNodes(nodes []CodeNode, top bool, depth int) []CodeNode {
	if depth > 4 || len(nodes) == 0 {
		return nil
	}
	if len(nodes) > 40 {
		nodes = nodes[:40]
	}
	var out []CodeNode
	for i, n := range nodes {
		if !knownKind(n.Kind) {
			continue
		}
		hat := strings.HasPrefix(string(n.Kind), "hat")
		if hat && (!top || i != 0) {
			continue
		}
		if !hat && top && i == 0 && len(out) == 0 {
			// action without a hat — keep, editor may add hat later
		}
		if n.ID == "" {
			n.ID = NewID()
		}
		if n.Kind == SetMessage {
			n.Text = sanitizeScriptText(n.Text)
		}
		if n.Kind == AddScore {
			if n.Amount < 1 {
				n.Amount = 1
			}
			if n.Amount > 100 {
				n.Amount = 100
			}
		}
		if n.Kind == Wait {
			if n.Seconds <= 0 {
				n.Seconds = 1
			}
			if n.Seconds > 5 {
				n.Seconds = 5
			}
		}
		if n.Kind == Repeat {
			if n.Times < 1 {
				n.Times = 1
			}
			if n.Times > 20 {
				n.Times = 20
			}
		}
		if n.Kind == HatTouch && (n.Block < 1 || n.Block > 15) {
			n.Block = int(Lava)
		}
		if n.Kind == IfScore || n.Kind == Repeat {
			n.Body = sanitizeNodes(n.Body, false, depth+1)
		} else {
			n.Body = nil
		}
		out = append(out, n)
	}
	return out
}

func knownKind(k CodeKind) bool {
	switch k {
	case HatStart, HatCoin, HatTouch, HatGoal, HatDeath, HatAlways,
		AddScore, SetMessage, Win, Lose, Teleport, Sound, Wait, IfScore, Repeat:
		return true
	}
	return false
}

func sanitizeScriptText(raw string) string {
	text := strings.Join(strings.Fields(raw), " ")
	runes := []rune(text)
	if len(runes) > 48 {
		text = string(runes[:48])
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "http") || strings.Contains(lower, "discord") || strings.Contains(lower, "www.") {
		return "Cześć!"
	}
	return text
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func ftoa(v float64) string {
	n := int(v*10 + 0.5)
	return itoa(n/10) + "." + itoa(n%10)
}

func RunScripts(scripts []ScriptStack, event ScriptEvent, api GameAPI) {
	vm := &ScriptVM{}
	vm.SetScripts(scripts)
	vm.Emit(event, api)
	vm.Tick(0.05, api)
}
