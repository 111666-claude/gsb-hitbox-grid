package hitbox

import "testing"

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
