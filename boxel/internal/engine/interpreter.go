package engine

const (
	maxOpsPerTick = 80
	maxFibers     = 16
	maxRepeat     = 20
	maxWait       = 5
)

type ScriptEvent struct {
	Kind  string
	Block int
}

type frame struct {
	blocks    []CodeNode
	index     int
	remaining *int
}

type fiber struct {
	stack      []frame
	wait       float64
	fromAlways bool
}

type ScriptVM struct {
	Scripts    []ScriptStack
	fibers     []fiber
	alwaysBusy bool
}

func (vm *ScriptVM) Reset() {
	vm.fibers = nil
	vm.alwaysBusy = false
}

func (vm *ScriptVM) SetScripts(scripts []ScriptStack) {
	vm.Scripts = scripts
}

func hatMatches(node *CodeNode, event ScriptEvent) bool {
	if node == nil {
		return false
	}
	switch event.Kind {
	case "start":
		return node.Kind == HatStart
	case "coin":
		return node.Kind == HatCoin
	case "goal":
		return node.Kind == HatGoal
	case "death":
		return node.Kind == HatDeath
	case "always":
		return node.Kind == HatAlways
	case "touch":
		return node.Kind == HatTouch && node.Block == event.Block
	default:
		return false
	}
}

func (vm *ScriptVM) Emit(event ScriptEvent, api GameAPI) {
	if api.IsStopped() && event.Kind != "start" {
		return
	}
	if len(vm.fibers) >= maxFibers {
		return
	}
	for _, stack := range vm.Scripts {
		if len(stack.Blocks) == 0 {
			continue
		}
		hat := stack.Blocks[0]
		if !hatMatches(&hat, event) {
			continue
		}
		if len(vm.fibers) >= maxFibers {
			return
		}
		vm.fibers = append(vm.fibers, fiber{
			stack:      []frame{{blocks: stack.Blocks, index: 1}},
			fromAlways: event.Kind == "always",
		})
	}
}

func (vm *ScriptVM) Tick(dt float64, api GameAPI) {
	if api.IsStopped() {
		vm.Reset()
		return
	}
	if !vm.alwaysBusy {
		vm.Emit(ScriptEvent{Kind: "always"}, api)
		for _, f := range vm.fibers {
			if f.fromAlways {
				vm.alwaysBusy = true
				break
			}
		}
	}
	ops := 0
	for i := range vm.fibers {
		f := &vm.fibers[i]
		if f.wait > 0 {
			f.wait -= dt
			if f.wait < 0 {
				f.wait = 0
			}
			if f.wait > 0 {
				continue
			}
		}
		for ops < maxOpsPerTick && f.wait <= 0 && len(f.stack) > 0 && !api.IsStopped() {
			ops++
			stepFiber(f, api)
		}
	}
	alive := vm.fibers[:0]
	vm.alwaysBusy = false
	for _, f := range vm.fibers {
		if len(f.stack) > 0 {
			alive = append(alive, f)
			if f.fromAlways {
				vm.alwaysBusy = true
			}
		}
	}
	vm.fibers = alive
}

func stepFiber(f *fiber, api GameAPI) {
	if len(f.stack) == 0 {
		return
	}
	fr := &f.stack[len(f.stack)-1]
	if fr.index >= len(fr.blocks) {
		if fr.remaining != nil && *fr.remaining > 1 {
			*fr.remaining--
			fr.index = 0
			return
		}
		f.stack = f.stack[:len(f.stack)-1]
		return
	}
	node := fr.blocks[fr.index]
	fr.index++
	switch node.Kind {
	case HatStart, HatCoin, HatTouch, HatGoal, HatDeath, HatAlways:
	case AddScore:
		n := node.Amount
		if n == 0 {
			n = 1
		}
		api.AddScore(n)
	case SetMessage:
		api.SetMessage(node.Text)
	case Win:
		api.DoWin()
	case Lose:
		api.DoLose()
	case Teleport:
		api.TeleportSpawn()
	case Sound:
		s := node.Sound
		if s == "" {
			s = SndPop
		}
		api.PlaySound(s)
	case Wait:
		if f.fromAlways {
			break
		}
		sec := node.Seconds
		if sec == 0 {
			sec = 1
		}
		if sec > maxWait {
			sec = maxWait
		}
		if sec < 0 {
			sec = 0
		}
		f.wait = sec
	case IfScore:
		body := node.Body
		if api.GetScore() >= node.Gte && len(body) > 0 {
			f.stack = append(f.stack, frame{blocks: body, index: 0})
		}
	case Repeat:
		times := node.Times
		if times == 0 {
			times = 1
		}
		if times > maxRepeat {
			times = maxRepeat
		}
		if times < 0 {
			times = 0
		}
		if times > 0 && len(node.Body) > 0 {
			rem := new(int)
			*rem = times
			f.stack = append(f.stack, frame{blocks: node.Body, index: 0, remaining: rem})
		}
	}
}
