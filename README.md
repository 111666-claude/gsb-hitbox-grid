# hitbox-grid

碰撞检测宽相（空间哈希网格），只用 Go 标准库。

```
go test ./...
go run ./cmd/hitbox-grid --sample cross
go run ./cmd/hitbox-grid --sample dup
go run ./cmd/hitbox-grid --sample remove
```

## 口径（README 为准）

- **查询要看邻居格**：候选格是所有与「圆心 ± 半径」相交的格子。`--sample cross`（两个半径 0.6、
  圆心在 (0.5,0.5) 与 (1.4,0.5) 的圆，网格边长 1）按口径是 `pairs=1`，现在只查中心格，给 `pairs=0`。
- **按 ID 去重**：同一个 ID 重复 Insert 只能留一份。`--sample dup` 按口径是 `dup_pairs=0`，
  现在留了两份，给 `dup_pairs=1`。
- **删除要删干净**：`--sample remove` 按口径是 `still_visible=0`，现在只删掉一份，给 `still_visible=1`。
- 不变量：一次查询返回的 ID 互不重复；`Query(A)` 命中 B 时 `Query(B)` 也要命中 A（对称）；
  存入后再删除的实体永远查不到。
- 规模：10 万个碰撞体、每帧 10 万次查询 1 秒内完成，内存 O(实体数)；查询只扫与半径矩形相交的格子，
  不许全表扫描。

## 现在的行为

- `Query` 只取 `g.key(circle.X, circle.Y)` 一个桶：跨格重叠直接漏掉。
- `Insert` 不看 ID 是否已存在：重复插入留多份，查询会返回重复 ID。
- `Remove` 找到第一份就返回，重复插入留下的副本还在。

## 计划新增的能力

```
--sample cross|dup|remove
--cell N
```

## 输出

```
pairs=1 hits=[2]
hits=1 dup_pairs=0
still_visible=0
```

## 目录

```
grid.go              空间哈希宽相
cmd/hitbox-grid      命令行入口
grid_test.go         go test 用例
```
