// Package service 是业务逻辑层：唯一允许写事务的地方。
// 调用方向约束：api/ws → service → repo，禁止反向与跨层调用（LLM_DEV_GUIDE.md §3.3）。
package service
