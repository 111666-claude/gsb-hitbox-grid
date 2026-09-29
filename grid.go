// Package hitbox 是碰撞检测的宽相：把圆形碰撞体放进空间哈希网格，查询与谁重叠。
// 缺陷：查询只查中心格（跨格重叠会漏）、同一个 ID 重复插入会留多份、删除只删一份。
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

// Grid 是空间哈希网格。
type Grid struct {
	cell    float64
	buckets map[[2]int][]Circle
}

// New 建网格。
func New(cell float64) *Grid {
	if cell <= 0 {
		cell = 1
	}
	return &Grid{cell: cell, buckets: map[[2]int][]Circle{}}
}

func (g *Grid) key(x, y float64) [2]int {
	return [2]int{int(math.Floor(x / g.cell)), int(math.Floor(y / g.cell))}
}

// Insert 插入一个碰撞体。缺陷：不按 ID 去重。
func (g *Grid) Insert(circle Circle) error {
	if circle.R <= 0 {
		return ErrInput
	}
	cell := g.key(circle.X, circle.Y)
	g.buckets[cell] = append(g.buckets[cell], circle)
	return nil
}

// Remove 删除一个碰撞体。缺陷：只删掉找到的第一份。
func (g *Grid) Remove(id int) bool {
	for cell, items := range g.buckets {
		for index, item := range items {
			if item.ID != id {
				continue
			}
			g.buckets[cell] = append(items[:index], items[index+1:]...)
			return true
		}
	}
	return false
}

func overlap(first, second Circle) bool {
	dx := first.X - second.X
	dy := first.Y - second.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance <= first.R+second.R
}

// Query 查询与 circle 重叠的 ID。缺陷：只查中心所在的那一格，也不去重。
func (g *Grid) Query(circle Circle) []int {
	hits := []int{}
	for _, other := range g.buckets[g.key(circle.X, circle.Y)] {
		if other.ID == circle.ID {
			continue
		}
		if overlap(circle, other) {
			hits = append(hits, other.ID)
		}
	}
	return hits
}
