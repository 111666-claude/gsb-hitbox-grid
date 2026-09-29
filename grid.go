// Package hitbox 是碰撞检测的宽相：把圆形碰撞体放进空间哈希网格，查询与谁重叠。
package hitbox

import (
	"errors"
	"math"
)

// ErrInput 表示输入不合法。
var ErrInput = errors.New("hitbox: 输入不合法")

// Circle 是一个圆形碰撞体。
type Circle struct {
	ID      int
	X, Y, R float64
}

// Grid 是空间哈希网格。同一个 ID 只保留一份；
// 每个圆铺进「圆心 ± 半径」覆盖到的所有格子。
type Grid struct {
	cell    float64
	buckets map[[2]int][]Circle
	members map[int]Circle // ID -> 当前保存的圆
}

// New 建网格。
func New(cell float64) *Grid {
	if cell <= 0 {
		cell = 1
	}
	return &Grid{
		cell:    cell,
		buckets: map[[2]int][]Circle{},
		members: map[int]Circle{},
	}
}

func (g *Grid) key(x, y float64) [2]int {
	return [2]int{int(math.Floor(x / g.cell)), int(math.Floor(y / g.cell))}
}

// cells 返回「圆心 ± 半径」矩形覆盖到的所有格子范围（含端点）。
func (g *Grid) cells(circle Circle) (lo, hi [2]int) {
	return g.key(circle.X-circle.R, circle.Y-circle.R),
		g.key(circle.X+circle.R, circle.Y+circle.R)
}

// Insert 插入一个碰撞体。同一个 ID 重复插入只保留最新的一份。
func (g *Grid) Insert(circle Circle) error {
	if circle.R <= 0 {
		return ErrInput
	}
	if _, ok := g.members[circle.ID]; ok {
		g.Remove(circle.ID)
	}
	lo, hi := g.cells(circle)
	for cx := lo[0]; cx <= hi[0]; cx++ {
		for cy := lo[1]; cy <= hi[1]; cy++ {
			cell := [2]int{cx, cy}
			g.buckets[cell] = append(g.buckets[cell], circle)
		}
	}
	g.members[circle.ID] = circle
	return nil
}

// Remove 删除一个碰撞体，删干净后任何查询都不会再返回它。
func (g *Grid) Remove(id int) bool {
	circle, ok := g.members[id]
	if !ok {
		return false
	}
	delete(g.members, id)
	lo, hi := g.cells(circle)
	for cx := lo[0]; cx <= hi[0]; cx++ {
		for cy := lo[1]; cy <= hi[1]; cy++ {
			cell := [2]int{cx, cy}
			items := g.buckets[cell]
			for index, item := range items {
				if item.ID != id {
					continue
				}
				items = append(items[:index], items[index+1:]...)
				break
			}
			if len(items) == 0 {
				delete(g.buckets, cell)
			} else {
				g.buckets[cell] = items
			}
		}
	}
	return true
}

func overlap(first, second Circle) bool {
	dx := first.X - second.X
	dy := first.Y - second.Y
	reach := first.R + second.R
	return dx*dx+dy*dy <= reach*reach
}

// Query 查询与 circle 重叠的 ID。只扫「圆心 ± 半径」矩形相交的格子，
// 结果里的 ID 互不重复。
func (g *Grid) Query(circle Circle) []int {
	hits := []int{}
	seen := map[int]struct{}{}
	lo, hi := g.cells(circle)
	for cx := lo[0]; cx <= hi[0]; cx++ {
		for cy := lo[1]; cy <= hi[1]; cy++ {
			for _, other := range g.buckets[[2]int{cx, cy}] {
				if other.ID == circle.ID {
					continue
				}
				if _, ok := seen[other.ID]; ok {
					continue
				}
				seen[other.ID] = struct{}{}
				if overlap(circle, other) {
					hits = append(hits, other.ID)
				}
			}
		}
	}
	return hits
}
