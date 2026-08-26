package engine

import "math"

type VoxelHit struct {
	Hit  *[3]int
	Prev *[3]int
}

func VoxelRaycast(w World, ox, oy, oz, dx, dy, dz, maxDist float64) VoxelHit {
	x, y, z := int(math.Floor(ox)), int(math.Floor(oy)), int(math.Floor(oz))
	stepX, stepY, stepZ := 0, 0, 0
	if dx > 0 {
		stepX = 1
	} else if dx < 0 {
		stepX = -1
	}
	if dy > 0 {
		stepY = 1
	} else if dy < 0 {
		stepY = -1
	}
	if dz > 0 {
		stepZ = 1
	} else if dz < 0 {
		stepZ = -1
	}

	tDeltaX, tDeltaY, tDeltaZ := math.Inf(1), math.Inf(1), math.Inf(1)
	if dx != 0 {
		tDeltaX = math.Abs(1 / dx)
	}
	if dy != 0 {
		tDeltaY = math.Abs(1 / dy)
	}
	if dz != 0 {
		tDeltaZ = math.Abs(1 / dz)
	}

	frac := func(v float64, step int) float64 {
		f := v - math.Floor(v)
		if step > 0 {
			return 1 - f
		}
		return f
	}

	tMaxX, tMaxY, tMaxZ := math.Inf(1), math.Inf(1), math.Inf(1)
	if stepX != 0 {
		tMaxX = frac(ox, stepX) * tDeltaX
	}
	if stepY != 0 {
		tMaxY = frac(oy, stepY) * tDeltaY
	}
	if stepZ != 0 {
		tMaxZ = frac(oz, stepZ) * tDeltaZ
	}

	var prev *[3]int
	t := 0.0
	for i := 0; i < 96 && t <= maxDist; i++ {
		if InBounds(w, x, y, z) && GetBlock(w, x, y, z) != int(Air) {
			h := [3]int{x, y, z}
			return VoxelHit{Hit: &h, Prev: prev}
		}
		cp := [3]int{x, y, z}
		prev = new([3]int)
		*prev = cp
		if tMaxX < tMaxY && tMaxX < tMaxZ {
			t = tMaxX
			x += stepX
			tMaxX += tDeltaX
		} else if tMaxY < tMaxZ {
			t = tMaxY
			y += stepY
			tMaxY += tDeltaY
		} else {
			t = tMaxZ
			z += stepZ
			tMaxZ += tDeltaZ
		}
	}
	return VoxelHit{}
}
