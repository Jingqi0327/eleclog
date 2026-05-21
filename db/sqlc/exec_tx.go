package db

import (
	"context"
	"fmt"
)

// execTx 执行一个函数内的所有数据库操作，并确保事务的原子性
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	// 1. 开启事务
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	// 2. 将 sqlc 生成的 Queries 与这个事务(tx)绑定
	q := New(tx)

	// 3. 执行回调函数 fn，并将绑定了事务的 q 传进去
	err = fn(q)

	// 4. 错误处理逻辑
	if err != nil {
		// 如果回调函数报错，尝试回滚
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			// 如果回滚也失败了，合并两个错误返回
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	// 5. 如果一切顺利，提交事务
	return tx.Commit(ctx)
}
