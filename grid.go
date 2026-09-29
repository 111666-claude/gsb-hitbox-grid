// Package hitbox 是碰撞检测的宽相：把圆形碰撞体放进空间哈希网格，查询与谁重叠。
package hitbox

import (
	"errors"
	"math"
	"sort"
)

// ErrInput 表示输入不合法。
var ErrInput = errors.New("hitbox: 输入不合法")

// Circle 是一个圆形碰撞体。
type Circle struct {
	ID      int
	X, Y, R float64
}

// Grid 是空间哈希网格。每个实体按「圆心 ± 半径」矩形覆盖的所有格子登记，
// 另用 cells 记录 ID 到格子范围的映射，保证按 ID 去重、删除删干净。
type Grid struct {
	cell    float64
	buckets map[[2]int]map[int]Circle
	cells   map[int]cellRange
}

// cellRange 是一个实体的包围盒覆盖的格子范围（闭区间）。
type cellRange struct {
	min, max [2]int
}

// New 建网格。
func New(cell float64) *Grid {
	if cell <= 0 {
		cell = 1
	}
	return &Grid{
		cell:    cell,
		buckets: map[[2]int]map[int]Circle{},
		cells:   map[int]cellRange{},
	}
}

func (g *Grid) key(x, y float64) [2]int {
	return [2]int{int(math.Floor(x / g.cell)), int(math.Floor(y / g.cell))}
}

// span 返回「圆心 ± 半径」矩形覆盖的格子范围。
func (g *Grid) span(circle Circle) cellRange {
	return cellRange{
		min: g.key(circle.X-circle.R, circle.Y-circle.R),
		max: g.key(circle.X+circle.R, circle.Y+circle.R),
	}
}

func (g *Grid) unlink(id int, span cellRange) {
	for cx := span.min[0]; cx <= span.max[0]; cx++ {
		for cy := span.min[1]; cy <= span.max[1]; cy++ {
			cell := [2]int{cx, cy}
			delete(g.buckets[cell], id)
			if len(g.buckets[cell]) == 0 {
				delete(g.buckets, cell)
			}
		}
	}
}

// Insert 插入一个碰撞体。同一个 ID 重复 Insert 只保留最新的一份。
func (g *Grid) Insert(circle Circle) error {
	if circle.R <= 0 {
		return ErrInput
	}
	if old, ok := g.cells[circle.ID]; ok {
		g.unlink(circle.ID, old)
	}
	span := g.span(circle)
	for cx := span.min[0]; cx <= span.max[0]; cx++ {
		for cy := span.min[1]; cy <= span.max[1]; cy++ {
			cell := [2]int{cx, cy}
			bucket, ok := g.buckets[cell]
			if !ok {
				bucket = map[int]Circle{}
				g.buckets[cell] = bucket
			}
			bucket[circle.ID] = circle
		}
	}
	g.cells[circle.ID] = span
	return nil
}

// Remove 删除一个碰撞体，删完后任何查询都不会再返回它。
func (g *Grid) Remove(id int) bool {
	span, ok := g.cells[id]
	if !ok {
		return false
	}
	g.unlink(id, span)
	delete(g.cells, id)
	return true
}

func overlap(first, second Circle) bool {
	dx := first.X - second.X
	dy := first.Y - second.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance <= first.R+second.R
}

// Query 查询与 circle 重叠的 ID。候选格是「圆心 ± 半径」矩形覆盖的所有格子，
// 结果按 ID 升序、无重复。
func (g *Grid) Query(circle Circle) []int {
	span := g.span(circle)
	found := map[int]bool{}
	for cx := span.min[0]; cx <= span.max[0]; cx++ {
		for cy := span.min[1]; cy <= span.max[1]; cy++ {
			for id, other := range g.buckets[[2]int{cx, cy}] {
				if id == circle.ID || found[id] {
					continue
				}
				if overlap(circle, other) {
					found[id] = true
				}
			}
		}
	}
	hits := make([]int, 0, len(found))
	for id := range found {
		hits = append(hits, id)
	}
	sort.Ints(hits)
	return hits
}
