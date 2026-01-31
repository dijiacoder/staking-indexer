package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:           "./internal/gen/query",
		Mode:              gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldCoverable:    true,
		FieldSignable:     true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		WithUnitTest:      true,
	})

	dsn := "root:st123456@tcp(localhost:4306)/stake_db?charset=utf8mb4&parseTime=True&loc=Local"
	gormdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	g.UseDB(gormdb)

	// 已有的表模型生成
	g.ApplyBasic(
		g.GenerateModel("chain_blocks"),
		g.GenerateModel("chain_scan_cursor"),
		g.GenerateModel("staking_events"),
		g.GenerateModel("staking_pools"),
		g.GenerateModel("staking_user_positions"),
	)

	g.Execute()
}
