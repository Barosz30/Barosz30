package render

import "math"

type Camera struct {
	EyeX, EyeY, EyeZ           float64
	OrbitYaw, OrbitPitch, Dist float64
	TargetX, TargetY, TargetZ  float64
}

func NewOrbit(cx, cy, cz float64) Camera {
	return Camera{
		OrbitYaw: 0.7, OrbitPitch: 0.55, Dist: 28,
		TargetX: cx, TargetY: cy, TargetZ: cz,
	}
}

func (c *Camera) SyncEye() {
	cp, sp := math.Cos(c.OrbitPitch), math.Sin(c.OrbitPitch)
	cy, sy := math.Cos(c.OrbitYaw), math.Sin(c.OrbitYaw)
	c.EyeX = c.TargetX + c.Dist*cp*sy
	c.EyeY = c.TargetY + c.Dist*sp
	c.EyeZ = c.TargetZ + c.Dist*cp*cy
}

func (c *Camera) Basis() (fx, fy, fz, rx, ry, rz, ux, uy, uz float64) {
	fx = c.TargetX - c.EyeX
	fy = c.TargetY - c.EyeY
	fz = c.TargetZ - c.EyeZ
	fl := math.Sqrt(fx*fx + fy*fy + fz*fz)
	if fl < 1e-6 {
		fl = 1
	}
	fx, fy, fz = fx/fl, fy/fl, fz/fl
	// right = forward × worldUp?  right = normalize(cross(forward, upWorld)) wait
	// cross(forward, (0,1,0)) = (fz, 0, -fx) ... actually cross(f, up) = (f.z, 0, -f.x)?
	// cross((fx,fy,fz),(0,1,0)) = (fy*0-fz*1, fz*0-fx*0, fx*1-fy*0) = (-fz, 0, fx)
	rx, ry, rz = -fz, 0, fx
	rl := math.Sqrt(rx*rx + ry*ry + rz*rz)
	if rl < 1e-6 {
		rx, ry, rz = 1, 0, 0
		rl = 1
	}
	rx, ry, rz = rx/rl, ry/rl, rz/rl
	// up = right × forward?  up = cross(right, forward) for RH...
	// actually up = cross(right, forward) if forward is look direction:
	// r × f
	ux = ry*fz - rz*fy
	uy = rz*fx - rx*fz
	uz = rx*fy - ry*fx
	return
}

func (c *Camera) Project(x, y, z float64, vw, vh int) (sx, sy, depth float64, ok bool) {
	fx, fy, fz, rx, ry, rz, ux, uy, uz := c.Basis()
	px, py, pz := x-c.EyeX, y-c.EyeY, z-c.EyeZ
	cx := px*rx + py*ry + pz*rz
	cy := px*ux + py*uy + pz*uz
	cz := px*fx + py*fy + pz*fz
	if cz < 0.12 {
		return 0, 0, cz, false
	}
	fov := 55 * math.Pi / 180
	syScale := 1 / math.Tan(fov/2)
	aspect := float64(vw) / math.Max(1, float64(vh))
	sxScale := syScale / aspect
	ndcX := cx * sxScale / cz
	ndcY := cy * syScale / cz
	sx = (ndcX + 1) / 2 * float64(vw)
	sy = (1 - ndcY) / 2 * float64(vh)
	return sx, sy, cz, true
}

func (c *Camera) ScreenRay(mx, my float64, vw, vh int) (ox, oy, oz, dx, dy, dz float64) {
	c.SyncEye()
	fx, fy, fz, rx, ry, rz, ux, uy, uz := c.Basis()
	ndcX := 2*mx/math.Max(1, float64(vw)) - 1
	ndcY := 1 - 2*my/math.Max(1, float64(vh))
	fov := 55 * math.Pi / 180
	syScale := 1 / math.Tan(fov/2)
	aspect := float64(vw) / math.Max(1, float64(vh))
	sxScale := syScale / aspect
	dirX := ndcX/sxScale*rx + ndcY/syScale*ux + fx
	dirY := ndcX/sxScale*ry + ndcY/syScale*uy + fy
	dirZ := ndcX/sxScale*rz + ndcY/syScale*uz + fz
	l := math.Sqrt(dirX*dirX + dirY*dirY + dirZ*dirZ)
	if l < 1e-6 {
		l = 1
	}
	return c.EyeX, c.EyeY, c.EyeZ, dirX / l, dirY / l, dirZ / l
}

func (c *Camera) OrbitDrag(dx, dy float64) {
	c.OrbitYaw -= dx * 0.008
	c.OrbitPitch += dy * 0.008
	if c.OrbitPitch < 0.08 {
		c.OrbitPitch = 0.08
	}
	if c.OrbitPitch > 1.45 {
		c.OrbitPitch = 1.45
	}
}

func (c *Camera) Zoom(delta float64) {
	c.Dist *= math.Pow(1.08, -delta)
	if c.Dist < 6 {
		c.Dist = 6
	}
	if c.Dist > 50 {
		c.Dist = 50
	}
}

func (c *Camera) FollowPlay(px, py, pz, yaw, dt float64) {
	dist, height := 10.0, 5.5
	tx := px + math.Sin(yaw)*dist
	ty := py + height
	tz := pz + math.Cos(yaw)*dist
	k := 1 - math.Pow(0.0008, math.Max(dt, 0.001))
	c.EyeX += (tx - c.EyeX) * k
	c.EyeY += (ty - c.EyeY) * k
	c.EyeZ += (tz - c.EyeZ) * k
	c.TargetX, c.TargetY, c.TargetZ = px, py+0.7, pz
}
