package hitbox

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestSingleCellOverlap(t *testing.T) {
	grid := New(1)
	grid.Insert(Circle{ID: 1, X: 0.2, Y: 0.2, R: 0.3})
	grid.Insert(Circle{ID: 2, X: 0.4, Y: 0.2, R: 0.3})
	hits := grid.Query(Circle{ID: 1, X: 0.2, Y: 0.2, R: 0.3})
	if len(hits) != 1 || hits[0] != 2 {
		t.Fatalf("同一格内的重叠应该查到：%v", hits)
	}
}

func TestFarApartDoesNotOverlap(t *testing.T) {
	grid := New(1)
	grid.Insert(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.2})
	grid.Insert(Circle{ID: 2, X: 9.5, Y: 9.5, R: 0.2})
	if hits := grid.Query(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.2}); len(hits) != 0 {
		t.Fatalf("离得远不该命中：%v", hits)
	}
}

func TestBadRadius(t *testing.T) {
	if err := New(1).Insert(Circle{ID: 1, X: 0, Y: 0, R: 0}); err != ErrInput {
		t.Fatalf("半径非正应该报错：%v", err)
	}
}

func TestRemoveMissing(t *testing.T) {
	if New(1).Remove(42) {
		t.Fatal("删除不存在的 ID 应该返回 false")
	}
}

func TestCrossCellOverlap(t *testing.T) {
	grid := New(1)
	grid.Insert(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.6})
	grid.Insert(Circle{ID: 2, X: 1.4, Y: 0.5, R: 0.6})
	hits := grid.Query(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.6})
	if len(hits) != 1 || hits[0] != 2 {
		t.Fatalf("跨格重叠应该查到：%v", hits)
	}
}

func TestInsertSameIDDedup(t *testing.T) {
	grid := New(1)
	grid.Insert(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.6})
	grid.Insert(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.6})
	grid.Insert(Circle{ID: 2, X: 0.9, Y: 0.5, R: 0.6})
	hits := grid.Query(Circle{ID: 2, X: 0.9, Y: 0.5, R: 0.6})
	if len(hits) != 1 || hits[0] != 1 {
		t.Fatalf("重复注册只该留一份：%v", hits)
	}
}

func TestInsertSameIDMovesToNewCell(t *testing.T) {
	grid := New(1)
	grid.Insert(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.2})
	grid.Insert(Circle{ID: 1, X: 5.5, Y: 5.5, R: 0.2})
	if hits := grid.Query(Circle{ID: 2, X: 0.5, Y: 0.5, R: 0.2}); len(hits) != 0 {
		t.Fatalf("旧位置不该再命中：%v", hits)
	}
	if hits := grid.Query(Circle{ID: 2, X: 5.6, Y: 5.5, R: 0.2}); len(hits) != 1 {
		t.Fatalf("新位置应该命中：%v", hits)
	}
}

func TestRemoveCleansAllCopies(t *testing.T) {
	grid := New(1)
	grid.Insert(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.6})
	grid.Insert(Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.6})
	if !grid.Remove(1) {
		t.Fatal("删除已注册的 ID 应该返回 true")
	}
	if hits := grid.Query(Circle{ID: 2, X: 0.5, Y: 0.5, R: 0.6}); len(hits) != 0 {
		t.Fatalf("删除后不该再查到：%v", hits)
	}
	if grid.Remove(1) {
		t.Fatal("重复删除应该返回 false")
	}
}

func TestQuerySymmetry(t *testing.T) {
	grid := New(1)
	a := Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.6}
	b := Circle{ID: 2, X: 1.4, Y: 0.5, R: 0.6}
	grid.Insert(a)
	grid.Insert(b)
	if hits := grid.Query(a); len(hits) != 1 || hits[0] != 2 {
		t.Fatalf("Query(A) 应该命中 B：%v", hits)
	}
	if hits := grid.Query(b); len(hits) != 1 || hits[0] != 1 {
		t.Fatalf("Query(B) 应该命中 A：%v", hits)
	}
}

// 随机 200 组位置与半径，与全量两两判断的暴力实现比对命中集合（顺序无关）。
func TestAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for trial := 0; trial < 200; trial++ {
		grid := New(1)
		circles := make([]Circle, 0, 32)
		count := 1 + rng.Intn(32)
		for i := 0; i < count; i++ {
			circle := Circle{
				ID: i + 1,
				X:  rng.Float64() * 20,
				Y:  rng.Float64() * 20,
				R:  0.1 + rng.Float64()*1.5,
			}
			if err := grid.Insert(circle); err != nil {
				t.Fatalf("插入失败：%v", err)
			}
			circles = append(circles, circle)
		}
		for _, circle := range circles {
			got := map[int]bool{}
			for _, id := range grid.Query(circle) {
				if got[id] {
					t.Fatalf("第 %d 组：查询结果里 ID %d 重复", trial, id)
				}
				got[id] = true
			}
			want := map[int]bool{}
			for _, other := range circles {
				if other.ID == circle.ID {
					continue
				}
				dx := circle.X - other.X
				dy := circle.Y - other.Y
				reach := circle.R + other.R
				if dx*dx+dy*dy <= reach*reach {
					want[other.ID] = true
				}
			}
			if len(got) != len(want) {
				t.Fatalf("第 %d 组 ID %d：命中数不同 got=%v want=%v", trial, circle.ID, got, want)
			}
			for id := range want {
				if !got[id] {
					t.Fatalf("第 %d 组 ID %d：漏了 %d", trial, circle.ID, id)
				}
			}
		}
	}
}

// 同一批插入与查询重复执行，结果必须一致。
func TestDeterministic(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	circles := make([]Circle, 0, 64)
	for i := 0; i < 64; i++ {
		circles = append(circles, Circle{
			ID: i + 1,
			X:  rng.Float64() * 50,
			Y:  rng.Float64() * 50,
			R:  0.1 + rng.Float64(),
		})
	}
	build := func() *Grid {
		grid := New(1)
		for _, circle := range circles {
			grid.Insert(circle)
		}
		return grid
	}
	first := build()
	second := build()
	for _, circle := range circles {
		a := first.Query(circle)
		b := second.Query(circle)
		sort.Ints(a)
		sort.Ints(b)
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("ID %d 两次查询结果不同：%v vs %v", circle.ID, a, b)
		}
	}
}
