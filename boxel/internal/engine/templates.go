package engine

type Template struct {
	ID, Name, Blurb string
	Build           func() (World, []ScriptStack, string)
}

func parkourScripts() []ScriptStack {
	return []ScriptStack{
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatStart, nil),
			MakeNode(SetMessage, &CodeNode{Text: "Zbierz 3 monety i dobiegnij do złotej flagi!"}),
			MakeNode(Sound, &CodeNode{Sound: SndPop}),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatCoin, nil),
			MakeNode(AddScore, &CodeNode{Amount: 10}),
			MakeNode(Sound, &CodeNode{Sound: SndCoin}),
			MakeNode(SetMessage, &CodeNode{Text: "Moneta! +10"}),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatTouch, &CodeNode{Block: int(Lava)}),
			MakeNode(SetMessage, &CodeNode{Text: "Auć! Lawa parzy."}),
			MakeNode(Sound, &CodeNode{Sound: SndBump}),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatTouch, &CodeNode{Block: int(Spike)}),
			MakeNode(SetMessage, &CodeNode{Text: "Ostrożnie, kolce!"}),
			MakeNode(Sound, &CodeNode{Sound: SndBump}),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatDeath, nil),
			MakeNode(SetMessage, &CodeNode{Text: "Spróbuj jeszcze raz!"}),
			MakeNode(Teleport, nil),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatGoal, nil),
			MakeNode(IfScore, &CodeNode{Gte: 30, Body: []CodeNode{
				MakeNode(Sound, &CodeNode{Sound: SndWin}),
				MakeNode(SetMessage, &CodeNode{Text: "WYGRANA!"}),
				MakeNode(Win, nil),
			}}),
			MakeNode(SetMessage, &CodeNode{Text: "Potrzeba 30 punktów (3 monety)."}),
		}},
	}
}

func buildSandbox() (World, []ScriptStack, string) {
	w := CreateWorld(SizeX, SizeY, SizeZ)
	FillBox(w, 0, 0, 0, 23, 0, 23, int(Grass))
	SetBlock(w, 12, 1, 12, int(Spawn))
	scripts := []ScriptStack{
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatStart, nil),
			MakeNode(SetMessage, &CodeNode{Text: "Buduj klockami, potem naciśnij GRAJ!"}),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatCoin, nil),
			MakeNode(AddScore, &CodeNode{Amount: 10}),
			MakeNode(Sound, &CodeNode{Sound: SndCoin}),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatGoal, nil),
			MakeNode(Win, nil),
		}},
	}
	return w, scripts, "Piaskownica"
}

func buildParkour() (World, []ScriptStack, string) {
	w := CreateWorld(SizeX, SizeY, SizeZ)
	FillBox(w, 0, 0, 0, 23, 0, 23, int(Lava))
	FillBox(w, 1, 1, 1, 5, 1, 5, int(Grass))
	SetBlock(w, 3, 2, 3, int(Spawn))
	FillBox(w, 4, 2, 8, 6, 2, 10, int(Stone))
	SetBlock(w, 5, 4, 9, int(Coin))
	FillBox(w, 8, 3, 12, 11, 3, 14, int(Brick))
	SetBlock(w, 10, 3, 13, int(Bounce))
	SetBlock(w, 9, 5, 13, int(Coin))
	FillBox(w, 12, 4, 14, 18, 4, 15, int(Ice))
	SetBlock(w, 16, 6, 14, int(Coin))
	FillBox(w, 18, 5, 17, 21, 5, 21, int(Wood))
	SetBlock(w, 20, 6, 19, int(Checkpoint))
	SetBlock(w, 19, 7, 20, int(Coin))
	SetBlock(w, 18, 6, 17, int(Spike))
	SetBlock(w, 21, 6, 21, int(Spike))
	FillBox(w, 12, 7, 20, 16, 7, 22, int(Brick))
	FillBox(w, 6, 8, 20, 10, 8, 22, int(Stone))
	FillBox(w, 2, 9, 19, 5, 9, 22, int(Wood))
	SetBlock(w, 3, 10, 20, int(Goal))
	return w, parkourScripts(), "Parkour monetkowy"
}

func buildArena() (World, []ScriptStack, string) {
	w := CreateWorld(SizeX, SizeY, SizeZ)
	FillBox(w, 0, 0, 0, 23, 0, 23, int(Stone))
	FillBox(w, 1, 1, 1, 22, 1, 22, int(Grass))
	FillBox(w, 8, 1, 8, 15, 1, 15, int(Wood))
	SetBlock(w, 12, 2, 12, int(Spawn))
	SetBlock(w, 4, 2, 4, int(Coin))
	SetBlock(w, 19, 2, 4, int(Coin))
	SetBlock(w, 4, 2, 19, int(Coin))
	SetBlock(w, 19, 2, 19, int(Coin))
	SetBlock(w, 12, 2, 20, int(Goal))
	FillBox(w, 10, 2, 10, 14, 4, 10, int(Brick))
	scripts := []ScriptStack{
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatStart, nil),
			MakeNode(SetMessage, &CodeNode{Text: "Zbierz monety w arenie!"}),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatCoin, nil),
			MakeNode(AddScore, &CodeNode{Amount: 10}),
			MakeNode(Sound, &CodeNode{Sound: SndCoin}),
		}},
		{ID: NewID(), Blocks: []CodeNode{
			MakeNode(HatGoal, nil),
			MakeNode(IfScore, &CodeNode{Gte: 40, Body: []CodeNode{
				MakeNode(Win, nil),
				MakeNode(Sound, &CodeNode{Sound: SndWin}),
			}}),
			MakeNode(SetMessage, &CodeNode{Text: "Zbierz wszystkie 4 monety (40 pkt)."}),
		}},
	}
	return w, scripts, "Arena monet"
}

var Templates = []Template{
	{ID: "parkour", Name: "Parkour monetkowy", Blurb: "Skacz, zbieraj monety, unikaj lawy.", Build: buildParkour},
	{ID: "sandbox", Name: "Piaskownica", Blurb: "Płaska trawa — buduj od zera.", Build: buildSandbox},
	{ID: "arena", Name: "Arena monet", Blurb: "Mały dziedziniec z monetami.", Build: buildArena},
}

func LoadTemplate(id string) (World, []ScriptStack, string) {
	for _, t := range Templates {
		if t.ID == id {
			return t.Build()
		}
	}
	return Templates[0].Build()
}
