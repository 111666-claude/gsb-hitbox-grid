package hitbox

import (
	"math/rand"
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
	a := Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.6}
	b := Circle{ID: 2, X: 1.4, Y: 0.5, R: 0.6}
	grid.Insert(a)
	grid.Insert(b)
	hits := grid.Query(a)
	if len(hits) != 1 || hits[0] != 2 {
		t.Fatalf("跨格重叠应该查到：%v", hits)
	}
}

func TestInsertDuplicateIDKeepsOne(t *testing.T) {
	grid := New(1)
	a := Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.5}
	b := Circle{ID: 2, X: 0.7, Y: 0.5, R: 0.5}
	grid.Insert(a)
	grid.Insert(a)
	grid.Insert(b)
	hits := grid.Query(b)
	if len(hits) != 1 || hits[0] != 1 {
		t.Fatalf("重复注册的 ID 只能查到一次：%v", hits)
	}
	// 重复插入应以后一次为准：挪走之后原位置查不到。
	grid.Insert(Circle{ID: 1, X: 9.5, Y: 9.5, R: 0.5})
	if hits := grid.Query(b); len(hits) != 0 {
		t.Fatalf("重复插入后旧位置不该再命中：%v", hits)
	}
}

func TestRemoveCompletely(t *testing.T) {
	grid := New(1)
	a := Circle{ID: 1, X: 0.5, Y: 0.5, R: 0.5}
	b := Circle{ID: 2, X: 0.7, Y: 0.5, R: 0.5}
	grid.Insert(a)
	grid.Insert(b)
	if !grid.Remove(2) {
		t.Fatal("删除存在的 ID 应该返回 true")
	}
	if hits := grid.Query(a); len(hits) != 0 {
		t.Fatalf("删除后不该再查到：%v", hits)
	}
	if grid.Remove(2) {
		t.Fatal("删除已删除的 ID 应该返回 false")
	}
	// 删除后重新插入同 ID 应恢复可见。
	grid.Insert(b)
	if hits := grid.Query(a); len(hits) != 1 || hits[0] != 2 {
		t.Fatalf("删除后重新插入应该能查到：%v", hits)
	}
}

func TestQueryNoDuplicateIDs(t *testing.T) {
	grid := New(1)
	grid.Insert(Circle{ID: 1, X: 0.5, Y: 0.5, R: 2.5})
	for id := 2; id <= 20; id++ {
		grid.Insert(Circle{ID: id, X: float64(id) * 0.2, Y: 0.5, R: 0.4})
	}
	seen := map[int]bool{}
	for _, id := range grid.Query(Circle{ID: 1, X: 0.5, Y: 0.5, R: 2.5}) {
		if seen[id] {
			t.Fatalf("查询结果里 ID 重复：%d", id)
		}
		seen[id] = true
	}
}

func TestSymmetry(t *testing.T) {
	grid := New(1)
	circles := randomCircles(200, 42)
	for _, c := range circles {
		grid.Insert(c)
	}
	for _, a := range circles {
		hitSet := map[int]bool{}
		for _, id := range grid.Query(a) {
			hitSet[id] = true
		}
		for _, b := range circles {
			if b.ID == a.ID {
				continue
			}
			if hitSet[b.ID] {
				back := map[int]bool{}
				for _, id := range grid.Query(b) {
					back[id] = true
				}
				if !back[a.ID] {
					t.Fatalf("不对称：Query(%d) 命中 %d，但 Query(%d) 没命中 %d", a.ID, b.ID, b.ID, a.ID)
				}
			}
		}
	}
}

func TestBruteForceCompare(t *testing.T) {
	grid := New(1)
	circles := randomCircles(200, 7)
	for _, c := range circles {
		grid.Insert(c)
	}
	for _, a := range circles {
		want := map[int]bool{}
		for _, b := range circles {
			if b.ID != a.ID && overlap(a, b) {
				want[b.ID] = true
			}
		}
		got := map[int]bool{}
		for _, id := range grid.Query(a) {
			if got[id] {
				t.Fatalf("Query(%d) 返回重复 ID %d", a.ID, id)
			}
			got[id] = true
		}
		if len(got) != len(want) {
			t.Fatalf("Query(%d) 命中数 %d，暴力结果 %d", a.ID, len(got), len(want))
		}
		for id := range want {
			if !got[id] {
				t.Fatalf("Query(%d) 漏了 %d", a.ID, id)
			}
		}
	}
}

func TestRepeatableResults(t *testing.T) {
	build := func() *Grid {
		grid := New(1)
		for _, c := range randomCircles(200, 99) {
			grid.Insert(c)
		}
		return grid
	}
	first := build()
	second := build()
	for _, c := range randomCircles(200, 99) {
		a, b := first.Query(c), second.Query(c)
		if len(a) != len(b) {
			t.Fatalf("重复执行结果不同：%v vs %v", a, b)
		}
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("重复执行结果不同：%v vs %v", a, b)
			}
		}
	}
}

// randomCircles 用确定性伪随机序列生成 ID 从 1 开始的圆。
func randomCircles(n int, seed int64) []Circle {
	rng := rand.New(rand.NewSource(seed))
	circles := make([]Circle, 0, n)
	for id := 1; id <= n; id++ {
		circles = append(circles, Circle{
			ID: id,
			X:  rng.Float64() * 20,
			Y:  rng.Float64() * 20,
			R:  0.2 + rng.Float64()*1.5,
		})
	}
	return circles
}
